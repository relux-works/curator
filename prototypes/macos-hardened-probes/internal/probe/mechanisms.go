package probe

import (
	"fmt"
	"sort"
	"strings"

	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/spec"
)

// Mechanism support statuses.
const (
	// StatusSupported means a published, non-deprecated interface.
	StatusSupported = "supported"
	// StatusDeprecated means a shipped interface that Apple documents as
	// deprecated. It may be used for a prototype observation, but a production
	// profile that depends on it is depending on a withdrawn interface.
	StatusDeprecated = "deprecated"
	// StatusPrivate means an SPI, an unpublished profile language, or an
	// entitlement not granted to third parties.
	StatusPrivate = "private"
	// StatusUnavailable means the mechanism does not exist on macOS.
	StatusUnavailable = "unavailable"
	// StatusConditional means the mechanism exists but only under a delivery
	// constraint the caller must state.
	StatusConditional = "conditional"
)

// supportedMechanisms is the platform mechanism inventory this prototype
// examined, with the support status that decides whether a production hardened
// profile could rely on it.
//
// The list is deliberately wider than what the probes exercise: a mechanism
// that was considered and rejected is as much a result as one that was used,
// and leaving it out would make the next implementer re-derive it. Which
// entries this run actually measured is not a property of this list — it is
// filled in by annotateMechanisms from the checks the run recorded, so a
// mechanism can never carry a conclusion wider than the observation behind it.
func supportedMechanisms() []Mechanism {
	return []Mechanism{
		{
			Name:    "/usr/bin/sandbox-exec",
			Status:  StatusDeprecated,
			Classes: []string{spec.ClassNetworkDenial, spec.ClassReadOnlySource, spec.ClassReadOnlyToolchain, spec.ClassWriteConfinement, spec.ClassViewRestriction, spec.ClassExecAllowlist},
			Note:    "shipped and functional on this host, but it is a thin wrapper over sandbox_init, which <sandbox.h> declares deprecated. Apple publishes no replacement for applying a dynamic profile to an arbitrary already-built binary.",
		},
		{
			Name:    "sandbox_init / sandbox_init_with_parameters (libsystem_sandbox)",
			Status:  StatusDeprecated,
			Classes: []string{spec.ClassNetworkDenial, spec.ClassWriteConfinement, spec.ClassExecAllowlist},
			Note:    "declared deprecated in the public SDK header. The named built-in profiles it accepts are coarse; a per-operation profile requires the profile language below.",
		},
		{
			Name:    "seatbelt profile language (version 1 S-expressions)",
			Status:  StatusPrivate,
			Classes: []string{spec.ClassNetworkDenial, spec.ClassReadOnlySource, spec.ClassReadOnlyToolchain, spec.ClassWriteConfinement, spec.ClassViewRestriction, spec.ClassExecAllowlist},
			Note:    "not a published, versioned interface. Operation names and filter semantics are discovered from the shipped profiles under /usr/share/sandbox and can change between releases without notice.",
		},
		{
			Name:    "App Sandbox (com.apple.security.app-sandbox entitlement)",
			Status:  StatusConditional,
			Classes: []string{spec.ClassNetworkDenial, spec.ClassWriteConfinement, spec.ClassViewRestriction},
			Note:    "supported and non-deprecated, but it applies to a signed, entitled bundle. It cannot be imposed on an arbitrary toolchain binary the manager did not build and sign, and its container model does not express an exact executable allowlist.",
		},
		{
			Name:    "Endpoint Security framework",
			Status:  StatusPrivate,
			Classes: []string{spec.ClassExecAllowlist, spec.ClassDomainMembership},
			Note:    "an authorization client could veto exec system-wide, but it requires the com.apple.developer.endpoint-security.client entitlement, which Apple grants case by case, plus a system extension and full disk access. It is also a global authorizer, not a per-operation domain.",
		},
		{
			Name:    "POSIX process group and session",
			Status:  StatusSupported,
			Classes: []string{spec.ClassDomainMembership, spec.ClassDomainTermination},
			Note:    "the only grouping a plain macOS supervisor can create. A descendant leaves it with setsid, so it is not an unescapable domain.",
		},
		{
			Name:    "Supervisor deadline (context cancellation plus a process-group signal)",
			Status:  StatusSupported,
			Classes: []string{spec.ClassAggregateBounds, spec.ClassDomainTermination},
			Note:    "the only way to bound wall-clock time, because no resource limit expresses it. Cancelling a supervised process signals that process alone; the descendants it created need a separate group-directed signal, and a descendant that left the group needs its pid to have been recorded in advance.",
		},
		{
			Name:    "setrlimit RLIMIT_NOFILE",
			Status:  StatusSupported,
			Classes: []string{spec.ClassAggregateBounds},
			Note:    "per process. A child inherits the same soft limit as a fresh budget, so a parent and a child together pass it. The soft limit can be lowered without the hard limit, in which case a domain member can raise it back; lowering the hard limit closes that route but is irreversible for the process that does it.",
		},
		{
			Name:    "setrlimit RLIMIT_CPU",
			Status:  StatusSupported,
			Classes: []string{spec.ClassAggregateBounds},
			Note:    "per process, in whole seconds, delivered as SIGXCPU at the soft limit and SIGKILL at the hard limit. Each descendant receives its own fresh CPU budget under the same declared bound.",
		},
		{
			Name:    "setrlimit RLIMIT_AS (RLIMIT_RSS on Darwin)",
			Status:  StatusConditional,
			Classes: []string{spec.ClassAggregateBounds},
			Note:    "the interface exists and the resource number is defined, but this kernel refuses any soft limit below the address space the calling process has already reserved. A Go process reserves far more than a build budget, so no usable address-space bound can be installed from inside one.",
		},
		{
			Name:    "setrlimit RLIMIT_DATA",
			Status:  StatusConditional,
			Classes: []string{spec.ClassAggregateBounds},
			Note:    "same refusal as RLIMIT_AS on this kernel, and it bounds the data segment rather than mapped memory, so it would not bound an mmap-based allocator even if it installed.",
		},
		{
			Name:    "setrlimit RLIMIT_NPROC",
			Status:  StatusConditional,
			Classes: []string{spec.ClassAggregateBounds},
			Note:    "accounted per real user id, not per operation. The budget is shared with every other process the user already has, so a bound sized for a build domain refuses the domain's first process.",
		},
		{
			Name:    "Mach task ports and task_policy_set",
			Status:  StatusPrivate,
			Classes: []string{spec.ClassAggregateBounds},
			Note:    "task_for_pid on another process requires elevated privilege and is blocked by SIP for protected processes. The policy interfaces set scheduling and QoS bands, not hard aggregate bounds.",
		},
		{
			Name:    "Linux cgroup v2 equivalent",
			Status:  StatusUnavailable,
			Classes: []string{spec.ClassAggregateBounds, spec.ClassDomainMembership, spec.ClassDomainTermination},
			Note:    "macOS has no cgroup-style controller. There is no pids.max, memory.max or cgroup.kill analogue exposed to third-party code.",
		},
		{
			Name:    "Windows Job Object equivalent",
			Status:  StatusUnavailable,
			Classes: []string{spec.ClassDomainMembership, spec.ClassDomainTermination, spec.ClassAggregateBounds},
			Note:    "macOS exposes nothing with JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE semantics: no kernel object that a descendant cannot leave and whose destruction destroys every member.",
		},
		{
			Name:    "Filesystem quotas (edquota, quotacheck)",
			Status:  StatusConditional,
			Classes: []string{spec.ClassAggregateBounds},
			Note:    "per user or group on a whole volume, requires root, and is not enabled on a default APFS install. It cannot bound the bytes one operation writes below one private directory.",
		},
		{
			Name:    "Disk image (hdiutil) as a size-bounded private volume",
			Status:  StatusSupported,
			Classes: []string{spec.ClassAggregateBounds},
			Note:    "a fixed-size attached image does bound bytes written below the build root, which is the one aggregate quantity macOS can enforce without private interfaces. It costs an attach and detach per operation and still does not bound memory, processes, CPU or time.",
		},
		{
			Name:    "Virtualization.framework guest",
			Status:  StatusSupported,
			Classes: []string{spec.ClassDomainMembership, spec.ClassDomainTermination, spec.ClassAggregateBounds, spec.ClassNetworkDenial},
			Note:    "a guest is an unescapable domain with aggregate memory and CPU bounds and atomic destruction. It is the only public macOS mechanism that satisfies the three blocking classes, at the cost of a guest image in the trusted computing base and a much larger per-operation setup.",
		},
		{
			Name:    "chroot",
			Status:  StatusConditional,
			Classes: []string{spec.ClassViewRestriction},
			Note:    "requires root and is not a security boundary on macOS. It does not remove the global path namespace for a process that can create a new root, and it does not bound anything else.",
		},
		{
			Name:    "posix_spawn POSIX_SPAWN_START_SUSPENDED plus a pre-exec sandbox call",
			Status:  StatusDeprecated,
			Classes: []string{spec.ClassNetworkDenial, spec.ClassWriteConfinement},
			Note:    "the usual way to apply a profile without sandbox-exec still calls the deprecated sandbox_init in the child. It changes who calls the interface, not whether the interface is deprecated.",
		},
	}
}

