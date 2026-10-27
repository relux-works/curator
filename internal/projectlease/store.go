package projectlease

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/homelease/persistence"
	homeleasestate "github.com/relux-works/curator/internal/homelease/state"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/stateread"
	"github.com/relux-works/curator/internal/transaction"
)

type entry struct {
	Binding           Binding     `json:"binding"`
	Composition       Composition `json:"composition"`
	State             Phase       `json:"state"`
	ProcessGeneration string      `json:"process_generation"`
}
type view struct {
	Composition Composition        `json:"composition"`
	Surfaces    map[string]Surface `json:"surfaces"`
}
type ledger struct {
	Kind   string           `json:"kind"`
	Leases map[string]entry `json:"leases"`
	Views  map[string]view  `json:"views"`
}

// Manager is read-only until an operation acquires the shared mutation lock.
// No background worker, TTL collector or launcher-exit hook owns these leases.
type Manager struct {
	home      string
	authority Authority
	sessions  Sessions
	hooks     transaction.Hooks
}

// New constructs a lease manager with explicit authority and session observers.
func New(home string, authority Authority, sessions Sessions) (*Manager, error) {
	locks, err := managerlock.New(home)
	if err != nil {
		return nil, err
	}
	return &Manager{home: locks.Home(), authority: authority, sessions: sessions}, nil
}
func statePath(home string) string {
	return filepath.Join(home, "state", "project-leases", "v1", "ledger.json")
}
func bindingPath(home, id string) string {
	return filepath.Join(home, "state", "project-leases", "v1", "bindings", id+".json")
}

// checkedRoute validates each existing component without walking native state.
// A missing final route is allowed for allocation; links are never followed.
func checkedRoute(root, path string) error {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("project lease path escapes manager home")
	}
	if err = pathboundary.ValidateRoot(root); err != nil {
		return err
	}
	current := root
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		if part == "." {
			continue
		}
		current = filepath.Join(current, part)
		observed, e := stateread.Lstat(current)
		if e != nil {
			return e
		}
		if observed.Kind == stateread.KindAbsent {
			return nil
		}
		info := observed.Info
		if info.IsDir() {
			if e = pathboundary.ValidateRoot(current); e != nil {
				return e
			}
		} else {
			if current != path {
				return fmt.Errorf("project lease parent is not a directory")
			}
			if e = pathboundary.CheckPrivateFile(current, info); e != nil {
				return e
			}
		}
	}
	return nil
}
func read(home, path string) ([]byte, error) {
	if err := checkedRoute(home, path); err != nil {
		return nil, err
	}
	f, err := pathboundary.OpenReadNoFollow(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > 16<<20 {
		return nil, fmt.Errorf("project lease record is not a bounded regular file")
	}
	return io.ReadAll(io.LimitReader(f, (16<<20)+1))
}
func load(home string) (*ledger, error) {
	path := statePath(home)
	// An absent project namespace preserves pre-project manager behavior.
	raw, err := read(home, path)
	if errors.Is(err, os.ErrNotExist) {
		if HasState(home) {
			return nil, fmt.Errorf("project lease ledger missing; retain all project roots")
		}
		// A dangling authority path is not absence.
		for p := path; p != home; p = filepath.Dir(p) {
			if info, e := os.Lstat(p); e == nil && info.Mode()&os.ModeSymlink != 0 {
				return nil, fmt.Errorf("unreadable project lease route")
			}
		}
		return &ledger{Kind: "project-lease-state/1", Leases: map[string]entry{}, Views: map[string]view{}}, nil
	}
	if err != nil {
		return nil, err
	}
	l, err := decodeLedger(raw)
	if err != nil {
		return nil, err
	}
	for id, e := range l.Leases {
		immutable, er := read(home, bindingPath(home, id))
		if er != nil {
			return nil, er
		}
		want, _ := canonical(e.Binding)
		if string(immutable) != string(want) {
			return nil, fmt.Errorf("immutable lease binding changed")
		}
	}
	return l, nil
}

func decodeLedger(raw []byte) (*ledger, error) {
	if len(raw) > maxRecord {
		return nil, fmt.Errorf("project lease state exceeds record limit")
	}
	var l ledger
	if err := protocoljson.UnmarshalCanonical(raw, &l); err != nil {
		return nil, err
	}
	if l.Kind != "project-lease-state/1" || l.Leases == nil || l.Views == nil {
		return nil, fmt.Errorf("invalid project lease state")
	}
	for id, e := range l.Leases {
		if err := e.Binding.validate(); err != nil {
			return nil, err
		}
		if id != e.Binding.LeaseID || e.Composition.validate() != nil || !matches(e.Binding, e.Composition) {
			return nil, fmt.Errorf("project lease identity mismatch")
		}
		if e.State != Prepared && e.State != Active && e.State != Unknown && e.State != Released {
			return nil, fmt.Errorf("invalid project lease phase")
		}
		if e.State == Active && e.ProcessGeneration == "" {
			return nil, fmt.Errorf("active project lease lacks runner identity")
		}
		if len(e.ProcessGeneration) > 256 {
			return nil, fmt.Errorf("invalid process generation")
		}
	}
	for key, v := range l.Views {
		if v.Composition.validate() != nil || key != homeRelative(v.Composition) || v.Surfaces == nil {
			return nil, fmt.Errorf("invalid project view identity")
		}
		for path, file := range v.Surfaces {
			if !safeSurface(path) || !hexPattern.MatchString(file.SHA256) || file.Size < 0 || file.Size > 16<<20 || (file.Mode != 0600 && file.Mode != 0700) {
				return nil, fmt.Errorf("invalid generated surface inventory")
			}
		}
	}
	roundtrip, er := canonical(l)
	if er != nil || string(roundtrip) != string(raw) {
		return nil, fmt.Errorf("project lease state has missing or null fields")
	}
	return &l, nil
}
func matches(b Binding, c Composition) bool {
	return b.ContextRef == c.Ref && b.ContextSHA256 == c.SHA256 && b.CheckoutID == c.CheckoutID && b.Environment == c.Environment && b.HomeRef == c.HomeRef
}
func nonce() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:])
}

