package install

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/manifest"
)

// Private compilation holds the global project lock, leaves the home lock
// available, and cannot publish profile or canonical state on failure/cancel.
// The fake compiler isolates the install boundary; the CLI witness uses real Go.
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
				probe, stop := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer stop()
				lock, err := manager.AcquireHomeOnly(probe, false)
				if err != nil {
					t.Fatalf("private staging held the home lock: %v", err)
				}
				_ = lock.Close()
				project := manager.NewOperation(false)
				defer func() { _ = project.Close() }()
				if _, err := project.AcquireProjects(probe, GlobalRoot(e.home)); !errors.Is(err, context.DeadlineExceeded) {
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
			result := Global(e.cfg, t.TempDir(), Options{Context: ctx, Build: deps, WithGlobalProfile: func(lock HomeLock, next func() Result) (Result, error) {
				if err := lock.AssertHeld(); err != nil {
					t.Fatal(err)
				}
				published = true
				return next(), nil
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
				WithGlobalProfile: func(_ HomeLock, next func() Result) (Result, error) {
					if len(builder.calls) != 1 {
						t.Fatal("prospective addition was not staged")
					}
					live, err := manifest.Load(GlobalRoot(e.home))
					if err != nil || len(live.Skills) != 0 {
						t.Fatalf("addition declaration published before private staging: %v", err)
					}
					if err := manifest.AddDecl(GlobalRoot(e.home), addition.Name, "tag", addition.Tag, "", ""); err != nil {
						t.Fatal(err)
					}
					if diverged {
						e.write(GlobalRoot(e.home), "Skillfile.json", `{"schema_version":1,"skills":[]}`)
					}
					return next(), nil
				},
			})
			if diverged {
				if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, " "), "differs from the staged addition") || len(cache.published) != 0 {
					t.Fatalf("divergent addition reached publication: %+v", result)
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
		WithGlobalProfile: func(_ HomeLock, next func() Result) (Result, error) {
			publications++
			if stages != 2 {
				t.Fatal("stale declaration reached profile publication before restart")
			}
			return next(), nil
		},
	})
	if result.Status != "ok" || stages != 2 || publications != 1 {
		t.Fatalf("stale declaration reached profile publication: stages=%d publications=%d result=%+v", stages, publications, result)
	}
	if _, err := os.Lstat(filepath.Join(GlobalRoot(e.home), "skills", "build-skill")); !os.IsNotExist(err) {
		t.Fatalf("stale staged canonical skill published: %v", err)
	}
}

// A continuation's generation restart retries only after successful profile
// compensation. A typed compensation failure terminates the first attempt and
// remains visible in the result instead of being discarded by the retry loop.
func TestGlobalProfileCompensationStopsRestart(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "restored", true: "failed"}[fail], func(t *testing.T) {
			e := newEnv(t)
			e.buildSkill("build-skill", "alpha")
			e.globalDeclare("build-skill")
			deps, _, _, _ := newFakeDeps(t)
			attempts := 0
			result := Global(e.cfg, t.TempDir(), Options{Build: deps,
				WithGlobalProfile: func(_ HomeLock, next func() Result) (Result, error) {
					attempts++
					if attempts == 1 {
						e.write(GlobalRoot(e.home), "Skillfile.json", `{"schema_version":1,"skills":[]}`)
					}
					result := next()
					if attempts == 1 && result.Status != "failed" {
						t.Fatalf("generation restart was not exercised: %+v", result)
					}
					if fail {
						return result, &envprofile.GlobalCompensationError{Cause: errors.New("planted surface conflict")}
					}
					return result, nil
				},
			})
			if fail {
				if attempts != 1 || result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "global profile compensation failed") {
					t.Fatalf("compensation failure was retried or hidden: attempts=%d result=%+v", attempts, result)
				}
			} else if attempts != 2 || result.Status != "ok" {
				t.Fatalf("successful compensation did not retry: attempts=%d result=%+v", attempts, result)
			}
		})
	}
}
