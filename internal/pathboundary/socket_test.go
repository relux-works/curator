//go:build unix || windows

package pathboundary

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// A real AF_UNIX socket belongs to regular_types, including Windows' socket
// reparse tag. An ordinary file passes and a symlink still fails link_safety.
func TestSocketIsRegularTypeFailure(t *testing.T) {
	root, err := os.MkdirTemp(os.TempDir(), "pb-sock-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(root); err != nil {
			t.Error(err)
		}
	})
	file := filepath.Join(root, "plain")
	if err := os.WriteFile(file, []byte("plain"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ProtectTree(root); err != nil {
		t.Fatal(err)
	}
	if err := ValidateWithin(root, file); err != nil {
		t.Fatalf("ordinary neighbor refused: %v", err)
	}
	socket := filepath.Join(root, "s")
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: socket, Net: "unix"})
	if err != nil {
		t.Fatalf("required real socket: %v", err)
	}
	defer func() { _ = listener.Close() }()
	info, err := os.Lstat(socket)
	if err != nil || info.Mode()&os.ModeSocket == 0 {
		t.Fatalf("not a real socket: %v %v", info, err)
	}
	err = ValidateWithin(root, socket)
	var failure *Failure
	if !errors.As(err, &failure) || failure.Check != CheckRegular || failure.Path != socket {
		t.Fatalf("socket missed regular_types refusal: %v", err)
	}
	t.Logf("real socket on %s mode=%s classified=%s", runtime.GOOS, info.Mode(), failure.Check)
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(file, link); err != nil {
		t.Fatalf("required symlink neighbor: %v", err)
	}
	err = ValidateWithin(root, link)
	if !errors.As(err, &failure) || failure.Check != CheckLinkSafety {
		t.Fatalf("symlink lost link_safety: %v", err)
	}
}
