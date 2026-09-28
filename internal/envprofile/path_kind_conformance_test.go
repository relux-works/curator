package envprofile

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/pathboundary"
)

const pathKindAdmissionVector = "environments-path-kind-admission.json"

type pathKindAdmissionVectors struct {
	MCPKindCases []pathMCPKindVector  `json:"mcp_kind_cases"`
	PathBoundary []pathBoundaryVector `json:"path_boundary_cases"`
	DryRun       []pathBoundaryVector `json:"dry_run_cases"`
}

type pathMCPKindVector struct {
	Name             string `json:"name"`
	Admitted         bool   `json:"admitted"`
	Conforming       *bool  `json:"conforming"`
	Diagnostic       string `json:"diagnostic"`
	Declaration      string `json:"declaration"`
	NamesDeclaration string `json:"names_declaration"`
	NamesPackage     string `json:"names_package"`
	Origin           string `json:"origin"`
	Package          string `json:"package"`
	SourceKind       string `json:"source_kind"`
}

type pathBoundaryVector struct {
	Conforming    *bool  `json:"conforming"`
	Name          string `json:"name"`
	Diagnostic    string `json:"diagnostic"`
	FailingCheck  string `json:"failing_check"`
	NamesCheck    string `json:"names_check"`
	Module        string `json:"module"`
	NamesModule   string `json:"names_module"`
	NamesPackage  string `json:"names_package"`
	Origin        string `json:"origin"`
	Package       string `json:"package"`
	Role          string `json:"role"`
	CarriesSystem bool   `json:"carries_system_modules"`
	Direct        bool   `json:"direct"`
	Rebuild       bool   `json:"rebuild_planned"`
	Current       bool   `json:"row_current"`
	Fragment      bool   `json:"fragment_emitted"`
	Outcome       string `json:"outcome"`
}

func readPathKindAdmissionVectors(t *testing.T) pathKindAdmissionVectors {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", pathKindAdmissionVector)) // #nosec G304 -- explicit pinned conformance input
	if err != nil {
		t.Fatal(err)
	}
	var vectors pathKindAdmissionVectors
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	return vectors
}

// TestPathKindMCPVectorsDriveInstallEntry executes the complete pinned MCP
// family through Install. The one git-positive row keeps the existing Git
// MCP admission path, while every path row, including the published bad
// admission shape, must refuse and name both package and declaration.
func TestPathKindMCPVectorsDriveInstallEntry(t *testing.T) {
	vectors := readPathKindAdmissionVectors(t)
	conformancecoverage.RunOutcomes(t, "environments-path-kind-admission/mcp_kind_cases", vectors.MCPKindCases,
		func(tc pathMCPKindVector) string { return tc.Name },
		func(caseT *testing.T, tc pathMCPKindVector) conformancecoverage.Observation {
			if tc.SourceKind == "git" {
				if !tc.Admitted || tc.Diagnostic != "" {
					caseT.Fatalf("Git MCP vector %q does not describe admission: %+v", tc.Name, tc)
				}
				driveGitMCPAdmission(caseT)
				return conformancecoverage.Observation{}
			}
			nonconformingAdmission := tc.Conforming != nil && !*tc.Conforming && tc.Admitted
			if tc.SourceKind != "path" ||
				(tc.Diagnostic != "mcp_declaration_path_source_refused" && !nonconformingAdmission) {
				caseT.Fatalf("unclassified MCP vector %q: %+v", tc.Name, tc)
			}
			drivePathMCPRefusal(caseT, tc)
			return conformancecoverage.Observation{}
		})
}

