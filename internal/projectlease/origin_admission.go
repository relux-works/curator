package projectlease

import "github.com/relux-works/curator/internal/envauthority"

// Logical profile references are effects despite a shared ledger pathname.
// Preserve exact global-retained preimages rather than requiring them to be
// launchable; new reservations independently pass the selected-profile gate.
func (m *Manager) admitLedger(l *ledger) error {
	raw, err := canonical(l)
	if err != nil {
		return err
	}
	return envauthority.AdmitLeaseReferenceBytes(m.home, raw)
}
