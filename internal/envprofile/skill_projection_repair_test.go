package envprofile

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/transaction"
)

// Missing, extra and changed files on both adapters stay stale under read-only
// resolve/status, then explicit repair restores the complete pinned byte map.
// Intact and whole-directory-missing cases are neighboring positive controls.
func TestProfileSkillProjectionDamagedStore(t *testing.T) {
	for _, id := range []string{"claude_code", "codex_cli"} {
		for _, mode := range []string{"intact", "missing-directory", "missing-file", "extra-file", "changed-file"} {
			t.Run(id+"/"+mode, func(t *testing.T) {
				fx := writeManagedFixture(t, "sample")
				req := fx.request(id)
				req.Repair = true
				if _, err := Resolve(req); err != nil {
					t.Fatal(err)
				}
				live := filepath.Join(ManagedHomeDir(fx.home, fx.profile, id), "skills/myskill")
				root, err := os.Readlink(live)
				if err != nil {
					t.Fatal(err)
				}
				want := projectionFiles(t, root)
				switch mode {
				case "missing-directory":
					err = os.RemoveAll(root)
				case "missing-file":
					err = os.Remove(filepath.Join(root, "SKILL.md"))
				case "extra-file":
					err = os.WriteFile(filepath.Join(root, "extra.md"), []byte("extra\n"), 0600)
				case "changed-file":
					err = os.WriteFile(filepath.Join(root, "SKILL.md"), []byte("changed\n"), 0600)
				}
				if err != nil {
					t.Fatal(err)
				}
				before, err := transaction.DigestPath(root)
				if err != nil {
					t.Fatal(err)
				}
				req.Repair = false
				_, err = Resolve(req)
				if mode == "intact" {
					if err != nil {
						t.Fatal(err)
					}
				} else if err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
					t.Fatalf("expected stale refusal: %v", err)
				}
				adapter, err := envregistry.ByID(id)
				if err != nil {
					t.Fatal(err)
				}
				state := homeState(StatusRequest{Home: fx.home, Machine: req.Machine, NativeHomeOf: req.NativeHomeOf, OperatorXDG: req.OperatorXDG}, fx.profile, adapter)
				if mode != "intact" && state.Current {
					t.Fatal("damaged projection status current")
				}
				if after, err := transaction.DigestPath(root); err != nil || after != before {
					t.Fatalf("read-only resolve/status changed projection: %v", err)
				}
				req.Repair = true
				if _, err := Resolve(req); err != nil {
					t.Fatalf("projection repair did not converge for %s: %v", mode, err)
				}
				req.Repair = false
				if _, err := Resolve(req); err != nil {
					t.Fatal(err)
				}
				if got := projectionFiles(t, root); !reflect.DeepEqual(got, want) {
					t.Fatal("repair bytes differ from pinned projection")
				}
			})
		}
	}
}

// A freshly staged tree that differs from the planned pin refuses once with
// ErrSkillProjectionIdentity, never prepares a transaction, preserves the old
// directory and immutable input, and removes its scratch. Clean retry succeeds.
func TestSkillProjectionRebuildSourceRefusal(t *testing.T) {
	home, source := t.TempDir(), t.TempDir()
	writeGitFile(t, source, "SKILL.md", "# Guide\n")
	skill, err := projectSkill(home, source, hashing.VersionV2)
	if err != nil {
		t.Fatal(err)
	}
	if err := publishTestSkillProjection(t, home, skill); err != nil {
		t.Fatal(err)
	}
	writeGitFile(t, skill.root, "extra.md", "old tree\n")
	writeGitFile(t, source, "SKILL.md", "# Changed source\n")
	old := projectionFiles(t, skill.root)
	raw := projectionFiles(t, source)
	prepared := 0
	op, err := beginOperation(home, transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
		if event.Point == transaction.PointPrepared {
			prepared++
		}
		return nil
	}}))
	if err != nil {
		t.Fatal(err)
	}
	err = publishSkillProjection(op, home, skill)
	if closeErr := op.close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if !errors.Is(err, ErrSkillProjectionIdentity) || !strings.Contains(err.Error(), "changed while projecting") || prepared != 0 {
		t.Fatalf("wrong staged identity refusal: %v prepared=%d", err, prepared)
	}
	if !reflect.DeepEqual(old, projectionFiles(t, skill.root)) || !reflect.DeepEqual(raw, projectionFiles(t, source)) {
		t.Fatal("refusal changed old projection or source")
	}
	entries, err := os.ReadDir(filepath.Dir(skill.root))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".projection-") {
			t.Fatal("refusal left staging tree")
		}
	}
	writeGitFile(t, source, "SKILL.md", "# Guide\n")
	if err := publishTestSkillProjection(t, home, skill); err != nil {
		t.Fatal(err)
	}
	if err := verifySkillProjection(home, skill); err != nil {
		t.Fatal(err)
	}
}