func driveGitMCPAdmission(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	pinHomes(t)
	ids := newGitIdentities(t)
	tool := gitRepo(t, map[string]string{
		"agent-mcp.json": pathMCPManifest,
	}, "v1.2.0")
	toolURL := ids.serve(tool, "https://example.com/mcp-figma-devmode")
	root := gitRepo(t, map[string]string{
		"agent-context.json": `{"schema_version":1,"name":"figma-root","version":"1.0.0","requires":{"mcp":{"figma-devmode":{"git":"` + toolURL + `","range":"*"}}}}` + "\n",
	}, "v1.0.0")
	operand := ids.serve(root, "https://example.com/figma-root")
	if _, _, _, err := Install(home, InstallOptions{Operand: operand}); err != nil {
		t.Fatalf("Git MCP package should remain admissible: %v", err)
	}
}

func drivePathMCPRefusal(t *testing.T, tc pathMCPKindVector) {
	t.Helper()
	home := t.TempDir()
	pinHomes(t)
	pathPackage := filepath.Join(t.TempDir(), "path-package")
	pathMCPPackage(t, pathPackage)
	var err error
	switch tc.Origin {
	case "root", "onboarding-import":
		_, _, _, err = Install(home, InstallOptions{Operand: pathPackage, Imported: tc.Origin == "onboarding-import"})
	case "overlay":
		root := filepath.Join(t.TempDir(), "root")
		writeManifestPackage(t, root,
			`{"schema_version":1,"name":"team-context","version":"1.0.0"}`+"\n", nil)
		policy := Policy{OverlaysAllowed: true, OverlayDefaultWeight: 1000,
			Overlays: map[string][]OverlaySpec{"team-context": {{Source: pathPackage}}}}
		_, _, _, err = Install(home, InstallOptions{Operand: root, Policy: policy})
	default:
		t.Fatalf("MCP vector %q has unknown path origin %q", tc.Name, tc.Origin)
	}
	wantDiagnostic := tc.Diagnostic
	if tc.Conforming != nil && !*tc.Conforming && tc.Admitted {
		wantDiagnostic = DiagMCPPathSourceRefused
	}
	wantPackage, wantDeclaration := tc.NamesPackage, tc.NamesDeclaration
	if wantPackage == "" {
		wantPackage = tc.Package
	}
	if wantDeclaration == "" {
		wantDeclaration = tc.Declaration
	}
	if err == nil || !strings.Contains(err.Error(), wantDiagnostic) ||
		!strings.Contains(err.Error(), wantPackage) || !strings.Contains(err.Error(), wantDeclaration) {
		t.Fatalf("Install err = %v, want %s naming package %q and declaration %q", err, wantDiagnostic, wantPackage, wantDeclaration)
	}
}

// TestPathKindBoundaryVectorsDriveProductionEntries executes every §4
// path-boundary vector through Install, Import, Resolve, Update or Status.
// Foreign-owner rows use the boundary validator's injected owner lookup;
// only a host that cannot create a required special file records a bound.
func TestPathKindBoundaryVectorsDriveProductionEntries(t *testing.T) {
	vectors := readPathKindAdmissionVectors(t)
	conformancecoverage.RunOutcomes(t, "environments-path-kind-admission/path_boundary_cases", vectors.PathBoundary,
		func(tc pathBoundaryVector) string { return tc.Name },
		func(caseT *testing.T, tc pathBoundaryVector) conformancecoverage.Observation {
			if reason := drivePathBoundaryVector(caseT, tc); reason != "" {
				caseT.Logf("bounded: %s", reason)
				return conformancecoverage.Observation{BoundReason: reason}
			}
			return conformancecoverage.Observation{}
		})
}

