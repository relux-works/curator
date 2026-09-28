//go:build !unix && !windows

package pathboundary

import "fmt"

func makeWorldWritableDirectoryForTest(string) (func() error, error) {
	return nil, fmt.Errorf("platform has no protected-boundary mutation helper")
}
