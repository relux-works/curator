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
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/identifiers"
	sourceidentity "github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/privatedir"
)

type boundedOutput struct {
	buffer   bytes.Buffer
	max      int
	exceeded bool
}

func (b *boundedOutput) Write(p []byte) (int, error) {
	if len(p) > b.max-b.buffer.Len() {
		b.exceeded = true
		return 0, fmt.Errorf("bounded local Git output exceeded")
	}
	return b.buffer.Write(p)
}

type localGitError struct {
	code  int
	class string
	errno string
	site  string
}

// Only static classes, call-site labels and errno names cross this boundary.
// In particular, exec errors contain paths and must never be interpolated.
func (e *localGitError) diagnostic() string {
	detail := e.class
	if detail == "" || detail == "exit" {
		detail = fmt.Sprintf("exit %d", e.code)
	}
	if e.class == "start refused" && e.errno != "" {
		detail += " (" + e.errno + ")"
	}
	site := ""
	switch e.site {
	case "rev-parse", "rev-list", "config":
		site = " at " + e.site
	}
	kind := "identity"
	if e.site == "ls-tree" || e.site == "tag" || e.site == "cat-file" {
		kind, site = "object", " at "+e.site
	}
	return "local Git " + kind + " read failed" + site + ": " + detail
}

func localGitFailure(err error, ctxErr error, bounded bool, site string) *localGitError {
	e := &localGitError{code: -1, site: site}
	var exit *exec.ExitError
	switch {
	case bounded:
		e.class = "output bound exceeded"
	case ctxErr != nil:
		e.class = "timed out"
	case errors.As(err, &exit):
		e.code = exit.ExitCode()
		e.class = "exit"
		if e.code < 0 {
			e.class = "signal"
		}
	default:
		e.class = "start refused"
		switch {
		case errors.Is(err, syscall.EACCES):
			e.errno = "EACCES"
		case errors.Is(err, syscall.EPERM):
			e.errno = "EPERM"
		case errors.Is(err, syscall.ENOENT):
			e.errno = "ENOENT"
		case errors.Is(err, syscall.ENOEXEC):
			e.errno = "ENOEXEC"
		case errors.Is(err, syscall.EMFILE):
			e.errno = "EMFILE"
		case errors.Is(err, syscall.ENFILE):
			e.errno = "ENFILE"
		case errors.Is(err, syscall.EAGAIN):
			e.errno = "EAGAIN"
		}
	}
	return e
}

func (e *localGitError) Error() string {
	code := "project_registration_mismatch"
	if e.site == "ls-tree" || e.site == "tag" || e.site == "cat-file" {
		code = "project_context_invalid"
	}
	return code + ": " + e.diagnostic()
}
func gitRead(dir string, out io.Writer, exceeded func() bool, args ...string) error {
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
	cmd.Stdout = out
	stderr := &boundedOutput{max: 4096}
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		site := ""
		if len(args) > 0 {
			site = args[0]
		}
		return localGitFailure(err, ctx.Err(), exceeded() || stderr.exceeded, site)
	}
	return nil
}
func gitBytes(dir string, byteLimit int, args ...string) ([]byte, error) {
	out := &boundedOutput{max: byteLimit}
	if err := gitRead(dir, out, func() bool { return out.exceeded }, args...); err != nil {
		return nil, err
	}
	return out.buffer.Bytes(), nil
}
func gitLocal(dir string, args ...string) (string, error) {
	b, err := gitBytes(dir, MaxRecord, args...)
	return strings.TrimSpace(string(b)), err
}

// bindingReadError carries only static, path-free detail owned by bind.
type bindingReadError struct {
	detail  string
	unknown bool
}

func (e *bindingReadError) Error() string {
	return "project_registration_mismatch: " + e.detail
}
func bindingFailure(detail string) error { return &bindingReadError{detail: detail} }

// Unknown observations refuse use without converting an unreadable tuple into
// a durable identity change. The diagnostic text remains the same safe refusal.
type bindingUnknownError struct{ error }

