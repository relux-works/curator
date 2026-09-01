// Package evidence builds and validates the closed hardened capability-evidence
// record of hardened-execution.md section 6.4.
//
// The record is deliberately closed: it carries exactly eleven fields, exactly
// one capability entry per class of the exhaustive inventory, and exactly one
// guarantee entry per guarantee. Validate implements the section 6.4 error
// table so a malformed record is a diagnostic, never a permitted variation.
package evidence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/spec"
)

// Capability is one entry of the record's capabilities array. It contains
// exactly name, availability, status and probed_at.
type Capability struct {
	Name         string `json:"name"`
	Availability string `json:"availability"`
	Status       string `json:"status"`
	ProbedAt     string `json:"probed_at"`
}

// Guarantee is one entry of the record's guarantees array. It contains exactly
// name and established.
type Guarantee struct {
	Name        string `json:"name"`
	Established bool   `json:"established"`
}

// Record is the closed hardened capability-evidence record. Field order matches
// the specification example so the emitted JSON reads like the normative form.
type Record struct {
	RecordVersion       string       `json:"record_version"`
	HardenedProfile     string       `json:"hardened_profile"`
	ExecutionPolicy     string       `json:"execution_policy"`
	Platform            string       `json:"platform"`
	EnforcementBackend  string       `json:"enforcement_backend"`
	QualificationStatus string       `json:"qualification_status"`
	Outcome             string       `json:"outcome"`
	RejectedBefore      *string      `json:"rejected_before"`
	Diagnostic          *string      `json:"diagnostic"`
	Capabilities        []Capability `json:"capabilities"`
	Guarantees          []Guarantee  `json:"guarantees"`
}

// Observation is the probe result for one capability class, reduced to the two
// values the closed record can carry.
type Observation struct {
	Availability string
	Applied      bool
}

// Build assembles a record from per-class observations.
//
// Build never invents an availability: a class missing from observations is
// reported unprobed and not-applied, which makes the record reject rather than
// silently claim a capability nobody measured. A guarantee is established only
// when every class mapped to it by section 6.2 is applied.
func Build(platform, backend, qualification string, observations map[string]Observation) Record {
	rec := Record{
		RecordVersion:       spec.RecordVersion,
		HardenedProfile:     spec.HardenedProfile,
		ExecutionPolicy:     spec.ExecutionPolicy,
		Platform:            platform,
		EnforcementBackend:  backend,
		QualificationStatus: qualification,
	}

	applied := map[string]bool{}
	for _, class := range spec.Classes() {
		obs, ok := observations[class]
		if !ok {
			obs = Observation{Availability: spec.AvailabilityUnprobed}
		}
		status := spec.StatusNotApplied
		availability := obs.Availability
		if obs.Applied && availability == spec.AvailabilityAvailable {
			status = spec.StatusApplied
		}
		// The closed record cannot say "available, but this run did not apply
		// it": section 6.4 requires available to imply applied. A class the host
		// could provide but that was not installed in this operation is therefore
		// reduced to unavailable, which is the same fail-closed direction section
		// 5.6 requires of an inconclusive probe. Reducing the other way would
		// claim a control that was never installed.
		if status == spec.StatusNotApplied && availability == spec.AvailabilityAvailable {
			availability = spec.AvailabilityUnavailable
		}
		applied[class] = status == spec.StatusApplied
		rec.Capabilities = append(rec.Capabilities, Capability{
			Name:         class,
			Availability: availability,
			Status:       status,
			ProbedAt:     spec.ProbedAt,
		})
	}

	allEstablished := true
	for _, guarantee := range spec.Guarantees() {
		established := true
		for _, class := range spec.GuaranteeClasses(guarantee) {
			if !applied[class] {
				established = false
				break
			}
		}
		if !established {
			allEstablished = false
		}
		rec.Guarantees = append(rec.Guarantees, Guarantee{Name: guarantee, Established: established})
	}

	if allEstablished {
		rec.Outcome = spec.OutcomeEstablished
		return rec
	}
	rec.Outcome = spec.OutcomeRejected
	rec.setRejection(spec.PhaseCapabilityProbe, spec.DiagCapabilityUnavailable)
	return rec
}

func (r *Record) setRejection(phase, diagnostic string) {
	p, d := phase, diagnostic
	r.RejectedBefore = &p
	r.Diagnostic = &d
}

