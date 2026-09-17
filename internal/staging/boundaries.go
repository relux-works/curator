package staging

// Draft source boundary checks for publication planning.
//
// This file implements the revision 1 physical-boundary half of
// protocol/skillfile-sources.md section 2 that belongs to the staging layer:
// canonicalize physical paths with symlink and actual filesystem
// case-equivalence rules, and recheck destination separation immediately
// before publication. Planning-time string prefixes alone are insufficient,
// so every comparison resolves the live filesystem and fails closed on
// inspection errors.
//
// Revision 2 (ports, mirrors, aliases) is a separate future leaf: these
// helpers carry no endpoint properties and need no v2 behaviour.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// Canonicalize resolves path to its physical form for boundary comparison:
// absolute, clean, with symlinks resolved. Missing trailing components are
// resolved against their closest existing prefix, so planned destinations
// that do not exist yet still compare against the filesystem that will
// hold them. Any inspection failure other than absence fails closed.
func Canonicalize(path string) (string, error) {
	if path == "" || !utf8.ValidString(path) || strings.ContainsRune(path, 0) || !filepath.IsAbs(path) {
		return "", fmt.Errorf("path is not valid absolute filesystem text")
	}
	absolute := filepath.Clean(path)
	resolved, err := filepath.EvalSymlinks(absolute)
	if err == nil {
		return filepath.Clean(resolved), nil
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("resolve symlinks for %q: %w", path, err)
	}
	prefix := absolute
	missing := make([]string, 0, 4)
	for {
		parent := filepath.Dir(prefix)
		if parent == prefix {
			return "", fmt.Errorf("resolve existing prefix for %q: %w", path, err)
		}
		missing = append(missing, filepath.Base(prefix))
		prefix = parent
		resolved, resolveErr := filepath.EvalSymlinks(prefix)
		if resolveErr == nil {
			for index := len(missing) - 1; index >= 0; index-- {
				resolved = filepath.Join(resolved, missing[index])
			}
			return filepath.Clean(resolved), nil
		}
		if !os.IsNotExist(resolveErr) {
			return "", fmt.Errorf("resolve symlinks for %q: %w", path, resolveErr)
		}
	}
}

// canonicalEntry resolves the parent of an owned directory entry while
// keeping its final component: the entry itself is the target, so resolving
// it would compare the manager's own link against whatever it points at.
// A newly introduced final-component link is therefore seen as a changed
// entry, never followed.
func canonicalEntry(path string) (string, error) {
	if path == "" || !utf8.ValidString(path) || strings.ContainsRune(path, 0) || !filepath.IsAbs(path) {
		return "", fmt.Errorf("path is not valid absolute filesystem text")
	}
	absolute := filepath.Clean(path)
	parent := filepath.Dir(absolute)
	if parent == absolute {
		return absolute, nil
	}
	resolvedParent, err := Canonicalize(parent)
	if err != nil {
		return "", err
	}
	return filepath.Join(resolvedParent, filepath.Base(absolute)), nil
}

// Within reports whether candidate lies within root using physical identity.
// It combines canonical string containment (symlinks resolved) with a
// SameFile ancestor walk, so the filesystem's actual case-equivalence
// rules count — including case-insensitive volumes where two spellings
// name one object. Inspection failures fail closed.
//
// The walk uses Lstat and never follows a link. Entry paths legitimately
// name links — a staged-then-committed mirror entry carries a destination
// string that did not resolve where os.Symlink created it, so on Windows
// the published entry is a file link pointing at a directory — and
// following one with Stat fails the whole check there (CreateFile Access
// denied) even when the destinations are disjoint. The walk keeps the
// caller's own spellings so an entry's location still counts: a link
// sitting inside admitted inputs is within them however it resolves,
// decided by its parent levels, while where it points is decided by the
// canonical comparison and the callers' explicit link-target arms.
// Case-variant spellings still name one object and compare equal, and a
// dangling link compares as the link itself rather than vanishing from
// the walk.
//
// A Unicode-normalization alias of a missing path is a stated blind spot:
// existing paths are covered through SameFile, but NFD-equivalent spellings
// of absent paths are not probed.
func Within(root, candidate string) (bool, error) {
	rootCanonical, err := Canonicalize(root)
	if err != nil {
		return false, err
	}
	candidateCanonical, err := Canonicalize(candidate)
	if err != nil {
		return false, err
	}
	if candidateCanonical == rootCanonical || strings.HasPrefix(candidateCanonical, rootCanonical+string(filepath.Separator)) {
		return true, nil
	}
	// SameFile covers aliases the string comparison cannot see: a
	// case-variant spelling on an insensitive volume, or a link the
	// missing-prefix walk could not fully resolve.
	rootInfo, err := os.Lstat(root) // #nosec G304 -- caller-supplied boundary root
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	for current := candidate; ; {
		if info, statErr := os.Lstat(current); statErr == nil && os.SameFile(rootInfo, info) { // #nosec G304 -- candidate ancestry under inspection
			return true, nil
		} else if statErr != nil && !os.IsNotExist(statErr) {
			return false, statErr
		}
		parent := filepath.Dir(current)
		if parent == current {
			return false, nil
		}
		current = parent
	}
}

