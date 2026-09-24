package scriptworker

import (
	"runtime"

	"github.com/relux-works/curator/internal/scriptpolicy"
	"github.com/relux-works/curator/internal/skillspec"
)

// Script inventory, probing, and capability evidence for the enforced
// `script-worker-v1` policy (manager profile §3.1 at the CI spec pin).
//
// The `script-worker-v1-native-control-inventory-v1` inventory is exhaustive:
// eight controls, three platforms, three availability states. The manager
// probes it once per invocation, before the worker starts, without
// substituting a host label, a cached result, or configuration for the
// probe. An `available` control MUST be applied and reported `applied`; an
// `unavailable` control MUST be neither applied nor reported applied; a
// `host-conditional` control is applied and reported `applied` exactly when
// this invocation's probe found it present, and reported `unavailable`
// otherwise — a control the probe did not find never rejects the
// invocation and never produces a diagnostic.
//
// This inventory is independent of `rc5-native-control-inventory-v1`. Its
// first five rows carry the same macOS and Windows verdicts because the
// underlying host facts do not depend on what the child is, and the probe
// and application mechanisms for those rows are the same ones the go-v1
// worker uses (process-group teardown, exact-limit Job Objects, RLIMIT_FSIZE
// across the fork, close-on-exec plus explicit release); Linux-only rows
// (delegated cgroup v2, Landlock, network namespaces) have no go-v1
// counterpart because the portable build policy is specified for macOS and
// Windows only.
const (
	// ScriptInventoryVersion is the exhaustive native-control inventory of
	// one enforced script invocation.
	ScriptInventoryVersion = "script-worker-v1-native-control-inventory-v1"
	// ScriptEvidenceVersion is the only capability-evidence record version
	// an enforced script invocation emits.
	ScriptEvidenceVersion = "script-capability-evidence-v1"
	// ScriptProbeTiming is the only permitted probed_at value. Availability
	// is probed once per invocation before the worker starts; an
	// install-generation record replayed at invocation time is a cached
	// result, not a probe.
	ScriptProbeTiming = "pre-worker-launch"

	// ScriptAvailabilityAvailable marks a control the inventory guarantees
	// for the host platform: the probe must confirm it and the invocation
	// must apply it.
	ScriptAvailabilityAvailable = "available"
	// ScriptAvailabilityHostConditional marks a control the probe may or
	// may not find present on this host.
	ScriptAvailabilityHostConditional = "host-conditional"
	// ScriptAvailabilityUnavailable marks a control the host platform
	// cannot provide: it is never probed, never applied, and never
	// rejects.
	ScriptAvailabilityUnavailable = "unavailable"

	// ScriptStatusApplied reports a control whose mechanism is installed
	// for this invocation.
	ScriptStatusApplied = "applied"
	// ScriptStatusUnavailable reports a control that is not installed for
	// this invocation.
	ScriptStatusUnavailable = "unavailable"

	// ScriptPlatformLinux is the inventory platform identifier for Linux.
	ScriptPlatformLinux = "linux"
	// ScriptPlatformMacOS is the inventory platform identifier for macOS.
	ScriptPlatformMacOS = "macos"
	// ScriptPlatformWindows is the inventory platform identifier for
	// Windows.
	ScriptPlatformWindows = "windows"
)

// Stable script capability-evidence diagnostics of manager profile §3.1.
// An unavailable inventory control, a host-conditional control the probe
// did not find, and the absence of any deferred guarantee never produce
// these; they reject only a record that contradicts the probe or claims a
// guarantee the policy defers.
const (
	CodeCapabilityEvidenceInvalid = "script_execution_capability_evidence_invalid"
	CodeHardenedClaimForbidden    = "script_execution_hardened_claim_forbidden"
)

// The exhaustive script inventory control names, in inventory order. Nothing
// outside this list may be applied, named, or reported.
const (
	ScriptControlDescendantDomainTermination = "descendant-domain-termination"
	ScriptControlActiveProcessCountLimit     = "active-process-count-limit"
	ScriptControlAggregateMemoryLimit        = "aggregate-memory-limit"
	ScriptControlPerFileSizeLimit            = "per-file-size-limit"
	ScriptControlInheritedHandleRestriction  = "inherited-handle-restriction"
	ScriptControlDescendantExecDenial        = "descendant-exec-denial"
	ScriptControlFilesystemWriteConfinement  = "filesystem-write-confinement"
	ScriptControlNetworkIsolationDomain      = "network-isolation-domain"
)

