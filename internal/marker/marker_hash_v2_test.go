package marker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/hashing"
)

func enableV1Writers(t *testing.T) {
	t.Helper()
	prior := hashing.EnableV2Writers
	hashing.EnableV2Writers = false
	t.Cleanup(func() { hashing.EnableV2Writers = prior })
}

func enableV2Writers(t *testing.T) {
	t.Helper()
	prior := hashing.EnableV2Writers
	hashing.EnableV2Writers = true
	t.Cleanup(func() { hashing.EnableV2Writers = prior })
}

// TestContentHashVersionFollowsTheMarkerNotTheWriter pins the reader rule
// status depends on: a marker's content framing comes from its own shape,
// never from the ambient writer switch. Frozen schemas and draft package
// markers read v1 under either switch; only a core v5 marker reads v2.
func TestContentHashVersionFollowsTheMarkerNotTheWriter(t *testing.T) {
	prior := hashing.EnableV2Writers
	t.Cleanup(func() { hashing.EnableV2Writers = prior })
	for _, writerV2 := range []bool{false, true} {
		hashing.EnableV2Writers = writerV2
		for _, tc := range []struct {
			name   string
			marker *Marker
			want   hashing.Version
		}{
			{"schema-1 reads v1", &Marker{SchemaVersion: LegacySchemaVersion}, hashing.VersionV1},
			{"schema-2 reads v1", &Marker{SchemaVersion: SchemaVersion}, hashing.VersionV1},
			{"schema-3 reads v1", &Marker{SchemaVersion: ExternalSchemaVersion}, hashing.VersionV1},
			{"schema-4 reads v1", &Marker{SchemaVersion: PolicySchemaVersion}, hashing.VersionV1},
			{"core v5 reads v2", &Marker{SchemaVersion: SchemaV5, HashVersion: hashing.VersionV2}, hashing.VersionV2},
			{"draft v5 reads v1", &Marker{SchemaVersion: SchemaV5, Package: &Package{Kind: "local-snapshot", Snapshot: "sha256:" + strings.Repeat("a", 64)}}, hashing.VersionV1},
		} {
			if got := tc.marker.ContentHashVersion(); got != tc.want {
				t.Fatalf("writer v2=%v: %s = %d, want %d", writerV2, tc.name, got, tc.want)
			}
		}
	}
}

func TestWritePreservesContentHashInRC13Mode(t *testing.T) {
	enableV1Writers(t)
	for _, tc := range []struct {
		skillSchema int
		marker      int
	}{{skillSchema: 6, marker: SchemaVersion}, {skillSchema: 7, marker: ExternalSchemaVersion}, {skillSchema: 8, marker: PolicySchemaVersion}} {
		t.Run(strconv.Itoa(tc.skillSchema), func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("legacy writer"), 0o600); err != nil {
				t.Fatal(err)
			}
			m := validMarkerV2()
			m.SkillSchemaVersion = tc.skillSchema
			callerHash := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
			m.ContentSHA256 = callerHash
			if err := Write(dir, m); err != nil {
				t.Fatal(err)
			}
			if m.SchemaVersion != tc.marker || m.HashVersion != 0 || m.ContentSHA256 != callerHash {
				t.Fatalf("default marker writer = schema:%d hash_version:%d content_sha256:%s, want %d/no hash_version/%s", m.SchemaVersion, m.HashVersion, m.ContentSHA256, tc.marker, callerHash)
			}
			payload, err := os.ReadFile(filepath.Join(dir, Name))
			if err != nil {
				t.Fatal(err)
			}
			var raw map[string]json.RawMessage
			if err := json.Unmarshal(payload, &raw); err != nil {
				t.Fatal(err)
			}
			if _, present := raw["hash_version"]; present {
				t.Fatal("default marker writer added hash_version to an rc.13 shape")
			}
			if string(raw["content_sha256"]) != `"`+callerHash+`"` {
				t.Fatalf("default marker wire hash = %s, want supplied v1 %q", raw["content_sha256"], callerHash)
			}
		})
	}
}

func TestWriteEmitsCoreMarkerV5WithHashVersion2(t *testing.T) {
	enableV2Writers(t)
	for _, skillSchema := range []int{6, 7, 8} {
		t.Run(strconv.Itoa(skillSchema), func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("v2 marker"), 0o600); err != nil {
				t.Fatal(err)
			}
			m := validMarkerV2()
			m.SkillSchemaVersion = skillSchema
			m.ContentSHA256 = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
			if err := Write(dir, m); err != nil {
				t.Fatal(err)
			}
			want, err := hashing.ContentSHA256WithVersion(dir, nil, hashing.VersionV2)
			if err != nil {
				t.Fatal(err)
			}
			if m.SchemaVersion != SchemaV5 || m.HashVersion != hashing.VersionV2 || m.ContentSHA256 != want {
				t.Fatalf("writer state = schema:%d hash_version:%d content_sha256:%s, want 5/2/%s", m.SchemaVersion, m.HashVersion, m.ContentSHA256, want)
			}
			var raw map[string]any
			payload, err := os.ReadFile(filepath.Join(dir, Name))
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(payload, &raw); err != nil {
				t.Fatal(err)
			}
			if raw["schema_version"] != float64(5) || raw["hash_version"] != float64(2) || raw["content_sha256"] != want {
				t.Fatalf("wire state = schema:%v hash:%v content:%v", raw["schema_version"], raw["hash_version"], raw["content_sha256"])
			}
			current, err := Current(dir, m)
			if err != nil || !current {
				t.Fatalf("fresh v2 marker current=%v err=%v", current, err)
			}
		})
	}
}

