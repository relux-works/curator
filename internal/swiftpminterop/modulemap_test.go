package swiftpminterop

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/swiftpmsource"
)

// H01: the conventional umbrella-header layout reproduces SwiftPM's automatic
// module map and hashes every admitted public header.
func TestH01GeneratedUmbrellaHeaderModuleMapIsReproduced(t *testing.T) {
	fixture := newFixture(t)
	result := fixture.mustClose()
	evidence := mustTarget(t, result, "root:CLib").ModuleMap
	if !evidence.Generated || evidence.GrammarID != GeneratedModuleMapGrammarID {
		t.Fatalf("module map evidence = %#v", evidence)
	}
	if len(evidence.Parsed.References) != 1 || evidence.Parsed.References[0].Kind != ReferenceUmbrellaHeader || evidence.Parsed.References[0].Path != "CLib.h" {
		t.Fatalf("generated references = %#v", evidence.Parsed.References)
	}
	if len(evidence.Parsed.Modules) != 1 || evidence.Parsed.Modules[0] != "CLib" {
		t.Fatalf("generated modules = %v", evidence.Parsed.Modules)
	}
	if len(evidence.ResolvedRefs) != 1 || evidence.ResolvedRefs[0].Class != ResolvedAdmitted {
		t.Fatalf("resolved references = %#v", evidence.ResolvedRefs)
	}
}

// H01: several directly contained headers fall back to SwiftPM's umbrella
// directory form rather than guessing an umbrella header.
func TestH01GeneratedUmbrellaDirectoryModuleMap(t *testing.T) {
	fixture := newFixture(t)
	fixture.addFiles(map[string]string{
		"Sources/CLib/include/extra.h": "int extra(void);\n",
		"Sources/CLib/include/more.h":  "int more(void);\n",
	})
	delete(fixture.files, "Sources/CLib/include/CLib.h")
	fixture.files["Sources/CLib/lib.c"] = "#include \"extra.h\"\nint value(void) { return 1; }\n"
	result := fixture.mustClose()
	evidence := mustTarget(t, result, "root:CLib").ModuleMap
	if len(evidence.Parsed.References) != 1 || evidence.Parsed.References[0].Kind != ReferenceUmbrellaDirectory {
		t.Fatalf("generated references = %#v", evidence.Parsed.References)
	}
	if len(evidence.PublicHeaderFiles) != 2 {
		t.Fatalf("public header inventory = %#v", evidence.PublicHeaderFiles)
	}
}

// H02: a custom module map wholly inside the package is admitted with its
// exact parsed reference manifest.
func TestH02CustomContainedModuleMapIsAdmitted(t *testing.T) {
	fixture := newFixture(t)
	fixture.addFiles(map[string]string{
		"Sources/CLib/include/module.modulemap": "// custom\nmodule CLib {\n    header \"CLib.h\"\n    export *\n}\n",
	})
	result := fixture.mustClose()
	evidence := mustTarget(t, result, "root:CLib").ModuleMap
	if evidence.Generated || evidence.Relative != "Sources/CLib/include/module.modulemap" || evidence.GrammarID != ModuleMapGrammarID {
		t.Fatalf("module map evidence = %#v", evidence)
	}
	if len(evidence.ResolvedRefs) != 1 || evidence.ResolvedRefs[0].Relative != "Sources/CLib/include/CLib.h" {
		t.Fatalf("resolved references = %#v", evidence.ResolvedRefs)
	}
}

// H03: SwiftPM itself builds a module map that reaches an absolute header
// outside the package. Curator rejects it before any compiler runs.
func TestH03AbsoluteModuleMapHeaderIsRejected(t *testing.T) {
	fixture := newFixture(t)
	fixture.addFiles(map[string]string{
		"Sources/CLib/include/module.modulemap": "module CLib {\n    header \"/etc/hosts\"\n    export *\n}\n",
	})
	_, err := fixture.close()
	requireCode(t, err, CodeModuleMapEscape)
}

// H03: a relative module-map reference that escapes the admitted package is
// the same rejection.
func TestH03EscapingModuleMapHeaderIsRejected(t *testing.T) {
	fixture := newFixture(t)
	fixture.addFiles(map[string]string{
		"Sources/CLib/include/module.modulemap": "module CLib {\n    header \"../../../../outside.h\"\n    export *\n}\n",
	})
	_, err := fixture.close()
	requireCode(t, err, CodeModuleMapEscape)
}

// An extern module reference outside the admitted package is rejected even
// when the named file happens to exist in a selected external root.
func TestExternModuleOutsidePackageIsRejected(t *testing.T) {
	fixture := newFixture(t)
	fixture.addFiles(map[string]string{
		"Sources/CLib/include/module.modulemap": "module CLib {\n    umbrella header \"CLib.h\"\n    export *\n}\nextern module Other \"../../../sdk/usr/include/other.modulemap\"\n",
	})
	_, err := fixture.close()
	requireCode(t, err, CodeModuleMapEscape)
}

// A module map this grammar cannot resolve exactly is a rejection, never a
// silent partial parse.
func TestMalformedModuleMapIsRejected(t *testing.T) {
	for name, payload := range map[string]string{
		"unterminated body":    "module CLib {\n    header \"CLib.h\"\n",
		"unterminated string":  "module CLib {\n    header \"CLib.h\n}\n",
		"unknown member":       "module CLib {\n    frobnicate \"CLib.h\"\n}\n",
		"unknown top level":    "import CLib\n",
		"unterminated comment": "/* module CLib {\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseModuleMap("module.modulemap", []byte(payload)); ErrorCode(err) != CodeModuleMapEscape {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

// H04: a source that includes a header outside the admitted closure is
// rejected with the undeclared-header diagnostic.
func TestH04UndeclaredHeaderIncludeIsRejected(t *testing.T) {
	t.Run("absolute", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"/etc/hosts\"\nint value(void) { return 1; }\n"
		_, err := fixture.close()
		requireCode(t, err, CodeHeaderInputUndeclared)
	})
	t.Run("parent escape", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"../../../../outside.h\"\nint value(void) { return 1; }\n"
		_, err := fixture.close()
		requireCode(t, err, CodeHeaderInputUndeclared)
	})
	t.Run("angled host header", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include <sneaky/secret.h>\nint value(void) { return 1; }\n"
		_, err := fixture.close()
		requireCode(t, err, CodeHeaderInputUndeclared)
	})
	t.Run("undeclared clang module", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "@import SecretKit;\nint value(void) { return 1; }\n"
		_, err := fixture.close()
		requireCode(t, err, CodeHeaderInputUndeclared)
	})
}

// H05: a header reachable only through an ambient include path or an unsafe
// flag never becomes admitted. The environment is not consulted at all, and
// the unsafe flag itself is rejected.
func TestH05AmbientIncludePathAndUnsafeFlagAreNotAdmitted(t *testing.T) {
	fixture := newFixture(t)
	fixture.writeExternal(map[string]string{"ambient.h": "int ambient(void);\n"}, filepath.Join(fixture.base, "ambient"))
	t.Setenv("CPATH", filepath.Join(fixture.base, "ambient"))
	t.Setenv("C_INCLUDE_PATH", filepath.Join(fixture.base, "ambient"))
	fixture.files["Sources/CLib/lib.c"] = "#include <ambient.h>\nint value(void) { return 1; }\n"
	_, err := fixture.close()
	requireCode(t, err, CodeHeaderInputUndeclared)

	unsafe := newFixture(t)
	unsafe.target("CLib").Settings = []swiftpmsource.BuildSetting{{Kind: "unsafeFlags", Value: "-I" + filepath.Join(unsafe.base, "ambient"), Unsafe: true}}
	_, unsafeErr := unsafe.close()
	requireCode(t, unsafeErr, CodeUnsafeSettingForbidden)
}

// H06: a system-library target inside a Curator-selected external root is
// admitted as an explicit external-toolchain binding.
func TestH06SystemLibraryInsideSelectedRootIsAdmitted(t *testing.T) {
	fixture := newSystemLibraryFixture(t, filepath.Join(newFixtureBase(t), "unused"))
	fixture.materializeHook = func() {
		fixture.interop.SystemLibraries = []SystemLibrary{{
			Package: "root", Target: "CFoo", ModuleMapPath: filepath.Join(fixture.systemRoot, "cfoo", "module.modulemap"),
			Component: systemComponent(filepath.Join(fixture.systemRoot, "cfoo")),
		}}
	}
	result := fixture.mustClose()
	system := mustTarget(t, result, "root:CFoo")
	if system.Kind != KindSystem || system.ModuleMap == nil || len(system.ModuleMap.ResolvedRefs) != 1 {
		t.Fatalf("system target evidence = %#v", system)
	}
	if system.ModuleMap.ResolvedRefs[0].Class != ResolvedBinding {
		t.Fatalf("system header did not resolve to a selected binding node: %#v", system.ModuleMap.ResolvedRefs[0])
	}
	if len(result.Boundaries) != 1 || result.Boundaries[0].Provider != "root:CFoo" {
		t.Fatalf("boundaries = %#v", result.Boundaries)
	}
	found := false
	for _, node := range result.Records.BindingNodes {
		if node.LogicalKey == "swiftpm.interop.component.system-cfoo" {
			found = true
		}
	}
	if !found {
		t.Fatal("selected system component is absent from the exact selection binding")
	}
}

// H07: the same declaration resolving to an arbitrary host path is untrusted.
// A provider hint is never provenance.
func TestH07SystemLibraryOutsideSelectedRootIsUntrusted(t *testing.T) {
	t.Run("host path", func(t *testing.T) {
		fixture := newSystemLibraryFixture(t, "")
		fixture.materializeHook = func() {
			host := filepath.Join(fixture.base, "homebrew", "opt", "cfoo")
			fixture.writeExternal(map[string]string{"module.modulemap": "module CFoo {\n    header \"cfoo.h\"\n}\n", "cfoo.h": "int cfoo(void);\n"}, host)
			fixture.interop.SystemLibraries = []SystemLibrary{{
				Package: "root", Target: "CFoo", ModuleMapPath: filepath.Join(host, "module.modulemap"),
				Component: systemComponent(filepath.Join(fixture.systemRoot, "cfoo")),
			}}
		}
		_, err := fixture.close()
		requireCode(t, err, CodeToolchainUntrusted)
	})
	t.Run("no binding at all", func(t *testing.T) {
		fixture := newSystemLibraryFixture(t, "")
		_, err := fixture.close()
		requireCode(t, err, CodeToolchainUntrusted)
	})
}

