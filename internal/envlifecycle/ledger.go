// Package envlifecycle holds private custody evidence. A retained entry never
// supplies native-origin discovery, launch authority, or a generic drift waiver.
package envlifecycle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/staging"
)

// Version identifies the closed private lifecycle carrier.
const Version = "environment-lifecycle/1"

// MaxBytes bounds the private lifecycle decoder.
const MaxBytes = 8 << 20

var namePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)
var hashPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

// Commit binds a lifecycle outcome to its durable transaction identity.
type Commit struct {
	Unit     string `json:"unit"`
	Evidence string `json:"evidence"`
}

// OwnedRecord retains an exact installed-record preimage.
type OwnedRecord struct {
	Path  string              `json:"path"`
	Image envunit.RecordImage `json:"image"`
}

// Revocation binds removal to the references it revoked.
type Revocation struct {
	Path   string              `json:"path"`
	Before envunit.RecordImage `json:"before"`
	After  envunit.RecordImage `json:"after"`
}

// ConflictReceipt binds the one handled global conflict to its phase and entry.
type ConflictReceipt struct {
	Code          string        `json:"code"`
	Phase         string        `json:"phase"`
	Profile       string        `json:"profile"`
	Adapter       string        `json:"adapter"`
	Target        string        `json:"target"`
	Surface       string        `json:"surface"`
	Entry         envunit.Image `json:"entry"`
	Publication   string        `json:"publication"`
	UnitPreimages string        `json:"unit_preimages"`
	Sequence      uint64        `json:"sequence"`
}

// GlobalRetained retains L1 and the complete old-home recovery baseline.
type GlobalRetained struct {
	ID                 string                      `json:"id"`
	Kind               string                      `json:"kind"`
	Phase              string                      `json:"phase"`
	CreatedByUnit      envunit.Unit                `json:"created_by_unit"`
	ProfileTransitions []envunit.ProfileTransition `json:"profile_transitions"`
	Homes              []envunit.Home              `json:"homes"`
	BoundaryProof      staging.DurableSnapshot     `json:"boundary_proof"`
	EvidenceDigest     string                      `json:"evidence_digest"`
	ClosureRefs        []string                    `json:"closure_refs"`
	RetainedRecords    []OwnedRecord               `json:"retained_records"`
	MutablePreimages   []envunit.Target            `json:"mutable_preimages"`
	ConflictReceipt    ConflictReceipt             `json:"conflict_receipt"`
	ContinuationCut    string                      `json:"continuation_cut"`
}

// RemovedRetained records intentional nonlaunchable home retention after removal.
type RemovedRetained struct {
	ID                   string                      `json:"id"`
	Kind                 string                      `json:"kind"`
	Phase                string                      `json:"phase"`
	CreatedByUnit        envunit.Unit                `json:"created_by_unit"`
	ProfileTransitions   []envunit.ProfileTransition `json:"profile_transitions"`
	Homes                []envunit.Home              `json:"homes"`
	BoundaryProof        staging.DurableSnapshot     `json:"boundary_proof"`
	EvidenceDigest       string                      `json:"evidence_digest"`
	RemovedRecords       []OwnedRecord               `json:"removed_records"`
	ExpectedAbsence      []string                    `json:"expected_absence"`
	ReferenceRevocations []Revocation                `json:"reference_revocations"`
	RemovalCommit        Commit                      `json:"removal_commit"`
}

// Purged records terminal verified absence and its name-reuse binding.
type Purged struct {
	ID               string          `json:"id"`
	Kind             string          `json:"kind"`
	Phase            string          `json:"phase"`
	Manager          envunit.Manager `json:"manager"`
	PriorRemovalID   string          `json:"prior_removal_id"`
	ProfileName      string          `json:"profile_name"`
	TargetNamespaces []string        `json:"target_namespaces"`
	RemovalCommit    Commit          `json:"removal_commit"`
	PurgeCommit      Commit          `json:"purge_commit"`
	ExpectedAbsence  []string        `json:"expected_absence"`
}

