package swiftpminterop

import (
	"testing"

	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/swiftpmsource"
)

// S02: a Swift target depending on a nested C target keeps separate capture
// targets, inventories every header SwiftPM itself omits, reproduces the
// generated module map, and declares one exact C ABI boundary.
func TestS02SwiftDependsOnNestedCTarget(t *testing.T) {
	fixture := newFixture(t)
	result := fixture.mustClose()
	swift, clang := mustTarget(t, result, "root:App"), mustTarget(t, result, "root:CLib")
	if swift.Kind != KindSwift || clang.Kind != KindClang {
		t.Fatalf("target kinds = %s/%s", swift.Kind, clang.Kind)
	}
	if len(clang.Headers) != 1 || clang.Headers[0].Relative != "Sources/CLib/include/CLib.h" || !clang.Headers[0].SHA256.Valid() {
		t.Fatalf("public header inventory = %#v", clang.Headers)
	}
	if clang.ModuleMap == nil || !clang.ModuleMap.Generated || clang.ModuleMap.Relative != "Sources/CLib/include/module.modulemap" {
		t.Fatalf("module map evidence = %#v", clang.ModuleMap)
	}
	if len(result.Boundaries) != 1 || result.Boundaries[0].Mode != closuregraph.InteropCABI {
		t.Fatalf("boundaries = %#v", result.Boundaries)
	}
	if result.Boundaries[0].Provider != "root:CLib" || result.Boundaries[0].Consumer != "root:App" {
		t.Fatalf("boundary endpoints = %#v", result.Boundaries[0])
	}
}

// S03: one Clang target may carry C, C++, Objective-C, and Objective-C++
// sources on an accepted Darwin profile; every language is recorded and the
// boundary binds the Objective-C runtime the target actually needs.
func TestS03SingleClangTargetCarriesEveryCFamilyLanguage(t *testing.T) {
	fixture := newFixture(t)
	fixture.addFiles(map[string]string{
		"Sources/CLib/lib.cpp": "#include \"CLib.h\"\nint cxx_value(void) { return 2; }\n",
		"Sources/CLib/lib.m":   "#include \"CLib.h\"\nint objc_value(void) { return 3; }\n",
		"Sources/CLib/lib.mm":  "#include \"CLib.h\"\nint objcxx_value(void) { return 4; }\n",
	})
	target := fixture.target("CLib")
	target.Sources = append(target.Sources, "Sources/CLib/lib.cpp", "Sources/CLib/lib.m", "Sources/CLib/lib.mm")
	result := fixture.mustClose()
	clang := mustTarget(t, result, "root:CLib")
	want := []Language{LanguageC, LanguageCXX, LanguageObjC, LanguageObjCXX}
	if joinLanguages(clang.Languages) != joinLanguages(want) {
		t.Fatalf("languages = %v, want %v", clang.Languages, want)
	}
	if result.Boundaries[0].Mode != closuregraph.InteropObjCRuntime || result.Boundaries[0].Runtime != ObjCRuntimeContract {
		t.Fatalf("boundary = %#v", result.Boundaries[0])
	}
	if result.Boundaries[0].ToolchainRole != fixture.interop.ClangCXX.Role {
		t.Fatalf("boundary toolchain = %q", result.Boundaries[0].ToolchainRole)
	}
}

// S04: Swift and C-family source in one target is rejected before any header,
// module, or boundary analysis.
func TestS04MixedLanguageTargetIsRejected(t *testing.T) {
	fixture := newFixture(t)
	fixture.addFiles(map[string]string{"Sources/App/helper.c": "int helper(void) { return 1; }\n"})
	fixture.target("App").Sources = append(fixture.target("App").Sources, "Sources/App/helper.c")
	_, err := fixture.close()
	requireCode(t, err, CodeMixedLanguageTarget)
}

