//go:build !windows

package scriptworker

import "io/fs"

// platformRequiresExecutableExtension is false on unix: the kernel executes
// the verified path itself however it is named, so the image header alone
// decides.
const platformRequiresExecutableExtension = false

func interpreterExecutable(mode fs.FileMode) bool { return mode.Perm()&0o111 != 0 }