// Entry is an in-memory tagged union. Each wire variant is closed separately;
// optional union members never expand the accepted persisted grammar.
type Entry struct {
	Global  *GlobalRetained
	Removed *RemovedRetained
	Purged  *Purged
}

// Ledger is a bounded collection of private action-specific lifecycle entries.
type Ledger struct {
	Record   string
	Manager  envunit.Manager
	Revision uint64
	Entries  []Entry
}
type wireLedger struct {
	Record   string            `json:"record"`
	Manager  envunit.Manager   `json:"manager"`
	Revision uint64            `json:"revision"`
	Entries  []json.RawMessage `json:"entries"`
}

func invalid(format string, args ...any) error {
	return fmt.Errorf(envorigin.DiagInvalid+": lifecycle: "+format, args...)
}

// Path returns the fixed private lifecycle carrier path.
func Path(manager string) string {
	return filepath.Join(manager, ".state", "environment-lifecycle.json")
}
func same(a, b any) bool {
	ar, _ := json.Marshal(a)
	br, _ := json.Marshal(b)
	return bytes.Equal(ar, br)
}
func validCommit(c Commit) bool {
	return namePattern.MatchString(c.Unit) && hashPattern.MatchString(c.Evidence)
}

func parseEntry(raw []byte) (Entry, error) {
	if len(raw) > envunit.MaxBytes {
		return Entry{}, invalid("entry bounds")
	}
	var tag struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(raw, &tag); err != nil {
		return Entry{}, invalid("variant: %v", err)
	}
	var e Entry
	switch tag.Kind {
	case "global-retained":
		e.Global = new(GlobalRetained)
		if err := protocoljson.DecodeClosed(raw, e.Global); err != nil {
			return Entry{}, invalid("global variant: %v", err)
		}
	case "removed-retained":
		e.Removed = new(RemovedRetained)
		if err := protocoljson.DecodeClosed(raw, e.Removed); err != nil {
			return Entry{}, invalid("removed variant: %v", err)
		}
	case "purged":
		e.Purged = new(Purged)
		if err := protocoljson.DecodeClosed(raw, e.Purged); err != nil {
			return Entry{}, invalid("purged variant: %v", err)
		}
	default:
		return Entry{}, invalid("unknown variant")
	}
	return e, nil
}

// ParseEntry validates a complete captured noncredential receipt independently
// of installed source/native lookup. It supplies custody evidence only.
func ParseEntry(raw []byte, manager envunit.Manager) (Entry, error) {
	e, err := parseEntry(raw)
	if err != nil {
		return Entry{}, err
	}
	if err := e.validate(manager); err != nil {
		return Entry{}, err
	}
	return e, nil
}

// Marshal encodes a validated private lifecycle entry.
func (e Entry) Marshal() ([]byte, error) {
	n := 0
	var v any
	if e.Global != nil {
		n++
		v = e.Global
	}
	if e.Removed != nil {
		n++
		v = e.Removed
	}
	if e.Purged != nil {
		n++
		v = e.Purged
	}
	if n != 1 {
		return nil, invalid("entry must have exactly one variant")
	}
	return json.Marshal(v)
}

// ID returns the action-specific durable entry identity.
func (e Entry) ID() string {
	if e.Global != nil {
		return e.Global.ID
	}
	if e.Removed != nil {
		return e.Removed.ID
	}
	if e.Purged != nil {
		return e.Purged.ID
	}
	return ""
}

// Profiles returns the complete profile names bound by this entry.
func (e Entry) Profiles() []string {
	if e.Purged != nil {
		return []string{e.Purged.ProfileName}
	}
	var p []envunit.ProfileTransition
	if e.Global != nil {
		p = e.Global.ProfileTransitions
	}
	if e.Removed != nil {
		p = e.Removed.ProfileTransitions
	}
	names := make([]string, 0, len(p))
	for _, x := range p {
		names = append(names, x.Name)
	}
	return names
}

