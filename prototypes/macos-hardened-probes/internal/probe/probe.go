// Package probe runs the macOS capability-class probes and reduces what they
// observed into the closed evidence record of hardened-execution.md section 6.4.
//
// Every class probe has the same shape:
//
//   - a positive capability test, which applies the candidate control in a probe
//     domain and observes whether the kernel refuses the operation;
//   - a negative control, which repeats the same operation with the control
//     removed and must observe success, so a clean sweep of denials cannot come
//     from a broken measurement;
//   - one or more adversarial escapes, which try to reach the same effect by a
//     different route.
//
// A class is available only when the positive test denies, the negative control
// succeeds, and no adversarial escape succeeds. Anything else — including a
// probe that could not run — is unavailable, because section 5.6 requires an
// inconclusive probe to reject.
package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/evidence"
	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/inside"
	"github.com/relux-works/curator/prototypes/macos-hardened-probes/internal/spec"
)

// Verdict is the reduced result of one class probe.
type Verdict string

const (
	// VerdictAvailable means the class was observed to hold in this run.
	VerdictAvailable Verdict = "available"
	// VerdictUnavailable means the class was observed not to hold.
	VerdictUnavailable Verdict = "unavailable"
	// VerdictInconclusive means the probe could not observe either way. It is
	// reduced to unavailable in the evidence record; it is kept distinct in the
	// report so an operator can tell "the host does not do this" from "the
	// probe did not work".
	VerdictInconclusive Verdict = "inconclusive"
)

// Check is one measured statement inside a class probe.
type Check struct {
	Name string `json:"name"`
	// Kind is "positive", "negative-control" or "adversarial".
	Kind string `json:"kind"`
	// Expectation states what the class requires of this check, in the words of
	// the specification, so the report can be read without the code.
	Expectation string `json:"expectation"`
	// Observed is what actually happened.
	Observed string `json:"observed"`
	// Pass reports whether Observed satisfies Expectation.
	Pass   bool   `json:"pass"`
	Detail string `json:"detail,omitempty"`
}

// ClassResult is the full record of one capability-class probe.
type ClassResult struct {
	Class     string   `json:"class"`
	Verdict   Verdict  `json:"verdict"`
	Mechanism string   `json:"mechanism"`
	Reasons   []string `json:"reasons,omitempty"`
	Checks    []Check  `json:"checks"`
	// Applied reports whether the control was actually installed in a probe
	// domain during this run. A class that is available but was never applied
	// must not be reported applied in the evidence record.
	Applied bool `json:"applied"`
	// DurationMS is how long the probe took, so a suspiciously instant result
	// is visible.
	DurationMS int64 `json:"duration_ms"`
}

// Report is the prototype-side detailed artifact. It is deliberately separate
// from the closed evidence record: section 6.4 forbids extra fields there, and
// an operator still needs the observations behind the verdicts.
type Report struct {
	ReportVersion  string          `json:"report_version"`
	SpecRevision   string          `json:"spec_revision"`
	Platform       string          `json:"platform"`
	Backend        string          `json:"enforcement_backend"`
	Host           HostInfo        `json:"host"`
	Classes        []ClassResult   `json:"classes"`
	Mechanisms     []Mechanism     `json:"mechanisms"`
	FailClosed     []FailClosed    `json:"fail_closed_controls"`
	StartedAt      string          `json:"started_at"`
	DurationMS     int64           `json:"duration_ms"`
	EvidenceRecord evidence.Record `json:"evidence_record"`
	ExitCode       int             `json:"exit_code"`
}

// ReportVersion identifies this prototype-only schema. It is not a curator-spec
// artifact and must never be mistaken for one.
const ReportVersion = "macos-hardened-probe-report-v1"

// HostInfo records what the probes ran on, because a capability observation is
// only meaningful together with the host that produced it.
type HostInfo struct {
	ProductName    string `json:"product_name"`
	ProductVersion string `json:"product_version"`
	BuildVersion   string `json:"build_version"`
	KernelVersion  string `json:"kernel_version"`
	Arch           string `json:"arch"`
	GoVersion      string `json:"go_version"`
	SIPStatus      string `json:"sip_status"`
	UID            int    `json:"uid"`
}

