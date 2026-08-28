package buildcache

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/closureexec"
)

// HomeLock is a caller-owned witness for the exclusive manager-home lock.
// Implementations must fail AssertHeld after release. Buildcache never creates
// a lock, which keeps dry-run inspection strictly read-only.
type HomeLock interface {
	AssertHeld() error
}

// Publication contains one fully built private artifact and its exact receipt.
// ArtifactSource is copied from its open file handle and is never executed.
type Publication struct {
	Input            buildmeta.Input
	ReceiptBytes     []byte
	Assurance        closureexec.AssuranceBinding
	ExecutionReceipt closureexec.BuildSessionReceipt
	ArtifactSource   string
}

// PublicationStatus distinguishes a newly selected directory winner from an
// identical winner selected by another publisher.
type PublicationStatus string

// Publication outcomes: a newly protected winner, or an identical existing one.
const (
	Published    PublicationStatus = "published"
	ReusedWinner PublicationStatus = "reused-winner"
)

// PublicationResult identifies the protected immutable winner.
type PublicationResult struct {
	Status       PublicationStatus
	ArtifactPath string
	ReceiptHash  buildmeta.ReceiptHash
	CacheKey     buildmeta.CacheKey
	// Quarantined is where an unusable predecessor was moved aside, and is
	// empty when this publication displaced nothing. It is the other half of
	// Revert: a caller that has to undo a publication it made cannot restore
	// what it displaced unless the publication says what that was.
	Quarantined string
}

// ConflictError reports different bytes for the same logical cache key.
// The existing protected winner is left unchanged.
type ConflictError struct {
	Key buildmeta.CacheKey
}

func (err *ConflictError) Error() string {
	return fmt.Sprintf("cache publication conflict for %s", err.Key)
}

// StateChangedError reports a failed cache operation that could not put the
// protected cache root back the way it found it.
//
// It is the one honest answer to a half-applied mutation. Publication and
// reversal both move live directories, and a fault between two of those moves
// leaves a shared slot holding something neither the caller nor the previous
// installation asked for. Presenting that as an ordinary failure would let a
// caller repeat the "the live build cache is unchanged" claim while it is not,
// so the boundary is marked in the error instead.
type StateChangedError struct {
	Key buildmeta.CacheKey
	Err error
}

func (err *StateChangedError) Error() string {
	return fmt.Sprintf("protected build cache state for %s was left changed: %v", err.Key, err.Err)
}

func (err *StateChangedError) Unwrap() error { return err.Err }

// StateChanged reports whether a failure left the protected cache root changed.
//
// A Publish error that does not satisfy it is a guarantee, not an absence of
// information: publication compensates its own mutations before returning, so
// the caller owes nothing for that call. Reversal is the mirror image — every
// error it returns means the entry this run published is still what the cache
// holds — so callers treat any reversal failure as a change regardless.
func StateChanged(err error) bool {
	var changed *StateChangedError
	return errors.As(err, &changed)
}

// faultPoint names one mutation boundary inside a publication or a reversal.
//
// These are exactly the boundaries that can leave the protected cache root
// half-changed. Several of them cannot be provoked from outside the package at
// all — a post-selection validation failure needs the live entry to break
// between the rename and the read, and directory syncing is a no-op on Windows
// — so the compensation they guard would otherwise be untestable on the
// production path.
type faultPoint string

// The mutation boundaries a deterministic in-package fault can be injected at.
//
// faultQuarantineSync is the interior one: it stands between the rename that
// empties the live slot and the sync that makes it durable, which is the only
// window where a quarantine has mutated the cache and has not yet returned.
// faultQuarantineRollbackSync stands one step further in, at the durability of
// the compensating rename that answers the first failure — the repair is a
// mutation too, and it is only worth what its own sync is worth.
const (
	faultQuarantine             faultPoint = "quarantine"
	faultQuarantineSync         faultPoint = "quarantine-sync"
	faultQuarantineRollbackSync faultPoint = "quarantine-rollback-sync"
	faultSelect                 faultPoint = "select"
	faultValidate               faultPoint = "validate"
	faultSync                   faultPoint = "sync"
	faultWithdraw               faultPoint = "withdraw"
	faultRestore                faultPoint = "restore"
	faultRestoreSync            faultPoint = "restore-sync"
)

