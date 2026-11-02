// Package envprofile implements explicit read-only credential migration plans
// and locked apply. Existing homes use their verified provisioning origin, never
// the current caller's selectors. Apply enlists credential entries, marker,
// origin and affected profile identities in one durable environment unit. It
// never copies credential bytes. Returned failures compensate mutable preimages;
// interrupted units recover by their proven phase through common admission.
// Old unsealed link-only journals cannot authorize writes; completed legacy
// receipts allow only exact verified cleanup. Repair does not perform migration.
package envprofile

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/stateread"
	"github.com/relux-works/curator/internal/transaction"
)

// Migration render phases for MigrateReport.Text. Inspect prints the
// inventory detail; plan prints the operations, conflicts, and hash.
const (
	MigratePhaseInspect = "inspect"
	MigratePhasePlan    = "plan"
)

// Migration operation kinds.
const (
	MigrateOpRelink = "relink"
	MigrateOpUnlink = "unlink"
)

// Entry classifications for one recorded-or-wanted credential path.
const (
	MigrateCurrent        = "current"
	MigratePending        = "dangling-to-declared"
	MigrateMistargeted    = "mis-targeted"
	MigrateStaleRecorded  = "stale-recorded"
	MigrateRegularFile    = "regular-file-at-link"
	MigrateForeign        = "foreign"
	MigrateAbsent         = "absent"
	MigrateEmptyDir       = "empty-dir"
	MigrateDirConflict    = "directory-at-link"
	MigrateUninspectable  = "uninspectable"
	MigrateNoDeclaredRoot = "no-declared-root"
)

// Native-root states for the Pi inventory: the path stats present,
// absent, or fails to stat (fail-closed, never read as absent).
const (
	MigrateRootPresent       = "present"
	MigrateRootAbsent        = "absent"
	MigrateRootUninspectable = "uninspectable"
	MigrateRootNA            = "-"
)

// MigrateRequest scopes one credential migration operation. Profile and
// EnvID filter the scope; empty means every installed profile or every
// registered environment. Expect carries the plan hash a previous --plan
// printed; every apply entry requires it and refuses — before any
// mutation — when it is missing or when the recomputed hash differs
// (plan drift). Print receives the locked, revalidated plan text before
// the first mutation (a write failure fails the apply before anything
// changes); nil prints nothing and the caller uses
// MigrateApplyResult.Report instead. The private transactionOptions seam injects raw transaction faults;
// it never decides origin authority or recovery policy.
type MigrateRequest struct {
	Home               string
	Profile            string
	EnvID              string
	Machine            envregistry.MachineConfig
	Detect             func(envregistry.Adapter) string
	NativeHomeOf       func(string) (string, error)
	Expect             string
	Print              io.Writer
	transactionOptions []transaction.Option
}

// resolveFor builds the resolve shim the migration reuses for the native
// home, release detection, and effective-passthrough computation, so the
// migration and repair cannot disagree about the declared targets.
func (req *MigrateRequest) resolveFor(profile, envID string) *ResolveRequest {
	return &ResolveRequest{
		Home:         req.Home,
		Profile:      profile,
		EnvID:        envID,
		Machine:      req.Machine,
		Detect:       req.Detect,
		NativeHomeOf: req.NativeHomeOf,
	}
}

// MigrateOp is one exact link operation: relink moves a recorded symlink
// from its current target to the declared native path, unlink removes a
// stale recorded link (From empty means the link is already gone and
// only the marker record is dropped). To is empty for unlink.
type MigrateOp struct {
	Profile string
	EnvID   string
	Kind    string
	Path    string
	From    string
	To      string
}

// MigrateConflict is one state the operator must resolve out of band
// before apply runs: Detail reports both sides, Choice names the exact
// operator action. Any conflict in scope blocks the whole apply.
type MigrateConflict struct {
	Profile string
	EnvID   string
	Path    string
	Detail  string
	Choice  string
}

// MigrateEntry is one inventoried recorded-or-wanted path: its record
// and wanted state, its classification, and the current versus declared
// link target ("-" where none applies).
type MigrateEntry struct {
	Path     string
	Strategy string
	Recorded bool
	Wanted   bool
	State    string
	Current  string
	Declared string
	Note     string
}

// MigrateHome is one profile × environment inventory: the marker and
// effective mode, every entry, the Pi native roots (pi homes only;
// "-" otherwise), and the derived operations, conflicts, and warnings.
type MigrateHome struct {
	Profile   string
	EnvID     string
	HomeDir   string
	Isolation string
	Mode      string
	Entries   []MigrateEntry
	PiOldPath string
	PiNewPath string
	PiOld     string
	PiNew     string
	Ops       []MigrateOp
	Conflicts []MigrateConflict
	Warnings  []string

	provisioned       bool
	marker            *envmarker.Marker
	markerRaw         []byte
	credentialRecords []envmarker.Passthrough
}

// MigrateRecovery summarizes a leftover migration journal from an
// interrupted apply: the original plan hash, the operation count, and
// how many link operations had completed. Broken is non-empty when the
// journal cannot be used (unreadable or corrupt) and carries the
// diagnostic; Plan and Ops are then empty. A nil Recovery means no
// journal is present.
type MigrateRecovery struct {
	Plan   string
	Ops    int
	Done   int
	Broken string
}

// notice renders the recovery banner for both report phases.
func (rec *MigrateRecovery) notice(home string) string {
	if rec.Broken != "" {
		return fmt.Sprintf("migration journal %s is unusable (%s); back it up out of band, verify every managed link by hand, and only then remove it and re-run `curator env migrate --plan`\n",
			migrationJournalPath(home), rec.Broken)
	}
	return fmt.Sprintf("completed legacy migration receipt detected: plan %s, %d of %d link operations completed; the next `curator env migrate --apply --expect %s` validates and cleans the receipt under the manager lock, preserving verified postimages (this read-only command changes nothing)\n",
		rec.Plan, rec.Done, rec.Ops, rec.Plan)
}