func drivePathBoundaryVector(t *testing.T, tc pathBoundaryVector) string {
	t.Helper()
	switch tc.Name {
	case "path-overlay-system-module-admitted":
		home, root, overlay, policy := newVectorPathOverlay(t, true)
		installVectorPathOverlay(t, home, root, policy)
		result := assertPathResolveCurrent(t, home, "team-context", policy, tc)
		var fragment struct {
			SystemPrompt *struct {
				Path string `json:"path"`
			} `json:"system_prompt"`
		}
		if err := json.Unmarshal(result.Document, &fragment); err != nil {
			t.Fatalf("decode emitted fragment: %v", err)
		}
		if fragment.SystemPrompt == nil || fragment.SystemPrompt.Path == "" {
			t.Fatalf("trusted direct system overlay emitted no system_prompt section: %s", result.Document)
		}
		payload, err := os.ReadFile(fragment.SystemPrompt.Path)
		if err != nil || string(payload) != "personal context\n" {
			t.Fatalf("system prompt at %q = %q, err=%v; want admitted module bytes", fragment.SystemPrompt.Path, payload, err)
		}
		if overlay == "" || !tc.Fragment || !tc.Current || tc.Diagnostic != "" || tc.Module != "90-system.md" {
			t.Fatalf("trusted direct system overlay did not reach its admitted resolve row: vector=%+v result=%s", tc, result.Document)
		}
		return ""
	case "path-root-no-system-modules-admitted":
		home := t.TempDir()
		pinHomes(t)
		root := filepath.Join(t.TempDir(), "root")
		writeManifestPackage(t, root, `{"schema_version":1,"name":"team-context","version":"1.0.0"}`+"\n", nil)
		if _, _, _, err := Install(home, InstallOptions{Operand: root}); err != nil {
			t.Fatalf("Install path root: %v", err)
		}
		assertPathResolveCurrent(t, home, "team-context", Policy{}, tc)
		return ""
	case "path-import-no-system-modules-admitted":
		home := t.TempDir()
		pinHomes(t)
		seams := pinImportSeams(t)
		seedCurrentDefault(t, home)
		if _, _, _, err := Import(home, seams.options(Policy{})); err != nil {
			t.Fatalf("Import path root: %v", err)
		}
		source, err := readSource(home, "imported")
		if err != nil || source.Path != filepath.Join(home, "imports", "imported") {
			t.Fatalf("Import source=%+v err=%v, want durable imports path", source, err)
		}
		assertPathResolveCurrent(t, home, "imported", Policy{}, tc)
		return ""
	case "path-overlay-no-system-modules-admitted":
		home, root, _, policy := newVectorPathOverlay(t, false)
		installVectorPathOverlay(t, home, root, policy)
		assertPathResolveCurrent(t, home, "team-context", policy, tc)
		return ""
	case "path-overlay-world-writable-untrusted", "path-overlay-no-system-world-writable-untrusted":
		return driveUntrustedOverlayInstall(t, tc, false)
	case "path-overlay-symlinked-component-untrusted":
		home, root, overlay, policy := newVectorPathOverlay(t, tc.CarriesSystem)
		if err := os.Symlink(filepath.Join(overlay, "context", "personal.md"), filepath.Join(overlay, "context", "alias.md")); err != nil {
			return "host cannot create the symlink needed by the link_safety vector: " + err.Error()
		}
		return assertOverlayInstallDiagnostic(t, home, root, policy, tc)
	case "path-overlay-containment-escape-untrusted":
		home, root, overlay, policy := newVectorPathOverlay(t, tc.CarriesSystem)
		outside := filepath.Join(t.TempDir(), "outside.md")
		if err := os.WriteFile(outside, []byte("outside\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(overlay, "context", "alias.md")); err != nil {
			return "host cannot create the symlink needed by the containment vector: " + err.Error()
		}
		return assertOverlayInstallDiagnostic(t, home, root, policy, tc)
	case "path-overlay-wrong-ownership-untrusted", "path-import-no-system-wrong-ownership-untrusted":
		return driveWrongOwnerVector(t, tc)
	case "path-overlay-untrusted-rebuilds":
		return drivePathOverlayUntrustedRepair(t, tc)
	case "path-overlay-non-regular-component-untrusted":
		return driveNonRegularVector(t, tc)
	case "path-transitive-system-module-refused":
		drivePathRootTransitiveSystemRefusal(t, tc)
		return ""
	case "path-overlay-untrusted-reported-current":
		return driveUntrustedOverlayInstall(t, tc, true)
	default:
		t.Fatalf("unclassified path-boundary vector %q", tc.Name)
		return ""
	}
}

func newVectorPathOverlay(t *testing.T, system bool) (string, string, string, Policy) {
	t.Helper()
	home := t.TempDir()
	pinHomes(t)
	root := filepath.Join(t.TempDir(), "root")
	writeManifestPackage(t, root, `{"schema_version":1,"name":"team-context","version":"1.0.0"}`+"\n", nil)
	class := ""
	if system {
		class = "system"
	}
	overlay := filepath.Join(t.TempDir(), "personal")
	module := `{"path":"personal.md"}`
	if class != "" {
		module = `{"path":"personal.md","class":"` + class + `"}`
	}
	writeManifestPackage(t, overlay,
		`{"schema_version":1,"name":"personal","version":"1.0.0","context":{"modules":[`+module+`]}}`+"\n",
		map[string]string{"personal.md": "personal context\n"})
	policy := Policy{OverlaysAllowed: true, OverlayDefaultWeight: 1000,
		Overlays: map[string][]OverlaySpec{"team-context": {{Source: overlay}}}}
	return home, root, overlay, policy
}

func installVectorPathOverlay(t *testing.T, home, root string, policy Policy) {
	t.Helper()
	if _, _, _, err := Install(home, InstallOptions{Operand: root, Policy: policy}); err != nil {
		t.Fatalf("Install trusted path overlay: %v", err)
	}
}

func resolveVectorPath(t *testing.T, home, profile string, policy Policy) *ResolveResult {
	t.Helper()
	result, err := Resolve(ResolveRequest{
		Home: home, Profile: profile, EnvID: "claude_code", Machine: envregistry.DefaultMachineConfig(),
		Policy: policy, LaunchDir: t.TempDir(), Repair: true, Format: "json",
		Detect:       func(envregistry.Adapter) string { return "unknown" },
		NativeHomeOf: func(id string) (string, error) { return filepath.Join(home, "native", id), nil },
		OperatorXDG:  t.TempDir(),
	})
	if err != nil {
		t.Fatalf("Resolve path source: %v", err)
	}
	return result
}

func assertPathResolveCurrent(t *testing.T, home, profile string, policy Policy, tc pathBoundaryVector) *ResolveResult {
	t.Helper()
	result := resolveVectorPath(t, home, profile, policy)
	if len(result.Document) == 0 || !tc.Fragment || !tc.Current || tc.Diagnostic != "" {
		t.Fatalf("trusted path source did not emit a current fragment: vector=%+v result=%s", tc, result.Document)
	}
	status, err := StatusOf(StatusRequest{
		Home: home, Machine: envregistry.DefaultMachineConfig(), Policy: policy, LaunchDir: t.TempDir(),
		Detect:       func(envregistry.Adapter) string { return "unknown" },
		NativeHomeOf: func(id string) (string, error) { return filepath.Join(home, "native", id), nil },
		OperatorXDG:  filepath.Join(home, "xdg"),
	})
	if err != nil {
		t.Fatalf("StatusOf trusted path source: %v", err)
	}
	for _, row := range status.Homes {
		if row.Profile == profile && row.Environment == "claude_code" {
			if !row.Current {
				t.Fatalf("StatusOf reports %s/claude_code non-current: %v", profile, row.Findings)
			}
			return result
		}
	}
	t.Fatalf("StatusOf has no %s/claude_code row", profile)
	return nil
}

func assertOverlayInstallDiagnostic(t *testing.T, home, root string, policy Policy, tc pathBoundaryVector) string {
	t.Helper()
	_, _, _, err := Install(home, InstallOptions{Operand: root, Policy: policy})
	if err == nil || !strings.Contains(err.Error(), tc.Diagnostic) || !strings.Contains(err.Error(), tc.NamesCheck) {
		t.Fatalf("Install err = %v, want %s naming %s", err, tc.Diagnostic, tc.NamesCheck)
	}
	return ""
}

func driveUntrustedOverlayInstall(t *testing.T, tc pathBoundaryVector, checkStatus bool) string {
	t.Helper()
	home, root, overlay, policy := newVectorPathOverlay(t, tc.CarriesSystem)
	if !checkStatus {
		restore, err := makeWorldWritableDirectoryForTest(overlay)
		if err != nil {
			t.Fatalf("create world-writable path boundary: %v", err)
			return ""
		}
		defer func() {
			if err := restore(); err != nil {
				t.Errorf("restore path boundary fixture: %v", err)
			}
		}()
		if err := validatePathDirectory(overlay); err == nil || !strings.Contains(err.Error(), "permissions") {
			t.Fatalf("world-writable path source validation = %v, want permissions refusal", err)
		}
		return assertOverlayInstallDiagnostic(t, home, root, policy, tc)
	}
	installVectorPathOverlay(t, home, root, policy)
	restore, err := makeWorldWritableDirectoryForTest(overlay)
	if err != nil {
		t.Fatalf("create world-writable path boundary: %v", err)
		return ""
	}
	defer func() {
		if err := restore(); err != nil {
			t.Errorf("restore path boundary fixture: %v", err)
		}
	}()
	if err := validatePathDirectory(overlay); err == nil || !strings.Contains(err.Error(), "permissions") {
		t.Fatalf("world-writable path source validation = %v, want permissions refusal", err)
	}
	request := ResolveRequest{Home: home, Profile: "team-context", EnvID: "claude_code", LaunchDir: t.TempDir(),
		Machine: envregistry.DefaultMachineConfig(), Policy: policy}
	result, err := Resolve(request)
	if err == nil || result != nil || !strings.Contains(err.Error(), tc.Diagnostic) || !strings.Contains(err.Error(), tc.NamesCheck) {
		t.Fatalf("Resolve = (%v, %v), want no fragment and %s naming %s", result, err, tc.Diagnostic, tc.NamesCheck)
	}
	status, err := StatusOf(StatusRequest{Home: home, Machine: envregistry.DefaultMachineConfig(), Policy: policy})
	if err != nil {
		t.Fatal(err)
	}
	if !status.NonCurrent {
		t.Fatal("Status reports a boundary-invalid overlay current")
	}
	for _, row := range status.Homes {
		if row.Profile == "team-context" && !row.Current && strings.Join(row.Findings, " ") != "" &&
			strings.Contains(strings.Join(row.Findings, " "), tc.Diagnostic) && strings.Contains(strings.Join(row.Findings, " "), tc.NamesCheck) {
			return ""
		}
	}
	t.Fatal("Status row did not report the path boundary failure")
	return ""
}

func driveWrongOwnerVector(t *testing.T, tc pathBoundaryVector) string {
	t.Helper()
	if tc.Origin == "onboarding-import" {
		home := t.TempDir()
		pinHomes(t)
		seams := pinImportSeams(t)
		seedCurrentDefault(t, home)
		info, _, _, err := Import(home, seams.options(Policy{}))
		if err != nil {
			t.Fatalf("Import: %v", err)
		}
		undo, reason := forceWrongOwnerForTest(info.Source.Path)
		if reason != "" {
			return reason
		}
		defer undo()
		return assertResolveUntrusted(t, home, info.Name, Policy{}, tc)
	}
	home, root, overlay, policy := newVectorPathOverlay(t, tc.CarriesSystem)
	undo, reason := forceWrongOwnerForTest(overlay)
	if reason != "" {
		return reason
	}
	defer undo()
	return assertOverlayInstallDiagnostic(t, home, root, policy, tc)
}

func drivePathOverlayUntrustedRepair(t *testing.T, tc pathBoundaryVector) string {
	t.Helper()
	if tc.Conforming == nil || *tc.Conforming || !tc.Rebuild || tc.FailingCheck != pathboundary.CheckOwnership {
		t.Fatalf("vector %q is not the pinned non-conforming rebuild probe: %+v", tc.Name, tc)
	}
	home, root, overlay, policy := newVectorPathOverlay(t, tc.CarriesSystem)
	installVectorPathOverlay(t, home, root, policy)
	undo, reason := forceWrongOwnerForTest(overlay)
	if reason != "" {
		return reason
	}
	defer undo()
	before := hashTreeForTest(t, home)
	result, err := Resolve(ResolveRequest{
		Home: home, Profile: "team-context", EnvID: "claude_code", Machine: envregistry.DefaultMachineConfig(),
		Policy: policy, LaunchDir: t.TempDir(), Repair: true,
		Detect:       func(envregistry.Adapter) string { return "unknown" },
		NativeHomeOf: func(id string) (string, error) { return filepath.Join(home, "native", id), nil },
		OperatorXDG:  filepath.Join(home, "xdg"),
	})
	if err == nil || result != nil || !strings.Contains(err.Error(), tc.Diagnostic) ||
		!strings.Contains(err.Error(), tc.NamesCheck) || strings.Contains(err.Error(), "would-rebuild-untrusted-store") {
		t.Fatalf("repair Resolve = (%v, %v), want no fragment, no rebuild plan, and %s naming %s", result, err, tc.Diagnostic, tc.NamesCheck)
	}
	if after := hashTreeForTest(t, home); after != before {
		t.Fatal("repair Resolve mutated manager state before refusing the untrusted path source")
	}
	return ""
}

func assertResolveUntrusted(t *testing.T, home, profile string, policy Policy, tc pathBoundaryVector) string {
	t.Helper()
	result, err := Resolve(ResolveRequest{Home: home, Profile: profile, EnvID: "claude_code", LaunchDir: t.TempDir(),
		Machine: envregistry.DefaultMachineConfig(), Policy: policy})
	if err == nil || result != nil || !strings.Contains(err.Error(), tc.Diagnostic) || !strings.Contains(err.Error(), tc.NamesCheck) {
		t.Fatalf("Resolve = (%v, %v), want %s naming %s", result, err, tc.Diagnostic, tc.NamesCheck)
	}
	return ""
}

func driveNonRegularVector(t *testing.T, tc pathBoundaryVector) string {
	t.Helper()
	home, root, overlay, policy := newVectorPathOverlay(t, tc.CarriesSystem)
	if err := makeFIFOVectorForTest(filepath.Join(overlay, "context", "pipe")); err != nil {
		return "host cannot create the non-regular-file vector: " + err.Error()
	}
	return assertOverlayInstallDiagnostic(t, home, root, policy, tc)
}

func drivePathRootTransitiveSystemRefusal(t *testing.T, tc pathBoundaryVector) {
	t.Helper()
	home := t.TempDir()
	pinHomes(t)
	ids := newGitIdentities(t)
	leaf := gitRepo(t, map[string]string{
		"agent-context.json":   `{"schema_version":1,"name":"leaf","version":"1.0.0","context":{"modules":[{"path":"90-system.md","class":"system"}]}}` + "\n",
		"context/90-system.md": "leaf system prompt\n",
	}, "v1.0.0")
	leafURL := ids.serve(leaf, "https://example.com/path-leaf")
	mid := gitRepo(t, map[string]string{
		"agent-context.json": `{"schema_version":1,"name":"mid","version":"1.0.0","requires":{"contexts":{"leaf":{"git":"` + leafURL + `","range":"*"}}}}` + "\n",
	}, "v1.0.0")
	midURL := ids.serve(mid, "https://example.com/path-mid")
	root := filepath.Join(t.TempDir(), "root")
	writeManifestPackage(t, root,
		`{"schema_version":1,"name":"team-context","version":"1.0.0","requires":{"contexts":{"mid":{"git":"`+midURL+`","range":"*"}}}}`+"\n", nil)
	policy := Policy{TransitiveSystemModules: contextmaterialize.TransitiveError}
	_, _, _, err := Install(home, InstallOptions{Operand: root, Policy: policy})
	if err == nil || !strings.Contains(err.Error(), tc.Diagnostic) || !strings.Contains(err.Error(), tc.NamesPackage) || !strings.Contains(err.Error(), tc.NamesModule) {
		t.Fatalf("Install err = %v, want %s naming package %s and module %s", err, tc.Diagnostic, tc.NamesPackage, tc.NamesModule)
	}
}

// TestPathKindDryRunVectorsDriveReadOnlyResolve verifies dry-run-equivalent
// lock-free Resolve behaviour for intact and untrusted path overlays: it
// never emits a fragment for an untrusted source, plans no rebuild, and
// leaves manager and source trees byte-identical.
func TestPathKindDryRunVectorsDriveReadOnlyResolve(t *testing.T) {
	vectors := readPathKindAdmissionVectors(t)
	conformancecoverage.RunOutcomes(t, "environments-path-kind-admission/dry_run_cases", vectors.DryRun,
		func(tc pathBoundaryVector) string { return tc.Name },
		func(caseT *testing.T, tc pathBoundaryVector) conformancecoverage.Observation {
			if reason := drivePathDryRunVector(caseT, tc); reason != "" {
				return conformancecoverage.Observation{BoundReason: reason}
			}
			return conformancecoverage.Observation{}
		})
}

func drivePathDryRunVector(t *testing.T, tc pathBoundaryVector) string {
	t.Helper()
	home, root, overlay, policy := newVectorPathOverlay(t, tc.CarriesSystem)
	installVectorPathOverlay(t, home, root, policy)
	var restore func() error
	if tc.FailingCheck != "" {
		var err error
		restore, err = makeWorldWritableDirectoryForTest(overlay)
		if err != nil {
			t.Fatalf("create world-writable path boundary: %v", err)
		}
		if err := validatePathDirectory(overlay); err == nil || !strings.Contains(err.Error(), "permissions") {
			t.Fatalf("world-writable path source validation = %v, want permissions refusal", err)
		}
		defer func() {
			if err := restore(); err != nil {
				t.Errorf("restore path boundary fixture: %v", err)
			}
		}()
	}
	beforeHome := hashTreeForTest(t, home)
	beforeSource := hashTreeForTest(t, overlay)
	result, err := Resolve(ResolveRequest{Home: home, Profile: "team-context", EnvID: "claude_code", LaunchDir: t.TempDir(),
		Machine: envregistry.DefaultMachineConfig(), Policy: policy, Repair: false})
	if tc.Diagnostic != "" {
		if err == nil || result != nil || !strings.Contains(err.Error(), tc.Diagnostic) || !strings.Contains(err.Error(), tc.FailingCheck) ||
			strings.Contains(err.Error(), "would-rebuild-untrusted-store") {
			t.Fatalf("Resolve=(%v,%v), want no fragment, %s/%s and no rebuild plan", result, err, tc.Diagnostic, tc.FailingCheck)
		}
	} else if err != nil && !strings.Contains(err.Error(), DiagHomeStale) {
		t.Fatalf("intact read-only Resolve: %v", err)
	}
	if after := hashTreeForTest(t, home); beforeHome != after {
		t.Fatal("read-only Resolve mutated manager state")
	}
	if after := hashTreeForTest(t, overlay); beforeSource != after {
		t.Fatal("read-only Resolve mutated the path source")
	}
	return ""
}

func hashTreeForTest(t *testing.T, root string) string {
	t.Helper()
	var paths []string
	if err := filepath.WalkDir(root, func(path string, _ os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		paths = append(paths, path)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	sort.Strings(paths)
	h := sha256.New()
	for _, path := range paths {
		info, err := os.Lstat(path)
		if err != nil {
			t.Fatal(err)
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = fmt.Fprint(h, relative, "\x00", info.Mode().String(), "\x00")
		if info.Mode().IsRegular() {
			payload, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			_, _ = h.Write(payload)
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}
