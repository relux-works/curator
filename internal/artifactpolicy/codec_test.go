package artifactpolicy

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/protocoljson"
)

func TestCanonicalArtifactManifestCodec(t *testing.T) {
	payload := []byte("package main\nfunc main() {}\n")
	result, err := admitDependency(t, "main.go", payload, ProfileGoV1)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.CanonicalBytes) == 0 || result.CanonicalBytes[len(result.CanonicalBytes)-1] == '\n' {
		t.Fatal("canonical manifest is empty or has a terminal newline")
	}
	if !bytes.HasPrefix(result.CanonicalBytes, []byte(`{"accounting":`)) {
		t.Fatalf("CCJ object keys are not lexical: %.80s", result.CanonicalBytes)
	}
	decoded, err := DecodeManifest(result.CanonicalBytes)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := EncodeManifest(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, result.CanonicalBytes) {
		t.Fatal("manifest codec is not byte-stable")
	}
	if decoded.ManifestDigest != result.Manifest.ManifestDigest {
		t.Fatal("decode changed manifest identity")
	}
	const goldenDigest = "sha256:66b11740d0ed814eaee0a3d141778b9fb21719366ea550c8385145a3556f5d8b"
	if decoded.ManifestDigest != goldenDigest {
		t.Fatalf("artifact-manifest-v1 golden digest = %q, want %q", decoded.ManifestDigest, goldenDigest)
	}

	withWhitespace := append(append([]byte(nil), result.CanonicalBytes...), '\n')
	if _, err := DecodeManifest(withWhitespace); err == nil {
		t.Fatal("noncanonical terminal whitespace admitted")
	}
	unknownField := bytes.Replace(result.CanonicalBytes, []byte(`"adapter_id"`), []byte(`"adapter_ix"`), 1)
	if _, err := DecodeManifest(unknownField); err == nil {
		t.Fatal("unknown manifest field admitted")
	}
	changedDigest := append([]byte(nil), result.CanonicalBytes...)
	digest := strings.TrimPrefix(result.Manifest.ManifestDigest, "sha256:")
	var replacement string
	if digest[0] == '0' {
		replacement = "1" + digest[1:]
	} else {
		replacement = "0" + digest[1:]
	}
	changedDigest = bytes.Replace(changedDigest, []byte(digest), []byte(replacement), 1)
	if _, err := DecodeManifest(changedDigest); err == nil {
		t.Fatal("manifest digest drift admitted")
	}
}

func TestCodecRejectsOpenEnumsAndNoncanonicalNodeOrder(t *testing.T) {
	payload := []byte("package main\n")
	result, err := admitDependency(t, "main.go", payload, ProfileGoV1)
	if err != nil {
		t.Fatal(err)
	}
	manifest := result.Manifest
	manifest.Nodes[0].Class = ArtifactClass("source.future")
	if _, err := EncodeManifest(manifest); err == nil {
		t.Fatal("open artifact class admitted")
	}

	archive := buildZIP(t, []zipFixtureEntry{
		{name: "b.txt", content: []byte("b\n")},
		{name: "a.txt", content: []byte("a\n")},
	})
	archiveResult, err := admitDependency(t, "source.zip", archive, ProfileCommonV1)
	if err != nil {
		t.Fatal(err)
	}
	manifest = archiveResult.Manifest
	manifest.Nodes[1], manifest.Nodes[2] = manifest.Nodes[2], manifest.Nodes[1]
	manifest.ManifestDigest = ""
	if _, err := EncodeManifest(manifest); err == nil {
		t.Fatal("noncanonical manifest node order admitted")
	}
}

func TestFindingsDigestBindsUntruncatedCanonicalSet(t *testing.T) {
	payload := buildZIP(t, []zipFixtureEntry{
		{name: "B.go", content: []byte("package b\n")},
		{name: "b.go", content: []byte("package b\n")},
		{name: "C.go", content: []byte("package c\n")},
		{name: "c.go", content: []byte("package c\n")},
	})
	result, err := admitDependency(t, "collision.zip", payload, ProfileGoV1)
	requireCode(t, err, CodeArchiveUnsafePath)
	if result.Manifest.Findings.Total < 2 || result.Manifest.Findings.Recorded != result.Manifest.Findings.Total {
		t.Fatalf("findings summary = %+v", result.Manifest.Findings)
	}
	if !sha256Identity.MatchString(result.Manifest.Findings.SHA256) {
		t.Fatalf("findings digest = %q", result.Manifest.Findings.SHA256)
	}
}

