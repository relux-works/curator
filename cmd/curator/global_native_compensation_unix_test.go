//go:build unix

package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/envregistry"
)

// These production-CLI witnesses pause real Git after profile publication.
// A clean manifest restart rolls back and converges; failed compensation must
// stop the restart, retain the pending profile lock, and preserve unknown state.
func TestGlobalNativeOwnerCompensationRestart(t *testing.T) {
	for _, fault := range []string{"none", "unwritable-parent", "replacement"} {
		t.Run(fault, func(t *testing.T) {
			if fault == "unwritable-parent" && os.Geteuid() == 0 {
				t.Skip("requires unprivileged directory permissions")
			}
			fixture := setupNativeConcurrent(t, false)
			child, release := pauseNativeProfilePublication(t, fixture)
			fixture.requireProfileOwner(t)
			path := filepath.Join(fixture.home, ".codex", "skills", "shared-skill")
			original, err := os.Readlink(path)
			if err != nil {
				t.Fatal(err)
			}
			if fault == "unwritable-parent" {
				if err := os.Chmod(filepath.Dir(path), 0500); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(filepath.Dir(path), 0755) })
			}
			if fault == "replacement" {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				writeFile(t, path, "operator replacement\n")
			}
			writeFile(t, filepath.Join(fixture.home, ".curator", "global", "Skillfile.json"), `{"schema_version":1,"agents":["codex_cli","claude_code"],"skills":[]}`)
			release()
			code, output := nativeCommandOutcome(t, child)
			if fault == "none" {
				if code != exitOK {
					t.Fatalf("clean restart failed: %d %s", code, output)
				}
				requireNativeClosure(t, fixture, false)
				requireNativeCanonical(t, fixture, false)
				return
			}
			if code != exitFail || !strings.Contains(output, "global profile compensation failed") {
				t.Fatalf("compensation failure was swallowed by restart: %d %s", code, output)
			}
			if fault == "replacement" {
				if data, err := os.ReadFile(path); err != nil || string(data) != "operator replacement\n" {
					t.Fatalf("compensation removed operator replacement: %q %v", data, err)
				}
			} else {
				if target, err := os.Readlink(path); err != nil || target != original {
					t.Fatalf("failed compensation lost active entry: %q %v", target, err)
				}
				if err := os.Chmod(filepath.Dir(path), 0755); err != nil {
					t.Fatal(err)
				}
			}
			// The durable extended lock and install record must still authorize
			// explicit actualization; no empty lock may conceal the active skill.
			requirePendingNativeSkill(t, fixture)
			fixture.command(t, exitOK, "profile", "sync", "--takeover")
			fixture.requireProfileOwner(t)
			fixture.command(t, exitOK, "global", "install")
			requireNativeClosure(t, fixture, false)
			requireNativeCanonical(t, fixture, false)
		})
	}
}

