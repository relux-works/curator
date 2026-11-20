package observation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/homelease/state"
	"github.com/relux-works/curator/internal/protocoljson"
)

// RecordV2 is the authority-bearing projection of one
// curator-run-session-observation-v2 reply. Only the members the lease
// contract compares are extracted; process and termination are closed, and
// their presence rules follow the observation contract exactly.
type RecordV2 struct {
	Manager struct {
		ID           string
		Executable   string
		RegisteredAt string
	}
	Session struct {
		Session     string
		Manager     string
		Readable    bool
		State       string
		Generation  *int64
		Process     *state.Process
		Termination *Termination
	}
	ProcessObservation bool
}

type Termination struct {
	Scope   string `json:"scope"`
	Witness string `json:"witness"`
}

var (
	v2ManagerID    = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,31}$`)
	v2RegisteredAt = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]+)?(Z|[+-]00:00)$`)
	v2ObservedAt   = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$`)
	v2SessionID    = regexp.MustCompile(`^SES-[A-Za-z0-9]{1,48}$`)
)

var processStates = map[string]bool{"starting": true, "idle": true, "busy": true, "draining": true}

// ParseV2 validates the closed v2 envelope and every authority-bearing
// member before extraction. A process member without the advertised
// process-observation capability, on an unreadable record, or outside its
// allowed states is a protocol error, never stripped input.
func ParseV2(raw []byte) (RecordV2, error) {
	var out RecordV2
	invalid := fmt.Errorf("invalid curator-run observation v2")
	if len(raw) > MaxBytes || protocoljson.Validate(raw) != nil {
		return out, invalid
	}
	envelope, err := objectFields(raw)
	if err != nil || !exactMembers(envelope, "schema", "manager", "session", "observed_at", "capabilities") {
		return out, invalid
	}
	schema, err := stringMember(envelope, "schema")
	if err != nil || schema != "curator-run-session-observation-v2" {
		return out, invalid
	}
	observedAt, err := stringMember(envelope, "observed_at")
	if err != nil || !v2ObservedAt.MatchString(observedAt) {
		return out, invalid
	}
	manager, err := objectFields(envelope["manager"])
	if err != nil || !exactMembers(manager, "id", "executable", "registered_at") {
		return out, invalid
	}
	if out.Manager.ID, err = stringMember(manager, "id"); err != nil || !v2ManagerID.MatchString(out.Manager.ID) {
		return out, invalid
	}
	if out.Manager.Executable, err = stringMember(manager, "executable"); err != nil ||
		!strings.HasPrefix(out.Manager.Executable, "/") || strings.ContainsRune(out.Manager.Executable, 0) {
		return out, invalid
	}
	if out.Manager.RegisteredAt, err = stringMember(manager, "registered_at"); err != nil || !v2RegisteredAt.MatchString(out.Manager.RegisteredAt) {
		return out, invalid
	}
	capabilities, err := objectFields(envelope["capabilities"])
	if err != nil || !exactMembers(capabilities, "process_observation") {
		return out, invalid
	}
	if err := json.Unmarshal(capabilities["process_observation"], &out.ProcessObservation); err != nil {
		return out, invalid
	}
	session, err := objectFields(envelope["session"])
	if err != nil || !closedSession(session) {
		return out, invalid
	}
	if out.Session.Session, err = stringMember(session, "session"); err != nil || !v2SessionID.MatchString(out.Session.Session) {
		return out, invalid
	}
	if out.Session.Manager, err = stringMember(session, "manager"); err != nil {
		return out, invalid
	}
	if err := json.Unmarshal(session["readable"], &out.Session.Readable); err != nil {
		return out, invalid
	}
	if out.Session.State, err = stringMember(session, "state"); err != nil {
		return out, invalid
	}
	switch out.Session.State {
	case "starting", "idle", "busy", "draining", "down", "unknown":
	default:
		return out, invalid
	}
	if rawGeneration, ok := session["generation"]; ok && string(bytes.TrimSpace(rawGeneration)) != "null" {
		generation, err := strconv.ParseInt(string(bytes.TrimSpace(rawGeneration)), 10, 64)
		if err != nil || generation < 1 || generation > protocoljson.MaxSafeInteger {
			return out, invalid
		}
		out.Session.Generation = &generation
	}
	hasProcess := false
	if rawProcess, ok := session["process"]; ok {
		hasProcess = true
		if !out.ProcessObservation || !out.Session.Readable || !processStates[out.Session.State] {
			return out, invalid
		}
		var process state.Process
		if err := json.Unmarshal(rawProcess, &process); err != nil || process.IsLegacy() ||
			state.ValidateProcessSemantics(&process) != nil {
			return out, invalid
		}
		out.Session.Process = &process
	}
	if rawTermination, ok := session["termination"]; ok {
		if hasProcess || !out.ProcessObservation || !out.Session.Readable || out.Session.State != "down" {
			return out, invalid
		}
		var termination Termination
		fields, err := objectFields(rawTermination)
		if err != nil || !exactMembers(fields, "scope", "witness") || json.Unmarshal(rawTermination, &termination) != nil {
			return out, invalid
		}
		known := (termination.Scope == "linux-cgroup" && termination.Witness == "populated-0") ||
			(termination.Scope == "dedicated-uid" && termination.Witness == "zero-process-proof")
		if !known {
			return out, invalid
		}
		out.Session.Termination = &termination
	}
	return out, nil
}

// sessionRequired is the closed member set every session-record carries
// (readable and unreadable variants alike, per the vendored schema);
// sessionOptional adds the conditional draining member and the two v2
// evidence members. A session object with a member outside this union, or
// missing a required one, is not a session-record observation at all.
var sessionRequired = []string{"session", "operation_id", "manager", "runtime", "native_id", "readable", "state", "labels", "generation", "started_at", "last_activity_at", "doorbell"}
var sessionOptional = []string{"draining", "process", "termination"}

func closedSession(session map[string]json.RawMessage) bool {
	allowed := make(map[string]bool, len(sessionRequired)+len(sessionOptional))
	for _, member := range append(append([]string{}, sessionRequired...), sessionOptional...) {
		allowed[member] = true
	}
	for member := range session {
		if !allowed[member] {
			return false
		}
	}
	for _, member := range sessionRequired {
		if _, ok := session[member]; !ok {
			return false
		}
	}
	return true
}

func objectFields(raw []byte) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return nil, fmt.Errorf("not an object")
	}
	return fields, nil
}

func exactMembers(fields map[string]json.RawMessage, members ...string) bool {
	if len(fields) != len(members) {
		return false
	}
	for _, member := range members {
		if _, ok := fields[member]; !ok {
			return false
		}
	}
	return true
}

func stringMember(fields map[string]json.RawMessage, member string) (string, error) {
	var value string
	if err := json.Unmarshal(fields[member], &value); err != nil {
		return "", err
	}
	return value, nil
}