// H08: a module map that links a library or framework no selected component
// declares is untrusted.
func TestH08UndeclaredLinkedLibraryIsUntrusted(t *testing.T) {
	fixture := newFixture(t)
	fixture.addFiles(map[string]string{
		"Sources/CLib/include/module.modulemap": "module CLib {\n    umbrella header \"CLib.h\"\n    link \"unvetted\"\n    export *\n}\n",
	})
	_, err := fixture.close()
	requireCode(t, err, CodeToolchainUntrusted)

	framework := newFixture(t)
	framework.addFiles(map[string]string{
		"Sources/CLib/include/module.modulemap": "module CLib {\n    umbrella header \"CLib.h\"\n    link framework \"Unvetted\"\n    export *\n}\n",
	})
	_, frameworkErr := framework.close()
	requireCode(t, frameworkErr, CodeToolchainUntrusted)
}

// A module map that links a library the selected SDK declares is admitted.
func TestDeclaredLinkedLibraryIsAdmitted(t *testing.T) {
	fixture := newFixture(t)
	fixture.addFiles(map[string]string{
		"Sources/CLib/include/module.modulemap": "module CLib {\n    umbrella header \"CLib.h\"\n    link \"c\"\n    link framework \"Foundation\"\n    export *\n}\n",
	})
	result := fixture.mustClose()
	links := mustTarget(t, result, "root:CLib").ModuleMap.Parsed.Links
	if len(links) != 2 || links[0].Name != "c" || !links[1].Framework {
		t.Fatalf("links = %#v", links)
	}
}

func newFixtureBase(t *testing.T) string { return t.TempDir() }

func systemComponent(root string) ExternalComponent {
	return ExternalComponent{
		Role: "system-cfoo", ExecutableRelativePath: "bin/pkg-config", PlatformABI: "darwin-arm64",
		PolicySelector: "curator-selected-system-v1", VersionOutput: "cfoo 1.2.3", Fingerprint: id('9'),
		Roots: []string{root}, Libraries: []string{"cfoo"}, Modules: []string{"CFoo"},
	}
}

func newSystemLibraryFixture(t *testing.T, _ string) *fixture {
	t.Helper()
	fixture := newFixture(t)
	delete(fixture.files, "Sources/CLib/lib.c")
	delete(fixture.files, "Sources/CLib/include/CLib.h")
	fixture.files["Sources/App/main.swift"] = "import CFoo\nprint(1)\n"
	fixture.manifest.Targets = []swiftpmsource.Target{
		{Name: "App", Type: "executable", Path: "Sources/App", Sources: []string{"Sources/App/main.swift"}, Dependencies: []swiftpmsource.TargetDependency{{Name: "CFoo"}}},
		{Name: "CFoo", Type: "system", Path: "Sources/CFoo"},
	}
	fixture.writeExternal(map[string]string{
		"cfoo/module.modulemap": "module CFoo [system] {\n    header \"cfoo.h\"\n    link \"cfoo\"\n    export *\n}\n",
		"cfoo/cfoo.h":           "int cfoo(void);\n",
	}, fixture.systemRoot)
	return fixture
}

var _ = strings.TrimSpace

// H09: an inclusion directive whose operand this grammar cannot resolve to an
// exact literal is a rejection, never a silently dropped reference. The
// macro-computed form is the concrete escape: `#define CURATOR_SECRET
// </etc/passwd>` followed by `#include CURATOR_SECRET` names a host header that
// no literal scanner can see, and portable mode has no compiler read set to
// catch it afterwards.
func TestH09ComputedIncludeDirectiveIsRejected(t *testing.T) {
	t.Run("macro operand", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\n#define CURATOR_SECRET </etc/passwd>\n#include CURATOR_SECRET\nint value(void) { return 1; }\n"
		_, err := fixture.close()
		requireCode(t, err, CodeHeaderInputUndeclared)
	})
	t.Run("macro operand in a public header", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/include/CLib.h"] = "#ifndef CLIB_H\n#define CLIB_H\n#define CURATOR_SECRET <secret.h>\n#include_next CURATOR_SECRET\nint value(void);\n#endif\n"
		_, err := fixture.close()
		requireCode(t, err, CodeHeaderInputUndeclared)
	})
	t.Run("unterminated literal", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include <stdio.h\nint value(void) { return 1; }\n"
		_, err := fixture.close()
		requireCode(t, err, CodeHeaderInputUndeclared)
	})
	t.Run("empty spelling", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include <>\nint value(void) { return 1; }\n"
		_, err := fixture.close()
		requireCode(t, err, CodeHeaderInputUndeclared)
	})
	t.Run("trailing token after the literal", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\" EXTRA\nint value(void) { return 1; }\n"
		_, err := fixture.close()
		requireCode(t, err, CodeHeaderInputUndeclared)
	})
	t.Run("trailing comment is still exact", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\" // the public interface\n#include <stdio.h> /* SDK */\nint value(void) { return 1; }\n"
		result := fixture.mustClose()
		spellings := []string{}
		for _, reference := range mustTarget(t, result, "root:CLib").Includes {
			spellings = append(spellings, reference.Spelling)
		}
		if strings.Join(spellings, ",") != "CLib.h,stdio.h" {
			t.Fatalf("scanned include spellings = %v", spellings)
		}
	})
}

// H10: a declared publicHeadersPath is honored exactly. The reviewer's probe is
// the regression: with the declaration dropped, the assumed `include` directory
// gets a generated module map while the real escaping map under the declared
// directory is never parsed, defeating H03 entirely.
func TestH10DeclaredPublicHeadersPathIsHonored(t *testing.T) {
	t.Run("custom directory is inventoried and parsed", func(t *testing.T) {
		fixture := newFixture(t)
		delete(fixture.files, "Sources/CLib/include/CLib.h")
		fixture.addFiles(map[string]string{
			"Sources/CLib/pub/CLib.h":           "#ifndef CLIB_H\n#define CLIB_H\nint value(void);\n#endif\n",
			"Sources/CLib/pub/module.modulemap": "module CLib {\n    umbrella header \"CLib.h\"\n    export *\n}\n",
		})
		fixture.target("CLib").PublicHeadersPath = "pub"
		result := fixture.mustClose()
		clang := mustTarget(t, result, "root:CLib")
		if clang.PublicHeaderRoot != "Sources/CLib/pub" {
			t.Fatalf("public header root = %q", clang.PublicHeaderRoot)
		}
		if clang.ModuleMap == nil || clang.ModuleMap.Generated || clang.ModuleMap.Relative != "Sources/CLib/pub/module.modulemap" {
			t.Fatalf("module map evidence = %#v", clang.ModuleMap)
		}
	})
	t.Run("escaping map under the declared directory is rejected", func(t *testing.T) {
		fixture := newFixture(t)
		delete(fixture.files, "Sources/CLib/include/CLib.h")
		fixture.addFiles(map[string]string{
			"Sources/CLib/pub/CLib.h":           "#ifndef CLIB_H\n#define CLIB_H\nint value(void);\n#endif\n",
			"Sources/CLib/pub/module.modulemap": "module CLib {\n    header \"/etc/passwd\"\n    export *\n}\n",
		})
		fixture.target("CLib").PublicHeadersPath = "pub"
		_, err := fixture.close()
		requireCode(t, err, CodeModuleMapEscape)
	})
}

// H11: a public-header layout this profile cannot represent exactly fails
// closed instead of silently falling back to SwiftPM's default directory.
func TestH11NonRepresentablePublicHeaderLayoutFailsClosed(t *testing.T) {
	for name, declared := range map[string]string{
		"absolute":        "/etc",
		"parent escape":   "../../elsewhere",
		"escaping suffix": "include/../../../elsewhere",
		"windows drive":   `C:\headers`,
	} {
		t.Run(name, func(t *testing.T) {
			fixture := newFixture(t)
			fixture.target("CLib").PublicHeadersPath = declared
			_, err := fixture.close()
			requireCode(t, err, CodeTargetPlatformUnsupported)
		})
	}
	t.Run("module map outside the resolved public-header root", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.addFiles(map[string]string{
			"Sources/CLib/module.modulemap":         "module CLib {\n    header \"/etc/passwd\"\n    export *\n}\n",
			"Sources/CLib/include/module.modulemap": "module CLib {\n    umbrella header \"CLib.h\"\n    export *\n}\n",
		})
		_, err := fixture.close()
		requireCode(t, err, CodeModuleMapEscape)
	})
	t.Run("module map nested below the public-header root", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.addFiles(map[string]string{
			"Sources/CLib/include/nested/module.modulemap": "module Nested {\n    header \"/etc/passwd\"\n}\n",
		})
		_, err := fixture.close()
		requireCode(t, err, CodeModuleMapEscape)
	})
}

