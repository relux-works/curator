package projectlease

import (
	"context"
	"errors"
	"github.com/relux-works/curator/internal/transaction"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type artifactLossSessions struct {
	Sessions
	remove func()
}

func (s *artifactLossSessions) Observe(ctx context.Context, ref string, generation int64) (Evidence, error) {
	e, err := s.Sessions.Observe(ctx, ref, generation)
	if s.remove != nil {
		remove := s.remove
		s.remove = nil
		remove()
	}
	return e, err
}

// Artifact presence is checked again after authority/session callbacks and
// before lease publication. Missing, linked or non-directory runtime roots
// yield the typed refusal without allocating the lease namespace; restoring
// the root admits a retry. The later commit gate cannot replace this preflight.
// This checks presence/boundaries, not the authority's receipt/hash proof.
func TestAcquireArtifactsBeforeCommit(t *testing.T) {
	for _, kind := range []string{"present", "removed-during-observe", "missing", "file", "link"} {
		t.Run(kind, func(t *testing.T) {
			f := setup(t)
			target := filepath.Join(f.home, "runtime", "tool", "revision")
			if kind != "present" && kind != "removed-during-observe" {
				if err := os.Remove(target); err != nil {
					t.Fatal(err)
				}
			}
			switch kind {
			case "removed-during-observe":
				f.m.sessions = &artifactLossSessions{Sessions: f.a, remove: func() {
					if err := os.Remove(target); err != nil {
						t.Fatal(err)
					}
				}}
			case "file":
				if err := os.WriteFile(target, []byte("not an artifact directory"), 0600); err != nil {
					t.Fatal(err)
				}
			case "link":
				if err := os.Symlink(t.TempDir(), target); err != nil {
					t.Skipf("symlink unavailable: %v", err)
				}
			}
			// Observe publication itself: a later commit rejection and clean
			// rollback cannot stand in for the required earlier artifact preflight.
			publicationEvents := 0
			f.m.hooks = transaction.Hooks{Fault: func(transaction.Event) error { publicationEvents++; return nil }}
			r := f.reserve("one")
			_, err := f.m.Acquire(context.Background(), r)
			if kind == "present" {
				if err != nil {
					t.Fatal("present artifact refused", err)
				}
				if publicationEvents == 0 {
					t.Fatal("present artifact did not exercise publication observation")
				}
				return
			}
			if !errors.Is(err, ErrArtifactsUnavailable) {
				t.Fatalf("missing artifact admission = %v, want typed refusal", err)
			}
			if publicationEvents != 0 {
				t.Fatal("missing artifact admission entered publication")
			}
			if HasState(f.home) {
				t.Fatal("missing artifact admission allocated lease namespace")
			}
			for _, path := range []string{bindingPath(f.home, "one"), statePath(f.home), filepath.Join(f.home, homeRelative(f.c))} {
				if _, err = os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
					t.Fatal("artifact refusal published state", err)
				}
			}
			if kind == "file" || kind == "link" {
				if err = os.Remove(target); err != nil {
					t.Fatal(err)
				}
			}
			if err = os.Mkdir(target, 0700); err != nil {
				t.Fatal(err)
			}
			if _, err = f.m.Acquire(context.Background(), r); err != nil {
				t.Fatal("restored artifact retry refused", err)
			}
			if publicationEvents == 0 {
				t.Fatal("restored artifact did not exercise publication observation")
			}
		})
	}
}

// Each supported GC build namespace supplies a present entry. Removing that
// entry refuses the next lease, preserves the first binding and creates no
// replacement binding. Full cache receipt validation stays with Authority.
func TestAcquireBuildArtifactNamespaces(t *testing.T) {
	for _, namespace := range []string{"cache/build/go-v1", "cache/build/go-v1-receipt-3", "external-build-cache/artifacts", "external-build-cache/artifacts-receipt-3"} {
		t.Run(namespace, func(t *testing.T) {
			f := setup(t)
			key := strings.TrimPrefix(f.c.BuildKeys[0], "sha256:")
			original := filepath.Join(f.home, "cache", "build", "go-v1", key)
			target := filepath.Join(f.home, filepath.FromSlash(namespace), key)
			if original != target {
				if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Rename(original, target); err != nil {
					t.Fatal(err)
				}
			}
			acquire(t, f, "one")
			before, err := os.ReadFile(bindingPath(f.home, "one"))
			if err != nil {
				t.Fatal(err)
			}
			if err = os.Remove(target); err != nil {
				t.Fatal(err)
			}
			if _, err = f.m.Acquire(context.Background(), f.reserve("two")); !errors.Is(err, ErrArtifactsUnavailable) {
				t.Fatalf("missing build admitted lease: %v", err)
			}
			after, err := os.ReadFile(bindingPath(f.home, "one"))
			if err != nil || string(before) != string(after) {
				t.Fatal("missing build changed existing binding", err)
			}
			if _, err = os.Lstat(bindingPath(f.home, "two")); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("missing build published replacement binding", err)
			}
		})
	}
}
