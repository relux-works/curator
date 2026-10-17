package envprofile

import (
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextresolve"
	"strings"
	"testing"
)

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
