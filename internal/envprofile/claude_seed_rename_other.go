//go:build !darwin && !linux && !windows

package envprofile

// Other platforms retain the existing publication behavior: preparation inside
// replaceManagedEntry can still occur after the final seed comparison. Do not
// introduce a new platform refusal merely to eliminate that residual window.
func prepareClaudeSeedRename(source, target string) (func() error, func(), error) {
	return func() error { return replaceManagedEntry(source, target) }, func() {}, nil
}
