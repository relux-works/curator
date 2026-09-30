package envprofile

import (
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/pathboundary"
)

// A lock-named store entry that is missing after lock creation is refused
// at the named-route boundary step (validateNamedStoreBoundaries), not only
// later by pin recomputation: the diagnostic names the regular_types check
// and never pin_hash. A vanishing-skip on the named route would let the
// entry through the boundary and surface only as pin_hash.
func TestResolveRefusesMissingLockNamedStoreEntryAtBoundary(t *testing.T) {
	fx := storeBoundaryFixture(t, "current", false)
	entry := storeBoundaryEntry(t, fx, contextlock.KindContext)
	if err := os.RemoveAll(entry); err != nil {
		t.Fatal(err)
	}

	lock, _, err := contextlock.Read(lockPath(fx.home, fx.profile))
	if err != nil {
		t.Fatal(err)
	}
	failure := validateNamedStoreBoundaries(fx.home, lock, nil)
	if failure == nil {
		t.Fatalf("boundary step admitted missing lock-named entry %s", entry)
	}
	if failure.Path != entry || failure.Check != pathboundary.CheckRegular || !errors.Is(failure.Err, fs.ErrNotExist) {
		t.Fatalf("boundary failure = %+v, want %s on %s wrapping ErrNotExist", failure, pathboundary.CheckRegular, entry)
	}

	before := hashTreeForTest(t, fx.home)
	result, err := Resolve(fx.request(envregistry.CodexCLI))
	if err == nil {
		t.Fatalf("Resolve admitted missing lock-named entry: %+v", result)
	}
	message := err.Error()
	if !strings.Contains(message, envregistry.DiagStoreUntrusted) ||
		!strings.Contains(message, "failed "+pathboundary.CheckRegular+" check") ||
		strings.Contains(message, "failed "+diagPinHash+" check") {
		t.Fatalf("Resolve error = %v, want %s from the boundary %s check, not %s", err, envregistry.DiagStoreUntrusted, pathboundary.CheckRegular, diagPinHash)
	}
	if result != nil && len(result.Document) != 0 {
		t.Fatalf("untrusted Resolve emitted fragment %s", result.Document)
	}
	if after := hashTreeForTest(t, fx.home); after != before {
		t.Fatal("refused Resolve mutated manager state")
	}
}
