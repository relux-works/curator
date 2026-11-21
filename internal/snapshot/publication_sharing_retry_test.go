package snapshot

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/transaction"
)

// An initially absent snapshot has one publisher and 15 authenticating callers.
// Each caller recovers from four synchronized sharing failures, then digests the
// real complete tree. Portable injection does not prove native Windows timing.
func TestConcurrentUncachedGetRetriesAuthenticateDestinationSharingViolation(t *testing.T) {
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	for name, payload := range map[string]string{
		"SKILL.md":         "concurrent publication",
		"runtime/main.txt": "runtime payload",
		"build/source.txt": "build payload",
		"nested/extra.txt": "complete tree authentication",
	} {
		path := filepath.Join(repo, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitRun(t, repo, "add", ".")
	gitRun(t, repo, "commit", "-qm", "one")
	head, err := gitops.Resolve(repo, "revision", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	expected := t.TempDir()
	if err := os.Chmod(expected, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := gitops.Extract(repo, head.Commit, expected); err != nil {
		t.Fatal(err)
	}
	expectedDigest, err := transaction.DigestPath(expected)
	if err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	target := Dir(home, "internal/skill-a", head.Commit)
	if _, err := os.Lstat(target); !os.IsNotExist(err) {
		t.Fatalf("initial destination must be absent, got %v", err)
	}

	const (
		workers        = 16
		authenticators = workers - 1
		failingRounds  = 4
	)
	var calls, realDigests, wrongPaths atomic.Int32
	var rounds [failingRounds]chan struct{}
	for index := range rounds {
		rounds[index] = make(chan struct{})
	}
	restore := injectSharingRetry(t, func(path string) (string, error) {
		if path != target {
			wrongPaths.Add(1)
		}
		call := int(calls.Add(1))
		if call <= authenticators*failingRounds {
			round := rounds[(call-1)/authenticators]
			if call%authenticators == 0 {
				close(round)
			}
			// The publisher never calls this seam. All 15 other callers must
			// reach each round; old three-retry callers finish the fourth
			// round with an error, so this barrier also completes on old code.
			<-round
			return "", &os.PathError{Op: "open", Path: path, Err: errInjectedSharingViolation}
		}
		digest, err := transaction.DigestPath(path)
		if err == nil {
			realDigests.Add(1)
		}
		return digest, err
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
			results[index], errorsByWorker[index] = Get(home, "internal/skill-a", repo, head.Commit)
		}(index)
	}
	close(start)
	group.Wait()
	restore()
	for index := range results {
		if errorsByWorker[index] != nil || results[index] != target {
			t.Errorf("worker %d authenticate-destination = %q, %v; want expected published snapshot %q after four sharing failures", index, results[index], errorsByWorker[index], target)
		}
	}
	if got, want := calls.Load(), int32(authenticators*(failingRounds+1)); got != want {
		t.Errorf("publication destination digest calls = %d; want %d (four failures and one real digest per authenticator)", got, want)
	}
	if got := realDigests.Load(); got != authenticators {
		t.Errorf("real publication destination digests = %d; want %d", got, authenticators)
	}
	if got := wrongPaths.Load(); got != 0 {
		t.Errorf("digest calls outside published destination = %d; want 0", got)
	}
	assertSharingSnapshotUnchanged(t, target, expectedDigest)
	if got, err := Get(home, "internal/skill-a", repo, head.Commit); err != nil || got != target {
		t.Fatalf("Get() after concurrent publication = %q, %v; want %q", got, err, target)
	}
	assertSharingSnapshotUnchanged(t, target, expectedDigest)
}