// H12: the include scan must cover the transitive closure of every admitted
// reference it resolves. The conventional private-header layout SwiftPM permits
// — a header inside the target tree but outside the public-header root — was
// admitted as a reference and then never opened, so every directive it declared
// was invisible. In portable mode this declared closure is the entire header
// proof, so an unopened admitted reference is an admission hole.
func TestH12TransitiveIncludeClosureIsScanned(t *testing.T) {
	t.Run("private header escape is reached", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\n#include \"private.h\"\nint value(void) { return 1; }\n"
		fixture.files["Sources/CLib/private.h"] = "#include </etc/passwd>\n"
		_, err := fixture.close()
		requireCode(t, err, CodeHeaderInputUndeclared)
	})
	t.Run("escape three levels deep is reached", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"detail/a.h\"\nint value(void) { return 1; }\n"
		fixture.addFiles(map[string]string{
			"Sources/CLib/detail/a.h": "#include \"b.h\"\n",
			"Sources/CLib/detail/b.h": "#include \"../../../../outside.h\"\n",
		})
		_, err := fixture.close()
		requireCode(t, err, CodeHeaderInputUndeclared)
	})
	t.Run("public header of a dependency is reached", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/include/CLib.h"] = "#ifndef CLIB_H\n#define CLIB_H\n#include \"detail.h\"\nint value(void);\n#endif\n"
		fixture.files["Sources/CLib/include/detail.h"] = "@import SecretKit;\n"
		_, err := fixture.close()
		requireCode(t, err, CodeHeaderInputUndeclared)
	})
	t.Run("admitted reference that cannot be opened fails closed", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\n#include \"detail\"\nint value(void) { return 1; }\n"
		fixture.files["Sources/CLib/detail/keep.h"] = "int keep(void);\n"
		_, err := fixture.close()
		requireCode(t, err, CodeSourceInventoryDrift)
	})
	t.Run("module-map header outside the public root is scanned", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.addFiles(map[string]string{
			"Sources/CLib/include/module.modulemap": "module CLib {\n    header \"CLib.h\"\n    header \"../hidden.h\"\n    export *\n}\n",
			"Sources/CLib/hidden.h":                 "#include </etc/passwd>\nint hidden(void);\n",
		})
		_, err := fixture.close()
		requireCode(t, err, CodeHeaderInputUndeclared)
	})
	t.Run("module-map header reaches an escape two hops away", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.addFiles(map[string]string{
			"Sources/CLib/include/module.modulemap": "module CLib {\n    header \"CLib.h\"\n    header \"../hidden.h\"\n    export *\n}\n",
			"Sources/CLib/hidden.h":                 "#include \"deeper.h\"\nint hidden(void);\n",
			"Sources/CLib/deeper.h":                 "#include </etc/passwd>\n",
		})
		_, err := fixture.close()
		requireCode(t, err, CodeHeaderInputUndeclared)
	})
	t.Run("module-map header inside the public root is scanned", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.addFiles(map[string]string{
			"Sources/CLib/include/module.modulemap": "module CLib {\n    header \"CLib.h\"\n    header \"extra.h\"\n    export *\n}\n",
			"Sources/CLib/include/extra.h":          "#include </etc/passwd>\nint extra(void);\n",
		})
		_, err := fixture.close()
		requireCode(t, err, CodeHeaderInputUndeclared)
	})
	t.Run("extern module map is confined and its headers are scanned", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.addFiles(map[string]string{
			"Sources/CLib/include/module.modulemap":         "module CLib {\n    umbrella header \"CLib.h\"\n    export *\n}\nextern module CLibPrivate \"module.private.modulemap\"\n",
			"Sources/CLib/include/module.private.modulemap": "module CLibPrivate {\n    header \"../hidden.h\"\n    export *\n}\n",
			"Sources/CLib/hidden.h":                         "#include </etc/passwd>\nint hidden(void);\n",
		})
		_, err := fixture.close()
		requireCode(t, err, CodeHeaderInputUndeclared)
	})
	t.Run("module-map umbrella directory members are scanned", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.addFiles(map[string]string{
			"Sources/CLib/include/module.modulemap": "module CLib {\n    umbrella \"../shared\"\n    export *\n}\n",
			"Sources/CLib/shared/member.h":          "#include </etc/passwd>\nint member(void);\n",
		})
		_, err := fixture.close()
		requireCode(t, err, CodeHeaderInputUndeclared)
	})
	t.Run("contained module-map closure is admitted and recorded", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.addFiles(map[string]string{
			"Sources/CLib/include/module.modulemap": "module CLib {\n    header \"CLib.h\"\n    header \"../hidden.h\"\n    export *\n}\n",
			"Sources/CLib/hidden.h":                 "#include <stdint.h>\nint hidden(void);\n",
		})
		result := fixture.mustClose()
		sources := map[string]bool{}
		for _, reference := range mustTarget(t, result, "root:CLib").Includes {
			sources[reference.Source] = true
		}
		if !sources["Sources/CLib/hidden.h"] {
			t.Fatalf("module-map member was never scanned (scanned %v)", sources)
		}
	})
	t.Run("contained transitive closure is admitted and recorded", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\n#include \"private.h\"\nint value(void) { return 1; }\n"
		fixture.addFiles(map[string]string{
			"Sources/CLib/private.h": "#include \"deeper.h\"\n#include <stdint.h>\n",
			"Sources/CLib/deeper.h":  "#include \"private.h\"\nint deeper(void);\n",
		})
		result := fixture.mustClose()
		sources := map[string]bool{}
		for _, reference := range mustTarget(t, result, "root:CLib").Includes {
			sources[reference.Source] = true
		}
		for _, wanted := range []string{"Sources/CLib/lib.c", "Sources/CLib/private.h", "Sources/CLib/deeper.h"} {
			if !sources[wanted] {
				t.Fatalf("transitive closure omitted %s (scanned %v)", wanted, sources)
			}
		}
	})
}

// H13: directive recognition must run on the translation the target compiler
// actually performs. Every spelling below was confirmed against the pinned
// Apple Clang on the accepted Darwin profile with `clang -std=c17 -H`: each one
// reads the named header, while the line-anchored regular expression this
// scanner used dropped it silently.
func TestH13DirectiveRecognitionMatchesTheCompilerTranslation(t *testing.T) {
	rejected := map[string]string{
		"spliced keyword":            "#inc\\\nlude </etc/passwd>\nint value(void) { return 1; }\n",
		"comment prefix":             "/* */ #include </etc/passwd>\nint value(void) { return 1; }\n",
		"form feed prefix":           "\f#include </etc/passwd>\nint value(void) { return 1; }\n",
		"digraph":                    "%:include </etc/passwd>\nint value(void) { return 1; }\n",
		"mid-line module import":     "int x = 1;\nint y = 1; @import SecretKit;\nint value(void) { return 1; }\n",
		"spliced operand control":    "#include \\\n</etc/passwd>\nint value(void) { return 1; }\n",
		"crlf literal control":       "#include </etc/passwd>\r\nint value(void) { return 1; }\n",
		"multi-line comment restart": "int a;\n/*\n*/ #include </etc/passwd>\nint value(void) { return 1; }\n",
		"unclassifiable directive":   "#curator_secret </etc/passwd>\nint value(void) { return 1; }\n",
		"line marker":                "# 1 \"/etc/passwd\"\nint value(void) { return 1; }\n",
		"unresolvable module import": "@import 3Secret;\nint value(void) { return 1; }\n",
	}
	for name, body := range rejected {
		t.Run(name, func(t *testing.T) {
			fixture := newFixture(t)
			fixture.files["Sources/CLib/lib.c"] = body
			_, err := fixture.close()
			requireCode(t, err, CodeHeaderInputUndeclared)
		})
	}
	t.Run("comment inside a directive does not end it", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include /*\n*/ \"CLib.h\"\nint value(void) { return 1; }\n"
		result := fixture.mustClose()
		if references := mustTarget(t, result, "root:CLib").Includes; len(references) != 1 || references[0].Spelling != "CLib.h" {
			t.Fatalf("scanned references = %#v", references)
		}
	})
	t.Run("classifiable directives are not rejected", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#if 0\n#warning skipped\n#elif defined(NOPE)\n#else\n#pragma once\n#endif\n#undef NOPE\n#\n#include \"CLib.h\" // the public interface\n#line 12\nint value(void) { return 1; }\n"
		result := fixture.mustClose()
		if references := mustTarget(t, result, "root:CLib").Includes; len(references) != 1 || references[0].Spelling != "CLib.h" {
			t.Fatalf("scanned references = %#v", references)
		}
	})
	t.Run("literals and quoted comment markers do not desynchronize", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#error unreachable\nstatic const char marker = '\\'';\nstatic const char *note = \"// not a comment\";\n#include \"CLib.h\"\nint value(void) { return 1; }\n"
		result := fixture.mustClose()
		if references := mustTarget(t, result, "root:CLib").Includes; len(references) != 1 || references[0].Spelling != "CLib.h" {
			t.Fatalf("scanned references = %#v", references)
		}
	})
	t.Run("swift sources are not scanned with the C grammar", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/App/main.swift"] = "#if os(macOS)\nimport CLib\n#endif\n#Preview { }\nprint(1)\n"
		fixture.mustClose()
	})
}

// H14: translation phase 1 replaces trigraphs, and whether it does depends on
// the language mode. Verified on the accepted Darwin profile with
// `clang -fsyntax-only -H`: `??=include "secret.h"` reads the header under
// `-std=c89`, `-std=c99`, `-std=c11`, `-std=c17`, and `-std=c++14`, and does
// not under the GNU default, `-std=gnu17`, `-std=c++17`, or Objective-C++;
// while `int a;??/`<newline>`#include "secret.h"` reads it under the GNU
// default and would splice the line away under an ISO mode. Neither reading is
// safe to assume per file, so a source holding any trigraph is rejected.
func TestH14TrigraphSequencesFailClosed(t *testing.T) {
	rejected := map[string]string{
		"trigraph hash directive":  "??=include </etc/passwd>\nint value(void) { return 1; }\n",
		"trigraph spliced keyword": "#inc??/\nlude </etc/passwd>\nint value(void) { return 1; }\n",
		"trigraph hides a splice":  "#include \"CLib.h\"\nint a;??/\n#include </etc/passwd>\nint value(void) { return 1; }\n",
		"trigraph admitted target": "??=include \"CLib.h\"\nint value(void) { return 1; }\n",
		"trigraph inside literal":  "#include \"CLib.h\"\nstatic const char *note = \"??!\";\nint value(void) { return 1; }\n",
		"trigraph in a header":     "#include \"CLib.h\"\nint value(void) { return 1; }\n",
	}
	for name, body := range rejected {
		t.Run(name, func(t *testing.T) {
			fixture := newFixture(t)
			fixture.files["Sources/CLib/lib.c"] = body
			if name == "trigraph in a header" {
				fixture.files["Sources/CLib/include/CLib.h"] = "#ifndef CLIB_H\n#define CLIB_H\n??=include </etc/passwd>\n#endif\n"
			}
			_, err := fixture.close()
			requireCode(t, err, CodeHeaderInputUndeclared)
		})
	}
	t.Run("a lone question mark pair is not a trigraph", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\nstatic const char *note = \"??x\";\nint value(void) { return 1; }\n"
		result := fixture.mustClose()
		if references := mustTarget(t, result, "root:CLib").Includes; len(references) != 1 || references[0].Spelling != "CLib.h" {
			t.Fatalf("scanned references = %#v", references)
		}
	})
}

