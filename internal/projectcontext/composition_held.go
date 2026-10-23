package projectcontext

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/pathboundary"
)

// readHeld borrows the caller's exclusive lock. The concrete unforgeable
// witness must validate this exact manager home before and after state reads.
// It never reacquires, releases, provisions or saves manager state.
func (s Store) readHeld(held managerlock.Held) (*operation, error) {
	manager, err := managerlock.New(s.Home)
	if err != nil {
		return nil, err
	}
	if err = held.Validate(manager.Home()); err != nil {
		return nil, err
	}
	if err = pathboundary.ValidateRoot(manager.Home()); err != nil {
		return nil, err
	}
	op := &operation{root: filepath.Join(manager.Home(), "project-context"), precedence: s.precedence()}
	if err = pathboundary.Validate(op.root); err != nil {
		return nil, fail("project_context_drift", "protected admission boundary unavailable")
	}
	root, err := os.OpenRoot(op.root)
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()
	b, _, err := read(root, "state.json", MaxRecord)
	if err != nil {
		return nil, fail("project_context_unreadable", "protected admission state unavailable")
	}
	if err = decode(b, &op.state); err != nil {
		return nil, err
	}
	canonicalState, err := canonical(op.state)
	if err != nil || !bytes.Equal(canonicalState, b) {
		return nil, fail("project_context_drift", "protected admission state is not canonical")
	}
	if err = op.validate(); err != nil {
		return nil, err
	}
	if err = held.Validate(manager.Home()); err != nil {
		return nil, err
	}
	return op, nil
}
