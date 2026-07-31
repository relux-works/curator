package runtimestore

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const windowsHelperEnv = "CURATOR_WINDOWS_ARTIFACT_HELPER"

func TestMain(m *testing.M) {
	if os.Getenv(windowsHelperEnv) == "1" {
		if marker := os.Getenv("CURATOR_WINDOWS_ARTIFACT_MARKER"); marker != "" {
			_ = os.WriteFile(marker, []byte("launched"), 0o600)
		}
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
			"args": os.Args[1:],
			"path": os.Getenv("PATH"),
		})
		code, _ := strconv.Atoi(os.Getenv("CURATOR_WINDOWS_ARTIFACT_EXIT"))
		os.Exit(code)
	}
	os.Exit(m.Run())
}

// decodeHelperOutput returns the JSON document the helper above writes to stdout
// when a staged wrapper launches it, failing the test with the raw combined
// output when the payload is absent.
func decodeHelperOutput(t *testing.T, output []byte) map[string]any {
	t.Helper()
	decoded, err := parseHelperOutput(output)
	if err != nil {
		t.Fatalf("%v; output:\n%s", err, output)
	}
	return decoded
}

func TestWindowsPostInstallWrappersForwardArgumentsPathAndExitCode(t *testing.T) {
	root := t.TempDir()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	artifactDir := filepath.Join(root, "immutable cache % Юникод")
	artifact := filepath.Join(artifactDir, "tool.exe")
	if err := os.MkdirAll(artifactDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := copyFile(executable, artifact); err != nil {
		t.Fatal(err)
	}
	compiled := compiledTargetFixture(t, artifact, "windows")
	helperPath := filepath.Join(root, "dependency path %PATH% Юникод")
	if err := os.MkdirAll(helperPath, 0o755); err != nil {
		t.Fatal(err)
	}
	destinations := []ManagedShim{
		mustManagedShim(t, ProjectShim, filepath.Join(root, "project bin %"), "tool", "windows"),
		mustManagedShim(t, GlobalCanonicalShim, filepath.Join(root, "global bin %"), "tool", "windows"),
		mustManagedShim(t, SafeForwardingShim, filepath.Join(root, "forward bin %"), "tool", "windows"),
	}
	desired := make([]ShimSpec, 0, len(destinations))
	for _, destination := range destinations {
		desired = append(desired, ShimSpec{Destination: destination, Target: compiled, PathEntries: []string{helperPath}})
	}
	plan, err := StageShimTransition(filepath.Join(root, "private stage"), desired, nil)
	if err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(root, "artifact-launched")
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("artifact launched while wrappers were staged: %v", err)
	}
	for _, target := range plan.Desired {
		installFixtureTarget(t, target)
	}

	comspec := os.Getenv("ComSpec")
	if comspec == "" {
		comspec = filepath.Join(os.Getenv("SystemRoot"), "System32", "cmd.exe")
	}
	for _, target := range plan.Desired {
		caller := filepath.Join(root, "call "+string(target.Role)+".cmd")
		shimPath := strings.ReplaceAll(target.LivePath, "%", "%%")
		content := "@echo off\r\n" +
			"\"" + shimPath + "\" \"space value\" \"quote\\\"value\" \"percent%%PATH%%value\" \"Юникод\" \"\"\r\n" +
			"exit /b %ERRORLEVEL%\r\n"
		if err := os.WriteFile(caller, []byte(content), 0o700); err != nil {
			t.Fatal(err)
		}
		command := exec.Command(comspec, "/d", "/s", "/c", `call "`+strings.ReplaceAll(caller, "%", "%%")+`"`)
		command.Env = []string{
			"PATH=",
			"SystemRoot=" + os.Getenv("SystemRoot"),
			"ComSpec=" + comspec,
			windowsHelperEnv + "=1",
			"CURATOR_WINDOWS_ARTIFACT_MARKER=" + marker,
			"CURATOR_WINDOWS_ARTIFACT_EXIT=37",
		}
		output, runErr := command.CombinedOutput()
		exitErr, ok := runErr.(*exec.ExitError)
		if !ok || exitErr.ExitCode() != 37 {
			t.Fatalf("%s wrapper exit = %v; output:\n%s", target.Role, runErr, output)
		}
		decoded := decodeHelperOutput(t, output)
		gotArgs, ok := decoded["args"].([]any)
		wantArgs := []string{"space value", `quote"value`, "percent%PATH%value", "Юникод", ""}
		if !ok || len(gotArgs) != len(wantArgs) {
			t.Fatalf("%s args = %#v", target.Role, decoded["args"])
		}
		for index, want := range wantArgs {
			if gotArgs[index] != want {
				t.Fatalf("%s arg %d = %#v, want %#v", target.Role, index, gotArgs[index], want)
			}
		}
		if decoded["path"] != helperPath {
			t.Fatalf("%s PATH = %#v, want %#v", target.Role, decoded["path"], helperPath)
		}
	}
}