func (m *Manager) locked(ctx context.Context, fn func(*ledger, *managerlock.HomeLock, *transaction.Engine) error) error {
	lm, err := managerlock.New(m.home)
	if err != nil {
		return err
	}
	lock, err := lm.AcquireHomeOnly(ctx, false)
	if err != nil {
		return err
	}
	defer func() { _ = lock.Close() }()
	if err = lock.AssertCurrent(); err != nil {
		return err
	}
	engine, err := transaction.New(m.home, transaction.WithHooks(m.hooks), transaction.WithCommitCheck(TransactionCheck(m.home)))
	if err != nil {
		return err
	}
	_, err = RecoverUnderLock(m.home, lock, engine)
	if err != nil {
		return err
	}
	l, err := load(m.home)
	if err != nil {
		return fmt.Errorf("project lease evidence unreadable; retain all roots: %w", err)
	}
	return fn(l, lock, engine)
}
func (m *Manager) publish(l *ledger, lock *managerlock.HomeLock, engine *transaction.Engine, files map[string][]byte, removals []string) error {
	targets := append([]string{}, removals...)
	for path := range files {
		targets = append(targets, path)
	}
	if err := homeleasestate.GuardMutation(m.home, lock, "project-leases", targets); err != nil {
		return err
	}
	raw, err := canonical(l)
	if err != nil {
		return err
	}
	if len(raw) > maxRecord {
		return fmt.Errorf("project lease state exceeds record limit")
	}
	files[statePath(m.home)] = raw
	return persistence.Publish(persistence.Publication{
		Home: m.home, StageDir: filepath.Dir(statePath(m.home)), TransactionID: "project-lease-" + nonce(),
		Owner: "project-leases", Class: "project-lease", Files: files, Removals: removals, Route: checkedRoute,
		Mode: func(path string) os.FileMode {
			if strings.HasPrefix(path, filepath.Join(m.home, "environments", "project-views")+string(filepath.Separator)) {
				for _, v := range l.Views {
					if rel, err := filepath.Rel(filepath.Join(m.home, homeRelative(v.Composition)), path); err == nil {
						if surface, ok := v.Surfaces[filepath.ToSlash(rel)]; ok {
							return os.FileMode(surface.Mode)
						}
					}
				}
			}
			return 0
		},
	}, lock, engine)
}
