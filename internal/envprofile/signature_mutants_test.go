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
		pipeFile, pipeBefore, pipeAfter = "signature_files.go", `return nil, fmt.Errorf("%w: private signature staging: %v", errSignerVerificationUnavailable, err)`, `return nil, fmt.Errorf("private signature staging: %v", err)`
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
		{"pe-runtime-detection", "signature_paths.go", "return signatureRuntimeFlavour(libraries), nil", "_ = libraries; return signatureRuntimeFlavour(nil), nil",
			"^TestSignatureWindowsPathConventions$", "TestSignatureWindowsPathConventions/pe-native", "PE runtime=", "TestSignatureWindowsPathConventions/native"},
		{"msys-native-path", "signature_paths.go",
			"return prefix + strings.ToLower(path[:1]) + path[2:], nil", "return path, nil",
			"^TestSignatureWindowsPathConventions$", "TestSignatureWindowsPathConventions/msys", "path=", "TestSignatureWindowsPathConventions/native"},
		{"ssh-namespace-negative", "signature_readonly.go", `strings.HasPrefix(line, "Couldn't verify signature: namespace ")`, `false`,
			"^TestReadPuritySSHNamespaceVerdict$", "TestReadPuritySSHNamespaceVerdict/not-git", "namespace rejection became unavailable/current", "TestReadPuritySSHNamespaceVerdict/git"},
		{"ssh-negative-status", "signature_readonly.go", `code == 255 || int64(code) == int64(0xffffffff)`, `true`,
			"^TestSignatureVerifierProcessFailures$", "TestSignatureVerifierProcessFailures/ssh-unexpected-status", "unexpected SSH exit became a verdict", ""},
		{"pe-symbol-name-bounds", "signature_pe.go", `if _, err := signaturePEName(file, value+2); err != nil`, `if _, err := signaturePEName(file, value+2); err != nil && false`,
			"^TestSignaturePEImportValidation$", "TestSignaturePEImportValidation/symbol-unterminated", "unsafe PE symbol-unterminated admitted", "TestSignaturePEImportValidation/mixed-ordinal-msys-2.0.dll"},
		{"pe-iat-bounds", "signature_pe.go", `signaturePERead(file, first+index*width, width)`, `signaturePERead(file, lookup+index*width, width)`,
			"^TestSignaturePEImportValidation$", "TestSignaturePEImportValidation/iat-rva", "unsafe PE iat-rva admitted", "TestSignaturePEImportValidation/mixed-ordinal-msys-2.0.dll"},
		{"pe-ordinal-only", "signature_pe.go", `if !sawNamed {`, `if !sawNamed && false {`,
			"^TestSignaturePEImportValidation$", "TestSignaturePEImportValidation/ordinal-only", "unsafe PE ordinal-only admitted", "TestSignaturePEImportValidation/mixed-ordinal-msys-2.0.dll"},
		{"pe-ordinal-runtime", "signature_pe.go", `libraries = append(libraries, library)`, `if named { libraries = append(libraries, library) }`,
			"^TestSignaturePEImportValidation$", "TestSignaturePEImportValidation/mixed-ordinal-msys-2.0.dll", "ordinal runtime lost", ""},
		{"regular-input-compatibility", "signature_files.go", `args[i] = converted`, `args[i] = ` + "`" + `\\.\pipe\curator-unsupported-signature` + "`" + `; _ = converted`,
			"^TestSignatureFileTransport$", "TestSignatureFileTransport/private-roundtrip", "private regular file roundtrip failed", "TestSignatureFileTransport/no-input"},
		{"private-input-creation", "signature_files.go", `privatedir.Create(path, 0600)`, `os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)`,
			"^TestSignatureFileTransport$", "TestSignatureFileTransport/private-roundtrip", "private regular file roundtrip failed", "TestSignatureFileTransport/no-input"},
		{"private-directory-creation", "signature_files.go", `privatedir.Make(path)`, `os.Mkdir(path, 0755)`,
			"^TestSignatureFileTransport$", "TestSignatureFileTransport/private-roundtrip", "private regular file roundtrip failed", "TestSignatureFileTransport/no-input"},

		{"private-input-cleanup", "signature_files.go", `os.RemoveAll(root)`, `func(string) error { return nil }(root)`,
			"^TestSignatureFileTransport$", "TestSignatureFileTransport/private-roundtrip", "signature staging leaked", "TestSignatureFileTransport/no-input"},
		{"staging-entry-witness", "read_purity_witness_test.go", `len(after) != 1`, `false`,
			"^TestSignatureScratchWitness$", "TestSignatureScratchWitness/entry", "staging entry plant escaped", "TestSignatureScratchWitness/transient"},
		{"staging-identity-witness", "read_purity_witness_test.go", `!os.SameFile(old.info, now.info) || old.info.Mode() != now.info.Mode()`, `old.info.Mode() != now.info.Mode()`,
			"^TestSignatureScratchWitness$", "TestSignatureScratchWitness/identity", "staging identity plant escaped", "TestSignatureScratchWitness/transient"},

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
