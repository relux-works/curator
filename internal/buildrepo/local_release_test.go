package buildrepo

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// scratchTempDir points every os.TempDir lookup of this test at a fresh
// directory, so a leaked operation-private root is observable as an entry
// rather than lost among the host's own temporary files.
func scratchTempDir(t *testing.T) string {
	t.Helper()
	scratch := t.TempDir()
	for _, name := range []string{"TMPDIR", "TMP", "TEMP"} {
		t.Setenv(name, scratch)
	}
	return scratch
}

func leakedPrivateRoots(t *testing.T, scratch string) []string {
	t.Helper()
	entries, err := os.ReadDir(scratch)
	if err != nil {
		t.Fatal(err)
	}
	var leaked []string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "curator-buildrepo-local-") {
			leaked = append(leaked, entry.Name())
		}
	}
	return leaked
}

func TestAdmitLocalReleasesSealedPrivateRoot(t *testing.T) {
	scratch := scratchTempDir(t)
	fixture := makeGitFixture(t, "sha1", false)
	if _, err := AdmitLocal(context.Background(), LocalRequest{Path: fixture.work, Tool: realGitTool(t)}); err != nil {
		t.Fatalf("local admission: %v", err)
	}
	if leaked := leakedPrivateRoots(t, scratch); len(leaked) != 0 {
		t.Fatalf("successful admission left its private root behind: %v", leaked)
	}
}

func TestAdmitLocalReleasesPrivateRootOnRefusal(t *testing.T) {
	scratch := scratchTempDir(t)
	fixture := makeGitFixture(t, "sha1", false)
	request := LocalRequest{Path: fixture.work, Tool: realGitTool(t), afterObjectCopy: func() {
		configPath := filepath.Join(fixture.work, ".git", "config")
		if err := os.WriteFile(configPath, []byte("[core]\n\trepositoryformatversion = 0\n\tbare = false\n\tchanged = true\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}}
	if _, err := AdmitLocal(context.Background(), request); err == nil {
		t.Fatal("admission accepted a repository that changed during object copy")
	}
	if leaked := leakedPrivateRoots(t, scratch); len(leaked) != 0 {
		t.Fatalf("refused admission left its private root behind: %v", leaked)
	}
}

func TestReleasePrivateRootRemovesSealedObjectStore(t *testing.T) {
	root := filepath.Join(t.TempDir(), "private")
	objects := filepath.Join(root, "repo.git", "objects", "ab")
	if err := os.MkdirAll(objects, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(objects, "cd"), []byte("blob"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := sealObjectStore(filepath.Join(root, "repo.git", "objects")); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && os.Getuid() != 0 {
		// The sealed store is exactly what a plain removal cannot take apart.
		if err := os.RemoveAll(root); err == nil {
			t.Fatal("plain RemoveAll removed a sealed object store; the regression this guards no longer reproduces")
		}
	}
	if err := releasePrivateRoot(root); err != nil {
		t.Fatalf("release sealed private root: %v", err)
	}
	if _, err := os.Lstat(root); !os.IsNotExist(err) {
		t.Fatalf("private root still present after release: %v", err)
	}
	if err := releasePrivateRoot(root); err != nil {
		t.Fatalf("releasing an absent root must be idempotent: %v", err)
	}
}
