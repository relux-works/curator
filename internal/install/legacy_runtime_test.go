package install

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/scopes"
)

// setupLegacyScriptRepo lays out one runnable schema-9 script skill as a
// tagged git repository under the skills root: the legacy-lane fixture for
// a schema-9 installation selected at the repository root.
func setupLegacyScriptRepo(e *env, name, command, scriptBody string) string {
	e.t.Helper()
	dir := filepath.Join(e.skillsRoot, name)
	writeDraftScriptSkill(e.t, dir, name, command, scriptBody, func(spec map[string]any) {
		spec["schema_version"] = 9
	})
	e.git(dir, "init", "-q", "-b", "main")
	e.git(dir, "add", ".")
	e.git(dir, "commit", "-qm", "schema9 script")
	e.git(dir, "tag", "v1")
	return e.gitOutput(dir, "rev-parse", "HEAD")
}

// legacyPackageDigest hashes the recorded marker package exactly the way
// GC marks it, and returns the source-v1 runtime key it requires.
func legacyPackageDigest(t *testing.T, recorded *marker.Marker) (digest, key string) {
	t.Helper()
	if recorded.Package == nil {
		t.Fatalf("marker for %s records no package", recorded.Name)
	}
	var err error
	digest, err = recorded.Package.Digest()
	if err != nil {
		t.Fatal(err)
	}
	key, err = runtimestore.SourceV1Key(digest)
	if err != nil {
		t.Fatal(err)
	}
	return digest, key
}

// launchInstalledShim runs one installed manager shim and returns its
// combined output. Windows shims are shell wrappers the test host cannot
// execute portably, so there the shim bytes must name the published
// runtime executable instead.
func launchInstalledShim(t *testing.T, shim, wantExecutable string) string {
	t.Helper()
	shimBytes, err := os.ReadFile(shim)
	if err != nil {
		t.Fatalf("installed shim missing: %v", err)
	}
	if installPlatform() != "unix" {
		if !strings.Contains(string(shimBytes), wantExecutable) {
			t.Fatalf("shim does not name the published runtime executable %q", wantExecutable)
		}
		return ""
	}
	command := exec.Command(shim) // #nosec G204 -- fixture launches the manager-derived test shim
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("shim launch failed: %v\n%s", err, output)
	}
	return string(output)
}

// TestLegacyLaneRuntimeMatchesPackageMarker is the permanent production
// entry for F1: a schema-9 script skill installed through the real
// schema-1 project entry records a v6 package marker, and its runnable
// script is published under the source-v1 key that marker requires — the
// same leaf GC marks — instead of a commit-keyed tree the marker leaves
// unreferenced. The installed command launches, a second install is
// up-to-date, and a maintenance sweep keeps the live tree.
func TestLegacyLaneRuntimeMatchesPackageMarker(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	commit := setupLegacyScriptRepo(e, "consumer", "ctool", "#!/bin/sh\necho consumer-ok\n")
	e.declare("consumer")

	result := e.install(Options{})
	if result.Status != "ok" {
		t.Fatalf("install: %+v", result)
	}

	installed := filepath.Join(e.project, ".agents", "skills", "consumer")
	recorded, _ := readLegacyMarker(t, installed)
	if recorded.SchemaVersion != marker.SchemaV6 || recorded.Package == nil {
		t.Fatalf("marker = %+v, want a v6 package marker", recorded)
	}
	digest, _ := legacyPackageDigest(t, recorded)
	runtimeDir, err := runtimestore.SourceV1Dir(e.home, "consumer", digest)
	if err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(runtimeDir, "scripts", "ctool.sh")
	if _, err := os.Stat(script); err != nil {
		t.Fatalf("package marker runtime missing under source-v1 namespace: %v", err)
	}
	// No commit-keyed twin may shadow the package tree: the migrated
	// runtime publishes under the package key only.
	if _, err := os.Stat(filepath.Join(e.home, "runtime", "consumer", commit)); !os.IsNotExist(err) {
		t.Fatalf("migrated runtime also published a commit-keyed tree: %v", err)
	}

	shim := filepath.Join(e.project, ".agents", "bin", shimName("ctool"))
	if output := launchInstalledShim(t, shim, script); installPlatform() == "unix" && !strings.Contains(output, "consumer-ok") {
		t.Fatalf("shim output = %q, want consumer-ok", output)
	}

	second := e.install(Options{})
	if second.Status != "ok" {
		t.Fatalf("second: %+v", second)
	}
	if joined := strings.Join(second.Messages, "\n"); !strings.Contains(joined, "up-to-date") {
		t.Fatalf("second install must be up-to-date:\n%s", joined)
	}

	removed, err := scopes.CollectRuntime(e.home)
	if err != nil {
		t.Fatalf("gc: %v", err)
	}
	if len(removed) != 0 {
		t.Fatalf("gc swept live entries: %v", removed)
	}
	if _, err := os.Stat(script); err != nil {
		t.Fatalf("gc swept the live package runtime: %v", err)
	}
}

// TestGlobalLaneRuntimeMatchesPackageMarker carries F1 to the machine-wide
// scope: a schema-9 script skill installed through the real global entry
// records a v6 package marker and publishes its runtime under the
// source-v1 key the marker requires.
func TestGlobalLaneRuntimeMatchesPackageMarker(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	setupLegacyScriptRepo(e, "consumer", "ctool", "#!/bin/sh\necho consumer-ok\n")
	e.globalDeclare("consumer")

	userHome := t.TempDir()
	result := Global(e.cfg, userHome, Options{Platform: installPlatform()})
	if result.Status != "ok" {
		t.Fatalf("global install: %+v", result)
	}

	installed := filepath.Join(GlobalRoot(e.home), "skills", "consumer")
	recorded, _ := readLegacyMarker(t, installed)
	if recorded.SchemaVersion != marker.SchemaV6 || recorded.Package == nil {
		t.Fatalf("marker = %+v, want a v6 package marker", recorded)
	}
	digest, _ := legacyPackageDigest(t, recorded)
	runtimeDir, err := runtimestore.SourceV1Dir(e.home, "consumer", digest)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(runtimeDir, "scripts", "ctool.sh")); err != nil {
		t.Fatalf("package marker runtime missing under source-v1 namespace: %v", err)
	}

	second := Global(e.cfg, userHome, Options{Platform: installPlatform()})
	if second.Status != "ok" {
		t.Fatalf("second: %+v", second)
	}
	if joined := strings.Join(second.Messages, "\n"); !strings.Contains(joined, "up-to-date") {
		t.Fatalf("second global install must be up-to-date:\n%s", joined)
	}
}
