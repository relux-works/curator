package artifactpolicy

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestCodecDerivesRejectedAndAdmittedNodeSemantics(t *testing.T) {
	elf, elfErr := admitDependency(t, "program", makeELF64(elfTypeExec, false, false, false), ProfileCommonV1)
	requireCode(t, elfErr, CodeCompiledDependency)
	machO, machErr := admitDependency(t, "program", makeMachO(2), ProfileCommonV1)
	requireCode(t, machErr, CodeCompiledDependency)
	pe, peErr := admitDependency(t, "program.exe", makePE(false), ProfileCommonV1)
	requireCode(t, peErr, CodeCompiledDependency)
	jvm, jvmErr := admitDependency(t, "Program.class", makeJVMClass(), ProfileCommonV1)
	requireCode(t, jvmErr, CodeCompiledDependency)
	wasm, wasmErr := admitDependency(t, "program.bin", makeWasm(), ProfileCommonV1)
	requireCode(t, wasmErr, CodeCompiledDependency)
	llvm, llvmErr := admitDependency(t, "program.bc", makeLLVMBitcodeWrapper(), ProfileCommonV1)
	requireCode(t, llvmErr, CodeCompiledDependency)
	text, textErr := admitDependency(t, "main.go", []byte("package main\n"), ProfileGoV1)
	if textErr != nil {
		t.Fatal(textErr)
	}
	archive := sourceArchiveResult(t)

	tests := []struct {
		name     string
		manifest Manifest
		mutate   func(*Manifest)
	}{
		{
			name: "ELF e_type contradicts selected executable", manifest: elf.Manifest,
			mutate: func(manifest *Manifest) {
				setObservationFact(t, &manifest.Nodes[0], "elf-v1", "e_type", "1")
				syncClassificationDiagnostics(t, manifest)
			},
		},
		{
			name: "Mach-O file type contradicts selected executable", manifest: machO.Manifest,
			mutate: func(manifest *Manifest) {
				setObservationFact(t, &manifest.Nodes[0], "macho-v1", "file_type", "1")
				syncClassificationDiagnostics(t, manifest)
			},
		},
		{
			name: "PE characteristics contradict selected image", manifest: pe.Manifest,
			mutate: func(manifest *Manifest) {
				setObservationFact(t, &manifest.Nodes[0], "pe-coff-v1", "characteristics", "8192")
				syncClassificationDiagnostics(t, manifest)
			},
		},
		{
			name: "JVM discriminator is outside the closed version range", manifest: jvm.Manifest,
			mutate: func(manifest *Manifest) {
				setObservationFact(t, &manifest.Nodes[0], "jvm-class-v1", "major", "44")
				syncClassificationDiagnostics(t, manifest)
			},
		},
		{
			name: "WebAssembly version contradicts the closed core format", manifest: wasm.Manifest,
			mutate: func(manifest *Manifest) {
				setObservationFact(t, &manifest.Nodes[0], "wasm-v1", "version", "2")
				syncClassificationDiagnostics(t, manifest)
			},
		},
		{
			name: "LLVM variant contradicts wrapper facts", manifest: llvm.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Nodes[0].Variant = "llvm.bitcode.raw"
				syncClassificationDiagnostics(t, manifest)
			},
		},
		{
			name: "text class contradicts source detector facts", manifest: text.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Nodes[0].Class = ClassSourceGeneratedText
			},
		},
		{
			name: "text detector cannot carry an unrecognized fact", manifest: text.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Nodes[0].Observations[0].Facts = append(
					manifest.Nodes[0].Observations[0].Facts,
					Fact{Key: "forged", Value: "true"},
				)
			},
		},
		{
			name: "ZIP entry size must bind the admitted child", manifest: archive.Manifest,
			mutate: func(manifest *Manifest) {
				setObservationFact(t, &manifest.Nodes[1], "archive-zip-v1", "uncompressed_size", "12")
			},
		},
		{
			name: "container decision rule is not canonical", manifest: archive.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Nodes[0].Rule = "class_role_decision"
			},
		},
		{
			name: "detector result cannot be rewritten after classification", manifest: elf.Manifest,
			mutate: func(manifest *Manifest) {
				for index := range manifest.Nodes[0].Observations {
					if manifest.Nodes[0].Observations[index].DetectorID == "elf-v1" {
						manifest.Nodes[0].Observations[index].Result = "NO_MATCH"
					}
				}
				syncClassificationDiagnostics(t, manifest)
			},
		},
		{
			name: "rejected rule cannot be replaced by an arbitrary label", manifest: elf.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Nodes[0].Rule = "forged_rejection"
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			manifest := cloneManifest(t, testCase.manifest)
			testCase.mutate(&manifest)
			if err := decodeRehashedManifest(manifest); err == nil {
				t.Fatal("self-consistently rehashed semantic forgery was admitted")
			}
		})
	}
}

