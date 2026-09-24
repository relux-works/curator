// Command forgeworker is a test-only lying script worker. It speaks the
// session framing just well enough to send a forged ready proof (or a
// forged result) and to record whether the parent answered with a permit,
// so tests prove the parent validates the proof before permitting the
// run. Behaviour is keyed by the request runtime entry suffix:
//
//   - "<entry>.forge-started": prove the true identities, attach the true
//     capability evidence, await the permit, then answer a result with
//     Started=2.
//   - "<entry>.forge-interp": lie about the interpreter identity only.
//   - "<entry>.forge-evidence-<mutation>": prove the true identities but
//     attach a mutated capability evidence record, where <mutation> is one
//     of the twelve ready-record mutations below.
//   - "<entry>.forge-evidence-second-record": prove the true identities,
//     attach the true record, await the permit, then answer a result that
//     carries a second record.
//   - anything else: lie about the manager identity.
//
// Every observed parent frame is recorded beside the runtime entry as
// "<entry>.forge-<kind>", so the test reads what the parent sent without
// any other channel.
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/relux-works/curator/internal/godriver"
)

type inventoryInput struct {
	Name         string `json:"name"`
	Availability string `json:"availability"`
	Present      bool   `json:"present"`
}

type request struct {
	ExecutablePath   string           `json:"executable_path"`
	ExecutableSHA256 string           `json:"executable_sha256"`
	ExecutableSize   int64            `json:"executable_size"`
	InterpreterSHA   string           `json:"interpreter_sha256"`
	InterpreterSize  int64            `json:"interpreter_size"`
	RuntimeEntry     string           `json:"runtime_entry"`
	Platform         string           `json:"inventory_platform"`
	Inventory        []inventoryInput `json:"inventory"`
}

type evidenceEntry struct {
	Name         string `json:"name"`
	Availability string `json:"availability"`
	Status       string `json:"status"`
	ProbedAt     string `json:"probed_at"`
}

type evidence struct {
	RecordVersion   string          `json:"record_version"`
	ExecutionPolicy string          `json:"execution_policy"`
	Platform        string          `json:"platform"`
	Controls        []evidenceEntry `json:"controls"`
}

type ready struct {
	ExecutablePath   string    `json:"executable_path"`
	ExecutableSHA256 string    `json:"executable_sha256"`
	ExecutableSize   int64     `json:"executable_size"`
	InterpreterSHA   string    `json:"interpreter_sha256"`
	InterpreterSize  int64     `json:"interpreter_size"`
	Evidence         *evidence `json:"evidence,omitempty"`
}

type result struct {
	Stdout   []byte    `json:"stdout"`
	Stderr   []byte    `json:"stderr"`
	ExitCode int       `json:"exit_code"`
	Started  int       `json:"started"`
	Overflow bool      `json:"overflow"`
	Evidence *evidence `json:"evidence,omitempty"`
}

type message struct {
	Kind    string   `json:"kind"`
	Nonce   string   `json:"nonce"`
	Request *request `json:"request,omitempty"`
	Ready   *ready   `json:"ready,omitempty"`
	Result  *result  `json:"result,omitempty"`
}

func main() {
	os.Exit(run())
}

