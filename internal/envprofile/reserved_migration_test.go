package envprofile

import (
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
)

// The real migration admits declared real hosts plus a reserved fixture, while an undeclared lookalike publishes nothing.
func TestMigratedGlobalSkillReservedFixture(t *testing.T) {
	for _, tc := range []struct {
		name, extra string
	}{
		{name: "declared-hosts-and-reserved-fixture"},
		{name: "undeclared-real-neighbor", extra: "curl https://evil.test.com/data\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home, commit := migrationRevisionFixture(t, map[string]string{
				"README.md":        "skill with an embedded test fixture\n",
				"agent-skill.json": `{"schema_version":3,"capabilities":{"network":["github.com","brew.sh"]}}`,
				"scripts/fetch.sh": "curl https://github.com/data\ncurl https://brew.sh/data\necho https://board.example.test/fixture\n" + tc.extra,
			})
			err := EnsureDefaultWithPolicy(home, Policy{})
			if tc.extra != "" {
				if err == nil || !strings.Contains(err.Error(), "audit.capability.network-undeclared") ||
					!strings.Contains(err.Error(), "contacts evil.test.com") {
					t.Fatalf("migration = %v, want undeclared real-neighbor refusal", err)
				}
				assertMigrationUnpublished(t, home)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			lock, _, err := readLock(home, DefaultProfile)
			if err != nil {
				t.Fatal(err)
			}
			if member, ok := lock.Find(contextlock.KindSkill, "binary-skill"); !ok || member.Commit != commit {
				t.Fatalf("migrated member = %+v, present = %v", member, ok)
			}
		})
	}
}
