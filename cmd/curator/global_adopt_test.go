package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/globalbins"
	"github.com/relux-works/curator/internal/runtimestore"
)

type globalAdoptCLIFixture struct {
	home       string
	configPath string
	userHome   string
	userBin    string
	canonical  string
	published  string
}

func newGlobalAdoptCLIFixture(t *testing.T) globalAdoptCLIFixture {
	t.Helper()
	home := globalScopeDeclaring(t, "")
	configPath := filepath.Join(home, "config.json")
	userHome := filepath.Join(filepath.Dir(home), "user")
	userBin := filepath.Join(userHome, ".local", "bin")
	if err := os.MkdirAll(userBin, 0o755); err != nil {
		t.Fatal(err)
	}
	canonicalName, publishedName := "tool", "tool"
	if runtime.GOOS == "windows" {
		canonicalName += ".cmd"
		publishedName += ".cmd"
	}
	canonical := filepath.Join(home, "global", "bin", canonicalName)
	published := filepath.Join(userBin, publishedName)
	if err := os.MkdirAll(filepath.Dir(canonical), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv(globalbins.UserBinEnv, userBin)
	t.Setenv("PATH", userBin)
	return globalAdoptCLIFixture{
		home: home, configPath: configPath, userHome: userHome, userBin: userBin,
		canonical: canonical, published: published,
	}
}

func (fixture globalAdoptCLIFixture) invoke(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	return captureWithUserHome(t, fixture.configPath, func() (string, error) { return fixture.userHome, nil }, args...)
}

func (fixture globalAdoptCLIFixture) canonicalShim() []byte {
	if runtime.GOOS == "windows" {
		return []byte(runtimestore.WindowsShimContent(fixture.canonical, nil))
	}
	return []byte(runtimestore.UnixShimContent(fixture.canonical, nil))
}

func TestGlobalAdoptCLIAdoptsAndIsIdempotent(t *testing.T) {
	fixture := newGlobalAdoptCLIFixture(t)
	if err := os.WriteFile(fixture.canonical, []byte("canonical target\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	want := fixture.canonicalShim()
	if err := os.WriteFile(fixture.published, want, 0o755); err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := fixture.invoke(t, "global", "adopt", "tool", "--dry-run")
	if code != exitOK || !strings.Contains(stdout, "would adopt command") {
		t.Fatalf("global adopt --dry-run = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	requireOnePermissivePostureWarning(t, stderr)
	if _, err := os.Lstat(filepath.Join(fixture.userBin, ".curator-managed.json")); !os.IsNotExist(err) {
		t.Fatalf("CLI dry-run wrote ownership marker: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(fixture.home, "backups")); !os.IsNotExist(err) {
		t.Fatalf("CLI dry-run wrote backup root: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(fixture.home, "state")); !os.IsNotExist(err) {
		t.Fatalf("CLI dry-run created manager lock state: %v", err)
	}

	code, stdout, stderr = fixture.invoke(t, "global", "adopt", "tool")
	if code != exitOK || !strings.Contains(stdout, "adopted command") || !strings.Contains(stdout, "backup:") {
		t.Fatalf("global adopt = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	requireOnePermissivePostureWarning(t, stderr)
	code, stdout, stderr = fixture.invoke(t, "global", "adopt", "tool")
	if code != exitOK || !strings.Contains(stdout, "already managed") {
		t.Fatalf("second global adopt = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	requireOnePermissivePostureWarning(t, stderr)
	backups, err := os.ReadDir(filepath.Join(fixture.home, "backups", "global-bins"))
	if err != nil || len(backups) != 1 {
		t.Fatalf("CLI backup count = (%d, %v), want 1", len(backups), err)
	}
}

func TestGlobalAdoptCLIRefusesByteMismatch(t *testing.T) {
	fixture := newGlobalAdoptCLIFixture(t)
	if err := os.WriteFile(fixture.canonical, []byte("canonical target\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fixture.published, []byte("manually edited shim\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(fixture.published)
	if err != nil {
		t.Fatal(err)
	}

	code, _, stderr := fixture.invoke(t, "global", "adopt", "tool")
	if code != exitFail || !strings.Contains(stderr, fixture.published) || !strings.Contains(stderr, "bytes differ") {
		t.Fatalf("global adopt mismatch = %d\nstderr:\n%s; want non-zero with path and byte-mismatch reason", code, stderr)
	}
	if _, err := os.Lstat(filepath.Join(fixture.userBin, ".curator-managed.json")); !os.IsNotExist(err) {
		t.Fatalf("mismatch refusal wrote ownership marker: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(fixture.home, "backups")); !os.IsNotExist(err) {
		t.Fatalf("mismatch refusal wrote backup root: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(fixture.home, "state")); !os.IsNotExist(err) {
		t.Fatalf("mismatch refusal created manager lock state: %v", err)
	}
	got, err := os.ReadFile(fixture.published)
	if err != nil || !bytes.Equal(got, original) {
		t.Fatalf("mismatch refusal changed user entry: %q, %v", got, err)
	}
}

func TestGlobalAdoptCLIRefusesUnknownAndMissingCommandsWithoutWrites(t *testing.T) {
	for _, testCase := range []struct {
		name       string
		command    string
		wantReason string
	}{
		{name: "unknown", command: "unknown-tool", wantReason: "no Curator global command has a canonical target"},
		{name: "missing entry", command: "tool", wantReason: "entry does not exist"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newGlobalAdoptCLIFixture(t)
			if err := os.WriteFile(fixture.canonical, []byte("canonical target\n"), 0o755); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(fixture.userBin, testCase.command)
			if runtime.GOOS == "windows" {
				path += ".cmd"
			}
			var original []byte
			if testCase.name == "unknown" {
				original = []byte("unrelated command\n")
				if err := os.WriteFile(path, original, 0o755); err != nil {
					t.Fatal(err)
				}
			}

			code, _, stderr := fixture.invoke(t, "global", "adopt", testCase.command)
			if code != exitFail || !strings.Contains(stderr, path) || !strings.Contains(stderr, testCase.wantReason) {
				t.Fatalf("global adopt %s = %d\nstderr:\n%s; want path %s and reason %q", testCase.command, code, stderr, path, testCase.wantReason)
			}
			if _, err := os.Lstat(filepath.Join(fixture.userBin, ".curator-managed.json")); !os.IsNotExist(err) {
				t.Fatalf("refusal wrote ownership marker: %v", err)
			}
			if _, err := os.Lstat(filepath.Join(fixture.home, "backups")); !os.IsNotExist(err) {
				t.Fatalf("refusal wrote backup root: %v", err)
			}
			if _, err := os.Lstat(filepath.Join(fixture.home, "state")); !os.IsNotExist(err) {
				t.Fatalf("refusal created manager lock state: %v", err)
			}
			if testCase.name == "unknown" {
				got, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(got, original) {
					t.Fatalf("unknown-command refusal changed user entry: %q, %v", got, err)
				}
			}
		})
	}
}
