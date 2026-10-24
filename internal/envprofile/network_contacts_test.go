package envprofile

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
)

// Every pinned network-contact vector runs through the public Install and
// EnsureDefaultWithPolicy entries. Refusals retain exact host/file evidence
// and publish no profile; successful migration retains the pinned skill.
// Fixture paths are prefixed by scripts/, the current extraction surface;
// all skill fixtures retain the required context entry alongside the corpus.
func TestProfileNetworkContactVectors(t *testing.T) {
	payload, err := os.ReadFile("../audit/testdata/network-contacts.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Cases []struct {
			Name     string
			Files    map[string]string
			Manifest map[string]any
			Expected struct {
				Findings []struct{ Class, Host, File string }
				Passes   bool `json:"capability_check_passes"`
			}
		}
	}
	if err := json.Unmarshal(payload, &corpus); err != nil {
		t.Fatal(err)
	}
	if len(corpus.Cases) != 40 {
		t.Fatalf("vectors=%d, want 40", len(corpus.Cases))
	}
	for _, tc := range corpus.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			files := map[string]string{
				"README.md": "network fixture\n",
				"SKILL.md":  "---\nname: network-fixture\ndescription: Network contact fixture.\n---\n",
			}
			for file, text := range tc.Files {
				files["scripts/"+file] = text
			}
			if tc.Manifest != nil {
				// A manifest without capabilities uses the valid legacy schema;
				// schema 3 requires that member independently of network checks.
				tc.Manifest["schema_version"] = 2
				if _, present := tc.Manifest["capabilities"]; present {
					tc.Manifest["schema_version"] = 3
				}
				manifest, err := json.Marshal(tc.Manifest)
				if err != nil {
					t.Fatal(err)
				}
				files["agent-skill.json"] = string(manifest)
			}
			t.Run("migration", func(t *testing.T) {
				home, commit := migrationRevisionFixture(t, files)
				err := EnsureDefaultWithPolicy(home, Policy{})
				if !tc.Expected.Passes {
					if err == nil {
						t.Fatal("migration admitted undeclared network contact")
					}
					for _, f := range tc.Expected.Findings {
						for _, want := range []string{f.Class, "contacts " + f.Host + ",", "scripts/" + f.File, "capabilities.network", "leave the skill unmigrated"} {
							if !strings.Contains(err.Error(), want) {
								t.Fatalf("migration refusal=%v, missing %q", err, want)
							}
						}
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
				member, ok := lock.Find(contextlock.KindSkill, "binary-skill")
				if !ok || member.Commit != commit {
					t.Fatalf("migrated skill=%+v, found=%v", member, ok)
				}
			})
			t.Run("install", func(t *testing.T) {
				ids := newGitIdentities(t)
				skill := gitRepo(t, files, "v1.0.0")
				operand := ids.serve(skill, "https://example.com/skills/network-fixture")
				profile := gitRepo(t, map[string]string{
					"agent-context.json": `{"schema_version":1,"name":"network-profile","version":"1.0.0","requires":{"skills":{"network-fixture":{"git":"` + operand + `","range":"*"}}}}`,
					"README.md":          "network profile\n",
				}, "v1.0.0")
				home := gitFixtureHome(t)
				pinHomes(t)
				info, _, _, err := Install(home, InstallOptions{Operand: ids.serve(profile, "https://example.com/profiles/network-profile")})
				if !tc.Expected.Passes {
					if err == nil {
						t.Fatal("Install admitted undeclared network contact")
					}
					for _, f := range tc.Expected.Findings {
						for _, want := range []string{f.Class, "contacts " + f.Host + ",", "scripts/" + f.File} {
							if !strings.Contains(err.Error(), want) {
								t.Fatalf("Install refusal=%v, missing %q", err, want)
							}
						}
					}
					if _, err := os.Stat(sourcePath(home, "network-profile")); !os.IsNotExist(err) {
						t.Fatalf("refused install published source: %v", err)
					}
					if _, err := os.Stat(lockPath(home, "network-profile")); !os.IsNotExist(err) {
						t.Fatalf("refused install published lock: %v", err)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if _, ok := info.Lock.Find(contextlock.KindSkill, "network-fixture"); !ok {
					t.Fatal("installed skill missing")
				}
			})
		})
	}
}