func (e *bindingUnknownError) Unwrap() error { return e.error }
func bindingUnknown(detail string) error     { return &bindingReadError{detail: detail, unknown: true} }
func bindingPathReadFailure(err error, detail string) error {
	// Missing-root lapse treatment is deliberately unchanged pending its decision.
	if errors.Is(err, os.ErrNotExist) {
		return bindingFailure(detail)
	}
	return bindingUnknown(detail)
}
func unknownBindingRead(err error) bool {
	var gitErr *localGitError
	if errors.As(err, &gitErr) {
		// Exit 128 can mean that the recorded path is no longer a repository.
		// Keep that conservative outcome until the open policy question is settled.
		return gitErr.class != "exit" || gitErr.code != 128
	}
	var readErr *bindingReadError
	if errors.As(err, &readErr) {
		return readErr.unknown
	}
	return true
}

func bind(p string) (binding, error) {
	var b binding
	abs, err := filepath.Abs(p)
	if err != nil {
		return b, bindingFailure("root path cannot be resolved")
	}
	canonical, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return b, bindingPathReadFailure(err, "root is missing, moved or unreadable; explicitly rebind")
	}
	// Explicit registration canonicalizes an operator path once. Use requires
	// the saved canonical spelling and the retained filesystem identity.
	if !validRoot(canonical) {
		return b, bindingFailure("root path is invalid or exceeds its bound")
	}
	st, err := os.Lstat(canonical)
	if err != nil {
		return b, bindingPathReadFailure(err, "root is unreadable or not a directory")
	}
	if !st.IsDir() || st.Mode().Perm()&0500 != 0500 {
		return b, bindingFailure("root is unreadable or not a directory")
	}
	f, err := os.Open(canonical)
	if err != nil {
		return b, bindingPathReadFailure(err, "root cannot be read; explicitly rebind")
	}
	_, readErr := f.Readdirnames(1)
	opened, statErr := f.Stat()
	_ = f.Close()
	if (readErr != nil && readErr != io.EOF) || statErr != nil {
		return b, bindingUnknown("root cannot be verified; explicitly rebind")
	}
	if !same(st, opened) {
		return b, bindingFailure("root cannot be verified; explicitly rebind")
	}
	fs, err := filesystemIdentity(st)
	if err != nil {
		return b, bindingUnknown("checkout identity is unavailable")
	}
	root, err := gitLocal(canonical, "rev-parse", "--show-toplevel")
	if err != nil {
		return b, err
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return b, bindingPathReadFailure(err, "registration requires the exact checkout root")
	}
	if resolvedRoot != canonical {
		return b, bindingFailure("registration requires the exact checkout root")
	}
	commit, err := gitLocal(canonical, "rev-list", "--max-parents=0", "HEAD")
	if err != nil {
		return b, err
	}
	if len(strings.Fields(commit)) != 1 || !contextCommit(commit) {
		return b, bindingFailure("repository root commit is unavailable or ambiguous")
	}
	var remote *string
	raw, e := gitLocal(canonical, "config", "--get-all", "remote.origin.url")
	if e == nil {
		if len(strings.Fields(raw)) != 1 {
			return b, bindingFailure("primary remote is ambiguous")
		}
		v, err := sourceidentity.Parse(raw)
		if err != nil || !sourceidentity.ValidCanonical(v) {
			return b, bindingFailure("primary remote is not a canonical network identity")
		}
		remote = &v
	} else {
		var gitErr *localGitError
		if !errors.As(e, &gitErr) || gitErr.code != 1 || gitErr.class != "exit" {
			return b, e
		}

	}
	after, err := os.Lstat(canonical)
	if err != nil {
		return b, bindingPathReadFailure(err, "root changed while binding")
	}
	if !same(st, after) {
		return b, bindingFailure("root changed while binding")
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
		refusal := fail("project_registration_mismatch", "checkout moved or is unreadable; explicitly rebind")
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return &bindingUnknownError{refusal}
		}
		return refusal
	}
	b, err := bind(r.Record.ProjectRoot)
	return verifyBindingResult(r, b, err)
}

