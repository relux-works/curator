//go:build !darwin && !linux

package projectcontext

import (
	"fmt"
	"os"
)

func filesystemIdentity(os.FileInfo) (map[string]string, error) {
	return nil, fmt.Errorf("project_context_unsupported: filesystem binding is qualified only on macOS and Linux")
}

func openCandidate(root *os.Root, name string) (*os.File, error) { return root.Open(name) }

func singleLink(os.FileInfo) bool { return false }