// scriptInventoryOrder is the exhaustive inventory order.
var scriptInventoryOrder = []string{
	ScriptControlDescendantDomainTermination,
	ScriptControlActiveProcessCountLimit,
	ScriptControlAggregateMemoryLimit,
	ScriptControlPerFileSizeLimit,
	ScriptControlInheritedHandleRestriction,
	ScriptControlDescendantExecDenial,
	ScriptControlFilesystemWriteConfinement,
	ScriptControlNetworkIsolationDomain,
}

// scriptInventoryRecord is the closed per-platform record of one inventory
// control: its availability, the mechanism the probe performs, and — for
// fixed-unavailable controls — the closed unavailable reason.
type scriptInventoryRecord struct {
	Availability      string
	Mechanism         string
	UnavailableReason string
}

// scriptInventoryPlatforms mirrors the normative per-platform inventory. A
// probe may only confirm or contradict an entry; it may not add or rename
// one.
var scriptInventoryPlatforms = map[string]map[string]scriptInventoryRecord{
	ScriptPlatformLinux: {
		ScriptControlDescendantDomainTermination: {Availability: ScriptAvailabilityAvailable, Mechanism: "process-group-and-session-teardown"},
		ScriptControlActiveProcessCountLimit:     {Availability: ScriptAvailabilityHostConditional, Mechanism: "delegated-cgroup-v2-pids.max"},
		ScriptControlAggregateMemoryLimit:        {Availability: ScriptAvailabilityHostConditional, Mechanism: "delegated-cgroup-v2-memory.max"},
		ScriptControlPerFileSizeLimit:            {Availability: ScriptAvailabilityAvailable, Mechanism: "rlimit-fsize"},
		ScriptControlInheritedHandleRestriction:  {Availability: ScriptAvailabilityAvailable, Mechanism: "close-on-exec-and-explicit-descriptor-release"},
		ScriptControlDescendantExecDenial:        {Availability: ScriptAvailabilityHostConditional, Mechanism: "landlock-execute-right"},
		ScriptControlFilesystemWriteConfinement:  {Availability: ScriptAvailabilityHostConditional, Mechanism: "landlock-write-rights"},
		ScriptControlNetworkIsolationDomain:      {Availability: ScriptAvailabilityHostConditional, Mechanism: "network-namespace-without-interfaces"},
	},
	ScriptPlatformMacOS: {
		ScriptControlDescendantDomainTermination: {Availability: ScriptAvailabilityAvailable, Mechanism: "process-group-and-session-teardown"},
		ScriptControlActiveProcessCountLimit:     {Availability: ScriptAvailabilityUnavailable, UnavailableReason: "no-private-aggregate-domain"},
		ScriptControlAggregateMemoryLimit:        {Availability: ScriptAvailabilityUnavailable, UnavailableReason: "no-private-aggregate-domain"},
		ScriptControlPerFileSizeLimit:            {Availability: ScriptAvailabilityAvailable, Mechanism: "rlimit-fsize"},
		ScriptControlInheritedHandleRestriction:  {Availability: ScriptAvailabilityAvailable, Mechanism: "close-on-exec-and-explicit-descriptor-release"},
		ScriptControlDescendantExecDenial:        {Availability: ScriptAvailabilityUnavailable, UnavailableReason: "no-unprivileged-per-process-exec-policy"},
		ScriptControlFilesystemWriteConfinement:  {Availability: ScriptAvailabilityUnavailable, UnavailableReason: "no-unprivileged-filesystem-domain"},
		ScriptControlNetworkIsolationDomain:      {Availability: ScriptAvailabilityUnavailable, UnavailableReason: "no-unprivileged-network-domain"},
	},
	ScriptPlatformWindows: {
		ScriptControlDescendantDomainTermination: {Availability: ScriptAvailabilityAvailable, Mechanism: "job-object-kill-on-close"},
		ScriptControlActiveProcessCountLimit:     {Availability: ScriptAvailabilityAvailable, Mechanism: "job-object-active-process-limit"},
		ScriptControlAggregateMemoryLimit:        {Availability: ScriptAvailabilityAvailable, Mechanism: "job-object-process-and-job-memory-limit"},
		ScriptControlPerFileSizeLimit:            {Availability: ScriptAvailabilityUnavailable, UnavailableReason: "no-private-aggregate-domain"},
		ScriptControlInheritedHandleRestriction:  {Availability: ScriptAvailabilityAvailable, Mechanism: "explicit-handle-inheritance-list"},
		ScriptControlDescendantExecDenial:        {Availability: ScriptAvailabilityUnavailable, UnavailableReason: "child-process-policy-requires-appcontainer"},
		ScriptControlFilesystemWriteConfinement:  {Availability: ScriptAvailabilityUnavailable, UnavailableReason: "no-unprivileged-filesystem-domain"},
		ScriptControlNetworkIsolationDomain:      {Availability: ScriptAvailabilityUnavailable, UnavailableReason: "no-unprivileged-network-domain"},
	},
}

