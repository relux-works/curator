package install

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/sourcelock"
)

const draftBuildPayload = `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["*"]}]}`

const externalLockedCommit = "0123456789abcdef0123456789abcdef01234567"

// writeDraftBuildSkill writes one local draft package declaring a local
// go-v1 build command ("ltool") and, for schema 7, an external
// go-repository-v1 command ("etool") over a declared build repository: both
// build arms of one marker-5 member.
func writeDraftBuildSkill(t *testing.T, dir, name string, external bool) {
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
	spec := map[string]any{"schema_version": 6, "build_roots": []string{"src"}, "capabilities": map[string]any{}, "commands": commands}
	if external {
		spec["schema_version"] = 7
		spec["build_repositories"] = map[string]any{"tools": map[string]any{
			"git":           "https://git.example.com/skills/tools.git",
			"locked_commit": map[string]any{"object_format": "sha1", "hex": externalLockedCommit},
		}}
		commands["etool"] = map[string]any{"type": "build", "driver": "go-repository-v1", "repository": "tools", "target": "tool"}
	}
	payload, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "agent-skill.json"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
}

// externalTestSnapshot frames one proved external repository snapshot the way
// the raw-object admission does (curator-build-source-v1 framing), so the
// fake acquisition hands the pipeline bytes it can materialize and revalidate.
func externalTestSnapshot() *buildrepo.Snapshot {
	files := []buildrepo.File{
		{Path: "skill-build.json", Content: []byte(`{"schema_version":1,"targets":{"tool":{"driver":"go-repository-v1","build_root":"tools","source_dir":"tools/cmd/tool"}}}`)},
		{Path: "tools/cmd/tool/main.go", Content: []byte("package main\n\nfunc main() {}\n")},
		{Path: "tools/go.mod", Content: []byte("module example.test/tool\n")},
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	canonical := []byte("curator-build-source-v1\x00")
	for _, file := range files {
		canonical = append(canonical, 'F')
		canonical = binary.BigEndian.AppendUint64(canonical, uint64(len(file.Path)))
		canonical = append(canonical, file.Path...)
		canonical = binary.BigEndian.AppendUint64(canonical, uint64(len(file.Content)))
		canonical = append(canonical, file.Content...)
	}
	sum := sha256.Sum256(canonical)
	return &buildrepo.Snapshot{ObjectFormat: "sha1", Commit: externalLockedCommit, Files: files, CanonicalBytes: canonical, Digest: "sha256:" + hex.EncodeToString(sum[:])}
}

func draftBuildDeps(t *testing.T) (BuildDeps, *fakeToolchain, *fakeBuilder) {
	t.Helper()
	toolchain := &fakeToolchain{t: t, target: testTarget(), toolchain: testToolchain()}
	builder := &fakeBuilder{t: t, failOn: map[string]error{}}
	// Cache is left nil so the real protected build cache under the manager
	// home is exercised: the receipt-3 namespace is a property of that store.
	deps := BuildDeps{Toolchain: toolchain, Builder: builder, Clock: fixedClock{at: time.Unix(1_700_000_000, 0).UTC()}, Generation: &countingGeneration{}}
	return deps, toolchain, builder
}

func draftExternalDeps(t *testing.T) ExternalDeps {
	t.Helper()
	return ExternalDeps{
		Acquire: func(_ context.Context, _ ExternalSource) (*buildrepo.Snapshot, error) {
			return externalTestSnapshot(), nil
		},
		Audit: func(context.Context, buildrepo.AuditSubject) error { return nil },
	}
}

func draftBuildInstall(t *testing.T, home, project string, deps BuildDeps, external ExternalDeps) Result {
	t.Helper()
	cfg := draftTestConfig(home, t.TempDir())
	return Project(cfg, project, "test", Options{Platform: installPlatform(), Build: deps, External: external})
}

func readReceiptObject(t *testing.T, path string) map[string]any {
	t.Helper()
	payload, err := os.ReadFile(path) // #nosec G304 -- test reads a manager-home receipt
	if err != nil {
		t.Fatalf("receipt missing: %v", err)
	}
	if err := protocoljson.Validate(payload); err != nil {
		t.Fatalf("receipt is not canonical: %v", err)
	}
	var object map[string]any
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&object); err != nil {
		t.Fatal(err)
	}
	return object
}

