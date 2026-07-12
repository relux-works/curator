package marker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/hashing"
)

func TestWriteUsesWireCompatibleEmptyValues(t *testing.T) {
	dir := t.TempDir()
	m := &Marker{Name: "skill-a", Source: "skill-a", RefKind: "tag", Ref: "v1", Commit: "abc"}
	if err := Write(dir, m); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(filepath.Join(dir, Name))
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil {
		t.Fatal(err)
	}
	if localeValue, present := object["locale"]; !present || localeValue != nil {
		t.Fatalf("locale = %#v, present=%v; want explicit null", localeValue, present)
	}
	for _, field := range []string{"agents", "commands", "dependencies", "runtime_roots", "files"} {
		value, ok := object[field].([]any)
		if !ok || len(value) != 0 {
			t.Fatalf("%s = %#v; want []", field, object[field])
		}
	}
}

func install(t *testing.T) (string, *Marker) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	hash, err := hashing.ContentSHA256(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	m := &Marker{
		Name: "skill-a", Source: "skill-a",
		RefKind: "tag", Ref: "v1", Commit: "abc",
		ContentSHA256: hash, Locale: "ru",
		Agents:     []string{"claude_code"},
		Activation: &Activation{Context: true, Commands: []string{"x"}},
	}
	if err := Write(dir, m); err != nil {
		t.Fatal(err)
	}
	return dir, m
}

func TestCurrentRoundTrip(t *testing.T) {
	dir, m := install(t)
	current, err := Current(dir, m)
	if err != nil || !current {
		t.Fatalf("Current = %v, %v; want true", current, err)
	}
}

func TestTamperDetectedByRehash(t *testing.T) {
	dir, m := install(t)
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("edited locally"), 0o644); err != nil {
		t.Fatal(err)
	}
	current, err := Current(dir, m)
	if err != nil {
		t.Fatal(err)
	}
	if current {
		t.Fatal("local edit must invalidate the installation")
	}
}

func TestDriftFieldsInvalidate(t *testing.T) {
	dir, m := install(t)
	cases := []func(x Marker) Marker{
		func(x Marker) Marker { x.Commit = "def"; return x },
		func(x Marker) Marker { x.Ref = "v2"; return x },
		func(x Marker) Marker { x.Locale = "en"; return x },
		func(x Marker) Marker { x.Agents = []string{"cursor"}; return x },
		func(x Marker) Marker { x.Activation = &Activation{Context: false}; return x },
		func(x Marker) Marker { x.Substituted = "path /x"; return x },
		func(x Marker) Marker {
			x.Attestation = &Attestation{Registry: "corp", Status: "audited"}
			return x
		},
	}
	for index, mutate := range cases {
		expected := mutate(*m)
		current, err := Current(dir, &expected)
		if err != nil {
			t.Fatal(err)
		}
		if current {
			t.Fatalf("case %d: drift must invalidate", index)
		}
	}
}

func TestActivationOrderInsensitive(t *testing.T) {
	dir, m := install(t)
	expected := *m
	expected.Activation = &Activation{Context: true, Commands: []string{"x"}}
	current, err := Current(dir, &expected)
	if err != nil || !current {
		t.Fatalf("Current = %v, %v", current, err)
	}
}

func TestUnsupportedSchemaErrors(t *testing.T) {
	dir, m := install(t)
	m.SchemaVersion = 99
	payload := `{"schema_version": 99, "name": "skill-a"}`
	if err := os.WriteFile(filepath.Join(dir, Name), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Current(dir, m); err == nil {
		t.Fatal("unsupported marker schema must error")
	}
}

func TestReplaceDirSwapsAndCleans(t *testing.T) {
	parent := t.TempDir()
	target := filepath.Join(parent, "skill-a")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "old.md"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	staged := filepath.Join(parent, ".skill-a.tmp")
	if err := os.MkdirAll(staged, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staged, "new.md"), []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ReplaceDir(staged, target); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(target, "new.md")); err != nil {
		t.Fatal("new content missing after replace")
	}
	if _, err := os.Stat(filepath.Join(target, "old.md")); err == nil {
		t.Fatal("old content survived")
	}
	entries, _ := os.ReadDir(parent)
	for _, entry := range entries {
		if entry.Name() != "skill-a" {
			t.Fatalf("leftover entry: %s", entry.Name())
		}
	}
}

func TestReadAbsentAndCorrupt(t *testing.T) {
	if Read(t.TempDir()) != nil {
		t.Fatal("absent marker must read nil")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, Name), []byte("{broken"), 0o644); err != nil {
		t.Fatal(err)
	}
	if Read(dir) != nil {
		t.Fatal("corrupt marker must read nil")
	}
}
