package scriptworker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/scriptpolicy"
)

// ShimSidecarSuffix is the filename suffix of the manager-published sidecar
// that pairs an enforced native launcher with its invocation contract. The
// sidecar path is always the launcher path plus this suffix, with no
// platform-dependent stem surgery.
const ShimSidecarSuffix = ".curator-shim.json"

// ShimSidecarVersion is the only sidecar layout this manager reads.
const ShimSidecarVersion = 1

// ShimSidecar is the install-published invocation contract of one enforced
// native launcher: the skill and command identity, the closed interpreter
// identifier, the manager-derived commit-keyed runtime entry and tree, the
// canonical project root ("" for global scope), and the declared
// capabilities bytes derivation reads. The launcher re-validates every
// field before use; a tampered or corrupt sidecar refuses fail-closed.
type ShimSidecar struct {
	Version       int             `json:"version"`
	Skill         string          `json:"skill"`
	Command       string          `json:"command"`
	Interpreter   string          `json:"interpreter"`
	RuntimeEntry  string          `json:"runtime_entry"`
	RuntimeDir    string          `json:"runtime_dir"`
	ProjectRoot   string          `json:"project_root"`
	SchemaVersion int             `json:"schema_version"`
	Capabilities  json.RawMessage `json:"capabilities"`
}

// NewShimSidecar builds the sidecar for one enforced command install. The
// capabilities are the raw declared bytes, re-validated here so install
// stores only what derivation can read.
func NewShimSidecar(skill, command, interpreter, runtimeEntry, runtimeDir, projectRoot string, schemaVersion int, capabilities json.RawMessage) (ShimSidecar, error) {
	sidecar := ShimSidecar{
		Version:       ShimSidecarVersion,
		Skill:         skill,
		Command:       command,
		Interpreter:   interpreter,
		RuntimeEntry:  runtimeEntry,
		RuntimeDir:    runtimeDir,
		ProjectRoot:   projectRoot,
		SchemaVersion: schemaVersion,
		Capabilities:  append(json.RawMessage(nil), capabilities...),
	}
	if err := sidecar.validate(); err != nil {
		return ShimSidecar{}, err
	}
	return sidecar, nil
}

// Marshal renders the sidecar for the install transaction.
func (sidecar ShimSidecar) Marshal() ([]byte, error) {
	payload, err := json.MarshalIndent(sidecar, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}

// LoadShimSidecar reads and validates one installed sidecar. Every field
// is re-checked: the sidecar is manager-published install state, and
// anything it carries that is not well-formed refuses rather than widening
// the invocation.
func LoadShimSidecar(path string) (ShimSidecar, error) {
	payload, err := os.ReadFile(path) // #nosec G304 -- manager-published sidecar beside the launcher
	if err != nil {
		return ShimSidecar{}, diagnosticErr(CodeWorkerProtocolInvalid, err,
			"cannot read the enforced launcher sidecar")
	}
	var sidecar ShimSidecar
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&sidecar); err != nil {
		return ShimSidecar{}, diagnosticErr(CodePackageInfluenceForbidden, err,
			"the enforced launcher sidecar is not a known contract")
	}
	if err := sidecar.validate(); err != nil {
		return ShimSidecar{}, err
	}
	return sidecar, nil
}

func (sidecar ShimSidecar) validate() error {
	if sidecar.Version != ShimSidecarVersion {
		return diagnostic(CodePackageInfluenceForbidden,
			"the enforced launcher sidecar version %d is not readable", sidecar.Version)
	}
	if !identifiers.Valid(sidecar.Skill) || !identifiers.Valid(sidecar.Command) {
		return diagnostic(CodePackageInfluenceForbidden,
			"the enforced launcher sidecar names an invalid skill or command")
	}
	if !closedInterpreters[sidecar.Interpreter] {
		return diagnostic(CodePackageInfluenceForbidden,
			"the enforced launcher sidecar names interpreter %q outside the closed set", sidecar.Interpreter)
	}
	if sidecar.RuntimeEntry == "" || !filepath.IsAbs(sidecar.RuntimeEntry) {
		return diagnostic(CodePackageInfluenceForbidden,
			"the enforced launcher sidecar runtime entry is not absolute")
	}
	if sidecar.RuntimeDir == "" || !filepath.IsAbs(sidecar.RuntimeDir) {
		return diagnostic(CodePackageInfluenceForbidden,
			"the enforced launcher sidecar runtime directory is not absolute")
	}
	if sidecar.ProjectRoot != "" && !filepath.IsAbs(sidecar.ProjectRoot) {
		return diagnostic(CodePackageInfluenceForbidden,
			"the enforced launcher sidecar project root is not absolute")
	}
	if sidecar.SchemaVersion < 8 {
		return diagnostic(CodePackageInfluenceForbidden,
			"the enforced launcher sidecar schema %d cannot carry an enforced command", sidecar.SchemaVersion)
	}
	if _, err := ParseDeclaredCapabilities(sidecar.Capabilities); err != nil {
		return err
	}
	return nil
}

