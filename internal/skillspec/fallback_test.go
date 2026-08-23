package skillspec

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// legacyFallbackPayload is a valid agents/runtime.json planted next to a
// broken skill manifest: if Load ever returns a spec, the fallback was
// reached, which Spec §4 forbids once a manifest is present.
const legacyFallbackPayload = `{"commands": {"fallback": "scripts/fallback"}}`

// plantRuntimeFallback writes the parseable legacy runtime manifest into dir.
func plantRuntimeFallback(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(RuntimeFallbackName)), []byte(legacyFallbackPayload), 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeManifestWithRuntimeFallback lays out a snapshot carrying the given
// manifest content under name plus a parseable runtime fallback.
func writeManifestWithRuntimeFallback(t *testing.T, name, manifest string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	plantRuntimeFallback(t, dir)
	return dir
}

// mustNotFallBack asserts that Load failed without producing a spec: any spec
// at all here would mean the runtime fallback (or the empty pure-context spec)
// was reached behind a manifest that is present.
func mustNotFallBack(t *testing.T, dir string) error {
	t.Helper()
	spec, err := Load(dir)
	if err == nil {
		t.Fatalf("present manifest degraded into a parsed spec: %+v", spec)
	}
	if spec != nil {
		t.Fatalf("failed load must not return a spec: %+v", spec)
	}
	return err
}

// TestBrokenManifestNeverFallsBackToRuntime pins the fail-loud rule for both
// modern manifest filenames: a present manifest is authoritative, so any
// failure to parse it is terminal and agents/runtime.json is never consulted.
func TestBrokenManifestNeverFallsBackToRuntime(t *testing.T) {
	manifests := []struct {
		name     string
		manifest string
	}{
		{"newer schema version", `{"schema_version": 99}`},
		{"malformed json", `{"schema_version": 5,`},
		{"not a json object", `[{"schema_version": 5}]`},
		{"empty file", ``},
		{"valid schema, invalid body", `{"schema_version": 3}`},
	}
	for _, file := range []string{CanonicalManifestName, LegacyManifestName} {
		for _, testCase := range manifests {
			t.Run(file+"/"+testCase.name, func(t *testing.T) {
				dir := writeManifestWithRuntimeFallback(t, file, testCase.manifest)
				_ = mustNotFallBack(t, dir)
			})
		}
	}
}

// TestNewerSchemaVersionReportsUpgradeHint pins the exact fail-loud signal for
// the forward-compatibility case: a schema this build does not know errors on
// schema_version and tells the operator to upgrade, rather than degrading to
// the runtime fallback that sits right next to it.
func TestNewerSchemaVersionReportsUpgradeHint(t *testing.T) {
	dir := writeManifestWithRuntimeFallback(t, CanonicalManifestName,
		`{"schema_version": 99, "commands": {"fallback": {"type": "system", "command": "sh"}}}`)
	err := mustNotFallBack(t, dir)
	if !strings.Contains(err.Error(), UpgradeHint) {
		t.Fatalf("error %q does not carry the upgrade hint %q", err, UpgradeHint)
	}
	if !strings.Contains(err.Error(), "99") {
		t.Fatalf("error %q does not name the unsupported version", err)
	}
}

// TestRuntimeFallbackNeedsAbsentManifest is the positive control for the tests
// above: the same fallback payload loads fine once no modern manifest is
// present, so the failures above come from precedence, not a broken fallback.
func TestRuntimeFallbackNeedsAbsentManifest(t *testing.T) {
	dir := writeManifestWithRuntimeFallback(t, CanonicalManifestName, `{"schema_version": 99}`)
	if err := os.Remove(filepath.Join(dir, CanonicalManifestName)); err != nil {
		t.Fatal(err)
	}
	spec, err := Load(dir)
	if err != nil {
		t.Fatalf("runtime fallback must load when no manifest exists: %v", err)
	}
	if spec.SourceFile != RuntimeFallbackName {
		t.Fatalf("source file: %q", spec.SourceFile)
	}
	if spec.Commands["fallback"].UnixPath != "scripts/fallback" {
		t.Fatalf("fallback command: %+v", spec.Commands)
	}
}

