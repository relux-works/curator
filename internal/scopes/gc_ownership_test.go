package scopes

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/projectlease"
)

// A real acquisition during executable enumeration invalidates the old mark
// set. The physical newly leased build must survive; with the original lock,
// the same child times out and the real sweeper removes the orphan.
func TestCollectCannotSweepNewOwnerLease(t *testing.T) {
	for _, replace := range []bool{false, true} {
		t.Run(map[bool]string{false: "held", true: "replaced"}[replace], func(t *testing.T) {
			home := protectedTestHome(t)
			store := openProtectedStore(t, home)
			key := publishRealEntry(t, store, "next", "leased artifact")
			backdateEntry(t, home, key, 30*24*time.Hour)
			lm, err := managerlock.New(home)
			if err != nil {
				t.Fatal(err)
			}
			lock, err := lm.AcquireHomeOnly(t.Context(), false)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = lock.Close() }()
			store.ExecutablePaths = func() ([]string, error) {
				if replace {
					replaceGCLock(t, home)
				}
				child := exec.Command(os.Args[0], "-test.run=^TestGCNewLeaseOwnerHelper$", "-test.v")
				child.Env = append(os.Environ(), "CURATOR_TEST_GC_OWNER_HOME="+home, "CURATOR_TEST_GC_OWNER_KEY="+string(key))
				out, err := child.CombinedOutput()
				if replace && (err != nil || !strings.Contains(string(out), "new lease committed")) {
					t.Fatalf("child admission: %v\n%s", err, out)
				}
				if !replace && (err != nil || !strings.Contains(string(out), "expected lock deadline")) {
					t.Fatalf("held lock control: %v\n%s", err, out)
				}
				return nil, nil
			}
			result, err := Collect(MaintenanceRequest{Home: home, Lock: lock, Cache: store})
			_, artifactErr := os.Stat(realEntryPath(home, key))
			if replace {
				if !errors.Is(err, managerlock.ErrNotHeld) || artifactErr != nil || len(result.RemovedBuilds) != 0 {
					t.Fatalf("stale GC removed newly leased build: %+v gc=%v artifact=%v", result, err, artifactErr)
				}
				m, _ := projectlease.New(home, nil, nil)
				if statuses, err := m.Status(t.Context()); err != nil || len(statuses) != 1 {
					t.Fatalf("stale GC damaged new ledger: %v %v", statuses, err)
				}
			} else if err != nil || !errors.Is(artifactErr, os.ErrNotExist) || len(result.RemovedBuilds) != 1 {
				t.Fatalf("held GC did not remove orphan: %+v %v %v", result, err, artifactErr)
			}
		})
	}
}

func replaceGCLock(t *testing.T, home string) {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(home, "state", "locks", "v1", "*.lock"))
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

// The helper performs actual Acquire and a durable Status check; a held-name
// sibling must produce the specific deadline instead of a setup failure.
func TestGCNewLeaseOwnerHelper(t *testing.T) {
	home := os.Getenv("CURATOR_TEST_GC_OWNER_HOME")
	if home == "" {
		return
	}
	a := leaseAuthority{projectlease.Composition{Ref: "composition:next", SHA256: strings.Repeat("a", 64), CheckoutID: "checkout", Profile: "default", Environment: "codex_cli", HomeRef: "home:next", ConcurrentState: true, BuildKeys: []string{os.Getenv("CURATOR_TEST_GC_OWNER_KEY")}, RuntimeEntries: []string{}, Generated: map[string]projectlease.Surface{}}}
	m, err := projectlease.New(home, a, a)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	_, err = m.Acquire(ctx, projectlease.Request{LeaseID: "next", ContextRef: a.c.Ref, ContextSHA256: a.c.SHA256, OwnerRef: "principal:operator", SessionRef: "session:first", Generation: 1})
	if errors.Is(err, context.DeadlineExceeded) {
		t.Log("expected lock deadline")
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if statuses, err := m.Status(t.Context()); err != nil || len(statuses) != 1 {
		t.Fatalf("new lease status: %v %v", statuses, err)
	}
	t.Log("new lease committed")
}

// Real runtime deletion and external artifact/snapshot deletion revalidate
// the retained acquisition after inventory reads. Replacement invalidates the
// same witness; unchanged acquisitions remove the identical orphan fixtures.
func TestGCSweepBoundariesRetainOwnership(t *testing.T) {
	for _, boundary := range []string{"runtime", "external-artifact", "external-snapshot"} {
		for _, replace := range []bool{false, true} {
			t.Run(boundary+map[bool]string{false: "/held", true: "/replaced"}[replace], func(t *testing.T) {
				home := protectedTestHome(t)
				lm, err := managerlock.New(home)
				if err != nil {
					t.Fatal(err)
				}
				lock, err := lm.AcquireHomeOnly(t.Context(), false)
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = lock.Close() }()
				witness := lock.Held()
				path := filepath.Join(home, "runtime", "tool", "orphan")
				if boundary == "external-artifact" {
					path = filepath.Join(home, "external-build-cache", buildrepo.ArtifactsDir(buildrepo.LegacyReceiptSchemaVersion), strings.Repeat("a", 64))
				}
				if boundary == "external-snapshot" {
					path = filepath.Join(home, "external-build-cache", "snapshots", strings.Repeat("a", 64))
				}
				if err = os.MkdirAll(path, 0700); err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(filepath.Join(path, "witness"), []byte("keep until acquisition verified"), 0600); err != nil {
					t.Fatal(err)
				}
				injected := false
				check := func() error {
					err := witness.Validate(home)
					// Lose ownership immediately after the first successful
					// observation, independently of any later deletion gate.
					if !injected {
						injected = true
						if replace {
							replaceGCLock(t, home)
						}
					}
					return err
				}
				var removed []string
				if boundary == "runtime" {
					removed, err = sweepRuntimeChecked(home, map[string]bool{}, check)
				} else {
					removed, err = buildrepo.CollectChecked(filepath.Join(home, "external-build-cache"), nil, time.Now().Add(time.Hour), time.Second, check)
				}
				_, statErr := os.Stat(path)
				if replace {
					if !errors.Is(err, managerlock.ErrNotHeld) || statErr != nil || len(removed) != 0 {
						t.Fatalf("%s stale sweep removed state: %v %v %v", boundary, removed, err, statErr)
					}
				} else if err != nil || !errors.Is(statErr, os.ErrNotExist) || len(removed) != 1 {
					t.Fatalf("%s held sweep orphan control: %v %v %v", boundary, removed, err, statErr)
				}
			})
		}
	}
}

