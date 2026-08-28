package artifactpolicy

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	policyconformance "github.com/relux-works/curator/internal/artifactpolicy/conformance"
)

func TestReusableArtifactManifestV1ConformanceCorpus(t *testing.T) {
	results := make(map[string]Result)
	fixtures := policyconformance.Cases()
	if len(fixtures) != 182 {
		t.Fatalf("shared corpus cases = %d, want 182", len(fixtures))
	}
	seen := make(map[string]struct{}, len(fixtures))
	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(fixture.Key(), func(t *testing.T) {
			if _, duplicate := seen[fixture.Key()]; duplicate {
				t.Fatalf("duplicate conformance key %q", fixture.Key())
			}
			seen[fixture.Key()] = struct{}{}
			result, err := executeConformanceCase(t, fixture)
			results[fixture.Key()] = result

			if fixture.Expected.PrimaryCode == "" {
				if err != nil {
					t.Fatalf("unexpected policy rejection: %v", err)
				}
			} else if got := string(ErrorCode(err)); got != fixture.Expected.PrimaryCode {
				t.Fatalf("primary code = %q (%v), want %q", got, err, fixture.Expected.PrimaryCode)
			}
			if got := string(result.Manifest.Decision); got != fixture.Expected.ManifestDecision {
				t.Fatalf("manifest decision = %q, want %q", got, fixture.Expected.ManifestDecision)
			}
			node := requireNode(t, result, fixture.Expected.Path)
			if got := string(node.Class); got != fixture.Expected.Class {
				t.Fatalf("node class = %q, want %q", got, fixture.Expected.Class)
			}
			if got := string(node.Decision); got != fixture.Expected.NodeDecision {
				t.Fatalf("node decision = %q, want %q", got, fixture.Expected.NodeDecision)
			}
			if (result.Admission != nil) != fixture.Expected.Authorization {
				t.Fatalf("authorization present = %t, want %t", result.Admission != nil, fixture.Expected.Authorization)
			}
			if !fixture.Expected.Authorization {
				assertNoAuthorization(t, result)
			}
			decoded, decodeErr := DecodeManifest(result.CanonicalBytes)
			if decodeErr != nil {
				t.Fatalf("decode canonical golden: %v", decodeErr)
			}
			if !reflect.DeepEqual(decoded, result.Manifest) {
				t.Fatal("canonical golden does not round-trip exactly")
			}
			if fixture.Expected.ManifestDigest == "" {
				t.Errorf("missing golden digest: %q: %q", fixture.Key(), result.Manifest.ManifestDigest)
			} else if result.Manifest.ManifestDigest != fixture.Expected.ManifestDigest {
				t.Fatalf("manifest digest = %q, want %q", result.Manifest.ManifestDigest, fixture.Expected.ManifestDigest)
			}
		})
	}

	if _, ok := results["F14/z-first"]; ok {
		assertF14CaptureAndOrderIdentity(t, results["F14/z-first"], results["F14/a-first"])
	}
	for _, format := range []string{"zip", "tar", "ar"} {
		firstKey := "F02/duplicate-" + format + "-source-first"
		secondKey := "F02/duplicate-" + format + "-compiled-first"
		if _, ok := results[firstKey]; ok {
			assertF14CaptureAndOrderIdentity(t, results[firstKey], results[secondKey])
		}
	}
	if _, ok := results["C12"]; ok {
		assertC12SharedLeafIdentity(t, results)
	}
}

