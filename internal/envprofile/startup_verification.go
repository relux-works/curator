package envprofile

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/skillintegrity"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/startupverification"
)

func (policy Policy) startupBaseline() int {
	if policy.StartupVerificationLevel == 0 {
		return startupverification.DefaultLevel
	}
	return policy.StartupVerificationLevel
}

func validateNamedStorePinsForRepair(req ResolveRequest, lock *contextlock.Lock) *storeEntryFailure {
	if !req.Repair || req.Policy.startupBaseline() < 3 {
		return nil
	}
	return validateNamedStorePins(req.Home, lock)
}

// prepareStartupEvidence runs only during explicit adoption/update. The complete
// candidate is prepared before replacing its lock/hash; no source file is edited.
func prepareStartupEvidence(home string, manager *gitManager, result *contextresolve.Result, policy Policy) error {
	var floors []int
	keys := make([]string, 0, len(result.Members))
	for key := range result.Members {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		p := result.Members[key].Package
		if p != nil && p.StartupVerificationLevel != nil {
			floors = append(floors, *p.StartupVerificationLevel)
		}
	}
	level, err := startupverification.EffectiveLevel(policy.startupBaseline(), floors...)
	if err != nil {
		return err
	}
	candidate := *result.Lock
	entries := []skillintegrity.Entry{}
	budget := skillintegrity.NewBudget()
	for _, member := range candidate.Members {
		if member.Kind != contextlock.KindSkill {
			continue
		}
		resolved := result.Members[contextresolve.Key(member.Kind, member.Name)]
		// Preserve the existing source/audit refusal before introducing new
		// evidence. The publication path still revalidates before storing.
		if _, err := auditMember(home, manager, contextresolve.Key(member.Kind, member.Name), resolved, policy, false); err != nil {
			return err
		}
		entry, cleanup, err := manager.inspectEntry(home, resolved)
		if err != nil {
			return err
		}
		root, err := selectedPackageRoot(entry, member.Directory)
		if err != nil {
			cleanup()
			return err
		}
		evidence, err := skillintegrity.Build(context.Background(), root, contextlock.SkillIdentity(member), budget)
		cleanup()
		if err != nil {
			return err
		}
		entries = append(entries, evidence)
	}
	if len(entries) == 0 && len(floors) == 0 {
		return nil
	}
	candidate.SchemaVersion = contextlock.SkillIntegritySchemaVersion
	candidate.HashVersion = 2
	candidate.SkillIntegrity = entries
	if len(floors) > 0 {
		candidate.StartupVerificationLevel = &level
	}
	canonical, err := candidate.Canonical()
	if err != nil {
		return err
	}
	result.Lock = &candidate
	result.LockHash = contextlock.HashBytes(canonical)
	return nil
}

