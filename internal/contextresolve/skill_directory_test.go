package contextresolve

import (
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
)

// Resolve itself must reject conflicting selectors before acquisition/audit
// can mask the agreement error. An agreeing diamond produces one member.
func TestSkillDirectoryAgreement(t *testing.T) {
	for _, tc := range []struct {
		name, directory, source string
		valid                   bool
	}{
		{"agreeing", "skills/formatter", "example.com/skills", true},
		{"omitted-selector", "", "example.com/skills", false},
		{"different-selector", "skills/alternate", "example.com/skills", false},
		{"different-source", "skills/formatter", "example.com/other", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rootID, aID, bID, skillID := strings.Repeat("1", 40), strings.Repeat("2", 40), strings.Repeat("3", 40), strings.Repeat("4", 40)
			first := Requirement{Kind: contextlock.KindSkill, Name: "formatter", Source: "example.com/skills", Revision: skillID, Directory: "skills/formatter"}
			second := first
			second.Directory = tc.directory
			second.Source = tc.source
			source := &stubSource{manifests: map[string]*Package{
				rootID: {SchemaVersion: 2, Version: "1.0.0", Requires: []Requirement{{Kind: contextlock.KindContext, Name: "consumer-a", Source: "example.com/a", Revision: aID}, {Kind: contextlock.KindContext, Name: "consumer-b", Source: "example.com/b", Revision: bID}}},
				aID:    {SchemaVersion: 2, Version: "1.0.0", Requires: []Requirement{first}}, bID: {SchemaVersion: 2, Version: "1.0.0", Requires: []Requirement{second}}, skillID: {},
			}}
			result, err := Resolve(source, Input{Root: Requirement{Kind: contextlock.KindContext, Name: "root", Source: "example.com/root", Revision: rootID}, Precedence: contextlock.Precedence{Winner: "lower-weight", Placement: "winner-first"}})
			if !tc.valid {
				if err == nil || !strings.Contains(err.Error(), DiagRangeConflict) {
					t.Fatalf("want context_range_conflict, got %v", err)
				}
				for _, part := range []string{"consumer-a", "consumer-b", "source", "directory"} {
					if !strings.Contains(err.Error(), part) {
						t.Fatalf("conflict omits %s: %v", part, err)
					}
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			member, ok := result.Lock.Find(contextlock.KindSkill, "formatter")
			if !ok || member.Directory != "skills/formatter" || len(member.RequiredBy) != 2 || len(result.Lock.Members) != 4 {
				t.Fatalf("diamond did not unify: %+v", result.Lock)
			}
			if result.Lock.SchemaVersion != 4 || result.Lock.Precedence.Winner != "lower-weight" || result.Lock.Precedence.Placement != "winner-first" {
				t.Fatalf("lock lost actual precedence: %+v", result.Lock)
			}
		})
	}
}