func executeConformanceCase(t *testing.T, fixture policyconformance.Case) (Result, error) {
	t.Helper()
	payload := append([]byte(nil), fixture.Payload...)
	descriptor := fixtureDescriptor(payload, ProfileID(fixture.Profile))
	descriptor.AdapterID = fixture.AdapterID
	if fixture.Scenario == policyconformance.OriginMissing {
		descriptor.Origin = OriginEvidence{}
	}
	if len(fixture.Uses) > 0 {
		descriptor.ResolvedUses = map[string][]UseEdge{fixture.Path: make([]UseEdge, len(fixture.Uses))}
		for index, use := range fixture.Uses {
			descriptor.ResolvedUses[fixture.Path][index] = UseEdge{Kind: UseKind(use.Kind), Origin: use.Origin}
		}
	}
	payloadRecord := Payload{Path: fixture.Path, Size: int64(len(payload)), Reader: bytes.NewReader(payload)}
	if fixture.PayloadSizeOverride > 0 {
		payloadRecord.Size = fixture.PayloadSizeOverride
	}
	service := NewService()
	switch fixture.Scenario {
	case policyconformance.Dependency, policyconformance.OriginMissing:
		return service.AdmitDependency(t.Context(), DependencyRequest{Descriptor: descriptor, Payload: payloadRecord})
	case policyconformance.ReaderFailure:
		payloadRecord.Reader = repeatReaderError(payload[:4], io.ErrUnexpectedEOF)
		return service.AdmitDependency(t.Context(), DependencyRequest{Descriptor: descriptor, Payload: payloadRecord})
	case policyconformance.Cancellation:
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		return service.AdmitDependency(ctx, DependencyRequest{Descriptor: descriptor, Payload: payloadRecord})
	case policyconformance.ToolchainAllowed:
		return service.AdmitToolchain(t.Context(), ToolchainRequest{
			Descriptor: descriptor, Payload: payloadRecord,
			Authorization: validToolchainAuthorization(t, fixture.Path, payload),
		})
	case policyconformance.ToolchainMissing:
		return service.AdmitToolchain(t.Context(), ToolchainRequest{Descriptor: descriptor, Payload: payloadRecord})
	case policyconformance.ToolchainDrift:
		record := validToolchainAuthorization(t, fixture.Path, payload).artifactPolicyToolchainAuthorization()
		record.timeOfUseFingerprintSHA256 = digestBytes([]byte("drifted-toolchain-tree"))
		return service.AdmitToolchain(t.Context(), ToolchainRequest{
			Descriptor: descriptor, Payload: payloadRecord,
			Authorization: sealedToolchainAuthorization{record: record},
		})
	case policyconformance.ToolchainLinkUnsafe, policyconformance.ToolchainSpecialUnsafe:
		record := validToolchainAuthorization(t, fixture.Path, payload).artifactPolicyToolchainAuthorization()
		if fixture.Scenario == policyconformance.ToolchainLinkUnsafe {
			record.containedLinksValidated = false
		} else {
			record.ordinaryNodesValidated = false
		}
		return service.AdmitToolchain(t.Context(), ToolchainRequest{
			Descriptor: descriptor, Payload: payloadRecord,
			Authorization: sealedToolchainAuthorization{record: record},
		})
	case policyconformance.OutputAllowed:
		return service.AdmitLocalOutput(t.Context(), LocalOutputRequest{
			Descriptor: descriptor, Payload: payloadRecord,
			Authorization: validLocalOutputAuthorization(t, fixture.Path, payload, ArtifactClass(fixture.Expected.Class)),
		})
	case policyconformance.OutputMissing, policyconformance.OutputPreexisting, policyconformance.OutputHardlink:
		// T04 uses the same production API available to an adapter. Caller-held
		// bytes, whether copied or hard-linked, carry no opaque protected-
		// executor receipt and therefore cannot obtain local-output authority.
		return service.AdmitLocalOutput(t.Context(), LocalOutputRequest{Descriptor: descriptor, Payload: payloadRecord})
	case policyconformance.OutputDrift:
		authorizationPayload := fixture.AuthorizationPayload
		if len(authorizationPayload) == 0 {
			t.Fatal("output-drift fixture lacks authorization payload")
		}
		authorizationPath := fixture.AuthorizationPath
		if authorizationPath == "" {
			authorizationPath = fixture.Path
		}
		return service.AdmitLocalOutput(t.Context(), LocalOutputRequest{
			Descriptor: descriptor, Payload: payloadRecord,
			Authorization: validLocalOutputAuthorization(
				t, authorizationPath, authorizationPayload, ArtifactClass(fixture.Expected.Class),
			),
		})
	case policyconformance.OutputInputDrift:
		record := validLocalOutputAuthorization(t, fixture.Path, payload, ArtifactClass(fixture.Expected.Class)).artifactPolicyLocalOutputAuthorization()
		record.completeInputMatched = false
		return service.AdmitLocalOutput(t.Context(), LocalOutputRequest{
			Descriptor: descriptor, Payload: payloadRecord,
			Authorization: sealedLocalOutputAuthorization{record: record},
		})
	case policyconformance.VerifiedUnavailable:
		return service.AdmitVerifiedBinary(t.Context(), VerifiedBinaryRequest{Descriptor: descriptor, Payload: payloadRecord})
	default:
		t.Fatalf("unsupported conformance scenario %q", fixture.Scenario)
		return Result{}, nil
	}
}

