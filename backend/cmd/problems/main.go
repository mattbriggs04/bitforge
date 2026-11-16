package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mattbriggs04/bitforge/backend/internal/authoring"
	"github.com/mattbriggs04/bitforge/backend/internal/config"
	"github.com/mattbriggs04/bitforge/backend/internal/db"
	"github.com/mattbriggs04/bitforge/backend/internal/sandbox"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: problems prompt -config request.json | validate/import -bundle draft.json | validate-interview/import-interview -bundle interview.json")
	}
	action := os.Args[1]
	fs := flag.NewFlagSet(action, flag.ContinueOnError)
	configPath := fs.String("config", "", "generation request JSON")
	bundlePath := fs.String("bundle", "", "draft bundle JSON")
	out := fs.String("out", "", "write prompt/report to file (default stdout)")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}
	if action == "prompt" {
		data, err := os.ReadFile(*configPath)
		if err != nil {
			return err
		}
		var c authoring.Config
		if err = authoring.Decode(data, &c); err != nil {
			return err
		}
		if c.Seed == 0 {
			n, err := rand.Int(rand.Reader, big.NewInt(1<<62))
			if err != nil {
				return err
			}
			c.Seed = n.Int64() + 1
		}
		prompt, err := authoring.Prompt(c)
		if err != nil {
			return err
		}
		return write(*out, []byte(prompt))
	}

	isInterview := action == "validate-interview" || action == "import-interview"
	if action != "validate" && action != "import" && !isInterview {
		return fmt.Errorf("unknown action %s", action)
	}
	var interview authoring.Interview
	var bundles []authoring.Bundle
	if isInterview {
		var err error
		interview, bundles, err = authoring.LoadInterview(*bundlePath)
		if err != nil {
			return err
		}
	} else {
		data, err := os.ReadFile(*bundlePath)
		if err != nil {
			return err
		}
		if len(data) > 2*1024*1024 {
			return fmt.Errorf("bundle exceeds 2 MiB")
		}
		var bundle authoring.Bundle
		if err = authoring.Decode(data, &bundle); err != nil {
			return err
		}
		if err = bundle.Check(); err != nil {
			return err
		}
		bundles = []authoring.Bundle{bundle}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	executor := sandbox.NewDocker()
	// Resolve the mutable local tag to an immutable image ID before validation.
	imageOut, err := exec.CommandContext(ctx, "docker", "image", "inspect", "--format", "{{.Id}}", executor.Image).CombinedOutput()
	if err != nil {
		return fmt.Errorf("sandbox image unavailable; start Docker and run make sandbox-image: %s", imageOut)
	}
	executor.Image = strings.TrimSpace(string(imageOut))
	reports := make([]authoring.Report, 0, len(bundles))
	for index, bundle := range bundles {
		if isInterview {
			fmt.Fprintf(os.Stderr, "Validating %d/%d: %s\n", index+1, len(bundles), bundle.Title)
		}
		report, err := authoring.Validate(ctx, bundle, executor, executor.Image)
		if err != nil {
			return fmt.Errorf("draft %s rejected: %w", bundle.Slug, err)
		}
		reports = append(reports, report)
	}

	if action == "import" || action == "import-interview" {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		conn, err := db.OpenDatabase(ctx, cfg.DatabaseURL)
		if err != nil {
			return err
		}
		defer conn.Close()
		if err = db.RunMigrations(ctx, conn); err != nil {
			return err
		}
		if isInterview {
			err = db.ImportInterview(ctx, conn, interview, bundles, reports)
		} else {
			err = db.ImportGenerated(ctx, conn, bundles[0], reports[0])
		}
		if err != nil {
			return err
		}
	}
	if isInterview {
		data, err := json.MarshalIndent(map[string]any{"interview": interview.Slug, "reports": reports}, "", "  ")
		if err != nil {
			return err
		}
		return write(*out, append(data, '\n'))
	}
	return write(*out, reports[0].JSON())
}
func write(path string, data []byte) error {
	if path == "" {
		_, err := os.Stdout.Write(data)
		return err
	}
	return os.WriteFile(path, data, 0600)
}
