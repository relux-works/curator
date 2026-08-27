package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/marker"
)

// Stable currentness codes.
//
// The first four are the historical `status` vocabulary and keep their exact
// meaning. The remaining ones classify compiled state. Every value is a stable
// machine-readable identifier: it appears verbatim in `status --json` and in
// the human line, and `status --check` treats every value except
// stateUpToDate and buildCurrent as a non-zero exit.
const (
	stateUpToDate     = "up-to-date"
	stateNotInstalled = "not-installed"
	stateNeedsInstall = "needs-install"
	stateContentDrift = "content-drift"
	stateUnresolvable = "unresolvable"

	// stateInvalidMarker is an install marker that is absent, unreadable, or
	// not a valid marker document.
	stateInvalidMarker = "invalid-marker"
	// stateUnsupportedMarker is a marker whose schema this manager cannot read.
	stateUnsupportedMarker = "unsupported-marker"

	// buildCurrent is the only compiled state that is up to date.
	buildCurrent = "current"
	// buildCommandDrift is a recorded compiled command set that differs from
	// the set the current closure activates.
	buildCommandDrift = "build-command-drift"
	// buildContextExposed is a build root that reached agent-facing context.
	buildContextExposed = "build-context-exposed"
	// buildSourceDrift is a recorded build-source identity that no longer
	// matches the frozen raw snapshot.
	buildSourceDrift = "build-source-drift"
	// buildInputDrift is a recorded logical key that does not match the key
	// derived from the current logical build input. The key is one opaque
	// digest over the whole input, so the code deliberately does not name a
	// cause; the accompanying cause subcode says exactly how much is provable.
	buildInputDrift = "build-input-drift"
	// buildUnsupportedDriver is a recorded or planned driver outside go-v1.
	buildUnsupportedDriver = "unsupported-build-driver"
	// buildUnusableToolchain is an active compiled command whose logical
	// identity could not be derived at all, because the trusted Go toolchain
	// could not be resolved or verified.
	buildUnusableToolchain = "unusable-build-toolchain"
	// buildMissingArtifact is a recorded key with no protected cache entry, so
	// the receipt and artifact the marker names are gone.
	buildMissingArtifact = "missing-build-artifact"
	// buildCorruptReceipt is a protected entry whose canonical receipt identity
	// differs from the one the marker recorded.
	buildCorruptReceipt = "corrupt-build-receipt"
	// buildArtifactDrift is a protected entry whose artifact path or hash
	// differs from the one the marker recorded.
	buildArtifactDrift = "build-artifact-drift"
	// buildCorruptCache is a protected entry that cannot be interpreted at all.
	buildCorruptCache = "corrupt-build-cache"
	// buildUntrustedCache is candidate cache state outside a provable
	// manager-protected boundary.
	buildUntrustedCache = "untrusted-build-cache"
	// buildUnsupportedPlatform is a host that cannot protect the build cache.
	buildUnsupportedPlatform = "unsupported-build-platform"
	// buildStateChanged is compiled state that moved while status classified
	// it, so the verdict this run derived is not authoritative.
	buildStateChanged = "build-state-changed"
	// buildUnknownState is a planner outcome this manager does not know. It
	// fails closed rather than being reported as current.
	buildUnknownState = "unknown-build-state"
)

// Stable cause subcodes. A cause refines a state without widening the state
// vocabulary, and is emitted only for the states that document one.
const (
	// causeBuildRoot means the marker's own recorded build roots do not contain
	// the build root the closure now activates. This is direct evidence.
	causeBuildRoot = "build-root"
	// causeTarget means the marker's own recorded artifact path is not the path
	// this manager derives for the current native target. This is direct
	// evidence that the recorded build was made for another target.
	causeTarget = "target"
	// causeUnattributed means the logical key differs but the install marker
	// records no prior input that could attribute it further. Source directory,
	// target tuning, toolchain identity, and the fixed manager build policy all
	// enter the same opaque digest, and none of them is recorded.
	causeUnattributed = "unattributed"
)

// currentnessCodes lists every stable code `status` can report, so a test can
// prove the documented set and the reachable set are the same.
func currentnessCodes() []string {
	return []string{
		stateUpToDate, stateNotInstalled, stateNeedsInstall, stateContentDrift, stateUnresolvable,
		stateInvalidMarker, stateUnsupportedMarker,
		buildCurrent, buildCommandDrift, buildContextExposed, buildSourceDrift, buildInputDrift,
		buildUnsupportedDriver, buildUnusableToolchain, buildMissingArtifact, buildCorruptReceipt,
		buildArtifactDrift, buildCorruptCache, buildUntrustedCache, buildUnsupportedPlatform,
		buildStateChanged, buildUnknownState,
	}
}

