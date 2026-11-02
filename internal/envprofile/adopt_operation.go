package envprofile

import (
	"context"
	"fmt"

	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/projectlease"
)

// Adoption alone can acquire the mutation lock for one attested legacy prefix.
// That read-side intent never authorizes another journal: recovery still uses
// the common strict scan, which needs a real complete pending unit for a legacy
// home. The actual adoption publication later binds all targets in Prepare.
func beginAdoptOperation(home string, intent *envunit.Unit, validate func() error) (*operation, error) {
	if err := envauthority.ValidateForUnit(home, intent); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), lockTimeout)
	defer cancel()
	manager, err := managerlock.New(home)
	if err != nil {
		return nil, err
	}
	lock, err := manager.AcquireHomeOnly(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("%s: acquire adoption lock: %w", DiagLockUnavailable, err)
	}
	op := &operation{lock: lock}
	fail := func(err error) (*operation, error) { _ = op.close(); return nil, err }
	if err := envauthority.ValidateForUnit(home, intent); err != nil {
		return fail(err)
	}
	if err := validate(); err != nil {
		return fail(err)
	}
	op.engine, err = newOperationEngine(home)
	if err != nil {
		return fail(err)
	}
	pending, err := op.engine.PendingRecovery(lock)
	if err != nil {
		return fail(err)
	}
	if pending {
		if err := envauthority.Validate(home); err != nil {
			return fail(fmt.Errorf("%w; unrelated legacy recovery cannot borrow an adoption intent", err))
		}
		op.recovered, err = projectlease.RecoverUnderLock(home, lock, op.engine)
		if err != nil {
			return fail(err)
		}
	}
	if err := envauthority.ValidateForUnit(home, intent); err != nil {
		return fail(op.recoveryRefusal(err))
	}
	return op, nil
}
