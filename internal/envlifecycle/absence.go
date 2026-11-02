package envlifecycle

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/pathboundary"
)

// VerifyAbsence checks only action-specific absent installed/terminal namespaces.
// It neither classifies retained homes nor authorizes a continuation. In
// particular, a retained conflict still needs its complete preimage verifier.
func (l *Ledger) VerifyAbsence() error {
	if err := l.Validate(); err != nil {
		return err
	}
	if err := l.Manager.Verify(); err != nil {
		return err
	}
	for _, entry := range l.Entries {
		var paths []string
		if entry.Purged != nil {
			paths = entry.Purged.ExpectedAbsence
		}
		if entry.Removed != nil {
			paths = entry.Removed.ExpectedAbsence
		}
		for _, path := range paths {
			if err := verifyAbsent(l.Manager.Path, path); err != nil {
				return err
			}
		}
	}
	return nil
}
func verifyAbsent(manager, path string) error {
	if !below(manager, path) || path == manager {
		return invalid("absence outside manager")
	}
	if err := pathboundary.ValidateRouteWithOwner(manager, filepath.Dir(path), pathboundary.DefaultOwnerLookup()); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return invalid("absence parent boundary: %v", err)
	}
	_, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return invalid("absence metadata unreadable: %v", err)
	}
	return invalid("expected absent namespace has reappeared")
}
