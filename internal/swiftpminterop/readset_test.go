package swiftpminterop

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/swiftpmsource"
)

// Portable assurance reports not-observed honestly instead of claiming a
// coverage it cannot have, and still proves the declared closure.
func TestPortableAssuranceReportsNotObservedReads(t *testing.T) {
	fixture := newFixture(t)
	result := fixture.mustClose()
	if result.Reads.Mode != "not-observed" || len(result.Reads.ReceiptIDs) != 0 {
		t.Fatalf("portable read evidence = %#v", result.Reads)
	}
}

// A provider that declines to observe keeps the aggregate honest even though
// the seam was consulted for every selected target.
func TestProviderWithoutObservationKeepsEvidenceHonest(t *testing.T) {
	fixture := newFixture(t)
	provider := &fakeReadSets{}
	fixture.materializeHook = func() { fixture.interop.ReadSets = provider }
	result := fixture.mustClose()
	if result.Reads.Mode != "not-observed" {
		t.Fatalf("read mode = %q", result.Reads.Mode)
	}
	if len(provider.calls) != len(result.Targets) {
		t.Fatalf("read-set seam calls = %v, want one per selected target", provider.calls)
	}
}

// Observed reads must resolve to admitted source or to exactly one selected
// binding node; every resolution is retained as evidence.
func TestObservedReadsResolveToAdmittedSourceOrSelectedBinding(t *testing.T) {
	fixture := newFixture(t)
	var provider *fakeReadSets
	fixture.materializeHook = func() {
		provider = &fakeReadSets{observed: true, reads: map[string][]ObservedRead{}}
		fixture.interop.ReadSets = provider
		fixture.interop.Assurance = closureexec.AssuranceVerified
	}
	fixture.materialize()
	capture, err := fixture.capture()
	if err != nil {
		t.Fatal(err)
	}
	root, err := capture.Packages[0].ProtectedRoot()
	if err != nil {
		t.Fatal(err)
	}
	provider.reads["root:CLib"] = []ObservedRead{
		{Path: filepath.Join(root, "Sources", "CLib", "lib.c"), Class: "source"},
		{Path: filepath.Join(root, "Sources", "CLib", "include", "CLib.h"), Class: "header"},
		{Path: filepath.Join(fixture.sdkRoot, "usr", "include", "stdio.h"), Class: "sdk-header"},
	}
	provider.reads["root:App"] = []ObservedRead{{Path: filepath.Join(root, "Sources", "App", "main.swift"), Class: "source"}}
	result, err := Close(t.Context(), fixture.interop, capture)
	if err != nil {
		t.Fatal(err)
	}
	if result.Reads.Mode != "observed" || len(result.Reads.ReceiptIDs) != 2 {
		t.Fatalf("observed read evidence = %#v", result.Reads)
	}
	classes := map[ResolutionClass]int{}
	for _, resolution := range result.Reads.Resolutions {
		classes[resolution.Class]++
	}
	if classes[ResolvedAdmitted] != 3 || classes[ResolvedBinding] != 1 || classes[ResolvedUndeclared] != 0 {
		t.Fatalf("resolution classes = %#v", classes)
	}
	for _, resolution := range result.Reads.Resolutions {
		if resolution.Class == ResolvedBinding && !resolution.BindingNode.Valid() {
			t.Fatalf("SDK read resolved to no exact binding node: %#v", resolution)
		}
	}
}

// An observed read outside the admitted closure and every selected binding
// root fails closed.
func TestObservedReadOutsideClosureFailsClosed(t *testing.T) {
	fixture := newFixture(t)
	fixture.materializeHook = func() {
		fixture.writeExternal(map[string]string{"secret.h": "int secret(void);\n"}, filepath.Join(fixture.base, "host"))
		fixture.interop.ReadSets = &fakeReadSets{observed: true, reads: map[string][]ObservedRead{
			"root:CLib": {{Path: filepath.Join(fixture.base, "host", "secret.h"), Class: "header"}},
		}}
	}
	_, err := fixture.close()
	requireCode(t, err, CodeHeaderInputUndeclared)
}

// Verified assurance without an observed read set fails closed rather than
// silently degrading to the portable evidence level.
func TestVerifiedAssuranceRequiresObservedReadSet(t *testing.T) {
	fixture := newFixture(t)
	fixture.materializeHook = func() { fixture.interop.Assurance = closureexec.AssuranceVerified }
	_, err := fixture.close()
	requireCode(t, err, CodeHeaderInputUndeclared)
}

// A provider that issues observations without its derivation receipt is
// unauthorized evidence.
func TestObservedReadSetWithoutReceiptIsUnauthorized(t *testing.T) {
	fixture := newFixture(t)
	fixture.materializeHook = func() { fixture.interop.ReadSets = receiptlessReadSets{} }
	_, err := fixture.close()
	requireCode(t, err, CodeDerivationUnauthorized)
}