func TestFindingsSummaryRequiresCapSaturationAndCompleteVerifiableEvidence(t *testing.T) {
	entries := make([]zipFixtureEntry, 0, 1_002)
	for index := 0; index < 1_001; index++ {
		entries = append(entries, zipFixtureEntry{name: "duplicate.go", content: []byte("package duplicate\n")})
	}
	entries = append(entries, zipFixtureEntry{name: "zzz-compiled", content: makeELF64(elfTypeExec, false, false, false)})
	result, err := admitDependency(t, "findings.zip", buildZIP(t, entries), ProfileGoV1)
	requireCode(t, err, CodeArchiveUnsafePath)
	if result.Manifest.Findings.Total != 1_001 || result.Manifest.Findings.Recorded != 1_000 {
		t.Fatalf("cap fixture findings = %+v", result.Manifest.Findings)
	}
	compiledEvidence := -1
	for index, evidence := range result.Manifest.Findings.Evidence {
		if evidence.Code == CodeCompiledDependency {
			compiledEvidence = index
			break
		}
	}
	if compiledEvidence < int(result.Manifest.Findings.Recorded) {
		t.Fatalf("compiled evidence index = %d, want hidden beyond recorded prefix", compiledEvidence)
	}

	t.Run("premature diagnostic truncation", func(t *testing.T) {
		manifest := cloneManifest(t, result.Manifest)
		manifest.Diagnostics = manifest.Diagnostics[:len(manifest.Diagnostics)-1]
		manifest.Findings = summarizeFindingEvidence(manifest.Findings.Evidence, int64(len(manifest.Diagnostics)))
		if err := decodeRehashedManifest(manifest); err == nil {
			t.Fatal("prematurely truncated diagnostic prefix was admitted")
		}
	})

	t.Run("invented count and duplicate evidence", func(t *testing.T) {
		manifest := cloneManifest(t, result.Manifest)
		manifest.Findings.Evidence = append(manifest.Findings.Evidence, manifest.Findings.Evidence[compiledEvidence])
		sort.Slice(manifest.Findings.Evidence, func(left, right int) bool {
			return findingEvidenceLess(manifest.Findings.Evidence[left], manifest.Findings.Evidence[right])
		})
		manifest.Findings = summarizeFindingEvidence(manifest.Findings.Evidence, manifest.Findings.Recorded)
		if err := decodeRehashedManifest(manifest); err == nil {
			t.Fatal("invented duplicate finding evidence was admitted")
		}
	})

	t.Run("hidden rejected node omitted", func(t *testing.T) {
		manifest := cloneManifest(t, result.Manifest)
		manifest.Findings.Evidence = append(
			manifest.Findings.Evidence[:compiledEvidence], manifest.Findings.Evidence[compiledEvidence+1:]...,
		)
		manifest.Findings = summarizeFindingEvidence(manifest.Findings.Evidence, manifest.Findings.Recorded)
		if err := decodeRehashedManifest(manifest); err == nil {
			t.Fatal("manifest hiding a rejected compiled node was admitted")
		}
	})

	t.Run("hidden finding semantic drift", func(t *testing.T) {
		manifest := cloneManifest(t, result.Manifest)
		diagnostic := diagnosticFromFindingEvidence(manifest.Findings.Evidence[compiledEvidence])
		diagnostic.Reason = "invented_hidden_reason"
		evidence, evidenceErr := findingEvidenceFromDiagnostic(diagnostic)
		if evidenceErr != nil {
			t.Fatal(evidenceErr)
		}
		manifest.Findings.Evidence[compiledEvidence] = evidence
		manifest.Findings = summarizeFindingEvidence(manifest.Findings.Evidence, manifest.Findings.Recorded)
		if err := decodeRehashedManifest(manifest); err == nil {
			t.Fatal("semantically drifted hidden finding was admitted")
		}
	})
}