// Faults after backup and after publication restore the exact old tree. The
// installed target is always a complete verified tree at the install hook;
// recovery and a clean retry converge without an in-place partial copy.
func TestSkillProjectionRebuildRollback(t *testing.T) {
	for _, point := range []transaction.Point{transaction.PointAfterBackup, transaction.PointAfterInstall, transaction.PointTargetCommitted} {
		t.Run(string(point), func(t *testing.T) {
			home, source := t.TempDir(), t.TempDir()
			writeGitFile(t, source, "SKILL.md", "# Guide\n")
			writeGitFile(t, source, "references/guide.md", "visible\n")
			skill, err := projectSkill(home, source, hashing.VersionV2)
			if err != nil {
				t.Fatal(err)
			}
			if err := publishTestSkillProjection(t, home, skill); err != nil {
				t.Fatal(err)
			}
			writeGitFile(t, skill.root, "extra.md", "old tree\n")
			old := projectionFiles(t, skill.root)
			fault := errors.New("injected projection rebuild")
			fired := 0
			op, err := beginOperation(home, transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
				if event.Point == point {
					fired++
					if point != transaction.PointAfterBackup {
						if err := verifySkillProjection(home, skill); err != nil {
							t.Errorf("published partial projection: %v", err)
						}
					}
					return fault
				}
				return nil
			}}))
			if err != nil {
				t.Fatal(err)
			}
			err = publishSkillProjection(op, home, skill)
			if closeErr := op.close(); closeErr != nil {
				t.Fatal(closeErr)
			}
			if fired != 1 || !errors.Is(err, fault) {
				t.Fatalf("intended rebuild fault missing: %v fired=%d", err, fired)
			}
			if !reflect.DeepEqual(old, projectionFiles(t, skill.root)) {
				t.Fatal("failed rebuild changed old tree")
			}
			if err := publishTestSkillProjection(t, home, skill); err != nil {
				t.Fatal(err)
			}
			if err := verifySkillProjection(home, skill); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Two profiles share a projection. Repairing one performs one replacement;
// the second resolves current, and explicit repair there does not rebuild it.
func TestSkillProjectionSharedRebuild(t *testing.T) {
	for _, id := range []string{"claude_code", "codex_cli"} {
		t.Run(id, func(t *testing.T) {
			fx := writeManagedFixture(t, "sample")
			req := fx.request(id)
			req.Repair = true
			if _, err := Resolve(req); err != nil {
				t.Fatal(err)
			}
			lock, _, err := readLock(fx.home, fx.profile)
			if err != nil {
				t.Fatal(err)
			}
			source, err := os.ReadFile(sourcePath(fx.home, fx.profile))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := contextlock.Write(lockPath(fx.home, "sibling"), lock); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(sourcePath(fx.home, "sibling"), source, 0600); err != nil {
				t.Fatal(err)
			}
			sibling := req
			sibling.Profile = "sibling"
			if _, err := Resolve(sibling); err != nil {
				t.Fatal(err)
			}
			first, err := os.Readlink(filepath.Join(ManagedHomeDir(fx.home, fx.profile, id), "skills/myskill"))
			if err != nil {
				t.Fatal(err)
			}
			second, err := os.Readlink(filepath.Join(ManagedHomeDir(fx.home, "sibling", id), "skills/myskill"))
			if err != nil {
				t.Fatal(err)
			}
			if first != second {
				t.Fatal("profiles do not share projection")
			}
			writeGitFile(t, first, "extra.md", "old tree\n")
			count := 0
			hooks := transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
				if event.Point == transaction.PointTargetCommitted && event.Class == "skill-projection" {
					count++
				}
				return nil
			}})
			req.transactionOptions = []transaction.Option{hooks}
			sibling.transactionOptions = []transaction.Option{hooks}
			if _, err := Resolve(req); err != nil {
				t.Fatal(err)
			}
			sibling.Repair = false
			if _, err := Resolve(sibling); err != nil {
				t.Fatal(err)
			}
			sibling.Repair = true
			if _, err := Resolve(sibling); err != nil {
				t.Fatal(err)
			}
			if count != 1 {
				t.Fatalf("shared rebuild count %d, want 1", count)
			}
		})
	}
}

