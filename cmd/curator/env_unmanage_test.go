package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/stateread"
)

type readFailureVectorDocument struct {
	Cases []readFailureVector `json:"cases"`
}

type readFailureVector struct {
	Name         string                    `json:"name"`
	Operation    string                    `json:"operation"`
	FileClass    string                    `json:"file_class"`
	Presence     string                    `json:"presence"`
	EntryKind    string                    `json:"entry_kind"`
	FailureClass string                    `json:"failure_class"`
	Expected     readFailureVectorExpected `json:"expected"`
}

type readFailureVectorExpected struct {
	Diagnostic *string `json:"diagnostic"`
	Written    *bool   `json:"written"`
}

type unmanageFixture struct {
	source      stubConfigSource
	managerHome string
	nativeHome  string
	backupRoot  string
	backupFile  string
	markerPath  string
	surfacePath string
}

func TestEnvUnmanageBackupRecordVectors(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "environments-read-failure.json"))
	if err != nil {
		t.Fatal(err)
	}
	var document readFailureVectorDocument
	if err := json.Unmarshal(payload, &document); err != nil {
		t.Fatalf("decode environments-read-failure.json: %v", err)
	}
	var cases []readFailureVector
	for _, published := range document.Cases {
		if published.Operation == "restore" && published.FileClass == "backup-record" {
			cases = append(cases, published)
		}
	}
	want := []string{
		"backup-record-unreadable-restore-stops",
		"backup-record-absent-restore-nothing",
	}
	if len(cases) != len(want) {
		t.Fatalf("rc.13 publishes %d backup-record restore vectors, want %d", len(cases), len(want))
	}
	got := map[string]bool{}
	for _, item := range cases {
		got[item.Name] = true
	}
	for _, name := range want {
		if !got[name] {
			t.Fatalf("rc.13 backup-record restore vector %q is missing", name)
		}
	}

	conformancecoverage.Run(t, "environments-read-failure/restore", cases,
		func(item readFailureVector) string { return item.Name },
		func(caseT *testing.T, item readFailureVector) {
			runBackupRecordRestoreVector(caseT, item)
		})
}

func runBackupRecordRestoreVector(t *testing.T, item readFailureVector) {
	t.Helper()
	withBackup := item.Presence == "present" && item.EntryKind == "directory"
	fixture := newUnmanageFixture(t, withBackup)
	if item.Expected.Written == nil || *item.Expected.Written {
		t.Fatalf("vector %s expected.written = %v; this consumer requires the rc.13 no-write assertion", item.Name, item.Expected.Written)
	}
	if item.Presence == "absent" {
		if _, err := os.Lstat(fixture.backupRoot); !os.IsNotExist(err) {
			t.Fatalf("absent backup-record fixture has inventory %s (%v)", fixture.backupRoot, err)
		}
	}

	var readDir func(string) (stateread.Directory, error)
	if item.FailureClass == "io-error" && item.Presence == "present" {
		readDir = func(directory string) (stateread.Directory, error) {
			if filepath.Clean(directory) == filepath.Clean(fixture.backupRoot) {
				return stateread.Directory{Kind: stateread.KindUnreadable}, &stateread.Error{
					Kind: stateread.KindUnreadable, Path: directory, Cause: errors.New("injected I/O error"),
				}
			}
			return stateread.ReadDir(directory)
		}
	}

	beforeSurface := readFileForUnmanageTest(t, fixture.surfacePath)
	beforeMarker := readFileForUnmanageTest(t, fixture.markerPath)
	beforeBackup := []byte(nil)
	if withBackup {
		beforeBackup = readFileForUnmanageTest(t, fixture.backupFile)
	}
	beforeCurrent, err := envprofile.Current(fixture.managerHome)
	if err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := runEnvUnmanageWithReadDir(t, fixture.source, readDir, "--restore-backups")
	if item.Expected.Diagnostic != nil {
		if code != exitFail || !strings.Contains(stderr, *item.Expected.Diagnostic) || !strings.Contains(stderr, stateread.DiagUnreadable) {
			t.Fatalf("%s CLI result = %d\nstdout:\n%s\nstderr:\n%s; want typed %s refusal", item.Name, code, stdout, stderr, *item.Expected.Diagnostic)
		}
		if got := readFileForUnmanageTest(t, fixture.surfacePath); string(got) != string(beforeSurface) {
			t.Fatal("unreadable backup inventory changed the managed surface before refusing")
		}
		if got := readFileForUnmanageTest(t, fixture.markerPath); string(got) != string(beforeMarker) {
			t.Fatal("unreadable backup inventory removed or replaced the marker")
		}
		if got := readFileForUnmanageTest(t, fixture.backupFile); string(got) != string(beforeBackup) {
			t.Fatal("unreadable backup inventory changed the recorded backup")
		}
		if got, err := envprofile.Current(fixture.managerHome); err != nil || got != beforeCurrent {
			t.Fatalf("unreadable backup inventory changed the current profile: %q (%v), want %q", got, err, beforeCurrent)
		}
		return
	}

	if code != exitOK {
		t.Fatalf("%s CLI result = %d\nstdout:\n%s\nstderr:\n%s", item.Name, code, stdout, stderr)
	}
	if strings.Contains(stderr, envregistry.DiagBackupRecordUnreadable) {
		t.Fatalf("absent inventory was treated as unreadable:\n%s", stderr)
	}
	if _, err := os.Lstat(fixture.surfacePath); !os.IsNotExist(err) {
		t.Fatalf("unmanage did not remove the recorded managed surface: %s (%v)", fixture.surfacePath, err)
	}
	if _, err := os.Lstat(fixture.markerPath); !os.IsNotExist(err) {
		t.Fatalf("unmanage did not remove the marker: %s (%v)", fixture.markerPath, err)
	}
	if _, err := os.Lstat(fixture.backupRoot); !os.IsNotExist(err) {
		t.Fatalf("unmanage created or changed the absent backup inventory: %s (%v)", fixture.backupRoot, err)
	}
	if got, err := envprofile.Current(fixture.managerHome); err != nil || got != "" {
		t.Fatalf("successful default unmanage left current profile %q (%v)", got, err)
	}
}

