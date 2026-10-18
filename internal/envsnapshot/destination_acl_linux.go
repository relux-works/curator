//go:build linux

package envsnapshot

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

// Refuse extended access and inherited/default ACLs. ENODATA is evidence of
// absence; an unsupported filesystem or a failed ACL read is not absence.
func destinationACL(f *os.File) error {
	for _, name := range []string{"system.posix_acl_access", "system.posix_acl_default"} {
		_, err := unix.Fgetxattr(int(f.Fd()), name, nil)
		if !errors.Is(err, unix.ENODATA) {
			return os.ErrPermission
		}
	}
	return nil
}
