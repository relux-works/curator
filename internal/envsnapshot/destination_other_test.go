//go:build !darwin && !linux

package envsnapshot

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// Unqualified platforms refuse without creating a destination. This is not
// owner/ACL/reparse or publication qualification for Windows.
func TestDestinationUnqualifiedPlatform(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "absent")
	d, err := OpenDestination(path, DestinationPolicy{ProtectedRoots: []string{root}})
	var refusal *Error
	if d != nil || !errors.As(err, &refusal) || refusal.Code != DestinationInvalid || refusal.Reason != "platform-unqualified" {
		t.Fatal("unqualified platform accepted destination", d, err)
	}
	if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("unqualified platform changed destination", err)
	}
	var unsupported Destination
	v, err := unsupported.Publish(context.Background(), Publication{})
	if v != nil || !errors.As(err, &refusal) || refusal.Code != DestinationInvalid || refusal.Reason != "platform-unqualified" {
		t.Fatal("unqualified platform published", v, err)
	}
}
