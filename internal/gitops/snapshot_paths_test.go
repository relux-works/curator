package gitops

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// The selected roster and existing projected pin agree with a real extraction
// in both object formats. A forbidden sibling is outside both selections.
func TestPinnedSnapshotPathsProjection(t *testing.T) {
	for _, format := range []string{"sha1", "sha256"} {
		t.Run(format, func(t *testing.T) {
			repo := t.TempDir()
			gitRun(t, repo, "init", "-q", "-b", "main", "--object-format="+format)
			for name, text := range map[string]string{"selected/run": "native capability, no SKILL.md\n", "outside": "sibling\n"} {
				path := filepath.Join(repo, name)
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(text), 0644); err != nil {
					t.Fatal(err)
				}
			}
			gitRun(t, repo, "add", ".")
			gitRun(t, repo, "commit", "-qm", "roster fixture")
			commit := gitRun(t, repo, "rev-parse", "HEAD")
			paths, err := PinnedDirectorySnapshotPaths(repo, commit, "selected")
			if err != nil || !reflect.DeepEqual(paths, []string{"selected/run"}) {
				t.Fatalf("selected roster: %v %v", paths, err)
			}
			pin, _, err := PinnedDirectorySnapshotTree(repo, commit, "selected")
			if err != nil {
				t.Fatal(err)
			}
			dest := filepath.Join(t.TempDir(), "snapshot")
			if err := ExtractDirectory(repo, commit, "selected", dest); err != nil {
				t.Fatal(err)
			}
			actual, err := TreeObjectIDFromDir(dest, format)
			if err != nil || actual != pin {
				t.Fatalf("roster/hash projection differs: %s %s %v", actual, pin, err)
			}
			if _, err := os.Stat(filepath.Join(dest, "outside")); !os.IsNotExist(err) {
				t.Fatalf("sibling entered projection: %v", err)
			}
			if paths, err := PinnedDirectorySnapshotPaths(repo, "HEAD", "selected"); err == nil || paths != nil {
				t.Fatalf("ref admitted: %v %v", paths, err)
			}
		})
	}
	repo, commit := directoryRepo(t, map[string]string{"unselected/link": "../sibling"})
	if paths, err := PinnedDirectorySnapshotPaths(repo, commit, "agents/skills/demo"); err != nil || len(paths) != 2 {
		t.Fatalf("outside link affected selection: %v %v", paths, err)
	}
	if paths, err := PinnedDirectorySnapshotPaths(repo, commit, ""); err == nil || paths != nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("whole-tree link admitted: %v %v", paths, err)
	}
}

