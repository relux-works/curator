package projectcontext

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/identifiers"
	sourceidentity "github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/privatedir"
)

type boundedOutput struct {
	bytes.Buffer
	max int
}

func (b *boundedOutput) Write(p []byte) (int, error) {
	if len(p) > b.max-b.Len() {
		return 0, fmt.Errorf("bounded local Git output exceeded")
	}
	return b.Buffer.Write(p)
}

type localGitError struct{ code int }

func (e *localGitError) Error() string {
	return "project_registration_mismatch: local Git identity or object is unavailable, ambiguous or exceeds its bound"
}
func gitBytes(dir string, byteLimit int, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	prefix := []string{"--no-optional-locks", "--no-replace-objects", "-c", "core.fsmonitor=false", "-c", "core.hooksPath=" + os.DevNull, "-c", "core.untrackedCache=false", "-C", dir}
	cmd := exec.CommandContext(ctx, "git", append(prefix, args...)...) // #nosec G204 -- Git is a fixed executable with separate arguments from read-only internal call sites.
	for _, e := range os.Environ() {
		k, _, _ := strings.Cut(e, "=")
		if strings.HasPrefix(k, "GIT_") || k == "SSH_AUTH_SOCK" {
			continue
		}
		cmd.Env = append(cmd.Env, e)
	}
	cmd.Env = append(cmd.Env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_TERMINAL_PROMPT=0", "GIT_NO_LAZY_FETCH=1")
	out := &boundedOutput{max: byteLimit}
	cmd.Stdout = out
	cmd.Stderr = &boundedOutput{max: 4096}
	if err := cmd.Run(); err != nil {
		code := -1
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			code = exit.ExitCode()
		}
		return nil, &localGitError{code}
	}
	return out.Bytes(), nil
}
func gitLocal(dir string, args ...string) (string, error) {
	b, err := gitBytes(dir, MaxRecord, args...)
	return strings.TrimSpace(string(b)), err
}

func bind(p string) (binding, error) {
	var b binding
	abs, err := filepath.Abs(p)
	if err != nil {
		return b, fail("project_registration_mismatch", "root path cannot be resolved")
	}
	canonical, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return b, fail("project_registration_mismatch", "root is missing, moved or unreadable; explicitly rebind")
	}
	// Explicit registration canonicalizes an operator path once. Use requires
	// the saved canonical spelling and the retained filesystem identity.
	if !validRoot(canonical) {
		return b, fail("project_registration_mismatch", "root path is invalid or exceeds its bound")
	}
	st, err := os.Lstat(canonical)
	if err != nil || !st.IsDir() || st.Mode().Perm()&0500 != 0500 {
		return b, fail("project_registration_mismatch", "root is unreadable or not a directory")
	}
	f, err := os.Open(canonical)
	if err != nil {
		return b, fail("project_registration_mismatch", "root cannot be read; explicitly rebind")
	}
	_, readErr := f.Readdirnames(1)
	opened, statErr := f.Stat()
	_ = f.Close()
	if (readErr != nil && readErr != io.EOF) || statErr != nil || !same(st, opened) {
		return b, fail("project_registration_mismatch", "root cannot be verified; explicitly rebind")
	}
	fs, err := filesystemIdentity(st)
	if err != nil {
		return b, err
	}
	root, err := gitLocal(canonical, "rev-parse", "--show-toplevel")
	if err != nil {
		return b, err
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil || resolvedRoot != canonical {
		return b, fail("project_registration_mismatch", "registration requires the exact checkout root")
	}
	commit, err := gitLocal(canonical, "rev-list", "--max-parents=0", "HEAD")
	if err != nil || len(strings.Fields(commit)) != 1 || !contextCommit(commit) {
		return b, fail("project_registration_mismatch", "repository root commit is unavailable or ambiguous")
	}
	var remote *string
	raw, e := gitLocal(canonical, "config", "--get-all", "remote.origin.url")
	if e == nil {
		if len(strings.Fields(raw)) != 1 {
			return b, fail("project_registration_mismatch", "primary remote is ambiguous")
		}
		v, err := sourceidentity.Parse(raw)
		if err != nil || !sourceidentity.ValidCanonical(v) {
			return b, fail("project_registration_mismatch", "primary remote is not a canonical network identity")
		}
		remote = &v
	} else {
		var gitErr *localGitError
		if !errors.As(e, &gitErr) || gitErr.code != 1 {
			return b, fail("project_registration_mismatch", "primary remote configuration is unreadable")
		}

	}
	after, err := os.Lstat(canonical)
	if err != nil || !same(st, after) {
		return b, fail("project_registration_mismatch", "root changed while binding")
	}
	b = binding{repository{canonical, remote, commit}, fs}
	return b, nil
}

