package projectcontext

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/stateread"
)

type binding struct {
	Repository repository        `json:"repository"`
	Filesystem map[string]string `json:"filesystem"`
}
type repository struct {
	Root          string  `json:"root"`
	PrimaryRemote *string `json:"primary_remote"`
	RootCommit    string  `json:"root_commit"`
}
type registered struct {
	Record     Registration           `json:"record"`
	Name       string                 `json:"name"`
	Binding    binding                `json:"binding"`
	Active     bool                   `json:"active"`
	Lapsed     bool                   `json:"lapsed,omitempty"`
	Precedence contextlock.Precedence `json:"precedence"`
}
type approved struct {
	Record       Approval   `json:"record"`
	ReviewSHA256 string     `json:"review_sha256"`
	Revoked      bool       `json:"revoked"`
	Candidate    *Candidate `json:"candidate,omitempty"`
}
type state struct {
	Version        int                   `json:"version"`
	ManagerID      string                `json:"manager_id"`
	Principal      string                `json:"principal"`
	ManagerBinding string                `json:"manager_binding"`
	Registrations  map[string]registered `json:"registrations"`
	Reviews        map[string]Preview    `json:"reviews"`
	Approvals      map[string]approved   `json:"approvals"`
}

// Store is scoped to an operator-controlled manager home. Protected records
// are trusted only inside that home, never from source paths or serialized
// records supplied by a repository. Same-UID hostile-code containment is not
// claimed. Agent approval requires a future manager-grant entry point; there
// is deliberately no caller-supplied approver or ceiling parameter here.
type Store struct {
	Home       string
	Precedence contextlock.Precedence
}

func (s Store) precedence() contextlock.Precedence {
	if s.Precedence.Winner == "" && s.Precedence.Placement == "" {
		return contextlock.Precedence{Winner: "higher-weight", Placement: "winner-last"}
	}
	return s.Precedence
}

type operation struct {
	state      state
	root       string
	lock       *managerlock.HomeLock
	precedence contextlock.Precedence
}

