// Package scriptpolicy implements the manager side of the schema-8 enforced
// script execution policy `script-worker-v1` (Spec §4.1.1, manager profile
// §3.6).
//
// curator parses the policy and does not implement it. Those are two separate
// statements and the spec treats them separately: a schema-8 manifest that
// selects the policy is a valid document, so `skillspec.Load` accepts it and
// the published schema cases that mark it valid stay green. Admission is the
// manager's own answer, and the profile leaves it no room —
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
// Refusal is unconditional because this manager has no worker. When the worker
// lands, this package is where the refusal is replaced by real containment;
// nothing above it has to move, because nothing above it decides the policy.
package scriptpolicy

import (
	"errors"
	"sort"

	"github.com/relux-works/curator/internal/skillspec"
)

// PolicyUnsupported is the closed §4.1.1 diagnostic for an enforced command
// read by a manager that does not implement `script-worker-v1`. State and
// Severity are the manager profile's fixed pair for it.
const (
	PolicyUnsupported = "script_execution_policy_unsupported"
	StateUnsupported  = "unsupported"
	SeverityError     = "error"
)

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
		return &Error{
			DiagnosticCode: PolicyUnsupported,
			State:          StateUnsupported,
			Severity:       SeverityError,
			Path:           "commands." + name + ".execution_policy",
			Detail: "this manager does not implement " + command.ExecutionPolicy +
				", and the policy forbids installing the command declared-only, downgrading it, or ignoring the field",
		}
	}
	return nil
}
