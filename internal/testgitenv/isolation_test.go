package testgitenv

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	cleanup := Isolate()
	code := m.Run()
	cleanup()
	os.Exit(code)
}

func fixtureGit(t *testing.T, dir string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
	return output
}

// Every inherited config channel and repository selector is removed; native
// home writes use scratch space, and cleanup restores exact previous values.
func TestIsolateEnvironmentAndHome(t *testing.T) {
	// This acceptance list is independent of managed: omitting a key from the
	// implementation must not also remove its assertion from the test.
	inheritedKeys := []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_CONFIG", "GIT_CONFIG_SYSTEM", "GIT_CONFIG_GLOBAL", "GIT_CONFIG_NOSYSTEM", "GIT_CONFIG_COUNT", "GIT_CONFIG_KEY_0", "GIT_CONFIG_VALUE_0", "GIT_CONFIG_PARAMETERS", "GIT_CEILING_DIRECTORIES", "SSH_AUTH_SOCK"}
	for _, key := range inheritedKeys {
		t.Setenv(key, "poison")
	}
	operatorHome := t.TempDir()
	t.Setenv("HOME", operatorHome)
	t.Setenv("XDG_CONFIG_HOME", "")
	sentinel := filepath.Join(operatorHome, ".claude", "CLAUDE.md")
	if err := os.MkdirAll(filepath.Dir(sentinel), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sentinel, []byte("operator"), 0o600); err != nil {
		t.Fatal(err)
	}
	before := os.Environ()
	cleanup := Isolate()
	home := os.Getenv("HOME")
	func() {
		defer cleanup()
		if home == operatorHome {
			t.Fatal("HOME was not isolated")
		}
		if resolved, err := os.UserHomeDir(); err != nil || resolved != home {
			t.Fatalf("native home = %q, want %q: %v", resolved, home, err)
		}
		for _, key := range inheritedKeys {
			if key == "GIT_CONFIG_GLOBAL" || key == "GIT_CONFIG_NOSYSTEM" {
				continue // These are replaced by owned, empty configuration below.
			}
			if value, exists := os.LookupEnv(key); exists {
				t.Errorf("inherited %s=%q", key, value)
			}
		}
		if config, err := os.ReadFile(os.Getenv("GIT_CONFIG_GLOBAL")); err != nil || len(config) != 0 || os.Getenv("GIT_CONFIG_NOSYSTEM") != "1" {
			t.Fatalf("unsafe Git configuration: %q, %v", config, err)
		}
		if info, err := os.Stat(os.Getenv("XDG_CONFIG_HOME")); err != nil || !info.IsDir() {
			t.Fatalf("XDG_CONFIG_HOME is not a scratch directory: %v", err)
		}
		native := filepath.Join(home, ".claude", "CLAUDE.md")
		if err := os.MkdirAll(filepath.Dir(native), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(native, []byte("fixture"), 0o600); err != nil {
			t.Fatal(err)
		}
	}()
	for _, entry := range before {
		key, value, _ := strings.Cut(entry, "=")
		if got, ok := os.LookupEnv(key); !ok || got != value {
			t.Errorf("%s was not restored", key)
		}
	}
	if _, err := os.Stat(home); !os.IsNotExist(err) {
		t.Fatalf("scratch home retained: %v", err)
	}
	if body, err := os.ReadFile(sentinel); err != nil || string(body) != "operator" {
		t.Fatalf("operator native home changed: %q, %v", body, err)
	}
}

// A child runs an actual scratch commit through TestMain, with hostile repo
// selectors inherited from its parent. Both success and decoy preservation
// are required; a helper that merely breaks all Git commands cannot pass.
func TestFixtureLeavesHostileRepositoryUntouched(t *testing.T) {
	root := t.TempDir()
	decoy := filepath.Join(root, "decoy")
	if err := os.Mkdir(decoy, 0o700); err != nil {
		t.Fatal(err)
	}
	fixtureGit(t, decoy, "init", "-q")
	fixtureGit(t, decoy, "config", "fixture.sentinel", "preserved")
	fixtureGit(t, decoy, "-c", "user.name=Decoy", "-c", "user.email=decoy@example.invalid", "commit", "-qm", "decoy", "--allow-empty")
	snapshot := func() [][]byte {
		config, err := os.ReadFile(filepath.Join(decoy, ".git", "config"))
		if err != nil {
			t.Fatal(err)
		}
		return [][]byte{fixtureGit(t, decoy, "show-ref"), config, fixtureGit(t, decoy, "count-objects", "-v")}
	}
	before := snapshot()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(executable, "-test.v", "-test.run=^TestScratchFixtureChild$")
	cmd.Env = append(os.Environ(), "CURATOR_TEST_HOSTILE_FIXTURE=1",
		"GIT_DIR="+filepath.Join(decoy, ".git"), "GIT_WORK_TREE="+decoy,
		"GIT_INDEX_FILE="+filepath.Join(decoy, ".git", "index"),
		"GIT_CEILING_DIRECTORIES="+root, "SSH_AUTH_SOCK="+filepath.Join(root, "agent"),
		"GIT_CONFIG_COUNT=1", "GIT_CONFIG_KEY_0=fixture.injected", "GIT_CONFIG_VALUE_0=hostile")
	output, childErr := cmd.CombinedOutput()
	after := snapshot()
	for i, label := range []string{"refs", "config", "objects"} {
		if !bytes.Equal(before[i], after[i]) {
			t.Errorf("decoy %s changed:\nbefore: %s\nafter: %s", label, before[i], after[i])
		}
	}
	if childErr != nil || !strings.Contains(string(output), "--- PASS: TestScratchFixtureChild") {
		t.Fatalf("scratch fixture failed: %v\n%s", childErr, output)
	}
}

// The child's fixture changes its own config and creates a commit. The resolved
// Git directory must belong to that scratch worktree, never the decoy.
func TestScratchFixtureChild(t *testing.T) {
	if os.Getenv("CURATOR_TEST_HOSTILE_FIXTURE") != "1" {
		return
	}
	fixture := t.TempDir()
	fixtureGit(t, fixture, "init", "-q")
	fixtureGit(t, fixture, "config", "fixture.sentinel", "fixture")
	fixtureGit(t, fixture, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "fixture", "--allow-empty")
	actual := strings.TrimSpace(string(fixtureGit(t, fixture, "rev-parse", "--absolute-git-dir")))
	want, err := filepath.EvalSymlinks(filepath.Join(fixture, ".git"))
	if err != nil {
		t.Fatal(err)
	}
	actual, err = filepath.EvalSymlinks(actual)
	if err != nil || actual != want {
		t.Fatalf("fixture Git directory = %q, want %q: %v", actual, want, err)
	}
}
