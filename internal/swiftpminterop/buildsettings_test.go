package swiftpminterop

import (
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/swiftpmsource"
)

// swiftpmSetting builds a build-setting record in exactly the shape
// `swiftpmsource.decodeBuildSetting` produces from real dump-package JSON: the
// folded `swiftpm-setting` kind, the raw setting object retained verbatim in
// `Value`, and `Unsafe` set by the same substring rule.
//
// Every payload below was taken from `swift package dump-package` run against a
// manifest declaring that setting on the accepted profile (SwiftPM 6.3.2, Swift
// 6.3.2, tools versions 5.9 and 6.2).
func swiftpmSetting(payload string) swiftpmsource.BuildSetting {
	return swiftpmsource.BuildSetting{Kind: "swiftpm-setting", Value: payload, Unsafe: strings.Contains(payload, `"unsafeFlags"`)}
}

// H24: the macro oracle behind the expanded-identifier positions, and the
// build-setting kind axis it stands on.
//
// H23 closed the two identifier positions the compiler macro-expands by asking
// "is this identifier macro-defined?". The answer came from an oracle populated
// only by source `#define`s, and a SwiftPM `.define` build setting is a macro
// the compiler binds that no admitted file spells. Both closed positions
// reopened one level down.
//
// D1 and D2 were confirmed against the pinned Apple Clang (21.0.0,
// `clang-2100.1.1.101`, `arm64-apple-darwin25.5.0`) with `-fmodules` against a
// module whose only header is an `#error` marker:
//
//	clang -fsyntax-only -fmodules -I secret -Dprotocol=import d1.m
//	  -> While building module 'SecretKit' … error: SECRET_MODULE_WAS_READ
//	clang -fsyntax-only -fmodules -I secret d1.m                    (control)
//	  -> clean
//	clang -fsyntax-only -fmodules -I secret d2.m                    (control)
//	  -> fatal error: module 'NoSuchKitXYZ' not found
//	clang -fsyntax-only -fmodules -I secret -DNoSuchKitXYZ=SecretKit d2.m
//	  -> While building module 'SecretKit' … error: SECRET_MODULE_WAS_READ
func TestH24BuildSettingMacroOracleAndKindAxis(t *testing.T) {
	// D1: the `@`-keyword position, rebound by a build setting instead of by a
	// `#define`. The body is ordinary Objective-C the scanner admits; the
	// setting is the whole vector, and it rejects before any file is scanned.
	t.Run("D1 at-position identifier rebound by a define build setting", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\n@ protocol SecretKit;\nint value(void) { return 1; }\n"
		fixture.target("CLib").Settings = []swiftpmsource.BuildSetting{swiftpmSetting(`{"kind":{"define":{"_0":"protocol=import"}},"tool":"c"}`)}
		result, err := fixture.close()
		requireCode(t, err, CodeHeaderInputUndeclared)
		if result != nil {
			t.Fatalf("rejected closure published a result: %#v", result)
		}
	})
	// D2: the `@import` module name, aliased by a build setting. Without the
	// route the closure admits and retains `ModuleImport: true, Spelling:
	// "CLib", ExpandedName: true` — its own admitted module, satisfying
	// moduleDeclared — while the compiler resolves SecretKit.
	t.Run("D2 module import name aliased by a define build setting", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\n@import CLib;\nint value(void) { return 1; }\n"
		fixture.target("CLib").Settings = []swiftpmsource.BuildSetting{swiftpmSetting(`{"kind":{"define":{"_0":"CLib=SecretKit"}},"tool":"c"}`)}
		result, err := fixture.close()
		requireCode(t, err, CodeHeaderInputUndeclared)
		if result != nil {
			t.Fatalf("rejected closure published a result: %#v", result)
		}
	})
	// D2 control: the identical source with no aliasing setting admits and
	// records the module the compiler really imports, so the route rejects the
	// aliased binding and nothing else.
	t.Run("D2 control the same import admits without the setting", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\n@import CLib;\nint value(void) { return 1; }\n"
		result := fixture.mustClose()
		imports := 0
		for _, reference := range mustTarget(t, result, "root:CLib").Includes {
			if reference.ModuleImport {
				if reference.Spelling != "CLib" {
					t.Fatalf("recorded module import = %#v", reference)
				}
				imports++
			}
		}
		if imports != 1 {
			t.Fatalf("recorded %d Clang module imports", imports)
		}
	})
	// A setting the destination prunes never reaches a compiler invocation, so
	// it binds nothing and must not reject. Conditions are the one axis where
	// the oracle is allowed to be narrower than the declaration.
	t.Run("a pruned define setting binds nothing", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\n@ protocol SecretKit;\nint value(void) { return 1; }\n"
		setting := swiftpmSetting(`{"kind":{"define":{"_0":"protocol=import"}},"tool":"c"}`)
		setting.Condition = swiftpmCondition("platform=linux")
		fixture.target("CLib").Settings = []swiftpmsource.BuildSetting{setting}
		fixture.mustClose()
	})
}

