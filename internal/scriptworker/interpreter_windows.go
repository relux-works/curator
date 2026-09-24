//go:build windows

package scriptworker

import "io/fs"

// platformRequiresExecutableExtension is true on Windows: os/exec rewrites
// an extensionless absolute path to a neighboring PATHEXT sibling, so only
// a binding that names the .exe itself executes the verified file.
const platformRequiresExecutableExtension = true

func interpreterExecutable(mode fs.FileMode) bool { return mode.IsRegular() }
