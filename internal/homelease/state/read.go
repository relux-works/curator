package state

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/protocoljson"
)

func Namespace(home string) string  { return filepath.Join(home, "state", "home-leases", "v1") }
func LedgerPath(home string) string { return filepath.Join(Namespace(home), "ledger.json") }
func BindingPath(home, id string) string {
	return filepath.Join(Namespace(home), "bindings", id+".json")
}
func EventPath(home, id string) string { return filepath.Join(Namespace(home), "events", id+".json") }

var idPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func ValidID(id string) bool    { return idPattern.MatchString(id) }
func ValidDigest(s string) bool { return digestPattern.MatchString(s) }

// CheckRoute never follows a state symlink, including dangling ancestors. It
// distinguishes a missing namespace from uncertainty without creating state.
func CheckRoute(home, path string) error {
	rel, err := filepath.Rel(home, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("state route escapes manager")
	}
	if err := pathboundary.ValidateRoot(home); err != nil {
		return err
	}
	current := home
	for _, component := range strings.Split(rel, string(filepath.Separator)) {
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("linked lease state")
		}
		if current == path && info.Mode().IsRegular() {
			return pathboundary.CheckPrivateFile(current, info)
		}
		if !info.IsDir() {
			return fmt.Errorf("invalid state route")
		}
		if err := pathboundary.ValidateRoot(current); err != nil {
			return err
		}
	}
	return nil
}

func Read(home, path string) ([]byte, error) {
	if err := CheckRoute(home, path); err != nil {
		return nil, err
	}
	f, err := pathboundary.OpenReadNoFollow(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	i, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !i.Mode().IsRegular() || i.Size() > MaxRecord {
		return nil, fmt.Errorf("invalid bounded lease state")
	}
	raw, err := io.ReadAll(io.LimitReader(f, MaxRecord+1))
	if len(raw) > MaxRecord {
		return nil, fmt.Errorf("lease state exceeds bound")
	}
	return raw, err
}

func HasState(home string) bool {
	for p := Namespace(home); p != home; p = filepath.Dir(p) {
		info, err := os.Lstat(p)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return true
		}
		if err == nil && (p == Namespace(home) || !info.IsDir()) {
			return true
		}
	}
	return false
}

func Load(home string) (*Ledger, error) {
	raw, err := Read(home, LedgerPath(home))
	if errors.Is(err, os.ErrNotExist) && !HasState(home) {
		return Empty(), nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil, Refuse("environment_lease_unreadable", "ledger_missing")
	}
	if err != nil {
		return nil, Refuse("environment_lease_unreadable", "state_unreadable")
	}
	l, err := DecodeLedger(home, raw, func(path string) ([]byte, error) { return Read(home, path) })
	if err != nil {
		return nil, err
	}
	if err := validateInventory(home, l); err != nil {
		return nil, err
	}
	return l, nil
}

// An orphan binding is a partially published reservation, never evidence of
// absence. Live readers inventory the bounded namespace before admitting any
// write; journal validation supplies its own complete projected artifact view.
func validateInventory(home string, l *Ledger) error {
	expected := map[string]map[string]bool{"bindings": {}, "events": {}}
	for id, entry := range l.Leases {
		expected["bindings"][id+".json"] = true
		if entry.TerminalEvent != nil {
			expected["events"][*entry.TerminalEvent+".json"] = true
		}
	}
	for dir, names := range expected {
		path := filepath.Join(Namespace(home), dir)
		if err := CheckRoute(home, path); err != nil {
			return Refuse("environment_lease_unreadable", "state_unreadable")
		}
		entries, err := boundedDirectory(path, 65537)
		if errors.Is(err, os.ErrNotExist) && len(names) == 0 {
			continue
		}
		if err != nil || len(entries) != len(names) {
			return invalid()
		}
		for _, entry := range entries {
			if !names[entry.Name()] || !entry.Type().IsRegular() {
				return invalid()
			}
		}
	}
	entries, err := boundedDirectory(Namespace(home), 4)
	if err != nil || len(entries) > 3 {
		return invalid()
	}
	for _, entry := range entries {
		if entry.Name() != "ledger.json" && entry.Name() != "bindings" && entry.Name() != "events" {
			return invalid()
		}
	}
	return nil
}