// Reject overrides the record's outcome with an explicit pre-domain-entry
// rejection. It is used when a phase earlier than capability-probe refuses, for
// example platform-qualification on an unqualified host.
func (r *Record) Reject(phase, diagnostic string) {
	r.Outcome = spec.OutcomeRejected
	r.setRejection(phase, diagnostic)
}

// UnavailableClasses lists, in normative order, the classes that are not
// applied. It is the operator-facing reason a rejection happened.
func (r Record) UnavailableClasses() []string {
	var out []string
	for _, capability := range r.Capabilities {
		if capability.Status != spec.StatusApplied {
			out = append(out, capability.Name)
		}
	}
	return out
}

// UnestablishedGuarantees lists, in normative order, the guarantees that were
// not established.
func (r Record) UnestablishedGuarantees() []string {
	var out []string
	for _, guarantee := range r.Guarantees {
		if !guarantee.Established {
			out = append(out, guarantee.Name)
		}
	}
	return out
}

// Validate implements the section 6.4 error table. It returns the mapped stable
// diagnostic and a human-readable reason, or an empty diagnostic when the
// record is well formed.
func (r Record) Validate() (diagnostic string, reason string) {
	if r.RecordVersion != spec.RecordVersion {
		return spec.DiagEvidenceInvalid, fmt.Sprintf("unknown record_version %q", r.RecordVersion)
	}
	if r.HardenedProfile != spec.HardenedProfile {
		return spec.DiagProfileClaimForbidden, fmt.Sprintf("hardened_profile %q is not %q", r.HardenedProfile, spec.HardenedProfile)
	}
	if r.ExecutionPolicy != spec.ExecutionPolicy {
		return spec.DiagProfileClaimForbidden, fmt.Sprintf("execution_policy %q is not %q", r.ExecutionPolicy, spec.ExecutionPolicy)
	}

	if diag, why := r.validateCapabilities(); diag != "" {
		return diag, why
	}
	if diag, why := r.validateGuarantees(); diag != "" {
		return diag, why
	}
	return r.validateOutcome()
}

func (r Record) validateCapabilities() (string, string) {
	seen := map[string]bool{}
	for _, capability := range r.Capabilities {
		if !spec.IsClass(capability.Name) {
			return spec.DiagEvidenceInvalid, fmt.Sprintf("unknown capability class %q", capability.Name)
		}
		if seen[capability.Name] {
			return spec.DiagEvidenceInvalid, fmt.Sprintf("duplicated capability class %q", capability.Name)
		}
		seen[capability.Name] = true

		switch capability.Availability {
		case spec.AvailabilityAvailable, spec.AvailabilityUnavailable, spec.AvailabilityUnprobed:
		default:
			return spec.DiagEvidenceInvalid, fmt.Sprintf("capability %q: unknown availability %q", capability.Name, capability.Availability)
		}
		switch capability.Status {
		case spec.StatusApplied, spec.StatusNotApplied:
		default:
			return spec.DiagEvidenceInvalid, fmt.Sprintf("capability %q: unknown status %q", capability.Name, capability.Status)
		}
		if capability.ProbedAt != spec.ProbedAt {
			return spec.DiagEvidenceInvalid, fmt.Sprintf("capability %q: probed_at %q is not %q", capability.Name, capability.ProbedAt, spec.ProbedAt)
		}
		if capability.Availability == spec.AvailabilityAvailable && capability.Status != spec.StatusApplied {
			return spec.DiagEvidenceInvalid, fmt.Sprintf("capability %q: available but %s", capability.Name, capability.Status)
		}
		if capability.Availability != spec.AvailabilityAvailable && capability.Status == spec.StatusApplied {
			return spec.DiagEvidenceInvalid, fmt.Sprintf("capability %q: %s but applied", capability.Name, capability.Availability)
		}
	}
	for _, class := range spec.Classes() {
		if !seen[class] {
			return spec.DiagEvidenceInvalid, fmt.Sprintf("missing capability class %q", class)
		}
	}
	return "", ""
}