// ContainsGit reports whether the canonical path carries a .git metadata
// component. Matching is case-insensitive: a case-variant .GIT directory on
// an insensitive volume is the same metadata.
func ContainsGit(canonical string) bool {
	for current := filepath.Clean(canonical); ; current = filepath.Dir(current) {
		if strings.EqualFold(filepath.Base(current), ".git") {
			return true
		}
		if filepath.Dir(current) == current {
			return false
		}
	}
}

// IsOutputPath reports whether candidate must be pruned as managed or
// generated output: inside any output root, or carrying .git metadata.
// Unknown output spellings and inspection failures fail closed.
func IsOutputPath(candidate string, outputs []string) (bool, error) {
	canonical, err := Canonicalize(candidate)
	if err != nil {
		return false, err
	}
	if ContainsGit(canonical) {
		return true, nil
	}
	for _, output := range outputs {
		within, err := Within(output, canonical)
		if err != nil {
			return false, err
		}
		if within {
			return true, nil
		}
	}
	return false, nil
}

// PinnedIdentity is one planning-time filesystem pin: the durable identity
// token (see FileIdentity) captured by a single inspection at Snapshot
// time. The token is the single authoritative pin for both the
// same-process Recheck and restart recovery: both verify the live token
// against it (see verifyDurableIdentity), so the two verification paths
// can never disagree about which object was planned. Durable serializes
// the stored token without touching the filesystem, so a swap between
// planning and conversion cannot substitute a later identity for the
// planned one.
type PinnedIdentity struct {
	Token string
}

// TargetSnapshot is the planning-time physical record of one destination:
// its resolved spelling plus the filesystem identity of every existing
// ancestor directory. The live path itself is deliberately not pinned —
// publication replaces it — but every directory holding it is.
type TargetSnapshot struct {
	Canonical string
	Ancestors map[string]PinnedIdentity
}

// Snapshot is the planning-time physical record Recheck compares against
// the live filesystem immediately before publication. Targets holds one
// record per planned destination; Admitted pins the identity of every
// admitted input, so a swapped or vanished input fails closed instead of
// comparing separation against whatever now sits at its spelling. The
// zero Snapshot records nothing: Recheck then skips the retarget and
// identity checks but still enforces separation.
type Snapshot struct {
	Targets  map[string]TargetSnapshot
	Admitted map[string]PinnedIdentity
}

// DurableTarget is the journal-stable form of TargetSnapshot: the resolved
// spelling plus the durable identity token (see FileIdentity) of every
// existing ancestor directory.
type DurableTarget struct {
	Canonical string            `json:"canonical"`
	Ancestors map[string]string `json:"ancestors,omitempty"`
}

// DurableSnapshot is the journal-stable form of Snapshot: the same canonical
// spellings with durable identity tokens instead of in-memory FileInfo, so
// crash recovery re-verifies the same boundary the original attempt pinned.
// A nil map means nothing recorded; an empty non-nil proof (no targets) is
// protected-but-unrestorable and fails closed, never legacy.
type DurableSnapshot struct {
	Targets  map[string]DurableTarget `json:"targets,omitempty"`
	Admitted map[string]string        `json:"admitted,omitempty"`
}

// Boundaries bundles the planning-time snapshot with the admitted inputs
// it pins: source package files, root-input selections, runtime, build and
// dependency inputs. The install commit phase carries one per scope and
// rechecks it immediately before journaling; a nil *Boundaries leaves the
// legacy path with byte-identical behavior. Durable is the journal-stable
// copy of Snapshot the transaction persists for restart recovery.
type Boundaries struct {
	Snapshot Snapshot
	Admitted []string
	Durable  DurableSnapshot
}