// H24 kind axis: every build-setting kind PackageDescription vends on the
// accepted profile, with its disposition pinned.
//
// The axis is closed reject-by-default: a kind is admitted only when it is
// provably macro-inert and resolution-inert. `define` is the only kind that
// reaches the compiler as `-D` and is routed through both macro rules;
// `headerSearchPath` is the only non-`unsafeFlags` kind that reaches it as
// `-I`; `linkedLibrary`/`linkedFramework` reach the linker and are gated on
// component declaration; everything else is a language mode, a feature name, a
// diagnostic severity, or an isolation default, and an unknown kind rejects.
func TestH24BuildSettingKindAxisDisposition(t *testing.T) {
	for name, expect := range map[string]struct {
		payload string
		code    Code
	}{
		// Routed: admitted when the bound name reaches no gated position.
		"define plain name":  {payload: `{"kind":{"define":{"_0":"FEATURE"}},"tool":"c"}`},
		"define keyed value": {payload: `{"kind":{"define":{"_0":"FEATURE=1"}},"tool":"c"}`},
		"define on the swift tool": {
			payload: `{"kind":{"define":{"_0":"SWIFTDEF"}},"tool":"swift"}`,
		},
		"define binding an at-position identifier": {
			payload: `{"kind":{"define":{"_0":"protocol=import"}},"tool":"c"}`,
			code:    CodeHeaderInputUndeclared,
		},
		"define binding the import spelling": {
			payload: `{"kind":{"define":{"_0":"import=protocol"}},"tool":"c"}`,
			code:    CodeHeaderInputUndeclared,
		},
		"define naming no identifier": {
			payload: `{"kind":{"define":{"_0":"=1"}},"tool":"c"}`,
			code:    CodeUnsafeSettingForbidden,
		},
		// Resolution axis: the search path changes where a reference resolves,
		// which this stage's include closure cannot follow.
		"header search path": {
			payload: `{"kind":{"headerSearchPath":{"_0":"include"}},"tool":"c"}`,
			code:    CodeUnsafeSettingForbidden,
		},
		// Unbounded on both axes. The kind rejects on its own name, so the
		// record's `Unsafe` flag is corroboration rather than the only gate.
		"unsafe flags": {
			payload: `{"kind":{"unsafeFlags":{"_0":["-I/tmp","-DX=1"]}},"tool":"c"}`,
			code:    CodeUnsafeSettingForbidden,
		},
		// Link edges: inert on both axes, still gated on component declaration.
		"linked library the sdk declares":   {payload: `{"kind":{"linkedLibrary":{"_0":"c"}},"tool":"linker"}`},
		"linked framework the sdk declares": {payload: `{"kind":{"linkedFramework":{"_0":"Foundation"}},"tool":"linker"}`},
		"linked library no component declares": {
			payload: `{"kind":{"linkedLibrary":{"_0":"z"}},"tool":"linker"}`,
			code:    CodeToolchainUntrusted,
		},
		"linked framework no component declares": {
			payload: `{"kind":{"linkedFramework":{"_0":"SecretKit"}},"tool":"linker"}`,
			code:    CodeToolchainUntrusted,
		},
		// Proven inert on both axes.
		"interoperability mode":       {payload: `{"kind":{"interoperabilityMode":{"_0":"Cxx"}},"tool":"swift"}`},
		"enable upcoming feature":     {payload: `{"kind":{"enableUpcomingFeature":{"_0":"ExistentialAny"}},"tool":"swift"}`},
		"enable experimental feature": {payload: `{"kind":{"enableExperimentalFeature":{"_0":"Foo"}},"tool":"swift"}`},
		"swift language mode":         {payload: `{"kind":{"swiftLanguageMode":{"_0":"5"}},"tool":"swift"}`},
		"treat all warnings":          {payload: `{"kind":{"treatAllWarnings":{"_0":"error"}},"tool":"c"}`},
		"treat one warning":           {payload: `{"kind":{"treatWarning":{"_0":"unused","_1":"error"}},"tool":"c"}`},
		"enable warning":              {payload: `{"kind":{"enableWarning":{"_0":"all"}},"tool":"c"}`},
		"disable warning":             {payload: `{"kind":{"disableWarning":{"_0":"unused"}},"tool":"c"}`},
		"strict memory safety":        {payload: `{"kind":{"strictMemorySafety":{}},"tool":"swift"}`},
		"default isolation":           {payload: `{"kind":{"defaultIsolation":{"_0":"MainActor"}},"tool":"swift"}`},
		// Reject-by-default: a kind a later SwiftPM release adds is not in the
		// table, so it fails closed instead of being admitted unexamined.
		"a kind this profile does not name": {
			payload: `{"kind":{"someFutureKind":{"_0":"x"}},"tool":"c"}`,
			code:    CodeUnsafeSettingForbidden,
		},
		// Shapes this stage cannot read are the same class as a kind it cannot
		// prove: the effect on both axes is unknown.
		"a payload with no kind member": {
			payload: `{"tool":"c"}`,
			code:    CodeUnsafeSettingForbidden,
		},
		"a payload with two kinds": {
			payload: `{"kind":{"define":{"_0":"A"},"headerSearchPath":{"_0":"include"}},"tool":"c"}`,
			code:    CodeUnsafeSettingForbidden,
		},
		"a payload whose operand is not a string": {
			payload: `{"kind":{"define":{"_0":1}},"tool":"c"}`,
			code:    CodeUnsafeSettingForbidden,
		},
		"a payload whose kind is not an object": {
			payload: `{"kind":"define","tool":"c"}`,
			code:    CodeUnsafeSettingForbidden,
		},
	} {
		t.Run(name, func(t *testing.T) {
			fixture := newFixture(t)
			fixture.target("CLib").Settings = []swiftpmsource.BuildSetting{swiftpmSetting(expect.payload)}
			result, err := fixture.close()
			if expect.code != "" {
				requireCode(t, err, expect.code)
				if result != nil {
					t.Fatalf("rejected closure published a result: %#v", result)
				}
				return
			}
			if err != nil {
				t.Fatalf("admitted kind rejected: %v", err)
			}
			if references := mustTarget(t, result, "root:CLib").Includes; len(references) != 2 {
				t.Fatalf("scanned references = %#v", references)
			}
		})
	}
	// The positive path the whole axis exists to preserve: an ordinary
	// C-family target carrying an ordinary benign define, a declared link
	// edge, and plain includes still admits with its references intact.
	t.Run("a normal target with benign settings still admits", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.target("CLib").Settings = []swiftpmsource.BuildSetting{
			swiftpmSetting(`{"kind":{"define":{"_0":"CLIB_FEATURE=1"}},"tool":"c"}`),
			swiftpmSetting(`{"kind":{"define":{"_0":"NDEBUG"}},"tool":"cxx"}`),
			swiftpmSetting(`{"kind":{"linkedLibrary":{"_0":"c"}},"tool":"linker"}`),
			swiftpmSetting(`{"kind":{"linkedFramework":{"_0":"Foundation"}},"tool":"linker"}`),
			swiftpmSetting(`{"kind":{"treatAllWarnings":{"_0":"error"}},"tool":"c"}`),
		}
		result := fixture.mustClose()
		spellings := []string{}
		for _, reference := range mustTarget(t, result, "root:CLib").Includes {
			spellings = append(spellings, reference.Spelling)
		}
		if len(spellings) != 2 || spellings[0] != "CLib.h" || spellings[1] != "stdio.h" {
			t.Fatalf("scanned references = %v", spellings)
		}
	})
}

