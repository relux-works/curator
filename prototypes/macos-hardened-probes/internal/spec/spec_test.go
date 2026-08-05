package spec

import (
	"strings"
	"testing"
)

// The transcription is only useful if it is exhaustive and internally
// consistent: a class nobody maps to a guarantee can never block a build, and a
// guarantee that maps to an unknown class can never be established. Both would
// be silent holes, so both are asserted here rather than left to review.

func TestClassesIsExhaustiveAndUnique(t *testing.T) {
	classes := Classes()
	if len(classes) != 11 {
		t.Fatalf("inventory has %d classes, want the 11 of section 6.1", len(classes))
	}
	seen := map[string]bool{}
	for _, class := range classes {
		if class == "" {
			t.Error("inventory contains an empty class name")
		}
		if seen[class] {
			t.Errorf("class %q appears twice in the inventory", class)
		}
		seen[class] = true
	}
}

func TestGuaranteesIsExhaustiveAndUnique(t *testing.T) {
	guarantees := Guarantees()
	if len(guarantees) != 6 {
		t.Fatalf("guarantee list has %d entries, want the 6 of section 5", len(guarantees))
	}
	seen := map[string]bool{}
	for _, guarantee := range guarantees {
		if seen[guarantee] {
			t.Errorf("guarantee %q appears twice", guarantee)
		}
		seen[guarantee] = true
	}
}

func TestGuaranteeClassesMapsOnlyKnownClasses(t *testing.T) {
	for _, guarantee := range Guarantees() {
		classes := GuaranteeClasses(guarantee)
		if len(classes) == 0 {
			t.Errorf("guarantee %q maps to no class, so it could never be blocked", guarantee)
			continue
		}
		seen := map[string]bool{}
		for _, class := range classes {
			if !IsClass(class) {
				t.Errorf("guarantee %q maps to unknown class %q", guarantee, class)
			}
			if seen[class] {
				t.Errorf("guarantee %q maps to class %q twice", guarantee, class)
			}
			seen[class] = true
		}
	}
}

// Every class must matter to at least one guarantee. A class that maps nowhere
// would be measured and then ignored, which is the failure mode the closed
// record exists to prevent.
func TestEveryClassIsMappedToAGuarantee(t *testing.T) {
	mapped := map[string]bool{}
	for _, guarantee := range Guarantees() {
		for _, class := range GuaranteeClasses(guarantee) {
			mapped[class] = true
		}
	}
	for _, class := range Classes() {
		if !mapped[class] {
			t.Errorf("class %q is in the inventory but maps to no guarantee", class)
		}
	}
}

func TestGuaranteeClassesUnknownGuarantee(t *testing.T) {
	for _, name := range []string{"", "not-a-guarantee", ClassNetworkDenial} {
		if got := GuaranteeClasses(name); got != nil {
			t.Errorf("GuaranteeClasses(%q) = %v, want nil", name, got)
		}
	}
}

// The exact mapping is normative, so it is pinned literally: a future edit that
// silently drops domain-membership from a guarantee would otherwise pass.
func TestGuaranteeClassesExactMapping(t *testing.T) {
	want := map[string][]string{
		GuaranteeNetworkDenial:       {ClassNetworkDenial, ClassEndpointRevocation, ClassDomainMembership},
		GuaranteeReadOnlyInputs:      {ClassReadOnlySource, ClassReadOnlyToolchain, ClassViewRestriction},
		GuaranteePrivateWrites:       {ClassWriteConfinement, ClassViewRestriction, ClassDomainMembership},
		GuaranteeResourceBounds:      {ClassAggregateBounds, ClassDomainMembership, ClassDomainTermination},
		GuaranteeExecAllowlist:       {ClassExecAllowlist, ClassDomainMembership},
		GuaranteeFailClosedPreflight: {ClassActiveProbe},
	}
	for guarantee, expected := range want {
		got := GuaranteeClasses(guarantee)
		if len(got) != len(expected) {
			t.Errorf("guarantee %q maps to %v, want %v", guarantee, got, expected)
			continue
		}
		for i := range expected {
			if got[i] != expected[i] {
				t.Errorf("guarantee %q class %d = %q, want %q", guarantee, i, got[i], expected[i])
			}
		}
	}
}

