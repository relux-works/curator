package marker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/hashing"
)

func v3Base() *Marker {
	m := validMarkerV2()
	m.SkillSchemaVersion = 7
	m.Commands = []string{"external", "local"}
	m.BuildRoots = []string{"build"}
	m.BuildSource = &buildsource.Identity{Algorithm: buildsource.Algorithm, ContentSHA256: "sha256:" + strings.Repeat("b", 64)}
	m.Builds = map[string]Build{}
	return m
}

func TestMarkerV3ExternalCurrentnessIsReadOnlyAndFailsClosed(t *testing.T) {
	dir := t.TempDir()
	m := v3Base()
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
	checks := 0
	state := BuildCurrentness{ContextFiles: []string{}, RuntimeFiles: []string{},
		InspectExternal: func(command string, recorded Build) (bool, error) {
			checks++
			return command == "external" && recorded.ReceiptSchemaVersion == 2, nil
		},
		VerifyShim: func(command string, recorded Build) (bool, error) {
			checks++
			return command == "external" && recorded.ArtifactPath == "bin/external", nil
		},
	}
	current, err := Current(dir, m, state)
	if err != nil || !current || checks != 2 {
		t.Fatalf("current=%v checks=%d err=%v", current, checks, err)
	}
	state.VerifyShim = nil
	if current, err = Current(dir, m, state); err != nil || current {
		t.Fatalf("missing shim proof current=%v err=%v", current, err)
	}
}

func localV3Build() Build {
	return Build{Driver: buildmeta.DriverGoV1, ReceiptSchemaVersion: 1, ExecutionPolicy: buildmeta.ExecutionPolicy,
		CacheKey: buildmeta.CacheKey("sha256:" + strings.Repeat("1", 64)), ReceiptSHA256: buildmeta.ReceiptHash("sha256:" + strings.Repeat("2", 64)),
		ArtifactSHA256: "sha256:" + strings.Repeat("3", 64), ArtifactPath: "bin/local"}
}

func externalV3Build() Build {
	return Build{Driver: "go-repository-v1", ReceiptSchemaVersion: 2, ExecutionPolicy: buildmeta.ExecutionPolicy,
		Repository: "tools", DeclaredIdentity: &RepositoryIdentity{Kind: "network-git", Value: "github.com/example/tools"},
		DeclaredLockedCommit: &RepositoryCommit{ObjectFormat: "sha1", Hex: strings.Repeat("a", 40)}, DeclaredTag: "v1.0.0",
		EffectiveIdentity: &RepositoryIdentity{Kind: "network-git", Value: "github.com/example/tools"}, ObjectFormat: "sha1", Commit: strings.Repeat("a", 40),
		BuildSource: &buildsource.Identity{Algorithm: buildsource.Algorithm, ContentSHA256: "sha256:" + strings.Repeat("b", 64)}, DescriptorTarget: "external",
		CacheKey: buildmeta.CacheKey("sha256:" + strings.Repeat("4", 64)), ReceiptSHA256: buildmeta.ReceiptHash("sha256:" + strings.Repeat("5", 64)),
		ArtifactSHA256: "sha256:" + strings.Repeat("6", 64), ArtifactPath: "bin/external"}
}

func TestMarkerV3StructurallyRepresentsLocalExternalAndMixed(t *testing.T) {
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
			m := v3Base()
			testCase.set(m)
			if err := Write(dir, m); err != nil {
				t.Fatal(err)
			}
			if m.SchemaVersion != ExternalSchemaVersion {
				t.Fatalf("schema=%d", m.SchemaVersion)
			}
			if got := Read(dir); got == nil || len(got.Builds) != len(m.Builds) {
				t.Fatalf("round trip=%+v", got)
			}
			payload, _ := os.ReadFile(filepath.Join(dir, Name))
			var raw map[string]any
			if err := json.Unmarshal(payload, &raw); err != nil {
				t.Fatal(err)
			}
			for command, value := range raw["builds"].(map[string]any) {
				record := value.(map[string]any)
				if command == "local" && record["receipt_schema_version"] != float64(1) {
					t.Fatalf("local receipt=%v", record)
				}
				if command == "external" && record["receipt_schema_version"] != float64(2) {
					t.Fatalf("external receipt=%v", record)
				}
			}
		})
	}
}

func TestMarkerV3RejectsReceiptInterpretationAliasing(t *testing.T) {
	dir := t.TempDir()
	m := v3Base()
	aliased := externalV3Build()
	aliased.Driver = buildmeta.DriverGoV1
	m.Builds["external"] = aliased
	if err := Write(dir, m); err == nil {
		t.Fatal("external receipt-v2 state aliased as local go-v1")
	}
	m = v3Base()
	aliased = localV3Build()
	aliased.Driver = "go-repository-v1"
	m.Builds["local"] = aliased
	if err := Write(dir, m); err == nil {
		t.Fatal("local receipt-v1 state aliased as external")
	}
}
