package envprofile

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/relux-works/curator/internal/envunit"
	"os"
	"sort"
	"time"

	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/projectlease"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/transaction"
)

// lockTimeout bounds one wait for the manager-home mutation lock. Tests
// shorten it for contention cases; production waits up to 30 seconds and
// then reports environment_lock_unavailable.
var lockTimeout = 30 * time.Second

// operation binds one profile mutation to the held manager-home mutation
// lock and the profile transaction journal.
type operation struct {
	lock interface {
		AssertHeld() error
		AssertCurrent() error
		Close() error
	}
	engine           *transaction.Engine
	policy           Policy
	publicationGroup *homePublicationGroup
	publishPlan      func(transaction.Plan) error
	joinedLegacy     *transaction.Plan
	stageLegacy      func(*envunit.Unit) (transaction.Plan, error)
	initialProfile   map[string][]byte // exact first-use installed records, private until Prepare
	retained         *envlifecycle.GlobalRetained
	// Prepared under this lock before recovery or repair publication.
	claudePrepared *claudeSeedPlan
	recovered      bool
}

// newOperationEngine keeps recovery observation available to entry-point tests.
var newOperationEngine = func(home string, options ...transaction.Option) (*transaction.Engine, error) {
	return transaction.New(home, append(options, transaction.WithCommitCheck(projectlease.TransactionCheck(home)))...)
}

// beginCaptureOperation never provisions a home or invokes recovery. The
// observer validates every journal before reporting pending recovery.
func beginCaptureOperation(home string, options ...transaction.Option) (*operation, error) {
	if err := envauthority.Validate(home); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), lockTimeout)
	defer cancel()
	manager, err := managerlock.New(home)
	if err != nil {
		return nil, err
	}
	lock, err := manager.AcquireExistingHomeOnly(ctx)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%s: source has no existing mutation lock", DiagHomeStale)
		}
		return nil, fmt.Errorf("%s: acquire capture lock: %w", DiagLockUnavailable, err)
	}
	if err := envauthority.Validate(home); err != nil {
		_ = lock.Close()
		return nil, err
	}
	op := &operation{lock: lock}
	op.engine, err = newOperationEngine(home, options...)
	if err == nil {
		var pending bool
		pending, err = op.engine.PendingRecovery(lock)
		if err == nil && pending {
			err = fmt.Errorf("%s: explicit recovery is required", DiagMaterializeRecoveryRequired)
		}
	}
	if err != nil {
		_ = op.close()
		return nil, err
	}
	return op, nil
}

// beginOperation takes the manager-home mutation lock and recovers
// incomplete profile journals before the caller mutates anything: every
// profile mutation runs serialized like any other manager-home
// transaction, and its manager-home records publish through op.publish.
// Managed and native home publishers additionally enlist their enumerated
// home/marker/origin and metadata effects in complete durable units.
func beginOperation(home string, options ...transaction.Option) (*operation, error) {
	return beginOperationValidated(home, nil, options...)
}

// beginOperationValidated checks read-only inputs under the lock before journal
// recovery can write. All mutations also validate manager-wide authority.
func beginOperationValidated(home string, validate func() error, options ...transaction.Option) (*operation, error) {
	// Refuse before even creating mutation-lock state in an existing node home.
	if err := envauthority.Validate(home); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), lockTimeout)
	defer cancel()
	manager, err := managerlock.New(home)
	if err != nil {
		return nil, err
	}
	lock, err := manager.AcquireHomeOnly(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("%s: acquire the manager-home mutation lock: %v", DiagLockUnavailable, err)
	}
	// Authority may have changed while waiting for the lock.
	if err := envauthority.Validate(home); err != nil {
		_ = lock.Close()
		return nil, err
	}
	if validate != nil {
		if err := validate(); err != nil {
			_ = lock.Close()
			return nil, err
		}
	}
	engine, err := newOperationEngine(home, options...)
	if err != nil {
		_ = lock.Close()
		return nil, err
	}
	recovered, err := projectlease.RecoverUnderLock(home, lock, engine)
	if err != nil {
		_ = lock.Close()
		if isCredentialRefusal(err) {
			return nil, err
		}
		return nil, fmt.Errorf("recover incomplete profile transactions: %w", err)
	}
	op := &operation{lock: lock, engine: engine, recovered: recovered}
	// Previously durable intent may itself replace the authority marker.
	if err := envauthority.Validate(home); err != nil {
		_ = op.close()
		return nil, op.recoveryRefusal(err)
	}
	return op, nil
}

// Preserve the original diagnostic and error chain while reporting only an
// actually completed recovery of previously durable intent.
func (op *operation) recoveryRefusal(err error) error {
	if op.recovered {
		return fmt.Errorf("%w; previously pending transaction recovery completed", err)
	}
	return err
}

