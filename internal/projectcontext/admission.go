package projectcontext

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/protocoljson"
)

// PackageInstruction is the explicit operator packaging act. Discovery never
// calls it. It selects exactly one inventoried bare instruction file and
// preserves its bytes; no normalization or implicit import expansion occurs.
func (s Store) PackageInstruction(id, itemID, source string, required bool) (result Preview, resultErr error) {
	defer func() { resultErr = projectError(resultErr) }()
	if !idRE.MatchString(itemID) || !relative(source) {
		return Preview{}, fail("project_context_invalid", "invalid selected instruction")
	}
	op, err := s.begin(false)
	if err != nil {
		return Preview{}, err
	}
	defer op.close()
	r, err := op.registered(id)
	if err != nil {
		return Preview{}, err
	}
	root, err := os.OpenRoot(r.Record.ProjectRoot)
	if err != nil {
		return Preview{}, err
	}
	defer root.Close()
	inventory, err := discovery(root)
	if err != nil {
		return Preview{}, err
	}
	found := false
	for _, d := range inventory {
		if d.Path == source && d.Kind == "instructions" {
			found = true
		}
	}
	if !found || (path.Base(source) != "AGENTS.md" && path.Base(source) != "CLAUDE.md") {
		return Preview{}, fail("project_context_invalid", "select one inventoried AGENTS.md or CLAUDE.md")
	}
	b, _, err := read(root, source, MaxFileBytes)
	if err != nil {
		return Preview{}, err
	}
	if err = contextpkg.ValidateModuleBytes(b); err != nil {
		return Preview{}, err
	}
	name := path.Base(source)
	man, err := canonical(map[string]any{"schema_version": 2, "name": itemID, "version": "0.0.0", "context": map[string]any{"modules": []any{map[string]any{"path": name}}}})
	if err != nil {
		return Preview{}, err
	}
	req := Request{ID: itemID, Type: "context", Environment: r.Record.Environment, Required: required, Source: source, Dependencies: map[string]string{}, Packaging: "instruction"}
	c, err := op.context(r, req, map[string][]byte{"agent-context.json": man, "context/" + name: b})
	if err != nil {
		return Preview{}, err
	}
	c.SourceSHA256 = digest(b)
	items := []Candidate{}
	if old, ok := op.state.Reviews[id]; ok {
		for _, item := range old.Review.Items {
			if item.Item.ID != itemID {
				items = append(items, item)
			}
		}
	}
	items = append(items, c)
	return op.preview(r, items, []Omission{})
}
func (s Store) Preview(id string, d Declaration) (result Preview, resultErr error) {
	defer func() { resultErr = projectError(resultErr) }()
	// Re-parse programmatic requests through the same closed reader.
	b, err := canonical(d)
	if err != nil {
		return Preview{}, err
	}
	d, err = ParseDeclaration(b)
	if err != nil {
		return Preview{}, err
	}
	op, err := s.begin(false)
	if err != nil {
		return Preview{}, err
	}
	defer op.close()
	r, err := op.registered(id)
	if err != nil {
		return Preview{}, err
	}
	items := []Candidate{}
	omitted := []Omission{}
	for _, req := range d.Items {
		if req.Environment != r.Record.Environment {
			return Preview{}, fail("project_context_invalid", "item environment differs from registration")
		}
		if req.Packaging != "" {
			return Preview{}, fail("project_context_invalid", "instruction packaging requires the explicit package operation")
		}
		if req.Type == "mcp" {
			if req.Required {
				return Preview{}, fail("project_context_unsupported", "required "+req.Type+" needs its owning admission contract")
			}
			root, err := os.OpenRoot(r.Record.ProjectRoot)
			if err != nil {
				return Preview{}, err
			}
			b, _, err := read(root, req.Source, MaxRecord)
			_ = root.Close()
			if err != nil {
				return Preview{}, fail("project_context_unreadable", "unsupported candidate is missing or unreadable")
			}
			if err = protocoljson.Validate(b); err != nil {
				return Preview{}, fail("project_context_invalid", "unsupported MCP candidate is malformed JSON")
			}
			omitted = append(omitted, Omission{ItemID: req.ID, Reason: "unsupported", Source: req.Source, SourceSHA256: digest(b)})
			continue
		}
		var c Candidate
		var err error
		if req.Type == "skill" {
			c, err = op.skill(r, req)
		} else {
			c, err = op.context(r, req, nil)
		}
		if err != nil {
			if !req.Required && isCode(err, "project_context_unsupported") {
				if !hexRE.MatchString(c.SourceSHA256) {
					return Preview{}, err
				}
				omitted = append(omitted, Omission{ItemID: req.ID, Reason: "unsupported", Source: req.Source, SourceSHA256: c.SourceSHA256})
				continue
			}
			return Preview{}, err
		}
		items = append(items, c)
	}
	return op.preview(r, items, omitted)
}
func (op *operation) preview(r registered, items []Candidate, omitted []Omission) (Preview, error) {
	changes := []Change{}
	next := map[string]bool{}
	for _, c := range items {
		next[c.Item.ID] = true
		a, ok := op.state.Approvals[r.Record.CheckoutID+":"+c.Item.ID]
		before := ""
		effect := "added"
		if ok && !a.Revoked {
			before = a.Record.ItemSHA256
			effect = "changed"
			if before == c.ItemSHA256 {
				effect = "unchanged"
			}
		}
		files, err := op.fileDiff(r.Record.CheckoutID, c)
		if err != nil {
			return Preview{}, err
		}
		change := Change{ItemID: c.Item.ID, Effect: effect, Before: before, After: c.ItemSHA256, Files: files}
		if c.Item.Type == "skill" {
			after, err := op.skillSemantics(c)
			if err != nil {
				return Preview{}, err
			}
			change.Semantics = &SemanticChange{After: after}
			if prior, err := op.approvedCandidate(r.Record.CheckoutID, c.Item.ID); err != nil {
				return Preview{}, err
			} else if prior != nil && prior.Item.Type == "skill" {
				change.Semantics.Before, err = op.skillSemantics(*prior)
				if err != nil {
					return Preview{}, err
				}
			}
		}
		changes = append(changes, change)
	}
	for _, a := range op.state.Approvals {
		if a.Record.CheckoutID == r.Record.CheckoutID && !a.Revoked {
			prior, err := op.approvedCandidate(r.Record.CheckoutID, a.Record.ItemID)
			if err != nil {
				return Preview{}, err
			}
			c := *prior
			if !next[c.Item.ID] {
				if c.Item.Required {
					return Preview{}, fail("project_context_unsupported", "revoke the required item before removing its selection")
				}
				changes = append(changes, Change{ItemID: c.Item.ID, Effect: "removed", Before: c.ItemSHA256, Files: []FileChange{}})
			}
		}
	}
	// Preserve explicit item order; semantic changes and omissions are sorted.
	sort.Slice(changes, func(i, j int) bool { return changes[i].ItemID < changes[j].ItemID })
	sort.Slice(omitted, func(i, j int) bool { return omitted[i].ItemID < omitted[j].ItemID })
	p := Preview{Review: Review{1, "registration:" + r.Record.CheckoutID, identity(r.Record), items, changes, omitted}}
	p.PackageSHA256 = identity(p.Review)
	if err := verifyBinding(r); err != nil {
		return Preview{}, err
	}
	op.state.Reviews[r.Record.CheckoutID] = p
	if err := op.save(); err != nil {
		return Preview{}, err
	}
	return p, nil
}

