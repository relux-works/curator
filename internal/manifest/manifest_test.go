package manifest

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/verr"
)

func writeManifest(t *testing.T, text string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, Name), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func mustFail(t *testing.T, dir, wantPathPrefix string) {
	t.Helper()
	_, err := Load(dir)
	if err == nil {
		t.Fatalf("expected error with path prefix %q", wantPathPrefix)
	}
	var v *verr.Error
	if !errors.As(err, &v) {
		t.Fatalf("not a validation error: %v", err)
	}
	if !strings.HasPrefix(v.Path, wantPathPrefix) {
		t.Fatalf("error path %q does not start with %q (%s)", v.Path, wantPathPrefix, v.Message)
	}
}

func TestMissingManifestIsNil(t *testing.T) {
	m, err := Load(t.TempDir())
	if err != nil || m != nil {
		t.Fatalf("Load = %v, %v; want nil, nil", m, err)
	}
}

func TestParseFull(t *testing.T) {
	dir := writeManifest(t, `{
		"schema_version": 1,
		"project": {"alias": "Demo iOS"},
		"agents": ["claude_code", "codex_cli"],
		"locale": "ru",
		"skills": [
			{"name": "skill-a", "git": "git@example.com:skills/skill-a.git", "tag": "v1.0.0"},
			{"name": "skill-b", "source": "références/文書", "branch": "main"},
			{"name": "skill-c", "revision": "abc123"}
		]
	}`)
	m, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.ProjectAlias != "Demo iOS" || m.Locale != "ru" || len(m.Agents) != 2 {
		t.Fatalf("manifest: %+v", m)
	}
	if m.Skills[0].Ref.Kind != "tag" || m.Skills[1].Source != "références/文書" || m.Skills[2].Ref.Kind != "revision" {
		t.Fatalf("skills: %+v", m.Skills)
	}
	if m.Skills[2].Source != "skill-c" {
		t.Fatalf("source default: %+v", m.Skills[2])
	}
}

func TestParseRejections(t *testing.T) {
	cases := []struct {
		name string
		text string
		path string
	}{
		{"no schema", `{"skills": []}`, "schema_version"},
		{"schema string", `{"schema_version": "1", "skills": []}`, "schema_version"},
		{"schema future", `{"schema_version": 2, "skills": []}`, "schema_version"},
		{"skills missing", `{"schema_version": 1}`, "skills"},
		{"agents type", `{"schema_version": 1, "agents": "all", "skills": []}`, "agents"},
		{"locale type", `{"schema_version": 1, "locale": 5, "skills": []}`, "locale"},
		{"locale path", `{"schema_version": 1, "locale": "../en", "skills": []}`, "locale"},
		{"project type", `{"schema_version": 1, "project": "x", "skills": []}`, "project"},
		{"alias empty", `{"schema_version": 1, "project": {"alias": ""}, "skills": []}`, "project.alias"},
		{"alias control", `{"schema_version": 1, "project": {"alias": "bad\u0001"}, "skills": []}`, "project.alias"},
		{"project unknown", `{"schema_version": 1, "project": {"alias": "x", "extra": true}, "skills": []}`, "project"},
		{"top unknown", `{"schema_version": 1, "extra": true, "skills": []}`, "Skillfile"},
		{"decl no name", `{"schema_version": 1, "skills": [{"tag": "v1"}]}`, "skills[0]"},
		{"bad name", `{"schema_version": 1, "skills": [{"name": "-x", "tag": "v1"}]}`, "skills[0].name"},
		{"dup name", `{"schema_version": 1, "skills": [{"name": "a", "tag": "v1"}, {"name": "a", "tag": "v2"}]}`, "skills[1].name"},
		{"no ref", `{"schema_version": 1, "skills": [{"name": "a"}]}`, "skills[0]"},
		{"two refs", `{"schema_version": 1, "skills": [{"name": "a", "tag": "v1", "branch": "main"}]}`, "skills[0]"},
		{"empty ref", `{"schema_version": 1, "skills": [{"name": "a", "tag": ""}]}`, "skills[0].tag"},
		{"bad source", `{"schema_version": 1, "skills": [{"name": "a", "source": "../up", "tag": "v1"}]}`, "skills[0].source"},
		{"empty git", `{"schema_version": 1, "skills": [{"name": "a", "git": "", "tag": "v1"}]}`, "skills[0].git"},
		{"decl unknown", `{"schema_version": 1, "skills": [{"name": "a", "tag": "v1", "extra": true}]}`, "skills[0]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { mustFail(t, writeManifest(t, tc.text), tc.path) })
	}
}

func TestLoadRejectsDuplicateJSONKeys(t *testing.T) {
	dir := writeManifest(t, `{"schema_version":1,"schema_version":1,"skills":[]}`)
	if _, err := Load(dir); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("Load duplicate keys = %v, want duplicate-key error", err)
	}
}

func TestEnsureEmptyAndEditing(t *testing.T) {
	dir := t.TempDir()
	if _, err := EnsureEmpty(dir); err != nil {
		t.Fatal(err)
	}
	if err := AddDecl(dir, "skill-a", "tag", "v1.0.0", "git@example.com:skills/skill-a.git", ""); err != nil {
		t.Fatal(err)
	}
	// replace the same name
	if err := AddDecl(dir, "skill-a", "tag", "v1.1.0", "git@example.com:skills/skill-a.git", ""); err != nil {
		t.Fatal(err)
	}
	m, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Skills) != 1 || m.Skills[0].Ref.Value != "v1.1.0" {
		t.Fatalf("after add: %+v", m.Skills)
	}
	if err := RemoveDecl(dir, "skill-a"); err != nil {
		t.Fatal(err)
	}
	if err := RemoveDecl(dir, "skill-a"); err == nil {
		t.Fatal("removing an absent skill must fail")
	}
	// invalid edit is rejected before write
	if err := AddDecl(dir, "-bad", "tag", "v1", "", ""); err == nil {
		t.Fatal("invalid declaration must not be written")
	}
	m, _ = Load(dir)
	if len(m.Skills) != 0 {
		t.Fatalf("manifest polluted by rejected edit: %+v", m.Skills)
	}
}