// S05: direct Swift/C++ interoperation is admitted only with the explicit
// opt-in and an accepted destination profile.
func TestS05DirectSwiftCxxInteropWithDeclaredMode(t *testing.T) {
	fixture := newCxxFixture(t)
	fixture.target("App").Settings = []swiftpmsource.BuildSetting{{Kind: "interoperabilityMode", Value: "Cxx"}}
	result := fixture.mustClose()
	if len(result.Boundaries) != 1 || result.Boundaries[0].Mode != closuregraph.InteropCXX {
		t.Fatalf("boundaries = %#v", result.Boundaries)
	}
	if result.Boundaries[0].ABI != CXXABIContract || result.Boundaries[0].Runtime != CXXRuntimeContract {
		t.Fatalf("C++ boundary contract = %#v", result.Boundaries[0])
	}
	if !mustTarget(t, result, "root:App").CxxInteropMode {
		t.Fatal("consumer interoperability mode was not recorded")
	}
}

// S06: the same graph without the opt-in fails closed; SwiftPM never
// propagates interoperabilityMode implicitly.
func TestS06MissingCxxInteropModeFailsClosed(t *testing.T) {
	fixture := newCxxFixture(t)
	_, err := fixture.close()
	requireCode(t, err, CodeInteropUndeclared)
}

// S05 restriction: an accepted opt-in still requires an accepted destination
// profile for direct C++ interoperation.
func TestS05CxxInteropRequiresAcceptedDestinationProfile(t *testing.T) {
	fixture := newCxxFixture(t)
	fixture.target("App").Settings = []swiftpmsource.BuildSetting{{Kind: "interoperabilityMode", Value: "Cxx"}}
	fixture.materializeHook = func() { fixture.interop.Profile.CxxInterop = false }
	_, err := fixture.close()
	requireCode(t, err, CodeTargetPlatformUnsupported)
}

// S07: an Objective-C target imported by Swift binds the Objective-C runtime
// boundary on the accepted Darwin profile.
func TestS07SwiftImportsObjectiveCTarget(t *testing.T) {
	fixture := newObjectiveCFixture(t, "m")
	result := fixture.mustClose()
	if result.Boundaries[0].Mode != closuregraph.InteropObjCRuntime {
		t.Fatalf("boundary = %#v", result.Boundaries[0])
	}
	if !containsLanguage(mustTarget(t, result, "root:ObjCLib").Languages, LanguageObjC) {
		t.Fatal("Objective-C language was not recorded")
	}
}

// S08: an Objective-C++ implementation exposed through Objective-C headers is
// admitted with the Objective-C++ language and C++ driver recorded.
func TestS08ObjectiveCXXBehindObjectiveCHeaders(t *testing.T) {
	fixture := newObjectiveCFixture(t, "mm")
	result := fixture.mustClose()
	if result.Boundaries[0].Mode != closuregraph.InteropObjCRuntime {
		t.Fatalf("boundary = %#v", result.Boundaries[0])
	}
	if result.Boundaries[0].ToolchainRole != fixture.interop.ClangCXX.Role {
		t.Fatalf("Objective-C++ boundary toolchain = %q", result.Boundaries[0].ToolchainRole)
	}
}

// S09: the same Objective-C package on a destination without an accepted
// runtime profile is unsupported rather than best-effort.
func TestS09ObjectiveCOnUnvalidatedDestinationIsUnsupported(t *testing.T) {
	fixture := newObjectiveCFixture(t, "m")
	fixture.useLinuxDestination()
	_, err := fixture.close()
	requireCode(t, err, CodeTargetPlatformUnsupported)
}

// A target whose declared destination triple has no accepted profile is
// rejected before any admitted byte is classified.
func TestUnacceptedDestinationTripleIsRejectedBeforeClassification(t *testing.T) {
	fixture := newFixture(t)
	fixture.materializeHook = func() { fixture.interop.Profile.TargetTriples = []string{"x86_64-unknown-linux-gnu"} }
	_, err := fixture.close()
	requireCode(t, err, CodeTargetPlatformUnsupported)
}

