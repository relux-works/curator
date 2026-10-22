package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/testtoolchain"
)

const cliPathChild = "CURATOR_TEST_CLI_PATH_CHILD"
const cliPathPlant = "CURATOR_TEST_CLI_PATH_PLANT"

// The actual test executable starts with planted ambient providers. Its
// TestMain must remove them while the unchanged legacy/draft status positives
// and the explicit provider acceptance/refusal cases still pass.
func TestCLIPathIsolation(t *testing.T) {
	home := t.TempDir()
	planted := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(planted, 0o755); err != nil {
		t.Fatal(err)
	}
	// Match the reported failure class: a user-bin directory where Curator
	// has published shims, rather than revision A's permitted warning-only
	// provider directory. Both the home and ownership ledger are disposable.
	if err := os.WriteFile(filepath.Join(planted, ".curator-managed.json"), []byte(`{"schema_version":1,"entries":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"curator-engines", "curator-run"} {
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		if err := os.WriteFile(filepath.Join(planted, name), []byte("ambient provider plant\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	child, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(child, "-test.v", "-test.run=^(TestCLIPathIsolationChild|TestCLIStatusReportsLegacyPackageMarkerCurrent|TestDraftStatusReportsUpToDateThroughCLI|TestCuratorStatusProviderPostureAndCheck)$")
	command.Env = append(os.Environ(), cliPathChild+"=1", cliPathPlant+"="+planted,
		"HOME="+home, "USERPROFILE="+home, "CURATOR_GLOBAL_USER_BIN=",
		"PATH="+planted+string(os.PathListSeparator)+os.Getenv("PATH"))
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("planted ambient provider suite: %v\n%s", err, output)
	}
	for _, name := range []string{"TestCLIPathIsolationChild", "TestCLIStatusReportsLegacyPackageMarkerCurrent", "TestDraftStatusReportsUpToDateThroughCLI", "TestCuratorStatusProviderPostureAndCheck"} {
		if !strings.Contains(string(output), "--- PASS: "+name+" ") {
			t.Fatalf("planted suite did not execute %s:\n%s", name, output)
		}
	}
}

// The ambient plant is hidden from default lookups, but explicitly passing
// it through the CLI helper still exposes the real untrusted-provider refusal.
func TestCLIPathIsolationChild(t *testing.T) {
	planted := os.Getenv(cliPathPlant)
	if os.Getenv(cliPathChild) != "1" || planted == "" {
		t.Skip("only runs in the planted PATH subprocess")
	}
	for _, name := range []string{"curator-engines", "curator-run"} {
		if path, err := exec.LookPath(name); err == nil || !errors.Is(err, exec.ErrNotFound) {
			t.Errorf("ambient provider %s leaked through TestMain: %q, %v", name, path, err)
		}
	}
	_, home := legacyProject(t)
	if code, _, stderr := capture(t, filepath.Join(home, "config.json"), "install", "app"); code != exitOK {
		t.Fatalf("install = %d\n%s", code, stderr)
	}
	code, stdout, stderr := captureWithEnv(t, filepath.Join(home, "config.json"), map[string]string{
		"PATH": statusProviderPath(t, planted+string(os.PathListSeparator)+os.Getenv("PATH")),
	}, "status", "app", "--json", "--check")
	if code != exitFail {
		t.Fatalf("explicit ambient plant = %d, want status refusal\n%s\n%s", code, stdout, stderr)
	}
	row := findProviderState(t, decodeStatus(t, stdout).Providers, "engines")
	if row.Current || row.Diagnostic == nil || *row.Diagnostic != "subcommand_provider_untrusted" {
		t.Fatalf("explicit plant did not reach product trust rules: %+v", row)
	}
}

// Appending the ambient PATH must fail the real unchanged-install status
// witness. A build error, unrelated failure, or skipped plant is not a kill.
func TestCLIPathIsolationNarrowingMutant(t *testing.T) {
	source, err := filepath.Abs("testpath_setup_test.go")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	const before = `return strings.Join(dirs, string(os.PathListSeparator))`
	const after = `return strings.Join(append(dirs, os.Getenv("PATH")), string(os.PathListSeparator))`
	if strings.Count(string(data), before) != 1 {
		t.Fatal("ambient PATH mutation anchor is not unique")
	}
	directory := t.TempDir()
	replacement := filepath.Join(directory, "testpath_test.go")
	if err := os.WriteFile(replacement, []byte(strings.Replace(string(data), before, after, 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	data, err = json.Marshal(map[string]any{"Replace": map[string]string{source: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlay, data, 0o600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(testtoolchain.Go(t), "test", "-overlay", overlay, "-count=1", "-timeout=3m", "-v", "-run", "^TestCLIPathIsolation$", ".")
	command.Env = append(os.Environ(), cliPathChild+"=1")
	output, err := command.CombinedOutput()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 ||
		!strings.Contains(string(output), "--- FAIL: TestCLIStatusReportsLegacyPackageMarkerCurrent") ||
		!strings.Contains(string(output), "status --check = 1, want 0") ||
		!strings.Contains(string(output), "--- FAIL: TestCLIPathIsolationChild") ||
		strings.Contains(string(output), "build failed") {
		t.Fatalf("ambient-appending mutant did not fail the intended status witness: %v\n%s", err, output)
	}
	t.Log("ambient-appending mutant: real legacy status --check fails with exit 1; ambient provider lookup also fails isolation")
}
