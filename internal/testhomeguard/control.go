package testhomeguard

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const controlMode = "CURATOR_TEST_NATIVE_GUARD_CONTROL"
const controlOrigin = "CURATOR_TEST_NATIVE_GUARD_ORIGIN"

// Control proves the actual package-exit guard detects a bounded write through
// an unpinned HOME. Both child original homes are temporary fixture directories;
// the plant never selects the OS account home. The paired isolated write passes.
func Control(t *testing.T, childTest string) {
	t.Helper()
	for _, mode := range []string{"isolated", "plant"} {
		t.Run(mode, func(t *testing.T) {
			origin := t.TempDir()
			if err := os.WriteFile(filepath.Join(origin, "guard-fixture"), []byte("synthetic origin"), 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.v", "-test.run=^"+childTest+"$")
			for _, entry := range os.Environ() {
				key, _, _ := strings.Cut(entry, "=")
				switch key {
				case "HOME", "USERPROFILE", "XDG_CONFIG_HOME", "XDG_DATA_HOME", "CODEX_HOME", "CLAUDE_CONFIG_DIR", "PI_CODING_AGENT_DIR", controlMode, controlOrigin:
					continue
				}
				cmd.Env = append(cmd.Env, entry)
			}
			cmd.Env = append(cmd.Env, "HOME="+origin, "USERPROFILE="+origin, "XDG_CONFIG_HOME="+filepath.Join(origin, ".config"), controlMode+"="+mode, controlOrigin+"="+origin)
			out, err := cmd.CombinedOutput()
			if mode == "plant" {
				exit, ok := err.(*exec.ExitError)
				if !ok || exit.ExitCode() != 1 || !bytes.Contains(out, []byte(Changed+" ["+fingerprint(filepath.Join(origin, ".claude"))+"]")) || !bytes.Contains(out, []byte("synthetic unpinned HOME plant completed")) {
					t.Fatalf("guard did not reject its intended plant: %v\n%s", err, out)
				}
				t.Log("expected exit 1: actual package-exit guard detected synthetic unpinned HOME write")
			} else if err != nil || !bytes.Contains(out, []byte("protected native homes unchanged")) {
				t.Fatalf("isolated positive control failed: %v\n%s", err, out)
			}
		})
	}
}

// Child is called only by a named test child after TestMain isolation. The
// supplied resolver is the package's production native-home resolver.
func Child(t *testing.T, resolve func() (string, error)) {
	t.Helper()
	mode := os.Getenv(controlMode)
	if mode == "" {
		return
	}
	origin := os.Getenv(controlOrigin)
	marker, err := os.ReadFile(filepath.Join(origin, "guard-fixture"))
	if err != nil || string(marker) != "synthetic origin" {
		t.Fatal("missing synthetic guard origin")
	}
	if mode == "plant" {
		if err := os.Setenv("HOME", origin); err != nil {
			t.Fatal(err)
		}
		if err := os.Setenv("USERPROFILE", origin); err != nil {
			t.Fatal(err)
		}
	}
	path, err := resolve()
	if err != nil {
		t.Fatal(err)
	}
	inside, err := filepath.Rel(origin, path)
	if err != nil || (mode == "plant") != (inside != ".." && !strings.HasPrefix(inside, ".."+string(filepath.Separator))) {
		t.Fatal("plant resolved an unexpected native home")
	}
	if err := os.MkdirAll(path, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "CLAUDE.md"), []byte("synthetic native context witness"), 0600); err != nil {
		t.Fatal(err)
	}
	if mode == "plant" {
		t.Log("synthetic unpinned HOME plant completed")
	}
}
