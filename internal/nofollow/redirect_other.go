//go:build !windows

package nofollow

import "io/fs"

func redirect(info fs.FileInfo) bool { return info.Mode()&fs.ModeSymlink != 0 }
