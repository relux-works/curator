package marker

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/hashing"
)

func validMarkerV2() *Marker {
	return &Marker{
		Name: "skill-a", Source: "skill-a", RefKind: "revision",
		Ref:           "0123456789abcdef0123456789abcdef01234567",
		Commit:        "0123456789abcdef0123456789abcdef01234567",
		ContentSHA256: "sha256:" + strings.Repeat("a", 64),
		Locale:        "en", Agents: []string{"codex_cli"}, Commands: []string{},
		Dependencies: []string{}, SkillSchemaVersion: 6, RuntimeRoots: []string{},
		BuildRoots: []string{}, InstalledAt: "2026-07-21T00:00:00Z", Files: []string{},
		Builds: map[string]Build{},
	}
}

func TestReadAuthoritativeMarkerV2SchemaCases(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	casesDir := filepath.Join(root, "schema-cases", "install-marker-v2")
	entries, err := os.ReadDir(casesDir)
	if err != nil {
		t.Fatal(err)
	}
	seen := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		seen++
		t.Run(entry.Name(), func(t *testing.T) {
			dir := t.TempDir()
			payload, err := os.ReadFile(filepath.Join(casesDir, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, Name), payload, 0o644); err != nil {
				t.Fatal(err)
			}
			wantValid := strings.HasPrefix(entry.Name(), "valid")
			if got := Read(dir) != nil; got != wantValid {
				t.Fatalf("Read valid = %v, want %v", got, wantValid)
			}
		})
	}
	if seen < 10 {
		t.Fatalf("only %d authoritative marker cases found", seen)
	}
}

func TestWriteAlwaysProducesCanonicalMarkerV2(t *testing.T) {
	dir := t.TempDir()
	m := validMarkerV2()
	m.SchemaVersion = LegacySchemaVersion
	m.SkillSchemaVersion = 5
	m.Agents = []string{"zed", "codex_cli"}
	m.BuildRoots = nil
	m.Builds = nil
	if err := Write(dir, m); err != nil {
		t.Fatal(err)
	}
	if m.SchemaVersion != SchemaVersion || m.BuildRoots == nil || m.Builds == nil {
		t.Fatalf("canonical marker = %+v", m)
	}
	if !reflect.DeepEqual(m.Agents, []string{"codex_cli", "zed"}) {
		t.Fatalf("agents = %v", m.Agents)
	}
	payload, err := os.ReadFile(filepath.Join(dir, Name))
	if err != nil {
		t.Fatal(err)
	}
	if len(payload) == 0 || payload[len(payload)-1] != '\n' {
		t.Fatal("marker must have one terminal LF")
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	if string(raw["schema_version"]) != "2" || string(raw["build_roots"]) != "[]" || string(raw["builds"]) != "{}" {
		t.Fatalf("unexpected v2 wire state: %s", payload)
	}
	if _, present := raw["build_source"]; present {
		t.Fatal("empty builds must omit build_source")
	}
	if Read(dir) == nil {
		t.Fatal("writer produced an unreadable marker")
	}
}

func TestAuthoritativeCompiledMarkerRoundTripsThroughWriter(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	sourceDir := t.TempDir()
	payload, err := os.ReadFile(filepath.Join(root, "expected", "build-driver", "marker.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, Name), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	m := Read(sourceDir)
	if m == nil {
		t.Fatal("authoritative compiled marker is unreadable")
	}
	destination := t.TempDir()
	if err := Write(destination, m); err != nil {
		t.Fatal(err)
	}
	written, err := os.ReadFile(filepath.Join(destination, Name))
	if err != nil {
		t.Fatal(err)
	}
	var got, want any
	if err := json.Unmarshal(written, &got); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(payload, &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("compiled marker round trip:\n got %s\nwant %s", written, payload)
	}
}

func TestWriteRejectsInvalidV2State(t *testing.T) {
	tests := map[string]func(*Marker){
		"unsatisfied build source": func(m *Marker) {
			m.Builds["tool"] = Build{Driver: buildmeta.DriverGoV1}
		},
		"source without builds": func(m *Marker) {
			m.BuildSource = &buildsource.Identity{Algorithm: buildsource.Algorithm, ContentSHA256: "sha256:" + strings.Repeat("b", 64)}
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			m := validMarkerV2()
			mutate(m)
			if err := Write(t.TempDir(), m); err == nil {
				t.Fatal("invalid marker write succeeded")
			}
		})
	}
}

func TestReadLegacyV1AndRewriteAsV2WithoutChangingContentHash(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("legacy"), 0o644); err != nil {
		t.Fatal(err)
	}
	contentHash, err := hashing.ContentSHA256(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	legacy := validMarkerV2()
	legacy.SchemaVersion = LegacySchemaVersion
	legacy.SkillSchemaVersion = 5
	legacy.ContentSHA256 = contentHash
	legacy.BuildRoots = nil
	legacy.Builds = nil
	payload, err := marshalLegacy(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, Name), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	for skillSchema := 1; skillSchema <= 5; skillSchema++ {
		legacy.SkillSchemaVersion = skillSchema
		payload, err = marshalLegacy(legacy)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, Name), payload, 0o644); err != nil {
			t.Fatal(err)
		}
		recorded := Read(dir)
		if recorded == nil || recorded.SchemaVersion != LegacySchemaVersion {
			t.Fatalf("schema %d legacy read = %+v", skillSchema, recorded)
		}
		current, err := Current(dir, legacy)
		if err != nil || !current {
			t.Fatalf("schema %d legacy Current = %v, %v", skillSchema, current, err)
		}
	}
	recorded := Read(dir)
	legacyJSON, err := json.Marshal(recorded)
	if err != nil {
		t.Fatal(err)
	}
	var legacyRaw map[string]json.RawMessage
	if err := json.Unmarshal(legacyJSON, &legacyRaw); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"build_roots", "build_source", "builds"} {
		if _, present := legacyRaw[field]; present {
			t.Fatalf("legacy marshal unexpectedly contains %s: %s", field, legacyJSON)
		}
	}
	if err := Write(dir, recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.SchemaVersion != SchemaVersion {
		t.Fatalf("rewritten schema = %d", recorded.SchemaVersion)
	}
	after, err := hashing.ContentSHA256(dir, nil)
	if err != nil || after != contentHash {
		t.Fatalf("marker-excluding hash = %s, %v; want %s", after, err, contentHash)
	}
}

