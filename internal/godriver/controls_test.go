package godriver

import (
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildsource"
)

// syntheticProbes builds this operation's probe set from the normative
// per-platform inventory, so validator behavior can be proved for both
// platforms on any host.
func syntheticProbes(platform string) []ControlProbe {
	records := nativeControlPlatforms[platform]
	probes := make([]ControlProbe, 0, len(nativeControlInventory))
	for _, name := range nativeControlInventory {
		probes = append(probes, ControlProbe{
			Name: name, Availability: records[name].Availability,
			Mechanism: records[name].Mechanism, ProbedAt: ProbeTiming,
		})
	}
	return probes
}

func validEvidence(platform string) CapabilityEvidence {
	probes := syntheticProbes(platform)
	applied := make([]string, 0, len(probes))
	for _, probe := range probes {
		if probe.Availability == AvailabilityAvailable {
			applied = append(applied, probe.Name)
		}
	}
	return evidenceFromApplied(platform, probes, applied)
}

func TestNativeControlInventoryIsExhaustiveAndClosed(t *testing.T) {
	if len(nativeControlInventory) != 5 {
		t.Fatalf("inventory has %d controls, want exactly five", len(nativeControlInventory))
	}
	if len(nativeControlPlatforms) != 2 {
		t.Fatalf("inventory covers %d platforms, want exactly macOS and Windows", len(nativeControlPlatforms))
	}
	for _, platform := range []string{PlatformMacOS, PlatformWindows} {
		records := nativeControlPlatforms[platform]
		if len(records) != len(nativeControlInventory) {
			t.Fatalf("%s has %d records, want one per inventory control", platform, len(records))
		}
		for _, name := range nativeControlInventory {
			record, present := records[name]
			if !present {
				t.Fatalf("%s has no record for %q", platform, name)
			}
			switch record.Availability {
			case AvailabilityAvailable:
				if record.Mechanism == "" || record.UnavailableReason != "" {
					t.Fatalf("%s/%s available record = %+v", platform, name, record)
				}
			case AvailabilityUnavailable:
				if record.Mechanism != "" || record.UnavailableReason != UnavailableReasonNoPrivateAggregateDomain {
					t.Fatalf("%s/%s unavailable record = %+v", platform, name, record)
				}
			default:
				t.Fatalf("%s/%s availability = %q", platform, name, record.Availability)
			}
			if isDeferredHardenedGuarantee(name) {
				t.Fatalf("inventory names deferred hardened guarantee %q", name)
			}
		}
	}
}

func TestMandatoryControlSetIsExactAndCarriesNoHardenedClaim(t *testing.T) {
	controls := MandatoryControls()
	if len(controls) != 18 {
		t.Fatalf("mandatory controls = %d, want exactly 18", len(controls))
	}
	seen := map[string]bool{}
	for _, name := range controls {
		if seen[name] {
			t.Fatalf("mandatory control %q is duplicated", name)
		}
		seen[name] = true
		if isDeferredHardenedGuarantee(name) {
			t.Fatalf("mandatory control set names deferred hardened guarantee %q", name)
		}
	}
	for _, required := range []string{
		"inventory-native-controls-applied", "closed-capability-evidence-record",
		"fixed-manager-selected-process-graph", "frozen-source-snapshot-integrity",
	} {
		if !seen[required] {
			t.Fatalf("mandatory control set is missing %q", required)
		}
	}
}

