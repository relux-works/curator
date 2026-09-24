//go:build windows

package snapshot

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/transaction"
	"golang.org/x/sys/windows"
)

func windowsSharingFixture(t *testing.T) (home, repo, commit string) {
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
	return t.TempDir(), repo, head.Commit
}

func replaceDestinationDigestForTest(replacement func(string) (string, error)) func() {
	previous := digestSnapshotDestination
	digestSnapshotDestination = replacement
	return func() { digestSnapshotDestination = previous }
}

func TestConcurrentGetRetriesInjectedWindowsSharingViolation(t *testing.T) {
	firstHome, repo, commit := windowsSharingFixture(t)
	const repetitions = 5
	for repetition := 0; repetition < repetitions; repetition++ {
		t.Run(fmt.Sprintf("iteration_%02d", repetition+1), func(t *testing.T) {
			home := firstHome
			if repetition > 0 {
				home = t.TempDir()
			}
			want := Dir(home, "internal/skill-a", commit)
			var injected atomic.Bool
			var destinationDigestCalls atomic.Int32
			restore := replaceDestinationDigestForTest(func(path string) (string, error) {
				if path == want {
					destinationDigestCalls.Add(1)
					if injected.CompareAndSwap(false, true) {
						return "", &os.PathError{Op: "open", Path: path, Err: windows.ERROR_SHARING_VIOLATION}
					}
				}
				return transaction.DigestPath(path)
			})
			defer restore()

			const workers = 16
			start := make(chan struct{})
			results := make([]string, workers)
			errorsByWorker := make([]error, workers)
			var group sync.WaitGroup
			for index := 0; index < workers; index++ {
				group.Add(1)
				go func(index int) {
					defer group.Done()
					<-start
					results[index], errorsByWorker[index] = Get(home, "internal/skill-a", repo, commit)
				}(index)
			}
			close(start)
			group.Wait()

			if !injected.Load() {
				t.Fatal("concurrent publication did not exercise the injected sharing violation")
			}
			if destinationDigestCalls.Load() < 2 {
				t.Fatalf("destination digest calls = %d, want retry after injected sharing violation", destinationDigestCalls.Load())
			}
			for index := range results {
				if errorsByWorker[index] != nil || results[index] != want {
					t.Fatalf("worker %d = %q, %v; want %q", index, results[index], errorsByWorker[index], want)
				}
			}
			payload, err := os.ReadFile(filepath.Join(want, "SKILL.md"))
			if err != nil || string(payload) != "concurrent" {
				t.Fatalf("published snapshot = %q, %v", payload, err)
			}
		})
	}
}

func TestGetFailsClosedAfterPersistentWindowsSharingViolation(t *testing.T) {
	home, repo, commit := windowsSharingFixture(t)
	if _, err := Get(home, "internal/skill-a", repo, commit); err != nil {
		t.Fatal(err)
	}
	want := Dir(home, "internal/skill-a", commit)
	var destinationDigestCalls atomic.Int32
	restore := replaceDestinationDigestForTest(func(path string) (string, error) {
		if path == want {
			destinationDigestCalls.Add(1)
			return "", &os.PathError{Op: "open", Path: path, Err: windows.ERROR_SHARING_VIOLATION}
		}
		return transaction.DigestPath(path)
	})
	defer restore()

	_, err := Get(home, "internal/skill-a", repo, commit)
	if !errors.Is(err, ErrDestinationConflict) {
		t.Fatalf("Get() error = %v; want destination conflict", err)
	}
	if got, wantCalls := destinationDigestCalls.Load(), int32(destinationSharingViolationRetries+1); got != wantCalls {
		t.Fatalf("destination digest calls = %d; want bounded %d attempts", got, wantCalls)
	}
}

func TestGetDoesNotRetryWindowsAccessDenied(t *testing.T) {
	home, repo, commit := windowsSharingFixture(t)
	if _, err := Get(home, "internal/skill-a", repo, commit); err != nil {
		t.Fatal(err)
	}
	want := Dir(home, "internal/skill-a", commit)
	var destinationDigestCalls atomic.Int32
	restore := replaceDestinationDigestForTest(func(path string) (string, error) {
		if path == want {
			destinationDigestCalls.Add(1)
			return "", &os.PathError{Op: "open", Path: path, Err: windows.ERROR_ACCESS_DENIED}
		}
		return transaction.DigestPath(path)
	})
	defer restore()

	_, err := Get(home, "internal/skill-a", repo, commit)
	if !errors.Is(err, ErrDestinationConflict) {
		t.Fatalf("Get() error = %v; want destination conflict", err)
	}
	if got := destinationDigestCalls.Load(); got != 1 {
		t.Fatalf("destination digest calls = %d; want access denied to fail without retry", got)
	}
	if !strings.Contains(err.Error(), windows.ERROR_ACCESS_DENIED.Error()) {
		t.Fatalf("Get() error = %v; want access-denied cause preserved", err)
	}
}
