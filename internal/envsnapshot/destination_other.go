//go:build !darwin && !linux

package envsnapshot

// Destination is unavailable until the platform's retained-handle owner/ACL
// backend is qualified. In particular, Unix mode synthesis is not Windows ACL
// evidence. This is a refusal, not a Windows support claim.
type Destination struct{}

// OpenDestination refuses without probing or modifying the supplied path.
func OpenDestination(string, DestinationPolicy) (*Destination, error) {
	return nil, invalidDestination("platform-unqualified")
}

// Kind returns no observation on an unqualified platform.
func (*Destination) Kind() DestinationKind { return "" }

// Revalidate refuses on an unqualified platform.
func (*Destination) Revalidate() error { return invalidDestination("platform-unqualified") }

// MatchPrior refuses on an unqualified platform.
func (*Destination) MatchPrior([]byte) (bool, error) {
	return false, invalidDestination("platform-unqualified")
}

// Close has no resources to release on an unqualified platform.
func (*Destination) Close() error { return nil }
