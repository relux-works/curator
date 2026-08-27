package buildcache

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildmeta"
)

func TestInvalidInputsFailClosed(t *testing.T) {
	if _, err := New(""); err == nil {
		t.Fatal("empty manager home was accepted")
	}
	unclean := t.TempDir() + string(filepath.Separator) + "child" + string(filepath.Separator) + ".."
	if _, err := New(unclean); err == nil {
		t.Fatal("unclean absolute manager home was accepted")
	}

	store := newTestStore(t)
	invalidInput := testInput("tool")
	invalidInput.Driver = "other"
	if result := store.Inspect(Expectation{Input: invalidInput}); result.Status != Corrupt {
		t.Fatalf("invalid expected input = %+v", result)
	}
	if result := (*Store)(nil).Inspect(Expectation{}); result.Status != Unsupported {
		t.Fatalf("nil store = %+v", result)
	}
	for _, key := range []buildmeta.CacheKey{"", "sha256:xyz", buildmeta.CacheKey("sha256:" + strings.Repeat("A", 64))} {
		if _, _, err := store.paths(key); err == nil {
			t.Fatalf("malformed key %q was accepted", key)
		}
	}
	if got := (Result{Status: Status("future")}).DryRunOutcome(); got != "unsupported" {
		t.Fatalf("unknown dry-run outcome = %q", got)
	}
}

func TestInvalidPublicationsLeaveCacheUnpublished(t *testing.T) {
	store := newTestStore(t)
	valid, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
	tests := []struct {
		name   string
		mutate func(t *testing.T, publication *Publication)
	}{
		{
			name: "noncanonical receipt",
			mutate: func(_ *testing.T, publication *Publication) {
				publication.ReceiptBytes = append(publication.ReceiptBytes, '\n')
			},
		},
		{
			name: "missing artifact source",
			mutate: func(_ *testing.T, publication *Publication) {
				publication.ArtifactSource = filepath.Join(store.Home(), "absent")
			},
		},
		{
			name: "artifact source is directory",
			mutate: func(t *testing.T, publication *Publication) {
				dir := filepath.Join(store.Home(), "source-directory")
				if err := os.Mkdir(dir, 0o700); err != nil {
					t.Fatal(err)
				}
				publication.ArtifactSource = dir
			},
		},
		{
			name: "artifact bytes mismatch",
			mutate: func(t *testing.T, publication *Publication) {
				writeFile(t, publication.ArtifactSource, []byte("different"), 0o600)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			publication := valid
			publication.ReceiptBytes = append([]byte(nil), valid.ReceiptBytes...)
			test.mutate(t, &publication)
			if _, err := store.Publish(publication, testHomeLock{}); err == nil {
				t.Fatal("invalid publication succeeded")
			}
			if result := store.Inspect(Expectation{Input: valid.Input}); result.Status != Miss {
				t.Fatalf("invalid publication created cache state: %+v", result)
			}
		})
	}
}

func TestExplicitQuarantineMovesEntryAndMissingIsNoop(t *testing.T) {
	store := newTestStore(t)
	publication, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
	key, err := publication.Input.CacheKey()
	if err != nil {
		t.Fatal(err)
	}
	if path, err := store.Quarantine(key, testHomeLock{}); err != nil || path != "" {
		t.Fatalf("missing quarantine = %q, %v", path, err)
	}
	if _, err := store.Quarantine("bad-key", testHomeLock{}); err == nil {
		t.Fatal("quarantine accepted malformed logical key")
	}
	if _, err := store.Publish(publication, testHomeLock{}); err != nil {
		t.Fatal(err)
	}
	path, err := store.Quarantine(key, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if path == "" {
		t.Fatal("present entry did not get a quarantine path")
	}
	if info, err := os.Lstat(path); err != nil || !info.IsDir() {
		t.Fatalf("quarantine path = %q, %v", path, err)
	}
	if result := store.Inspect(Expectation{Input: publication.Input}); result.Status != Miss {
		t.Fatalf("quarantined entry remains live: %+v", result)
	}
}

func TestConflictErrorCarriesLogicalKey(t *testing.T) {
	key := buildmeta.CacheKey("sha256:" + strings.Repeat("1", 64))
	var conflict error = &ConflictError{Key: key}
	if !strings.Contains(conflict.Error(), string(key)) {
		t.Fatalf("conflict error = %q", conflict)
	}
	var typed *ConflictError
	if !errors.As(conflict, &typed) || typed.Key != key {
		t.Fatalf("typed conflict = %#v", typed)
	}
}
