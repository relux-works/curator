package evidence

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/spec"
)

// allApplied is the only observation set that may produce an established
// outcome. Every other test starts from it and removes exactly one thing, so a
// failure names the single property that was violated.
func allApplied() map[string]Observation {
	obs := map[string]Observation{}
	for _, class := range spec.Classes() {
		obs[class] = Observation{Availability: spec.AvailabilityAvailable, Applied: true}
	}
	return obs
}

func buildEstablished(t *testing.T) Record {
	t.Helper()
	rec := Build(spec.PlatformMacOS, spec.BackendMacOSSandbox, spec.QualificationQualified, allApplied())
	if rec.Outcome != spec.OutcomeEstablished {
		t.Fatalf("fixture is not established: outcome %q", rec.Outcome)
	}
	if diag, why := rec.Validate(); diag != "" {
		t.Fatalf("fixture does not validate: %s: %s", diag, why)
	}
	return rec
}

func capabilityIndex(t *testing.T, rec Record, class string) int {
	t.Helper()
	for i, capability := range rec.Capabilities {
		if capability.Name == class {
			return i
		}
	}
	t.Fatalf("record has no capability %q", class)
	return -1
}

func guaranteeIndex(t *testing.T, rec Record, name string) int {
	t.Helper()
	for i, guarantee := range rec.Guarantees {
		if guarantee.Name == name {
			return i
		}
	}
	t.Fatalf("record has no guarantee %q", name)
	return -1
}

// ------------------------------------------------------------------- Build

func TestBuildEstablishedCarriesNoRejection(t *testing.T) {
	rec := buildEstablished(t)
	if rec.RejectedBefore != nil || rec.Diagnostic != nil {
		t.Errorf("established record carries rejected_before=%v diagnostic=%v", rec.RejectedBefore, rec.Diagnostic)
	}
	if rec.RecordVersion != spec.RecordVersion ||
		rec.HardenedProfile != spec.HardenedProfile ||
		rec.ExecutionPolicy != spec.ExecutionPolicy {
		t.Errorf("closed identifiers not stamped: %+v", rec)
	}
	if len(rec.Capabilities) != len(spec.Classes()) {
		t.Errorf("record has %d capabilities, want %d", len(rec.Capabilities), len(spec.Classes()))
	}
	if len(rec.Guarantees) != len(spec.Guarantees()) {
		t.Errorf("record has %d guarantees, want %d", len(rec.Guarantees), len(spec.Guarantees()))
	}
	if len(rec.UnavailableClasses()) != 0 || len(rec.UnestablishedGuarantees()) != 0 {
		t.Errorf("established record reports gaps: classes=%v guarantees=%v",
			rec.UnavailableClasses(), rec.UnestablishedGuarantees())
	}
}

// A class nobody measured must never be reported available. This is the single
// property that keeps an unmeasured host from claiming a guarantee.
func TestBuildTreatsUnobservedClassAsUnprobed(t *testing.T) {
	rec := Build(spec.PlatformMacOS, spec.BackendMacOSSandbox, spec.QualificationUnqualified, nil)
	for _, capability := range rec.Capabilities {
		if capability.Availability != spec.AvailabilityUnprobed {
			t.Errorf("capability %q with no observation is %q, want %q",
				capability.Name, capability.Availability, spec.AvailabilityUnprobed)
		}
		if capability.Status != spec.StatusNotApplied {
			t.Errorf("capability %q with no observation is %q, want %q",
				capability.Name, capability.Status, spec.StatusNotApplied)
		}
		if capability.ProbedAt != spec.ProbedAt {
			t.Errorf("capability %q probed_at %q, want %q", capability.Name, capability.ProbedAt, spec.ProbedAt)
		}
	}
	if rec.Outcome != spec.OutcomeRejected {
		t.Fatalf("outcome %q, want %q", rec.Outcome, spec.OutcomeRejected)
	}
	if rec.RejectedBefore == nil || *rec.RejectedBefore != spec.PhaseCapabilityProbe {
		t.Errorf("rejected_before %v, want %q", rec.RejectedBefore, spec.PhaseCapabilityProbe)
	}
	if rec.Diagnostic == nil || *rec.Diagnostic != spec.DiagCapabilityUnavailable {
		t.Errorf("diagnostic %v, want %q", rec.Diagnostic, spec.DiagCapabilityUnavailable)
	}
	if len(rec.UnestablishedGuarantees()) != len(spec.Guarantees()) {
		t.Errorf("unestablished guarantees %v, want all of them", rec.UnestablishedGuarantees())
	}
}

