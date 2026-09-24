// Package scriptworker implements the enforced `script-worker-v1` command
// launch path of Protocol Core §4.1.1 and manager profile §3.1: the fixed
// hidden-mode re-execution of the installed manager as the script worker,
// interpreter resolution from operator-trusted machine configuration, and
// worker-domain teardown.
//
// The fixed process graph of one enforced invocation is:
//
//	manager parent
//	  -> identity-verified manager-owned script worker
//	       -> identity-verified interpreter for the declared identifier
//
// Manager-resolved `exec` names are copied into the PATH farm only when the
// declaration names them. Unresolved names stay absent from the farm and
// appear in the invocation report.
//
// Launch is the production entry. It preflights the mandatory portable
// controls through scriptpolicy before anything starts: while that table
// names missing controls the invocation refuses with
// `script_execution_control_unavailable` and no worker exists. The worker
// session behind the preflight is real production code — resolve and bind
// the interpreter, recheck identity at the launch boundary, prove identity
// over a fresh session nonce, bind streams explicitly, run the interpreter
// against the commit-keyed runtime entry with verbatim arguments, return
// the child's exit status, and terminate the complete worker domain — and
// is proven at the real process boundary by this package's tests calling
// the session directly. There is no environment, flag, or configuration
// switch that opens the preflight: it opens only when the control table is
// complete.
package scriptworker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/relux-works/curator/internal/scriptpolicy"
)

// Stable execution-boundary diagnostics of manager profile §3.1. The
// policy-level `script_execution_control_unavailable` and
// `script_execution_policy_unsupported` live in scriptpolicy, which owns
// admission; the three worker-session codes live here, mirroring the
// equivalent `build_execution_*` codes of §2.2.1.
const (
	CodeWorkerIdentityInvalid     = "script_execution_worker_identity_invalid"
	CodeWorkerProtocolInvalid     = "script_execution_worker_protocol_invalid"
	CodePackageInfluenceForbidden = "script_execution_package_influence_forbidden"
)

// Diagnostic is a stable, machine-testable failure at a script execution
// boundary. Detail is intended for operators; callers branch on Code.
type Diagnostic struct {
	Code   string
	Detail string
	Err    error
}

func (err *Diagnostic) Error() string {
	message := "script-worker-v1 " + err.Code
	if err.Detail != "" {
		message += ": " + err.Detail
	}
	return message
}

func (err *Diagnostic) Unwrap() error { return err.Err }

func diagnostic(code, format string, args ...any) error {
	return &Diagnostic{Code: code, Detail: fmt.Sprintf(format, args...)}
}

func diagnosticErr(code string, err error, format string, args ...any) error {
	return &Diagnostic{Code: code, Detail: fmt.Sprintf(format, args...), Err: err}
}

// DiagnosticCode returns the stable script-worker code carried by err, or an
// empty string when err did not originate at a script trust boundary. It also
// recognizes the admission codes scriptpolicy owns, so one call classifies
// every refusal of this execution path.
func DiagnosticCode(err error) string {
	var failure *Diagnostic
	if errors.As(err, &failure) {
		return failure.Code
	}
	return scriptpolicy.Code(err)
}

// LaunchRequest is the complete manager-owned launch input. The interpreter
// mapping is the operator-trusted machine configuration, the runtime entry
// is the manager-derived commit-keyed path, and the capabilities are the
// declared manifest bytes derivation reads. The environment, PATH,
// working directory, and private roots are derived from those bytes at
// launch; no caller-supplied environment reaches the worker.
type LaunchRequest struct {
	// ManagerPath names the installed manager executable to re-execute. Empty
	// resolves the running process, which is what production passes.
	ManagerPath string
	// Interpreters is the operator-trusted identifier-to-path mapping from
	// machine configuration. It is the only source the interpreter is
	// resolved from.
	Interpreters map[string]string
	// ForbiddenRoots are package-controlled or runtime roots the
	// interpreter and exec names must never resolve under: snapshots, the
	// runtime store, .agents/bin.
	ForbiddenRoots []string
	// InterpreterID is the command's declared closed identifier.
	InterpreterID string
	// RuntimeEntry is the manager-derived commit-keyed script path the
	// interpreter runs.
	RuntimeEntry string
	// Args are forwarded to the interpreter verbatim after the runtime entry.
	Args []string
	// CapabilitiesRaw is the declared `capabilities` object the
	// containment profile derives from. Nil derives every field absent.
	CapabilitiesRaw json.RawMessage
	// HostEnvironment is the manager's own environment, read for
	// non-reserved env_read passthrough. Nil reads the process
	// environment, which is what production passes.
	HostEnvironment []string
	// ExecSearchDirs overrides the manager-owned exec search list. Nil
	// uses DefaultExecSearchDirs; tests inject fixture directories.
	ExecSearchDirs []string
	// ProjectRoot is the canonical project root of the invocation, or ""
	// when the invocation has none (global scope).
	ProjectRoot string
	// Stdin is the explicit standard-input payload. Nil binds the platform
	// null device instead of an inherited descriptor.
	Stdin []byte
	// PrivateBase holds the operation-private runtime area. It must be an
	// absolute manager-owned directory outside every forbidden root.
	PrivateBase string
}

// Result is one completed enforced invocation.
type Result struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
	// Report records the derivation decisions behind the invocation:
	// network mode, recorded hosts, withheld names, resolved and
	// unresolved exec names, and secret identifiers.
	Report DerivationReport
	// Evidence is the invocation's closed capability-evidence record:
	// exactly one per invocation, result-only, validated by the parent
	// before the permit.
	Evidence ScriptEvidence
}

// Launch runs one enforced script invocation through the fixed
// manager-owned worker. It preflights the mandatory portable controls
// first: the implementation table must be complete, and this host must
// provide every mandatory control, or the invocation refuses with
// `script_execution_control_unavailable` before any worker starts and runs
// nothing. Behind the preflight the containment profile derives from the
// declared capabilities and the session applies it. The context is an
// operator or invocation-level bound, never a policy control.
func Launch(ctx context.Context, request LaunchRequest) (Result, error) {
	if missing := scriptpolicy.MissingControls(); len(missing) != 0 {
		return Result{}, scriptpolicy.ControlError("")
	}
	return runSession(ctx, request, nil)
}

// PreflightHostControls probes the native-control inventory for one
// install or update of an enforced command and refuses with
// `script_execution_control_unavailable` when this host cannot provide a
// mandatory control, before anything is published. It is an additional
// moment, not a substitute: every invocation probes again.
func PreflightHostControls() error {
	_, _, err := probeScriptInventory()
	return err
}
