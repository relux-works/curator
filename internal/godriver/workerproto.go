package godriver

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"strings"
)

// WorkerMode is the fixed hidden mode that re-executes the installed manager
// as the go-v1 build worker. It is an implementation boundary, not a
// user-visible command surface: no package file, manifest value, descriptor
// value, environment value, PATH lookup, shell, or user option selects it.
const WorkerMode = "__curator-go-worker-v1"

// protocolVersion pins the framing and message vocabulary of one session.
const protocolVersion = "curator-go-worker-v1"

// maxProtocolFrame bounds every framed message. The list result carries the
// bounded go list stream, so the ceiling is above the largest permitted
// combined-output bound and below any value that could exhaust the worker.
const maxProtocolFrame = int64(64 * 1024 * 1024)

// Message kinds. The session state machine admits no other kind.
const (
	kindRequest     = "request"
	kindReady       = "ready"
	kindList        = "list"
	kindListResult  = "list-result"
	kindPermit      = "permit"
	kindBuildResult = "build-result"
	kindShutdown    = "shutdown"
	kindFailure     = "failure"
)

// wireLimits is the bounded resource request carried to the worker.
type wireLimits struct {
	TimeoutMillis int64 `json:"timeout_millis"`
	OutputBytes   int64 `json:"output_bytes"`
	ArtifactBytes int64 `json:"artifact_bytes"`
	FileBytes     int64 `json:"file_bytes"`
	DiskBytes     int64 `json:"disk_bytes"`
	MemoryBytes   int64 `json:"memory_bytes"`
	Processes     int   `json:"processes"`
}

// workerRequest is the single length-bounded canonical request. Every value is
// manager-owned; package bytes select nothing here.
type workerRequest struct {
	Version          string         `json:"version"`
	Secret           string         `json:"secret"`
	ExecutablePath   string         `json:"executable_path"`
	ExecutableSHA256 string         `json:"executable_sha256"`
	ExecutableSize   int64          `json:"executable_size"`
	GoExecutable     string         `json:"go_executable"`
	GOROOT           string         `json:"goroot"`
	ToolDirectory    string         `json:"tool_directory"`
	Directory        string         `json:"directory"`
	Environment      []string       `json:"environment"`
	ListArgv         []string       `json:"list_argv"`
	BuildArgv        []string       `json:"build_argv"`
	ArtifactPath     string         `json:"artifact_path"`
	ReadOnlyRoots    []string       `json:"readonly_roots"`
	PrivateRoots     []string       `json:"private_roots"`
	Platform         string         `json:"platform"`
	Probes           []ControlProbe `json:"probes"`
	Limits           wireLimits     `json:"limits"`
}

// workerReady proves the worker executable identity and reports the controls
// the worker confirms are in effect, together with the closed
// capability-evidence-v1 record it derives from them. The manager requires that
// record to be identical to the one it derived from what it installed.
type workerReady struct {
	ExecutablePath   string             `json:"executable_path"`
	ExecutableSHA256 string             `json:"executable_sha256"`
	ExecutableSize   int64              `json:"executable_size"`
	Applied          []string           `json:"applied"`
	Evidence         CapabilityEvidence `json:"evidence"`
}

// workerResult is one bounded Go child result. Started counts how many times the
// worker used its single permitted process-creation site in this session, so a
// creation attempt the operating system refused still consumes one and an extra
// program is still detected. StartFailed marks the refusal, which leaves no exit
// status at all.
type workerResult struct {
	Stdout      []byte `json:"stdout"`
	Stderr      []byte `json:"stderr"`
	ExitCode    int    `json:"exit_code"`
	Started     int    `json:"started"`
	TimedOut    bool   `json:"timed_out"`
	Overflow    bool   `json:"overflow"`
	StartFailed bool   `json:"start_failed"`
	Detail      string `json:"detail"`
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
	Permit  string         `json:"permit,omitempty"`
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
	case kindRequest, kindReady, kindList, kindListResult, kindPermit, kindBuildResult, kindShutdown, kindFailure:
	default:
		return workerMessage{}, diagnostic(CodeWorkerProtocolInvalid, "unknown session message kind %q", message.Kind)
	}
	return message, nil
}

// buildPermit derives the authenticated fixed build permit. It binds the
// session secret, the session nonce, and the exact build argument vector, so a
// permit cannot be replayed into another session or cover another vector.
func buildPermit(secret, nonce string, argv []string) string {
	key, err := hex.DecodeString(secret)
	if err != nil {
		key = []byte(secret)
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte("curator-go-build-permit\x00"))
	mac.Write([]byte(nonce))
	mac.Write([]byte{0})
	mac.Write([]byte(strings.Join(argv, "\x00")))
	return hex.EncodeToString(mac.Sum(nil))
}

func validPermit(permit, secret, nonce string, argv []string) bool {
	expected := buildPermit(secret, nonce, argv)
	return hmac.Equal([]byte(expected), []byte(permit))
}