// unsyncedRollbackError reports a compensating rename that put an entry back at
// its live pathname but could not make that repair durable.
//
// It is the one quarantine outcome where the visible namespace is right and the
// durable state is not, so neither ordinary answer is honest on its own: the
// bytes are live and recoverable now, and a crash can still leave the slot
// empty. It is deliberately not a *StateChangedError, because the helper that
// raises it names entries by path while every caller names them by cache key.
// The callers re-key it through stateChangeFor so the reported boundary carries
// the entry it belongs to.
type unsyncedRollbackError struct{ Err error }

func (err *unsyncedRollbackError) Error() string { return err.Err.Error() }

func (err *unsyncedRollbackError) Unwrap() error { return err.Err }

// stateChangeFor marks a quarantine failure that left durable state uncertain as
// the changed state it is, and passes every other failure through untouched.
//
// A quarantine that put the entry back durably compensated its own mutation and
// owes the caller nothing, which is the guarantee an untyped error carries.
func stateChangeFor(key buildmeta.CacheKey, err error) error {
	var unsynced *unsyncedRollbackError
	if err == nil || !errors.As(err, &unsynced) {
		return err
	}
	return &StateChangedError{Key: key, Err: err}
}

// quarantinePrefix names an entry this store moved aside instead of deleting.
// The ordinary sweep collects it; nothing else in the cache root carries it.
const quarantinePrefix = ".quarantine-"

// fault consults the in-package test seam. New never sets it, so no store a
// caller can construct outside this package ever takes the hooked path.
func (store *Store) fault(point faultPoint) error {
	if store == nil || store.faults == nil {
		return nil
	}
	return store.faults(point)
}

