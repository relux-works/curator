//go:build !linux

package buildcache

func nativeExecutableSnapshot() (executableSnapshot, error) {
	paths, err := liveExecutablePaths()
	return executableSnapshot{paths: paths}, err
}

// Only Linux uses write denial as an extra signal; native enumeration on other
// platforms retains its existing all-or-nothing behavior.
func entryExecutionInUse(_ *sweepRoot, _ string, _ executableSnapshot) (bool, error) {
	return false, nil
}
