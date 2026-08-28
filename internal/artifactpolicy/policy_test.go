package artifactpolicy

import (
	"archive/zip"
	"bytes"
	"path/filepath"
	"testing"
)

func TestAdmissionVectorsA01ThroughA08(t *testing.T) {
	t.Run("A01 authored source lock manifest and license", func(t *testing.T) {
		payload := buildZIP(t, []zipFixtureEntry{
			{name: "go.mod", content: []byte("module example.test/a\n\ngo 1.25\n"), method: zip.Store},
			{name: "go.sum", content: []byte("example.test/b v1.0.0 h1:fixture\n"), method: zip.Store},
			{name: "LICENSE", content: []byte("Permission is hereby granted.\n"), method: zip.Store},
			{name: "main.go", content: []byte("package main\nfunc main() {}\n"), method: zip.Store},
		})
		result, err := admitDependency(t, "source.zip", payload, ProfileGoV1)
		if err != nil {
			t.Fatal(err)
		}
		requireDecision(t, result, DecisionAdmitInput)
		if result.Admission == nil || result.Admission.Role() != RoleDependencyInput {
			t.Fatal("dependency admission token missing")
		}
		if got := requireNode(t, result, "source.zip!/main.go").Class; got != ClassSourceAuthoredText {
			t.Fatalf("main.go class = %q", got)
		}
	})

	t.Run("A02 executable shell source", func(t *testing.T) {
		payload := []byte("#!/bin/sh\nset -eu\nprintf '%s\\n' ok\n")
		request := dependencyRequest("bin/tool", payload, ProfileCommonV1)
		result, err := NewService().AdmitDependency(t.Context(), request)
		if err != nil {
			t.Fatal(err)
		}
		node := requireNode(t, result, "bin/tool")
		if node.Class != ClassSourceAuthoredText || node.Decision != DecisionAdmitInput {
			t.Fatalf("shell node = %+v", node)
		}
	})

	t.Run("A03 shipped generated JavaScript and source map", func(t *testing.T) {
		payload := buildZIP(t, []zipFixtureEntry{
			{name: "dist/app.min.js", content: []byte("(()=>{console.log('ok')})();\n"), method: zip.Store},
			{name: "dist/app.min.js.map", content: []byte(`{"version":3,"sources":["app.ts"],"mappings":""}`), method: zip.Store},
			{name: "package.json", content: []byte(`{"name":"fixture","version":"1.0.0"}`), method: zip.Store},
		})
		result, err := admitDependency(t, "package.zip", payload, ProfileNodeV1)
		if err != nil {
			t.Fatal(err)
		}
		if got := requireNode(t, result, "package.zip!/dist/app.min.js").Class; got != ClassSourceGeneratedText {
			t.Fatalf("minified JS class = %q", got)
		}
		if got := requireNode(t, result, "package.zip!/dist/app.min.js.map").Class; got != ClassTextMetadata {
			t.Fatalf("source map class = %q", got)
		}
	})

	t.Run("A04 Swift interface plus source", func(t *testing.T) {
		payload := buildZIP(t, []zipFixtureEntry{
			{name: "Sources/API.swift", content: []byte("public struct API { public init() {} }\n"), method: zip.Store},
			{name: "Sources/API.swiftinterface", content: []byte("// swift-interface-format-version: 1.0\npublic struct API {}\n"), method: zip.Store},
		})
		result, err := admitDependency(t, "swift-source.zip", payload, ProfileSwiftPMV1)
		if err != nil {
			t.Fatal(err)
		}
		if got := requireNode(t, result, "swift-source.zip!/Sources/API.swiftinterface").Class; got != ClassSourceGeneratedText {
			t.Fatalf("swiftinterface class = %q", got)
		}
	})

	t.Run("A05 ZIP to tar gzip source recursion", func(t *testing.T) {
		tarPayload := buildTar(t, []tarFixtureEntry{{name: "src/main.rs", content: []byte("pub fn answer() -> u8 { 42 }\n")}})
		gzipPayload := buildGZIP(t, tarPayload, "bundle.tar")
		payload := buildZIP(t, []zipFixtureEntry{{name: "bundle.tar.gz", content: gzipPayload, method: zip.Store}})
		result, err := admitDependency(t, "source.zip", payload, ProfileRustV1)
		if err != nil {
			t.Fatal(err)
		}
		requireDecision(t, result, DecisionAdmitInput)
		leaf := requireNode(t, result, "source.zip!/bundle.tar.gz!/bundle.tar!/src/main.rs")
		if leaf.Class != ClassSourceAuthoredText || len(leaf.ContainerChain) != 3 {
			t.Fatalf("nested leaf = %+v", leaf)
		}
	})

	t.Run("A06 source-only wheel with RECORD", func(t *testing.T) {
		payload := buildWheel(t, map[string][]byte{
			"fixture/__init__.py":              []byte("VALUE = 42\n"),
			"fixture-1.0.0.dist-info/METADATA": []byte("Metadata-Version: 2.1\nName: fixture\nVersion: 1.0.0\n"),
			"fixture-1.0.0.dist-info/WHEEL":    []byte("Wheel-Version: 1.0\nRoot-Is-Purelib: true\n"),
		})
		result, err := admitDependency(t, "fixture-1.0.0-py3-none-any.whl", payload, ProfilePythonSourceV1)
		if err != nil {
			t.Fatal(err)
		}
		requireDecision(t, result, DecisionAdmitInput)
		if got := requireNode(t, result, "fixture-1.0.0-py3-none-any.whl!/fixture/__init__.py").Class; got != ClassSourceAuthoredText {
			t.Fatalf("wheel source class = %q", got)
		}
	})

	t.Run("A07 selected external compiler", func(t *testing.T) {
		payload := makeELF64(elfTypeExec, false, false, false)
		request := ToolchainRequest{
			Descriptor:    fixtureDescriptor(payload, ProfileCommonV1),
			Payload:       Payload{Path: "bin/clang", Size: int64(len(payload)), Reader: bytes.NewReader(payload)},
			Authorization: validToolchainAuthorization(t, "bin/clang", payload),
		}
		result, err := NewService().AdmitToolchain(t.Context(), request)
		if err != nil {
			t.Fatal(err)
		}
		requireDecision(t, result, DecisionAllowToolchain)
		if requireNode(t, result, "bin/clang").Class != ClassNativeExecutable {
			t.Fatal("toolchain executable was not classified")
		}
	})

	t.Run("A08 protected local object library addon and executable outputs", func(t *testing.T) {
		fixtures := []struct {
			path    string
			payload []byte
			class   ArtifactClass
		}{
			{path: "obj/main.o", payload: makeELF64(elfTypeRel, false, false, false), class: ClassNativeObject},
			{path: "lib/libmain.a", payload: buildAR(t, map[string][]byte{"main.o": makeELF64(elfTypeRel, false, false, false)}), class: ClassNativeLibraryStatic},
			{path: "lib/addon.node", payload: makeELF64(elfTypeDyn, false, false, true), class: ClassNodeExtension},
			{path: "bin/tool", payload: makeELF64(elfTypeExec, false, false, false), class: ClassNativeExecutable},
		}
		for _, fixture := range fixtures {
			t.Run(fixture.path, func(t *testing.T) {
				request := validOutputRequest(t, fixture.path, fixture.payload, fixture.class)
				result, err := NewService().AdmitLocalOutput(t.Context(), request)
				if err != nil {
					t.Fatal(err)
				}
				requireDecision(t, result, DecisionAllowOutput)
				if node := requireNode(t, result, fixture.path); node.Class != fixture.class {
					t.Fatalf("output class = %q, want %q", node.Class, fixture.class)
				}
				if err := AuthorizeCachePublication(result.Admission); err != nil {
					t.Fatal(err)
				}
			})
		}
	})
}

