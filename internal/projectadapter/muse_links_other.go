//go:build !darwin && !linux

package projectadapter

import "os"

// Offline filesystem qualification is restricted to Darwin and Linux.
func museSingleLink(os.FileInfo) bool { return false }