// H15: the white space that may precede a directive is not only ASCII. Verified
// on the accepted Darwin profile with `clang -fsyntax-only -H`: a leading UTF-8
// BOM, U+0085, U+00A0, U+1680, U+2000, U+200A, U+2028, U+2029, U+202F, U+205F,
// U+3000, and an embedded NUL each keep `#include` a directive that reads the
// named header, while U+200B and an ordinary non-ASCII identifier byte do not.
// Treating those bytes as content silently demoted the line and dropped the
// header it named; a byte sequence this grammar cannot classify as white space
// is rejected when a directive follows it rather than classified on a guess.
func TestH15NonASCIIWhiteSpaceBeforeADirective(t *testing.T) {
	rejected := map[string]string{
		"leading utf-8 bom":          "\ufeff#include </etc/passwd>\nint value(void) { return 1; }\n",
		"mid-file utf-8 bom":         "int a;\n\ufeff#include </etc/passwd>\nint value(void) { return 1; }\n",
		"no-break space":             "\u00a0#include </etc/passwd>\nint value(void) { return 1; }\n",
		"ogham space mark":           "\u1680#include </etc/passwd>\nint value(void) { return 1; }\n",
		"en quad":                    "\u2000#include </etc/passwd>\nint value(void) { return 1; }\n",
		"hair space":                 "\u200a#include </etc/passwd>\nint value(void) { return 1; }\n",
		"line separator":             "\u2028#include </etc/passwd>\nint value(void) { return 1; }\n",
		"paragraph separator":        "\u2029#include </etc/passwd>\nint value(void) { return 1; }\n",
		"narrow no-break space":      "\u202f#include </etc/passwd>\nint value(void) { return 1; }\n",
		"medium mathematical space":  "\u205f#include </etc/passwd>\nint value(void) { return 1; }\n",
		"ideographic space":          "\u3000#include </etc/passwd>\nint value(void) { return 1; }\n",
		"bom after ascii space":      "  \ufeff\t#include </etc/passwd>\nint value(void) { return 1; }\n",
		"bom in a header":            "#include \"CLib.h\"\nint value(void) { return 1; }\n",
		"space inside the directive": "#\u00a0include </etc/passwd>\nint value(void) { return 1; }\n",
		"zero width space":           "\u200b#include </etc/passwd>\nint value(void) { return 1; }\n",
		"identifier byte":            "\u00e9#include </etc/passwd>\nint value(void) { return 1; }\n",
	}
	for name, body := range rejected {
		t.Run(name, func(t *testing.T) {
			fixture := newFixture(t)
			fixture.files["Sources/CLib/lib.c"] = body
			if name == "bom in a header" {
				fixture.files["Sources/CLib/include/CLib.h"] = "#ifndef CLIB_H\n#define CLIB_H\n\ufeff#include </etc/passwd>\n#endif\n"
			}
			_, err := fixture.close()
			requireCode(t, err, CodeHeaderInputUndeclared)
		})
	}
	// An embedded NUL, a C1 control such as U+0085, and an invalid UTF-8 byte
	// never reach the scanner: the shared recursive artifact classifier rejects
	// the source as opaque first. The scanner still classifies a NUL as the
	// white space Clang skips, so the two layers agree rather than depend on
	// each other's ordering.
	for name, body := range map[string]string{
		"embedded nul":       "\x00#include </etc/passwd>\nint value(void) { return 1; }\n",
		"next line":          "\u0085#include </etc/passwd>\nint value(void) { return 1; }\n",
		"invalid utf-8 byte": "\xff#include </etc/passwd>\nint value(void) { return 1; }\n",
	} {
		t.Run(name, func(t *testing.T) {
			fixture := newFixture(t)
			fixture.files["Sources/CLib/lib.c"] = body
			_, err := fixture.close()
			if ErrorCode(err) != "" || err == nil {
				t.Fatalf("expected the shared artifact policy to reject the source first, got %v", err)
			}
			if !strings.Contains(err.Error(), "artifact_opaque_dependency_forbidden") {
				t.Fatalf("error = %v, want an opaque-dependency rejection", err)
			}
		})
	}
	t.Run("a bom before an admitted directive is scanned", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "\ufeff#include \"CLib.h\"\nint value(void) { return 1; }\n"
		result := fixture.mustClose()
		if references := mustTarget(t, result, "root:CLib").Includes; len(references) != 1 || references[0].Spelling != "CLib.h" {
			t.Fatalf("scanned references = %#v", references)
		}
	})
	t.Run("a non-ascii identifier that introduces no directive is content", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\n\u00e9_marker = 1;\nint value(void) { return 1; }\n"
		result := fixture.mustClose()
		if references := mustTarget(t, result, "root:CLib").Includes; len(references) != 1 || references[0].Spelling != "CLib.h" {
			t.Fatalf("scanned references = %#v", references)
		}
	})
	t.Run("non-ascii white space mid line does not introduce a directive", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\nint a = 1; \u00a0 int b = 2;\nint value(void) { return 1; }\n"
		result := fixture.mustClose()
		if references := mustTarget(t, result, "root:CLib").Includes; len(references) != 1 || references[0].Spelling != "CLib.h" {
			t.Fatalf("scanned references = %#v", references)
		}
	})
}

// H16: a Clang module import is recognized at the token level, not by literal
// adjacency. Every rejected spelling below was confirmed to import on the
// pinned Apple Clang with `-fmodules -fimplicit-module-maps` against a module
// whose only header is an `#error` marker, so a read really does happen; each
// was invisible to a scanner that required the exact bytes `@import` and that
// consumed an unterminated quote to the end of its line.
func TestH16ModuleImportRecognitionMatchesTheCompiler(t *testing.T) {
	rejected := map[string]string{
		"space between tokens":    "@ import SecretKit;\nint value(void) { return 1; }\n",
		"comment between tokens":  "@/*c*/import SecretKit;\nint value(void) { return 1; }\n",
		"line break between":      "@\nimport SecretKit;\nint value(void) { return 1; }\n",
		"non-ascii space between": "@\u00a0import SecretKit;\nint value(void) { return 1; }\n",
		"comment before the name": "@import /*c*/ SecretKit;\nint value(void) { return 1; }\n",
		"line break before semi":  "@import SecretKit\n;\nint value(void) { return 1; }\n",
		"import inside a token":   "int x@import SecretKit;\nint value(void) { return 1; }\n",
	}
	for name, body := range rejected {
		t.Run(name, func(t *testing.T) {
			fixture := newFixture(t)
			fixture.files["Sources/CLib/lib.c"] = body
			_, err := fixture.close()
			requireCode(t, err, CodeHeaderInputUndeclared)
		})
	}
	// An unbalanced quote never reaches the scanner: the shared recursive
	// artifact classifier rejects the source as opaque first. The scanner still
	// treats a delimiter with no partner on its logical line as one ordinary
	// byte, so both layers reject the import rather than depend on each other's
	// ordering — verified for `int x = 1'0; @import Secret;`, which the pinned
	// compiler does import under `-std=c++14 -fcxx-modules`.
	for name, body := range map[string]string{
		"digit separator prefix": "int x = 1'0; @import SecretKit;\nint value(void) { return 1; }\n",
		"unterminated string":    "const char *s = \"open; @import SecretKit;\nint value(void) { return 1; }\n",
	} {
		t.Run(name, func(t *testing.T) {
			fixture := newFixture(t)
			fixture.files["Sources/CLib/lib.c"] = body
			_, err := fixture.close()
			if err == nil || !strings.Contains(err.Error(), "artifact_opaque_dependency_forbidden") {
				t.Fatalf("expected the shared artifact policy to reject the source first, got %v", err)
			}
		})
	}
	t.Run("an objective-c string literal is not a module import", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\nstatic const void *s = @\"@import SecretKit;\";\nint value(void) { return 1; }\n"
		result := fixture.mustClose()
		if references := mustTarget(t, result, "root:CLib").Includes; len(references) != 1 || references[0].Spelling != "CLib.h" {
			t.Fatalf("scanned references = %#v", references)
		}
	})
	t.Run("balanced digit separators are ordinary content", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\nstatic long v = 1'000'000;\nint value(void) { return 1; }\n"
		result := fixture.mustClose()
		if references := mustTarget(t, result, "root:CLib").Includes; len(references) != 1 || references[0].Spelling != "CLib.h" {
			t.Fatalf("scanned references = %#v", references)
		}
	})
	t.Run("an at sign that introduces no import is content", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\n@interface Marker\n@end\nint value(void) { return 1; }\n"
		result := fixture.mustClose()
		if references := mustTarget(t, result, "root:CLib").Includes; len(references) != 1 || references[0].Spelling != "CLib.h" {
			t.Fatalf("scanned references = %#v", references)
		}
	})
}