// mechanismEvidence names, for each mechanism this run exercised, the checks
// whose observations are the evidence about it. Keys are "class/check".
//
// A mechanism absent from this map was considered and not exercised, and the
// report says so rather than letting a declared status read like a measurement.
// That distinction is the whole point of the map: without it, a status written
// by hand next to a mechanism nobody probed is indistinguishable from one the
// run established.
func mechanismEvidence() map[string][]string {
	bounds := spec.ClassAggregateBounds + "/"
	return map[string][]string{
		"/usr/bin/sandbox-exec": {
			spec.ClassActiveProbe + "/probe-domain-can-be-created",
		},
		"seatbelt profile language (version 1 S-expressions)": {
			spec.ClassNetworkDenial + "/connect-offhost-tcp",
			spec.ClassWriteConfinement + "/write-absolute-outside",
			spec.ClassExecAllowlist + "/exec-shell",
			spec.ClassViewRestriction + "/readdir-root",
		},
		"POSIX process group and session": {
			spec.ClassDomainMembership + "/detached-descendant-membership",
			spec.ClassDomainTermination + "/detached-descendant-survives",
			spec.ClassDomainTermination + "/attached-descendant-terminated",
		},
		"Supervisor deadline (context cancellation plus a process-group signal)": {
			bounds + "wall-clock:deadline-terminates-the-domain-root",
			bounds + "wall-clock:deadline-reaches-the-attached-descendant",
			bounds + "wall-clock:deadline-reaches-the-detached-descendant",
			bounds + "wall-clock:deadline-cancellation-leaves-no-descendant",
		},
		"setrlimit RLIMIT_NOFILE": {
			bounds + "descriptors:per-process-bound-binds",
			bounds + "descriptors:descendant-shares-the-declared-budget",
		},
		"setrlimit RLIMIT_CPU": {
			bounds + "cpu-milliseconds:bound-can-be-declared",
			bounds + "cpu-milliseconds:bound-binds-the-process-that-set-it",
			bounds + "cpu-milliseconds:descendant-shares-the-declared-budget",
			bounds + "cpu-milliseconds:member-cannot-raise-its-own-bound",
		},
		"setrlimit RLIMIT_AS (RLIMIT_RSS on Darwin)": {
			bounds + "address-space-bytes:bound-can-be-declared",
			bounds + "address-space-bytes:some-bound-is-accepted-by-the-kernel",
		},
		"setrlimit RLIMIT_DATA": {
			bounds + "data-segment-bytes:bound-can-be-declared",
			bounds + "data-segment-bytes:some-bound-is-accepted-by-the-kernel",
		},
		"setrlimit RLIMIT_NPROC": {
			bounds + "process-count:bound-can-be-declared",
			bounds + "process-count:bound-binds-the-process-that-set-it",
			bounds + "process-count:declared-budget-is-available-to-the-domain",
		},
	}
}

