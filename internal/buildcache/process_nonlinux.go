//go:build !linux

package buildcache

func nativeExecutableSnapshot() (executableSnapshot, error) {
	paths, err := liveExecutablePaths()
	return executableSnapshot{paths: paths}, err
}