func run() int {
	incoming, err := readFrame(os.Stdin)
	if err != nil || incoming.Kind != "request" || incoming.Request == nil {
		return 2
	}
	entry := incoming.Request.RuntimeEntry
	record := func(kind string) {
		_ = os.WriteFile(entry+".forge-"+kind, []byte(kind+"\n"), 0o644)
	}
	zero := "sha256:" + strings.Repeat("0", 64)
	proof := &ready{
		ExecutablePath:   incoming.Request.ExecutablePath,
		ExecutableSHA256: zero,
		ExecutableSize:   incoming.Request.ExecutableSize,
		InterpreterSHA:   incoming.Request.InterpreterSHA,
		InterpreterSize:  incoming.Request.InterpreterSize,
	}
	truthful := func() bool {
		self, digest, size, ok := selfIdentity()
		if !ok {
			return false
		}
		proof.ExecutablePath = self
		proof.ExecutableSHA256 = digest
		proof.ExecutableSize = size
		return true
	}
	trueRecord := func() *evidence {
		return trueEvidence(incoming.Request.Platform, incoming.Request.Inventory)
	}
	switch {
	case strings.HasSuffix(entry, ".forge-started"):
		// Prove the true identities so the parent permits the run, then
		// lie in the result instead.
		if !truthful() {
			return 2
		}
		proof.Evidence = trueRecord()
		if err := writeFrame(os.Stdout, message{Kind: "ready", Nonce: incoming.Nonce, Ready: proof}); err != nil {
			return 2
		}
		next, err := readFrame(os.Stdin)
		if err != nil {
			return 2
		}
		record(next.Kind)
		if next.Kind == "permit" {
			_ = writeFrame(os.Stdout, message{
				Kind: "result", Nonce: incoming.Nonce,
				Result: &result{Stdout: []byte("{}"), Started: 2},
			})
		}
		return 0
	case strings.HasSuffix(entry, ".forge-evidence-second-record"):
		// Prove everything truthfully so the parent permits, then
		// answer a result carrying a second record.
		if !truthful() {
			return 2
		}
		proof.Evidence = trueRecord()
		if err := writeFrame(os.Stdout, message{Kind: "ready", Nonce: incoming.Nonce, Ready: proof}); err != nil {
			return 2
		}
		next, err := readFrame(os.Stdin)
		if err != nil {
			return 2
		}
		record(next.Kind)
		if next.Kind == "permit" {
			second := trueRecord()
			_ = writeFrame(os.Stdout, message{
				Kind: "result", Nonce: incoming.Nonce,
				Result: &result{Stdout: []byte("{}"), Started: 1, Evidence: second},
			})
		}
		return 0
	case evidenceMutation(entry) != "":
		// Prove the true identities but attach a mutated record, so only
		// the evidence gate refuses.
		if !truthful() {
			return 2
		}
		mutated, ok := mutateEvidence(trueRecord(), evidenceMutation(entry), incoming.Request.Inventory)
		if !ok {
			_ = os.WriteFile(entry+".forge-no-candidate", []byte(evidenceMutation(entry)+"\n"), 0o644)
			return 2
		}
		proof.Evidence = mutated
	case strings.HasSuffix(entry, ".forge-interp"):
		// Lie about the interpreter identity only: the manager proof is
		// true, so only the interpreter expectation refuses.
		if !truthful() {
			return 2
		}
		proof.InterpreterSHA = zero
		proof.InterpreterSize = 1
		proof.Evidence = trueRecord()
	default:
		proof.Evidence = trueRecord()
	}
	if err := writeFrame(os.Stdout, message{Kind: "ready", Nonce: incoming.Nonce, Ready: proof}); err != nil {
		return 2
	}
	next, err := readFrame(os.Stdin)
	if err != nil {
		return 2
	}
	record(next.Kind)
	return 0
}

// evidenceMutation extracts the ready-record mutation suffix, or "".
func evidenceMutation(entry string) string {
	const prefix = ".forge-evidence-"
	index := strings.LastIndex(entry, prefix)
	if index < 0 {
		return ""
	}
	mutation := entry[index+len(prefix):]
	if mutation == "" || mutation == "second-record" {
		return ""
	}
	return mutation
}

// trueEvidence builds the truthful record for the request inventory: every
// installable control applied, every other control unavailable.
func trueEvidence(platform string, inventory []inventoryInput) *evidence {
	record := &evidence{
		RecordVersion:   "script-capability-evidence-v1",
		ExecutionPolicy: "script-worker-v1",
		Platform:        platform,
	}
	for _, input := range inventory {
		status := "unavailable"
		if input.Present && (input.Availability == "available" || input.Availability == "host-conditional") {
			status = "applied"
		}
		record.Controls = append(record.Controls, evidenceEntry{
			Name: input.Name, Availability: input.Availability, Status: status, ProbedAt: "pre-worker-launch",
		})
	}
	return record
}

