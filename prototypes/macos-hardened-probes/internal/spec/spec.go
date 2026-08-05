// Package spec mirrors the normative constants of the Curator Hardened
// Execution Profile 1.0 (curator-spec protocol/hardened-execution.md) that the
// macOS capability probes are measured against.
//
// This package is a prototype-side transcription, not the normative source.
// The authority remains curator-spec; SpecRevision records which revision the
// transcription was taken from so a drift check has something to compare.
package spec

// SpecRevision names the hardened-profile candidate this transcription follows.
const SpecRevision = "hardened-1.0.0-rc.1"

// Closed record identifiers of hardened-execution.md section 6.4.
const (
	RecordVersion   = "hardened-capability-evidence-v1"
	HardenedProfile = "hardened-profile-v1"
	ExecutionPolicy = "hardened-worker-v1"

	InventoryVersion = "hardened-capability-inventory-v1"
)

// Platform declarations of hardened-execution.md section 6.3 for macOS.
const (
	PlatformMacOS            = "macos"
	BackendMacOSSandbox      = "macos-sandbox-v1"
	QualificationUnqualified = "unqualified"
	QualificationQualified   = "qualified"
)

// Capability class names of hardened-execution.md section 6.1. The inventory is
// exhaustive: a manager probes exactly these classes, and reports exactly one
// entry per class.
const (
	ClassDomainMembership   = "domain-membership-enforcement"
	ClassDomainTermination  = "domain-atomic-termination"
	ClassNetworkDenial      = "network-syscall-denial"
	ClassEndpointRevocation = "preexisting-endpoint-revocation"
	ClassReadOnlySource     = "read-only-source-view"
	ClassReadOnlyToolchain  = "read-only-toolchain-view"
	ClassWriteConfinement   = "write-path-confinement"
	ClassViewRestriction    = "filesystem-view-restriction"
	ClassExecAllowlist      = "exec-path-allowlist"
	ClassAggregateBounds    = "aggregate-resource-bounds"
	ClassActiveProbe        = "active-capability-probe"
)

// Guarantee names of hardened-execution.md section 5.
const (
	GuaranteeNetworkDenial       = "total-network-denial"
	GuaranteeReadOnlyInputs      = "read-only-source-and-toolchain"
	GuaranteePrivateWrites       = "private-build-root-only-writes"
	GuaranteeResourceBounds      = "hard-aggregate-descendant-resource-bounds"
	GuaranteeExecAllowlist       = "exact-executable-allowlisting"
	GuaranteeFailClosedPreflight = "fail-closed-capability-preflight"
)

// Classes is the exhaustive capability-class inventory in normative order.
func Classes() []string {
	return []string{
		ClassDomainMembership,
		ClassDomainTermination,
		ClassNetworkDenial,
		ClassEndpointRevocation,
		ClassReadOnlySource,
		ClassReadOnlyToolchain,
		ClassWriteConfinement,
		ClassViewRestriction,
		ClassExecAllowlist,
		ClassAggregateBounds,
		ClassActiveProbe,
	}
}

// Guarantees is the guarantee list in normative order.
func Guarantees() []string {
	return []string{
		GuaranteeNetworkDenial,
		GuaranteeReadOnlyInputs,
		GuaranteePrivateWrites,
		GuaranteeResourceBounds,
		GuaranteeExecAllowlist,
		GuaranteeFailClosedPreflight,
	}
}

// GuaranteeClasses is the normative, exhaustive guarantee-to-class mapping of
// hardened-execution.md section 6.2. A guarantee is established only when every
// class mapped to it is available and applied.
func GuaranteeClasses(guarantee string) []string {
	switch guarantee {
	case GuaranteeNetworkDenial:
		return []string{ClassNetworkDenial, ClassEndpointRevocation, ClassDomainMembership}
	case GuaranteeReadOnlyInputs:
		return []string{ClassReadOnlySource, ClassReadOnlyToolchain, ClassViewRestriction}
	case GuaranteePrivateWrites:
		return []string{ClassWriteConfinement, ClassViewRestriction, ClassDomainMembership}
	case GuaranteeResourceBounds:
		return []string{ClassAggregateBounds, ClassDomainMembership, ClassDomainTermination}
	case GuaranteeExecAllowlist:
		return []string{ClassExecAllowlist, ClassDomainMembership}
	case GuaranteeFailClosedPreflight:
		return []string{ClassActiveProbe}
	default:
		return nil
	}
}