// inputCauses lists every stable cause subcode a build-input-drift row can
// carry, so the documented set and the emitted set stay the same.
func inputCauses() []string {
	return []string{causeBuildRoot, causeTarget, causeUnattributed}
}

// currentCode reports whether a code means "exactly current". Everything else
// — invalid, unsupported, missing, corrupt, drifted, or unknown — is drift.
func currentCode(code string) bool {
	return code == stateUpToDate || code == buildCurrent
}

// checkFailed is the `status --check` verdict of one scope. It fails closed:
// only an exactly current skill set and an exactly current compiled command set
// exit zero. Compiled commands are consulted separately from the declared-skill
// map because a build can belong to a transitively resolved node that no
// project declaration names.
func checkFailed(drift map[string]string, builds []buildReport) bool {
	for _, state := range drift {
		if !currentCode(state) {
			return true
		}
	}
	for _, build := range builds {
		if !currentCode(build.State) {
			return true
		}
	}
	return false
}

// buildFacts is the presentation view of one planned compiled command. It is a
// value type so classification is a pure function of independently derived
// evidence rather than of a live plan.
type buildFacts struct {
	Skill       string
	Command     string
	Driver      string
	BuildRoot   string
	SourceDir   string
	Source      buildsource.Identity
	Target      buildmeta.Target
	CacheKey    buildmeta.CacheKey
	Outcome     string
	Reason      string
	Diagnostic  string
	ReceiptHash buildmeta.ReceiptHash
	Artifact    buildmeta.Artifact
}

// factsOf projects one planned build onto the presentation view. Only
// protocol-relative paths and portable identities cross this boundary: the
// absolute protected-cache artifact path of a hit stays inside the planner.
func factsOf(build install.PlannedBuild) buildFacts {
	return buildFacts{
		Skill:       build.Skill(),
		Command:     build.Command(),
		Driver:      build.Driver(),
		BuildRoot:   build.BuildRoot(),
		SourceDir:   build.SourceDir(),
		Source:      build.Source(),
		Target:      build.Target(),
		CacheKey:    build.CacheKey(),
		Outcome:     string(build.Outcome()),
		Reason:      build.Reason(),
		Diagnostic:  build.DiagnosticCode(),
		ReceiptHash: build.ReceiptSHA256(),
		Artifact:    build.Artifact(),
	}
}

// buildReport is one machine-readable compiled-command diagnostic. Every field
// is either a portable identity, a protocol-relative path, or a stable code.
type buildReport struct {
	Skill        string               `json:"skill"`
	Command      string               `json:"command"`
	Driver       string               `json:"driver"`
	BuildRoot    string               `json:"build_root,omitempty"`
	SourceDir    string               `json:"source_dir,omitempty"`
	BuildSource  buildsource.Identity `json:"build_source"`
	Target       string               `json:"target,omitempty"`
	CacheKey     string               `json:"cache_key"`
	ArtifactPath string               `json:"artifact_path,omitempty"`
	CacheOutcome string               `json:"cache_outcome,omitempty"`
	State        string               `json:"state"`
	// Cause is a stable subcode that refines State without widening the state
	// vocabulary. A build-input-drift row carries one of inputCauses(); an
	// unusable-build-toolchain row carries the go-v1 boundary code that refused
	// the operation. Every other state leaves it empty.
	Cause  string `json:"cause,omitempty"`
	Detail string `json:"detail,omitempty"`
}

// Describe renders one compiled-command diagnostic as a single report line in
// the same shape the installation plan uses.
func (report buildReport) Describe() string {
	line := fmt.Sprintf("%s.%s build driver=%s", report.Skill, report.Command, report.Driver)
	for _, field := range []struct{ name, value string }{
		{"root", report.BuildRoot},
		{"dir", report.SourceDir},
		{"source", buildSourceText(report.BuildSource)},
		{"target", report.Target},
		{"key", report.CacheKey},
		{"artifact", report.ArtifactPath},
		{"cache", report.CacheOutcome},
	} {
		if field.value != "" {
			line += " " + field.name + "=" + field.value
		}
	}
	line += " state=" + report.State
	if report.Cause != "" {
		line += " cause=" + report.Cause
	}
	if report.Detail != "" {
		line += fmt.Sprintf(" detail=%q", report.Detail)
	}
	return line
}

