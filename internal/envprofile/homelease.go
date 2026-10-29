package envprofile

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/homelease"
	"github.com/relux-works/curator/internal/homelease/state"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/projectlease"
	"github.com/relux-works/curator/internal/transaction"
)

type homeLeaseAuthority struct{ request ResolveRequest }

// HomeLeaseManager supplies authority from the existing read-only materializer.
// Observer is a separately qualified consumer integration, never holder input.
func HomeLeaseManager(req ResolveRequest, observer homelease.Observer) (*homelease.Manager, error) {
	return homelease.New(req.Home, homeLeaseAuthority{request: req}, observer)
}

func (a homeLeaseAuthority) Admit(_ context.Context, held managerlock.Held, environment, profile, digest string) (homelease.Admission, error) {
	req := a.request
	if err := held.Validate(req.Home); err != nil {
		return homelease.Admission{}, err
	}
	req.EnvID, req.Profile, req.Format = environment, profile, "json"
	req.Repair, req.DryRun, req.PreserveLock = false, false, false
	req.homeLeaseHeld = &held
	_, lock, hash, err := loadResolveInputs(req.Home, profile)
	if err != nil {
		if strings.Contains(err.Error(), DiagProfileUnknown) {
			return homelease.Admission{}, state.Refuse("profile_unknown", "profile_absent")
		}
		return homelease.Admission{}, err
	}
	if strings.TrimPrefix(hash, "sha256:") != digest {
		return homelease.Admission{}, state.Refuse("environment_lease_lock_mismatch", "profile_lock_changed")
	}
	if _, err := resolveVerified(req, verifyHomeContent); err != nil {
		if strings.Contains(err.Error(), DiagHomeStale) {
			return homelease.Admission{}, state.Refuse("environment_home_stale", "home_not_current")
		}
		return homelease.Admission{}, err
	}
	actual := ManagedHomeDir(req.Home, profile, environment)
	protected := []string{actual, sourcePath(req.Home, profile), lockPath(req.Home, profile), prevLockPath(req.Home, profile), storeDocPath(req.Home, profile, environment, "")}
	for _, member := range lock.Members {
		protected = append(protected, contextstore.EntryDir(req.Home, member.Kind, member.Name, member.PinKey()))
	}
	skills, err := skillsOf(req.Home, newGitManager(req.Home), lock)
	if err != nil {
		return homelease.Admission{}, err
	}
	for _, skill := range skills {
		protected = append(protected, skill.root)
	}
	return homelease.Admission{Home: actual, Protected: protected, BuildKeys: []string{}, RuntimeEntries: []string{}}, nil
}

// homeLeaseGuard is shared by direct home writers and borrowed global phases.
// Its caller keeps the same operation acquisition until the final write.
func homeLeaseGuard(home string, lock transaction.HomeLock, targets []string) error {
	if state.HasState(home) {
		engine, err := transaction.New(home)
		if err != nil {
			return err
		}
		pending, err := engine.PendingRecoveryFor(lock, "home-leases")
		if err != nil {
			return state.Refuse("environment_lease_unreadable", "state_unreadable")
		}
		if pending {
			return state.Refuse("environment_lease_recovery_required", "pending_mutation")
		}
		if _, err := projectlease.ReferencesUnderLock(home, lock); err != nil {
			return state.Refuse("environment_lease_unreadable", "state_unreadable")
		}
	}
	return state.GuardMutation(home, lock, "profiles", targets)
}

func homeLeaseGuardSelection(op *operation, home, name, environment string) error {
	targets, err := homeLeaseSelectionTargets(home, name, environment)
	if err != nil {
		return err
	}
	return homeLeaseGuard(home, op.lock, targets)
}

func homeLeaseProfileTargets(home, profile, environment string) []string {
	if profile == "" {
		return nil
	}
	if environment != "" {
		return []string{ManagedHomeDir(home, profile, envregistry.NormalizeEnvID(environment))}
	}
	return []string{filepath.Join(EnvRoot(home), profile), ProfileDir(home, profile)}
}

func homeLeaseRefusal(err error) bool {
	var refusal *state.Refusal
	return errors.As(err, &refusal)
}