func TestCodecRejectsSemanticallyNoncanonicalArrays(t *testing.T) {
	payload := makeELF64(elfTypeExec, false, false, false)
	result, err := admitDependency(t, "tool", payload, ProfileCommonV1)
	requireCode(t, err, CodeCompiledDependency)

	tests := map[string]func(*Manifest){
		"detector registry": func(manifest *Manifest) {
			manifest.Detectors[0], manifest.Detectors[1] = manifest.Detectors[1], manifest.Detectors[0]
		},
		"observation facts": func(manifest *Manifest) {
			factsValue := manifest.Nodes[0].Observations[0].Facts
			factsValue[0], factsValue[len(factsValue)-1] = factsValue[len(factsValue)-1], factsValue[0]
		},
		"diagnostics": func(manifest *Manifest) {
			manifest.Diagnostics = append(manifest.Diagnostics, manifest.Diagnostics[0])
			manifest.Diagnostics[0].Path = "z-path"
			manifest.Findings.Total++
			manifest.Findings.Recorded++
			findingsBytes, marshalErr := marshalCanonicalStruct(manifest.Diagnostics)
			if marshalErr != nil {
				t.Fatal(marshalErr)
			}
			manifest.Findings.SHA256 = digestBytes(findingsBytes)
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			manifest := result.Manifest
			manifest.Detectors = append([]DetectorIdentity(nil), manifest.Detectors...)
			manifest.Nodes = append([]ManifestNode(nil), manifest.Nodes...)
			manifest.Diagnostics = append([]Diagnostic(nil), manifest.Diagnostics...)
			manifest.Nodes[0].Observations = append([]Observation(nil), manifest.Nodes[0].Observations...)
			manifest.Nodes[0].Observations[0].Facts = append([]Fact(nil), manifest.Nodes[0].Observations[0].Facts...)
			mutate(&manifest)
			digest, digestErr := manifestDigest(manifest)
			if digestErr != nil {
				t.Fatal(digestErr)
			}
			manifest.ManifestDigest = digest
			canonical, marshalErr := marshalCanonicalStruct(manifest)
			if marshalErr != nil {
				t.Fatal(marshalErr)
			}
			if _, decodeErr := DecodeManifest(canonical); decodeErr == nil {
				t.Fatal("semantically noncanonical manifest admitted")
			}
		})
	}
}

