package godriver

import (
	"runtime"
	"sort"

	"github.com/relux-works/curator/internal/buildmeta"
)

// Stable execution-boundary diagnostics of Manager Profile section 2.2.1. They
// take precedence over the generic driver codes in this package.
const (
	CodeWorkerIdentityInvalid     = "build_execution_worker_identity_invalid"
	CodeWorkerProtocolInvalid     = "build_execution_worker_protocol_invalid"
	CodeControlUnavailable        = "build_execution_control_unavailable"
	CodeCapabilityEvidenceInvalid = "build_execution_capability_evidence_invalid"
	CodeHardenedClaimForbidden    = "build_execution_hardened_claim_forbidden"
	CodePackageInfluenceForbidden = "build_execution_package_influence_forbidden"
)

// Portable execution-policy identity and the two closed record versions of
// Protocol Core section 4.2.1.
const (
	ExecutionPolicy               = buildmeta.ExecutionPolicy
	NativeControlInventoryVersion = "rc5-native-control-inventory-v1"
	CapabilityEvidenceVersion     = "capability-evidence-v1"

	// ProbeTiming is the only permitted probed_at value. Availability is probed
	// once per operation before the worker starts; a host label, build-time
	// constant, configuration value, or cached result is not a probe.
	ProbeTiming = "pre-worker-launch"

	AvailabilityAvailable   = "available"
	AvailabilityUnavailable = "unavailable"

	StatusApplied     = "applied"
	StatusUnavailable = "unavailable"

	// UnavailableReasonNoPrivateAggregateDomain is the only reason vocabulary
	// entry defined by rc5-native-control-inventory-v1.
	UnavailableReasonNoPrivateAggregateDomain = "no-private-aggregate-domain"

	PlatformMacOS   = "macos"
	PlatformWindows = "windows"
)

// The exhaustive rc5-native-control-inventory-v1 control names, in inventory
// order. Nothing outside this list may be applied or reported.
const (
	ControlDescendantDomainTermination = "descendant-domain-termination"
	ControlActiveProcessCountLimit     = "active-process-count-limit"
	ControlAggregateMemoryLimit        = "aggregate-memory-limit"
	ControlPerFileSizeLimit            = "per-file-size-limit"
	ControlInheritedHandleRestriction  = "inherited-handle-restriction"
)

// nativeControlInventory is the exhaustive inventory order.
var nativeControlInventory = []string{
	ControlDescendantDomainTermination,
	ControlActiveProcessCountLimit,
	ControlAggregateMemoryLimit,
	ControlPerFileSizeLimit,
	ControlInheritedHandleRestriction,
}

// inventoryRecord is the closed per-platform record of one inventory control.
type inventoryRecord struct {
	Availability      string
	Mechanism         string
	UnavailableReason string
}

// nativeControlPlatforms mirrors the normative per-platform inventory. A probe
// may only confirm or contradict an entry; it may not add or rename one.
var nativeControlPlatforms = map[string]map[string]inventoryRecord{
	PlatformMacOS: {
		ControlDescendantDomainTermination: {Availability: AvailabilityAvailable, Mechanism: "process-group-and-session-teardown"},
		ControlActiveProcessCountLimit:     {Availability: AvailabilityUnavailable, UnavailableReason: UnavailableReasonNoPrivateAggregateDomain},
		ControlAggregateMemoryLimit:        {Availability: AvailabilityUnavailable, UnavailableReason: UnavailableReasonNoPrivateAggregateDomain},
		ControlPerFileSizeLimit:            {Availability: AvailabilityAvailable, Mechanism: "rlimit-fsize"},
		ControlInheritedHandleRestriction:  {Availability: AvailabilityAvailable, Mechanism: "close-on-exec-and-explicit-descriptor-release"},
	},
	PlatformWindows: {
		ControlDescendantDomainTermination: {Availability: AvailabilityAvailable, Mechanism: "job-object-kill-on-close"},
		ControlActiveProcessCountLimit:     {Availability: AvailabilityAvailable, Mechanism: "job-object-active-process-limit"},
		ControlAggregateMemoryLimit:        {Availability: AvailabilityAvailable, Mechanism: "job-object-process-and-job-memory-limit"},
		ControlPerFileSizeLimit:            {Availability: AvailabilityUnavailable, UnavailableReason: UnavailableReasonNoPrivateAggregateDomain},
		ControlInheritedHandleRestriction:  {Availability: AvailabilityAvailable, Mechanism: "explicit-handle-inheritance-list"},
	},
}

