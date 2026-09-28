package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

const pathMCPManifest = `{"schema_version":1,"name":"figma-devmode","version":"1.2.0","server":{"transport":"stdio","command":"npx","args":["-y","figma-developer-mcp","--stdio"],"env_names":["FIGMA_API_KEY"],"environments":["claude_code","codex_cli","opencode"]}}` + "\n"

func pathMCPPackage(t *testing.T, root string) {
	t.Helper()
	writeManifestPackage(t, root,
		`{"schema_version":1,"name":"figma-devmode","version":"1.2.0"}`+"\n", nil)
	if err := os.WriteFile(filepath.Join(root, "agent-mcp.json"), []byte(pathMCPManifest), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertPathMCPRefusal(t *testing.T, err error) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), DiagMCPPathSourceRefused) ||
		strings.Count(err.Error(), "figma-devmode") < 2 {
		t.Fatalf("err = %v, want %s naming package and declaration figma-devmode", err, DiagMCPPathSourceRefused)
	}
}

// TestPathRootAndImportedMCPDeclarationsRefused drives the installed path
// source admission used by both an ordinary path root and Import's
// ImportedFromNative path. The refusal names the package and declaration.
// Production entry point: Install, the entry Import delegates to.
func TestPathRootAndImportedMCPDeclarationsRefused(t *testing.T) {
	for _, imported := range []bool{false, true} {
		t.Run(map[bool]string{false: "root", true: "onboarding-import"}[imported], func(t *testing.T) {
			home := t.TempDir()
			pinHomes(t)
			source := filepath.Join(t.TempDir(), "path-mcp")
			pathMCPPackage(t, source)
			_, _, _, err := Install(home, InstallOptions{Operand: source, Imported: imported})
			assertPathMCPRefusal(t, err)
			if _, err := readSource(home, "figma-devmode"); err == nil {
				t.Fatal("refused path package left an installed profile source")
			}
		})
	}
}

// TestPathOverlayMCPDeclarationRefused drives a path overlay through the
// ordinary install pipeline and refuses its MCP declaration before the
// overlay can join the profile closure.
func TestPathOverlayMCPDeclarationRefused(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	root := filepath.Join(t.TempDir(), "root")
	writeManifestPackage(t, root,
		`{"schema_version":1,"name":"team-context","version":"1.0.0"}`+"\n", nil)
	overlay := filepath.Join(t.TempDir(), "path-mcp")
	pathMCPPackage(t, overlay)
	policy := Policy{
		OverlaysAllowed:      true,
		OverlayDefaultWeight: 1000,
		Overlays:             map[string][]OverlaySpec{"team-context": {{Source: overlay}}},
	}
	_, _, _, err := Install(home, InstallOptions{Operand: root, Policy: policy})
	assertPathMCPRefusal(t, err)
}

func pathOverlayFixture(t *testing.T, moduleClass string) (string, string, Policy) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "root")
	writeManifestPackage(t, root,
		`{"schema_version":1,"name":"team-context","version":"1.0.0"}`+"\n", nil)
	overlay := filepath.Join(t.TempDir(), "personal")
	module := `{"path":"personal.md"}`
	if moduleClass != "" {
		module = `{"path":"personal.md","class":"` + moduleClass + `"}`
	}
	writeManifestPackage(t, overlay,
		`{"schema_version":1,"name":"personal","version":"1.0.0","context":{"modules":[`+module+`]}}`+"\n",
		map[string]string{"personal.md": "personal context\n"})
	policy := Policy{
		OverlaysAllowed:      true,
		OverlayDefaultWeight: 1000,
		Overlays:             map[string][]OverlaySpec{"team-context": {{Source: overlay}}},
	}
	return root, overlay, policy
}

// TestPathOverlayBoundaryRejectsEscapingAndInternalLinks proves both
// containment and link_safety at Install's production path-overlay entry.
func TestPathOverlayBoundaryRejectsEscapingAndInternalLinks(t *testing.T) {
	for _, tc := range []struct {
		name   string
		target func(root, overlay string) string
		check  string
	}{
		{name: "escape", target: func(_, _ string) string { return filepath.Join(t.TempDir(), "outside.md") }, check: "containment"},
		{name: "within-tree", target: func(_ string, overlay string) string { return filepath.Join(overlay, "context", "personal.md") }, check: "link_safety"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			pinHomes(t)
			root, overlay, policy := pathOverlayFixture(t, "")
			link := filepath.Join(overlay, "context", "alias.md")
			if err := os.Symlink(tc.target(root, overlay), link); err != nil {
				t.Skipf("host cannot create the symlink vector: %v", err)
			}
			_, _, _, err := Install(home, InstallOptions{Operand: root, Policy: policy})
			if err == nil || !strings.Contains(err.Error(), DiagPathSourceUntrusted) || !strings.Contains(err.Error(), tc.check) {
				t.Fatalf("err = %v, want %s with %s", err, DiagPathSourceUntrusted, tc.check)
			}
		})
	}
}

