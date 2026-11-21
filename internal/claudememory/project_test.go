package claudememory

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A real checkout requires Git to verify its boundary. Missing Git refuses
// that checkout, while an outside-Git directory still has a usable boundary.
func TestProjectRootMissingGitRefusesCheckout(t *testing.T) {
	root := filepath.Join(t.TempDir(), "project")
	if output, err := exec.Command("git", "init", "--quiet", root).CombinedOutput(); err != nil {
		t.Fatalf("fixture Git init: %v: %s", err, output)
	}
	if got, err := ProjectRoot(root); err != nil || got != root {
		t.Fatalf("Git-present positive control: %q, %v", got, err)
	}
	t.Setenv("PATH", t.TempDir())
	if got, err := ProjectRoot(root); err == nil || got != "" || !strings.Contains(err.Error(), "resolve launch Git toplevel failed") {
		t.Fatalf("checkout with missing Git did not specifically refuse: %q, %v", got, err)
	}
	outside := t.TempDir()
	if got, err := ProjectRoot(outside); err != nil || got != outside {
		t.Fatalf("outside-Git directory with missing Git: %q, %v", got, err)
	}
}

// Git toplevel, rather than nested launch directory, bounds ancestor policy.
// Ambient GIT_DIR cannot redirect it; a lexical root alias is retained.
func TestProjectRootCheckoutAndAlias(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "project")
	launch := filepath.Join(root, "nested", "deeper")
	if err := os.MkdirAll(launch, 0o700); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "init", "--quiet", root)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fixture Git init: %v: %s", err, output)
	}
	t.Setenv("GIT_DIR", filepath.Join(base, "foreign", ".git"))
	if got, err := ProjectRoot(launch); err != nil || got != root {
		t.Fatalf("wrong checkout boundary: %q, %v", got, err)
	}
	alias := filepath.Join(base, "alias")
	if err := os.Symlink(root, alias); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if got, err := ProjectRoot(filepath.Join(alias, "nested")); err != nil || got != alias {
		t.Fatalf("lexical checkout alias lost: %q, %v", got, err)
	}
}

// Outside Git uses the launch directory. Corrupt metadata and invalid paths
// refuse without claiming that a broken checkout is an outside-Git project.
func TestProjectRootOutsideAndBrokenRepository(t *testing.T) {
	root := t.TempDir()
	if got, err := ProjectRoot(root); err != nil || got != root {
		t.Fatalf("outside-Git root: %q, %v", got, err)
	}
	if err := os.Mkdir(filepath.Join(root, "objects"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "HEAD"), []byte("ordinary fixture document"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got, err := ProjectRoot(root); err != nil || got != root {
		t.Fatalf("ordinary HEAD/objects names falsely treated as Git: %q, %v", got, err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git"), []byte("invalid fixture metadata\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, launch := range []string{root, "relative", filepath.Join(root, "missing")} {
		if got, err := ProjectRoot(launch); err == nil || got != "" {
			t.Fatalf("invalid launch was admitted: %q", launch)
		}
	}
}

// A legitimate .git indirection still resolves the working-tree boundary;
// an actual bare repository refuses rather than becoming an outside-Git root.
func TestProjectRootGitFileAndBareRepository(t *testing.T) {
	t.Run("separate-git-directory", func(t *testing.T) {
		base := t.TempDir()
		root := filepath.Join(base, "project")
		metadata := filepath.Join(base, "metadata")
		cmd := exec.Command("git", "init", "--quiet", "--separate-git-dir", metadata, root)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("fixture separate Git directory: %v: %s", err, output)
		}
		launch := filepath.Join(root, "nested")
		if err := os.Mkdir(launch, 0o700); err != nil {
			t.Fatal(err)
		}
		if got, err := ProjectRoot(launch); err != nil || got != root {
			t.Fatalf("legitimate Git file refused: %q, %v", got, err)
		}
	})
	t.Run("bare-repository", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "bare")
		cmd := exec.Command("git", "init", "--quiet", "--bare", root)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("fixture bare repository: %v: %s", err, output)
		}
		got, err := ProjectRoot(root)
		if err == nil || got != "" || !strings.Contains(err.Error(), "Git toplevel") {
			t.Fatalf("bare repository did not refuse the absent working-tree boundary: %q, %v", got, err)
		}
	})
}
