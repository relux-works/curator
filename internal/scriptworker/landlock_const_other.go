//go:build !linux

package scriptworker

// Off-Linux mirror of the Landlock UAPI values in
// include/uapi/linux/landlock.h. x/sys exposes the LANDLOCK_ACCESS_FS_*
// constants on Linux only, so the pure mask helpers below keep one
// definition per platform: these literals never reach a system call off
// Linux (the probe reports the controls absent and confirmation refuses),
// and the mask tests pin them against the same header values on every
// platform, so a drift fails the suite on the platform that carries it.
const (
	landlockAccessFSExecute    = 1 << 0
	landlockAccessFSWriteFile  = 1 << 1
	landlockAccessFSRemoveDir  = 1 << 4
	landlockAccessFSRemoveFile = 1 << 5
	landlockAccessFSMakeChar   = 1 << 6
	landlockAccessFSMakeDir    = 1 << 7
	landlockAccessFSMakeReg    = 1 << 8
	landlockAccessFSMakeSock   = 1 << 9
	landlockAccessFSMakeFifo   = 1 << 10
	landlockAccessFSMakeBlock  = 1 << 11
	landlockAccessFSMakeSym    = 1 << 12
	landlockAccessFSRefer      = 1 << 13
	landlockAccessFSTruncate   = 1 << 14
	landlockAccessFSIoctlDev   = 1 << 15
)
