package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// Templates contain only regular files/directories, never installed state,
// hardlinks, symlinks, toolchain probes, or absolute-path-bearing configuration.
// Each consumer receives new files, including its own Git index and refs.
type sourceFixtureEntry struct {
	path   string
	mode   fs.FileMode
	data   []byte
	digest [32]byte
}
type sourceFixtureTree []sourceFixtureEntry

var sourceFixtures = struct {
	sync.Mutex
	trees map[string]sourceFixtureTree
}{trees: make(map[string]sourceFixtureTree)}

// A source snapshot must be quiescent when capture starts. Git's automatic
// post-commit maintenance otherwise removes lock files while WalkDir reads them.
func configureSourceFixtureGit(t *testing.T, root string) {
	t.Helper()
	for _, setting := range [][2]string{{"maintenance.auto", "false"}, {"gc.auto", "0"}} {
		cmd := exec.Command("git", "-C", root, "config", "--local", setting[0], setting[1])
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("configure source fixture Git: %v: %s", err, output)
		}
	}
}

func captureSourceFixture(root string) (sourceFixtureTree, error) {
	var tree sourceFixtureTree
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		file := sourceFixtureEntry{path: rel, mode: info.Mode()}
		if info.Mode().IsRegular() {
			file.data, err = os.ReadFile(path)
			if err != nil {
				return err
			}
			file.digest = sha256.Sum256(file.data)
		} else if !info.IsDir() {
			return fmt.Errorf("unsupported source fixture entry: %s", rel)
		}
		tree = append(tree, file)
		return nil
	})
	return tree, err
}

func (tree sourceFixtureTree) copyTo(root string) error {
	// Validate all shared bytes before writing any consumer state.
	for _, file := range tree {
		if file.mode.IsRegular() && sha256.Sum256(file.data) != file.digest {
			return fmt.Errorf("shared source fixture changed: %s", file.path)
		}
	}
	for _, file := range tree {
		path := filepath.Join(root, file.path)
		if file.mode.IsDir() {
			if err := os.MkdirAll(path, file.mode.Perm()); err != nil {
				return err
			}
		} else {
			if err := os.WriteFile(path, file.data, file.mode.Perm()); err != nil {
				return err
			}
		}
	}
	return nil
}

func copySourceFixture(t *testing.T, key, target string, build func(string)) {
	t.Helper()
	// Only setup is serialized; tests operate on independent writable copies.
	func() {
		sourceFixtures.Lock()
		defer sourceFixtures.Unlock()
		if _, ok := sourceFixtures.trees[key]; !ok {
			root := t.TempDir()
			build(root)
			tree, err := captureSourceFixture(root)
			if err != nil {
				t.Fatal(err)
			}
			sourceFixtures.trees[key] = tree
		}
	}()
	sourceFixtures.Lock()
	tree := sourceFixtures.trees[key]
	sourceFixtures.Unlock()
	if err := tree.copyTo(target); err != nil {
		t.Fatal(err)
	}
}