// Targets returns the exact managed homes bound by this entry.
func (e Entry) Targets() []string {
	if e.Purged != nil {
		return append([]string(nil), e.Purged.TargetNamespaces...)
	}
	var homes []envunit.Home
	if e.Global != nil {
		homes = e.Global.Homes
	}
	if e.Removed != nil {
		homes = e.Removed.Homes
	}
	out := make([]string, 0, len(homes))
	for _, h := range homes {
		out = append(out, h.Root)
	}
	return out
}

// Parse decodes and validates the closed private lifecycle ledger.
func Parse(raw []byte, manager envunit.Manager) (*Ledger, error) {
	if len(raw) == 0 || len(raw) > MaxBytes {
		return nil, invalid("ledger bounds")
	}
	var wire wireLedger
	if err := protocoljson.DecodeClosed(raw, &wire); err != nil {
		return nil, invalid("closed ledger: %v", err)
	}
	if wire.Record != Version || wire.Manager != manager || wire.Revision == 0 || len(wire.Entries) > 1024 {
		return nil, invalid("header/manager/revision")
	}
	l := &Ledger{Record: wire.Record, Manager: wire.Manager, Revision: wire.Revision, Entries: make([]Entry, 0, len(wire.Entries))}
	for _, rawEntry := range wire.Entries {
		entry, err := parseEntry(rawEntry)
		if err != nil {
			return nil, err
		}
		l.Entries = append(l.Entries, entry)
	}
	if err := l.Validate(); err != nil {
		return nil, err
	}
	return l, nil
}

// Marshal validates and encodes a canonical private ledger.
func (l *Ledger) Marshal() ([]byte, error) {
	if err := l.Validate(); err != nil {
		return nil, err
	}
	w := wireLedger{Record: l.Record, Manager: l.Manager, Revision: l.Revision, Entries: make([]json.RawMessage, 0, len(l.Entries))}
	for _, entry := range l.Entries {
		raw, err := entry.Marshal()
		if err != nil {
			return nil, err
		}
		w.Entries = append(w.Entries, raw)
	}
	raw, err := json.Marshal(w)
	if len(raw) > MaxBytes {
		return nil, invalid("ledger bounds")
	}
	return raw, err
}

// Validate checks closed entry, revision and identity invariants.
func (l *Ledger) Validate() error {
	if l == nil || l.Record != Version || l.Revision == 0 || len(l.Entries) > 1024 {
		return invalid("header/revision/bounds")
	}
	if err := l.Manager.Validate(); err != nil {
		return err
	}
	ids := map[string]bool{}
	profiles := map[string]bool{}
	paths := []string{}
	for _, entry := range l.Entries {
		if _, err := entry.Marshal(); err != nil {
			return err
		}
		if !namePattern.MatchString(entry.ID()) || ids[entry.ID()] {
			return invalid("lifecycle ID/duplicate")
		}
		ids[entry.ID()] = true
		if err := entry.validate(l.Manager); err != nil {
			return err
		}
		for _, profile := range entry.Profiles() {
			if profiles[profile] {
				return invalid("conflicting profile lifecycle ownership")
			}
			profiles[profile] = true
		}
		for _, target := range entry.Targets() {
			for _, prior := range paths {
				if pathOverlap(prior, target) {
					return invalid("conflicting target lifecycle ownership")
				}
			}
			paths = append(paths, target)
		}
	}
	return nil
}
func pathOverlap(a, b string) bool { return a == b || below(a, b) || below(b, a) }
func below(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !filepath.IsAbs(rel) && len(rel) > 0 && (len(rel) < 3 || rel[:3] != ".."+string(filepath.Separator))
}

