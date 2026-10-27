// Package skillruntime stages Slice A's isolated machine-current command domain.
// Trusted supervisor adapters supply admission; profiles and ambient environment
// cannot grant it. This package supplies no supervisor or final-child qualification.
package skillruntime

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/stateread"
)

// Unavailable is the required Slice A B1 diagnostic for non-current profiles.
const Unavailable = "profile commands unavailable: slice A serves only the machine-current profile"

// CompanionRequired keeps allocation and qualification separate from adoption.
const CompanionRequired = "profile_skill_companion_required"

// RoutingUnproven refuses mutation without exclusive supervisor authority.
const RoutingUnproven = "profile_skill_routing_unproven"

// DomainDir is provisioned by the trusted supervisor, before any runtime journal.
// Its presence fences legacy writers, including recovery of prepared journals.
// Creating this directory alone is never admission.
func DomainDir(home string) string { return filepath.Join(home, "profile-command-domain") }

// RefuseUnadmitted fences legacy writers/recovery even when no shim has swapped
// yet. Domain enrollment is a supervisor responsibility, not a side effect of
// compilation or journal recovery.
func RefuseUnadmitted(home string) error {
	state, err := stateread.Lstat(DomainDir(home))
	if err != nil {
		return fmt.Errorf("%s: routing-domain boundary unreadable: %w", RoutingUnproven, err)
	}
	if state.Kind != stateread.KindAbsent {
		return fmt.Errorf("%s: exclusive routing admission required", RoutingUnproven)
	}
	return nil
}

// Request binds an admission to a manager and the protected observed current.
// The supervisor must exclude every other writer/consumer, drain old-profile
// consumers for a switch, and hold that exclusion until Close.
type Request struct{ Home, Profile, Current string }

// Supervisor is implemented by a trusted manager/supervisor adapter, never by
// a package manifest. There is deliberately no environment/boolean fallback.
type Supervisor interface {
	Acquire(context.Context, Request) (Admission, error)
}

// Admission independently pins native filesystem identities, a fixed PATH,
// current-profile routing and the admitted consumer set. Revalidate must check
// all of them without enrolling consumers or repairing state.
type Admission interface {
	Domain() Domain
	Revalidate() error
	Close() error
}

// Domain is the supervisor-owned routing description, not profile configuration.
type Domain struct{ Home, UserHome, UserBin, FixedPATH string }

// CommandChecks are mandatory companion seams. Inventory must check the union
// of locked/current machine protections (including shell/manager witnesses).
// BasePATH must use the destination's admitted PATH, with exact owned-forwarder
// subtraction, and refuse missing/unsafe authority. A producer PATH is invalid.
// No implementation or runtime qualification is implied by this interface.
type CommandChecks interface {
	Inventory(Binding, []Export) error
	BasePATH(Domain, Binding, []Export) error
}

// Session keeps the supervisor exclusion outside and before the home lock.
type Session struct {
	request   Request
	domain    Domain
	admission Admission
	checks    CommandChecks
}

// Open acquires and checks admission before the caller takes the home lock.
func Open(ctx context.Context, request Request, supervisor Supervisor, checks CommandChecks) (*Session, error) {
	if supervisor == nil {
		return nil, fmt.Errorf("%s: supervisor admission required", RoutingUnproven)
	}
	if checks == nil {
		return nil, fmt.Errorf("%s: protected inventory and admitted base PATH required", CompanionRequired)
	}
	admission, err := supervisor.Acquire(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("%s: acquire exclusive routing admission: %w", RoutingUnproven, err)
	}
	if admission == nil {
		return nil, fmt.Errorf("%s: missing exclusive routing admission", RoutingUnproven)
	}
	session := &Session{request: request, domain: admission.Domain(), admission: admission, checks: checks}
	if err := session.Revalidate(); err != nil {
		_ = admission.Close()
		return nil, err
	}
	return session, nil
}

// Close releases consumer/writer exclusion after the complete operation.
func (s *Session) Close() error { return s.admission.Close() }

// Domain returns the captured immutable routing description.
func (s *Session) Domain() Domain { return s.domain }

// Request returns the current/profile observations covered by this admission.
func (s *Session) Request() Request { return s.request }

// Revalidate checks current exclusion and the protected routing boundaries.
func (s *Session) Revalidate() error {
	if s == nil || s.admission == nil || s.checks == nil {
		return fmt.Errorf("%s: missing routing admission", RoutingUnproven)
	}
	if s.domain != s.admission.Domain() {
		return fmt.Errorf("%s: routing admission changed", RoutingUnproven)
	}
	if err := s.admission.Revalidate(); err != nil {
		return fmt.Errorf("%s: %w", RoutingUnproven, err)
	}
	d := s.domain
	if d.Home != s.request.Home || !filepath.IsAbs(d.Home) || !filepath.IsAbs(d.UserHome) || !filepath.IsAbs(d.UserBin) || filepath.Clean(d.UserBin) != d.UserBin {
		return fmt.Errorf("%s: mismatched routing domain", RoutingUnproven)
	}
	// Only a distinct private user-bin already present in the supervisor's fixed
	// PATH qualifies. We neither discover a user-bin nor change the process PATH.
	relative, err := filepath.Rel(d.UserHome, d.UserBin)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("%s: user-bin outside admitted home", RoutingUnproven)
	}
	found := false
	for _, entry := range filepath.SplitList(d.FixedPATH) {
		if entry == d.UserBin {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("%s: private user-bin missing from fixed PATH", RoutingUnproven)
	}
	for _, route := range [][2]string{{d.Home, DomainDir(d.Home)}, {d.Home, filepath.Join(d.Home, "global", "bin")}, {d.UserHome, d.UserBin}} {
		if err := pathboundary.ValidateRouteWithOwner(route[0], route[1], pathboundary.DefaultOwnerLookup()); err != nil {
			return fmt.Errorf("%s: unsafe routing boundary: %w", RoutingUnproven, err)
		}
	}
	inside, err := filepath.Rel(d.Home, d.UserBin)
	if err != nil || (inside != ".." && !strings.HasPrefix(inside, ".."+string(filepath.Separator))) {
		return fmt.Errorf("%s: forwarding root overlaps manager", RoutingUnproven)
	}
	return nil
}
