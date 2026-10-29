package install

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/transaction"
)

// Private compilation holds the global project lock, leaves the home lock
// available, and cannot publish profile or canonical state on failure/cancel.
// The fake compiler isolates the install boundary; the CLI witness uses real Go.
// Separate probes cancel on actual contention, not elapsed acquisition time.
func TestGlobalProfilePrivateStaging(t *testing.T) {
	for _, mode := range []string{"success", "failure", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			e := newEnv(t)
			e.buildSkill("build-skill", "alpha")
			e.globalDeclare("build-skill")
			deps, toolchain, cache, builder := newFakeDeps(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			published := false
			before := snapshotState(t, e)
			builder.observe = func(StageRequest) {
				manager, err := managerlock.New(e.home)
				if err != nil {
					t.Fatal(err)
				}
				homeProbe, stopHome := context.WithCancel(t.Context())
				defer stopHome()
				lock, err := manager.AcquireHomeOnly(privateStagingContentionContext{homeProbe, stopHome}, false)
				if err != nil {
					t.Fatalf("private staging held the home lock: %v", err)
				}
				_ = lock.Close()
				project := manager.NewOperation(false)
				defer func() { _ = project.Close() }()
				probe, stop := context.WithCancel(t.Context())
				defer stop()
				if _, err := project.AcquireProjects(privateStagingContentionContext{probe, stop}, GlobalRoot(e.home)); !errors.Is(err, context.Canceled) {
					t.Fatalf("private staging did not preserve project contention: %v", err)
				}
				if published {
					t.Fatal("profile published before private staging")
				}
				if mode == "cancel" {
					cancel()
				}
			}
			if mode == "failure" {
				builder.failOn["alpha"] = errors.New("compiler refused fixture")
			}
			if mode == "cancel" {
				builder.failOn["alpha"] = context.Canceled
			}
			result := Global(e.cfg, t.TempDir(), Options{Context: ctx, Build: deps, PublishGlobalPlan: func(lock HomeLock, stage func(envprofile.NativeSkillOwnership) (transaction.Plan, error)) error {
				if err := lock.AssertHeld(); err != nil {
					t.Fatal(err)
				}
				published = true
				return publishStagingProfile(t, e, lock, stage)
			}})
			if mode == "success" {
				if result.Status != "ok" || !published {
					t.Fatalf("valid staged build did not publish: %+v", result)
				}
			} else {
				if result.Status != "failed" || published || !reflect.DeepEqual(before, snapshotState(t, e)) {
					t.Fatalf("failed staged build published state: %+v", result)
				}
				want := "compiler refused fixture"
				if mode == "cancel" {
					want = "context canceled"
				}
				if !strings.Contains(strings.Join(result.Errors, " "), want) {
					t.Fatalf("wrong build refusal: %+v", result)
				}
				if len(cache.published) != 0 || toolchain.session.released != 1 {
					t.Fatal("failed staging published cache or retained its private session")
				}
			}
			manager, _ := managerlock.New(e.home)
			probe, stop := context.WithCancel(t.Context())
			defer stop()
			lock, err := manager.AcquireHomeOnly(probe, false)
			if err != nil {
				t.Fatalf("staging left the home lock held: %v", err)
			}
			_ = lock.Close()
		})
	}
}

// Done is consulted by the real manager lock after an unsuccessful OS lock
// attempt. A held lock therefore cancels the probe without a timer; preparing
// and acquiring an available lock have only the enclosing test's hang guard.
type privateStagingContentionContext struct {
	context.Context
	cancel context.CancelFunc
}

func (c privateStagingContentionContext) Done() <-chan struct{} {
	c.cancel()
	return c.Context.Done()
}