// Report the failed observation separately from an observed identity change.
// Keep the digest comparison authoritative; field labels add diagnostics only.
func verifyBindingResult(r registered, b binding, err error) error {
	if err != nil {
		detail := "checkout identity is unavailable"
		var readErr *bindingReadError
		var gitErr *localGitError
		switch {
		case errors.As(err, &readErr):
			detail = strings.TrimSuffix(readErr.detail, "; explicitly rebind")
		case errors.As(err, &gitErr):
			detail = gitErr.diagnostic()
		}
		refusal := fail("project_registration_mismatch", "checkout binding recheck failed: "+detail+"; explicitly rebind")
		if unknownBindingRead(err) {
			return &bindingUnknownError{refusal}
		}
		return refusal
	}
	if identity(b) != r.Record.RootIdentity {
		return fail("project_registration_mismatch", "checkout binding changed: "+strings.Join(bindingChangedFields(r.Binding, b), ", ")+"; explicitly rebind")
	}
	return nil
}

func bindingChangedFields(saved, live binding) []string {
	fields := []string{}
	if saved.Repository.Root != live.Repository.Root {
		fields = append(fields, "repository.root")
	}
	a, b := saved.Repository.PrimaryRemote, live.Repository.PrimaryRemote
	if (a == nil) != (b == nil) || (a != nil && b != nil && *a != *b) {
		fields = append(fields, "repository.primary_remote")
	}
	if saved.Repository.RootCommit != live.Repository.RootCommit {
		fields = append(fields, "repository.root_commit")
	}
	for _, key := range []string{"device", "inode"} {
		a, aok := saved.Filesystem[key]
		b, bok := live.Filesystem[key]
		if aok != bok || a != b {
			fields = append(fields, "filesystem."+key)
		}
	}
	// Unknown future filesystem fields and digest-only inconsistencies still
	// refuse, without echoing arbitrary keys, paths or identity values.
	if len(fields) == 0 {
		if identity(saved.Filesystem) != identity(live.Filesystem) {
			fields = append(fields, "filesystem")
		} else {
			fields = append(fields, "root_identity")
		}
	}
	return fields
}

// verifyRegistration refuses unknown metadata without writing a lapse. It
// retains observed loss of root authority independently of
// the immutable registration record. Restoring the old inode or permissions
// cannot revive approvals; only Rebind allocates new authority. Lapsed slots
// remain active shadows so a parent registration cannot inherit their CWDs.
func (op *operation) verifyRegistration(r registered) error {
	if r.Lapsed {
		return fail("project_registration_mismatch", "registration authority lapsed; explicitly rebind")
	}
	if err := verifyBinding(r); err != nil {
		var unknown *bindingUnknownError
		if errors.As(err, &unknown) {
			return err
		}
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

// Register binds a canonical checkout root to a fresh strict registration.
func (s Store) Register(name, root, profileName, env string) (Registration, error) {
	return s.RegisterMode(name, root, profileName, env, "strict")
}

// RegisterMode is an explicit protected operator operation. Repository data and
// composition requests cannot enable native discovery or change a registration.
func (s Store) RegisterMode(name, root, profileName, env, mode string) (result Registration, resultErr error) {
	defer func() { resultErr = projectError(resultErr) }()
	if !validSourceMode(mode) {
		return Registration{}, fail("project_context_invalid", "unknown source mode")
	}
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
	return op.register(name, b, p, env, mode)
}
func (op *operation) register(name string, b binding, p Profile, env, mode string) (Registration, error) {
	for _, old := range op.state.Registrations {
		if old.Active && ((old.Name == name && old.Record.ProjectRoot != b.Repository.Root) || (old.Record.ProjectRoot == b.Repository.Root && old.Record.Profile == p && old.Record.Environment == env)) {
			return Registration{}, fail("project_registration_mismatch", "registration already exists; explicitly rebind")
		}
	}
	id := newID()
	r := Registration{"project-registration/1", id, b.Repository.Root, identity(b), p, env, mode, "home:" + id, "worktree:" + id}
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
	return op.register(old.Name, b, p, old.Record.Environment, old.Record.Mode)
}

// Locate proves the resolved launch CWD belongs to the most specific registered
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
	return op.locateFor(cwd, profileName, env)
}
func (op *operation) locateFor(cwd, profileName, env string) (Registration, error) {
	canonical, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		return Registration{}, op.observeUnavailableCWD(cwd)
	}
	st, err := os.Stat(canonical)
	if err != nil || !st.IsDir() {
		return Registration{}, op.observeUnavailableCWD(cwd)
	}
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
	return inspectionStatus(r.Record, items), nil
}