// An ordinary canonical refusal also preserves an operator replacement. The
// adjacent canonical-refusal test covers a successful exact-preimage rollback.
func TestGlobalNativeOwnerCompensationReplacement(t *testing.T) {
	fixture := setupNativeConcurrent(t, false)
	foreign := filepath.Join(fixture.home, ".curator", "global", "skills", "shared-skill", "SKILL.md")
	writeFile(t, foreign, "foreign canonical\n")
	child, release := pauseNativeProfilePublication(t, fixture)
	path := filepath.Join(fixture.home, ".codex", "skills", "shared-skill")
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	writeFile(t, path, "operator replacement\n")
	release()
	code, output := nativeCommandOutcome(t, child)
	if code != exitFail || !strings.Contains(output, "source_ownership_unproven") {
		t.Fatalf("canonical refusal missing: %d %s", code, output)
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != "operator replacement\n" {
		t.Fatalf("compensation removed operator replacement: %q %v", data, err)
	}
	if !strings.Contains(output, "global profile compensation failed") {
		t.Fatalf("replacement conflict not reported: %s", output)
	}
	requirePendingNativeSkill(t, fixture)
	if data, err := os.ReadFile(foreign); err != nil || string(data) != "foreign canonical\n" {
		t.Fatalf("canonical foreign entry changed: %q %v", data, err)
	}
}

func requirePendingNativeSkill(t *testing.T, fixture nativeGlobalFixture) {
	t.Helper()
	lock, err := os.ReadFile(filepath.Join(fixture.home, ".curator", "profiles", "selected-profile", "lock.json"))
	if err != nil || !strings.Contains(string(lock), `"name":"shared-skill"`) {
		t.Fatalf("compensation failure lost pending lock: %s %v", lock, err)
	}
	if _, err := os.Stat(filepath.Join(fixture.home, ".curator", "profiles", "selected-profile", "source.json")); err != nil {
		t.Fatalf("compensation failure lost install record: %v", err)
	}
}

func pauseNativeProfilePublication(t *testing.T, fixture nativeGlobalFixture) (*nativeCommand, func()) {
	t.Helper()
	published := filepath.Join(envprofile.ManagedHomeDir(filepath.Join(fixture.home, ".curator"), "selected-profile", envregistry.CodexCLI), envmarker.Name)
	dir := t.TempDir()
	ready, release := filepath.Join(dir, "ready"), filepath.Join(dir, "release")
	git, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'" }
	script := "#!/bin/sh\n"
	for key, value := range map[string]string{cliHelperProcess: "1", "CURATOR_NATIVE_GIT_HELPER": "1", "CURATOR_NATIVE_REAL_GIT": git, "CURATOR_NATIVE_GIT_CONFIG": os.Getenv("GIT_CONFIG_GLOBAL"), "CURATOR_NATIVE_PUBLISHED": published, "CURATOR_NATIVE_READY": ready, "CURATOR_NATIVE_RELEASE": release} {
		script += key + "=" + quote(value) + "\nexport " + key + "\n"
	}
	script += "exec " + quote(os.Args[0]) + " -test.run=^TestNativeGitBarrierProcess$ -- \"$@\"\n"
	writeFile(t, filepath.Join(dir, "git"), script)
	if err := os.Chmod(filepath.Join(dir, "git"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	child := startNativeCommand(t, fixture, "global", "install")
	unblock := func() {
		if err := os.WriteFile(release, []byte("release"), 0600); err != nil {
			t.Error(err)
		}
	}
	// This cleanup is registered after the child cleanup and therefore runs first.
	t.Cleanup(unblock)
	deadline := time.Now().Add(60 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		select {
		case err := <-child.done:
			child.finished = true
			t.Fatalf("production CLI exited before publication barrier: %v %s", err, child.output.String())
		default:
		}
		if time.Now().After(deadline) {
			t.Fatal("production Git pause did not reach native publication")
		}
		time.Sleep(10 * time.Millisecond)
	}
	return child, unblock
}

func nativeCommandOutcome(t *testing.T, child *nativeCommand) (int, string) {
	t.Helper()
	select {
	case err := <-child.done:
		child.finished = true
		if err == nil {
			return 0, child.output.String()
		}
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatal(err)
		}
		return exit.ExitCode(), child.output.String()
	case <-time.After(40 * time.Second):
		t.Fatal("production CLI did not finish after release")
		return -1, ""
	}
}

// A real post-publication Git pause leaves room for an operator chmod. An
// unchanged parent permits rollback and restart; a narrower operator mode must
// survive, report a compensation conflict and retain the pending profile lock.
// The unrelated file remains byte-exact in both cases.
func TestGlobalNativeOwnerCompensationParentMode(t *testing.T) {
	for _, changed := range []bool{false, true} {
		t.Run(map[bool]string{false: "unchanged", true: "operator-chmod"}[changed], func(t *testing.T) {
			fixture := setupNativeConcurrent(t, false)
			parent := filepath.Join(fixture.home, ".codex", "skills")
			if err := os.MkdirAll(parent, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(parent, 0755); err != nil {
				t.Fatal(err)
			}
			unrelated := filepath.Join(parent, "operator-note.txt")
			writeFile(t, unrelated, "operator bytes\n")
			child, release := pauseNativeProfilePublication(t, fixture)
			info, err := os.Lstat(parent)
			if err != nil || info.Mode().Perm() != 0755 {
				t.Fatalf("parent publication did not retain the fixture mode: %v %v", info, err)
			}
			want := os.FileMode(0755)
			if changed {
				want = 0700
				if err := os.Chmod(parent, want); err != nil {
					t.Fatal(err)
				}
			}
			writeFile(t, filepath.Join(fixture.home, ".curator", "global", "Skillfile.json"), `{"schema_version":1,"agents":["codex_cli","claude_code"],"skills":[]}`)
			release()
			code, output := nativeCommandOutcome(t, child)
			info, err = os.Lstat(parent)
			if err != nil || info.Mode().Perm() != want {
				t.Fatalf("operator parent mode overwritten: want %o, got %v %v", want, info, err)
			}
			if data, err := os.ReadFile(unrelated); err != nil || string(data) != "operator bytes\n" {
				t.Fatalf("unrelated operator file changed: %q %v", data, err)
			}
			if !changed {
				if code != exitOK {
					t.Fatalf("unchanged parent blocked compensation: %d %s", code, output)
				}
				requireNativeClosure(t, fixture, false)
				requireNativeCanonical(t, fixture, false)
				return
			}
			if code != exitFail || !strings.Contains(output, "global profile compensation failed") || !strings.Contains(output, "compensation parent postimage changed") {
				t.Fatalf("operator parent change was not a reported conflict: %d %s", code, output)
			}
			requirePendingNativeSkill(t, fixture)
			fixture.requireProfileOwner(t)
			requireNativeCanonical(t, fixture, false)
		})
	}
}