// MigrateReport is the migration plan over the requested scope: every
// inventoried home plus the plan hash over the canonical inventory,
// operations, and conflicts.
type MigrateReport struct {
	Home           string
	ScopeProf      string
	ScopeEnv       string
	Homes          []MigrateHome
	Recovery       *MigrateRecovery
	Hash           string
	identityInputs map[string]string
}

// Ops flattens the report's operations in plan order.
func (r *MigrateReport) Ops() []MigrateOp {
	var out []MigrateOp
	for _, home := range r.Homes {
		out = append(out, home.Ops...)
	}
	return out
}

// Conflicts flattens the report's conflicts in plan order.
func (r *MigrateReport) Conflicts() []MigrateConflict {
	var out []MigrateConflict
	for _, home := range r.Homes {
		out = append(out, home.Conflicts...)
	}
	return out
}

// conflictError renders the same operator choice for locked apply and early
// read-only authority refusal; neither path applies a conflicted plan.
func (r *MigrateReport) conflictError() error {
	if conflicts := r.Conflicts(); len(conflicts) > 0 {
		first := conflicts[0]
		where := fmt.Sprintf("profile %s, environment %s", first.Profile, first.EnvID)
		if first.Path != "" {
			where += ", " + first.Path
		}
		extra := ""
		if len(conflicts) > 1 {
			extra = fmt.Sprintf(" (+%d more; see plan)", len(conflicts)-1)
		}
		return fmt.Errorf("%s: migration blocked by %d conflicts: %s: %s. Operator choice: %s%s",
			envregistry.DiagCredentialConflict, len(conflicts), where, first.Detail, first.Choice, extra)
	}
	return nil
}

// PlanMigration inventories the scope and derives the exact operations,
// conflicts, and plan hash. Read-only: it takes no lock and never reads
// credential bytes (lstat/readlink/stat only). It serves both the
// inspect and the plan renderings; use MigrateReport.Text to print.
func PlanMigration(req MigrateRequest) (*MigrateReport, error) {
	// Read-only planning still reaches backend observation. Validate every
	// authority/lifecycle/recovery owner first, including disjoint homes; selected
	// scope and a standing journal never waive the common admission boundary.
	if err := envauthority.Validate(req.Home); err != nil {
		return nil, err
	}
	profiles, err := migrationProfiles(req.Home, req.Profile)
	if err != nil {
		return nil, err
	}
	adapters, err := migrationAdapters(req.EnvID)
	if err != nil {
		return nil, err
	}
	report := &MigrateReport{Home: req.Home, ScopeProf: req.Profile, ScopeEnv: req.EnvID}
	for _, profile := range profiles {
		for _, adapter := range adapters {
			report.Homes = append(report.Homes, inventoryHome(&req, profile, adapter))
		}
	}
	report.identityInputs, err = identityMigrationInputs(&req, profiles)
	if err != nil {
		return nil, err
	}
	report.Recovery = readRecoveryBestEffort(req.Home)
	report.Hash = migrationHash(report)
	return report, nil
}

// readRecoveryBestEffort surfaces a leftover migration journal for the
// read-only phases: the summary, a broken marker, or nil when no
// journal is present. It never mutates.
func readRecoveryBestEffort(home string) *MigrateRecovery {
	journal, err := readMigrationJournal(home)
	if err != nil {
		return &MigrateRecovery{Broken: err.Error()}
	}
	if journal == nil {
		return nil
	}
	done := 0
	for _, op := range journal.Ops {
		if op.Done {
			done++
		}
	}
	return &MigrateRecovery{Plan: journal.Plan, Ops: len(journal.Ops), Done: done}
}

// migrationProfiles resolves the profile scope without touching the
// manager lock and without creating the default profile: inspect and
// plan are read-only, so unlike List they never ensure anything.
func migrationProfiles(home, explicit string) ([]string, error) {
	if explicit != "" {
		installed, err := installedProfiles(home)
		if err != nil {
			return nil, err
		}
		for _, name := range installed {
			if name == explicit {
				return []string{explicit}, nil
			}
		}
		return nil, fmt.Errorf("%s: profile %q is not installed", DiagProfileUnknown, explicit)
	}
	return installedProfiles(home)
}

// installedProfiles enumerates installed profiles lock-free: a directory
// with a valid name carrying source.json. Reserved state (scoped/) and
// anything else in the store directory is not a profile.
func installedProfiles(home string) ([]string, error) {
	profilesDir := ProfilesDir(home)
	state, err := stateread.ReadDir(profilesDir)
	if err != nil {
		return nil, err
	}
	if state.Kind == stateread.KindAbsent {
		return nil, nil
	}
	if state.Kind != stateread.KindPresent {
		return nil, stateread.UnusableError(profilesDir, fmt.Errorf("unknown directory state %q", state.Kind))
	}
	var out []string
	for _, entry := range state.Entries {
		if !entry.IsDir() || !validProfileName(entry.Name()) {
			continue
		}
		source := sourcePath(home, entry.Name())
		metadata, err := stateread.Lstat(source)
		if err != nil {
			return nil, err
		}
		if metadata.Kind == stateread.KindAbsent {
			continue
		}
		if metadata.Kind != stateread.KindPresent || metadata.Info == nil {
			return nil, stateread.UnusableError(source, fmt.Errorf("unknown source metadata state %q", metadata.Kind))
		}
		out = append(out, entry.Name())
	}
	sort.Strings(out)
	return out, nil
}

// migrationAdapters resolves the environment scope in closed-registry
// order, so plans are deterministic across machines.
func migrationAdapters(explicit string) ([]envregistry.Adapter, error) {
	if explicit != "" {
		adapter, err := envregistry.ByID(explicit)
		if err != nil {
			return nil, err
		}
		return []envregistry.Adapter{adapter}, nil
	}
	return append([]envregistry.Adapter{}, envregistry.Registry...), nil
}

