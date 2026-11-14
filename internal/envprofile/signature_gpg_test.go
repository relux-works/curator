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
	home, err := os.MkdirTemp("", "gpg-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(home); err != nil {
			t.Errorf("remove GPG fixture: %v", err)
		}
	})
	gpgHome := gpgFixturePath(t, "gpg", home)
	if _, err := exec.LookPath("gpgconf"); err != nil {
		t.Skip("gpg-fixture-agent-cleanup-unavailable: gpgconf missing")
	}
	gpgconfHome := gpgFixturePath(t, "gpgconf", home)
	t.Cleanup(func() {
		if output, err := exec.Command("gpgconf", "--homedir", gpgconfHome, "--kill", "all").CombinedOutput(); err != nil {
			t.Errorf("stop fixture-only GPG agents: %v %s", err, output)
		}
	})
	t.Logf("%s GPG fixture path convention qualified", runtime.GOOS)
	gen := exec.Command("gpg", "--homedir", gpgHome, "--batch", "--pinentry-mode", "loopback", "--passphrase", "", "--quick-generate-key", "fixture@example.test", "ed25519", "sign", "0")
	if output, err := gen.CombinedOutput(); err != nil {
		t.Fatalf("fixture generation: %v\n%s", err, output)
	}
	signer := exec.Command("gpg", "--homedir", gpgHome, "--batch", "--pinentry-mode", "loopback", "--passphrase", "", "--armor", "--detach-sign")
	signer.Stdin = strings.NewReader("signed bytes\n")
	signature, err := signer.Output()
	if err != nil {
		t.Fatal(err)
	}
	public, err := exec.Command("gpg", "--homedir", gpgHome, "--batch", "--export").Output()
	if err != nil {
		t.Fatal(err)
	}
	readonly := t.TempDir()
	if err := os.Chmod(readonly, 0700); err != nil {
		t.Fatal(err)
	}
	ring := filepath.Join(readonly, "pubring.gpg")
	gpgvHome := gpgFixturePath(t, "gpgv", readonly)
	gpgvRing := gpgFixturePath(t, "gpgv", ring)
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

			command := exec.Command("gpgv", "--homedir", gpgvHome, "--keyring", gpgvRing, "--status-fd", "1", "--", "/dev/fd/3", "-")
			out, err := runSignaturePipes(command, []byte(tc.message), signature)

			if (err == nil) != tc.valid {
				t.Fatalf("gpgv = %v valid=%v", err, tc.valid)
			}
			if !tc.valid && !strings.Contains(string(out), "[GNUPG:] BADSIG") {
				t.Fatalf("tampered fixture did not produce a bad-signature verdict: %v %s", err, out)
			}
			t.Logf("%s real OpenPGP %s control PASS", runtime.GOOS, tc.name)
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

func gpgFixturePath(t *testing.T, executable, path string) string {
	t.Helper()
	converted, err := signatureVerifierPath(executable, path)
	if err != nil {
		if strings.Contains(err.Error(), "gpg-path-flavour-unknown:") {
			t.Skipf("gpg-fixture-path-convention-unavailable: %v", err)
		}
		t.Fatalf("GPG fixture path setup failed: %v", err)
	}
	t.Logf("%s executable=%s fixture path convention PASS", runtime.GOOS, executable)
	return converted
}

// Native and both POSIX runtime conventions keep spaces and Unicode intact;
// unknown runtimes and paths without an absolute drive are named refusals.
func TestSignatureWindowsPathConventions(t *testing.T) {
	for _, tc := range []struct{ flavour, want string }{
		{"native", `C:\fixture space\unicode-π\keyring`},
		{"msys", "/c/fixture space/unicode-π/keyring"},
		{"cygwin", "/cygdrive/c/fixture space/unicode-π/keyring"},
	} {
		t.Run(tc.flavour, func(t *testing.T) {
			got, err := signatureWindowsPath(tc.flavour, `C:\fixture space\unicode-π\keyring`)
			if err != nil || got != tc.want {
				t.Fatalf("path=%q err=%v want=%q", got, err, tc.want)
			}
		})
	}
	for _, path := range []string{`relative`, `\\server\share`, "/tmp/path"} {
		if _, err := signatureWindowsPath("msys", path); err == nil || !strings.Contains(err.Error(), "gpg-path-unrepresentable") {
			t.Fatalf("unsupported path admitted: %q %v", path, err)
		}
	}
	if _, err := signatureWindowsPath("unknown", `C:\fixture`); err == nil || !strings.Contains(err.Error(), "gpg-path-flavour-unknown") {
		t.Fatal("unknown runtime admitted")
	}
	for _, tc := range []struct {
		libraries []string
		want      string
	}{
		{[]string{"KERNEL32.dll"}, "native"},
		{[]string{"KERNEL32.dll", "MSYS-2.0.DLL"}, "msys"},
		{[]string{"KERNEL32.dll", "cygwin1.dll"}, "cygwin"},
		{nil, "unknown"},
	} {
		if got := signatureRuntimeFlavour(tc.libraries); got != tc.want {
			t.Fatalf("flavour=%q want=%q", got, tc.want)
		}
	}
}
