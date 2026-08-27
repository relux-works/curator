package crossconformance_test

import (
	"context"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/artifactpolicy"
	policyconformance "github.com/relux-works/curator/internal/artifactpolicy/conformance"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/crossconformance"
	"github.com/relux-works/curator/internal/nodesource"
	"github.com/relux-works/curator/internal/npmsource"
	"github.com/relux-works/curator/internal/pnpmsource"
	"github.com/relux-works/curator/internal/rustsource"
	"github.com/relux-works/curator/internal/swiftpmsource"
	"github.com/relux-works/curator/internal/yarnclassicsource"
	"github.com/relux-works/curator/internal/yarnmodernsource"
)

// pathProfiles binds every delivered path to the exact adapter and profile
// identity it declares to the shared artifact classifier.
var pathProfiles = map[crossconformance.PathID]struct {
	adapter string
	profile artifactpolicy.ProfileID
}{
	crossconformance.PathRust:        {adapter: rustsource.ProfileID, profile: artifactpolicy.ProfileRustV1},
	crossconformance.PathNPM:         {adapter: npmsource.ProfileID, profile: artifactpolicy.ProfileNodeV1},
	crossconformance.PathPNPM:        {adapter: pnpmsource.ProfileID, profile: artifactpolicy.ProfileNodeV1},
	crossconformance.PathYarnClassic: {adapter: yarnclassicsource.ProfileID, profile: artifactpolicy.ProfileNodeV1},
	crossconformance.PathYarnModern:  {adapter: yarnmodernsource.ProfileID, profile: artifactpolicy.ProfileNodeV1},
	crossconformance.PathSwiftPM:     {adapter: swiftpmsource.ProfileID, profile: artifactpolicy.ProfileSwiftPMV1},
}