// Unsafe build settings on a selected C-family target are rejected.
func TestUnsafeFlagsAreRejectedBeforeInteropClosure(t *testing.T) {
	fixture := newFixture(t)
	fixture.target("CLib").Settings = []swiftpmsource.BuildSetting{{Kind: "unsafeFlags", Value: "-I/usr/local/include", Unsafe: true}}
	_, err := fixture.close()
	requireCode(t, err, CodeUnsafeSettingForbidden)
}

func newCxxFixture(t *testing.T) *fixture {
	t.Helper()
	fixture := newFixture(t)
	delete(fixture.files, "Sources/CLib/lib.c")
	delete(fixture.files, "Sources/CLib/include/CLib.h")
	fixture.addFiles(map[string]string{
		"Sources/CxxLib/lib.cpp":            "#include \"CxxLib.hpp\"\nint Value::get() const { return 42; }\n",
		"Sources/CxxLib/include/CxxLib.hpp": "#pragma once\nstruct Value { int get() const; };\n",
		"Sources/App/main.swift":            "import CxxLib\nprint(1)\n",
	})
	fixture.manifest.Targets = []swiftpmsource.Target{
		{Name: "App", Type: "executable", Path: "Sources/App", Sources: []string{"Sources/App/main.swift"}, Dependencies: []swiftpmsource.TargetDependency{{Name: "CxxLib"}}},
		{Name: "CxxLib", Type: "regular", Path: "Sources/CxxLib", Sources: []string{"Sources/CxxLib/lib.cpp"}},
	}
	return fixture
}

func newObjectiveCFixture(t *testing.T, extension string) *fixture {
	t.Helper()
	fixture := newFixture(t)
	delete(fixture.files, "Sources/CLib/lib.c")
	delete(fixture.files, "Sources/CLib/include/CLib.h")
	fixture.addFiles(map[string]string{
		"Sources/ObjCLib/lib." + extension:  "#include \"ObjCLib.h\"\nint objc_value(void) { return 7; }\n",
		"Sources/ObjCLib/include/ObjCLib.h": "#ifndef OBJCLIB_H\n#define OBJCLIB_H\nint objc_value(void);\n#endif\n",
		"Sources/App/main.swift":            "import ObjCLib\nprint(1)\n",
	})
	fixture.manifest.Targets = []swiftpmsource.Target{
		{Name: "App", Type: "executable", Path: "Sources/App", Sources: []string{"Sources/App/main.swift"}, Dependencies: []swiftpmsource.TargetDependency{{Name: "ObjCLib"}}},
		{Name: "ObjCLib", Type: "regular", Path: "Sources/ObjCLib", Sources: []string{"Sources/ObjCLib/lib." + extension}},
	}
	return fixture
}

func mustTarget(t *testing.T, result *Result, key string) TargetInterop {
	t.Helper()
	for _, target := range result.Targets {
		if target.Package+":"+target.Target == key {
			return target
		}
	}
	t.Fatalf("result has no target %s", key)
	return TargetInterop{}
}

