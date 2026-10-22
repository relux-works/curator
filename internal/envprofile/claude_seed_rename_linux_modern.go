//go:build linux && (loong64 || riscv64)

package envprofile

import "golang.org/x/sys/unix"

// These architectures expose renameat2 instead of renameat. Zero flags have
// the same replacement semantics; their supported kernels provide this call.
const claudeSeedRenameTrap = unix.SYS_RENAMEAT2
