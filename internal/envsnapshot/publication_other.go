//go:build !darwin && !linux

package envsnapshot

import "context"

// Publish refuses until retained-handle admission and no-replace publication
// are qualified together. There is no pathname or copy fallback.
func (*Destination) Publish(context.Context, Publication) (*Verified, error) {
	return nil, invalidDestination("platform-unqualified")
}
