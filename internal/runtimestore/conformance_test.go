package runtimestore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
		LauncherCases []struct {
			Name                  string   `json:"name"`
			Platforms             []string `json:"platforms"`
			ForwardArguments      bool     `json:"forward_arguments"`
			PreserveExitStatus    bool     `json:"preserve_exit_status"`
			PreserveInheritedPath bool     `json:"preserve_inherited_path"`
			RequiredPathRoles     []string `json:"required_path_roles"`
		} `json:"launcher_cases"`
	}
	if err := json.Unmarshal(payload, &document); err != nil {
		t.Fatal(err)
	}
	wantCases := map[string]bool{
		"skill-command-without-shell-activation":  false,
		"declared-system-command-without-profile": false,
	}
	for _, testCase := range document.LauncherCases {
		if _, relevant := wantCases[testCase.Name]; !relevant {
			continue
		}
		wantCases[testCase.Name] = true
		if !testCase.ForwardArguments || !testCase.PreserveExitStatus || !testCase.PreserveInheritedPath ||
			strings.Join(testCase.Platforms, ",") != "unix,windows" ||
			strings.Join(testCase.RequiredPathRoles, ",") != "command_directory,implementation_runtime,system_dependencies" {
			t.Fatalf("candidate launcher contract changed: %+v", testCase)
		}
	}
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
