package devsub

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSchema2BuildRepositorySubstitutions(t *testing.T) {
	payload := []byte(`{"schema_version":2,"substitutions":{},"build_repository_substitutions":{"golden-skill":{"local":{"path":"tools/tmp/../golden"},"network":{"git":"ssh://git@git.example.com/skills/tools.git","ref":{"kind":"branch","value":"release/v2"}}}}}`)
	manifest, err := ParseManifestBytes(payload, "/project")
	if err != nil {
		t.Fatal(err)
	}
	if manifest.SchemaVersion != 2 || len(manifest.Substitutions) != 0 {
		t.Fatalf("manifest = %+v", manifest)
	}
	local := manifest.BuildRepositorySubstitutions["golden-skill"]["local"]
	wantLocalPath := filepath.Clean(filepath.Join("/project", "tools/golden"))
	if local.Selector != "tools/golden" || local.Path != wantLocalPath {
		t.Fatalf("local substitution = %+v", local)
	}
	network := manifest.BuildRepositorySubstitutions["golden-skill"]["network"]
	if network.Identity != "git.example.com/skills/tools" || network.Transport != "ssh" || network.RefKind != "branch" || network.RefValue != "release/v2" {
		t.Fatalf("network substitution = %+v", network)
	}
}

func TestReleasedSchema2Cases(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	directory := filepath.Join(root, "schema-cases", "skillfile-dev-v2")
	entries, err := os.ReadDir(directory)
	if os.IsNotExist(err) {
		t.Skipf("supplied conformance root publishes no schema-cases/skillfile-dev-v2: %v", err)
	}
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			payload, err := os.ReadFile(filepath.Join(directory, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			_, err = ParseManifestBytes(payload, "/project")
			wantValid := strings.HasPrefix(entry.Name(), "valid")
			if (err == nil) != wantValid {
				t.Fatalf("valid=%v, error=%v", wantValid, err)
			}
		})
	}
}

func TestSchema2BuildRepositorySubstitutionClosedShape(t *testing.T) {
	base := `{"schema_version":2,"substitutions":{},"build_repository_substitutions":{"skill":{"repo":%s}}}`
	invalid := []string{
		`{"path":"repo","git":"https://git.example.com/repo"}`,
		`{"path":"repo","ref":{"kind":"branch","value":"main"}}`,
		`{"git":"https://git.example.com/repo","ref":"main"}`,
		`{"git":"https://git.example.com/repo","ref":{"kind":"revision","value":"deadbeef"}}`,
		`{"git":"https://git.example.com/repo","ref":{"kind":"tag","value":"bad tag"}}`,
		`{"git":"https://git.example.com/repo","ref":{"kind":"branch","value":"main"},"credentials":{}}`,
		`{"git":"https://git.example.com/repo","ref":{"kind":"branch","value":"main"},"command":"tool"}`,
		`{"git":"https://git.example.com/repo","ref":{"kind":"branch","value":"main"},"driver":"go-v1"}`,
		`{"git":"https://git.example.com/repo","ref":{"kind":"branch","value":"main"},"target":"tool"}`,
		`{"git":"https://git.example.com/repo","ref":{"kind":"branch","value":"main"},"output":"bin/tool"}`,
	}
	for _, entry := range invalid {
		if _, err := ParseManifestBytes([]byte(strings.Replace(base, "%s", entry, 1)), "/project"); err == nil {
			t.Errorf("accepted invalid substitution %s", entry)
		}
	}
}

func TestSchema2LocalRepositoryPathIsBoundedPortableSelector(t *testing.T) {
	base := `{"schema_version":2,"substitutions":{},"build_repository_substitutions":{"skill":{"repo":{"path":"%s"}}}}`
	for _, selector := range []string{
		`C:\\absolute\\repo`,
		`C:/absolute/repo`,
		`C:relative/repo`,
		`tools\\repo`,
		`/absolute/repo`,
		strings.Repeat("a", 8193),
	} {
		payload := []byte(strings.Replace(base, "%s", selector, 1))
		if _, err := ParseManifestBytes(payload, "/project"); err == nil {
			t.Errorf("accepted non-portable or oversized schema-2 path %q", selector)
		}
	}

	selector := strings.Repeat("界", 8192)
	payload := []byte(strings.Replace(base, "%s", selector, 1))
	if _, err := ParseManifestBytes(payload, "/project"); err != nil {
		t.Fatalf("rejected 8192-scalar schema-2 path: %v", err)
	}
}

