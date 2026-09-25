package runtimestore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
)

func TestCandidateManagerLauncherContract(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "manager-lifecycle.json")) // #nosec G304 -- explicit candidate conformance input
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		LauncherCases []authoritativeLauncherCase `json:"launcher_cases"`
	}
	if err := json.Unmarshal(payload, &document); err != nil {
		t.Fatal(err)
	}
	wantCases := map[string]bool{
		"skill-command-without-shell-activation":  false,
		"declared-system-command-without-profile": false,
	}
	conformancecoverage.RunOutcomes(t, "manager-lifecycle/launcher-cases", document.LauncherCases,
		func(tc authoritativeLauncherCase) string { return tc.Name }, func(t *testing.T, testCase authoritativeLauncherCase) conformancecoverage.Observation {
			if _, relevant := wantCases[testCase.Name]; !relevant {
				return conformancecoverage.Observation{BoundReason: "launcher case is outside this runtime-store contract check"}
			}
			wantCases[testCase.Name] = true
			if !testCase.ForwardArguments || !testCase.PreserveExitStatus || !testCase.PreserveInheritedPath ||
				strings.Join(testCase.Platforms, ",") != "unix,windows" ||
				strings.Join(testCase.RequiredPathRoles, ",") != "command_directory,implementation_runtime,system_dependencies" {
				t.Fatalf("candidate launcher contract changed: %+v", testCase)
			}
			return conformancecoverage.Observation{}
		})
	for name, found := range wantCases {
		if !found {
			t.Fatalf("candidate launcher case %q is absent", name)
		}
	}

	unix := UnixShimContent("/immutable/artifact", []string{"/system/dependency"})
	if !strings.Contains(unix, `exec '/immutable/artifact' "$@"`) || !strings.Contains(unix, `:"$PATH"`) {
		t.Fatalf("Unix launcher no longer satisfies candidate forwarding contract:\n%s", unix)
	}
	windows := WindowsShimContent(`C:\immutable\artifact.exe`, []string{`C:\system\dependency`})
	for _, required := range []string{`call "C:\immutable\artifact.exe" %*`, `;%PATH%`, `exit /b %ERRORLEVEL%`} {
		if !strings.Contains(windows, required) {
			t.Fatalf("Windows launcher lacks %q:\n%s", required, windows)
		}
	}
}