func TestReadLegacyV1AndRewriteV2PreservesEmptyRequirer(t *testing.T) {
	dir := t.TempDir()
	legacy := validMarkerV2()
	legacy.SchemaVersion = LegacySchemaVersion
	legacy.SkillSchemaVersion = 5
	legacy.BuildRoots = nil
	legacy.Builds = nil
	legacy.Requirers = []string{""}
	payload, err := marshalLegacy(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, Name), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	recorded := Read(dir)
	if recorded == nil || !reflect.DeepEqual(recorded.Requirers, []string{""}) {
		t.Fatalf("legacy requirers = %#v; want one empty string", recorded)
	}
	if err := Write(dir, recorded); err != nil {
		t.Fatalf("rewrite v2: %v", err)
	}
	rewritten := Read(dir)
	if rewritten == nil || rewritten.SchemaVersion != SchemaVersion ||
		!reflect.DeepEqual(rewritten.Requirers, []string{""}) {
		t.Fatalf("rewritten marker = %#v", rewritten)
	}
}

func TestCompiledCurrentnessRoundTripAndCallbackOrder(t *testing.T) {
	fixture := compiledFixture(t)
	order := []string{}
	options := fixture.options
	open := options.RawSnapshot
	options.RawSnapshot = func() (*buildsource.Token, error) {
		order = append(order, "snapshot")
		return open()
	}
	inspect := options.InspectCache
	options.InspectCache = func(command string, expectation buildcache.Expectation) buildcache.Result {
		order = append(order, "cache")
		return inspect(command, expectation)
	}
	current, err := Current(fixture.installed, fixture.marker, options)
	if err != nil || !current {
		t.Fatalf("Current = %v, %v", current, err)
	}
	if !reflect.DeepEqual(order, []string{"snapshot", "cache"}) {
		t.Fatalf("callback order = %v", order)
	}
}

