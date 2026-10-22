package projectlease

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"

	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/transaction"
)

type collectProbe struct {
	KeyCount int
	Acquired bool
	Error    string
	Removed  []string
}

func collectInProcess(t *testing.T, home, mode string) collectProbe {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestLeaseCollectorProcessHelper$", "-test.v")
	cmd.Env = append(os.Environ(), "CURATOR_TEST_PENDING_GC_HOME="+home, "CURATOR_TEST_PENDING_GC_MODE="+mode)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("collector helper failed: %v\n%s", err, out)
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "collect-result:") {
			var p collectProbe
			if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "collect-result:")), &p); err != nil {
				t.Fatal(err)
			}
			return p
		}
	}
	t.Fatalf("collector helper produced no observation: %s", out)
	return collectProbe{}
}

func nextAdmission(t *testing.T, f *fixture) Request {
	t.Helper()
	next := f.c
	next.Ref, next.SHA256, next.CheckoutID, next.HomeRef = "composition:next", strings.Repeat("c", 64), "checkout-next", "home:next"
	next.RuntimeEntries = []string{"tool/next"}
	next.BuildKeys = []string{"sha256:" + strings.Repeat("d", 64)}
	for _, rel := range []string{"runtime/tool/next", "cache/build/go-v1/" + strings.Repeat("d", 64), "runtime/tool/orphan"} {
		if err := os.MkdirAll(filepath.Join(f.home, filepath.FromSlash(rel)), 0700); err != nil {
			t.Fatal(err)
		}
	}
	f.c = next
	f.a.comps[next.Ref] = next
	return f.reserve("next")
}

// After durable Prepare, an unchanged home lock excludes a second owner and
// the lease commits. Replacing its stable name admits the second owner but
// must invalidate the first acquisition before any new binding is published.
// The second owner runs the real collector; its pending-intent result is
// asserted separately by TestAcquirePreparedRecoveryArtifacts.
func TestAcquireRevalidatesPreparedLock(t *testing.T) {
	for _, replace := range []bool{false, true} {
		t.Run(map[bool]string{false: "unchanged", true: "replaced"}[replace], func(t *testing.T) {
			f := setup(t)
			acquire(t, f, "first")
			before, err := os.ReadFile(statePath(f.home))
			if err != nil {
				t.Fatal(err)
			}
			request := nextAdmission(t, &f)
			reached := false
			f.m.hooks = transaction.Hooks{Observe: func(e transaction.Event) {
				if e.Point != transaction.PointPrepared || reached {
					return
				}
				reached = true
				if replace {
					paths, err := filepath.Glob(filepath.Join(f.home, "state", "locks", "v1", "*.lock"))
					if err != nil || len(paths) != 1 {
						t.Fatalf("lock fixture: %v %v", paths, err)
					}
					if err = os.Rename(paths[0], paths[0]+".old"); err != nil {
						if runtime.GOOS == "windows" && errors.Is(err, syscall.Errno(32)) {
							t.Skip("host refuses open-lock replacement")
						}
						t.Fatal(err)
					}
					if err = os.WriteFile(paths[0], nil, 0600); err != nil {
						t.Fatal(err)
					}
				}
				observation := collectInProcess(t, f.home, "try")
				if observation.Acquired != replace {
					t.Fatalf("second owner acquired=%v with replace=%v: %s", observation.Acquired, replace, observation.Error)
				}
			}}
			_, err = f.m.Acquire(context.Background(), request)
			if !reached {
				t.Fatal("prepared boundary not reached")
			}
			if !replace {
				if err != nil {
					t.Fatal("unchanged lock refused admission", err)
				}
				return
			}
			if !errors.Is(err, managerlock.ErrNotHeld) {
				t.Fatalf("stale Acquire committed or wrong refusal: %v", err)
			}
			if _, err = os.Lstat(bindingPath(f.home, "next")); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("stale acquisition allocated binding", err)
			}
			after, err := os.ReadFile(statePath(f.home))
			if err != nil || string(before) != string(after) {
				t.Fatal("stale acquisition changed prior ledger", err)
			}
		})
	}
}