// deferredScriptGuarantees are the seven guarantees Protocol Core §4.1.1
// defers for this policy. None of them may appear in the mandatory-control
// set, the native-control inventory, or an evidence record.
var deferredScriptGuarantees = []string{
	"script-total-network-denial",
	"script-network-host-allowlisting",
	"script-exact-executable-allowlisting",
	"script-private-runtime-area-only-writes",
	"script-read-only-runtime-tree",
	"script-hard-aggregate-descendant-resource-bounds",
	"script-fail-closed-capability-preflight",
}

// deferredBuildGuarantees are the six guarantees manager profile §2.2.1
// defers. A record carrying one of these names — a build guarantee claimed
// for a script invocation — is a forbidden hardened claim, and none of them
// aliases a control of either inventory.
var deferredBuildGuarantees = []string{
	"total-network-denial",
	"read-only-source-and-toolchain",
	"private-build-root-only-writes",
	"hard-aggregate-descendant-resource-bounds",
	"exact-executable-allowlisting",
	"fail-closed-capability-preflight",
}

// ScriptControlProbe is one per-invocation availability determination made
// in the manager parent before the worker exists. Present reports whether
// the mechanism probe found the control usable on this host for exactly
// this invocation; it is meaningful for `available` and `host-conditional`
// controls and always false for fixed-`unavailable` ones, which are never
// probed.
type ScriptControlProbe struct {
	Name         string
	Availability string
	Mechanism    string
	Present      bool
	ProbedAt     string
}

// scriptProbeFault is nil in production. Tests assign it to inject a real
// probe failure for one control and prove that the install and invoke
// preflights refuse with `script_execution_control_unavailable`, start no
// worker, claim no applied control, and publish nothing. It mirrors the
// go-v1 `controlSeamFault` precedent: a fault is a probe outcome, never a
// cached or configured availability.
var scriptProbeFault func(control string) error

// OverrideScriptProbeFaultForTest installs a probe fault injector and
// returns a restore function. Tests use it to force one control's probe to
// fail through the production install and invoke entries; production has no
// other way to set it. Callers must be sequential and must defer restore.
func OverrideScriptProbeFaultForTest(fault func(control string) error) (restore func()) {
	scriptProbeFault = fault
	return func() { scriptProbeFault = nil }
}

// testInventoryPlatform overrides the probed platform when non-empty. Tests
// use it to drive another unix-family inventory (macOS on Linux, Linux on
// macOS) through the production session with fixture control roots;
// production leaves it empty, so the host GOOS selects the inventory. The
// override never crosses the process boundary: the parent sends the probed
// list to the worker in the request, and the worker applies exactly that
// list, so availability is still probed once, before the worker launches.
// Cross-OS-family forcing (a unix inventory on Windows) is refused: the
// mechanisms do not exist there.
var testInventoryPlatform string