// TestCrossAdapterConformance is the single integration proof over the already
// accepted adapter contracts. It runs one normative semantic suite against
// every delivered path, runs the published rejection matrix, and then refuses
// an incomplete coverage matrix so that a path which stops proving an
// obligation is a failure rather than a silent gap.
func TestCrossAdapterConformance(t *testing.T) {
	coverage := crossconformance.NewCoverage()
	projections := map[crossconformance.PathID][2]crossconformance.TargetProjection{}
	captureText := map[crossconformance.PathID]string{}
	var rustManifestIDs []string
	rustUnavailableReason := ""

	t.Run("project-every-path", func(t *testing.T) {
		nodeSources := map[crossconformance.PathID]nodeCapture{
			crossconformance.PathNPM:         npmCapture(t),
			crossconformance.PathPNPM:        pnpmCapture(t),
			crossconformance.PathYarnClassic: yarnClassicCapture(t),
			crossconformance.PathYarnModern:  yarnModernCapture(t),
		}
		for path, source := range nodeSources {
			projections[path] = [2]crossconformance.TargetProjection{projectNodePath(t, source, 0), projectNodePath(t, source, 1)}
			captureText[path] = nodeCaptureText(t, source)
		}
		projections[crossconformance.PathSwiftPM] = [2]crossconformance.TargetProjection{projectSwiftPath(t, 0), projectSwiftPath(t, 1)}
		captureText[crossconformance.PathSwiftPM] = swiftCaptureText(t)

		t.Run("rust", func(t *testing.T) {
			target, approved := rustsource.NativeCargoDescriptorAvailable()
			if target != "" && !approved {
				rustUnavailableReason = "no operator-approved Cargo descriptor for native target " + target
				t.Skip(rustUnavailableReason)
			}
			rust := runRustManager(t)
			rustManifestIDs = rust.artifactManifestIDs
			first := projectRustPath(t, 0, rustManifestIDs)
			second := projectRustPath(t, 1, rustManifestIDs)
			// The Rust manager owns the C0 Cargo registration, the pinned vendor
			// transform, and the metadata derivations; the reconciliation seam
			// owns the selection-neutral lock superset and the exact active
			// identity. Both belong to one rust-source-v1 projection.
			first.DerivationReceipts, second.DerivationReceipts = rust.receipts, rust.receipts
			first.ToolIdentities = append(first.ToolIdentities, rust.tools...)
			second.ToolIdentities = append(second.ToolIdentities, rust.tools...)
			projections[crossconformance.PathRust] = [2]crossconformance.TargetProjection{first, second}
			captureText[crossconformance.PathRust] = rustCaptureText(t)
		})
	})
	if t.Failed() {
		t.Fatal("one or more available paths could not be projected; the normative suite cannot run")
	}

	t.Run("normative-suite", func(t *testing.T) {
		for _, path := range crossconformance.DeliveredPaths() {
			path := path
			pair, present := projections[path]
			if !present {
				if path == crossconformance.PathRust && rustUnavailableReason != "" {
					t.Run(string(path), func(t *testing.T) { t.Skip(rustUnavailableReason) })
					continue
				}
				t.Errorf("%s produced no projection", path)
				continue
			}
			t.Run(string(path), func(t *testing.T) {
				for _, projection := range pair {
					if err := crossconformance.CheckSelectionNeutralCapture(projection); err != nil {
						t.Error(err)
					}
					if err := crossconformance.CheckCaptureExcludesToolIdentities(projection, captureText[path]); err != nil {
						t.Error(err)
					}
					if err := crossconformance.CheckBindingOwnsTargetAuthority(projection); err != nil {
						t.Error(err)
					}
					if err := crossconformance.CheckCausalEvidenceChain(projection); err != nil {
						t.Error(err)
					}
				}
				if err := crossconformance.CheckTargetDivergence(pair[0], pair[1]); err != nil {
					t.Error(err)
				}
				if err := crossconformance.CheckDeterministicProjection([]crossconformance.TargetProjection{pair[0], repeatProjection(t, path, pair[0], rustManifestIDs)}); err != nil {
					t.Error(err)
				}
				if t.Failed() {
					return
				}
				for _, obligation := range []crossconformance.Obligation{
					crossconformance.ObligationSelectionNeutralCapture,
					crossconformance.ObligationCaptureStableAcrossTargets,
					crossconformance.ObligationBindingOwnsTargetAuthority,
					crossconformance.ObligationSelectionDivergesPerTarget,
					crossconformance.ObligationDeterministicProjection,
					crossconformance.ObligationCausalEvidenceChain,
				} {
					coverage.RecordObligation(obligation, path)
				}
			})
		}
	})

	t.Run("shared-artifact-admission", func(t *testing.T) {
		proveSharedArtifactAdmission(t, coverage)
	})

	t.Run("rejection-matrix", func(t *testing.T) {
		proveRejectionMatrix(t, coverage, rustUnavailableReason != "")
	})

	t.Run("coverage-is-complete", func(t *testing.T) {
		// This gate only means something when the whole test ran: a filtered
		// -run that skips the proving subtests must fail here rather than
		// report a green integration proof over an empty matrix.
		missing := coverage.MissingObligations()
		if rustUnavailableReason == "" {
			if len(missing) != 0 {
				t.Errorf("obligations never proved (a filtered -run cannot satisfy this gate): %s", strings.Join(missing, ", "))
			}
		} else {
			expected := []string{}
			for _, obligation := range crossconformance.Obligations() {
				if obligation != crossconformance.ObligationSharedArtifactAdmission {
					expected = append(expected, string(obligation)+"/"+string(crossconformance.PathRust))
				}
			}
			sort.Strings(expected)
			if strings.Join(missing, "\n") != strings.Join(expected, "\n") {
				t.Errorf("available-path obligations never proved: got %s, want only unavailable Rust obligations %s", strings.Join(missing, ", "), strings.Join(expected, ", "))
			} else {
				t.Logf("rust manager path unavailable on this host: %s", rustUnavailableReason)
			}
		}
		if uncovered := coverage.UncoveredRejections(); len(uncovered) != 0 {
			t.Errorf("rejection vectors never proved: %s", strings.Join(uncovered, ", "))
		}
		t.Logf("rejection coverage:\n%s", strings.Join(coverage.RejectionPaths(), "\n"))
		t.Logf("delegated to the owning accepted suites:\n%s", strings.Join(coverage.DelegatedRejections(), "\n"))
	})
}

