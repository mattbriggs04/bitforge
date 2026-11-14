// Package sandbox compiles and executes untrusted programs in disposable containers.
package sandbox

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const OutputLimit = 64 * 1024

type Output struct {
	Stdout, Stderr      string
	ExitCode            int
	TimedOut, Truncated bool
}
type Program interface {
	Run(context.Context, string, time.Duration) (Output, error)
	Close()
}
type Executor interface {
	Prepare(context.Context, string, string, time.Duration) (Program, Output, error)
}
type Docker struct{ Image, Runtime string }

func NewDocker() *Docker {
	image := os.Getenv("SANDBOX_IMAGE")
	if image == "" {
		image = "bitforge-runner:v1"
	}
	return &Docker{Image: image, Runtime: os.Getenv("SANDBOX_RUNTIME")}
}

type program struct {
	docker  *Docker
	dir     string
	command []string
}

func (p *program) Close() { _ = os.RemoveAll(p.dir) }
func (p *program) Run(ctx context.Context, input string, timeout time.Duration) (Output, error) {
	return p.docker.command(ctx, p.dir, p.command, input, timeout, false)
}
func (d *Docker) Prepare(ctx context.Context, language, source string, timeout time.Duration) (Program, Output, error) {
	var file string
	var compile, run []string
	switch language {
	case "c":
		file = "solution.c"
		compile = []string{"gcc", "-std=c11", "-O2", "-pthread", file, "-lm", "-o", "program"}
		run = []string{"/src/program"}
	case "cpp":
		file = "solution.cpp"
		compile = []string{"g++", "-std=c++20", "-O2", "-pthread", file, "-o", "program"}
		run = []string{"/src/program"}
	case "rust":
		file = "solution.rs"
		compile = []string{"rustc", "--edition=2021", "-O", "-C", "strip=symbols", file, "-o", "program"}
		run = []string{"/src/program"}
	case "python":
		file = "solution.py"
		compile = []string{"python3", "-I", "-c", "import ast; ast.parse(open('solution.py').read())"}
		run = []string{"python3", "-I", "-B", "/src/solution.py"}
	default:
		return nil, Output{}, fmt.Errorf("unsupported language %q", language)
	}
	dir, err := os.MkdirTemp("", "bitforge-program-")
	if err != nil {
		return nil, Output{}, err
	}
	p := &program{docker: d, dir: dir, command: run}
	if err = os.Chmod(dir, 0755); err != nil {
		p.Close()
		return nil, Output{}, err
	}
	if err = os.WriteFile(filepath.Join(dir, file), []byte(source), 0644); err != nil {
		p.Close()
		return nil, Output{}, err
	}
	script := `cp /src/* /work/ && exec "$@"`
	if language != "python" {
		script = `cp /src/* /work/ && "$@" >&2 && base64 /work/program`
	}
	out, err := d.command(ctx, dir, append([]string{"sh", "-c", script, "compile"}, compile...), "", timeout, language != "python")
	if err != nil || out.ExitCode != 0 || out.TimedOut || out.Truncated {
		p.Close()
		if language != "python" {
			out.Stdout = ""
		}
		return nil, out, err
	}
	if language != "python" {
		binary, err := io.ReadAll(base64.NewDecoder(base64.StdEncoding, strings.NewReader(out.Stdout)))
		out.Stdout = ""
		if err != nil || len(binary) > 16*1024*1024 {
			p.Close()
			return nil, out, fmt.Errorf("invalid or oversized compiled artifact")
		}
		if err = os.WriteFile(filepath.Join(dir, "program"), binary, 0555); err != nil {
			p.Close()
			return nil, out, err
		}
	}
	return p, out, nil
}

// The cap applies while reading, not after buffering an arbitrarily large output.
type cappedWriter struct {
	mu        sync.Mutex
	data      bytes.Buffer
	remaining int
	exceeded  bool
	cancel    context.CancelFunc
}

func (w *cappedWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n := len(b)
	if len(b) > w.remaining {
		b = b[:w.remaining]
		w.exceeded = true
		w.cancel()
	}
	w.data.Write(b)
	w.remaining -= len(b)
	return n, nil
}
func (w *cappedWriter) String() string { w.mu.Lock(); defer w.mu.Unlock(); return w.data.String() }
func (w *cappedWriter) Exceeded() bool { w.mu.Lock(); defer w.mu.Unlock(); return w.exceeded }
func (d *Docker) command(parent context.Context, dir string, command []string, input string, timeout time.Duration, captureBinary bool) (Output, error) {
	if timeout <= 0 || timeout > 30*time.Second {
		timeout = 10 * time.Second
	}
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return Output{}, err
	}
	name := "bitforge-" + hex.EncodeToString(id[:])
	args := []string{"run", "--rm", "--pull=never", "--name", name, "--network=none", "--user=65534:65534", "--read-only", "--cap-drop=ALL", "--security-opt=no-new-privileges", "--memory=256m", "--memory-swap=256m", "--cpus=1", "--pids-limit=64", "--ulimit", "fsize=16777216:16777216", "--ulimit", "nofile=128:128", "--tmpfs", "/tmp:rw,noexec,nosuid,size=64m", "--mount", "type=bind,src=" + dir + ",dst=/src,readonly", "--tmpfs", "/work:rw,noexec,nosuid,size=64m,mode=1777", "--workdir=/work", "-i"}
	if d.Runtime != "" {
		args = append(args, "--runtime", d.Runtime)
	}
	args = append(args, d.Image)
	// The daemon-side deadline survives an abrupt worker crash. When timeout
	// exits, Docker tears down the container's remaining process tree.
	args = append(args, "timeout", "--signal=KILL", fmt.Sprintf("%.3fs", timeout.Seconds()))
	args = append(args, command...)
	deadline, cancelDeadline := context.WithTimeout(parent, timeout)
	defer cancelDeadline()
	ctx, cancel := context.WithCancel(deadline)
	defer cancel()
	stdoutLimit := OutputLimit
	if captureBinary {
		stdoutLimit = 24 * 1024 * 1024
	}
	stdout := &cappedWriter{remaining: stdoutLimit, cancel: cancel}
	stderr := &cappedWriter{remaining: OutputLimit, cancel: cancel}
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.WaitDelay = time.Second
	err := cmd.Run()
	// Killing the docker client alone does not kill the container/process tree.
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cleanupCancel()
	cleanup := exec.CommandContext(cleanupCtx, "docker", "rm", "-f", name)
	cleanup.Stdout = io.Discard
	cleanup.Stderr = io.Discard
	_ = cleanup.Run()
	out := Output{Stdout: stdout.String(), Stderr: stderr.String(), TimedOut: errors.Is(deadline.Err(), context.DeadlineExceeded), Truncated: stdout.Exceeded() || stderr.Exceeded()}
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			out.ExitCode = exit.ExitCode()
			if out.ExitCode == 125 || out.ExitCode == 126 || out.ExitCode == 127 {
				return out, fmt.Errorf("sandbox unavailable: %s", strings.TrimSpace(out.Stderr))
			}
		} else {
			return out, fmt.Errorf("start Docker sandbox: %w", err)
		}
	}
	if parent.Err() != nil {
		return out, parent.Err()
	}
	return out, nil
}