// OverrideInventoryPlatformForTest forces the probed inventory platform and
// returns a restore function. Production never sets it. See
// testInventoryPlatform for the boundary rules.
func OverrideInventoryPlatformForTest(platform string) (restore func()) {
	testInventoryPlatform = platform
	return func() { testInventoryPlatform = "" }
}

// ScriptInventoryPlatform maps a Go GOOS to the inventory platform
// identifier. An empty string means the inventory covers no record for the
// host, so no enforced invocation can proceed there.
func ScriptInventoryPlatform(goos string) string {
	switch goos {
	case "linux":
		return ScriptPlatformLinux
	case "darwin":
		return ScriptPlatformMacOS
	case "windows":
		return ScriptPlatformWindows
	default:
		return ""
	}
}

// effectiveInventoryPlatform is the platform the probe reads: the host
// inventory, or the test-only override when one is active.
func effectiveInventoryPlatform() string {
	if testInventoryPlatform != "" {
		return testInventoryPlatform
	}
	return ScriptInventoryPlatform(runtime.GOOS)
}

// probeScriptInventory determines, once for this invocation, which
// inventory controls this host provides. It runs in the parent before the
// worker exists and refuses with `script_execution_control_unavailable`
// when an `available` control cannot be confirmed, naming it. A
// host-conditional control the probe does not find is recorded absent and
// never refuses; a fixed-unavailable control is recorded without probing.
func probeScriptInventory() (string, []ScriptControlProbe, error) {
	platform := effectiveInventoryPlatform()
	records := scriptInventoryPlatforms[platform]
	if platform == "" || records == nil {
		return "", nil, scriptpolicy.HostControlError("",
			append([]string(nil), scriptInventoryOrder...),
			"script-worker-v1-native-control-inventory-v1 defines no record for this host")
	}
	if testInventoryPlatform != "" && !inventoryPlatformUsable(testInventoryPlatform) {
		return "", nil, diagnostic(CodeWorkerProtocolInvalid,
			"inventory platform %q cannot be probed on host %s", testInventoryPlatform, runtime.GOOS)
	}
	probes := make([]ScriptControlProbe, 0, len(scriptInventoryOrder))
	for _, name := range scriptInventoryOrder {
		record := records[name]
		if record.Availability == ScriptAvailabilityUnavailable {
			probes = append(probes, ScriptControlProbe{
				Name: name, Availability: ScriptAvailabilityUnavailable, ProbedAt: ScriptProbeTiming,
			})
			continue
		}
		present, err := probeScriptControl(platform, name)
		if err != nil {
			return "", nil, scriptpolicy.HostControlError("",
				[]string{name}, "cannot probe native control")
		}
		if !present {
			if record.Availability == ScriptAvailabilityHostConditional {
				probes = append(probes, ScriptControlProbe{
					Name: name, Availability: ScriptAvailabilityHostConditional,
					Mechanism: record.Mechanism, ProbedAt: ScriptProbeTiming,
				})
				continue
			}
			return "", nil, scriptpolicy.HostControlError("",
				[]string{name}, "native control is unavailable on this host")
		}
		probes = append(probes, ScriptControlProbe{
			Name: name, Availability: record.Availability,
			Mechanism: record.Mechanism, Present: true, ProbedAt: ScriptProbeTiming,
		})
	}
	return platform, probes, nil
}

// probeScriptControl performs the per-invocation availability determination
// for one control: the exact operation the control will perform, or the
// closest applicability proof the host allows without side effects. It
// starts no program except the self-reexec probes (network isolation and
// Landlock, whose mechanisms cannot be probed in-process without changing
// the probing process itself), caches nothing, and reads no configuration.
func probeScriptControl(platform, name string) (bool, error) {
	if scriptProbeFault != nil {
		if err := scriptProbeFault(name); err != nil {
			return false, err
		}
	}
	return probeScriptMechanism(platform, name)
}

// scriptControlInstallable reports whether the probe found the control
// present for this invocation: every `available` control (the probe refused
// unless present) and every `host-conditional` control the probe found.
func scriptControlInstallable(probe ScriptControlProbe) bool {
	return probe.Present &&
		(probe.Availability == ScriptAvailabilityAvailable ||
			probe.Availability == ScriptAvailabilityHostConditional)
}

