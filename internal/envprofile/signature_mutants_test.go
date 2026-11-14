package envprofile

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/testtoolchain"
)

// Each overlay must reach the named behavioral assertion with exit 1. Build
// errors, unrelated failures, and absent positive controls do not kill a plant.
func TestSignatureVerifierNarrowingMutants(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	pipeFile, pipeBefore, pipeAfter := "signature_pipes.go", "if writeErr != nil {", "if runErr == nil && writeErr != nil {"
	if runtime.GOOS == "windows" {
		pipeFile = "signature_pipe_results.go"
		pipeBefore = "if fatal != nil {"
		pipeAfter = "if fatal != nil && runErr == nil {"
	}
	plants := []struct{ name, file, before, after, run, fails, message, control string }{
		{"exit-is-negative", "signature_readonly.go",
			"func signatureVerifierResult(kind string, output []byte, runErr error) error {\n\tif runErr == nil {", "func signatureVerifierResult(kind string, output []byte, runErr error) error {\n\tif runErr == nil || func() bool { var exit *exec.ExitError; return errors.As(runErr, &exit) }() {",
			"^TestSignatureVerifierProcessFailures$", "TestSignatureVerifierProcessFailures/SSH", "process failure was not safely unavailable", "TestSignatureVerifierProcessFailures/ssh-policy"},
		{"gpg-error-is-negative", "signature_readonly.go",
			"case \"NO_PUBKEY\", \"ERRSIG\", \"FAILURE\", \"ERROR\":", "case \"NO_PUBKEY\", \"ERRSIG\", \"ERROR\":",
			"^TestSignatureVerifierProcessFailures$", "TestSignatureVerifierProcessFailures/gpg-failure-after-badsig", "verifier verdict negative=false", "TestSignatureVerifierProcessFailures/gpg-cryptographic"},
		{"transport-precedence", pipeFile, pipeBefore, pipeAfter,
			"^TestSignatureVerifierProcessFailures$", "TestSignatureVerifierProcessFailures", "input transport did not dominate process exit", "TestSignatureVerifierProcessFailures/ssh-policy"},
		{"output-overflow-verdict", "signature_output.go", "if output.truncated {", "if output.truncated && false {",
			"^TestSignatureVerifierProcessFailures$", "TestSignatureVerifierProcessFailures", "verifier output capture was not bounded", "TestSignatureVerifierProcessFailures/ssh-policy"},
		{"raw-status-hint", "status.go", "cause = signatureDiagnostic([]byte(verifyErr.Error()))", "cause = verifyErr.Error()",
			"^TestReadPuritySignerVerdictsKeepPin$", "TestReadPuritySignerVerdictsKeepPin", "unavailable signer hint was not bounded and escaped", ""},
		{"unknown-path-runtime", "signature_paths.go", `if flavour != "msys" && flavour != "cygwin" {`, `if false && flavour != "msys" && flavour != "cygwin" {`,
			"^TestSignatureWindowsPathConventions$", "TestSignatureWindowsPathConventions", "unknown runtime admitted", "TestSignatureWindowsPathConventions/native"},
		{"pe-import-error", "signature_paths.go", "if readErr != nil || closeErr != nil {", "if (readErr != nil && false) || closeErr != nil {",
			"^TestSignatureWindowsPathConventions$", "TestSignatureWindowsPathConventions/pe-malformed-imports", "malformed PE imports did not produce named refusal", "TestSignatureWindowsPathConventions/pe-native"},
		{"pe-runtime-detection", "signature_paths.go", "file.ImportedSymbols()", "func() ([]string, error) { return nil, nil }()",
			"^TestSignatureWindowsPathConventions$", "TestSignatureWindowsPathConventions/pe-native", "PE runtime=", "TestSignatureWindowsPathConventions/native"},
		{"msys-native-path", "signature_paths.go",
			"return prefix + strings.ToLower(path[:1]) + path[2:], nil", "return path, nil",
			"^TestSignatureWindowsPathConventions$", "TestSignatureWindowsPathConventions/msys", "path=", "TestSignatureWindowsPathConventions/native"},
		{"metadata-is-delivery", "signature_pipe_results.go",
			"delivered[result.input] = delivered[result.input] || result.delivered", "delivered[result.input] = delivered[result.input] || result.delivered || result.connected",
			"^TestSignaturePipeResultControls$", "TestSignaturePipeResultControls/metadata-only", "undelivered input became a verdict", "TestSignaturePipeResultControls/consumed"},
		{"keyring-boundary-ignored", "signature_readonly.go",
			"if err := privatedir.Validate(home); err != nil {", "if err := privatedir.Validate(home); err != nil && false {",
			"^TestReadOnlyGPGKeyringPrivateBoundary$", "TestReadOnlyGPGKeyringPrivateBoundary", "unsafe keyring boundary admitted", ""},
		{"directory-cached-metadata", "read_purity_test.go",
			"return current, closeErr", "_ = current; return info, closeErr",
			"^(TestReadPurityDirectoryHandleWitness|TestReadPurityWitnessControl)$", "TestReadPurityDirectoryHandleWitness", "stale directory enumeration survived handle snapshot", "TestReadPurityWitnessControl"},
		{"directory-mtime-exemption", "read_purity_test.go",
			"!old.info.ModTime().Equal(now.info.ModTime())", "(!old.info.IsDir() && !old.info.ModTime().Equal(now.info.ModTime()))",
			"^TestReadPurityDirectoryHandleWitness$", "TestReadPurityDirectoryHandleWitness", "mtime plant escaped directory witness", "TestReadPurityDirectoryHandleWitness/manager/state/locks/v1/none"},
		{"witness-identity-exemption", "read_purity_test.go",
			"!os.SameFile(old.info, now.info)", "false",
			"^TestReadPurityDirectoryHandleWitness$", "TestReadPurityDirectoryHandleWitness", "identity plant escaped directory witness", "TestReadPurityDirectoryHandleWitness/manager/state/locks/v1/none"},
	}
	if runtime.GOOS == "windows" {
		plants = append(plants, struct{ name, file, before, after, run, fails, message, control string }{
			"single-pipe-instance", "signature_pipes_windows.go", "const signaturePipeInstances = 2", "const signaturePipeInstances = 1",
			"^(TestSignaturePipeMetadataProbeAndDrain|TestSignaturePipeTransportReadOnlyAndCancels)$", "TestSignaturePipeMetadataProbeAndDrain", "metadata probe stole signature input or drain lost bytes", "TestSignaturePipeTransportReadOnlyAndCancels",
		})
	}
	for _, plant := range plants {
		t.Run(plant.name, func(t *testing.T) {
			original := filepath.Join(root, "internal", "envprofile", plant.file)
			source, err := os.ReadFile(original)
			if err != nil {
				t.Fatal(err)
			}
			text := strings.ReplaceAll(string(source), "\r\n", "\n")
			if strings.Count(text, plant.before) != 1 {
				t.Fatal("mutation anchor is not unique")
			}
			dir := t.TempDir()
			replacement := filepath.Join(dir, "replacement.go")
			mutated := strings.Replace(text, plant.before, plant.after, 1)
			// Keep the POSIX prefix calculation live in this deliberate semantic plant.
			if plant.name == "msys-native-path" {
				mutated = strings.Replace(mutated, "return path, nil\n}", "_ = prefix; return path, nil\n}", 1)
			}
			if err := os.WriteFile(replacement, []byte(mutated), 0600); err != nil {
				t.Fatal(err)
			}
			overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{original: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			overlayPath := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(overlayPath, overlay, 0600); err != nil {
				t.Fatal(err)
			}
			command := exec.Command(testtoolchain.Go(t), "test", "-p=1", "-parallel=2", "-overlay", overlayPath, "-count=1", "-timeout=2m", "-v", "-run", plant.run, "./internal/envprofile")
			command.Dir = root
			output, err := command.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: "+plant.fails+" ") || !strings.Contains(string(output), plant.message) || (plant.control != "" && !strings.Contains(string(output), "--- PASS: "+plant.control+" ")) {
				t.Fatalf("mutant survived or lacked intended red/control: %v\n%s", err, output)
			}
			t.Logf("expected-red %s exit=1 with positive control:\n%s", plant.name, output)
		})
	}
}
