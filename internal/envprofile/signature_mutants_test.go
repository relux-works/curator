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
		pipeFile = "signature_pipes_windows.go"
		pipeBefore = "if writeErr != nil || processCloseErr != nil {"
		pipeAfter = "if runErr == nil && (writeErr != nil || processCloseErr != nil) {"
	}
	for _, plant := range []struct{ name, file, before, after, run, fails, message, control string }{
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
		{"msys-native-path", "signature_paths.go",
			"return prefix + strings.ToLower(path[:1]) + path[2:], nil", "return path, nil",
			"^TestSignatureWindowsPathConventions$", "TestSignatureWindowsPathConventions/msys", "path=", "TestSignatureWindowsPathConventions/native"},
	} {
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
