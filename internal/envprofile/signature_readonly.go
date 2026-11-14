package envprofile

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/stateread"
)

var errSignerVerificationUnavailable = errors.New("signer_verification_unavailable")

func verifyReadOnlyObjectSignature(objectType string, payload []byte, allowed []contextresolve.Signer) (contextresolve.SignatureEvidence, error) {
	evidence := contextresolve.SignatureEvidence{Present: true}
	data, signature, err := splitGitSignedPayload(objectType, payload)
	if err != nil {
		// Malformed signature framing is a failed cryptographic verdict, while
		// the caller's already-read object remains the same recorded pin.
		return evidence, nil
	}
	if runtime.GOOS == "plan9" {
		return evidence, fmt.Errorf("%w: read-only signature input transport is unavailable on this platform", errSignerVerificationUnavailable)
	}
	if bytes.Contains(signature, []byte("-----BEGIN SSH SIGNATURE-----")) {
		// Armor inspection is informational; only the verifier can admit a signer.
		if key, err := sshPublicKeyFromSignature(signature); err == nil {
			evidence.Signer = contextresolve.Signer{Type: "ssh", Key: key}
		}
		if _, err := exec.LookPath("ssh-keygen"); err != nil {
			return evidence, fmt.Errorf("%w: SSH verifier is unavailable", errSignerVerificationUnavailable)
		}
		principal := gitSignaturePrincipal(objectType, payload)
		if principal == "" {
			principal = "curator"
		}
		for _, signer := range allowed {
			if signer.Type != "ssh" {
				continue
			}
			fields := strings.Fields(signer.Key)
			if len(fields) < 2 {
				continue
			}
			policy := []byte(principal + " " + fields[0] + " " + fields[1] + "\n")
			command := exec.Command("ssh-keygen", "-Y", "verify", "-f", "/dev/fd/3", "-I", principal, "-n", "git", "-s", "/dev/fd/4") // #nosec G204 -- literal verifier and argv-only principal; no shell is invoked
			output, runErr := runSignaturePipes(command, data, policy, signature)
			err := signatureVerifierResult("SSH", output, runErr)
			if err != nil {
				return evidence, err
			}
			if runErr == nil {
				evidence.Valid = true
				evidence.Signer = signer
				return evidence, nil
			}
		}
		return evidence, nil
	}
	if !bytes.Contains(signature, []byte("-----BEGIN PGP SIGNATURE-----")) {
		return evidence, nil
	}
	home, keyring, err := readOnlyGPGKeyring()
	if err != nil {
		return evidence, err
	}
	if _, err := exec.LookPath("gpgv"); err != nil {
		return evidence, fmt.Errorf("%w: read-only OpenPGP verifier is unavailable", errSignerVerificationUnavailable)
	}
	home, err = signatureVerifierPath("gpgv", home)
	if err != nil {
		return evidence, fmt.Errorf("%w: %v", errSignerVerificationUnavailable, err)
	}
	keyring, err = signatureVerifierPath("gpgv", keyring)
	if err != nil {
		return evidence, fmt.Errorf("%w: %v", errSignerVerificationUnavailable, err)
	}
	command := exec.Command("gpgv", "--homedir", home, "--keyring", keyring, "--status-fd", "1", "--", "/dev/fd/3", "-") // #nosec G204 -- literal verifier and argv-only paths; no shell is invoked
	output, verifyErr := runSignaturePipes(command, data, signature)
	if err := signatureVerifierResult("OpenPGP", output, verifyErr); err != nil {
		return evidence, err
	}
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 || fields[0] != "[GNUPG:]" {
			continue
		}
		switch fields[1] {
		case "VALIDSIG":
			evidence.Valid = verifyErr == nil
			evidence.Signer = contextresolve.Signer{Type: "gpg", Fingerprint: fields[2]}
			return evidence, nil
		case "BADSIG", "ERRSIG":
			evidence.Signer = contextresolve.Signer{Type: "gpg", Fingerprint: fields[2]}
		case "NO_PUBKEY":
			return evidence, fmt.Errorf("%w: the existing public keyring lacks the signing key", errSignerVerificationUnavailable)
		}
	}
	return evidence, nil
}