// H17: the compiler's file-reading channels are wider than the translation
// phases the scanner models. Each rejected spelling below was confirmed on the
// accepted Darwin profile to make the pinned Apple Clang open a file, name a
// module, or redirect include resolution, while the delivered v3 grammar
// dropped it as a benign directive or never saw it at all.
//
// Under the reject-by-default posture the `#embed` family no longer needs a
// per-spelling rule: the directive is outside the admitted inclusion set, so
// every operand form rejects at the directive name. The vectors are kept as
// they were written because they pin the outcome, not the mechanism.
func TestH17CompilerFileReadingChannelsAreClosed(t *testing.T) {
	rejected := map[string]string{
		// `#embed` is the C23 resource-inclusion directive, honoured in the
		// default GNU C mode SwiftPM selects with only a `-Wc23-extensions`
		// warning. A deliberately wrong `_Static_assert` on the resulting array
		// size proves the bytes really arrive. Portable mode admits no operand
		// form of it.
		"embed of an absolute path":    "static const unsigned char d[] = {\n#embed </etc/passwd>\n};\nint value(void) { return 1; }\n",
		"embed escaping the package":   "static const unsigned char d[] = {\n#embed \"../../../../../../../../etc/passwd\"\n};\nint value(void) { return 1; }\n",
		"embed with a macro operand":   "#define SECRET </etc/passwd>\nstatic const unsigned char d[] = {\n#embed SECRET\n};\nint value(void) { return 1; }\n",
		"embed with a limit parameter": "static const unsigned char d[] = {\n#embed \"data.h\" limit(1)\n};\nint value(void) { return 1; }\n",
		"embed in a header":            "#include \"CLib.h\"\nint value(void) { return 1; }\n",
		// A wholly contained operand rejects too: the directive is the channel,
		// not the escape. This vector was a positive control before the pivot.
		"embed of a contained file": "#include \"CLib.h\"\nstatic const unsigned char d[] = {\n#embed \"data.h\"\n};\nint value(void) { return 1; }\n",
		// The pragma spelling is the only module-import form in a plain C
		// translation unit: `@import` there is a syntax error.
		"pragma module import":          "#pragma clang module import SecretKit\nint value(void) { return 1; }\n",
		"_Pragma module import":         "_Pragma(\"clang module import SecretKit\")\nint value(void) { return 1; }\n",
		"pragma module import no name":  "#pragma clang module import\nint value(void) { return 1; }\n",
		"pragma module import trailing": "#pragma clang module import SecretKit extra\nint value(void) { return 1; }\n",
		// An inline module map between build/endbuild can name an absolute
		// out-of-package header and Clang parses it — the H03 escape through a
		// channel the module-map stage never sees.
		"pragma module build":          "#pragma clang module build SecretKit\nmodule SecretKit { header \"/etc/passwd\" }\n#pragma clang module endbuild\nint value(void) { return 1; }\n",
		"pragma module endbuild alone": "#pragma clang module endbuild\nint value(void) { return 1; }\n",
		"pragma module load":           "#pragma clang module load \"secret.pcm\"\nint value(void) { return 1; }\n",
		"pragma module begin":          "#pragma clang module begin SecretKit\n#pragma clang module end\nint value(void) { return 1; }\n",
		// `#pragma include_alias` really substitutes the aliased file under
		// `-fms-extensions`; it is inert but silently accepted without it.
		"pragma include_alias":  "#pragma include_alias(\"CLib.h\", \"/etc/passwd\")\n#include \"CLib.h\"\nint value(void) { return 1; }\n",
		"pragma GCC dependency": "#pragma GCC dependency \"/etc/passwd\"\nint value(void) { return 1; }\n",
		// A `_Pragma` operand that is not an exact literal cannot be classified,
		// and `_Pragma(M)` with `M` defined as the import string does import.
		"_Pragma with a macro operand": "#define M \"clang module import SecretKit\"\n_Pragma(M)\nint value(void) { return 1; }\n",
		// An encoding-prefixed literal is a spelling this grammar declines to
		// read rather than guess at: `_Pragma(u8"clang module import SecretKit")`
		// really does import on the pinned compiler.
		"_Pragma with a wide literal": "_Pragma(L\"clang module import SecretKit\")\nint value(void) { return 1; }\n",
		"_Pragma with a u8 literal":   "_Pragma(u8\"clang module import SecretKit\")\nint value(void) { return 1; }\n",
		// The Microsoft `__pragma` spelling takes raw tokens and imports the
		// module under `-fms-extensions`.
		"__pragma module import":      "__pragma(clang module import SecretKit)\nint value(void) { return 1; }\n",
		"__pragma include_alias":      "__pragma(include_alias(\"CLib.h\", \"/etc/passwd\"))\n#include \"CLib.h\"\nint value(void) { return 1; }\n",
		"_Pragma without parentheses": "_Pragma \"clang diagnostic push\"\nint value(void) { return 1; }\n",
		// A `_Pragma` hidden in a macro definition expands at a site no grammar
		// short of a preprocessor can recognize, so the definition is where it
		// has to be classified.
		"_Pragma inside a define":    "#define IMP _Pragma(\"clang module import SecretKit\")\nIMP\nint value(void) { return 1; }\n",
		"stringizing _Pragma define": "#define DO(x) _Pragma(#x)\nDO(clang module import SecretKit)\nint value(void) { return 1; }\n",
	}
	for name, body := range rejected {
		t.Run(name, func(t *testing.T) {
			fixture := newFixture(t)
			fixture.files["Sources/CLib/lib.c"] = body
			fixture.files["Sources/CLib/data.h"] = "0x41, 0x42,\n"
			if name == "embed in a header" {
				fixture.files["Sources/CLib/include/CLib.h"] = "#ifndef CLIB_H\n#define CLIB_H\nstatic const unsigned char d[] = {\n#embed </etc/passwd>\n};\n#endif\n"
			}
			_, err := fixture.close()
			requireCode(t, err, CodeHeaderInputUndeclared)
		})
	}
	// An unbalanced `__pragma` operand never reaches the scanner: the shared
	// recursive artifact classifier rejects the source as opaque first. The
	// scanner still rejects an operand whose parentheses it cannot balance, so
	// both layers close the channel rather than depend on each other's ordering.
	t.Run("unbalanced __pragma operand", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "__pragma(clang module import SecretKit\nint value(void) { return 1; }\n"
		_, err := fixture.close()
		if err == nil || !strings.Contains(err.Error(), "artifact_opaque_dependency_forbidden") {
			t.Fatalf("expected the shared artifact policy to reject the source first, got %v", err)
		}
		root := t.TempDir()
		if writeErr := os.WriteFile(filepath.Join(root, "probe.c"), []byte("__pragma(clang module import SecretKit\n"), 0o600); writeErr != nil {
			t.Fatal(writeErr)
		}
		if _, scanErr := scanIncludes("root", "CLib", "root", root, "probe.c"); ErrorCode(scanErr) != CodeHeaderInputUndeclared {
			t.Fatalf("scanner verdict = %v", scanErr)
		}
	})
	t.Run("a pragma module import of a declared module resolves", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\n#pragma clang module import Foundation\n_Pragma(\"clang module import CLib\")\nint value(void) { return 1; }\n"
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
	})
	// Positive control for the pragma allowlist: every spelling an ordinary
	// SwiftPM C-family target uses stays content. `#pragma comment(lib, …)` is
	// deliberately absent — it was proven inert in round 5 and is rejected now
	// anyway, because it names a library and the allowlist admits only
	// spellings that name nothing.
	t.Run("allowlisted pragmas are content", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\n#pragma once\n#pragma pack(1)\n#pragma mark - section\n#pragma push_macro(\"X\")\n#pragma pop_macro(\"X\")\n#pragma unused(x)\n#pragma STDC FP_CONTRACT ON\n#pragma clang diagnostic push\n#pragma clang attribute push\n#pragma clang system_header\n#pragma clang assume_nonnull begin\n#pragma GCC visibility push(default)\n#pragma GCC diagnostic ignored \"-Wall\"\n_Pragma(\"clang diagnostic pop\")\n__pragma(pack(1))\n#pragma\nint value(void) { return 1; }\n"
		result := fixture.mustClose()
		if references := mustTarget(t, result, "root:CLib").Includes; len(references) != 1 || references[0].Spelling != "CLib.h" {
			t.Fatalf("scanned references = %#v", references)
		}
	})
	// `#define JOIN(a, b) a##b` used to sit in this positive control. Finding M
	// proved parameter pasting is a channel, so it moved to H21 as a deliberate
	// narrowing; the rest of the vector is unchanged.
	t.Run("an ordinary define is not a pragma channel", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\n#define WIDTH 8\n#define TEXT \"_Pragma is only a word here\"\nint value(void) { return WIDTH; }\n"
		result := fixture.mustClose()
		if references := mustTarget(t, result, "root:CLib").Includes; len(references) != 1 || references[0].Spelling != "CLib.h" {
			t.Fatalf("scanned references = %#v", references)
		}
	})
	t.Run("an identifier that ends in _Pragma is content", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\nstatic int my_Pragma = 1;\nstatic int _Pragmatic = 2;\nint value(void) { return my_Pragma + _Pragmatic; }\n"
		result := fixture.mustClose()
		if references := mustTarget(t, result, "root:CLib").Includes; len(references) != 1 || references[0].Spelling != "CLib.h" {
			t.Fatalf("scanned references = %#v", references)
		}
	})
}

