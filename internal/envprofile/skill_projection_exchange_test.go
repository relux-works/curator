//go:build darwin || linux

package envprofile

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/transaction"
)

// A separately running reader, synchronized at every transaction hook during
// each repair, can resolve two shared links and read a retained context file
// without ENOENT. Missing/extra/changed neighboring files are repaired exactly;
// this does not claim snapshot isolation for a reader traversing multiple files.
func TestSkillProjectionRepairNoGap(t *testing.T) {
	for _, damage := range []string{"missing", "extra", "changed"} {
		t.Run(damage, func(t *testing.T) {
			home, source := t.TempDir(), t.TempDir()
			writeGitFile(t, source, "SKILL.md", "# Guide\n")
			writeGitFile(t, source, "references/stable.md", "retained\n")
			skill, err := projectSkill(home, source, hashing.VersionV2)
			if err != nil {
				t.Fatal(err)
			}
			if err := publishTestSkillProjection(t, home, skill); err != nil {
				t.Fatal(err)
			}
			want := projectionFiles(t, skill.root)
			wantDigest, err := transaction.DigestPath(skill.root)
			if err != nil {
				t.Fatal(err)
			}
			links := []string{filepath.Join(home, "first"), filepath.Join(home, "second")}
			for _, link := range links {
				if err := os.Symlink(skill.root, link); err != nil {
					t.Fatal(err)
				}
			}
			switch damage {
			case "missing":
				err = os.Remove(filepath.Join(skill.root, "SKILL.md"))
			case "extra":
				err = os.WriteFile(filepath.Join(skill.root, "extra.md"), []byte("extra\n"), 0600)
			case "changed":
				err = os.WriteFile(filepath.Join(skill.root, "SKILL.md"), []byte("changed\n"), 0600)
			}
			if err != nil {
				t.Fatal(err)
			}
			read := make(chan struct{})
			result := make(chan error)
			stopped := make(chan struct{})
			go func() {
				defer close(stopped)
				for range read {
					var failure error
					for _, link := range links {
						body, err := os.ReadFile(filepath.Join(link, "references/stable.md"))
						if err != nil {
							failure = fmt.Errorf("shared projection unavailable: %w", err)
							break
						}
						if string(body) != "retained\n" {
							failure = fmt.Errorf("retained context bytes changed")
							break
						}
					}
					result <- failure
				}
			}()
			defer func() { close(read); <-stopped }()
			seen := map[transaction.Point]bool{}
			observe := func(event transaction.Event) {
				seen[event.Point] = true
				read <- struct{}{}
				if err := <-result; err != nil {
					t.Errorf("%s: no-gap assertion: %v", event.Point, err)
				}
			}
			op, err := beginOperation(home, transaction.WithHooks(transaction.Hooks{Observe: observe}))
			if err != nil {
				t.Fatal(err)
			}
			err = publishSkillProjection(op, home, skill)
			if closeErr := op.close(); closeErr != nil {
				t.Fatal(closeErr)
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, point := range []transaction.Point{transaction.PointPrepared, transaction.PointAfterBackup, transaction.PointBeforeInstall, transaction.PointAfterInstall, transaction.PointTargetCommitted, transaction.PointBeforeCleanup, transaction.PointAfterCleanupRename, transaction.PointDuringCleanupRemoval} {
				if !seen[point] {
					t.Errorf("reader hook not exercised: %s", point)
				}
			}
			if got := projectionFiles(t, skill.root); !reflect.DeepEqual(got, want) {
				t.Fatal("repair changed pinned bytes or private permissions")
			}
			if got, err := transaction.DigestPath(skill.root); err != nil || got != wantDigest {
				t.Fatalf("repair changed bytes or permissions: %s %v", got, err)
			}
		})
	}
}

// Restoring the former backup-then-install selection must fail all three
// deterministic reader cases specifically with the no-gap ENOENT assertion.
func TestSkillProjectionExchangeNarrowingMutant(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "internal/envprofile/skill_projection_exchange.go")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	const anchor = "const skillProjectionReplacementKind = transaction.KindExchange"
	if strings.Count(string(body), anchor) != 1 {
		t.Fatal("mutation anchor is not unique")
	}
	scratch := t.TempDir()
	replacement := filepath.Join(scratch, "replacement.go")
	if err := os.WriteFile(replacement, []byte(strings.Replace(string(body), anchor, "const skillProjectionReplacementKind = transaction.KindBytes", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{path: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(scratch, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("env", "-u", "GIT_DIR", "-u", "GIT_WORK_TREE", "go", "test", "-overlay", overlayPath, "-count=1", "-timeout=2m", "-v", "./internal/envprofile", "-run", "^TestSkillProjectionRepairNoGap$")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 1 || strings.Count(string(output), "--- FAIL: TestSkillProjectionRepairNoGap/") != 3 || !strings.Contains(string(output), "after_backup: no-gap assertion: shared projection unavailable:") || !strings.Contains(string(output), "no such file or directory") {
		t.Fatalf("mutant survived or lacked intended failure: %v\n%s", err, output)
	}
	t.Logf("expected-red exit %d:\n%s", exit.ExitCode(), output)
}
