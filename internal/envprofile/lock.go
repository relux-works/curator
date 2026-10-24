package envprofile

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/projectlease"
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
	engine *transaction.Engine
	// Prepared under this lock before recovery or repair publication.
	claudePrepared *claudeSeedPlan
	recovered      bool
	undo           *globalUndo
}

// newOperationEngine keeps recovery observation available to entry-point tests.
var newOperationEngine = transaction.New

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
	op.engine, err = newOperationEngine(home, append(options, transaction.WithCommitCheck(projectlease.TransactionCheck(home)))...)
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
// The per-entry agent-home payloads are not transaction targets in this
// stage (see the switch.go package doc for why); they stay direct writes
// under the held lock.
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
	engine, err := newOperationEngine(home, append(options, transaction.WithCommitCheck(projectlease.TransactionCheck(home)))...)
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
	for path := range files {
		if err := op.capture(path); err != nil {
			return err
		}
	}
	if err := op.capture(removals...); err != nil {
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
	for _, item := range changes {
		live := item.path
		if err := os.MkdirAll(filepath.Dir(live), 0o755); err != nil {
			return err
		}
		preimage, err := transaction.DigestTarget(transaction.KindBytes, live)
		if err != nil {
			return fmt.Errorf("read the current state of %s: %w", live, err)
		}
		target := transaction.Target{
			Class: "profile", Identifier: live, Kind: transaction.KindBytes,
			LivePath: live, PreimageDigest: preimage,
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
		if err := privatedir.ProtectFile(item.path); err != nil {
			return fmt.Errorf("protect published profile record %s: %w", item.path, err)
		}
	}
	var publishedPaths []string
	for _, item := range changes {
		publishedPaths = append(publishedPaths, item.path)
	}
	return op.recordPostimages(publishedPaths...)
}

// newTransactionID names one profile journal.
func newTransactionID() string {
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return fmt.Sprintf("profile-%d", time.Now().UnixNano())
	}
	return "profile-" + hex.EncodeToString(random[:])
}