// installableScriptControls lists, in inventory order, the controls this
// invocation's probes marked present. It is the exact set the session must
// apply before the interpreter starts.
func installableScriptControls(probes []ScriptControlProbe) []string {
	names := make([]string, 0, len(probes))
	for _, probe := range probes {
		if scriptControlInstallable(probe) {
			names = append(names, probe.Name)
		}
	}
	return names
}

// landlockGrants is the exact Landlock grant set of one invocation:
// manager-resolved executables for descendant-exec-denial, and the
// operation-private area plus the derived path set for
// filesystem-write-confinement. The worker's own null standard input
// carries no rule — the handle binds before any restriction, and
// Landlock never restricts an already-open descriptor — but the
// enforced write-confinement ruleset additionally grants the null
// device itself (file-typed write rights, so opens with O_TRUNC work
// on the sink too), so the confined interpreter and its descendants
// can open it for redirected standard streams.
type landlockGrants struct {
	executables []string
	writables   []string
}

// ScriptEvidenceEntry is one closed script-capability-evidence-v1 control
// entry. It contains exactly name, availability, status, and probed_at.
type ScriptEvidenceEntry struct {
	Name         string `json:"name"`
	Availability string `json:"availability"`
	Status       string `json:"status"`
	ProbedAt     string `json:"probed_at"`
}

// ScriptEvidence is the single closed script-capability-evidence-v1 record
// emitted per enforced invocation. It is result-only: it never enters a
// cache key, receipt input, install marker, or conformance claim, and it is
// never written to the command's standard output or standard error.
type ScriptEvidence struct {
	RecordVersion   string                `json:"record_version"`
	ExecutionPolicy string                `json:"execution_policy"`
	Platform        string                `json:"platform"`
	Controls        []ScriptEvidenceEntry `json:"controls"`
}

// buildScriptEvidence builds the closed record from this invocation's
// probes and the controls whose mechanism is actually installed. The parent
// derives it only for the installable set, and the worker derives the same
// record independently from what it confirms; the parent requires the two
// to be identical before permitting the run.
func buildScriptEvidence(platform string, probes []ScriptControlProbe, applied []string) ScriptEvidence {
	appliedSet := make(map[string]bool, len(applied))
	for _, name := range applied {
		appliedSet[name] = true
	}
	record := ScriptEvidence{
		RecordVersion:   ScriptEvidenceVersion,
		ExecutionPolicy: skillspec.ScriptExecutionPolicy,
		Platform:        platform,
		Controls:        make([]ScriptEvidenceEntry, 0, len(probes)),
	}
	for _, probe := range probes {
		status := ScriptStatusUnavailable
		if scriptControlInstallable(probe) && appliedSet[probe.Name] {
			status = ScriptStatusApplied
		}
		// Any other combination — an installable control the session did
		// not apply, a fixed-unavailable control, a host-conditional
		// control the probe did not find — is recorded faithfully as
		// unavailable, so validation rejects a contradiction instead of
		// hiding it.
		record.Controls = append(record.Controls, ScriptEvidenceEntry{
			Name: probe.Name, Availability: probe.Availability, Status: status, ProbedAt: probe.ProbedAt,
		})
	}
	return record
}

// scriptEvidenceEqual reports whether two records carry the same entries in
// the same order.
func scriptEvidenceEqual(first, second ScriptEvidence) bool {
	if first.RecordVersion != second.RecordVersion ||
		first.ExecutionPolicy != second.ExecutionPolicy ||
		first.Platform != second.Platform ||
		len(first.Controls) != len(second.Controls) {
		return false
	}
	for index := range first.Controls {
		if first.Controls[index] != second.Controls[index] {
			return false
		}
	}
	return true
}

func inScriptInventory(name string) bool {
	for _, control := range scriptInventoryOrder {
		if control == name {
			return true
		}
	}
	return false
}

func isDeferredScriptGuarantee(name string) bool {
	for _, guarantee := range deferredScriptGuarantees {
		if guarantee == name {
			return true
		}
	}
	for _, guarantee := range deferredBuildGuarantees {
		if guarantee == name {
			return true
		}
	}
	return false
}