// mutateEvidence applies one vector evidence mutation to the truthful
// record. It reports false when the request inventory offers no candidate
// for the mutation.
func mutateEvidence(record *evidence, mutation string, inventory []inventoryInput) (*evidence, bool) {
	firstWith := func(availability string, present *bool) int {
		for index, input := range inventory {
			if input.Availability != availability {
				continue
			}
			if present != nil && input.Present != *present {
				continue
			}
			return index
		}
		return -1
	}
	absent := false
	switch mutation {
	case "available-status-unavailable":
		index := firstWith("available", nil)
		if index < 0 {
			return nil, false
		}
		record.Controls[index].Status = "unavailable"
	case "unavailable-status-applied":
		index := firstWith("unavailable", nil)
		if index < 0 {
			return nil, false
		}
		record.Controls[index].Status = "applied"
	case "remove-control":
		if len(record.Controls) == 0 {
			return nil, false
		}
		record.Controls = record.Controls[:len(record.Controls)-1]
	case "duplicate-control":
		if len(record.Controls) == 0 {
			return nil, false
		}
		record.Controls = append(record.Controls, record.Controls[0])
	case "add-unknown-control":
		record.Controls = append(record.Controls, evidenceEntry{
			Name: "script-fast-start", Availability: "available", Status: "applied", ProbedAt: "pre-worker-launch",
		})
	case "record-version-v2":
		record.RecordVersion = "script-capability-evidence-v2"
	case "probe-unavailable-status-applied":
		index := firstWith("host-conditional", &absent)
		if index < 0 {
			return nil, false
		}
		record.Controls[index].Status = "applied"
	case "probed-at-install":
		if len(record.Controls) == 0 {
			return nil, false
		}
		record.Controls[0].ProbedAt = "at-install"
	case "capability-evidence-v1":
		record.RecordVersion = "capability-evidence-v1"
	case "manager-worker-v1":
		record.ExecutionPolicy = "manager-worker-v1"
	case "script-total-network-denial":
		record.Controls = append(record.Controls, evidenceEntry{
			Name: "script-total-network-denial", Availability: "available", Status: "applied", ProbedAt: "pre-worker-launch",
		})
	case "total-network-denial":
		record.Controls = append(record.Controls, evidenceEntry{
			Name: "total-network-denial", Availability: "available", Status: "applied", ProbedAt: "pre-worker-launch",
		})
	case "reorder-controls":
		// Every entry is individually valid; only the order contradicts
		// the parent's derivation, so only the identity comparison
		// refuses.
		for left, right := 0, len(record.Controls)-1; left < right; left, right = left+1, right-1 {
			record.Controls[left], record.Controls[right] = record.Controls[right], record.Controls[left]
		}
	default:
		return nil, false
	}
	return record, true
}

// selfIdentity hashes this helper's own canonical executable through the
// same primitive the parent resolves with, so the truthful modes prove
// what the parent's launch-boundary check expects on every platform.
func selfIdentity() (string, string, int64, bool) {
	path, err := os.Executable()
	if err != nil {
		return "", "", 0, false
	}
	resolved, err := godriver.CanonicalPhysicalPath(path)
	if err != nil {
		return "", "", 0, false
	}
	payload, err := os.ReadFile(resolved)
	if err != nil {
		return "", "", 0, false
	}
	digest := sha256.Sum256(payload)
	return resolved, "sha256:" + hex.EncodeToString(digest[:]), int64(len(payload)), true
}

func readFrame(reader io.Reader) (message, error) {
	var header [4]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return message{}, err
	}
	length := binary.BigEndian.Uint32(header[:])
	if length == 0 || length > 64*1024*1024 {
		return message{}, errors.New("frame too large")
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return message{}, err
	}
	var decoded message
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return message{}, err
	}
	return decoded, nil
}

func writeFrame(writer io.Writer, outgoing message) error {
	payload, err := json.Marshal(outgoing)
	if err != nil {
		return err
	}
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(payload)))
	if _, err := writer.Write(header[:]); err != nil {
		return err
	}
	_, err = writer.Write(payload)
	return err
}
