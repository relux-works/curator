package install

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const (
	fixtureVersion      = "v1.2.3"
	fixtureArchive      = "curator_1.2.3_darwin_amd64.tar.gz"
	fixtureSignerPrefix = "https://github.com/relux-works/curator/.github/workflows/release.yml@refs/tags/v"
	fixtureIdentityRE   = `^https://github.com/relux-works/curator/\.github/workflows/release\.yml@refs/tags/v`
	fixtureOIDCIssuer   = "https://token.actions.githubusercontent.com"
	fixtureWorkflow     = "relux-works/curator/.github/workflows/release.yml"
)

type installScriptFixture struct {
	ghAvailable              bool
	ghAttestationCommand     bool
	ghSupportsSignerWorkflow bool
	ghAttestationValid       bool
	cosignAvailable          bool
	cosignSignatureValid     bool
	signerIdentity           string
	tamperArchive            bool
	omitChecksumEntry        bool
	duplicateChecksumEntry   bool
	forceShasum              bool
	skipVerification         bool
}

type installScriptResult struct {
	output       string
	exitCode     int
	installed    bool
	ghCalls      string
	cosignCalls  string
	curlRequests string
}

func TestInstallScriptSecurityRows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("install.sh supports macOS and Linux")
	}

	type row struct {
		name                 string
		acceptance           bool
		fixture              installScriptFixture
		install              bool
		refusalMessage       string
		warning              bool
		expectCosignPolicy   bool
		expectNoGH           bool
		expectNoCosign       bool
		expectedGHArgs       []string
		expectedGHCall       string
		expectedCurlRequests []string
	}
	rows := []row{
		{
			name:       "valid_attestation_installs",
			acceptance: true,
			fixture: installScriptFixture{
				ghAvailable:              true,
				ghAttestationCommand:     true,
				ghSupportsSignerWorkflow: true,
				ghAttestationValid:       true,
			},
			install: true,
			expectedGHArgs: []string{
				"attestation verify",
				"--repo relux-works/curator",
				"checksums.txt",
			},
			expectedGHCall: "attestation verify checksums.txt --repo relux-works/curator --signer-workflow " + fixtureWorkflow + " --cert-oidc-issuer " + fixtureOIDCIssuer,
			expectNoCosign: true,
		},
		{
			name:       "tampered_archive_refused",
			acceptance: true,
			fixture: installScriptFixture{
				ghAvailable:              true,
				ghAttestationCommand:     true,
				ghSupportsSignerWorkflow: true,
				ghAttestationValid:       true,
				tamperArchive:            true,
			},
			refusalMessage: "SHA-256 verification failed",
			expectedGHCall: "attestation verify checksums.txt --repo relux-works/curator --signer-workflow " + fixtureWorkflow + " --cert-oidc-issuer " + fixtureOIDCIssuer,
			expectNoCosign: true,
		},
		{
			name:       "missing_archive_checksum_refused",
			acceptance: true,
			fixture: installScriptFixture{
				ghAvailable:              true,
				ghAttestationCommand:     true,
				ghSupportsSignerWorkflow: true,
				ghAttestationValid:       true,
				omitChecksumEntry:        true,
			},
			refusalMessage: "SHA-256 verification failed",
			expectedGHCall: "attestation verify checksums.txt --repo relux-works/curator --signer-workflow " + fixtureWorkflow + " --cert-oidc-issuer " + fixtureOIDCIssuer,
			expectNoCosign: true,
		},
		{
			name: "duplicate_archive_checksum_refused",
			fixture: installScriptFixture{
				ghAvailable:              true,
				ghAttestationCommand:     true,
				ghSupportsSignerWorkflow: true,
				ghAttestationValid:       true,
				duplicateChecksumEntry:   true,
			},
			refusalMessage: "SHA-256 verification failed",
			expectedGHCall: "attestation verify checksums.txt --repo relux-works/curator --signer-workflow " + fixtureWorkflow + " --cert-oidc-issuer " + fixtureOIDCIssuer,
			expectNoCosign: true,
		},
		{
			name:       "unattested_checksums_refused",
			acceptance: true,
			fixture: installScriptFixture{
				ghAvailable:              true,
				ghAttestationCommand:     true,
				ghSupportsSignerWorkflow: true,
				ghAttestationValid:       false,
			},
			refusalMessage: "GitHub attestation verification failed",
			expectedGHCall: "attestation verify checksums.txt --repo relux-works/curator --signer-workflow " + fixtureWorkflow + " --cert-oidc-issuer " + fixtureOIDCIssuer,
			expectNoCosign: true,
		},
		{
			name:       "bad_cosign_signature_refused",
			acceptance: true,
			fixture: installScriptFixture{
				cosignAvailable:      true,
				cosignSignatureValid: false,
			},
			refusalMessage:     "cosign signature verification failed",
			expectCosignPolicy: true,
			expectNoGH:         true,
		},
		{
			name:       "wrong_cosign_identity_refused",
			acceptance: true,
			fixture: installScriptFixture{
				cosignAvailable:      true,
				cosignSignatureValid: true,
				signerIdentity:       "https://github.com/untrusted/repository/.github/workflows/release.yml@refs/tags/v1.2.3",
			},
			refusalMessage:     "cosign signature verification failed",
			expectCosignPolicy: true,
			expectNoGH:         true,
		},
		{
			name:           "missing_verifier_refused",
			acceptance:     true,
			fixture:        installScriptFixture{},
			refusalMessage: "install GitHub CLI",
		},
		{
			name:           "explicit_optout_warns_and_installs",
			acceptance:     true,
			fixture:        installScriptFixture{skipVerification: true},
			install:        true,
			warning:        true,
			expectNoGH:     true,
			expectNoCosign: true,
		},
		{
			name: "cosign_fallback_installs",
			fixture: installScriptFixture{
				ghAvailable:          true,
				ghAttestationCommand: false,
				cosignAvailable:      true,
				cosignSignatureValid: true,
				forceShasum:          runtime.GOOS == "darwin",
			},
			install:              true,
			expectCosignPolicy:   true,
			expectNoGH:           true,
			expectedCurlRequests: []string{"checksums.txt.sig", "checksums.txt.pem"},
		},
		{
			name: "failed_attestation_does_not_downgrade",
			fixture: installScriptFixture{
				ghAvailable:              true,
				ghAttestationCommand:     true,
				ghSupportsSignerWorkflow: true,
				ghAttestationValid:       false,
				cosignAvailable:          true,
				cosignSignatureValid:     true,
			},
			refusalMessage: "GitHub attestation verification failed",
			expectedGHCall: "attestation verify checksums.txt --repo relux-works/curator --signer-workflow " + fixtureWorkflow + " --cert-oidc-issuer " + fixtureOIDCIssuer,
			expectNoCosign: true,
		},
		{
			name: "older_gh_pins_identity_without_signer_workflow_flag",
			fixture: installScriptFixture{
				ghAvailable:          true,
				ghAttestationCommand: true,
				ghAttestationValid:   true,
			},
			install:        true,
			expectedGHCall: "attestation verify checksums.txt --repo relux-works/curator --cert-identity-regex " + fixtureIdentityRE + " --cert-oidc-issuer " + fixtureOIDCIssuer,
			expectNoCosign: true,
		},
	}

	const expectedAcceptanceRows = 8
	const expectedTotalRows = 12
	acceptanceRows := 0
	for _, testRow := range rows {
		if testRow.acceptance {
			acceptanceRows++
		}
	}
	if acceptanceRows != expectedAcceptanceRows || len(rows) != expectedTotalRows {
		t.Fatalf("production-entry row coverage = %d/%d acceptance rows and %d/%d total rows",
			acceptanceRows, expectedAcceptanceRows, len(rows), expectedTotalRows)
	}

	passed := 0
	for _, testRow := range rows {
		testRow := testRow
		if t.Run(testRow.name, func(t *testing.T) {
			result := runInstallScriptFixture(t, testRow.fixture)
			if testRow.install {
				if result.exitCode != 0 || !result.installed {
					t.Fatalf("expected install: exit=%d installed=%t\n%s", result.exitCode, result.installed, result.output)
				}
			} else {
				assertInstallRefused(t, result, testRow.refusalMessage)
			}
			if testRow.warning && (!strings.Contains(result.output, "WARNING: CURATOR_INSTALL_INSECURE_SKIP_VERIFY=1") ||
				!strings.Contains(result.output, "without integrity verification")) {
				t.Errorf("opt-out warning was not loud and explicit:\n%s", result.output)
			}
			if testRow.expectCosignPolicy {
				assertCosignPolicyArgs(t, result.cosignCalls)
			}
			if testRow.expectNoGH && result.ghCalls != "" {
				t.Errorf("unexpected gh call: %s", result.ghCalls)
			}
			if testRow.expectNoCosign && result.cosignCalls != "" {
				t.Errorf("unexpected cosign call: %s", result.cosignCalls)
			}
			for _, expected := range testRow.expectedGHArgs {
				if !strings.Contains(result.ghCalls, expected) {
					t.Errorf("gh invocation does not contain %q:\n%s", expected, result.ghCalls)
				}
			}
			if testRow.expectedGHCall != "" && result.ghCalls != testRow.expectedGHCall+"\n" {
				t.Errorf("gh argv = %q, want exactly %q", result.ghCalls, testRow.expectedGHCall+"\n")
			}
			for _, expected := range testRow.expectedCurlRequests {
				if !strings.Contains(result.curlRequests, expected) {
					t.Errorf("release request list does not contain %q:\n%s", expected, result.curlRequests)
				}
			}
		}) {
			passed++
		}
	}
	t.Logf("install.sh production-entry coverage: %d/%d rows passed; %d/%d task acceptance rows", passed, expectedTotalRows, acceptanceRows, expectedAcceptanceRows)
	if passed != expectedTotalRows {
		t.Errorf("production-entry rows passed = %d/%d", passed, expectedTotalRows)
	}
}

