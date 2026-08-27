// Package managerlock provides cross-process locks for Curator operations.
// Lock files are stable manager state; release drops only the operating-system
// lock and never removes a path that another process may already have opened.
package managerlock

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"
)

var (
	// ErrLockOrder reports an acquisition that would invert the required
	// project -> optional build key -> manager-home order.
	ErrLockOrder = errors.New("manager lock acquisition order violation")
	// ErrDryRun reports an attempted acquisition on a dry-run path. The error
	// is returned before any lock directory or file is created.
	ErrDryRun = errors.New("dry-run must not acquire manager locks")
	// ErrNotHeld reports a released or otherwise invalid lock witness.
	ErrNotHeld = errors.New("manager lock is not held")
)

// Manager derives every lock path from one configured Curator home. New is
// read-only; lock state is created lazily by a real acquisition.
type Manager struct {
	mu             sync.RWMutex
	configuredHome string
	home           string
	lockRoot       string
	process        *processState
	prepared       bool
}

// New constructs a manager without creating filesystem state.
func New(home string) (*Manager, error) {
	canonical, err := canonicalAbsolute(home, "manager home")
	if err != nil {
		return nil, err
	}
	configuredHome, err := filepath.Abs(home)
	if err != nil {
		return nil, fmt.Errorf("resolve manager home: %w", err)
	}
	configuredHome = filepath.Clean(configuredHome)
	return &Manager{
		configuredHome: configuredHome,
		home:           canonical,
		lockRoot:       filepath.Join(canonical, "state", "locks", "v1"),
		process:        stateForHome(canonical),
	}, nil
}

// Home is the canonical absolute configured Curator home.
func (manager *Manager) Home() string {
	if manager == nil {
		return ""
	}
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return manager.home
}

// prepare creates the configured home for a real acquisition, then resolves
// its now-observable physical identity before any process reservation or lock
// path is selected. This closes the first-use gap on Windows, where case
// sensitivity belongs to each directory and cannot be known for a
// multi-component suffix until those directories exist. Dry-run paths never
// call prepare.
func (manager *Manager) prepare() error {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if manager.prepared {
		return nil
	}
	if err := os.MkdirAll(manager.configuredHome, 0o700); err != nil {
		return fmt.Errorf("create manager home: %w", err)
	}
	canonical, err := canonicalAbsolute(manager.configuredHome, "manager home")
	if err != nil {
		return err
	}
	manager.home = canonical
	manager.lockRoot = filepath.Join(canonical, "state", "locks", "v1")
	manager.process = stateForHome(canonical)
	manager.prepared = true
	return nil
}

func (manager *Manager) processState() *processState {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return manager.process
}

// Operation owns one ordered project operation. A dry-run operation rejects
// every acquisition before filesystem access.
type Operation struct {
	mu           sync.Mutex
	manager      *Manager
	dryRun       bool
	closed       bool
	projects     []*fileLock
	identities   []ProjectIdentity
	buildKey     *fileLock
	buildKeyUsed bool
	home         *HomeLock
}

// NewOperation creates an operation state machine. It performs no I/O.
func (manager *Manager) NewOperation(dryRun bool) *Operation {
	return &Operation{manager: manager, dryRun: dryRun}
}

// AcquireProjects canonicalizes and sorts the complete batch before acquiring
// it. Later batches are permitted only when every new identity follows the
// last held identity, so an inversion fails before waiting on an OS lock.
func (operation *Operation) AcquireProjects(ctx context.Context, projects ...string) ([]ProjectIdentity, error) {
	if operation == nil || operation.manager == nil {
		return nil, fmt.Errorf("manager operation is nil")
	}
	if operation.dryRun {
		return nil, ErrDryRun
	}
	identities, err := CanonicalProjects(projects...)
	if err != nil {
		return nil, err
	}
	if len(identities) == 0 {
		return nil, fmt.Errorf("at least one project lock is required")
	}

	operation.mu.Lock()
	defer operation.mu.Unlock()
	if operation.closed {
		return nil, orderError("operation is closed")
	}
	if operation.buildKeyUsed || operation.home != nil {
		return nil, orderError("project lock requested after a later lock class")
	}
	if len(operation.identities) > 0 && compareProject(operation.identities[len(operation.identities)-1], identities[0]) >= 0 {
		return nil, orderError("project identities are not strictly increasing")
	}
	if err := operation.manager.prepare(); err != nil {
		return nil, err
	}
	process := operation.manager.processState()
	if err := process.reserve(acquireProject); err != nil {
		return nil, err
	}
	acquired := false
	defer func() { process.finish(acquireProject, acquired) }()

	batch := make([]*fileLock, 0, len(identities))
	for _, identity := range identities {
		lock, acquireErr := acquireFileLock(ctx, operation.manager.projectLockPath(identity), nil)
		if acquireErr != nil {
			for index := len(batch) - 1; index >= 0; index-- {
				acquireErr = errors.Join(acquireErr, batch[index].Close())
			}
			return nil, acquireErr
		}
		batch = append(batch, lock)
	}
	acquired = true
	operation.projects = append(operation.projects, batch...)
	operation.identities = append(operation.identities, identities...)
	return append([]ProjectIdentity(nil), operation.identities...), nil
}