func TestCodecRejectsSemanticallyImpossibleManifests(t *testing.T) {
	source, err := admitDependency(t, "main.go", []byte("package main\n"), ProfileGoV1)
	if err != nil {
		t.Fatal(err)
	}
	compiled, compiledErr := admitDependency(t, "tool", makeELF64(elfTypeExec, false, false, false), ProfileCommonV1)
	requireCode(t, compiledErr, CodeCompiledDependency)
	archivePayload := buildZIP(t, []zipFixtureEntry{{
		name: "tool", content: makeELF64(elfTypeExec, false, false, false), method: zip.Store,
	}})
	archive, archiveErr := admitDependency(t, "package.zip", archivePayload, ProfileCommonV1)
	requireCode(t, archiveErr, CodeCompiledDependency)
	gzipResult, gzipErr := admitDependency(
		t, "main.go.gz", buildGZIP(t, []byte("package main\n"), "main.go"), ProfileGoV1,
	)
	if gzipErr != nil {
		t.Fatal(gzipErr)
	}
	toolBytes := makeELF64(elfTypeExec, false, false, false)
	toolchain, toolchainErr := NewService().AdmitToolchain(t.Context(), ToolchainRequest{
		Descriptor:    fixtureDescriptor(toolBytes, ProfileCommonV1),
		Payload:       Payload{Path: "bin/clang", Size: int64(len(toolBytes)), Reader: bytes.NewReader(toolBytes)},
		Authorization: validToolchainAuthorization(t, "bin/clang", toolBytes),
	})
	if toolchainErr != nil {
		t.Fatal(toolchainErr)
	}
	outputBytes := makeELF64(elfTypeRel, false, false, false)
	output, outputErr := NewService().AdmitLocalOutput(t.Context(), validOutputRequest(
		t, "obj/main.o", outputBytes, ClassNativeObject,
	))
	if outputErr != nil {
		t.Fatal(outputErr)
	}
	unsafePayload := buildZIP(t, []zipFixtureEntry{{
		name: "../escape.go", content: []byte("package escape\n"), method: zip.Store,
	}})
	unsafe, unsafeErr := admitDependency(t, "unsafe.zip", unsafePayload, ProfileGoV1)
	requireCode(t, unsafeErr, CodeArchiveUnsafePath)

	tests := []struct {
		name     string
		manifest Manifest
		mutate   func(*Manifest)
	}{
		{
			name: "file root hash differs from raw payload", manifest: source.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Nodes[0].SHA256 = digestBytes([]byte("forged-root"))
			},
		},
		{
			name: "rejecting byte node omits digest", manifest: archive.Manifest,
			mutate: func(manifest *Manifest) {
				for index := range manifest.Nodes {
					if manifest.Nodes[index].Path == "package.zip!/tool" {
						manifest.Nodes[index].SHA256 = ""
					}
				}
			},
		},
		{
			name: "detector error appears on admitted complete node", manifest: source.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Nodes[0].Observations[0].Result = "ERROR"
			},
		},
		{
			name: "container chain is not derived from parent graph", manifest: archive.Manifest,
			mutate: func(manifest *Manifest) {
				for index := range manifest.Nodes {
					if manifest.Nodes[index].Path == "package.zip!/tool" {
						manifest.Nodes[index].ContainerChain = []string{}
					}
				}
			},
		},
		{
			name: "complete compressed stream loses its only child", manifest: gzipResult.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Nodes = manifest.Nodes[:1]
			},
		},
		{
			name: "compiled dependency rewritten as admitted", manifest: compiled.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Nodes[0].Decision = DecisionAdmitInput
				manifest.Nodes[0].Rule = "forged_allow"
				manifest.Decision = DecisionAdmitInput
				setManifestDiagnostics(t, manifest, nil)
			},
		},
		{
			name: "dependency rewritten with toolchain decision", manifest: source.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Nodes[0].Decision = DecisionAllowToolchain
				manifest.Decision = DecisionAllowToolchain
			},
		},
		{
			name: "incomplete container admitted", manifest: sourceArchiveResult(t).Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Nodes[0].InspectionComplete = false
			},
		},
		{
			name: "rejecting descendant has admitted ancestor", manifest: archive.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Nodes[0].Decision = DecisionDescend
				manifest.Nodes[0].Rule = "forged_descend"
			},
		},
		{
			name: "rejected final decision has admitted node", manifest: source.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Decision = DecisionReject
				setManifestDiagnostics(t, manifest, []Diagnostic{{
					Code: CodeOriginUnverified, Path: "main.go", Reason: "forged_origin_failure",
				}})
			},
		},
		{
			name: "regular file rewritten as directory class", manifest: source.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Nodes[0].Class = ClassDirectory
			},
		},
		{
			name: "dependency role evidence omits origin binding", manifest: source.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.RoleEvidence = []Fact{}
			},
		},
		{
			name: "admitted dependency rewrites verified origin", manifest: source.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Origin.Verified = false
				for index := range manifest.RoleEvidence {
					if manifest.RoleEvidence[index].Key == "origin_verified" {
						manifest.RoleEvidence[index].Value = "false"
					}
				}
			},
		},
		{
			name: "allowed toolchain rewrites contained-link evidence", manifest: toolchain.Manifest,
			mutate: func(manifest *Manifest) {
				setFactValue(manifest.RoleEvidence, "contained_links_validated", "false")
			},
		},
		{
			name: "allowed output rewrites complete-input evidence", manifest: output.Manifest,
			mutate: func(manifest *Manifest) {
				setFactValue(manifest.RoleEvidence, "complete_input_matched", "false")
			},
		},
		{
			name: "entry accounting omits manifested node", manifest: sourceArchiveResult(t).Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Accounting.EntryCount = 0
			},
		},
		{
			name: "rejecting entry accounting invents unmanifested entries", manifest: archive.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Accounting.EntryCount++
				manifest.Accounting.UnmanifestedEntryCount++
			},
		},
		{
			name: "unsafe traversal accounting exceeds its explicit failure binding", manifest: unsafe.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Accounting.EntryCount++
				manifest.Accounting.UnmanifestedEntryCount++
			},
		},
		{
			name: "traversal failure binding points at another path", manifest: unsafe.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.TraversalFailure.Path = "unsafe.zip!/other"
			},
		},
		{
			name: "rejecting emitted accounting invents unmanifested bytes", manifest: archive.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Accounting.TotalEmittedBytes++
				manifest.Accounting.UnmanifestedEmittedBytes++
				manifest.Accounting.AggregateExpansionRatio = ceilingRatio(
					manifest.Accounting.TotalEmittedBytes, manifest.Accounting.RawPayloadBytes,
				)
			},
		},
		{
			name: "admitted emitted-byte accounting is inflated", manifest: sourceArchiveResult(t).Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Accounting.TotalEmittedBytes++
				manifest.Accounting.AggregateExpansionRatio = ceilingRatio(
					manifest.Accounting.TotalEmittedBytes, manifest.Accounting.RawPayloadBytes,
				)
			},
		},
		{
			name: "diagnostic role mismatch", manifest: compiled.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Diagnostics[0].Code = CodeToolchainUntrusted
				setManifestDiagnostics(t, manifest, manifest.Diagnostics)
			},
		},
		{
			name: "classification diagnostic path drift", manifest: compiled.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Diagnostics[0].Path = "other"
				setManifestDiagnostics(t, manifest, manifest.Diagnostics)
			},
		},
		{
			name: "classification diagnostic size drift", manifest: compiled.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Diagnostics[0].Size++
				setManifestDiagnostics(t, manifest, manifest.Diagnostics)
			},
		},
		{
			name: "classification diagnostic hash drift", manifest: compiled.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Diagnostics[0].SHA256 = digestBytes([]byte("forged-diagnostic"))
				setManifestDiagnostics(t, manifest, manifest.Diagnostics)
			},
		},
		{
			name: "classification diagnostic collision drift", manifest: compiled.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Diagnostics[0].CollisionKey = "forged"
				setManifestDiagnostics(t, manifest, manifest.Diagnostics)
			},
		},
		{
			name: "classification diagnostic detector drift", manifest: compiled.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Diagnostics[0].DetectorID = "wasm-v1"
				setManifestDiagnostics(t, manifest, manifest.Diagnostics)
			},
		},
		{
			name: "classification diagnostic detail drift", manifest: compiled.Manifest,
			mutate: func(manifest *Manifest) {
				manifest.Diagnostics[0].Details = append(manifest.Diagnostics[0].Details, Fact{Key: "forged", Value: "true"})
				setManifestDiagnostics(t, manifest, manifest.Diagnostics)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			manifest := cloneManifest(t, testCase.manifest)
			testCase.mutate(&manifest)
			if err := decodeRehashedManifest(manifest); err == nil {
				t.Fatal("semantically impossible manifest was accepted after a self-consistent rehash")
			}
		})
	}
}

