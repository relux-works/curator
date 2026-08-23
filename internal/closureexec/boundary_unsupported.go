//go:build !darwin

package closureexec

// NewOSBoundary fails closed until this platform has a lossless authoritative
// enforcement-and-observation implementation. Platform providers remain
// pluggable through NewExecutor.
func NewOSBoundary(_, _ string) (EnforceObserveProvider, error) {
	return nil, failure("closure_derivation_unauthorized", "lossless enforce-and-observe provider unavailable on this platform")
}
