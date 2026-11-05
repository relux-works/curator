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
// Foreign origins and cache links outside home refuse after an intact control;
// file/gitlink selections and case/normalization collisions never publish.
func TestSkillPlanDirectorySnapshots(t *testing.T) {
	for _, mode := range []string{"selected", "retained", "discovery-link", "root", "root-link", "selected-link", "ancestor-link", "missing", "selection-file", "selection-submodule", "case-collision", "normalization-collision", "cache-foreign-origin", "cache-outside-home"} {
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
			// Index objects make collision/gitlink witnesses independent of the
			// host filesystem's case folding, normalization and submodule checkout.
			addObject := func(mode, name, content string) {
				t.Helper()
				object := content
				if mode != "160000" {
					cmd := exec.Command("git", "hash-object", "-w", "--stdin")
					cmd.Dir = repo
					cmd.Stdin = strings.NewReader(content)
					out, err := cmd.CombinedOutput()
					if err != nil {
						t.Fatalf("fixture object: %v %s", err, out)
					}
					object = strings.TrimSpace(string(out))
				}
				gitRun(t, repo, "update-index", "--add", "--cacheinfo", mode+","+object+","+name)
			}
			link := ""
			switch mode {
			case "discovery-link", "root-link":
				link = ".claude/skills/memory"
			case "selected-link":
				link = directory + "/linked"
			case "ancestor-link":
				gitRun(t, repo, "rm", "-r", "skills")
				link = "skills"
			case "selection-file", "selection-submodule":
				gitRun(t, repo, "rm", "-r", directory)
				if mode == "selection-file" {
					addObject("100644", directory, "file at selection\n")
				} else {
					addObject("160000", directory, fixtureGitOutput(t, repo, "rev-parse", "HEAD"))
				}
			case "case-collision":
				addObject("100644", prefix+"Case", "first\n")
				addObject("100644", prefix+"case", "second\n")
			case "normalization-collision":
				addObject("100644", prefix+"caf\u00e9", "first\n")
				addObject("100644", prefix+"cafe\u0301", "second\n")
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
			refusal := map[string]string{
				"root-link": "links in git snapshots", "selected-link": "links in git snapshots",
				"ancestor-link": "links in git snapshots", "missing": "does not exist",
				"selection-file": "is not a Git tree", "selection-submodule": "is not a Git tree",
				"case-collision":          "duplicate platform path in git snapshot",
				"normalization-collision": "duplicate platform path in git snapshot",
			}[mode]
			if refusal != "" {
				if err == nil || !strings.Contains(err.Error(), refusal) {
					t.Fatalf("intended snapshot refusal %q: %v", refusal, err)
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
			if strings.HasPrefix(mode, "cache-") {
				// Source binding is required even when every pinned tree byte is
				// intact. The same lock/cache just passed the positive control.
				cache := newSkillPlanGitManager(home).repoDir(member.Source)
				want := "source cache origin differs from declaration"
				if mode == "cache-foreign-origin" {
					gitRun(t, cache, "remote", "set-url", "origin", "https://example.com/foreign")
				} else {
					outside := filepath.Join(t.TempDir(), "cache")
					if err := os.Rename(cache, outside); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink(outside, cache); err != nil {
						t.Fatal(err)
					}
					want = "source cache is unavailable"
				}
				before := snapshotPlanState(t, home)
				failure := validateNamedStorePins(home, result.Lock)
				if failure == nil || failure.Check != diagPinHash || !strings.Contains(failure.Error(), want) {
					t.Fatalf("source-bound cache refusal %q: %v", want, failure)
				}
				if after := snapshotPlanState(t, home); before != after {
					t.Fatal("cache refusal changed lock or store")
				}
				return
			}
			if mode == "retained" {
				// A retained member still uses the legacy route, which must reject
				// changed payload bytes and admit the restored pinned bytes.
				retainedMember, _ := result.Lock.Find("skill", "retained")
				payload := filepath.Join(memberEntryPath(home, retainedMember), "SKILL.md")
				writeGitFile(t, filepath.Dir(payload), "SKILL.md", "tampered retained\n")
				if failure := validateNamedStorePins(home, result.Lock); failure == nil || failure.Check != diagPinHash || failure.Member.Name != "retained" {
					t.Fatalf("retained pin tamper admitted: %v", failure)
				}
				writeGitFile(t, filepath.Dir(payload), "SKILL.md", "# Retained\n")
				if failure := validateNamedStorePins(home, result.Lock); failure != nil {
					t.Fatal("restored retained pin refused", failure)
				}
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
			if directory == "." {
				// A whole-repository pin also rejects a newly injected sibling.
				writeGitFile(t, entry, "injected.txt", "injected sibling\n")
				if failure := validateNamedStorePins(home, result.Lock); failure == nil || failure.Check != diagPinHash {
					t.Fatalf("root sibling pin drift admitted: %v", failure)
				}
				if err := os.Remove(filepath.Join(entry, "injected.txt")); err != nil {
					t.Fatal(err)
				}
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

// Snapshot only the publication surfaces; cache mutation itself is the input.
func snapshotPlanState(t *testing.T, home string) string {
	t.Helper()
	var state strings.Builder
	for _, root := range []string{filepath.Join(home, "profiles"), contextstore.Root(home)} {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			state.WriteString(path)
			if !entry.IsDir() {
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				state.Write(data)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return state.String()
}
