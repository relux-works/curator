package testgitenv

import (
	"crypto/sha256"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Independent of managed: every repository, storage, discovery, namespace and
// template override must be absent in the child, whose real Git commit must
// still succeed without changing any disposable decoy bytes or identities.
var repositorySelectors = []string{
	"GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_COMMON_DIR",
	"GIT_OBJECT_DIRECTORY", "GIT_ALTERNATE_OBJECT_DIRECTORIES", "GIT_NAMESPACE",
	"GIT_TEMPLATE_DIR", "GIT_CEILING_DIRECTORIES", "GIT_DISCOVERY_ACROSS_FILESYSTEM",
	"GIT_IMPLICIT_WORK_TREE", "GIT_GRAFT_FILE", "GIT_SHALLOW_FILE",
	"GIT_NO_REPLACE_OBJECTS", "GIT_REPLACE_REF_BASE", "GIT_PREFIX",
	"GIT_QUARANTINE_PATH", "GIT_REFERENCE_BACKEND", "GIT_DEFAULT_HASH", "GIT_DEFAULT_REF_FORMAT",
}

type decoyEntry struct {
	Mode fs.FileMode
	Size int64
	Hash [32]byte
	Link string
	Info fs.FileInfo
}

func decoyImage(t *testing.T, root string) map[string]decoyEntry {
	t.Helper()
	image := map[string]decoyEntry{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		v := decoyEntry{Mode: info.Mode(), Info: info}
		if info.Mode().IsRegular() {
			body, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			v.Size, v.Hash = info.Size(), sha256.Sum256(body)
		} else if info.Mode()&os.ModeSymlink != 0 {
			v.Link, err = os.Readlink(path)
			if err != nil {
				return err
			}
		}
		image[path] = v
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return image
}

func requireDecoyImage(t *testing.T, before, after map[string]decoyEntry) {
	t.Helper()
	if len(before) != len(after) {
		t.Errorf("decoy entry count changed: %d -> %d", len(before), len(after))
	}
	for path, b := range before {
		a, ok := after[path]
		if !ok || b.Mode != a.Mode || b.Size != a.Size || b.Hash != a.Hash || b.Link != a.Link || !os.SameFile(b.Info, a.Info) || (!b.Mode.IsDir() && !b.Info.ModTime().Equal(a.Info.ModTime())) {
			t.Errorf("decoy changed: %s", path) // Paths only; never emit payloads.
		}
	}
}

func newDecoy(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	fixtureGit(t, root, "init", "-q", "-b", "main")
	fixtureGit(t, root, "config", "fixture.sentinel", "preserved")
	if err := os.WriteFile(filepath.Join(root, "sentinel"), []byte("disposable decoy"), 0600); err != nil {
		t.Fatal(err)
	}
	fixtureGit(t, root, "add", "sentinel")
	fixtureGit(t, root, "-c", "user.name=Decoy", "-c", "user.email=decoy@example.invalid", "commit", "-qm", "decoy")
	// An untracked ancestry fixture leaf lets the old shell bypass really
	// stage and commit into this sacrificial repository, rather than merely
	// failing an unrelated missing-file setup.
	if err := os.WriteFile(filepath.Join(root, "content.txt"), []byte("decoy ancestry leaf"), 0600); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestRepositorySelectorIsolation(t *testing.T) {
	for _, selector := range append([]string{"clean"}, repositorySelectors...) {
		t.Run(selector, func(t *testing.T) {
			decoy := newDecoy(t)
			before := decoyImage(t, decoy)
			cmd := exec.Command(os.Args[0], "-test.v", "-test.run=^TestRepositorySelectorChild$")
			cmd.Env = append(os.Environ(), "CURATOR_SELECTOR_CHILD=1")
			if selector != "clean" {
				value := filepath.Join(decoy, ".git")
				switch selector {
				case "GIT_WORK_TREE", "GIT_CEILING_DIRECTORIES", "GIT_TEMPLATE_DIR":
					value = decoy
				case "GIT_OBJECT_DIRECTORY", "GIT_ALTERNATE_OBJECT_DIRECTORIES", "GIT_QUARANTINE_PATH":
					value = filepath.Join(decoy, ".git", "objects")
				case "GIT_INDEX_FILE", "GIT_GRAFT_FILE", "GIT_SHALLOW_FILE":
					value = filepath.Join(decoy, ".git", "index")
				case "GIT_NAMESPACE", "GIT_REPLACE_REF_BASE":
					value = "refs/hostile/"
				case "GIT_DISCOVERY_ACROSS_FILESYSTEM", "GIT_IMPLICIT_WORK_TREE", "GIT_NO_REPLACE_OBJECTS":
					value = "1"
				}
				cmd.Env = append(cmd.Env, selector+"="+value)
			}
			output, err := cmd.CombinedOutput()
			requireDecoyImage(t, before, decoyImage(t, decoy))
			if err != nil || !strings.Contains(string(output), "--- PASS: TestRepositorySelectorChild") {
				t.Fatalf("selector child: %v\n%s", err, output)
			}
		})
	}
}

func TestRepositorySelectorChild(t *testing.T) {
	if os.Getenv("CURATOR_SELECTOR_CHILD") == "" {
		return
	}
	for _, key := range repositorySelectors {
		if _, ok := os.LookupEnv(key); ok {
			t.Errorf("repository selector survived: %s", key)
		}
	}
	if t.Failed() {
		return // Never let an unfixed helper write through a surviving selector.
	}
	root := t.TempDir()
	fixtureGit(t, root, "init", "-q")
	for _, leaf := range []string{"config", "objects"} {
		got := strings.TrimSpace(string(fixtureGit(t, root, "rev-parse", "--path-format=absolute", "--git-path", leaf)))
		got, err := filepath.EvalSymlinks(got)
		want, wantErr := filepath.EvalSymlinks(filepath.Join(root, ".git", leaf))
		if err != nil || wantErr != nil || got != want {
			t.Fatalf("fixture %s escaped: %v / %v", leaf, err, wantErr)
		}
	}
	fixtureGit(t, root, "config", "fixture.sentinel", "fixture")
	fixtureGit(t, root, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "fixture", "--allow-empty")
	fixtureGit(t, root, "cat-file", "-e", "HEAD^{commit}")
}

// Unlike the environment-absence check, these probes ask real Git where it
// would read config and objects. No Git writes run after selector injection;
// an unfixed helper therefore exposes the decoy without modifying it.
func TestGitStoragePathIsolation(t *testing.T) {
	for _, selector := range []string{"clean", "GIT_COMMON_DIR", "GIT_OBJECT_DIRECTORY"} {
		t.Run(selector, func(t *testing.T) {
			decoy, fixture := newDecoy(t), newDecoy(t)
			before := decoyImage(t, decoy)
			cmd := exec.Command(os.Args[0], "-test.v", "-test.run=^TestGitStoragePathChild$")
			cmd.Env = append(os.Environ(), "CURATOR_STORAGE_PROBE="+fixture)
			switch selector {
			case "GIT_COMMON_DIR":
				cmd.Env = append(cmd.Env, selector+"="+filepath.Join(decoy, ".git"))
			case "GIT_OBJECT_DIRECTORY":
				cmd.Env = append(cmd.Env, selector+"="+filepath.Join(decoy, ".git", "objects"))
			}
			output, err := cmd.CombinedOutput()
			requireDecoyImage(t, before, decoyImage(t, decoy))
			if err != nil || !strings.Contains(string(output), "--- PASS: TestGitStoragePathChild") {
				t.Fatalf("read-only storage probe: %v\n%s", err, output)
			}
		})
	}
}

func TestGitStoragePathChild(t *testing.T) {
	fixture := os.Getenv("CURATOR_STORAGE_PROBE")
	if fixture == "" {
		return
	}
	for _, leaf := range []string{"config", "objects"} {
		got := strings.TrimSpace(string(fixtureGit(t, fixture, "rev-parse", "--path-format=absolute", "--git-path", leaf)))
		got, err := filepath.EvalSymlinks(got)
		want, wantErr := filepath.EvalSymlinks(filepath.Join(fixture, ".git", leaf))
		if err != nil || wantErr != nil || got != want {
			t.Errorf("read-only Git %s path escaped the fixture", leaf)
		}
	}
}

// A relative or symlinked temporary root cannot change the meaning of HOME,
// config or cleanup when the child changes cwd. The absolute row is a control.
func TestPhysicalTemporaryRoots(t *testing.T) {
	for _, mode := range []string{"absolute", "relative", "symlink", "unavailable"} {
		t.Run(mode, func(t *testing.T) {
			root, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			tmp := filepath.Join(root, ".tmp")
			if err := os.Mkdir(tmp, 0700); err != nil {
				t.Fatal(err)
			}
			selected := tmp
			switch mode {
			case "relative":
				selected = ".tmp"
			case "unavailable":
				selected = filepath.Join(root, "missing")
			case "symlink":
				selected = filepath.Join(root, "link")
				if err := os.Symlink(tmp, selected); err != nil {
					t.Skipf("directory symlink unavailable: %v", err)
				}
			}
			report := filepath.Join(root, "home-report")
			cmd := exec.Command(os.Args[0], "-test.v", "-test.run=^TestPhysicalTemporaryRootChild$")
			cmd.Dir = root
			cmd.Env = append(os.Environ(), "TMPDIR="+selected, "TMP="+selected, "TEMP="+selected, "CURATOR_TEMP_CHILD="+root)
			output, childErr := cmd.CombinedOutput()
			if mode == "unavailable" {
				exit, ok := childErr.(*exec.ExitError)
				if !ok || exit.ExitCode() != 2 || !strings.Contains(string(output), "testgitenv: resolve temporary root") || strings.Contains(string(output), "=== RUN") {
					t.Fatalf("unavailable root did not refuse before dispatch: %v\n%s", childErr, output)
				}
				if _, err := os.Stat(report); !os.IsNotExist(err) {
					t.Fatal("unavailable root reached fixture writes")
				}
				return
			}
			if childErr != nil || !strings.Contains(string(output), "--- PASS: TestPhysicalTemporaryRootChild") {
				t.Fatalf("temporary root child: %v\n%s", childErr, output)
			}
			home, err := os.ReadFile(report)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := os.Lstat(filepath.Dir(string(home))); !os.IsNotExist(err) {
				t.Fatalf("cleanup after cwd change retained scratch root: %v", err)
			}
		})
	}
}

func TestPhysicalTemporaryRootChild(t *testing.T) {
	root := os.Getenv("CURATOR_TEMP_CHILD")
	if root == "" {
		return
	}
	for _, key := range []string{"HOME", "XDG_CONFIG_HOME", "GIT_CONFIG_GLOBAL", "TMPDIR", "TMP", "TEMP"} {
		value := os.Getenv(key)
		physical, err := filepath.EvalSymlinks(value)
		if !filepath.IsAbs(value) || err != nil || physical != value {
			t.Fatalf("%s is not absolute and physical", key)
		}
	}
	home, config := os.Getenv("HOME"), os.Getenv("GIT_CONFIG_GLOBAL")
	if err := os.Mkdir(filepath.Join(root, "other"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(filepath.Join(root, "other")); err != nil {
		t.Fatal(err)
	}
	short := ShortTempDir(t)
	physical, err := filepath.EvalSymlinks(short)
	if !filepath.IsAbs(short) || err != nil || physical != short {
		t.Fatal("short fixture root moved with cwd")
	}
	fixtureGit(t, ".", "config", "--global", "fixture.sentinel", "isolated")
	body, err := os.ReadFile(config)
	if err != nil || !strings.Contains(string(body), "isolated") {
		t.Fatal("global fixture config moved with cwd")
	}
	if err := os.WriteFile(filepath.Join(home, "native-fixture"), []byte("isolated"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "home-report"), []byte(home), 0600); err != nil {
		t.Fatal(err)
	}
}

// Nested boundaries restore the exact selector values (including absent/empty)
// and remove new selectors introduced inside the inner scope.
func TestSelectorRestoration(t *testing.T) {
	for _, key := range repositorySelectors {
		t.Setenv(key, "")
	}
	before := append([]string(nil), os.Environ()...)
	cleanup := Isolate()
	if err := os.Setenv("GIT_CONFIG_KEY_REWORK_NEW", "fixture.new"); err != nil {
		t.Fatal(err)
	}
	cleanup()
	if !reflect.DeepEqual(envMap(before), envMap(os.Environ())) {
		t.Fatal("selector boundary did not restore exact environment")
	}
}

func envMap(env []string) map[string]string {
	result := map[string]string{}
	for _, entry := range env {
		key, value, _ := strings.Cut(entry, "=")
		result[key] = value
	}
	return result
}

// The installed Git's own repository-local inventory cannot silently grow
// beyond the independent admission list and dynamic configuration channels.
func TestInstalledGitSelectorInventory(t *testing.T) {
	output := fixtureGit(t, t.TempDir(), "rev-parse", "--local-env-vars")
	for _, key := range strings.Fields(string(output)) {
		if strings.HasPrefix(key, "GIT_CONFIG") {
			continue
		}
		found := false
		for _, known := range repositorySelectors {
			found = found || known == key
		}
		if !found {
			t.Errorf("unqualified repository selector: %s", key)
		}
	}
}
