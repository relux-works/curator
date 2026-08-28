// Command swiftpm_stub is the deterministic cross-platform stand-in for the
// selected SwiftPM driver in this package's tests. The previous stand-in was a
// POSIX shell script, which Windows cannot execute; this binary reproduces the
// same observable effect from a declarative scenario instead of shell text.
//
// It reads swiftpm-scenario.json from its own directory and applies each
// action relative to the process working directory — the exact working
// directory the committed permit binds. The literal token {{PWD}} inside a
// payload is replaced with that absolute working directory, which is how the
// compiler dependency files spell absolute source paths.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type action struct {
	// Op is mkdir, write, chmod-readonly-exec, remove, or remove-all.
	Op string `json:"op"`
	// Path is slash-separated and relative to the working directory.
	Path string `json:"path"`
	// Payload is the file content for write, with {{PWD}} expanded.
	Payload string `json:"payload,omitempty"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	scenarioBytes, err := os.ReadFile(filepath.Join(filepath.Dir(self), "swiftpm-scenario.json"))
	if err != nil {
		return err
	}
	var actions []action
	if err := json.Unmarshal(scenarioBytes, &actions); err != nil {
		return err
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		return err
	}
	for _, step := range actions {
		target := filepath.FromSlash(step.Path)
		switch step.Op {
		case "mkdir":
			if err := os.MkdirAll(target, 0o700); err != nil {
				return err
			}
		case "write":
			payload := strings.ReplaceAll(step.Payload, "{{PWD}}", filepath.ToSlash(workingDirectory))
			if err := os.WriteFile(target, []byte(payload), 0o600); err != nil {
				return err
			}
		case "chmod-readonly-exec":
			if err := os.Chmod(target, 0o500); err != nil {
				return err
			}
		case "remove":
			if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
				return err
			}
		case "remove-all":
			if err := os.RemoveAll(target); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown scenario op %q", step.Op)
		}
	}
	return nil
}
