package contextlock

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/skillintegrity"
)

func integrityFixture(t *testing.T) *Lock {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "SKILL.md"), []byte("---\nname: fixture-skill\ndescription: Fixture skill.\n---\nFixture context.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Windows mode bits alone do not establish a private skill boundary.
	if runtime.GOOS == "windows" {
		if err := privatedir.ProtectTree(root); err != nil {
			t.Fatal(err)
		}
	}
	member := Member{Kind: KindSkill, Name: "fixture-skill", Source: "https://example.org/skill.git", Commit: strings.Repeat("a", 40), RequiredBy: []string{"root"}}
	e, err := skillintegrity.Build(context.Background(), root, SkillIdentity(member), nil)
	if err != nil {
		t.Fatal(err)
	}
	return &Lock{SchemaVersion: SkillIntegritySchemaVersion, HashVersion: 2, Root: "root", Precedence: Precedence{Winner: "higher-weight", Placement: "winner-last"}, Members: []Member{{Kind: KindContext, Name: "root", Source: "https://example.org/context.git", Version: "1.0.0", Commit: strings.Repeat("b", 40), RequiredBy: []string{}}, member}, SkillIntegrity: []skillintegrity.Entry{e}}
}

// Lock6 roundtrips exact CCJ evidence and a bound profile floor without enabling
// lock5's skill-plan fields. A legacy lock4 keeps its original object shape.
func TestIntegrityLockRoundTrip(t *testing.T) {
	lock := integrityFixture(t)
	floor := 4
	lock.StartupVerificationLevel = &floor
	raw, err := lock.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	again, err := decoded.Canonical()
	if err != nil || string(raw) != string(again) || decoded.StartupVerificationLevel == nil || *decoded.StartupVerificationLevel != 4 {
		t.Fatalf("roundtrip: %s,%v", again, err)
	}
	lock.SchemaVersion = SchemaVersion4
	lock.SkillIntegrity = nil
	lock.StartupVerificationLevel = nil
	raw, err = lock.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "skill_integrity") || strings.Contains(string(raw), "startup_verification_level") {
		t.Fatal("legacy serialization acquired new fields")
	}
}

// Missing/extra evidence, mismatched identity/hash, unsupported versions and
// legacy-carrier field smuggling refuse instead of dropping or defaulting data.
func TestIntegrityLockRefusals(t *testing.T) {
	lock := integrityFixture(t)
	raw, err := lock.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"missing", "null", "extra", "legacy", "legacy-null", "floor-null", "floor-fraction", "floor-string", "plan", "codex-settings", "hash"} {
		t.Run(kind, func(t *testing.T) {
			var object map[string]any
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.UseNumber()
			if err := decoder.Decode(&object); err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "missing":
				delete(object, "skill_integrity")
			case "null":
				object["skill_integrity"] = nil
			case "extra":
				entries := object["skill_integrity"].([]any)
				object["skill_integrity"] = append(entries, entries[0])
			case "legacy":
				object["schema_version"] = 4
			case "legacy-null":
				object["schema_version"] = 4
				object["skill_integrity"] = nil
			case "floor-null":
				object["startup_verification_level"] = nil
			case "floor-fraction":
				object["startup_verification_level"] = json.Number("3.0")
			case "floor-string":
				object["startup_verification_level"] = "3"
			case "plan":
				object["skill_plan"] = nil
			case "codex-settings":
				object["codex_settings"] = map[string]any{"extra": true}
			case "hash":
				entries := object["skill_integrity"].([]any)
				entries[0].(map[string]any)["manifest_sha256"] = "sha256:" + strings.Repeat("0", 64)
			}
			// Preserve deliberately noncanonical number tokens for the reader
			// rejection control; the canonical writer must reject them earlier.
			changed, err := json.Marshal(object)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := Parse(changed); err == nil {
				t.Fatalf("accepted %s", kind)
			}
		})
	}
	lock.SchemaVersion = SchemaVersion4
	if err := lock.Validate(); err == nil {
		t.Fatal("legacy programmatic lock silently drops evidence")
	}
}
