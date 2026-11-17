package homelease

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/relux-works/curator/internal/homelease/observation"
	"github.com/relux-works/curator/internal/homelease/state"
)

var errObservationUnavailable = errors.New("curator-run observation unavailable")
var errObservationIncomplete = errors.New("curator-run observation lacks home-lease incarnation and reconciliation proofs")

// ErrProcessIncarnationMismatch marks any unequal field among the requested
// bound holder, the authoritative observation and Curator's own kernel
// reading. Bind maps it to the closed process-incarnation mismatch row.
var ErrProcessIncarnationMismatch = errors.New("process incarnation mismatch")

var managerSelector = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,31}$`)
var sessionSelector = regexp.MustCompile(`^SES-[A-Za-z0-9]{1,48}$`)

// BindVerifier is implemented by an observer that can independently prove a
// requested bound holder against the authoritative observation and the
// kernel. The manager's assertion alone never satisfies it.
type BindVerifier interface {
	VerifyBind(ctx context.Context, prepared, requested Holder) error
}

type sessionManagerObserver struct {
	executable string
	// kernel reads one incarnation independently; tests substitute a fake
	// source without changing the production transport.
	kernel func(pid int64) (state.Process, error)
	// frozenExecutable, when set, supplies the executable frozen in the
	// launcher's operation binding for --expect-executable. The holder
	// itself carries no executable member.
	frozenExecutable func(prepared, requested Holder) (string, bool)
}

// SessionManagerObserver reads only curator-run's public observe command and
// proves binds through Curator's own kernel reading. It never speaks the
// manager plugin protocol directly.
func SessionManagerObserver() Observer {
	return sessionManagerObserver{executable: "curator-run", kernel: readKernelProcess}
}

type boundedObservation struct{ bytes.Buffer }

func (b *boundedObservation) Write(p []byte) (int, error) {
	if len(p) > observation.MaxBytes-b.Len() {
		return 0, errObservationUnavailable
	}
	return b.Buffer.Write(p)
}

// observeV2 runs the read-only v2 observation with the frozen binding
// expectations: registered_at from the prepared holder, generation from the
// requested bound holder and, when a frozen executable is supplied, that
// executable. The returned record's identity members are compared against
// the same frozen values before any process member is read.
func (o sessionManagerObserver) observeV2(ctx context.Context, prepared, requested Holder) (observation.RecordV2, error) {
	var none observation.RecordV2
	if requested.Kind != "bound" || requested.Session == nil || requested.Generation == nil ||
		prepared.ManagerID != requested.ManagerID || !managerSelector.MatchString(requested.ManagerID) ||
		!sessionSelector.MatchString(*requested.Session) {
		return none, errObservationUnavailable
	}
	args := []string{"session-manager", "observe", requested.ManagerID,
		"--session", *requested.Session, "--json", "--schema", "v2",
		"--expect-registered-at", prepared.ManagerRegisteredAt,
		"--expect-generation", strconv.FormatInt(*requested.Generation, 10)}
	if o.frozenExecutable != nil {
		if executable, ok := o.frozenExecutable(prepared, requested); ok && executable != "" {
			args = append(args, "--expect-executable", executable)
		}
	}
	// The producer allows 10s for describe and another 10s for status.
	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, o.executable, args...)
	cmd.WaitDelay = time.Second
	var stdout boundedObservation
	cmd.Stdout, cmd.Stderr = &stdout, io.Discard
	if cmd.Run() != nil {
		return none, errObservationUnavailable
	}
	record, err := observation.ParseV2(stdout.Bytes())
	if err != nil {
		return none, errObservationUnavailable
	}
	if record.Manager.ID != requested.ManagerID || record.Manager.RegisteredAt != prepared.ManagerRegisteredAt ||
		record.Session.Manager != requested.ManagerID || record.Session.Session != *requested.Session ||
		!record.Session.Readable || record.Session.Generation == nil || *record.Session.Generation != *requested.Generation {
		return none, errObservationUnavailable
	}
	return record, nil
}

// VerifyBind obtains unavailable evidence first and compares incarnations
// only once both the authoritative observation and Curator's own kernel
// reading exist. All four fields must be byte-equal across the requested
// holder, the observation tuple and the kernel reading.
func (o sessionManagerObserver) VerifyBind(ctx context.Context, prepared, requested Holder) error {
	if requested.Process == nil || requested.Process.IsLegacy() || requested.Process.StartTime == nil {
		return errObservationUnavailable
	}
	if nativeStartTimeKind() == "" || requested.Process.StartTime.Kind != nativeStartTimeKind() {
		// A well-formed variant this platform cannot read natively is
		// unsupported evidence, never a coerced comparison.
		return errObservationUnavailable
	}
	record, err := o.observeV2(ctx, prepared, requested)
	if err != nil {
		return err
	}
	if record.Session.Process == nil {
		return errObservationUnavailable
	}
	kernel := o.kernel
	if kernel == nil {
		kernel = readKernelProcess
	}
	observed, err := kernel(requested.Process.PID)
	if err != nil {
		if errors.Is(err, errKernelChanged) {
			return ErrProcessIncarnationMismatch
		}
		return errObservationUnavailable
	}
	if !state.ProcessEqual(record.Session.Process, requested.Process) || !kernelProcessEqual(observed, *requested.Process) {
		return ErrProcessIncarnationMismatch
	}
	return nil
}

// kernelProcessEqual compares Curator's own reading field by field, with
// the tagged birth identity as its own comparison step.
func kernelProcessEqual(kernel, requested state.Process) bool {
	return kernel.PID == requested.PID && kernel.UID == requested.UID &&
		kernel.BootID == requested.BootID && state.StartTimeEqual(kernel.StartTime, requested.StartTime)
}

// Observe serves release reconciliation only. The first use adds no release
// authority: a v2 process or termination member is an input to future
// release rules, not proof of terminal session, terminated descendants or
// no-start. Until that authoritative integration exists, every release
// observation stays incomplete and the lease keeps blocking.
func (o sessionManagerObserver) Observe(ctx context.Context, h Holder) (Evidence, error) {
	// The published command requires a session; prepared no-start evidence
	// cannot be requested through it. Never issue a different manager command.
	if h.Kind != "bound" || h.Session == nil || h.Generation == nil || !managerSelector.MatchString(h.ManagerID) || !sessionSelector.MatchString(*h.Session) {
		return Evidence{}, errObservationUnavailable
	}
	prepared := h
	prepared.Kind, prepared.Session, prepared.Generation, prepared.Process = "prepared", nil, nil, nil
	if _, err := o.observeV2(ctx, prepared, h); err != nil {
		return Evidence{}, err
	}
	// Neither "down" nor any other state proves process identity or complete
	// termination. Do not copy h.Process, h.RequestID or labels into evidence.
	return Evidence{}, errObservationIncomplete
}
