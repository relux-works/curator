//go:build darwin || linux

package projectadapter

import (
	"golang.org/x/sys/unix"
	"os"
)

func openGenerated(root *os.Root, name string) (*os.File, error) {
	return root.OpenFile(name, os.O_RDONLY|unix.O_NONBLOCK|unix.O_NOFOLLOW, 0)
}