// Review returns retained bytes and metadata for an exact operator preview.
// It never rereads live files as a replacement baseline.
func (s Store) Review(id string) (result Preview, resultErr error) {
	defer func() { resultErr = projectError(resultErr) }()
	op, err := s.begin(false)
	if err != nil {
		return Preview{}, err
	}
	defer op.close()
	if _, err := op.registered(id); err != nil {
		return Preview{}, err
	}
	p, ok := op.state.Reviews[id]
	if !ok {
		return Preview{}, fail("project_item_unapproved", "no preview exists")
	}
	for _, c := range p.Review.Items {
		if err := op.verify(c); err != nil {
			return Preview{}, err
		}
	}
	return p, nil
}
func (op *operation) verify(c Candidate) error {
	snap, err := op.get(c.Item.Payload.SnapshotRef, c.Item.Payload.SnapshotSHA256)
	if err != nil {
		return fail("project_context_drift", "protected snapshot differs from preview")
	}
	if c.Item.Type == "skill" {
		if len(c.Members) != 1 {
			return fail("project_context_drift", "incomplete skill pin")
		}
		pin, err := auditSkill(filepath.Dir(op.root), filepath.Join(op.root, "snapshots", c.Item.Payload.SnapshotSHA256, "files"))
		if err != nil {
			return err
		}
		if c.Members[0].Pin != "state sha256:"+pin || c.Members[0].SnapshotSHA256 != c.Item.Payload.SnapshotSHA256 {
			return fail("project_context_drift", "whole-skill pin changed")
		}
		return nil
	}
	b, err := op.lockBytes(c.Item.Payload.LockRef, c.Item.Payload.LockSHA256)
	if err != nil {
		return err
	}
	lock, err := contextlock.Parse(b)
	if err != nil || lock.SchemaVersion != 4 {
		return fail("project_context_drift", "invalid retained context lock")
	}
	members := map[string]Member{}
	for _, m := range c.Members {
		if _, exists := members[m.Name]; exists {
			return fail("project_context_drift", "duplicate retained closure member")
		}
		members[m.Name] = m
	}
	if len(members) != len(lock.Members) {
		return fail("project_context_drift", "incomplete retained closure")
	}
	for _, m := range lock.Members {
		record, ok := members[m.Name]
		if !ok || record.Pin != m.Pin() || m.Kind != "context" {
			return fail("project_context_drift", "closure member binding changed")
		}
		capt, err := op.get(record.SnapshotRef, record.SnapshotSHA256)
		if err != nil {
			return fail("project_context_drift", "protected dependency changed")
		}
		if _, err = manifest(capt); err != nil {
			return err
		}
		if m.StateHash != "" {
			pin, err := hashing.ContentSHA256Files(capt.Bytes, hashing.VersionV2)
			if err != nil || strings.TrimPrefix(pin, "sha256:") != m.StateHash {
				return fail("project_context_drift", "context content pin changed")
			}
		}
		if m.Name == lock.Root && !equalCapture(snap, capt) {
			return fail("project_context_drift", "root snapshot differs from closure")
		}
	}
	return nil
}
func (op *operation) fresh(r registered, c Candidate) error {
	root, err := os.OpenRoot(r.Record.ProjectRoot)
	if err != nil {
		return err
	}
	defer root.Close()
	if c.Request.Packaging == "instruction" {
		b, _, err := read(root, c.Request.Source, MaxFileBytes)
		if err != nil || digest(b) != c.SourceSHA256 {
			return fail("project_item_unapproved", "selected instruction changed; package and preview again")
		}
		return nil
	}
	live, err := capture(root, c.Request.Source)
	if err != nil {
		return fail("project_context_invalid", "candidate manifest or inventory is unreadable")
	}
	if identity(live.Manifest) != c.SourceSHA256 {
		return fail("project_item_unapproved", "candidate bytes changed; preview again")
	}
	return nil
}
func (s Store) Approve(id, itemID, itemDigest, packageDigest string) (result Approval, resultErr error) {
	defer func() { resultErr = projectError(resultErr) }()
	if !idRE.MatchString(itemID) || !hexRE.MatchString(itemDigest) || !hexRE.MatchString(packageDigest) {
		return Approval{}, fail("project_item_unapproved", "approval requires exact item and review digests")
	}
	op, err := s.begin(false)
	if err != nil {
		return Approval{}, err
	}
	defer op.close()
	r, err := op.registered(id)
	if err != nil {
		return Approval{}, err
	}
	p, ok := op.state.Reviews[id]
	if !ok || p.PackageSHA256 != packageDigest {
		return Approval{}, fail("project_item_unapproved", "review digest is not the protected current preview")
	}
	for _, c := range p.Review.Items {
		if c.Item.ID != itemID {
			continue
		}
		if c.ItemSHA256 != itemDigest {
			return Approval{}, fail("project_item_unapproved", "item digest differs from preview")
		}
		if err := op.verify(c); err != nil {
			return Approval{}, err
		}
		if err := op.fresh(r, c); err != nil {
			return Approval{}, err
		}
		if err := verifyBinding(r); err != nil {
			return Approval{}, err
		}
		a := Approval{"project-approval/1", id, itemID, itemDigest, r.Record.Environment, op.state.Principal, ceiling(r.Record)}
		op.state.Approvals[id+":"+itemID] = approved{Record: a, ReviewSHA256: packageDigest, Candidate: &c}
		if err := op.save(); err != nil {
			return Approval{}, err
		}
		return a, nil
	}
	return Approval{}, fail("project_item_unapproved", "item has no supported protected preview")
}
func (s Store) Revoke(id, itemID string) (resultErr error) {
	defer func() { resultErr = projectError(resultErr) }()
	if !idRE.MatchString(itemID) {
		return fail("project_item_unapproved", "invalid item ID")
	}
	op, err := s.begin(false)
	if err != nil {
		return err
	}
	defer op.close()
	if _, ok := op.state.Registrations[id]; !ok {
		return fail("project_registration_mismatch", "unknown checkout ID")
	}
	key := id + ":" + itemID
	a, ok := op.state.Approvals[key]
	if !ok {
		return fail("project_item_unapproved", "item has no protected approval")
	}
	a.Revoked = true
	op.state.Approvals[key] = a
	return op.save()
}

