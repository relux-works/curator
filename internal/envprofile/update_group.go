package envprofile

import (
	"fmt"
	"sort"

	"github.com/relux-works/curator/internal/contextlock"
)

// publishUpdatedScopes owns the entire ordinary update/reinstall invocation.
// Native planning uses the candidate lock without exposing it at its live name.
// A refusal or later scoped fault therefore cannot strand a committed L1.
func publishUpdatedScopes(op *operation, home, name string, source Source, next *contextlock.Lock, files map[string][]byte, activate bool, policy Policy) (_ []EntryResult, resultErr error) {
	if activate {
		if err := policy.CheckMachineUse(name); err != nil {
			return nil, err
		}
	}
	machine, err := Current(home)
	if err != nil {
		return nil, err
	}
	scopes, err := ScopedCurrents(home)
	if err != nil {
		return nil, err
	}
	g := newHomePublicationGroup()
	defer finishPublicationCleanup(&resultErr, g.close)
	if op.publicationGroup != nil {
		return nil, fmt.Errorf("ordinary update cannot nest publication groups")
	}
	op.publicationGroup = g
	defer func() { op.publicationGroup = nil }()
	if _, err := op.publishManagedRebind(home, files, nil, nil, nil); err != nil {
		return nil, err
	}
	canonical, err := next.Canonical()
	if err != nil {
		return nil, err
	}
	hash := contextlock.HashBytes(canonical)
	var results []EntryResult
	plan := func(environment string) error {
		rows, err := materializeScopeObserved(op, home, name, environment, policy, nil, source, next, hash)
		results = append(results, rows...)
		if err != nil {
			return err
		}
		for _, row := range rows {
			if !row.OK {
				return fmt.Errorf("%s: update scope planning refused; lock, homes and selection are unchanged", DiagUsePartial)
			}
		}
		return nil
	}
	if machine == name || activate {
		if err := plan(""); err != nil {
			return results, err
		}
	}
	keys := make([]string, 0, len(scopes))
	for scope, current := range scopes {
		if current == name {
			keys = append(keys, scope)
		}
	}
	sort.Strings(keys)
	for _, scope := range keys {
		environment, target := splitScope(scope)
		if target != "" {
			return results, fmt.Errorf("secondary fixed-home target %q writes are deferred", target)
		}
		if err := plan(environment); err != nil {
			return results, err
		}
	}
	var removals []string
	if activate {
		activation, err := planScopePublication(home, "", name, false)
		if err != nil {
			return results, err
		}
		for path, body := range activation.files {
			files[path] = body
		}
		removals = activation.removals
	}
	if g.unit != nil {
		return results, op.commitHomeGroup(home, g, files, removals, nil, nil)
	}
	// No materialized native/managed surface exists; the ordinary record owner
	// still performs strict whole-manager effect admission.
	op.publicationGroup = nil
	return results, op.publish(files, removals...)
}