// validateScriptEvidence enforces the closed consistency rules before the
// parent permits the run. The probes are this invocation's parent-side
// determination; a record that reports an availability the invocation did
// not probe is rejected, as is a status inconsistent with the availability
// rules, a non-`pre-worker-launch` timing, a missing, duplicated, extra, or
// unknown entry, and an unknown record version. A deferred guarantee named
// as an entry, and a record execution policy other than `script-worker-v1`,
// are forbidden hardened claims.
func validateScriptEvidence(record ScriptEvidence, platform string, probes []ScriptControlProbe) error {
	if record.ExecutionPolicy != skillspec.ScriptExecutionPolicy {
		return diagnostic(CodeHardenedClaimForbidden,
			"capability evidence execution_policy %q is not the enforced script identity %q",
			record.ExecutionPolicy, skillspec.ScriptExecutionPolicy)
	}
	for _, entry := range record.Controls {
		if isDeferredScriptGuarantee(entry.Name) {
			return diagnostic(CodeHardenedClaimForbidden,
				"capability evidence names deferred guarantee %q", entry.Name)
		}
	}
	if record.RecordVersion != ScriptEvidenceVersion {
		return diagnostic(CodeCapabilityEvidenceInvalid,
			"unknown capability evidence record_version %q", record.RecordVersion)
	}
	if record.Platform != platform || scriptInventoryPlatforms[record.Platform] == nil {
		return diagnostic(CodeCapabilityEvidenceInvalid,
			"capability evidence platform %q is not the probed platform %q", record.Platform, platform)
	}
	probed := make(map[string]ScriptControlProbe, len(probes))
	for _, probe := range probes {
		probed[probe.Name] = probe
	}
	seen := make(map[string]bool, len(record.Controls))
	for _, entry := range record.Controls {
		if !inScriptInventory(entry.Name) {
			return diagnostic(CodeCapabilityEvidenceInvalid,
				"capability evidence reports control %q outside script-worker-v1-native-control-inventory-v1", entry.Name)
		}
		if seen[entry.Name] {
			return diagnostic(CodeCapabilityEvidenceInvalid, "capability evidence duplicates control %q", entry.Name)
		}
		seen[entry.Name] = true
		if entry.ProbedAt != ScriptProbeTiming {
			return diagnostic(CodeCapabilityEvidenceInvalid,
				"capability evidence control %q reports probed_at %q", entry.Name, entry.ProbedAt)
		}
		probe, present := probed[entry.Name]
		if !present || entry.Availability != probe.Availability {
			return diagnostic(CodeCapabilityEvidenceInvalid,
				"capability evidence control %q reports availability %q that this invocation did not probe", entry.Name, entry.Availability)
		}
		switch entry.Availability {
		case ScriptAvailabilityAvailable:
			if entry.Status != ScriptStatusApplied {
				return diagnostic(CodeCapabilityEvidenceInvalid,
					"available control %q reports status %q", entry.Name, entry.Status)
			}
		case ScriptAvailabilityUnavailable:
			if entry.Status != ScriptStatusUnavailable {
				return diagnostic(CodeCapabilityEvidenceInvalid,
					"unavailable control %q reports status %q", entry.Name, entry.Status)
			}
		case ScriptAvailabilityHostConditional:
			want := ScriptStatusUnavailable
			if probe.Present {
				want = ScriptStatusApplied
			}
			if entry.Status != want {
				return diagnostic(CodeCapabilityEvidenceInvalid,
					"host-conditional control %q reports status %q against this invocation's probe", entry.Name, entry.Status)
			}
		default:
			return diagnostic(CodeCapabilityEvidenceInvalid,
				"control %q reports availability %q", entry.Name, entry.Availability)
		}
	}
	if len(seen) != len(scriptInventoryOrder) {
		missing := make([]string, 0, len(scriptInventoryOrder))
		for _, control := range scriptInventoryOrder {
			if !seen[control] {
				missing = append(missing, control)
			}
		}
		return diagnostic(CodeCapabilityEvidenceInvalid,
			"capability evidence is missing exactly one entry per inventory control: %v", missing)
	}
	return nil
}