func TestTrustBoundaryVectorsT01ThroughT05AndV01(t *testing.T) {
	t.Run("T01 package path cannot claim toolchain", func(t *testing.T) {
		payload := buildZIP(t, []zipFixtureEntry{{
			name: "vendor/toolchain/bin/rustc", content: makeELF64(elfTypeExec, false, false, false), method: zip.Store,
		}})
		result, err := admitDependency(t, "package.zip", payload, ProfileRustV1)
		requireCode(t, err, CodeCompiledDependency)
		if got := requireNode(t, result, "package.zip!/vendor/toolchain/bin/rustc").Class; got != ClassNativeExecutable {
			t.Fatalf("rustc class = %q", got)
		}
	})

	t.Run("T02 arbitrary host component is not a toolchain", func(t *testing.T) {
		payload := makeELF64(elfTypeDyn, false, false, true)
		result, err := NewService().AdmitToolchain(t.Context(), ToolchainRequest{
			Descriptor: fixtureDescriptor(payload, ProfileCommonV1),
			Payload:    Payload{Path: "lib/libfoo.so", Size: int64(len(payload)), Reader: bytes.NewReader(payload)},
		})
		requireCode(t, err, CodeToolchainUntrusted)
		requireDecision(t, result, DecisionReject)
	})

	t.Run("T03 toolchain identity changes before use", func(t *testing.T) {
		payload := makeELF64(elfTypeExec, false, false, false)
		record := validToolchainAuthorization(t, "bin/tool", payload).artifactPolicyToolchainAuthorization()
		for name, mutate := range map[string]func(*toolchainAuthorizationRecord){
			"escaping selected path": func(record *toolchainAuthorizationRecord) {
				record.environmentSearchResolution = filepath.Join(record.resolvedRoot, "..", "escaped-tool")
			},
			"unvalidated contained link": func(record *toolchainAuthorizationRecord) {
				record.containedLinksValidated = false
			},
			"unvalidated special node": func(record *toolchainAuthorizationRecord) {
				record.ordinaryNodesValidated = false
			},
		} {
			t.Run(name, func(t *testing.T) {
				invalid := record
				mutate(&invalid)
				if code, _ := validateToolchainRecord(invalid, managerAuthorizationSeal); code == "" {
					t.Fatal("unsafe toolchain checkpoint passed central validation")
				}
				result, err := NewService().AdmitToolchain(t.Context(), ToolchainRequest{
					Descriptor:    fixtureDescriptor(payload, ProfileCommonV1),
					Payload:       Payload{Path: "bin/tool", Size: int64(len(payload)), Reader: bytes.NewReader(payload)},
					Authorization: sealedToolchainAuthorization{record: invalid},
				})
				requireCode(t, err, CodeToolchainUntrusted)
				assertNoAuthorization(t, result)
			})
		}
		record.timeOfUseFingerprintSHA256 = digestBytes([]byte("after"))
		result, err := NewService().AdmitToolchain(t.Context(), ToolchainRequest{
			Descriptor:    fixtureDescriptor(payload, ProfileCommonV1),
			Payload:       Payload{Path: "bin/tool", Size: int64(len(payload)), Reader: bytes.NewReader(payload)},
			Authorization: sealedToolchainAuthorization{record: record},
		})
		requireCode(t, err, CodeToolchainIdentityChanged)
		requireDecision(t, result, DecisionReject)
	})

	t.Run("T04 preexisting output is unreceipted", func(t *testing.T) {
		payload := makeELF64(elfTypeRel, false, false, false)
		request := validOutputRequest(t, "obj/main.o", payload, ClassNativeObject)
		request.Authorization = nil
		result, err := NewService().AdmitLocalOutput(t.Context(), request)
		requireCode(t, err, CodeLocalOutputUnreceipted)
		requireDecision(t, result, DecisionReject)
	})

	t.Run("T05 output expectation drifts", func(t *testing.T) {
		payload := makeELF64(elfTypeRel, false, false, false)
		base := validOutputRequest(t, "obj/main.o", payload, ClassNativeObject)
		fixtures := []struct {
			name    string
			path    string
			payload []byte
		}{
			{name: "digest", path: "obj/main.o", payload: func() []byte {
				value := append([]byte(nil), payload...)
				value[len(value)-1] ^= 1
				return value
			}()},
			{name: "path", path: "obj/copied.o", payload: payload},
			{name: "size", path: "obj/main.o", payload: append(append([]byte(nil), payload...), 0)},
		}
		for _, fixture := range fixtures {
			t.Run(fixture.name, func(t *testing.T) {
				request := base
				request.Payload = Payload{Path: fixture.path, Size: int64(len(fixture.payload)), Reader: bytes.NewReader(fixture.payload)}
				result, err := NewService().AdmitLocalOutput(t.Context(), request)
				requireCode(t, err, CodeLocalOutputDrift)
				requireDecision(t, result, DecisionReject)
				assertNoAuthorization(t, result)
			})
		}

		t.Run("complete input mismatch cannot be issued", func(t *testing.T) {
			record := base.Authorization.artifactPolicyLocalOutputAuthorization()
			record.completeInputMatched = false
			if code, _ := validateLocalOutputRecord(record, managerAuthorizationSeal); code != CodeLocalOutputDrift {
				t.Fatalf("incomplete input validation code = %q, want %q", code, CodeLocalOutputDrift)
			}
			base.Authorization = sealedLocalOutputAuthorization{record: record}
			result, err := NewService().AdmitLocalOutput(t.Context(), base)
			requireCode(t, err, CodeLocalOutputDrift)
			assertNoAuthorization(t, result)
		})
	})

	t.Run("V01 verified binary capability unavailable", func(t *testing.T) {
		payload := makePE(false)
		result, err := NewService().AdmitVerifiedBinary(t.Context(), VerifiedBinaryRequest{
			Descriptor: fixtureDescriptor(payload, ProfileCommonV1),
			Payload:    Payload{Path: "signed.exe", Size: int64(len(payload)), Reader: bytes.NewReader(payload)},
		})
		requireCode(t, err, CodeBinaryAdmissionUnavailable)
		requireDecision(t, result, DecisionReject)
		if result.Admission != nil {
			t.Fatal("unavailable verified-binary capability returned a token")
		}
	})
}