// Compensation is a writer too. Validate the original continuous witness and
// the complete rollback set before restoring a single surface or parent.
func homeLeaseGuardUndo(op *operation) error {
	if op.undo == nil {
		return nil
	}
	targets := make([]string, 0, len(op.undo.entries)+len(op.undo.parentPostimages))
	for path := range op.undo.entries {
		targets = append(targets, path)
	}
	for path := range op.undo.parentPostimages {
		targets = append(targets, path)
	}
	return homeLeaseGuard(op.undo.managerHome, op.lock, targets)
}

func homeLeaseRead(home string, held *managerlock.Held) error {
	if !state.HasState(home) && !projectlease.HasState(home) {
		return homeLeaseReadLegacy(home, held)
	}
	if held != nil {
		if err := held.Validate(home); err != nil {
			return err
		}
		if err := state.LegacyPending(home); err != nil {
			return err
		}
		lock := homeLeaseReadWitness{home: home, held: *held}
		if _, err := projectlease.ReferencesUnderLock(home, lock); err != nil {
			return state.Refuse("environment_lease_unreadable", "state_unreadable")
		}
		engine, err := transaction.New(home)
		if err != nil {
			return err
		}
		pending, err := engine.PendingRecovery(lock)
		if err != nil {
			return state.Refuse("environment_lease_unreadable", "state_unreadable")
		}
		if pending {
			return state.Refuse("environment_lease_recovery_required", "pending_mutation")
		}
		_, err = state.Load(home)
		return err
	}
	m, err := homelease.New(home, nil, nil)
	if err != nil {
		return err
	}
	_, err = m.List(context.Background(), "", "")
	return err
}

type homeLeaseReadWitness struct {
	home string
	held managerlock.Held
}

func (w homeLeaseReadWitness) AssertHeld() error      { return w.held.Validate(w.home) }
func (w homeLeaseReadWitness) Held() managerlock.Held { return w.held }

func homeLeaseReadRequest(req ResolveRequest) error {
	if req.homeLeaseHeld != nil {
		return homeLeaseRead(req.Home, req.homeLeaseHeld)
	}
	if req.borrowedOperation != nil {
		witness, ok := req.borrowedOperation.lock.(interface{ Held() managerlock.Held })
		if !ok {
			return managerlock.ErrNotHeld
		}
		held := witness.Held()
		return homeLeaseRead(req.Home, &held)
	}
	return homeLeaseRead(req.Home, nil)
}

func homeLeaseGuardPlan(op *operation, req *ResolveRequest, plan *homePlan) error {
	targets := homeLeaseProfileTargets(req.Home, req.Profile, req.EnvID)
	for path := range plan.docs {
		targets = append(targets, path)
	}
	for _, skill := range plan.skillProjections {
		if verifySkillProjection(req.Home, skill) != nil {
			targets = append(targets, skill.root)
		}
	}
	return homeLeaseGuard(req.Home, op.lock, targets)
}

func beginOperationGuarded(home string, validate func() error, targets []string, options ...transaction.Option) (*operation, error) {
	return beginOperationSelected(home, validate, func() ([]string, error) { return targets, nil }, options...)
}

func homeLeaseDefaultTargets(home string) ([]string, error) {
	_, err := readSource(home, DefaultProfile)
	if err == nil {
		return []string{}, nil
	}
	if !isStateAbsent(err) {
		return nil, err
	}
	return homeLeaseProfileTargets(home, DefaultProfile, ""), nil
}

func homeLeaseInstallTargets(home string, options InstallOptions) ([]string, error) {
	if !state.HasState(home) && !projectlease.HasState(home) {
		return []string{}, nil
	}
	name := options.As
	if name == "" {
		if installOperandKind(options.Operand) != identity.SourcePath {
			// A remote package's installed name is unknown until resolution.
			// Never claim a disjoint write set from the caller's URL spelling.
			return nil, nil
		}
		manifest, err := contextpkg.LoadManifest(options.Operand)
		if err != nil {
			return nil, nil // Preserve the install entry point's existing typed refusal.
		}
		name = manifest.Name
	}
	targets := homeLeaseProfileTargets(home, name, "")
	activate, err := reinstallActivation(home, name, options.Use)
	if err != nil {
		return nil, err
	}
	if activate {
		selection, err := homeLeaseSelectionTargets(home, name, "")
		if err != nil {
			return nil, err
		}
		targets = append(targets, selection...)
	}
	defaults, err := homeLeaseDefaultTargets(home)
	return append(targets, defaults...), err
}