// Snapshot records the plan-time physical identity of every live path and
// every admitted input: full resolution for byte targets, parent
// resolution for owned entries, and the filesystem identity (device+inode
// on unix, volume serial plus file index on Windows) of every existing
// destination ancestor and every admitted input. Each pin comes from a
// single inspection (see captureIdentity): the durable token is the
// single authoritative pin for both the same-process Recheck and restart
// recovery, so the two verifications can never name different objects.
// Comparing spellings alone would admit a same-path replacement —
// renaming the planned parent away and creating a new directory at its
// spelling — so Recheck compares identities, not strings. Absent
// ancestors are not pinned: publication legitimately creates them. An
// admitted input must exist; a missing one is an error, never a blind
// pass. Recheck compares this record against the live filesystem
// immediately before publication; a changed boundary fails and rolls back
// unpublished state.
func (plan Plan) Snapshot(admitted []string) (Snapshot, error) {
	snapshot := Snapshot{
		Targets:  make(map[string]TargetSnapshot, len(plan.Targets)),
		Admitted: make(map[string]PinnedIdentity, len(admitted)),
	}
	for _, input := range admitted {
		canonical, err := Canonicalize(input)
		if err != nil {
			return Snapshot{}, fmt.Errorf("boundary_identity_unreadable: admitted input %s: %v", input, err)
		}
		pin, err := captureIdentity(canonical)
		if err != nil {
			if os.IsNotExist(err) {
				return Snapshot{}, fmt.Errorf("source_member_missing: admitted input %s: %v", input, err)
			}
			return Snapshot{}, fmt.Errorf("boundary_identity_unreadable: admitted input %s: %v", input, err)
		}
		snapshot.Admitted[canonical] = pin
	}
	for _, target := range plan.Targets {
		var (
			canonical string
			err       error
		)
		if target.Kind == KindEntry {
			canonical, err = canonicalEntry(target.LivePath)
		} else {
			canonical, err = Canonicalize(target.LivePath)
		}
		if err != nil {
			return Snapshot{}, fmt.Errorf("boundary_identity_unreadable: snapshot %s: %v", target.LivePath, err)
		}
		ancestors, err := snapshotAncestors(canonical)
		if err != nil {
			return Snapshot{}, fmt.Errorf("boundary_identity_unreadable: snapshot %s: %v", target.LivePath, err)
		}
		snapshot.Targets[target.LivePath] = TargetSnapshot{Canonical: canonical, Ancestors: ancestors}
	}
	return snapshot, nil
}

// captureHook is the test-only identity-capture seam. captureIdentity fires
// it after its single filesystem inspection and before the pure token
// derivation: the exact point where a second pathname-based read would
// observe a swapped object. Production leaves it nil. Tests install it
// with SetCaptureHook and clear it before returning; it is not safe for
// concurrent use.
var captureHook func(canonical string)

// SetCaptureHook installs the test-only identity-capture hook fired by
// captureIdentity between its inspection and its token derivation, or
// clears it with nil. Deterministic capture-window tests swap the object
// under capture from this hook and assert the pin still names the
// pre-swap object; a capture that re-reads the path after the hook fails
// those tests. Production code never sets this.
func SetCaptureHook(hook func(canonical string)) {
	captureHook = hook
}

// captureIdentity pins one canonical path with a single filesystem
// inspection: the durable token derives from that one inspection and
// nothing else, so a swap landing between two reads cannot split the pin
// from its token. On unix the inspection is one Stat whose buffer feeds
// the token; on Windows it is one opened handle whose file index feeds
// the token — never a second pathname-based open, because FileInfo
// carries no public file index there and os.SameFile would resolve it
// lazily from the path. An uncapturable identity fails closed at
// planning.
func captureIdentity(canonical string) (PinnedIdentity, error) {
	inspected, err := inspectPath(canonical)
	if err != nil {
		return PinnedIdentity{}, err
	}
	if captureHook != nil {
		captureHook(canonical)
	}
	token, err := tokenOfInspection(inspected, canonical)
	if err != nil {
		return PinnedIdentity{}, err
	}
	return PinnedIdentity{Token: token}, nil
}