// AcquireBuildKey takes the optional per-key deduplication lock. An operation
// may use at most one key, and must release it before acquiring the home lock.
func (operation *Operation) AcquireBuildKey(ctx context.Context, logicalKey string) error {
	if operation == nil || operation.manager == nil {
		return fmt.Errorf("manager operation is nil")
	}
	if operation.dryRun {
		return ErrDryRun
	}
	if logicalKey == "" || !utf8.ValidString(logicalKey) || strings.IndexByte(logicalKey, 0) >= 0 {
		return fmt.Errorf("logical build key is empty or invalid UTF-8")
	}

	operation.mu.Lock()
	defer operation.mu.Unlock()
	if operation.closed || len(operation.projects) == 0 || operation.buildKeyUsed || operation.home != nil {
		return orderError("build-key lock requires held project locks and may be used once before the home lock")
	}
	process := operation.manager.processState()
	if err := process.reserve(acquireBuildKey); err != nil {
		return err
	}
	acquired := false
	defer func() { process.finish(acquireBuildKey, acquired) }()
	lock, err := acquireFileLock(ctx, operation.manager.buildKeyLockPath(logicalKey), func() {
		process.released(acquireBuildKey)
	})
	if err != nil {
		return err
	}
	acquired = true
	operation.buildKey = lock
	operation.buildKeyUsed = true
	return nil
}

// ReleaseBuildKey releases the optional key lock before the serialized commit
// phase. The persistent lock file deliberately remains in manager state.
func (operation *Operation) ReleaseBuildKey() error {
	if operation == nil {
		return nil
	}
	operation.mu.Lock()
	defer operation.mu.Unlock()
	if operation.buildKey == nil {
		return orderError("build-key lock is not held")
	}
	err := operation.buildKey.Close()
	operation.buildKey = nil
	return err
}

// HomeLock is an exclusive manager-home mutation witness. It satisfies the
// caller-held witness expected by buildcache and future transaction code.
type HomeLock struct {
	lock *fileLock
}

// AssertHeld verifies that the operating-system lock has not been released.
func (lock *HomeLock) AssertHeld() error {
	if lock == nil {
		return ErrNotHeld
	}
	return lock.lock.AssertHeld()
}

// Close releases the operating-system lock without deleting its stable file.
func (lock *HomeLock) Close() error {
	if lock == nil {
		return nil
	}
	return lock.lock.Close()
}

// AcquireHome acquires the mutation lock after project locks and after an
// optional build-key lock has been released.
func (operation *Operation) AcquireHome(ctx context.Context) (*HomeLock, error) {
	if operation == nil || operation.manager == nil {
		return nil, fmt.Errorf("manager operation is nil")
	}
	if operation.dryRun {
		return nil, ErrDryRun
	}
	operation.mu.Lock()
	defer operation.mu.Unlock()
	if operation.closed || len(operation.projects) == 0 || operation.buildKey != nil || operation.home != nil {
		return nil, orderError("home lock requires held project locks and no held build-key lock")
	}
	return operation.acquireHome(ctx)
}

func (operation *Operation) acquireHome(ctx context.Context) (*HomeLock, error) {
	process := operation.manager.processState()
	if err := process.reserve(acquireHome); err != nil {
		return nil, err
	}
	acquired := false
	defer func() { process.finish(acquireHome, acquired) }()
	fileLock, err := acquireFileLock(ctx, operation.manager.homeLockPath(), func() {
		process.released(acquireHome)
	})
	if err != nil {
		return nil, err
	}
	acquired = true
	operation.home = &HomeLock{lock: fileLock}
	return operation.home, nil
}

