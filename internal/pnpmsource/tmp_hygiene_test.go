package pnpmsource

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// The real conformance body and its capture-only early failure restore all modes before TempDir removal.
func TestPNPMTemporaryTreeRemovable(t *testing.T) {
	if testing.Short() {
		t.Skip("real pnpm cleanup integration")
	}
	newConcretePNPMRunner(t) // Require the same pinned runtime as the real conformance test.
	for _, mode := range []string{"pass", "failure", "capture-failure"} {
		t.Run(mode, func(t *testing.T) {
			report := filepath.Join(t.TempDir(), "restored")
			childTmp := filepath.Join(writableTempDir(t), "child-tmp")
			if err := os.Mkdir(childTmp, 0o700); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(os.Args[0], "-test.run=^TestPNPMTemporaryTreeChild$", "-test.v")
			cmd.Env = append(os.Environ(), "CURATOR_PNPM_CLEANUP_CHILD="+mode, "CURATOR_PNPM_CLEANUP_REPORT="+report, "TMPDIR="+childTmp, "TMP="+childTmp, "TEMP="+childTmp)
			out, err := cmd.CombinedOutput()
			want := 0
			if mode != "pass" {
				want = 1
			}
			code := 0
			if err != nil {
				exit, ok := err.(*exec.ExitError)
				if !ok {
					t.Fatal(err)
				}
				code = exit.ExitCode()
			}
			if code != want {
				t.Fatalf("child exit=%d want=%d: %s", code, want, out)
			}
			data, readErr := os.ReadFile(report)
			if readErr != nil || string(data) != "writable and removed\n" {
				t.Fatalf("cleanup witness missing: %v; child: %s", readErr, out)
			}
			if mode != "pass" && !strings.Contains(string(out), "forced cleanup failure") {
				t.Fatalf("child failed for an unrelated reason: %s", out)
			}
		})
	}
}

// Subprocess isolation allows an actual Fatal path without marking the parent test failed.
func TestPNPMTemporaryTreeChild(t *testing.T) {
	mode := os.Getenv("CURATOR_PNPM_CLEANUP_CHILD")
	if mode == "" {
		t.Skip("helper process")
	}
	base := filepath.Dir(t.TempDir())
	// Registered after TempDir removal and before the body: LIFO restores, audits, then removes.
	t.Cleanup(func() {
		err := filepath.WalkDir(base, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.Type()&os.ModeSymlink != 0 {
				return nil
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			if info.Mode().Perm()&0o200 == 0 {
				return fmt.Errorf("non-writable entry after body cleanup: %s", path)
			}
			return nil
		})
		if err != nil {
			t.Errorf("audit tree: %v", err)
			return
		}
		if t.Failed() && mode == "pass" {
			return
		}
		// Don't publish success merely because the child's forced failure was expected.
		if err = os.RemoveAll(base); err != nil {
			t.Errorf("remove tree: %v", err)
			return
		}
		if err = os.WriteFile(os.Getenv("CURATOR_PNPM_CLEANUP_REPORT"), []byte("writable and removed\n"), 0o600); err != nil {
			t.Error(err)
		}
	})
	if mode == "capture-failure" {
		captureFixture(t, newHostPNPMFixture(t))
	} else {
		TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization(t)
	}
	if mode != "pass" {
		t.Fatal("forced cleanup failure")
	}
}

// Nested sealed fixture trees are owner-writable before removal, preserving executability and external symlink targets.
func TestWritableTempDirRestoresNestedTrees(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission fixture is exercised on Linux and macOS")
	}
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("outside"), 0o400); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(outside, 0o600) })
	t.Run("sealed", func(t *testing.T) {
		_ = t.TempDir() // Register TempDir removal before the audit.
		var root string
		t.Cleanup(func() {
			for _, path := range []string{root, filepath.Join(root, "nested"), filepath.Join(root, "nested", "executable")} {
				info, err := os.Stat(path)
				if err != nil || info.Mode().Perm()&0o200 == 0 {
					t.Errorf("tree not restored: %s: %v", path, err)
				}
			}
			info, err := os.Stat(filepath.Join(root, "nested", "executable"))
			if err != nil || info.Mode().Perm()&0o100 == 0 {
				t.Error("executable bit lost")
			}
		})
		root = writableTempDir(t)
		dir := filepath.Join(root, "nested")
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "executable"), []byte("x"), 0o500); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(dir, "link")); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(dir, 0o500); err != nil {
			t.Fatal(err)
		}
	})
	if info, err := os.Stat(outside); err != nil || info.Mode().Perm()&0o200 != 0 {
		t.Errorf("outside target changed: %v", err)
	}
}