func validRoot(p string) bool {
	if !utf8.ValidString(p) || len(p) > 4096 || !filepath.IsAbs(p) || filepath.Clean(p) != p || strings.HasPrefix(p, "//") {
		return false
	}
	for _, r := range p {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}
func contextCommit(s string) bool {
	if len(s) != 40 && len(s) != 64 {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}
func verifyBinding(r registered) error {
	resolvedRoot, err := filepath.EvalSymlinks(r.Record.ProjectRoot)
	if err != nil || resolvedRoot != r.Record.ProjectRoot {
		return fail("project_registration_mismatch", "checkout moved or is unreadable; explicitly rebind")
	}
	b, err := bind(r.Record.ProjectRoot)
	if err != nil || identity(b) != r.Record.RootIdentity {
		return fail("project_registration_mismatch", "checkout binding changed; explicitly rebind")
	}
	return nil
}

// verifyRegistration retains observed loss of root authority independently of
// the immutable registration record. Restoring the old inode or permissions
// cannot revive approvals; only Rebind allocates new authority. Lapsed slots
// remain active shadows so a parent registration cannot inherit their CWDs.
func (op *operation) verifyRegistration(r registered) error {
	if r.Lapsed {
		return fail("project_registration_mismatch", "registration authority lapsed; explicitly rebind")
	}
	if err := verifyBinding(r); err != nil {
		for id, slot := range op.state.Registrations {
			if slot.Active && slot.Record.ProjectRoot == r.Record.ProjectRoot {
				slot.Lapsed = true
				op.state.Registrations[id] = slot
			}
		}
		if saveErr := op.retainLapses(); saveErr != nil {
			return fail("project_context_drift", "cannot retain root-authority lapse")
		}
		return err
	}
	return nil
}

// A missing CWD must not hide an observed missing registered root. Errors in
// an unrelated child directory leave a still-verifiable root's authority alone.
func (op *operation) observeUnavailableCWD(cwd string) error {
	abs, err := filepath.Abs(cwd)
	if err != nil {
		return fail("project_registration_mismatch", "CWD cannot be resolved")
	}
	// Canonicalize the surviving ancestor even when the registered leaf has
	// disappeared. A symlinked ancestor must not hide an observed root loss.
	suffix := []string{}
	for {
		if resolvedRoot, err := filepath.EvalSymlinks(abs); err == nil {
			abs = filepath.Join(append([]string{resolvedRoot}, suffix...)...)
			break
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return fail("project_registration_mismatch", "CWD cannot be resolved")
		}
		suffix = append([]string{filepath.Base(abs)}, suffix...)
		abs = parent
	}
	for _, id := range sortedRegistrationIDs(op.state) {
		r := op.state.Registrations[id]
		if r.Active && within(r.Record.ProjectRoot, abs) {
			if err := op.verifyRegistration(r); err != nil {
				return err
			}
		}
	}
	return fail("project_registration_mismatch", "CWD is unreadable or not a directory")
}

func profile(home, name string) (Profile, *contextlock.Lock, error) {
	if !identifiers.Valid(name) {
		return Profile{}, nil, fail("project_registration_mismatch", "invalid selected profile")
	}
	p := filepath.Join(home, "profiles", name)
	if err := pathboundary.Validate(p); err != nil {
		return Profile{}, nil, err
	}
	r, err := os.OpenRoot(p)
	if err != nil {
		return Profile{}, nil, err
	}
	defer func() { _ = r.Close() }()
	b, _, err := read(r, "lock.json", MaxRecord)
	if err != nil {
		return Profile{}, nil, err
	}
	l, err := contextlock.Parse(b)
	if err != nil {
		return Profile{}, nil, err
	}
	c, err := l.Canonical()
	if err != nil {
		return Profile{}, nil, err
	}
	return Profile{name, digest(c)}, l, nil
}

// Register binds a canonical checkout root to a fresh protected registration.
func (s Store) Register(name, root, profileName, env string) (result Registration, resultErr error) {
	defer func() { resultErr = projectError(resultErr) }()
	if name == "" {
		name = filepath.Base(filepath.Clean(root))
	}
	if !identifiers.Valid(name) || !environment(env) {
		return Registration{}, fail("project_registration_mismatch", "invalid name or environment")
	}
	b, err := bind(root)
	if err != nil {
		return Registration{}, err
	}
	// A manager store inside a source tree can never be protected from that
	// tree's namespace. Refuse before the first manager write.
	home, err := filepath.Abs(s.Home)
	if err != nil {
		return Registration{}, err
	}
	if resolvedRoot, e := filepath.EvalSymlinks(home); e == nil {
		home = resolvedRoot
	}
	if within(b.Repository.Root, home) {
		return Registration{}, fail("project_registration_mismatch", "manager home must be outside the source tree")
	}
	p, _, err := profile(home, profileName)
	if err != nil {
		return Registration{}, err
	}
	op, err := s.begin(true)
	if err != nil {
		return Registration{}, err
	}
	defer op.close()
	return op.register(name, b, p, env)
}
func (op *operation) register(name string, b binding, p Profile, env string) (Registration, error) {
	for _, old := range op.state.Registrations {
		if old.Active && ((old.Name == name && old.Record.ProjectRoot != b.Repository.Root) || (old.Record.ProjectRoot == b.Repository.Root && old.Record.Profile == p && old.Record.Environment == env)) {
			return Registration{}, fail("project_registration_mismatch", "registration already exists; explicitly rebind")
		}
	}
	id := newID()
	r := Registration{"project-registration/1", id, b.Repository.Root, identity(b), p, env, "home:" + id, "worktree:" + id}
	for _, part := range []string{"homes", "worktrees"} {
		p := filepath.Join(op.root, part, id)
		if err := privatedir.MakeAll(p); err != nil {
			return Registration{}, err
		}
	}
	current, l, err := profile(filepath.Dir(op.root), p.Name)
	if err != nil {
		return Registration{}, err
	}
	if current != p || !validPrecedence(op.precedence) || (l.Precedence.Winner != "" && l.Precedence != op.precedence) {
		return Registration{}, fail("project_context_drift", "profile or manager precedence changed")
	}
	op.state.Registrations[id] = registered{Record: r, Name: name, Binding: b, Active: true, Precedence: op.precedence}
	if err := verifyBinding(op.state.Registrations[id]); err != nil {
		return Registration{}, err
	}
	if err := op.save(); err != nil {
		return Registration{}, err
	}
	return r, nil
}

// Rebind allocates a new checkout ID. Old approvals remain inert historical
// records and cannot migrate merely because name, Git metadata or bytes agree.
func (s Store) Rebind(id, root string) (result Registration, resultErr error) {
	defer func() { resultErr = projectError(resultErr) }()
	b, err := bind(root)
	if err != nil {
		return Registration{}, err
	}
	home, err := filepath.EvalSymlinks(s.Home)
	if err != nil {
		return Registration{}, err
	}
	if within(b.Repository.Root, home) {
		return Registration{}, fail("project_registration_mismatch", "manager home is inside source tree")
	}
	op, err := s.begin(false)
	if err != nil {
		return Registration{}, err
	}
	defer op.close()
	old, ok := op.state.Registrations[id]
	if !ok || !old.Active {
		return Registration{}, fail("project_registration_mismatch", "unknown registration")
	}
	old.Active = false
	op.state.Registrations[id] = old
	p, _, err := profile(home, old.Record.Profile.Name)
	if err != nil {
		return Registration{}, err
	}
	return op.register(old.Name, b, p, old.Record.Environment)
}

// Locate proves the real launch CWD belongs to the most specific registered
// root. A stale nested binding refuses instead of falling back to its parent.
func (s Store) Locate(cwd string) (result Registration, resultErr error) {
	defer func() { resultErr = projectError(resultErr) }()
	return s.LocateFor(cwd, "", "")
}

// LocateFor disambiguates a checkout's separately registered profile/environment
// slots using explicit manager selection. A deeper registered root shadows its
// parent even when that deeper root has no slot for the requested environment.
func (s Store) LocateFor(cwd, profileName, env string) (result Registration, resultErr error) {
	defer func() { resultErr = projectError(resultErr) }()
	op, err := s.begin(false)
	if err != nil {
		return Registration{}, err
	}
	defer op.close()
	canonical, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		return Registration{}, op.observeUnavailableCWD(cwd)
	}
	st, err := os.Stat(canonical)
	if err != nil || !st.IsDir() {
		return Registration{}, op.observeUnavailableCWD(cwd)
	}
	return op.locateFor(canonical, profileName, env)
}
func (op *operation) locateFor(canonical, profileName, env string) (Registration, error) {
	best := ""
	matches := []registered{}
	for _, id := range sortedRegistrationIDs(op.state) {
		r := op.state.Registrations[id]
		if !r.Active || !within(r.Record.ProjectRoot, canonical) {
			continue
		}
		if len(r.Record.ProjectRoot) > len(best) {
			best = r.Record.ProjectRoot
			matches = []registered{r}
		} else if r.Record.ProjectRoot == best {
			matches = append(matches, r)
		}
	}
	if best == "" {
		return Registration{}, fail("project_registration_mismatch", "CWD has no registration")
	}
	selected := []registered{}
	for _, r := range matches {
		if (profileName == "" || r.Record.Profile.Name == profileName) && (env == "" || r.Record.Environment == env) {
			selected = append(selected, r)
		}
	}
	if len(selected) != 1 {
		return Registration{}, fail("project_registration_mismatch", "ambiguous or unavailable profile/environment registration")
	}
	r, err := op.registered(selected[0].Record.CheckoutID)
	if err != nil {
		return Registration{}, err
	}
	return r.Record, nil
}

// Inspect returns a registration and its inert project discovery inventory.
func (s Store) Inspect(id string) (result Inspection, resultErr error) {
	defer func() { resultErr = projectError(resultErr) }()
	op, err := s.begin(false)
	if err != nil {
		return Inspection{}, err
	}
	defer op.close()
	r, err := op.registered(id)
	if err != nil {
		return Inspection{}, err
	}
	root, err := os.OpenRoot(r.Record.ProjectRoot)
	if err != nil {
		return Inspection{}, err
	}
	defer func() { _ = root.Close() }()
	items, err := discovery(root)
	if err != nil {
		return Inspection{}, err
	}
	if err := op.verifyRegistration(r); err != nil {
		return Inspection{}, err
	}
	return Inspection{r.Record, "off", "off", AppendAfterProfile, items}, nil
}