// AcquireHomeOnly is the recovery/GC path. It acquires no project or key lock.
// dryRun rejects the request without creating any manager state.
func (manager *Manager) AcquireHomeOnly(ctx context.Context, dryRun bool) (*HomeLock, error) {
	if manager == nil {
		return nil, fmt.Errorf("manager is nil")
	}
	if dryRun {
		return nil, ErrDryRun
	}
	if err := manager.prepare(); err != nil {
		return nil, err
	}
	process := manager.processState()
	if err := process.reserve(acquireHome); err != nil {
		return nil, err
	}
	acquired := false
	defer func() { process.finish(acquireHome, acquired) }()
	fileLock, err := acquireFileLock(ctx, manager.homeLockPath(), func() {
		process.released(acquireHome)
	})
	if err != nil {
		return nil, err
	}
	acquired = true
	return &HomeLock{lock: fileLock}, nil
}

// Close releases held locks in reverse class and project order.
func (operation *Operation) Close() error {
	if operation == nil {
		return nil
	}
	operation.mu.Lock()
	defer operation.mu.Unlock()
	if operation.closed {
		return nil
	}
	operation.closed = true
	var result error
	if operation.home != nil {
		result = errors.Join(result, operation.home.Close())
		operation.home = nil
	}
	if operation.buildKey != nil {
		result = errors.Join(result, operation.buildKey.Close())
		operation.buildKey = nil
	}
	for index := len(operation.projects) - 1; index >= 0; index-- {
		result = errors.Join(result, operation.projects[index].Close())
	}
	operation.projects = nil
	return result
}

func (manager *Manager) projectLockPath(identity ProjectIdentity) string {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return filepath.Join(manager.lockRoot, "projects", lockName("project", string(identity)))
}

func (manager *Manager) buildKeyLockPath(logicalKey string) string {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return filepath.Join(manager.lockRoot, "build", lockName("build-key", logicalKey))
}

func (manager *Manager) homeLockPath() string {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return filepath.Join(manager.lockRoot, lockName("manager-home", manager.home))
}

func compareProject(left, right ProjectIdentity) int {
	leftBytes, rightBytes := []byte(left), []byte(right)
	for index := 0; index < len(leftBytes) && index < len(rightBytes); index++ {
		if leftBytes[index] < rightBytes[index] {
			return -1
		}
		if leftBytes[index] > rightBytes[index] {
			return 1
		}
	}
	switch {
	case len(leftBytes) < len(rightBytes):
		return -1
	case len(leftBytes) > len(rightBytes):
		return 1
	default:
		return 0
	}
}

func orderError(detail string) error {
	return fmt.Errorf("%w: %s", ErrLockOrder, detail)
}

type acquisitionKind uint8

const (
	acquireProject acquisitionKind = iota
	acquireBuildKey
	acquireHome
)

type processState struct {
	mu             sync.Mutex
	projectPending int
	keyPending     int
	keysHeld       int
	homePending    int
	homesHeld      int
}

var processStates = struct {
	sync.Mutex
	byHome map[string]*processState
}{byHome: map[string]*processState{}}

func stateForHome(home string) *processState {
	processStates.Lock()
	defer processStates.Unlock()
	if state := processStates.byHome[home]; state != nil {
		return state
	}
	state := new(processState)
	processStates.byHome[home] = state
	return state
}

func (state *processState) reserve(kind acquisitionKind) error {
	state.mu.Lock()
	defer state.mu.Unlock()
	switch kind {
	case acquireProject:
		if state.homesHeld > 0 || state.homePending > 0 {
			return orderError("project lock requested while the process holds or is acquiring the home lock")
		}
		state.projectPending++
	case acquireBuildKey:
		if state.homesHeld > 0 || state.homePending > 0 {
			return orderError("build-key lock requested while the process holds or is acquiring the home lock")
		}
		state.keyPending++
	case acquireHome:
		if state.keysHeld > 0 || state.keyPending > 0 || state.projectPending > 0 {
			return orderError("home lock requested before an in-process project/key acquisition completed or key release")
		}
		state.homePending++
	}
	return nil
}

func (state *processState) finish(kind acquisitionKind, acquired bool) {
	state.mu.Lock()
	defer state.mu.Unlock()
	switch kind {
	case acquireProject:
		state.projectPending--
	case acquireBuildKey:
		state.keyPending--
		if acquired {
			state.keysHeld++
		}
	case acquireHome:
		state.homePending--
		if acquired {
			state.homesHeld++
		}
	}
}

func (state *processState) released(kind acquisitionKind) {
	state.mu.Lock()
	defer state.mu.Unlock()
	switch kind {
	case acquireBuildKey:
		if state.keysHeld > 0 {
			state.keysHeld--
		}
	case acquireHome:
		if state.homesHeld > 0 {
			state.homesHeld--
		}
	}
}
