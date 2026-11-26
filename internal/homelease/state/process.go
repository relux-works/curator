package state

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// StartTime is the closed tagged kernel birth identity of a bound process.
// Exactly one variant is populated; variants never convert into each other
// or into a wall-clock time.
type StartTime struct {
	Kind  string
	Ticks string // linux-stat-starttime: /proc/<pid>/stat field 22
	Hz    int64  // linux-stat-starttime: sysconf(_SC_CLK_TCK)
	Sec   string // darwin-pbi-start: pbi_start_tvsec
	Usec  int64  // darwin-pbi-start: pbi_start_tvusec
}

// Supported kernel process start-time discriminator values.
const (
	StartLinux  = "linux-stat-starttime"
	StartDarwin = "darwin-pbi-start"
)

// Process is one holder process member: either the bound four-field tuple
// {pid, uid, boot_id, start_time} or, for prepared holders only, the legacy
// {pid, started_at} form. The legacy form can never bind.
type Process struct {
	PID       int64
	UID       int64
	BootID    string
	StartTime *StartTime
	StartedAt string // legacy prepared-only form
	legacy    bool
}

// LegacyProcess returns the prepared-only legacy form.
func LegacyProcess(pid int64, startedAt string) *Process {
	return &Process{PID: pid, StartedAt: startedAt, legacy: true}
}

// IsLegacy reports whether this process carries only legacy wall-clock data.
func (p *Process) IsLegacy() bool { return p != nil && p.legacy }

// MarshalJSON encodes the closed platform-specific start_time object.
func (s StartTime) MarshalJSON() ([]byte, error) {
	switch s.Kind {
	case StartLinux:
		return json.Marshal(struct {
			Kind  string `json:"kind"`
			Ticks string `json:"ticks"`
			Hz    int64  `json:"hz"`
		}{s.Kind, s.Ticks, s.Hz})
	case StartDarwin:
		return json.Marshal(struct {
			Kind string `json:"kind"`
			Sec  string `json:"sec"`
			Usec int64  `json:"usec"`
		}{s.Kind, s.Sec, s.Usec})
	}
	return nil, fmt.Errorf("unknown start_time kind")
}

// UnmarshalJSON decodes and checks the closed platform-specific start_time object.
func (s *StartTime) UnmarshalJSON(raw []byte) error {
	fields, err := closedObject(raw)
	if err != nil {
		return err
	}
	kind, err := stringField(fields, "kind")
	if err != nil {
		return err
	}
	out := StartTime{Kind: kind}
	switch kind {
	case StartLinux:
		if !exactKeys(fields, "kind", "ticks", "hz") {
			return fmt.Errorf("invalid linux start_time members")
		}
		if out.Ticks, err = stringField(fields, "ticks"); err != nil {
			return err
		}
		if out.Hz, err = intField(fields, "hz"); err != nil {
			return err
		}
	case StartDarwin:
		if !exactKeys(fields, "kind", "sec", "usec") {
			return fmt.Errorf("invalid darwin start_time members")
		}
		if out.Sec, err = stringField(fields, "sec"); err != nil {
			return err
		}
		if out.Usec, err = intField(fields, "usec"); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown start_time kind")
	}
	*s = out
	return nil
}

// MarshalJSON encodes either the bound four-field tuple or the legacy prepared form.
func (p Process) MarshalJSON() ([]byte, error) {
	if p.legacy {
		return json.Marshal(struct {
			PID       int64  `json:"pid"`
			StartedAt string `json:"started_at"`
		}{p.PID, p.StartedAt})
	}
	return json.Marshal(struct {
		PID       int64      `json:"pid"`
		UID       int64      `json:"uid"`
		BootID    string     `json:"boot_id"`
		StartTime *StartTime `json:"start_time"`
	}{p.PID, p.UID, p.BootID, p.StartTime})
}

// UnmarshalJSON decodes either the bound four-field tuple or the legacy prepared form.
func (p *Process) UnmarshalJSON(raw []byte) error {
	fields, err := closedObject(raw)
	if err != nil {
		return err
	}
	out := Process{}
	if out.PID, err = intField(fields, "pid"); err != nil {
		return err
	}
	if exactKeys(fields, "pid", "started_at") {
		if out.StartedAt, err = stringField(fields, "started_at"); err != nil {
			return err
		}
		out.legacy = true
		*p = out
		return nil
	}
	if !exactKeys(fields, "pid", "uid", "boot_id", "start_time") {
		return fmt.Errorf("invalid process members")
	}
	if out.UID, err = intField(fields, "uid"); err != nil {
		return err
	}
	if out.BootID, err = stringField(fields, "boot_id"); err != nil {
		return err
	}
	var start StartTime
	if err := json.Unmarshal(fields["start_time"], &start); err != nil {
		return err
	}
	out.StartTime = &start
	*p = out
	return nil
}