func TestAuthoritativeCompiledMarkerRoundTripsThroughV2Writer(t *testing.T) {
	enableV2Writers(t)
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
	// The fixture predates v5's local-build policy fields; provide the source
	// boundary supplied by a current install plan before publishing it.
	m.BuildRoots = []string{"build"}
	m.BuildSource = &buildsource.Identity{
		Algorithm: buildsource.Algorithm, ContentSHA256: "sha256:" + strings.Repeat("b", 64),
	}
	expectedBuilds := make(map[string]Build, len(m.Builds))
	for name, build := range m.Builds {
		if build.Driver == buildmeta.DriverGoV1 {
			build.ReceiptSchemaVersion = 1
			build.ExecutionPolicy = buildmeta.ExecutionPolicy
		}
		m.Builds[name] = build
		expectedBuilds[name] = build
	}
	destination := t.TempDir()
	if err := Write(destination, m); err != nil {
		t.Fatal(err)
	}
	recorded := Read(destination)
	if recorded == nil || recorded.SchemaVersion != SchemaV5 || recorded.HashVersion != hashing.VersionV2 {
		t.Fatalf("writer did not upgrade the compiled marker to core v5/hash v2: %+v", recorded)
	}
	if !reflect.DeepEqual(recorded.Builds, expectedBuilds) {
		t.Fatalf("compiled build records do not carry core v5 policy fields:\n got %+v\nwant %+v", recorded.Builds, expectedBuilds)
	}
}

func TestReadLegacyV1AndRewriteAsCoreMarkerV5WithHashVersion2(t *testing.T) {
	enableV2Writers(t)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("legacy"), 0o644); err != nil {
		t.Fatal(err)
	}
	legacyHash, err := hashing.ContentSHA256(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	legacy := validMarkerV2()
	legacy.SchemaVersion = LegacySchemaVersion
	legacy.SkillSchemaVersion = 5
	legacy.ContentSHA256 = legacyHash
	legacy.BuildRoots = nil
	legacy.Builds = nil
	payload, err := marshalLegacy(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, Name), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	recorded := Read(dir)
	if recorded == nil || recorded.SchemaVersion != LegacySchemaVersion {
		t.Fatalf("legacy marker read = %+v", recorded)
	}
	if err := Write(dir, recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.SchemaVersion != SchemaV5 || recorded.HashVersion != hashing.VersionV2 {
		t.Fatalf("rewritten marker versions = schema:%d hash:%d, want 5/2", recorded.SchemaVersion, recorded.HashVersion)
	}
	v2Hash, err := hashing.ContentSHA256WithVersion(dir, nil, hashing.VersionV2)
	if err != nil || recorded.ContentSHA256 != v2Hash || recorded.ContentSHA256 == legacyHash {
		t.Fatalf("rewritten hashes = marker:%s v2:%s v1:%s err:%v", recorded.ContentSHA256, v2Hash, legacyHash, err)
	}
	if current, err := Current(dir, recorded); err != nil || !current {
		t.Fatalf("rewritten v2 marker current=%v err=%v", current, err)
	}
}

func TestReadLegacyV1AndRewriteAsCoreMarkerV5PreservesEmptyRequirer(t *testing.T) {
	enableV2Writers(t)
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
		t.Fatalf("rewrite marker: %v", err)
	}
	rewritten := Read(dir)
	if rewritten == nil || rewritten.SchemaVersion != SchemaV5 || rewritten.HashVersion != hashing.VersionV2 ||
		!reflect.DeepEqual(rewritten.Requirers, []string{""}) {
		t.Fatalf("rewritten marker = %#v", rewritten)
	}
}

func TestCurrentRejectsV1MarkerForV2Expectation(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("legacy marker"), 0o600); err != nil {
		t.Fatal(err)
	}
	legacyHash, err := hashing.ContentSHA256WithVersion(dir, nil, hashing.VersionV1)
	if err != nil {
		t.Fatal(err)
	}
	legacy := validMarkerV2()
	legacy.SchemaVersion = PolicySchemaVersion
	legacy.SkillSchemaVersion = 8
	legacy.ContentSHA256 = legacyHash
	payload, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, Name), append(payload, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	expected := *legacy
	expected.SchemaVersion = SchemaV5
	expected.HashVersion = hashing.VersionV2
	current, err := Current(dir, &expected)
	if err != nil {
		t.Fatal(err)
	}
	if current {
		t.Fatal("a v1 marker was current for a v2 expectation")
	}
}