func commonValid(manager envunit.Manager, unit envunit.Unit, profiles []envunit.ProfileTransition, homes []envunit.Home, proof staging.DurableSnapshot) error {
	if unit.Manager != manager {
		return invalid("retained manager differs from creating unit")
	}
	if err := unit.Validate(); err != nil {
		return err
	}
	if !same(unit.Profiles, profiles) || !same(unit.Homes, homes) || !same(unit.BoundaryProof, proof) {
		return invalid("retained inventory differs from admitted creating unit")
	}
	return nil
}
func (e Entry) validate(manager envunit.Manager) error {
	if g := e.Global; g != nil {
		if g.Kind != "global-retained" || g.Phase != "retained_conflict" || g.CreatedByUnit.Action != envunit.GlobalExtend || g.ContinuationCut != "profile_in_place_conflict" {
			return invalid("global retained phase/action")
		}
		if err := commonValid(manager, g.CreatedByUnit, g.ProfileTransitions, g.Homes, g.BoundaryProof); err != nil {
			return err
		}
		if !validEvidence(g, g.EvidenceDigest) {
			return invalid("global retained evidence digest")
		}
		c := g.ConflictReceipt
		if c.Code != "environment_surface_unmanaged_conflict" || c.Phase != "profile_in_place" || c.Sequence == 0 || !hashPattern.MatchString(c.Publication) || !hashPattern.MatchString(c.UnitPreimages) || c.Surface == "" || c.Target == "" || c.Entry.Kind == "absent" {
			return invalid("typed handled conflict receipt")
		}
		found := false
		for _, p := range g.ProfileTransitions {
			if p.Name == c.Profile {
				found = true
			}
		}
		if !found {
			return invalid("conflict profile outside creating group")
		}
		if _, err := envorigin.NativeTarget(filepath.Dir(c.Target), c.Adapter); err != nil {
			return invalid("conflict adapter")
		}
		if !same(g.MutablePreimages, g.CreatedByUnit.Targets) {
			return invalid("global retained preimages differ from complete target set")
		}
		if len(g.ClosureRefs) == 0 || len(g.RetainedRecords) == 0 {
			return invalid("global retained closure/publication inventory")
		}
		seen := map[string]bool{}
		for _, ref := range g.ClosureRefs {
			if !hashPattern.MatchString(ref) || seen[ref] {
				return invalid("global retained closure reference")
			}
			seen[ref] = true
		}
		for _, p := range g.ProfileTransitions {
			for _, ref := range p.Closure {
				if !seen[ref] {
					return invalid("retained closure omits a creating dependency")
				}
			}
		}
		return validateRecords(manager, g.ProfileTransitions, g.RetainedRecords, false)
	}
	if r := e.Removed; r != nil {
		if r.Kind != "removed-retained" || r.Phase != "retained" || r.CreatedByUnit.Action != envunit.Remove || !validCommit(r.RemovalCommit) || r.RemovalCommit.Unit != r.CreatedByUnit.ID {
			return invalid("removed retained commit/phase/action")
		}
		if err := commonValid(manager, r.CreatedByUnit, r.ProfileTransitions, r.Homes, r.BoundaryProof); err != nil {
			return err
		}
		if !validEvidence(r, r.EvidenceDigest) {
			return invalid("removed retained evidence digest")
		}
		if len(r.RemovedRecords) == 0 || len(r.ExpectedAbsence) == 0 || r.ReferenceRevocations == nil {
			return invalid("complete removal inventory required")
		}
		if err := validateRecords(manager, r.ProfileTransitions, r.RemovedRecords, true); err != nil {
			return err
		}
		for _, p := range r.ProfileTransitions {
			if p.NewLock.Kind != "absent" || !contains(r.ExpectedAbsence, filepath.Join(manager.Path, "profiles", p.Name)) {
				return invalid("removed profile namespace lacks absent postcondition")
			}
		}
		for _, h := range r.Homes {
			if h.Before.Kind != "recorded" || h.After.Kind != "recorded" || !bytes.Equal(h.Before.Marker, h.After.Marker) || !bytes.Equal(h.Before.Origin, h.After.Origin) || h.DirectoryBefore != h.DirectoryAfter {
				return invalid("remove retained changed home")
			}
		}
		for _, path := range r.ExpectedAbsence {
			valid := false
			for _, p := range r.ProfileTransitions {
				valid = valid || path == filepath.Join(manager.Path, "profiles", p.Name)
			}
			if !valid {
				return invalid("unexpected removed absence namespace")
			}
		}
		for _, revocation := range r.ReferenceRevocations {
			if !allowedReference(manager.Path, revocation.Path) || same(revocation.Before, revocation.After) {
				return invalid("invalid removal reference revocation")
			}
			found := false
			for _, target := range r.CreatedByUnit.Targets {
				if target.Path == revocation.Path && target.Class == "reference-record" && target.Before.Kind == revocation.Before.Kind && target.After.Kind == revocation.After.Kind && target.Before.Digest == revocation.Before.Digest && target.After.Digest == revocation.After.Digest {
					found = true
				}
			}
			if !found {
				return invalid("revocation not bound to actual removal target")
			}
		}
		return nil
	}
	p := e.Purged
	if p == nil || p.Kind != "purged" || p.Phase != "complete" || p.Manager != manager || !namePattern.MatchString(p.PriorRemovalID) || !namePattern.MatchString(p.ProfileName) || !validCommit(p.RemovalCommit) || !validCommit(p.PurgeCommit) || p.TargetNamespaces == nil || len(p.TargetNamespaces) > 16 {
		return invalid("purged identity/commit/namespace")
	}
	want := []string{filepath.Join(manager.Path, "profiles", p.ProfileName)}
	seen := map[string]bool{}
	for _, target := range p.TargetNamespaces {
		valid := false
		for _, adapter := range []string{"codex_cli", "claude_code", "pi", "opencode", "muse"} {
			valid = valid || target == envorigin.ManagedRoot(manager.Path, p.ProfileName, adapter)
		}
		if !valid || seen[target] {
			return invalid("purged namespace differs from fixed owned home")
		}
		seen[target] = true
		want = append(want, target)
	}
	sort.Strings(want)
	actual := append([]string{}, p.ExpectedAbsence...)
	sort.Strings(actual)
	if !same(want, actual) {
		return invalid("purged absence set differs from complete target set")
	}
	return nil
}
func contains(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}
func allowedReference(manager, path string) bool {
	return path == filepath.Join(manager, "profiles", "current") || below(filepath.Join(manager, "profiles", "scoped"), path) || path == filepath.Join(manager, "state", "project-leases", "v1", "ledger.json")
}
func validateRecords(manager envunit.Manager, profiles []envunit.ProfileTransition, records []OwnedRecord, removed bool) error {
	seen := map[string]bool{}
	for _, record := range records {
		if seen[record.Path] || record.Image.Kind != "file" || len(record.Image.Bytes) == 0 || envorigin.Digest(record.Image.Bytes) != record.Image.Digest {
			return invalid("record byte identity/duplicate")
		}
		seen[record.Path] = true
		valid := false
		for _, p := range profiles {
			root := filepath.Join(manager.Path, "profiles", p.Name)
			for _, name := range []string{"source.json", "lock.json", "lock.prev.json"} {
				valid = valid || record.Path == filepath.Join(root, name)
			}
			if record.Path == filepath.Join(root, "lock.json") {
				want := p.NewLock
				if removed {
					want = p.OldLock
				}
				if !same(want, record.Image) {
					return invalid("retained/removed lock differs from creating transition")
				}
			}
			if record.Path == filepath.Join(root, "source.json") && record.Image.Digest != p.Source {
				return invalid("retained/removed source differs from creating transition")
			}
		}
		if !valid {
			return invalid("record outside installed inventory")
		}
	}
	for _, p := range profiles {
		for _, name := range []string{"source.json", "lock.json"} {
			if !seen[filepath.Join(manager.Path, "profiles", p.Name, name)] {
				return invalid("installed source/lock missing from lifecycle inventory")
			}
		}
	}
	return nil
}
func validEvidence(v any, want string) bool {
	if !hashPattern.MatchString(want) {
		return false
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return false
	}
	var members map[string]json.RawMessage
	_ = json.Unmarshal(raw, &members)
	delete(members, "evidence_digest")
	canonical, _ := json.Marshal(members)
	return envorigin.Digest(canonical) == want
}