// SidecarPathForExecutable derives the sidecar path paired with one
// launcher executable: the executable path plus the sidecar suffix.
func SidecarPathForExecutable(executable string) string {
	return executable + ShimSidecarSuffix
}

// ShimSidecarFor reports whether executable is an enforced native launcher
// by the presence of its manager-published sidecar. The sidecar is the
// whole signal: a launcher without one is an ordinary binary, and the
// executable name alone never selects shim mode.
func ShimSidecarFor(executable string) (string, bool) {
	if executable == "" {
		return "", false
	}
	sidecar := SidecarPathForExecutable(executable)
	info, err := os.Stat(sidecar)
	if err != nil || !info.Mode().IsRegular() {
		return "", false
	}
	return sidecar, true
}

// ShimRequest is one enforced launcher invocation. The executable replays
// the manager role: it derives the containment profile from the sidecar
// contract and runs the fixed worker session.
type ShimRequest struct {
	// ExePath is the launcher executable itself: the manager identity the
	// session re-executes in the fixed hidden worker mode.
	ExePath string
	// Args are forwarded to the interpreter verbatim after the runtime entry.
	Args []string
	// Stdin is the launcher's standard input. It is forwarded as the
	// explicit session payload unless StdinIsTerminal, which binds the
	// null device: a live terminal cannot be pre-read into a payload.
	Stdin io.Reader
	// StdinIsTerminal reports whether Stdin is a character device.
	StdinIsTerminal bool
	Stdout          io.Writer
	Stderr          io.Writer
	// Environ is the manager's own environment, read for non-reserved
	// env_read passthrough.
	Environ []string
	// Interpreters is the operator-trusted identifier-to-path mapping
	// from machine configuration.
	Interpreters map[string]string
	// DiagnosticsDir is the operator-selected machine-local directory
	// that receives the invocation's result-only record. "" disables
	// reporting. Package data can never choose it: it arrives from
	// operator machine configuration only.
	DiagnosticsDir string
}

