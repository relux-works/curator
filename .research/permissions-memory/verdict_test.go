package permissionsmemory

import (
	"os"
	"testing"
)

type observation struct {
	Atom            string `json:"atom"`
	Verdict         string `json:"verdict"`
	Positive        bool   `json:"positive_effect"`
	Denial          bool   `json:"specific_denial"`
	ForbiddenEffect bool   `json:"forbidden_effect"`
	Limit           string `json:"limit"`
}

func verdict(positive, denial, forbidden bool) string {
	if forbidden {
		return "refused"
	}
	if positive && denial {
		return "qualified"
	}
	if os.Getenv("QUAL_MUTANT") == "unknown-qualified" {
		return "qualified"
	}
	return "unknown"
}

// Claim: qualification needs a working positive sibling and a specific negative;
// a timeout/setup error, missing control, generic error or forbidden effect
// cannot produce a qualified row. A real observed forbidden effect is refused.
func TestVerdictGate(t *testing.T) {
	for _, row := range []struct {
		positive, denial, forbidden bool
		want                        string
	}{
		{true, true, false, "qualified"}, {false, true, false, "unknown"},
		{true, false, false, "unknown"}, {false, false, false, "unknown"},
		{true, true, true, "refused"}, {false, false, true, "refused"},
	} {
		if got := verdict(row.positive, row.denial, row.forbidden); got != row.want {
			t.Fatalf("false qualification: %s", got)
		}
	}
}