// Publish verifies a private build, creates a protected staging directory, and
// selects one complete directory winner atomically. Existing corrupt or
// untrusted entries are quarantined rather than adopted or permission-repaired.
//
// Publication is fail-closed across every live mutation it makes. Selecting a
// winner is not one atomic step: an unusable predecessor is quarantined first,
// the staged directory is renamed into the freed slot second, and the winner is
// validated and the cache root synced third. A fault at any of those later steps
// used to return an empty result with an error while the shared cache already
// held something else. It now compensates before it returns, so an error from
// this call means the cache root is exactly what this call found and the caller
// owes it nothing. Only a compensation that itself fails returns a
// *StateChangedError, which says the opposite in the one case where it is true.
func (store *Store) Publish(publication Publication, lock HomeLock) (result PublicationResult, err error) {
	if err := requireHomeLock(lock); err != nil {
		return PublicationResult{}, err
	}
	if store == nil || store.supported == nil || !store.supported() {
		return PublicationResult{}, fmt.Errorf("persistent build cache protection is unsupported")
	}
	receipt, err := buildmeta.DecodeExpectedReceipt(publication.ReceiptBytes, publication.Input)
	if err != nil {
		return PublicationResult{}, fmt.Errorf("validate publication receipt: %w", err)
	}
	logicalKey, err := publication.Input.CacheKey()
	if err != nil {
		return PublicationResult{}, fmt.Errorf("derive publication key: %w", err)
	}
	if receipt.CacheKey != logicalKey {
		return PublicationResult{}, fmt.Errorf("publication receipt key mismatch")
	}
	if err := publication.ExecutionReceipt.ValidateFor(publication.Assurance, publication.Input, receipt.Artifact); err != nil {
		return PublicationResult{}, fmt.Errorf("validate publication execution receipt: %w", err)
	}
	executionReceiptBytes, err := publication.ExecutionReceipt.CanonicalBytes()
	if err != nil {
		return PublicationResult{}, fmt.Errorf("encode publication execution receipt: %w", err)
	}
	assuredID, err := (closureexec.AssuredBuildCacheInput{BuildInput: publication.Input, Binding: publication.Assurance}).ID()
	if err != nil {
		return PublicationResult{}, fmt.Errorf("derive assured publication key: %w", err)
	}
	key := buildmeta.CacheKey(assuredID)
	receiptHash, err := buildmeta.HashReceiptBytes(publication.ReceiptBytes)
	if err != nil {
		return PublicationResult{}, fmt.Errorf("hash publication receipt: %w", err)
	}

	source, err := openRegularSource(publication.ArtifactSource)
	if err != nil {
		return PublicationResult{}, err
	}
	defer func() { _ = source.Close() }()
	artifactHash, artifactSize, err := hashOpenFile(source)
	if err != nil {
		return PublicationResult{}, fmt.Errorf("hash publication artifact: %w", err)
	}
	if artifactHash != receipt.Artifact.SHA256 || artifactSize != receipt.Artifact.Size {
		return PublicationResult{}, fmt.Errorf("publication artifact does not match receipt")
	}
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return PublicationResult{}, fmt.Errorf("rewind publication artifact: %w", err)
	}

	entryPath, base, err := store.paths(key)
	if err != nil {
		return PublicationResult{}, err
	}

	// displaced and selected are the only live mutations the selection below can
	// make: the predecessor it moved aside, and whether the staged directory
	// became the entry a launcher resolves. The compensation is armed here,
	// before either can happen, so every later exit — including the ones inside
	// the loop and the one that runs out of attempts — unwinds exactly what was
	// done. A call that changed neither returns its error untouched.
	displaced := ""
	selected := false
	defer func() {
		if err == nil || (displaced == "" && !selected) {
			return
		}
		result = PublicationResult{}
		if restoreErr := store.restoreDisplaced(entryPath, base, displaced, lock); restoreErr != nil {
			err = &StateChangedError{Key: key, Err: errors.Join(err, restoreErr)}
		}
	}()

	if err := ensureProtectedBase(store.home, base); err != nil {
		return PublicationResult{}, fmt.Errorf("prepare protected cache root: %w", err)
	}
	stage, err := makeProtectedTempDir(base, ".stage-"+strings.TrimPrefix(string(key), "sha256:")+"-")
	if err != nil {
		return PublicationResult{}, fmt.Errorf("create publication staging: %w", err)
	}
	keepStage := false
	defer func() {
		if !keepStage {
			_ = os.RemoveAll(stage)
		}
	}()

	artifactRel := filepath.FromSlash(receipt.Artifact.Path)
	binDir := filepath.Join(stage, "bin")
	if err := createProtectedDir(binDir); err != nil {
		return PublicationResult{}, fmt.Errorf("create staged artifact directory: %w", err)
	}
	stagedArtifact := filepath.Join(stage, artifactRel)
	if err := writeProtectedFile(stagedArtifact, 0o700, source); err != nil {
		return PublicationResult{}, fmt.Errorf("write staged artifact: %w", err)
	}
	if err := writeProtectedFile(filepath.Join(stage, ReceiptFilename), 0o600, bytes.NewReader(publication.ReceiptBytes)); err != nil {
		return PublicationResult{}, fmt.Errorf("write staged receipt: %w", err)
	}
	if err := writeProtectedFile(filepath.Join(stage, ExecutionReceiptFilename), 0o600, bytes.NewReader(executionReceiptBytes)); err != nil {
		return PublicationResult{}, fmt.Errorf("write staged execution receipt: %w", err)
	}
	if err := syncDirectory(binDir); err != nil {
		return PublicationResult{}, fmt.Errorf("sync staged artifact directory: %w", err)
	}
	if err := syncDirectory(stage); err != nil {
		return PublicationResult{}, fmt.Errorf("sync staged cache entry: %w", err)
	}

	expect := Expectation{Input: publication.Input, ReceiptHash: receiptHash, Assurance: publication.Assurance}
	winnerExpect := Expectation{Input: publication.Input, Assurance: publication.Assurance}
	staged := store.inspectEntry(stage, expect, key, logicalKey)
	if staged.Status != Hit {
		return PublicationResult{}, fmt.Errorf("staged cache entry failed protected validation: %s", staged.Reason)
	}

	// displaced remembers the unusable predecessor this publication moved aside,
	// so it can be put back — by the compensation above if this call fails, or
	// by the caller's Revert if the operation that needed the replacement never
	// becomes durable. Only the first one is kept: a second pass can only find
	// the slot empty, because nothing but this publisher holds the lock.
	for attempts := 0; attempts < 3; attempts++ {
		if err := requireHomeLock(lock); err != nil {
			return PublicationResult{}, err
		}
		winner := store.Inspect(winnerExpect)
		switch winner.Status {
		case Hit:
			if bytes.Equal(winner.ReceiptBytes, publication.ReceiptBytes) &&
				winner.Receipt.Artifact.SHA256 == artifactHash && winner.Receipt.Artifact.Size == artifactSize {
				return PublicationResult{
					Status: ReusedWinner, ArtifactPath: winner.ArtifactPath,
					ReceiptHash: winner.ReceiptHash, CacheKey: key, Quarantined: displaced,
				}, nil
			}
			return PublicationResult{}, &ConflictError{Key: key}
		case Unsupported:
			return PublicationResult{}, fmt.Errorf("persistent build cache protection is unsupported")
		case Corrupt, UntrustedProvenance:
			// The moved path is recorded before the error is acted on, and not
			// only on success. A withdrawal that empties the slot and then fails
			// is still a mutation this publication made, and the compensation
			// above can only unwind what displaced names.
			moved, quarantineErr := store.withdrawEntry(entryPath, faultQuarantine, lock)
			if displaced == "" {
				displaced = moved
			}
			if quarantineErr != nil {
				return PublicationResult{}, stateChangeFor(key, quarantineErr)
			}
		case Miss:
		}

		// The fault stands in for the rename itself rather than short-circuiting
		// the loop, so a selection that never succeeds reaches the exhaustion exit
		// below with the predecessor already quarantined — which is one of the
		// paths that has to put the cache back.
		selectErr := store.fault(faultSelect)
		if selectErr == nil {
			selectErr = renameDirectoryNoReplace(stage, entryPath)
		}
		if selectErr == nil {
			keepStage = true
			selected = true
			if err := store.fault(faultValidate); err != nil {
				return PublicationResult{}, fmt.Errorf("published cache winner failed validation: %v", err)
			}
			winner = store.Inspect(expect)
			if winner.Status != Hit {
				return PublicationResult{}, fmt.Errorf("published cache winner failed validation: %s", winner.Reason)
			}
			if err := store.fault(faultSync); err != nil {
				return PublicationResult{}, fmt.Errorf("sync cache root: %w", err)
			}
			if err := syncDirectory(base); err != nil {
				return PublicationResult{}, fmt.Errorf("sync cache root: %w", err)
			}
			return PublicationResult{
				Status: Published, ArtifactPath: winner.ArtifactPath,
				ReceiptHash: winner.ReceiptHash, CacheKey: key, Quarantined: displaced,
			}, nil
		}
		// A racing publisher may have selected a winner. Loop to validate it;
		// no directory is merged and this staging directory remains private.
	}
	return PublicationResult{}, fmt.Errorf("cache publication could not select or validate a winner")
}