// deferredHardenedGuarantees are the six guarantees specified separately by the
// hardened profile of STORY-260728-327soo. None of them may appear in the
// mandatory-control set, the native-control inventory, or an evidence record,
// and the absence of any of them never rejects a portable build.
var deferredHardenedGuarantees = []string{
	"total-network-denial",
	"read-only-source-and-toolchain",
	"private-build-root-only-writes",
	"hard-aggregate-descendant-resource-bounds",
	"exact-executable-allowlisting",
	"fail-closed-capability-preflight",
}

// mandatoryControls is the exact portable control set. Only the absence of one
// of these rejects an operation, with CodeControlUnavailable before the worker
// starts.
var mandatoryControls = []string{
	"fixed-offline-vendored-go",
	"fixed-argument-vectors",
	"fixed-empty-environment",
	"fixed-manager-selected-process-graph",
	"identity-verified-manager-owned-worker",
	"pre-launch-worker-identity-verification",
	"post-exec-identity-reverification",
	"frozen-source-snapshot-integrity",
	"manager-private-staging-roots",
	"manager-derived-output-path",
	"bounded-wall-clock-deadline",
	"bounded-combined-output",
	"bounded-artifact-size",
	"closed-standard-input-and-descriptors",
	"worker-domain-teardown",
	"no-artifact-execution",
	"inventory-native-controls-applied",
	"closed-capability-evidence-record",
}

// ControlProbe is one per-operation availability determination made in the
// manager parent before the worker is launched. A probe performs the exact
// operation the control will perform for this build — the exact job limits, the
// exact per-file byte bound, the exact attribute list — so applicability is
// proved rather than assumed.
type ControlProbe struct {
	Name         string
	Availability string
	Mechanism    string
	ProbedAt     string
}

// controlSeam names one point at which a native control is proved applicable,
// created, or installed. Every seam is crossed in the manager parent, and the
// worker is released to execute only after the last one succeeds, so
// CodeControlUnavailable can never first appear after worker execution begins.
type controlSeam string

const (
	seamProbe   controlSeam = "availability-probe"
	seamPrepare controlSeam = "domain-preparation"
	seamInstall controlSeam = "domain-installation"
)

// controlSeamFault is nil in production. Tests assign it to inject a real
// failure at one seam and prove that the pre-worker failure boundary rejects
// with CodeControlUnavailable, starts no worker and no compiler, claims no
// applied control, and publishes nothing.
var controlSeamFault func(controlSeam) error

func seamFault(seam controlSeam) error {
	if controlSeamFault == nil {
		return nil
	}
	return controlSeamFault(seam)
}

// CapabilityEvidenceEntry is one closed capability-evidence-v1 control entry.
// It contains exactly name, availability, status, and probed_at.
type CapabilityEvidenceEntry struct {
	Name         string `json:"name"`
	Availability string `json:"availability"`
	Status       string `json:"status"`
	ProbedAt     string `json:"probed_at"`
}

// CapabilityEvidence is the single closed capability-evidence-v1 record emitted
// per operation. It is result-only: it never enters a cache key, receipt input,
// install marker, or conformance claim.
type CapabilityEvidence struct {
	RecordVersion   string                    `json:"record_version"`
	ExecutionPolicy string                    `json:"execution_policy"`
	Platform        string                    `json:"platform"`
	Controls        []CapabilityEvidenceEntry `json:"controls"`
}

// MandatoryControls returns the exact portable mandatory-control names.
func MandatoryControls() []string { return append([]string(nil), mandatoryControls...) }

// NativeControlInventory returns the exhaustive inventory control names.
func NativeControlInventory() []string { return append([]string(nil), nativeControlInventory...) }

