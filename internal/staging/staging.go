// Package staging describes operation-private replacements that the
// transaction layer publishes atomically.
//
// A producer writes desired bytes below an operation-private stage root and
// returns the live path they belong to. It never creates, replaces, or removes
// the live path itself: that is the exclusive job of the manager-home-locked
// commit phase, which journals every target before the first swap and restores
// them in exact reverse order when any class fails.
//
// Classes carry a numeric prefix because the transaction engine orders targets
// bytewise by (class, identifier). The prefix is therefore the commit order
// itself, not decoration, and the machine-wide consumer ledger deliberately
// sorts last.
package staging

import (
	"fmt"
	"path/filepath"
	"sort"
)

// Commit classes in normative order (Spec §6.1 step 12).
const (
	// ClassContext holds project, global, and hybrid context directories and
	// the install markers inside them.
	ClassContext = "10-context"
	// ClassRuntime holds commit-keyed script runtime trees.
	ClassRuntime = "20-runtime"
	// ClassCanonicalShim holds project and global canonical launchers.
	ClassCanonicalShim = "30-shim-canonical"
	// ClassForwardingShim holds safe user-bin forwarding launchers.
	ClassForwardingShim = "40-shim-forwarding"
	// ClassEnvFile holds the generated PATH helper files of a scope.
	ClassEnvFile = "50-env-file"
	// ClassAdapterLedger holds per-adapter-root ownership ledgers.
	ClassAdapterLedger = "60-adapter-ledger"
	// ClassMirrorLedger holds the user-bin mirror ownership ledger.
	ClassMirrorLedger = "70-mirror-ledger"
	// ClassRemoval holds managed live paths that the next state does not keep.
	ClassRemoval = "80-removal"
	// ClassConsumer holds the machine-wide consumer ledger and always commits
	// last, so a failed first install leaves no consumer record behind.
	ClassConsumer = "90-consumer"
)

// Kind selects how the commit phase owns one target path.
const (
	// KindBytes is the default: a regular file or a link-free directory tree.
	KindBytes = ""
	// KindEntry is a managed directory entry that the manager owns as itself and
	// which may be a symbolic link — an agent adapter mirror in the default
	// symlink-first mode is exactly this. Its final path component is never
	// resolved, and a link is journaled, restored, and removed as the link it is.
	KindEntry = "entry"
)

// Target is one journaled replacement. An empty StagedPath means the desired
// state is absence, which is how a stale managed path is removed.
type Target struct {
	Class      string
	Identifier string
	Kind       string
	LivePath   string
	StagedPath string
}

// Removal reports whether the desired state of the target is absence.
func (target Target) Removal() bool { return target.StagedPath == "" }

// Plan accumulates the complete target set of one commit.
type Plan struct {
	Targets []Target
}

// Replace records a desired replacement of one live path.
func (plan *Plan) Replace(class, identifier, livePath, stagedPath string) {
	plan.replace(class, identifier, KindBytes, livePath, stagedPath)
}

// ReplaceEntry records a desired replacement of one managed directory entry,
// which may be or become a symbolic link.
func (plan *Plan) ReplaceEntry(class, identifier, livePath, stagedPath string) {
	plan.replace(class, identifier, KindEntry, livePath, stagedPath)
}

func (plan *Plan) replace(class, identifier, kind, livePath, stagedPath string) {
	plan.Targets = append(plan.Targets, Target{
		Class: class, Identifier: identifier, Kind: kind, LivePath: livePath, StagedPath: stagedPath,
	})
}

// Remove records a managed live path whose desired state is absence. Every
// removal commits in the single removal class, after the classes that install
// the state which replaces it.
func (plan *Plan) Remove(identifier, livePath string) {
	plan.Targets = append(plan.Targets, Target{
		Class: ClassRemoval, Identifier: identifier, LivePath: livePath,
	})
}

// RemoveEntry records a managed directory entry whose desired state is absence.
// A stale adapter mirror is removed through here, so its exact prior state —
// including a symbolic link's destination — is restored by a rollback.
func (plan *Plan) RemoveEntry(identifier, livePath string) {
	plan.Targets = append(plan.Targets, Target{
		Class: ClassRemoval, Identifier: identifier, Kind: KindEntry, LivePath: livePath,
	})
}

// Merge appends every target of another plan.
func (plan *Plan) Merge(other Plan) {
	plan.Targets = append(plan.Targets, other.Targets...)
}

// Empty reports whether the commit would change nothing.
func (plan Plan) Empty() bool { return len(plan.Targets) == 0 }

// Sorted returns the targets in commit order: bytewise by class, then by
// identifier inside the class.
func (plan Plan) Sorted() []Target {
	sorted := append([]Target(nil), plan.Targets...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Class != sorted[j].Class {
			return sorted[i].Class < sorted[j].Class
		}
		return sorted[i].Identifier < sorted[j].Identifier
	})
	return sorted
}

// Validate rejects a plan the commit phase could not journal safely: an
// unusable path, a duplicate canonical identifier, or two targets that name the
// same live path. The transaction layer validates namespaces independently;
// this catches producer defects with a message that names the producer.
func (plan Plan) Validate() error {
	seenKey := map[string]bool{}
	seenLive := map[string]bool{}
	for _, target := range plan.Targets {
		if target.Class == "" || target.Identifier == "" {
			return fmt.Errorf("staged target has no class or identifier")
		}
		if target.Kind != KindBytes && target.Kind != KindEntry {
			return fmt.Errorf("staged target %s/%s has unknown kind %q", target.Class, target.Identifier, target.Kind)
		}
		if !filepath.IsAbs(target.LivePath) || filepath.Clean(target.LivePath) != target.LivePath {
			return fmt.Errorf("staged target %s/%s live path must be clean and absolute", target.Class, target.Identifier)
		}
		if target.StagedPath != "" && (!filepath.IsAbs(target.StagedPath) || filepath.Clean(target.StagedPath) != target.StagedPath) {
			return fmt.Errorf("staged target %s/%s staged path must be clean and absolute", target.Class, target.Identifier)
		}
		key := target.Class + "\x00" + target.Identifier
		if seenKey[key] {
			return fmt.Errorf("duplicate staged target %s/%s", target.Class, target.Identifier)
		}
		seenKey[key] = true
		if seenLive[target.LivePath] {
			return fmt.Errorf("two staged targets claim the live path %s", target.LivePath)
		}
		seenLive[target.LivePath] = true
	}
	return nil
}