// Hostile ambient Git selectors cannot supply a different repository, object
// database, config or replacement ref. A missing local object stays missing.
func TestPinnedSnapshotPathsIsolation(t *testing.T) {
	repo, commit := directoryRepo(t, nil)
	other := makeRepo(t)
	if err := os.WriteFile(filepath.Join(other, "foreign-only"), []byte("not in the selected repository\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, other, "add", "foreign-only")
	gitRun(t, other, "commit", "-qm", "distinct foreign object")
	otherCommit := gitRun(t, other, "rev-parse", "HEAD")
	probe := gitCommand("-C", repo, "cat-file", "-e", otherCommit+"^{commit}")
	probe.Env = gitEnv()
	if err := probe.Run(); err == nil {
		t.Fatal("isolation control is invalid: foreign object already exists locally")
	}
	want, err := PinnedDirectorySnapshotPaths(repo, commit, "agents/skills/demo")
	if err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(t.TempDir(), "hostile.gitconfig")
	if err := os.WriteFile(config, []byte("[core]\n repositoryformatversion=999\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for k, v := range map[string]string{"GIT_DIR": filepath.Join(other, ".git"), "GIT_WORK_TREE": other, "GIT_OBJECT_DIRECTORY": filepath.Join(other, ".git/objects"), "GIT_ALTERNATE_OBJECT_DIRECTORIES": filepath.Join(other, ".git/objects"), "GIT_CONFIG_GLOBAL": config, "GIT_CONFIG_SYSTEM": config, "GIT_CONFIG_NOSYSTEM": "0", "GIT_CONFIG_COUNT": "1", "GIT_CONFIG_KEY_0": "core.repositoryformatversion", "GIT_CONFIG_VALUE_0": "999", "GIT_CONFIG_PARAMETERS": "'core.repositoryformatversion=999'", "GIT_REPLACE_REF_BASE": "refs/hostile/", "GIT_GRAFT_FILE": config} {
		t.Setenv(k, v)
	}
	t.Setenv("GIT_NO_REPLACE_OBJECTS", "0")
	if paths, err := PinnedDirectorySnapshotPaths(repo, commit, "agents/skills/demo"); err != nil || !reflect.DeepEqual(paths, want) {
		t.Fatalf("ambient selectors changed roster: %v %v", paths, err)
	}
	if paths, err := PinnedDirectorySnapshotPaths(repo, otherCommit, ""); err == nil || paths != nil || !strings.Contains(err.Error(), "pinned commit object unavailable") {
		t.Fatalf("foreign object admitted: %v %v", paths, err)
	}
}

// Repository-local alternate/graft metadata is refused before resolving the
// snapshot. The same unmodified local repository is a positive control.
func TestPinnedSnapshotPathsRejectsAlternateAndGraft(t *testing.T) {
	for _, metadata := range []string{"objects/info/alternates", "info/grafts"} {
		t.Run(metadata, func(t *testing.T) {
			repo, commit := directoryRepo(t, nil)
			if _, err := PinnedDirectorySnapshotPaths(repo, commit, ""); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(repo, ".git", filepath.FromSlash(metadata))
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte{}, 0600); err != nil {
				t.Fatal(err)
			}
			if paths, err := PinnedDirectorySnapshotPaths(repo, commit, ""); err == nil || paths != nil || !strings.Contains(err.Error(), metadata) {
				t.Fatalf("metadata refusal: %v %v", paths, err)
			}
		})
	}
}

// Metadata/path limits fail without a truncated roster. Nearby valid paths
// include non-ASCII Unicode; collisions include folded directory prefixes.
func TestPinnedSnapshotPathMetadataControls(t *testing.T) {
	// Exercise the public reader with real Git output above the byte bound,
	// without creating thousands of filesystem payloads. A nearby small tree
	// proves repository setup and the same read path work before the plant.
	repo := makeRepo(t)
	commit := gitRun(t, repo, "rev-parse", "HEAD")
	if paths, err := PinnedDirectorySnapshotPaths(repo, commit, ""); err != nil || len(paths) == 0 {
		t.Fatalf("metadata control baseline: %v %v", paths, err)
	}
	blob := gitRun(t, repo, "hash-object", "-w", "--stdin")
	var treeInput strings.Builder
	for i := 0; i < maxPinnedPaths; i++ {
		fmt.Fprintf(&treeInput, "100644 blob %s\t%05d-%s\x00", blob, i, strings.Repeat("x", 1024))
	}
	cmd := gitCommand("-C", repo, "mktree", "-z")
	cmd.Env = gitEnv()
	cmd.Stdin = strings.NewReader(treeInput.String())
	tree, err := cmd.Output()
	if err != nil {
		t.Fatalf("metadata control tree setup: %v", err)
	}
	commit = gitRun(t, repo, "commit-tree", strings.TrimSpace(string(tree)), "-m", "metadata bound control")
	if paths, err := PinnedDirectorySnapshotPaths(repo, commit, ""); err == nil || paths != nil || !strings.Contains(err.Error(), "path metadata exceeds 16 MiB") {
		t.Fatalf("actual Git metadata bound refusal returned %d paths: %v", len(paths), err)
	}
	for name, paths := range map[string][]string{
		"positive": {"run", "nested/é.txt"}, "dotdot": {"../escape"}, "absolute": {"/escape"}, "nul": {"bad\x00path"},
		"git": {"nested/.Git/config"}, "case": {"A/x", "a/y"}, "unicode": {"é.txt", "e\u0301.txt"}, "duplicate": {"same", "same"}, "file-directory": {"folder/x", "folder"},
	} {
		t.Run(name, func(t *testing.T) {
			entries := make([]treeEntry, len(paths))
			for i, path := range paths {
				entries[i].path = path
			}
			got, err := pinnedPaths(entries, "")
			if name == "positive" {
				if err != nil || !reflect.DeepEqual(got, paths) {
					t.Fatalf("valid paths: %v %v", got, err)
				}
				return
			}
			if err == nil || got != nil {
				t.Fatalf("unsafe paths admitted: %v %v", got, err)
			}
		})
	}
	for _, mode := range []string{"120000 blob", "160000 commit", "040000 tree"} {
		if got, err := parseTreeListing("fixture", []byte(mode+" abc 0\tbad\x00")); err == nil || got != nil {
			t.Fatalf("unsafe type %s: %v %v", mode, got, err)
		}
	}
	entries := make([]treeEntry, maxPinnedPaths+1)
	for i := range entries {
		entries[i].path = fmt.Sprintf("file-%d", i)
	}
	if got, err := pinnedPaths(entries[:maxPinnedPaths], ""); err != nil || len(got) != maxPinnedPaths {
		t.Fatalf("count boundary valid: %d %v", len(got), err)
	}
	if got, err := pinnedPaths(entries, ""); err == nil || got != nil || !strings.Contains(err.Error(), "16384") {
		t.Fatalf("count refusal: %v", err)
	}
	var b pinnedListingBuffer
	if n, err := b.Write(bytes.Repeat([]byte{'x'}, maxPinnedPathBytes)); err != nil || n != maxPinnedPathBytes {
		t.Fatalf("byte boundary valid: %d %v", n, err)
	}
	if n, err := b.Write([]byte{'x'}); err == nil || n != 0 || b.Len() != maxPinnedPathBytes || !strings.Contains(err.Error(), "16 MiB") {
		t.Fatalf("metadata limit truncated: %d %v", n, err)
	}
}