// DeferredHardenedGuarantees returns the six guarantees the portable profile
// does not provide and must never claim.
func DeferredHardenedGuarantees() []string {
	return append([]string(nil), deferredHardenedGuarantees...)
}

// InventoryPlatform maps a Go GOOS to the inventory platform identifier. An
// unsupported host returns an empty string; the portable policy is defined for
// exactly macOS and Windows.
func InventoryPlatform(goos string) string {
	switch goos {
	case "darwin":
		return PlatformMacOS
	case "windows":
		return PlatformWindows
	default:
		return ""
	}
}

func isDeferredHardenedGuarantee(name string) bool {
	for _, guarantee := range deferredHardenedGuarantees {
		if guarantee == name {
			return true
		}
	}
	return false
}

func inInventory(name string) bool {
	for _, control := range nativeControlInventory {
		if control == name {
			return true
		}
	}
	return false
}

// probeNativeControls determines, once for this operation, which inventory
// controls this host provides for exactly these limits. It is called in the
// parent before the worker exists and starts no program.
func probeNativeControls(limits ResourceLimits) (string, []ControlProbe, error) {
	return probeNativeControlsFor(InventoryPlatform(runtime.GOOS), limits)
}

// probeNativeControlsFor is the platform-parameterized probe. A platform the
// exhaustive inventory does not cover cannot satisfy the mandatory
// inventory-native-controls-applied control, so the operation rejects before
// the worker starts.
func probeNativeControlsFor(platform string, limits ResourceLimits) (string, []ControlProbe, error) {
	if platform == "" || nativeControlPlatforms[platform] == nil {
		return "", nil, diagnostic(CodeControlUnavailable,
			"rc5-native-control-inventory-v1 defines no record for host %s (platform %q); the portable execution policy is specified for macOS and Windows only",
			runtime.GOOS, platform)
	}
	records := nativeControlPlatforms[platform]
	probes := make([]ControlProbe, 0, len(nativeControlInventory))
	for _, name := range nativeControlInventory {
		record := records[name]
		if record.Availability == AvailabilityUnavailable {
			probes = append(probes, ControlProbe{Name: name, Availability: AvailabilityUnavailable, ProbedAt: ProbeTiming})
			continue
		}
		available, err := probeNativeControl(name, limits)
		if err != nil {
			return "", nil, diagnosticErr(CodeControlUnavailable, err, "cannot probe native control %s", name)
		}
		if !available {
			// The inventory marks this control available for the platform, so a
			// host that cannot provide it cannot satisfy
			// inventory-native-controls-applied.
			return "", nil, diagnostic(CodeControlUnavailable, "native control %s is unavailable on this host", name)
		}
		if err := seamFault(seamProbe); err != nil {
			return "", nil, diagnosticErr(CodeControlUnavailable, err,
				"native control %s is not applicable to this operation", name)
		}
		probes = append(probes, ControlProbe{
			Name: name, Availability: AvailabilityAvailable, Mechanism: record.Mechanism, ProbedAt: ProbeTiming,
		})
	}
	return platform, probes, nil
}

// installableControls lists, in inventory order, the controls this operation's
// probes marked available. It is the exact set a control domain must install
// before the worker executes.
func installableControls(probes []ControlProbe) []string {
	names := make([]string, 0, len(probes))
	for _, probe := range probes {
		if probe.Availability == AvailabilityAvailable {
			names = append(names, probe.Name)
		}
	}
	return names
}

