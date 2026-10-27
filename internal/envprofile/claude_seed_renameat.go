//go:build linux && !loong64 && !riscv64

package envprofile

import "golang.org/x/sys/unix"

const claudeSeedRenameTrap = unix.SYS_RENAMEAT