// snapshotAncestors pins the identity of every existing ancestor
// directory of a resolved destination spelling, nearest first. Absent
// ancestors carry no record; inspection failures other than absence fail
// closed. Each pin is a single-inspection capture, so Durable serializes
// planning-time identity without re-inspection.
func snapshotAncestors(canonical string) (map[string]PinnedIdentity, error) {
	ancestors := map[string]PinnedIdentity{}
	for current := filepath.Dir(canonical); ; {
		pin, err := captureIdentity(current)
		if err == nil {
			ancestors[current] = pin
		} else if !os.IsNotExist(err) {
			return nil, err
		}
		parent := filepath.Dir(current)
		if parent == current {
			return ancestors, nil
		}
		current = parent
	}
}

// Recheck validates destination separation immediately before publication
// writes, under the caller's serialized transaction. For every target it:
//   - re-resolves the live path without following a newly introduced final
//     link, and fails when the resolved spelling moved since Snapshot;
//   - re-stats every recorded destination ancestor and admitted input, and
//     fails when an identity changed or a recorded path vanished — a
//     same-spelling parent replacement fails here even though the spelling
//     comparison alone would pass;
//   - refuses byte targets whose live final component is now a link;
//   - refuses destinations that would overwrite an admitted input, in either
//     direction, including through links, casing aliases or a changed parent.
//
// Admitted holds canonical or absolute admitted input paths. Every entry
// must have been recorded by Snapshot; an unrecorded entry fails closed
// rather than skipping its identity check. Targets added after Snapshot
// (the commit phase merges the consumer ledger after scope planning) have
// no record: they are still separation-checked, and their bytes stay
// covered by engine preimages. Outputs themselves are where destinations
// belong and are not checked here; pass the zero Snapshot only when the
// caller recorded none, in which case the retarget and identity checks
// are skipped but separation is still enforced.
func (plan Plan) Recheck(snapshot Snapshot, admitted []string) error {
	admittedCanonical, err := recheckAdmitted(snapshot, admitted)
	if err != nil {
		return err
	}
	for _, target := range plan.Targets {
		if err := recheckTarget(target, snapshot, admittedCanonical); err != nil {
			return err
		}
	}
	return nil
}

// RecheckOne is the single-target form of Recheck for the per-write
// publication guard: the transaction engine calls it (through the
// install layer's BoundaryCheck) immediately before each publication
// write. It re-verifies the admitted identities and then the one target
// the engine is about to write, with exactly the refusal classes of
// Recheck. Callers recheck every written target; skipping one would leave
// its write unguarded.
func RecheckOne(target Target, snapshot Snapshot, admitted []string) error {
	admittedCanonical, err := recheckAdmitted(snapshot, admitted)
	if err != nil {
		return err
	}
	return recheckTarget(target, snapshot, admittedCanonical)
}

// recheckAdmitted re-resolves every admitted input and compares it against
// the planning-time snapshot: every entry must have been recorded, and no
// recorded identity may have changed or vanished. It returns the admitted
// canonical paths the target checks compare separation against. Proven
// changes refuse as source_output_overlap; inspection failures refuse as
// boundary_identity_unreadable, never as a proven overlap.
func recheckAdmitted(snapshot Snapshot, admitted []string) ([]string, error) {
	type admittedInput struct {
		original  string
		canonical string
	}
	inputs := make([]admittedInput, 0, len(admitted))
	admittedCanonical := make([]string, 0, len(admitted))
	for _, input := range admitted {
		canonical, err := Canonicalize(input)
		if err != nil {
			return nil, fmt.Errorf("boundary_identity_unreadable: admitted input %s: %v", input, err)
		}
		inputs = append(inputs, admittedInput{original: input, canonical: canonical})
		admittedCanonical = append(admittedCanonical, canonical)
	}
	recorded := len(snapshot.Targets) > 0 || len(snapshot.Admitted) > 0
	if recorded {
		for _, input := range inputs {
			if _, pinned := snapshot.Admitted[input.canonical]; !pinned {
				return nil, fmt.Errorf("source_output_overlap: admitted input %s was not recorded at planning", input.original)
			}
		}
		for canonical, pinned := range snapshot.Admitted {
			unreadable, err := verifyDurableIdentity("admitted input "+canonical, canonical, pinned.Token)
			if err != nil {
				if unreadable {
					return nil, fmt.Errorf("boundary_identity_unreadable: %v", err)
				}
				return nil, fmt.Errorf("source_output_overlap: %v", err)
			}
		}
	}
	return admittedCanonical, nil
}

