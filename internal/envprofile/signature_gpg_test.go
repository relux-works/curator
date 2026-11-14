package envprofile

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// gpgv verifies against a pre-existing public-only keyring through a signature
// pipe; the read phase neither invokes gpg nor initializes a keyring.
func TestGPGSignaturePipe(t *testing.T) {
	if runtime.GOOS == "plan9" {
		t.Skip("read-only signature input transport unavailable")
	}
	if _, err := exec.LookPath("gpgv"); err != nil {
		t.Skip("gpgv unavailable")
	}
	if _, err := exec.LookPath("gpg"); err != nil {
		t.Skip("fixture generator unavailable")
	}
	home := t.TempDir()
	gen := exec.Command("gpg", "--homedir", home, "--batch", "--pinentry-mode", "loopback", "--passphrase", "", "--quick-generate-key", "fixture@example.test", "ed25519", "sign", "0")
	if output, err := gen.CombinedOutput(); err != nil {
		t.Fatalf("fixture generation: %v\n%s", err, output)
	}
	signer := exec.Command("gpg", "--homedir", home, "--batch", "--pinentry-mode", "loopback", "--passphrase", "", "--armor", "--detach-sign")
	signer.Stdin = strings.NewReader("signed bytes\n")
	signature, err := signer.Output()
	if err != nil {
		t.Fatal(err)
	}
	public, err := exec.Command("gpg", "--homedir", home, "--batch", "--export").Output()
	if err != nil {
		t.Fatal(err)
	}
	readonly := t.TempDir()
	if err := os.Chmod(readonly, 0700); err != nil {
		t.Fatal(err)
	}
	ring := filepath.Join(readonly, "pubring.gpg")
	if err := os.WriteFile(ring, public, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GNUPGHOME", readonly)
	witness := readWitnesses(t, readonly)
	payload := append([]byte("signed bytes\n"), signature...)
	evidence, err := verifyReadOnlyObjectSignature("tag", payload, nil)
	if err != nil || !evidence.Valid || evidence.Signer.Type != "gpg" {
		t.Fatalf("read-only public keyring verification = %+v, %v", evidence, err)
	}
	assertReadWitnesses(t, readonly, witness)
	before, err := os.ReadDir(readonly)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, message string
		valid         bool
	}{{"valid", "signed bytes\n", true}, {"tampered", "changed bytes\n", false}} {
		t.Run(tc.name, func(t *testing.T) {

			command := exec.Command("gpgv", "--homedir", readonly, "--keyring", ring, "--status-fd", "1", "--", "/dev/fd/3", "-")
			out, err := runSignaturePipes(command, []byte(tc.message), signature)

			if (err == nil) != tc.valid {
				t.Fatalf("gpgv = %v valid=%v", err, tc.valid)
			}
			if tc.valid && !strings.Contains(string(out), "[GNUPG:] VALIDSIG") {
				t.Fatal("validity evidence missing")
			}
		})
	}
	after, err := os.ReadDir(readonly)
	if err != nil || len(before) != len(after) {
		t.Fatalf("keyring state changed: %v", err)
	}
}