func TestNativeArchiveMetadataIsCanonicalAndRoleBound(t *testing.T) {
	zero32 := make([]byte, 4)
	zero64 := make([]byte, 8)
	binary.BigEndian.PutUint32(zero32, 0)
	binary.LittleEndian.PutUint32(zero64[:4], 0)
	binary.LittleEndian.PutUint32(zero64[4:], 0)
	fixtures := map[string][]byte{
		"gnu": buildAROrdered(t, []arFixtureEntry{
			{name: "/", content: zero32},
			{name: "//", content: []byte("long-object-name.o/\n")},
			{name: "/0", content: makeELF64(elfTypeRel, false, false, false)},
		}),
		"bsd": buildAROrdered(t, []arFixtureEntry{
			{name: "__.SYMDEF", content: zero64},
			{name: "unit.o", content: makeELF64(elfTypeRel, false, false, false)},
		}),
		"coff-import": buildAROrdered(t, []arFixtureEntry{
			{name: "/", content: zero32},
			{name: "/", content: zero64},
			{name: "//", content: []byte("long-object.obj\x00")},
			{name: "/0", content: makeCOFFObject()},
		}),
	}

	for name, payload := range fixtures {
		t.Run(name, func(t *testing.T) {
			pathValue := name + ".a"
			dependency, dependencyErr := admitDependency(t, pathValue, payload, ProfileCommonV1)
			requireCode(t, dependencyErr, CodeCompiledDependency)
			assertNativeArchiveMetadata(t, dependency, pathValue, DecisionReject)

			toolchain, toolchainErr := NewService().AdmitToolchain(t.Context(), ToolchainRequest{
				Descriptor:    fixtureDescriptor(payload, ProfileCommonV1),
				Payload:       Payload{Path: pathValue, Size: int64(len(payload)), Reader: bytes.NewReader(payload)},
				Authorization: validToolchainAuthorization(t, pathValue, payload),
			})
			if toolchainErr != nil {
				t.Fatal(toolchainErr)
			}
			requireDecision(t, toolchain, DecisionAllowToolchain)
			assertNativeArchiveMetadata(t, toolchain, pathValue, DecisionAllowToolchain)

			output, outputErr := NewService().AdmitLocalOutput(t.Context(), validOutputRequest(
				t, pathValue, payload, ClassNativeLibraryStatic,
			))
			if outputErr != nil {
				t.Fatal(outputErr)
			}
			requireDecision(t, output, DecisionAllowOutput)
			assertNativeArchiveMetadata(t, output, pathValue, DecisionAllowOutput)
		})
	}
}

