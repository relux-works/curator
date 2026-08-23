// Package conformance publishes deterministic, reusable artifact-policy byte
// vectors. Adapter wrappers consume the same payloads and expected
// artifact-manifest-v1 tuple instead of recreating format fixtures locally.
package conformance

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha1" // #nosec G505 -- constructs a format-correct DEX fixture.
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"hash/adler32"
	"io/fs"
	"sort"
	"strings"
	"time"
)

// Scenario tells a conformance harness which public admission API and causal
// evidence state to exercise.
type Scenario string

const (
	// Dependency exercises ordinary immutable dependency admission.
	Dependency Scenario = "dependency"
	// ToolchainAllowed exercises a manager-issued trusted-toolchain checkpoint.
	ToolchainAllowed Scenario = "toolchain_allowed"
	// ToolchainMissing exercises the fail-closed absent-checkpoint path.
	ToolchainMissing Scenario = "toolchain_missing"
	// ToolchainDrift exercises a time-of-use fingerprint mismatch.
	ToolchainDrift Scenario = "toolchain_drift"
	// ToolchainLinkUnsafe exercises incomplete contained-link validation.
	ToolchainLinkUnsafe Scenario = "toolchain_link_unsafe"
	// ToolchainSpecialUnsafe exercises an unvalidated special toolchain node.
	ToolchainSpecialUnsafe Scenario = "toolchain_special_unsafe"
	// OutputAllowed exercises a causally receipted local output.
	OutputAllowed Scenario = "output_allowed"
	// OutputMissing exercises the fail-closed absent-receipt path.
	OutputMissing Scenario = "output_missing"
	// OutputDrift exercises exact protected-output expectation drift.
	OutputDrift Scenario = "output_drift"
	// OutputPreexisting exercises a copied pre-existing output.
	OutputPreexisting Scenario = "output_preexisting"
	// OutputHardlink exercises a hard-linked pre-existing output.
	OutputHardlink Scenario = "output_hardlink"
	// OutputInputDrift exercises a complete receipt-input mismatch.
	OutputInputDrift Scenario = "output_input_drift"
	// VerifiedUnavailable exercises the unavailable verified-binary-v1 seam.
	VerifiedUnavailable Scenario = "verified_unavailable"
	// OriginMissing exercises immutable dependency-origin refusal.
	OriginMissing Scenario = "origin_missing"
	// ReaderFailure exercises incomplete capture refusal.
	ReaderFailure Scenario = "reader_failure"
	// Cancellation exercises an already-cancelled pre-execution inspection.
	Cancellation Scenario = "cancellation"
)

// Use declares a manager-resolved ELF use edge without importing the policy
// package and creating an import cycle.
type Use struct {
	Kind   string
	Origin string
}

// Expected is the complete stable tuple shared by adapter conformance tests.
type Expected struct {
	Path             string
	Class            string
	NodeDecision     string
	ManifestDecision string
	PrimaryCode      string
	ManifestDigest   string
	Authorization    bool
}

// Case is one immutable vector. Payload is freshly allocated by Cases.
type Case struct {
	ID                   string
	Variant              string
	Scenario             Scenario
	Path                 string
	Profile              string
	AdapterID            string
	Payload              []byte
	AuthorizationPayload []byte
	AuthorizationPath    string
	PayloadSizeOverride  int64
	Uses                 []Use
	Expected             Expected
}

// Key is the stable corpus lookup key.
func (fixture Case) Key() string {
	if fixture.Variant == "" {
		return fixture.ID
	}
	return fixture.ID + "/" + fixture.Variant
}

