package envprofile

import (
	"fmt"
	"github.com/relux-works/curator/internal/envregistry"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func plantHomePrecedenceRepositoryLoss(t *testing.T, repo, plant string) {
	t.Helper()
	git := filepath.Join(repo, ".git")
	var err error
	switch plant {
	case "git-head":
		err = os.RemoveAll(filepath.Join(git, "HEAD"))
	case "objects-dir":
		err = os.RemoveAll(filepath.Join(git, "objects"))
	case "head-garbage":
		err = os.WriteFile(filepath.Join(git, "HEAD"), []byte("not a ref\n"), 0600)
	case "refs-dir":
		err = os.RemoveAll(filepath.Join(git, "refs"))
	case "git-as-file":
		if err = os.RemoveAll(git); err == nil {
			err = os.WriteFile(git, []byte("gitdir: /nonexistent-fixture-gitdir\n"), 0600)
		}
	default:
		t.Fatalf("unknown plant %s", plant)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func resolveHomePrecedenceRepositoryLoss(t *testing.T, plant, drift, mode string, level int) (string, bool, bool) {
	fx := storeBoundaryFixture(t, "current", true)
	req := fx.request(envregistry.ClaudeCode)
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	req.Policy.StartupVerificationLevel = level
	req.Repair = false
	if r, err := Resolve(req); err != nil || r == nil || len(r.Document) == 0 {
		t.Fatalf("intact neighbor did not launch: %v", err)
	}
	member, entry := storeBoundaryRootMemberFor(fx)
	repo := newGitManager(fx.home).repoDir(member.Source)
	if drift == "home" || drift == "combined" {
		path := filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.ClaudeCode), "CLAUDE.md")
		if err := os.WriteFile(path, []byte("home drift\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if drift == "source" || drift == "combined" {
		if err := os.WriteFile(filepath.Join(entry, "context", "s.md"), []byte("source drift\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if plant != "none" {
		plantHomePrecedenceRepositoryLoss(t, repo, plant)
	}
	before := hashTreeForTest(t, fx.home)
	req.Repair, req.DryRun = mode != "observe", mode == "dry-run"
	r, err := Resolve(req)
	launched := err == nil && r != nil && len(r.Document) != 0
	msg := "LAUNCH"
	if err != nil {
		msg = err.Error()
	} else if r != nil && len(r.Warnings) > 0 {
		msg = "LAUNCH warn0=" + r.Warnings[0]
	}
	wrote := hashTreeForTest(t, fx.home) != before
	return msg, launched, wrote
}

// Negative (reviewer-authored) on the changed refusal ordering: home-first
// precedence must be caused by home drift, never by the repository failure.
// With source-only or no drift, a broken repository must refuse as pin_hash
// and must NOT be reported as environment_home_stale. Positive control in the
// same case: the identical plant plus home drift does lead with home_stale.
// Limitation: Claude Code adapter, git-head and objects-dir plants, L0 and L3.
func TestHomeFirstIsNotBorrowedByRepositoryLoss(t *testing.T) {
	for _, plant := range []string{"git-head", "objects-dir"} {
		for _, level := range []int{0, 3} {
			for _, mode := range []string{"observe", "repair"} {
				t.Run(fmt.Sprintf("%s/L%d/%s", plant, level, mode), func(t *testing.T) {
					t.Parallel()
					for _, drift := range []string{"none", "source"} {
						msg, launched, wrote := resolveHomePrecedenceRepositoryLoss(t, plant, drift, mode, level)
						if wrote {
							t.Fatal("repository-loss refusal wrote state")
						}
						if launched || strings.HasPrefix(msg, DiagHomeStale+":") || strings.Contains(msg, DiagHomeStale) {
							t.Fatalf("%s drift: repository loss borrowed home precedence: %s", drift, msg)
						}
						if !strings.Contains(msg, "pin_hash") {
							t.Fatalf("%s drift: lost pin class: %s", drift, msg)
						}
					}
					msg, launched, wrote := resolveHomePrecedenceRepositoryLoss(t, plant, "home", mode, level)
					if wrote {
						t.Fatal("home-first positive control wrote state")
					}
					if launched || !strings.HasPrefix(msg, DiagHomeStale+":") || !strings.Contains(msg, "pin_hash") {
						t.Fatalf("positive control: home drift plus repository loss must lead home-first and keep pin_hash: %s", msg)
					}
				})
			}
		}
	}
}