// TestUnreadableManifestNeverFallsBackToRuntime covers the present-but-
// unreadable manifest: the file is there, so the fallback stays out of reach
// and the permission failure surfaces.
func TestUnreadableManifestNeverFallsBackToRuntime(t *testing.T) {
	// A manifest body that parses cleanly when readable, so the only thing
	// the assertion can be reacting to is the read failure itself.
	dir := writeManifestWithRuntimeFallback(t, CanonicalManifestName, `{"schema_version": 1}`)
	manifest := filepath.Join(dir, CanonicalManifestName)
	if err := os.Chmod(manifest, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(manifest, 0o644) })
	if _, err := os.ReadFile(manifest); err == nil {
		t.Skip("this host cannot create an unreadable file: mode 000 stays readable here")
	}
	err := mustNotFallBack(t, dir)
	if !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("error does not carry the read failure: %v", err)
	}
}

// TestManifestDirectoryNeverFallsBackToRuntime covers a manifest entry that
// exists but can never be read as a manifest.
func TestManifestDirectoryNeverFallsBackToRuntime(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, CanonicalManifestName), 0o755); err != nil {
		t.Fatal(err)
	}
	plantRuntimeFallback(t, dir)
	err := mustNotFallBack(t, dir)
	if !strings.Contains(err.Error(), CanonicalManifestName) {
		t.Fatalf("error does not name the manifest that failed: %v", err)
	}
}

// TestDanglingManifestSymlinkNeverFallsBackToRuntime pins presence against
// Lstat, not Stat: a manifest symlink whose target is gone is a broken
// manifest, not an absent one, so it must not open the fallback. Before the
// Lstat fix this load returned the runtime fallback spec with err == nil.
func TestDanglingManifestSymlinkNeverFallsBackToRuntime(t *testing.T) {
	for _, file := range []string{CanonicalManifestName, LegacyManifestName} {
		t.Run(file, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.Symlink(filepath.Join(dir, "absent-target.json"), filepath.Join(dir, file)); err != nil {
				t.Skipf("symlinks unavailable: %v", err)
			}
			plantRuntimeFallback(t, dir)
			err := mustNotFallBack(t, dir)
			if !strings.Contains(err.Error(), file) {
				t.Fatalf("error does not name the manifest that failed: %v", err)
			}
		})
	}
}

// TestUnreadableSnapshotNeverDegradesToEmptySpec pins the same rule one level
// up: a snapshot whose contents cannot be listed is an error, never the empty
// pure-context spec, which would silently strip every command a skill has.
func TestUnreadableSnapshotNeverDegradesToEmptySpec(t *testing.T) {
	dir := writeManifestWithRuntimeFallback(t, CanonicalManifestName, `{"schema_version": 1}`)
	if err := os.Chmod(dir, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	if _, err := os.Lstat(filepath.Join(dir, CanonicalManifestName)); err == nil {
		t.Skip("this environment can inspect a mode-000 directory")
	}
	err := mustNotFallBack(t, dir)
	if !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("error does not carry the inspection failure: %v", err)
	}
}

// TestDanglingRuntimeFallbackNeverDegradesToEmptySpec applies the presence
// rule to the runtime fallback itself: a broken agents/runtime.json is an
// error, not a pure context skill.
func TestDanglingRuntimeFallbackNeverDegradesToEmptySpec(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "agents", "absent-target.json")
	if err := os.Symlink(target, filepath.Join(dir, filepath.FromSlash(RuntimeFallbackName))); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	err := mustNotFallBack(t, dir)
	if !strings.Contains(err.Error(), filepath.FromSlash(RuntimeFallbackName)) {
		t.Fatalf("error does not name the manifest that failed: %v", err)
	}
}

// TestManifestSourcePathNamesADanglingManifest keeps the diagnostic surface in
// step with Load: the file Load will fail on is the file the diagnostic names.
func TestManifestSourcePathNamesADanglingManifest(t *testing.T) {
	dir := t.TempDir()
	if err := os.Symlink(filepath.Join(dir, "absent-target.json"), filepath.Join(dir, CanonicalManifestName)); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	plantRuntimeFallback(t, dir)
	if got := ManifestSourcePath(dir); got != CanonicalManifestName {
		t.Fatalf("ManifestSourcePath = %q, want %q", got, CanonicalManifestName)
	}
}