func assertCosignPolicyArgs(t *testing.T, calls string) {
	t.Helper()
	for _, expected := range []string{
		"verify-blob",
		"--signature ",
		"--certificate ",
		"--certificate-identity-regexp " + fixtureIdentityRE,
		"--certificate-oidc-issuer " + fixtureOIDCIssuer,
	} {
		if !strings.Contains(calls, expected) {
			t.Errorf("cosign invocation does not contain %q:\n%s", expected, calls)
		}
	}
}

func assertInstallRefused(t *testing.T, result installScriptResult, message string) {
	t.Helper()
	if result.exitCode == 0 || result.installed {
		t.Fatalf("installer accepted a release that must be refused: exit=%d installed=%t\n%s", result.exitCode, result.installed, result.output)
	}
	if !strings.Contains(result.output, message) || !strings.Contains(result.output, "refusing to install") {
		t.Fatalf("refusal lacks a clear reason (%q):\n%s", message, result.output)
	}
}

func runInstallScriptFixture(t *testing.T, fixture installScriptFixture) installScriptResult {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("install.sh supports macOS and Linux")
	}

	root := t.TempDir()
	releaseDir := filepath.Join(root, "release")
	toolDir := filepath.Join(root, "tools")
	installDir := filepath.Join(root, "installed")
	for _, dir := range []string{releaseDir, toolDir, installDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	archivePath := filepath.Join(releaseDir, fixtureArchive)
	writeFixtureArchive(t, archivePath, false)
	archiveBytes, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(archiveBytes)
	checksums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(digest[:]), fixtureArchive)
	if fixture.omitChecksumEntry {
		checksums = "# checksums file has no row for this archive\n"
	} else if fixture.duplicateChecksumEntry {
		checksums += checksums
	}
	if err := os.WriteFile(filepath.Join(releaseDir, "checksums.txt"), []byte(checksums), 0o644); err != nil {
		t.Fatal(err)
	}
	if fixture.tamperArchive {
		writeFixtureArchive(t, archivePath, true)
	}
	for name, content := range map[string]string{
		"checksums.txt.sig": "fixture signature",
		"checksums.txt.pem": "fixture certificate",
	} {
		if err := os.WriteFile(filepath.Join(releaseDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	for _, name := range []string{"awk", "cp", "cut", "grep", "gzip", "head", "install", "mkdir", "mktemp", "rm", "sha256sum", "shasum", "tar", "tr"} {
		if name == "sha256sum" && fixture.forceShasum {
			continue
		}
		path, err := exec.LookPath(name)
		if err != nil {
			if name == "sha256sum" || name == "shasum" {
				continue
			}
			t.Fatalf("required test utility %q is unavailable: %v", name, err)
		}
		if err := os.Symlink(path, filepath.Join(toolDir, name)); err != nil {
			t.Fatal(err)
		}
	}
	if fixture.forceShasum {
		if _, err := exec.LookPath("shasum"); err != nil {
			t.Fatalf("shasum is required for the macOS checksum fallback row: %v", err)
		}
	}
	writeInstallerExecutable(t, filepath.Join(toolDir, "uname"), `#!/bin/sh
case "${1:-}" in
  -s) echo Darwin ;;
  -m) echo x86_64 ;;
  *) exit 2 ;;
esac
`)

	curlLog := filepath.Join(root, "curl.log")
	writeInstallerExecutable(t, filepath.Join(toolDir, "curl"), `#!/bin/sh
out=
url=
while [ "$#" -gt 0 ]; do
  case "$1" in
    -o) out=$2; shift 2 ;;
    -*) shift ;;
    *) url=$1; shift ;;
  esac
done
name=${url##*/}
printf '%s\n' "$name" >> "$CURL_LOG"
if [ -z "$out" ] || [ ! -f "$FIXTURE_RELEASE_DIR/$name" ]; then
  echo "fixture release asset not found: $name" >&2
  exit 22
fi
cp "$FIXTURE_RELEASE_DIR/$name" "$out"
`)

	ghLog := filepath.Join(root, "gh.log")
	if fixture.ghAvailable {
		writeInstallerExecutable(t, filepath.Join(toolDir, "gh"), `#!/bin/sh
if [ "${1:-}" = attestation ] && [ "${2:-}" = verify ] && [ "${3:-}" = --help ]; then
  if [ "$GH_ATTESTATION_COMMAND" != 1 ]; then
    echo "unknown command: attestation verify" >&2
    exit 127
  fi
  if [ "$GH_SUPPORTS_SIGNER_WORKFLOW" = 1 ]; then echo --signer-workflow; fi
  exit 0
fi
pins=0
line=
for arg in "$@"; do
  case "$arg" in
    --cert-identity|--cert-identity-regex|--signer-repo|--signer-workflow) pins=$((pins + 1)) ;;
  esac
  case "$arg" in
    */checksums.txt) arg=checksums.txt ;;
  esac
  line="${line:+$line }$arg"
done
printf '%s\n' "$line" >> "$GH_LOG"
if [ "$pins" -gt 1 ]; then
  echo "if any flags in the group [cert-identity cert-identity-regex signer-repo signer-workflow] are set none of the others can be" >&2
  exit 1
fi
if [ "$GH_ATTESTATION_VALID" != 1 ]; then
  echo "fixture attestation rejected" >&2
  exit 1
fi
exit 0
`)
	}

	cosignLog := filepath.Join(root, "cosign.log")
	if fixture.cosignAvailable {
		writeInstallerExecutable(t, filepath.Join(toolDir, "cosign"), `#!/bin/sh
printf '%s\n' "$*" >> "$COSIGN_LOG"
if [ "${1:-}" != verify-blob ]; then exit 2; fi
shift
file=$1
shift
signature=
certificate=
identity=
issuer=
while [ "$#" -gt 0 ]; do
  case "$1" in
    --signature) signature=$2; shift 2 ;;
    --certificate) certificate=$2; shift 2 ;;
    --certificate-identity-regexp) identity=$2; shift 2 ;;
    --certificate-oidc-issuer) issuer=$2; shift 2 ;;
    *) shift ;;
  esac
done
if [ ! -s "$file" ] || [ ! -s "$signature" ] || [ ! -s "$certificate" ]; then
  echo "fixture signature inputs missing" >&2
  exit 1
fi
if [ "$identity" != "$EXPECTED_COSIGN_IDENTITY_REGEX" ] || [ "$issuer" != "$EXPECTED_COSIGN_ISSUER" ]; then
  echo "fixture verifier policy mismatch" >&2
  exit 1
fi
if ! printf '%s\n' "$FIXTURE_SIGNER_IDENTITY" | grep -Eq "$identity"; then
  echo "fixture certificate identity mismatch" >&2
  exit 1
fi
if [ "$COSIGN_SIGNATURE_VALID" != 1 ]; then
  echo "fixture signature invalid" >&2
  exit 1
fi
exit 0
`)
	}

	signerIdentity := fixture.signerIdentity
	if signerIdentity == "" {
		signerIdentity = fixtureSignerPrefix + "1.2.3"
	}
	installerPath, err := filepath.Abs(filepath.Join("..", "..", "install.sh"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("/bin/sh", installerPath)
	cmd.Env = []string{
		"PATH=" + toolDir,
		"HOME=" + filepath.Join(root, "home"),
		"TMPDIR=" + root,
		"CURATOR_VERSION=" + fixtureVersion,
		"CURATOR_BIN_DIR=" + installDir,
		"FIXTURE_RELEASE_DIR=" + releaseDir,
		"CURL_LOG=" + curlLog,
		"GH_LOG=" + ghLog,
		"GH_ATTESTATION_COMMAND=" + boolEnv(fixture.ghAttestationCommand),
		"GH_SUPPORTS_SIGNER_WORKFLOW=" + boolEnv(fixture.ghSupportsSignerWorkflow),
		"GH_ATTESTATION_VALID=" + boolEnv(fixture.ghAttestationValid),
		"COSIGN_LOG=" + cosignLog,
		"COSIGN_SIGNATURE_VALID=" + boolEnv(fixture.cosignSignatureValid),
		"EXPECTED_COSIGN_IDENTITY_REGEX=" + fixtureIdentityRE,
		"EXPECTED_COSIGN_ISSUER=" + fixtureOIDCIssuer,
		"FIXTURE_SIGNER_IDENTITY=" + signerIdentity,
	}
	if fixture.skipVerification {
		cmd.Env = append(cmd.Env, "CURATOR_INSTALL_INSECURE_SKIP_VERIFY=1")
	}

	output, runErr := cmd.CombinedOutput()
	result := installScriptResult{output: string(output)}
	if runErr != nil {
		var exitErr *exec.ExitError
		if !errors.As(runErr, &exitErr) {
			t.Fatalf("running install.sh: %v\n%s", runErr, output)
		}
		result.exitCode = exitErr.ExitCode()
	}
	if _, err := os.Stat(filepath.Join(installDir, "curator")); err == nil {
		result.installed = true
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("checking installed binary: %v", err)
	}
	result.ghCalls = readOptionalFixtureLog(t, ghLog)
	result.cosignCalls = readOptionalFixtureLog(t, cosignLog)
	result.curlRequests = readOptionalFixtureLog(t, curlLog)
	return result
}

func writeFixtureArchive(t *testing.T, path string, includeExtraFile bool) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gz)
	binary := []byte("#!/bin/sh\nif [ \"${1:-}\" = --version ]; then echo 'curator v1.2.3'; exit 0; fi\nexit 0\n")
	if err := tarWriter.WriteHeader(&tar.Header{Name: "curator", Mode: 0o755, Size: int64(len(binary)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write(binary); err != nil {
		t.Fatal(err)
	}
	if includeExtraFile {
		note := []byte("valid archive with a changed digest\n")
		if err := tarWriter.WriteHeader(&tar.Header{Name: "README.txt", Mode: 0o644, Size: int64(len(note)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tarWriter.Write(note); err != nil {
			t.Fatal(err)
		}
	}
	for _, closer := range []io.Closer{tarWriter, gz, file} {
		if err := closer.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func writeInstallerExecutable(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
}

func readOptionalFixtureLog(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return ""
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func boolEnv(value bool) string {
	if value {
		return "1"
	}
	return "0"
}
