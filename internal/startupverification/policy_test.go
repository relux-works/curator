package startupverification

import (
	"strings"
	"testing"
)

// The full local truth table distinguishes definite failures from unavailable
// evidence; a later unavailable check never erases an earlier failure.
func TestLocalLevels(t *testing.T) {
	for _, tc := range []struct {
		name     string
		checks   []Observation
		allowed  [4]bool
		verified bool
	}{
		{"passed", []Observation{{Status: Passed}}, [4]bool{true, true, true, true}, true},
		{"failed", []Observation{{Status: Failed}}, [4]bool{true, false, false, false}, false},
		{"unavailable", []Observation{{Status: Unavailable}}, [4]bool{true, true, false, false}, false},
		{"mixed", []Observation{{Status: Failed}, {Status: Unavailable}}, [4]bool{true, false, false, false}, false},
	} {
		for level := 1; level <= 4; level++ {
			t.Run(tc.name+string(rune('0'+level)), func(t *testing.T) {
				got, err := Decide(level, tc.checks, CentrePassed)
				if err != nil || got.Allow != tc.allowed[level-1] || got.Verified != tc.verified {
					t.Fatalf("verdict=%+v err=%v", got, err)
				}
				if !tc.verified && got.Allow && len(got.Warnings) == 0 {
					t.Fatal("degraded start has no warning")
				}
			})
		}
	}
}

// Only absent/unreachable centres fall back; denial/invalid reply never does.
// Local failures remain blocking even when the centre cannot be reached.
func TestCentreFallback(t *testing.T) {
	for _, tc := range []struct {
		centre          Centre
		allow, verified bool
		warning         bool
	}{
		{CentrePassed, true, true, false}, {CentreAbsent, true, false, true},
		{CentreUnavailable, true, false, true}, {CentreDenied, false, false, false},
		{CentreInvalid, false, false, false},
	} {
		got, err := Decide(4, []Observation{{Status: Passed}}, tc.centre)
		if err != nil || got.Allow != tc.allow || got.Verified != tc.verified || (len(got.Warnings) > 0) != tc.warning {
			t.Fatalf("%s: %+v, %v", tc.centre, got, err)
		}
		got, err = Decide(4, []Observation{{Status: Failed}}, tc.centre)
		if err != nil || got.Allow {
			t.Fatalf("%s clears local failure: %+v, %v", tc.centre, got, err)
		}
	}
}

// Structural invalidity cannot be downgraded even at level1. A profile may
// equal/tighten the baseline and dependency minima combine by maximum.
func TestInvalidPolicyAndFloors(t *testing.T) {
	for _, level := range []int{0, 5, -1} {
		if _, err := Decide(level, nil, CentreAbsent); err == nil {
			t.Fatal("invalid level accepted")
		}
	}
	if _, err := Decide(1, []Observation{{Status: Status("invalid")}}, CentreAbsent); err == nil {
		t.Fatal("unknown status accepted")
	}
	if _, err := Decide(1, nil, Centre("invalid-status")); err == nil {
		t.Fatal("unknown centre status accepted")
	}
	for _, tc := range []struct {
		baseline int
		floors   []int
		want     int
		code     string
	}{
		{3, nil, 3, ""}, {2, []int{2, 4, 3}, 4, ""}, {3, []int{2, 4}, 0, "startup_verification_weakening"}, {3, []int{5}, 0, "startup_verification_invalid"},
	} {
		got, err := EffectiveLevel(tc.baseline, tc.floors...)
		if tc.code == "" {
			if err != nil || got != tc.want {
				t.Fatalf("floor %v: %d,%v", tc, got, err)
			}
		} else if d, ok := err.(*Diagnostic); !ok || d.Code != tc.code {
			t.Fatalf("floor %v: %v", tc, err)
		}
	}
}

// Hardening below an explicit profile minimum preserves the refusal contract;
// remediation names the editable manifests and recomposition, beside equality.
func TestFloorHardeningRemediation(t *testing.T) {
	if level, err := EffectiveLevel(4, 4); err != nil || level != 4 {
		t.Fatalf("equal neighbor: %d %v", level, err)
	}
	_, err := EffectiveLevel(4, 3)
	if err == nil || !strings.HasPrefix(err.Error(), "startup_verification_weakening:") || !strings.Contains(err.Error(), "selected context manifests") || !strings.Contains(err.Error(), "profile update") {
		t.Fatalf("hardening remediation lost: %v", err)
	}
}
