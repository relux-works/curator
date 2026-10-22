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
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/privatedir"
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
	if len(raw) > maxRecord {
		return nil, fmt.Errorf("project lease state exceeds record limit")
	}
	var l ledger
	if err = protocoljson.UnmarshalCanonical(raw, &l); err != nil {
		return nil, err
	}
	if l.Kind != "project-lease-state/1" || l.Leases == nil || l.Views == nil {
		return nil, fmt.Errorf("invalid project lease state")
	}
	for id, e := range l.Leases {
		if err = e.Binding.validate(); err != nil {
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
		immutable, er := read(home, bindingPath(home, id))
		if er != nil {
			return nil, er
		}
		want, _ := canonical(e.Binding)
		if string(immutable) != string(want) {
			return nil, fmt.Errorf("immutable lease binding changed")
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
	engine, err := transaction.New(m.home, transaction.WithHooks(m.hooks))
	if err != nil {
		return err
	}
	recovered, err := engine.RecoverWithOutcome(lock)
	if err != nil {
		return err
	}
	l, err := load(m.home)
	if err != nil {
		return fmt.Errorf("project lease evidence unreadable; retain all roots: %w", err)
	}
	if recovered {
		for id, e := range l.Leases {
			if e.State != Released {
				e.State = Unknown
				l.Leases[id] = e
			}
		}
		if err = m.publish(l, lock, engine, map[string][]byte{}, nil); err != nil {
			return err
		}
	}
	return fn(l, lock, engine)
}
func (m *Manager) publish(l *ledger, lock *managerlock.HomeLock, engine *transaction.Engine, files map[string][]byte, removals []string) error {
	raw, err := canonical(l)
	if err != nil {
		return err
	}
	if len(raw) > maxRecord {
		return fmt.Errorf("project lease state exceeds record limit")
	}
	files[statePath(m.home)] = raw
	if err = checkedRoute(m.home, statePath(m.home)); err != nil {
		return err
	}
	if err = privatedir.MakeAll(filepath.Dir(statePath(m.home))); err != nil {
		return err
	}
	plan := transaction.Plan{TransactionID: "project-lease-" + nonce(), ProjectIdentity: "project-leases"}
	paths := make([]string, 0, len(files)+len(removals))
	for p := range files {
		paths = append(paths, p)
	}
	paths = append(paths, removals...)
	sort.Strings(paths)
	temps := []string{}
	defer func() {
		for _, p := range temps {
			_ = os.Remove(p)
		}
	}()
	for i, path := range paths {
		if i > 0 && paths[i-1] == path {
			return fmt.Errorf("duplicate project lease transaction target")
		}
		if err = checkedRoute(m.home, path); err != nil {
			return err
		}
		if err = privatedir.MakeAll(filepath.Dir(path)); err != nil {
			return err
		}
		pre, er := transaction.DigestTarget(transaction.KindBytes, path)
		if er != nil {
			return er
		}
		t := transaction.Target{Class: "project-lease", Identifier: path, LivePath: path, PreimageDigest: pre}
		if payload, ok := files[path]; ok {
			f, er := privatedir.CreateTemp(filepath.Dir(statePath(m.home)), "stage-*")
			if er != nil {
				return er
			}
			temps = append(temps, f.Name())
			if strings.HasPrefix(path, filepath.Join(m.home, "environments", "project-views")+string(filepath.Separator)) {
				for _, v := range l.Views {
					if rel, e := filepath.Rel(filepath.Join(m.home, homeRelative(v.Composition)), path); e == nil {
						if surface, ok := v.Surfaces[filepath.ToSlash(rel)]; ok {
							if e = f.Chmod(os.FileMode(surface.Mode)); e != nil {
								_ = f.Close()
								return e
							}
						}
					}
				}
			}
			_, er = f.Write(payload)
			er = errors.Join(er, f.Close())
			if er != nil {
				return er
			}
			t.StagedSource = f.Name()
		}
		plan.Targets = append(plan.Targets, t)
	}
	if err = lock.AssertCurrent(); err != nil {
		return err
	}
	if _, err = engine.Prepare(lock, plan); err != nil {
		return err
	}
	return engine.Commit(lock, plan.TransactionID)
}
