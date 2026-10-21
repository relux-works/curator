package envprofile

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/audit"
	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/skillspec"
)

// strictAuditMember must ignore an opposite-capability verdict with the SAME
// package identity. Each member kind runs declared-first and none-first;
// repeating each side also exercises writable reuse through the real Gate.
func TestStrictAuditMemberCapabilityCacheIsolation(t *testing.T) {
	for _, kind := range []string{contextlock.KindSkill, contextlock.KindContext, contextlock.KindMCP} {
		for _, declaredFirst := range []bool{true, false} {
			t.Run(kind+map[bool]string{true: "/skill-first", false: "/non-skill-first"}[declaredFirst], func(t *testing.T) {
				home, entry := t.TempDir(), t.TempDir()
				writeStoreTestFiles(t, entry, map[string][]byte{
					"agent-skill.json": []byte(`{"schema_version":3,"capabilities":{"network":["api.example.com"]}}`),
					"scripts/fetch.sh": []byte("curl https://api.example.com/data\n"),
				})
				resolved := contextresolve.Resolved{Kind: kind, Name: "same", Source: "example.com/skills/same", Commit: strings.Repeat("a", 40)}
				spec, err := skillspec.Load(entry)
				if err != nil {
					t.Fatal(err)
				}
				cfg := &config.Config{Path: filepath.Join(home, "config.json"), Audit: config.Audit{Enabled: true, Mode: "strict", FailOn: "high", Backend: "null"}}
				subject := audit.Subject{Kind: kind, Name: resolved.Name, Source: resolved.Source, Git: resolved.Source, Commit: resolved.Commit, Snapshot: entry, SchemaVersion: 3, HashVersion: resolvedHashVersion(resolved)}
				if kind == contextlock.KindSkill {
					var ok bool
					subject.Package, ok = audit.SkillPackage(resolved.Source, "", resolved.Name, resolved.Commit, resolved.Directory)
					if !ok {
						t.Fatal("fixture package identity invalid")
					}
				}
				for _, declared := range []bool{declaredFirst, !declaredFirst, declaredFirst} {
					for repeat := 0; repeat < 2; repeat++ {
						// For the other declaration, seed through the production Gate while
						// preserving every non-capability input used by strictAuditMember.
						if declared == (kind == contextlock.KindSkill) {
							_, err = strictAuditMember(home, &gitManager{}, resolved, entry, Policy{})
						} else {
							subject.Capabilities = capabilities.ImplicitNone()
							if declared {
								subject.Capabilities = spec.Capabilities
							}
							_, errs := audit.Gate(cfg, []audit.Subject{subject})
							err = nil
							if len(errs) != 0 {
								err = fmt.Errorf("%s", strings.Join(errs, ";"))
							}
						}
						if declared {
							if err != nil {
								t.Fatalf("declared skill refused after another audit: %v", err)
							}
						} else if err == nil || !strings.Contains(err.Error(), "audit.capability.network-undeclared") || !strings.Contains(err.Error(), "contacts api.example.com") {
							t.Fatalf("%s audit reused skill allowance: %v", kind, err)
						}
					}
				}
			})
		}
	}
}
