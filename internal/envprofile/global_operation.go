package envprofile

import (
	"fmt"
	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/transaction"
)

type borrowedGlobalLock struct{ transaction.HomeLock }

func (borrowedGlobalLock) Close() error { return nil }
func (lock borrowedGlobalLock) AssertCurrent() error {
	if current, ok := lock.HomeLock.(interface{ AssertCurrent() error }); ok {
		return current.AssertCurrent()
	}
	return fmt.Errorf("global operation requires current manager lock")
}

type globalInPlaceConflict struct {
	error
	capture envlifecycle.ConflictReceipt
}