// S05/S06: a conditional `.interoperabilityMode(.Cxx)` opt-in is a declaration,
// not a destination fact. Testing the destination-evaluated setting in the
// declaration-level gate hard-rejected a package whose entire C++ declaration is
// pruned on a destination, and would have published two different capture
// records for one admitted closure.
func TestS05ConditionalCxxInteropOptInIsSelectionNeutral(t *testing.T) {
	build := func(linux bool) *Result {
		t.Helper()
		fixture := newCxxFixture(t)
		fixture.target("App").Dependencies = []swiftpmsource.TargetDependency{{Name: "CxxLib", Condition: swiftpmCondition("platform=macos")}}
		fixture.target("App").Settings = []swiftpmsource.BuildSetting{{Kind: "interoperabilityMode", Value: "Cxx", Condition: swiftpmCondition("platform=macos")}}
		if linux {
			fixture.useLinuxDestination()
		}
		result, err := fixture.close()
		if err != nil {
			t.Fatalf("conditional C++ opt-in closure failed (linux=%v): %v", linux, err)
		}
		return result
	}
	darwin, linux := build(false), build(true)
	if darwin.GraphDigest != linux.GraphDigest {
		t.Fatalf("a conditional C++ opt-in changed the selection-neutral capture: %s != %s", darwin.GraphDigest, linux.GraphDigest)
	}
	if darwin.Boundaries[0].Mode != closuregraph.InteropCXX || linux.Boundaries[0].Mode != closuregraph.InteropCXX {
		t.Fatalf("boundary modes = %q / %q", darwin.Boundaries[0].Mode, linux.Boundaries[0].Mode)
	}
	if !darwin.Boundaries[0].Selected || linux.Boundaries[0].Selected {
		t.Fatalf("selection verdicts = darwin %v, linux %v", darwin.Boundaries[0].Selected, linux.Boundaries[0].Selected)
	}
	darwinApp, linuxApp := mustTarget(t, darwin, "root:App"), mustTarget(t, linux, "root:App")
	if !darwinApp.CxxInteropDeclared || !linuxApp.CxxInteropDeclared {
		t.Fatal("the condition-neutral declaration was not recorded on both destinations")
	}
	if !darwinApp.CxxInteropMode || linuxApp.CxxInteropMode {
		t.Fatalf("destination verdicts = darwin %v, linux %v", darwinApp.CxxInteropMode, linuxApp.CxxInteropMode)
	}
	if darwin.EvidenceDigest == linux.EvidenceDigest {
		t.Fatal("destination did not change the exact interop evidence record")
	}
}

// S05/S06 control: when the C++ dependency itself is unconditional, the target
// is selected on every destination, so a destination without an accepted C++
// standard/toolchain profile still fails closed before any boundary is derived.
func TestS05ConditionalCxxInteropOptInStillRequiresAnAcceptedProfile(t *testing.T) {
	fixture := newCxxFixture(t)
	fixture.target("App").Settings = []swiftpmsource.BuildSetting{{Kind: "interoperabilityMode", Value: "Cxx", Condition: swiftpmCondition("platform=macos")}}
	fixture.useLinuxDestination()
	_, err := fixture.close()
	requireCode(t, err, CodeTargetPlatformUnsupported)
}

