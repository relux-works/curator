package artifactpolicy

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestCompiledVectorsC01ELFResolution(t *testing.T) {
	t.Run("C01 ET_EXEC and ET_REL ignore suffix", func(t *testing.T) {
		fixtures := []struct {
			name  string
			bytes []byte
			class ArtifactClass
		}{
			{"program", makeELF64(elfTypeExec, false, false, false), ClassNativeExecutable},
			{"program.elf", makeELF64(elfTypeExec, false, false, false), ClassNativeExecutable},
			{"program.txt", makeELF64(elfTypeExec, false, false, false), ClassNativeExecutable},
			{"object", makeELF64(elfTypeRel, false, false, false), ClassNativeObject},
			{"object.o", makeELF64(elfTypeRel, false, false, false), ClassNativeObject},
			{"object.data", makeELF64(elfTypeRel, false, false, false), ClassNativeObject},
		}
		for _, fixture := range fixtures {
			t.Run(fixture.name, func(t *testing.T) {
				result, err := admitDependency(t, fixture.name, fixture.bytes, ProfileCommonV1)
				requireCode(t, err, CodeCompiledDependency)
				node := requireNode(t, result, fixture.name)
				if node.Class != fixture.class {
					t.Fatalf("class = %q, want %q", node.Class, fixture.class)
				}
			})
		}
	})

	t.Run("C01a dynamic PIE", func(t *testing.T) {
		payload := makeELF64(elfTypeDyn, true, true, false)
		for _, name := range []string{"program", "program.so", "program.dat", "suffixless"} {
			result, err := admitDependency(t, name, payload, ProfileCommonV1)
			requireCode(t, err, CodeCompiledDependency)
			node := requireNode(t, result, name)
			if node.Class != ClassNativeExecutable || node.Variant != "elf.pie.interpreter" {
				t.Fatalf("%s = %s/%s", name, node.Class, node.Variant)
			}
		}
	})

	t.Run("C01b no-interpreter PIE", func(t *testing.T) {
		payload := makeELF64(elfTypeDyn, true, false, false)
		for _, name := range []string{"static-pie", "static-pie.so", "static-pie.dat"} {
			result, err := admitDependency(t, name, payload, ProfileCommonV1)
			requireCode(t, err, CodeCompiledDependency)
			node := requireNode(t, result, name)
			if node.Class != ClassNativeExecutable || node.Variant != "elf.pie.no_interpreter" {
				t.Fatalf("%s node = %+v", name, node)
			}
		}
	})

	t.Run("C01c shared object", func(t *testing.T) {
		payload := makeELF64(elfTypeDyn, false, false, true)
		for _, name := range []string{"libcase.so", "renamed.dat", "suffixless"} {
			result, err := admitDependency(t, name, payload, ProfileCommonV1)
			requireCode(t, err, CodeCompiledDependency)
			node := requireNode(t, result, name)
			if node.Class != ClassNativeLibraryDynamic || node.Variant != "elf.shared_object" {
				t.Fatalf("%s = %s/%s", name, node.Class, node.Variant)
			}
		}
	})

	t.Run("C01d ambiguous ET_DYN", func(t *testing.T) {
		for name, payload := range map[string][]byte{
			"no-evidence": makeELF64(elfTypeDyn, false, false, false),
			"interp-only": makeELF64(elfTypeDyn, false, true, false),
		} {
			result, err := admitDependency(t, name, payload, ProfileCommonV1)
			requireCode(t, err, CodeCompiledDependency)
			node := requireNode(t, result, name)
			if node.Class != ClassELFETDYNAmbiguous || node.Variant != "insufficient_evidence" {
				t.Fatalf("%s = %s/%s", name, node.Class, node.Variant)
			}
		}
	})

	t.Run("C01e conflicting ET_DYN facts", func(t *testing.T) {
		payload := makeELF64(elfTypeDyn, false, true, true)
		result, err := admitDependency(t, "conflict", payload, ProfileCommonV1)
		requireCode(t, err, CodeCompiledDependency)
		node := requireNode(t, result, "conflict")
		if node.Class != ClassELFETDYNAmbiguous || node.Variant != "interp_soname_conflict" {
			t.Fatalf("node = %+v", node)
		}

		useConflict := makeELF64(elfTypeDyn, false, false, false)
		request := dependencyRequest("use-conflict", useConflict, ProfileCommonV1)
		request.Descriptor.ResolvedUses = map[string][]UseEdge{
			"use-conflict": {
				{Kind: UseExecute, Origin: "active_graph.edges[0]"},
				{Kind: UseLinkOrLoad, Origin: "active_graph.edges[1]"},
			},
		}
		result, err = NewService().AdmitDependency(t.Context(), request)
		requireCode(t, err, CodeCompiledDependency)
		node = requireNode(t, result, "use-conflict")
		if node.Class != ClassELFETDYNAmbiguous || node.Variant != "use_conflict" {
			t.Fatalf("use conflict node = %+v", node)
		}
	})

	t.Run("C01f manager-resolved use edges", func(t *testing.T) {
		executable := makeELF64(elfTypeDyn, false, true, false)
		executableRequest := dependencyRequest("legacy", executable, ProfileCommonV1)
		executableRequest.Descriptor.ResolvedUses = map[string][]UseEdge{
			"legacy": {{Kind: UseExecute, Origin: "active_graph.edges[0]"}},
		}
		result, err := NewService().AdmitDependency(t.Context(), executableRequest)
		requireCode(t, err, CodeCompiledDependency)
		if node := requireNode(t, result, "legacy"); node.Class != ClassNativeExecutable || node.Variant != "elf.et_dyn.executable_by_use" {
			t.Fatalf("execute edge node = %+v", node)
		}

		library := makeELF64(elfTypeDyn, false, false, false)
		libraryRequest := dependencyRequest("legacy-lib", library, ProfileCommonV1)
		libraryRequest.Descriptor.ResolvedUses = map[string][]UseEdge{
			"legacy-lib": {{Kind: UseLinkOrLoad, Origin: "active_graph.edges[1]"}},
		}
		result, err = NewService().AdmitDependency(t.Context(), libraryRequest)
		requireCode(t, err, CodeCompiledDependency)
		if node := requireNode(t, result, "legacy-lib"); node.Class != ClassNativeLibraryDynamic || node.Variant != "elf.shared_object" {
			t.Fatalf("link edge node = %+v", node)
		}
	})
}