func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return "p" + hex.EncodeToString(b)
}
func (s Store) begin(create bool) (*operation, error) {
	abs, err := filepath.Abs(s.Home)
	if err != nil {
		return nil, err
	}
	s.Home = filepath.Clean(abs)
	m, err := managerlock.New(s.Home)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var lock *managerlock.HomeLock
	if create {
		lock, err = m.AcquireHomeOnly(ctx, false)
	} else {
		lock, err = m.AcquireExistingHomeOnly(ctx)
	}
	if err != nil {
		return nil, err
	}
	op := &operation{root: filepath.Join(m.Home(), "project-context"), lock: lock, precedence: s.precedence()}
	ok := false
	defer func() {
		if !ok {
			_ = lock.Close()
		}
	}()
	if err := pathboundary.ValidateRoot(m.Home()); err != nil {
		return nil, err
	}
	rootState, rootErr := stateread.Lstat(op.root)
	if rootErr != nil {
		return nil, rootErr
	}
	freshRoot := rootState.Kind == stateread.KindAbsent
	if create {
		if err := privatedir.MakeAll(op.root); err != nil {
			return nil, err
		}
	}
	if err := pathboundary.Validate(op.root); err != nil {
		return nil, fail("project_context_drift", "protected admission boundary is invalid")
	}
	r, err := os.OpenRoot(op.root)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	b, _, err := read(r, "state.json", MaxRecord)
	if errors.Is(err, os.ErrNotExist) && create && freshRoot {
		homeBinding, err := managerBinding(m.Home())
		if err != nil {
			return nil, err
		}
		op.state = state{Version: 2, ManagerID: newID(), Principal: "principal:" + newID(), ManagerBinding: homeBinding, Registrations: map[string]registered{}, Reviews: map[string]Preview{}, Approvals: map[string]approved{}}
		if err := op.retainLapses(); err != nil {
			return nil, err
		}
		if err := op.save(); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, fail("project_context_drift", "cannot read protected admission state")
	} else {
		if err := decode(b, &op.state); err != nil {
			return nil, err
		}
		can, err := canonical(op.state)
		if err != nil || string(can) != string(b) {
			return nil, fail("project_context_drift", "noncanonical protected admission state")
		}
		if err := op.validate(); err != nil {
			return nil, err
		}
	}
	if err := op.loadLapses(r); err != nil {
		return nil, err
	}
	if op.state.Version == 1 {
		// Establish the companion before requiring it. Changing the envelope
		// version uses the same number of bytes, even in a full old ledger.
		if err := op.retainLapses(); err != nil {
			return nil, err
		}
		op.state.Version = 2
		if err := op.save(); err != nil {
			return nil, err
		}
	}
	ok = true
	return op, nil
}
func (op *operation) close() { _ = op.lock.Close() }
func (op *operation) validate() error {
	s := op.state
	binding, err := managerBinding(filepath.Dir(op.root))
	if err != nil || s.ManagerBinding != binding {
		return fail("project_context_drift", "admission records belong to another manager allocation")
	}
	if (s.Version != 1 && s.Version != 2) || !idRE.MatchString(s.ManagerID) || !strings.HasPrefix(s.Principal, "principal:") || !idRE.MatchString(strings.TrimPrefix(s.Principal, "principal:")) || s.Registrations == nil || s.Reviews == nil || s.Approvals == nil || len(s.Registrations) > 256 || len(s.Reviews) > 256 || len(s.Approvals) > 256 {
		return fail("project_context_drift", "invalid protected state envelope")
	}
	for id, r := range s.Registrations {
		if !validPrecedence(r.Precedence) {
			return fail("project_context_drift", "invalid retained precedence")
		}
		v := r.Record
		if id != v.CheckoutID || !idRE.MatchString(id) || v.Kind != "project-registration/1" || !identifiers.Valid(r.Name) || !environment(v.Environment) || !validRoot(v.ProjectRoot) || !validRoot(r.Binding.Repository.Root) || v.RootIdentity != identity(r.Binding) || !identifiers.Valid(v.Profile.Name) || !hexRE.MatchString(v.Profile.LockSHA256) || v.HomeRef != "home:"+id || v.WorktreeRef != "worktree:"+id {
			return fail("project_context_drift", "invalid protected registration")
		}
	}
	for id, p := range s.Reviews {
		r, ok := s.Registrations[id]
		if !ok || p.PackageSHA256 != identity(p.Review) || p.Review.Version != 1 || p.Review.RegistrationRef != "registration:"+id || p.Review.RegistrationSHA256 != identity(r.Record) || len(p.Review.Items) > 256 {
			return fail("project_context_drift", "invalid protected review")
		}
		seen := map[string]bool{}
		for _, c := range p.Review.Items {
			if seen[c.Item.ID] || validateCandidate(c, r.Record) != nil {
				return fail("project_context_drift", "invalid protected item")
			}
			seen[c.Item.ID] = true
		}
	}
	for key, a := range s.Approvals {
		r, ok := s.Registrations[a.Record.CheckoutID]
		v := a.Record
		if !ok || key != v.CheckoutID+":"+v.ItemID || v.Kind != "project-approval/1" || !idRE.MatchString(v.ItemID) || !hexRE.MatchString(v.ItemSHA256) || v.Environment != r.Record.Environment || v.ApproverRef != s.Principal || v.CeilingSHA256 != ceiling(r.Record) || !hexRE.MatchString(a.ReviewSHA256) {
			return fail("project_context_drift", "invalid protected approval")
		}
		if a.Candidate != nil && (validateCandidate(*a.Candidate, r.Record) != nil || a.Candidate.ItemSHA256 != v.ItemSHA256 || a.Candidate.Item.ID != v.ItemID) {
			return fail("project_context_drift", "approved candidate differs from its immutable identity")
		}
	}
	return nil
}
func ceiling(r Registration) string {
	return identity(map[string]any{"version": 1, "environment": r.Environment, "context": "additive-only", "permissions": "manager-only-default-deny", "memory": "off", "executable": "owning-admission-required"})
}
func validateCandidate(c Candidate, r Registration) error {
	i := c.Item
	if i.Kind != "project-item/1" || !idRE.MatchString(i.ID) || i.ID != c.Request.ID || i.Environment != r.Environment || i.Environment != c.Request.Environment || i.Required != c.Request.Required || i.Type != c.Request.Type || (i.Type != "context" && i.Type != "skill") || !hexRE.MatchString(i.Payload.SnapshotSHA256) || i.Payload.SnapshotRef != snapshotRef(i.Payload.SnapshotSHA256) || c.ItemSHA256 != identity(i) || c.Rendering != AppendAfterProfile || !hexRE.MatchString(c.SourceSHA256) || !relative(c.Request.Source) {
		return fail("project_context_invalid", "invalid closed candidate")
	}
	if i.Type == "context" {
		if !hexRE.MatchString(i.Payload.LockSHA256) || i.Payload.LockRef != lockRef(i.Payload.LockSHA256) {
			return fail("project_context_invalid", "invalid context lock binding")
		}
	} else if i.Payload.LockRef != "" || i.Payload.LockSHA256 != "" || c.Request.Packaging != "" {
		return fail("project_context_invalid", "skill cannot carry context lock or packaging")
	}
	return nil
}