// H18: inline assembly is a construct portable mode rejects, not a spelling it
// classifies.
//
// The integrated assembler is a second file-reading stage inside the same
// `clang -c` invocation and it shares no token with the preprocessor grammar:
// `__asm__(".incbin \"/etc/passwd\"");` at file scope in a plain `.c` exits 0
// and puts that file's bytes in the object, while `-H` reports no read at all.
// Rounds 5 through 8 each closed one more spelling of that channel and each
// time a further one was found, because proving what an assembler reads means
// reproducing the assembler. Portable mode stops trying: any `asm`, `__asm`, or
// `__asm__` at a token boundary rejects the target, and the adversarially
// complete acceptance of inline assembly is deferred to the observed-read
// provider.
//
// Every vector below is a spelling some earlier round had to name individually
// — the two file-reading directives, the linker-option and secure-log
// directives, the macro-expansion layer, the literal and escape forms that hid
// a directive name, the aliased and macro-built keywords. They are retained
// verbatim so the pivot is proved to lose no coverage, and they now all reject
// through one rule.
func TestH18InlineAssemblyRejectsTheTarget(t *testing.T) {
	bodies := map[string]string{
		"inline asm incbin absolute":        "__asm__(\".incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
		"inline asm incbin escaping":        "__asm__(\".incbin \\\"../../../../../../../../etc/passwd\\\"\");\nint value(void) { return 1; }\n",
		"inline asm incbin relative":        "__asm__(\".incbin \\\"CLib.h\\\"\");\nint value(void) { return 1; }\n",
		"inline asm include":                "asm(\".include \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
		"inline asm dump":                   "__asm__(\".dump \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
		"inline asm load":                   "__asm__(\".load \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
		"inline asm linker option":          "__asm__(\".linker_option \\\"-lSecretProbeLib\\\"\");\nint value(void) { return 1; }\n",
		"inline asm secure log":             "__asm__(\".secure_log_unique \\\"marker\\\"\");\nint value(void) { return 1; }\n",
		"inline asm incbin in header":       "#include \"CLib.h\"\nint value(void) { return 1; }\n",
		"inline asm with a macro operand":   "#define TPL \".incbin \\\"/etc/passwd\\\"\"\n__asm__(TPL);\nint value(void) { return 1; }\n",
		"inline asm with a wide template":   "__asm__(L\".incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
		"reserved asm keyword with a block": "__asm { nop }\nint value(void) { return 1; }\n",
		"reserved asm keyword bare":         "__asm__ nop;\nint value(void) { return 1; }\n",
		"aliased bare asm keyword":          "#define K asm\nK(\".incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
		"aliased reserved asm keyword":      "#define K __asm__\nK(\".incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
		"asm statement built by a macro":    "#define STMT(x) __asm__(x)\nSTMT(\".incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
		"asm operand assembled by include":  "__asm__(\n#include \"tpl.h\"\n);\nint value(void) { return 1; }\n",
		"bare asm used as an identifier":    "struct holder { int asm; };\nint value(void) { return 1; }\n",
		"inline asm literal concatenation":  "__asm__(\".incbin \" \"\\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
		"inline asm hex escaped directive":  "__asm__(\"\\x2eincbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
		"inline asm octal escaped comment":  "__asm__(\"\\056incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
		"inline asm delimited hex escape":   "__asm__(\"\\x{2e}incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
		"inline asm delimited octal escape": "__asm__(\"\\o{56}incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
		"inline asm universal character":    "__asm__(\"\\u{2e}incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
		"inline asm named character":        "__asm__(\"\\N{FULL STOP}incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
		"inline asm inside a define":        "#define BOOT __asm__(\".incbin \\\"/etc/passwd\\\"\")\nBOOT;\nint value(void) { return 1; }\n",
		"incbin inside an asm macro body":   "__asm__(\".macro emb\\n.incbin \\\"/etc/passwd\\\"\\n.endm\\nemb\\n\");\nint value(void) { return 1; }\n",

		"asm macro argument builds a directive name": "__asm__(\".macro D a\\n.\\\\a \\\"/etc/passwd\\\"\\n.endm\\nD incbin\\n\");\nint value(void) { return 1; }\n",
		"asm irp builds a directive name":            "__asm__(\".irp x,incbin\\n.\\\\x \\\"/etc/passwd\\\"\\n.endr\\n\");\nint value(void) { return 1; }\n",
		"asm empty separator splices a directive":    "__asm__(\".macro D a\\n.inc\\\\a\\\\()bin \\\"/etc/passwd\\\"\\n.endm\\nD \\\"\\\"\\n\");\nint value(void) { return 1; }\n",
		"asm two arguments splice a directive":       "__asm__(\".macro D a b\\n.\\\\a\\\\b \\\"/etc/passwd\\\"\\n.endm\\nD inc bin\\n\");\nint value(void) { return 1; }\n",
		"asm macro argument builds an include":       "__asm__(\".macro D a\\n.\\\\a \\\"/etc/passwd\\\"\\n.endm\\nD include\\n\");\nint value(void) { return 1; }\n",
		"asm altmacro ampersand concatenation":       "__asm__(\".altmacro\\n.macro D a\\n.inc&a&bin \\\"/etc/passwd\\\"\\n.endm\\nD bin\\n\");\nint value(void) { return 1; }\n",

		// The construct rejects even when its template names nothing. These
		// four were the round-7/8 positive control: portable mode no longer
		// distinguishes an ordinary template from a file-reading one, because
		// that distinction is exactly the emulation it declines to carry.
		"an ordinary assembler body":     "#include \"CLib.h\"\n__asm__(\".align 4\\n.arch armv8-a\\nnop\\n\");\nint value(void) { return 1; }\n",
		"a qualified extended statement": "#include \"CLib.h\"\nint value(void) { int r = 0; __asm__ volatile (\"mov %0, #1\" : \"=r\"(r) : : \"memory\"); return r; }\n",
		"an assembler symbol label":      "#include \"CLib.h\"\nextern int named __asm__(\"_named_symbol\");\nint value(void) { return 1; }\n",
		"a bare volatile statement":      "#include \"CLib.h\"\n__asm__ __volatile__(\"nop\");\nint value(void) { return 1; }\n",

		// Finding L: a backslash separated from its newline by horizontal white
		// space is a splice on the pinned compiler in all six separator
		// spellings and all twelve language modes tried. Before phase 2 was
		// completed a split *inside* the keyword left no residual for any
		// grammar to fail closed on, and `__as\`+ws+newline+`m__` produced an
		// object byte-identical to the direct `.incbin` control.
		"asm keyword split by a bare splice":      "__as\\\nm__(\".incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
		"asm keyword split by a space splice":     "__as\\ \nm__(\".incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
		"asm keyword split by a tab splice":       "__as\\\t\nm__(\".incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
		"asm keyword split by a form feed splice": "__as\\\f\nm__(\".incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
		"asm keyword split by a mixed splice":     "__as\\ \t \nm__(\".incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
		"asm keyword split by a CRLF splice":      "__as\\ \r\nm__(\".incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
	}
	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			fixture := newFixture(t)
			fixture.files["Sources/CLib/lib.c"] = body
			if name == "asm operand assembled by include" {
				fixture.files["Sources/CLib/tpl.h"] = "\".incbin \\\"/etc/passwd\\\"\"\n"
			}
			if name == "inline asm incbin in header" {
				fixture.files["Sources/CLib/include/CLib.h"] = "#ifndef CLIB_H\n#define CLIB_H\n__asm__(\".incbin \\\"/etc/passwd\\\"\");\nint value(void);\n#endif\n"
			}
			_, err := fixture.close()
			requireCode(t, err, CodeTargetPlatformUnsupported)
		})
	}
	// A C++ raw-string template is a construct this grammar declines to read
	// and the pinned compiler really does read through it: `__asm__(R"(.incbin
	// "/etc/passwd")")` in a `.cpp` embeds those bytes.
	t.Run("raw string asm template in a C++ source", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\nint value(void) { return 1; }\n"
		fixture.files["Sources/CLib/impl.cpp"] = "__asm__(R\"(.incbin \"/etc/passwd\")\");\nint cxx_value(void) { return 2; }\n"
		fixture.target("CLib").Sources = append(fixture.target("CLib").Sources, "Sources/CLib/impl.cpp")
		_, err := fixture.close()
		requireCode(t, err, CodeTargetPlatformUnsupported)
	})
	// A C-family target that declares an assembly source is unsupported rather
	// than partially inspected. The C preprocessor grammar models no assembler
	// directive, so `.incbin "/etc/passwd"` in a `.S` file was invisible; a
	// lowercase `.s` file is not preprocessed at all, so scanning it with that
	// grammar would also reject an ordinary `# comment` line.
	for name, source := range map[string]string{
		"uppercase assembly source": "Sources/CLib/boot.S",
		"lowercase assembly source": "Sources/CLib/boot.s",
	} {
		t.Run(name, func(t *testing.T) {
			fixture := newFixture(t)
			fixture.files[source] = ".incbin \"/etc/passwd\"\n.globl _boot\n_boot:\n  ret\n"
			fixture.target("CLib").Sources = append(fixture.target("CLib").Sources, source)
			_, err := fixture.close()
			requireCode(t, err, CodeTargetPlatformUnsupported)
		})
	}
	// Positive control: recognition is anchored to a token boundary, so an
	// identifier that merely contains an assembly keyword is content. Without
	// this the reject-by-default rule would swallow ordinary C.
	t.Run("an identifier that contains an assembly keyword is content", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\nstatic int asmx = 1;\nstatic int my__asm__ = 2;\nstatic int __asmt = 3;\nint value(void) { return asmx + my__asm__ + __asmt; }\n"
		result := fixture.mustClose()
		if references := mustTarget(t, result, "root:CLib").Includes; len(references) != 1 || references[0].Spelling != "CLib.h" {
			t.Fatalf("scanned references = %#v", references)
		}
	})
}