func buildSourceText(identity buildsource.Identity) string {
	if identity.Algorithm == "" && identity.ContentSHA256 == "" {
		return ""
	}
	return identity.Algorithm + ":" + identity.ContentSHA256
}

// renderTarget renders one native target and its single tuning input in the
// deterministic order the planner uses.
func renderTarget(target buildmeta.Target) string {
	if target.GOOS == "" && target.GOARCH == "" {
		return ""
	}
	rendered := target.GOOS + "/" + target.GOARCH
	keys := make([]string, 0, len(target.Tuning))
	for key := range target.Tuning {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		rendered += "+" + key + "=" + target.Tuning[key]
	}
	return rendered
}

// classifySkillBuilds compares the compiled state one install marker recorded
// with the state the read-only plan independently derived for the same skill.
//
// It never repairs, quarantines, launches, or otherwise touches cache state:
// every verdict comes from the marker plus the planner's read-only inspection.
// The returned state is the skill-level code; the reports carry one row per
// command in the union of the recorded and planned command sets.
func classifySkillBuilds(installedDir string, recorded *marker.Marker, planned []buildFacts) (string, []buildReport) {
	plannedByCommand := map[string]buildFacts{}
	for _, facts := range planned {
		plannedByCommand[facts.Command] = facts
	}
	recordedBuilds := map[string]marker.Build{}
	if recorded != nil {
		recordedBuilds = recorded.Builds
	}
	if len(plannedByCommand) == 0 && len(recordedBuilds) == 0 {
		return "", nil
	}

	if recorded == nil {
		// The marker was refused. Why it was refused is itself a stable code, so
		// an unreadable document, a schema this manager cannot read, and a driver
		// outside the closed set stay distinguishable to an operator.
		state, detail := markerRefusal(installedDir)
		return state, plannedRows(planned, state, "", detail)
	}
	if recorded.SchemaVersion != marker.SchemaVersion {
		// A marker below schema 2 cannot describe a compiled command at all, so
		// an installation that now activates one has to be installed again.
		return stateNeedsInstall, plannedRows(planned, stateNeedsInstall, "", fmt.Sprintf(
			"install marker schema %d cannot describe a compiled command; reinstall to record marker schema %d",
			recorded.SchemaVersion, marker.SchemaVersion))
	}

	if exposure := contextExposure(installedDir, recorded); exposure != "" {
		return buildContextExposed, plannedRows(planned, buildContextExposed, "", exposure)
	}

	skillState := buildCurrent
	demote := func(state string) {
		if skillState == buildCurrent {
			skillState = state
		}
	}

	commands := unionKeys(recordedBuilds, plannedByCommand)
	reports := make([]buildReport, 0, len(commands))
	for _, command := range commands {
		facts, isPlanned := plannedByCommand[command]
		build, isRecorded := recordedBuilds[command]
		switch {
		case !isRecorded:
			demote(buildCommandDrift)
			reports = append(reports, report(facts, buildCommandDrift, "",
				"the closure activates this compiled command, but the install marker records no build for it"))
		case !isPlanned:
			demote(buildCommandDrift)
			reports = append(reports, buildReport{
				Skill: recorded.Name, Command: command, Driver: build.Driver,
				CacheKey: string(build.CacheKey), ArtifactPath: build.ArtifactPath,
				BuildSource: identityOf(recorded.BuildSource), State: buildCommandDrift,
				Detail: "the install marker records this compiled command, but the current closure does not activate it",
			})
		default:
			state, cause, detail := classifyBuildCommand(build, recorded, facts)
			demote(state)
			reports = append(reports, report(facts, state, cause, detail))
		}
	}
	return skillState, reports
}

