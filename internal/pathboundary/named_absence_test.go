package pathboundary

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/privatedir"
)

// Named targets (lock, marker, store entry) fail closed when missing; only
// children discovered by the tree walk may vanish.
func TestNamedRoutesRejectMissingTarget(t *testing.T) {
	root := privateRoot(t)
	if err := os.Mkdir(filepath.Join(root, "store"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{
		filepath.Join(root, "missing"),
		filepath.Join(root, "store", "missing"),
		filepath.Join(root, "missing", "entry"),
	} {
		for name, validate := range map[string]func(string, string, OwnerLookup) error{
			"within": ValidateWithinWithOwner,
			"route":  ValidateRouteWithOwner,
			"leaf":   ValidateLeafWithOwner,
		} {
			err := validate(root, target, DefaultOwnerLookup())
			var failure *Failure
			if !errors.As(err, &failure) || failure.Check != CheckRegular || !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("%s(%s) = %v, want *Failure %s wrapping ErrNotExist", name, target, err, CheckRegular)
			}
		}
	}
}

// The walked target itself is named by the caller: if it is gone when the
// walk starts, validateTree fails instead of treating it as vanished.
func TestValidateTreeRejectsMissingWalkTarget(t *testing.T) {
	root := privateRoot(t)
	operator, err := effectiveOwner()
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "entry")
	err = validateTree(root, target, operator, DefaultOwnerLookup(), defaultEntryInfo)
	var failure *Failure
	if !errors.As(err, &failure) || failure.Path != target || failure.Check != CheckRegular || !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("validateTree(missing target) = %v, want *Failure %s wrapping ErrNotExist", err, CheckRegular)
	}
}

// privateRoot carries the owner-only protected DACL on Windows, so the
// enclosing-root check passes and the missing target is the first failure.
func privateRoot(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "root")
	if err := privatedir.MakeAll(root); err != nil {
		t.Fatal(err)
	}
	return root
}
