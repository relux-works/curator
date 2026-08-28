package swiftpminterop

import (
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/swiftpmsource"
)

// The module-map grammar must resolve every construct SwiftPM and Clang can
// emit, because an unparsed clause is an unproved input.
func TestModuleMapGrammarResolvesEveryDeclaredReference(t *testing.T) {
	payload := `// leading comment
/* block
   comment */
framework module Root [system] [extern_c] {
    umbrella header "Root.h"
    private textual header "Private.h"
    exclude header "Skip.h"
    header "Extra.h" { size 100 mtime 200 }
    link "root"
    link framework "Foundation"
    requires objc, !cplusplus
    config_macros [exhaustive] ROOT_DEBUG, ROOT_TRACE
    export *
    export Other
    use Helper
    explicit module Nested {
        header "Nested.h"
    }
}
extern module Detached "Detached/module.modulemap"
`
	parsed, err := ParseModuleMap("module.modulemap", []byte(payload))
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"Detached", "Root", "Root.Nested"}; !reflect.DeepEqual(parsed.Modules, want) {
		t.Fatalf("modules = %v, want %v", parsed.Modules, want)
	}
	if want := []string{"Root"}; !reflect.DeepEqual(parsed.Frameworks, want) {
		t.Fatalf("frameworks = %v, want %v", parsed.Frameworks, want)
	}
	if want := []string{"!cplusplus", "objc"}; !reflect.DeepEqual(parsed.Requires, want) {
		t.Fatalf("requires = %v, want %v", parsed.Requires, want)
	}
	kinds := map[ReferenceKind][]string{}
	for _, reference := range parsed.References {
		kinds[reference.Kind] = append(kinds[reference.Kind], reference.Path)
	}
	if want := []string{"Root.h"}; !reflect.DeepEqual(kinds[ReferenceUmbrellaHeader], want) {
		t.Fatalf("umbrella headers = %v", kinds[ReferenceUmbrellaHeader])
	}
	if want := []string{"Private.h", "Skip.h", "Extra.h", "Nested.h"}; !reflect.DeepEqual(kinds[ReferenceHeader], want) {
		t.Fatalf("headers = %v", kinds[ReferenceHeader])
	}
	if want := []string{"Detached/module.modulemap"}; !reflect.DeepEqual(kinds[ReferenceExternModule], want) {
		t.Fatalf("extern modules = %v", kinds[ReferenceExternModule])
	}
	if len(parsed.Links) != 2 || parsed.Links[0].Name != "root" || parsed.Links[0].Framework || !parsed.Links[1].Framework {
		t.Fatalf("links = %#v", parsed.Links)
	}
}