// The C++ interoperability gate reads the decoded kind, not the folded record
// kind, so a mode SwiftPM actually emits reaches it. Before the kind axis
// existed the gate only ever saw a directly constructed record.
func TestH24InteroperabilityModeIsReadFromTheDecodedKind(t *testing.T) {
	for name, expect := range map[string]struct {
		setting swiftpmsource.BuildSetting
		want    bool
	}{
		"the shape SwiftPM emits": {
			setting: swiftpmSetting(`{"kind":{"interoperabilityMode":{"_0":"Cxx"}},"tool":"swift"}`),
			want:    true,
		},
		"the C mode SwiftPM emits": {
			setting: swiftpmSetting(`{"kind":{"interoperabilityMode":{"_0":"C"}},"tool":"swift"}`),
		},
		"a directly constructed record": {
			setting: swiftpmsource.BuildSetting{Kind: "interoperabilityMode", Value: "Cxx"},
			want:    true,
		},
		"an unrelated decoded kind": {
			setting: swiftpmSetting(`{"kind":{"define":{"_0":"Cxx"}},"tool":"c"}`),
		},
	} {
		t.Run(name, func(t *testing.T) {
			if got := cxxInteropSetting(expect.setting); got != expect.want {
				t.Fatalf("cxxInteropSetting = %v, want %v", got, expect.want)
			}
		})
	}
}