// Revert undoes one publication this run performed and puts back whatever was
// live before it.
//
// It exists because publication is not a transaction target. A launcher can
// only point at a protected entry that is already live, so the replacement is
// selected before the installation that needs it is durable; an operation that
// then fails owns putting the shared cache back, or a run that committed
// nothing would still have left the predecessor quarantined and its
// replacement live.
//
// Every step is a rename inside the protected cache root, under the same
// caller-held home lock the publication ran under. No byte is deleted here:
// the withdrawn entry is quarantined exactly like any other unusable one and
// is collected by the ordinary sweep. A publication that selected no new
// winner changed nothing this call can undo and is a no-op.
//
// Every failure it reports is a *StateChangedError, and deliberately so: a
// reversal that refuses before it moves anything, and one that stops between
// its own two renames, both leave the cache holding something other than what
// the run found. The caller has one correct response to either, which is to
// stop claiming the live cache is unchanged.
func (store *Store) Revert(key buildmeta.CacheKey, published PublicationResult, lock HomeLock) error {
	if _, _, err := store.paths(key); err != nil {
		return &StateChangedError{Key: key, Err: err}
	}
	if published.CacheKey != "" {
		key = published.CacheKey
	}
	if err := requireHomeLock(lock); err != nil {
		return &StateChangedError{Key: key, Err: err}
	}
	if store == nil || store.supported == nil || !store.supported() {
		return &StateChangedError{Key: key, Err: fmt.Errorf("persistent build cache protection is unsupported")}
	}
	if published.Status != Published {
		return nil
	}
	entryPath, base, err := store.paths(key)
	if err != nil {
		return &StateChangedError{Key: key, Err: err}
	}
	// The predecessor is only ever restored from inside the cache root this
	// store owns, so a caller cannot use a reversal to move foreign bytes in.
	// It is checked before anything moves: a reversal that cannot complete must
	// not have withdrawn the live entry on its way to refusing.
	if published.Quarantined != "" && !pathWithin(base, published.Quarantined) {
		return &StateChangedError{
			Key: key,
			Err: fmt.Errorf("quarantined cache entry is outside the protected cache root"),
		}
	}
	if err := store.restoreDisplaced(entryPath, base, published.Quarantined, lock); err != nil {
		return &StateChangedError{Key: key, Err: err}
	}
	return nil
}