func TestPreExecutionAndPublicationTokensAreRoleBound(t *testing.T) {
	source := []byte("package main\nfunc main() {}\n")
	dependency, err := admitDependency(t, "main.go", source, ProfileGoV1)
	if err != nil {
		t.Fatal(err)
	}
	toolBytes := makeELF64(elfTypeExec, false, false, false)
	toolchain, err := NewService().AdmitToolchain(t.Context(), ToolchainRequest{
		Descriptor:    fixtureDescriptor(toolBytes, ProfileCommonV1),
		Payload:       Payload{Path: "bin/go", Size: int64(len(toolBytes)), Reader: bytes.NewReader(toolBytes)},
		Authorization: validToolchainAuthorization(t, "bin/go", toolBytes),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := AuthorizeAdapterExecution([]*Admission{dependency.Admission}, toolchain.Admission); err != nil {
		t.Fatal(err)
	}
	if err := AuthorizeAdapterExecution([]*Admission{toolchain.Admission}, dependency.Admission); err == nil {
		t.Fatal("role-swapped tokens authorized execution")
	}
	if err := AuthorizeCachePublication(dependency.Admission); err == nil {
		t.Fatal("dependency token authorized cache publication")
	}
}

func TestManagerOwnedRoleAuthorizationCannotBeForgedOrReplayed(t *testing.T) {
	toolBytes := makeELF64(elfTypeExec, false, false, false)
	validTool := validToolchainAuthorization(t, "bin/clang", toolBytes)
	toolRecord := validTool.artifactPolicyToolchainAuthorization()

	t.Run("fabricated toolchain checkpoint", func(t *testing.T) {
		forged := toolRecord
		forged.seal = &authorizationSeal{}
		result, err := NewService().AdmitToolchain(t.Context(), ToolchainRequest{
			Descriptor: fixtureDescriptor(toolBytes, ProfileCommonV1),
			Payload:    Payload{Path: "bin/clang", Size: int64(len(toolBytes)), Reader: bytes.NewReader(toolBytes)},
			Authorization: forgedToolchainAuthorization{
				record: forged,
			},
		})
		requireCode(t, err, CodeToolchainUntrusted)
		assertNoAuthorization(t, result)
	})

	t.Run("toolchain path replay", func(t *testing.T) {
		result, err := NewService().AdmitToolchain(t.Context(), ToolchainRequest{
			Descriptor:    fixtureDescriptor(toolBytes, ProfileCommonV1),
			Payload:       Payload{Path: "bin/clang-copy", Size: int64(len(toolBytes)), Reader: bytes.NewReader(toolBytes)},
			Authorization: validTool,
		})
		requireCode(t, err, CodeToolchainUntrusted)
		assertNoAuthorization(t, result)
	})

	t.Run("toolchain payload replay", func(t *testing.T) {
		drifted := append([]byte(nil), toolBytes...)
		drifted[len(drifted)-1] ^= 1
		result, err := NewService().AdmitToolchain(t.Context(), ToolchainRequest{
			Descriptor:    fixtureDescriptor(drifted, ProfileCommonV1),
			Payload:       Payload{Path: "bin/clang", Size: int64(len(drifted)), Reader: bytes.NewReader(drifted)},
			Authorization: validTool,
		})
		requireCode(t, err, CodeToolchainIdentityChanged)
		assertNoAuthorization(t, result)
	})

	outputBytes := makeELF64(elfTypeRel, false, false, false)
	validOutput := validLocalOutputAuthorization(t, "obj/main.o", outputBytes, ClassNativeObject)
	outputRecord := validOutput.artifactPolicyLocalOutputAuthorization()

	t.Run("fabricated protected receipt", func(t *testing.T) {
		forged := outputRecord
		forged.seal = &authorizationSeal{}
		result, err := NewService().AdmitLocalOutput(t.Context(), LocalOutputRequest{
			Descriptor: fixtureDescriptor(outputBytes, ProfileCommonV1),
			Payload:    Payload{Path: "obj/main.o", Size: int64(len(outputBytes)), Reader: bytes.NewReader(outputBytes)},
			Authorization: forgedLocalOutputAuthorization{
				record: forged,
			},
		})
		requireCode(t, err, CodeLocalOutputUnreceipted)
		assertNoAuthorization(t, result)
	})

	t.Run("copied or hard-linked preexisting output cannot be issued", func(t *testing.T) {
		copied := outputRecord
		copied.observedProduction = false
		copied.preexistingInputExcluded = false
		copied.hardlinkSourceExcluded = false
		if code, _ := validateLocalOutputRecord(copied, managerAuthorizationSeal); code != CodeLocalOutputUnreceipted {
			t.Fatalf("preexisting output validation code = %q, want %q", code, CodeLocalOutputUnreceipted)
		}
		result, err := NewService().AdmitLocalOutput(t.Context(), LocalOutputRequest{
			Descriptor:    fixtureDescriptor(outputBytes, ProfileCommonV1),
			Payload:       Payload{Path: "obj/main.o", Size: int64(len(outputBytes)), Reader: bytes.NewReader(outputBytes)},
			Authorization: sealedLocalOutputAuthorization{record: copied},
		})
		requireCode(t, err, CodeLocalOutputUnreceipted)
		assertNoAuthorization(t, result)
	})

	t.Run("local output path replay", func(t *testing.T) {
		result, err := NewService().AdmitLocalOutput(t.Context(), LocalOutputRequest{
			Descriptor:    fixtureDescriptor(outputBytes, ProfileCommonV1),
			Payload:       Payload{Path: "obj/copied.o", Size: int64(len(outputBytes)), Reader: bytes.NewReader(outputBytes)},
			Authorization: validOutput,
		})
		requireCode(t, err, CodeLocalOutputDrift)
		assertNoAuthorization(t, result)
	})
}

type forgedToolchainAuthorization struct {
	record toolchainAuthorizationRecord
}

func (authorization forgedToolchainAuthorization) artifactPolicyToolchainAuthorization() toolchainAuthorizationRecord {
	return authorization.record
}

type forgedLocalOutputAuthorization struct {
	record localOutputAuthorizationRecord
}

func (authorization forgedLocalOutputAuthorization) artifactPolicyLocalOutputAuthorization() localOutputAuthorizationRecord {
	return authorization.record
}

func assertNoAuthorization(t *testing.T, result Result) {
	t.Helper()
	if result.Admission != nil {
		t.Fatal("rejected evidence returned an authorization token")
	}
	if err := AuthorizeCachePublication(result.Admission); err == nil {
		t.Fatal("rejected evidence authorized cache publication")
	}
	if err := AuthorizeAdapterExecution([]*Admission{result.Admission}, result.Admission); err == nil {
		t.Fatal("rejected evidence authorized adapter execution")
	}
}

func TestServiceZeroValueUsesClosedProductionConfiguration(t *testing.T) {
	payload := []byte("package main\nfunc main() {}\n")
	request := dependencyRequest("main.go", payload, ProfileGoV1)
	var service Service
	result, err := service.AdmitDependency(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}
	requireDecision(t, result, DecisionAdmitInput)
	if len(result.Manifest.Detectors) != len(detectorIdentities()) {
		t.Fatal("zero-value service did not bind the production detector registry")
	}

	misconfigured := &Service{limits: DefaultLimits(), detectors: detectorIdentities()}
	misconfigured.detectors[0], misconfigured.detectors[1] = misconfigured.detectors[1], misconfigured.detectors[0]
	if _, err := misconfigured.AdmitDependency(t.Context(), request); err == nil {
		t.Fatal("non-production detector registry was accepted")
	}
}

func TestDeclaredSourceGrammarMustLexCompletely(t *testing.T) {
	result, err := admitDependency(t, "main.go", []byte("package main\nvar value = \\q\n"), ProfileGoV1)
	requireCode(t, err, CodeOpaqueDependency)
	requireDecision(t, result, DecisionReject)
	if result.Admission != nil {
		t.Fatal("lexically invalid declared source returned an admission token")
	}
}

func validOutputRequest(t *testing.T, path string, payload []byte, class ArtifactClass) LocalOutputRequest {
	t.Helper()
	return LocalOutputRequest{
		Descriptor:    fixtureDescriptor(payload, ProfileCommonV1),
		Payload:       Payload{Path: path, Size: int64(len(payload)), Reader: bytes.NewReader(payload)},
		Authorization: validLocalOutputAuthorization(t, path, payload, class),
	}
}