// evidenceFromApplied builds the closed record from this operation's probes and
// the controls whose mechanism is actually installed. The manager parent derives
// it only after the control domain is complete, so status applied is never
// reported for a mechanism that has not been installed; the worker derives the
// same record independently from what it confirms, and the parent requires the
// two to be identical.
func evidenceFromApplied(platform string, probes []ControlProbe, applied []string) CapabilityEvidence {
	appliedSet := make(map[string]bool, len(applied))
	for _, name := range applied {
		appliedSet[name] = true
	}
	record := CapabilityEvidence{
		RecordVersion:   CapabilityEvidenceVersion,
		ExecutionPolicy: ExecutionPolicy,
		Platform:        platform,
		Controls:        make([]CapabilityEvidenceEntry, 0, len(probes)),
	}
	for _, probe := range probes {
		status := StatusUnavailable
		if probe.Availability == AvailabilityAvailable && appliedSet[probe.Name] {
			status = StatusApplied
		} else if probe.Availability == AvailabilityAvailable {
			// An available control the worker did not apply is a contradiction;
			// record it faithfully so validation rejects it instead of hiding it.
			status = StatusUnavailable
		}
		record.Controls = append(record.Controls, CapabilityEvidenceEntry{
			Name: probe.Name, Availability: probe.Availability, Status: status, ProbedAt: probe.ProbedAt,
		})
	}
	return record
}

// validateCapabilityEvidence enforces the eight closed consistency rules. The
// probes are this operation's parent-side determination; a record that reports
// an availability the operation did not probe is rejected.
func validateCapabilityEvidence(record CapabilityEvidence, platform string, probes []ControlProbe) error {
	if record.ExecutionPolicy != ExecutionPolicy {
		return diagnostic(CodeHardenedClaimForbidden,
			"capability evidence execution_policy %q is not the portable identity %q", record.ExecutionPolicy, ExecutionPolicy)
	}
	for _, entry := range record.Controls {
		if isDeferredHardenedGuarantee(entry.Name) {
			return diagnostic(CodeHardenedClaimForbidden,
				"capability evidence names deferred hardened guarantee %q", entry.Name)
		}
	}
	if record.RecordVersion != CapabilityEvidenceVersion {
		return diagnostic(CodeCapabilityEvidenceInvalid,
			"unknown capability evidence record_version %q", record.RecordVersion)
	}
	if record.Platform != platform || nativeControlPlatforms[record.Platform] == nil {
		return diagnostic(CodeCapabilityEvidenceInvalid,
			"capability evidence platform %q is not the probed platform %q", record.Platform, platform)
	}
	probed := make(map[string]string, len(probes))
	for _, probe := range probes {
		probed[probe.Name] = probe.Availability
	}
	seen := make(map[string]bool, len(record.Controls))
	for _, entry := range record.Controls {
		if !inInventory(entry.Name) {
			return diagnostic(CodeCapabilityEvidenceInvalid,
				"capability evidence reports control %q outside rc5-native-control-inventory-v1", entry.Name)
		}
		if seen[entry.Name] {
			return diagnostic(CodeCapabilityEvidenceInvalid, "capability evidence duplicates control %q", entry.Name)
		}
		seen[entry.Name] = true
		if entry.ProbedAt != ProbeTiming {
			return diagnostic(CodeCapabilityEvidenceInvalid,
				"capability evidence control %q reports probed_at %q", entry.Name, entry.ProbedAt)
		}
		availability, present := probed[entry.Name]
		if !present || entry.Availability != availability {
			return diagnostic(CodeCapabilityEvidenceInvalid,
				"capability evidence control %q reports availability %q that this operation did not probe", entry.Name, entry.Availability)
		}
		switch entry.Availability {
		case AvailabilityAvailable:
			if entry.Status != StatusApplied {
				return diagnostic(CodeCapabilityEvidenceInvalid,
					"available control %q reports status %q", entry.Name, entry.Status)
			}
		case AvailabilityUnavailable:
			if entry.Status != StatusUnavailable {
				return diagnostic(CodeCapabilityEvidenceInvalid,
					"unavailable control %q reports status %q", entry.Name, entry.Status)
			}
		default:
			return diagnostic(CodeCapabilityEvidenceInvalid,
				"control %q reports availability %q", entry.Name, entry.Availability)
		}
	}
	if len(seen) != len(nativeControlInventory) {
		missing := make([]string, 0, len(nativeControlInventory))
		for _, control := range nativeControlInventory {
			if !seen[control] {
				missing = append(missing, control)
			}
		}
		sort.Strings(missing)
		return diagnostic(CodeCapabilityEvidenceInvalid,
			"capability evidence is missing exactly one entry per inventory control: %v", missing)
	}
	return nil
}