// Real child exit at Prepare or after a first target install leaves durable
// admission intent. The real same-home collector must refuse without sweeping
// runtime (even with JournalKeys=[] and without a prior ledger). Recovery with
// intact artifacts commits conservatively; missing runtime/build rolls back
// the new binding and preserves the previous ledger. Without a prior ledger,
// the uncertain first-admission namespace keeps the missing-ledger refusal.
func TestAcquirePreparedRecoveryArtifacts(t *testing.T) {
	for _, prior := range []bool{false, true} {
		for _, point := range []transaction.Point{transaction.PointPrepared, transaction.PointAfterInstall} {
			for _, loss := range []string{"present", "runtime", "build"} {
				name := map[bool]string{false: "first", true: "prior"}[prior] + "/" + string(point) + "/" + loss
				t.Run(name, func(t *testing.T) {
					f := setup(t)
					var before []byte
					if prior {
						acquire(t, f, "first")
						var err error
						before, err = os.ReadFile(statePath(f.home))
						if err != nil {
							t.Fatal(err)
						}
					}
					_ = nextAdmission(t, &f)
					cmd := exec.Command(os.Args[0], "-test.run=^TestLeasePrepareExitHelper$")
					cmd.Env = append(os.Environ(), "CURATOR_TEST_PREPARE_EXIT_HOME="+f.home, "CURATOR_TEST_PREPARE_EXIT_POINT="+string(point))
					out, err := cmd.CombinedOutput()
					exit, ok := err.(*exec.ExitError)
					if !ok || exit.ExitCode() != 57 {
						t.Fatalf("admission crash plant missed: %v %s", err, out)
					}
					p := collectInProcess(t, f.home, "collect")
					if p.KeyCount != 0 {
						t.Fatal("pending admission fixture unexpectedly has journal build keys", p.KeyCount)
					}
					if !p.Acquired || !strings.Contains(p.Error, "pending project lease transaction") || len(p.Removed) != 0 {
						t.Fatalf("pending GC did not refuse safely: %+v", p)
					}
					if _, err = os.Stat(filepath.Join(f.home, "runtime", "tool", "next")); err != nil {
						t.Fatal("pending GC removed admission runtime", err)
					}
					if loss != "present" {
						target := filepath.Join(f.home, "runtime", "tool", "next")
						if loss == "build" {
							target = filepath.Join(f.home, "cache", "build", "go-v1", strings.Repeat("d", 64))
						}
						if err = os.Remove(target); err != nil {
							t.Fatal(err)
						}
					}
					err = f.m.Recover(context.Background())
					if loss == "present" {
						if err != nil {
							t.Fatal("intact prepared recovery refused", err)
						}
						if row := status(t, f, "next"); row.State != Unknown {
							t.Fatalf("recovered admission is not unknown: %+v", row)
						}
					} else {
						if !errors.Is(err, ErrArtifactsUnavailable) {
							t.Fatalf("missing recovery artifact committed or wrong refusal: %v", err)
						}
						if _, err = os.Lstat(bindingPath(f.home, "next")); !errors.Is(err, os.ErrNotExist) {
							t.Fatal("failed recovery allocated binding", err)
						}
						after, readErr := os.ReadFile(statePath(f.home))
						if prior {
							if readErr != nil || string(before) != string(after) {
								t.Fatal("failed recovery changed prior ledger", readErr)
							}
						} else if !errors.Is(readErr, os.ErrNotExist) {
							t.Fatal("failed first admission allocated ledger", readErr)
						}
						if !prior {
							// A first-admission crash leaves an uncommitted namespace
							// and source staging files. Preserve that uncertain state;
							// absence of its ledger must not become permission to sweep.
							err = f.m.Recover(context.Background())
							if err == nil || !strings.Contains(err.Error(), "ledger missing") {
								t.Fatal("first rollback lost conservative namespace refusal", err)
							}
							p = collectInProcess(t, f.home, "collect")
							if !strings.Contains(p.Error, "ledger missing") || len(p.Removed) != 0 {
								t.Fatalf("first rollback swept uncertain namespace: %+v", p)
							}
							return
						}
						// A second recovery proves the rolled-back intent cannot resurrect later.
						if err = f.m.Recover(context.Background()); err != nil {
							t.Fatal("rolled-back recovery remained pending", err)
						}
					}
					p = collectInProcess(t, f.home, "collect")
					if p.Error != "" || !p.Acquired {
						t.Fatalf("settled GC refused: %+v", p)
					}
					found := false
					for _, r := range p.Removed {
						if r == "tool/orphan" {
							found = true
						}
					}
					if !found {
						t.Fatal("settled GC failed to remove orphan positive control")
					}
					if loss == "present" {
						if _, err = os.Stat(filepath.Join(f.home, "runtime", "tool", "next")); err != nil {
							t.Fatal("settled GC removed recovered leased runtime", err)
						}
					}
				})
			}
		}
	}
}

// Child fixture for TestAcquirePreparedRecoveryArtifacts; exit 57 proves the
// selected durable boundary was reached rather than an ordinary setup error.
func TestLeasePrepareExitHelper(t *testing.T) {
	home := os.Getenv("CURATOR_TEST_PREPARE_EXIT_HOME")
	if home == "" {
		return
	}
	f := atHome(t, home)
	r := nextAdmission(t, &f)
	point := transaction.Point(os.Getenv("CURATOR_TEST_PREPARE_EXIT_POINT"))
	f.m.hooks = transaction.Hooks{Observe: func(e transaction.Event) {
		if e.Point == point {
			os.Exit(57)
		}
	}}
	if _, err := f.m.Acquire(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	t.Fatal("admission crash boundary not reached")
}

// Exact lower-case closed binding members are accepted. Replacing any member
// with an upper-case alias, or adding such an alias alongside it, refuses even
// when the input remains canonical JSON and every value is otherwise valid.
func TestBindingExactMembers(t *testing.T) {
	f := setup(t)
	b := acquire(t, f, "one")
	raw, _ := canonical(b)
	if got, err := ParseBinding(raw); err != nil || got != b {
		t.Fatal("exact binding refused", err)
	}
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	for key := range fields {
		for _, alongside := range []bool{false, true} {
			t.Run(key+map[bool]string{false: "/replaced", true: "/alongside"}[alongside], func(t *testing.T) {
				changed := map[string]any{}
				for k, v := range fields {
					changed[k] = v
				}
				changed[strings.ToUpper(key)] = changed[key]
				if !alongside {
					delete(changed, key)
				}
				bad, _ := canonical(changed)
				if _, err := ParseBinding(bad); err == nil {
					t.Fatalf("case-alias binding accepted: %s", bad)
				}
			})
		}
	}
}