// recheckTarget validates one planned destination immediately before
// publication writes, under the caller's serialized transaction. See
// Recheck for the checks; unrecorded targets (planned after Snapshot)
// are separation-checked without the retarget and identity comparisons.
func recheckTarget(target Target, snapshot Snapshot, admittedCanonical []string) error {
	var current string
	if target.Kind == KindEntry {
		resolved, err := canonicalEntry(target.LivePath)
		if err != nil {
			return fmt.Errorf("boundary_identity_unreadable: %s: %v", target.LivePath, err)
		}
		current = resolved
		if info, err := os.Lstat(target.LivePath); err == nil && info.Mode()&os.ModeSymlink != 0 { // #nosec G304 -- planned live entry under recheck
			destination, err := os.Readlink(target.LivePath) // #nosec G304 -- planned live entry under recheck
			if err != nil {
				return fmt.Errorf("boundary_identity_unreadable: %s: %v", target.LivePath, err)
			}
			linkTarget := destination
			if !filepath.IsAbs(linkTarget) {
				linkTarget = filepath.Join(filepath.Dir(target.LivePath), linkTarget)
			}
			if err := checkAdmittedOverlap(target.LivePath, linkTarget, admittedCanonical); err != nil {
				return err
			}
		} else if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("boundary_identity_unreadable: %s: %v", target.LivePath, err)
		}
	} else {
		resolved, err := Canonicalize(target.LivePath)
		if err != nil {
			return fmt.Errorf("boundary_identity_unreadable: %s: %v", target.LivePath, err)
		}
		current = resolved
		if info, err := os.Lstat(target.LivePath); err == nil && info.Mode()&os.ModeSymlink != 0 { // #nosec G304 -- planned live path under recheck
			return fmt.Errorf("source_output_overlap: %s is now a link", target.LivePath)
		} else if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("boundary_identity_unreadable: %s: %v", target.LivePath, err)
		}
	}
	if record, pinned := snapshot.Targets[target.LivePath]; pinned {
		if record.Canonical != current {
			return fmt.Errorf("source_output_overlap: %s retargeted since planning", target.LivePath)
		}
		for ancestor, pin := range record.Ancestors {
			unreadable, err := verifyDurableIdentity("destination ancestor "+ancestor, ancestor, pin.Token)
			if err != nil {
				if unreadable {
					return fmt.Errorf("boundary_identity_unreadable: %s: %v", target.LivePath, err)
				}
				return fmt.Errorf("source_output_overlap: %s: %v", target.LivePath, err)
			}
		}
	}
	if err := checkAdmittedOverlap(target.LivePath, current, admittedCanonical); err != nil {
		return err
	}
	return nil
}

func checkAdmittedOverlap(live, current string, admitted []string) error {
	for _, input := range admitted {
		within, err := Within(input, current)
		if err != nil {
			return fmt.Errorf("boundary_identity_unreadable: %s: %v", live, err)
		}
		if within {
			return fmt.Errorf("source_output_overlap: %s overwrites admitted input", live)
		}
		reverse, err := Within(current, input)
		if err != nil {
			return fmt.Errorf("boundary_identity_unreadable: %s: %v", live, err)
		}
		if reverse {
			return fmt.Errorf("source_output_overlap: %s overwrites admitted input", live)
		}
	}
	return nil
}

// Durable converts an in-memory Snapshot to its journal-stable form: the
// same canonical spellings with the durable identity tokens captured at
// Snapshot time. It never touches the filesystem: re-statting here would
// record whatever occupies each spelling now instead of the planned
// identity, so a swap between planning and conversion would persist the
// replacement. A pin without a stored token (never produced by Snapshot,
// which fails closed at planning when a token is uncapturable) fails
// closed as boundary_identity_unreadable, never as a proven overlap.
func (snapshot Snapshot) Durable() (DurableSnapshot, error) {
	durable := DurableSnapshot{
		Targets:  make(map[string]DurableTarget, len(snapshot.Targets)),
		Admitted: make(map[string]string, len(snapshot.Admitted)),
	}
	for canonical, pin := range snapshot.Admitted {
		if pin.Token == "" {
			return DurableSnapshot{}, fmt.Errorf("boundary_identity_unreadable: admitted input %s: identity was not captured at planning", canonical)
		}
		durable.Admitted[canonical] = pin.Token
	}
	for live, record := range snapshot.Targets {
		ancestors := make(map[string]string, len(record.Ancestors))
		for ancestor, pin := range record.Ancestors {
			if pin.Token == "" {
				return DurableSnapshot{}, fmt.Errorf("boundary_identity_unreadable: snapshot %s: destination ancestor %s identity was not captured at planning", live, ancestor)
			}
			ancestors[ancestor] = pin.Token
		}
		durable.Targets[live] = DurableTarget{Canonical: record.Canonical, Ancestors: ancestors}
	}
	return durable, nil
}