// restoreDisplaced puts one protected slot back the way a publication found it:
// it withdraws whatever that publication selected and renames the predecessor
// it displaced back into place.
//
// It is the single implementation behind both the compensation Publish runs
// before returning an error and the caller-driven Revert, so a failure in one
// path cannot behave differently from the same failure in the other.
//
// It is fail-closed at its own seam too. Withdrawal and restoration are two
// renames, and a fault between them would leave the slot empty while a
// perfectly usable entry sat in quarantine — which is strictly worse than
// either end state, because a launcher that already points into that slot
// resolves to nothing. So a restoration that fails puts the withdrawn entry
// back before it reports. No byte is deleted on any path: everything moved
// aside is quarantined exactly like any other unusable entry and the ordinary
// sweep collects it.
func (store *Store) restoreDisplaced(entryPath, base, displaced string, lock HomeLock) error {
	// A withdrawal that fails after it has already emptied the slot is the worst
	// of the three outcomes — the predecessor is still in quarantine and the
	// launcher now resolves to nothing — so the entry goes back before the
	// failure is reported, exactly as it does when the restoration below fails.
	withdrawn, err := store.withdrawEntry(entryPath, faultWithdraw, lock)
	if err != nil {
		return fmt.Errorf("withdraw the published cache entry: %w",
			errors.Join(err, returnWithdrawn(withdrawn, entryPath)))
	}
	if displaced == "" {
		return nil
	}
	restoreErr := store.fault(faultRestore)
	if restoreErr == nil {
		restoreErr = renameDirectoryNoReplace(displaced, entryPath)
	}
	if restoreErr != nil {
		return fmt.Errorf("restore the quarantined cache entry: %w",
			errors.Join(restoreErr, returnWithdrawn(withdrawn, entryPath)))
	}
	if err := store.fault(faultRestoreSync); err != nil {
		return fmt.Errorf("sync cache root: %w", err)
	}
	if err := syncDirectory(base); err != nil {
		return fmt.Errorf("sync cache root: %w", err)
	}
	return nil
}