func TestCompiledCurrentnessFailureMatrix(t *testing.T) {
	tests := map[string]func(*compiledState){
		"missing raw snapshot": func(state *compiledState) {
			state.options.RawSnapshot = func() (*buildsource.Token, error) { return nil, nil }
		},
		"changed build roots": func(state *compiledState) { state.expected.BuildRoots = []string{"other"} },
		"build source mismatch": func(state *compiledState) {
			state.expected.BuildSource.ContentSHA256 = "sha256:" + strings.Repeat("0", 64)
		},
		"context-visible build file": func(state *compiledState) { state.options.ContextFiles = []string{"build/main.go"} },
		"runtime-copied build file":  func(state *compiledState) { state.options.RuntimeFiles = []string{"build/main.go"} },
		"missing currentness proof":  func(state *compiledState) { state.options.ContextFiles = nil },
		"untrusted cache":            func(state *compiledState) { state.result.Status = buildcache.UntrustedProvenance },
		"missing cache":              func(state *compiledState) { state.result.Status = buildcache.Miss },
		"corrupt receipt":            func(state *compiledState) { state.result.Status = buildcache.Corrupt },
		"wrong input":                func(state *compiledState) { state.result.Receipt.Input.SourceDir = "build/other" },
		"wrong key": func(state *compiledState) {
			state.result.Receipt.CacheKey = buildmeta.CacheKey("sha256:" + strings.Repeat("0", 64))
		},
		"receipt hash drift": func(state *compiledState) {
			state.result.ReceiptHash = buildmeta.ReceiptHash("sha256:" + strings.Repeat("0", 64))
		},
		"receipt bytes drift": func(state *compiledState) { state.result.ReceiptBytes = append(state.result.ReceiptBytes, '\n') },
		"artifact drift":      func(state *compiledState) { state.result.Receipt.Artifact.SHA256 = "sha256:" + strings.Repeat("0", 64) },
		"path mismatch":       func(state *compiledState) { state.result.Receipt.Artifact.Path = "bin/other" },
		"wrong target": func(state *compiledState) {
			state.options.Inputs["tool"] = mutateInput(state.options.Inputs["tool"], func(input *buildmeta.Input) { input.Target.GOOS = "linux" })
		},
		"wrong toolchain": func(state *compiledState) {
			state.options.Inputs["tool"] = mutateInput(state.options.Inputs["tool"], func(input *buildmeta.Input) { input.Toolchain.GoVersion = "go version go1.25.6 darwin/arm64" })
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			fixture := compiledFixture(t)
			state := &compiledState{expected: cloneMarker(fixture.marker), options: fixture.options, result: fixture.result}
			state.options.Inputs = cloneInputs(state.options.Inputs)
			state.options.InspectCache = func(string, buildcache.Expectation) buildcache.Result { return state.result }
			mutate(state)
			current, err := Current(fixture.installed, state.expected, state.options)
			if err != nil {
				t.Fatal(err)
			}
			if current {
				t.Fatal("drift must be non-current")
			}
		})
	}
}

func TestCompiledCurrentnessTreatsSnapshotIOAsUnknown(t *testing.T) {
	fixture := compiledFixture(t)
	fixture.options.RawSnapshot = func() (*buildsource.Token, error) { return nil, errors.New("storage unavailable") }
	if current, err := Current(fixture.installed, fixture.marker, fixture.options); err == nil || current {
		t.Fatalf("Current = %v, %v; want unknown error", current, err)
	}
}

func TestBuildCurrentnessResultPropagatesSnapshotCloseFailure(t *testing.T) {
	closeErr := errors.New("close snapshot")
	current, err := buildCurrentnessResult(true, buildsource.ErrSnapshotMutated, closeErr)
	if current || !errors.Is(err, closeErr) || !errors.Is(err, buildsource.ErrSnapshotMutated) {
		t.Fatalf("buildCurrentnessResult = %v, %v; want joined close and mutation error", current, err)
	}
}