// Selection is a composition consumer's explicit item choice. Disabled
// optional items still need their own approval; required items cannot be
// omitted. Unsupported optional requests are returned as named omissions by
// preview, for the composition lane's effective-preview approval contract.
type Selection struct {
	ItemID   string
	Disabled bool
}
type Admitted struct {
	Items     []Candidate
	Omitted   []Omission
	Rendering string
}

func (s Store) Admitted(id string, selections []Selection) (result Admitted, resultErr error) {
	defer func() { resultErr = projectError(resultErr) }()
	op, err := s.begin(false)
	if err != nil {
		return Admitted{}, err
	}
	defer op.close()
	r, err := op.registered(id)
	if err != nil {
		return Admitted{}, err
	}
	p, ok := op.state.Reviews[id]
	if !ok {
		return Admitted{}, fail("project_item_unapproved", "no protected review")
	}
	wanted := map[string]bool{}
	// A mutable candidate preview cannot erase an actively approved requirement.
	for _, a := range op.state.Approvals {
		if a.Record.CheckoutID != id || a.Revoked {
			continue
		}
		prior, err := op.approvedCandidate(id, a.Record.ItemID)
		if err != nil {
			return Admitted{}, err
		}
		if !prior.Item.Required {
			continue
		}
		found := false
		for _, c := range p.Review.Items {
			if c.Item.ID == prior.Item.ID && c.ItemSHA256 == prior.ItemSHA256 {
				found = true
			}
		}
		if !found {
			return Admitted{}, fail("project_item_unapproved", "required approved item differs from the candidate preview")
		}
	}
	for _, sel := range selections {
		if _, ok := wanted[sel.ItemID]; ok {
			return Admitted{}, fail("project_context_invalid", "duplicate selection")
		}
		wanted[sel.ItemID] = sel.Disabled
	}
	out := Admitted{Items: []Candidate{}, Omitted: []Omission{}, Rendering: AppendAfterProfile}
	for _, c := range p.Review.Items {
		disabled, selected := wanted[c.Item.ID]
		if !selected || disabled {
			if c.Item.Required {
				return Admitted{}, fail("project_context_unsupported", "required item cannot be omitted")
			}
			if !selected {
				return Admitted{}, fail("project_item_unapproved", "optional omission must be explicit")
			}
		}
		a, ok := op.state.Approvals[id+":"+c.Item.ID]
		if !ok || a.Revoked || a.Record.CheckoutID != r.Record.CheckoutID || a.Record.ItemSHA256 != c.ItemSHA256 || a.Record.Environment != c.Item.Environment || a.Record.CeilingSHA256 != ceiling(r.Record) {
			return Admitted{}, fail("project_item_unapproved", "exact checkout item has no current approval")
		}
		if err := op.verify(c); err != nil {
			return Admitted{}, err
		}
		if disabled {
			out.Omitted = append(out.Omitted, Omission{ItemID: c.Item.ID, Reason: "disabled"})
		} else {
			out.Items = append(out.Items, c)
		}
		delete(wanted, c.Item.ID)
	}
	if len(wanted) != 0 {
		return Admitted{}, fail("project_item_unapproved", "unknown selected item")
	}
	out.Omitted = append(out.Omitted, p.Review.Omitted...)
	return out, nil
}

