package envprofile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/audit"
	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/hashing"
)

// Production entry: EnsureDefaultWithPolicy -> ensureDefault ->
// migrateGlobalSkills -> strictAuditMember. Only a covering declaration
// admits the contacted host; every refusal leaves both profile files absent.
func TestMigratedGlobalSkillNetworkCapabilities(t *testing.T) {
	previous := hashing.EnableV2Writers
	hashing.EnableV2Writers = true
	t.Cleanup(func() { hashing.EnableV2Writers = previous })
	for _, tc := range []struct {
		name, file, manifest string
		admitted             bool
	}{
		{"exact-host-declared", "agent-skill.json", `{"schema_version":3,"capabilities":{"network":["downloads.example.com"]}}`, true},
		{"glob-host-declared", "agent-skill.json", `{"schema_version":3,"capabilities":{"network":["*.example.com"]}}`, true},
		{"legacy-alias", "csk-skill.json", `{"schema_version":3,"capabilities":{"network":["downloads.example.com"]}}`, true},
		{"schema-nine", "agent-skill.json", `{"schema_version":9,"capabilities":{"network":["downloads.example.com"]}}`, true},
		{"no-manifest", "", "", false},
		{"network-none", "agent-skill.json", `{"schema_version":3,"capabilities":{"network":"none"}}`, false},
		{"network-absent", "agent-skill.json", `{"schema_version":3,"capabilities":{}}`, false},
		{"other-host-declared", "agent-skill.json", `{"schema_version":3,"capabilities":{"network":["api.example.org"]}}`, false},
		{"pre-capabilities-schema", "agent-skill.json", `{"schema_version":2}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := map[string]string{"README.md": "network skill\n", "assets/small.bin": "a\x00b", "scripts/fetch.sh": "curl https://downloads.example.com/archive\n"}
			if tc.file != "" {
				files[tc.file] = tc.manifest
			}
			home, commit := migrationRevisionFixture(t, files)
			err := EnsureDefaultWithPolicy(home, Policy{})
			if tc.admitted {
				if err != nil {
					t.Fatalf("covering declaration refused: %v", err)
				}
				lock, _, err := readLock(home, DefaultProfile)
				if err != nil {
					t.Fatal(err)
				}
				if member, ok := lock.Find(contextlock.KindSkill, "binary-skill"); !ok || member.Commit != commit {
					t.Fatalf("migrated skill = %+v, present = %v", member, ok)
				}
				if _, err := readSource(home, DefaultProfile); err != nil {
					t.Fatal(err)
				}
				return
			}
			assertMigrationNetworkRefusal(t, home, err, "downloads.example.com")
		})
	}
}

// The real migration must neither reuse a refusal produced before it read
// manifest capabilities nor accept an unstated-package allow for undeclared
// network use. Seed each cache through the real audit gate, then drive the
// public migration entry; only the covering manifest may publish.
func TestMigratedGlobalSkillIgnoresPreManifestVerdictCache(t *testing.T) {
	previous := hashing.EnableV2Writers
	hashing.EnableV2Writers = true
	t.Cleanup(func() { hashing.EnableV2Writers = previous })
	for _, declared := range []bool{true, false} {
		name := "undeclared-skill-old-allow"
		if declared {
			name = "declared-skill-old-refusal"
		}
		t.Run(name, func(t *testing.T) {
			files := map[string]string{"README.md": "network skill\n", "scripts/fetch.sh": "curl https://downloads.example.com/a\n"}
			if declared {
				files["agent-skill.json"] = `{"schema_version":3,"capabilities":{"network":["downloads.example.com"]}}`
			}
			home, commit := migrationRevisionFixture(t, files)
			// Populate the immutable entry without publishing a profile.
			_, firstErr := migrateGlobalSkills(home, Policy{})
			if declared && firstErr != nil || !declared && (firstErr == nil || !strings.Contains(firstErr.Error(), "audit.capability.network-undeclared")) {
				t.Fatalf("initial member audit: %v", firstErr)
			}
			assertMigrationUnpublished(t, home)
			// Remove only this isolated fixture's verdict so the following
			// production Gate call computes and persists the old subject.
			verdicts, err := filepath.Glob(filepath.Join(home, "audit", "*", "verdict-*.json"))
			if err != nil || len(verdicts) != 1 {
				t.Fatalf("fixture verdicts = %v, err = %v", verdicts, err)
			}
			if err := os.Remove(verdicts[0]); err != nil {
				t.Fatal(err)
			}
			caps := capabilities.ImplicitNone()
			if !declared {
				caps.Network = []string{"downloads.example.com"}
			}
			cfg := &config.Config{Path: filepath.Join(home, "config.json"), Audit: config.Audit{Enabled: true, Mode: "strict", FailOn: "high", Backend: "null"}}
			_, errs := audit.Gate(cfg, []audit.Subject{{Name: "binary-skill", Commit: commit,
				Snapshot:      contextstore.EntryDir(home, contextlock.KindSkill, "binary-skill", commit),
				SchemaVersion: 3, Capabilities: caps, HashVersion: hashing.WriteVersion()}})
			if declared && (len(errs) == 0 || !strings.Contains(strings.Join(errs, ";"), "audit.capability.network-undeclared")) || !declared && len(errs) != 0 {
				t.Fatalf("old unstated-package cache seed: %v", errs)
			}
			err = EnsureDefaultWithPolicy(home, Policy{})
			if !declared {
				assertMigrationNetworkRefusal(t, home, err, "downloads.example.com")
				return
			}
			if err != nil {
				t.Fatalf("stale implicit-none refusal reused: %v", err)
			}
			payload, err := os.ReadFile(verdicts[0])
			if err != nil {
				t.Fatal(err)
			}
			var record struct {
				Package audit.SourcePackage `json:"package"`
			}
			if err := json.Unmarshal(payload, &record); err != nil {
				t.Fatal(err)
			}
			// Capability equality must not erase the independently required
			// package identity when the migration writes its fresh verdict.
			if record.Package.Kind == "" {
				t.Fatal("migration verdict lost package identity")
			}

			lock, _, err := readLock(home, DefaultProfile)
			if err != nil {
				t.Fatal(err)
			}
			if member, ok := lock.Find(contextlock.KindSkill, "binary-skill"); !ok || member.Commit != commit {
				t.Fatalf("migrated skill = %+v, present = %v", member, ok)
			}
		})
	}
}

func assertMigrationNetworkRefusal(t *testing.T, home string, err error, host string) {
	t.Helper()
	if err == nil {
		t.Fatal("migration admitted undeclared network use")
	}
	if strings.Contains(err.Error(), "require_pin") || strings.Contains(err.Error(), "audit requires pin") {
		t.Fatalf("legacy profile audit hid the network finding behind a pin: %v", err)
	}
	for _, want := range []string{DiagSourceInvalid, `"binary-skill"`, "audit.capability.network-undeclared", "contacts " + host, "scripts/fetch.sh", "capabilities.network", "agent-skill.json", "csk-skill.json", "leave the skill unmigrated"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("migration = %v, missing %q", err, want)
		}
	}
	assertMigrationUnpublished(t, home)
}

// A declaration only covers its named host. Revocation and the canary remain
// independent blocking gates, including when all contacted hosts are declared.
func TestMigratedGlobalSkillDeclarationKeepsOtherGates(t *testing.T) {
	for _, tc := range []struct {
		name, script, want string
		policy             Policy
		failCanary         bool
	}{
		{name: "neighbouring-host", script: "curl https://downloads.example.com/a\ncurl https://uploads.example.com/b\n", want: "contacts uploads.example.com"},
		{name: "revoked-source", script: "curl https://downloads.example.com/a\n", policy: Policy{Revocations: []string{"source:example.com/skills/binary-skill"}}, want: "revoked"},
		{name: "canary-failure", script: "curl https://downloads.example.com/a\n", failCanary: true, want: "canary"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.failCanary {
				previous := canaryPasses
				canaryPasses = func() bool { return false }
				t.Cleanup(func() { canaryPasses = previous })
			}
			home, _ := migrationRevisionFixture(t, map[string]string{"README.md": "network skill\n", "agent-skill.json": `{"schema_version":3,"capabilities":{"network":["downloads.example.com"]}}`, "scripts/fetch.sh": tc.script})
			err := EnsureDefaultWithPolicy(home, tc.policy)
			if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("migration = %v, want refusal carrying %q", err, tc.want)
			}
			if tc.name == "neighbouring-host" {
				assertMigrationNetworkRefusal(t, home, err, "uploads.example.com")
			}
			assertMigrationUnpublished(t, home)
		})
	}
}

// Invalid and unreadable manifests cannot fall back to implicit-none, even
// without network use. A directory at the manifest path is a portable failed
// read witness; this does not simulate an OS permission-denied read.
func TestMigratedGlobalSkillManifestErrorsBlock(t *testing.T) {
	for _, tc := range []struct{ name, file, payload, want string }{
		{"malformed-network", "agent-skill.json", `{"schema_version":3,"capabilities":{"network":7}}`, "capabilities.network"},
		{"malformed-json", "agent-skill.json", `{`, "agent-skill.json"},
		{"unreadable-manifest", "agent-skill.json/entry", "unreadable as a manifest\n", "agent-skill.json"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home, _ := migrationRevisionFixture(t, map[string]string{"README.md": "clean skill\n", tc.file: tc.payload})
			err := EnsureDefaultWithPolicy(home, Policy{})
			if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("migration = %v, want manifest refusal naming %q", err, tc.want)
			}
			assertMigrationUnpublished(t, home)
		})
	}
}

// Production Install audits required skills with their own capabilities.
// A context member ignores skill manifests: malformed manifests do not block a
// clean context, and covering skill declarations cannot authorize its network.
func TestInstalledMemberCapabilitiesStayKindScoped(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		skill, network, admitted bool
	}{
		{"declared-skill", true, true, true},
		{"clean-context-ignores-skill-manifest", false, false, true},
		{"context-network-not-authorized", false, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ids := newGitIdentities(t)
			files := map[string]string{"agent-context.json": `{"schema_version":1,"name":"clean","version":"1.0.0"}`, "README.md": "clean\n"}
			declared := `{"schema_version":3,"capabilities":{"network":["downloads.example.com"]}}`
			if tc.skill {
				skill := gitRepo(t, map[string]string{"agent-skill.json": declared, "scripts/fetch.sh": "curl https://downloads.example.com/a\n"}, "v1.0.0")
				operand := ids.serve(skill, "https://example.com/skills/declared")
				files["agent-context.json"] = `{"schema_version":1,"name":"clean","version":"1.0.0","requires":{"skills":{"declared":{"git":"` + operand + `","range":"*"}}}}`
			} else if tc.network {
				files["agent-skill.json"] = declared
				files["scripts/fetch.sh"] = "curl https://downloads.example.com/a\n"
			} else {
				files["agent-skill.json"] = `{"schema_version":3,"capabilities":{"network":7}}`
			}
			root := gitRepo(t, files, "v1.0.0")
			home := t.TempDir()
			pinHomes(t)
			info, _, _, err := Install(home, InstallOptions{Operand: ids.serve(root, "https://example.com/profiles/clean")})
			if tc.admitted {
				if err != nil {
					t.Fatal(err)
				}
				if tc.skill {
					if _, ok := info.Lock.Find(contextlock.KindSkill, "declared"); !ok {
						t.Fatal("skill missing from installed lock")
					}
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), "audit.capability.network-undeclared") {
					t.Fatalf("context audit = %v, want network refusal", err)
				}
				if _, err := os.Stat(sourcePath(home, "clean")); !os.IsNotExist(err) {
					t.Fatalf("refused context published: %v", err)
				}
			}
		})
	}
}
