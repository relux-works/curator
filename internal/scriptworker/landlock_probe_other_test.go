//go:build !linux

package scriptworker

import "testing"

// Process-group cancellation and subreaper verification are Linux-only; other
// platforms never launch a Landlock availability probe.
func TestLandlockProbeCancellationKillsProcessGroup(t *testing.T) {
	t.Skip("Landlock probe cancellation is a Linux-only control in script-worker-v1-native-control-inventory-v1 (requires process groups and a subreaper)")
}