// Callers append to and sort these slices. If a call handed out shared backing
// storage, one caller's edit would rewrite the normative inventory for every
// later caller.
func TestListsAreFreshOnEachCall(t *testing.T) {
	for name, list := range map[string]func() []string{
		"Classes":     Classes,
		"Guarantees":  Guarantees,
		"Phases":      Phases,
		"Diagnostics": Diagnostics,
	} {
		first := list()
		if len(first) == 0 {
			t.Fatalf("%s() is empty", name)
		}
		original := first[0]
		first[0] = "clobbered"
		if second := list(); second[0] != original {
			t.Errorf("%s() hands out shared storage: after clobbering, got %q", name, second[0])
		}
	}
}

func TestPhasesAreOrderedAndUnique(t *testing.T) {
	phases := Phases()
	if len(phases) != 9 {
		t.Fatalf("phase list has %d entries, want 9", len(phases))
	}
	// capability-probe is the phase every capability rejection names, so its
	// position relative to domain entry is load-bearing.
	probeIndex, entryIndex := -1, -1
	seen := map[string]bool{}
	for i, phase := range phases {
		if seen[phase] {
			t.Errorf("phase %q appears twice", phase)
		}
		seen[phase] = true
		switch phase {
		case PhaseCapabilityProbe:
			probeIndex = i
		case PhaseDomainEntry:
			entryIndex = i
		}
	}
	if probeIndex < 0 || entryIndex < 0 {
		t.Fatalf("capability-probe (%d) or domain-entry (%d) missing from the phase list", probeIndex, entryIndex)
	}
	if probeIndex >= entryIndex {
		t.Errorf("capability-probe is at %d and domain-entry at %d; the probe must precede entry", probeIndex, entryIndex)
	}
}

func TestDiagnosticsAreStableAndPrefixed(t *testing.T) {
	diagnostics := Diagnostics()
	if len(diagnostics) != 9 {
		t.Fatalf("diagnostic list has %d entries, want 9", len(diagnostics))
	}
	seen := map[string]bool{}
	for _, diagnostic := range diagnostics {
		if !strings.HasPrefix(diagnostic, "hardened_") {
			t.Errorf("diagnostic %q is not in the hardened_ namespace", diagnostic)
		}
		if seen[diagnostic] {
			t.Errorf("diagnostic %q appears twice", diagnostic)
		}
		seen[diagnostic] = true
	}
}

func TestMembershipPredicates(t *testing.T) {
	cases := []struct {
		name      string
		predicate func(string) bool
		member    string
		stranger  string
	}{
		{"IsClass", IsClass, ClassNetworkDenial, GuaranteeNetworkDenial},
		{"IsGuarantee", IsGuarantee, GuaranteeNetworkDenial, ClassNetworkDenial},
		{"IsDiagnostic", IsDiagnostic, DiagCapabilityUnavailable, "capability_unavailable"},
		{"IsPhase", IsPhase, PhaseCapabilityProbe, "capability_probe"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.predicate(tc.member) {
				t.Errorf("%s(%q) = false, want true", tc.name, tc.member)
			}
			if tc.predicate(tc.stranger) {
				t.Errorf("%s(%q) = true, want false", tc.name, tc.stranger)
			}
			if tc.predicate("") {
				t.Errorf("%s(\"\") = true, want false", tc.name)
			}
		})
	}
}

// The record identifiers are cache and receipt inputs. A silent change to one
// would make two different execution policies share a cache entry.
func TestClosedRecordIdentifiersArePinned(t *testing.T) {
	pinned := map[string]string{
		"SpecRevision":     "hardened-1.0.0-rc.1",
		"RecordVersion":    "hardened-capability-evidence-v1",
		"HardenedProfile":  "hardened-profile-v1",
		"ExecutionPolicy":  "hardened-worker-v1",
		"InventoryVersion": "hardened-capability-inventory-v1",
		"ProbedAt":         "pre-domain-entry",
		"PlatformMacOS":    "macos",
		"BackendMacOS":     "macos-sandbox-v1",
	}
	got := map[string]string{
		"SpecRevision":     SpecRevision,
		"RecordVersion":    RecordVersion,
		"HardenedProfile":  HardenedProfile,
		"ExecutionPolicy":  ExecutionPolicy,
		"InventoryVersion": InventoryVersion,
		"ProbedAt":         ProbedAt,
		"PlatformMacOS":    PlatformMacOS,
		"BackendMacOS":     BackendMacOSSandbox,
	}
	for name, want := range pinned {
		if got[name] != want {
			t.Errorf("%s = %q, want %q", name, got[name], want)
		}
	}
}
