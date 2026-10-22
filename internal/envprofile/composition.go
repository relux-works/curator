package envprofile

import (
	"fmt"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/identifiers"
)

// VisitCompositionProfile reads only the explicitly selected lock and its
// complete protected closure. The caller holds the manager home lock and must
// copy any needed bytes during visit. No machine-current fallback, acquisition,
// repair or managed-home publication occurs here. Pins and boundaries are
// checked on both sides of the visitor, including members it does not consume.
func VisitCompositionProfile(home, name, expected string, visit func(*contextlock.Lock, contextlock.Member, string) error) error {
	if !identifiers.Valid(name) || expected == "" || visit == nil {
		return fmt.Errorf("invalid composition profile selection")
	}
	check := func() (*contextlock.Lock, error) {
		if err := validateLockBoundary(home, name, nil); err != nil {
			return nil, err
		}
		lock, hash, err := readLock(home, name)
		if err != nil {
			return nil, err
		}
		if strings.TrimPrefix(hash, "sha256:") != expected {
			return nil, fmt.Errorf("selected profile lock changed")
		}
		if failure := validateNamedStoreBoundaries(home, lock, nil); failure != nil {
			return nil, failure.Err
		}
		if failure := validateNamedStorePins(home, lock); failure != nil {
			return nil, failure.Err
		}
		return lock, nil
	}
	lock, err := check()
	if err != nil {
		return err
	}
	manager := newGitManager(home)

	for _, member := range lock.Members {
		entry := manager.entryPath(home, resolvedOf(member))
		cleanup := func() {}
		if member.Commit != "" {
			var e error
			entry, cleanup, _, e = extractPinnedSnapshot(home, member)
			if e != nil {
				return e
			}
		}
		root, e := selectedPackageRoot(entry, member.Directory)
		if e == nil {
			e = visit(lock, member, root)
		}
		cleanup()
		if e != nil {
			return e
		}
	}

	_, err = check()
	return err
}