// classifyBuildCommand is the per-command decision procedure. It returns the
// stable state, an optional stable cause subcode, and the operator detail.
//
// The recorded portable identities are compared with the derived ones before
// the cache inspection is interpreted, so identity drift is never reported as a
// cache miss. The logical key is deliberately not treated as evidence about any
// single input: it is one opaque digest over the whole build input, and the
// marker records no prior input to compare against.
func classifyBuildCommand(recorded marker.Build, recordedMarker *marker.Marker, planned buildFacts) (string, string, string) {
	// A command the plan could not derive at all is reported as exactly that.
	// Nothing below it is knowable, so nothing below it is claimed.
	if install.BuildOutcome(planned.Outcome) == install.BuildToolchainUnavailable {
		return buildUnusableToolchain, planned.Diagnostic, withReason(
			"the trusted Go toolchain could not be resolved or verified, so no compiled command could be planned",
			planned.Reason)
	}
	if recorded.Driver != buildmeta.DriverGoV1 || planned.Driver != buildmeta.DriverGoV1 {
		return buildUnsupportedDriver, "", fmt.Sprintf(
			"recorded driver %q and planned driver %q must both be the closed %s driver",
			recorded.Driver, planned.Driver, buildmeta.DriverGoV1)
	}
	recordedSource := recordedMarker.BuildSource
	if recordedSource == nil {
		return stateInvalidMarker, "", "the install marker records a build without a build-source identity"
	}
	if *recordedSource != planned.Source {
		return buildSourceDrift, "", fmt.Sprintf(
			"recorded build source %s does not match the frozen snapshot identity %s",
			buildSourceText(*recordedSource), buildSourceText(planned.Source))
	}
	if recorded.CacheKey != planned.CacheKey {
		cause, detail := attributeInputDrift(recorded, recordedMarker, planned)
		return buildInputDrift, cause, detail
	}

	switch install.BuildOutcome(planned.Outcome) {
	case install.BuildCacheHit:
		switch {
		case planned.ReceiptHash != recorded.ReceiptSHA256:
			return buildCorruptReceipt, "", fmt.Sprintf(
				"the protected entry carries receipt %s, but the install marker recorded %s",
				planned.ReceiptHash, recorded.ReceiptSHA256)
		case planned.Artifact.Path != recorded.ArtifactPath:
			return buildArtifactDrift, "", fmt.Sprintf(
				"the protected entry carries artifact path %q, but the install marker recorded %q",
				planned.Artifact.Path, recorded.ArtifactPath)
		case planned.Artifact.SHA256 != recorded.ArtifactSHA256:
			return buildArtifactDrift, "", fmt.Sprintf(
				"the protected artifact hashes to %s, but the install marker recorded %s",
				planned.Artifact.SHA256, recorded.ArtifactSHA256)
		}
		return buildCurrent, "", ""
	case install.BuildWouldPreflightAndBuild:
		return buildMissingArtifact, "", withReason(
			"the protected build cache holds no entry for the recorded logical key", planned.Reason)
	case install.BuildWouldRebuildUntrustedCache:
		return buildUntrustedCache, "", withReason(
			"candidate cache bytes are outside a provable manager-protected boundary", planned.Reason)
	case install.BuildCorrupt:
		return buildCorruptCache, "", withReason(
			"the protected cache entry cannot be interpreted", planned.Reason)
	case install.BuildUnsupported:
		return buildUnsupportedPlatform, "", withReason(
			"this host cannot prove manager-protected build cache state", planned.Reason)
	default:
		return buildUnknownState, "", withReason(
			fmt.Sprintf("planner outcome %q is not a known compiled state", planned.Outcome), planned.Reason)
	}
}