// An addition absent from the live Skillfile is compiled before its declaration
// publishes. Only the exact prospective bytes may advance that observation;
// a divergent write refuses before publishing the compiled canonical entry.
func TestGlobalProfileStagingAddition(t *testing.T) {
	for _, diverged := range []bool{false, true} {
		t.Run(map[bool]string{false: "exact", true: "diverged"}[diverged], func(t *testing.T) {
			e := newEnv(t)
			e.buildSkill("build-skill", "alpha")
			if _, err := GlobalInit(e.home); err != nil {
				t.Fatal(err)
			}
			deps, _, cache, builder := newFakeDeps(t)
			addition := &envprofile.DirectSkill{Name: "build-skill", Tag: "v1"}
			result := Global(e.cfg, t.TempDir(), Options{Build: deps, GlobalAddition: addition,
				PublishGlobalPlan: func(lock HomeLock, stage func(envprofile.NativeSkillOwnership) (transaction.Plan, error)) error {
					if len(builder.calls) != 1 {
						t.Fatal("prospective addition was not staged")
					}
					live, err := manifest.Load(GlobalRoot(e.home))
					if err != nil || len(live.Skills) != 0 {
						t.Fatalf("addition declaration published before private staging: %v", err)
					}

					if diverged {
						e.write(GlobalRoot(e.home), "Skillfile.json", `{"schema_version":1,"skills":[],"agents":["codex_cli"]}`)
					}
					return publishStagingProfile(t, e, lock, stage)
				},
			})
			if diverged {
				if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, " "), "declaration changed during publication planning") || len(cache.byCommand) != 0 || len(cache.reverted) != len(cache.published) {
					t.Fatalf("divergent addition reached publication: %+v", result)
				}
				if _, err := os.Lstat(filepath.Join(GlobalRoot(e.home), "skills", addition.Name)); !os.IsNotExist(err) {
					t.Fatalf("divergent declaration installed canonical skill: %v", err)
				}
			} else if result.Status != "ok" || len(cache.published) != 1 {
				t.Fatalf("exact staged addition did not publish: %+v", result)
			}
		})
	}
}

// A changed declaration restarts before profile publication; the stable retry
// publishes exactly once from the new empty declaration, never the staged skill.
func TestGlobalProfileStagingRevalidatesDeclaration(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	e.globalDeclare("build-skill")
	deps, _, _, _ := newFakeDeps(t)
	stages, publications := 0, 0
	result := Global(e.cfg, t.TempDir(), Options{Build: deps,
		OnStaged: func(Staged) error {
			stages++
			if stages == 1 {
				e.write(GlobalRoot(e.home), "Skillfile.json", `{"schema_version":1,"skills":[]}`)
			}
			return nil
		},
		PublishGlobalPlan: func(lock HomeLock, stage func(envprofile.NativeSkillOwnership) (transaction.Plan, error)) error {
			publications++
			if stages != 2 {
				t.Fatal("stale declaration reached profile publication before restart")
			}
			return publishStagingProfile(t, e, lock, stage)
		},
	})
	if result.Status != "ok" || stages != 2 || publications != 1 {
		t.Fatalf("stale declaration reached profile publication: stages=%d publications=%d result=%+v", stages, publications, result)
	}
	if _, err := os.Lstat(filepath.Join(GlobalRoot(e.home), "skills", "build-skill")); !os.IsNotExist(err) {
		t.Fatalf("stale staged canonical skill published: %v", err)
	}
}

// A joined publication refusal is final for that attempt. No arbitrary callback
// may request an optimistic restart after profile state has already committed.
// Actual rollback/substitution behavior is covered by joined owner tests; this
// control checks propagation and single execution at the install boundary.
func TestGlobalProfilePublicationRefusalStopsRestart(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	e.globalDeclare("build-skill")
	deps, _, _, _ := newFakeDeps(t)
	attempts := 0
	result := Global(e.cfg, t.TempDir(), Options{Build: deps,
		PublishGlobalPlan: func(lock HomeLock, stage func(envprofile.NativeSkillOwnership) (transaction.Plan, error)) error {
			attempts++
			if err := lock.AssertHeld(); err != nil {
				t.Fatal(err)
			}
			return errors.New("joined publication fixture refusal")
		},
	})
	if attempts != 1 || result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "joined publication fixture refusal") {
		t.Fatalf("publication refusal retried or hidden: attempts=%d result=%+v", attempts, result)
	}
	if _, err := os.Lstat(filepath.Join(GlobalRoot(e.home), "skills", "build-skill")); !os.IsNotExist(err) {
		t.Fatalf("refusal installed canonical skill: %v", err)
	}
}

// This delegates to the actual production profile planner and executor. Raw
// native roots are owned empty fixture directories; no account lookup is used.
// Local-only legacy skills have no profile Git member, so addition validation
// is tested separately through the real Git-backed CLI owner controls.
func publishStagingProfile(t *testing.T, e *env, lock HomeLock, stage func(envprofile.NativeSkillOwnership) (transaction.Plan, error)) error {
	t.Helper()
	native := t.TempDir()
	resolve := func(id string) (string, error) {
		path := filepath.Join(native, id)
		if err := os.MkdirAll(path, 0700); err != nil {
			return "", err
		}
		return path, nil
	}
	_, err := envprofile.PublishGlobalSkillPlanWithPlanner(e.home, envprofile.PolicyFromConfig(e.cfg), envregistry.DefaultMachineConfig(), nil, resolve, lock, stage)
	return err
}
