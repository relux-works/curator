package envprofile

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	if runtime.GOOS == "windows" || runtime.GOOS == "plan9" {
		return evidence, fmt.Errorf("%w: inherited signature descriptors are unavailable on this platform", errSignerVerificationUnavailable)
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
			_, err := runSignaturePipes("ssh-keygen", []string{"-Y", "verify", "-f", "/dev/fd/3", "-I", principal, "-n", "git", "-s", "/dev/fd/4"}, data, policy, signature)
			if err != nil {
				var exit *exec.ExitError
				if !errors.As(err, &exit) {
					return evidence, fmt.Errorf("%w: SSH verification could not execute", errSignerVerificationUnavailable)
				}
			}
			if err == nil {
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
	output, verifyErr := runSignaturePipes("gpgv", []string{"--homedir", home, "--keyring", keyring, "--status-fd", "1", "--", "/dev/fd/3", "-"}, data, signature)
	if verifyErr != nil {
		var exit *exec.ExitError
		if !errors.As(verifyErr, &exit) {
			return evidence, fmt.Errorf("%w: OpenPGP verification could not execute", errSignerVerificationUnavailable)
		}
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