// Cases returns reusable immutable bytes and exact artifact-manifest-v1
// expectations for every accepted branch of A01-A08, C01-C12, F01-F14,
// T01-T05, and the currently unavailable V01 capability.
func Cases() []Case {
	const goSourceText = "package main\nfunc main() {}\n"
	const goSourceSize = uint32(len(goSourceText))
	goSource := []byte(goSourceText)
	elfExec := ELF64(2, false, false, false)
	elfObject := ELF64(1, false, false, false)
	elfShared := ELF64(3, false, false, true)
	gnuDynamicPIE := GNUDynamicPIE()
	gnuStaticPIE := GNUStaticPIE()
	gnuSharedObject := GNUSharedObject()
	cases := []Case{
		fixture("A01", "", Dependency, "source.zip", "go-source-v1", ZIP([]ZIPEntry{
			{Name: "go.mod", Data: []byte("module example.test/a\n\ngo 1.25\n")},
			{Name: "go.sum", Data: []byte("example.test/b v1.0.0 h1:fixture\n")},
			{Name: "LICENSE", Data: []byte("Permission is hereby granted.\n")},
			{Name: "main.go", Data: goSource},
		}), "source.zip!/main.go", "source.authored_text", "ADMIT_INPUT", "ADMIT_INPUT", "", true),
		fixture("A02", "", Dependency, "bin/tool", "common-source-v1",
			[]byte("#!/bin/sh\nset -eu\nprintf '%s\\n' ok\n"),
			"bin/tool", "source.authored_text", "ADMIT_INPUT", "ADMIT_INPUT", "", true),
		fixture("A03", "", Dependency, "package.zip", "node-source-v1", ZIP([]ZIPEntry{
			{Name: "dist/app.min.js", Data: []byte("(()=>{console.log('ok')})();\n")},
			{Name: "dist/app.min.js.map", Data: []byte(`{"version":3,"sources":["app.ts"],"mappings":""}`)},
			{Name: "package.json", Data: []byte(`{"name":"fixture","version":"1.0.0"}`)},
		}), "package.zip!/dist/app.min.js", "source.generated_text", "ADMIT_INPUT", "ADMIT_INPUT", "", true),
		fixture("A04", "", Dependency, "swift-source.zip", "swiftpm-source-v1", ZIP([]ZIPEntry{
			{Name: "Sources/API.swift", Data: []byte("public struct API { public init() {} }\n")},
			{Name: "Sources/API.swiftinterface", Data: []byte("// swift-interface-format-version: 1.0\npublic struct API {}\n")},
		}), "swift-source.zip!/Sources/API.swiftinterface", "source.generated_text", "ADMIT_INPUT", "ADMIT_INPUT", "", true),
		fixture("A05", "", Dependency, "source.zip", "rust-source-v1",
			ZIP([]ZIPEntry{{Name: "bundle.tar.gz", Data: GZIP(Tar([]TarEntry{{Name: "src/main.rs", Data: []byte("pub fn answer() -> u8 { 42 }\n")}}), "bundle.tar")}}),
			"source.zip!/bundle.tar.gz!/bundle.tar!/src/main.rs", "source.authored_text", "ADMIT_INPUT", "ADMIT_INPUT", "", true),
		fixture("A06", "", Dependency, "fixture-1.0.0-py3-none-any.whl", "python-source-container-v1", Wheel(map[string][]byte{
			"fixture/__init__.py":              []byte("VALUE = 42\n"),
			"fixture-1.0.0.dist-info/METADATA": []byte("Metadata-Version: 2.1\nName: fixture\nVersion: 1.0.0\n"),
			"fixture-1.0.0.dist-info/WHEEL":    []byte("Wheel-Version: 1.0\nRoot-Is-Purelib: true\n"),
		}), "fixture-1.0.0-py3-none-any.whl!/fixture/__init__.py", "source.authored_text", "ADMIT_INPUT", "ADMIT_INPUT", "", true),
		fixture("A07", "", ToolchainAllowed, "bin/clang", "common-source-v1", elfExec,
			"bin/clang", "native.executable", "ALLOW_TOOLCHAIN", "ALLOW_TOOLCHAIN", "", true),
		fixture("A08", "", OutputAllowed, "obj/main.o", "common-source-v1", elfObject,
			"obj/main.o", "native.object", "ALLOW_OUTPUT", "ALLOW_OUTPUT", "", true),

		fixture("C01", "", Dependency, "program.txt", "common-source-v1", elfExec,
			"program.txt", "native.executable", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C01a", "", Dependency, "program.dat", "common-source-v1", gnuDynamicPIE,
			"program.dat", "native.executable", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C01b", "", Dependency, "static-pie.so", "common-source-v1", gnuStaticPIE,
			"static-pie.so", "native.executable", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C01c", "", Dependency, "renamed.dat", "common-source-v1", gnuSharedObject,
			"renamed.dat", "native.library.dynamic", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C01d", "", Dependency, "no-evidence", "common-source-v1", ELF64(3, false, false, false),
			"no-evidence", "native.elf.et_dyn_ambiguous", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C01e", "", Dependency, "conflict", "common-source-v1", ELF64(3, false, true, true),
			"conflict", "native.elf.et_dyn_ambiguous", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		withUses(fixture("C01f", "", Dependency, "legacy", "common-source-v1", ELF64(3, false, true, false),
			"legacy", "native.executable", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
			Use{Kind: "execute", Origin: "active_graph.edges[0]"}),
		fixture("C02", "", Dependency, "case.exe", "common-source-v1", PE(false),
			"case.exe", "native.executable", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C03", "", Dependency, "fat-exec", "common-source-v1", FatMachO(MachO(2)),
			"fat-exec", "native.executable", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C04", "", Dependency, "depth-2.zip", "common-source-v1", NestedZIP(1, "case.a", AR(map[string][]byte{"case.o": elfObject})),
			"depth-2.zip!/case.a", "native.library.static", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C05", "", Dependency, "bundle.zip", "swiftpm-source-v1", ZIP([]ZIPEntry{{Name: "Thing.framework/README.txt", Data: []byte("resource only\n")}}),
			"bundle.zip!/Thing.framework", "apple.framework", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C06", "", Dependency, "addon.node", "common-source-v1", elfShared,
			"addon.node", "native.extension.node", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C07", "", Dependency, "outer.zip", "common-source-v1", ZIP([]ZIPEntry{{Name: "lib/fixture.jar", Data: ZIP([]ZIPEntry{{Name: "Fixture.class", Data: JVMClass()}})}}),
			"outer.zip!/lib/fixture.jar!/Fixture.class", "vm.jvm_bytecode", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C08", "", Dependency, "package.zip", "node-source-v1", ZIP([]ZIPEntry{{Name: "assets/module.bin", Data: Wasm()}}),
			"package.zip!/assets/module.bin", "ir.webassembly", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C09", "", Dependency, "module-wrapper.bc", "common-source-v1", LLVMBitcodeWrapper(),
			"module-wrapper.bc", "ir.compiler_serialized", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C10", "", Dependency, "compiled.txt", "common-source-v1", elfExec,
			"compiled.txt", "native.executable", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C11", "", Dependency, "vendor/toolchain/bin/clang", "common-source-v1", elfExec,
			"vendor/toolchain/bin/clang", "native.executable", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		withAdapter(fixture("C12", "", Dependency, "renamed.dat", "common-source-v1", Wasm(),
			"renamed.dat", "ir.webassembly", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false), "go-v1"),

		fixture("F01", "", Dependency, "unsafe.zip", "common-source-v1", ZIP([]ZIPEntry{{Name: "../escape.go", Data: goSource}}),
			"unsafe.zip", "container.archive", "REJECT", "REJECT", "artifact_archive_unsafe_path", false),
		fixture("F02", "", Dependency, "duplicate.zip", "go-source-v1", ZIP([]ZIPEntry{{Name: "main.go", Data: goSource}, {Name: "main.go", Data: goSource}}),
			"duplicate.zip", "container.archive", "REJECT", "REJECT", "artifact_archive_unsafe_path", false),
		fixture("F03", "", Dependency, "sparse.tar", "common-source-v1", Tar([]TarEntry{{Name: "sparse.bin", Data: []byte("extent"), PAX: map[string]string{"CURATOR.sparse.external_extent": "0,6"}}}),
			"sparse.tar!/sparse.bin", "fs.special", "REJECT", "REJECT", "artifact_archive_unsafe_entry", false),
		fixture("F04", "", Dependency, "encrypted.zip", "common-source-v1", PatchZIPFlags(ZIP([]ZIPEntry{{Name: "main.go", Data: goSource}}), 1),
			"encrypted.zip!/main.go", "opaque.unknown", "REJECT", "REJECT", "artifact_archive_encrypted", false),
		fixture("F05", "", Dependency, "unsupported.zip", "common-source-v1", PatchZIPMethod(ZIP([]ZIPEntry{{Name: "main.go", Data: goSource}}), 99),
			"unsupported.zip!/main.go", "opaque.unknown", "REJECT", "REJECT", "artifact_archive_unsupported", false),
		fixture("F06", "", Dependency, "trailing.zip", "common-source-v1", append(ZIP([]ZIPEntry{{Name: "main.go", Data: goSource}}), []byte("trailing")...),
			"trailing.zip", "container.archive", "REJECT", "REJECT", "artifact_archive_invalid", false),
		fixture("F07", "", Dependency, "deep.zip", "go-source-v1", NestedZIP(9, "main.go", goSource),
			"deep.zip!/layer-07.zip!/layer-06.zip!/layer-05.zip!/layer-04.zip!/layer-03.zip!/layer-02.zip!/layer-01.zip!/layer-00.zip", "container.archive", "REJECT", "REJECT", "artifact_inspection_limit_exceeded", false),
		fixture("F08", "", Dependency, "ratio.zip", "common-source-v1", PatchZIPDeclaredSizes(ZIP([]ZIPEntry{{Name: "ratio.bin"}}), []uint32{201}, []uint32{1}),
			"ratio.zip", "container.archive", "REJECT", "REJECT", "artifact_inspection_limit_exceeded", false),
		fixture("F09", "", ReaderFailure, "main.go", "go-source-v1", goSource,
			"main.go", "opaque.unknown", "REJECT", "REJECT", "artifact_inspection_unavailable", false),
		fixture("F10", "", Dependency, "library.so", "common-source-v1", []byte("ordinary text\n"),
			"library.so", "opaque.unknown", "REJECT", "REJECT", "artifact_type_ambiguous", false),
		fixture("F11", "", Dependency, "looks-safe.txt", "common-source-v1", elfExec,
			"looks-safe.txt", "native.executable", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("F12", "", Dependency, "unknown.bin", "common-source-v1", []byte{0x12, 0x34, 0x00, 0xff, 0x55},
			"unknown.bin", "opaque.unknown", "REJECT", "REJECT", "artifact_opaque_dependency_forbidden", false),
		fixture("F13", "", OriginMissing, "main.go", "go-source-v1", goSource,
			"main.go", "source.authored_text", "REJECT", "REJECT", "artifact_origin_unverified", false),
		fixture("F14", "z-first", Dependency, "order.zip", "common-source-v1", ZIP([]ZIPEntry{{Name: "z.txt", Data: []byte("z\n")}, {Name: "a.txt", Data: []byte("a\n")}}),
			"order.zip!/a.txt", "text.metadata", "ADMIT_INPUT", "ADMIT_INPUT", "", true),
		fixture("F14", "a-first", Dependency, "order.zip", "common-source-v1", ZIP([]ZIPEntry{{Name: "a.txt", Data: []byte("a\n")}, {Name: "z.txt", Data: []byte("z\n")}}),
			"order.zip!/a.txt", "text.metadata", "ADMIT_INPUT", "ADMIT_INPUT", "", true),

		fixture("T01", "", Dependency, "package.zip", "rust-source-v1", ZIP([]ZIPEntry{{Name: "vendor/toolchain/bin/rustc", Data: elfExec}}),
			"package.zip!/vendor/toolchain/bin/rustc", "native.executable", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("T02", "", ToolchainMissing, "lib/libfoo.so", "common-source-v1", elfShared,
			"lib/libfoo.so", "native.library.dynamic", "REJECT", "REJECT", "artifact_toolchain_untrusted", false),
		fixture("T03", "", ToolchainDrift, "bin/tool", "common-source-v1", elfExec,
			"bin/tool", "native.executable", "REJECT", "REJECT", "artifact_toolchain_identity_changed", false),
		fixture("T04", "", OutputMissing, "obj/main.o", "common-source-v1", elfObject,
			"obj/main.o", "native.object", "REJECT", "REJECT", "artifact_local_output_unreceipted", false),
		withAuthorizationPayload(fixture("T05", "", OutputDrift, "obj/main.o", "common-source-v1", append(append([]byte(nil), elfObject...), 0),
			"obj/main.o", "native.object", "REJECT", "REJECT", "artifact_local_output_drift", false), elfObject),
		fixture("V01", "", VerifiedUnavailable, "signed.exe", "common-source-v1", PE(false),
			"signed.exe", "native.executable", "REJECT", "REJECT", "artifact_binary_admission_unavailable", false),
	}

	// Compound accepted vectors publish every named byte/evidence branch. The
	// unqualified case above is one branch and these stable variants cover the
	// remainder; consumers never need package-private fixture builders.
	cases = append(cases,
		fixture("A08", "static-library", OutputAllowed, "lib/libmain.a", "common-source-v1", AR(map[string][]byte{"main.o": elfObject}),
			"lib/libmain.a", "native.library.static", "ALLOW_OUTPUT", "ALLOW_OUTPUT", "", true),
		fixture("A08", "node-addon", OutputAllowed, "lib/addon.node", "common-source-v1", elfShared,
			"lib/addon.node", "native.extension.node", "ALLOW_OUTPUT", "ALLOW_OUTPUT", "", true),
		fixture("A08", "executable", OutputAllowed, "bin/tool", "common-source-v1", elfExec,
			"bin/tool", "native.executable", "ALLOW_OUTPUT", "ALLOW_OUTPUT", "", true),

		fixture("C01", "exec-correct-suffix", Dependency, "program.elf", "common-source-v1", elfExec,
			"program.elf", "native.executable", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C01", "exec-no-suffix", Dependency, "program", "common-source-v1", elfExec,
			"program", "native.executable", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C01", "object-correct-suffix", Dependency, "case.o", "common-source-v1", elfObject,
			"case.o", "native.object", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C01", "object-wrong-suffix", Dependency, "case.txt", "common-source-v1", elfObject,
			"case.txt", "native.object", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C01", "object-no-suffix", Dependency, "case-object", "common-source-v1", elfObject,
			"case-object", "native.object", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C01a", "executable-looking", Dependency, "bin/pie", "common-source-v1", gnuDynamicPIE,
			"bin/pie", "native.executable", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C01a", "shared-object-suffix", Dependency, "program.so", "common-source-v1", gnuDynamicPIE,
			"program.so", "native.executable", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C01a", "no-suffix", Dependency, "dynamic-pie", "common-source-v1", gnuDynamicPIE,
			"dynamic-pie", "native.executable", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C01b", "no-suffix", Dependency, "static-pie", "common-source-v1", gnuStaticPIE,
			"static-pie", "native.executable", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C01c", "shared-object-suffix", Dependency, "libcase.so", "common-source-v1", gnuSharedObject,
			"libcase.so", "native.library.dynamic", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C01c", "no-suffix", Dependency, "libcase", "common-source-v1", gnuSharedObject,
			"libcase", "native.library.dynamic", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C01d", "interp-without-use", Dependency, "interp-no-use", "common-source-v1", ELF64(3, false, true, false),
			"interp-no-use", "native.elf.et_dyn_ambiguous", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		withUses(fixture("C01e", "use-conflict", Dependency, "use-conflict", "common-source-v1", ELF64(3, false, true, false),
			"use-conflict", "native.elf.et_dyn_ambiguous", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
			Use{Kind: "execute", Origin: "active_graph.edges[0]"}, Use{Kind: "link_or_load", Origin: "active_graph.edges[1]"}),
		withUses(fixture("C01f", "link-or-load", Dependency, "legacy-lib", "common-source-v1", ELF64(3, false, false, false),
			"legacy-lib", "native.library.dynamic", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
			Use{Kind: "link_or_load", Origin: "active_graph.edges[1]"}),

		fixture("C02", "dll", Dependency, "case.dll", "common-source-v1", PE(true),
			"case.dll", "native.library.dynamic", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C02", "coff-object", Dependency, "case.obj", "common-source-v1", COFFObject(),
			"case.obj", "native.object", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C02", "archive-library", Dependency, "case.lib", "common-source-v1", AR(map[string][]byte{"case.obj": COFFObject()}),
			"case.lib", "native.library.static", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C02", "import-library", Dependency, "case-import.lib", "common-source-v1", AR(map[string][]byte{"case.obj": COFFObject()}),
			"case-import.lib", "native.library.static", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),

		fixture("C03", "thin-executable", Dependency, "thin-exec", "common-source-v1", MachO(2),
			"thin-exec", "native.executable", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C03", "thin-object", Dependency, "thin-object", "common-source-v1", MachO(1),
			"thin-object", "native.object", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C03", "thin-dylib", Dependency, "thin-dylib", "common-source-v1", MachO(6),
			"thin-dylib", "native.library.dynamic", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C03", "thin-bundle", Dependency, "thin-bundle", "common-source-v1", MachO(8),
			"thin-bundle", "native.library.dynamic", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C03", "fat-object", Dependency, "fat-object", "common-source-v1", FatMachO(MachO(1)),
			"fat-object", "native.object", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C03", "fat-dylib", Dependency, "fat-dylib", "common-source-v1", FatMachO(MachO(6)),
			"fat-dylib", "native.library.dynamic", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C03", "fat-bundle", Dependency, "fat-bundle", "common-source-v1", FatMachO(MachO(8)),
			"fat-bundle", "native.library.dynamic", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C03", "mixed-dynamic-first", Dependency, "mixed-dynamic-first", "common-source-v1", FatMachOSlices(MachO(6), MachO(2)),
			"mixed-dynamic-first", "native.executable", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C03", "mixed-executable-first", Dependency, "mixed-executable-first", "common-source-v1", FatMachOSlices(MachO(2), MachO(6)),
			"mixed-executable-first", "native.executable", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),

		fixture("C05", "xcframework-resource-only", Dependency, "bundle.zip", "swiftpm-source-v1", ZIP([]ZIPEntry{{Name: "Thing.xcframework/Info.plist", Data: []byte("resource only\n")}}),
			"bundle.zip!/Thing.xcframework", "apple.xcframework", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C05", "framework-renamed-nested", Dependency, "renamed.dat", "swiftpm-source-v1", ZIP([]ZIPEntry{{Name: "nested/Thing.framework/README.txt", Data: []byte("resource only\n")}}),
			"renamed.dat!/nested/Thing.framework", "apple.framework", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C05", "xcframework-renamed-nested", Dependency, "renamed.dat", "swiftpm-source-v1", ZIP([]ZIPEntry{{Name: "nested/Thing.xcframework/Info.plist", Data: []byte("resource only\n")}}),
			"renamed.dat!/nested/Thing.xcframework", "apple.xcframework", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C06", "pe-node", Dependency, "addon-pe.node", "common-source-v1", PE(true),
			"addon-pe.node", "native.extension.node", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C06", "macho-node", Dependency, "addon-macho.node", "common-source-v1", MachO(8),
			"addon-macho.node", "native.extension.node", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C06", "python-so", Dependency, "module.cpython-313-x86_64-linux-gnu.so", "common-source-v1", elfShared,
			"module.cpython-313-x86_64-linux-gnu.so", "native.extension.python", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C06", "python-pyd", Dependency, "module.pyd", "common-source-v1", PE(true),
			"module.pyd", "native.extension.python", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C07", "direct-class", Dependency, "Fixture.class", "common-source-v1", JVMClass(),
			"Fixture.class", "vm.jvm_bytecode", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C07", "dex-in-aar", Dependency, "fixture.aar", "common-source-v1", ZIP([]ZIPEntry{{Name: "classes.dex", Data: DEX()}}),
			"fixture.aar!/classes.dex", "vm.jvm_bytecode", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C08", "direct", Dependency, "module.wasm", "common-source-v1", Wasm(),
			"module.wasm", "ir.webassembly", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C08", "renamed", Dependency, "module.dat", "common-source-v1", Wasm(),
			"module.dat", "ir.webassembly", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C08", "wheel-nested", Dependency, "fixture-1.0.0-py3-none-any.whl", "python-source-container-v1", Wheel(map[string][]byte{
			"fixture/__init__.py": []byte("VALUE = 1\n"), "fixture/module.dat": Wasm(),
		}), "fixture-1.0.0-py3-none-any.whl!/fixture/module.dat", "ir.webassembly", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
	)

	pyc := make([]byte, 16)
	copy(pyc[:4], []byte{0xa7, 0x0d, '\r', '\n'})
	serialized := []struct {
		variant, path, class string
		payload              []byte
	}{
		{"python-bytecode", "module.pyc", "vm.python_bytecode", pyc},
		{"v8-cache", "module.v8cache", "vm.javascript_code_cache", []byte("V8CS\x01\x00\x00\x00payload")},
		{"node-snapshot", "startup.snapshot", "vm.javascript_code_cache", []byte("NODE\x01\x00\x00\x00snapshot")},
		{"llvm-bitcode", "module.bc", "ir.compiler_serialized", LLVMBitcode()},
		{"swiftmodule", "module.swiftmodule", "ir.compiler_serialized", []byte{0, 1, 2, 3, 4}},
		{"swiftdoc", "module.swiftdoc", "ir.compiler_serialized", []byte{0, 2, 3, 4, 5}},
		{"pcm", "module.pcm", "ir.compiler_serialized", []byte{0, 4, 3, 2, 1}},
		{"pch", "module.pch", "ir.compiler_serialized", []byte{0, 9, 8, 7}},
		{"gch", "module.gch", "ir.compiler_serialized", []byte{0, 8, 7, 6}},
		{"ifc", "module.ifc", "ir.compiler_serialized", []byte{0, 7, 6, 5}},
		{"rmeta", "module.rmeta", "ir.compiler_serialized", []byte{0, 6, 5, 4}},
	}
	for _, item := range serialized {
		cases = append(cases, fixture("C09", item.variant, Dependency, item.path, "common-source-v1", item.payload,
			item.path, item.class, "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false))
	}
	for _, suffix := range []string{"swiftmodule", "swiftdoc", "pcm", "pch", "gch", "ifc", "rmeta"} {
		pathValue := "README." + suffix
		cases = append(cases, fixture("C09", "printable-"+suffix, Dependency, pathValue, "common-source-v1",
			[]byte("printable text cannot satisfy a serialized compiler claim\n"), pathValue,
			"opaque.unknown", "REJECT", "REJECT", "artifact_type_ambiguous", false))
	}
	cases = append(cases,
		fixture("C10", "mode-cleared", Dependency, "compiled-no-mode", "common-source-v1", elfExec,
			"compiled-no-mode", "native.executable", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
	)
	for _, adapter := range []string{"rust-v1", "node-v1", "swiftpm-v1", "python-reference-v1"} {
		cases = append(cases, withAdapter(fixture("C12", adapter, Dependency, "renamed.dat", "common-source-v1", Wasm(),
			"renamed.dat", "ir.webassembly", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false), adapter))
	}

	archives := map[string][]byte{
		"a":    AR(map[string][]byte{"case.o": elfObject}),
		"lib":  AR(map[string][]byte{"case.obj": COFFObject()}),
		"rlib": AR(map[string][]byte{"case.o": elfObject}),
	}
	for _, extension := range []string{"a", "lib", "rlib"} {
		for _, depth := range []int{1, 2, 8} {
			if extension == "a" && depth == 2 {
				continue // the unqualified C04 case is this branch.
			}
			root := fmt.Sprintf("depth-%d-%s", depth, extension)
			payload := archives[extension]
			expectedPath := root
			if depth > 1 {
				root += ".zip"
				leafName := "case." + extension
				payload = NestedZIP(depth-1, leafName, payload)
				expectedPath = NestedPath(root, depth-1, leafName)
			}
			cases = append(cases, fixture("C04", fmt.Sprintf("%s-depth-%d", extension, depth), Dependency,
				root, "common-source-v1", payload, expectedPath, "native.library.static",
				"REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false))
		}
	}

	unsafePaths := []struct{ variant, name string }{
		{"absolute", "/absolute.go"}, {"drive", "C:/drive.go"}, {"unc", "//server/share.go"},
		{"backslash", `back\slash.go`}, {"nul", "nul\x00byte.go"}, {"control", "control\x01.go"},
		{"overlong-path", strings.Repeat("a", 4_097)}, {"overlong-component", strings.Repeat("b", 256)},
	}
	for _, item := range unsafePaths {
		cases = append(cases, fixture("F01", item.variant, Dependency, "unsafe.zip", "go-source-v1",
			ZIP([]ZIPEntry{{Name: item.name, Data: goSource}}), "unsafe.zip", "container.archive",
			"REJECT", "REJECT", "artifact_archive_unsafe_path", false))
	}
	cases = append(cases,
		fixture("F02", "case-fold", Dependency, "collision.zip", "common-source-v1", ZIP([]ZIPEntry{
			{Name: "Readme.txt", Data: []byte("one\n")}, {Name: "README.txt", Data: []byte("two\n")},
		}), "collision.zip", "container.archive", "REJECT", "REJECT", "artifact_archive_unsafe_path", false),
		fixture("F02", "nfc", Dependency, "collision.zip", "common-source-v1", ZIP([]ZIPEntry{{Name: "e\u0301/file.go", Data: goSource}}),
			"collision.zip", "container.archive", "REJECT", "REJECT", "artifact_archive_unsafe_path", false),
		fixture("F02", "trailing-dot", Dependency, "collision.zip", "common-source-v1", ZIP([]ZIPEntry{{Name: "trailing./file.go", Data: goSource}}),
			"collision.zip", "container.archive", "REJECT", "REJECT", "artifact_archive_unsafe_path", false),
		fixture("F03", "symlink", Dependency, "link.zip", "common-source-v1", ZIP([]ZIPEntry{{Name: "link", Data: []byte("target"), Mode: fs.ModeSymlink | 0o777}}),
			"link.zip!/link", "fs.symlink_or_hardlink", "REJECT", "REJECT", "artifact_archive_unsafe_entry", false),
		fixture("F03", "hardlink", Dependency, "hardlink.tar", "common-source-v1", Tar([]TarEntry{{Name: "unsafe", Typeflag: tar.TypeLink, Linkname: "target"}}),
			"hardlink.tar!/unsafe", "fs.symlink_or_hardlink", "REJECT", "REJECT", "artifact_archive_unsafe_entry", false),
		fixture("F03", "device", Dependency, "device.tar", "common-source-v1", Tar([]TarEntry{{Name: "unsafe", Typeflag: tar.TypeChar}}),
			"device.tar!/unsafe", "fs.special", "REJECT", "REJECT", "artifact_archive_unsafe_entry", false),
		fixture("F03", "fifo", Dependency, "fifo.tar", "common-source-v1", Tar([]TarEntry{{Name: "unsafe", Typeflag: tar.TypeFifo}}),
			"fifo.tar!/unsafe", "fs.special", "REJECT", "REJECT", "artifact_archive_unsafe_entry", false),
		fixture("F03", "socket", Dependency, "socket.zip", "common-source-v1", ZIP([]ZIPEntry{{Name: "unsafe", Mode: fs.ModeSocket | 0o600}}),
			"socket.zip!/unsafe", "fs.special", "REJECT", "REJECT", "artifact_archive_unsafe_entry", false),
	)

	unsupported := []struct {
		variant, path string
		payload       []byte
	}{
		{"7z", "case.7z", []byte{0x37, 0x7a, 0xbc, 0xaf, 0x27, 0x1c, 0, 0}},
		{"xz", "case.xz", []byte{0xfd, 0x37, 0x7a, 0x58, 0x5a, 0, 0, 0}},
		{"bzip2", "case.bz2", []byte{'B', 'Z', 'h', '9', 0}},
		{"zstd", "case.zst", []byte{0x28, 0xb5, 0x2f, 0xfd, 0}},
		{"iso9660", "case.iso", ISO9660()}, {"dmg", "case.dmg", DMG()},
		{"multi-volume", "split.zip", PatchZIPDisk(ZIP([]ZIPEntry{{Name: "main.go", Data: goSource}}), 1)},
	}
	for _, item := range unsupported {
		class := "opaque.unknown"
		if item.variant == "multi-volume" {
			class = "container.archive"
		}
		cases = append(cases, fixture("F05", item.variant, Dependency, item.path, "common-source-v1", item.payload,
			item.path, class, "REJECT", "REJECT", "artifact_archive_unsupported", false))
	}
	ordinaryZIP := ZIP([]ZIPEntry{{Name: "main.go", Data: goSource}})
	cases = append(cases,
		fixture("F06", "truncated", Dependency, "truncated.zip", "go-source-v1", ordinaryZIP[:len(ordinaryZIP)-7],
			"truncated.zip", "container.archive", "REJECT", "REJECT", "artifact_archive_invalid", false),
		fixture("F06", "crc", Dependency, "crc.zip", "go-source-v1", CorruptZIPBody(ordinaryZIP),
			"crc.zip", "container.archive", "REJECT", "REJECT", "artifact_archive_invalid", false),
		fixture("F06", "size", Dependency, "size.zip", "go-source-v1", PatchZIPCentralSize(ordinaryZIP, goSourceSize+1),
			"size.zip", "container.archive", "REJECT", "REJECT", "artifact_archive_invalid", false),
		fixture("F06", "tar-trailing", Dependency, "trailing.tar", "go-source-v1", append(Tar([]TarEntry{{Name: "main.go", Data: goSource}}), []byte("not-zero-trailing-data")...),
			"trailing.tar", "container.archive", "REJECT", "REJECT", "artifact_archive_invalid", false),
	)

	limitsRaw := fixture("F08", "raw-payload", Dependency, "oversized.bin", "common-source-v1", []byte{0},
		"oversized.bin", "opaque.unknown", "REJECT", "REJECT", "artifact_inspection_limit_exceeded", false)
	limitsRaw.PayloadSizeOverride = 512<<20 + 1
	limitLeaf := PatchZIPDeclaredSizes(ZIP([]ZIPEntry{{Name: "large.bin"}}), []uint32{256<<20 + 1}, []uint32{1})
	totalEntries := make([]ZIPEntry, 9)
	totalDeclared, totalCompressed := make([]uint32, 9), make([]uint32, 9)
	for index := range totalEntries {
		totalEntries[index] = ZIPEntry{Name: fmt.Sprintf("part-%02d.bin", index)}
		totalDeclared[index] = 256 << 20
		totalCompressed[index] = 2 << 20
	}
	emptyZIP := ZIP(nil)
	containerEntries := make([]ZIPEntry, 1_024)
	for index := range containerEntries {
		containerEntries[index] = ZIPEntry{Name: fmt.Sprintf("archive-%04d.zip", index), Data: emptyZIP}
	}
	cases = append(cases, limitsRaw,
		fixture("F08", "single-leaf", Dependency, "large.zip", "common-source-v1", limitLeaf,
			"large.zip", "container.archive", "REJECT", "REJECT", "artifact_inspection_limit_exceeded", false),
		fixture("F08", "total-emitted", Dependency, "total.zip", "common-source-v1", PatchZIPDeclaredSizes(ZIP(totalEntries), totalDeclared, totalCompressed),
			"total.zip", "container.archive", "REJECT", "REJECT", "artifact_inspection_limit_exceeded", false),
		fixture("F08", "container-count", Dependency, "containers.zip", "common-source-v1", ZIP(containerEntries),
			"containers.zip", "container.archive", "REJECT", "REJECT", "artifact_inspection_limit_exceeded", false),
		fixture("F08", "entry-count", Dependency, "entries.zip", "common-source-v1", ManyZIP(100_001),
			"entries.zip", "container.archive", "REJECT", "REJECT", "artifact_inspection_limit_exceeded", false),
		fixture("F09", "cancelled", Cancellation, "main.go", "go-source-v1", goSource,
			"main.go", "opaque.unknown", "REJECT", "REJECT", "artifact_inspection_unavailable", false),
		fixture("F10", "node", Dependency, "addon.node", "common-source-v1", []byte("ordinary text\n"),
			"addon.node", "opaque.unknown", "REJECT", "REJECT", "artifact_type_ambiguous", false),
		fixture("F10", "wasm", Dependency, "module.wasm", "common-source-v1", []byte("ordinary text\n"),
			"module.wasm", "opaque.unknown", "REJECT", "REJECT", "artifact_type_ambiguous", false),
		fixture("F10", "archive-name", Dependency, "library.a", "go-source-v1", ZIP([]ZIPEntry{{Name: "main.go", Data: goSource}}),
			"library.a", "container.archive", "REJECT", "REJECT", "artifact_type_ambiguous", false),
		fixture("F12", "undeclared-text", Dependency, "undeclared.blob", "common-source-v1", []byte("valid UTF-8 without a declared grammar\n"),
			"undeclared.blob", "opaque.unknown", "REJECT", "REJECT", "artifact_opaque_dependency_forbidden", false),
	)

	for _, tool := range []string{"node", "swiftc", "clang"} {
		cases = append(cases, fixture("T01", tool, Dependency, "package.zip", "common-source-v1",
			ZIP([]ZIPEntry{{Name: "vendor/toolchain/bin/" + tool, Data: elfExec}}),
			"package.zip!/vendor/toolchain/bin/"+tool, "native.executable", "REJECT", "REJECT",
			"artifact_compiled_dependency_forbidden", false))
	}
	cases = append(cases,
		fixture("T03", "link-escape", ToolchainLinkUnsafe, "bin/tool", "common-source-v1", elfExec,
			"bin/tool", "native.executable", "REJECT", "REJECT", "artifact_toolchain_untrusted", false),
		fixture("T03", "special-node", ToolchainSpecialUnsafe, "bin/tool", "common-source-v1", elfExec,
			"bin/tool", "native.executable", "REJECT", "REJECT", "artifact_toolchain_untrusted", false),
		fixture("T04", "copied-preexisting", OutputPreexisting, "obj/main.o", "common-source-v1", elfObject,
			"obj/main.o", "native.object", "REJECT", "REJECT", "artifact_local_output_unreceipted", false),
		fixture("T04", "hard-link-preexisting", OutputHardlink, "obj/main.o", "common-source-v1", elfObject,
			"obj/main.o", "native.object", "REJECT", "REJECT", "artifact_local_output_unreceipted", false),
	)
	digestDrift := append([]byte(nil), elfObject...)
	digestDrift[len(digestDrift)-1] ^= 1
	pathDrift := withAuthorizationPayload(fixture("T05", "path", OutputDrift, "obj/copied.o", "common-source-v1", elfObject,
		"obj/copied.o", "native.object", "REJECT", "REJECT", "artifact_local_output_drift", false), elfObject)
	pathDrift.AuthorizationPath = "obj/main.o"
	cases = append(cases,
		withAuthorizationPayload(fixture("T05", "digest", OutputDrift, "obj/main.o", "common-source-v1", digestDrift,
			"obj/main.o", "native.object", "REJECT", "REJECT", "artifact_local_output_drift", false), elfObject),
		pathDrift,
		fixture("T05", "complete-input", OutputInputDrift, "obj/main.o", "common-source-v1", elfObject,
			"obj/main.o", "native.object", "REJECT", "REJECT", "artifact_local_output_drift", false),
	)

	gnuMetadataArchive := AROrdered([]AREntry{
		{Name: "/", Data: []byte{0, 0, 0, 0}},
		{Name: "//", Data: []byte("long-object-name.o/\n")},
		{Name: "/0", Data: elfObject},
	})
	bsdMetadataArchive := AROrdered([]AREntry{
		{Name: "__.SYMDEF", Data: make([]byte, 8)},
		{Name: "unit.o", Data: elfObject},
	})
	coffImportArchive := AROrdered([]AREntry{
		{Name: "/", Data: []byte{0, 0, 0, 0}},
		{Name: "/", Data: make([]byte, 8)},
		{Name: "//", Data: []byte("long-object.obj\x00")},
		{Name: "/0", Data: COFFObject()},
	})
	duplicateSource := []byte("package duplicate\n")
	duplicateZIPSourceFirst := ZIP([]ZIPEntry{{Name: "unit.go", Data: duplicateSource}, {Name: "unit.go", Data: elfExec}})
	duplicateZIPCompiledFirst := ZIP([]ZIPEntry{{Name: "unit.go", Data: elfExec}, {Name: "unit.go", Data: duplicateSource}})
	duplicateTarSourceFirst := Tar([]TarEntry{{Name: "unit.go", Data: duplicateSource}, {Name: "unit.go", Data: elfExec}})
	duplicateTarCompiledFirst := Tar([]TarEntry{{Name: "unit.go", Data: elfExec}, {Name: "unit.go", Data: duplicateSource}})
	duplicateARSourceFirst := AROrdered([]AREntry{{Name: "unit.go", Data: duplicateSource}, {Name: "unit.go", Data: elfExec}})
	duplicateARCompiledFirst := AROrdered([]AREntry{{Name: "unit.go", Data: elfExec}, {Name: "unit.go", Data: duplicateSource}})
	gzipBomb := GZIP(bytes.Repeat([]byte{0}, 1<<20), "bomb.bin")
	cases = append(cases,
		withUses(fixture("A02", "resolved-execute", Dependency, "bin/interpreted", "common-source-v1",
			[]byte("#!/bin/sh\nset -eu\nprintf ok\\n\n"), "bin/interpreted", "source.authored_text",
			"ADMIT_INPUT", "ADMIT_INPUT", "", true),
			Use{Kind: "execute", Origin: "active_graph.edges[2]"}),
		fixture("A07", "native-archive-metadata", ToolchainAllowed, "lib/toolchain.a", "common-source-v1", gnuMetadataArchive,
			"lib/toolchain.a", "native.library.static", "ALLOW_TOOLCHAIN", "ALLOW_TOOLCHAIN", "", true),
		fixture("A08", "native-archive-metadata", OutputAllowed, "lib/output.a", "common-source-v1", gnuMetadataArchive,
			"lib/output.a", "native.library.static", "ALLOW_OUTPUT", "ALLOW_OUTPUT", "", true),
		fixture("C04", "gnu-metadata", Dependency, "gnu.a", "common-source-v1", gnuMetadataArchive,
			"gnu.a", "native.library.static", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C04", "bsd-metadata", Dependency, "bsd.a", "common-source-v1", bsdMetadataArchive,
			"bsd.a", "native.library.static", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("C04", "coff-import-metadata", Dependency, "import.lib", "common-source-v1", coffImportArchive,
			"import.lib", "native.library.static", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("F02", "duplicate-zip-source-first", Dependency, "duplicates.zip", "go-source-v1", duplicateZIPSourceFirst,
			"duplicates.zip", "container.archive", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("F02", "duplicate-zip-compiled-first", Dependency, "duplicates.zip", "go-source-v1", duplicateZIPCompiledFirst,
			"duplicates.zip", "container.archive", "REJECT", "REJECT", "artifact_compiled_dependency_forbidden", false),
		fixture("F02", "duplicate-tar-source-first", Dependency, "duplicates.tar", "go-source-v1", duplicateTarSourceFirst,
			"duplicates.tar", "container.archive", "REJECT", "REJECT", "artifact_archive_unsafe_path", false),
		fixture("F02", "duplicate-tar-compiled-first", Dependency, "duplicates.tar", "go-source-v1", duplicateTarCompiledFirst,
			"duplicates.tar", "container.archive", "REJECT", "REJECT", "artifact_archive_unsafe_path", false),
		fixture("F02", "duplicate-ar-source-first", Dependency, "duplicates.a", "go-source-v1", duplicateARSourceFirst,
			"duplicates.a", "native.library.static", "REJECT", "REJECT", "artifact_archive_unsafe_path", false),
		fixture("F02", "duplicate-ar-compiled-first", Dependency, "duplicates.a", "go-source-v1", duplicateARCompiledFirst,
			"duplicates.a", "native.library.static", "REJECT", "REJECT", "artifact_archive_unsafe_path", false),
		fixture("F08", "gzip-padded-ratio", Dependency, "bomb.gz", "common-source-v1",
			append(append([]byte(nil), gzipBomb...), bytes.Repeat([]byte{0}, 2<<20)...),
			"bomb.gz", "container.compressed_stream", "REJECT", "REJECT", "artifact_inspection_limit_exceeded", false),
		fixture("F08", "gzip-concatenated-ratio", Dependency, "bomb.gz", "common-source-v1",
			append(append([]byte(nil), gzipBomb...), GZIP([]byte("second stream\n"), "second.txt")...),
			"bomb.gz", "container.compressed_stream", "REJECT", "REJECT", "artifact_inspection_limit_exceeded", false),
		withUses(fixture("F10", "resolved-link-or-load-text", Dependency, "library.go", "go-source-v1",
			[]byte("package library\n"), "library.go", "opaque.unknown", "REJECT", "REJECT", "artifact_type_ambiguous", false),
			Use{Kind: "link_or_load", Origin: "active_graph.edges[0]"}),
	)

	for index := range cases {
		cases[index].Expected.ManifestDigest = goldenDigests[cases[index].Key()]
	}
	return cases
}

func fixture(
	id, variant string,
	scenario Scenario,
	pathValue, profile string,
	payload []byte,
	expectedPath, class, nodeDecision, manifestDecision, code string,
	authorization bool,
) Case {
	return Case{
		ID: id, Variant: variant, Scenario: scenario, Path: pathValue, Profile: profile,
		AdapterID: "artifact-conformance-v1", Payload: append([]byte(nil), payload...),
		Expected: Expected{
			Path: expectedPath, Class: class, NodeDecision: nodeDecision,
			ManifestDecision: manifestDecision, PrimaryCode: code, Authorization: authorization,
		},
	}
}

func withUses(fixture Case, uses ...Use) Case {
	fixture.Uses = append([]Use(nil), uses...)
	return fixture
}

func withAdapter(fixture Case, adapter string) Case {
	fixture.AdapterID = adapter
	return fixture
}

func withAuthorizationPayload(fixture Case, payload []byte) Case {
	fixture.AuthorizationPayload = append([]byte(nil), payload...)
	return fixture
}

// ZIPEntry is one deterministic ZIP fixture member.
type ZIPEntry struct {
	Name string
	Data []byte
	Mode fs.FileMode
}

// ZIP returns deterministic stored ZIP bytes in the supplied physical order.
func ZIP(entries []ZIPEntry) []byte {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, entry := range entries {
		header := &zip.FileHeader{Name: entry.Name, Method: zip.Store}
		header.Modified = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)
		if entry.Mode != 0 {
			header.SetMode(entry.Mode)
		}
		file, err := writer.CreateHeader(header)
		must(err)
		_, err = file.Write(entry.Data)
		must(err)
	}
	must(writer.Close())
	return buffer.Bytes()
}

// TarEntry is one deterministic PAX tar fixture member.
type TarEntry struct {
	Name     string
	Data     []byte
	Typeflag byte
	Mode     int64
	Linkname string
	PAX      map[string]string
}

// AREntry is one physical native-archive member. Name is the exact ar header
// spelling for structural metadata (/, /SYM64/, //, or /<offset>) and an
// ordinary logical name otherwise.
type AREntry struct {
	Name string
	Data []byte
}

// Tar returns deterministic PAX tar bytes.
func Tar(entries []TarEntry) []byte {
	var buffer bytes.Buffer
	writer := tar.NewWriter(&buffer)
	for _, entry := range entries {
		typeflag := entry.Typeflag
		if typeflag == 0 {
			typeflag = tar.TypeReg
		}
		mode := entry.Mode
		if mode == 0 {
			mode = 0o644
		}
		size := int64(len(entry.Data))
		if typeflag != tar.TypeReg && typeflag != tar.TypeGNUSparse {
			size = 0
		}
		header := &tar.Header{
			Name: entry.Name, Mode: mode, Size: size, Typeflag: typeflag,
			Linkname: entry.Linkname, ModTime: time.Unix(0, 0).UTC(),
			Format: tar.FormatPAX, PAXRecords: entry.PAX,
		}
		must(writer.WriteHeader(header))
		if size > 0 {
			_, err := writer.Write(entry.Data)
			must(err)
		}
	}
	must(writer.Close())
	return buffer.Bytes()
}

// GZIP returns a deterministic gzip envelope.
func GZIP(payload []byte, name string) []byte {
	var buffer bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buffer, gzip.BestSpeed)
	must(err)
	writer.Name = name
	writer.ModTime = time.Unix(0, 0).UTC()
	writer.OS = 255
	_, err = writer.Write(payload)
	must(err)
	must(writer.Close())
	return buffer.Bytes()
}

// Wheel returns a source-only wheel with a valid complete RECORD.
func Wheel(files map[string][]byte) []byte {
	paths := make([]string, 0, len(files))
	for pathValue := range files {
		paths = append(paths, pathValue)
	}
	sort.Strings(paths)
	var record strings.Builder
	for _, pathValue := range paths {
		digest := sha256.Sum256(files[pathValue])
		fmt.Fprintf(&record, "%s,sha256=%s,%d\n", pathValue, base64.RawURLEncoding.EncodeToString(digest[:]), len(files[pathValue]))
	}
	recordPath := "fixture-1.0.0.dist-info/RECORD"
	fmt.Fprintf(&record, "%s,,\n", recordPath)
	entries := make([]ZIPEntry, 0, len(files)+1)
	for _, pathValue := range paths {
		entries = append(entries, ZIPEntry{Name: pathValue, Data: files[pathValue]})
	}
	entries = append(entries, ZIPEntry{Name: recordPath, Data: []byte(record.String())})
	return ZIP(entries)
}

// NestedZIP wraps a leaf in depth successive ZIP containers.
func NestedZIP(depth int, leafName string, leaf []byte) []byte {
	data := append([]byte(nil), leaf...)
	name := leafName
	for level := 0; level < depth; level++ {
		data = ZIP([]ZIPEntry{{Name: name, Data: data}})
		name = fmt.Sprintf("layer-%02d.zip", level)
	}
	return data
}

// NestedPath returns the canonical path to a leaf wrapped by NestedZIP.
func NestedPath(root string, depth int, leafName string) string {
	pathValue := root
	for level := depth - 2; level >= 0; level-- {
		pathValue += fmt.Sprintf("!/layer-%02d.zip", level)
	}
	return pathValue + "!/" + leafName
}

// ELF64 constructs a structurally valid minimal 64-bit little-endian ELF.
func ELF64(eType uint16, pie, interpreter, soname bool) []byte {
	const (
		ptInterp  = 3
		ptDynamic = 2
		dtFlags1  = 0x6ffffffb
		dtSoname  = 14
		df1PIE    = 0x08000000
	)
	programCount := 0
	if interpreter {
		programCount++
	}
	if pie || soname {
		programCount++
	}
	headerSize, programSize := 64, 56
	offset := headerSize + programCount*programSize
	interp := []byte("/lib64/ld-linux-x86-64.so.2\x00")
	interpOffset := 0
	if interpreter {
		interpOffset = offset
		offset += len(interp)
	}
	dynamicOffset, dynamicEntries := 0, 0
	if pie {
		dynamicEntries++
	}
	if soname {
		dynamicEntries++
	}
	if dynamicEntries > 0 {
		dynamicEntries++
		dynamicOffset = offset
		offset += dynamicEntries * 16
	}
	payload := make([]byte, offset)
	copy(payload[:16], []byte{0x7f, 'E', 'L', 'F', 2, 1, 1, 0, 0})
	binary.LittleEndian.PutUint16(payload[16:18], eType)
	binary.LittleEndian.PutUint16(payload[18:20], 62)
	binary.LittleEndian.PutUint32(payload[20:24], 1)
	binary.LittleEndian.PutUint64(payload[24:32], 0x401000)
	if programCount > 0 {
		binary.LittleEndian.PutUint64(payload[32:40], uint64(headerSize))
	}
	binary.LittleEndian.PutUint16(payload[52:54], uint16(headerSize))
	if programCount > 0 {
		binary.LittleEndian.PutUint16(payload[54:56], uint16(programSize))
	}
	binary.LittleEndian.PutUint16(payload[56:58], uint16(programCount))
	programOffset := headerSize
	if interpreter {
		program := payload[programOffset : programOffset+programSize]
		binary.LittleEndian.PutUint32(program[0:4], ptInterp)
		binary.LittleEndian.PutUint32(program[4:8], 4)
		binary.LittleEndian.PutUint64(program[8:16], uint64(interpOffset))
		binary.LittleEndian.PutUint64(program[32:40], uint64(len(interp)))
		binary.LittleEndian.PutUint64(program[40:48], uint64(len(interp)))
		binary.LittleEndian.PutUint64(program[48:56], 1)
		copy(payload[interpOffset:], interp)
		programOffset += programSize
	}
	if dynamicEntries > 0 {
		program := payload[programOffset : programOffset+programSize]
		binary.LittleEndian.PutUint32(program[0:4], ptDynamic)
		binary.LittleEndian.PutUint32(program[4:8], 4)
		binary.LittleEndian.PutUint64(program[8:16], uint64(dynamicOffset))
		binary.LittleEndian.PutUint64(program[32:40], uint64(dynamicEntries*16))
		binary.LittleEndian.PutUint64(program[40:48], uint64(dynamicEntries*16))
		binary.LittleEndian.PutUint64(program[48:56], 8)
		entryOffset := dynamicOffset
		if pie {
			binary.LittleEndian.PutUint64(payload[entryOffset:entryOffset+8], dtFlags1)
			binary.LittleEndian.PutUint64(payload[entryOffset+8:entryOffset+16], df1PIE)
			entryOffset += 16
		}
		if soname {
			binary.LittleEndian.PutUint64(payload[entryOffset:entryOffset+8], dtSoname)
			binary.LittleEndian.PutUint64(payload[entryOffset+8:entryOffset+16], 1)
		}
	}
	return payload
}

// PE constructs a minimal structurally valid PE image.
func PE(dll bool) []byte {
	const peOffset, optionalSize = 64, 112
	payload := make([]byte, peOffset+24+optionalSize+40)
	copy(payload[:2], "MZ")
	binary.LittleEndian.PutUint32(payload[0x3c:0x40], peOffset)
	copy(payload[peOffset:peOffset+4], "PE\x00\x00")
	header := payload[peOffset+4 : peOffset+24]
	binary.LittleEndian.PutUint16(header[0:2], 0x8664)
	binary.LittleEndian.PutUint16(header[2:4], 1)
	binary.LittleEndian.PutUint16(header[16:18], optionalSize)
	if dll {
		binary.LittleEndian.PutUint16(header[18:20], 0x2000)
	}
	binary.LittleEndian.PutUint16(payload[peOffset+24:peOffset+26], 0x20b)
	return payload
}

// COFFObject constructs a minimal structurally valid AMD64 COFF object.
func COFFObject() []byte {
	payload := make([]byte, 20+40)
	binary.LittleEndian.PutUint16(payload[0:2], 0x8664)
	binary.LittleEndian.PutUint16(payload[2:4], 1)
	return payload
}

// MachO constructs a minimal thin Mach-O fixture.
func MachO(fileType uint32) []byte {
	payload := make([]byte, 32)
	binary.BigEndian.PutUint32(payload[0:4], 0xcffaedfe)
	binary.LittleEndian.PutUint32(payload[4:8], 0x01000007)
	binary.LittleEndian.PutUint32(payload[8:12], 3)
	binary.LittleEndian.PutUint32(payload[12:16], fileType)
	return payload
}

// FatMachO wraps one thin slice in a universal Mach-O fixture.
func FatMachO(slice []byte) []byte {
	return FatMachOSlices(slice)
}

// FatMachOSlices wraps all supplied thin slices in one universal Mach-O.
func FatMachOSlices(slices ...[]byte) []byte {
	const headerSize = 8
	tableSize := 20 * len(slices)
	offset := headerSize + tableSize
	total := offset
	for _, slice := range slices {
		if uint64(len(slice)) > uint64(^uint32(0)) || uint64(total) > uint64(^uint32(0))-uint64(len(slice)) {
			panic("conformance Mach-O slices exceed 32-bit fat-archive fields")
		}
		total += len(slice)
	}
	payload := make([]byte, total)
	binary.BigEndian.PutUint32(payload[0:4], 0xcafebabe)
	binary.BigEndian.PutUint32(payload[4:8], uint32(len(slices))) // #nosec G115 -- fixture slice count is bounded by memory.
	for index, slice := range slices {
		entry := headerSize + index*20
		binary.BigEndian.PutUint32(payload[entry:entry+4], 0x01000007+uint32(index)) // #nosec G115 -- tiny fixture index.
		binary.BigEndian.PutUint32(payload[entry+4:entry+8], 3)
		binary.BigEndian.PutUint32(payload[entry+8:entry+12], uint32(offset))      // #nosec G115 -- bounded above.
		binary.BigEndian.PutUint32(payload[entry+12:entry+16], uint32(len(slice))) // #nosec G115 -- bounded above.
		copy(payload[offset:], slice)
		offset += len(slice)
	}
	return payload
}

// DEX returns a format-correct minimal DEX fixture for adapter extensions.
func DEX() []byte {
	payload := make([]byte, 112)
	copy(payload[:8], []byte{'d', 'e', 'x', '\n', '0', '3', '5', 0})
	binary.LittleEndian.PutUint32(payload[32:36], uint32(len(payload))) // #nosec G115 -- this fixture has the fixed 112-byte DEX header size.
	binary.LittleEndian.PutUint32(payload[36:40], 112)
	binary.LittleEndian.PutUint32(payload[40:44], 0x12345678)
	signature := sha1.Sum(payload[32:]) // #nosec G401 -- SHA-1 is mandated by the DEX format and provides no trust decision.
	copy(payload[12:32], signature[:])
	binary.LittleEndian.PutUint32(payload[8:12], adler32.Checksum(payload[12:]))
	return payload
}

// Wasm returns a minimal core WebAssembly module.
func Wasm() []byte { return []byte{0, 'a', 's', 'm', 1, 0, 0, 0} }

// LLVMBitcode returns a structurally valid minimal LLVM bitcode stream.
func LLVMBitcode() []byte {
	payload := make([]byte, 16)
	copy(payload[:4], []byte{'B', 'C', 0xc0, 0xde})
	binary.LittleEndian.PutUint32(payload[4:8], 1|(8<<2)|(2<<10))
	binary.LittleEndian.PutUint32(payload[8:12], 1)
	return payload
}

// LLVMBitcodeWrapper returns the native wrapper form around LLVM bitcode.
func LLVMBitcodeWrapper() []byte {
	bitcode := LLVMBitcode()
	payload := make([]byte, 20+len(bitcode))
	copy(payload[:4], []byte{0xde, 0xc0, 0x17, 0x0b})
	binary.LittleEndian.PutUint32(payload[4:8], 1)
	binary.LittleEndian.PutUint32(payload[8:12], 20)
	binary.LittleEndian.PutUint32(payload[12:16], uint32(len(bitcode))) // #nosec G115 -- LLVMBitcode returns a fixed 16-byte fixture.
	binary.LittleEndian.PutUint32(payload[16:20], 0xffffffff)
	copy(payload[20:], bitcode)
	return payload
}

// AR returns a deterministic ordinary native archive.
func AR(members map[string][]byte) []byte {
	var buffer bytes.Buffer
	buffer.WriteString("!<arch>\n")
	names := make([]string, 0, len(members))
	for name := range members {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if len(name) > 15 || strings.Contains(name, " ") {
			panic("conformance ar member name is not short-form")
		}
		content := members[name]
		header := fmt.Sprintf("%-16s%-12d%-6d%-6d%-8s%-10d`\n", name+"/", 0, 0, 0, "100644", len(content))
		if len(header) != 60 {
			panic("conformance ar header has wrong length")
		}
		buffer.WriteString(header)
		buffer.Write(content)
		if len(content)%2 != 0 {
			buffer.WriteByte('\n')
		}
	}
	return buffer.Bytes()
}

// AROrdered returns a deterministic ar archive while preserving physical
// member order, duplicate names, and GNU/BSD/COFF metadata entries.
func AROrdered(members []AREntry) []byte {
	var buffer bytes.Buffer
	buffer.WriteString("!<arch>\n")
	for _, member := range members {
		name := member.Name
		if name != "/" && name != "/SYM64/" && name != "//" &&
			!strings.HasPrefix(name, "/") && !strings.HasPrefix(name, "#1/") &&
			!strings.HasSuffix(name, "/") {
			name += "/"
		}
		if len(name) > 16 || strings.Contains(name, " ") {
			panic("conformance ar member name does not fit its raw header")
		}
		header := fmt.Sprintf("%-16s%-12d%-6d%-6d%-8s%-10d`\n", name, 0, 0, 0, "100644", len(member.Data))
		if len(header) != 60 {
			panic("conformance ar header has wrong length")
		}
		buffer.WriteString(header)
		buffer.Write(member.Data)
		if len(member.Data)%2 != 0 {
			buffer.WriteByte('\n')
		}
	}
	return buffer.Bytes()
}

// PatchZIPFlags sets general-purpose flags in all local and central entries.
func PatchZIPFlags(payload []byte, flags uint16) []byte {
	result := append([]byte(nil), payload...)
	patch16(result, []byte{'P', 'K', 3, 4}, 6, flags)
	patch16(result, []byte{'P', 'K', 1, 2}, 8, flags)
	return result
}

// PatchZIPMethod sets a deliberately unsupported compression method.
func PatchZIPMethod(payload []byte, method uint16) []byte {
	result := append([]byte(nil), payload...)
	patch16(result, []byte{'P', 'K', 3, 4}, 8, method)
	patch16(result, []byte{'P', 'K', 1, 2}, 10, method)
	return result
}

// PatchZIPDeclaredSizes changes entry declarations without adding content, so
// limit preflight vectors prove refusal occurs before member materialization.
func PatchZIPDeclaredSizes(payload []byte, uncompressed, compressed []uint32) []byte {
	if len(uncompressed) != len(compressed) {
		panic("ZIP declared-size vectors differ in length")
	}
	result := append([]byte(nil), payload...)
	patch32Vector(result, []byte{'P', 'K', 3, 4}, 18, 22, compressed, uncompressed)
	patch32Vector(result, []byte{'P', 'K', 1, 2}, 20, 24, compressed, uncompressed)
	return result
}

// PatchZIPDisk marks an ordinary ZIP as a split/multi-volume archive.
func PatchZIPDisk(payload []byte, disk uint16) []byte {
	result := append([]byte(nil), payload...)
	eocd := bytes.LastIndex(result, []byte{'P', 'K', 5, 6})
	if eocd < 0 || eocd+6 > len(result) {
		panic("ZIP fixture lacks an end record")
	}
	binary.LittleEndian.PutUint16(result[eocd+4:eocd+6], disk)
	return result
}

// CorruptZIPBody changes member bytes without updating their CRC.
func CorruptZIPBody(payload []byte) []byte {
	result := append([]byte(nil), payload...)
	if len(result) < 31 || binary.LittleEndian.Uint32(result[:4]) != 0x04034b50 {
		panic("fixture is not a local ZIP entry")
	}
	nameLength := int(binary.LittleEndian.Uint16(result[26:28]))
	extraLength := int(binary.LittleEndian.Uint16(result[28:30]))
	offset := 30 + nameLength + extraLength
	if offset >= len(result) {
		panic("ZIP fixture has no body")
	}
	result[offset] ^= 0xff
	return result
}

// PatchZIPCentralSize creates a local/central declared-size disagreement.
func PatchZIPCentralSize(payload []byte, size uint32) []byte {
	result := append([]byte(nil), payload...)
	central := bytes.Index(result, []byte{'P', 'K', 1, 2})
	if central < 0 || central+28 > len(result) {
		panic("ZIP fixture lacks a central header")
	}
	binary.LittleEndian.PutUint32(result[central+24:central+28], size)
	return result
}

// ManyZIP returns a deterministic empty-member ZIP for entry-count limits.
func ManyZIP(count int) []byte {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for index := 0; index < count; index++ {
		header := &zip.FileHeader{Name: fmt.Sprintf("file-%06d.txt", index), Method: zip.Store}
		header.Modified = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)
		header.SetMode(0o644)
		_, err := writer.CreateHeader(header)
		must(err)
	}
	must(writer.Close())
	return buffer.Bytes()
}

// ISO9660 returns a minimal recognized unsupported filesystem image.
func ISO9660() []byte {
	payload := make([]byte, 0x8006)
	copy(payload[0x8001:], "CD001")
	return payload
}

// DMG returns a minimal recognized unsupported disk image.
func DMG() []byte {
	payload := make([]byte, 512)
	copy(payload, "koly")
	return payload
}

func patch16(payload, signature []byte, fieldOffset int, value uint16) {
	for search := 0; search < len(payload); {
		relative := bytes.Index(payload[search:], signature)
		if relative < 0 {
			return
		}
		offset := search + relative
		if offset+fieldOffset+2 > len(payload) {
			panic("ZIP fixture field is out of bounds")
		}
		binary.LittleEndian.PutUint16(payload[offset+fieldOffset:offset+fieldOffset+2], value)
		search = offset + len(signature)
	}
}

func patch32Vector(payload, signature []byte, firstOffset, secondOffset int, first, second []uint32) {
	count := 0
	for search := 0; search < len(payload); {
		relative := bytes.Index(payload[search:], signature)
		if relative < 0 {
			break
		}
		offset := search + relative
		if count >= len(first) || offset+secondOffset+4 > len(payload) {
			panic("ZIP fixture header count or bounds are invalid")
		}
		binary.LittleEndian.PutUint32(payload[offset+firstOffset:offset+firstOffset+4], first[count])
		binary.LittleEndian.PutUint32(payload[offset+secondOffset:offset+secondOffset+4], second[count])
		count++
		search = offset + len(signature)
	}
	if count != len(first) {
		panic("ZIP fixture header count does not match size vector")
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

var goldenDigests = map[string]string{
	"A01":                              "sha256:1d5fc4a3ec2e8c207177199ffd6e3567f2eb93f276c646e5c46bbddbf935b3e8",
	"A02":                              "sha256:c858feef1c686516c4a136cb41ac8778cac8c56b87306d55cd28581d46345475",
	"A02/resolved-execute":             "sha256:b411100fd52a742780693e3d36f78f3e504a4d479aa469fa4d76aa6c119a4dd9",
	"A03":                              "sha256:f20f61bb93e52ff0dc6e0073f34650831865304cadac0fc7317826e6df089771",
	"A04":                              "sha256:b11391b63a114dee3b6bd621dc017c83df2e10a2158f2cc4136a55253b30985c",
	"A05":                              "sha256:f4137fe56c6ffdc0cb471d97771042d533a7431dde089908fabd7d00acb5afca",
	"A06":                              "sha256:12399be90df084bf5dedd309c489714076c0ffa5a57b6c43dcdd2b6b01126662",
	"A07":                              "sha256:2b99d05b1af0bbb86e65fbd596090e00c847edfee02f3ab70d3d72f216f77e9a",
	"A07/native-archive-metadata":      "sha256:8ad79211c1d8edfde2b702e18864df7ebf0d9e7919dbfc4a07627afbad40c07d",
	"A08":                              "sha256:0b8155af58c9f62af506945dfac19295def82ff5744e5bb4d68e7e84980acec8",
	"A08/executable":                   "sha256:11961e393958919d0be280c4b27ca6e27480713278f8abdfd7d0f5e53e0e4a68",
	"A08/native-archive-metadata":      "sha256:c52810153436669840fd4ea05b60e5a0e65c35481e58282ee3b57486ee5c3869",
	"A08/node-addon":                   "sha256:120c66707377a2e6d70c922ef50265f6621320018c90ca611d91eee32550fbcb",
	"A08/static-library":               "sha256:7251ee37389c75f3cb1424c6f4bf901b1bd31b6fe5326a80b69efa080445f396",
	"C01":                              "sha256:e031cd9baf3c88a21a53779a8ce393064880806d6ffc13af3950bc8ae542a62d",
	"C01/exec-correct-suffix":          "sha256:01b54fdf46cf445bbf9dbdb4acf817ca8ba0e0a0dca09cb6dde33ac2c7ddbd56",
	"C01/exec-no-suffix":               "sha256:cd5f1345c1af854f687d9c7d04653fc2f0a8c33cf2ddd4be419b0bc147f0a07d",
	"C01/object-correct-suffix":        "sha256:eb1720ab89928c0095b5df7087b3ef27b574ce723b5a65f4f18f5e623691f125",
	"C01/object-no-suffix":             "sha256:d2fac3bc9b54237802cfea2e9e2d93891ce2dfc7f6e6104f4ddd67f3b4fe997a",
	"C01/object-wrong-suffix":          "sha256:ee8fe00d55bf8ff99ba9233947840dd6fc71800066330061bb26d112d3ace8ed",
	"C01a":                             "sha256:78d05eae40a1bf5b807a4ee36d18cdf1d7aca5371757f99baf3a3e925849add2",
	"C01a/executable-looking":          "sha256:273bef3f85e0f0fee2fd69968e3b05032f018c4a0f05fa0fd7e3a881bfa28e9e",
	"C01a/no-suffix":                   "sha256:e30d8b36e3cb36b2d0aec26be5cefc1d9a36211323d43793ca343973126947f5",
	"C01a/shared-object-suffix":        "sha256:4bf19f1075e96e6cf39119ee1d84803d99ba64e1375f426225020c20a84cd8df",
	"C01b":                             "sha256:6051eda7ef97c65f6ee8d09f56439c58127090a70c05131c0c6ebf873a3f250c",
	"C01b/no-suffix":                   "sha256:ac010d89f35c71db46d400afd6964ffad4ae145470c3efa9dd6259824c6942e6",
	"C01c":                             "sha256:b7a2011ef81374d97d78a8f837f6bcd2860e98e4b6ce70856daf1088eeb4ebce",
	"C01c/no-suffix":                   "sha256:a9a8bae3c2664170dd67e263affbf90f15061e0024ad3855ffd23c92944824e1",
	"C01c/shared-object-suffix":        "sha256:6e48a13bdc714e32bc035ce8f55793998144e3fd9bb8f4abc47d5394f4cd80fd",
	"C01d":                             "sha256:d8cc45031bece8955cd5064ba0e1c0566dcad8c93394bb8de98aae73f3194ba5",
	"C01d/interp-without-use":          "sha256:2f6e94857a9c415d3aa95b0a7e004eed52ea0de6df79ec055048a524fe3ad5a8",
	"C01e":                             "sha256:44314ae1544996287009d6cf2e94507d1b17131c3340bdb6d440e3b07b369bc2",
	"C01e/use-conflict":                "sha256:bab015e50edc6bacacc478da7296aa72b68025849dce19274ed6dd82d74a7b26",
	"C01f":                             "sha256:35936aeb4dbbc5287e8bb78a2f5925efabf851ec69813245d143a982bed32a5f",
	"C01f/link-or-load":                "sha256:91b67633aae16acfcd7a87995942a3536d617e77222919b4af1ffeb82819c73a",
	"C02":                              "sha256:d66f3c197c6c2cbc4e5b56e37724c4633413bc69571219106a8a0c5b13069635",
	"C02/archive-library":              "sha256:f32f052c6680760e82e076b6f3b5064e231f115b8e40f5fefea5f6df41b57aba",
	"C02/coff-object":                  "sha256:d304c9df94af53736da275bfa9cd96e319a49bf180a150c5518cbe65fd02c95e",
	"C02/dll":                          "sha256:c71d6c0cd459612d0f1484fe75ea0e9cff036f8fac2d02da092cfc00326d4ac7",
	"C02/import-library":               "sha256:14838ea4b7552561f4dd4d40788281c722f98f1a4e75784339c264029c9b63d1",
	"C03":                              "sha256:b0be8b665c3a5596245f15a258b229ac3226a54728cf8a6c052f3ed2c72f99b6",
	"C03/fat-bundle":                   "sha256:6e9d8763ab791ba56e90ef56cc119ca6fc6107afca8c1c14eefb3478858c3183",
	"C03/fat-dylib":                    "sha256:adf7006d028670c4d68ec1119fbcf0d4199f676b1b085c00ce7b6e60f2ce351c",
	"C03/fat-object":                   "sha256:f8d7cd7a4c21b16393ff0ceecc71b589855f905976e74ce94d4a924f8c6120d3",
	"C03/mixed-dynamic-first":          "sha256:1f26129d0d66da50628c23f360dfa4c14c272a4d7526579517f5f3e87e6c0d1d",
	"C03/mixed-executable-first":       "sha256:3c0684c9514838b6b22f9e90f71656d49260bf0674a0536052b5b549ff55cf52",
	"C03/thin-bundle":                  "sha256:afc8c9d40f1eff9b32f58dc0a414f3702e69977338edee96ffd0de877eab3edd",
	"C03/thin-dylib":                   "sha256:44275a1049ab9ae95ab0328883dd2a73cfb1c966ccd7bad16fe7112a4dd71978",
	"C03/thin-executable":              "sha256:b06ee983b27dd9c0f07285d498304e606a939026103c3fa9d1c1084fe80c0908",
	"C03/thin-object":                  "sha256:1f66a2eb371ad37165a6bbde41e13d8e19e0e33e31d8c913eceacd94dd780557",
	"C04":                              "sha256:c86d73df2f27dc918f7d1048f3055aa672618fc804111cc4f869e95418347575",
	"C04/a-depth-1":                    "sha256:41dcaf695dfd3eade3b614511c133c8bf7e8abb2768784af60dd0d13359a1906",
	"C04/a-depth-8":                    "sha256:b36b8d3e7107da6ff5f455f2c2e1cdb3282c22e186d3e4d220f5396a232762e8",
	"C04/bsd-metadata":                 "sha256:09a2f04fb84ebaacbe50512f0dc8e712ba7fa406e6c395e959c1421170c92090",
	"C04/coff-import-metadata":         "sha256:576c6aee824cb05a5eff8e9aeb007e466af90dcf68a4ef2884719ea8085015f2",
	"C04/gnu-metadata":                 "sha256:f248e0d791704afa0800ad327830a800ecb3e428b033a898f10fd83fb8a341bc",
	"C04/lib-depth-1":                  "sha256:ae9075e3a869fd39d47de06b4817e5dd1ee839efc6be962bea201409d5c3faa4",
	"C04/lib-depth-2":                  "sha256:98f35c0d00a0badc499294c92b8e15e83fdf774576d70a6d349f007c2b23d92f",
	"C04/lib-depth-8":                  "sha256:478778ab54b6166331892218ab226eb10503d66cdb97901bc7a70fa28fea300b",
	"C04/rlib-depth-1":                 "sha256:206808552af066660dd40e78b9fb45740c04315dd88c65f3c03ad5dba03f9c16",
	"C04/rlib-depth-2":                 "sha256:3bbfd8807131c9e07e06133e5d29f464eb468b44fb3599a58e9b6b64c011cd40",
	"C04/rlib-depth-8":                 "sha256:29cb72caa90b23ded2f1be86de5cb01dc035a8d1055c772d8b6003e7ea5729f7",
	"C05":                              "sha256:f84b203e39f667f8af79e942e6c890193187636aafff52184de2f3ad93fd504f",
	"C05/framework-renamed-nested":     "sha256:bee2428c73b660e5e641b0f651f00e2e7d831d3d50b64967a9f2a03e1884e25c",
	"C05/xcframework-renamed-nested":   "sha256:d9762274b3aec8bee0a5e814fd963dc6ccf5176256ecd6135c3cec08059a9402",
	"C05/xcframework-resource-only":    "sha256:29890ac5340af0e3de498aa24f8234d73e970ccf34734eb7fe69008b650609c9",
	"C06":                              "sha256:b83af1d489ab51900f0b3a2055f7a51cd0a546251c494005fe86eeb3d166cdce",
	"C06/macho-node":                   "sha256:66e66ef61c85a8d468312ab4be7fb5631cbda5e8b07a9d9982baedd41b97389f",
	"C06/pe-node":                      "sha256:5d94f11b6738379a441c549190d78bc2bfe6781c75083a4019dd3145bbc74206",
	"C06/python-pyd":                   "sha256:5a8237db44ef0df2e5cb755abfb42e6daa8fc72c392a457259795144d67e88b7",
	"C06/python-so":                    "sha256:71310e1567013270ca49b9fd8ed172ddc70b2064e7e0ed2880967bf461e37e1d",
	"C07":                              "sha256:ee9fab1f76a7a2ca01259aef4c19425e0b3ed180f06c756e6bb22b3ca9b43bb3",
	"C07/dex-in-aar":                   "sha256:21a9afa6f4547de16fd4c7a8762c158baff16d69ced1e995a4d3e4b261c4207c",
	"C07/direct-class":                 "sha256:be52b65331e0983f7fd6122a785da68f35c89d33ef5acfee6cd5529c9ad0616f",
	"C08":                              "sha256:4d286ad9b4f560ea6a5c9f49c4383f7db5a57c0fbcad8765729bc7b823dd5d1a",
	"C08/direct":                       "sha256:09e702674439e6e24dfbdadbec1c583aa2c4c0d0824449753e157e8938cc17bc",
	"C08/renamed":                      "sha256:b47c9cdd8dd838f7f8712dae82c637fa195aef8fbd5b7e4f12795a19aa450411",
	"C08/wheel-nested":                 "sha256:56657e799b5e987062032a6adc434fe099c65e86597dbe08400c1b97417fb2d7",
	"C09":                              "sha256:68105e25a8d707c33caf612a9d21c99a91737e3892487facf57ac02fc420c4d7",
	"C09/gch":                          "sha256:cc4a1085e453a4e560028427c225bb1af6ab5886f1bc382168f9189757fa518b",
	"C09/ifc":                          "sha256:893e039a618fe55c431f802e26331d0db1098ee6a4f75beda092b05e8ffb8ea6",
	"C09/llvm-bitcode":                 "sha256:41766d7f75b5552597de5c77e77efafa61467e4b668f0c0977f58fe350fb1746",
	"C09/node-snapshot":                "sha256:fbd0eb91a9231c1d27f31c24c3a8d70a392e00672caecf4fe0fa55ee899616fd",
	"C09/pch":                          "sha256:54a49870312d8df3d9e0cb0785733da6f5d5a2dc8fc9ac75a506eca8b4350bb8",
	"C09/pcm":                          "sha256:3dbffefd15f65ae699e6b27c2d08f72ebb231452af068d04885cba77ba102bea",
	"C09/printable-gch":                "sha256:16926131577cfb3fd2a06f946e27debded3a575766e572a31b98cc294b93d04a",
	"C09/printable-ifc":                "sha256:a77f76ec50672b354c8ee38e3d2b1fefb54e1e505ac65a8133a50fa9daf6cbb8",
	"C09/printable-pch":                "sha256:40656617def290b85c41368617f27c5e17d1262863a819a966269155279f5511",
	"C09/printable-pcm":                "sha256:4b557223ae8c2113e62f53b0c7e895c548d605b5c284048c7e5972d4b66741a3",
	"C09/printable-rmeta":              "sha256:160abaadd687c61a1fc5f93d897e918553e5e430810ab9dc18d68af5d509b15d",
	"C09/printable-swiftdoc":           "sha256:506c803e45e53f39ffdd1e19af0bedaabaf9e2d13b25b8648925a09ec1597360",
	"C09/printable-swiftmodule":        "sha256:b18d737c7d1930203943b8bcfaeb9d37da4aecea321d6d7a88ac9749b5296ab6",
	"C09/python-bytecode":              "sha256:2fbe1b326fde6c5f86531e986720abc2eddb231907f9bff162e2d0e5438d250e",
	"C09/rmeta":                        "sha256:45e9d3181e68542eb207fcb0cff4cf204ed62a1dc4b6a5e819a88e665285fb27",
	"C09/swiftdoc":                     "sha256:4f500cd617e9b71444583672da37fa7d67ca51ee69ae254e1cb86e0422f01e18",
	"C09/swiftmodule":                  "sha256:139b18ad54534821b24ba460e73cf7674322973e4c19c5e3b464695313aaeb01",
	"C09/v8-cache":                     "sha256:7db17ddafde9ace002cc028f1fa8b8bdf58ea8f5b949661513fcfd8ad8a91e53",
	"C10":                              "sha256:b34358e0bccdc81b7b5e665e01959e5ca6333c678e2d34eba88e2c83443fda2b",
	"C10/mode-cleared":                 "sha256:9c6f67e8e4cda382bf4649dd66d49a7458860b3c9150a19281d536bd9b7e3a46",
	"C11":                              "sha256:006893b5d373a53be04d1e5d5912fc74f6d576f88ea1f9afdaec25d9a6b1378d",
	"C12":                              "sha256:a54027a19b325ce47a32d14cbc26738e2fcc5a91641a7941c648ef60c501ad9b",
	"C12/node-v1":                      "sha256:89fa880e3b562c244e8e9c4b07e1e5a4dc2e806f8109fc4616ad99ef3e190b76",
	"C12/python-reference-v1":          "sha256:9167725bf453f5a6c26a9a0498d85142cbbcf04d47ca56a3f2a74ef8bed440c5",
	"C12/rust-v1":                      "sha256:48f6391850b13fdc33210ccee4f9030bf1978e7c6a9b5f38a418b5c32c20962c",
	"C12/swiftpm-v1":                   "sha256:5b54c2784a111428c9144118acce9f9f9fedd65b40183464d0f0824054e3f4ed",
	"F01":                              "sha256:3f0d1863ae06e908525bc4b0703f5c962886a781d1bc10085fcff0b62c9cecbd",
	"F01/absolute":                     "sha256:47f779eeb02a0301129639ed1c74f1072f415784942b04d089e168b02f3bfe85",
	"F01/backslash":                    "sha256:b434218dbec25523d7dcabefd86f89d6393372bf38e4317f6213920d9571db75",
	"F01/control":                      "sha256:015c3a0aa4371fa25d5991c6fe81c3d169bf163558d378691fd18e94cd117d39",
	"F01/drive":                        "sha256:f99d5e21148aabb31c7baec9f1f9bbbeb4af4fa8db5c16391a288ece2028bf0a",
	"F01/nul":                          "sha256:9999d4c529afe1f8c5a6979521e674124dc958a2c67f95d9904c2a4d931efc52",
	"F01/overlong-component":           "sha256:ea28dc5d2cc15bf19c1ef1462bcfacbb32bcf338b7c0abb4a176006811631a47",
	"F01/overlong-path":                "sha256:190cce507aae763c93216af722bfb2de0b3aa3de36edf4d743ad269ebe5d1b68",
	"F01/unc":                          "sha256:5eec083e48060967c08dd6f499b868bdc5845e9a7a5842f7f9a370d350be04ef",
	"F02":                              "sha256:de626870ebd4b4556a8a0d54bf121a97bf5e3a83fc79db8c7796663b18bfe6e2",
	"F02/case-fold":                    "sha256:35cde0e7891b98fbec93182feea7a32fcb66c3a1f417eaf18bade638eef23dc1",
	"F02/duplicate-ar-compiled-first":  "sha256:5ec5002b1a9d2621b23c49db7c863c595e5bfb2b7f8d022d1c864e22b72a04e9",
	"F02/duplicate-ar-source-first":    "sha256:ef1f260fcad0a65c9991907a1937eb0411568b4d6979627787706df8335c2296",
	"F02/duplicate-tar-compiled-first": "sha256:142bbb7dbe2fb818cf2e5eaa6828719b09979815df0a9824ee7ecf42c97ddd8b",
	"F02/duplicate-tar-source-first":   "sha256:90c51e1f8fee3e9d84cdb0c560467a231582972552c43ab187c07f57d41bb4e8",
	"F02/duplicate-zip-compiled-first": "sha256:6c3ee69b2b65bd3065c0fdd70ab4ae9a2f7cfb8a80660417af8b16872508670a",
	"F02/duplicate-zip-source-first":   "sha256:7e6e8fae087aa25d7a7c615bcde3226cdd6999622418829d42a6e890660a187f",
	"F02/nfc":                          "sha256:87e269bcf0e521fc537036400e3c905d23a9e669a44676c6d4c877b2268eae02",
	"F02/trailing-dot":                 "sha256:49a4f0ee493d875593a03e450eb8e575cc1f26c1256165e90976c679d8f686bd",
	"F03":                              "sha256:acaaba206bf209559b97c8163ef3c05fae6b2b31422bd08b536c0cf34215de1f",
	"F03/device":                       "sha256:40a0c9ab4e69c67955f04574ae24e184ca4695f7cf78e852999561c39ee011e7",
	"F03/fifo":                         "sha256:ca050db0967a902c50e80830bba6757a6ae880fb14a9015bafd7393dcfcbd084",
	"F03/hardlink":                     "sha256:b58699d4ed12e255aedbe1ea05b046ac8cba3e1df9d877da8d0702453079bd01",
	"F03/socket":                       "sha256:6c05ebf0a77f517a7f03ef4c1a23051cb7d4734d39ac22556fa88a006921f59d",
	"F03/symlink":                      "sha256:adfffb990969be6882ea83c1cd57dc1bc7eb077191695adaad33db794bcec69c",
	"F04":                              "sha256:0a346745d250fab0337b2ebd5753dadeaeb913837046d1ca60ba01fed0e21aa0",
	"F05":                              "sha256:25d0b41ad85cca41b177da8e7af6364299d7c5a2e202a1bd65fee9569d0d98ef",
	"F05/7z":                           "sha256:4ba5868b20addd97a63e4286f70bb0651b29412322b5a9dc7c84ec079166af22",
	"F05/bzip2":                        "sha256:ff5fc10d919f7a095a956fae664c81d1736fd9c147940f5b79fc58c45fde7a08",
	"F05/dmg":                          "sha256:926224c818dda9f016049b79f6e61463f4bacd5d3ac3dffa8b7d797c88c11d9c",
	"F05/iso9660":                      "sha256:494af0231950d156e9cd703ae9fff40af5355d9ca2eb8b653930977d01a1c005",
	"F05/multi-volume":                 "sha256:53031edf2439ad2aecfbec6294a3f897c7746e5628679cf4d31ce0d238d2fb0d",
	"F05/xz":                           "sha256:a05b97b394bbe0a917977a6c82443ee52fe1fa7cc026d678915d2ef58dcace70",
	"F05/zstd":                         "sha256:b337d10c3d0088565f182f7dd0927e7c97f874f5aaafc090e28da8a2dbcc78c4",
	"F06":                              "sha256:eef6952cf682728a1bde9830c5cd9b24bcb7ab523cecb2958a7717d5d7e9b871",
	"F06/crc":                          "sha256:f064ec0c14b4727a095c13d15d7002e6cc54b84920155c2e6855589fbbfeb35b",
	"F06/size":                         "sha256:a7adbaa77d3f8a86e84230a46fbdac38cf3dd2bd0b177aa6fde8a92a9045380a",
	"F06/tar-trailing":                 "sha256:a440fcde4ccaf2e443a502b6302e2a93d62bdefcf822f064ab269853892b2fea",
	"F06/truncated":                    "sha256:dd73fef2495b864eb81e01e5d7098f26090543f49ff15d95f69c1269fc0f4918",
	"F07":                              "sha256:1868ae9fa0bc602caa87ceee1162f0f09e24bb704c03bc523a3005ace4f46cd4",
	"F08":                              "sha256:f4c2709c120d68e54c3d03847f60875416eeb6899247155df0c8b9af1a3d85d6",
	"F08/container-count":              "sha256:d8ae6971e7b67cf7a14b327c75f7b183efbc3b69324e52c78da0236cb89afdac",
	"F08/entry-count":                  "sha256:06e011c5abb0a50cec33529dc1fdaa31511d1c226ddb5775b368831268040aec",
	"F08/gzip-concatenated-ratio":      "sha256:42d5233a5e16b2c77c642475b03d7c9369f96821d21779e4707388dd051d45f0",
	"F08/gzip-padded-ratio":            "sha256:aed737678f6d26a01b0697bc3d7f5a7aae2b7c9f40aaac76166a9141092cbf55",
	"F08/raw-payload":                  "sha256:b9ac273d9a333231cc4d9b4369306691f3a0db78aeea2fc6c9c31544d73cb44f",
	"F08/single-leaf":                  "sha256:a57df5d98cc3be72eb5be6efe452f3fc4474ec726f89cda5f2f80a5dbbf5fa9c",
	"F08/total-emitted":                "sha256:4abae851dd6bdb4b672e448cb77795b99f104c07a184dd091618828d178d731a",
	"F09":                              "sha256:7a67f4cac1b48374e6b3dea7c104c24903f130f2df615f7359147c27b3a94d8f",
	"F09/cancelled":                    "sha256:e84f439ea20cf447a7688f22f49a861bf91d568887e88251741afa72810238f5",
	"F10":                              "sha256:f045bbbacb6da993b092c511660d77bf228cdd24d701f056624205c2c0dfba89",
	"F10/archive-name":                 "sha256:1316d3b1ddf1ab33b6140b75a42ead55d32b865a7bd3049c919f2f38fcb13d07",
	"F10/node":                         "sha256:3c88a9f516cdaef744ce4c7ee3a8c784100c26fd9058a08d6b6b022d00595ff8",
	"F10/resolved-link-or-load-text":   "sha256:965b6bd120ae5b645a6ac12b3ef2f548e36f9e61c268317d75ad9db58d4bae50",
	"F10/wasm":                         "sha256:2ff96fbd730e9428686ed7be18310d843b979383100eec81696e10a74a0ce4a9",
	"F11":                              "sha256:7cb55a8892e69b595b913014b7dfdc0ee97d5325d73ac0e139974e029c229dce",
	"F12":                              "sha256:966edae62e871dfb22be3c4183c63643d776cbb28f9d5eac3cf50f0805fc06ac",
	"F12/undeclared-text":              "sha256:6998ec2d6f90b8c15855208942a11c55bfb003cd0e4ba293223cdcd33d3745ec",
	"F13":                              "sha256:a223c2ab43884e7a426983be0eb53c3b7f691e481a8f647b20d38ea192a8669a",
	"F14/a-first":                      "sha256:064f09122ffafc3438057d5c0782df505b1713bd9c8a92e5bfb21c9228c0cec9",
	"F14/z-first":                      "sha256:02076c7b888ea56ec893f773e6d2035fdd25eba08032bab8f8dfb47c2cb7ca49",
	"T01":                              "sha256:cc4f2f193c358c29c235ee8c9019c9aecb3e071f626bf80d23b0c2a19c36bade",
	"T01/clang":                        "sha256:e01a52df5d740901c754a6e01d118fb8b89c8b0c87ccaa570d3ae4ab7b834aad",
	"T01/node":                         "sha256:63e2ab02aa347050b2ecb555f6826b8538298a23be6f21059a5b40d6f96eb649",
	"T01/swiftc":                       "sha256:8d6ccc4cfd6c0c4fc0a9f984c73ea89c4b24afc6901c84936f0187501e4b5d5d",
	"T02":                              "sha256:7e22f62f6e1954a68181ce55caa2d3b1b3fbc886b486a8a068557c52e916b932",
	"T03":                              "sha256:bbb6b179f62526f8a2fd04f5b25823971d1ce9c8e5e5a07577910f1d57dfb08a",
	"T03/link-escape":                  "sha256:2a3c792a9c5058a1d5d142bf1a7675fb2f447199cb8c02acbe181432054f29a4",
	"T03/special-node":                 "sha256:b8aa253fd1a4bbecb19a339fde8d3a843e4b9b5bf65ebeab30e1f15806e3f029",
	"T04":                              "sha256:ba6bd3ed536380d1e181e6435f956de9ef65e913a2dffcbcd932643f582a6ce8",
	"T04/copied-preexisting":           "sha256:ba6bd3ed536380d1e181e6435f956de9ef65e913a2dffcbcd932643f582a6ce8",
	"T04/hard-link-preexisting":        "sha256:ba6bd3ed536380d1e181e6435f956de9ef65e913a2dffcbcd932643f582a6ce8",
	"T05":                              "sha256:e06b031598b5fa1ed034ba0ff7ba21b0845d8033312f4f238ea820f1617b7cfc",
	"T05/complete-input":               "sha256:902376e5f24a748f3114fb71b9fcad8d1bae47247eeef3eecc7ea2431f8bc42d",
	"T05/digest":                       "sha256:e430bb056e3e810e87c908662bff764cb5f954381db5d469f5e3c79b88f8a943",
	"T05/path":                         "sha256:16b03deb70cb2bb02bebcf3d667a7fefe633fc99e915c73d994151e120dc8289",
	"V01":                              "sha256:d2e2f8d4880bacfb9606a3b60f490fb15b0084d0750a94ce56e900f59d8f034e",
}