func TestPackageRootMarkerBytesAffectBuildSourceButNotInstalledContentHash(t *testing.T) {
	fixture := compiledFixture(t)
	before, err := hashing.ContentSHA256(fixture.installed, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixture.raw, Name), []byte("different package marker bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	current, err := Current(fixture.installed, fixture.marker, fixture.options)
	if err != nil {
		t.Fatal(err)
	}
	if current {
		t.Fatal("package root marker drift must invalidate build source")
	}
	after, err := hashing.ContentSHA256(fixture.installed, nil)
	if err != nil || after != before {
		t.Fatalf("installed marker-excluding hash = %s, %v; want %s", after, err, before)
	}
}

func TestSnapshotMutationDuringCacheInspectionIsNonCurrent(t *testing.T) {
	fixture := compiledFixture(t)
	inspect := fixture.options.InspectCache
	fixture.options.InspectCache = func(command string, expectation buildcache.Expectation) buildcache.Result {
		if err := os.WriteFile(filepath.Join(fixture.raw, "SKILL.md"), []byte("mutated during inspection"), 0o644); err != nil {
			t.Fatal(err)
		}
		return inspect(command, expectation)
	}
	current, err := Current(fixture.installed, fixture.marker, fixture.options)
	if err != nil {
		t.Fatal(err)
	}
	if current {
		t.Fatal("snapshot mutation through the cache boundary must be non-current")
	}
}

type compiledFixtureState struct {
	installed string
	raw       string
	marker    *Marker
	options   BuildCurrentness
	result    buildcache.Result
}

type compiledState struct {
	expected *Marker
	options  BuildCurrentness
	result   buildcache.Result
}

func compiledFixture(t *testing.T) compiledFixtureState {
	t.Helper()
	installed := t.TempDir()
	if err := os.WriteFile(filepath.Join(installed, "SKILL.md"), []byte("context"), 0o644); err != nil {
		t.Fatal(err)
	}
	contentHash, err := hashing.ContentSHA256(installed, nil)
	if err != nil {
		t.Fatal(err)
	}
	raw := t.TempDir()
	if err := os.MkdirAll(filepath.Join(raw, "build", "cmd", "tool"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(raw, "SKILL.md"), []byte("raw"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(raw, Name), []byte("package marker bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	token, err := buildsource.Validate(raw)
	if err != nil {
		t.Fatal(err)
	}
	identity := token.Identity()
	if err := token.Close(); err != nil {
		t.Fatal(err)
	}
	input := buildmeta.Input{
		SchemaVersion: buildmeta.SchemaVersion, Driver: buildmeta.DriverGoV1, BuildSource: identity,
		BuildRoot: "build", Command: "tool", SourceDir: "build/cmd/tool",
		Target:    buildmeta.Target{GOOS: "darwin", GOARCH: "arm64", Tuning: map[string]string{"GOARM64": "v8.0"}},
		Toolchain: buildmeta.Toolchain{Algorithm: buildmeta.ToolchainAlgorithm, GoRelpath: buildmeta.ToolchainGoRelpath, GoVersion: "go version go1.25.5 darwin/arm64", ContentSHA256: "sha256:" + strings.Repeat("c", 64)},
		Policy:    buildmeta.FixedPolicy(),
	}
	key, err := input.CacheKey()
	if err != nil {
		t.Fatal(err)
	}
	artifact := buildmeta.Artifact{Path: "bin/tool", SHA256: "sha256:" + strings.Repeat("d", 64), Size: 7}
	receipt, err := buildmeta.NewReceipt(input, artifact)
	if err != nil {
		t.Fatal(err)
	}
	receiptBytes, err := receipt.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	receiptHash, err := buildmeta.HashReceiptBytes(receiptBytes)
	if err != nil {
		t.Fatal(err)
	}
	m := validMarkerV2()
	m.ContentSHA256 = contentHash
	m.Commands = []string{"tool"}
	m.BuildRoots = []string{"build"}
	m.BuildSource = &identity
	m.Files = []string{"SKILL.md"}
	m.Builds = map[string]Build{"tool": {Driver: buildmeta.DriverGoV1, CacheKey: key, ReceiptSHA256: receiptHash, ArtifactSHA256: artifact.SHA256, ArtifactPath: artifact.Path}}
	if err := Write(installed, m); err != nil {
		t.Fatal(err)
	}
	result := buildcache.Result{Status: buildcache.Hit, Receipt: receipt, ReceiptBytes: receiptBytes, ReceiptHash: receiptHash, ArtifactPath: filepath.Join(t.TempDir(), "bin", "tool")}
	options := BuildCurrentness{
		RawSnapshot:  func() (*buildsource.Token, error) { return buildsource.Validate(raw) },
		Inputs:       map[string]buildmeta.Input{"tool": input},
		ContextFiles: []string{"SKILL.md"}, RuntimeFiles: []string{},
	}
	options.InspectCache = func(_ string, expectation buildcache.Expectation) buildcache.Result {
		if !reflect.DeepEqual(expectation.Input, input) || expectation.ReceiptHash != receiptHash {
			return buildcache.Result{Status: buildcache.Corrupt}
		}
		return result
	}
	return compiledFixtureState{installed: installed, raw: raw, marker: m, options: options, result: result}
}

func mutateInput(input buildmeta.Input, mutate func(*buildmeta.Input)) buildmeta.Input {
	mutate(&input)
	return input
}

func cloneInputs(inputs map[string]buildmeta.Input) map[string]buildmeta.Input {
	clone := make(map[string]buildmeta.Input, len(inputs))
	for command, input := range inputs {
		clone[command] = input
	}
	return clone
}

func cloneMarker(marker *Marker) *Marker {
	clone := *marker
	clone.BuildRoots = append([]string(nil), marker.BuildRoots...)
	clone.Builds = make(map[string]Build, len(marker.Builds))
	for command, build := range marker.Builds {
		clone.Builds[command] = build
	}
	if marker.BuildSource != nil {
		identity := *marker.BuildSource
		clone.BuildSource = &identity
	}
	return &clone
}

func marshalLegacy(marker *Marker) ([]byte, error) {
	value := map[string]any{
		"schema_version": LegacySchemaVersion, "name": marker.Name, "source": marker.Source,
		"ref_kind": marker.RefKind, "ref": marker.Ref, "commit": marker.Commit,
		"content_sha256": marker.ContentSHA256, "locale": marker.Locale, "agents": marker.Agents,
		"commands": marker.Commands, "dependencies": marker.Dependencies,
		"skill_schema_version": marker.SkillSchemaVersion, "runtime_roots": marker.RuntimeRoots,
		"installed_at": marker.InstalledAt, "files": marker.Files,
	}
	if marker.Requirers != nil {
		value["requirers"] = marker.Requirers
	}
	return json.Marshal(value)
}

func TestV2WriterSortsEverySetLikeField(t *testing.T) {
	m := validMarkerV2()
	m.Agents = []string{"zed", "alpha"}
	m.Commands = []string{"zed", "alpha"}
	m.Dependencies = []string{"zed", "alpha"}
	m.RuntimeRoots = []string{"zed", "alpha"}
	m.BuildRoots = []string{"zed", "alpha"}
	m.Files = []string{"zed", "alpha"}
	m.Requirements = []string{"zed", "alpha"}
	m.Requirers = []string{"zed", "<project>"}
	m.Activation = &Activation{Commands: []string{"zed", "alpha"}}
	m.McpServers = map[string][]string{"mcp": {"zed", "alpha"}}
	if err := Write(t.TempDir(), m); err != nil {
		t.Fatal(err)
	}
	sets := [][]string{m.Agents, m.Commands, m.Dependencies, m.RuntimeRoots, m.BuildRoots, m.Files, m.Requirements, m.Requirers, m.Activation.Commands, m.McpServers["mcp"]}
	for _, set := range sets {
		if !sort.StringsAreSorted(set) {
			t.Fatalf("unsorted set: %v", set)
		}
	}
}