// H19: translation phase 2 is reproduced completely, so a line splice can
// neither hide a channel keyword nor invent one.
//
// Finding L proved the incomplete form was an admission hole rather than a
// conservative divergence: `spliceTranslationLines` joined only `\`+newline, so
// a split inside `__asm__`, `_Pragma`, or `@import` left ordinary content bytes
// behind and the construct was never entered, while the compiler reconstituted
// the keyword and performed the read. Unlike phase-1 trigraph replacement the
// behaviour is mode-independent, so the splice is performed rather than
// rejected — which also means an ordinary spliced `#include` still resolves.
func TestH19LineSplicesAreReproducedBeforeRecognition(t *testing.T) {
	rejected := map[string]string{
		"module import split by a space splice":   "@imp\\ \nort SecretKit;\nint value(void) { return 1; }\n",
		"module import split by a tab splice":     "@imp\\\t\nort SecretKit;\nint value(void) { return 1; }\n",
		"pragma operator split by a space splice": "_Pra\\ \ngma(\"clang module import SecretKit\")\nint value(void) { return 1; }\n",
		"pragma operator split by a mixed splice": "_Pra\\ \t \ngma(\"clang module import SecretKit\")\nint value(void) { return 1; }\n",
		"ms pragma operator split by a splice":    "__prag\\ \nma(clang module import SecretKit)\nint value(void) { return 1; }\n",
		"include directive split mid-name":        "#inc\\ \nlude </etc/passwd>\nint value(void) { return 1; }\n",
		"include operand split into an escape":    "#include \"../../../../../../../../etc/pas\\ \nswd\"\nint value(void) { return 1; }\n",
		"pragma directive split mid-name":         "#prag\\ \nma clang module import SecretKit\nint value(void) { return 1; }\n",
	}
	for name, body := range rejected {
		t.Run(name, func(t *testing.T) {
			fixture := newFixture(t)
			fixture.files["Sources/CLib/lib.c"] = body
			_, err := fixture.close()
			requireCode(t, err, CodeHeaderInputUndeclared)
		})
	}
	// Positive control: the splice really is performed rather than rejected, so
	// an ordinary directive split across lines still resolves to its header.
	// `#include "CLi\`+space+newline+`b.h"` above is the same mechanism used to
	// name a file the closure does not hold, and it fails closed there.
	for name, body := range map[string]string{
		"an include split mid-name":           "#inc\\ \nlude \"CLib.h\"\nint value(void) { return 1; }\n",
		"an operand split inside the literal": "#include \"CLi\\ \nb.h\"\nint value(void) { return 1; }\n",
		"an include split before the operand": "#include\\ \t\n \"CLib.h\"\nint value(void) { return 1; }\n",
		"a bare splice inside a macro body":   "#include \"CLib.h\"\n#define WIDE(a, \\\n               b) ((a) + (b))\nint value(void) { return WIDE(1, 0); }\n",
	} {
		t.Run(name, func(t *testing.T) {
			fixture := newFixture(t)
			fixture.files["Sources/CLib/lib.c"] = body
			result := fixture.mustClose()
			if references := mustTarget(t, result, "root:CLib").Includes; len(references) != 1 || references[0].Spelling != "CLib.h" {
				t.Fatalf("scanned references = %#v", references)
			}
		})
	}
}

// H20: the pragma surface is reject-by-default.
//
// A pragma is open by construction to every vendor, so a deny-list grammar can
// never be closed over it: a spelling nobody enumerated falls through as
// content. Portable mode admits a closed allowlist of spellings that name
// nothing — macro state, layout, diagnostics, symbol binding, presentation —
// plus `clang module import NAME`, which is resolved through the same module
// confinement `@import` uses. Everything else rejects, including the three
// spellings earlier rounds had to identify individually and every spelling a
// later toolchain may add.
func TestH20PragmaChannelsAreRejectByDefault(t *testing.T) {
	for name, body := range map[string]string{
		"a comment library pragma":           "#pragma comment(lib, \"SecretProbeLib\")\nint value(void) { return 1; }\n",
		"an unenumerated vendor pragma":      "#pragma acme read_from(\"/etc/passwd\")\nint value(void) { return 1; }\n",
		"an openmp pragma":                   "#pragma omp parallel for\nint value(void) { return 1; }\n",
		"an unknown clang pragma":            "#pragma clang __debug module_map\nint value(void) { return 1; }\n",
		"an unknown GCC pragma":              "#pragma GCC unknown_extension \"x\"\nint value(void) { return 1; }\n",
		"a comment pragma via _Pragma":       "_Pragma(\"comment(lib, \\\"SecretProbeLib\\\")\")\nint value(void) { return 1; }\n",
		"a comment pragma via __pragma":      "__pragma(comment(lib, \"SecretProbeLib\"))\nint value(void) { return 1; }\n",
		"an unenumerated pragma in a define": "#define P _Pragma(\"acme read_from(\\\"/etc/passwd\\\")\")\nP\nint value(void) { return 1; }\n",
	} {
		t.Run(name, func(t *testing.T) {
			fixture := newFixture(t)
			fixture.files["Sources/CLib/lib.c"] = body
			_, err := fixture.close()
			requireCode(t, err, CodeHeaderInputUndeclared)
		})
	}
}

// H21: translation phase 4 cannot deliver a channel keyword the scanner never
// sees.
//
// Phases 1-3 close the lexical reconstitution of `asm`, `__asm`, `__asm__`,
// `_Pragma`, `__pragma`, and `import`, but macro expansion runs after the
// scanner has read the file and the `##` operator joins two tokens into a new
// one. Every rejected spelling below was confirmed on the accepted Darwin
// profile: the two paste forms produce an object byte-identical to the direct
// `.incbin` control, and the three module forms build a module the target never
// declared and read its header, which is an `#error` marker.
//
// The disposition is reject-by-default rather than emulation. A paste that
// takes a fragment from the call site rejects, because the definition cannot
// bound it; a paste of fixed fragments is joined and then scanned like any
// other token stream; and an `@` followed by an identifier outside Objective-C's
// closed `@`-keyword set rejects, because such an identifier can only be a
// macro and a macro there expands to a module import.
func TestH21MacroReconstitutedChannelKeywordsReject(t *testing.T) {
	for name, expect := range map[string]struct {
		body string
		code Code
	}{
		"M1 pasted asm keyword": {
			body: "#define J(a,b) a##b\nJ(a,sm)(\".incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		"M2 pasted reserved asm keyword": {
			body: "#define J(a,b) a##b\nJ(__as,m__)(\".incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		"M3 pasted pragma operator": {
			body: "#define J(a,b) a##b\nJ(_Prag,ma)(\"clang module import SecretKit\")\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		"M4 macro import after an at sign": {
			body: "#define I import\n@ I SecretKit;\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		"M5 macro at sign before an import": {
			body: "#define AT @\nAT import SecretKit;\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		"a fixed paste that builds an asm keyword": {
			body: "#define A __as##m__\nA(\".incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
			code: CodeTargetPlatformUnsupported,
		},
		"a fixed paste chain that builds an asm keyword": {
			body: "#define A a##s##m\nA(\".incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
			code: CodeTargetPlatformUnsupported,
		},
		"a fixed paste that builds a pragma operator": {
			body: "#define A _Prag##ma\nA(\"clang module import SecretKit\")\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		"a fixed paste that builds a module import": {
			body: "#define A @im##port\nA SecretKit;\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		"a paste of a parameter with a fixed fragment": {
			body: "#define S(a) __as##a\nS(m__)(\".incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		"a variadic parameter pasted to an identifier": {
			body: "#define V(...) as##__VA_ARGS__\nV(m)(\".incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		"an at sign before an unknown keyword": {
			body: "#include \"CLib.h\"\n@secretly SecretKit;\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		"an at sign before the older import spelling": {
			body: "#include \"CLib.h\"\n@__experimental_modules_import SecretKit;\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
	} {
		t.Run(name, func(t *testing.T) {
			fixture := newFixture(t)
			fixture.files["Sources/CLib/lib.c"] = expect.body
			_, err := fixture.close()
			requireCode(t, err, expect.code)
		})
	}
	// An unbalanced parameter list never reaches the scanner: the shared
	// recursive artifact classifier rejects the source as opaque first. The
	// scanner still rejects a parameter list it cannot read, so both layers
	// close the channel rather than depend on each other's ordering.
	t.Run("an unreadable function-like parameter list", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#define J(a,b a##b\nint value(void) { return 1; }\n"
		_, err := fixture.close()
		if err == nil || !strings.Contains(err.Error(), "artifact_opaque_dependency_forbidden") {
			t.Fatalf("expected the shared artifact policy to reject the source first, got %v", err)
		}
		root := t.TempDir()
		if writeErr := os.WriteFile(filepath.Join(root, "probe.c"), []byte("#define J(a,b a##b\n"), 0o600); writeErr != nil {
			t.Fatal(writeErr)
		}
		if _, scanErr := scanIncludes("root", "CLib", "root", root, "probe.c"); ErrorCode(scanErr) != CodeHeaderInputUndeclared {
			t.Fatalf("scanner verdict = %v", scanErr)
		}
	})
	// Positive controls. The narrowing is deliberate but bounded: a paste of
	// fixed fragments still resolves, the GNU comma-deletion idiom is untouched
	// because a punctuator contributes no identifier characters, and every
	// Objective-C construct `@` really introduces stays content.
	for name, body := range map[string]string{
		"a fixed paste that builds an ordinary token": "#include \"CLib.h\"\n#define AB foo##bar\nint foobar = 1;\nint value(void) { return AB; }\n",
		"the comma deletion idiom":                    "#include \"CLib.h\"\n#define LOG(fmt, ...) printf(fmt, ##__VA_ARGS__)\nint value(void) { return 1; }\n",
		"objective-c literals and collections":        "#include \"CLib.h\"\nstatic const void *a = @\"s\";\nstatic const void *b = @[ @1, @YES, @NO ];\nstatic const void *c = @{ @\"k\": @(1) };\nint value(void) { return 1; }\n",
		"objective-c at keywords":                     "#include \"CLib.h\"\n@interface Marker\n@property int x;\n@end\n@implementation Marker\n@synthesize x;\n@end\nint value(void) { return @selector(x) != 0; }\n",
	} {
		t.Run(name, func(t *testing.T) {
			fixture := newFixture(t)
			fixture.files["Sources/CLib/lib.c"] = body
			result := fixture.mustClose()
			if references := mustTarget(t, result, "root:CLib").Includes; len(references) != 1 || references[0].Spelling != "CLib.h" {
				t.Fatalf("scanned references = %#v", references)
			}
		})
	}
	// The literal spellings this profile admits are unaffected by the `@` rule.
	t.Run("literal module imports still resolve", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\n@import Foundation;\n_Pragma(\"clang module import CLib\")\n#pragma clang module import Foundation\nint value(void) { return 1; }\n"
		result := fixture.mustClose()
		imports := 0
		for _, reference := range mustTarget(t, result, "root:CLib").Includes {
			if reference.ModuleImport {
				imports++
			}
		}
		if imports != 3 {
			t.Fatalf("recorded %d Clang module imports", imports)
		}
	})
}