// H25 finding Q: the `.define` build setting's BODY, routed through the same
// phase-4 analyzer a source `#define` body goes through.
//
// H24 closed the setting's NAME against both macro-oracle rules and left the
// replacement list unread, which reopened round 9 finding M one input down: a
// body can deliver a channel keyword into a call site the token scanner reads
// as ordinary content, because the call site names only the macro. Confirmed on
// the pinned Apple Clang (21.0.0, `clang-2100.1.1.101`,
// `arm64-apple-darwin25.5.0`), where `payload.bin` holds a marker string and
// `SecretKit`'s only header is an `#error` marker:
//
//	clang -c -D'A=__asm__' d.c        with  A(".incbin \"payload.bin\"");
//	  -> exit 0; payload.bin's bytes present in d.o at 0x180, and the object is
//	     byte-identical to the direct `__asm__(…)` control
//	     (sha256 8e639ccfe3a3ebcab9c95b2aa6a0872f71c1e60ece92d1eb6cd41adfa40f7334)
//	clang -c d.c                                             (negative control)
//	  -> exit 1: `A` is undeclared, so the setting is the entire vector
//	clang -fsyntax-only -fmodules -I secret -D'A=_Pragma' p.c
//	                                 with  A("clang module import SecretKit")
//	  -> While building module 'SecretKit' … error: SECRET_MODULE_WAS_READ
//	clang -fsyntax-only -fmodules -I secret p.c              (control)
//	  -> expected parameter declarator; nothing imported
//
// Each vector is asserted twice: once with the macro bound by the build setting
// and once with the identical body bound by a source `#define` and no setting
// at all. Both spellings must reject with the same code, which is what makes
// this one analyzer over two inputs rather than two analyzers that happen to
// agree.
func TestH25BuildSettingDefineBodyIsAnalyzedLikeASourceDefine(t *testing.T) {
	for name, expect := range map[string]struct {
		operand string
		body    string
		use     string
		want    Code
	}{
		// Q1/Q1b: the reserved and bare inline-assembly keywords, delivered as
		// the replacement list. `.incbin` reads an arbitrary file through the
		// assembler stage portable mode declares it does not admit.
		"Q1 reserved asm keyword as the body": {
			operand: "A=__asm__", body: "A __asm__",
			use:  "A(\".incbin \\\"payload.bin\\\"\");",
			want: CodeTargetPlatformUnsupported,
		},
		"Q1b bare asm keyword as the body": {
			operand: "A=asm", body: "A asm",
			use:  "A(\".incbin \\\"payload.bin\\\"\");",
			want: CodeTargetPlatformUnsupported,
		},
		// Q3/Q4: the two pragma operators. Bare, they are already
		// unclassifiable operands; bound to a name, the operand arrives from the
		// call site the compiler expands.
		"Q3 _Pragma operator as the body": {
			operand: "A=_Pragma", body: "A _Pragma",
			use:  "A(\"clang module import SecretKit\")",
			want: CodeHeaderInputUndeclared,
		},
		"Q4 __pragma operator as the body": {
			operand: "A=__pragma", body: "A __pragma",
			use:  "A(clang module import SecretKit)",
			want: CodeHeaderInputUndeclared,
		},
		// Q5: the same keyword built by a definition-resolvable paste, which
		// collapseMacroPastes joins before the channel scan sees it.
		"Q5 asm keyword pasted from fixed fragments": {
			operand: "A=__as##m__", body: "A __as##m__",
			use:  "A(\".incbin \\\"payload.bin\\\"\");",
			want: CodeTargetPlatformUnsupported,
		},
		// Q6: the function-like form, whose paste takes a fragment from the
		// call site. The parameter list has to be parsed out of the operand for
		// the parameter-paste rule to fire at all.
		"Q6 function-like body pasting a call-site fragment": {
			operand: "J(a,b)=a##b", body: "J(a,b) a##b",
			use:  "J(__as,m__)(\".incbin \\\"payload.bin\\\"\");",
			want: CodeHeaderInputUndeclared,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Run("bound by the build setting", func(t *testing.T) {
				fixture := newFixture(t)
				fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\n" + expect.use + "\nint value(void) { return 1; }\n"
				fixture.target("CLib").Settings = []swiftpmsource.BuildSetting{swiftpmSetting(`{"kind":{"define":{"_0":"` + expect.operand + `"}},"tool":"c"}`)}
				result, err := fixture.close()
				requireCode(t, err, expect.want)
				if result != nil {
					t.Fatalf("rejected closure published a result: %#v", result)
				}
			})
			t.Run("control the same body bound by a source define", func(t *testing.T) {
				fixture := newFixture(t)
				fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\n#define " + expect.body + "\n" + expect.use + "\nint value(void) { return 1; }\n"
				result, err := fixture.close()
				requireCode(t, err, expect.want)
				if result != nil {
					t.Fatalf("rejected closure published a result: %#v", result)
				}
			})
		})
	}
	// A body that resolves to a module import belongs to no admitted source, so
	// it can neither be attributed nor confined and is refused rather than
	// silently attached to the target's include closure.
	t.Run("a define body that performs a module import rejects", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.target("CLib").Settings = []swiftpmsource.BuildSetting{swiftpmSetting(`{"kind":{"define":{"_0":"IMP=_Pragma(\"clang module import SecretKit\")"}},"tool":"c"}`)}
		result, err := fixture.close()
		requireCode(t, err, CodeHeaderInputUndeclared)
		if result != nil {
			t.Fatalf("rejected closure published a result: %#v", result)
		}
	})
	// The operand grammar is closed the same way: a name and body this stage
	// cannot separate is a define whose effect it cannot bound.
	t.Run("an operand this stage cannot separate rejects", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.target("CLib").Settings = []swiftpmsource.BuildSetting{swiftpmSetting(`{"kind":{"define":{"_0":"A B=__asm__"}},"tool":"c"}`)}
		result, err := fixture.close()
		requireCode(t, err, CodeUnsafeSettingForbidden)
		if result != nil {
			t.Fatalf("rejected closure published a result: %#v", result)
		}
	})
	// A pruned setting is still never analysed: the condition axis stays ahead
	// of the body axis, exactly as it is ahead of the name axis.
	t.Run("a pruned define setting body binds nothing", func(t *testing.T) {
		fixture := newFixture(t)
		setting := swiftpmSetting(`{"kind":{"define":{"_0":"A=__asm__"}},"tool":"c"}`)
		setting.Condition = swiftpmCondition("platform=linux")
		fixture.target("CLib").Settings = []swiftpmsource.BuildSetting{setting}
		fixture.mustClose()
	})
	// The positive finding Q must not cost: ordinary object-like and
	// function-like defines with inert bodies still admit a normal target with
	// its references intact.
	t.Run("benign define bodies still admit a normal target", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.target("CLib").Settings = []swiftpmsource.BuildSetting{
			swiftpmSetting(`{"kind":{"define":{"_0":"FEATURE=1"}},"tool":"c"}`),
			swiftpmSetting(`{"kind":{"define":{"_0":"MAX=256"}},"tool":"c"}`),
			swiftpmSetting(`{"kind":{"define":{"_0":"BANNER=\"portable mode\""}},"tool":"c"}`),
			swiftpmSetting(`{"kind":{"define":{"_0":"SQUARE(x)=((x) * (x))"}},"tool":"c"}`),
			swiftpmSetting(`{"kind":{"define":{"_0":"NDEBUG"}},"tool":"c"}`),
		}
		result := fixture.mustClose()
		spellings := []string{}
		for _, reference := range mustTarget(t, result, "root:CLib").Includes {
			spellings = append(spellings, reference.Spelling)
		}
		if len(spellings) != 2 || spellings[0] != "CLib.h" || spellings[1] != "stdio.h" {
			t.Fatalf("scanned references = %v", spellings)
		}
	})
}
