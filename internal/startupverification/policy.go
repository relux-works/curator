// Package startupverification decides startup from typed verification outcomes.
// Authority, structural and path gates return errors before this policy is used.
package startupverification

import "fmt"

const DefaultLevel = 3

type Diagnostic struct {
	Code   string
	Detail string
}

func (d *Diagnostic) Error() string { return d.Code + ": " + d.Detail }
func Invalid(detail string) error {
	return &Diagnostic{Code: "startup_verification_invalid", Detail: detail}
}

type Status string

const (
	Passed      Status = "passed"
	Failed      Status = "failed"
	Unavailable Status = "unavailable"
)

type Observation struct {
	Status Status
	Code   string
}
type Centre string

const (
	CentreAbsent      Centre = "absent"
	CentrePassed      Centre = "passed"
	CentreUnavailable Centre = "unavailable"
	CentreDenied      Centre = "denied"
	CentreInvalid     Centre = "invalid"
)

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
			return 0, &Diagnostic{Code: "startup_verification_weakening", Detail: "profile minimum is below machine level"}
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
