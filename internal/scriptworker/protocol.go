package scriptworker

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
)

// WorkerMode is the fixed hidden mode that re-executes the installed manager
// as the script-worker-v1 worker. It is an implementation boundary, not a
// user-visible command surface: no package file, manifest value, environment
// value, PATH lookup, shell, or user option selects it.
const WorkerMode = "__curator-script-worker-v1"

// protocolVersion pins the framing and message vocabulary of one session.
const protocolVersion = "curator-script-worker-v1"

// maxProtocolFrame bounds every framed message. The result frame carries the
// bounded captured child output, so the ceiling sits above the largest
// permitted capture and below any value that could exhaust the worker.
const maxProtocolFrame = int64(64 * 1024 * 1024)

// sessionNonceLength is the exact length of the hex-encoded fresh nonce that
// binds every message of one session.
const sessionNonceLength = 64

// Message kinds. The session state machine admits no other kind.
const (
	kindRequest  = "request"
	kindReady    = "ready"
	kindPermit   = "permit"
	kindResult   = "result"
	kindShutdown = "shutdown"
	kindFailure  = "failure"
)

// workerRequest is the single length-bounded canonical request. Every value
// is manager-owned; package bytes select nothing here.
type workerRequest struct {
	Version          string   `json:"version"`
	ExecutablePath   string   `json:"executable_path"`
	ExecutableSHA256 string   `json:"executable_sha256"`
	ExecutableSize   int64    `json:"executable_size"`
	InterpreterID    string   `json:"interpreter_id"`
	InterpreterPath  string   `json:"interpreter_path"`
	InterpreterSHA   string   `json:"interpreter_sha256"`
	InterpreterSize  int64    `json:"interpreter_size"`
	RuntimeEntry     string   `json:"runtime_entry"`
	Args             []string `json:"args"`
	Environment      []string `json:"environment"`
	WorkingDir       string   `json:"working_dir"`
	PrivateBase      string   `json:"private_base"`
	PrivateTmp       string   `json:"private_tmp"`
	PrivateConfig    string   `json:"private_config"`
	PrivateCache     string   `json:"private_cache"`
	FarmDir          string   `json:"farm_dir"`
	NetworkOffline   bool     `json:"network_offline"`
	ProjectRoot      string   `json:"project_root"`
	Stdin            []byte   `json:"stdin"`
	StdinIsNull      bool     `json:"stdin_is_null"`
	OutputLimit      int64    `json:"output_limit"`
	// InventoryPlatform is the probed inventory platform identifier, and
	// Inventory is this invocation's parent-side probe list. Availability
	// is probed once, before the worker launches; the worker applies and
	// confirms exactly this list and returns the resulting record with
	// its acknowledgement.
	InventoryPlatform string                 `json:"inventory_platform"`
	Inventory         []ScriptInventoryInput `json:"inventory"`
	// WritePaths is the derived path set filesystem-write-confinement
	// grants: absolute paths beneath the canonical project root, plus the
	// operation-private area the worker adds itself.
	WritePaths []string `json:"write_paths"`
	// CgroupPath is the prepared invocation cgroup, relative to the
	// hierarchy root, or "" when no cgroup control is installable.
	// CgroupRoot overrides the hierarchy root the worker confirms
	// against; "" selects the production hierarchy. Tests point it at a
	// fixture hierarchy through the parent's cgroup seam.
	CgroupPath string `json:"cgroup_path"`
	CgroupRoot string `json:"cgroup_root"`
	// FileSizeBound is the exact per-file soft bound the worker must
	// observe, or 0 when per-file-size-limit is not installable.
	FileSizeBound uint64 `json:"file_size_bound"`
}

// ScriptInventoryInput is one parent-probed inventory control as the worker
// receives it: the closed name, the probed availability, and whether the
// probe found the mechanism present for this invocation.
type ScriptInventoryInput struct {
	Name         string `json:"name"`
	Availability string `json:"availability"`
	Present      bool   `json:"present"`
}

// workerReady proves the worker executable identity and acknowledges the
// session nonce. The interpreter verification it reports is the worker's own
// launch-boundary check, performed before the interpreter starts, and the
// evidence record is the worker's own confirmation of the installed
// controls. The parent validates both before permitting the run.
type workerReady struct {
	ExecutablePath   string          `json:"executable_path"`
	ExecutableSHA256 string          `json:"executable_sha256"`
	ExecutableSize   int64           `json:"executable_size"`
	InterpreterSHA   string          `json:"interpreter_sha256"`
	InterpreterSize  int64           `json:"interpreter_size"`
	Evidence         *ScriptEvidence `json:"evidence,omitempty"`
}

// workerResult is one bounded interpreter result. Started counts how many
// times the worker used its single permitted process-creation site in this
// session, so an extra program below the worker is detected. Overflow marks
// a capture bound the session enforced; the output it carries is truncated.
// Evidence must stay nil: the invocation already returned its one record
// with the acknowledgement, and a second record would be a contradiction
// rather than information.
type workerResult struct {
	Stdout   []byte          `json:"stdout"`
	Stderr   []byte          `json:"stderr"`
	ExitCode int             `json:"exit_code"`
	Started  int             `json:"started"`
	Overflow bool            `json:"overflow"`
	Evidence *ScriptEvidence `json:"evidence,omitempty"`
}

// workerFailure carries one stable diagnostic across the session boundary.
type workerFailure struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

// workerMessage is the only framed value on the session channel.
type workerMessage struct {
	Kind    string         `json:"kind"`
	Nonce   string         `json:"nonce"`
	Request *workerRequest `json:"request,omitempty"`
	Ready   *workerReady   `json:"ready,omitempty"`
	Result  *workerResult  `json:"result,omitempty"`
	Failure *workerFailure `json:"failure,omitempty"`
}

var errProtocolClosed = errors.New("worker session channel closed")

// writeMessage writes one length-prefixed canonical frame.
func writeMessage(writer io.Writer, message workerMessage) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	if int64(len(payload)) > maxProtocolFrame {
		return diagnostic(CodeWorkerProtocolInvalid, "outgoing %s frame exceeds the session bound", message.Kind)
	}
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(payload))) // #nosec G115 -- the length was just bounded by maxProtocolFrame
	if _, err := writer.Write(header[:]); err != nil {
		return err
	}
	_, err = writer.Write(payload)
	return err
}

// readMessage reads one length-prefixed canonical frame. An oversize declared
// length is rejected before any payload byte is buffered.
func readMessage(reader io.Reader) (workerMessage, error) {
	var header [4]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		if errors.Is(err, io.EOF) {
			return workerMessage{}, errProtocolClosed
		}
		return workerMessage{}, diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot read a session frame header")
	}
	length := int64(binary.BigEndian.Uint32(header[:]))
	if length == 0 || length > maxProtocolFrame {
		return workerMessage{}, diagnostic(CodeWorkerProtocolInvalid, "session frame length %d is outside the bound", length)
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return workerMessage{}, diagnosticErr(CodeWorkerProtocolInvalid, err, "cannot read a session frame payload")
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var message workerMessage
	if err := decoder.Decode(&message); err != nil {
		return workerMessage{}, diagnosticErr(CodeWorkerProtocolInvalid, err, "session frame is not a known message")
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return workerMessage{}, diagnostic(CodeWorkerProtocolInvalid, "session frame has trailing data")
	}
	switch message.Kind {
	case kindRequest, kindReady, kindPermit, kindResult, kindShutdown, kindFailure:
	default:
		return workerMessage{}, diagnostic(CodeWorkerProtocolInvalid, "unknown session message kind %q", message.Kind)
	}
	return message, nil
}
