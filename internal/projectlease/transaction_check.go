package projectlease

import (
	"fmt"

	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/transaction"
)

// TransactionCheck supplies the project-lease policy to every manager engine
// that may commit or recover these journals, including install and CLI GC.
// Other transaction owners retain their existing behavior. The current owner
// must retain the same-home lock throughout commit and recovery.
func TransactionCheck(home string) transaction.CommitCheck {
	return func(lock transaction.HomeLock, journal transaction.Journal) error {
		if journal.ProjectIdentity != "project-leases" {
			return nil
		}
		witness, ok := lock.(interface{ Held() managerlock.Held })
		if !ok {
			return managerlock.ErrNotHeld
		}
		if err := witness.Held().Validate(home); err != nil {
			return err
		}
		canonical, err := managerlock.New(home)
		if err != nil {
			return err
		}
		m := &Manager{home: canonical.Home()}
		return m.verifyTransactionArtifacts(journal)
	}
}

// The desired ledger remains in its staged file until installed. An interrupted
// install may already have moved it to LivePath before recording target state.
// Only bytes matching the journal's desired digest supply pending references;
// an absent/mismatched candidate cannot be replaced by the old live ledger.
func (m *Manager) verifyTransactionArtifacts(journal transaction.Journal) error {
	for _, target := range journal.Targets {
		if target.LivePath != statePath(m.home) || target.Class != "project-lease" {
			continue
		}
		for _, path := range []string{target.StagedPath, target.LivePath} {
			if path == "" {
				continue
			}
			if err := checkedRoute(m.home, path); err != nil {
				return err
			}
			got, err := transaction.DigestTarget(target.Kind, path)
			if err != nil {
				return err
			}
			if got != target.DesiredDigest {
				continue
			}
			raw, err := read(m.home, path)
			if err != nil {
				return err
			}
			l, err := decodeLedger(raw)
			if err != nil {
				return err
			}
			for _, e := range l.Leases {
				if err = m.verifyArtifacts(e.Composition); err != nil {
					return err
				}
			}
			for _, v := range l.Views {
				if err = m.verifyArtifacts(v.Composition); err != nil {
					return err
				}
			}
			return nil
		}
		return fmt.Errorf("pending project lease desired ledger is unavailable")
	}
	return fmt.Errorf("pending project lease transaction lacks its ledger")
}