// A writer's checkout cannot change a later consumer, even for Git metadata.
func TestSharedSourceFixtureCopiesAreIsolated(t *testing.T) {
	t.Parallel()
	build := func(root string) {
		if err := os.MkdirAll(filepath.Join(root, ".git"), 0755); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"context.md", ".git/HEAD"} {
			if err := os.WriteFile(filepath.Join(root, name), []byte("original"), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	first := t.TempDir()
	copySourceFixture(t, t.Name(), first, build)
	for _, name := range []string{"context.md", ".git/HEAD"} {
		if err := os.WriteFile(filepath.Join(first, name), []byte("writer"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Run("later-reader", func(t *testing.T) {
		second := t.TempDir()
		copySourceFixture(t, "TestSharedSourceFixtureCopiesAreIsolated", second, func(string) { t.Fatal("fixture rebuilt") })
		for _, name := range []string{"context.md", ".git/HEAD"} {
			got, err := os.ReadFile(filepath.Join(second, name))
			if err != nil || string(got) != "original" {
				t.Fatalf("copy %s = %q, %v", name, got, err)
			}
		}
	})
}

// A write to the shared template is detected by the next consumer before any
// partial output. This local corrupt template never enters the package cache.
func TestSharedSourceFixtureRejectsMutation(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "context.md"), []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	tree, err := captureSourceFixture(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := tree.copyTo(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	tree[0].data[0] = 'X'
	t.Run("later-reader", func(t *testing.T) {
		target := t.TempDir()
		err := tree.copyTo(target)
		if err == nil || err.Error() != "shared source fixture changed: context.md" {
			t.Fatalf("mutated template error = %v", err)
		}
		entries, err := os.ReadDir(target)
		if err != nil || len(entries) != 0 {
			t.Fatalf("partial copy = %v, %v", entries, err)
		}
	})
}

// Narrow only context.md's integrity check, retaining the check for every
// other file and its diagnostic token. The real copy consumer must fail.
func TestSharedSourceFixtureMutationControl(t *testing.T) {
	source, err := os.ReadFile("source_fixture_test.go")
	if err != nil {
		t.Fatal(err)
	}
	source = []byte(strings.ReplaceAll(string(source), "\r\n", "\n"))
	anchor := "\n\t\tif file.mode.IsRegular() && sha256.Sum256(file.data) != file.digest {"
	if strings.Count(string(source), anchor) != 1 {
		t.Fatal("mutation control not applied")
	}
	dir := t.TempDir()
	mutant := filepath.Join(dir, "source_fixture_test.go")
	changed := strings.Replace(string(source), anchor, strings.Replace(anchor, "file.mode.IsRegular() &&", `file.mode.IsRegular() && file.path != "context.md" &&`, 1), 1)
	if err := os.WriteFile(mutant, []byte(changed), 0600); err != nil {
		t.Fatal(err)
	}
	original, err := filepath.Abs("source_fixture_test.go")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{original: mutant}})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(overlay, data, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "test", "-count=1", "-overlay", overlay, "-run", "^TestSharedSourceFixture(CopiesAreIsolated|RejectsMutation)$", ".")
	// Reuse the established helper protocol: the parent holds the host GOROOT lock.
	cmd.Env = append(os.Environ(), cliHelperProcess+"=1")
	output, err := cmd.CombinedOutput()
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: TestSharedSourceFixtureRejectsMutation/later-reader") || !strings.Contains(string(output), "mutated template error = <nil>") {
		t.Fatalf("mutation survived or unrelated failure: %v\n%s", err, output)
	}
	t.Log("mutant=context-md-integrity-bypass child_exit=1 killer=TestSharedSourceFixtureRejectsMutation/later-reader")
}

// Real post-commit Git maintenance is disabled before snapshotting. The positive
// control forces it back on and must observe its child, so a silent trace cannot pass.
func TestSharedSourceFixtureGitMaintenanceStaysQuiescent(t *testing.T) {
	t.Parallel()
	for _, automatic := range []bool{false, true} {
		t.Run(fmt.Sprint(automatic), func(t *testing.T) {
			root := t.TempDir()
			cmd := exec.Command("git", "init", "-q", root)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("init: %v: %s", err, output)
			}
			configureSourceFixtureGit(t, root)
			trace := filepath.Join(t.TempDir(), "trace.json")
			args := []string{"-C", root, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "-c", "maintenance.autoDetach=false", "-c", "gc.autoDetach=false"}
			if automatic {
				args = append(args, "-c", "maintenance.auto=true")
			}
			args = append(args, "commit", "--allow-empty", "-qm", "fixture")
			cmd = exec.Command("git", args...)
			cmd.Env = append(os.Environ(), "GIT_TRACE2_EVENT="+filepath.ToSlash(trace))
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("commit: %v: %s", err, output)
			}
			data, err := os.ReadFile(trace)
			if err != nil {
				t.Fatal(err)
			}
			seen := false
			for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
				var e struct {
					Event string
					Argv  []string
				}
				if err := json.Unmarshal([]byte(line), &e); err != nil {
					t.Fatal(err)
				}
				if e.Event == "child_start" && strings.Contains(strings.Join(e.Argv, " "), "maintenance run --auto") {
					seen = true
				}
			}
			if seen != automatic {
				t.Fatalf("automatic maintenance child = %v, want %v", seen, automatic)
			}
			if !automatic {
				if _, err := captureSourceFixture(root); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