// TestPathOverlayTrustedSystemModuleRemainsAdmitted follows rc.13 §3 and
// the pinned path-overlay-system-module-admitted vector: direct trusted
// path overlays share the existing system-module admission rule.
func TestPathOverlayTrustedSystemModuleRemainsAdmitted(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	root, _, policy := pathOverlayFixture(t, "system")
	info, _, _, err := Install(home, InstallOptions{Operand: root, Policy: policy})
	if err != nil {
		t.Fatalf("trusted direct system overlay was refused: %v", err)
	}
	found := false
	for _, warning := range info.Warnings {
		if strings.Contains(warning, "context-system-module-present") {
			found = true
		}
	}
	if !found {
		t.Fatalf("direct system overlay produced no system-module finding: warnings=%v", info.Warnings)
	}
}

// TestPathOverlayFailureBlocksResolveAndMakesStatusNonCurrent drives the
// path directory contract after install. The path is inspected again by
// Resolve and Status, no fragment is emitted, and no rebuild is planned.
func TestPathOverlayFailureBlocksResolveAndMakesStatusNonCurrent(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	root, overlay, policy := pathOverlayFixture(t, "")
	if _, _, _, err := Install(home, InstallOptions{Operand: root, Policy: policy}); err != nil {
		t.Fatal(err)
	}
	restore, err := makeWorldWritableDirectoryForTest(overlay)
	if err != nil {
		t.Fatalf("create world-writable permission vector: %v", err)
	}
	defer func() {
		if err := restore(); err != nil {
			t.Errorf("restore path boundary fixture: %v", err)
		}
	}()
	if err := validatePathDirectory(overlay); err == nil || !strings.Contains(err.Error(), "permissions") {
		t.Fatalf("world-writable path source validation = %v, want permissions refusal", err)
	}
	req := ResolveRequest{
		Home: home, Profile: "team-context", EnvID: "claude_code", LaunchDir: t.TempDir(),
		Machine: envregistry.DefaultMachineConfig(), Policy: policy,
		Detect: func(envregistry.Adapter) string { return "unknown" },
	}
	if result, err := Resolve(req); err == nil || result != nil || !strings.Contains(err.Error(), DiagPathSourceUntrusted) || !strings.Contains(err.Error(), "permissions") {
		t.Fatalf("Resolve = (%v, %v), want no fragment and %s/permissions", result, err, DiagPathSourceUntrusted)
	}
	status, err := StatusOf(StatusRequest{Home: home, Machine: envregistry.DefaultMachineConfig(), Policy: policy})
	if err != nil {
		t.Fatal(err)
	}
	if !status.NonCurrent {
		t.Fatal("StatusOf reports a profile current while its path overlay is untrusted")
	}
	for _, row := range status.Homes {
		if row.Profile != "team-context" || row.Current {
			continue
		}
		for _, finding := range row.Findings {
			if strings.Contains(finding, DiagPathSourceUntrusted) && strings.Contains(finding, "permissions") {
				return
			}
		}
	}
	t.Fatal("StatusOf did not report environment_store_untrusted naming the permissions check")
}

func TestPathSourceMutationsRejectUntrustedOverlayBeforeDefaultWrites(t *testing.T) {
	for _, operation := range []string{"update", "use", "sync"} {
		t.Run(operation, func(t *testing.T) {
			home := t.TempDir()
			pinHomes(t)
			root, overlay, policy := pathOverlayFixture(t, "")
			if _, _, _, err := Install(home, InstallOptions{Operand: root, Policy: policy}); err != nil {
				t.Fatal(err)
			}
			if err := os.RemoveAll(ProfileDir(home, DefaultProfile)); err != nil {
				t.Fatal(err)
			}
			if err := SetCurrent(home, "team-context"); err != nil {
				t.Fatal(err)
			}
			undo, reason := forceWrongOwnerForTest(overlay)
			if reason != "" {
				t.Skip(reason)
			}
			defer undo()

			var err error
			switch operation {
			case "update":
				_, _, err = UpdateWithPolicy(home, "team-context", policy)
			case "use":
				_, err = UseWithPolicy(home, "team-context", "", "", false, policy)
			case "sync":
				_, err = SyncWithPolicy(home, policy)
			}
			if err == nil || !strings.Contains(err.Error(), DiagPathSourceUntrusted) || !strings.Contains(err.Error(), "ownership") {
				t.Fatalf("%s error = %v, want %s naming ownership", operation, err, DiagPathSourceUntrusted)
			}
			if _, err := os.Stat(ProfileDir(home, DefaultProfile)); !os.IsNotExist(err) {
				t.Fatalf("%s created the default profile before refusing the untrusted path: stat err=%v", operation, err)
			}
		})
	}
}