// attributeInputDrift explains a logical-key mismatch without overclaiming.
//
// The cache key is a digest over the complete build input: schema version,
// driver, build source, build root, command, source directory, native target
// and its tuning, trusted toolchain identity, and the fixed manager build
// policy. Driver and build source were already compared above and still match,
// and the command name is the map key, so the difference is in one of the rest.
//
// Only two of those leave independent evidence in the install marker: the
// recorded build-root set, and the recorded artifact path, whose file-name form
// is derived from the target operating system. When one of them proves the
// cause, the row says so; otherwise the row states plainly that the marker
// cannot attribute it further rather than blaming the target.
func attributeInputDrift(recorded marker.Build, recordedMarker *marker.Marker, planned buildFacts) (string, string) {
	if planned.BuildRoot != "" && !containsString(recordedMarker.BuildRoots, planned.BuildRoot) {
		return causeBuildRoot, fmt.Sprintf(
			"the closure builds this command under build root %q, which the install marker does not record",
			planned.BuildRoot)
	}
	if derived, err := buildmeta.ArtifactPath(planned.Command, planned.Target.GOOS); err == nil &&
		recorded.ArtifactPath != "" && recorded.ArtifactPath != derived {
		return causeTarget, fmt.Sprintf(
			"the install marker recorded artifact %q, but target %s derives %q",
			recorded.ArtifactPath, renderTarget(planned.Target), derived)
	}
	return causeUnattributed, fmt.Sprintf(
		"recorded logical key %s does not match the key %s derived from the current build input; "+
			"the install marker records no prior input, so the source directory, the target tuning, "+
			"the trusted toolchain, and the manager build policy remain equally possible causes",
		recorded.CacheKey, planned.CacheKey)
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

// markerRefusal explains, in the same stable vocabulary, why the marker reader
// refused one installed marker.
//
// The reader itself is a strict validator and returns nothing but "refused", so
// this is a presentation-only second look at the same bytes. It never accepts a
// marker the reader rejected: it only decides which stable code an operator is
// told, so a document from a newer manager and a document with a driver outside
// the closed set do not both collapse into "invalid".
func markerRefusal(installedDir string) (string, string) {
	payload, err := os.ReadFile(filepath.Join(installedDir, marker.Name)) // #nosec G304 -- path derives from an installed skill directory
	if err != nil {
		return stateInvalidMarker,
			"the installed skill has no readable install marker, so no compiled command can be proven current"
	}
	var refused struct {
		SchemaVersion *int `json:"schema_version"`
		Builds        map[string]struct {
			Driver string `json:"driver"`
		} `json:"builds"`
	}
	if json.Unmarshal(payload, &refused) != nil || refused.SchemaVersion == nil {
		return stateInvalidMarker, "the install marker is not a readable install marker document"
	}
	if *refused.SchemaVersion != marker.SchemaVersion && *refused.SchemaVersion != marker.LegacySchemaVersion {
		return stateUnsupportedMarker, fmt.Sprintf(
			"install marker schema %d cannot be read by this manager; the newest supported schema is %d",
			*refused.SchemaVersion, marker.SchemaVersion)
	}
	commands := make([]string, 0, len(refused.Builds))
	for command := range refused.Builds {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	for _, command := range commands {
		if driver := refused.Builds[command].Driver; driver != buildmeta.DriverGoV1 {
			return buildUnsupportedDriver, fmt.Sprintf(
				"the install marker records command %q under build driver %q, and %s is the only supported driver",
				command, driver, buildmeta.DriverGoV1)
		}
	}
	return stateInvalidMarker, "the install marker is present but is not a valid install marker document"
}

// contextExposure reports the build-root exclusion violation of one installed
// skill, or an empty string when the boundary is intact. Both directions are
// checked: the file set the marker recorded and what is on disk right now.
func contextExposure(installedDir string, recorded *marker.Marker) string {
	if len(recorded.Builds) == 0 {
		return ""
	}
	for _, file := range recorded.Files {
		for _, root := range recorded.BuildRoots {
			if file == root || strings.HasPrefix(file, root+"/") {
				return fmt.Sprintf("install marker records context file %q inside build root %q", file, root)
			}
		}
	}
	for _, root := range recorded.BuildRoots {
		if _, err := os.Lstat(filepath.Join(installedDir, filepath.FromSlash(root))); err == nil {
			return fmt.Sprintf("build root %q is materialized in agent-facing context", root)
		}
	}
	return ""
}

func withReason(detail, reason string) string {
	if reason == "" {
		return detail
	}
	return detail + ": " + install.RedactDiagnostic(reason)
}

func report(facts buildFacts, state, cause, detail string) buildReport {
	return buildReport{
		Skill:        facts.Skill,
		Command:      facts.Command,
		Driver:       facts.Driver,
		BuildRoot:    facts.BuildRoot,
		SourceDir:    facts.SourceDir,
		BuildSource:  facts.Source,
		Target:       renderTarget(facts.Target),
		CacheKey:     string(facts.CacheKey),
		ArtifactPath: facts.Artifact.Path,
		CacheOutcome: facts.Outcome,
		State:        state,
		Cause:        cause,
		Detail:       install.RedactDiagnostic(detail),
	}
}

func plannedRows(planned []buildFacts, state, cause, detail string) []buildReport {
	rows := make([]buildReport, 0, len(planned))
	for _, facts := range planned {
		rows = append(rows, report(facts, state, cause, detail))
	}
	return rows
}

func identityOf(source *buildsource.Identity) buildsource.Identity {
	if source == nil {
		return buildsource.Identity{}
	}
	return *source
}

func unionKeys(recorded map[string]marker.Build, planned map[string]buildFacts) []string {
	seen := map[string]bool{}
	var names []string
	for name := range recorded {
		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}
	for name := range planned {
		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// markerDigests fingerprints every install marker below the given stores, so a
// classification can tell an operator that installed state moved underneath it
// instead of publishing a verdict that was already stale when it was printed.
func markerDigests(stores ...string) map[string]string {
	digests := map[string]string{}
	for _, store := range stores {
		entries, err := os.ReadDir(store)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			path := filepath.Join(store, entry.Name(), marker.Name)
			payload, readErr := os.ReadFile(path) // #nosec G304 -- path derives from a manager-owned store
			if readErr != nil {
				digests[path] = "absent"
				continue
			}
			sum := sha256.Sum256(payload)
			digests[path] = hex.EncodeToString(sum[:])
		}
	}
	return digests
}

// markerMoved reports whether the marker of one installed directory differs
// between two fingerprints of the same stores.
func markerMoved(before, after map[string]string, installedDir string) bool {
	path := filepath.Join(installedDir, marker.Name)
	return before[path] != after[path]
}

// printRepairNotices reports what a real installation did with compiled state
// it was not allowed to reuse.
//
// Curator has no separate repair command: install and upgrade are the
// reconciliation path. They rebuild a missing, corrupt, drifted, or untrusted
// entry into new protected state after every declaration, closure, collision,
// audit, registry, and moved-tag gate has already passed, and they never adopt
// candidate bytes by changing permissions or rewriting a marker. These notices
// state which of the two happened, so an operator can see that a failed run
// left the previous installation in place.
func printRepairNotices(result install.Result) {
	for _, notice := range repairNotices(result.Alias, factsList(result.Builds), result.Status) {
		fmt.Println(notice)
	}
}

// factsList projects a whole plan onto the presentation view.
func factsList(builds []install.PlannedBuild) []buildFacts {
	facts := make([]buildFacts, 0, len(builds))
	for _, build := range builds {
		facts = append(facts, factsOf(build))
	}
	return facts
}

// repairedState names the compiled states install and upgrade repair by
// rebuilding, in the operator's words. A state that is absent from this map is
// either already current or one a rebuild cannot resolve.
func repairedState(outcome string) string {
	switch install.BuildOutcome(outcome) {
	case install.BuildWouldRebuildUntrustedCache:
		return "untrusted"
	case install.BuildCorrupt:
		return "corrupt"
	default:
		return ""
	}
}

func repairNotices(alias string, builds []buildFacts, status string) []string {
	var notices []string
	for _, build := range builds {
		notice := ""
		switch repaired := repairedState(build.Outcome); {
		case repaired != "" && status == "ok":
			notice = fmt.Sprintf("rebuilt %s build cache state into a new protected entry; "+
				"the previous installation stayed in place until every gate, the private build, "+
				"and the journaled commit succeeded", repaired)
		case repaired != "":
			notice = fmt.Sprintf("did not repair %s build cache state; "+
				"the previous installation, its consumers, and the live build cache are unchanged", repaired)
		case build.Outcome == string(install.BuildUnsupported),
			build.Outcome == string(install.BuildToolchainUnavailable):
			notice = fmt.Sprintf("refused %s build state before any mutation; "+
				"the previous installation, its consumers, and the live build cache are unchanged", build.Outcome)
		default:
			continue
		}
		notices = append(notices, fmt.Sprintf("%s: %s.%s: %s", alias, build.Skill, build.Command, notice))
	}
	return notices
}

// goToolchainGuidance renders operator guidance for a failed go-v1 trust
// boundary. It names the complete accepted selection mechanisms and the tested
// release families, and deliberately never suggests a PATH lookup or an
// automatic toolchain download: neither is an accepted selection mechanism.
func goToolchainGuidance(code string) string {
	selection := fmt.Sprintf(
		"select a trusted Go installation with %s=<GOROOT>/bin/go (bin/go.exe on Windows) or %s=<GOROOT>; "+
			"Curator never searches PATH and never downloads a toolchain",
		godriver.SelectionCuratorGo, godriver.SelectionGOROOT)
	families := "tested Go release families: " + strings.Join(godriver.TestedFamilies(), ", ")

	switch code {
	case "":
		return ""
	case "unsupported_go_family", "malformed_go_version":
		return "go-v1 " + code + ": the selected Go release is not tested against the go-v1 vectors; " +
			families + "; " + selection
	case "go_toolchain_missing", "untrusted_go_executable", "toolchain_executable_mismatch", "toolchain_mutated":
		return "go-v1 " + code + ": the trusted Go installation could not be resolved or verified; " +
			selection + "; " + families
	default:
		return "go-v1 " + code + ": the go-v1 build boundary refused this operation; " +
			selection + "; " + families
	}
}
