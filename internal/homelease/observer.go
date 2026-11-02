package homelease

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"regexp"
	"time"

	"github.com/relux-works/curator/internal/homelease/observation"
)

var errObservationUnavailable = errors.New("curator-run observation unavailable")
var errObservationIncomplete = errors.New("curator-run observation lacks home-lease incarnation and reconciliation proofs")
var managerSelector = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,31}$`)
var sessionSelector = regexp.MustCompile(`^SES-[A-Za-z0-9]{1,48}$`)

type sessionManagerObserver struct{ executable string }

// SessionManagerObserver reads only curator-run's public observe command.
// The current v1 output cannot authorize bind/release: it lacks exact OS
// incarnation, request/Werkstatt binding, descendants/ownership and no-start
// proof. This adapter keeps that uncertainty rather than borrowing caller data.
func SessionManagerObserver() Observer { return sessionManagerObserver{executable: "curator-run"} }

type boundedObservation struct{ bytes.Buffer }

func (b *boundedObservation) Write(p []byte) (int, error) {
	if len(p) > observation.MaxBytes-b.Len() {
		return 0, errObservationUnavailable
	}
	return b.Buffer.Write(p)
}

func (o sessionManagerObserver) Observe(ctx context.Context, h Holder) (Evidence, error) {
	// The published command requires a session; prepared no-start evidence
	// cannot be requested through it. Never issue a different manager command.
	if h.Kind != "bound" || h.Session == nil || !managerSelector.MatchString(h.ManagerID) || !sessionSelector.MatchString(*h.Session) {
		return Evidence{}, errObservationUnavailable
	}
	// The producer allows 10s for describe and another 10s for status.
	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, o.executable, "session-manager", "observe", h.ManagerID, "--session", *h.Session, "--json")
	cmd.WaitDelay = time.Second
	var stdout boundedObservation
	cmd.Stdout, cmd.Stderr = &stdout, io.Discard
	if cmd.Run() != nil {
		return Evidence{}, errObservationUnavailable
	}
	r, err := observation.Parse(stdout.Bytes())
	if err != nil {
		return Evidence{}, errObservationUnavailable
	}
	if r.Manager.ID != h.ManagerID || r.Manager.RegisteredAt != h.ManagerRegisteredAt || r.Session.Manager != h.ManagerID || r.Session.Session != *h.Session || !r.Session.Readable || r.Session.Generation == nil || h.Generation == nil || *r.Session.Generation != *h.Generation {
		return Evidence{}, errObservationUnavailable
	}
	// Neither "down" nor any other state proves process identity or complete
	// termination. Do not copy h.Process, h.RequestID or labels into evidence.
	return Evidence{}, errObservationIncomplete
}
