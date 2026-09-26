package scriptworker

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/stateread"
)

// TestLoadShimSidecarRejectsMalformed proves the launcher contract gate:
// every malformed sidecar shape refuses fail-closed with package
// influence, and a launcher paired with one exits 1 through the
// production RunShim entry instead of running anything.
func TestLoadShimSidecarRejectsMalformed(t *testing.T) {
	root := mustPhysical(t, t.TempDir())
	entry := filepath.Join(root, "runtime", "tool")
	writeTestFile(t, entry, []byte("print('tool')\n"), 0o644)
	runtimeDir := filepath.Join(root, "runtime")
	binDir := mustPhysical(t, t.TempDir())
	launcher := filepath.Join(binDir, "tool"+exeSuffix())
	copyTestFile(t, mustPhysical(t, os.Args[0]), launcher)
	stub := mustPhysical(t, stubInterpreterBinary(t))
	shape := func(mutate func(map[string]any)) string {
		t.Helper()
		object := map[string]any{
			"version": 1, "skill": "skill", "command": "tool",
			"interpreter": "python3-v1", "runtime_entry": entry,
			"runtime_dir": runtimeDir, "project_root": "",
			"schema_version": 8, "capabilities": map[string]any{},
		}
		mutate(object)
		payload, err := json.Marshal(object)
		if err != nil {
			t.Fatal(err)
		}
		return string(payload)
	}
	cases := []struct {
		name    string
		payload string
	}{
		{"unknown-field", shape(func(object map[string]any) { object["extra"] = true })},
		{"version-2", shape(func(object map[string]any) { object["version"] = 2 })},
		{"bash-interpreter", shape(func(object map[string]any) { object["interpreter"] = "bash-v1" })},
		{"relative-runtime-entry", shape(func(object map[string]any) { object["runtime_entry"] = "scripts/tool" })},
		{"schema-7", shape(func(object map[string]any) { object["schema_version"] = 7 })},
		{"exec-escape", shape(func(object map[string]any) { object["capabilities"] = map[string]any{"exec": []string{"../evil"}} })},
		{"garbage", "{not json"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			path := filepath.Join(root, testCase.name+".curator-shim.json")
			writeTestFile(t, path, []byte(testCase.payload), 0o644)
			// Drive every malformed shape through the launcher entry too:
			// a direct parser test cannot prove the shipped launcher calls
			// this gate before starting its worker.
			writeTestFile(t, launcher+ShimSidecarSuffix, []byte(testCase.payload), 0o644)
			var stdout, stderr bytes.Buffer
			if code := RunShim(ShimRequest{
				ExePath: launcher, Stdout: &stdout, Stderr: &stderr,
				Environ: []string{}, Interpreters: map[string]string{"python3-v1": stub},
			}); code != 1 {
				t.Fatalf("RunShim exit = %d, want refusal for %s (stdout=%q stderr=%q)", code, testCase.name, stdout.String(), stderr.String())
			}
			if !strings.Contains(stderr.String(), CodePackageInfluenceForbidden) {
				t.Fatalf("RunShim stderr %q does not name %s for %s", stderr.String(), CodePackageInfluenceForbidden, testCase.name)
			}
			if stdout.Len() != 0 {
				t.Fatalf("RunShim wrote %q to stdout for refused sidecar %s", stdout.String(), testCase.name)
			}
			if _, err := LoadShimSidecar(path); DiagnosticCode(err) != CodePackageInfluenceForbidden {
				t.Fatalf("LoadShimSidecar error = %v, want %s", err, CodePackageInfluenceForbidden)
			}
		})
	}
}

func TestLoadShimSidecarRefusesUnreadablePath(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "not-a-directory")
	writeTestFile(t, blocker, []byte("blocker"), 0o600)
	path := filepath.Join(blocker, "sidecar"+ShimSidecarSuffix)
	_, err := LoadShimSidecar(path)
	var stateErr *stateread.Error
	if DiagnosticCode(err) != CodeWorkerProtocolInvalid || !errors.As(err, &stateErr) ||
		stateErr.Kind != stateread.KindUnreadable || stateErr.Path != path {
		t.Fatalf("LoadShimSidecar error = %v, want worker-protocol refusal with typed unreadable path", err)
	}

	executable := filepath.Join(blocker, "curator")
	var stderr bytes.Buffer
	if code := RunShim(ShimRequest{ExePath: executable, Stderr: &stderr}); code != 1 {
		t.Fatalf("RunShim exit = %d, want refusal for unreadable sidecar (stderr=%q)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "cannot inspect the enforced launcher sidecar") {
		t.Fatalf("RunShim stderr = %q, want the unreadable-sidecar diagnostic", stderr.String())
	}
}