// An umbrella directory declaration is distinct from an umbrella header.
func TestModuleMapUmbrellaDirectory(t *testing.T) {
	parsed, err := ParseModuleMap("module.modulemap", []byte("module Root {\n    umbrella \".\"\n    export *\n}\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed.References) != 1 || parsed.References[0].Kind != ReferenceUmbrellaDirectory || parsed.References[0].Path != "." {
		t.Fatalf("references = %#v", parsed.References)
	}
}

// Additional malformed shapes must reject rather than parse partially.
func TestModuleMapRejectsAdditionalMalformedShapes(t *testing.T) {
	for name, payload := range map[string]string{
		"missing module name":     "module {\n}\n",
		"missing extern path":     "extern module Other\n",
		"unterminated attributes": "module Root [system {\n}\n",
		"literal attribute":       "module Root [\"system\"] {\n}\n",
		"string at top level":     "\"Root\"\n",
		"unsupported character":   "module Root {\n    header @\n}\n",
		"missing header path":     "module Root {\n    header\n}\n",
		"missing umbrella path":   "module Root {\n    umbrella header\n}\n",
		"missing link name":       "module Root {\n    link\n}\n",
		"dangling requires":       "module Root {\n    requires\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseModuleMap("module.modulemap", []byte(payload)); ErrorCode(err) != CodeModuleMapEscape {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

// The generated module map reproduces SwiftPM's documented selection rules.
func TestGenerateModuleMapSelectionRules(t *testing.T) {
	headers := []HeaderFile{{Relative: "include/CLib.h"}, {Relative: "include/nested/Other.h"}}
	generated, err := GenerateModuleMap("CLib", "include", headers)
	if err != nil || generated != "module CLib {\n    umbrella header \"CLib.h\"\n    export *\n}\n" {
		t.Fatalf("umbrella-header form = %q, %v", generated, err)
	}
	single, err := GenerateModuleMap("CLib", "include", []HeaderFile{{Relative: "include/Only.h"}})
	if err != nil || single != "module CLib {\n    umbrella header \"Only.h\"\n    export *\n}\n" {
		t.Fatalf("single-header form = %q, %v", single, err)
	}
	directory, err := GenerateModuleMap("CLib", "include", []HeaderFile{{Relative: "include/A.h"}, {Relative: "include/B.h"}})
	if err != nil || directory != "module CLib {\n    umbrella \".\"\n    export *\n}\n" {
		t.Fatalf("umbrella-directory form = %q, %v", directory, err)
	}
	if _, err = GenerateModuleMap("CLib", "include", nil); ErrorCode(err) != CodeGraphIncomplete {
		t.Fatalf("empty public header set = %v", err)
	}
	if got := c99Identifier("my-lib+2"); got != "my_lib_2" {
		t.Fatalf("c99 identifier = %q", got)
	}
}

// Source classification is driven by the exact declared extension set.
func TestSourceClassificationCoversTheClosedLanguageSet(t *testing.T) {
	for relative, want := range map[string]Language{
		"a.swift": LanguageSwift, "a.c": LanguageC, "a.cpp": LanguageCXX, "a.cc": LanguageCXX,
		"a.cxx": LanguageCXX, "a.m": LanguageObjC, "a.mm": LanguageObjCXX,
	} {
		if got, ok := sourceLanguage(relative); !ok || got != want {
			t.Fatalf("sourceLanguage(%q) = %q, %v", relative, got, ok)
		}
	}
	for _, relative := range []string{"a.h", "a.hpp", "module.modulemap", "README.md"} {
		if _, ok := sourceLanguage(relative); ok {
			t.Fatalf("%q was classified as a compiler source", relative)
		}
	}
	if _, _, err := classifyTarget("root", "Empty", swiftpmsource.Target{Name: "Empty", Type: "regular", Sources: []string{"README.md"}}); ErrorCode(err) != CodeGraphIncomplete {
		t.Fatalf("unclassifiable target = %v", err)
	}
}

// A Clang module import resolves against the admitted module maps of the
// declared target edges or against a selected external module.
func TestClangModuleImportsResolveAgainstDeclaredModules(t *testing.T) {
	fixture := newFixture(t)
	fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\n@import Foundation;\n@import CLib;\nint value(void) { return 1; }\n"
	result := fixture.mustClose()
	imports := 0
	for _, reference := range mustTarget(t, result, "root:CLib").Includes {
		if reference.ModuleImport {
			imports++
		}
	}
	if imports != 2 {
		t.Fatalf("recorded %d Clang module imports", imports)
	}
}

// A conditional product dependency is captured once with its condition and the
// active projection is the authority that reports it selected.
func TestConditionalProductDependencyIsProjectedExactlyOnce(t *testing.T) {
	fixture := newFixture(t)
	fixture.manifest.Products = append(fixture.manifest.Products, swiftpmsource.Product{Name: "CLibProd", Type: "library", Targets: []string{"CLib"}})
	fixture.target("App").Dependencies = []swiftpmsource.TargetDependency{{Product: "CLibProd", Condition: swiftpmCondition("platform=macos")}}
	result := fixture.mustClose()
	if len(result.Boundaries) != 1 || result.Boundaries[0].Provider != "root:CLib" {
		t.Fatalf("boundaries = %#v", result.Boundaries)
	}
	boundary := result.Boundaries[0]
	if boundary.Condition == nil || boundary.Condition.Expression != "platform=macos" {
		t.Fatalf("boundary condition = %#v", boundary.Condition)
	}
	if !boundary.Selected {
		t.Fatal("the selected destination did not select the conditional boundary")
	}
	if state := activationForID(t, result, boundary.NodeID); state != closuregraph.ActivationSelected {
		t.Fatalf("boundary activation = %q", state)
	}
	if state := activationForKey(t, result, "swiftpm.interop.compile.root.CLib"); state != closuregraph.ActivationSelected {
		t.Fatalf("provider compile activation = %q", state)
	}
	consumes := edgeActivation(t, result, "swiftpm.interop.consumes.root:CLib->root:App")
	if !consumes.Evaluation || consumes.State != closuregraph.ActivationSelected {
		t.Fatalf("conditional consumer edge activation = %#v", consumes)
	}
}

// A conditional dependency the destination does not select is still captured
// in full: the pruned verdict is recorded by the active projection, never by
// omitting the declaration from the selection-neutral capture.
func TestPrunedConditionalDependencyIsCapturedAndProjectedPruned(t *testing.T) {
	fixture := newFixture(t)
	fixture.target("App").Dependencies = []swiftpmsource.TargetDependency{{Name: "CLib", Condition: swiftpmCondition("platform=windows")}}
	result := fixture.mustClose()
	if len(result.Boundaries) != 1 || result.Boundaries[0].Provider != "root:CLib" {
		t.Fatalf("pruned dependency dropped its capture declaration: %#v", result.Boundaries)
	}
	boundary := result.Boundaries[0]
	if boundary.Selected {
		t.Fatal("a pruned conditional dependency was reported as selected")
	}
	if state := activationForID(t, result, boundary.NodeID); state != closuregraph.ActivationPruned {
		t.Fatalf("boundary activation = %q", state)
	}
	provider := mustTarget(t, result, "root:CLib")
	if provider.Selected {
		t.Fatal("a pruned C-family target was reported as selected")
	}
	for _, key := range []string{"swiftpm.interop.compile.root.CLib", "swiftpm.interop.sources.root.CLib", "swiftpm.interop.headers.root.CLib", "swiftpm.interop.object.root.CLib.0000"} {
		if state := activationForKey(t, result, key); state != closuregraph.ActivationPruned {
			t.Fatalf("%s activation = %q", key, state)
		}
	}
	for _, edge := range result.Records.BindingEdges {
		if strings.Contains(edge.EdgeKey, "root.CLib") || strings.Contains(edge.EdgeKey, "root:CLib->") {
			t.Fatalf("pruned declaration gained an exact binding edge: %s", edge.EdgeKey)
		}
	}
}

// The assembly template grammar is closed: only a run of plain adjacent string
// literals is readable, and its escape decoding has to match the compiler's,
// because the assembler sees the decoded bytes.
// Translation phase 2 is a small closed set and it has to be exact: a
// reject-by-default keyword match is only closed if a splice cannot
// reconstitute the keyword past the scanner. The pinned Apple Clang joins a
// backslash to the following newline across any run of horizontal white space,
// in all six separator spellings and all twelve `-std`/language modes tried, so
// the splice is unconditional here. A backslash that is not part of a splice is
// preserved byte for byte, because it is content — and, after C escape
// decoding, the assembler's own substitution marker.
func TestTranslationPhaseTwoSplicing(t *testing.T) {
	for name, want := range map[string]struct{ payload, spliced string }{
		"a bare splice":             {"a\\\nb", "ab"},
		"a space splice":            {"a\\ \nb", "ab"},
		"a tab splice":              {"a\\\t\nb", "ab"},
		"a vertical tab splice":     {"a\\\v\nb", "ab"},
		"a form feed splice":        {"a\\\f\nb", "ab"},
		"a mixed run splice":        {"a\\ \t\v\f \nb", "ab"},
		"a CRLF splice":             {"a\\ \r\nb", "ab"},
		"a CR splice":               {"a\\ \rb", "ab"},
		"consecutive splices":       {"a\\\n\\ \n\\\t\nb", "ab"},
		"a keyword split":           {"__as\\ \nm__(\"x\")", "__asm__(\"x\")"},
		"a backslash before text":   {"a\\b", "a\\b"},
		"a backslash before space":  {"a\\ b", "a\\ b"},
		"a trailing backslash":      {"a\\", "a\\"},
		"a trailing spaced slash":   {"a\\ ", "a\\ "},
		"an escaped backslash pair": {"a\\\\\nb", "a\\b"},
		"line endings only":         {"a\r\nb\rc\n", "a\nb\nc\n"},
		"no backslash at all":       {"#include \"a.h\"\n", "#include \"a.h\"\n"},
	} {
		t.Run(name, func(t *testing.T) {
			if got := spliceTranslationLines(want.payload); got != want.spliced {
				t.Fatalf("splice(%q) = %q; want %q", want.payload, got, want.spliced)
			}
		})
	}
}

// The pragma allowlist is the closed channel surface, so it is pinned directly
// as well as through the target-level vectors: a spelling that names nothing is
// content, `clang module import NAME` is a resolved module reference, and every
// other body — including the three spellings earlier rounds proved to be real
// channels and any spelling a later toolchain adds — rejects.
func TestPragmaAllowlistIsClosed(t *testing.T) {
	for name, body := range map[string]string{
		"an empty body":            "",
		"an include guard":         "once",
		"a layout pragma":          "pack(push, 1)",
		"a macro table pragma":     "push_macro(\"X\")",
		"a presentation pragma":    "mark - section",
		"a standard pragma":        "STDC FENV_ACCESS ON",
		"a clang diagnostic":       "clang diagnostic ignored \"-Wall\"",
		"a clang nullability span": "clang assume_nonnull begin",
		"a GCC visibility pragma":  "GCC visibility push(default)",
	} {
		t.Run("content: "+name, func(t *testing.T) {
			if verdict := classifyPragmaBody(body); verdict.rejected || verdict.module != "" {
				t.Fatalf("pragma %q verdict = %#v", body, verdict)
			}
		})
	}
	for name, body := range map[string]string{
		"a library comment":      "comment(lib, \"SecretProbeLib\")",
		"an include alias":       "include_alias(\"a.h\", \"b.h\")",
		"a GCC dependency":       "GCC dependency \"/etc/passwd\"",
		"an unknown vendor":      "acme read_from(\"/etc/passwd\")",
		"an unknown clang body":  "clang __debug module_map",
		"an unknown GCC body":    "GCC unknown_extension",
		"a module build":         "clang module build SecretKit",
		"a module endbuild":      "clang module endbuild",
		"a module load":          "clang module load \"secret.pcm\"",
		"a module begin":         "clang module begin SecretKit",
		"a nameless module":      "clang module import",
		"a trailing module name": "clang module import SecretKit extra",
	} {
		t.Run("rejected: "+name, func(t *testing.T) {
			if verdict := classifyPragmaBody(body); !verdict.rejected {
				t.Fatalf("pragma %q verdict = %#v", body, verdict)
			}
		})
	}
	for name, want := range map[string]string{
		"a plain module import":  "SecretKit",
		"a dotted module import": "Secret.Kit",
	} {
		t.Run("module: "+name, func(t *testing.T) {
			if verdict := classifyPragmaBody("clang module import " + want); verdict.rejected || verdict.module != want {
				t.Fatalf("pragma import %q verdict = %#v", want, verdict)
			}
		})
	}
}
