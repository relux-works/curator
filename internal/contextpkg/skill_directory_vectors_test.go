package contextpkg_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextpkg"
)

// The 56 published structural vectors exercise the actual strict readers,
// including the frozen schema-1 refusals. They do not prove lifecycle behavior.
func TestSkillDirectorySchemaVectors(t *testing.T) {
	payload, err := os.ReadFile("testdata/profile-skill-directories-v1/schema-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name, Schema string
		Instance     json.RawMessage
		Valid        bool
	}
	if err := json.Unmarshal(payload, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 56 {
		t.Fatalf("vector count = %d", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var err error
			if strings.Contains(tc.Schema, "agent-context") {
				_, err = contextpkg.ParseManifest(tc.Instance)
			} else {
				_, err = contextlock.Parse(tc.Instance)
			}
			if (err == nil) != tc.Valid {
				t.Fatalf("valid=%v, reader error=%v", tc.Valid, err)
			}
		})
	}
}

// The independently published canonical bytes and pins change only directory.
func TestSkillDirectoryLockIdentityVectors(t *testing.T) {
	payload, err := os.ReadFile("testdata/profile-skill-directories-v1/lock-identity.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases map[string]struct {
		Lock json.RawMessage
		CCJ1 string
		Hash string `json:"lock_sha256"`
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	delete(raw, "difference")
	payload, _ = json.Marshal(raw)
	if err := json.Unmarshal(payload, &cases); err != nil {
		t.Fatal(err)
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			lock, err := contextlock.Parse(tc.Lock)
			if err != nil {
				t.Fatal(err)
			}
			bytes, err := lock.Canonical()
			if err != nil {
				t.Fatal(err)
			}
			hash, err := lock.Hash()
			if err != nil {
				t.Fatal(err)
			}
			if string(bytes) != tc.CCJ1 || hash != tc.Hash {
				t.Fatalf("canonical or identity mismatch: %s %s", bytes, hash)
			}
		})
	}
}

// Successor readers keep the complete closed carriers; a discriminator cannot
// lend schema-4 fields to frozen locks, or machine settings to packages.
func TestSkillDirectoryClosedSuccessors(t *testing.T) {
	payload, err := os.ReadFile("testdata/profile-skill-directories-v1/fixtures.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Manifest json.RawMessage `json:"profile_manifest"`
		Lock     json.RawMessage `json:"candidate_lock"`
	}
	if err = json.Unmarshal(payload, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		lock bool
		edit func(map[string]any)
	}{
		{"manifest-nonempty-settings", false, func(m map[string]any) { m["codex_settings"] = map[string]any{"settings": map[string]any{}} }},
		{"manifest-null-settings", false, func(m map[string]any) { m["codex_settings"] = nil }},
		{"manifest-schema-one-settings", false, func(m map[string]any) {
			m["schema_version"] = 1
			delete(m, "requires")
			m["codex_settings"] = map[string]any{}
		}},
		{"manifest-unsupported-three", false, func(m map[string]any) { m["schema_version"] = 3 }},
		{"manifest-unknown-top-level", false, func(m map[string]any) { m["unexpected"] = true }},
		{"lock-no-precedence", true, func(m map[string]any) { delete(m, "precedence") }},
		{"lock-no-settings", true, func(m map[string]any) { delete(m, "codex_settings") }},
		{"lock-no-contributions", true, func(m map[string]any) { delete(m, "codex_settings_contributions") }},
		{"lock-null-settings", true, func(m map[string]any) { m["codex_settings"] = nil }},
		{"lock-null-contributions", true, func(m map[string]any) { m["codex_settings_contributions"] = nil }},
		{"lock-nonempty-settings", true, func(m map[string]any) { m["codex_settings"] = map[string]any{"model": "unowned"} }},
		{"lock-nonempty-contributions", true, func(m map[string]any) { m["codex_settings_contributions"] = []any{map[string]any{}} }},
		{"lock-invalid-precedence", true, func(m map[string]any) { m["precedence"].(map[string]any)["winner"] = "unknown" }},
		{"lock-unknown-precedence", true, func(m map[string]any) { m["precedence"].(map[string]any)["extra"] = true }},
		{"lock-wrong-hash-version", true, func(m map[string]any) { m["hash_version"] = 1 }},
		{"lock-draft-three", true, func(m map[string]any) { m["schema_version"] = 3 }},
		{"lock-empty-directory", true, func(m map[string]any) { m["members"].([]any)[1].(map[string]any)["directory"] = "" }},
		{"lock-null-directory", true, func(m map[string]any) { m["members"].([]any)[1].(map[string]any)["directory"] = nil }},
		{"lock-frozen-two-directory", true, func(m map[string]any) {
			m["schema_version"] = 2
			delete(m, "precedence")
			delete(m, "codex_settings")
			delete(m, "codex_settings_contributions")
		}},
		{"lock-frozen-two-null-precedence", true, func(m map[string]any) {
			m["schema_version"] = 2
			m["precedence"] = nil
			delete(m, "codex_settings")
			delete(m, "codex_settings_contributions")
			delete(m["members"].([]any)[1].(map[string]any), "directory")
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := fixture.Manifest
			if tc.lock {
				raw = fixture.Lock
			}
			var m map[string]any
			if err = json.Unmarshal(raw, &m); err != nil {
				t.Fatal(err)
			}
			tc.edit(m)
			raw, err = json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			if tc.lock {
				_, err = contextlock.Parse(raw)
			} else {
				_, err = contextpkg.ParseManifest(raw)
			}
			if err == nil {
				t.Fatal("closed carrier admitted forbidden mutation")
			}
		})
	}
	// Nearby valid controls exercise both empty settings and frozen lock 2.
	var m map[string]any
	_ = json.Unmarshal(fixture.Manifest, &m)
	m["codex_settings"] = map[string]any{}
	raw, _ := json.Marshal(m)
	if _, err = contextpkg.ParseManifest(raw); err != nil {
		t.Fatal(err)
	}
	m = nil
	_ = json.Unmarshal(fixture.Lock, &m)
	m["schema_version"] = 2
	delete(m, "precedence")
	delete(m, "codex_settings")
	delete(m, "codex_settings_contributions")
	delete(m["members"].([]any)[1].(map[string]any), "directory")
	raw, _ = json.Marshal(m)
	if _, err = contextlock.Parse(raw); err != nil {
		t.Fatal(err)
	}
}