func sourceArchiveResult(t *testing.T) Result {
	t.Helper()
	payload := buildZIP(t, []zipFixtureEntry{{
		name: "main.go", content: []byte("package main\n"), method: zip.Store,
	}})
	result, err := admitDependency(t, "source.zip", payload, ProfileGoV1)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func setFactValue(input []Fact, key, value string) {
	for index := range input {
		if input[index].Key == key {
			input[index].Value = value
			return
		}
	}
	panic("test fact not found: " + key)
}

func setManifestDiagnostics(t *testing.T, manifest *Manifest, diagnostics []Diagnostic) {
	t.Helper()
	if diagnostics == nil {
		diagnostics = []Diagnostic{}
	}
	findings, err := findingsFromDiagnostics(manifest.LimitVector.MaxRecordedFindings, diagnostics)
	if err != nil {
		t.Fatal(err)
	}
	manifest.Diagnostics = findings.recorded()
	manifest.Findings = findings.summary()
}

func cloneManifest(t *testing.T, manifest Manifest) Manifest {
	t.Helper()
	payload, err := marshalCanonicalStruct(manifest)
	if err != nil {
		t.Fatal(err)
	}
	var result Manifest
	if err := protocoljson.UnmarshalCanonical(payload, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func decodeRehashedManifest(manifest Manifest) error {
	manifest.ManifestDigest = ""
	digest, err := manifestDigest(manifest)
	if err != nil {
		return err
	}
	manifest.ManifestDigest = digest
	payload, err := marshalCanonicalStruct(manifest)
	if err != nil {
		return err
	}
	_, err = DecodeManifest(payload)
	return err
}