// Seal hashes only the closed noncredential custody representation. It is an
// integrity/CAS aid within the same-UID trust boundary, not a signature.
func (e *Entry) Seal() error {
	raw, err := e.Marshal()
	if err != nil {
		return err
	}
	var members map[string]json.RawMessage
	if err := json.Unmarshal(raw, &members); err != nil {
		return err
	}
	delete(members, "evidence_digest")
	raw, err = json.Marshal(members)
	if err != nil {
		return err
	}
	digest := envorigin.Digest(raw)
	if e.Global != nil {
		e.Global.EvidenceDigest = digest
	} else if e.Removed != nil {
		e.Removed.EvidenceDigest = digest
	} else {
		return invalid("purged receipt has action commit evidence instead")
	}
	return nil
}

// Read returns an absent ledger only after a protected no-follow route proves
// absence. It does not classify homes or authorize a caller's continuation.
func Read(manager envunit.Manager) (*Ledger, []byte, error) {
	raw, present, err := envorigin.ReadPrivateAuthority(manager.Path, Path(manager.Path), MaxBytes)
	if err != nil {
		return nil, nil, err
	}
	if !present {
		return &Ledger{Record: Version, Manager: manager, Entries: []Entry{}}, nil, nil
	}
	l, err := Parse(raw, manager)
	return l, raw, err
}