// annotateMechanisms fills in what this run measured about each mechanism.
//
// It is the answer to a specific failure mode: an inventory written by hand can
// say more than the probes established, and a reader has no way to tell which
// entries are measurements and which are claims. After this, every entry says
// which it is.
func annotateMechanisms(base []Mechanism, results []ClassResult) []Mechanism {
	observed := map[string]Check{}
	for _, result := range results {
		for _, check := range result.Checks {
			observed[result.Class+"/"+check.Name] = check
		}
	}

	evidence := mechanismEvidence()
	out := make([]Mechanism, 0, len(base))
	for _, mechanism := range base {
		keys, named := evidence[mechanism.Name]
		if !named {
			mechanism.Observation = "not exercised by this run; the status above is a reading of the published interface, not a measurement"
			out = append(out, mechanism)
			continue
		}
		var parts []string
		for _, key := range keys {
			check, ok := observed[key]
			if !ok {
				parts = append(parts, fmt.Sprintf("%s=not-recorded", shortKey(key)))
				continue
			}
			parts = append(parts, fmt.Sprintf("%s=%s (%s)", shortKey(key), check.Observed, passWord(check.Pass)))
		}
		if len(parts) == 0 {
			mechanism.Observation = "no observation was recorded for this mechanism in this run"
			out = append(out, mechanism)
			continue
		}
		mechanism.Exercised = true
		mechanism.Observation = strings.Join(parts, "; ")
		out = append(out, mechanism)
	}
	return out
}

// shortKey drops the class prefix, which is already implied by the check name.
func shortKey(key string) string {
	if i := strings.Index(key, "/"); i >= 0 {
		return key[i+1:]
	}
	return key
}

func passWord(pass bool) string {
	if pass {
		return "expectation held"
	}
	return "expectation did not hold"
}

// UnexercisedMechanisms lists the mechanisms the inventory names but this run
// did not measure. It exists so a reader, and the tests, can see the boundary of
// what the evidence covers without reading every entry.
func UnexercisedMechanisms(mechanisms []Mechanism) []string {
	var out []string
	for _, mechanism := range mechanisms {
		if !mechanism.Exercised {
			out = append(out, mechanism.Name)
		}
	}
	sort.Strings(out)
	return out
}