type afterInternalSweep struct{ after func() }

func (cache afterInternalSweep) Sweep(buildcache.SweepRequest, buildcache.HomeLock) (buildcache.SweepResult, error) {
	cache.after()
	return buildcache.SweepResult{}, nil
}

// Collect must validate again between its internal and external sweep. An
// actual replacement at the returned cache boundary cannot reuse old marks.
func TestCollectRechecksBeforeExternalSweep(t *testing.T) {
	for _, replace := range []bool{false, true} {
		t.Run(map[bool]string{false: "held", true: "replaced"}[replace], func(t *testing.T) {
			home := protectedTestHome(t)
			path := filepath.Join(home, "external-build-cache", buildrepo.ArtifactsDir(buildrepo.LegacyReceiptSchemaVersion), strings.Repeat("a", 64))
			if err := os.MkdirAll(path, 0700); err != nil {
				t.Fatal(err)
			}
			lm, err := managerlock.New(home)
			if err != nil {
				t.Fatal(err)
			}
			lock, err := lm.AcquireHomeOnly(t.Context(), false)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = lock.Close() }()
			result, err := Collect(MaintenanceRequest{Home: home, Lock: lock, Now: time.Now().Add(time.Hour), Grace: time.Second, Cache: afterInternalSweep{after: func() {
				if replace {
					replaceGCLock(t, home)
				}
			}}})
			_, statErr := os.Stat(path)
			if replace {
				if !errors.Is(err, managerlock.ErrNotHeld) || statErr != nil || len(result.RemovedBuilds) != 0 {
					t.Fatalf("stale external boundary: %+v %v %v", result, err, statErr)
				}
			} else if err != nil || !errors.Is(statErr, os.ErrNotExist) || len(result.RemovedBuilds) != 1 {
				t.Fatalf("held external control: %+v %v %v", result, err, statErr)
			}
		})
	}
}

// A stale mark set cannot overwrite the consumer registry after ownership
// changes during pruning. The held sibling prunes the same known-dead entry.
func TestGCPruningRetainsOwnership(t *testing.T) {
	for _, replace := range []bool{false, true} {
		t.Run(map[bool]string{false: "held", true: "replaced"}[replace], func(t *testing.T) {
			home := protectedTestHome(t)
			if err := ReplaceConsumers(home, []string{t.TempDir()}); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(home, ConsumersName)
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			lm, err := managerlock.New(home)
			if err != nil {
				t.Fatal(err)
			}
			lock, err := lm.AcquireHomeOnly(t.Context(), false)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = lock.Close() }()
			witness := lock.Held()
			injected := false
			check := func() error {
				err := witness.Validate(home)
				// The loss occurs even if the later write gate is removed.
				if !injected {
					injected = true
					if replace {
						replaceGCLock(t, home)
					}
				}
				return err
			}
			err = pruneConsumersChecked(home, marks{}, check)
			after, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if replace {
				if !errors.Is(err, managerlock.ErrNotHeld) || string(before) != string(after) {
					t.Fatalf("stale pruning overwrote registry: %v before=%s after=%s", err, before, after)
				}
			} else if err != nil || string(before) == string(after) {
				t.Fatalf("held pruning control: %v", err)
			}
		})
	}
}
