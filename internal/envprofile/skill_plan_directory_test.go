package envprofile

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/contextstore"
)

// Selected plans omit siblings and discovery links; root selection remains a
// whole snapshot. Selected/ancestor links and missing selections refuse before
// preparation or publication. Scoped v5 pins reject sibling and payload drift;
// retained members keep verification through their existing legacy cache.
func TestSkillPlanDirectorySnapshots(t *testing.T) {
	for _, mode := range []string{"selected", "retained", "discovery-link", "root", "root-link", "selected-link", "ancestor-link", "missing"} {
		t.Run(mode, func(t *testing.T) {
			home := gitFixtureHome(t)
			ids := newGitIdentities(t)
			key, pub := statusSignerKey(t)
			directory := "skills/memory"
			if strings.HasPrefix(mode, "root") {
				directory = "."
			}
			prefix := ""
			if directory != "." {
				prefix = directory + "/"
			}
			repo := gitRepo(t, map[string]string{
				prefix + "SKILL.md":         "---\nname: memory\ndescription: Memory fixture\n---\n# Memory\n",
				prefix + "agent-skill.json": `{"schema_version":4,"capabilities":{},"commands":{}}`,
				"outside.txt":               "sibling\n",
			}, "v1")
			link := ""
			switch mode {
			case "discovery-link", "root-link":
				link = ".claude/skills/memory"
			case "selected-link":
				link = directory + "/linked"
			case "ancestor-link":
				gitRun(t, repo, "rm", "-r", "skills")
				link = "skills"
			case "missing":
				directory = "skills/missing"
			}
			if link != "" {
				// Write the Git link object directly: no host symlink privilege is needed.
				cmd := exec.Command("git", "hash-object", "-w", "--stdin")
				cmd.Dir = repo
				cmd.Stdin = strings.NewReader("../../skills/memory")
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("link object: %v %s", err, out)
				}
				gitRun(t, repo, "update-index", "--add", "--cacheinfo", "120000,"+strings.TrimSpace(string(out))+","+link)
			}
			gitRun(t, repo, "-c", "gpg.format=ssh", "-c", "user.signingkey="+key, "commit", "--amend", "-S", "--no-edit")
			pin := fixtureGitOutput(t, repo, "rev-parse", "HEAD")
			source := ids.serve(repo, "https://example.com/memory")
			root := t.TempDir()
			writePackage(t, root, "profile", "1.0.0", "fixture\n")
			selector := ""
			if directory != "." {
				selector = fmt.Sprintf(`,"directory":%q`, directory)
			}
			writeGitFile(t, root, "agent-context.json", fmt.Sprintf(`{"schema_version":2,"name":"profile","version":"1.0.0","requires":{"skills":{"memory":{"git":%q,"revision":%q%s}}}}`, source, pin, selector))
			_, state, err := contextstore.EnsureState(home, "context", "profile", root)
			if err != nil {
				t.Fatal(err)
			}
			var retained []contextlock.Member
			if mode == "retained" {
				retainedRepo := gitRepo(t, map[string]string{"SKILL.md": "# Retained\n"}, "v1")
				retainedPin := fixtureGitOutput(t, retainedRepo, "rev-parse", "HEAD")
				ids.serve(retainedRepo, "https://example.com/retained")
				legacyRepo, err := newGitManager(home).ensureRepo("example.com/retained")
				if err != nil {
					t.Fatal(err)
				}
				if _, err := contextstore.EnsureGit(home, "skill", "retained", legacyRepo, retainedPin); err != nil {
					t.Fatal(err)
				}
				retained = []contextlock.Member{{Kind: "skill", Name: "retained", Source: "example.com/retained", Commit: retainedPin, RequiredBy: []string{}}}
			}
			hash, err := contextlock.Write(lockPath(home, "profile"), &contextlock.Lock{SchemaVersion: 2, HashVersion: 2, Root: "profile", Members: append([]contextlock.Member{{Kind: "context", Name: "profile", Version: "1.0.0", StateHash: state, RequiredBy: []string{}}}, retained...)})
			if err != nil {
				t.Fatal(err)
			}
			old, err := os.ReadFile(lockPath(home, "profile"))
			if err != nil {
				t.Fatal(err)
			}
			policy := Policy{RequireSourceSigners: true, SourceSigners: map[string][]contextresolve.Signer{"example.com/memory": {{Type: "ssh", Key: pub}}}}
			calls := 0
			result, err := PrepareSkillPlan(home, "profile", hash, policy, func(nodes []*closure.Node) error {
				calls++
				if len(nodes) != 1 || nodes[0].Directory != directory {
					t.Fatal("wrong selected closure")
				}
				return nil
			})
			refused := mode == "root-link" || mode == "selected-link" || mode == "ancestor-link" || mode == "missing"
			if refused {
				if err == nil || (!strings.Contains(err.Error(), "links in git snapshots") && !strings.Contains(err.Error(), "does not exist")) {
					t.Fatalf("intended snapshot refusal: %v", err)
				}
				after, _ := os.ReadFile(lockPath(home, "profile"))
				if calls != 0 || !bytes.Equal(old, after) {
					t.Fatal("invalid snapshot reached preparation or published lock")
				}
				entries, _ := os.ReadDir(filepath.Join(contextstore.Root(home), "skill", "memory"))
				if len(entries) != 0 {
					t.Fatal("invalid snapshot entered store")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if calls != 1 || result.Lock.SchemaVersion != 5 {
				t.Fatal("missing v5 preparation")
			}
			member, ok := result.Lock.Find("skill", "memory")
			if !ok || member.Directory != directory {
				t.Fatal("lost directory binding")
			}
			selectedDirectory := directory
			if directory == "." {
				selectedDirectory = ""
			}
			entry := contextstore.GitEntryDir(home, "skill", "memory", pin, selectedDirectory)
			if got := memberEntryPath(home, member); got != entry {
				t.Fatalf("pin route %s, want %s", got, entry)
			}
			if failure := validateNamedStorePins(home, result.Lock); failure != nil {
				t.Fatalf("valid v5 pins: %v", failure)
			}
			if directory != "." {
				if _, err := os.Stat(contextstore.EntryDir(home, "skill", "memory", pin)); !os.IsNotExist(err) {
					t.Fatalf("whole snapshot published: %v", err)
				}
				if _, err := os.Stat(filepath.Join(entry, "outside.txt")); !os.IsNotExist(err) {
					t.Fatalf("sibling published: %v", err)
				}
				writeGitFile(t, entry, "outside.txt", "injected sibling\n")
				if failure := validateNamedStorePins(home, result.Lock); failure == nil || failure.Check != diagPinHash {
					t.Fatalf("sibling pin drift admitted: %v", failure)
				}
				if err := os.Remove(filepath.Join(entry, "outside.txt")); err != nil {
					t.Fatal(err)
				}
			} else if raw, err := os.ReadFile(filepath.Join(entry, "outside.txt")); err != nil || string(raw) != "sibling\n" {
				t.Fatal("root snapshot lost sibling", err)
			}
			if failure := validateNamedStorePins(home, result.Lock); failure != nil {
				t.Fatal("restored pin refused", failure)
			}
			payload := filepath.Join(entry, filepath.FromSlash(prefix), "SKILL.md")
			if err := os.WriteFile(payload, []byte("tampered\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if failure := validateNamedStorePins(home, result.Lock); failure == nil || failure.Check != diagPinHash {
				t.Fatalf("selected pin tamper admitted: %v", failure)
			}
		})
	}
}
