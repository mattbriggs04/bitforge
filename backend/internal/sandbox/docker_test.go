package sandbox

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func integration(t *testing.T) *Docker {
	t.Helper()
	if os.Getenv("BITFORGE_SANDBOX_TEST") != "1" {
		t.Skip("set BITFORGE_SANDBOX_TEST=1 to test Docker isolation")
	}
	return NewDocker()
}
func TestDockerLanguages(t *testing.T) {
	d := integration(t)
	for lang, source := range map[string]string{
		"c":      "#include <stdio.h>\nint main(){int x;scanf(\"%d\",&x);printf(\"%d\\n\",x*x);}",
		"cpp":    "#include <iostream>\nint main(){int x;std::cin>>x;std::cout<<x*x;}",
		"python": "import sys, numpy as np\nprint(int(np.square(int(sys.stdin.read()))))\n",
		"rust":   "use std::io::{self,Read};fn main(){let mut s=String::new();io::stdin().read_to_string(&mut s).unwrap();let x:i32=s.trim().parse().unwrap();println!(\"{}\",x*x);}",
	} {
		t.Run(lang, func(t *testing.T) {
			p, out, err := d.Prepare(context.Background(), lang, source, 20*time.Second)
			if err != nil || p == nil {
				t.Fatalf("prepare %+v %v", out, err)
			}
			defer p.Close()
			for i := 0; i < 2; i++ {
				out, err = p.Run(context.Background(), "7", 3*time.Second)
				if err != nil || out.ExitCode != 0 || strings.TrimSpace(out.Stdout) != "49" {
					t.Fatalf("run %+v %v", out, err)
				}
			}
		})
	}
}
func TestDockerIsolationAndLimits(t *testing.T) {
	d := integration(t)
	p, out, err := d.Prepare(context.Background(), "python", `import os, socket
assert os.getuid() == 65534
for path in ('/src/solution.py', '/etc/bitforge-write-test'):
    try:
        open(path,'w').write('bad')
    except OSError:
        pass
    else:
        raise AssertionError('filesystem writable')
assert not os.path.exists('/var/run/docker.sock')
assert len(open("/proc/net/route").read().strip().splitlines()) == 1
print('isolated')
`, 20*time.Second)
	if err != nil || p == nil {
		t.Fatalf("prepare %+v %v", out, err)
	}
	defer p.Close()
	out, err = p.Run(context.Background(), "", 3*time.Second)
	if err != nil || out.ExitCode != 0 || strings.TrimSpace(out.Stdout) != "isolated" {
		t.Fatalf("isolation %+v %v", out, err)
	}
	for _, source := range []string{"while True: print('x'*10000)", "import os,time\nif os.fork()==0: time.sleep(60)\nelse: time.sleep(60)"} {
		p, out, err := d.Prepare(context.Background(), "python", source, 20*time.Second)
		if err != nil || p == nil {
			t.Fatal(out, err)
		}
		out, err = p.Run(context.Background(), "", time.Second)
		p.Close()
		if err != nil || (!out.TimedOut && !out.Truncated) {
			t.Fatalf("limits %+v %v", out, err)
		}
	}
	remaining, err := exec.Command("docker", "ps", "-q", "--filter", "name=bitforge-").Output()
	if err != nil || strings.TrimSpace(string(remaining)) != "" {
		t.Fatalf("containers left running: %s %v", remaining, err)
	}
}