// inventoryHome inventories one profile × environment. It never fails:
// anything that prevents planning the home — an invalid marker, an
// unresolvable native home, a refused configuration, an unreadable
// native selector — becomes a conflict naming the out-of-band fix, so
// apply blocks instead of guessing.
func inventoryHome(req *MigrateRequest, profile string, adapter envregistry.Adapter) MigrateHome {
	home := MigrateHome{
		Profile: profile, EnvID: adapter.ID,
		HomeDir: ManagedHomeDir(req.Home, profile, adapter.ID),
		PiOld:   MigrateRootNA, PiNew: MigrateRootNA,
	}
	if err := validateMarkerBoundary(req.Home, profile, adapter.ID, nil); err != nil {
		home.addConflict("", err.Error(), "restore the owned no-follow marker boundary before migration")
		return home
	}
	markerPath := filepath.Join(home.HomeDir, envmarker.Name)
	markerFile, err := stateread.ReadFile(markerPath)
	if err != nil {
		home.addConflict("", fmt.Sprintf("marker is unreadable: %v", err),
			"restore access to the marker out of band (back it up first) and re-run")
		return home
	}
	if markerFile.Kind == stateread.KindAbsent {
		return home
	}
	if markerFile.Kind != stateread.KindPresent {
		home.addConflict("", fmt.Sprintf("marker is unreadable: %v", stateread.UnusableError(markerPath, fmt.Errorf("unknown file state %q", markerFile.Kind))),
			"restore access to the marker out of band (back it up first) and re-run")
		return home
	}
	payload := markerFile.Bytes
	marker, err := envmarker.Parse(payload)
	if err != nil {
		home.addConflict("", fmt.Sprintf("marker is invalid: %v", err),
			"back the marker up, remove it, and re-run `curator env resolve --repair` to re-provision")
		return home
	}
	if err := requirePerHomeCredentialSource(marker); err != nil {
		home.addConflict("", err.Error(), "use a manager supporting the recorded node credential source")
		return home
	}
	home.provisioned = true
	home.marker = marker
	home.markerRaw = payload
	home.Mode = marker.Mode
	rr := req.resolveFor(profile, adapter.ID)
	if err := rr.observeOrigin(adapter.ID, marker); err != nil {
		home.addConflict("", err.Error(), "verify the original provisioning history; use explicit env adopt only with trusted attestation for a legacy home")
		return home
	}
	policy, err := rr.resolveCredentialPolicy(adapter)
	if err != nil {
		home.addConflict("", fmt.Sprintf("cannot establish the credential strategy: %v", err),
			"fix the native credential configuration out of band or correct the machine credential store/isolation setting, then re-run")
		return home
	}
	isolation := policy.isolation
	home.Isolation = isolation
	links, strategies, err := rr.effectivePassthrough(adapter)
	if err != nil {
		home.addConflict("", fmt.Sprintf("cannot establish the credential strategy: %v", err),
			"fix the native credential configuration out of band and re-run")
		return home
	}
	credentials, err := rr.credentialRecords(adapter, "migrated")
	if err != nil {
		home.addConflict("", fmt.Sprintf("cannot establish the credential record: %v", err),
			"fix the credential configuration out of band and re-run")
		return home
	}
	home.credentialRecords = credentials
	native, err := rr.credentialNativeRoot(adapter.ID)
	if err != nil {
		home.addConflict("", fmt.Sprintf("cannot resolve the native home: %v", err),
			"restore the native home out of band and re-run")
		return home
	}
	recorded := map[string]string{}
	if marker.Passthrough != nil {
		for _, entry := range *marker.Passthrough {
			if entry.Path != "" {
				recorded[entry.Path] = entry.Strategy
			}
		}
	}
	paths := map[string]bool{}
	for path := range recorded {
		paths[path] = true
	}
	for path := range links {
		paths[path] = true
	}
	for _, path := range sortedKeysBool(paths) {
		_, isRecorded := recorded[path]
		wantedTarget, isWanted := links[path]
		home.classifyPath(adapter, native, path, isRecorded, isWanted, wantedTarget, strategies[path])
	}
	if adapter.ID == envregistry.Pi {
		home.inventoryPiRoots(native)
	}
	return home
}

