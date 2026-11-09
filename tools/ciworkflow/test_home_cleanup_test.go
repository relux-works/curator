package ciworkflow

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Exercise the real wrapper with newly allocated fixture roots only. A marker
// outside the child home proves both dispatch and removal of the exact home.
func TestUnixTestHomeCleanup(t *testing.T) {
	wrapper, err := filepath.Abs("../../.github/ci/isolated-test.cjs")
	if err != nil {
		t.Fatal(err)
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, program string
		status        int
	}{
		{"success", "", 0},
		{"child-failure", "process.exit(7);", 7},
		{"child-signal", "process.kill(process.pid,'SIGTERM');", -1},
		{"readonly-cache", `const cache=p.join(process.env.HOME,'cache'); fs.mkdirSync(cache); fs.writeFileSync(p.join(cache,'payload'),'cache'); fs.chmodSync(cache,0o500);`, 0},
		{"removed-home", `fs.rmSync(process.env.HOME,{recursive:true});`, 0},
		{"home-symlink", `fs.rmSync(process.env.HOME,{recursive:true}); fs.symlinkSync(process.env.OUTSIDE,process.env.HOME,'dir');`, 0},
		{"nested-symlink", `fs.symlinkSync(process.env.OUTSIDE,p.join(process.env.HOME,'link'),'dir');`, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if runtime.GOOS == "windows" && (strings.Contains(tc.name, "symlink") || tc.name == "child-signal") {
				t.Skip("Unix wrapper control requires Unix symlink or signal semantics")
			}
			root, marker, outside := t.TempDir(), filepath.Join(t.TempDir(), "home"), t.TempDir()
			payload := filepath.Join(outside, "payload")
			if err := os.WriteFile(payload, []byte("unchanged"), 0o400); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(outside, 0o500); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(outside, 0o700) })
			before, err := os.Stat(outside)
			if err != nil {
				t.Fatal(err)
			}
			program := `const fs=require('node:fs'),p=require('node:path'); fs.writeFileSync(process.env.MARKER,process.env.HOME); ` + tc.program
			cmd := exec.Command(node, wrapper, node, "-e", program)
			cmd.Env = wrapperFixtureEnv(root, marker, outside)
			out, runErr := cmd.CombinedOutput()
			status := 0
			if runErr != nil {
				exit, ok := runErr.(*exec.ExitError)
				if !ok {
					t.Fatalf("wrapper launch: %v: %s", runErr, out)
				}
				status = exit.ExitCode()
			}
			if status != tc.status {
				t.Fatalf("wrapper exit %d, want %d: %s", status, tc.status, out)
			}
			home, err := os.ReadFile(marker)
			if err != nil {
				t.Fatal(err)
			}
			if !filepath.IsAbs(string(home)) || !strings.HasPrefix(filepath.Base(string(home)), "curator-test-home-") {
				t.Fatalf("unexpected fixture home: %q", home)
			}
			if _, err := os.Lstat(string(home)); !os.IsNotExist(err) {
				t.Fatalf("child home retained: %v", err)
			}
			assertNoWrapperHomes(t, root)
			after, err := os.Stat(outside)
			if err != nil || after.Mode() != before.Mode() {
				t.Fatalf("outside directory mode changed: %v", err)
			}
			data, err := os.ReadFile(payload)
			if err != nil || string(data) != "unchanged" {
				t.Fatalf("outside payload changed: %v", err)
			}
		})
	}
	for _, name := range []string{"missing-command", "launch-failure"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			args := []string{wrapper}
			if name == "launch-failure" {
				args = append(args, filepath.Join(root, "absent-command"))
			}
			cmd := exec.Command(node, args...)
			cmd.Env = wrapperFixtureEnv(root, "", "")
			out, err := cmd.CombinedOutput()
			if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 {
				t.Fatalf("expected exit 1: %v: %s", err, out)
			}
			assertNoWrapperHomes(t, root)
		})
	}
}

// An unremovable home turns child success into failure without hiding a child
// failure. Writable-parent neighbours prove that ordinary cleanup still succeeds.
func TestUnixTestHomeCleanupFailure(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("requires unprivileged Unix directory permissions")
	}
	wrapper, err := filepath.Abs("../../.github/ci/isolated-test.cjs")
	if err != nil {
		t.Fatal(err)
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name                    string
		blocked                 bool
		childStatus, wantStatus int
	}{
		{"success", false, 0, 0},
		{"cleanup-failure", true, 0, 1},
		{"child-failure", false, 7, 7},
		{"cleanup-and-child-failure", true, 7, 7},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, marker := t.TempDir(), filepath.Join(t.TempDir(), "home")
			t.Cleanup(func() {
				if err := os.Chmod(root, 0o700); err != nil {
					t.Error(err)
				}
			})
			program := `const fs=require('node:fs'); fs.writeFileSync(process.env.MARKER,process.env.HOME);`
			if tc.blocked {
				program += `fs.chmodSync(process.env.TMPDIR,0o500);`
			}
			if tc.childStatus != 0 {
				program += `process.exit(7);`
			}
			cmd := exec.Command(node, wrapper, node, "-e", program)
			cmd.Env = wrapperFixtureEnv(root, marker, "")
			out, runErr := cmd.CombinedOutput()
			status := 0
			if runErr != nil {
				exit, ok := runErr.(*exec.ExitError)
				if !ok {
					t.Fatalf("wrapper launch: %v: %s", runErr, out)
				}
				status = exit.ExitCode()
			}
			if status != tc.wantStatus {
				t.Errorf("wrapper exit %d, want %d: %s", status, tc.wantStatus, out)
			}
			const diagnostic = "isolated-test: temporary home cleanup failed"
			if strings.Contains(string(out), diagnostic) != tc.blocked {
				t.Errorf("cleanup failure diagnostic: got %q, blocked=%v", out, tc.blocked)
			}
			home, err := os.ReadFile(marker)
			if err != nil {
				t.Fatal(err)
			}
			physicalRoot, err := filepath.EvalSymlinks(root)
			if err != nil {
				t.Fatal(err)
			}
			if filepath.Dir(string(home)) != physicalRoot || !strings.HasPrefix(filepath.Base(string(home)), "curator-test-home-") {
				t.Fatalf("unexpected fixture home: %q", home)
			}
			info, err := os.Lstat(string(home))
			if tc.blocked {
				if err != nil || !info.IsDir() {
					t.Fatalf("unremovable home not retained for recovery: %v", err)
				}
				parent, err := os.Stat(root)
				if err != nil || parent.Mode().Perm() != 0o500 {
					t.Fatalf("cleanup changed denied parent permissions: %v", err)
				}
				// Only the fixture owner restores access after observing the refusal.
				if err := os.Chmod(root, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.RemoveAll(string(home)); err != nil {
					t.Fatal(err)
				}
			} else if !os.IsNotExist(err) {
				t.Fatalf("child home retained after successful cleanup: %v", err)
			}
			assertNoWrapperHomes(t, root)
		})
	}
}

func wrapperFixtureEnv(root, marker, outside string) []string {
	var env []string
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if key != "TMPDIR" && key != "TMP" && key != "TEMP" && key != "MARKER" && key != "OUTSIDE" {
			env = append(env, entry)
		}
	}
	return append(env, "TMPDIR="+root, "TMP="+root, "TEMP="+root, "MARKER="+marker, "OUTSIDE="+outside)
}

func assertNoWrapperHomes(t *testing.T, root string) {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatalf("wrapper retained allocations in its private fixture root: count=%d, err=%v", len(entries), err)
	}
}