// applied is a statement about this run, not about the platform, so it is only
// ever granted when the class was both available and installed. The awkward
// case is "available, but this run did not install it": the closed record has
// no way to say that, so it must reduce to unavailable rather than emit a
// record that fails its own Validate.
func TestBuildAppliedRequiresAvailable(t *testing.T) {
	cases := []struct {
		name             string
		obs              Observation
		wantAvailability string
		wantStatus       string
	}{
		{
			"available and applied",
			Observation{Availability: spec.AvailabilityAvailable, Applied: true},
			spec.AvailabilityAvailable, spec.StatusApplied,
		},
		{
			"available but not applied reduces to unavailable",
			Observation{Availability: spec.AvailabilityAvailable, Applied: false},
			spec.AvailabilityUnavailable, spec.StatusNotApplied,
		},
		{
			"unavailable but claimed applied",
			Observation{Availability: spec.AvailabilityUnavailable, Applied: true},
			spec.AvailabilityUnavailable, spec.StatusNotApplied,
		},
		{
			"unprobed but claimed applied",
			Observation{Availability: spec.AvailabilityUnprobed, Applied: true},
			spec.AvailabilityUnprobed, spec.StatusNotApplied,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			obs := allApplied()
			obs[spec.ClassNetworkDenial] = tc.obs
			rec := Build(spec.PlatformMacOS, spec.BackendMacOSSandbox, spec.QualificationUnqualified, obs)
			got := rec.Capabilities[capabilityIndex(t, rec, spec.ClassNetworkDenial)]
			if got.Availability != tc.wantAvailability {
				t.Errorf("availability %q, want %q", got.Availability, tc.wantAvailability)
			}
			if got.Status != tc.wantStatus {
				t.Errorf("status %q, want %q", got.Status, tc.wantStatus)
			}
			// The point of the reduction: whatever the observation said, the
			// emitted record must survive its own error table.
			if diag, why := rec.Validate(); diag != "" {
				t.Errorf("built record does not validate: %s: %s", diag, why)
			}
		})
	}
}

// Build must never hand probe.Run a record that Validate refuses; that path
// reports "the harness broke" (exit 2) instead of the fail-closed "the host
// rejected" (exit 1), which is the wrong answer about the host.
func TestBuildOutputAlwaysValidates(t *testing.T) {
	availabilities := []string{spec.AvailabilityAvailable, spec.AvailabilityUnavailable, spec.AvailabilityUnprobed, ""}
	for _, availability := range availabilities {
		for _, applied := range []bool{true, false} {
			for _, class := range spec.Classes() {
				obs := allApplied()
				obs[class] = Observation{Availability: availability, Applied: applied}
				rec := Build(spec.PlatformMacOS, spec.BackendMacOSSandbox, spec.QualificationUnqualified, obs)
				diag, why := rec.Validate()
				if availability == "" {
					// An empty availability is not a value of the closed record;
					// Build passes it through and Validate is expected to catch it.
					if diag != spec.DiagEvidenceInvalid {
						t.Errorf("%s: empty availability accepted (diag %q)", class, diag)
					}
					continue
				}
				if diag != "" {
					t.Errorf("Build(%s=%s/applied=%v) produced an invalid record: %s: %s",
						class, availability, applied, diag, why)
				}
			}
		}
	}
}

// Dropping one class must unestablish exactly the guarantees that map to it and
// leave the others alone. This is the guarantee-to-class mapping under test, not
// just the aggregate rejection.
func TestBuildGuaranteeEstablishmentFollowsTheMapping(t *testing.T) {
	for _, class := range spec.Classes() {
		t.Run(class, func(t *testing.T) {
			obs := allApplied()
			obs[class] = Observation{Availability: spec.AvailabilityUnavailable}
			rec := Build(spec.PlatformMacOS, spec.BackendMacOSSandbox, spec.QualificationUnqualified, obs)

			for _, guarantee := range spec.Guarantees() {
				wantEstablished := true
				for _, mapped := range spec.GuaranteeClasses(guarantee) {
					if mapped == class {
						wantEstablished = false
					}
				}
				got := rec.Guarantees[guaranteeIndex(t, rec, guarantee)].Established
				if got != wantEstablished {
					t.Errorf("with %q unavailable, guarantee %q established=%v, want %v",
						class, guarantee, got, wantEstablished)
				}
			}
			if rec.Outcome != spec.OutcomeRejected {
				t.Errorf("outcome %q, want %q", rec.Outcome, spec.OutcomeRejected)
			}
			if diag, why := rec.Validate(); diag != "" {
				t.Errorf("built record does not validate: %s: %s", diag, why)
			}
		})
	}
}

