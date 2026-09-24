// Package scriptpolicy implements the manager side of the schema-8 enforced
// script execution policy `script-worker-v1` (Spec §4.1.1, manager profile
// §3.6).
//
// curator parses the policy and admits it in two tiers. Those are two separate
// statements and the spec treats them separately: a schema-8 manifest that
// selects the policy is a valid document, so `skillspec.Load` accepts it and
// the published schema cases that mark it valid stay green. Admission is the
// manager's own answer, and the profile leaves it no room for an unknown
// policy —
//
//	A manager that does not implement this policy MUST reject such a command
//	with `script_execution_policy_unsupported`. It MUST NOT install the
//	command declared-only, downgrade it, or ignore the field, because the
//	resulting shim would run package code the manifest says is contained.
//
// — so admission lives here, one layer below the parser. Every surface that
// would turn a declared command into an installed shim, and every surface that
// merely reads a package's commands back to an operator, asks this package
// first and fails closed on an enforced command.
//
// For the implemented `script-worker-v1` policy the manager enters the spec's
// preflight instead: it enumerates the 11 mandatory portable controls in
// MandatoryControls, and while any control is unimplemented the install and
// the invocation refuse with `script_execution_control_unavailable` naming
// the missing controls, before any worker starts. No enforced script launches
// uncontained while that list is non-empty.
package scriptpolicy

import (
	"errors"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/skillspec"
)

// Closed §4.1.1 diagnostics for an enforced command. State and Severity are
// the manager profile's fixed pairs for them.
const (
	// PolicyUnsupported is read by a manager that does not implement the
	// selected execution policy at all.
	PolicyUnsupported = "script_execution_policy_unsupported"
	// ControlUnavailable is read when the policy is implemented but a
	// mandatory portable control cannot be applied on this host, at shim
	// install or update and again before worker launch.
	ControlUnavailable = "script_execution_control_unavailable"
	StateUnsupported   = "unsupported"
	SeverityError      = "error"
)

// MandatoryControl is one entry of the 11-control implementation table. The
// order is the vector order of `mandatory_controls` in
// script-host-execution-policy.json. Owner names the residual leaf that
// delivers an unimplemented control; it is empty once Implemented is true.
type MandatoryControl struct {
	Name        string
	Implemented bool
	Owner       string
}

// mandatoryControls is the 11-control implementation table in vector
// order. R3 implements the two evidence controls, so the table is
// complete: every install and every invocation of an enforced command
// proceeds to the per-invocation host probe, and refuses with
// `script_execution_control_unavailable` only when this host cannot
// provide a mandatory control. The R2 table-injection seam is removed:
// there is no override, so no test or production path can force the
// table, and MissingControls is empty in every process.
var mandatoryControls = []MandatoryControl{
	{Name: "fixed-process-graph", Implemented: true},
	{Name: "worker-identity-verification", Implemented: true},
	{Name: "interpreter-resolution-and-identity-verification", Implemented: true},
	{Name: "manager-built-environment", Implemented: true},
	{Name: "manager-built-path", Implemented: true},
	{Name: "offline-network-configuration", Implemented: true},
	{Name: "operation-private-runtime-area", Implemented: true},
	{Name: "explicit-standard-stream-binding", Implemented: true},
	{Name: "inventory-controls-applied", Implemented: true},
	{Name: "closed-script-capability-evidence-record", Implemented: true},
	{Name: "worker-domain-teardown", Implemented: true},
}

// MandatoryControls returns a copy of the 11-control implementation table
// in vector order. The copy keeps callers from mutating the table the
// preflight enforces.
func MandatoryControls() []MandatoryControl {
	return append([]MandatoryControl(nil), mandatoryControls...)
}

// MissingControls lists the mandatory portable controls this build cannot yet
// apply, in vector order. A non-empty list means the worker must not launch.
func MissingControls() []string {
	var missing []string
	for _, control := range MandatoryControls() {
		if !control.Implemented {
			missing = append(missing, control.Name)
		}
	}
	return missing
}

// ControlError builds the preflight refusal naming the missing controls.
// Path is the manifest field path at install, or "" at invocation where no
// manifest field is being read.
func ControlError(path string) *Error {
	return &Error{
		DiagnosticCode: ControlUnavailable,
		State:          StateUnsupported,
		Severity:       SeverityError,
		Path:           path,
		Detail: "this manager cannot apply mandatory controls: " +
			strings.Join(MissingControls(), ", ") +
			"; enforced launch is refused until the control set is complete",
	}
}

// HostControlError builds the host-preflight refusal naming the mandatory
// controls this host cannot provide. Path is the manifest field path at
// install, or "" at invocation where no manifest field is being read;
// reason is the probe outcome (unavailable, or a probe failure). Unlike
// ControlError, which names unimplemented controls, this names implemented
// controls the host does not provide, so install and invocation refuse
// without starting any worker.
func HostControlError(path string, unavailable []string, reason string) *Error {
	detail := "this host cannot provide mandatory controls: " +
		strings.Join(append([]string(nil), unavailable...), ", ") +
		"; enforced launch is refused (" + reason + ")"
	return &Error{
		DiagnosticCode: ControlUnavailable,
		State:          StateUnsupported,
		Severity:       SeverityError,
		Path:           path,
		Detail:         detail,
	}
}

// Error is an execution-policy refusal bound to a closed diagnostic. Path is
// the manifest field path that carries the policy, so an operator is pointed
// at the field rather than at the command.
type Error struct {
	DiagnosticCode string
	State          string
	Severity       string
	Path           string
	Detail         string
}

func (err *Error) Error() string {
	message := err.DiagnosticCode
	if err.Detail != "" {
		message += ": " + err.Detail
	}
	if err.Path != "" {
		return err.Path + ": " + message
	}
	return message
}

// Code returns the stable execution-policy diagnostic carried by err, or an
// empty string when err did not originate at this boundary.
func Code(err error) string {
	var diagnostic *Error
	if errors.As(err, &diagnostic) {
		return diagnostic.DiagnosticCode
	}
	return ""
}

// Enforced reports whether a command opted into an execution policy. Only
// script commands can carry one, and `script-worker-v1` is the single closed
// value, so any non-empty policy is an enforced command.
func Enforced(command skillspec.Command) bool {
	return command.ExecutionPolicy != ""
}

// Admit checks one skill's commands and refuses the first enforced one in
// command-lexical order, so the reported command is the same on every host and
// every run. A nil return means every command is declared-only and installs
// through the ordinary schema-7 path unchanged.
//
// An unknown policy is refused with `script_execution_policy_unsupported`
// exactly as before. The implemented `script-worker-v1` policy enters the
// preflight: while MandatoryControls names missing controls the refusal is
// `script_execution_control_unavailable` instead, and no shim is published.
func Admit(commands map[string]skillspec.Command) error {
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		command := commands[name]
		if !Enforced(command) {
			continue
		}
		if command.ExecutionPolicy != skillspec.ScriptExecutionPolicy {
			return &Error{
				DiagnosticCode: PolicyUnsupported,
				State:          StateUnsupported,
				Severity:       SeverityError,
				Path:           "commands." + name + ".execution_policy",
				Detail: "this manager does not implement " + command.ExecutionPolicy +
					", and the policy forbids installing the command declared-only, downgrading it, or ignoring the field",
			}
		}
		if missing := MissingControls(); len(missing) != 0 {
			return ControlError("commands." + name + ".execution_policy")
		}
	}
	return nil
}