// TestRunShimWritesDiagnosticsRecord proves the operator-selected
// diagnostic destination: with DiagnosticsDir set, a successful launcher
// invocation persists exactly one result-only document carrying the
// invocation's evidence record and derivation report — and never the
// command's own output.
func TestRunShimWritesDiagnosticsRecord(t *testing.T) {
	stub := mustPhysical(t, stubInterpreterBinary(t))
	root := mustPhysical(t, t.TempDir())
	entry := filepath.Join(root, "runtime", "tool")
	writeTestFile(t, entry, []byte("print('tool')\n"), 0o644)
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	launcher := filepath.Join(binDir, "tool"+exeSuffix())
	copyTestFile(t, mustPhysical(t, os.Args[0]), launcher)
	sidecar, err := NewShimSidecar("skill", "tool", "python3-v1",
		entry, filepath.Join(root, "runtime"), "", 8,
		json.RawMessage(`{"env_read": ["STUB_STDERR", "STUB_EXIT"]}`))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := sidecar.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, launcher+ShimSidecarSuffix, payload, 0o644)
	diagnostics := filepath.Join(root, "diagnostics")
	const canary = "CANARY-9f8e7d6c-stdout-stderr"
	var stdout, stderr bytes.Buffer
	code := RunShim(ShimRequest{
		ExePath: launcher, Args: []string{"hello"},
		Stdin: strings.NewReader(""), Stdout: &stdout, Stderr: &stderr,
		Environ:        []string{"STUB_STDERR=" + canary, "STUB_EXIT=0"},
		Interpreters:   map[string]string{"python3-v1": stub},
		DiagnosticsDir: diagnostics,
	})
	if code != 0 {
		t.Fatalf("RunShim exit = %d, stderr %q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), canary) {
		t.Fatalf("the canary did not reach the launcher stderr: %q", stderr.String())
	}
	recordPath := filepath.Join(diagnostics, "skill-tool."+ScriptEvidenceVersion+".json")
	recordPayload, err := os.ReadFile(recordPath)
	if err != nil {
		t.Fatalf("the diagnostics destination carries no record: %v", err)
	}
	if strings.Contains(string(recordPayload), canary) {
		t.Fatalf("the result-only record carries command output: %s", recordPayload)
	}
	var record invocationRecord
	if err := json.Unmarshal(recordPayload, &record); err != nil {
		t.Fatalf("the record is not JSON: %v", err)
	}
	assertValidScriptEvidence(t, record.Evidence, ScriptInventoryPlatform(runtime.GOOS))
	if len(record.Evidence.Controls) != len(scriptInventoryOrder) {
		t.Fatalf("the record carries %d entries, want one per inventory control", len(record.Evidence.Controls))
	}
	if record.Derivation.NetworkMode != networkModeOffline {
		t.Fatalf("derivation network mode = %q, want offline", record.Derivation.NetworkMode)
	}
	for _, withheld := range record.Derivation.WithheldEnv {
		if withheld == "STUB_STDERR" || withheld == "STUB_EXIT" {
			t.Fatalf("derivation withholds %q, want it passed through", withheld)
		}
	}
	// Exactly one document per command: a second invocation replaces the
	// record rather than adding one.
	entries, err := os.ReadDir(diagnostics)
	if err != nil || len(entries) != 1 {
		t.Fatalf("diagnostics entries = %v, %v; want exactly the one record", entries, err)
	}
	var secondStdout, secondStderr bytes.Buffer
	secondCode := RunShim(ShimRequest{
		ExePath: launcher, Stdout: &secondStdout, Stderr: &secondStderr,
		Environ:        []string{"STUB_EXIT=0"},
		Interpreters:   map[string]string{"python3-v1": stub},
		DiagnosticsDir: diagnostics,
	})
	if secondCode != 0 {
		t.Fatalf("second RunShim exit = %d, stderr %q", secondCode, secondStderr.String())
	}
	secondEntries, err := os.ReadDir(diagnostics)
	if err != nil || len(secondEntries) != 1 {
		t.Fatalf("diagnostics entries after a second invocation = %v, %v; want still one record", secondEntries, err)
	}
}

// TestRunShimToleratesUnusableDiagnosticsDir proves a reporting failure
// never fails the invocation: when the destination cannot be created the
// command still runs and its exit status is preserved.
func TestRunShimToleratesUnusableDiagnosticsDir(t *testing.T) {
	stub := mustPhysical(t, stubInterpreterBinary(t))
	root := mustPhysical(t, t.TempDir())
	entry := filepath.Join(root, "runtime", "tool")
	writeTestFile(t, entry, []byte("print('tool')\n"), 0o644)
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	launcher := filepath.Join(binDir, "tool"+exeSuffix())
	copyTestFile(t, mustPhysical(t, os.Args[0]), launcher)
	sidecar, err := NewShimSidecar("skill", "tool", "python3-v1",
		entry, filepath.Join(root, "runtime"), "", 8, json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := sidecar.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, launcher+ShimSidecarSuffix, payload, 0o644)
	blocker := filepath.Join(root, "blocker")
	writeTestFile(t, blocker, []byte("not a directory"), 0o644)
	var stdout, stderr bytes.Buffer
	code := RunShim(ShimRequest{
		ExePath: launcher, Stdout: &stdout, Stderr: &stderr,
		Environ:        []string{},
		Interpreters:   map[string]string{"python3-v1": stub},
		DiagnosticsDir: filepath.Join(blocker, "diagnostics"),
	})
	if code != 0 {
		t.Fatalf("RunShim exit = %d with an unusable diagnostics dir, want 0: %q", code, stderr.String())
	}
	if stdout.Len() == 0 {
		t.Fatal("the invocation produced no output")
	}
}