func TestDuplicateMembersAreInspectedIndependentlyOfPhysicalOrder(t *testing.T) {
	source := []byte("package duplicate\n")
	compiled := makeELF64(elfTypeExec, false, false, false)
	fixtures := map[string][2][]byte{
		"zip": {
			buildZIP(t, []zipFixtureEntry{{name: "unit.go", content: source}, {name: "unit.go", content: compiled}}),
			buildZIP(t, []zipFixtureEntry{{name: "unit.go", content: compiled}, {name: "unit.go", content: source}}),
		},
		"tar": {
			buildTar(t, []tarFixtureEntry{{name: "unit.go", content: source}, {name: "unit.go", content: compiled}}),
			buildTar(t, []tarFixtureEntry{{name: "unit.go", content: compiled}, {name: "unit.go", content: source}}),
		},
		"ar": {
			buildAROrdered(t, []arFixtureEntry{{name: "unit.go", content: source}, {name: "unit.go", content: compiled}}),
			buildAROrdered(t, []arFixtureEntry{{name: "unit.go", content: compiled}, {name: "unit.go", content: source}}),
		},
	}

	for format, payloads := range fixtures {
		t.Run(format, func(t *testing.T) {
			pathValue := "duplicates." + format
			first, firstErr := admitDependency(t, pathValue, payloads[0], ProfileGoV1)
			second, secondErr := admitDependency(t, pathValue, payloads[1], ProfileGoV1)
			firstCode, secondCode := ErrorCode(firstErr), ErrorCode(secondErr)
			if firstCode != secondCode || (firstCode != CodeCompiledDependency && firstCode != CodeArchiveUnsafePath) {
				t.Fatalf("duplicate primary codes changed with order: first=%q second=%q", firstCode, secondCode)
			}
			assertDuplicateClassInventory(t, first)
			assertDuplicateClassInventory(t, second)
			if !reflect.DeepEqual(first.Manifest.Nodes[1:], second.Manifest.Nodes[1:]) {
				t.Fatalf("physical order changed duplicate node evidence:\nfirst=%+v\nsecond=%+v", first.Manifest.Nodes[1:], second.Manifest.Nodes[1:])
			}
			if !reflect.DeepEqual(first.Manifest.Findings.Evidence, second.Manifest.Findings.Evidence) {
				t.Fatal("physical order changed complete duplicate finding evidence")
			}
		})
	}
}

func TestResolvedTextUsesAreFailClosedWithoutBreakingScripts(t *testing.T) {
	t.Run("link or load use makes benign source ambiguous", func(t *testing.T) {
		payload := []byte("package library\n")
		request := dependencyRequest("library.go", payload, ProfileGoV1)
		request.Descriptor.ResolvedUses = map[string][]UseEdge{
			"library.go": {{Kind: UseLinkOrLoad, Origin: "active_graph.edges[0]"}},
		}
		result, err := NewService().AdmitDependency(t.Context(), request)
		requireCode(t, err, CodeTypeAmbiguous)
		node := requireNode(t, result, "library.go")
		if node.Class != ClassOpaqueUnknown || result.Manifest.Diagnostics[0].Reason != "resolved_link_or_load_with_noncompiled_bytes" {
			t.Fatalf("link/load text decision = %+v diagnostics=%+v", node, result.Manifest.Diagnostics)
		}
	})

	t.Run("deny suffix plus link or load remains deterministic", func(t *testing.T) {
		payload := []byte("ordinary text\n")
		request := dependencyRequest("library.so", payload, ProfileCommonV1)
		request.Descriptor.ResolvedUses = map[string][]UseEdge{
			"library.so": {{Kind: UseLinkOrLoad, Origin: "active_graph.edges[1]"}},
		}
		result, err := NewService().AdmitDependency(t.Context(), request)
		requireCode(t, err, CodeTypeAmbiguous)
		if result.Manifest.Diagnostics[0].Reason != "resolved_link_or_load_with_noncompiled_bytes" {
			t.Fatalf("ambiguity reason = %q", result.Manifest.Diagnostics[0].Reason)
		}
	})

	t.Run("resolved execute preserves a declared interpreted script", func(t *testing.T) {
		payload := []byte("#!/bin/sh\nset -eu\nprintf ok\\n\n")
		request := dependencyRequest("tool", payload, ProfileCommonV1)
		request.Descriptor.ResolvedUses = map[string][]UseEdge{
			"tool": {{Kind: UseExecute, Origin: "active_graph.edges[2]"}},
		}
		result, err := NewService().AdmitDependency(t.Context(), request)
		if err != nil {
			t.Fatal(err)
		}
		node := requireNode(t, result, "tool")
		if node.Class != ClassSourceAuthoredText || node.Decision != DecisionAdmitInput {
			t.Fatalf("interpreted script decision = %+v", node)
		}
	})
}