// Only existing public keyrings are read. No secret key, trust database,
// configuration file, keyserver or GPG agent participates in this verifier.
func readOnlyGPGKeyring() (home, keyring string, err error) {
	home = os.Getenv("GNUPGHOME")
	if home == "" {
		operator, err := os.UserHomeDir()
		if err != nil {
			return "", "", fmt.Errorf("%w: OpenPGP public keyring home is unavailable", errSignerVerificationUnavailable)
		}
		home = filepath.Join(operator, ".gnupg")
	}
	state, err := stateread.Lstat(home)
	if err != nil || state.Kind == stateread.KindAbsent {
		return "", "", fmt.Errorf("%w: no existing OpenPGP public keyring", errSignerVerificationUnavailable)
	}
	if err := privatedir.Validate(home); err != nil {
		return "", "", fmt.Errorf("%w: OpenPGP public keyring boundary is untrusted", errSignerVerificationUnavailable)
	}
	for _, name := range []string{"pubring.kbx", "pubring.gpg"} {
		candidate := filepath.Join(home, name)
		state, err := stateread.Lstat(candidate)
		if err != nil {
			return "", "", fmt.Errorf("%w: OpenPGP public keyring is unreadable", errSignerVerificationUnavailable)
		}
		if state.Kind == stateread.KindAbsent {
			continue
		}
		if state.Info == nil || !state.Info.Mode().IsRegular() {
			return "", "", fmt.Errorf("%w: OpenPGP public keyring must be a regular file", errSignerVerificationUnavailable)
		}
		return home, candidate, nil
	}
	return "", "", fmt.Errorf("%w: no existing OpenPGP public keyring", errSignerVerificationUnavailable)
}

// Exit status alone cannot distinguish a bad signature from unreadable input,
// unsupported options, or a failed verifier. Only recognized negative evidence
// from a completed transport is a cryptographic verdict.
func signatureVerifierResult(kind string, output []byte, runErr error) error {
	if runErr == nil {
		return nil
	}
	// A C verifier returning -1 is exit 255 on Unix and 0xffffffff on
	// Windows. Match its negative evidence, not one OS's truncated status.
	var exit *exec.ExitError
	if !errors.Is(runErr, errSignerVerificationUnavailable) && errors.As(runErr, &exit) && exit.Exited() && signatureNegativeVerdict(kind, output) {
		return nil
	}
	return fmt.Errorf("%w: %s verifier could not complete (%s); output=%s", errSignerVerificationUnavailable, kind, signatureDiagnostic([]byte(runErr.Error())), signatureDiagnostic(output))
}

func signatureNegativeVerdict(kind string, output []byte) bool {
	if kind == "SSH" {
		text := strings.ReplaceAll(string(output), "\r\n", "\n")
		// A policy/key mismatch emits only the final line. A cryptographic
		// mismatch emits its specific error followed by the same final line.
		return text == "Could not verify signature.\n" || text == "Signature verification failed: incorrect signature\nCould not verify signature.\n"
	}
	bad := false
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != "[GNUPG:]" {
			continue
		}
		switch fields[1] {
		case "NO_PUBKEY", "ERRSIG", "FAILURE", "ERROR":
			return false
		case "BADSIG":
			bad = true
		}
	}
	return bad
}

// Bound the escaped diagnostic itself, including quotes, so control bytes
// cannot inject terminal lines or expand a bounded input into unbounded output.
func signatureDiagnostic(data []byte) string {
	const limit = 1024
	var text strings.Builder
	text.WriteByte('"')
	for _, char := range string(data) {
		escaped := strconv.QuoteToASCII(string(char))
		escaped = escaped[1 : len(escaped)-1]
		if text.Len()+len(escaped)+5 > limit {
			text.WriteString("...")
			break
		}
		text.WriteString(escaped)
	}
	text.WriteByte('"')
	return text.String()
}
