package envprofile

import (
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextresolve"
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

// A cached skill allowance cannot authorize a context/MCP member; a cached non-skill refusal cannot reject the skill.
func TestStrictAuditMemberCapabilityCacheIsolation(t *testing.T) {
	for _, kind := range []string{contextlock.KindContext, contextlock.KindMCP} {
		for _, skillFirst := range []bool{true, false} {
			t.Run(kind+map[bool]string{true: "/skill-first", false: "/non-skill-first"}[skillFirst], func(t *testing.T) {
				home, entry := t.TempDir(), t.TempDir()
				writeStoreTestFiles(t, entry, map[string][]byte{
					"agent-skill.json": []byte(`{"schema_version":3,"capabilities":{"network":["github.com"]}}`),
					"scripts/fetch.sh": []byte("curl https://github.com/data\n"),
				})
				order := []string{contextlock.KindSkill, kind, contextlock.KindSkill, kind}
				if !skillFirst {
					order = []string{kind, contextlock.KindSkill, kind, contextlock.KindSkill}
				}
				for _, memberKind := range order {
					resolved := contextresolve.Resolved{Kind: memberKind, Name: "same", Source: "same", Commit: strings.Repeat("a", 40)}
					_, err := strictAuditMember(home, &gitManager{}, resolved, entry, Policy{})
					if memberKind == contextlock.KindSkill {
						if err != nil {
							t.Fatalf("declared skill refused after another audit: %v", err)
						}
					} else if err == nil || !strings.Contains(err.Error(), "audit.capability.network-undeclared") || !strings.Contains(err.Error(), "contacts github.com") {
						t.Fatalf("%s audit reused skill allowance: %v", memberKind, err)
					}
				}
			})
		}
	}
}