func TestDeferredHardenedGuaranteesNeverAppearAndNeverReject(t *testing.T) {
	guarantees := DeferredHardenedGuarantees()
	if len(guarantees) != 6 {
		t.Fatalf("deferred guarantees = %d, want exactly six", len(guarantees))
	}
	for _, name := range guarantees {
		if inInventory(name) {
			t.Fatalf("%q is in the native-control inventory", name)
		}
		for _, control := range MandatoryControls() {
			if control == name {
				t.Fatalf("%q is a mandatory control", name)
			}
		}
		for _, platform := range []string{PlatformMacOS, PlatformWindows} {
			for _, entry := range validEvidence(platform).Controls {
				if entry.Name == name {
					t.Fatalf("%s evidence names %q", platform, name)
				}
			}
		}
	}
	// The absence of every deferred guarantee still yields a valid record for
	// both platforms: no rejection code, no diagnostic, publication permitted.
	for _, platform := range []string{PlatformMacOS, PlatformWindows} {
		if err := validateCapabilityEvidence(validEvidence(platform), platform, syntheticProbes(platform)); err != nil {
			t.Fatalf("%s portable evidence rejected: %v", platform, err)
		}
	}
}

func TestCapabilityEvidenceConsistencyRules(t *testing.T) {
	for _, testCase := range []struct {
		name, platform, code string
		mutate               func(*CapabilityEvidence)
	}{
		{name: "available-native-control-is-applied", platform: PlatformMacOS},
		{name: "unavailable-native-control-does-not-reject", platform: PlatformMacOS},
		{name: "unavailable-control-cannot-be-reported-as-applied", platform: PlatformMacOS, code: CodeCapabilityEvidenceInvalid,
			mutate: func(record *CapabilityEvidence) {
				for index := range record.Controls {
					if record.Controls[index].Availability == AvailabilityUnavailable {
						record.Controls[index].Status = StatusApplied
						return
					}
				}
			}},
		{name: "available-control-cannot-be-reported-as-unavailable", platform: PlatformMacOS, code: CodeCapabilityEvidenceInvalid,
			mutate: func(record *CapabilityEvidence) {
				for index := range record.Controls {
					if record.Controls[index].Availability == AvailabilityAvailable {
						record.Controls[index].Status = StatusUnavailable
						return
					}
				}
			}},
		{name: "unknown-native-control-is-rejected", platform: PlatformWindows, code: CodeCapabilityEvidenceInvalid,
			mutate: func(record *CapabilityEvidence) {
				record.Controls[0].Name = "host-firewall-profile"
			}},
		{name: "missing-native-control-entry-is-rejected", platform: PlatformWindows, code: CodeCapabilityEvidenceInvalid,
			mutate: func(record *CapabilityEvidence) {
				record.Controls = record.Controls[:len(record.Controls)-1]
			}},
		{name: "duplicate-native-control-entry-is-rejected", platform: PlatformMacOS, code: CodeCapabilityEvidenceInvalid,
			mutate: func(record *CapabilityEvidence) {
				record.Controls = append(record.Controls, record.Controls[3])
			}},
		{name: "unknown-evidence-record-version-is-rejected", platform: PlatformMacOS, code: CodeCapabilityEvidenceInvalid,
			mutate: func(record *CapabilityEvidence) { record.RecordVersion = "capability-evidence-v2" }},
		{name: "unprobed-availability-is-rejected", platform: PlatformMacOS, code: CodeCapabilityEvidenceInvalid,
			mutate: func(record *CapabilityEvidence) {
				for index := range record.Controls {
					if record.Controls[index].Availability == AvailabilityUnavailable {
						record.Controls[index].Availability = AvailabilityAvailable
						record.Controls[index].Status = StatusApplied
						return
					}
				}
			}},
		{name: "unknown-probe-timing-is-rejected", platform: PlatformMacOS, code: CodeCapabilityEvidenceInvalid,
			mutate: func(record *CapabilityEvidence) { record.Controls[0].ProbedAt = "install-time" }},
		{name: "foreign-platform-is-rejected", platform: PlatformMacOS, code: CodeCapabilityEvidenceInvalid,
			mutate: func(record *CapabilityEvidence) { record.Platform = PlatformWindows }},
		{name: "hardened-guarantee-claimed-under-portable-policy", platform: PlatformMacOS, code: CodeHardenedClaimForbidden,
			mutate: func(record *CapabilityEvidence) { record.Controls[0].Name = "total-network-denial" }},
		{name: "hardened-execution-policy-in-evidence-record", platform: PlatformMacOS, code: CodeHardenedClaimForbidden,
			mutate: func(record *CapabilityEvidence) { record.ExecutionPolicy = buildmeta.ReservedHardenedExecutionPolicy }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			record := validEvidence(testCase.platform)
			if testCase.mutate != nil {
				testCase.mutate(&record)
			}
			err := validateCapabilityEvidence(record, testCase.platform, syntheticProbes(testCase.platform))
			if DiagnosticCode(err) != testCase.code {
				t.Fatalf("error = %v, want %q", err, testCase.code)
			}
			// A reporting fault is never promoted to the mandatory-control
			// rejection; that boundary is reserved for a control that cannot be
			// applied before the worker starts.
			if DiagnosticCode(err) == CodeControlUnavailable {
				t.Fatal("a capability-evidence fault became a mandatory-control rejection")
			}
		})
	}
}