// ReadSnapshot exposes only already-approved immutable context bytes. No path
// supplied by a repository can select manager state or approval authority.
func (s Store) ReadSnapshot(id, itemID string) (result Snapshot, files map[string][]byte, resultErr error) {
	defer func() { resultErr = projectError(resultErr) }()
	op, err := s.begin(false)
	if err != nil {
		return Snapshot{}, nil, err
	}
	defer op.close()
	if _, err := op.registered(id); err != nil {
		return Snapshot{}, nil, err
	}
	p, ok := op.state.Reviews[id]
	if !ok {
		return Snapshot{}, nil, fail("project_item_unapproved", "no review")
	}
	for _, c := range p.Review.Items {
		if c.Item.ID == itemID {
			a, ok := op.state.Approvals[id+":"+itemID]
			if !ok || a.Revoked || a.Record.ItemSHA256 != c.ItemSHA256 {
				return Snapshot{}, nil, fail("project_item_unapproved", "item not approved")
			}
			if err := op.verify(c); err != nil {
				return Snapshot{}, nil, err
			}
			capt, err := op.get(c.Item.Payload.SnapshotRef, c.Item.Payload.SnapshotSHA256)
			return capt.Manifest, capt.Bytes, err
		}
	}
	return Snapshot{}, nil, fail("project_item_unapproved", "unknown item")
}