// repeatProjection re-derives the first target's projection from freshly
// captured inputs so determinism is a property of the adapter rather than of
// one retained value.
func repeatProjection(t *testing.T, path crossconformance.PathID, original crossconformance.TargetProjection, rustManifestIDs []string) crossconformance.TargetProjection {
	t.Helper()
	switch path {
	case crossconformance.PathNPM:
		return projectNodePath(t, npmCapture(t), 0)
	case crossconformance.PathPNPM:
		return projectNodePath(t, pnpmCapture(t), 0)
	case crossconformance.PathYarnClassic:
		return projectNodePath(t, yarnClassicCapture(t), 0)
	case crossconformance.PathYarnModern:
		return projectNodePath(t, yarnModernCapture(t), 0)
	case crossconformance.PathSwiftPM:
		return projectSwiftPath(t, 0)
	case crossconformance.PathRust:
		// The manager evidence is not re-derived here: determinism under test
		// is the reconciliation of the same lock superset and artifact
		// manifest set onto the same target.
		repeated := projectRustPath(t, 0, rustManifestIDs)
		repeated.PlanIdentity = original.PlanIdentity
		return repeated
	default:
		t.Fatalf("no repeat projection for %s", path)
		return crossconformance.TargetProjection{}
	}
}

// nodeCaptureText renders every capture record of a Node path so the suite can
// prove no bound tool or platform identity is spelled inside it.
func nodeCaptureText(t *testing.T, source nodeCapture) string {
	t.Helper()
	return recordText(t, source.capture.Nodes, source.capture.Edges)
}

func swiftCaptureText(t *testing.T) string {
	t.Helper()
	fixture := newSwiftFixture(t, 0, nil)
	capture, err := swiftpmsource.CaptureAndClose(context.Background(), fixture.config, swiftpmsource.Request{Root: fixture.root, Product: "cli", Resolved: swiftLock()})
	if err != nil {
		t.Fatal(err)
	}
	return recordText(t, capture.Records.CaptureNodes, capture.Records.CaptureEdges)
}

func recordText(t *testing.T, nodes []closuregraph.Node, edges []closuregraph.Edge) string {
	t.Helper()
	var builder strings.Builder
	for _, node := range nodes {
		payload, err := node.CanonicalBytes()
		if err != nil {
			t.Fatal(err)
		}
		builder.Write(payload)
		builder.WriteByte('\n')
	}
	for _, edge := range edges {
		payload, err := edge.CanonicalBytes()
		if err != nil {
			t.Fatal(err)
		}
		builder.Write(payload)
		builder.WriteByte('\n')
	}
	return builder.String()
}

// proveSharedArtifactAdmission proves the accepted C12 claim across every
// delivered path: one deny-class dependency payload presented through each
// adapter's own declared adapter and profile identity produces one shared
// class, node decision, manifest decision, primary diagnostic, and leaf digest.
//
// Positive source admission is deliberately not asserted to be identical. The
// accepted policy lets an adapter narrow its allowed source grammars, so a Go
// or Python source fixture is legitimately opaque under the Rust profile. What
// no adapter may do is admit a class the shared policy denies, which is
// exactly what the deny corpus below checks.
func proveSharedArtifactAdmission(t *testing.T, coverage *crossconformance.Coverage) {
	denyCases := []policyconformance.Case{}
	profileCases := map[artifactpolicy.ProfileID][]policyconformance.Case{}
	for _, fixture := range policyconformance.Cases() {
		if fixture.Scenario != policyconformance.Dependency || len(fixture.Uses) != 0 || fixture.AuthorizationPayload != nil {
			continue
		}
		if fixture.Expected.PrimaryCode == "artifact_compiled_dependency_forbidden" && fixture.Path == fixture.Expected.Path {
			denyCases = append(denyCases, fixture)
		}
		if fixture.Expected.PrimaryCode == "" {
			profile := artifactpolicy.ProfileID(fixture.Profile)
			profileCases[profile] = append(profileCases[profile], fixture)
		}
	}
	if len(denyCases) < 20 {
		t.Fatalf("shared deny corpus has %d bare compiled leaves, want the accepted corpus", len(denyCases))
	}
	paths := crossconformance.DeliveredPaths()

	t.Run("deny-classes-are-identical-across-paths", func(t *testing.T) {
		for _, fixture := range denyCases {
			fixture := fixture
			t.Run(fixture.Key(), func(t *testing.T) {
				first := admissionOutcome(t, paths[0], fixture)
				if first.code != fixture.Expected.PrimaryCode || first.class != fixture.Expected.Class {
					t.Fatalf("%s reported %+v, want class %q code %q", paths[0], first, fixture.Expected.Class, fixture.Expected.PrimaryCode)
				}
				for _, path := range paths[1:] {
					if got := admissionOutcome(t, path, fixture); got != first {
						t.Errorf("%s and %s disagree about %s: %+v vs %+v", paths[0], path, fixture.Key(), first, got)
					}
				}
			})
		}
	})

	t.Run("each-profile-admits-exactly-its-own-source-grammars", func(t *testing.T) {
		for _, path := range paths {
			path := path
			profile := pathProfiles[path]
			t.Run(string(path), func(t *testing.T) {
				admitted := profileCases[profile.profile]
				if len(admitted) == 0 {
					t.Fatalf("the accepted corpus publishes no admitted source vector for %s", profile.profile)
				}
				for _, fixture := range admitted {
					got := admissionOutcome(t, path, fixture)
					if got.code != "" || got.class != fixture.Expected.Class || got.nodeDecision != fixture.Expected.NodeDecision {
						t.Errorf("%s did not admit its own %s vector: %+v", path, fixture.Key(), got)
					}
				}
			})
		}
	})

	if !t.Failed() {
		for _, path := range paths {
			coverage.RecordObligation(crossconformance.ObligationSharedArtifactAdmission, path)
		}
	}
}