// returnWithdrawn moves the entry a failed restoration had already withdrawn
// back into the live slot. It reports its own failure rather than hiding it:
// the caller is already returning an error, and which of the two entries ended
// up live is the difference between a cache miss and an empty slot.
func returnWithdrawn(withdrawn, entryPath string) error {
	if withdrawn == "" {
		return nil
	}
	if err := renameDirectoryNoReplace(withdrawn, entryPath); err != nil {
		return fmt.Errorf("return the withdrawn cache entry to the live slot: %w", err)
	}
	return nil
}

// Quarantine atomically moves a present entry aside under the caller-held home
// lock. It never traverses or repairs the entry.
//
// The returned path is what the caller owns afterwards, and it is empty in every
// case where the caller owns nothing: a missing key, and a failure that put the
// entry back. It is non-empty alongside an error only when the move could
// neither be made durable nor undone, because then the path is the sole record
// that makes the entry recoverable at all.
//
// A failure that put the entry back but could not make that repair durable is a
// *StateChangedError. The path is empty there too — the caller owns nothing, the
// bytes are live — but the slot is not provably what this call found, and
// StateChanged is how a caller stops claiming otherwise.
func (store *Store) Quarantine(key buildmeta.CacheKey, lock HomeLock) (string, error) {
	if err := requireHomeLock(lock); err != nil {
		return "", err
	}
	if store == nil || store.supported == nil || !store.supported() {
		return "", fmt.Errorf("persistent build cache protection is unsupported")
	}
	entryPath, _, err := store.paths(key)
	if err != nil {
		return "", err
	}
	moved, err := store.quarantinePath(entryPath, lock)
	return moved, stateChangeFor(key, err)
}

func (store *Store) quarantinePath(entryPath string, lock HomeLock) (string, error) {
	if err := requireHomeLock(lock); err != nil {
		return "", err
	}
	if _, err := os.Lstat(entryPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", fmt.Errorf("inspect cache entry for quarantine: %w", err)
	}
	parent := filepath.Dir(entryPath)
	name := quarantinePrefix + filepath.Base(entryPath) + "-"
	placeholder, err := os.CreateTemp(parent, name)
	if err != nil {
		return "", fmt.Errorf("reserve quarantine name: %w", err)
	}
	quarantinePath := placeholder.Name()
	if err := placeholder.Close(); err != nil {
		_ = os.Remove(quarantinePath)
		return "", err
	}
	if err := os.Remove(quarantinePath); err != nil {
		return "", err
	}
	if err := os.Rename(entryPath, quarantinePath); err != nil {
		return "", fmt.Errorf("quarantine cache entry: %w", err)
	}
	// The rename is the mutation, and every step after it has to answer for a
	// live slot that is already empty. Returning an error here without the path
	// it just moved was the whole hazard: the caller reads an ordinary failure,
	// has nothing to put back, and the logical entry a launcher resolves through
	// is simply gone. So a sync that fails puts the entry straight back, which
	// keeps the ordinary contract — an error means the slot is untouched — and
	// only a return that also fails reports the quarantine path, because at that
	// point the path is the only thing that makes the move recoverable at all.
	syncErr := store.fault(faultQuarantineSync)
	if syncErr == nil {
		syncErr = syncDirectory(parent)
	}
	if syncErr != nil {
		syncErr = fmt.Errorf("sync quarantined cache root: %w", syncErr)
		if back := renameDirectoryNoReplace(quarantinePath, entryPath); back != nil {
			return quarantinePath, errors.Join(syncErr,
				fmt.Errorf("return the quarantined cache entry to the live slot: %w", back))
		}
		// The move back is a mutation of the same directory the first sync failed
		// on, so it is worth exactly what its own sync is worth. Reporting the
		// ordinary "the slot is untouched" failure here would promise a repair
		// that has not reached the disk: the pathname resolves again now, and
		// whether it still does after a crash is the one thing that could not be
		// established. So the second sync gets the same treatment as the first,
		// and only a failure at it is reported as changed durable state — the
		// bytes stay live and recoverable either way.
		rollbackErr := store.fault(faultQuarantineRollbackSync)
		if rollbackErr == nil {
			rollbackErr = syncDirectory(parent)
		}
		if rollbackErr != nil {
			return "", &unsyncedRollbackError{Err: errors.Join(syncErr,
				fmt.Errorf("sync the cache root the entry was returned to: %w", rollbackErr))}
		}
		return "", syncErr
	}
	return quarantinePath, nil
}

