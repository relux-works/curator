package install

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/buildcache"
)

// cacheRebuildWatchdog bounds the production install entry below. The
// slowest authoritative cache-rebuild sample on the hosted Windows lane is
// 13.0s (gate run 35418260010, artifact-hash-mismatch); the watchdog gives
// that sample ~14x headroom while still failing ~20x faster than the lane's
// hour-long package timeout. A future hang in the rebuild path trips the
// watchdog instead of consuming the lane.
const cacheRebuildWatchdog = 180 * time.Second

// TestCacheWrongTargetRebuildCompletesWithinDeadline reproduces the
// authoritative cache-wrong-target rebuild at the production entry
// (install.Project via env.install): a refused protected-cache entry must
// be rebuilt privately, never adopted — and the rebuild must return within
// a short test-side deadline on every OS. The test is self-contained (no
// CURATOR_CONFORMANCE_ROOT, no skips) and covers both non-reusable
// boundary verdicts the authoritative suite assigns by position.
func TestCacheWrongTargetRebuildCompletesWithinDeadline(t *testing.T) {
	t.Parallel()
	for _, status := range nonReusableStatuses {
		t.Run(string(status), func(t *testing.T) {
			e := newEnv(t)
			e.buildSkill("build-skill", "alpha")
			e.declare("build-skill")
			deps, _, cache, builder := newFakeDeps(t)

			// The refused entry still advertises an artifact path.
			// Adopting it is exactly what the rebuild must refuse.
			refused := filepath.Join(t.TempDir(), "refused-artifact")
			if err := os.WriteFile(refused, []byte("refused"), 0o600); err != nil {
				t.Fatal(err)
			}
			cache.byCommand["alpha"] = buildcache.Result{
				Status: status, Reason: "cache-wrong-target: wrong target", ArtifactPath: refused,
			}

			result := installWithWatchdog(t, e, Options{Build: deps}, cacheRebuildWatchdog)
			if result.Status != "ok" {
				t.Fatalf("the refused entry was not repaired: %+v", result)
			}
			if got := string(result.Builds[0].Outcome()); got == string(BuildCacheHit) {
				t.Fatalf("a refused entry was adopted as a cache hit")
			}
			if len(builder.calls) != 1 || builder.calls[0] != "alpha" {
				t.Fatalf("the refused entry was not rebuilt privately: %v", builder.calls)
			}
			if got := result.Builds[0].ArtifactPath(); got == refused {
				t.Fatalf("the installation adopted the refused artifact %q", got)
			}
			for _, staged := range result.Staged {
				if staged.Receipt().CacheKey != result.Builds[0].logicalKey {
					t.Fatalf("the rebuild published a receipt for another key: %+v", staged.Receipt())
				}
			}
		})
	}
}

// installWithWatchdog runs the production install entry in a dedicated
// goroutine and fails the test fast when it does not return within the
// deadline. The bound is test-side only: production keeps its own restart
// bound (runWithRestarts), and a watchdog trip reports the hang instead of
// waiting out the package timeout.
func installWithWatchdog(t *testing.T, e *env, opts Options, deadline time.Duration) Result {
	t.Helper()
	done := make(chan Result, 1)
	go func() { done <- e.install(opts) }()
	timer := time.NewTimer(deadline)
	defer timer.Stop()
	select {
	case result := <-done:
		return result
	case <-timer.C:
		t.Fatalf("install did not return within %s; refusing to wait out the package timeout", deadline)
		panic("unreachable")
	}
}