type admissionResult struct{ class, nodeDecision, manifestDecision, code, leaf string }

func admissionOutcome(t *testing.T, path crossconformance.PathID, fixture policyconformance.Case) admissionResult {
	t.Helper()
	profile := pathProfiles[path]
	result, err := artifactpolicy.NewService().AdmitDependency(context.Background(), artifactpolicy.DependencyRequest{
		Descriptor: crossDescriptor(profile.adapter, profile.profile, string(path), fixture.Payload),
		Payload:    artifactpolicy.Payload{Path: fixture.Path, Size: int64(len(fixture.Payload)), Reader: strings.NewReader(string(fixture.Payload))},
	})
	node := findManifestNode(result, fixture.Expected.Path)
	return admissionResult{
		class: string(node.Class), nodeDecision: string(node.Decision),
		manifestDecision: string(result.Manifest.Decision),
		code:             string(artifactpolicy.ErrorCode(err)), leaf: node.SHA256,
	}
}

func findManifestNode(result artifactpolicy.Result, path string) artifactpolicy.ManifestNode {
	for _, node := range result.Manifest.Nodes {
		if node.Path == path {
			return node
		}
	}
	return artifactpolicy.ManifestNode{}
}

// diagnosticCode extracts the stable diagnostic every delivered path is
// required to carry, without depending on human error text.
func diagnosticCode(err error) string {
	if err == nil {
		return ""
	}
	if code := artifactpolicy.ErrorCode(err); code != "" {
		return string(code)
	}
	if code := closuregraph.ErrorCode(err); code != "" {
		return string(code)
	}
	for _, code := range []string{
		string(rustsource.ErrorCode(err)), string(swiftpmsource.ErrorCode(err)),
		nodesource.ErrorCode(err), npmsource.ErrorCode(err), pnpmsource.ErrorCode(err),
		yarnclassicsource.ErrorCode(err), yarnmodernsource.ErrorCode(err),
	} {
		if code != "" {
			return code
		}
	}
	text := err.Error()
	if index := strings.IndexAny(text, ": "); index > 0 {
		candidate := text[:index]
		if isStableCode(candidate) {
			return candidate
		}
	}
	return ""
}

func isStableCode(candidate string) bool {
	if candidate == "" {
		return false
	}
	for _, character := range candidate {
		if (character < 'a' || character > 'z') && character != '_' {
			return false
		}
	}
	return strings.Contains(candidate, "_")
}

// sharedCompiledPayload is the pinned GNU shared object every path must reject
// as a dependency byte, injected into each ecosystem's own package payload.
func sharedCompiledPayload() map[string][]byte {
	return map[string][]byte{"prebuilds/addon.node": policyconformance.GNUSharedObject()}
}

func compiledSwiftPayload() map[string][]byte {
	return map[string][]byte{"vendor-addon.node": policyconformance.GNUSharedObject()}
}

func rustCompiledWorkspace(t *testing.T) string {
	t.Helper()
	workspace := t.TempDir()
	for name, payload := range rustWorkspaceFiles() {
		writeFile(t, filepath.Join(workspace, filepath.FromSlash(name)), []byte(payload))
	}
	writeFile(t, filepath.Join(workspace, "dep", "src", "prebuilt.so"), policyconformance.GNUSharedObject())
	return workspace
}