// startupChecks reads the installed evidence; it never synthesizes or changes
// it. Hard refusals remain errors, while known verification outcomes are typed.
func startupChecks(home string, lock *contextlock.Lock) ([]startupverification.Observation, error) {
	checks := []startupverification.Observation{}
	budget := skillintegrity.NewBudget()
	index := 0
	for _, member := range lock.Members {
		entry := memberEntryPath(home, member)
		if member.Kind == contextlock.KindSkill {
			if err := skillintegrity.ValidateLegacyBoundary(entry); err != nil {
				if skillintegrity.IsCode(err, "skill_integrity_boundary_invalid") || !skillintegrity.IsCode(err, "skill_integrity_unavailable") {
					return nil, err
				}
				checks = append(checks, startupverification.Observation{Status: startupverification.Unavailable, Code: "profile_boundary_unavailable"})
			}
		}
		actual, expected, err := storeEntryPinHashes(home, member, entry, lock.ContentHashVersion())
		if err != nil {
			var unavailable *pinUnavailable
			var fileError *os.PathError
			if !errors.As(err, &unavailable) && !errors.As(err, &fileError) {
				return nil, &storeEntryFailure{Member: member, Path: entry, Check: diagPinHash, Err: err}
			}
			detail := fmt.Sprintf("%s: store entry %s/%s failed %s check", envregistry.DiagStoreUntrusted, member.Kind, member.Name, diagPinHash)
			if unavailable != nil {
				detail += ": pinned commit object unavailable"
			}
			checks = append(checks, startupverification.Observation{Status: startupverification.Unavailable, Code: detail})
		} else if actual != expected {
			checks = append(checks, startupverification.Observation{Status: startupverification.Failed, Code: fmt.Sprintf("%s: store entry %s/%s failed %s check", envregistry.DiagStoreUntrusted, member.Kind, member.Name, diagPinHash)})
		} else {
			checks = append(checks, startupverification.Observation{Status: startupverification.Passed, Code: "profile_pin_verified"})
		}
		if member.Kind != contextlock.KindSkill {
			continue
		}
		root, err := selectedPackageRoot(entry, member.Directory)
		if err != nil {
			return nil, err
		}
		if lock.SchemaVersion != contextlock.SkillIntegritySchemaVersion {
			// Legacy native declarations retain the existing parser and pin
			// algorithm. A manifestless legacy lock does not acquire evidence.
			if _, err := skillspec.Load(root); err != nil {
				var fileError *os.PathError
				if !errors.As(err, &fileError) {
					return nil, err
				}
				checks = append(checks, startupverification.Observation{Status: startupverification.Unavailable, Code: "skill_integrity_unavailable"})
			} else if skillspec.ManifestSourcePath(root) == "" {
				checks = append(checks, startupverification.Observation{Status: startupverification.Unavailable, Code: "skill_integrity_evidence_missing"})
			} else {
				checks = append(checks, startupverification.Observation{Status: startupverification.Passed, Code: "skill_integrity_native_verified"})
			}
			continue
		}
		if index >= len(lock.SkillIntegrity) {
			return nil, fmt.Errorf("skill_integrity_evidence_missing: incomplete lock")
		}
		evidence := lock.SkillIntegrity[index]
		index++
		err = skillintegrity.Verify(context.Background(), root, contextlock.SkillIdentity(member), evidence, budget)
		if err == nil {
			checks = append(checks, startupverification.Observation{Status: startupverification.Passed, Code: "skill_integrity_verified"})
			continue
		}
		for _, code := range []string{"skill_integrity_boundary_invalid", "skill_integrity_manifest_invalid", "skill_integrity_unsupported_version"} {
			if skillintegrity.IsCode(err, code) {
				return nil, err
			}
		}
		failed := skillintegrity.IsCode(err, "skill_integrity_mismatch")
		unavailable := skillintegrity.IsCode(err, "skill_integrity_unavailable")
		if failed {
			checks = append(checks, startupverification.Observation{Status: startupverification.Failed, Code: "skill_integrity_mismatch"})
		}
		if unavailable {
			checks = append(checks, startupverification.Observation{Status: startupverification.Unavailable, Code: "skill_integrity_unavailable"})
		}
		if !failed && !unavailable {
			return nil, err
		}
	}
	return checks, nil
}

type pinUnavailable struct{ cause error }

func (e *pinUnavailable) Error() string { return e.cause.Error() }
func (e *pinUnavailable) Unwrap() error { return e.cause }

func startupVerdict(policy Policy, lock *contextlock.Lock, checks []startupverification.Observation) (startupverification.Verdict, error) {
	var floors []int
	if lock.StartupVerificationLevel != nil {
		floors = append(floors, *lock.StartupVerificationLevel)
	}
	level, err := startupverification.EffectiveLevel(policy.startupBaseline(), floors...)
	if err != nil {
		return startupverification.Verdict{}, err
	}
	// Centre metadata has no admitted provider protocol in this release.
	return startupverification.Decide(level, checks, startupverification.CentreAbsent)
}