// Availability values of the closed record.
const (
	AvailabilityAvailable   = "available"
	AvailabilityUnavailable = "unavailable"
	AvailabilityUnprobed    = "unprobed"
)

// Applied status values of the closed record.
const (
	StatusApplied    = "applied"
	StatusNotApplied = "not-applied"
)

// ProbedAt is the only permitted probe point.
const ProbedAt = "pre-domain-entry"

// Outcome values of the closed record.
const (
	OutcomeEstablished = "established"
	OutcomeRejected    = "rejected"
)

// Ordered phase names of hardened-execution.md section 7.2. Only the phases up
// to and including capability-probe are exercised by this prototype; the rest
// are listed so rejected_before can name a real phase.
const (
	PhaseProfileSelection = "profile-selection"
	PhasePlatformQual     = "platform-qualification"
	PhaseCapabilityProbe  = "capability-probe"
	PhaseToolchainFreeze  = "toolchain-probe-and-snapshot-freeze"
	PhaseTCBVerification  = "tcb-identity-verification"
	PhaseCacheLookup      = "build-input-and-cache-lookup"
	PhaseDomainEstablish  = "domain-establishment"
	PhaseDomainEntry      = "domain-entry"
	PhaseSelfTest         = "in-domain-guarantee-self-test"
)

// Phases returns the ordered phase names this prototype knows about, in
// normative order.
func Phases() []string {
	return []string{
		PhaseProfileSelection,
		PhasePlatformQual,
		PhaseCapabilityProbe,
		PhaseToolchainFreeze,
		PhaseTCBVerification,
		PhaseCacheLookup,
		PhaseDomainEstablish,
		PhaseDomainEntry,
		PhaseSelfTest,
	}
}

// Stable hardened diagnostics of hardened-execution.md section 10.
const (
	DiagProfileUnsupported    = "hardened_profile_unsupported"
	DiagCapabilityUnavailable = "hardened_capability_unavailable"
	DiagTCBIdentityInvalid    = "hardened_tcb_identity_invalid"
	DiagDomainEstablishFailed = "hardened_domain_establishment_failed"
	DiagDomainProtocolInvalid = "hardened_domain_protocol_invalid"
	DiagDomainBreachDetected  = "hardened_domain_breach_detected"
	DiagEvidenceInvalid       = "hardened_evidence_invalid"
	DiagProfileClaimForbidden = "hardened_profile_claim_forbidden"
	DiagPackageInfluence      = "hardened_package_influence_forbidden"
)

// Diagnostics lists the stable hardened diagnostics.
func Diagnostics() []string {
	return []string{
		DiagProfileUnsupported,
		DiagCapabilityUnavailable,
		DiagTCBIdentityInvalid,
		DiagDomainEstablishFailed,
		DiagDomainProtocolInvalid,
		DiagDomainBreachDetected,
		DiagEvidenceInvalid,
		DiagProfileClaimForbidden,
		DiagPackageInfluence,
	}
}

// IsClass reports whether name is a member of the exhaustive inventory.
func IsClass(name string) bool { return contains(Classes(), name) }

// IsGuarantee reports whether name is a member of the guarantee set.
func IsGuarantee(name string) bool { return contains(Guarantees(), name) }

// IsDiagnostic reports whether name is a stable hardened diagnostic.
func IsDiagnostic(name string) bool { return contains(Diagnostics(), name) }

// IsPhase reports whether name is a phase this prototype knows about.
func IsPhase(name string) bool { return contains(Phases(), name) }

func contains(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}
