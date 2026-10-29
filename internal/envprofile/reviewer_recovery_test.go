package envprofile

import "testing"

// An edited marker after an actual credential-prefix stop must survive recovery
// refusal byte-for-byte, together with every pending sidecar and native fixture.
func TestReviewerRecoveryMarkerDrift(t *testing.T) { migrationTransactionCase(t, "marker-drift") }

// A similarly named unrelated regular file never becomes transaction-owned.
// A nearby genuine interrupted unit still recovers and reaches current homes.
func TestReviewerRecoveryPreservesRegularTemp(t *testing.T) {
	migrationTransactionCase(t, "unrelated-temp")
}