func TestSchema2OrdinarySubstitutionScalarBounds(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		field    string
		limit    int
		manifest func(string) string
	}{
		{
			name:  "path",
			field: "substitutions.skill.path",
			limit: 8192,
			manifest: func(value string) string {
				return `{"schema_version":2,"substitutions":{"skill":{"path":"` + value + `"}}}`
			},
		},
		{
			name:  "git",
			field: "substitutions.skill.git",
			limit: 8192,
			manifest: func(value string) string {
				return `{"schema_version":2,"substitutions":{"skill":{"git":"` + value + `","ref":{"kind":"branch","value":"main"}}}}`
			},
		},
		{
			name:  "ref value",
			field: "substitutions.skill.ref.value",
			limit: 1024,
			manifest: func(value string) string {
				return `{"schema_version":2,"substitutions":{"skill":{"git":"repo","ref":{"kind":"branch","value":"` + value + `"}}}}`
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			atLimit := strings.Repeat("界", testCase.limit)
			if _, err := ParseManifestBytes([]byte(testCase.manifest(atLimit)), "/project"); err != nil {
				t.Fatalf("rejected %d-scalar %s: %v", testCase.limit, testCase.field, err)
			}
			overLimit := strings.Repeat("界", testCase.limit+1)
			if _, err := ParseManifestBytes([]byte(testCase.manifest(overLimit)), "/project"); err == nil {
				t.Fatalf("accepted %d-scalar %s", testCase.limit+1, testCase.field)
			}
		})
	}
}

func TestSchema2StructuredRevisionWidths(t *testing.T) {
	base := `{"schema_version":2,"substitutions":{},"build_repository_substitutions":{"skill":{"repo":{"git":"https://git.example.com/repo","ref":{"kind":"revision","value":"%s"}}}}}`
	for _, revision := range []string{strings.Repeat("a", 40), strings.Repeat("b", 64)} {
		if _, err := ParseManifestBytes([]byte(strings.Replace(base, "%s", revision, 1)), "/project"); err != nil {
			t.Errorf("valid revision width rejected: %v", err)
		}
	}
	for _, revision := range []string{strings.Repeat("a", 39), strings.Repeat("b", 41), strings.Repeat("A", 40)} {
		if _, err := ParseManifestBytes([]byte(strings.Replace(base, "%s", revision, 1)), "/project"); err == nil {
			t.Errorf("invalid revision accepted: %q", revision)
		}
	}
}

func TestSchema1FacadeRemainsCompatible(t *testing.T) {
	payload := []byte(`{"substitutions":{"skill":{"path":"../skill"}}}`)
	legacy, err := ParseBytes(payload, "/project")
	wantLegacyPath := filepath.Clean(filepath.Join("/project", "..", "skill"))
	if err != nil || legacy["skill"].Path != wantLegacyPath {
		t.Fatalf("schema-1 facade = %+v, %v", legacy, err)
	}
	manifest, err := ParseManifestBytes(payload, "/project")
	if err != nil || manifest.SchemaVersion != 1 || len(manifest.BuildRepositorySubstitutions) != 0 {
		t.Fatalf("schema-1 manifest = %+v, %v", manifest, err)
	}
}

func TestSchema1OrdinarySubstitutionBoundsRemainLegacyCompatible(t *testing.T) {
	pathValue := strings.Repeat("界", 8193)
	pathPayload := []byte(`{"substitutions":{"skill":{"path":"` + pathValue + `"}}}`)
	if _, err := ParseManifestBytes(pathPayload, "/project"); err != nil {
		t.Fatalf("schema-1 oversized path compatibility changed: %v", err)
	}

	gitValue := strings.Repeat("界", 8193)
	refValue := strings.Repeat("界", 1025)
	gitPayload := []byte(`{"substitutions":{"skill":{"git":"` + gitValue + `","ref":{"kind":"branch","value":"` + refValue + `"}}}}`)
	if _, err := ParseManifestBytes(gitPayload, "/project"); err != nil {
		t.Fatalf("schema-1 oversized git/ref compatibility changed: %v", err)
	}
}