// H23: the identifier positions the compiler macro-expands.
//
// H21 closed the reconstitution of a channel *keyword*. This family closes the
// two shapes that need no reconstitution at all: an identifier the scanner
// reads literally that the compiler expands before acting on it. Every vector
// was first confirmed against the pinned Apple Clang (21.0.0,
// `clang-2100.1.1.101`, `arm64-apple-darwin25.5.0`) with `-fmodules` against a
// module whose only header is an `#error` marker, or by comparing object bytes
// against a direct `.incbin` control.
func TestH23MacroExpandedIdentifierPositionsReject(t *testing.T) {
	for name, expect := range map[string]struct {
		body string
		code Code
	}{
		// N1: every member of the `@`-keyword allowlist is an ordinary
		// identifier the preprocessor can rebind, so the allowlist itself was
		// the delivery vehicle. The rejection lands on the definition.
		"N1a macro bound to the protocol keyword": {
			body: "#define protocol import\n@ protocol SecretKit;\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		"N1b macro bound to the class keyword": {
			body: "#define class import\n@ class SecretKit;\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		"N1c macro bound to the selector keyword": {
			body: "#define selector import\n@ selector SecretKit;\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		"N1d macro bound to the end keyword": {
			body: "#define end import\n@ end SecretKit;\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		"N1e macro bound to the YES keyword": {
			body: "#define YES import\n@ YES SecretKit;\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		"N1f keyword macro built by a paste": {
			body: "#define class im##port\n@ class SecretKit;\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		"N1g macro bound to the import spelling itself": {
			body: "#define import protocol\n@import CLib;\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		"N1h macro bound to the older import spelling": {
			body: "#define __experimental_modules_import import\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		// The definition is rejected wherever it sits, including a header the
		// `.c` file merely includes — which is the realistic vector, and the
		// reason this rule lives on `#define` and not on `@`.
		"N1i keyword macro bound in an included header": {
			body: "#include \"CLib.h\"\n@ protocol SecretKit;\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		// N2: the `@import` module name is macro-expanded, so the recorded
		// spelling is neither the module the compiler resolves nor a name the
		// `moduleDeclared` gate may be satisfied by. `CLib` is the fixture's own
		// admitted module, so without this rule the target admits while the
		// compiler imports SecretKit.
		"N2 aliased module import": {
			body: "#define CLib SecretKit\n@import CLib;\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		"N2b aliased module import through the microsoft pragma operator": {
			body: "#define CLib SecretKit\nvoid f(void) { __pragma(clang module import CLib) }\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		"N2c aliased qualified module import": {
			body: "#define CLib SecretKit\n@import CLib.Sub;\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		// N3: `%:%:` is the digraph spelling of `##`. Each of these produced the
		// channel on the real compiler while the paste layer saw nothing.
		"N3a object-like digraph paste to an asm keyword": {
			body: "#define A __as%:%:m__\nA(\".incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
			code: CodeTargetPlatformUnsupported,
		},
		"N3b function-like digraph paste to an asm keyword": {
			body: "#define J(a,b) a%:%:b\nJ(a,sm)(\".incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		"N3c digraph paste to a pragma operator": {
			body: "#define A _Prag%:%:ma\nA(\"clang module import SecretKit\")\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		"N3d digraph paste to a module import": {
			body: "#define A @im%:%:port\nA SecretKit;\nint value(void) { return 1; }\n",
			code: CodeHeaderInputUndeclared,
		},
		"N3e digraph paste chain to an asm keyword": {
			body: "#define A a%:%:s%:%:m\nA(\".incbin \\\"/etc/passwd\\\"\");\nint value(void) { return 1; }\n",
			code: CodeTargetPlatformUnsupported,
		},
	} {
		t.Run(name, func(t *testing.T) {
			fixture := newFixture(t)
			if strings.Contains(name, "included header") {
				fixture.files["Sources/CLib/include/CLib.h"] = "#ifndef CLIB_H\n#define CLIB_H\n#define protocol import\nint value(void);\n#endif\n"
			}
			fixture.files["Sources/CLib/lib.c"] = expect.body
			_, err := fixture.close()
			requireCode(t, err, expect.code)
		})
	}
	// The verified asymmetry, encoded as a control rather than as an assumption:
	// `#pragma clang module import NAME` is NOT macro-expanded, so the same
	// aliasing definition that makes the `@import` form unresolvable leaves the
	// pragma form recording exactly the module the compiler resolves.
	t.Run("the pragma import spelling is not expanded and still admits", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\n#define CLib SecretKit\n#pragma clang module import CLib\n_Pragma(\"clang module import CLib\")\nint value(void) { return 1; }\n"
		result := fixture.mustClose()
		imports := 0
		for _, reference := range mustTarget(t, result, "root:CLib").Includes {
			if reference.ModuleImport {
				if reference.ExpandedName {
					t.Fatalf("pragma module import recorded as macro-expanded: %#v", reference)
				}
				imports++
			}
		}
		if imports != 2 {
			t.Fatalf("recorded %d Clang module imports", imports)
		}
	})
	// Positive controls. The two rules narrow exactly one identifier position
	// each: an ordinary macro name, an ordinary digraph paste, a literal module
	// import, and every Objective-C construct `@` really introduces still admit.
	for name, body := range map[string]string{
		"an ordinary macro definition":                  "#include \"CLib.h\"\n#define WIDTH 8\nint value(void) { return WIDTH; }\n",
		"a digraph paste that builds an ordinary token": "#include \"CLib.h\"\n#define AB foo%:%:bar\nint foobar = 1;\nint value(void) { return AB; }\n",
		"a literal at-keyword after an at sign":         "#include \"CLib.h\"\n@protocol Marker;\n@class Marker2;\nint value(void) { return 1; }\n",
	} {
		t.Run(name, func(t *testing.T) {
			fixture := newFixture(t)
			fixture.files["Sources/CLib/lib.c"] = body
			result := fixture.mustClose()
			if references := mustTarget(t, result, "root:CLib").Includes; len(references) != 1 || references[0].Spelling != "CLib.h" {
				t.Fatalf("scanned references = %#v", references)
			}
		})
	}
	// A literal `@import` of an admitted module still admits and is still
	// recorded, so the N2 rule rejects the aliased name and nothing else.
	t.Run("a literal module import still admits", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\n@import CLib;\n@import Foundation;\nint value(void) { return 1; }\n"
		result := fixture.mustClose()
		imports := 0
		for _, reference := range mustTarget(t, result, "root:CLib").Includes {
			if reference.ModuleImport {
				if !reference.ExpandedName {
					t.Fatalf("@import recorded as not macro-expanded: %#v", reference)
				}
				imports++
			}
		}
		if imports != 2 {
			t.Fatalf("recorded %d Clang module imports", imports)
		}
	})
}

// H22: a C++ raw string literal is a construct this grammar rejects rather than
// reads.
//
// The divergence was load-bearing, not merely conservative. `R"x(" /* )x"`
// hands the scanner an unmatched `"` and then a `/*`; skipBlockComment swallows
// the rest of the file while the compiler sees no comment at all. Verified on
// the accepted Darwin profile: that prologue followed by `__asm__(".incbin
// \"payload.bin\"")` compiles and puts the named file's bytes in the object.
// Whether `R"` opens a raw string is also language-mode dependent, which is the
// same reason a trigraph is rejected rather than translated.
func TestH22RawStringLiteralsRejectTheTarget(t *testing.T) {
	// The scanner's own verdict, independent of which layer sees the source
	// first, so a future artifactpolicy relaxation cannot silently open this.
	for name, payload := range map[string]string{
		"comment swallowing prologue": "const char* s = R\"x(\" /* )x\";\n__asm__(\".incbin \\\"/etc/passwd\\\"\");\n",
		"an ordinary raw string":      "const char* s = R\"(hi)\";\n",
		"a wide raw string":           "const char* s = LR\"(hi)\";\n",
		"a utf8 raw string":           "const char* s = u8R\"(hi)\";\n",
		"a raw string in a define":    "#define S R\"(hi)\"\n",
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			if writeErr := os.WriteFile(filepath.Join(root, "probe.cpp"), []byte(payload), 0o600); writeErr != nil {
				t.Fatal(writeErr)
			}
			if _, scanErr := scanIncludes("root", "CLib", "root", root, "probe.cpp"); ErrorCode(scanErr) != CodeHeaderInputUndeclared {
				t.Fatalf("scanner verdict = %v", scanErr)
			}
		})
	}
	// End to end: an ordinary raw string clears the shared artifact classifier,
	// so this rejection is the interop stage's own.
	t.Run("an ordinary raw string rejects the target", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/impl.cpp"] = "static const char *s = R\"(hi)\";\nint cxx_value(void) { return 2; }\n"
		fixture.target("CLib").Sources = append(fixture.target("CLib").Sources, "Sources/CLib/impl.cpp")
		_, err := fixture.close()
		requireCode(t, err, CodeHeaderInputUndeclared)
	})
	// Positive control: recognition is anchored to a token boundary, so an
	// identifier that merely ends in an encoding prefix is content and ordinary
	// literal concatenation still works.
	t.Run("an identifier that ends in a raw prefix is content", func(t *testing.T) {
		fixture := newFixture(t)
		fixture.files["Sources/CLib/lib.c"] = "#include \"CLib.h\"\n#define myR \"head\"\nstatic const char *t = myR\"tail\";\nstatic int R = 1;\nint value(void) { return R; }\n"
		result := fixture.mustClose()
		if references := mustTarget(t, result, "root:CLib").Includes; len(references) != 1 || references[0].Spelling != "CLib.h" {
			t.Fatalf("scanned references = %#v", references)
		}
	})
}
