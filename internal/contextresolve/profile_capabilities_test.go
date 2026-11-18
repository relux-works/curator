package contextresolve

import (
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/contextlock"
)

func profileSkillEdge(present bool, hosts ...string) constraint {
	return constraint{requirement: Requirement{
		Kind: contextlock.KindSkill, Name: "guide", ProfileSkillRequirement: true,
		CapabilitiesPresent: present, Capabilities: capabilities.Manifest{Network: hosts},
	}}
}

// TestProfileSkillCapabilityEdgesRequireAgreement proves shared profile edges
// neither union host sets nor let an omitted declaration erase an explicit one.
func TestProfileSkillCapabilityEdgesRequireAgreement(t *testing.T) {
	tests := []struct {
		name  string
		edges []constraint
		want  string
	}{
		{name: "different hosts", edges: []constraint{profileSkillEdge(true, "api.example.net"), profileSkillEdge(true, "docs.example.net")}, want: DiagProfileSkillCapabilityConflict},
		{name: "one omitted", edges: []constraint{profileSkillEdge(true, "api.example.net"), profileSkillEdge(false)}, want: DiagProfileSkillCapabilityConflict},
		{name: "equal different order", edges: []constraint{profileSkillEdge(true, "api.example.net", "docs.example.net"), profileSkillEdge(true, "docs.example.net", "api.example.net")}},
		{name: "equal host case", edges: []constraint{profileSkillEdge(true, "API.EXAMPLE.NET"), profileSkillEdge(true, "api.example.net")}},
		{name: "all omitted", edges: []constraint{profileSkillEdge(false), profileSkillEdge(false)}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &resolver{constraints: map[string][]constraint{"guide": tc.edges}}
			caps, present, err := r.profileSkillCapabilities("guide")
			if tc.want != "" {
				if err == nil || !strings.Contains(err.Error(), tc.want) || present {
					t.Fatalf("got capabilities=%+v present=%v error=%v, want %s refusal", caps, present, err, tc.want)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			wantPresent := tc.name == "equal different order" || tc.name == "equal host case"
			if present != wantPresent {
				t.Fatalf("present=%v, want %v", present, wantPresent)
			}
			if tc.name == "equal different order" && (len(caps.Network) != 2 || caps.Network[0] != "api.example.net" || caps.Network[1] != "docs.example.net") {
				t.Fatalf("capability order not normalized: %+v", caps)
			}
			if tc.name == "equal host case" && (len(caps.Network) != 1 || caps.Network[0] != "api.example.net") {
				t.Fatalf("host spelling not normalized by audit semantics: %+v", caps)
			}
		})
	}
}