func TestGZIPExpansionUsesOnlyFirstStreamCompressedBytes(t *testing.T) {
	first := buildGZIP(t, bytes.Repeat([]byte{0}, 1<<20), "bomb.bin")
	tests := map[string][]byte{
		"large trailing padding":     append(append([]byte(nil), first...), bytes.Repeat([]byte{0}, 2<<20)...),
		"concatenated second stream": append(append([]byte(nil), first...), buildGZIP(t, []byte("second stream\n"), "second.txt")...),
	}
	for name, payload := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := admitDependency(t, "bomb.gz", payload, ProfileCommonV1)
			requireLimitDiagnostic(t, result, err, "max_expansion_ratio", 201)
			accounting := result.Manifest.Accounting
			if accounting.MaxStreamInputBytes <= 0 || accounting.MaxStreamInputBytes >= int64(len(payload)) {
				t.Fatalf("first-stream compressed-byte accounting = %+v payload=%d", accounting, len(payload))
			}
			if accounting.UnmanifestedEmittedBytes > accounting.MaxStreamInputBytes*DefaultLimits().MaxExpansionRatio+1 {
				t.Fatalf("gzip traversal exceeded its dynamic first-stream budget: %+v", accounting)
			}
		})
	}
}

func setObservationFact(t *testing.T, node *ManifestNode, detectorID, key, value string) {
	t.Helper()
	for observationIndex := range node.Observations {
		observation := &node.Observations[observationIndex]
		if observation.DetectorID != detectorID {
			continue
		}
		for factIndex := range observation.Facts {
			if observation.Facts[factIndex].Key == key {
				observation.Facts[factIndex].Value = value
				return
			}
		}
	}
	t.Fatalf("observation fact %s.%s is absent", detectorID, key)
}

func syncClassificationDiagnostics(t *testing.T, manifest *Manifest) {
	t.Helper()
	for diagnosticIndex := range manifest.Diagnostics {
		diagnostic := &manifest.Diagnostics[diagnosticIndex]
		switch diagnostic.Code {
		case CodeCompiledDependency, CodeTypeAmbiguous, CodeOpaqueDependency, CodeGeneratedInputUndeclared:
		default:
			continue
		}
		for _, node := range manifest.Nodes {
			if node.Path != diagnostic.Path {
				continue
			}
			diagnostic.OriginalNameBase64 = node.OriginalNameBase64
			diagnostic.CollisionKey = node.CollisionKey
			diagnostic.Class = node.Class
			diagnostic.Variant = node.Variant
			diagnostic.DetectorID = node.SelectedDetectorID
			diagnostic.ContainerChain = append([]string{}, node.ContainerChain...)
			diagnostic.SHA256 = node.SHA256
			diagnostic.Size = node.Size
			diagnostic.Details = observationFacts(node.Observations)
		}
	}
	setManifestDiagnostics(t, manifest, manifest.Diagnostics)
}

func assertNativeArchiveMetadata(t *testing.T, result Result, root string, want Decision) {
	t.Helper()
	count := 0
	for _, node := range result.Manifest.Nodes {
		if !strings.HasPrefix(node.Path, root+"!/$ar-metadata/") || node.Kind != NodeRegularFile {
			continue
		}
		count++
		if node.Class != ClassNativeLibraryStatic || !strings.HasPrefix(node.Variant, "ar.metadata.") ||
			node.SelectedDetectorID != "archive-ar-v1" || node.SHA256 == "" || node.Decision != want {
			t.Fatalf("native archive metadata node = %+v, want decision %q", node, want)
		}
	}
	if count == 0 {
		t.Fatalf("native archive %q omitted its structural metadata nodes", root)
	}
	if _, err := DecodeManifest(result.CanonicalBytes); err != nil {
		t.Fatalf("decode native archive manifest: %v", err)
	}
}

func assertDuplicateClassInventory(t *testing.T, result Result) {
	t.Helper()
	classes := map[ArtifactClass]int{}
	for _, node := range result.Manifest.Nodes[1:] {
		if node.Kind == NodeRegularFile {
			classes[node.Class]++
		}
	}
	if classes[ClassSourceAuthoredText] != 1 || classes[ClassNativeExecutable] != 1 {
		t.Fatalf("duplicate member class inventory = %+v", classes)
	}
}
