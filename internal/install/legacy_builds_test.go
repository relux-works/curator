package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/marker"
)

// writeLegacyBuildSkill writes one schema-9 skill declaring a local go-v1
// build command ("ltool") over a build root: core §4.4 permits build
// commands from schema-9 providers, so the legacy lane must record them,
// not refuse them.
func writeLegacySchema9BuildSkill(t *testing.T, dir, name string) {
	t.Helper()
	for _, rel := range []string{"src/cmd/ltool", "references"} {
		if err := os.MkdirAll(filepath.Join(dir, filepath.FromSlash(rel)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]string{
		"SKILL.md":              "---\nname: " + name + "\ndescription: Test\n---\n# " + name + "\n",
		"src/go.mod":            "module example.com/" + name + "\n",
		"src/cmd/ltool/main.go": "package main\n\nfunc main() {}\n",
		"references/info.md":    "context",
	}
	for rel, content := range files {
		if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(rel)), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	commands := map[string]any{"ltool": map[string]any{"type": "build", "driver": "go-v1", "source_dir": "src/cmd/ltool"}}
	spec := map[string]any{"schema_version": 9, "build_roots": []string{"src"}, "capabilities": map[string]any{}, "commands": commands}
	payload, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "agent-skill.json"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
}

// setupLegacyBuildRepo lays out one schema-9 build skill as a tagged git
// repository under the skills root.
func setupLegacyBuildRepo(e *env, name string) {
	e.t.Helper()
	dir := filepath.Join(e.skillsRoot, name)
	writeLegacySchema9BuildSkill(e.t, dir, name)
	e.git(dir, "init", "-q", "-b", "main")
	e.git(dir, "add", ".")
	e.git(dir, "commit", "-qm", "schema9 build")
	e.git(dir, "tag", "v1")
}

// TestLegacyLaneBuildsPublishReceipt3 is the F3 acceptance at the real
// install entry: a schema-9 skill with a compiled command installs on the
// schema-1 lane (no schema-2 refusal), records a v6 package marker whose
// build entry binds receipt version 3 with the execution policy, and
// publishes the entry under the distinct receipt-3 namespace with a
// receipt whose input.package equals the marker package. A reinstall
// takes exact receipt-3 cache hits and rebuilds nothing.
func TestLegacyLaneBuildsPublishReceipt3(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	setupLegacyBuildRepo(e, "tooling")
	e.declare("tooling")
	deps, _, builder := draftBuildDeps(t)

	result := e.install(Options{Build: deps})
	if result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}

	recorded := marker.Read(filepath.Join(e.project, ".agents", "skills", "tooling"))
	if recorded == nil || recorded.SchemaVersion != marker.SchemaV6 || recorded.Package == nil {
		t.Fatalf("marker = %+v, want a v6 package marker", recorded)
	}
	local, ok := recorded.Builds["ltool"]
	if !ok {
		t.Fatalf("marker builds = %+v, want ltool", recorded.Builds)
	}
	if local.Driver != buildmeta.DriverGoV1 || local.ReceiptSchemaVersion != 3 || local.ExecutionPolicy != buildmeta.ExecutionPolicy {
		t.Fatalf("marker record = %+v, want receipt 3 with the execution policy", local)
	}

	// The entry lives in the receipt-3 namespace only: a legacy receipt
	// or key here would prove the package wrapper was never applied.
	legacyEntries := dirNames(t, filepath.Join(e.home, "cache", "build", buildcache.Namespace))
	sourceAwareEntries := dirNames(t, filepath.Join(e.home, "cache", "build", buildcache.SourceAwareNamespace))
	if len(legacyEntries) != 0 || len(sourceAwareEntries) != 1 || "sha256:"+sourceAwareEntries[0] != string(local.CacheKey) {
		t.Fatalf("namespaces: legacy=%v receipt-3=%v key=%s", legacyEntries, sourceAwareEntries, local.CacheKey)
	}
	receipt := readReceiptObject(t, filepath.Join(e.home, "cache", "build", buildcache.SourceAwareNamespace, sourceAwareEntries[0], buildcache.ReceiptFilename))
	input := receipt["input"].(map[string]any)
	if receipt["schema_version"] != json.Number("3") || input["schema_version"] != json.Number("3") || len(input) != 3 {
		t.Fatalf("receipt shape = %v", receipt)
	}
	wantPackage, err := recorded.Package.Digest()
	if err != nil {
		t.Fatal(err)
	}
	receiptPackageDigest := receiptPackageDigestOf(t, input["package"])
	if receiptPackageDigest != wantPackage {
		t.Fatalf("receipt package %s != marker package %s", receiptPackageDigest, wantPackage)
	}
	if build := input["build"].(map[string]any); build["schema_version"] != json.Number("1") || build["driver"] != "go-v1" || build["command"] != "ltool" || build["build_root"] != "src" {
		t.Fatalf("wrapped build = %v", build)
	}
	if strings.Join(builder.calls, ",") != "ltool" {
		t.Fatalf("builder calls = %v, want one ltool compilation", builder.calls)
	}

	again := e.install(Options{Build: deps})
	if again.Status != "ok" {
		t.Fatalf("reinstall = %+v", again)
	}
	messages := strings.Join(again.Messages, "\n")
	if !strings.Contains(messages, "tooling.ltool build source=") || !strings.Contains(messages, "key="+string(local.CacheKey)+" outcome=cache-hit") {
		t.Fatalf("reinstall did not take an exact receipt-3 cache hit: %q", again.Messages)
	}
	if len(builder.calls) != 1 {
		t.Fatalf("reinstall rebuilt: %v", builder.calls)
	}
	if reread := marker.Read(filepath.Join(e.project, ".agents", "skills", "tooling")); reread == nil || !reflect.DeepEqual(reread.Builds, recorded.Builds) {
		t.Fatalf("reinstall changed the marker build records: %+v", reread)
	}
}

// receiptPackageDigestOf hashes one decoded receipt input.package object
// the way the package digest does, so the test compares identities
// without reimplementing the arm shapes.
func receiptPackageDigestOf(t *testing.T, value any) string {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Kind       string `json:"kind"`
		Snapshot   string `json:"snapshot"`
		Repository string `json:"repository"`
		Source     string `json:"source"`
		Commit     *struct {
			ObjectFormat string `json:"object_format"`
			Hex          string `json:"hex"`
		} `json:"commit"`
		Directory string `json:"directory"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	pkg := &marker.Package{
		Kind: decoded.Kind, Snapshot: decoded.Snapshot,
		Repository: decoded.Repository, Source: decoded.Source, Directory: decoded.Directory,
	}
	if decoded.Commit != nil {
		pkg.Commit = &marker.Commit{ObjectFormat: decoded.Commit.ObjectFormat, Hex: decoded.Commit.Hex}
	}
	digest, err := pkg.Digest()
	if err != nil {
		t.Fatal(err)
	}
	return digest
}

// TestLegacyLaneBuildMarkerRebindsTamperedReceipt is the F3 negative at
// the real install entry: a marker whose recorded cache key no longer
// binds the protected receipt-3 entry is not adopted as trust — the
// reinstall re-derives the binding from the cache and restores the true
// receipt-3 record.
func TestLegacyLaneBuildMarkerRebindsTamperedReceipt(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	setupLegacyBuildRepo(e, "tooling")
	e.declare("tooling")
	deps, _, _ := draftBuildDeps(t)

	result := e.install(Options{Build: deps})
	if result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	installed := filepath.Join(e.project, ".agents", "skills", "tooling")
	recorded := marker.Read(installed)
	if recorded == nil {
		t.Fatal("marker missing after install")
	}
	wantBuilds := recorded.Builds

	markerPath := filepath.Join(installed, marker.Name)
	payload, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	builds := raw["builds"].(map[string]any)
	entry := builds["ltool"].(map[string]any)
	entry["cache_key"] = "sha256:" + strings.Repeat("ff", 32)
	tampered, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(markerPath, append(tampered, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}

	again := e.install(Options{Build: deps})
	if again.Status != "ok" {
		t.Fatalf("reinstall = %+v", again)
	}
	restored := marker.Read(installed)
	if restored == nil || !reflect.DeepEqual(restored.Builds, wantBuilds) {
		t.Fatalf("reinstall left %+v, want the true receipt-3 records %+v", restored, wantBuilds)
	}
	restoredDigest, err := restored.Package.Digest()
	if err != nil {
		t.Fatal(err)
	}
	recordedDigest, err := recorded.Package.Digest()
	if err != nil {
		t.Fatal(err)
	}
	if restoredDigest != recordedDigest {
		t.Fatalf("reinstall rebound the package identity: %+v", restored.Package)
	}
}
