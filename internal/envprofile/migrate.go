// Package envprofile credential migration runs the explicit inspect →
// plan → apply step (environments §7.4, §10.1; manager §12.4; decision
// 0017 choice 3). A mode change or a native-root correction that leaves
// a recorded credential link behind is
// never applied by resolve --repair: repair re-links an absent link and
// refuses anything else, pointing at this step where the explicit
// migration can fix it. The migration inventories the old marker, every
// recorded link target, and both Pi roots; preserves the effective mode;
// relinks recorded symlinks at the declared native path and unlinks
// stale recorded links; and never copies, moves, or rewrites credential
// bytes — a relink moves the pointer, never the file.
//
// Inspect and plan share one read-only computation (no lock, never a
// credential-byte read) and differ only in what they print; apply takes
// the manager-home mutation lock, requires the --expect plan hash,
// prints the locked, revalidated plan before the first mutation,
// refuses on plan drift and on conflicts, and executes link operations
// through a durable migration journal (temp-link + atomic rename, never
// remove-then-create), publishing changed markers through the journaled
// record publication. A syscall failure mid-op or a failed marker
// publication reverts the links; a leftover journal from an
// interrupted apply is reported by inspect/plan and recovered — rolled
// back to the prior state — by the next apply carrying a plan hash.
package envprofile

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/stateread"
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

