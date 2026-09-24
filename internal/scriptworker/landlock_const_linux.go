//go:build linux

package scriptworker

import "golang.org/x/sys/unix"

// The Landlock filesystem access rights are the kernel UAPI values from
// golang.org/x/sys/unix — never hand-written literals. A wrong literal
// silently handles the wrong right while leaving the intended one
// unrestricted (Landlock permits every unhandled action), and the mask
// tests pin the resulting values against the header on every platform.
const (
	landlockAccessFSExecute    = unix.LANDLOCK_ACCESS_FS_EXECUTE
	landlockAccessFSWriteFile  = unix.LANDLOCK_ACCESS_FS_WRITE_FILE
	landlockAccessFSRemoveDir  = unix.LANDLOCK_ACCESS_FS_REMOVE_DIR
	landlockAccessFSRemoveFile = unix.LANDLOCK_ACCESS_FS_REMOVE_FILE
	landlockAccessFSMakeChar   = unix.LANDLOCK_ACCESS_FS_MAKE_CHAR
	landlockAccessFSMakeDir    = unix.LANDLOCK_ACCESS_FS_MAKE_DIR
	landlockAccessFSMakeReg    = unix.LANDLOCK_ACCESS_FS_MAKE_REG
	landlockAccessFSMakeSock   = unix.LANDLOCK_ACCESS_FS_MAKE_SOCK
	landlockAccessFSMakeFifo   = unix.LANDLOCK_ACCESS_FS_MAKE_FIFO
	landlockAccessFSMakeBlock  = unix.LANDLOCK_ACCESS_FS_MAKE_BLOCK
	landlockAccessFSMakeSym    = unix.LANDLOCK_ACCESS_FS_MAKE_SYM
	landlockAccessFSRefer      = unix.LANDLOCK_ACCESS_FS_REFER
	landlockAccessFSTruncate   = unix.LANDLOCK_ACCESS_FS_TRUNCATE
	landlockAccessFSIoctlDev   = unix.LANDLOCK_ACCESS_FS_IOCTL_DEV
)
