//go:build darwin

package closureexec

// NewOSBoundary fails closed on Darwin. sandbox-exec enforces a policy but
// does not return the lossless process, filesystem, network, and output event
// stream required by EnforceObserveProvider. A caller with an entitled
// Endpoint Security implementation can inject that provider into NewExecutor.
func NewOSBoundary(_, _ string) (EnforceObserveProvider, error) {
	return nil, failure("closure_derivation_unauthorized", "lossless Darwin enforce-and-observe provider unavailable")
}
