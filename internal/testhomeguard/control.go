package testhomeguard

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

const controlMode = "CURATOR_TEST_NATIVE_GUARD_CONTROL"
const controlOrigin = "CURATOR_TEST_NATIVE_GUARD_ORIGIN"

// Control proves the actual package-exit guard detects a bounded write through
// an unpinned HOME. Both child original homes are temporary fixture directories;
// the plant never selects the OS account home. The paired isolated write passes.
func Control(t *testing.T, childTest string, childEnv ...string) {
	t.Helper()
	for _, mode := range []string{"isolated", "plant", "settings", "skill-copy", "synced-skill", "env-state", "panic", "exit", "panic-clean", "exit-clean", "forged", "forged-overflow", "overflow"} {
		t.Run(mode, func(t *testing.T) {
			origin := t.TempDir()
			if err := os.WriteFile(filepath.Join(origin, "guard-fixture"), []byte("synthetic origin"), 0600); err != nil {
				t.Fatal(err)
			}
			// Existing decoys make settings/copied-skill controls discriminate
			// the old shallow snapshot: no parent or skill-entry identity changes.
			for _, name := range []string{"CLAUDE.md", "settings.json", ".agent-environment.json", "skills/copied/nested/SKILL.md", "skills/synced/fixture/SKILL.md"} {
				path := filepath.Join(origin, ".claude", filepath.FromSlash(name))
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(strings.Repeat("b", len("synthetic native context witness"))), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "overflow" || mode == "forged-overflow" {
				f, err := os.OpenFile(filepath.Join(origin, ".claude", "CLAUDE.md"), os.O_WRONLY, 0600)
				if err != nil {
					t.Fatal(err)
				}
				err = f.Truncate(maxFileBytes + 1)
				_ = f.Close()
				if err != nil {
					t.Fatal(err)
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.v", "-test.run=^"+childTest+"$")
			for _, entry := range os.Environ() {
				key, _, _ := strings.Cut(entry, "=")
				switch key {
				case "HOME", "USERPROFILE", "XDG_CONFIG_HOME", "XDG_DATA_HOME", "CODEX_HOME", "CLAUDE_CONFIG_DIR", "PI_CODING_AGENT_DIR", controlMode, controlOrigin, supervisorPID, observerAddress, observerKey:
					continue
				}
				cmd.Env = append(cmd.Env, entry)
			}
			cmd.Env = append(cmd.Env, "HOME="+origin, "USERPROFILE="+origin, "XDG_CONFIG_HOME="+filepath.Join(origin, ".config"), controlMode+"="+mode, controlOrigin+"="+origin)
			cmd.Env = append(cmd.Env, childEnv...)
			if strings.HasPrefix(mode, "forged") {
				cmd.Env = append(cmd.Env, supervisorPID+"="+strconv.Itoa(os.Getpid()))
			}
			out, err := cmd.CombinedOutput()
			if strings.HasPrefix(mode, "forged") || mode == "overflow" {
				want := "live observer handoff required"
				if mode == "overflow" {
					want = "cannot establish protected preimage"
				}
				exit, ok := err.(*exec.ExitError)
				if !ok || exit.ExitCode() != 2 || !bytes.Contains(out, []byte(want)) || bytes.Contains(out, []byte("synthetic child reached")) {
					t.Fatalf("pre-test refusal missing: %v\n%s", err, out)
				}
				return
			}
			if mode != "isolated" && mode != "panic-clean" && mode != "exit-clean" {
				exit, ok := err.(*exec.ExitError)
				if !ok || exit.ExitCode() != 1 || !bytes.Contains(out, []byte(Changed+" ["+fingerprint(filepath.Join(origin, ".claude"))+"]")) || !bytes.Contains(out, []byte("synthetic "+mode+" plant completed")) {
					t.Fatalf("guard did not reject its intended plant: %v\n%s", err, out)
				}
				t.Log("expected exit 1: actual package-exit guard detected synthetic unpinned HOME write")
			} else {
				want := 0
				if mode == "panic-clean" {
					want = 2
				}
				if mode == "exit-clean" {
					want = 23
				}
				got := 0
				if err != nil {
					if exit, ok := err.(*exec.ExitError); ok {
						got = exit.ExitCode()
					} else {
						t.Fatalf("child setup failed: %v", err)
					}
				}
				if got != want || !bytes.Contains(out, []byte("protected native homes unchanged")) {
					t.Fatalf("positive exit control: got %d want %d: %s", got, want, out)
				}
				if mode != "isolated" && !bytes.Contains(out, []byte("synthetic "+mode+" exit witness")) {
					t.Fatalf("intended abnormal exit was not exercised: %s", out)
				}
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
	t.Log("synthetic child reached")
	origin := os.Getenv(controlOrigin)
	marker, err := os.ReadFile(filepath.Join(origin, "guard-fixture"))
	if err != nil || string(marker) != "synthetic origin" {
		t.Fatal("missing synthetic guard origin")
	}
	if mode != "isolated" && mode != "panic-clean" && mode != "exit-clean" {
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
	allowedRoot, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	if mode == "isolated" && allowedRoot == origin {
		t.Fatal("neutral helper did not isolate ambient HOME")
	}
	if mode != "isolated" && mode != "panic-clean" && mode != "exit-clean" {
		allowedRoot = origin
	}
	if !filepath.IsAbs(allowedRoot) {
		t.Fatal("missing synthetic write boundary")
	}
	inside, err := filepath.Rel(allowedRoot, path)
	if err != nil || inside == ".." || strings.HasPrefix(inside, ".."+string(filepath.Separator)) {
		t.Fatal("control resolved outside its synthetic write boundary")
	}
	if mode != "panic-clean" && mode != "exit-clean" {
		name := "CLAUDE.md"
		switch mode {
		case "settings":
			name = "settings.json"
		case "skill-copy":
			name = "skills/copied/nested/SKILL.md"
		case "synced-skill":
			name = "skills/synced/fixture/SKILL.md"
		case "env-state":
			name = ".agent-environment.json"
		}
		target := filepath.Join(path, filepath.FromSlash(name))
		previous, statErr := os.Stat(target)
		if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte("synthetic native context witness"), 0600); err != nil {
			t.Fatal(err)
		}
		if mode == "settings" || mode == "skill-copy" {
			if statErr != nil {
				t.Fatal(statErr)
			}
			if err := os.Chtimes(target, previous.ModTime(), previous.ModTime()); err != nil {
				t.Fatal(err)
			}
		}
		if mode != "isolated" {
			t.Log("synthetic " + mode + " plant completed")
		}
	}
	switch mode {
	case "panic", "panic-clean":
		t.Log("synthetic " + mode + " exit witness")
		panic("synthetic guarded panic")
	case "exit", "exit-clean":
		t.Log("synthetic " + mode + " exit witness")
		os.Exit(23)
	}
}
