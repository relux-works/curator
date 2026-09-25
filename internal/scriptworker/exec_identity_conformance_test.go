package scriptworker

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

//go:embed testdata/executable_identity_cases.json
var pinnedExecutableIdentityFixture embed.FS

const pinnedExecutableIdentityFixtureSHA256 = "sha256:124e00757b3add8c2ba639a8403cba97eec7f5d1bdd923d14b51db9a104292a2"

type pinnedExecutableIdentityCase struct {
	Accepted       bool    `json:"accepted"`
	AdditionalLink string  `json:"additional_links"`
	Interpreter    *string `json:"interpreter"`
	Name           string  `json:"name"`
	Platform       string  `json:"platform"`
	PlatformOwned  bool    `json:"platform_owned"`
	Reason         string  `json:"reason"`
	Resolution     string  `json:"resolution"`
	SystemRoot     *string `json:"system_root"`
	Target         string  `json:"target"`
	Use            string  `json:"use"`
}

// TestUncapturedSystemRootHardlinkRegression proves an ambient SYSTEMROOT
// cannot authorize the System32 hard-link exception. The exhaustive family
// driver accounts for the other published cases, including known gaps.
func TestUncapturedSystemRootHardlinkRegression(t *testing.T) {
	payload, err := pinnedExecutableIdentityFixture.ReadFile("testdata/executable_identity_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(payload)
	if got := "sha256:" + hex.EncodeToString(digest[:]); got != pinnedExecutableIdentityFixtureSHA256 {
		t.Fatalf("executable identity fixture digest = %s, want %s", got, pinnedExecutableIdentityFixtureSHA256)
	}
	var cases []pinnedExecutableIdentityCase
	if err := json.Unmarshal(payload, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 8 {
		t.Fatalf("executable identity case count = %d, want 8", len(cases))
	}
	var target *pinnedExecutableIdentityCase
	for i := range cases {
		if cases[i].Name != "windows-exec-uncaptured-systemroot-hardlinks" {
			continue
		}
		if target != nil {
			t.Fatal("pinned identity fixture repeats the uncaptured-SystemRoot case")
		}
		target = &cases[i]
	}
	if target == nil {
		t.Fatal("pinned identity fixture omits the uncaptured-SystemRoot case")
	}
	if target.Platform != "windows" || target.Use != "declared-exec-name" ||
		target.Accepted || target.SystemRoot == nil || *target.SystemRoot != "caller-or-package-value" {
		t.Fatalf("uncaptured-SystemRoot case has unexpected contract: %+v", *target)
	}
	if drivePinnedDeclaredExecIdentityCase(t, *target) {
		t.Fatalf("production resolver accepted uncaptured SystemRoot hard-link case (reason %s)", target.Reason)
	}
}

func drivePinnedDeclaredExecIdentityCase(t *testing.T, testCase pinnedExecutableIdentityCase) bool {
	t.Helper()
	fixture := newLaunchFixture(t, os.Args[0])
	root := mustPhysical(t, t.TempDir())
	system32 := filepath.Join(root, "System32")
	if err := os.MkdirAll(system32, 0o755); err != nil {
		t.Fatal(err)
	}
	var target string
	switch testCase.Target {
	case "physically-below-canonical-systemroot-system32", "physically-below-system32-from-uncaptured-value":
		target = filepath.Join(system32, "cmd.exe")
	case "outside-canonical-systemroot-system32":
		target = filepath.Join(root, "cmd.exe")
	default:
		t.Fatalf("unsupported executable target %q", testCase.Target)
	}
	copyTestFile(t, fixture.stub, target)

	var aliasDir string
	switch testCase.AdditionalLink {
	case "systemroot-winsxs-component-store-only":
		aliasDir = filepath.Join(root, "WinSxS")
	case "not-platform-component-store-or-unknown":
		aliasDir = filepath.Join(root, "OtherLinks")
	default:
		t.Fatalf("unsupported executable additional_links %q", testCase.AdditionalLink)
	}
	if err := os.MkdirAll(aliasDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(target, filepath.Join(aliasDir, "cmd.exe")); err != nil {
		t.Fatalf("create executable hard-link fixture: %v", err)
	}

	private, err := createPrivateArea(fixture.request.PrivateBase, fixture.request.ForbiddenRoots)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(private.base) }()
	interpreter, err := ResolveInterpreter("python3-v1", fixture.request.Interpreters, fixture.request.ForbiddenRoots)
	if err != nil {
		t.Fatal(err)
	}
	declared, err := ParseDeclaredCapabilities(json.RawMessage(`{"exec":["cmd.exe"]}`))
	if err != nil {
		t.Fatal(err)
	}
	var managerEnvironment []string
	if testCase.SystemRoot != nil && *testCase.SystemRoot == "manager-captured" {
		managerEnvironment = []string{"SYSTEMROOT=" + root}
	}
	if testCase.SystemRoot != nil && *testCase.SystemRoot == "caller-or-package-value" {
		// A nil manager snapshot uses ambient SYSTEMROOT for search, but it
		// must not turn that value into captured System32 authority.
		t.Setenv("SYSTEMROOT", root)
	}
	var searchDirs []string
	if testCase.Resolution == "caller-path-or-package-controlled" {
		searchDirs = []string{system32}
	} else if testCase.Resolution != "manager-default-windows-search-list" {
		t.Fatalf("unsupported exec resolution %q", testCase.Resolution)
	}
	profile, err := deriveProfileForPlatform(DerivationInput{
		Declared: declared, InterpreterID: "python3-v1", Interpreter: interpreter,
		HostEnv: managerEnvironment, ExecSearchDirs: searchDirs,
		ForbiddenRoots: fixture.request.ForbiddenRoots,
		Private:        private, FarmParent: private.base,
	}, "windows")
	if err != nil {
		t.Fatalf("derive Windows profile: %v", err)
	}
	_, resolved := profile.Report.ResolvedExec["cmd.exe"]
	return resolved
}