func boundedDirectory(path string, bound int) ([]os.DirEntry, error) {
	f, err := openDirectory(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	entries, err := f.ReadDir(bound)
	if errors.Is(err, io.EOF) {
		err = nil
	}
	return entries, err
}

// DecodeLedger validates the complete closed projection and its immutable
// evidence. Journal admission can supply the desired artifact view without
// treating a partially published live ledger as empty.
func DecodeLedger(home string, raw []byte, readArtifact func(string) ([]byte, error)) (*Ledger, error) {
	var l Ledger
	if err := protocoljson.UnmarshalCanonical(raw, &l); err != nil {
		return nil, invalid()
	}
	roundtrip, err := Canonical(l)
	if err != nil || !bytes.Equal(raw, roundtrip) || l.Kind != "home-lease-state/1" || l.Leases == nil || l.Requests == nil {
		return nil, invalid()
	}
	if len(l.Leases) != len(l.Requests) || len(l.Leases) > 65536 {
		return nil, invalid()
	}
	for id, e := range l.Leases {
		if !ValidID(id) || e.Record.LeaseID != id || e.Initial.LeaseID != id || !ValidDigest(e.PayloadSHA256) {
			return nil, invalid()
		}
		for _, r := range []Record{e.Record, e.Initial} {
			b, err := Canonical(r)
			if err != nil || Validate("record", b) != nil {
				return nil, invalid()
			}
		}
		if e.Initial.Phase != "prepared" || e.Initial.Revision != 1 || e.Record.Revision < 1 || l.Requests[e.Initial.Holder.RequestID] != id {
			return nil, invalid()
		}
		if e.Initial.Holder.Kind != "prepared" || e.Initial.Observation.State != "not_observed" || e.Initial.Observation.ObservedAt != nil {
			return nil, invalid()
		}
		payload, err := Canonical(struct {
			Environment string `json:"environment"`
			Profile     string `json:"profile"`
			LockSHA256  string `json:"lock_sha256"`
			RequestID   string `json:"request_id"`
			Holder      Holder `json:"holder"`
		}{e.Initial.Environment, e.Initial.Profile, e.Initial.LockSHA256, e.Initial.Holder.RequestID, e.Initial.Holder})
		if err != nil || e.PayloadSHA256 != digest(payload) || e.Initial.ManagerRootID != "root-"+digest([]byte(home)) || e.Initial.HomeID != "home-"+digest([]byte(e.Admission.Home)) {
			return nil, invalid()
		}
		reservation := e.Record.Holder
		reservation.Kind, reservation.Session, reservation.Generation, reservation.Process = "prepared", nil, nil, nil
		if !reflect.DeepEqual(reservation, e.Initial.Holder) {
			return nil, invalid()
		}
		if e.Record.Holder.Kind == "bound" {
			if e.BindRevision == nil || *e.BindRevision != 1 || e.Record.Revision < 2 || e.Record.Phase == "prepared" {
				return nil, invalid()
			}
		} else if e.BindRevision != nil || e.Record.Phase == "active" {
			return nil, invalid()
		}
		currentBase, initialBase := e.Record, e.Initial
		currentBase.Revision, currentBase.Phase, currentBase.Holder, currentBase.Observation = initialBase.Revision, initialBase.Phase, initialBase.Holder, initialBase.Observation
		if !reflect.DeepEqual(currentBase, initialBase) {
			return nil, invalid()
		}
		binding, err := readArtifact(BindingPath(home, id))
		want, _ := Canonical(e.Initial)
		if err != nil || !bytes.Equal(want, binding) {
			return nil, invalid()
		}
		if !filepath.IsAbs(e.Admission.Home) || filepath.Clean(e.Admission.Home) != e.Admission.Home || len(e.Admission.Protected) == 0 || len(e.Admission.Protected) > 16384 || e.Admission.BuildKeys == nil || e.Admission.RuntimeEntries == nil || len(e.Admission.BuildKeys) > 16384 || len(e.Admission.RuntimeEntries) > 16384 {
			return nil, invalid()
		}
		protectsHome := false
		for _, path := range e.Admission.Protected {
			if !filepath.IsAbs(path) || filepath.Clean(path) != path {
				return nil, invalid()
			}
			protectsHome = protectsHome || path == e.Admission.Home
		}
		if !protectsHome {
			return nil, invalid()
		}
		for _, key := range e.Admission.BuildKeys {
			if !strings.HasPrefix(key, "sha256:") || !ValidDigest(strings.TrimPrefix(key, "sha256:")) {
				return nil, invalid()
			}
		}
		for _, entry := range e.Admission.RuntimeEntries {
			parts := strings.Split(entry, "/")
			if len(parts) != 2 || !identifiers.Valid(parts[0]) || !identifiers.Valid(parts[1]) {
				return nil, invalid()
			}
		}
		if e.Record.Blocking() {
			if e.TerminalEvent != nil {
				return nil, invalid()
			}
		} else {
			if e.TerminalEvent == nil || !ValidID(*e.TerminalEvent) {
				return nil, invalid()
			}
			raw, err := readArtifact(EventPath(home, *e.TerminalEvent))
			if err != nil || Validate("event", raw) != nil {
				return nil, invalid()
			}
			var event Event
			if protocoljson.UnmarshalCanonical(raw, &event) != nil || event.LeaseID != id || event.EventID != *e.TerminalEvent || event.NewRevision != e.Record.Revision || event.Previous.Revision+1 != event.NewRevision || !event.Previous.Blocking() {
				return nil, invalid()
			}
			if (e.Record.Phase == "broken") != (event.Operation == "break") {
				return nil, invalid()
			}
			if !reflect.DeepEqual(event.Previous.Holder, e.Record.Holder) {
				return nil, invalid()
			}
			previousBase := event.Previous
			previousBase.Revision, previousBase.Phase, previousBase.Holder, previousBase.Observation = initialBase.Revision, initialBase.Phase, initialBase.Holder, initialBase.Observation
			if !reflect.DeepEqual(previousBase, initialBase) || e.Record.Observation.ObservedAt == nil || *e.Record.Observation.ObservedAt != event.RecordedAt {
				return nil, invalid()
			}
		}
	}
	return &l, nil
}

func digest(raw []byte) string { sum := sha256.Sum256(raw); return hex.EncodeToString(sum[:]) }

func invalid() *Refusal { return Refuse("environment_lease_unreadable", "invalid_record") }

func Records(l *Ledger) []Record {
	rows := make([]Record, 0, len(l.Leases))
	for _, entry := range l.Leases {
		rows = append(rows, entry.Record)
	}
	sort.Slice(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if a.Environment != b.Environment {
			return a.Environment < b.Environment
		}
		if a.Profile != b.Profile {
			return a.Profile < b.Profile
		}
		if a.HomeID != b.HomeID {
			return a.HomeID < b.HomeID
		}
		return a.LeaseID < b.LeaseID
	})
	return rows
}
