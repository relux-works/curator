package envprofile

import (
	"fmt"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/identifiers"
)

// VisitCompositionProfile retains default local verification for existing
// callers. Configured callers use VisitCompositionProfileWithPolicy.
func VisitCompositionProfile(home, name, expected string, visit func(*contextlock.Lock, contextlock.Member, string) error) error {
	_, err := VisitCompositionProfileWithPolicy(home, name, expected, Policy{}, visit)
	return err
}

// VisitCompositionProfileWithPolicy reads the explicit protected selection and
// every required observation before applying policy. It acquires no source,
// synthesizes no evidence and publishes no home. The caller holds the home lock.
// Source generation must remain stable even when initial mismatches are allowed.
func VisitCompositionProfileWithPolicy(home, name, expected string, policy Policy, visit func(*contextlock.Lock, contextlock.Member, string) error) ([]string, error) {
	if !identifiers.Valid(name) || expected == "" || visit == nil {
		return nil, fmt.Errorf("invalid composition profile selection")
	}
	var initialGeneration []string
	check := func(initial bool) (*contextlock.Lock, []string, error) {
		if err := envauthority.Validate(home); err != nil {
			return nil, nil, err
		}
		if err := validateLockBoundary(home, name, nil); err != nil {
			return nil, nil, err
		}
		lock, hash, err := readLock(home, name)
		if err != nil {
			return nil, nil, err
		}
		if strings.TrimPrefix(hash, "sha256:") != expected {
			return nil, nil, fmt.Errorf("selected profile lock changed")
		}
		if failure := validateNamedStoreBoundaries(home, lock, nil); failure != nil {
			return nil, nil, failure.Err
		}
		checks, err := startupChecks(home, lock)
		if err != nil {
			return nil, nil, err
		}
		verdict, err := startupVerdict(policy, lock, checks)
		if err != nil {
			return nil, nil, err
		}
		if !verdict.Allow {
			return nil, verdict.Warnings, fmt.Errorf("%s: composition verification refused", strings.Join(verdict.Refusals, "; "))
		}
		generation := make([]string, 0, len(lock.Members))
		for _, member := range lock.Members {
			digest, err := hashing.ContentSHA256WithVersion(memberEntryPath(home, member), map[string]bool{}, hashing.VersionV2)
			if err != nil {
				return nil, verdict.Warnings, err
			}
			generation = append(generation, digest)
		}
		if initial {
			initialGeneration = generation
		} else if !equalStrings(initialGeneration, generation) {
			return nil, verdict.Warnings, fmt.Errorf("composition profile source generation changed")
		}
		return lock, verdict.Warnings, nil
	}
	lock, warnings, err := check(true)
	if err != nil {
		return warnings, err
	}
	manager := newGitManager(home)
	for _, member := range lock.Members {
		entry := manager.entryPath(home, resolvedOf(member))
		cleanup := func() {}
		if member.Commit != "" {
			var err error
			entry, cleanup, _, err = extractPinnedSnapshot(home, member)
			if err != nil {
				return warnings, err
			}
		}
		root, err := selectedPackageRoot(entry, member.Directory)
		if err == nil {
			err = visit(lock, member, root)
		}
		cleanup()
		if err != nil {
			return warnings, err
		}
		if err := envauthority.Validate(home); err != nil {
			return warnings, err
		}
	}
	_, _, err = check(false)
	return warnings, err
}