// RecheckDurableOne is the restart-recovery form of RecheckOne: it
// re-verifies one journal target against the durable proof the transaction
// persisted, immediately before the remaining write. Admitted identities
// come from the proof itself (there is no caller list to cross-check after
// a restart); unrecorded targets (planned after Snapshot, such as the
// consumer ledger merged after scope planning) are separation-checked
// without retarget and identity comparisons, exactly as in Recheck.
// Proven changes refuse as source_output_overlap; inspection failures
// refuse as boundary_identity_unreadable, never as a proven overlap.
func RecheckDurableOne(target Target, durable DurableSnapshot) error {
	admittedCanonical := make([]string, 0, len(durable.Admitted))
	for canonical, token := range durable.Admitted {
		unreadable, err := verifyDurableIdentity("admitted input "+canonical, canonical, token)
		if err != nil {
			if unreadable {
				return fmt.Errorf("boundary_identity_unreadable: %v", err)
			}
			return fmt.Errorf("source_output_overlap: %v", err)
		}
		admittedCanonical = append(admittedCanonical, canonical)
	}
	var current string
	if target.Kind == KindEntry {
		resolved, err := canonicalEntry(target.LivePath)
		if err != nil {
			return fmt.Errorf("boundary_identity_unreadable: %s: %v", target.LivePath, err)
		}
		current = resolved
		if info, err := os.Lstat(target.LivePath); err == nil && info.Mode()&os.ModeSymlink != 0 { // #nosec G304 -- planned live entry under durable recheck
			destination, err := os.Readlink(target.LivePath) // #nosec G304 -- planned live entry under durable recheck
			if err != nil {
				return fmt.Errorf("boundary_identity_unreadable: %s: %v", target.LivePath, err)
			}
			linkTarget := destination
			if !filepath.IsAbs(linkTarget) {
				linkTarget = filepath.Join(filepath.Dir(target.LivePath), linkTarget)
			}
			if err := checkAdmittedOverlap(target.LivePath, linkTarget, admittedCanonical); err != nil {
				return err
			}
		} else if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("boundary_identity_unreadable: %s: %v", target.LivePath, err)
		}
	} else {
		resolved, err := Canonicalize(target.LivePath)
		if err != nil {
			return fmt.Errorf("boundary_identity_unreadable: %s: %v", target.LivePath, err)
		}
		current = resolved
		if info, err := os.Lstat(target.LivePath); err == nil && info.Mode()&os.ModeSymlink != 0 { // #nosec G304 -- planned live path under durable recheck
			return fmt.Errorf("source_output_overlap: %s is now a link", target.LivePath)
		} else if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("boundary_identity_unreadable: %s: %v", target.LivePath, err)
		}
	}
	if record, pinned := durable.Targets[target.LivePath]; pinned {
		if record.Canonical != current {
			return fmt.Errorf("source_output_overlap: %s retargeted since planning", target.LivePath)
		}
		for ancestor, token := range record.Ancestors {
			unreadable, err := verifyDurableIdentity("destination ancestor "+ancestor, ancestor, token)
			if err != nil {
				if unreadable {
					return fmt.Errorf("boundary_identity_unreadable: %s: %v", target.LivePath, err)
				}
				return fmt.Errorf("source_output_overlap: %s: %v", target.LivePath, err)
			}
		}
	}
	if err := checkAdmittedOverlap(target.LivePath, current, admittedCanonical); err != nil {
		return err
	}
	return nil
}

// verifyDurableIdentity compares the live identity token at a recorded
// spelling against the planning-time token: a mismatch means the directory
// was replaced under its name, and a vanished path means a recorded
// ancestor or admitted input is gone. It is the single identity
// comparison behind both the same-process Recheck and restart recovery,
// so both paths verify the same authoritative pin. It returns
// unreadable=true for inspection failures (boundary_identity_unreadable)
// and unreadable=false for proven changes (vanished or changed,
// source_output_overlap) or nil.
func verifyDurableIdentity(label, canonical, recorded string) (bool, error) {
	current, err := FileIdentity(canonical)
	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Errorf("%s no longer exists", label)
		}
		return true, fmt.Errorf("%s: %v", label, err)
	}
	if current != recorded {
		return false, fmt.Errorf("%s changed since planning", label)
	}
	return false, nil
}