func (r Record) validateGuarantees() (string, string) {
	applied := map[string]bool{}
	for _, capability := range r.Capabilities {
		applied[capability.Name] = capability.Status == spec.StatusApplied
	}

	seen := map[string]bool{}
	for _, guarantee := range r.Guarantees {
		if !spec.IsGuarantee(guarantee.Name) {
			return spec.DiagEvidenceInvalid, fmt.Sprintf("unknown guarantee %q", guarantee.Name)
		}
		if seen[guarantee.Name] {
			return spec.DiagEvidenceInvalid, fmt.Sprintf("duplicated guarantee %q", guarantee.Name)
		}
		seen[guarantee.Name] = true

		if !guarantee.Established {
			continue
		}
		for _, class := range spec.GuaranteeClasses(guarantee.Name) {
			if !applied[class] {
				return spec.DiagEvidenceInvalid, fmt.Sprintf("guarantee %q established while class %q is not applied", guarantee.Name, class)
			}
		}
	}
	for _, guarantee := range spec.Guarantees() {
		if !seen[guarantee] {
			return spec.DiagEvidenceInvalid, fmt.Sprintf("missing guarantee %q", guarantee)
		}
	}
	return "", ""
}

func (r Record) validateOutcome() (string, string) {
	switch r.Outcome {
	case spec.OutcomeEstablished:
		if r.RejectedBefore != nil || r.Diagnostic != nil {
			return spec.DiagEvidenceInvalid, "established outcome carries rejected_before or diagnostic"
		}
		for _, capability := range r.Capabilities {
			if capability.Status != spec.StatusApplied {
				return spec.DiagEvidenceInvalid, fmt.Sprintf("established outcome while capability %q is %s", capability.Name, capability.Status)
			}
		}
		for _, guarantee := range r.Guarantees {
			if !guarantee.Established {
				return spec.DiagEvidenceInvalid, fmt.Sprintf("established outcome while guarantee %q is not established", guarantee.Name)
			}
		}
	case spec.OutcomeRejected:
		if r.RejectedBefore == nil || r.Diagnostic == nil {
			return spec.DiagEvidenceInvalid, "rejected outcome without both rejected_before and diagnostic"
		}
		if !spec.IsPhase(*r.RejectedBefore) {
			return spec.DiagEvidenceInvalid, fmt.Sprintf("rejected_before %q is not a known phase", *r.RejectedBefore)
		}
		if !spec.IsDiagnostic(*r.Diagnostic) {
			return spec.DiagEvidenceInvalid, fmt.Sprintf("diagnostic %q is not a stable hardened diagnostic", *r.Diagnostic)
		}
	default:
		return spec.DiagEvidenceInvalid, fmt.Sprintf("unknown outcome %q", r.Outcome)
	}
	return "", ""
}

// closedFields is the exact field set of the record. Anything else in a decoded
// document is an extra field, which section 6.4 treats as an error.
var closedFields = []string{
	"record_version", "hardened_profile", "execution_policy", "platform",
	"enforcement_backend", "qualification_status", "outcome", "rejected_before",
	"diagnostic", "capabilities", "guarantees",
}

// Decode parses a record and rejects unknown or missing top-level fields before
// running Validate. It exists so a record produced elsewhere can be checked
// against the closed shape, not only records this package built.
func Decode(data []byte) (Record, string, string) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return Record{}, spec.DiagEvidenceInvalid, fmt.Sprintf("record is not a JSON object: %v", err)
	}
	known := map[string]bool{}
	for _, field := range closedFields {
		known[field] = true
	}
	var extra []string
	for field := range raw {
		if !known[field] {
			extra = append(extra, field)
		}
	}
	if len(extra) > 0 {
		sort.Strings(extra)
		return Record{}, spec.DiagEvidenceInvalid, fmt.Sprintf("extra field(s) %v", extra)
	}
	for _, field := range closedFields {
		if _, ok := raw[field]; !ok {
			return Record{}, spec.DiagEvidenceInvalid, fmt.Sprintf("missing field %q", field)
		}
	}

	var rec Record
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&rec); err != nil {
		return Record{}, spec.DiagEvidenceInvalid, fmt.Sprintf("record does not match the closed shape: %v", err)
	}
	diag, reason := rec.Validate()
	return rec, diag, reason
}

// JSON renders the record as indented JSON with a trailing newline.
func (r Record) JSON() []byte {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		// Record contains only strings, bools and slices of the same; marshal
		// cannot fail. Panicking here would hide a programming error behind a
		// partial artifact, so surface it as an obviously broken document.
		return []byte(fmt.Sprintf("{\"record_version\":%q,\"marshal_error\":%q}\n", spec.RecordVersion, err.Error()))
	}
	return append(data, '\n')
}