// A failing read-set observation is a rejection, not a silent downgrade.
func TestFailingReadSetObservationIsRejected(t *testing.T) {
	fixture := newFixture(t)
	fixture.materializeHook = func() { fixture.interop.ReadSets = failingReadSets{} }
	_, err := fixture.close()
	requireCode(t, err, CodeHeaderInputUndeclared)
}

// The manager refuses an incomplete authority and forwards a complete one.
func TestManagerAuthorityContract(t *testing.T) {
	if _, err := NewManager(Config{}); ErrorCode(err) != CodeDerivationUnauthorized {
		t.Fatalf("empty manager config = %v", err)
	}
	var absent *Manager
	if _, err := absent.Close(context.Background(), nil); ErrorCode(err) != CodeDerivationUnauthorized {
		t.Fatalf("absent manager = %v", err)
	}
	fixture := newFixture(t)
	fixture.materialize()
	manager, err := NewManager(fixture.interop)
	if err != nil {
		t.Fatal(err)
	}
	capture, err := fixture.capture()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = manager.Close(t.Context(), capture); err != nil {
		t.Fatal(err)
	}
	if _, err = manager.Close(t.Context(), nil); ErrorCode(err) != CodeDerivationUnauthorized {
		t.Fatalf("absent capture = %v", err)
	}
}

// An incomplete or overlapping external identity is untrusted before any
// admitted byte is read.
func TestIncompleteExternalIdentityIsUntrusted(t *testing.T) {
	t.Run("no contained root", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.materializeHook = func() { fixture.interop.SDK.Roots = nil }
		_, err := fixture.close()
		requireCode(t, err, CodeToolchainUntrusted)
	})
	t.Run("absent root", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.materializeHook = func() { fixture.interop.SDK.Roots = []string{filepath.Join(fixture.base, "missing-sdk")} }
		_, err := fixture.close()
		requireCode(t, err, CodeToolchainUntrusted)
	})
	t.Run("incomplete C-family driver", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.materializeHook = func() { fixture.interop.Clang.Fingerprint = "" }
		_, err := fixture.close()
		requireCode(t, err, CodeTargetPlatformUnsupported)
	})
	t.Run("duplicate system binding", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.materializeHook = func() {
			library := SystemLibrary{Package: "root", Target: "CLib", ModuleMapPath: "/dev/null", Component: systemComponent(fixture.sdkRoot)}
			fixture.interop.SystemLibraries = []SystemLibrary{library, library}
		}
		_, err := fixture.close()
		requireCode(t, err, CodeGraphReferenceInvalid)
	})
	t.Run("system binding names no system target", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.materializeHook = func() {
			fixture.interop.SystemLibraries = []SystemLibrary{{Package: "root", Target: "CLib", ModuleMapPath: "/dev/null", Component: systemComponent(filepath.Join(fixture.systemRoot, "cfoo"))}}
			fixture.writeExternal(map[string]string{"cfoo/keep": "x\n"}, fixture.systemRoot)
		}
		_, err := fixture.close()
		requireCode(t, err, CodeGraphReferenceInvalid)
	})
	t.Run("selected root overlaps the admitted closure", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.materializeHook = func() { fixture.interop.SDK.Roots = []string{fixture.base} }
		_, err := fixture.close()
		requireCode(t, err, CodeToolchainUntrusted)
	})
}

// A selected sysroot is bound as its own external component.
func TestSelectedSysrootIsBoundAsExternalComponent(t *testing.T) {
	fixture := newFixture(t)
	fixture.materializeHook = func() {
		sysroot := filepath.Join(fixture.base, "sysroot")
		fixture.writeExternal(map[string]string{"usr/include/unistd.h": "int close(int);\n"}, sysroot)
		fixture.interop.Sysroot = &ExternalComponent{
			Role: "macos-sysroot", ExecutableRelativePath: "usr/bin/ld", PlatformABI: "darwin-arm64", PolicySelector: "apple-sysroot-v1",
			VersionOutput: "ld 1234", Fingerprint: id('b'), Roots: []string{filepath.Join(sysroot, "usr", "include")},
		}
	}
	result := fixture.mustClose()
	found := false
	for _, node := range result.Records.BindingNodes {
		if node.LogicalKey == "swiftpm.interop.component.macos-sysroot" {
			found = true
		}
	}
	if !found {
		t.Fatal("selected sysroot is absent from the exact selection binding")
	}
}

type receiptlessReadSets struct{}

func (receiptlessReadSets) ObserveReads(context.Context, ReadSetRequest) (ReadSetResult, error) {
	return ReadSetResult{Observed: true}, nil
}

type failingReadSets struct{}

func (failingReadSets) ObserveReads(context.Context, ReadSetRequest) (ReadSetResult, error) {
	return ReadSetResult{}, errors.New("observation unavailable")
}

var _ = swiftpmsource.ProfileID
