// Package startupverification decides startup from typed verification outcomes.
// Authority, structural and path gates return errors before this policy is used.
package startupverification

import "fmt"

// DefaultLevel requires complete local verification unless explicitly configured.
const DefaultLevel = 3

// Diagnostic exposes a stable startup policy error code and its detail.
type Diagnostic struct {
	Code   string
	Detail string
}

func (d *Diagnostic) Error() string { return d.Code + ": " + d.Detail }

// Invalid reports a malformed policy input or unknown verification outcome.
func Invalid(detail string) error {
	return &Diagnostic{Code: "startup_verification_invalid", Detail: detail}
}

// Status classifies a required local verification observation.
type Status string

// Local outcomes distinguish a definite failure from unavailable evidence.
const (
	Passed      Status = "passed"
	Failed      Status = "failed"
	Unavailable Status = "unavailable"
)

// Observation carries one local check's typed outcome and diagnostic code.
type Observation struct {
	Status Status
	Code   string
}

// Centre classifies the verification centre's availability and verdict.
type Centre string

// Centre outcomes keep fallback conditions separate from explicit refusals.
const (
	CentreAbsent      Centre = "absent"
	CentrePassed      Centre = "passed"
	CentreUnavailable Centre = "unavailable"
	CentreDenied      Centre = "denied"
	CentreInvalid     Centre = "invalid"
)

// Verdict records admission, verification completeness and all policy diagnostics.
type Verdict struct {
	Allow    bool
	Verified bool
	Warnings []string
	Refusals []string
}

// EffectiveLevel admits explicit profile floors only at or above the baseline.
func EffectiveLevel(baseline int, floors ...int) (int, error) {
	if baseline < 1 || baseline > 4 {
		return 0, Invalid("level must be an integer from 1 to 4")
	}
	level := baseline
	for _, floor := range floors {
		if floor < 1 || floor > 4 {
			return 0, Invalid("profile level must be an integer from 1 to 4")
		}
		if floor < baseline {
			return 0, &Diagnostic{Code: "startup_verification_weakening", Detail: "profile minimum is below machine level; raise startup_verification.level in the selected context manifests, then run profile update"}
		}
		if floor > level {
			level = floor
		}
	}
	return level, nil
}

// Decide evaluates every supplied observation before deciding. The caller must
// supply all required checks, including Unavailable when evidence is missing.
func Decide(level int, checks []Observation, centre Centre) (Verdict, error) {
	if _, err := EffectiveLevel(level); err != nil {
		return Verdict{}, err
	}
	switch centre {
	case CentreAbsent, CentrePassed, CentreUnavailable, CentreDenied, CentreInvalid:
	default:
		return Verdict{}, Invalid("unknown centre outcome")
	}
	v := Verdict{Allow: true, Verified: true}
	for _, check := range checks {
		var code string
		switch check.Status {
		case Passed:
			continue
		case Failed:
			code = "startup_verification_failed"
		case Unavailable:
			code = "startup_verification_unavailable"
		default:
			return Verdict{}, Invalid("unknown local outcome")
		}
		v.Verified = false
		if level == 1 || (level == 2 && check.Status == Unavailable) {
			v.Warnings = append(v.Warnings, code)
		} else {
			v.Refusals = append(v.Refusals, code)
			v.Allow = false
		}
	}
	if level == 4 {
		switch centre {
		case CentreAbsent, CentreUnavailable:
			v.Warnings = append(v.Warnings, "verification_centre_unavailable")
			v.Verified = false
		case CentreDenied, CentreInvalid:
			v.Allow = false
			v.Verified = false
			v.Refusals = append(v.Refusals, fmt.Sprintf("verification_centre_%s", centre))
		}
	}
	return v, nil
}
