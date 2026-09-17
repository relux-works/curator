//go:build !windows

package staging

import (
	"fmt"
	"os"
	"syscall"
)

// inspectedIdentity is the result of one filesystem inspection behind an
// identity capture: the FileInfo a single Stat returned.
type inspectedIdentity struct {
	info os.FileInfo
}

// inspectPath performs the single filesystem inspection behind an identity
// capture, with Stat semantics (following links), matching Snapshot and
// Recheck. A missing path returns an IsNotExist error (a vanished
// boundary, proven overlap); any other failure returns a plain error the
// caller reports as boundary_identity_unreadable, never as a proven
// overlap.
func inspectPath(canonical string) (inspectedIdentity, error) {
	info, err := os.Stat(canonical) // #nosec G304 -- boundary identity under inspection
	if err != nil {
		return inspectedIdentity{}, err
	}
	return inspectedIdentity{info: info}, nil
}

// tokenOfInspection serializes the identity of an already-inspected object
// without touching the filesystem, so the token Snapshot stores is
// coherent with the inspection it came from. It must never Stat the path
// again: a second pathname-based read is the capture window a swap
// between the two reads would corrupt (see captureIdentity). A FileInfo
// without a stat buffer fails closed at planning.
func tokenOfInspection(inspected inspectedIdentity, canonical string) (string, error) {
	if inspected.info == nil {
		return "", fmt.Errorf("cannot read file identity for %s", canonical)
	}
	stat, ok := inspected.info.Sys().(*syscall.Stat_t)
	if !ok || stat == nil {
		return "", fmt.Errorf("cannot read file identity for %s", canonical)
	}
	return fmt.Sprintf("%d:%d", stat.Dev, stat.Ino), nil
}

// FileIdentity returns the durable filesystem identity token for the object
// at canonical (a resolved absolute path): device and inode, the same pair
// os.SameFile compares on unix. Both the same-process Recheck and restart
// recovery verify the live token against the planning-time pin (see
// verifyDurableIdentity). A missing path returns an IsNotExist error (a
// vanished boundary, proven overlap); any other failure returns a plain
// error the caller reports as boundary_identity_unreadable, never as a
// proven overlap.
func FileIdentity(canonical string) (string, error) {
	inspected, err := inspectPath(canonical)
	if err != nil {
		return "", err
	}
	return tokenOfInspection(inspected, canonical)
}