func closedObject(raw []byte) (map[string]json.RawMessage, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	var fields map[string]json.RawMessage
	if err := dec.Decode(&fields); err != nil || fields == nil {
		return nil, fmt.Errorf("process value is not an object")
	}
	return fields, nil
}

func exactKeys(fields map[string]json.RawMessage, keys ...string) bool {
	if len(fields) != len(keys) {
		return false
	}
	for _, key := range keys {
		if _, ok := fields[key]; !ok {
			return false
		}
	}
	return true
}

func stringField(fields map[string]json.RawMessage, key string) (string, error) {
	var value string
	if err := json.Unmarshal(fields[key], &value); err != nil {
		return "", fmt.Errorf("invalid %s", key)
	}
	return value, nil
}

// intField accepts only a JSON integer token, never a float or exponent form.
func intField(fields map[string]json.RawMessage, key string) (int64, error) {
	raw := bytes.TrimSpace(fields[key])
	value, err := strconv.ParseInt(string(raw), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s", key)
	}
	return value, nil
}

// validDecimal reports the unsigned decimal-string grammar shared by ticks
// and sec: 1-20 ASCII digits, no sign, whitespace or leading zero padding.
func validDecimal(s string) bool {
	if len(s) == 0 || len(s) > 20 {
		return false
	}
	if len(s) > 1 && s[0] == '0' {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// decimalWithinUint64 enforces the semantic ceiling: a grammar-valid digit
// string may still name a value above 2^64-1, which no kernel field can hold.
func decimalWithinUint64(s string) bool {
	if !validDecimal(s) {
		return false
	}
	_, err := strconv.ParseUint(s, 10, 64)
	return err == nil
}

func validBootID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i := 0; i < len(s); i++ {
		switch i {
		case 8, 13, 18, 23:
			if s[i] != '-' {
				return false
			}
		default:
			c := s[i]
			if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
				return false
			}
		}
	}
	return true
}

// ValidateProcessSemantics applies the numeric ceilings that are semantic
// checks rather than schema grammar: pid within pid_t (and positive), uid
// within uid_t with (uid_t)-1 excluded, ticks/sec at most 2^64-1. A nil
// process is not process evidence and is rejected here; callers deciding
// admission treat absence separately.
func ValidateProcessSemantics(p *Process) error {
	if p == nil {
		return fmt.Errorf("missing process")
	}
	if p.legacy {
		if p.PID < 1 || p.StartedAt == "" || p.StartTime != nil {
			return fmt.Errorf("invalid legacy process")
		}
		return nil
	}
	if p.PID < 1 || p.PID > 2147483647 {
		return fmt.Errorf("pid outside range")
	}
	if p.UID < 0 || p.UID > 4294967294 {
		return fmt.Errorf("uid outside range")
	}
	if !validBootID(p.BootID) {
		return fmt.Errorf("invalid boot_id")
	}
	if p.StartTime == nil || p.StartedAt != "" {
		return fmt.Errorf("invalid bound process")
	}
	switch p.StartTime.Kind {
	case StartLinux:
		if !decimalWithinUint64(p.StartTime.Ticks) || p.StartTime.Hz < 1 || p.StartTime.Hz > 100000 || p.StartTime.Sec != "" || p.StartTime.Usec != 0 {
			return fmt.Errorf("invalid linux start_time")
		}
	case StartDarwin:
		if !decimalWithinUint64(p.StartTime.Sec) || p.StartTime.Usec < 0 || p.StartTime.Usec > 999999 || p.StartTime.Ticks != "" || p.StartTime.Hz != 0 {
			return fmt.Errorf("invalid darwin start_time")
		}
	default:
		return fmt.Errorf("unknown start_time kind")
	}
	return nil
}

// ValidateHolderSemantics applies semantic process checks to a holder whose
// closed schema shape already validated. A holder without a process carries
// no semantic process claim; admission rules refuse its bind separately.
func ValidateHolderSemantics(h Holder) error {
	if h.Process == nil {
		return nil
	}
	return ValidateProcessSemantics(h.Process)
}

// ProcessEqual reports byte equality of two process tuples through their
// canonical encodings. A missing process never equals a present one.
func ProcessEqual(a, b *Process) bool {
	if a == nil || b == nil {
		return a == b
	}
	rawA, errA := Canonical(a)
	rawB, errB := Canonical(b)
	return errA == nil && errB == nil && bytes.Equal(rawA, rawB)
}

// StartTimeEqual is the birth-identity half of tuple equality, kept as its
// own comparison so a mutant skipping it is observable.
func StartTimeEqual(a, b *StartTime) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