func (op *operation) close() error {
	if op == nil || op.lock == nil {
		return nil
	}
	return op.lock.Close()
}

// publish commits manager-home record writes and removals as one journaled
// transaction. Write paths are staged and swapped under the held lock;
// removal paths use the transaction engine's absent desired state. A crash
// mid-publication resumes from the journal on the next operation instead of
// stranding a half-written record set.
func (op *operation) publish(files map[string][]byte, removals ...string) error {
	return op.publishProtected(files, nil, nil, removals...)
}

// publishProtected adds planning-time boundary identities and a current-input
// check to the existing journal publisher. Legacy profile calls remain nil.
func (op *operation) publishProtected(files map[string][]byte, boundaries *staging.Boundaries, validate func() error, removals ...string) error {
	if handled, err := op.publishPurgedReuse(files, removals, boundaries, validate); handled || err != nil {
		return err
	}
	if handled, err := op.publishManagedRebind(op.engine.Home(), files, removals, boundaries, validate); handled || err != nil {
		return err
	}
	paths := append([]string(nil), removals...)
	for path := range files {
		paths = append(paths, path)
	}
	if err := envauthority.AdmitEffects(op.engine.Home(), nil, paths); err != nil {
		return err
	}
	if len(files) == 0 && len(removals) == 0 {
		return nil
	}
	type change struct {
		path    string
		payload []byte
		remove  bool
	}
	changesByPath := make(map[string]change, len(files)+len(removals))
	for path := range files {
		changesByPath[path] = change{path: path, payload: files[path]}
	}
	for _, path := range removals {
		if _, exists := changesByPath[path]; exists {
			return fmt.Errorf("profile transaction both writes and removes %s", path)
		}
		changesByPath[path] = change{path: path, remove: true}
	}
	changes := make([]change, 0, len(changesByPath))
	for _, item := range changesByPath {
		changes = append(changes, item)
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].path < changes[j].path })
	var temps []string
	defer func() {
		for _, temp := range temps {
			_ = os.Remove(temp)
		}
	}()
	plan := transaction.Plan{TransactionID: newTransactionID(), ProjectIdentity: "profiles"}
	var writePaths []string
	for _, item := range changes {
		if !item.remove {
			writePaths = append(writePaths, item.path)
		}
	}
	var err error
	plan.Directories, err = op.engine.PublicationDirectories(writePaths...)
	if err != nil {
		return fmt.Errorf("capture profile record parents: %w", err)
	}
	for _, item := range changes {
		live := item.path
		preimage, err := transaction.DigestTarget(transaction.KindBytes, live)
		if err != nil {
			return fmt.Errorf("read the current state of %s: %w", live, err)
		}
		target := transaction.Target{
			Class: "profile", Identifier: live, Kind: transaction.KindBytes,
			LivePath: live, PreimageDigest: preimage, PrivateFile: !item.remove,
		}
		if !item.remove {
			temp, err := privatedir.CreateTemp("", "curator-profile-record-*")
			if err != nil {
				return err
			}
			if err := privatedir.ValidateFile(temp.Name()); err != nil {
				_ = temp.Close()
				_ = os.Remove(temp.Name())
				return err
			}
			temps = append(temps, temp.Name())
			if _, err := temp.Write(item.payload); err != nil {
				_ = temp.Close()
				return err
			}
			if err := temp.Close(); err != nil {
				return err
			}
			target.StagedSource = temp.Name()
		}
		plan.Targets = append(plan.Targets, target)
	}
	if boundaries != nil {
		plan.BoundaryProof = &boundaries.Durable
		plan.BoundaryCheck = func(_ int, path string) error {
			if err := staging.RecheckOne(staging.Target{LivePath: path, Kind: staging.KindBytes}, boundaries.Snapshot, boundaries.Admitted); err != nil {
				return err
			}
			if validate != nil {
				return validate()
			}
			return nil
		}
	}
	if validate != nil {
		if err := validate(); err != nil {
			return err
		}
	}
	if _, err := op.engine.Prepare(op.lock, plan); err != nil {
		return fmt.Errorf("prepare profile records: %w", err)
	}
	for _, temp := range temps {
		_ = os.Remove(temp)
	}
	temps = nil
	if err := op.engine.Commit(op.lock, plan.TransactionID); err != nil {
		return fmt.Errorf("commit profile records: %w", err)
	}
	for _, item := range changes {
		if item.remove {
			continue
		}
		if err := privatedir.ValidateFile(item.path); err != nil {
			return fmt.Errorf("protect published profile record %s: %w", item.path, err)
		}
	}
	return nil
}

// newTransactionID names one profile journal.
func newTransactionID() string {
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return fmt.Sprintf("profile-%d", time.Now().UnixNano())
	}
	return "profile-" + hex.EncodeToString(random[:])
}
