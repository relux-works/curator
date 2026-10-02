//go:build !darwin && !linux && !windows

package buildcache

import "fmt"

func liveExecutablePaths() ([]string, error) {
	return nil, fmt.Errorf("process image enumeration is unavailable on this platform")
}