func assertC12SharedLeafIdentity(t *testing.T, results map[string]Result) {
	t.Helper()
	keys := []string{"C12", "C12/rust-v1", "C12/node-v1", "C12/swiftpm-v1", "C12/python-reference-v1"}
	want := requireNode(t, results[keys[0]], "renamed.dat")
	for _, key := range keys[1:] {
		got := requireNode(t, results[key], "renamed.dat")
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s shared leaf evidence differs:\ngot=%+v\nwant=%+v", key, got, want)
		}
	}
}

func assertF14CaptureAndOrderIdentity(t *testing.T, first, second Result) {
	t.Helper()
	if first.Manifest.Decision != second.Manifest.Decision ||
		ErrorCode(policyErrorForResult(first)) != ErrorCode(policyErrorForResult(second)) {
		t.Fatal("F14 physical order changed decision or primary diagnostic")
	}
	if bytes.Equal(first.CanonicalBytes, second.CanonicalBytes) || first.Manifest.ManifestDigest == second.Manifest.ManifestDigest {
		t.Fatal("F14 distinct immutable archive bytes aliased one artifact-manifest identity")
	}
	firstProjection := archiveOrderProjection(first.Manifest)
	secondProjection := archiveOrderProjection(second.Manifest)
	firstBytes, firstErr := marshalCanonicalStruct(firstProjection)
	secondBytes, secondErr := marshalCanonicalStruct(secondProjection)
	if firstErr != nil || secondErr != nil {
		t.Fatalf("encode F14 order projection: first=%v second=%v", firstErr, secondErr)
	}
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatalf("F14 logical manifest changed with archive enumeration order:\nfirst=%s\nsecond=%s", firstBytes, secondBytes)
	}
}

// archiveOrderProjection removes only the fields that must differ when two
// physical archive serializations differ. artifact-manifest-v1 deliberately
// binds those raw bytes; equating the full manifest would violate immutable
// origin binding. Every classification, path, observation, accounting value,
// decision, and finding remains in this exact canonical comparison.
func archiveOrderProjection(manifest Manifest) Manifest {
	manifest.ManifestDigest = ""
	manifest.RawPayload.SHA256 = ""
	manifest.Origin.ChecksumSHA256 = ""
	for index := range manifest.RoleEvidence {
		if manifest.RoleEvidence[index].Key == "origin_checksum_sha256" {
			manifest.RoleEvidence[index].Value = ""
		}
	}
	for index := range manifest.Nodes {
		if manifest.Nodes[index].Parent == "" {
			manifest.Nodes[index].SHA256 = ""
		}
	}
	return manifest
}

func policyErrorForResult(result Result) error {
	if result.Manifest.Decision != DecisionReject || len(result.Manifest.Diagnostics) == 0 {
		return nil
	}
	return &PolicyError{Primary: result.Manifest.Diagnostics[0]}
}

func TestConformanceCorpusBuildersAreDeterministic(t *testing.T) {
	first := policyconformance.Cases()
	second := policyconformance.Cases()
	if !reflect.DeepEqual(first, second) {
		t.Fatal("published conformance byte corpus is nondeterministic")
	}
	for _, fixture := range first {
		if fixture.AdapterID == "" || fixture.Path == "" || fixture.Profile == "" || len(fixture.Payload) == 0 {
			t.Fatalf("incomplete reusable fixture %q", fixture.Key())
		}
	}
	root := "/curator-test-toolchain"
	if runtime.GOOS == "windows" {
		root = `C:\curator-test-toolchain`
	}
	if !filepath.IsAbs(root) {
		t.Fatalf("test toolchain root %q is not absolute", root)
	}
}
