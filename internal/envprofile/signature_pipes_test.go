package envprofile

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Inherited signer/signature pipes verify the signed bytes and reject changed
// bytes or another principal; failure does not deadlock unread pipe writers.
func TestSignaturePipes(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "plan9" {
		t.Skip("inherited descriptor backend unavailable")
	}
	key := filepath.Join(t.TempDir(), "signer")
	if out, err := exec.Command("ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", key).CombinedOutput(); err != nil {
		t.Fatalf("fixture key: %v %s", err, out)
	}
	public, err := os.ReadFile(key + ".pub")
	if err != nil {
		t.Fatal(err)
	}
	signer := exec.Command("ssh-keygen", "-Y", "sign", "-n", "git", "-f", key)
	signer.Stdin = strings.NewReader("signed bytes\n")
	signature, err := signer.Output()
	if err != nil {
		t.Fatal(err)
	}
	fields := strings.Fields(string(public))
	allowed := []byte("fixture " + fields[0] + " " + fields[1] + "\n")
	for _, tc := range []struct {
		name, message, principal string
		valid                    bool
	}{{"valid", "signed bytes\n", "fixture", true}, {"tampered", "changed bytes\n", "fixture", false}, {"wrong-principal", "signed bytes\n", "other", false}} {
		t.Run(tc.name, func(t *testing.T) {

			_, err := runSignaturePipes("ssh-keygen", []string{"-Y", "verify", "-f", "/dev/fd/3", "-I", tc.principal, "-n", "git", "-s", "/dev/fd/4"}, []byte(tc.message), allowed, signature)

			if (err == nil) != tc.valid {
				t.Fatalf("verification = %v, valid=%v", err, tc.valid)
			}
		})
	}
	// Failed exec must also release pipe goroutines, even above pipe capacity.
	if _, err := runSignaturePipes("/missing-verifier", nil, nil, make([]byte, 1024*1024)); err == nil {
		t.Fatal("missing verifier accepted")
	}
}