// Migration fault points for the MigrateRequest.InjectFault seam:
// after each applied link operation ("link"); after an applied link
// operation simulating a process kill ("crash": the journal stands and
// no rollback runs, so the next apply must recover); before the marker
// publication ("publish-begin": simulates a publication failure, which
// reverts the links); and after the marker publication ("publish").
// Production never sets InjectFault.
const (
	MigrateFaultLink         = "link"
	MigrateFaultCrash        = "crash"
	MigrateFaultPublishBegin = "publish-begin"
	MigrateFaultPublish      = "publish"
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
// MigrateApplyResult.Report instead. InjectFault, symlink, and rename
// are test seams: production never sets them.
type MigrateRequest struct {
	Home         string
	Profile      string
	EnvID        string
	Machine      envregistry.MachineConfig
	Detect       func(envregistry.Adapter) string
	NativeHomeOf func(string) (string, error)
	Expect       string
	InjectFault  func(point string, applied int) error
	Print        io.Writer
	symlink      func(oldname, newpath string) error
	rename       func(oldpath, newpath string) error
}

// symlinkFn resolves the link-creation primitive: os.Symlink unless a
// test injected a syscall-boundary failure.
func (req *MigrateRequest) symlinkFn() func(string, string) error {
	if req.symlink != nil {
		return req.symlink
	}
	return os.Symlink
}

// renameFn resolves the atomic-replacement primitive: os.Rename unless
// a test injected a syscall-boundary failure.
func (req *MigrateRequest) renameFn() func(string, string) error {
	if req.rename != nil {
		return req.rename
	}
	return os.Rename
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
	return fmt.Sprintf("interrupted migration apply detected: plan %s, %d of %d link operations completed; the next `curator env migrate --apply --expect %s` recovers to the prior state under the manager lock before executing (this read-only command recovers nothing)\n",
		rec.Plan, rec.Done, rec.Ops, rec.Plan)
}

// MigrateReport is the migration plan over the requested scope: every
// inventoried home plus the plan hash over the canonical inventory,
// operations, and conflicts.
type MigrateReport struct {
	Home      string
	ScopeProf string
	ScopeEnv  string
	Homes     []MigrateHome
	Recovery  *MigrateRecovery
	Hash      string
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

// PlanMigration inventories the scope and derives the exact operations,
// conflicts, and plan hash. Read-only: it takes no lock and never reads
// credential bytes (lstat/readlink/stat only). It serves both the
// inspect and the plan renderings; use MigrateReport.Text to print.
func PlanMigration(req MigrateRequest) (*MigrateReport, error) {
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
	home.provisioned = true
	home.marker = marker
	home.markerRaw = payload
	home.Mode = marker.Mode
	rr := req.resolveFor(profile, adapter.ID)
	atAbove := true
	if ok, known := adapter.AtOrAbovePinned(rr.detected(adapter)); known {
		atAbove = ok
	}
	isolation, err := req.Machine.EffectiveIsolation(profile, adapter, atAbove)
	if err != nil {
		home.addConflict("", fmt.Sprintf("machine configuration refused: %v", err),
			"fix the isolation setting out of band and re-run")
		return home
	}
	home.Isolation = isolation
	if err := rr.checkIsolatedCredentialStore(adapter, isolation); err != nil {
		home.addConflict("", fmt.Sprintf("machine configuration refused: %v", err),
			"fix the isolation setting out of band and re-run")
		return home
	}
	links, strategies, err := rr.effectivePassthrough(adapter, isolation)
	if err != nil {
		home.addConflict("", fmt.Sprintf("cannot establish the credential strategy: %v", err),
			"fix the native credential configuration out of band and re-run")
		return home
	}
	credentials, err := rr.credentialRecords(adapter, isolation, "migrated")
	if err != nil {
		home.addConflict("", fmt.Sprintf("cannot establish the credential record: %v", err),
			"fix the credential configuration out of band and re-run")
		return home
	}
	home.credentialRecords = credentials
	native, err := rr.nativeHome(adapter.ID)
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
	h.PiOldPath = filepath.Join(native, "auth.json")
	h.PiNewPath = filepath.Join(native, "agent", "auth.json")
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
	canonical.WriteString("migration-plan-v2\n")
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
// executing, plus the operations it applied and any interrupted apply
// it recovered first (nil when no journal was present).
type MigrateApplyResult struct {
	Report    *MigrateReport
	Applied   []MigrateOp
	Recovered *MigrateRecovery
}

// ApplyMigration executes the migration under the manager-home mutation
// lock in strict order: require the --expect plan hash (refusing with
// zero writes when it is missing); validate the entire recovery
// inventory before any recovery write and recover a leftover journal
// back to the prior state, announcing the recovery before mutating;
// re-inventory and print the locked, revalidated plan before the first
// mutation (a print failure fails before anything changes); refuse on
// plan drift and on conflicts with zero writes; then journal the intent
// and execute exactly the printed operations link by link — each relink
// as a journaled temp-link + atomic rename, never remove-then-create —
// publishing changed markers as one journaled transaction. A link-phase
// syscall failure or a failed marker publication reverts the links; a
// post-publish failure restores the prior markers first, then the
// links. Every outcome is the prior state or the migrated state, never
// a mix; a kill between mutations leaves the journal for the next apply
// carrying a plan hash to recover.
func ApplyMigration(req MigrateRequest) (*MigrateApplyResult, error) {
	op, err := beginOperation(req.Home)
	if err != nil {
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
	recovered, err := recoverMigrationJournal(req.Home, op, req.Print, req.symlinkFn(), req.renameFn())
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
	if conflicts := report.Conflicts(); len(conflicts) > 0 {
		first := conflicts[0]
		where := fmt.Sprintf("profile %s, environment %s", first.Profile, first.EnvID)
		if first.Path != "" {
			where += ", " + first.Path
		}
		extra := ""
		if len(conflicts) > 1 {
			extra = fmt.Sprintf(" (+%d more; see plan)", len(conflicts)-1)
		}
		return result, fmt.Errorf("%s: migration blocked by %d conflicts: %s: %s. Operator choice: %s%s",
			envregistry.DiagCredentialConflict, len(conflicts), where, first.Detail, first.Choice, extra)
	}
	ops := report.Ops()
	if len(ops) == 0 {
		return result, nil
	}
	changed, priors, err := changedMarkers(report)
	if err != nil {
		return result, err
	}
	journal := newMigrationJournal(report, changed, priors)
	if err := writeMigrationJournal(req.Home, journal); err != nil {
		return result, fmt.Errorf("migration apply cannot journal the plan, refusing before any mutation: %v", err)
	}
	applied, err := executeMigrationOps(req, ops, journal)
	if err != nil {
		var crash *migrationCrashError
		if errors.As(err, &crash) {
			result.Applied = crash.applied
			return result, fmt.Errorf("migration apply interrupted after %d of %d operations (the journal stands at %s): re-run `curator env migrate --apply --expect %s%s` to recover the prior state and retry",
				len(crash.applied), len(ops), migrationJournalPath(req.Home), report.Hash, report.scopeFlags())
		}
		rollbackErr := reconcileJournalLinks(req.Home, journal, req.symlinkFn(), req.renameFn())
		journalErr := removeMigrationJournalIfClean(req.Home, rollbackErr)
		return result, migrationFaultError(req.Home, err, rollbackErr, journalErr)
	}
	result.Applied = applied
	if req.InjectFault != nil {
		if err := req.InjectFault(MigrateFaultPublishBegin, len(applied)); err != nil {
			rollbackErr := reconcileJournalLinks(req.Home, journal, req.symlinkFn(), req.renameFn())
			journalErr := removeMigrationJournalIfClean(req.Home, rollbackErr)
			return result, fmt.Errorf("migration marker publish failed: %v; %s; re-run `curator env migrate --plan%s` to confirm the restored state",
				err, rollbackOutcome(req.Home, rollbackErr, journalErr), report.scopeFlags())
		}
	}
	if len(changed) > 0 {
		if err := op.publish(changed); err != nil {
			rollbackErr := reconcileJournalLinks(req.Home, journal, req.symlinkFn(), req.renameFn())
			journalErr := removeMigrationJournalIfClean(req.Home, rollbackErr)
			return result, fmt.Errorf("migration marker publish failed: %v; %s; re-run `curator env migrate --plan%s` to confirm the restored state",
				err, rollbackOutcome(req.Home, rollbackErr, journalErr), report.scopeFlags())
		}
		journal.MarkersDone = true
		if err := writeMigrationJournal(req.Home, journal); err != nil {
			rbErr := rollbackMigrationMarkersAndLinks(op, req, journal)
			journalErr := removeMigrationJournalIfClean(req.Home, rbErr)
			return result, fmt.Errorf("migration apply failed and was rolled back: cannot update the journal after publishing markers: %v; %s",
				err, rollbackOutcome(req.Home, rbErr, journalErr))
		}
	}
	if req.InjectFault != nil {
		if err := req.InjectFault(MigrateFaultPublish, len(applied)); err != nil {
			rbErr := rollbackMigrationMarkersAndLinks(op, req, journal)
			journalErr := removeMigrationJournalIfClean(req.Home, rbErr)
			return result, migrationFaultError(req.Home, err, rbErr, journalErr)
		}
	}
	journal.Complete = true
	if err := writeMigrationJournal(req.Home, journal); err != nil {
		rbErr := rollbackMigrationMarkersAndLinks(op, req, journal)
		journalErr := removeMigrationJournalIfClean(req.Home, rbErr)
		return result, fmt.Errorf("migration apply failed and was rolled back: cannot seal the journal: %v; %s",
			err, rollbackOutcome(req.Home, rbErr, journalErr))
	}
	if err := removeMigrationJournal(req.Home); err != nil {
		return result, fmt.Errorf("migration applied %d operations but cannot remove the completed journal %s: %v; the sealed journal is inert — the next apply deletes it without touching links — but back up manager state before re-running",
			len(applied), migrationJournalPath(req.Home), err)
	}
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

// migrationCrashError marks a simulated process kill: the journal stands
// with the operations applied so far and no rollback runs, so the next
// apply must recover. Only the InjectFault seam produces it.
type migrationCrashError struct {
	applied []MigrateOp
}

func (e *migrationCrashError) Error() string {
	return fmt.Sprintf("simulated crash after %d operations", len(e.applied))
}

// executeMigrationOps re-validates each operation against live state
// under the held lock and executes it through the durable journal: the
// intent for every operation is on disk before the first mutation, and
// each operation is marked done durably right after it lands, so a kill
// between any two steps recovers deterministically. Relinks move via a
// journaled temp-link + atomic rename (never remove-then-create), so a
// failed syscall leaves the old link standing. Only symlink, rename, and
// unlink syscalls run here — credential bytes are never read.
func executeMigrationOps(req MigrateRequest, ops []MigrateOp, journal *migrationJournal) ([]MigrateOp, error) {
	var applied []MigrateOp
	for i, op := range ops {
		full := filepath.Join(ManagedHomeDir(req.Home, op.Profile, op.EnvID), filepath.FromSlash(op.Path))
		if err := validateMigrationOp(full, op); err != nil {
			return applied, err
		}
		switch op.Kind {
		case MigrateOpRelink:
			if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
				return applied, fmt.Errorf("relink %s: %v", op.Path, err)
			}
			if err := atomicRelinkJournaled(req.Home, journal, i, full, op.To, req.symlinkFn(), req.renameFn()); err != nil {
				return applied, fmt.Errorf("relink %s: %v", op.Path, err)
			}
		case MigrateOpUnlink:
			if op.From != "" {
				if err := os.Remove(full); err != nil {
					return applied, fmt.Errorf("unlink %s: %v", op.Path, err)
				}
			}
		default:
			return applied, fmt.Errorf("unknown migration operation %q", op.Kind)
		}
		applied = append(applied, op)
		journal.Ops[i].Done = true
		journal.Ops[i].Temp = ""
		journal.Ops[i].TempTarget = ""
		if err := writeMigrationJournal(req.Home, journal); err != nil {
			return applied, fmt.Errorf("cannot journal the completed %s %s: %v", op.Kind, op.Path, err)
		}
		if req.InjectFault != nil {
			if err := req.InjectFault(MigrateFaultLink, len(applied)); err != nil {
				return applied, err
			}
			if err := req.InjectFault(MigrateFaultCrash, len(applied)); err != nil {
				return applied, &migrationCrashError{applied: applied}
			}
		}
	}
	return applied, nil
}

// atomicRelinkJournaled re-points the symlink at full to target without
// a remove-then-create window: the replacement is created under a
// temporary name in the same directory, then renamed over the old link
// in one atomic step. The temporary path and its expected target are
// journaled before the symlink runs, so a kill between the symlink and
// the rename leaves an owned temp the next recovery cleans; cleanup
// removes only that recorded path and only when it is a symlink to the
// recorded target. Any failure leaves the old link standing; a failed
// rename removes its own temporary link best effort (a leftover stays
// journaled). On success the recorded Temp stands until the caller
// marks the operation done and clears it in the same journal write — a
// crash in between leaves Temp pointing at an absent path, which
// cleans as a no-op.
func atomicRelinkJournaled(home string, journal *migrationJournal, index int, full, target string, symlink func(string, string) error, rename func(string, string) error) error {
	dir := filepath.Dir(full)
	for attempt := 0; attempt < 10; attempt++ {
		tmp := filepath.Join(dir, ".migrate-"+migrationRandHex()+".tmp")
		metadata, err := stateread.Lstat(tmp)
		if err != nil {
			return err
		}
		if metadata.Kind == stateread.KindPresent {
			continue
		}
		if metadata.Kind != stateread.KindAbsent {
			return stateread.UnusableError(tmp, fmt.Errorf("unknown temporary metadata state %q", metadata.Kind))
		}
		journal.Ops[index].Temp = tmp
		journal.Ops[index].TempTarget = target
		if err := writeMigrationJournal(home, journal); err != nil {
			journal.Ops[index].Temp = ""
			journal.Ops[index].TempTarget = ""
			return fmt.Errorf("cannot journal the temporary link intent: %v", err)
		}
		if err := symlink(target, tmp); err != nil {
			if os.IsExist(err) {
				journal.Ops[index].Temp = ""
				journal.Ops[index].TempTarget = ""
				continue
			}
			return err
		}
		if err := rename(tmp, full); err != nil {
			_ = removeOwnedTemp(tmp, target)
			return err
		}
		return nil
	}
	return fmt.Errorf("cannot pick a temporary link name in %s", dir)
}

// removeOwnedTemp removes the recorded temporary link only when it is a
// symlink to the recorded target. An absent path is a no-op; a regular
// file, directory, or symlink to any other target is never deleted and
// reports a mismatch for the caller to refuse on.
func removeOwnedTemp(tmp, want string) error {
	metadata, err := stateread.Lstat(tmp)
	if err != nil {
		return err
	}
	if metadata.Kind == stateread.KindAbsent {
		return nil
	}
	if metadata.Kind != stateread.KindPresent || metadata.Info == nil {
		return stateread.UnusableError(tmp, fmt.Errorf("unknown temporary metadata state %q", metadata.Kind))
	}
	info := metadata.Info
	if info.Mode()&os.ModeSymlink == 0 {
		shape := "a regular file"
		if info.IsDir() {
			shape = "a directory"
		}
		return fmt.Errorf("temporary path %s holds %s, not the owned link: move it aside out of band and re-run", tmp, shape)
	}
	got, err := os.Readlink(tmp)
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("temporary path %s links to %s, not the owned target %s: remove it out of band and re-run", tmp, got, want)
	}
	return os.Remove(tmp)
}

// migrationRandHex names one temporary link. Uniqueness is best effort
// (the journaled relink retries on collision); ownership comes from the
// journal record, never from the .migrate-*.tmp name pattern.
func migrationRandHex() string {
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(random[:])
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
		updated := *home.marker
		updated.Version = envmarker.VersionV2
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

// validateRecoveryInventory re-checks the entire recovery inventory
// before any recovery write: every journaled marker must read back as
// the recorded prior or the intended bytes, every journaled link must
// sit at its recorded prior or intended target, and every recorded
// temporary path must be absent or the owned symlink. Anything else —
// an edited marker, an unexpected link target, a foreign file at a
// recorded temp path — refuses naming the exact operator action,
// preserving the unknown state and the journal. Only states proven to
// belong to this transaction reconcile.
func validateRecoveryInventory(home string, journal *migrationJournal) error {
	var failures []string
	for _, marker := range journal.Markers {
		markerFile, err := stateread.ReadFile(marker.Live)
		if err != nil {
			failures = append(failures, fmt.Sprintf("marker %s cannot be inspected: %v; restore access out of band and re-run", marker.Live, err))
			continue
		}
		if markerFile.Kind == stateread.KindAbsent {
			failures = append(failures, fmt.Sprintf("marker %s is missing (expected the prior or the intended bytes): restore it out of band from backup and re-run", marker.Live))
			continue
		}
		if markerFile.Kind != stateread.KindPresent {
			failures = append(failures, fmt.Sprintf("marker %s cannot be inspected: %v; restore access out of band and re-run", marker.Live, stateread.UnusableError(marker.Live, fmt.Errorf("unknown file state %q", markerFile.Kind))))
			continue
		}
		current := markerFile.Bytes
		if !bytes.Equal(current, marker.Prior) && !bytes.Equal(current, marker.New) {
			failures = append(failures, fmt.Sprintf("marker %s was edited out of band (neither the prior nor the intended bytes): decide which bytes win, back %s up out of band, reconcile it by hand, and re-run", marker.Live, marker.Live))
		}
	}
	for _, op := range journal.Ops {
		full := filepath.Join(ManagedHomeDir(home, op.Profile, op.EnvID), filepath.FromSlash(op.Path))
		where := fmt.Sprintf("profile %s, environment %s, %s", op.Profile, op.EnvID, op.Path)
		metadata, err := stateread.Lstat(full)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s (%s) cannot be inspected: %v; restore access out of band and re-run", op.Path, where, err))
			continue
		}
		if metadata.Kind == stateread.KindAbsent {
			switch {
			case op.Kind == MigrateOpRelink:
				failures = append(failures, fmt.Sprintf("%s (%s) is missing (expected a link to %s or %s): restore the recorded link out of band and re-run", op.Path, where, op.From, op.To))
			case op.Kind == MigrateOpUnlink && op.From == "":
				// Absent is both the prior and the intended state.
			default:
				// Absent is the intended unlinked state.
			}
			continue
		}
		if metadata.Kind != stateread.KindPresent || metadata.Info == nil {
			failures = append(failures, fmt.Sprintf("%s (%s) cannot be inspected: %v; restore access out of band and re-run", op.Path, where, stateread.UnusableError(full, fmt.Errorf("unknown metadata state %q", metadata.Kind))))
			continue
		}
		info := metadata.Info
		if info.Mode()&os.ModeSymlink == 0 {
			shape := "a regular file"
			if info.IsDir() {
				shape = "a directory"
			}
			failures = append(failures, fmt.Sprintf("%s (%s) holds %s where the recorded link was: decide which side wins and restore the recorded link out of band (the manager never moves credential bytes), then re-run", op.Path, where, shape))
			continue
		}
		got, err := os.Readlink(full)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s (%s) link target cannot be read: restore access out of band and re-run", op.Path, where))
			continue
		}
		switch {
		case op.Kind == MigrateOpRelink:
			if got != op.From && got != op.To {
				failures = append(failures, fmt.Sprintf("%s (%s) links to %s, expected %s or %s: decide which target wins out of band, restore the recorded link by hand (the manager never moves credential bytes), and re-run", op.Path, where, got, op.From, op.To))
			}
		case op.Kind == MigrateOpUnlink && op.From == "":
			failures = append(failures, fmt.Sprintf("%s (%s) links to %s where no link is expected: remove it out of band and re-run", op.Path, where, got))
		default:
			if got != op.From {
				failures = append(failures, fmt.Sprintf("%s (%s) links to %s, expected %s or absent: decide which side wins out of band, restore the recorded link by hand (the manager never moves credential bytes), and re-run", op.Path, where, got, op.From))
			}
		}
	}
	for _, op := range journal.Ops {
		if op.Temp == "" {
			continue
		}
		metadata, err := stateread.Lstat(op.Temp)
		if err != nil {
			failures = append(failures, fmt.Sprintf("temporary path %s cannot be inspected: %v; restore access out of band and re-run", op.Temp, err))
			continue
		}
		if metadata.Kind == stateread.KindAbsent {
			continue
		}
		if metadata.Kind != stateread.KindPresent || metadata.Info == nil {
			failures = append(failures, fmt.Sprintf("temporary path %s cannot be inspected: %v; restore access out of band and re-run", op.Temp, stateread.UnusableError(op.Temp, fmt.Errorf("unknown metadata state %q", metadata.Kind))))
			continue
		}
		info := metadata.Info
		if info.Mode()&os.ModeSymlink == 0 {
			shape := "a regular file"
			if info.IsDir() {
				shape = "a directory"
			}
			failures = append(failures, fmt.Sprintf("temporary path %s holds %s, not the owned link to %s: move it aside out of band and re-run", op.Temp, shape, op.TempTarget))
			continue
		}
		got, err := os.Readlink(op.Temp)
		if err != nil {
			failures = append(failures, fmt.Sprintf("temporary path %s target cannot be read: restore access out of band and re-run", op.Temp))
			continue
		}
		if got != op.TempTarget {
			failures = append(failures, fmt.Sprintf("temporary path %s links to %s, not the owned target %s: remove it out of band and re-run", op.Temp, got, op.TempTarget))
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("migration recovery found %d unexpected states, refusing before any write (the journal stands at %s and nothing was changed): %s",
			len(failures), migrationJournalPath(home), strings.Join(failures, "; "))
	}
	return nil
}

// reconcileJournalLinks converges every journaled link to its prior
// (From) state in reverse plan order: already-prior links are left
// alone, moved links move back through a journaled temp-link + atomic
// rename, and removed links are re-created. Each operation's recorded
// temporary link is removed first, and only when it is a symlink to the
// recorded target — foreign paths are never touched. Anything
// unexpected fails closed naming the operator choice instead of
// guessing. Journal deletes are the caller's: in-process rollback
// deletes the journal on success, while recovery deletes it after the
// markers restore.
func reconcileJournalLinks(home string, journal *migrationJournal, symlink func(string, string) error, rename func(string, string) error) error {
	var failures []string
	for i := len(journal.Ops) - 1; i >= 0; i-- {
		op := journal.Ops[i]
		if op.Temp != "" {
			if err := removeOwnedTemp(op.Temp, op.TempTarget); err != nil {
				failures = append(failures, err.Error())
			}
		}
		if op.Kind == MigrateOpUnlink && op.From == "" {
			continue
		}
		full := filepath.Join(ManagedHomeDir(home, op.Profile, op.EnvID), filepath.FromSlash(op.Path))
		if err := reconcileLinkToPriorJournaled(home, journal, i, full, op, symlink, rename); err != nil {
			failures = append(failures, err.Error())
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("cannot restore %d prior links: %s", len(failures), strings.Join(failures, "; "))
	}
	return nil
}

// reconcileLinkToPriorJournaled converges one link path to its prior
// target: absent re-creates it, already-prior is a no-op, and a link at
// the intended target moves back through a journaled temp-link + atomic
// rename. Anything else — a non-symlink, or a symlink to an unexpected
// target — fails closed: the operator resolves it out of band and
// re-runs. The intended-target check re-validates at mutation time, so
// an out-of-band edit between the recovery validation and this write
// refuses instead of erasing the operator's change.
func reconcileLinkToPriorJournaled(home string, journal *migrationJournal, index int, full string, op migrationJournalOp, symlink func(string, string) error, rename func(string, string) error) error {
	want := op.From
	rel := op.Path
	metadata, err := stateread.Lstat(full)
	if err != nil {
		return fmt.Errorf("%s: cannot inspect the link path: %v; restore access out of band and re-run", rel, err)
	}
	if metadata.Kind == stateread.KindAbsent {
		if op.Kind == MigrateOpRelink {
			return fmt.Errorf("%s: is missing (expected a link to %s or %s): restore the recorded link out of band and re-run", rel, op.From, op.To)
		}
		if err := symlink(want, full); err != nil {
			return fmt.Errorf("%s: cannot re-create the prior link: %v", rel, err)
		}
		return nil
	}
	if metadata.Kind != stateread.KindPresent || metadata.Info == nil {
		return fmt.Errorf("%s: cannot inspect the link path: %v; restore access out of band and re-run", rel, stateread.UnusableError(full, fmt.Errorf("unknown metadata state %q", metadata.Kind)))
	}
	info := metadata.Info
	if info.Mode()&os.ModeSymlink == 0 {
		shape := "a regular file"
		if info.IsDir() {
			shape = "a directory"
		}
		return fmt.Errorf("%s: holds %s where the recorded link was: decide which side wins and restore the recorded link out of band (the manager never moves credential bytes), then re-run", rel, shape)
	}
	got, err := os.Readlink(full)
	if err != nil {
		return fmt.Errorf("%s: link target cannot be read: restore access out of band and re-run", rel)
	}
	if got == want {
		return nil
	}
	if op.Kind == MigrateOpRelink {
		if got != op.To {
			return fmt.Errorf("%s: links to %s, expected %s or %s: decide which target wins out of band, restore the recorded link by hand (the manager never moves credential bytes), and re-run", rel, got, op.From, op.To)
		}
	} else if got != want {
		return fmt.Errorf("%s: links to %s, expected %s or absent: decide which side wins out of band, restore the recorded link by hand (the manager never moves credential bytes), and re-run", rel, got, want)
	}
	if err := atomicRelinkJournaled(home, journal, index, full, want, symlink, rename); err != nil {
		return fmt.Errorf("%s: cannot move the link back: %v", rel, err)
	}
	return nil
}

// rollbackMigrationMarkersAndLinks restores the journal's prior marker
// bytes through the journaled record publication, then converges the
// links back. Best effort per stage; failures are collected.
func rollbackMigrationMarkersAndLinks(op *operation, req MigrateRequest, journal *migrationJournal) error {
	var stages []string
	if len(journal.Markers) > 0 {
		priors := make(map[string][]byte, len(journal.Markers))
		for _, marker := range journal.Markers {
			priors[marker.Live] = marker.Prior
		}
		if err := op.publish(priors); err != nil {
			stages = append(stages, fmt.Sprintf("cannot restore the prior markers: %v", err))
		}
	}
	if err := reconcileJournalLinks(req.Home, journal, req.symlinkFn(), req.renameFn()); err != nil {
		stages = append(stages, err.Error())
	}
	if len(stages) > 0 {
		return fmt.Errorf("%s", strings.Join(stages, "; "))
	}
	return nil
}

// removeMigrationJournalIfClean deletes the journal after a clean
// rollback; a dirty rollback keeps it so the next apply retries the
// recovery. A delete failure is reported so the caller can warn that
// the next apply re-runs the idempotent recovery.
func removeMigrationJournalIfClean(home string, rollbackErr error) error {
	if rollbackErr != nil {
		return nil
	}
	return removeMigrationJournal(home)
}

// rollbackOutcome phrases a rollback result for error text.
func rollbackOutcome(home string, rollbackErr, journalErr error) string {
	switch {
	case rollbackErr != nil:
		return fmt.Sprintf("rollback incomplete (%v); the journal stands at %s and the next apply retries the recovery", rollbackErr, migrationJournalPath(home))
	case journalErr != nil:
		return fmt.Sprintf("rolled back, but the journal could not be removed (%v); the next apply re-runs the idempotent recovery", journalErr)
	default:
		return "rolled back"
	}
}

// migrationFaultError joins an apply failure with its rollback outcome.
func migrationFaultError(home string, fault, rollbackErr, journalErr error) error {
	return fmt.Errorf("migration apply failed: %v; %s", fault, rollbackOutcome(home, rollbackErr, journalErr))
}

// migrationJournalVersion is the durable journal schema below the
// manager state directory. Journals are single-flight under the
// manager-home mutation lock: at most one apply mutates at a time.
const migrationJournalVersion = 1

// migrationJournalOp is one journaled link operation: the intent (Kind,
// Path, From, To) plus whether it landed (Done). From is the rollback
// data: the prior link target ("" for an unlink whose link was already
// gone). To is empty for unlink. Temp records the exact temporary-link
// path the journal owns for this operation's most recent atomic
// replacement (empty when none is owned); TempTarget is the expected
// symlink target of that temporary link. Cleanup removes only the
// recorded Temp path and only when it is a symlink to TempTarget — a
// regular file, directory, or foreign symlink is never deleted.
type migrationJournalOp struct {
	Profile    string `json:"profile"`
	EnvID      string `json:"env"`
	Kind       string `json:"kind"`
	Path       string `json:"path"`
	From       string `json:"from"`
	To         string `json:"to"`
	Done       bool   `json:"done"`
	Temp       string `json:"temp,omitempty"`
	TempTarget string `json:"temp_target,omitempty"`
}

// migrationJournalMarker is one journaled marker intent: the live path,
// the prior bytes to restore on rollback, and the bytes the apply
// publishes. Marker bytes are manager records, never credentials.
type migrationJournalMarker struct {
	Live  string `json:"live"`
	Prior []byte `json:"prior"`
	New   []byte `json:"new"`
}

// migrationJournal is the durable intent of one apply: the plan hash it
// executes, the link operations with rollback targets, and the marker
// intents. MarkersDone records a completed publication; Complete seals
// a fully applied plan so a crash before the journal delete cannot roll
// back completed work — recovery of a sealed journal only deletes it.
type migrationJournal struct {
	Version     int                      `json:"version"`
	Plan        string                   `json:"plan"`
	Ops         []migrationJournalOp     `json:"ops"`
	Markers     []migrationJournalMarker `json:"markers"`
	MarkersDone bool                     `json:"markers_done"`
	Complete    bool                     `json:"complete"`
}

// migrationJournalPath is the fixed journal below manager state (never
// below the credential scope, and holding no credential bytes).
func migrationJournalPath(home string) string {
	return filepath.Join(home, "state", "env-migration.json")
}

// newMigrationJournal builds the journal for one validated plan: every
// operation plus the marker intents in live-path order.
func newMigrationJournal(report *MigrateReport, changed, priors map[string][]byte) *migrationJournal {
	journal := &migrationJournal{
		Version: migrationJournalVersion,
		Plan:    report.Hash,
		Ops:     []migrationJournalOp{},
		Markers: []migrationJournalMarker{},
	}
	for _, op := range report.Ops() {
		journal.Ops = append(journal.Ops, migrationJournalOp{
			Profile: op.Profile, EnvID: op.EnvID, Kind: op.Kind, Path: op.Path, From: op.From, To: op.To,
		})
	}
	lives := make([]string, 0, len(changed))
	for live := range changed {
		lives = append(lives, live)
	}
	sort.Strings(lives)
	for _, live := range lives {
		journal.Markers = append(journal.Markers, migrationJournalMarker{Live: live, Prior: priors[live], New: changed[live]})
	}
	return journal
}

// writeMigrationJournal persists the journal durably (temp + sync +
// atomic rename) so the intent for every not-yet-executed operation is
// on disk before its mutation runs.
func writeMigrationJournal(home string, journal *migrationJournal) error {
	payload, err := json.Marshal(journal)
	if err != nil {
		return err
	}
	dir := filepath.Dir(migrationJournalPath(home))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".env-migration-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(payload); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, migrationJournalPath(home)); err != nil {
		return err
	}
	return nil
}

// readMigrationJournal loads the leftover journal: nil when no apply
// was interrupted. An unreadable or invalid journal is an error, never
// silence — the caller refuses and names the out-of-band fix.
func readMigrationJournal(home string) (*migrationJournal, error) {
	journalPath := migrationJournalPath(home)
	state, err := stateread.ReadFile(journalPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read: %v", err)
	}
	if state.Kind == stateread.KindAbsent {
		return nil, nil
	}
	if state.Kind != stateread.KindPresent {
		return nil, stateread.UnusableError(journalPath, fmt.Errorf("unknown file state %q", state.Kind))
	}
	payload := state.Bytes
	var journal migrationJournal
	if err := json.Unmarshal(payload, &journal); err != nil {
		return nil, fmt.Errorf("invalid JSON: %v", err)
	}
	if journal.Version != migrationJournalVersion {
		return nil, fmt.Errorf("unsupported version %d", journal.Version)
	}
	if journal.Plan == "" {
		return nil, fmt.Errorf("missing plan hash")
	}
	for _, op := range journal.Ops {
		switch op.Kind {
		case MigrateOpRelink, MigrateOpUnlink:
		default:
			return nil, fmt.Errorf("unknown operation %q", op.Kind)
		}
		if op.Path == "" || filepath.IsAbs(op.Path) || strings.Contains(op.Path, "..") {
			return nil, fmt.Errorf("invalid operation path %q", op.Path)
		}
		if op.Temp != "" {
			if !filepath.IsAbs(op.Temp) || !withinHome(home, op.Temp) {
				return nil, fmt.Errorf("invalid temporary path %q", op.Temp)
			}
			base := filepath.Base(op.Temp)
			if !strings.HasPrefix(base, ".migrate-") || !strings.HasSuffix(base, ".tmp") {
				return nil, fmt.Errorf("invalid temporary path %q", op.Temp)
			}
			if op.TempTarget == "" {
				return nil, fmt.Errorf("temporary path %q misses its target", op.Temp)
			}
		} else if op.TempTarget != "" {
			return nil, fmt.Errorf("temporary target without a path")
		}
	}
	for _, marker := range journal.Markers {
		if marker.Live == "" || !filepath.IsAbs(marker.Live) || !withinHome(home, marker.Live) {
			return nil, fmt.Errorf("invalid marker path %q", marker.Live)
		}
	}
	return &journal, nil
}

// withinHome reports whether the absolute target stays below the manager
// home: journaled absolute paths (marker lives, owned temps) must never
// escape it.
func withinHome(home, target string) bool {
	rel, err := filepath.Rel(home, target)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// removeMigrationJournal deletes the journal after a completed apply, a
// clean rollback, or a completed recovery.
func removeMigrationJournal(home string) error {
	if err := os.Remove(migrationJournalPath(home)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// recoverMigrationJournal rolls a leftover journal back to the prior
// state under the held lock: it validates the entire recovery inventory
// before any write, then links converge to their From targets in
// reverse plan order, prior marker bytes republish, and the journal is
// deleted. A sealed (fully applied) journal only deletes. The recovery
// is announced on print before the first mutation; an announcement
// failure fails before anything changes. It returns the pre-recovery
// summary for reporting, or nil when no journal was present.
func recoverMigrationJournal(home string, op *operation, announce io.Writer, symlink func(string, string) error, rename func(string, string) error) (*MigrateRecovery, error) {
	journal, err := readMigrationJournal(home)
	if err != nil {
		return nil, fmt.Errorf("migration journal %s is unusable (%v): back it up out of band, verify every managed link by hand, and only then remove it and re-run `curator env migrate --plan`",
			migrationJournalPath(home), err)
	}
	if journal == nil {
		return nil, nil
	}
	done := 0
	for _, entry := range journal.Ops {
		if entry.Done {
			done++
		}
	}
	recovered := &MigrateRecovery{Plan: journal.Plan, Ops: len(journal.Ops), Done: done}
	if journal.Complete {
		if announce != nil {
			if _, err := fmt.Fprintf(announce, "recovering interrupted migration apply: plan %s, %d of %d link operations completed; restoring the prior state under the manager lock\n",
				journal.Plan, done, len(journal.Ops)); err != nil {
				return recovered, fmt.Errorf("migration recovery cannot announce itself, refusing before any mutation: %v", err)
			}
		}
		if err := removeMigrationJournal(home); err != nil {
			return recovered, fmt.Errorf("migration recovery cannot remove the completed journal %s: %v", migrationJournalPath(home), err)
		}
		return recovered, nil
	}
	if err := validateRecoveryInventory(home, journal); err != nil {
		return recovered, err
	}
	if announce != nil {
		if _, err := fmt.Fprintf(announce, "recovering interrupted migration apply: plan %s, %d of %d link operations completed; restoring the prior state under the manager lock\n",
			journal.Plan, done, len(journal.Ops)); err != nil {
			return recovered, fmt.Errorf("migration recovery cannot announce itself, refusing before any mutation: %v", err)
		}
	}
	if err := reconcileJournalLinks(home, journal, symlink, rename); err != nil {
		return recovered, fmt.Errorf("migration recovery cannot restore the prior links: %v; the journal stands at %s and the next apply retries the recovery",
			err, migrationJournalPath(home))
	}
	if len(journal.Markers) > 0 {
		if err := validateRecoveryMarkers(journal); err != nil {
			return recovered, fmt.Errorf("migration recovery found the markers changed during the link restore, refusing before republishing: %v; the journal stands at %s and the next apply retries the recovery",
				err, migrationJournalPath(home))
		}
		priors := make(map[string][]byte, len(journal.Markers))
		for _, marker := range journal.Markers {
			priors[marker.Live] = marker.Prior
		}
		if err := op.publish(priors); err != nil {
			return recovered, fmt.Errorf("migration recovery cannot restore the prior markers: %v; the journal stands at %s and the next apply retries the recovery",
				err, migrationJournalPath(home))
		}
	}
	if err := removeMigrationJournal(home); err != nil {
		return recovered, fmt.Errorf("migration recovery restored the prior state but cannot remove the journal %s: %v; the next apply re-runs the idempotent recovery",
			migrationJournalPath(home), err)
	}
	return recovered, nil
}

// validateRecoveryMarkers re-checks the journaled markers immediately
// before the recovery republishes the priors: an out-of-band edit
// between the pre-recovery validation and this write refuses instead of
// erasing the operator's change.
func validateRecoveryMarkers(journal *migrationJournal) error {
	for _, marker := range journal.Markers {
		current, err := os.ReadFile(marker.Live) // #nosec G304 -- journaled manager record below the resolved home
		if err != nil {
			return fmt.Errorf("marker %s cannot be re-checked: %v; restore access out of band and re-run", marker.Live, err)
		}
		if !bytes.Equal(current, marker.Prior) && !bytes.Equal(current, marker.New) {
			return fmt.Errorf("marker %s was edited out of band (neither the prior nor the intended bytes): decide which bytes win, back %s up out of band, reconcile it by hand, and re-run", marker.Live, marker.Live)
		}
	}
	return nil
}