// ReviewContent is an inert operator view. It includes exact selected context
// manifests/modules and retained locks, plus complete file inventories. The
// digest covers their immutable identities; reviewing them grants no authority.
type ReviewContent struct {
	Preview     Preview             `json:"preview"`
	Inventories map[string]Snapshot `json:"inventories"`
	Text        map[string]string   `json:"text"`
}

func (s Store) ReviewContent(id string) (result ReviewContent, resultErr error) {
	defer func() { resultErr = projectError(resultErr) }()
	op, err := s.begin(false)
	if err != nil {
		return ReviewContent{}, err
	}
	defer op.close()
	if _, err = op.registered(id); err != nil {
		return ReviewContent{}, err
	}
	p, ok := op.state.Reviews[id]
	if !ok {
		return ReviewContent{}, fail("project_item_unapproved", "no protected preview")
	}
	out := ReviewContent{p, map[string]Snapshot{}, map[string]string{}}
	for _, c := range p.Review.Items {
		if err = op.verify(c); err != nil {
			return ReviewContent{}, err
		}
		for _, m := range c.Members {
			capt, err := op.get(m.SnapshotRef, m.SnapshotSHA256)
			if err != nil {
				return ReviewContent{}, err
			}
			key := c.Item.ID + "/" + m.Name
			out.Inventories[key] = capt.Manifest
			if c.Item.Type == "context" {
				man, err := manifest(capt)
				if err != nil {
					return ReviewContent{}, err
				}
				out.Text[key+"/agent-context.json"] = string(capt.Bytes["agent-context.json"])
				for _, mod := range man.Modules {
					p := path.Join("context", mod.Path)
					out.Text[key+"/"+p] = string(capt.Bytes[p])
				}
			} else if c.Item.Type == "skill" {
				for p, b := range capt.Bytes {
					if utf8.Valid(b) {
						out.Text[key+"/"+p] = string(b)
					}
				}
			}
		}
		if c.Item.Type == "context" {
			b, err := op.lockBytes(c.Item.Payload.LockRef, c.Item.Payload.LockSHA256)
			if err != nil {
				return ReviewContent{}, err
			}
			out.Text[c.Item.ID+"/context-lock.json"] = string(b)
		}
	}
	return out, nil
}

func (op *operation) fileDiff(id string, next Candidate) ([]FileChange, error) {
	current := map[string]string{}
	prior := map[string]string{}
	collect := func(c Candidate, m map[string]string) error {
		for _, member := range c.Members {
			snap, err := op.get(member.SnapshotRef, member.SnapshotSHA256)
			if err != nil {
				return err
			}
			for _, f := range snap.Manifest.Files {
				m[member.Name+"/"+f.Path] = f.SHA256 + ":" + f.Mode
			}
		}
		return nil
	}
	if err := collect(next, current); err != nil {
		return nil, err
	}
	if c, err := op.approvedCandidate(id, next.Item.ID); err != nil {
		return nil, err
	} else if c != nil {
		if err := collect(*c, prior); err != nil {
			return nil, err
		}
	}
	names := map[string]bool{}
	for p := range prior {
		names[p] = true
	}
	for p := range current {
		names[p] = true
	}
	paths := make([]string, 0, len(names))
	for p := range names {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	changes := []FileChange{}
	for _, p := range paths {
		before, after := prior[p], current[p]
		if before == after {
			continue
		}
		effect := "changed"
		if before == "" {
			effect = "added"
		}
		if after == "" {
			effect = "removed"
		}
		changes = append(changes, FileChange{p, effect, before, after})
	}
	return changes, nil
}

// Older protected approvals may recover their candidate only from an exact
// matching retained review. A newer unapproved preview is never a baseline.
func (op *operation) approvedCandidate(id, itemID string) (*Candidate, error) {
	a, ok := op.state.Approvals[id+":"+itemID]
	if !ok || a.Revoked {
		return nil, nil
	}
	if a.Candidate != nil {
		return a.Candidate, nil
	}
	for _, c := range op.state.Reviews[id].Review.Items {
		if c.Item.ID == itemID && c.ItemSHA256 == a.Record.ItemSHA256 {
			a.Candidate = &c
			op.state.Approvals[id+":"+itemID] = a
			return &c, nil
		}
	}
	return nil, fail("project_context_drift", "approved candidate evidence is missing; revoke before replacing it")
}
