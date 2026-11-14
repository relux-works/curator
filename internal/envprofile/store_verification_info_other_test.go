//go:build !windows

package envprofile

import "os"

func storeVerificationFileInfo(path string) (os.FileInfo, error) { return os.Lstat(path) }