func (op *operation) save() error {
	if err := op.validate(); err != nil {
		return err
	}
	b, err := canonical(op.state)
	if err != nil {
		return err
	}
	if len(b) > MaxRecord {
		return fail("project_context_invalid", "protected state capacity exceeded")
	}
	return writeAtomic(op.root, "state.json", b)
}
func writeAtomic(root, name string, b []byte) error {
	if err := pathboundary.ValidateRoot(root); err != nil {
		return err
	}
	f, err := os.CreateTemp(root, ".write-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(b)
	}
	if err == nil {
		err = f.Sync()
	}
	e := f.Close()
	if err == nil {
		err = e
	}
	if err != nil {
		return err
	}
	target := filepath.Join(root, name)
	state, err := stateread.Lstat(target)
	if err != nil {
		return err
	}
	if state.Kind == stateread.KindPresent && (!state.Info.Mode().IsRegular() || state.Info.Mode()&os.ModeSymlink != 0) {
		return fail("project_context_drift", "invalid protected record target")
	}
	if err = os.Rename(tmp, target); err != nil {
		return err
	}
	return syncDir(root)
}
func syncDir(p string) error {
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}
func hashID(hash string) string {
	b, err := hex.DecodeString(hash)
	if err != nil {
		panic(err)
	}
	return "p" + strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b))
}
func snapshotRef(hash string) string { return "snapshot:" + hashID(hash) }
func lockRef(hash string) string     { return "lock:" + hashID(hash) }
func (op *operation) put(c captured) (string, string, error) {
	h := identity(c.Manifest)
	ref := snapshotRef(h)
	base := filepath.Join(op.root, "snapshots")
	if err := privatedir.MakeAll(base); err != nil {
		return "", "", err
	}
	target := filepath.Join(base, h)
	state, err := stateread.Lstat(target)
	if err != nil {
		return "", "", err
	}
	if state.Kind == stateread.KindPresent {
		if _, err := op.get(ref, h); err != nil {
			return "", "", err
		}
		return ref, h, nil
	}
	stage, err := os.MkdirTemp(base, ".snapshot-*")
	if err != nil {
		return "", "", err
	}
	defer os.RemoveAll(stage)
	for _, e := range c.Manifest.Files {
		if !relative(e.Path) {
			return "", "", fail("project_context_invalid", "unsafe snapshot path")
		}
		p := filepath.Join(stage, "files", filepath.FromSlash(e.Path))
		if err = privatedir.MakeAll(filepath.Dir(p)); err != nil {
			return "", "", err
		}
		mode := os.FileMode(0600)
		if e.Mode == "0700" {
			mode = 0700
		}
		if err = os.WriteFile(p, c.Bytes[e.Path], mode); err != nil {
			return "", "", err
		}
	}
	b, _ := canonical(c.Manifest)
	if err = os.WriteFile(filepath.Join(stage, "manifest.json"), b, 0600); err != nil {
		return "", "", err
	}
	// No active record can reference a partially published tree. Re-read exact
	// inventory before atomic rename; an orphan after a crash has no authority.
	if err = syncTree(stage); err != nil {
		return "", "", err
	}
	if err = os.Rename(stage, target); err != nil {
		return "", "", err
	}
	if err = syncDir(base); err != nil {
		return "", "", err
	}
	if _, err = op.get(ref, h); err != nil {
		return "", "", err
	}
	return ref, h, nil
}
func (op *operation) get(ref, hash string) (captured, error) {
	var empty captured
	if !hexRE.MatchString(hash) || ref != snapshotRef(hash) {
		return empty, fail("project_context_drift", "invalid snapshot reference")
	}
	p := filepath.Join(op.root, "snapshots", hash)
	if err := pathboundary.Validate(p); err != nil {
		return empty, fail("project_context_drift", "invalid protected snapshot boundary")
	}
	r, err := os.OpenRoot(p)
	if err != nil {
		return empty, err
	}
	defer r.Close()
	b, _, err := read(r, "manifest.json", MaxRecord)
	if err != nil {
		return empty, err
	}
	var manifest Snapshot
	if err = decode(b, &manifest); err != nil {
		return empty, err
	}
	if identity(manifest) != hash {
		return empty, fail("project_context_drift", "snapshot manifest changed")
	}
	c, err := captureTree(r, "files", false)
	if err != nil {
		return empty, err
	}
	if err := validateInventory(c); err != nil {
		return empty, fail("project_context_drift", "protected snapshot inventory is invalid")
	}
	for _, entry := range c.Manifest.Files {
		st, e := r.Lstat("files/" + entry.Path)
		if e != nil || (st.Mode().Perm() != 0600 && st.Mode().Perm() != 0700) {
			return empty, fail("project_context_drift", "protected file mode changed")
		}
	}
	if identity(c.Manifest) != hash {
		return empty, fail("project_context_drift", "snapshot bytes, entry set or modes changed")
	}
	return c, nil
}
func (op *operation) putLock(b []byte) (string, string, error) {
	h := digest(b)
	dir := filepath.Join(op.root, "locks")
	if err := privatedir.MakeAll(dir); err != nil {
		return "", "", err
	}
	name := h + ".json"
	p := filepath.Join(dir, name)
	rr, err := os.OpenRoot(dir)
	if err != nil {
		return "", "", err
	}
	defer rr.Close()
	_ = p
	if old, _, err := read(rr, name, MaxRecord); err == nil {
		if string(old) != string(b) {
			return "", "", fail("project_context_drift", "retained lock changed")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", "", err
	} else if err = writeAtomic(dir, name, b); err != nil {
		return "", "", err
	}
	return lockRef(h), h, nil
}
func (op *operation) lockBytes(ref, hash string) ([]byte, error) {
	if !hexRE.MatchString(hash) || ref != lockRef(hash) {
		return nil, fail("project_context_drift", "invalid lock reference")
	}
	r, err := os.OpenRoot(filepath.Join(op.root, "locks"))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	b, _, err := read(r, hash+".json", MaxRecord)
	if err != nil || digest(b) != hash {
		return nil, fail("project_context_drift", "retained lock changed or missing")
	}
	return b, nil
}
func (op *operation) registered(id string) (registered, error) {
	if !idRE.MatchString(id) {
		return registered{}, fail("project_registration_mismatch", "invalid checkout ID")
	}
	r, ok := op.state.Registrations[id]
	if !ok || !r.Active {
		return r, fail("project_registration_mismatch", "unknown or retired registration")
	}
	if err := op.verifyRegistration(r); err != nil {
		return r, err
	}
	if err := op.policyFresh(r); err != nil {
		return r, err
	}
	return r, nil
}
func sortedRegistrationIDs(s state) []string {
	ids := make([]string, 0, len(s.Registrations))
	for id := range s.Registrations {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// MarshalRecord returns exact CCJ-1 for schema validation and consumers.
func MarshalRecord(v any) ([]byte, error) { return canonical(v) }

func syncTree(root string) error {
	dirs := []string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			dirs = append(dirs, p)
			return nil
		}
		f, err := os.Open(p)
		if err != nil {
			return err
		}
		err = f.Sync()
		closeErr := f.Close()
		if err != nil {
			return err
		}
		return closeErr
	})
	if err != nil {
		return err
	}
	for i := len(dirs) - 1; i >= 0; i-- {
		if err := syncDir(dirs[i]); err != nil {
			return err
		}
	}
	return nil
}

func managerBinding(home string) (string, error) {
	st, err := os.Lstat(home)
	if err != nil {
		return "", err
	}
	fs, err := filesystemIdentity(st)
	if err != nil {
		return "", err
	}
	return identity(map[string]any{"root": home, "filesystem": fs}), nil
}

func validPrecedence(p contextlock.Precedence) bool {
	return (p.Winner == "higher-weight" || p.Winner == "lower-weight") && (p.Placement == "winner-last" || p.Placement == "winner-first")
}
func (op *operation) policyFresh(r registered) error {
	p, l, err := profile(filepath.Dir(op.root), r.Record.Profile.Name)
	if err != nil {
		return fail("project_context_unreadable", "selected profile lock is unavailable")
	}
	if p != r.Record.Profile || r.Precedence != op.precedence || (l.Precedence.Winner != "" && l.Precedence != r.Precedence) {
		return fail("project_context_drift", "selected profile or precedence changed; explicitly rebind")
	}
	return nil
}