func sortedKeysBool(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (h *MigrateHome) addConflict(path, detail, choice string) {
	h.Conflicts = append(h.Conflicts, MigrateConflict{
		Profile: h.Profile, EnvID: h.EnvID, Path: path, Detail: detail, Choice: choice,
	})
}

// classifyPath classifies one recorded-or-wanted path by inspection only
// (lstat/readlink/stat — never a credential-byte read) and derives at
// most one operation. Anything the migration must not touch — a regular
// file, a foreign link, a non-empty directory, an uninspectable path —
// becomes a conflict naming the operator choice instead of an operation.
func (h *MigrateHome) classifyPath(adapter envregistry.Adapter, native, path string, recorded, wanted bool, wantedTarget, strategy string) {
	if strategy == "" {
		strategy = "-"
	}
	entry := MigrateEntry{Path: path, Strategy: strategy, Recorded: recorded, Wanted: wanted, Current: "-", Declared: "-"}
	if wanted {
		entry.Declared = wantedTarget
	}
	if !recorded && strategy == "" {
		entry.Strategy = "-"
	}
	full := filepath.Join(h.HomeDir, filepath.FromSlash(path))
	metadata, err := stateread.Lstat(full)
	if err != nil {
		entry.State = MigrateUninspectable
		h.Entries = append(h.Entries, entry)
		h.addConflict(path, fmt.Sprintf("%s cannot be inspected: %v", full, err),
			"restore access to the link path out of band and re-run")
		return
	}
	if metadata.Kind == stateread.KindAbsent {
		if wanted {
			entry.State = MigrateAbsent
			entry.Note = "repair re-links the absent path"
			h.Entries = append(h.Entries, entry)
			return
		}
		// A stale record whose link is already gone needs only the
		// record drop (crash-recovery and hand-removed shapes converge
		// here); there is no link to remove.
		entry.State = MigrateStaleRecorded
		h.Entries = append(h.Entries, entry)
		h.Ops = append(h.Ops, MigrateOp{Profile: h.Profile, EnvID: h.EnvID, Kind: MigrateOpUnlink, Path: path})
		return
	}
	if metadata.Kind != stateread.KindPresent || metadata.Info == nil {
		entry.State = MigrateUninspectable
		h.Entries = append(h.Entries, entry)
		h.addConflict(path, fmt.Sprintf("%s cannot be inspected: %v", full, stateread.UnusableError(full, fmt.Errorf("unknown metadata state %q", metadata.Kind))),
			"restore access to the link path out of band and re-run")
		return
	}
	info := metadata.Info
	if info.Mode()&os.ModeSymlink != 0 {
		h.classifyLink(adapter, native, full, &entry, wantedTarget, recorded, wanted)
		return
	}
	if info.IsDir() {
		entries, err := os.ReadDir(full)
		if err != nil {
			entry.State = MigrateUninspectable
			h.Entries = append(h.Entries, entry)
			h.addConflict(path, fmt.Sprintf("%s holds a directory that cannot be listed: %v", full, err),
				"restore access to the link path out of band and re-run")
			return
		}
		if len(entries) == 0 && wanted {
			entry.State = MigrateEmptyDir
			entry.Note = "repair replaces the empty directory with the declared link"
			h.Entries = append(h.Entries, entry)
			return
		}
		entry.State = MigrateDirConflict
		h.Entries = append(h.Entries, entry)
		shape := "a non-empty directory"
		if len(entries) == 0 {
			shape = "an empty directory where no link is wanted"
		}
		h.addConflict(path, fmt.Sprintf("%s holds %s", full, shape),
			"clear the directory out of band and re-run")
		return
	}
	entry.State = MigrateRegularFile
	h.Entries = append(h.Entries, entry)
	nativeSide := h.staleDeclaredTarget(adapter, native, path)
	if wanted {
		nativeSide = wantedTarget
	}
	if nativeSide == "" {
		nativeSide = "the declared native store"
	}
	h.addConflict(path,
		fmt.Sprintf("%s holds a regular file (the managed home holds its own credential bytes — e.g. after running isolated, or the tool severed the link), expected a link to %s", full, nativeSide),
		fmt.Sprintf("decide which credential bytes win (managed %s vs native %s); move the loser aside out of band (the manager never moves credential bytes) and re-run", full, nativeSide))
}

// classifyLink classifies a symlink at a recorded-or-wanted path: a
// recorded link to the wrong target is re-pointed, a stale recorded
// link to the declared store is unlinked, and anything unprovable is a
// conflict. A correctly targeted link is current, or pending when the
// native target does not exist yet — pending entries get no operation.
func (h *MigrateHome) classifyLink(adapter envregistry.Adapter, native, full string, entry *MigrateEntry, wantedTarget string, recorded, wanted bool) {
	got, err := os.Readlink(full)
	if err != nil {
		entry.State = MigrateUninspectable
		h.Entries = append(h.Entries, *entry)
		h.addConflict(entry.Path, fmt.Sprintf("%s link target cannot be read", full),
			"restore access to the link out of band and re-run")
		return
	}
	entry.Current = got
	if wanted {
		if got == wantedTarget {
			metadata, err := stateread.Stat(wantedTarget)
			if err != nil {
				entry.State = MigrateUninspectable
				h.Entries = append(h.Entries, *entry)
				h.addConflict(entry.Path, fmt.Sprintf("%s target %s cannot be inspected: %v", full, wantedTarget, err),
					"restore access to the native target out of band and re-run")
				return
			}
			if metadata.Kind == stateread.KindAbsent {
				entry.State = MigratePending
				entry.Note = "dangling-to-declared: nothing to migrate; log in to populate the native target"
				h.Entries = append(h.Entries, *entry)
				return
			}
			if metadata.Kind != stateread.KindPresent || metadata.Info == nil {
				entry.State = MigrateUninspectable
				h.Entries = append(h.Entries, *entry)
				h.addConflict(entry.Path, fmt.Sprintf("%s target %s cannot be inspected: %v", full, wantedTarget, stateread.UnusableError(wantedTarget, fmt.Errorf("unknown metadata state %q", metadata.Kind))),
					"restore access to the native target out of band and re-run")
				return
			}
			entry.State = MigrateCurrent
			if !recorded {
				entry.Note = "unrecorded; repair adopts the record"
			}
			h.Entries = append(h.Entries, *entry)
			return
		}
		if !recorded {
			entry.State = MigrateForeign
			h.Entries = append(h.Entries, *entry)
			h.addConflict(entry.Path, fmt.Sprintf("%s links to %s, expected %s", full, got, wantedTarget),
				"remove the unrecorded link out of band and re-run (repair then links the declared store)")
			return
		}
		entry.State = MigrateMistargeted
		h.Entries = append(h.Entries, *entry)
		h.Ops = append(h.Ops, MigrateOp{Profile: h.Profile, EnvID: h.EnvID, Kind: MigrateOpRelink, Path: entry.Path, From: got, To: wantedTarget})
		return
	}
	declared, ok := declaredPassthroughTarget(adapter, native, entry.Path)
	if !ok {
		entry.State = MigrateNoDeclaredRoot
		h.Entries = append(h.Entries, *entry)
		h.addConflict(entry.Path, fmt.Sprintf("%s has no declared native store for the recorded credential link %s", full, entry.Path),
			"remove the link out of band and re-run")
		return
	}
	if got != declared {
		entry.State = MigrateForeign
		h.Entries = append(h.Entries, *entry)
		h.addConflict(entry.Path, fmt.Sprintf("%s links to %s, not the recorded credential target %s", full, got, declared),
			"remove the link out of band and re-run")
		return
	}
	entry.State = MigrateStaleRecorded
	h.Entries = append(h.Entries, *entry)
	h.Ops = append(h.Ops, MigrateOp{Profile: h.Profile, EnvID: h.EnvID, Kind: MigrateOpUnlink, Path: entry.Path, From: got})
}

// staleDeclaredTarget reconstructs the declared target of a stale
// recorded path for conflict wording, or "" where none is declared.
func (h *MigrateHome) staleDeclaredTarget(adapter envregistry.Adapter, native, path string) string {
	declared, ok := declaredPassthroughTarget(adapter, native, path)
	if !ok {
		return ""
	}
	return declared
}

// inventoryPiRoots stats both Pi native roots without reading bytes. Two
// live roots block a re-point (the operator chooses the account); bytes
// at the old root alone warn (the relink orphans them and the manager
// never touches native files). Roots are informational where no
// re-point is planned.
func (h *MigrateHome) inventoryPiRoots(native string) {
	h.PiOldPath = filepath.Join(filepath.Dir(native), "auth.json")
	h.PiNewPath = filepath.Join(native, "auth.json")
	h.PiOld = statRoot(h.PiOldPath)
	h.PiNew = statRoot(h.PiNewPath)
	relink := false
	for _, op := range h.Ops {
		if op.Kind == MigrateOpRelink {
			relink = true
		}
	}
	if relink && (h.PiOld == MigrateRootUninspectable || h.PiNew == MigrateRootUninspectable) {
		h.addConflict("", fmt.Sprintf("Pi native roots cannot be inspected (old %s: %s; new %s: %s)", h.PiOldPath, h.PiOld, h.PiNewPath, h.PiNew),
			"restore access to the native roots out of band and re-run")
		return
	}
	if relink && h.PiOld == MigrateRootPresent && h.PiNew == MigrateRootPresent {
		h.addConflict("", fmt.Sprintf("two live Pi credentials: %s and %s both hold bytes", h.PiOldPath, h.PiNewPath),
			fmt.Sprintf("decide which of %s and %s holds the live credential; reconcile out of band (the manager never copies, moves, or deletes credential bytes) and re-run", h.PiOldPath, h.PiNewPath))
		return
	}
	if h.PiOld == MigrateRootPresent {
		h.Warnings = append(h.Warnings, fmt.Sprintf("native bytes exist at the old Pi root %s (declared root %s): the relink leaves them untouched — if they are live, move them out of band (the manager never copies credential bytes)", h.PiOldPath, h.PiNewPath))
	}
}

// statRoot stats one native path: present, absent, or uninspectable.
// Absence and a failed read are different facts; only NotExist is
// absence.
func statRoot(path string) string {
	metadata, err := stateread.Stat(path)
	if err != nil {
		return MigrateRootUninspectable
	}
	switch metadata.Kind {
	case stateread.KindAbsent:
		return MigrateRootAbsent
	case stateread.KindPresent:
		if metadata.Info != nil {
			return MigrateRootPresent
		}
	}
	return MigrateRootUninspectable
}

// migrationHash hashes the canonical plan bytes: every inventoried home
// (marker identity as a raw-bytes digest — no credential reads — mode,
// entries with targets, Pi roots), every operation, every conflict with
// its choice, and any leftover recovery state. Any inventory change
// re-hashes, so --expect refuses on plan drift.
func migrationHash(report *MigrateReport) string {
	var canonical strings.Builder
	canonical.WriteString("migration-plan-v3\n")
	fmt.Fprintf(&canonical, "writer %d identity-inputs %s\n", hashing.WriteVersion(), identityInputHash(report.identityInputs))
	for _, home := range report.Homes {
		provisioned := "0"
		if home.provisioned {
			provisioned = "1"
		}
		fmt.Fprintf(&canonical, "home %s %s provisioned=%s isolation=%s mode=%s\n", home.Profile, home.EnvID, provisioned, home.Isolation, home.Mode)
		if home.provisioned {
			sum := sha256.Sum256(home.markerRaw)
			fmt.Fprintf(&canonical, "marker %s %s sha256=%x\n", home.Profile, home.EnvID, sum)
		}
		for _, entry := range home.Entries {
			fmt.Fprintf(&canonical, "entry %s %s %s recorded=%t wanted=%t strategy=%s state=%s current=%s declared=%s\n",
				home.Profile, home.EnvID, entry.Path, entry.Recorded, entry.Wanted, entry.Strategy, entry.State, entry.Current, entry.Declared)
		}
		if home.EnvID == envregistry.Pi && home.provisioned {
			fmt.Fprintf(&canonical, "piroots %s old=%s new=%s\n", home.Profile, home.PiOld, home.PiNew)
		}
		for _, op := range home.Ops {
			fmt.Fprintf(&canonical, "op %s %s %s %s from=%s to=%s\n", op.Profile, op.EnvID, op.Kind, op.Path, op.From, op.To)
		}
		for _, conflict := range home.Conflicts {
			fmt.Fprintf(&canonical, "conflict %s %s %s detail=%s choice=%s\n", conflict.Profile, conflict.EnvID, conflict.Path, conflict.Detail, conflict.Choice)
		}
		for _, warning := range home.Warnings {
			fmt.Fprintf(&canonical, "warning %s %s %s\n", home.Profile, home.EnvID, warning)
		}
		if len(home.credentialRecords) > 0 {
			payload, _ := json.Marshal(home.credentialRecords)
			fmt.Fprintf(&canonical, "credentials %s %s %s\n", home.Profile, home.EnvID, payload)
		}
	}
	if report.Recovery != nil {
		if report.Recovery.Broken != "" {
			fmt.Fprintf(&canonical, "recovery broken=%s\n", report.Recovery.Broken)
		} else {
			fmt.Fprintf(&canonical, "recovery plan=%s ops=%d done=%d\n",
				report.Recovery.Plan, report.Recovery.Ops, report.Recovery.Done)
		}
	}
	sum := sha256.Sum256([]byte(canonical.String()))
	return hex.EncodeToString(sum[:])
}

// Text renders the report deterministically for one phase: the inspect
// inventory detail, or the plan with operations, conflicts, warnings,
// and hash. A leftover recovery journal banners first in both phases:
// the listed inventory is live pre-recovery state, and the plan footer
// points at recovery instead of the printed hash (recovery runs before
// the drift check, so the pre-recovery hash never applies). Apply
// prints the plan text before the first mutation.
func (r *MigrateReport) Text(phase string) string {
	var out strings.Builder
	if r.Recovery != nil {
		out.WriteString(r.Recovery.notice(r.Home))
	}
	if phase == MigratePhaseInspect {
		fmt.Fprintf(&out, "credential migration inventory: manager home %s\n", r.Home)
		for _, home := range r.Homes {
			home.writeInventory(&out)
		}
	} else {
		fmt.Fprintf(&out, "credential migration plan %s: manager home %s\n", r.Hash, r.Home)
		shown := false
		for _, home := range r.Homes {
			if len(home.Ops) == 0 && len(home.Conflicts) == 0 && len(home.Warnings) == 0 {
				continue
			}
			shown = true
			fmt.Fprintf(&out, "profile %s, environment %s:\n", home.Profile, home.EnvID)
			for _, op := range home.Ops {
				op.writePlan(&out)
			}
			for _, warning := range home.Warnings {
				fmt.Fprintf(&out, "  warning: %s\n", warning)
			}
			for _, conflict := range home.Conflicts {
				conflict.writePlan(&out)
			}
		}
		if !shown {
			out.WriteString("no operations, conflicts, or warnings\n")
		}
	}
	if len(r.identityInputs) > 0 {
		fmt.Fprintf(&out, "atomic v1-to-v2 profile identity migration: %d bound inputs; all provisioned environments of each selected legacy profile participate\n", len(r.identityInputs))
	}
	ops := len(r.Ops())
	conflicts := len(r.Conflicts())
	migrating := 0
	for _, home := range r.Homes {
		if len(home.Ops) > 0 || len(home.Conflicts) > 0 {
			migrating++
		}
	}
	fmt.Fprintf(&out, "summary: %d homes inventoried, %d need migration, %d operations, %d conflicts\n",
		len(r.Homes), migrating, ops, conflicts)
	if phase != MigratePhaseInspect {
		switch {
		case r.Recovery != nil && r.Recovery.Broken != "":
			fmt.Fprintf(&out, "blocked: the migration journal is unusable; %s\n", r.Recovery.Broken)
		case r.Recovery != nil:
			fmt.Fprintf(&out, "recover: run `curator env migrate --apply --expect %s%s` to recover the prior state and execute the plan\n", r.Recovery.Plan, r.scopeFlags())
		case conflicts > 0:
			fmt.Fprintf(&out, "blocked: resolve the conflicts out of band, then re-run `curator env migrate --plan%s`\n", r.scopeFlags())
		case ops > 0:
			fmt.Fprintf(&out, "ready: run `curator env migrate --apply --expect %s%s`\n", r.Hash, r.scopeFlags())
		default:
			out.WriteString("nothing to migrate\n")
		}
	}
	return out.String()
}

// scopeFlags renders the scope back as CLI flags for re-runnable
// commands, or "" for the full scope.
func (r *MigrateReport) scopeFlags() string {
	var flags strings.Builder
	if r.ScopeProf != "" {
		fmt.Fprintf(&flags, " --profile %s", r.ScopeProf)
	}
	if r.ScopeEnv != "" {
		fmt.Fprintf(&flags, " --env %s", r.ScopeEnv)
	}
	return flags.String()
}

func (h *MigrateHome) writeInventory(out *strings.Builder) {
	if !h.provisioned {
		fmt.Fprintf(out, "profile %s, environment %s: not provisioned, nothing to migrate\n", h.Profile, h.EnvID)
		return
	}
	fmt.Fprintf(out, "profile %s, environment %s: provisioned, effective isolation %s, marker mode %s\n",
		h.Profile, h.EnvID, h.Isolation, h.Mode)
	for _, entry := range h.Entries {
		recorded := "unrecorded"
		if entry.Recorded {
			recorded = "recorded"
		}
		wanted := "unwanted"
		if entry.Wanted {
			wanted = "wanted"
		}
		fmt.Fprintf(out, "  %s (%s, %s, %s): %s", entry.Path, recorded, wanted, entry.Strategy, entry.State)
		switch entry.State {
		case MigrateCurrent:
			fmt.Fprintf(out, ": link targets the declared %s", entry.Declared)
		case MigratePending:
			fmt.Fprintf(out, ": link targets the declared %s, which does not exist yet", entry.Declared)
		case MigrateMistargeted:
			fmt.Fprintf(out, ": link targets %s, declared %s", entry.Current, entry.Declared)
		case MigrateStaleRecorded:
			if entry.Current == "-" {
				fmt.Fprintf(out, ": recorded link already gone, record still present")
			} else {
				fmt.Fprintf(out, ": recorded link targets %s, no longer wanted", entry.Current)
			}
		case MigrateRegularFile, MigrateForeign, MigrateDirConflict, MigrateUninspectable, MigrateNoDeclaredRoot:
			fmt.Fprintf(out, ": see conflict below")
		case MigrateAbsent, MigrateEmptyDir:
			fmt.Fprintf(out, ": %s", entry.Note)
		}
		if entry.Note != "" && (entry.State == MigrateCurrent) {
			fmt.Fprintf(out, " (%s)", entry.Note)
		}
		out.WriteString("\n")
	}
	if h.EnvID == envregistry.Pi {
		fmt.Fprintf(out, "  native Pi roots: %s %s; %s %s\n", h.PiOldPath, h.PiOld, h.PiNewPath, h.PiNew)
	}
	for _, warning := range h.Warnings {
		fmt.Fprintf(out, "  warning: %s\n", warning)
	}
	for _, conflict := range h.Conflicts {
		conflict.writePlan(out)
	}
}

func (op MigrateOp) writePlan(out *strings.Builder) {
	switch op.Kind {
	case MigrateOpRelink:
		fmt.Fprintf(out, "  relink %s: %s -> %s\n", op.Path, op.From, op.To)
	case MigrateOpUnlink:
		if op.From == "" {
			fmt.Fprintf(out, "  unlink %s: link already gone, drop the record\n", op.Path)
		} else {
			fmt.Fprintf(out, "  unlink %s (%s)\n", op.Path, op.From)
		}
	}
}

func (c MigrateConflict) writePlan(out *strings.Builder) {
	where := fmt.Sprintf("profile %s, environment %s", c.Profile, c.EnvID)
	if c.Path != "" {
		where += ", " + c.Path
	}
	fmt.Fprintf(out, "  conflict %s: %s. Operator choice: %s\n", where, c.Detail, c.Choice)
}

// MigrateApplyResult carries the recomputed plan apply printed before
// executing, or a read-only refusal plan when authority admission fails. It
// records operations applied and any interrupted apply recovered first (nil
// when no journal was present).
type MigrateApplyResult struct {
	Report    *MigrateReport
	Applied   []MigrateOp
	Recovered *MigrateRecovery
}

// ApplyMigration requires a current plan hash and holds the manager mutation
// lock. Common phase-aware recovery precedes revalidation. Drift or conflict
// refuses before new publication; actual link/marker/origin changes publish as
// one complete unit with durable preimages and protection metadata.
func ApplyMigration(req MigrateRequest) (*MigrateApplyResult, error) {
	op, err := beginOperation(req.Home, req.transactionOptions...)
	if err != nil {
		// Admission must not recover an unreadable authority marker, but the
		// read-only inventory still explains the existing migration conflict.
		if strings.HasPrefix(err.Error(), envmarker.DiagMarkerUnreadable+":") || strings.HasPrefix(err.Error(), envmarker.DiagMarkerInvalid+":") {
			if report, planErr := PlanMigration(req); planErr == nil && len(report.Conflicts()) != 0 {
				result := &MigrateApplyResult{Report: report}
				if printErr := printMigrationPlan(req.Print, report); printErr != nil {
					return result, fmt.Errorf("migration authority refusal plan cannot be printed: %w; %w", printErr, err)
				}
				return result, fmt.Errorf("%w; %w", report.conflictError(), err)
			}
		}
		return nil, err
	}
	defer func() { _ = op.close() }()
	if strings.TrimSpace(req.Expect) == "" {
		report, err := PlanMigration(req)
		if err != nil {
			return nil, err
		}
		result := &MigrateApplyResult{Report: report}
		if err := printMigrationPlan(req.Print, report); err != nil {
			return result, fmt.Errorf("migration apply needs a prior plan but cannot print it: %v; re-run `curator env migrate --plan%s`", err, report.scopeFlags())
		}
		return result, fmt.Errorf("%s: migration apply needs a prior plan: run `curator env migrate --plan%s`, then apply with `--expect %s`",
			envregistry.DiagCredentialConflict, report.scopeFlags(), report.Hash)
	}
	recovered, err := recoverMigrationJournal(req.Home, op, req.Print)
	if err != nil {
		report, planErr := PlanMigration(req)
		if planErr != nil {
			return nil, err
		}
		return &MigrateApplyResult{Report: report, Recovered: recovered}, err
	}
	report, err := PlanMigration(req)
	if err != nil {
		return nil, err
	}
	result := &MigrateApplyResult{Report: report, Recovered: recovered}
	if err := printMigrationPlan(req.Print, report); err != nil {
		return result, fmt.Errorf("migration apply cannot print the locked plan, refusing before any mutation: %v", err)
	}
	if !strings.EqualFold(report.Hash, strings.TrimSpace(req.Expect)) {
		return result, fmt.Errorf("%s: migration plan drift: expected plan %s, the inventory now hashes %s; re-run `curator env migrate --plan%s` and apply with the new hash",
			envregistry.DiagCredentialConflict, strings.TrimSpace(req.Expect), report.Hash, report.scopeFlags())
	}
	if err := report.conflictError(); err != nil {
		return result, err
	}
	ops := report.Ops()
	if len(report.identityInputs) > 0 {
		profiles, err := migrationProfiles(req.Home, req.Profile)
		if err != nil {
			return result, err
		}
		rr := req.resolveFor(req.Profile, req.EnvID)
		migration, err := prepareIdentityMigration(op, rr, profiles, report)
		if err != nil {
			return result, err
		}
		defer migration.cleanup()
		inputs, err := identityMigrationInputs(&req, profiles)
		if err != nil {
			return result, err
		}
		if identityInputHash(inputs) != identityInputHash(report.identityInputs) {
			return result, fmt.Errorf("migration plan drift during identity staging; re-run --plan")
		}
		if err := migration.publish(op); err != nil {
			return result, err
		}
		result.Applied = ops
		return result, nil
	}
	if len(ops) == 0 {
		return result, nil
	}
	if err := publishCredentialMigration(op, req, report); err != nil {
		return result, err
	}
	result.Applied = ops
	return result, nil
}

// printMigrationPlan writes the locked, revalidated plan text before the
// first mutation; a nil writer prints nothing and the caller uses the
// returned report instead.
func printMigrationPlan(w io.Writer, report *MigrateReport) error {
	if w == nil {
		return nil
	}
	_, err := io.WriteString(w, report.Text(MigratePhasePlan))
	return err
}

// validateMigrationOp re-checks one operation's preconditions at
// execution time: the link must still be exactly as inventoried. The
// lock rules out legitimate races, so a mismatch fails closed.
func validateMigrationOp(full string, op MigrateOp) error {
	metadata, err := stateread.Lstat(full)
	if err != nil {
		return fmt.Errorf("%s %s: link path changed since the plan: %v", op.Kind, op.Path, err)
	}
	if metadata.Kind == stateread.KindAbsent && op.Kind == MigrateOpUnlink && op.From == "" {
		return nil
	}
	if metadata.Kind != stateread.KindPresent || metadata.Info == nil {
		return fmt.Errorf("%s %s: link path changed since the plan: %v", op.Kind, op.Path, stateread.UnusableError(full, fmt.Errorf("unknown metadata state %q", metadata.Kind)))
	}
	info := metadata.Info
	if op.Kind == MigrateOpUnlink && op.From == "" {
		return fmt.Errorf("%s %s: link path changed since the plan: expected absent", op.Kind, op.Path)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("%s %s: link path changed since the plan: not a symlink", op.Kind, op.Path)
	}
	got, err := os.Readlink(full)
	if err != nil || got != op.From {
		return fmt.Errorf("%s %s: link path changed since the plan", op.Kind, op.Path)
	}
	return nil
}

// changedMarkers renders marker updates for every schema-2 home with a
// migration operation so provenance records the migration, plus schema-1
// homes whose unlink operation requires changing the legacy marker. A
// schema-1 relink keeps its exact bytes because no marker replacement is
// otherwise required. Prior raw bytes are returned for rollback.
func changedMarkers(report *MigrateReport) (map[string][]byte, map[string][]byte, error) {
	changed := map[string][]byte{}
	priors := map[string][]byte{}
	for _, home := range report.Homes {
		var dropped []string
		for _, op := range home.Ops {
			if op.Kind == MigrateOpUnlink {
				dropped = append(dropped, op.Path)
			}
		}
		if len(home.Ops) == 0 || home.marker == nil {
			continue
		}
		if home.marker.Version == envmarker.VersionV1 && len(dropped) == 0 {
			continue
		}
		_, lock, _, err := loadResolveInputs(report.Home, home.Profile)
		if err != nil {
			return nil, nil, err
		}
		if markerHashVersion(home.marker) != lock.ContentHashVersion() {
			return nil, nil, fmt.Errorf("marker/lock hash version mismatch for %s/%s", home.Profile, home.EnvID)
		}
		updated := *home.marker
		if markerHashVersion(home.marker) == hashing.VersionV2 {
			if updated.Version != envmarker.VersionV7 {
				updated.Version = envmarker.VersionV3
			}
			updated.HashVersion = int(hashing.VersionV2)
		} else {
			updated.Version = envmarker.VersionV2
			updated.HashVersion = 0
		}
		complete := append([]envmarker.Passthrough{}, home.credentialRecords...)
		updated.Passthrough = &complete
		payload, err := updated.Marshal()
		if err != nil {
			return nil, nil, fmt.Errorf("render the migrated marker for %s/%s: %v", home.Profile, home.EnvID, err)
		}
		live := filepath.Join(home.HomeDir, envmarker.Name)
		changed[live] = payload
		priors[live] = home.markerRaw
	}
	return changed, priors, nil
}

const migrationJournalVersion = 1

type migrationJournalOp = envauthority.LegacyMigrationOp
type migrationJournalMarker = envauthority.LegacyMigrationMarker
type migrationJournal = envauthority.LegacyMigration

// migrationJournalPath is the fixed journal below manager state (never
// below the credential scope, and holding no credential bytes).
func migrationJournalPath(home string) string {
	return filepath.Join(home, "state", "env-migration.json")
}

// Legacy journals cannot authorize new forward writes or recreate lost origin.
// Only a closed, completed journal with verified stable postimages is eligible
// for cleanup. Unsealed records lack durable parent/protection and origin
// preimages and are preserved for explicit recovery, across every owner.
func readMigrationJournal(home string) (*migrationJournal, error) {
	return envauthority.ReadCompletedLegacyMigration(home)
}

// removeMigrationJournal deletes the journal after a completed apply, a
// clean rollback, or a completed recovery.
func removeMigrationJournal(home string) error {
	if err := os.Remove(migrationJournalPath(home)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// recoverMigrationJournal cleans only a validated completed legacy receipt.
// Common admission has already refused unsealed or substituted records. New
// migration writes and recovery use the complete environment transaction unit.
func recoverMigrationJournal(home string, op *operation, announce io.Writer) (*MigrateRecovery, error) {
	if err := op.lock.AssertHeld(); err != nil {
		return nil, err
	}
	if err := envauthority.Validate(home); err != nil {
		return nil, err
	}
	journal, err := readMigrationJournal(home)
	if err != nil {
		return nil, fmt.Errorf("migration journal is unusable: %w", err)
	}
	if journal == nil {
		return nil, nil
	}
	if !journal.Complete {
		return nil, fmt.Errorf("unsealed legacy migration journal cannot authorize recovery")
	}
	recovered := &MigrateRecovery{Plan: journal.Plan, Ops: len(journal.Ops), Done: len(journal.Ops)}
	if announce != nil {
		if _, err := fmt.Fprintf(announce, "cleaning completed legacy migration receipt: plan %s, %d completed link operations; preserving verified postimages\n", journal.Plan, len(journal.Ops)); err != nil {
			return recovered, fmt.Errorf("migration cleanup cannot announce itself, refusing before mutation: %w", err)
		}
	}
	if err := removeMigrationJournal(home); err != nil {
		return recovered, fmt.Errorf("completed migration receipt cleanup remains pending: %w", err)
	}
	return recovered, nil
}
