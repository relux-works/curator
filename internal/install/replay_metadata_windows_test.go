//go:build windows

package install

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/snapshot"
	"github.com/relux-works/curator/internal/sourcelock"
	"github.com/relux-works/curator/internal/testgitenv"
)

// Fresh production replay at roots 132/133 succeeds with exact locked bytes.
// At 134, the acquisition helper preserves MetadataPathError and Project
// reports it without fetching, publishing a snapshot, changing the lock or
// even creating source-replay. The second endpoint cannot retry a local bound.
func TestWindowsSourceReplayMetadataBoundary(t *testing.T) {
	project, _, _, remote, plan := setupDraftGitSource(t, `{"repository":"example.org/kit","tag":"v1"}`)
	member, ok := plan.Lock.Find("review")
	if !ok {
		t.Fatal("missing locked review package")
	}
	before, err := os.ReadFile(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	bin := testgitenv.ShortTempDir(t)
	git, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	wrapper := filepath.Join(bin, "git.exe")
	cmd := exec.Command("go", "build", "-o", wrapper, "./testdata/replaygit")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build real-Git transport: %v: %s", err, output)
	}
	contacts := filepath.Join(bin, "contacts")
	route, err := json.Marshal(map[string]string{"Git": git, "Local": draftReplayFileURL(remote), "Contacts": contacts})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "route.json"), route, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	for _, length := range []int{132, 133, 134} {
		t.Run(fmt.Sprint(length), func(t *testing.T) {
			base := testgitenv.ShortTempDir(t)
			// source-replay separator/name/separator/digest contributes 35.
			home := filepath.Join(base, strings.Repeat("h", length-35-len(utf16.Encode([]rune(base)))-1))
			if err := os.Mkdir(home, 0o700); err != nil {
				t.Fatal(err)
			}
			policy := `{"schema_version":1,"repositories":{"example.org/kit":{"endpoints":[{"url":"https://example.org/kit.git","authentication":"team-https"},{"url":"git@example.org:kit.git","authentication":"team-ssh"}],"fallback":"availability-auth"}}}`
			if err := os.WriteFile(filepath.Join(home, config.SourcePolicyFileName), []byte(policy), 0o600); err != nil {
				t.Fatal(err)
			}
			cfg := draftTestConfig(home, testgitenv.ShortTempDir(t))
			if err := os.Remove(contacts); err != nil && !os.IsNotExist(err) {
				t.Fatal(err)
			}
			if length > 133 {
				source := draftReplaySource{Alias: "s"}
				source.Source.Repository = "example.org/kit"
				repo, err := fetchLockedCommitFromDeclaredSource(cfg, project, source, member)
				var refusal *gitops.MetadataPathError
				if repo != "" || !errors.As(err, &refusal) || refusal.RootLength != length || refusal.Actual != 260 || refusal.Limit != 259 || refusal.RootLimit != 133 || !strings.HasPrefix(err.Error(), "git_metadata_path_too_long:") {
					t.Fatalf("replay lost typed metadata refusal: %q %v", repo, err)
				}
			}
			result := Project(cfg, project, "test", Options{DryRun: true, Platform: installPlatform()})
			contactBytes, contactErr := os.ReadFile(contacts)
			if length > 133 {
				if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "git_metadata_path_too_long:") || !os.IsNotExist(contactErr) {
					t.Fatalf("Project replay lost pre-contact metadata refusal: %+v contacts=%q error=%v", result, contactBytes, contactErr)
				}
				if _, err := os.Lstat(filepath.Join(home, "source-replay")); !os.IsNotExist(err) {
					t.Fatalf("replay refusal created cache parent: %v", err)
				}
				entries, err := os.ReadDir(home)
				if err != nil || len(entries) != 1 || entries[0].Name() != config.SourcePolicyFileName {
					t.Fatalf("replay refusal changed home: %v %v", entries, err)
				}
			} else {
				if result.Status != "ok" || contactErr != nil || strings.Count(string(contactBytes), "contact") != 1 {
					t.Fatalf("admitted fresh Project replay failed: %+v contacts=%q error=%v", result, contactBytes, contactErr)
				}
				tree := snapshot.Dir(home, member.Package.Repository, member.Package.Commit.Hex)
				if body, err := os.ReadFile(filepath.Join(tree, "skills", "review", "references", "info.md")); err != nil || string(body) != "context" {
					t.Fatalf("admitted replay changed locked bytes: %q %v", body, err)
				}
			}
			after, err := os.ReadFile(sourcelock.PathIn(project))
			if err != nil || !bytes.Equal(before, after) {
				t.Fatalf("replay changed portable lock: %v", err)
			}
		})
	}
}