// Selection is reread under the acquired lock: a pre-lock current profile is
// not authority for the homes a switch or direct pointer setter can affect.
func homeLeaseSelectionTargets(home, name, environment string) ([]string, error) {
	current, err := Current(home)
	if err != nil {
		return nil, err
	}
	scoped, err := ScopedCurrents(home)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0)
	for _, profile := range []string{name, current} {
		if profile != "" {
			paths = append(paths, homeLeaseProfileTargets(home, profile, environment)...)
		}
	}
	for scope, profile := range scoped {
		env, _ := splitScope(scope)
		if profile != "" && (environment == "" || env == environment) {
			paths = append(paths, homeLeaseProfileTargets(home, profile, environment)...)
		}
	}
	native, err := homeLeaseNativeTargets(environment, nil)
	return append(paths, native...), err
}

func homeLeaseNativeTargets(environment string, resolver func(string) (string, error)) ([]string, error) {
	paths := []string{}
	for _, adapter := range Adapters {
		if environment != "" && adapter.ID != envregistry.NormalizeEnvID(environment) {
			continue
		}
		path, err := resolveNativeHome(adapter, resolver)
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func homeLeaseUnmanageTargets(req UnmanageRequest) ([]string, error) {
	paths, err := homeLeaseSelectionTargets(req.Home, "", envregistry.NormalizeEnvID(req.EnvID))
	if err != nil {
		return nil, err
	}
	native, err := homeLeaseNativeTargets(req.EnvID, req.NativeHomeOf)
	return append(paths, native...), err
}

// Reuse the read-only ownership planner so a newly added lease gate cannot
// mask an existing MCP ownership/policy refusal. Preview never recovers or
// publishes; the operation callback repeats it under the real lock.
func homeLeaseMCPPreflight(req ProfileMCPRequest) error {
	if req.DryRun || req.Mode == "status" || !state.HasState(req.Home) {
		return nil
	}
	req.DryRun, req.beforePublish, req.transactionOptions = true, nil, nil
	_, err := ProfileMCP(req)
	return err
}

func SetCurrent(home, name string) error {
	op, err := beginOperationSelectedRecovery(home, nil, func() ([]string, error) { return homeLeaseSelectionTargets(home, name, "") }, false)
	if err != nil {
		return err
	}
	defer func() { _ = op.close() }()
	return setCurrentDirect(op, home, name)
}

func SetScoped(home, scope, name string, clearScope bool) error {
	environment, _ := splitScope(scope)
	op, err := beginOperationSelectedRecovery(home, nil, func() ([]string, error) { return homeLeaseSelectionTargets(home, name, environment) }, false)
	if err != nil {
		return err
	}
	defer func() { _ = op.close() }()
	return setScopedDirect(op, home, scope, name, clearScope)
}

// SyncHomesWithPolicy retains one operation across the command's managed and
// native phases. Neither phase releases the admission lock to a new acquire.
func SyncHomesWithPolicy(home string, policy Policy, machine envregistry.MachineConfig) (managed, native []EntryResult, managedErr, nativeErr error) {
	op, err := beginOperation(home)
	if err != nil {
		return nil, nil, err, nil
	}
	defer func() { _ = op.close() }()
	managed, managedErr = syncManagedHomesWithOperation(op, home, policy, machine)
	if managedErr != nil {
		return
	}
	native, nativeErr = syncWithOperation(op, home, policy)
	return
}

// UpdateProfilesWithOptions preflights the entire command before the first
// profile changes and borrows the same operation for every update.
func UpdateProfilesWithOptions(home string, names []string, options UpdateOptions, visit func(string, Info, bool, error) bool) error {
	if err := options.Policy.checkHardenedProfileMutation(); err != nil {
		return err
	}
	targets := make([]string, 0)
	for _, name := range names {
		targets = append(targets, homeLeaseProfileTargets(home, name, "")...)
	}
	op, err := beginOperationGuarded(home, nil, targets)
	if err != nil {
		return err
	}
	defer func() { _ = op.close() }()
	for _, name := range names {
		info, moved, err := updateLocked(op, home, name, options.Policy, options.SurfacingSink, options.WarningSink, options.ConfirmSystemDelta, options.readRegularFile)
		if !visit(name, info, moved, err) {
			break
		}
	}
	return nil
}