// Mechanism records the support status of a platform mechanism the probes
// touched or deliberately rejected.
type Mechanism struct {
	Name string `json:"name"`
	// Status is "supported", "deprecated", "private", "unavailable" or
	// "not-applicable".
	Status string `json:"status"`
	// Classes lists the capability classes the mechanism was considered for.
	Classes []string `json:"classes,omitempty"`
	Note    string   `json:"note"`
	// Exercised reports whether this run executed a probe against the
	// mechanism. A mechanism that was only considered carries a status read
	// from the published interface, which is a weaker statement than a
	// measurement and must not be mistaken for one.
	Exercised bool `json:"exercised"`
	// Observation is what this run measured about the mechanism, named check by
	// check, or why there is no measurement.
	Observation string `json:"observation"`
}

// FailClosed records one forced-unavailable injection and the rejection it
// produced, which is the executable form of the section 11 preflight evidence.
type FailClosed struct {
	ForcedClass    string `json:"forced_class"`
	Outcome        string `json:"outcome"`
	RejectedBefore string `json:"rejected_before"`
	Diagnostic     string `json:"diagnostic"`
	ExitCode       int    `json:"exit_code"`
	Pass           bool   `json:"pass"`
}

// JSON renders the detailed report.
func (r Report) JSON() []byte {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return []byte(fmt.Sprintf("{\"report_version\":%q,\"marshal_error\":%q}\n", ReportVersion, err.Error()))
	}
	return append(data, '\n')
}

// reduce turns a set of checks into a verdict. Every check must pass; a check
// that could not be evaluated makes the class inconclusive.
func reduce(checks []Check) (Verdict, []string) {
	var reasons []string
	inconclusive := false
	for _, check := range checks {
		if check.Pass {
			continue
		}
		reasons = append(reasons, fmt.Sprintf("%s (%s): %s", check.Name, check.Kind, check.Observed))
		if check.Observed == string(inside.OutcomeInconclusive) || check.Observed == "probe-failed" {
			inconclusive = true
		}
	}
	switch {
	case len(reasons) == 0:
		return VerdictAvailable, nil
	case inconclusive:
		return VerdictInconclusive, reasons
	default:
		return VerdictUnavailable, reasons
	}
}

// observation builds a Check from an in-domain attempt.
func observation(name, kind, expectation string, a inside.Attempt, found bool, want string) Check {
	if !found {
		return Check{
			Name:        name,
			Kind:        kind,
			Expectation: expectation,
			Observed:    "probe-failed",
			Pass:        false,
			Detail:      "the in-domain agent did not report this attempt",
		}
	}
	check := Check{
		Name:        name,
		Kind:        kind,
		Expectation: expectation,
		Observed:    a.Outcome,
		Pass:        a.Outcome == want,
		Detail:      a.Detail,
	}
	if a.Errno != "" {
		if check.Detail != "" {
			check.Detail = a.Errno + ": " + check.Detail
		} else {
			check.Detail = a.Errno
		}
	}
	return check
}

// failedCheck records a check that could not be run at all.
func failedCheck(name, kind, expectation, detail string) Check {
	return Check{
		Name:        name,
		Kind:        kind,
		Expectation: expectation,
		Observed:    "probe-failed",
		Pass:        false,
		Detail:      detail,
	}
}

// timed measures a class probe.
func timed(class, mechanism string, run func() ([]Check, bool)) ClassResult {
	start := time.Now()
	checks, applied := run()
	verdict, reasons := reduce(checks)
	return ClassResult{
		Class:      class,
		Verdict:    verdict,
		Mechanism:  mechanism,
		Reasons:    reasons,
		Checks:     checks,
		Applied:    applied && verdict == VerdictAvailable,
		DurationMS: time.Since(start).Milliseconds(),
	}
}

// Observations reduces class results to the two values the closed record can
// carry. Inconclusive collapses to unavailable, which is the fail-closed
// direction required by section 5.6.
func Observations(results []ClassResult) map[string]evidence.Observation {
	out := map[string]evidence.Observation{}
	for _, result := range results {
		availability := spec.AvailabilityUnavailable
		if result.Verdict == VerdictAvailable {
			availability = spec.AvailabilityAvailable
		}
		out[result.Class] = evidence.Observation{
			Availability: availability,
			Applied:      result.Applied,
		}
	}
	return out
}

// runContext is the shared state a class probe needs.
type runContext struct {
	ctx context.Context
	env *Environment
}