// RunShim runs one enforced native launcher invocation to completion and
// returns the process exit code: the interpreter's own status on success,
// or 1 with a stable diagnostic on the launcher's standard error when the
// session itself refuses. Captured child output is written verbatim.
func RunShim(request ShimRequest) int {
	sidecarPath, ok := ShimSidecarFor(request.ExePath)
	if !ok {
		_, _ = fmt.Fprintf(request.Stderr, "script-worker-v1 %s: no enforced launcher sidecar beside this executable\n", CodeWorkerProtocolInvalid)
		return 1
	}
	sidecar, err := LoadShimSidecar(sidecarPath)
	if err != nil {
		writeShimDiagnostic(request.Stderr, err)
		return 1
	}
	stdin, err := shimStdin(request.Stdin, request.StdinIsTerminal)
	if err != nil {
		writeShimDiagnostic(request.Stderr, err)
		return 1
	}
	binDir := filepath.Dir(request.ExePath)
	// The interpreter and exec names must never resolve under
	// package-controlled or runtime roots: the staged runtime tree, the
	// installed binary directory, and the checked-out project. The
	// snapshot store root itself is not named in the sidecar contract,
	// so operator bindings live outside every repository by hygiene;
	// the staged runtime tree it contains is covered here.
	forbidden := []string{sidecar.RuntimeDir, binDir}
	if sidecar.ProjectRoot != "" {
		forbidden = append(forbidden, sidecar.ProjectRoot)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	result, err := Launch(ctx, LaunchRequest{
		ManagerPath:     request.ExePath,
		Interpreters:    request.Interpreters,
		ForbiddenRoots:  forbidden,
		InterpreterID:   sidecar.Interpreter,
		RuntimeEntry:    sidecar.RuntimeEntry,
		Args:            append([]string(nil), request.Args...),
		CapabilitiesRaw: sidecar.Capabilities,
		HostEnvironment: request.Environ,
		ProjectRoot:     sidecar.ProjectRoot,
		Stdin:           stdin,
		PrivateBase:     os.TempDir(),
	})
	if err != nil {
		writeShimDiagnostic(request.Stderr, err)
		return 1
	}
	writeInvocationRecord(request.DiagnosticsDir, sidecar, result)
	if _, err := request.Stdout.Write(result.Stdout); err != nil {
		writeShimDiagnostic(request.Stderr, diagnosticErr(CodeWorkerProtocolInvalid, err,
			"cannot forward the interpreter standard output"))
		return 1
	}
	if _, err := request.Stderr.Write(result.Stderr); err != nil {
		return 1
	}
	return result.ExitCode
}

// shimStdin resolves the explicit session input. Piped or file input is
// read fully into the session payload, bounded by the session frame; a
// terminal binds the null device instead, because a live terminal has no
// end to pre-read. Transparent terminal streaming needs the pass-through
// binding owned by R3.
func shimStdin(input io.Reader, isTerminal bool) ([]byte, error) {
	if isTerminal || input == nil {
		return nil, nil
	}
	payload, err := io.ReadAll(io.LimitReader(input, maxProtocolFrame+1))
	if err != nil {
		return nil, diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot read the launcher standard input")
	}
	if int64(len(payload)) > maxProtocolFrame {
		return nil, diagnostic(CodeWorkerProtocolInvalid, "launcher standard input exceeds the session bound")
	}
	return payload, nil
}

// invocationRecord is the result-only diagnostic document of one enforced
// invocation: the closed capability-evidence record plus the derivation
// decisions behind it. It carries identifiers and decisions only — never
// the command's standard output or standard error.
type invocationRecord struct {
	Evidence   ScriptEvidence   `json:"evidence"`
	Derivation DerivationReport `json:"derivation"`
}

// writeInvocationRecord persists the invocation's result-only record at
// the operator-selected diagnostic destination: at most the most recent
// record per command, named by the validated skill and command identity.
// A destination failure is reported nowhere the caller's pipeline reads:
// the invocation already completed, so there is no refusal to return and
// no stream to write. "" disables reporting.
func writeInvocationRecord(directory string, sidecar ShimSidecar, result Result) {
	if directory == "" {
		return
	}
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return
	}
	payload, err := json.MarshalIndent(invocationRecord{Evidence: result.Evidence, Derivation: result.Report}, "", "  ")
	if err != nil {
		return
	}
	name := sidecar.Skill + "-" + sidecar.Command + "." + ScriptEvidenceVersion + ".json"
	_ = os.WriteFile(filepath.Join(directory, name), append(payload, '\n'), 0o600)
}

func writeShimDiagnostic(writer io.Writer, err error) {
	code := DiagnosticCode(err)
	if code == "" {
		code = CodeWorkerProtocolInvalid
	}
	// A Diagnostic already renders as `script-worker-v1 <code>: <detail>`,
	// and a policy refusal already leads with its closed code, so neither
	// is prefixed again here.
	var failure *Diagnostic
	if errors.As(err, &failure) {
		if failure.Detail == "" {
			_, _ = fmt.Fprintf(writer, "script-worker-v1 %s\n", code)
		} else {
			_, _ = fmt.Fprintf(writer, "script-worker-v1 %s: %s\n", code, failure.Detail)
		}
		return
	}
	if scriptpolicy.Code(err) != "" {
		_, _ = fmt.Fprintf(writer, "%s\n", err.Error())
		return
	}
	_, _ = fmt.Fprintf(writer, "script-worker-v1 %s: %s\n", code, err.Error())
}