func canonicalOf(t *testing.T, value any) []byte {
	t.Helper()
	payload, err := protocoljson.MarshalCanonical(value)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

func dirNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

// TestDraftBuildsPublishReceipt3OnBothArms is the acceptance core at
// install.Project: a marker-5 member with a local go-v1 command and an
// external go-repository-v1 command publishes both under the distinct
// receipt-3 namespaces with receipts whose input.package equals the marker
// package, whose build keeps every legacy field, and whose marker records
// bind receipt version 3 on both arms; a reinstall is current.
func TestDraftBuildsPublishReceipt3OnBothArms(t *testing.T) {
	project, home, _ := draftProject(t, draftBuildPayload, nil)
	writeDraftBuildSkill(t, filepath.Join(project, "skills", "tooling"), "tooling", true)
	resolveDraftForInstall(t, project, home, draftBuildPayload)
	deps, _, builder := draftBuildDeps(t)

	result := draftBuildInstall(t, home, project, deps, draftExternalDeps(t))
	if result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	lock, err := sourcelock.Read(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find("tooling")
	if !ok {
		t.Fatal("lock misses tooling")
	}
	wantPackage, err := member.Package.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	recorded := marker.Read(filepath.Join(project, ".agents", "skills", "tooling"))
	if recorded == nil || recorded.SchemaVersion != marker.SchemaV5 || recorded.Package == nil {
		t.Fatalf("marker = %+v", recorded)
	}
	if len(recorded.Builds) != 2 {
		t.Fatalf("marker builds = %+v", recorded.Builds)
	}
	local, external := recorded.Builds["ltool"], recorded.Builds["etool"]
	if local.Driver != buildmeta.DriverGoV1 || local.ReceiptSchemaVersion != 3 || local.ExecutionPolicy != buildmeta.ExecutionPolicy {
		t.Fatalf("local marker record = %+v", local)
	}
	if external.Driver != "go-repository-v1" || external.ReceiptSchemaVersion != 3 || external.Repository != "tools" ||
		external.DeclaredLockedCommit == nil || external.DeclaredLockedCommit.Hex != externalLockedCommit || external.DeclaredIdentity == nil ||
		external.EffectiveIdentity == nil || external.Commit != externalLockedCommit || external.BuildSource == nil || external.DescriptorTarget != "tool" || external.Substituted {
		t.Fatalf("external marker record lost retained fields: %+v", external)
	}

	// Local arm: the entry lives in the receipt-3 namespace only.
	legacyLocal := dirNames(t, filepath.Join(home, "cache", "build", buildcache.Namespace))
	sourceAwareLocal := dirNames(t, filepath.Join(home, "cache", "build", buildcache.SourceAwareNamespace))
	if len(legacyLocal) != 0 || len(sourceAwareLocal) != 1 || "sha256:"+sourceAwareLocal[0] != string(local.CacheKey) {
		t.Fatalf("local namespaces: legacy=%v receipt-3=%v key=%s", legacyLocal, sourceAwareLocal, local.CacheKey)
	}
	localReceipt := readReceiptObject(t, filepath.Join(home, "cache", "build", buildcache.SourceAwareNamespace, sourceAwareLocal[0], buildcache.ReceiptFilename))
	localInput := localReceipt["input"].(map[string]any)
	if localReceipt["schema_version"] != json.Number("3") || localInput["schema_version"] != json.Number("3") || len(localInput) != 3 {
		t.Fatalf("local receipt shape = %v", localReceipt)
	}
	if !bytes.Equal(canonicalOf(t, localInput["package"]), wantPackage) {
		t.Fatalf("local receipt package %s != lock package %s", canonicalOf(t, localInput["package"]), wantPackage)
	}
	if build := localInput["build"].(map[string]any); build["schema_version"] != json.Number("1") || build["driver"] != "go-v1" || build["command"] != "ltool" || build["build_root"] != "src" {
		t.Fatalf("local wrapped build = %v", build)
	}
	if decoded, err := buildmeta.DecodeReceipt(canonicalOf(t, localReceipt)); err != nil || decoded.CacheKey == "" {
		t.Fatalf("local receipt does not decode as receipt-3: %v", err)
	}

	// External arm: artifacts-receipt-3 only, receipt-3 wrapper around the
	// unchanged receipt-2 input with every declared/effective field.
	externalRoot := filepath.Join(home, "external-build-cache")
	legacyExternal := dirNames(t, filepath.Join(externalRoot, buildrepo.ArtifactsDir(buildrepo.LegacyReceiptSchemaVersion)))
	sourceAwareExternal := dirNames(t, filepath.Join(externalRoot, buildrepo.ArtifactsDir(buildrepo.SourceAwareReceiptSchemaVersion)))
	if len(legacyExternal) != 0 || len(sourceAwareExternal) != 1 || "sha256:"+sourceAwareExternal[0] != string(external.CacheKey) {
		t.Fatalf("external namespaces: legacy=%v receipt-3=%v key=%s", legacyExternal, sourceAwareExternal, external.CacheKey)
	}
	externalReceipt := readReceiptObject(t, filepath.Join(externalRoot, buildrepo.ArtifactsDir(buildrepo.SourceAwareReceiptSchemaVersion), sourceAwareExternal[0], "receipt.json"))
	externalInput := externalReceipt["input"].(map[string]any)
	if externalReceipt["schema_version"] != json.Number("3") || externalInput["schema_version"] != json.Number("3") || len(externalInput) != 3 {
		t.Fatalf("external receipt shape = %v", externalReceipt)
	}
	if !bytes.Equal(canonicalOf(t, externalInput["package"]), wantPackage) {
		t.Fatalf("external receipt package %s != lock package %s", canonicalOf(t, externalInput["package"]), wantPackage)
	}
	externalBuild := externalInput["build"].(map[string]any)
	externalSource := externalBuild["source"].(map[string]any)
	declared := externalSource["declared"].(map[string]any)
	effective := externalSource["effective"].(map[string]any)
	if externalBuild["schema_version"] != json.Number("2") || externalBuild["driver"] != "go-repository-v1" || externalSource["repository"] != "tools" ||
		declared["transport"] != "https" || declared["locked_commit"].(map[string]any)["hex"] != externalLockedCommit ||
		effective["commit"] != externalLockedCommit || effective["substituted"] != false || externalBuild["assurance"] == nil {
		t.Fatalf("external wrapped build lost fields: %v", externalBuild)
	}
	sum := sha256.Sum256(canonicalOf(t, externalInput))
	if string(external.CacheKey) != "sha256:"+hex.EncodeToString(sum[:]) {
		t.Fatalf("external key %s is not SHA-256 of the wrapped input", external.CacheKey)
	}
	receiptPayload, _ := os.ReadFile(filepath.Join(externalRoot, buildrepo.ArtifactsDir(buildrepo.SourceAwareReceiptSchemaVersion), sourceAwareExternal[0], "receipt.json"))
	receiptSum := sha256.Sum256(receiptPayload)
	if string(external.ReceiptSHA256) != "sha256:"+hex.EncodeToString(receiptSum[:]) {
		t.Fatalf("external marker receipt hash does not bind the protected receipt")
	}
	if strings.Join(builder.calls, ",") != "ltool,etool" && strings.Join(builder.calls, ",") != "etool,ltool" {
		t.Fatalf("builder calls = %v", builder.calls)
	}

	// A reinstall reuses both receipt-3 entries: the plan takes exact cache
	// hits for both arms and nothing is compiled again (build-bearing nodes
	// are re-staged from the cache on every lane, so the context line still
	// reads "installed").
	again := draftBuildInstall(t, home, project, deps, draftExternalDeps(t))
	if again.Status != "ok" {
		t.Fatalf("reinstall = %+v", again)
	}
	messages := strings.Join(again.Messages, "\n")
	if !strings.Contains(messages, "tooling.ltool build source=") || !strings.Contains(messages, "key="+string(local.CacheKey)+" outcome=cache-hit") ||
		!strings.Contains(messages, "tooling.etool external build key="+string(external.CacheKey)+" outcome=cache-hit") {
		t.Fatalf("reinstall did not take exact receipt-3 cache hits: %q", again.Messages)
	}
	if len(builder.calls) != 2 {
		t.Fatalf("reinstall rebuilt: %v", builder.calls)
	}
	if reread := marker.Read(filepath.Join(project, ".agents", "skills", "tooling")); reread == nil || !reflect.DeepEqual(reread.Builds, recorded.Builds) {
		t.Fatalf("reinstall changed the marker build records: %+v", reread)
	}
}

// TestDraftBuildKeyFollowsThePackageIdentity: a runtime-only edit followed by
// an explicit refresh changes the frozen package, so both arms are keyed
// anew — the previous entries can never answer for the new identity — and
// the marker rebinds to the new receipts.
func TestDraftBuildKeyFollowsThePackageIdentity(t *testing.T) {
	project, home, _ := draftProject(t, draftBuildPayload, nil)
	writeDraftBuildSkill(t, filepath.Join(project, "skills", "tooling"), "tooling", true)
	resolveDraftForInstall(t, project, home, draftBuildPayload)
	deps, _, builder := draftBuildDeps(t)
	first := draftBuildInstall(t, home, project, deps, draftExternalDeps(t))
	if first.Status != "ok" {
		t.Fatalf("install = %+v", first)
	}
	before := marker.Read(filepath.Join(project, ".agents", "skills", "tooling"))

	// A context-only edit: the build roots are untouched, so the build source
	// identity is unchanged and only the package identity moves.
	if err := os.WriteFile(filepath.Join(project, "skills", "tooling", "references", "info.md"), []byte("context v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	pinned := draftBuildInstall(t, home, project, deps, draftExternalDeps(t))
	if pinned.Status != "ok" || len(builder.calls) != 2 {
		t.Fatalf("pinned reinstall rebuilt or failed: %+v calls=%v", pinned, builder.calls)
	}
	resolveDraftForInstall(t, project, home, draftBuildPayload)
	refreshed := draftBuildInstall(t, home, project, deps, draftExternalDeps(t))
	if refreshed.Status != "ok" {
		t.Fatalf("refreshed install = %+v", refreshed)
	}
	after := marker.Read(filepath.Join(project, ".agents", "skills", "tooling"))
	if after == nil || before == nil || after.Package.Snapshot == before.Package.Snapshot {
		t.Fatalf("package identity did not move: before=%+v after=%+v", before, after)
	}
	for _, command := range []string{"ltool", "etool"} {
		if after.Builds[command].CacheKey == before.Builds[command].CacheKey || after.Builds[command].ReceiptSHA256 == before.Builds[command].ReceiptSHA256 {
			t.Fatalf("%s kept its key across a package change: %+v", command, after.Builds[command])
		}
		if after.Builds[command].ReceiptSchemaVersion != 3 {
			t.Fatalf("%s lost receipt version 3: %+v", command, after.Builds[command])
		}
	}
	if len(builder.calls) != 4 {
		t.Fatalf("refresh did not rebuild both arms: %v", builder.calls)
	}
	if entries := dirNames(t, filepath.Join(home, "cache", "build", buildcache.SourceAwareNamespace)); len(entries) != 2 {
		t.Fatalf("local receipt-3 namespace = %v", entries)
	}
	// The external cache is collected immediately after commit for keys no
	// marker references (existing external GC behavior), so exactly the new
	// key survives there.
	if entries := dirNames(t, filepath.Join(home, "external-build-cache", buildrepo.ArtifactsDir(buildrepo.SourceAwareReceiptSchemaVersion))); len(entries) != 1 || "sha256:"+entries[0] != string(after.Builds["etool"].CacheKey) {
		t.Fatalf("external receipt-3 namespace = %v, want only %s", entries, after.Builds["etool"].CacheKey)
	}
}

// TestDraftBuildRefusesASessionThatBindsTheLegacyDigest: an execution receipt
// over the context-only digest cannot attest a receipt-3 entry. The install
// fails closed on both arms before any cache publication or marker write.
func TestDraftBuildRefusesASessionThatBindsTheLegacyDigest(t *testing.T) {
	for _, external := range []bool{false, true} {
		name := "local"
		if external {
			name = "external"
		}
		t.Run(name, func(t *testing.T) {
			project, home, _ := draftProject(t, draftBuildPayload, nil)
			writeDraftBuildSkill(t, filepath.Join(project, "skills", "tooling"), "tooling", external)
			resolveDraftForInstall(t, project, home, draftBuildPayload)
			deps, _, builder := draftBuildDeps(t)
			builder.dropPackage = true
			result := draftBuildInstall(t, home, project, deps, draftExternalDeps(t))
			if result.Status == "ok" {
				t.Fatalf("legacy-bound execution receipt accepted: %+v", result)
			}
			joined := strings.Join(append(result.Errors, result.Messages...), "\n")
			if !strings.Contains(joined, "receipt") {
				t.Fatalf("refusal does not name the receipt binding: %s", joined)
			}
			for _, dir := range []string{
				filepath.Join(home, "cache", "build", buildcache.Namespace),
				filepath.Join(home, "cache", "build", buildcache.SourceAwareNamespace),
				filepath.Join(home, "external-build-cache", buildrepo.ArtifactsDir(buildrepo.LegacyReceiptSchemaVersion)),
				filepath.Join(home, "external-build-cache", buildrepo.ArtifactsDir(buildrepo.SourceAwareReceiptSchemaVersion)),
			} {
				if entries := dirNames(t, dir); len(entries) != 0 {
					t.Fatalf("refused install published %v under %s", entries, dir)
				}
			}
			if marker.Read(filepath.Join(project, ".agents", "skills", "tooling")) != nil {
				t.Fatal("refused install wrote a marker")
			}
		})
	}
}

// TestDraftBuildToolchainFailureBehaviorIsPreserved: the trusted toolchain
// boundary fails exactly as on the legacy lane — before any cache lookup,
// compilation or persistent state — and the package wrapper adds no path
// around it.
func TestDraftBuildToolchainFailureBehaviorIsPreserved(t *testing.T) {
	project, home, _ := draftProject(t, draftBuildPayload, nil)
	writeDraftBuildSkill(t, filepath.Join(project, "skills", "tooling"), "tooling", false)
	resolveDraftForInstall(t, project, home, draftBuildPayload)
	deps, toolchain, builder := draftBuildDeps(t)
	toolchain.establishErr = errors.New("trusted toolchain unavailable")
	result := draftBuildInstall(t, home, project, deps, draftExternalDeps(t))
	if result.Status != "failed" || len(builder.calls) != 0 {
		t.Fatalf("result = %+v calls=%v", result, builder.calls)
	}
	if _, err := os.Lstat(filepath.Join(home, "cache")); !os.IsNotExist(err) {
		t.Fatalf("toolchain failure left cache state: %v", err)
	}
	if marker.Read(filepath.Join(project, ".agents", "skills", "tooling")) != nil {
		t.Fatal("toolchain failure wrote a marker")
	}
	// The row is reported with the identities it already had, like on v1.
	if len(result.Builds) != 1 || result.Builds[0].Driver() != buildmeta.DriverGoV1 || result.Builds[0].CacheKey() != "" {
		t.Fatalf("builds inventory = %+v", result.Builds)
	}
}

// TestLegacyBuildsKeepReceipt1 pins the schema-1 receipt contract: a Git
// skill publishes a schema-1 receipt under the legacy namespace and no
// receipt-3 namespace exists.
func TestLegacyBuildsKeepReceipt1(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	e.declare("build-skill")
	deps, _, builder := draftBuildDeps(t)
	result := e.install(Options{Build: deps})
	if result.Status != "ok" || len(builder.calls) != 1 {
		t.Fatalf("legacy install = %+v calls=%v", result, builder.calls)
	}
	legacy := dirNames(t, filepath.Join(e.home, "cache", "build", buildcache.Namespace))
	if len(legacy) != 1 {
		t.Fatalf("legacy namespace = %v", legacy)
	}
	if _, err := os.Lstat(filepath.Join(e.home, "cache", "build", buildcache.SourceAwareNamespace)); !os.IsNotExist(err) {
		t.Fatalf("legacy install created the receipt-3 namespace: %v", err)
	}
	receipt := readReceiptObject(t, filepath.Join(e.home, "cache", "build", buildcache.Namespace, legacy[0], buildcache.ReceiptFilename))
	input := receipt["input"].(map[string]any)
	if receipt["schema_version"] != json.Number("1") || input["schema_version"] != json.Number("1") || input["package"] != nil || len(input) != 9 {
		t.Fatalf("legacy receipt shape = %v", receipt)
	}
	recorded := marker.Read(filepath.Join(e.project, ".agents", "skills", "build-skill"))
	if recorded == nil || recorded.SchemaVersion == marker.SchemaV5 || recorded.Builds["alpha"].ReceiptSchemaVersion == 3 {
		t.Fatalf("legacy marker = %+v", recorded)
	}
}

// TestDraftBuildFinalRootParentsAreStoreCreated: the external arm publishes
// into final-root namespaces the protected store created privately before the
// transaction, so a lookup on the next install proves every parent (root,
// snapshots, receipt-3 artifacts) as its own instead of a parent the commit's
// generic scaffolding invented; on unix each parent is 0700. The negative row
// pre-creates a foreign world-writable artifact namespace: the install refuses
// with the protected-boundary code before publishing or writing a marker.
func TestDraftBuildFinalRootParentsAreStoreCreated(t *testing.T) {
	project, home, _ := draftProject(t, draftBuildPayload, nil)
	writeDraftBuildSkill(t, filepath.Join(project, "skills", "tooling"), "tooling", true)
	resolveDraftForInstall(t, project, home, draftBuildPayload)
	deps, _, builder := draftBuildDeps(t)
	if result := draftBuildInstall(t, home, project, deps, draftExternalDeps(t)); result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	finalRoot := filepath.Join(home, "external-build-cache")
	store := &buildrepo.DiskProtectedStore{Root: finalRoot}
	for _, name := range []string{"", "snapshots", buildrepo.ArtifactsDir(buildrepo.SourceAwareReceiptSchemaVersion)} {
		path := filepath.Join(finalRoot, name)
		info, err := os.Lstat(path)
		if err != nil || !info.IsDir() {
			t.Fatalf("%s: %v", path, err)
		}
		if runtime.GOOS != "windows" && info.Mode().Perm() != 0o700 {
			t.Fatalf("%s mode = %o, want 0700", path, info.Mode().Perm())
		}
	}
	if _, err := os.Lstat(filepath.Join(finalRoot, buildrepo.ArtifactsDir(buildrepo.LegacyReceiptSchemaVersion))); !os.IsNotExist(err) {
		t.Fatalf("a receipt-3 publication prepared the legacy namespace: %v", err)
	}
	recorded := marker.Read(filepath.Join(project, ".agents", "skills", "tooling"))
	if recorded == nil || recorded.Builds["etool"].CacheKey == "" {
		t.Fatalf("marker = %+v", recorded)
	}
	// The store proves the published entry through the prepared parents.
	if err := store.PrepareNamespaces(buildrepo.SourceAwareReceiptSchemaVersion); err != nil {
		t.Fatalf("prepared parents are not proved private: %v", err)
	}
	if again := draftBuildInstall(t, home, project, deps, draftExternalDeps(t)); again.Status != "ok" || len(builder.calls) != 2 {
		t.Fatalf("reinstall over the prepared parents = %+v calls=%v", again, builder.calls)
	}

	t.Run("foreign parent refused", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("mode-bit foreign-parent fixture is exercised on the unix runners; the Windows DACL refusal is asserted by TestWindowsProtectedSecurityDescriptorRejectsWrongOwnerAndDACL")
		}
		project, home, _ := draftProject(t, draftBuildPayload, nil)
		writeDraftBuildSkill(t, filepath.Join(project, "skills", "tooling"), "tooling", true)
		resolveDraftForInstall(t, project, home, draftBuildPayload)
		deps, _, builder := draftBuildDeps(t)
		foreign := filepath.Join(home, "external-build-cache", buildrepo.ArtifactsDir(buildrepo.SourceAwareReceiptSchemaVersion))
		if err := os.MkdirAll(foreign, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(foreign, 0o777); err != nil {
			t.Fatal(err)
		}
		result := draftBuildInstall(t, home, project, deps, draftExternalDeps(t))
		if result.Status == "ok" || !strings.Contains(strings.Join(append(result.Errors, result.Messages...), "\n"), buildrepo.CodeProtectedBoundaryUntrusted) {
			t.Fatalf("foreign parent accepted: %+v", result)
		}
		if entries := dirNames(t, foreign); len(entries) != 0 {
			t.Fatalf("refused install published %v under the foreign parent", entries)
		}
		if marker.Read(filepath.Join(project, ".agents", "skills", "tooling")) != nil {
			t.Fatal("refused install wrote a marker")
		}
		_ = builder
	})
}

// TestReviewExternalExecutionBindsExactReceipt3Input is the revision-5 review
// regression: the published external receipt's cache_key equals the
// independent hash of canonical receipt.input, and the execution receipt's
// build_input_sha256 equals that same digest, on both fresh publication and
// cache reuse. A compiler-view digest with the same package is a different
// value and must never satisfy this arm.
func TestReviewExternalExecutionBindsExactReceipt3Input(t *testing.T) {
	project, home, _ := draftProject(t, draftBuildPayload, nil)
	writeDraftBuildSkill(t, filepath.Join(project, "skills", "tooling"), "tooling", true)
	resolveDraftForInstall(t, project, home, draftBuildPayload)
	deps, _, builder := draftBuildDeps(t)
	result := draftBuildInstall(t, home, project, deps, draftExternalDeps(t))
	if result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	recorded := marker.Read(filepath.Join(project, ".agents", "skills", "tooling"))
	if recorded == nil {
		t.Fatal("missing marker")
	}
	entry := filepath.Join(home, "external-build-cache", buildrepo.ArtifactsDir(buildrepo.SourceAwareReceiptSchemaVersion), strings.TrimPrefix(string(recorded.Builds["etool"].CacheKey), "sha256:"))
	receipt := readReceiptObject(t, filepath.Join(entry, "receipt.json"))
	execution := readReceiptObject(t, filepath.Join(entry, "execution-receipt.ccj.json"))
	input := canonicalOf(t, receipt["input"])
	sum := sha256.Sum256(input)
	want := "sha256:" + hex.EncodeToString(sum[:])
	if receipt["cache_key"] != want {
		t.Fatal("test control: receipt cache key differs from wrapper digest")
	}
	if string(recorded.Builds["etool"].CacheKey) != want {
		t.Fatalf("marker cache key %s != wrapper digest %s", recorded.Builds["etool"].CacheKey, want)
	}
	if execution["build_input_sha256"] != want {
		t.Fatalf("successful install accepted wrong execution binding: got %v; exact receipt-3 input digest %s", execution["build_input_sha256"], want)
	}
	// Cache reuse binds the same digest without recompiling.
	again := draftBuildInstall(t, home, project, deps, draftExternalDeps(t))
	if again.Status != "ok" {
		t.Fatalf("reinstall = %+v", again)
	}
	if len(builder.calls) != 2 {
		t.Fatalf("reuse recompiled: %v", builder.calls)
	}
	reused := readReceiptObject(t, filepath.Join(entry, "execution-receipt.ccj.json"))
	if reused["build_input_sha256"] != want {
		t.Fatalf("reused execution binds %v, want %s", reused["build_input_sha256"], want)
	}
}

// TestDraftBuildRefusesCompilerViewDigest: a session that keeps the package
// but binds only the compiler go-v1 view digest is refused before any cache
// publication or marker write, even though its package matches.
func TestDraftBuildRefusesCompilerViewDigest(t *testing.T) {
	project, home, _ := draftProject(t, draftBuildPayload, nil)
	writeDraftBuildSkill(t, filepath.Join(project, "skills", "tooling"), "tooling", true)
	resolveDraftForInstall(t, project, home, draftBuildPayload)
	deps, _, builder := draftBuildDeps(t)
	builder.bindCompilerView = true
	result := draftBuildInstall(t, home, project, deps, draftExternalDeps(t))
	if result.Status == "ok" {
		t.Fatalf("compiler-view execution receipt accepted: %+v", result)
	}
	joined := strings.Join(append(result.Errors, result.Messages...), "\n")
	if !strings.Contains(joined, "receipt") {
		t.Fatalf("refusal does not name the receipt binding: %s", joined)
	}
	for _, dir := range []string{
		filepath.Join(home, "cache", "build", buildcache.Namespace),
		filepath.Join(home, "cache", "build", buildcache.SourceAwareNamespace),
		filepath.Join(home, "external-build-cache", buildrepo.ArtifactsDir(buildrepo.LegacyReceiptSchemaVersion)),
		filepath.Join(home, "external-build-cache", buildrepo.ArtifactsDir(buildrepo.SourceAwareReceiptSchemaVersion)),
	} {
		if entries := dirNames(t, dir); len(entries) != 0 {
			t.Fatalf("refused install published %v under %s", entries, dir)
		}
	}
	if marker.Read(filepath.Join(project, ".agents", "skills", "tooling")) != nil {
		t.Fatal("refused install wrote a marker")
	}
}
