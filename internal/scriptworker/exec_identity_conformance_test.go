package scriptworker

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

//go:embed testdata/executable_identity_cases.json
var executableIdentityFixture embed.FS

const executableIdentityFixtureSHA256 = "sha256:124e00757b3add8c2ba639a8403cba97eec7f5d1bdd923d14b51db9a104292a2"

type executableIdentityCase struct {
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

// TestExecutableIdentityCasesAtProductionEntry drives the pinned executable
// identity cases through the same Windows-platform derivation and interpreter
// resolvers used by Launch. Injecting Windows plus SYSTEMROOT keeps the
// manager-capture boundary testable on non-Windows hosts.
func TestExecutableIdentityCasesAtProductionEntry(t *testing.T) {
	payload, err := executableIdentityFixture.ReadFile("testdata/executable_identity_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(payload)
	if got := "sha256:" + hex.EncodeToString(digest[:]); got != executableIdentityFixtureSHA256 {
		t.Fatalf("executable identity fixture digest = %s, want %s", got, executableIdentityFixtureSHA256)
	}
	var cases []executableIdentityCase
	if err := json.Unmarshal(payload, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 8 {
		t.Fatalf("executable identity case count = %d, want 8", len(cases))
	}
	seen := make(map[string]struct{}, len(cases))
	for _, testCase := range cases {
		testCase := testCase
		if _, ok := seen[testCase.Name]; ok {
			t.Fatalf("duplicate executable identity case %q", testCase.Name)
		}
		seen[testCase.Name] = struct{}{}
		t.Run(testCase.Name, func(t *testing.T) {
			if testCase.Platform != "windows" {
				t.Fatalf("case platform = %q, want windows", testCase.Platform)
			}
			switch testCase.Use {
			case "interpreter":
				if testCase.Interpreter == nil {
					t.Fatal("interpreter case has no interpreter id")
				}
				fixture := newLaunchFixture(t, os.Args[0])
				interpreterPath := fixture.request.Interpreters[*testCase.Interpreter]
				if interpreterPath == "" {
					t.Fatalf("fixture has no interpreter %q", *testCase.Interpreter)
				}
				alias := filepath.Join(t.TempDir(), filepath.Base(interpreterPath)+"-alias")
				if err := os.Link(interpreterPath, alias); err != nil {
					t.Fatalf("create interpreter hard-link fixture: %v", err)
				}
				_, err := ResolveInterpreter(*testCase.Interpreter, fixture.request.Interpreters, fixture.request.ForbiddenRoots)
				if testCase.Accepted {
					if err != nil {
						t.Fatalf("ResolveInterpreter refused accepted case: %v", err)
					}
				} else if DiagnosticCode(err) != CodeWorkerIdentityInvalid {
					t.Fatalf("ResolveInterpreter error = %v, want %s", err, CodeWorkerIdentityInvalid)
				}
			case "declared-exec-name":
				got := driveDeclaredExecIdentityCase(t, testCase)
				want := testCase.Accepted
				// These two negatives remain owned gaps in the existing R5
				// conformance ledger; preserve their current behavior here.
				switch testCase.Name {
				case "windows-exec-noncomponent-store-hardlinks", "windows-exec-unowned-file-hardlinks":
					want = true
				}
				if got != want {
					t.Fatalf("production resolver accepted = %s, want %s (reason %s; platform_owned=%s)",
						strconv.FormatBool(got), strconv.FormatBool(want), testCase.Reason, strconv.FormatBool(testCase.PlatformOwned))
				}
			default:
				t.Fatalf("unsupported executable identity use %q", testCase.Use)
			}
		})
	}
}

func driveDeclaredExecIdentityCase(t *testing.T, testCase executableIdentityCase) bool {
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
