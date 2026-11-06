//go:build !windows

package buildrepo

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func runWindowsPolicyHandleProbe() bool { return false }

// Unix accepts a bounded regular read-only descriptor, never a pathname,
// standard stream, pipe or oversized record. Refusal leaves bytes unchanged;
// this locator-only control starts no client.
func TestUnixSSHPolicyRecordBoundary(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "record")
	if err := os.WriteFile(path, []byte("policy"), 0600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	locator := strconv.Itoa(int(f.Fd()))
	if got, err := sshPolicyPayload(locator); err != nil || string(got) != "policy" {
		t.Fatalf("positive: %q %v", got, err)
	}
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Close(); _ = w.Close() }()
	for _, value := range []string{path, "0", "1", "2", strconv.Itoa(int(r.Fd()))} {
		if got, err := sshPolicyPayload(value); err == nil || got != nil {
			t.Fatalf("invalid locator accepted: %q %v", got, err)
		}
	}
	if err := os.WriteFile(path, []byte(strings.Repeat("x", 8193)), 0600); err != nil {
		t.Fatal(err)
	}
	if got, err := sshPolicyPayload(locator); err == nil || got != nil {
		t.Fatalf("oversize: %q %v", got, err)
	}
	if got, _ := os.ReadFile(path); len(got) != 8193 {
		t.Fatal("refusal changed bytes")
	}
}