// A cleanup interruption leaves a durable journal and complete desired tree.
// A fresh operation recovers without source bytes, clears the journal and keeps
// the installed identity; recovery does not need the caller's staging directory.
func TestSkillProjectionRebuildRecovery(t *testing.T) {
	home, source := t.TempDir(), t.TempDir()
	writeGitFile(t, source, "SKILL.md", "# Guide\n")
	skill, err := projectSkill(home, source, hashing.VersionV2)
	if err != nil {
		t.Fatal(err)
	}
	if err := publishTestSkillProjection(t, home, skill); err != nil {
		t.Fatal(err)
	}
	writeGitFile(t, skill.root, "extra.md", "old tree\n")
	fault := errors.New("injected projection cleanup")
	fired := 0
	op, err := beginOperation(home, transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
		if event.Point == transaction.PointBeforeCleanup {
			fired++
			return fault
		}
		return nil
	}}))
	if err != nil {
		t.Fatal(err)
	}
	err = publishSkillProjection(op, home, skill)
	if fired != 1 || !errors.Is(err, fault) {
		t.Fatalf("intended cleanup interruption missing: %v fired=%d", err, fired)
	}
	pending, err := op.engine.PendingRecovery(op.lock)
	if err != nil || !pending {
		t.Fatalf("no durable recovery: %v %t", err, pending)
	}
	if err := op.close(); err != nil {
		t.Fatal(err)
	}
	if err := verifySkillProjection(home, skill); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(source); err != nil {
		t.Fatal(err)
	}
	op, err = beginOperation(home)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := op.close(); err != nil {
			t.Error(err)
		}
	}()
	pending, err = op.engine.PendingRecovery(op.lock)
	if err != nil || pending {
		t.Fatalf("recovery did not clear journal: %v %t", err, pending)
	}
	if err := verifySkillProjection(home, skill); err != nil {
		t.Fatal(err)
	}
}

// Native materialization receives the same held operation as managed repair:
// stale native links stay on one target path, which is recoverably rebuilt.
func TestSkillProjectionNativeRebuild(t *testing.T) {
	for _, id := range []string{"claude_code", "codex_cli"} {
		t.Run(id, func(t *testing.T) {
			fx := writeManagedFixture(t, "sample")
			req := fx.request(id)
			results, err := materializeTestScope(t, fx.home, fx.profile, id, Policy{}, req.NativeHomeOf)
			if err != nil || len(results) != 1 || !results[0].OK {
				t.Fatalf("native setup: %v %+v", err, results)
			}
			lock, _, err := readLock(fx.home, fx.profile)
			if err != nil {
				t.Fatal(err)
			}
			skills, err := skillsOf(fx.home, newGitManager(fx.home), lock)
			if err != nil || len(skills) != 1 {
				t.Fatalf("skill setup: %v", err)
			}
			skill := skills[0]
			writeGitFile(t, skill.root, "extra.md", "old tree\n")
			row, err := nativeSkillsState(fx.home, fx.profile, id, req.NativeHomeOf)
			if err != nil || row == nil || row.State == "current" {
				t.Fatalf("damaged native projection current: %v %+v", err, row)
			}
			results, err = materializeTestScope(t, fx.home, fx.profile, id, Policy{}, req.NativeHomeOf)
			if err != nil || len(results) != 1 || !results[0].OK {
				t.Fatalf("native repair: %v %+v", err, results)
			}
			if err := verifySkillProjection(fx.home, skill); err != nil {
				t.Fatal(err)
			}
			row, err = nativeSkillsState(fx.home, fx.profile, id, req.NativeHomeOf)
			if err != nil || row == nil || row.State != "current" {
				t.Fatalf("native repair not current: %v %+v", err, row)
			}
		})
	}
}