func newUnmanageFixture(t *testing.T, withTakeoverBackup bool) unmanageFixture {
	t.Helper()
	source, managerHome := profileHome(t)
	native := os.Getenv("CLAUDE_CONFIG_DIR")
	if err := os.MkdirAll(native, 0o755); err != nil {
		t.Fatal(err)
	}
	surface := filepath.Join(native, "CLAUDE.md")
	if withTakeoverBackup {
		if err := os.WriteFile(surface, []byte("operator-owned context\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// A first install activates automatically when there is no current
	// profile. Seed a distinct current so the explicit scoped profile use
	// below is the operation that records any takeover backup.
	if err := envprofile.SetCurrent(managerHome, "prior-profile"); err != nil {
		t.Fatal(err)
	}
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "managed context\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("profile install = %d\nstderr:\n%s", code, stderr)
	}
	arguments := []string{"profile", "use", "acme", "--env", "claude_code"}
	if withTakeoverBackup {
		arguments = append(arguments, "--takeover")
	}
	if code, _, stderr := runProfile(t, source, arguments...); code != exitOK {
		t.Fatalf("profile use = %d\nstderr:\n%s", code, stderr)
	}
	if err := envprofile.SetCurrent(managerHome, "acme"); err != nil {
		t.Fatal(err)
	}
	fixture := unmanageFixture{
		source: source, managerHome: managerHome, nativeHome: native,
		backupRoot:  filepath.Join(native, ".agent-environment-backup"),
		markerPath:  filepath.Join(native, ".agent-environment.json"),
		surfacePath: surface,
	}
	if withTakeoverBackup {
		fixture.backupFile = filepath.Join(fixture.backupRoot, "1", "CLAUDE.md")
		if got := readFileForUnmanageTest(t, fixture.backupFile); string(got) != "operator-owned context\n" {
			t.Fatalf("takeover backup = %q, want original operator bytes", got)
		}
	}
	return fixture
}

func runEnvUnmanageWithReadDir(t *testing.T, source stubConfigSource, readDir func(string) (stateread.Directory, error), args ...string) (int, string, string) {
	return runEnvUnmanageWithReaders(t, source, nil, readDir, args...)
}

func runEnvUnmanageWithReaders(t *testing.T, source stubConfigSource, lstat func(string) (stateread.Metadata, error), readDir func(string) (stateread.Directory, error), args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr strings.Builder
	command := cli{
		config: source, stdout: &stdout, stderr: &stderr, userHome: os.UserHomeDir,
		unmanageBackupLstat: lstat, unmanageBackupReadDir: readDir,
	}
	code := command.run(append([]string{"env", "unmanage"}, args...))
	return code, stdout.String(), stderr.String()
}

func readFileForUnmanageTest(t *testing.T, path string) []byte {
	t.Helper()
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	return payload
}

func TestEnvUnmanageDoesNotRestoreWithoutAnObservedRecord(t *testing.T) {
	fixture := newUnmanageFixture(t, true)
	lstat := func(path string) (stateread.Metadata, error) {
		if filepath.Clean(path) == filepath.Clean(fixture.backupRoot) {
			return stateread.Metadata{Kind: stateread.KindAbsent}, nil
		}
		return stateread.Lstat(path)
	}
	code, stdout, stderr := runEnvUnmanageWithReaders(t, fixture.source, lstat, nil, "--restore-backups", "--env", "claude_code")
	if code != exitOK {
		t.Fatalf("unmanage with absent backup-record view = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if payload, err := os.ReadFile(fixture.surfacePath); !os.IsNotExist(err) {
		t.Fatalf("a physical generation was restored without an observed backup record: %q (%v)", payload, err)
	}
	if got := readFileForUnmanageTest(t, fixture.backupFile); string(got) != "operator-owned context\n" {
		t.Fatalf("the backup record itself changed: %q", got)
	}
}

func TestEnvUnmanageRestoresNewestBackupGeneration(t *testing.T) {
	fixture := newUnmanageFixture(t, true)
	newest := filepath.Join(fixture.backupRoot, "2", "CLAUDE.md")
	if err := os.MkdirAll(filepath.Dir(newest), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newest, []byte("newest operator context\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := runEnvUnmanageWithReadDir(t, fixture.source, nil, "--restore-backups", "--env", "claude_code")
	if code != exitOK {
		t.Fatalf("env unmanage --restore-backups = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if got := readFileForUnmanageTest(t, fixture.surfacePath); string(got) != "newest operator context\n" {
		t.Fatalf("restored surface = %q, want newest generation bytes", got)
	}
	if _, err := os.Lstat(fixture.markerPath); !os.IsNotExist(err) {
		t.Fatalf("unmanage did not remove the marker: %s (%v)", fixture.markerPath, err)
	}
	if got := readFileForUnmanageTest(t, fixture.backupFile); string(got) != "operator-owned context\n" {
		t.Fatalf("restore changed the prior backup generation: %q", got)
	}
	if got := readFileForUnmanageTest(t, newest); string(got) != "newest operator context\n" {
		t.Fatalf("restore changed the newest backup generation: %q", got)
	}
	if got, err := envprofile.Current(fixture.managerHome); err != nil || got != "acme" {
		t.Fatalf("scoped unmanage changed machine current profile: %q (%v), want acme", got, err)
	}
	if got, err := envprofile.ScopedCurrents(fixture.managerHome); err != nil {
		t.Fatal(err)
	} else if _, exists := got["env:claude_code"]; exists {
		t.Fatalf("scoped unmanage retained its current profile record: %v", got)
	}
}

func TestEnvUnmanageRestoresRecordedTakeoverBackup(t *testing.T) {
	fixture := newUnmanageFixture(t, true)
	code, stdout, stderr := runEnvUnmanageWithReadDir(t, fixture.source, nil, "--restore-backups", "--env", "claude_code")
	if code != exitOK {
		t.Fatalf("env unmanage --restore-backups = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if got := readFileForUnmanageTest(t, fixture.surfacePath); string(got) != "operator-owned context\n" {
		t.Fatalf("restored surface = %q, want the original takeover bytes", got)
	}
	if _, err := os.Lstat(fixture.markerPath); !os.IsNotExist(err) {
		t.Fatalf("unmanage did not remove the marker: %s (%v)", fixture.markerPath, err)
	}
	if got := readFileForUnmanageTest(t, fixture.backupFile); string(got) != "operator-owned context\n" {
		t.Fatalf("restore modified its source generation: %q", got)
	}
}

func TestEnvUnmanageLeavesBackupsWithoutRestoreFlag(t *testing.T) {
	fixture := newUnmanageFixture(t, true)
	code, stdout, stderr := runEnvUnmanageWithReadDir(t, fixture.source, nil)
	if code != exitOK {
		t.Fatalf("env unmanage = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if _, err := os.Lstat(fixture.surfacePath); !os.IsNotExist(err) {
		t.Fatalf("managed surface remains after unmanage: %s (%v)", fixture.surfacePath, err)
	}
	if got := readFileForUnmanageTest(t, fixture.backupFile); string(got) != "operator-owned context\n" {
		t.Fatalf("unmanage without restore changed the retained backup: %q", got)
	}
	if !strings.Contains(stderr, fixture.backupRoot) {
		t.Fatalf("unmanage did not tell the operator where backups remain:\n%s", stderr)
	}
}