func TestCapabilityEvidenceIsNotACacheOrReceiptInput(t *testing.T) {
	input := portableVectorInput()
	payload, err := input.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	for _, term := range append([]string{
		CapabilityEvidenceVersion, NativeControlInventoryVersion, ProbeTiming,
		"capability", "controls", "availability", "probed_at",
	}, nativeControlInventory...) {
		if strings.Contains(text, term) {
			t.Fatalf("canonical build input carries capability evidence term %q", term)
		}
	}
	receipt, err := buildmeta.NewReceipt(input, buildmeta.Artifact{
		Path: "bin/golden-tool", SHA256: "sha256:" + strings.Repeat("d", 64), Size: 1234567,
	})
	if err != nil {
		t.Fatal(err)
	}
	receiptBytes, err := receipt.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	for _, term := range append([]string{CapabilityEvidenceVersion, NativeControlInventoryVersion}, nativeControlInventory...) {
		if strings.Contains(string(receiptBytes), term) {
			t.Fatalf("canonical receipt carries capability evidence term %q", term)
		}
	}
}

func TestPortableExecutionPolicyIsABindingCacheInput(t *testing.T) {
	input := portableVectorInput()
	key, err := input.CacheKey()
	if err != nil {
		t.Fatal(err)
	}
	if string(key) != portableVectorCacheKey {
		t.Fatalf("portable cache key = %s, want the accepted rc.5 value %s", key, portableVectorCacheKey)
	}
	hardened := input
	hardened.Policy.ExecutionPolicy = buildmeta.ReservedHardenedExecutionPolicy
	if _, err := hardened.CacheKey(); err == nil {
		t.Fatal("go-v1 accepted the reserved hardened execution policy")
	}
	legacy := input
	legacy.Policy.ExecutionPolicy = ""
	if _, err := legacy.CacheKey(); err == nil {
		t.Fatal("go-v1 accepted a pre-revision input without an execution policy")
	}
}

const portableVectorCacheKey = "sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b"

// portableVectorInput reproduces the accepted rc.5 cache-identity vector input.
func portableVectorInput() buildmeta.Input {
	return buildmeta.Input{
		SchemaVersion: buildmeta.SchemaVersion,
		Driver:        buildmeta.DriverGoV1,
		BuildSource: buildsource.Identity{
			Algorithm: buildsource.Algorithm, ContentSHA256: "sha256:" + strings.Repeat("b", 64),
		},
		BuildRoot: "build",
		Command:   "golden-tool",
		SourceDir: "build/cmd/golden-tool",
		Target:    buildmeta.Target{GOOS: "darwin", GOARCH: "arm64", Tuning: map[string]string{"GOARM64": "v8.0"}},
		Toolchain: buildmeta.Toolchain{
			Algorithm: buildmeta.ToolchainAlgorithm, GoRelpath: buildmeta.ToolchainGoRelpath,
			GoVersion: "go version go1.26.1 darwin/arm64", ContentSHA256: "sha256:" + strings.Repeat("c", 64),
		},
		Policy: buildmeta.FixedPolicy(),
	}
}
