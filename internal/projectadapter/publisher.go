package projectadapter

import (
	"context"
	"fmt"

	"github.com/relux-works/curator/internal/identifiers"
)

// PreparedPublisher is I-LEASE's published prepared-lease publication method.
// The lease owner verifies current authority and the exact final receipt and
// owns transactional publication. The adapter never acquires/releases a lease.
type PreparedPublisher interface {
	PublishPrepared(context.Context, string, map[string][]byte, bool) error
}

// PublishPrepared contributes immutable bytes to the existing lease owner.
// The consumer must bind this plan's final receipt before preparing the lease;
// this method cannot convert an old strict composition into native-project.
func (p *Plan) PublishPrepared(ctx context.Context, leaseID string, publisher PreparedPublisher) error {
	if publisher == nil || !identifiers.Valid(leaseID) {
		return fmt.Errorf("project_lease_mismatch: prepared lease publisher and identifier required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := p.RequireLaunch(); err != nil {
		return err
	}
	return publisher.PublishPrepared(ctx, leaseID, p.Files(), false)
}
