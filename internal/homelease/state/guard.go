package state

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/managerlock"
)

// GuardMutation is the transaction/recovery entry gate. The home-leases owner
// publishes its own atomic lifecycle state, rather than rewriting a leased
// home. Its journal is still subject to the engine's ownership, preimage and
// canonical-journal checks. Other owners must present the real held witness.
func GuardMutation(home string, lock interface{ AssertHeld() error }, owner string, targets []string) error {
	if !HasState(home) {
		return nil
	}
	if lock == nil {
		return managerlock.ErrNotHeld
	}
	if err := lock.AssertHeld(); err != nil {
		return err
	}
	witness, ok := lock.(interface{ Held() managerlock.Held })
	if !ok {
		return managerlock.ErrNotHeld
	}
	if err := witness.Held().Validate(home); err != nil {
		return err
	}
	if owner == "home-leases" {
		return nil
	}
	return Guard(home, witness.Held(), targets)
}

// Guard checks the complete target set before the first write. Nil targets
// denotes all manager-owned homes; an empty nonnil set denotes no writes.
// Callers retain and revalidate this same witness through publication.
func Guard(home string, held managerlock.Held, targets []string) error {
	if err := held.Validate(home); err != nil {
		return err
	}
	l, err := Load(home)
	if err != nil {
		return err
	}
	return GuardLedger(l, targets)
}

func GuardLedger(l *Ledger, targets []string) error {
	blocked := []Record{}
	for _, record := range Records(l) {
		if !record.Blocking() {
			continue
		}
		entry := l.Leases[record.LeaseID]
		overlaps := targets == nil
		for _, target := range targets {
			for _, protected := range entry.Admission.Protected {
				match, err := Overlap(target, protected)
				if err != nil {
					return Refuse("environment_lease_unreadable", "state_unreadable")
				}
				overlaps = overlaps || match
			}
		}
		if overlaps {
			blocked = append(blocked, record)
		}
	}
	if len(blocked) == 0 {
		return nil
	}
	err := Refuse("environment_home_leased", "blocking_holders")
	err.Details.Holders, err.Details.Count = blocked, len(blocked)
	return err
}

// Overlap uses physical identity for existing paths and canonical manager path
// rules for absent suffixes. It does not fold case independently of the host.
func Overlap(a, b string) (bool, error) {
	if !filepath.IsAbs(a) || !filepath.IsAbs(b) || filepath.Clean(a) != a || filepath.Clean(b) != b {
		return false, fmt.Errorf("ambiguous lease path")
	}
	ai, ae := os.Stat(a)
	bi, be := os.Stat(b)
	if ae == nil && be == nil && os.SameFile(ai, bi) {
		return true, nil
	}
	// A non-existing child of an existing case alias still belongs to that
	// physical directory. Comparing cleaned path spellings would miss it on
	// case-insensitive filesystems; inspect actual ancestor identity instead.
	for _, pair := range [][2]string{{a, b}, {b, a}} {
		base, err := os.Stat(pair[1])
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return false, err
		}
		for parent := pair[0]; ; parent = filepath.Dir(parent) {
			info, err := os.Stat(parent)
			if err == nil && os.SameFile(base, info) {
				return true, nil
			}
			if err != nil && !os.IsNotExist(err) {
				return false, err
			}
			if parent == filepath.Dir(parent) {
				break
			}
		}
	}
	ca, err := managerlock.New(a)
	if err != nil {
		return false, err
	}
	cb, err := managerlock.New(b)
	if err != nil {
		return false, err
	}
	a, b = ca.Home(), cb.Home()
	return contains(a, b) || contains(b, a), nil
}

func contains(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