func TestCompiledVectorsC02ThroughC12(t *testing.T) {
	t.Run("C02 PE COFF and archive", func(t *testing.T) {
		archive := buildAR(t, map[string][]byte{"case.obj": makeCOFFObject()})
		fixtures := []struct {
			name  string
			bytes []byte
			class ArtifactClass
		}{
			{"case.exe", makePE(false), ClassNativeExecutable},
			{"case.dll", makePE(true), ClassNativeLibraryDynamic},
			{"case.obj", makeCOFFObject(), ClassNativeObject},
			{"case.lib", archive, ClassNativeLibraryStatic},
			{"case-import.lib", archive, ClassNativeLibraryStatic},
		}
		for _, fixture := range fixtures {
			result, err := admitDependency(t, fixture.name, fixture.bytes, ProfileCommonV1)
			requireCode(t, err, CodeCompiledDependency)
			if got := requireNode(t, result, fixture.name).Class; got != fixture.class {
				t.Fatalf("%s class = %q, want %q", fixture.name, got, fixture.class)
			}
		}
	})

	t.Run("C03 thin and fat Mach-O", func(t *testing.T) {
		fixtures := []struct {
			name  string
			bytes []byte
			class ArtifactClass
		}{
			{"thin-exec", makeMachO(2), ClassNativeExecutable},
			{"thin-object", makeMachO(1), ClassNativeObject},
			{"thin-dylib", makeMachO(6), ClassNativeLibraryDynamic},
			{"thin-bundle", makeMachO(8), ClassNativeLibraryDynamic},
			{"fat-exec", makeFatMachO(makeMachO(2)), ClassNativeExecutable},
			{"fat-object", makeFatMachO(makeMachO(1)), ClassNativeObject},
			{"fat-dylib", makeFatMachO(makeMachO(6)), ClassNativeLibraryDynamic},
			{"fat-bundle", makeFatMachO(makeMachO(8)), ClassNativeLibraryDynamic},
		}
		for _, fixture := range fixtures {
			result, err := admitDependency(t, fixture.name, fixture.bytes, ProfileCommonV1)
			requireCode(t, err, CodeCompiledDependency)
			if got := requireNode(t, result, fixture.name).Class; got != fixture.class {
				t.Fatalf("%s class = %q, want %q", fixture.name, got, fixture.class)
			}
		}

		t.Run("mixed universal slice order is deny-dominant", func(t *testing.T) {
			executable := makeMachO(2)
			dynamic := makeMachO(6)
			for _, fixture := range []struct {
				name   string
				slices [][]byte
			}{
				{name: "dynamic-first", slices: [][]byte{dynamic, executable}},
				{name: "executable-first", slices: [][]byte{executable, dynamic}},
			} {
				result, err := admitDependency(t, fixture.name, makeFatMachOSlices(fixture.slices...), ProfileCommonV1)
				requireCode(t, err, CodeCompiledDependency)
				node := requireNode(t, result, fixture.name)
				if node.Class != ClassNativeExecutable || node.Variant != "macho.universal" {
					t.Fatalf("%s mixed universal = %s/%s", fixture.name, node.Class, node.Variant)
				}
				classes := map[string]bool{}
				for _, observation := range node.Observations {
					if observation.DetectorID != "macho-v1" {
						continue
					}
					for _, fact := range observation.Facts {
						if strings.HasSuffix(fact.Key, ".class") {
							classes[fact.Value] = true
						}
					}
				}
				if !classes[string(ClassNativeExecutable)] || !classes[string(ClassNativeLibraryDynamic)] {
					t.Fatalf("%s did not retain both slice classes: %+v", fixture.name, classes)
				}
			}
		})
	})

	t.Run("C04 native archive at depths one two and eight", func(t *testing.T) {
		archives := map[string][]byte{
			"a":    buildAR(t, map[string][]byte{"case.o": makeELF64(elfTypeRel, false, false, false)}),
			"lib":  buildAR(t, map[string][]byte{"case.obj": makeCOFFObject()}),
			"rlib": buildAR(t, map[string][]byte{"case.o": makeELF64(elfTypeRel, false, false, false)}),
		}
		for extension, archive := range archives {
			for _, depth := range []int{1, 2, 8} {
				name := fmt.Sprintf("depth-%d.%s", depth, extension)
				payload := archive
				if depth > 1 {
					name = fmt.Sprintf("depth-%d-%s.zip", depth, extension)
					payload = buildNestedZIP(t, depth-1, "case."+extension, archive)
				}
				result, err := admitDependency(t, name, payload, ProfileCommonV1)
				requireCode(t, err, CodeCompiledDependency)
				found := false
				for _, node := range result.Manifest.Nodes {
					if node.Class == ClassNativeLibraryStatic {
						found = true
					}
				}
				if !found {
					t.Fatalf("%s has no static-library node", name)
				}
			}
		}
	})

	t.Run("C05 frameworks reject as bundles", func(t *testing.T) {
		for _, bundle := range []string{"Thing.framework", "Thing.xcframework"} {
			for _, fixture := range []struct {
				outer, prefix string
			}{
				{outer: "bundle.zip"},
				{outer: "renamed.dat", prefix: "nested"},
			} {
				member := bundle + "/README.txt"
				bundlePath := fixture.outer + "!/" + bundle
				if fixture.prefix != "" {
					member = fixture.prefix + "/" + member
					bundlePath = fixture.outer + "!/" + fixture.prefix + "/" + bundle
				}
				payload := buildZIP(t, []zipFixtureEntry{{
					name: member, content: []byte("resource-only bundle\n"), method: zip.Store,
				}})
				result, err := admitDependency(t, fixture.outer, payload, ProfileSwiftPMV1)
				requireCode(t, err, CodeCompiledDependency)
				class := requireNode(t, result, bundlePath).Class
				if class != ClassAppleFramework && class != ClassAppleXCFramework {
					t.Fatalf("%s class = %q", bundlePath, class)
				}
			}
		}
	})

	t.Run("C06 Node and Python native extensions", func(t *testing.T) {
		fixtures := []struct {
			name  string
			bytes []byte
			class ArtifactClass
		}{
			{"addon.node", makeELF64(elfTypeDyn, false, false, true), ClassNodeExtension},
			{"addon-pe.node", makePE(true), ClassNodeExtension},
			{"addon-macho.node", makeMachO(8), ClassNodeExtension},
			{"module.cpython-313-x86_64-linux-gnu.so", makeELF64(elfTypeDyn, false, false, true), ClassPythonExtension},
			{"module.pyd", makePE(true), ClassPythonExtension},
		}
		for _, fixture := range fixtures {
			result, err := admitDependency(t, fixture.name, fixture.bytes, ProfileCommonV1)
			requireCode(t, err, CodeCompiledDependency)
			if got := requireNode(t, result, fixture.name).Class; got != fixture.class {
				t.Fatalf("%s class = %q, want %q", fixture.name, got, fixture.class)
			}
		}
	})

	t.Run("C07 class in nested JAR and DEX in AAR", func(t *testing.T) {
		classBytes := makeJVMClass()
		result, err := admitDependency(t, "Fixture.class", classBytes, ProfileCommonV1)
		requireCode(t, err, CodeCompiledDependency)
		if got := requireNode(t, result, "Fixture.class").Class; got != ClassJVMBytecode {
			t.Fatalf("class = %q", got)
		}
		jar := buildZIP(t, []zipFixtureEntry{{name: "Fixture.class", content: classBytes, method: zip.Store}})
		outer := buildZIP(t, []zipFixtureEntry{{name: "lib/fixture.jar", content: jar, method: zip.Store}})
		result, err = admitDependency(t, "outer.zip", outer, ProfileCommonV1)
		requireCode(t, err, CodeCompiledDependency)
		if got := requireNode(t, result, "outer.zip!/lib/fixture.jar!/Fixture.class").Class; got != ClassJVMBytecode {
			t.Fatalf("nested class = %q", got)
		}
		aar := buildZIP(t, []zipFixtureEntry{{name: "classes.dex", content: makeDEX(), method: zip.Store}})
		result, err = admitDependency(t, "fixture.aar", aar, ProfileCommonV1)
		requireCode(t, err, CodeCompiledDependency)
		if got := requireNode(t, result, "fixture.aar!/classes.dex").Class; got != ClassJVMBytecode {
			t.Fatalf("DEX class = %q", got)
		}
	})

	t.Run("C08 WebAssembly direct renamed and nested", func(t *testing.T) {
		for _, name := range []string{"module.wasm", "module.dat"} {
			result, err := admitDependency(t, name, makeWasm(), ProfileCommonV1)
			requireCode(t, err, CodeCompiledDependency)
			if got := requireNode(t, result, name).Class; got != ClassWebAssembly {
				t.Fatalf("%s class = %q", name, got)
			}
		}
		payload := buildZIP(t, []zipFixtureEntry{{name: "assets/module.bin", content: makeWasm(), method: zip.Store}})
		result, err := admitDependency(t, "package.zip", payload, ProfileNodeV1)
		requireCode(t, err, CodeCompiledDependency)
		if got := requireNode(t, result, "package.zip!/assets/module.bin").Class; got != ClassWebAssembly {
			t.Fatalf("nested wasm class = %q", got)
		}
		wheel := buildWheel(t, map[string][]byte{
			"fixture/__init__.py": []byte("VALUE = 1\n"),
			"fixture/module.dat":  makeWasm(),
		})
		result, err = admitDependency(t, "fixture-1.0.0-py3-none-any.whl", wheel, ProfilePythonSourceV1)
		requireCode(t, err, CodeCompiledDependency)
		if got := requireNode(t, result, "fixture-1.0.0-py3-none-any.whl!/fixture/module.dat").Class; got != ClassWebAssembly {
			t.Fatalf("wheel wasm class = %q", got)
		}
	})

	t.Run("C09 VM caches and compiler serialization", func(t *testing.T) {
		pyc := make([]byte, 16)
		copy(pyc[:4], []byte{0xa7, 0x0d, '\r', '\n'})
		fixtures := []struct {
			name  string
			bytes []byte
			class ArtifactClass
		}{
			{"module.pyc", pyc, ClassPythonBytecode},
			{"module.v8cache", []byte("V8CS\x01\x00\x00\x00payload"), ClassJavaScriptCodeCache},
			{"startup.snapshot", []byte("NODE\x01\x00\x00\x00snapshot"), ClassJavaScriptCodeCache},
			{"module.bc", makeLLVMBitcode(), ClassCompilerSerialized},
			{"module-wrapper.bc", makeLLVMBitcodeWrapper(), ClassCompilerSerialized},
			{"module.swiftmodule", []byte{0, 1, 2, 3, 4}, ClassCompilerSerialized},
			{"module.swiftdoc", []byte{0, 2, 3, 4, 5}, ClassCompilerSerialized},
			{"module.pcm", []byte{0, 4, 3, 2, 1}, ClassCompilerSerialized},
			{"module.pch", []byte{0, 9, 8, 7}, ClassCompilerSerialized},
			{"module.gch", []byte{0, 8, 7, 6}, ClassCompilerSerialized},
			{"module.ifc", []byte{0, 7, 6, 5}, ClassCompilerSerialized},
			{"module.rmeta", []byte{0, 6, 5, 4}, ClassCompilerSerialized},
		}
		for _, fixture := range fixtures {
			result, err := admitDependency(t, fixture.name, fixture.bytes, ProfileCommonV1)
			requireCode(t, err, CodeCompiledDependency)
			if got := requireNode(t, result, fixture.name).Class; got != fixture.class {
				t.Fatalf("%s class = %q, want %q", fixture.name, got, fixture.class)
			}
		}
		for _, suffix := range []string{".swiftmodule", ".swiftdoc", ".pcm", ".pch", ".gch", ".ifc", ".rmeta"} {
			name := "README" + suffix
			result, err := admitDependency(t, name, []byte("printable text cannot satisfy a serialized compiler claim\n"), ProfileCommonV1)
			requireCode(t, err, CodeTypeAmbiguous)
			if got := requireNode(t, result, name).Class; got != ClassOpaqueUnknown {
				t.Fatalf("%s printable conflict class = %q, want %q", name, got, ClassOpaqueUnknown)
			}
		}
	})

	t.Run("C10 and C11 mode name and identical tool bytes do not change dependency role", func(t *testing.T) {
		payload := makeELF64(elfTypeExec, false, false, false)
		for _, name := range []string{"compiled.txt", "vendor/toolchain/bin/clang"} {
			result, err := admitDependency(t, name, payload, ProfileCommonV1)
			requireCode(t, err, CodeCompiledDependency)
			if got := requireNode(t, result, name).Class; got != ClassNativeExecutable {
				t.Fatalf("%s class = %q", name, got)
			}
		}
		toolchain, toolchainErr := NewService().AdmitToolchain(t.Context(), ToolchainRequest{
			Descriptor: fixtureDescriptor(payload, ProfileCommonV1),
			Payload: Payload{
				Path: "bin/clang", Size: int64(len(payload)), Reader: bytes.NewReader(payload),
			},
			Authorization: validToolchainAuthorization(t, "bin/clang", payload),
		})
		if toolchainErr != nil {
			t.Fatal(toolchainErr)
		}
		dependency, dependencyErr := admitDependency(t, "vendor/toolchain/bin/clang", payload, ProfileCommonV1)
		requireCode(t, dependencyErr, CodeCompiledDependency)
		if requireNode(t, toolchain, "bin/clang").Class != requireNode(t, dependency, "vendor/toolchain/bin/clang").Class {
			t.Fatal("identical toolchain and dependency bytes received different content classes")
		}
		if toolchain.Manifest.Decision != DecisionAllowToolchain || dependency.Manifest.Decision != DecisionReject {
			t.Fatal("causal role did not control the decision for identical bytes")
		}
	})

	t.Run("C12 adapters share leaf classification", func(t *testing.T) {
		payload := makeWasm()
		var want ManifestNode
		for index, adapter := range []string{"go-v1", "rust-v1", "node-v1", "swiftpm-v1", "python-reference-v1"} {
			request := dependencyRequest("renamed.dat", payload, ProfileCommonV1)
			request.Descriptor.AdapterID = adapter
			result, err := NewService().AdmitDependency(t.Context(), request)
			requireCode(t, err, CodeCompiledDependency)
			node := requireNode(t, result, "renamed.dat")
			if index == 0 {
				want = node
				continue
			}
			if node.Class != want.Class || node.SHA256 != want.SHA256 || node.Size != want.Size || node.Decision != want.Decision {
				t.Fatalf("adapter %s leaf differs: got=%+v want=%+v", adapter, node, want)
			}
		}
	})
}

