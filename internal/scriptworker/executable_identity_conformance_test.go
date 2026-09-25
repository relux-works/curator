package scriptworker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
)

type executableIdentityVector struct {
	Cases []executableIdentityCase `json:"executable_identity_cases"`
}

type executableIdentityCase struct {
	Name            string  `json:"name"`
	Accepted        bool    `json:"accepted"`
	AdditionalLinks string  `json:"additional_links"`
	Platform        string  `json:"platform"`
	PlatformOwned   bool    `json:"platform_owned"`
	Reason          string  `json:"reason"`
	Resolution      string  `json:"resolution"`
	SystemRoot      *string `json:"system_root"`
	Target          string  `json:"target"`
	Use             string  `json:"use"`
	Interpreter     *string `json:"interpreter"`
}

func loadExecutableIdentityVector(t *testing.T) executableIdentityVector {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "script-host-execution-policy.json")) // #nosec G304 -- explicit conformance input
	if err != nil {
		t.Fatal(err)
	}
	var vector executableIdentityVector
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatal(err)
	}
	return vector
}

// TestExecutableIdentityCasesAtProductionEntry drives the dcc7f015 identity
// vectors through ResolveInterpreter and deriveProfileForPlatform, the same
// resolver used to build the manager-owned exec PATH farm. The platform is
// injected as Windows so the filesystem cases run on all three CI hosts.
func TestExecutableIdentityCasesAtProductionEntry(t *testing.T) {
	vector := loadExecutableIdentityVector(t)
	if len(vector.Cases) != 8 {
		t.Fatalf("executable_identity_cases has %d cases, want 8", len(vector.Cases))
	}
	conformancecoverage.RunOutcomes(t, "script-host-execution-policy/executable-identity-cases", vector.Cases,
		func(testCase executableIdentityCase) string { return testCase.Name },
		func(t *testing.T, testCase executableIdentityCase) conformancecoverage.Observation {
			if testCase.Platform != "windows" {
				t.Fatalf("case platform = %q, want windows", testCase.Platform)
			}
			switch testCase.Use {
			case "interpreter":
				if testCase.Interpreter == nil {
					t.Fatal("interpreter identity case has no interpreter id")
				}
				fixture := newLaunchFixture(t, os.Args[0])
				interpreterPath := fixture.request.Interpreters[*testCase.Interpreter]
				if interpreterPath == "" {
					t.Fatalf("fixture has no configured interpreter %q", *testCase.Interpreter)
				}
				alias := filepath.Join(t.TempDir(), filepath.Base(interpreterPath)+"-alias")
				if err := os.Link(interpreterPath, alias); err != nil {
					t.Fatalf("create interpreter hard-link fixture: %v", err)
				}
				_, err := ResolveInterpreter(*testCase.Interpreter,
					fixture.request.Interpreters, fixture.request.ForbiddenRoots)
				if testCase.Accepted {
					if err != nil {
						return conformancecoverage.Observation{FailureReason: "ResolveInterpreter refused an accepted identity: " + err.Error()}
					}
				} else if DiagnosticCode(err) != CodeWorkerIdentityInvalid {
					return conformancecoverage.Observation{FailureReason: "ResolveInterpreter did not refuse interpreter hard-link substitution"}
				}
			case "declared-exec-name":
				if failure := driveDeclaredExecIdentityCase(t, testCase); failure != "" {
					return conformancecoverage.Observation{FailureReason: failure}
				}
			default:
				t.Fatalf("unsupported identity case use %q", testCase.Use)
			}
			return conformancecoverage.Observation{}
		})
}

func driveDeclaredExecIdentityCase(t *testing.T, testCase executableIdentityCase) string {
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
	switch testCase.AdditionalLinks {
	case "systemroot-winsxs-component-store-only":
		aliasDir = filepath.Join(root, "WinSxS")
	case "not-platform-component-store-or-unknown":
		aliasDir = filepath.Join(root, "OtherLinks")
	default:
		t.Fatalf("unsupported executable additional_links %q", testCase.AdditionalLinks)
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
	var managerEnv []string
	if testCase.SystemRoot != nil && *testCase.SystemRoot == "manager-captured" {
		managerEnv = []string{"SYSTEMROOT=" + root}
	}
	if testCase.SystemRoot != nil && *testCase.SystemRoot == "caller-or-package-value" {
		// This value is available to ordinary process search but is not part of
		// the manager-captured snapshot that grants the System32 exception.
		t.Setenv("SYSTEMROOT", root)
	}
	var searchDirs []string
	if testCase.Resolution == "caller-path-or-package-controlled" {
		searchDirs = []string{system32}
	}
	profile, err := deriveProfileForPlatform(DerivationInput{
		Declared: declared, InterpreterID: "python3-v1", Interpreter: interpreter,
		HostEnv: managerEnv, ExecSearchDirs: searchDirs,
		ForbiddenRoots: fixture.request.ForbiddenRoots,
		Private:        private, FarmParent: private.base,
	}, "windows")
	if err != nil {
		t.Fatalf("derive Windows profile: %v", err)
	}
	_, resolved := profile.Report.ResolvedExec["cmd.exe"]
	if resolved != testCase.Accepted {
		return "production resolver accepted = " + strconv.FormatBool(resolved) +
			", want " + strconv.FormatBool(testCase.Accepted) +
			" (reason " + testCase.Reason + "; platform_owned=" + strconv.FormatBool(testCase.PlatformOwned) + ")"
	}
	return ""
}