// S10: `.C` and `.M` are C++ and Objective-C++ translation units. The Clang
// driver is case-sensitive on exactly those two suffixes — `clang -### -c up.C`
// selects `-x c++` and `clang -### -c up.M` selects `-x objective-c++` — so
// lowercasing the extension recorded `c` and `objective-c` evidence for a
// translation unit the compiler compiles as C++ and Objective-C++, and every
// gate keyed on the recorded language set was bypassed.
func TestS10UpperCaseExtensionsSelectTheCompilersLanguage(t *testing.T) {
	t.Run("dot C is a C++ translation unit", func(t *testing.T) {
		fixture := newUpperCaseFixture(t, "Sources/CLib/impl.C")
		clang := mustTarget(t, fixture.mustClose(), "root:CLib")
		if joinLanguages(clang.Languages) != string(LanguageCXX) {
			t.Fatalf("languages = %v, want %v", clang.Languages, LanguageCXX)
		}
	})
	t.Run("dot M is an Objective-C++ translation unit", func(t *testing.T) {
		fixture := newUpperCaseFixture(t, "Sources/CLib/impl.M")
		clang := mustTarget(t, fixture.mustClose(), "root:CLib")
		if joinLanguages(clang.Languages) != string(LanguageObjCXX) {
			t.Fatalf("languages = %v, want %v", clang.Languages, LanguageObjCXX)
		}
	})
	// The restricted-profile gate reads the recorded language set, so a `.C`
	// source on a destination with no accepted C++ standard profile is
	// unsupported exactly as `.cpp` is.
	t.Run("dot C is gated by the C++ standard profile", func(t *testing.T) {
		fixture := newUpperCaseFixture(t, "Sources/CLib/impl.C")
		fixture.materializeHook = func() { fixture.interop.Profile.CxxStandardModes = nil }
		_, err := fixture.close()
		requireCode(t, err, CodeTargetPlatformUnsupported)
	})
	// The Objective-C++ runtime gate likewise.
	t.Run("dot M is gated by the Objective-C runtime profile", func(t *testing.T) {
		fixture := newUpperCaseFixture(t, "Sources/CLib/impl.M")
		fixture.materializeHook = func() { fixture.interop.Profile.ObjectiveCRuntime = "" }
		_, err := fixture.close()
		requireCode(t, err, CodeTargetPlatformUnsupported)
	})
	// A Swift consumer that declares the opt-in against a `.C` provider binds
	// the direct C++ boundary, which the accepted-profile gate then governs.
	// While `.C` reported `c`, this graph bound a plain C ABI boundary and the
	// gate never ran.
	t.Run("dot C provider binds the direct C++ boundary", func(t *testing.T) {
		fixture := newUpperCaseCxxFixture(t)
		fixture.target("App").Settings = []swiftpmsource.BuildSetting{{Kind: "interoperabilityMode", Value: "Cxx"}}
		result := fixture.mustClose()
		if len(result.Boundaries) != 1 || result.Boundaries[0].Mode != closuregraph.InteropCXX {
			t.Fatalf("boundaries = %#v", result.Boundaries)
		}
		if !containsLanguage(result.Boundaries[0].ProviderLanguages, LanguageCXX) {
			t.Fatalf("provider languages = %v", result.Boundaries[0].ProviderLanguages)
		}
	})
	t.Run("dot C provider without an accepted profile is unsupported", func(t *testing.T) {
		fixture := newUpperCaseCxxFixture(t)
		fixture.target("App").Settings = []swiftpmsource.BuildSetting{{Kind: "interoperabilityMode", Value: "Cxx"}}
		fixture.materializeHook = func() { fixture.interop.Profile.CxxInterop = false }
		_, err := fixture.close()
		requireCode(t, err, CodeTargetPlatformUnsupported)
	})
	// S06 against a `.C` provider: a Swift consumer that declares no
	// `.interoperabilityMode(.Cxx)` is rejected, because SwiftPM never
	// propagates the mode implicitly.
	t.Run("dot C provider without the Swift opt-in is rejected", func(t *testing.T) {
		fixture := newUpperCaseCxxFixture(t)
		_, err := fixture.close()
		requireCode(t, err, CodeInteropUndeclared)
	})
}

// newUpperCaseFixture replaces the C implementation of the standard C-family
// target with one upper-case-suffixed source.
func newUpperCaseFixture(t *testing.T, source string) *fixture {
	t.Helper()
	fixture := newFixture(t)
	delete(fixture.files, "Sources/CLib/lib.c")
	fixture.addFiles(map[string]string{source: "#include \"CLib.h\"\nint value(void) { return 1; }\n"})
	fixture.target("CLib").Sources = []string{source}
	return fixture
}

// newUpperCaseCxxFixture is the S05/S06 graph with a `.C` implementation.
func newUpperCaseCxxFixture(t *testing.T) *fixture {
	t.Helper()
	fixture := newCxxFixture(t)
	delete(fixture.files, "Sources/CxxLib/lib.cpp")
	fixture.addFiles(map[string]string{"Sources/CxxLib/impl.C": "#include \"CxxLib.hpp\"\nint Value::get() const { return 42; }\n"})
	fixture.target("CxxLib").Sources = []string{"Sources/CxxLib/impl.C"}
	return fixture
}