func TestBuildOrdersEntriesNormatively(t *testing.T) {
	rec := buildEstablished(t)
	classes := spec.Classes()
	for i, capability := range rec.Capabilities {
		if capability.Name != classes[i] {
			t.Errorf("capability %d is %q, want %q", i, capability.Name, classes[i])
		}
	}
	guarantees := spec.Guarantees()
	for i, guarantee := range rec.Guarantees {
		if guarantee.Name != guarantees[i] {
			t.Errorf("guarantee %d is %q, want %q", i, guarantee.Name, guarantees[i])
		}
	}
}

func TestUnavailableClassesReportsInNormativeOrder(t *testing.T) {
	obs := allApplied()
	obs[spec.ClassActiveProbe] = Observation{Availability: spec.AvailabilityUnavailable}
	obs[spec.ClassNetworkDenial] = Observation{Availability: spec.AvailabilityUnavailable}
	rec := Build(spec.PlatformMacOS, spec.BackendMacOSSandbox, spec.QualificationUnqualified, obs)

	got := rec.UnavailableClasses()
	want := []string{spec.ClassNetworkDenial, spec.ClassActiveProbe}
	if len(got) != len(want) {
		t.Fatalf("UnavailableClasses() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("UnavailableClasses()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestRejectOverridesAnEstablishedOutcome(t *testing.T) {
	rec := buildEstablished(t)
	rec.Reject(spec.PhasePlatformQual, spec.DiagProfileUnsupported)

	if rec.Outcome != spec.OutcomeRejected {
		t.Errorf("outcome %q, want %q", rec.Outcome, spec.OutcomeRejected)
	}
	if rec.RejectedBefore == nil || *rec.RejectedBefore != spec.PhasePlatformQual {
		t.Errorf("rejected_before %v, want %q", rec.RejectedBefore, spec.PhasePlatformQual)
	}
	if rec.Diagnostic == nil || *rec.Diagnostic != spec.DiagProfileUnsupported {
		t.Errorf("diagnostic %v, want %q", rec.Diagnostic, spec.DiagProfileUnsupported)
	}
	// A rejection at an earlier phase is still a well-formed record: the
	// capabilities it carries were genuinely applied, they just do not matter.
	if diag, why := rec.Validate(); diag != "" {
		t.Errorf("rejected record does not validate: %s: %s", diag, why)
	}
}

// ---------------------------------------------------------------- Validate

func TestValidateAcceptsBuiltRecords(t *testing.T) {
	for _, obs := range []map[string]Observation{allApplied(), nil, {}} {
		rec := Build(spec.PlatformMacOS, spec.BackendMacOSSandbox, spec.QualificationUnqualified, obs)
		if diag, why := rec.Validate(); diag != "" {
			t.Errorf("Build output does not validate: %s: %s", diag, why)
		}
	}
}

func TestValidateRejectsTamperedRecords(t *testing.T) {
	cases := []struct {
		name       string
		tamper     func(*Record)
		wantDiag   string
		wantReason string
	}{
		{
			name:       "unknown record version",
			tamper:     func(r *Record) { r.RecordVersion = "hardened-capability-evidence-v2" },
			wantDiag:   spec.DiagEvidenceInvalid,
			wantReason: "unknown record_version",
		},
		{
			name:       "foreign hardened profile",
			tamper:     func(r *Record) { r.HardenedProfile = "hardened-profile-v2" },
			wantDiag:   spec.DiagProfileClaimForbidden,
			wantReason: "hardened_profile",
		},
		{
			name:       "foreign execution policy",
			tamper:     func(r *Record) { r.ExecutionPolicy = "direct-go-v1" },
			wantDiag:   spec.DiagProfileClaimForbidden,
			wantReason: "execution_policy",
		},
		{
			name:       "unknown capability class",
			tamper:     func(r *Record) { r.Capabilities[0].Name = "cgroup-kill" },
			wantDiag:   spec.DiagEvidenceInvalid,
			wantReason: "unknown capability class",
		},
		{
			name: "duplicated capability class",
			tamper: func(r *Record) {
				r.Capabilities = append(r.Capabilities, r.Capabilities[0])
			},
			wantDiag:   spec.DiagEvidenceInvalid,
			wantReason: "duplicated capability class",
		},
		{
			name: "missing capability class",
			tamper: func(r *Record) {
				r.Capabilities = r.Capabilities[1:]
			},
			wantDiag:   spec.DiagEvidenceInvalid,
			wantReason: "missing capability class",
		},
		{
			name:       "unknown availability",
			tamper:     func(r *Record) { r.Capabilities[0].Availability = "probably" },
			wantDiag:   spec.DiagEvidenceInvalid,
			wantReason: "unknown availability",
		},
		{
			name:       "unknown status",
			tamper:     func(r *Record) { r.Capabilities[0].Status = "partially-applied" },
			wantDiag:   spec.DiagEvidenceInvalid,
			wantReason: "unknown status",
		},
		{
			name:       "probed at the wrong point",
			tamper:     func(r *Record) { r.Capabilities[0].ProbedAt = "post-domain-entry" },
			wantDiag:   spec.DiagEvidenceInvalid,
			wantReason: "probed_at",
		},
		{
			name:       "available but not applied",
			tamper:     func(r *Record) { r.Capabilities[0].Status = spec.StatusNotApplied },
			wantDiag:   spec.DiagEvidenceInvalid,
			wantReason: "available but not-applied",
		},
		{
			name:       "unavailable but applied",
			tamper:     func(r *Record) { r.Capabilities[0].Availability = spec.AvailabilityUnavailable },
			wantDiag:   spec.DiagEvidenceInvalid,
			wantReason: "unavailable but applied",
		},
		{
			name:       "unknown guarantee",
			tamper:     func(r *Record) { r.Guarantees[0].Name = "total-silence" },
			wantDiag:   spec.DiagEvidenceInvalid,
			wantReason: "unknown guarantee",
		},
		{
			name: "duplicated guarantee",
			tamper: func(r *Record) {
				r.Guarantees = append(r.Guarantees, r.Guarantees[0])
			},
			wantDiag:   spec.DiagEvidenceInvalid,
			wantReason: "duplicated guarantee",
		},
		{
			name: "missing guarantee",
			tamper: func(r *Record) {
				r.Guarantees = r.Guarantees[1:]
			},
			wantDiag:   spec.DiagEvidenceInvalid,
			wantReason: "missing guarantee",
		},
		{
			name:       "unknown outcome",
			tamper:     func(r *Record) { r.Outcome = "partially-established" },
			wantDiag:   spec.DiagEvidenceInvalid,
			wantReason: "unknown outcome",
		},
		{
			name: "established outcome carrying a rejection",
			tamper: func(r *Record) {
				phase, diagnostic := spec.PhaseCapabilityProbe, spec.DiagCapabilityUnavailable
				r.RejectedBefore, r.Diagnostic = &phase, &diagnostic
			},
			wantDiag:   spec.DiagEvidenceInvalid,
			wantReason: "established outcome carries",
		},
		{
			name: "rejected outcome without a diagnostic",
			tamper: func(r *Record) {
				r.Outcome = spec.OutcomeRejected
			},
			wantDiag:   spec.DiagEvidenceInvalid,
			wantReason: "without both rejected_before and diagnostic",
		},
		{
			name: "rejected before an unknown phase",
			tamper: func(r *Record) {
				phase, diagnostic := "vibes-check", spec.DiagCapabilityUnavailable
				r.Outcome, r.RejectedBefore, r.Diagnostic = spec.OutcomeRejected, &phase, &diagnostic
			},
			wantDiag:   spec.DiagEvidenceInvalid,
			wantReason: "is not a known phase",
		},
		{
			name: "rejected with an unstable diagnostic",
			tamper: func(r *Record) {
				phase, diagnostic := spec.PhaseCapabilityProbe, "it_did_not_work"
				r.Outcome, r.RejectedBefore, r.Diagnostic = spec.OutcomeRejected, &phase, &diagnostic
			},
			wantDiag:   spec.DiagEvidenceInvalid,
			wantReason: "not a stable hardened diagnostic",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := buildEstablished(t)
			tc.tamper(&rec)
			diag, reason := rec.Validate()
			if diag != tc.wantDiag {
				t.Fatalf("diagnostic %q (%s), want %q", diag, reason, tc.wantDiag)
			}
			if !strings.Contains(reason, tc.wantReason) {
				t.Errorf("reason %q does not mention %q", reason, tc.wantReason)
			}
		})
	}
}

// A record cannot claim a guarantee whose classes it also reports unapplied.
// Build never produces one, so the inconsistency has to be injected.
func TestValidateRejectsGuaranteeWithoutItsClasses(t *testing.T) {
	rec := buildEstablished(t)
	rec.Outcome = spec.OutcomeRejected
	phase, diagnostic := spec.PhaseCapabilityProbe, spec.DiagCapabilityUnavailable
	rec.RejectedBefore, rec.Diagnostic = &phase, &diagnostic

	i := capabilityIndex(t, rec, spec.ClassActiveProbe)
	rec.Capabilities[i].Availability = spec.AvailabilityUnavailable
	rec.Capabilities[i].Status = spec.StatusNotApplied
	// fail-closed-capability-preflight maps to active-capability-probe alone, so
	// leaving it established is now a contradiction.

	diag, reason := rec.Validate()
	if diag != spec.DiagEvidenceInvalid {
		t.Fatalf("diagnostic %q (%s), want %q", diag, reason, spec.DiagEvidenceInvalid)
	}
	if !strings.Contains(reason, "established while class") {
		t.Errorf("reason %q does not name the unapplied class", reason)
	}
}

func TestValidateRejectsEstablishedOutcomeWithGaps(t *testing.T) {
	t.Run("unapplied capability", func(t *testing.T) {
		rec := buildEstablished(t)
		i := capabilityIndex(t, rec, spec.ClassActiveProbe)
		rec.Capabilities[i].Availability = spec.AvailabilityUnavailable
		rec.Capabilities[i].Status = spec.StatusNotApplied
		rec.Guarantees[guaranteeIndex(t, rec, spec.GuaranteeFailClosedPreflight)].Established = false

		diag, reason := rec.Validate()
		if diag != spec.DiagEvidenceInvalid || !strings.Contains(reason, "established outcome while capability") {
			t.Errorf("got %q / %q, want an established-outcome-with-unapplied-capability error", diag, reason)
		}
	})

	t.Run("unestablished guarantee", func(t *testing.T) {
		rec := buildEstablished(t)
		rec.Guarantees[0].Established = false

		diag, reason := rec.Validate()
		if diag != spec.DiagEvidenceInvalid || !strings.Contains(reason, "established outcome while guarantee") {
			t.Errorf("got %q / %q, want an established-outcome-with-unestablished-guarantee error", diag, reason)
		}
	})
}

// ------------------------------------------------------------------ Decode

func TestDecodeRoundTripsBuiltRecords(t *testing.T) {
	original := buildEstablished(t)
	decoded, diag, reason := Decode(original.JSON())
	if diag != "" {
		t.Fatalf("Decode rejected a freshly built record: %s: %s", diag, reason)
	}
	if decoded.Outcome != original.Outcome ||
		len(decoded.Capabilities) != len(original.Capabilities) ||
		len(decoded.Guarantees) != len(original.Guarantees) {
		t.Errorf("round trip changed the record:\n got %+v\nwant %+v", decoded, original)
	}
}

func TestDecodeRoundTripsRejectedRecords(t *testing.T) {
	original := Build(spec.PlatformMacOS, spec.BackendMacOSSandbox, spec.QualificationUnqualified, nil)
	decoded, diag, reason := Decode(original.JSON())
	if diag != "" {
		t.Fatalf("Decode rejected a rejected record: %s: %s", diag, reason)
	}
	if decoded.RejectedBefore == nil || *decoded.RejectedBefore != *original.RejectedBefore {
		t.Errorf("rejected_before lost in the round trip: %v", decoded.RejectedBefore)
	}
	if decoded.Diagnostic == nil || *decoded.Diagnostic != *original.Diagnostic {
		t.Errorf("diagnostic lost in the round trip: %v", decoded.Diagnostic)
	}
}

func TestDecodeRejectsMalformedDocuments(t *testing.T) {
	valid := buildEstablished(t)

	mutate := func(t *testing.T, edit func(map[string]json.RawMessage)) []byte {
		t.Helper()
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(valid.JSON(), &raw); err != nil {
			t.Fatalf("fixture is not a JSON object: %v", err)
		}
		edit(raw)
		data, err := json.Marshal(raw)
		if err != nil {
			t.Fatalf("re-marshal: %v", err)
		}
		return data
	}

	t.Run("not a JSON object", func(t *testing.T) {
		for _, data := range [][]byte{[]byte("["), []byte("[]"), []byte("7"), nil} {
			_, diag, reason := Decode(data)
			if diag != spec.DiagEvidenceInvalid {
				t.Errorf("Decode(%q) diagnostic %q (%s), want %q", data, diag, reason, spec.DiagEvidenceInvalid)
			}
		}
	})

	t.Run("extra field", func(t *testing.T) {
		data := mutate(t, func(raw map[string]json.RawMessage) {
			raw["enforcement_notes"] = json.RawMessage(`"trust me"`)
		})
		_, diag, reason := Decode(data)
		if diag != spec.DiagEvidenceInvalid || !strings.Contains(reason, "extra field") {
			t.Errorf("got %q / %q, want an extra-field error", diag, reason)
		}
		if !strings.Contains(reason, "enforcement_notes") {
			t.Errorf("reason %q does not name the extra field", reason)
		}
	})

	t.Run("missing field", func(t *testing.T) {
		for _, field := range []string{"record_version", "outcome", "capabilities", "guarantees", "rejected_before"} {
			data := mutate(t, func(raw map[string]json.RawMessage) { delete(raw, field) })
			_, diag, reason := Decode(data)
			if diag != spec.DiagEvidenceInvalid || !strings.Contains(reason, "missing field") {
				t.Errorf("dropping %q gave %q / %q, want a missing-field error", field, diag, reason)
			}
		}
	})

	t.Run("wrong field type", func(t *testing.T) {
		data := mutate(t, func(raw map[string]json.RawMessage) {
			raw["capabilities"] = json.RawMessage(`"none"`)
		})
		_, diag, reason := Decode(data)
		if diag != spec.DiagEvidenceInvalid || !strings.Contains(reason, "closed shape") {
			t.Errorf("got %q / %q, want a closed-shape error", diag, reason)
		}
	})

	t.Run("valid shape but invalid content", func(t *testing.T) {
		data := mutate(t, func(raw map[string]json.RawMessage) {
			raw["outcome"] = json.RawMessage(`"maybe"`)
		})
		_, diag, reason := Decode(data)
		if diag != spec.DiagEvidenceInvalid || !strings.Contains(reason, "unknown outcome") {
			t.Errorf("got %q / %q, want Validate to run after the shape check", diag, reason)
		}
	})
}

// A null rejected_before is the established form and must survive Decode; the
// field is present, it just carries null.
func TestDecodeAcceptsExplicitNulls(t *testing.T) {
	rec := buildEstablished(t)
	data := rec.JSON()
	if !strings.Contains(string(data), `"rejected_before": null`) {
		t.Fatalf("established record does not emit a null rejected_before:\n%s", data)
	}
	decoded, diag, reason := Decode(data)
	if diag != "" {
		t.Fatalf("Decode rejected explicit nulls: %s: %s", diag, reason)
	}
	if decoded.RejectedBefore != nil {
		t.Errorf("rejected_before decoded to %v, want nil", decoded.RejectedBefore)
	}
}

// -------------------------------------------------------------------- JSON

func TestJSONIsIndentedAndNewlineTerminated(t *testing.T) {
	data := buildEstablished(t).JSON()
	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Errorf("record JSON does not end with a newline")
	}
	if !json.Valid(data[:len(data)-1]) {
		t.Errorf("record JSON is not valid JSON:\n%s", data)
	}
	if !strings.Contains(string(data), "\n  \"record_version\"") {
		t.Errorf("record JSON is not indented:\n%s", data)
	}
}

// Field order in the emitted document is part of the artifact: reviewers read
// it against the specification example.
func TestJSONFieldOrderMatchesTheSpecificationExample(t *testing.T) {
	data := string(buildEstablished(t).JSON())
	order := []string{
		"record_version", "hardened_profile", "execution_policy", "platform",
		"enforcement_backend", "qualification_status", "outcome", "rejected_before",
		"diagnostic", "capabilities", "guarantees",
	}
	position := -1
	for _, field := range order {
		next := strings.Index(data, `"`+field+`"`)
		if next < 0 {
			t.Fatalf("field %q missing from the emitted record", field)
		}
		if next <= position {
			t.Errorf("field %q appears out of order at %d (previous field ended at %d)", field, next, position)
		}
		position = next
	}
}