// quarantineNames lists the quarantine directories of one entry that are
// present in a cache root right now.
//
// The manager home lock is exclusive, so this set only ever changes because of
// the caller holding it. That is what makes a before/after difference a sound
// answer to "did my own withdrawal move the entry before it failed" rather than
// a guess about a racing process.
func quarantineNames(parent, entryName string) map[string]bool {
	prefix := quarantinePrefix + entryName + "-"
	present := map[string]bool{}
	entries, err := os.ReadDir(parent)
	if err != nil {
		return present
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), prefix) {
			present[entry.Name()] = true
		}
	}
	return present
}

// withdrawnTo reports where a withdrawal that failed had already moved the live
// entry, and an empty path when the slot is untouched.
//
// It answers by looking at the cache root rather than by trusting the step it
// follows. A withdrawal is a rename plus the work that makes it durable, and
// anything in that tail can fail after the slot is already empty; a caller that
// cannot tell those two outcomes apart cannot compensate for either.
func withdrawnTo(entryPath, parent string, before map[string]bool) string {
	if _, err := os.Lstat(entryPath); err == nil || !errors.Is(err, os.ErrNotExist) {
		return ""
	}
	for name := range quarantineNames(parent, filepath.Base(entryPath)) {
		if !before[name] {
			return filepath.Join(parent, name)
		}
	}
	return ""
}

// withdrawEntry moves one live entry aside and reports what it did even when it
// fails.
//
// This is the single contract every compensation in this file rests on: if it
// returns an error, either the entry is still live and the returned path is
// empty, or the returned path names exactly where the entry went. Both callers
// own the result unconditionally — Publish records it as the predecessor to put
// back, a reversal puts it straight back — so neither can be left holding an
// error it cannot act on.
func (store *Store) withdrawEntry(entryPath string, point faultPoint, lock HomeLock) (string, error) {
	parent := filepath.Dir(entryPath)
	before := quarantineNames(parent, filepath.Base(entryPath))
	if err := store.fault(point); err != nil {
		return withdrawnTo(entryPath, parent, before), fmt.Errorf("quarantine cache entry: %w", err)
	}
	moved, err := store.quarantinePath(entryPath, lock)
	if err != nil && moved == "" {
		moved = withdrawnTo(entryPath, parent, before)
	}
	return moved, err
}

func requireHomeLock(lock HomeLock) error {
	if lock == nil || (reflect.ValueOf(lock).Kind() == reflect.Pointer && reflect.ValueOf(lock).IsNil()) {
		return fmt.Errorf("caller-held manager-home lock is required")
	}
	if err := lock.AssertHeld(); err != nil {
		return fmt.Errorf("manager-home lock is not held: %w", err)
	}
	return nil
}

func openRegularSource(path string) (*os.File, error) {
	if path == "" {
		return nil, fmt.Errorf("publication artifact source is empty")
	}
	lstat, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("inspect publication artifact: %w", err)
	}
	if !lstat.Mode().IsRegular() {
		return nil, fmt.Errorf("publication artifact is not a regular file")
	}
	file, err := os.Open(path) // #nosec G304 -- caller-provided private staging input is verified below
	if err != nil {
		return nil, fmt.Errorf("open publication artifact: %w", err)
	}
	opened, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	if !opened.Mode().IsRegular() || !os.SameFile(lstat, opened) {
		_ = file.Close()
		return nil, fmt.Errorf("publication artifact changed while opening")
	}
	return file, nil
}
