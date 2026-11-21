package snapshot

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/transaction"
)

var errInjectedSharingViolation = errors.New("injected sharing violation")

func sharingRetryFixture(t *testing.T) (home, repo, commit, target string) {
	t.Helper()
	repo = t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(repo, "SKILL.md"), []byte("concurrent"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, repo, "add", ".")
	gitRun(t, repo, "commit", "-qm", "one")
	head, err := gitops.Resolve(repo, "revision", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	home = t.TempDir()
	target, err = Get(home, "internal/skill-a", repo, head.Commit)
	if err != nil {
		t.Fatal(err)
	}
	return home, repo, head.Commit, target
}

func injectSharingRetry(t *testing.T, digest func(string) (string, error)) func() {
	t.Helper()
	previousDigest, previousClassifier := digestSnapshotDestination, destinationSharingViolation
	digestSnapshotDestination = digest
	destinationSharingViolation = func(err error) bool { return errors.Is(err, errInjectedSharingViolation) }
	restore := func() {
		digestSnapshotDestination, destinationSharingViolation = previousDigest, previousClassifier
	}
	t.Cleanup(restore)
	return restore
}

func assertSharingSnapshotUnchanged(t *testing.T, target, before string) {
	t.Helper()
	after, err := transaction.DigestPath(target)
	if err != nil || after != before {
		t.Fatalf("destination digest = %q, %v; want unchanged %q", after, err, before)
	}
	entries, err := os.ReadDir(filepath.Dir(target))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "snapshot" {
		t.Fatalf("snapshot parent entries = %v; want only snapshot, with no staging residue", entries)
	}
}

// All 16 cached readers recover after four synchronized sharing failures each,
// beyond the previous three retries. The injected classifier makes this portable;
// native Windows publication and error classification have separate controls.
func TestConcurrentGetRetriesExtendedSharingViolation(t *testing.T) {
	home, repo, commit, target := sharingRetryFixture(t)
	before, err := transaction.DigestPath(target)
	if err != nil {
		t.Fatal(err)
	}
	const workers, failingRounds = 16, 4
	var calls atomic.Int32
	var rounds [failingRounds]chan struct{}
	for index := range rounds {
		rounds[index] = make(chan struct{})
	}
	restore := injectSharingRetry(t, func(path string) (string, error) {
		call := int(calls.Add(1))
		if call <= workers*failingRounds {
			round := rounds[(call-1)/workers]
			if call%workers == 0 {
				close(round)
			}
			// Each reader must reach this round before any can retry it.
			<-round
			return "", &os.PathError{Op: "open", Path: path, Err: errInjectedSharingViolation}
		}
		return transaction.DigestPath(path)
	})
	start := make(chan struct{})
	results := make([]string, workers)
	errorsByWorker := make([]error, workers)
	var group sync.WaitGroup
	for index := range results {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			<-start
			results[index], errorsByWorker[index] = Get(home, "internal/skill-a", repo, commit)
		}(index)
	}
	close(start)
	group.Wait()
	restore()
	for index := range results {
		if errorsByWorker[index] != nil || results[index] != target {
			t.Errorf("worker %d = %q, %v; want expected snapshot %q after four sharing failures", index, results[index], errorsByWorker[index], target)
		}
	}
	if got, want := calls.Load(), int32(workers*(failingRounds+1)); got != want {
		t.Errorf("destination digest calls = %d; want %d (four failures and one real digest per reader)", got, want)
	}
	assertSharingSnapshotUnchanged(t, target, before)
}

// Persistent sharing errors exhaust the retry bound; other read errors return
// immediately. Each refusal preserves the destination, removes staging, and
// permits a valid cached read once the injected fault is removed.
func TestGetSharingRetryRefusals(t *testing.T) {
	for _, tc := range []struct {
		name  string
		cause error
		calls int32
	}{
		{"persistent-sharing", errInjectedSharingViolation, destinationSharingViolationRetries + 1},
		{"permission-denied", fs.ErrPermission, 1},
		{"not-found", fs.ErrNotExist, 1},
		{"other-read-error", errors.New("injected read failure"), 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home, repo, commit, target := sharingRetryFixture(t)
			before, err := transaction.DigestPath(target)
			if err != nil {
				t.Fatal(err)
			}
			var calls atomic.Int32
			restore := injectSharingRetry(t, func(path string) (string, error) {
				calls.Add(1)
				return "", &os.PathError{Op: "open", Path: path, Err: tc.cause}
			})
			got, err := Get(home, "internal/skill-a", repo, commit)
			restore()
			if got != "" || !errors.Is(err, ErrDestinationConflict) {
				t.Fatalf("Get() = %q, %v; want no snapshot and destination conflict", got, err)
			}
			if got := calls.Load(); got != tc.calls {
				t.Fatalf("destination digest calls = %d; want %d", got, tc.calls)
			}
			assertSharingSnapshotUnchanged(t, target, before)
			if got, err := Get(home, "internal/skill-a", repo, commit); err != nil || got != target {
				t.Fatalf("Get() after removing fault = %q, %v; want %q", got, err, target)
			}
		})
	}
}
