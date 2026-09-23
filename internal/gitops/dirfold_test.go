package gitops

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// commitCaseTree commits path->content pairs through update-index --cacheinfo
// plus write-tree, so trees whose names fold on disk (Dir/x.txt + dir/y.txt)
// still commit: only the index and the object database are touched, never the
// working tree.
func commitCaseTree(t *testing.T, files map[string]string) (string, string) {
	t.Helper()
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		gitRun(t, repo, "update-index", "--add", "--cacheinfo", "100644,"+blobOID(t, repo, files[name])+","+name)
	}
	return repo, gitRun(t, repo, "commit-tree", gitRun(t, repo, "write-tree"), "-m", "case")
}

// TestExtractRefusesDirectoryComponentFold proves the platform-path gate folds
// per path component: Dir/x.txt + dir/y.txt share no folded FULL path, but on
// a case-folding destination they land in ONE physical directory, so the pair
// is refused with the duplicate-platform-path diagnostic before cat-file
// starts and nothing is written.
func TestExtractRefusesDirectoryComponentFold(t *testing.T) {
	files := map[string]string{"Dir/x.txt": "x-bytes", "dir/y.txt": "y-bytes"}
	repo, commit := commitCaseTree(t, files)
	parent := t.TempDir()
	folds, err := destinationFoldsCase(parent)
	if err != nil {
		t.Fatal(err)
	}
	if !folds {
		t.Skip("test filesystem is case-sensitive: Dir and dir coexist, no directory-component collision to refuse")
	}
	dest := filepath.Join(parent, "snap")
	err = Extract(repo, commit, dest)
	if err == nil || !strings.Contains(err.Error(), "duplicate platform path") {
		t.Fatalf("err = %v, want directory-component fold refusal", err)
	}
	assertNothingWritten(t, parent, dest)
}

// TestExtractAdmitsDirectoryComponentFoldWhenCaseSensitive is the coexistence
// half of the gate: the same Dir/x.txt + dir/y.txt tree extracts with exact
// bytes on a case-sensitive destination.
func TestExtractAdmitsDirectoryComponentFoldWhenCaseSensitive(t *testing.T) {
	files := map[string]string{"Dir/x.txt": "x-bytes", "dir/y.txt": "y-bytes"}
	repo, commit := commitCaseTree(t, files)
	parent := t.TempDir()
	folds, err := destinationFoldsCase(parent)
	if err != nil {
		t.Fatal(err)
	}
	if folds {
		t.Skip("test filesystem lacks case sensitivity: Dir and dir fold into one directory, no coexistence to assert")
	}
	dest := filepath.Join(parent, "snap")
	if err := Extract(repo, commit, dest); err != nil {
		t.Fatalf("case-distinct directories must coexist on a case-sensitive filesystem: %v", err)
	}
	for name, want := range files {
		got, err := os.ReadFile(filepath.Join(dest, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, []byte(want)) {
			t.Fatalf("%s = %q, want %q", name, got, want)
		}
	}
}

// TestExtractRefusesFileDirectoryFold proves a regular file that folds onto a
// planned directory is refused with the duplicate-platform-path diagnostic on
// a case-folding destination, in both listing orders: the file sorts first
// (a/x.txt + A) and the directory sorts first (Dir/x.txt + dir). The
// file-first order is the discriminating one: without the pre-pass the file
// streams first and the directory then fails mid-stream with a raw MkdirAll
// error instead of the duplicate-platform-path class.
func TestExtractRefusesFileDirectoryFold(t *testing.T) {
	parent := t.TempDir()
	folds, err := destinationFoldsCase(parent)
	if err != nil {
		t.Fatal(err)
	}
	if !folds {
		t.Skip("test filesystem is case-sensitive: the folded file and directory coexist, no file-directory collision to refuse")
	}
	for _, files := range []map[string]string{
		{"a/x.txt": "x-bytes", "A": "file-bytes"},
		{"Dir/x.txt": "x-bytes", "dir": "file-bytes"},
	} {
		repo, commit := commitCaseTree(t, files)
		dest := filepath.Join(parent, "snap")
		err := Extract(repo, commit, dest)
		if err == nil || !strings.Contains(err.Error(), "duplicate platform path") {
			t.Fatalf("%v: err = %v, want file-directory fold refusal", files, err)
		}
		assertNothingWritten(t, parent, dest)
	}
}

// TestExtractRefusesNestedDirectoryComponentFold proves the gate tracks every
// ancestor prefix: A/B/x + a/b/y folds at the top component already and is
// refused on a case-folding destination.
func TestExtractRefusesNestedDirectoryComponentFold(t *testing.T) {
	files := map[string]string{"A/B/x": "x-bytes", "a/b/y": "y-bytes"}
	repo, commit := commitCaseTree(t, files)
	parent := t.TempDir()
	folds, err := destinationFoldsCase(parent)
	if err != nil {
		t.Fatal(err)
	}
	if !folds {
		t.Skip("test filesystem is case-sensitive: A/B and a/b coexist, no nested directory-component collision to refuse")
	}
	dest := filepath.Join(parent, "snap")
	err = Extract(repo, commit, dest)
	if err == nil || !strings.Contains(err.Error(), "duplicate platform path") {
		t.Fatalf("err = %v, want nested directory-component fold refusal", err)
	}
	assertNothingWritten(t, parent, dest)
}
