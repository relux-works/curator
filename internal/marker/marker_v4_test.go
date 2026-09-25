package marker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/hashing"
)

// v4Base is v3Base moved onto manifest schema 8. Marker v4 is marker v3 with
// `schema_version` 4 and `skill_schema_version` 8 and no other difference, so
// every build-record rule below is the v3 rule unchanged.
func v4Base() *Marker {
	m := v3Base()
	m.SkillSchemaVersion = 8
	return m
}

func TestMarkerV4IsWrittenForSchema8Installations(t *testing.T) {
	for _, testCase := range []struct {
		name string
		set  func(*Marker)
	}{
		{name: "local-only", set: func(m *Marker) { m.Commands = []string{"local"}; m.Builds["local"] = localV3Build() }},
		{name: "external-only", set: func(m *Marker) {
			m.Commands = []string{"external"}
			m.BuildRoots = []string{}
			m.BuildSource = nil
			m.Builds["external"] = externalV3Build()
		}},
		{name: "mixed", set: func(m *Marker) { m.Builds["local"] = localV3Build(); m.Builds["external"] = externalV3Build() }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			dir := t.TempDir()
			m := v4Base()
			testCase.set(m)
			if err := Write(dir, m); err != nil {
				t.Fatal(err)
			}
			if m.SchemaVersion != PolicySchemaVersion {
				t.Fatalf("schema=%d, want %d", m.SchemaVersion, PolicySchemaVersion)
			}
			recorded := Read(dir)
			if recorded == nil || recorded.SchemaVersion != PolicySchemaVersion || recorded.SkillSchemaVersion != 8 ||
				len(recorded.Builds) != len(m.Builds) {
				t.Fatalf("round trip=%+v", recorded)
			}
			payload, err := os.ReadFile(filepath.Join(dir, Name))
			if err != nil {
				t.Fatal(err)
			}
			var raw map[string]any
			if err := json.Unmarshal(payload, &raw); err != nil {
				t.Fatal(err)
			}
			if raw["schema_version"] != float64(4) || raw["skill_schema_version"] != float64(8) {
				t.Fatalf("payload versions = %v/%v", raw["schema_version"], raw["skill_schema_version"])
			}
			for command, value := range raw["builds"].(map[string]any) {
				record := value.(map[string]any)
				if record["execution_policy"] != "manager-worker-v1" {
					t.Fatalf("%s records execution policy %v", command, record["execution_policy"])
				}
			}
		})
	}
}

// TestMarkerV4BandIsExact keeps every marker version bound to exactly the
// manifest band it was defined for: v4 never represents schema 7, and v3 never
// represents schema 8.
func TestMarkerV4BandIsExact(t *testing.T) {
	for _, testCase := range []struct {
		name          string
		markerVersion int
		skillVersion  int
	}{
		{"v4 with schema 7", PolicySchemaVersion, 7},
		{"v4 with schema 6", PolicySchemaVersion, 6},
		{"v3 with schema 8", ExternalSchemaVersion, 8},
		{"v2 with schema 8", SchemaVersion, 8},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			m := v4Base()
			m.Commands = []string{"local"}
			m.Builds["local"] = localV3Build()
			m.SchemaVersion = testCase.markerVersion
			m.SkillSchemaVersion = testCase.skillVersion
			payload, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			var raw map[string]json.RawMessage
			if err := json.Unmarshal(payload, &raw); err != nil {
				t.Fatal(err)
			}
			if validMarker(m, raw) {
				t.Fatalf("marker v%d accepted skill schema %d", testCase.markerVersion, testCase.skillVersion)
			}
		})
	}
}

// TestMarkerV4CurrentnessRequiresBuildProof proves a schema-8 marker fails
// closed exactly as a schema-7 one does when the build evidence is missing.
func TestMarkerV4CurrentnessRequiresBuildProof(t *testing.T) {
	dir := t.TempDir()
	m := v4Base()
	m.Commands, m.BuildRoots, m.BuildSource = []string{"external"}, []string{}, nil
	m.Builds["external"] = externalV3Build()
	hash, err := hashing.ContentSHA256(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	m.ContentSHA256 = hash
	if err := Write(dir, m); err != nil {
		t.Fatal(err)
	}
	state := BuildCurrentness{ContextFiles: []string{}, RuntimeFiles: []string{},
		InspectExternal: func(command string, recorded Build) (bool, error) {
			return command == "external" && recorded.ReceiptSchemaVersion == 2, nil
		},
		VerifyShim: func(command string, recorded Build) (bool, error) {
			return command == "external" && recorded.ArtifactPath == "bin/external", nil
		},
	}
	current, err := Current(dir, m, state)
	if err != nil || !current {
		t.Fatalf("current=%v err=%v", current, err)
	}
	state.VerifyShim = nil
	if current, err = Current(dir, m, state); err != nil || current {
		t.Fatalf("missing shim proof current=%v err=%v", current, err)
	}
	current, err = Current(dir, m)
	if err != nil || current {
		t.Fatalf("build state omitted: current=%v err=%v", current, err)
	}
}

// TestMarkerV4LocalBuildRequiresBuildSource keeps the local-build rule of
// marker v3 in force for v4: a recorded go-v1 build without a frozen source
// identity is not a representable marker.
func TestMarkerV4LocalBuildRequiresBuildSource(t *testing.T) {
	dir := t.TempDir()
	m := v4Base()
	m.Commands = []string{"local"}
	m.Builds["local"] = localV3Build()
	m.BuildSource = nil
	if err := Write(dir, m); err == nil {
		t.Fatal("a local schema-8 build without a build source must not be writable")
	}
	m.BuildSource = &buildsource.Identity{Algorithm: buildsource.Algorithm, ContentSHA256: "sha256:" + strings.Repeat("b", 64)}
	if err := Write(dir, m); err != nil {
		t.Fatal(err)
	}
}
