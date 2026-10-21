//go:build unix

package envprofile

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/transaction"
)

// Both publishers report staging cleanup failures, including alongside a
// transaction refusal. A clean success removes staging; cleanup failure after
// commit preserves the complete published tree, and refusal publishes nothing.
// Permission fault injection is limited to the test's own staging directory.
func TestSkillProjectionCleanupErrors(t *testing.T) {
	probe := t.TempDir()
	if err := os.Chmod(probe, 0500); err != nil {
		t.Fatal(err)
	}
	err := os.WriteFile(filepath.Join(probe, "probe"), nil, 0600)
	if restoreErr := os.Chmod(probe, 0700); restoreErr != nil {
		t.Fatal(restoreErr)
	}
	if err == nil {
		t.Skip("this process can write through a read-only directory")
	}
	if !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("permission control failed for another reason: %v", err)
	}
	for _, publisher := range []string{"projection", "fallback"} {
		for _, mode := range []string{"clean", "cleanup-error", "combined-error"} {
			t.Run(publisher+"/"+mode, func(t *testing.T) {
				home, source := t.TempDir(), t.TempDir()
				writeGitFile(t, source, "SKILL.md", "# Guide\n")
				skill, err := projectSkill(home, source, hashing.VersionV2)
				if err != nil {
					t.Fatal(err)
				}
				root, rel := EnvRoot(home), "sample/codex_cli/skills/guide"
				live, pattern := skill.root, filepath.Join(filepath.Dir(skill.root), ".projection-*")
				if publisher == "fallback" {
					live = filepath.Join(root, rel)
					if err := os.MkdirAll(filepath.Dir(live), 0700); err != nil {
						t.Fatal(err)
					}
					pattern = filepath.Join(filepath.Dir(live), ".curator-skill-copy-*")
				}
				stage := ""
				fault := errors.New("injected prepared refusal")
				op, err := beginOperation(home, transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
					if event.Point != transaction.PointPrepared {
						return nil
					}
					matches, err := filepath.Glob(pattern)
					if err != nil || len(matches) != 1 {
						t.Fatalf("staging control is not unique: %v %v", matches, err)
					}
					stage = matches[0]
					if mode != "clean" {
						if err := os.Chmod(stage, 0500); err != nil {
							t.Fatal(err)
						}
					}
					if mode == "combined-error" {
						return fault
					}
					return nil
				}}))
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					if err := op.close(); err != nil {
						t.Error(err)
					}
					if stage != "" && mode != "clean" {
						if err := os.Chmod(stage, 0700); err != nil {
							t.Error(err)
						}
						if err := os.RemoveAll(stage); err != nil {
							t.Error(err)
						}
					}
				}()
				if publisher == "projection" {
					err = publishSkillProjection(op, home, skill)
				} else {
					err = publishSkillFallback(op, root, rel, source)
				}
				if stage == "" || errors.Is(err, fs.ErrPermission) != (mode != "clean") || errors.Is(err, fault) != (mode == "combined-error") {
					t.Fatalf("cleanup error identity lost: %v stage=%q", err, stage)
				}
				if mode == "clean" {
					if err != nil {
						t.Fatal(err)
					}
					if _, err := os.Stat(stage); !os.IsNotExist(err) {
						t.Fatalf("successful cleanup retained staging: %v", err)
					}
				}
				if mode == "combined-error" {
					if _, err := os.Lstat(live); !os.IsNotExist(err) {
						t.Fatalf("prepared refusal published a target: %v", err)
					}
				} else if !reflect.DeepEqual(projectionFiles(t, live), projectionFiles(t, source)) {
					t.Fatal("cleanup changed committed projection bytes")
				}
			})
		}
	}
}