func TestCompiledDetectorErrorsNeverAdmitOpaqueBytes(t *testing.T) {
	fixtures := map[string][]byte{
		"broken-elf.bin":     append([]byte{0x7f, 'E', 'L', 'F'}, bytes.Repeat([]byte{0}, 12)...),
		"broken-pe.txt":      []byte("MZ this is not a structurally valid PE image\n"),
		"broken-bitcode.bin": {'B', 'C', 0xc0, 0xde, 1, 2, 3, 4},
	}
	for name, payload := range fixtures {
		result, err := admitDependency(t, name, payload, ProfileCommonV1)
		requireCode(t, err, CodeOpaqueDependency)
		requireDecision(t, result, DecisionReject)
	}
	result, err := admitDependency(t, "broken.exe", []byte("MZ truncated"), ProfileCommonV1)
	requireCode(t, err, CodeTypeAmbiguous)
	requireDecision(t, result, DecisionReject)

	t.Run("sound compiled match remains primary but cannot hide a detector error", func(t *testing.T) {
		payload := append([]byte{0x7f, 'E', 'L', 'F'}, bytes.Repeat([]byte{0}, 12)...)
		result, err := admitDependency(t, "module.v8cache", payload, ProfileCommonV1)
		requireCode(t, err, CodeCompiledDependency)
		node := requireNode(t, result, "module.v8cache")
		if node.Class != ClassJavaScriptCodeCache || node.InspectionComplete {
			t.Fatalf("compiled match with detector error = %+v", node)
		}
		hasError := false
		for _, observation := range node.Observations {
			hasError = hasError || observation.Result == "ERROR"
		}
		if !hasError {
			t.Fatal("compiled match lost the competing detector error")
		}
		assertNoAuthorization(t, result)
	})
}
