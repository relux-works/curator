//go:build darwin || linux

package testtoolchain

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// A FIFO cannot extend the waiter's deadline; the real OS lock remains held.
// Replacing it with a valid record restores the named diagnostic.
func TestHostGOROOTLockFIFO(t *testing.T) {
	root := t.TempDir()
	_, release := startHostGOROOTHolder(t, root)
	path := filepath.Join(root, "holder.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := unix.Mkfifo(path, 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		lock, err := acquireHostGOROOT(ctx, root, "fifo-waiter", func(string, ...any) {})
		if lock != nil {
			_ = lock.Close()
		}
		done <- err
	}()
	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) || !strings.Contains(err.Error(), "holder=unknown") {
			t.Fatalf("want deadline: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("FIFO exceeded deadline watchdog")
	}
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeNamedPipe == 0 {
		t.Fatal("waiter changed FIFO")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if readHostGOROOTOwner(path) == nil {
		t.Fatal("restored regular record refused")
	}
	release()
}