// PlanReplace preserves every protected entry byte-for-byte unless one exact
// typed lifecycle action is entitled to consume it. The returned replacement
// remains a journal target with a full-file preimage; no write occurs here.
func PlanReplace(raw []byte, manager envunit.Manager, revision uint64, next []Entry, action *envunit.Unit) ([]byte, error) {
	var old *Ledger
	var err error
	if len(raw) == 0 {
		if revision != 0 {
			return nil, invalid("absent ledger revision")
		}
		old = &Ledger{Record: Version, Manager: manager, Entries: []Entry{}}
	} else {
		old, err = Parse(raw, manager)
		if err != nil {
			return nil, err
		}
		if old.Revision != revision {
			return nil, invalid("ledger revision conflict")
		}
	}
	if revision == ^uint64(0) {
		return nil, invalid("ledger revision overflow")
	}
	if action != nil {
		if err := action.Validate(); err != nil {
			return nil, err
		}
		publication := action.LedgerPublication
		digest := "absent"
		if len(raw) != 0 {
			digest = envorigin.Digest(raw)
		}
		if publication == nil || publication.Path != Path(manager.Path) || publication.PreimageDigest != digest || publication.Revision != revision || publication.NextRevision != revision+1 {
			return nil, invalid("action differs from structured ledger CAS preimage")
		}
		ids := make([]string, 0, len(next))
		for _, entry := range next {
			ids = append(ids, entry.ID())
		}
		sort.Strings(ids)
		bound := append([]string{}, publication.EntryIDs...)
		sort.Strings(bound)
		if !same(ids, bound) {
			return nil, invalid("action differs from complete successor entry set")
		}
	}
	proposed := &Ledger{Record: Version, Manager: manager, Revision: revision + 1, Entries: next}
	if err := proposed.Validate(); err != nil {
		return nil, err
	}
	byID := map[string]Entry{}
	for _, entry := range next {
		byID[entry.ID()] = entry
	}
	for _, entry := range old.Entries {
		if successor, ok := byID[entry.ID()]; ok {
			before, _ := entry.Marshal()
			after, _ := successor.Marshal()
			if bytes.Equal(before, after) {
				continue
			}
		}
		if action == nil || action.Manager != manager {
			return nil, invalid("structured ledger replacement changes protected entry")
		}
		if err := action.Validate(); err != nil {
			return nil, err
		}
		if err := consumeEntry(entry, action, proposed); err != nil {
			return nil, err
		}
	}
	return proposed.Marshal()
}
func consumeEntry(entry Entry, action *envunit.Unit, next *Ledger) error {
	// Consumption is restricted to the exact lifecycle transitions. Production
	// callers must additionally verify live receipts and the engine target set.
	raw, err := entry.Marshal()
	if err != nil {
		return err
	}
	if action.RetainedEntry == nil || action.RetainedEntry.ID != entry.ID() || action.RetainedEntry.Digest != envorigin.Digest(raw) {
		return invalid("consumption is not bound to exact retained receipt")
	}
	if p := entry.Purged; p != nil && action.Action == envunit.Reuse {
		if len(action.Profiles) != 1 || action.Profiles[0].Name != p.ProfileName {
			return invalid("name reuse profile differs from purged receipt")
		}
		for _, home := range action.Homes {
			if home.Before.Kind != "absent" || home.After.Kind != "recorded" {
				return invalid("name reuse is not genuine new capture")
			}
		}
		return nil
	}
	if r := entry.Removed; r != nil && action.Action == envunit.Purge {
		if !same(r.ProfileTransitions, action.Profiles) || len(r.Homes) != len(action.Homes) {
			return invalid("purge must bind exact removal inventory")
		}
		roots := make([]string, 0, len(r.Homes))
		for i, prior := range r.Homes {
			home := action.Homes[i]
			if home.Profile != prior.Profile || home.Adapter != prior.Adapter || home.Mode != prior.Mode || home.Target != prior.Target || home.Root != prior.Root || !same(home.Before, prior.After) || home.DirectoryBefore != prior.DirectoryAfter || home.After.Kind != "absent" || home.DirectoryAfter.Kind != "absent" {
				return invalid("purge changed retained home preimage or target inventory")
			}
			roots = append(roots, home.Root)
		}
		sort.Strings(roots)
		for _, candidate := range next.Entries {
			if p := candidate.Purged; p != nil && p.PriorRemovalID == r.ID && p.PurgeCommit.Unit == action.ID {
				targets := append([]string{}, p.TargetNamespaces...)
				sort.Strings(targets)
				if !same(roots, targets) || p.RemovalCommit != r.RemovalCommit || len(action.Profiles) != 1 || p.ProfileName != action.Profiles[0].Name {
					return invalid("purge terminal successor differs from removal")
				}
				return nil
			}
		}
		return invalid("purge consumption lacks terminal successor")
	}
	if g := entry.Global; g != nil && action.Action == envunit.Takeover {
		if len(g.ProfileTransitions) != len(action.Profiles) || len(g.Homes) != len(action.Homes) {
			return invalid("takeover differs from complete retained group")
		}
		for i, prior := range g.ProfileTransitions {
			p := action.Profiles[i]
			if p.Name != prior.Name || p.Source != prior.Source || !same(p.Closure, prior.Closure) || !same(p.OldLock, prior.NewLock) || !same(p.NewLock, prior.NewLock) || !same(action.RetainedEntry.HomeLocks[p.Name], prior.OldLock) {
				return invalid("takeover must retain L1 and cannot perform L2")
			}
		}
		for i, prior := range g.Homes {
			home := action.Homes[i]
			if home.Profile != prior.Profile || home.Adapter != prior.Adapter || home.Mode != prior.Mode || home.Target != prior.Target || home.Root != prior.Root || !same(home.Before, prior.Before) || home.DirectoryBefore != prior.DirectoryBefore || home.After.Kind != "recorded" {
				return invalid("takeover does not start at exact retained old-home state")
			}
		}
		return nil
	}
	return invalid("action cannot consume protected lifecycle entry")
}

// ValidateConsumption rechecks the exact serialized predecessor even after its
// ledger backup was durably cleaned. A missing ID alone is not consumption.
func ValidateConsumption(action *envunit.Unit, next *Ledger) error {
	if action == nil || action.RetainedEntry == nil {
		return invalid("consumption lacks predecessor")
	}
	if err := action.Validate(); err != nil {
		return err
	}
	raw := action.RetainedEntry.Serialized
	if envorigin.Digest(raw) != action.RetainedEntry.Digest {
		return invalid("consumption predecessor digest differs")
	}
	entry, err := ParseEntry(raw, action.Manager)
	if err != nil {
		return err
	}
	if entry.ID() != action.RetainedEntry.ID {
		return invalid("consumption predecessor identity differs")
	}
	return consumeEntry(entry, action, next)
}
