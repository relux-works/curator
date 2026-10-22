package projectlease

import (
	"context"
	"errors"
	"fmt"
	"github.com/relux-works/curator/internal/privatedir"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/transaction"
)

func (m *Manager) resolve(ctx context.Context, held managerlock.Held, ref, sha, owner string) (Composition, error) {
	if !tagged(ref, "composition") || !hexPattern.MatchString(sha) || !tagged(owner, "principal") {
		return Composition{}, fmt.Errorf("invalid project composition or principal identity")
	}
	if err := held.Validate(m.home); err != nil {
		return Composition{}, err
	}
	if m.authority == nil {
		return Composition{}, fmt.Errorf("project composition authority unavailable")
	}
	c, err := m.authority.Resolve(ctx, held, ref, sha)
	if err != nil {
		return c, err
	}
	if err = c.validate(); err != nil {
		return c, err
	}
	if c.Ref != ref || c.SHA256 != sha {
		return c, fmt.Errorf("project composition authority returned a different identity")
	}
	if err = held.Validate(m.home); err != nil {
		return c, err
	}
	if err = m.authority.Authorize(ctx, held, c, owner); err != nil {
		return c, err
	}
	if err = held.Validate(m.home); err != nil {
		return c, err
	}
	return c, nil
}
func (m *Manager) observe(ctx context.Context, b Binding) (Evidence, error) {
	if m.sessions == nil {
		return Evidence{}, fmt.Errorf("tracked-session authority unavailable")
	}
	e, err := m.sessions.Observe(ctx, b.SessionRef, b.Generation)
	if err != nil {
		return e, err
	}
	if e.SessionRef != b.SessionRef || e.Generation != b.Generation || e.OwnerRef != b.OwnerRef || len(e.ProcessGeneration) > 256 || (e.Running && e.Terminated) {
		return e, fmt.Errorf("tracked-session generation or owner mismatch")
	}
	return e, nil
}
func busy(l *ledger, key string, exceptPrepared string) error {
	sessions := []string{}
	for _, e := range l.Leases {
		if homeRelative(e.Composition) == key && e.State != Released {
			if exceptPrepared != "" && e.State == Prepared && e.Binding.ContextSHA256 == exceptPrepared {
				continue
			}
			sessions = append(sessions, e.Binding.SessionRef)
		}
	}
	if len(sessions) == 0 {
		return nil
	}
	sort.Strings(sessions)
	return fmt.Errorf("%s: sessions %s retain this home; wait or end the owning sessions", Busy, strings.Join(sessions, ", "))
}
func consistentHome(l *ledger, c Composition) error {
	// A protected reference must never alias a different stable allocation.
	check := func(other Composition) error {
		if (foldPath(homeRelative(c)) == foldPath(homeRelative(other))) != (c.HomeRef == other.HomeRef) || (c.HomeRef == other.HomeRef && homeRelative(c) != homeRelative(other)) {
			return fmt.Errorf("project home allocation identity mismatch")
		}
		return nil
	}
	for _, e := range l.Leases {
		if err := check(e.Composition); err != nil {
			return err
		}
	}
	for _, v := range l.Views {
		if err := check(v.Composition); err != nil {
			return err
		}
	}
	return nil
}

// Acquire reserves a prepared lease before any fragment can be published. It
// requires a reserved tracked session; PID/TTL/caller lifetime are not inputs.
func (m *Manager) Acquire(ctx context.Context, r Request) (Binding, error) {
	var result Binding
	if !idPattern.MatchString(r.LeaseID) || !tagged(r.ContextRef, "composition") || !hexPattern.MatchString(r.ContextSHA256) || !tagged(r.OwnerRef, "principal") || !tagged(r.SessionRef, "session") || r.Generation < 1 || r.Generation > 9007199254740991 {
		return result, fmt.Errorf("invalid lease request")
	}
	err := m.locked(ctx, func(l *ledger, lock *managerlock.HomeLock, engine *transaction.Engine) error {
		return m.acquireUnderLock(ctx, r, l, lock, engine, &result)
	})
	return result, err
}

func (m *Manager) acquireUnderLock(ctx context.Context, r Request, l *ledger, lock *managerlock.HomeLock, engine *transaction.Engine, result *Binding) error {
	if !idPattern.MatchString(r.LeaseID) {
		return fmt.Errorf("invalid lease ID")
	}
	if _, ok := l.Leases[r.LeaseID]; ok {
		return fmt.Errorf("lease ID already allocated")
	}
	c, err := m.resolve(ctx, lock.Held(), r.ContextRef, r.ContextSHA256, r.OwnerRef)
	if err != nil {
		return err
	}
	if err = consistentHome(l, c); err != nil {
		return err
	}
	b := Binding{"project-lease/1", r.LeaseID, c.Ref, c.SHA256, c.CheckoutID, c.Environment, c.HomeRef, r.OwnerRef, r.SessionRef, r.Generation}
	if err = b.validate(); err != nil {
		return err
	}
	evidence, err := m.observe(ctx, b)
	if err != nil {
		return err
	}
	if !evidence.NoChildStarted || evidence.Terminated {
		return fmt.Errorf("session is not reserved before provider start")
	}
	for _, e := range l.Leases {
		if e.State == Released {
			continue
		}
		if e.Binding.SessionRef == b.SessionRef && e.Binding.Generation == b.Generation {
			return fmt.Errorf("session generation already leased")
		}
		if homeRelative(e.Composition) == homeRelative(c) && (!c.ConcurrentState || !e.Composition.ConcurrentState || e.Binding.ContextRef != b.ContextRef || e.Binding.ContextSHA256 != b.ContextSHA256) {
			return fmt.Errorf("%s: %s owns the home; wait or end the owning session", Busy, e.Binding.SessionRef)
		}
	}
	l.Leases[b.LeaseID] = entry{Binding: b, Composition: c, State: Prepared}
	raw, err := canonical(b)
	if err != nil {
		return err
	}
	if _, err = os.Lstat(bindingPath(m.home, b.LeaseID)); !os.IsNotExist(err) {
		return fmt.Errorf("immutable lease allocation already exists or is unreadable")
	}
	if err = m.publish(l, lock, engine, map[string][]byte{bindingPath(m.home, b.LeaseID): raw}, nil); err != nil {
		return err
	}
	*result = b
	return nil
}

// PrepareResume reserves a new session generation for a previously released
// lease's ORIGINAL composition. It never selects the current generation or
// falls back to a fresh launch. PublishPrepared and Transfer still precede start.
func (m *Manager) PrepareResume(ctx context.Context, previous string, r Request) (Binding, error) {
	var result Binding
	err := m.locked(ctx, func(l *ledger, lock *managerlock.HomeLock, engine *transaction.Engine) error {
		old, ok := l.Leases[previous]
		if !ok || old.State != Released {
			return fmt.Errorf("resume reservation requires a released reconciled original lease")
		}
		if (r.ContextRef != "" && r.ContextRef != old.Binding.ContextRef) || (r.ContextSHA256 != "" && r.ContextSHA256 != old.Binding.ContextSHA256) {
			return fmt.Errorf("resume cannot change the recorded composition")
		}
		r.ContextRef = old.Binding.ContextRef
		r.ContextSHA256 = old.Binding.ContextSHA256
		if !tagged(r.OwnerRef, "principal") || !tagged(r.SessionRef, "session") || r.Generation < 1 || r.Generation > 9007199254740991 {
			return fmt.Errorf("invalid resume reservation")
		}
		return m.acquireUnderLock(ctx, r, l, lock, engine, &result)
	})
	return result, err
}

func safeSurface(path string) bool {
	return identifiers.PortablePath(path) && !strings.ContainsAny(path, "*?[]") && len(path) <= 240 && len(strings.Split(path, "/")) <= 32
}
func inventory(files map[string][]byte) map[string]Surface {
	result := map[string]Surface{}
	for path, data := range files {
		result[path] = Surface{SHA256: digest(data), Size: int64(len(data)), Mode: 0600}
	}
	return result
}
func (m *Manager) verifyView(c Composition, v view) error {
	if v.Composition.Ref != c.Ref || v.Composition.SHA256 != c.SHA256 || !reflect.DeepEqual(v.Surfaces, c.Generated) {
		return fmt.Errorf("project home does not materialize the leased composition")
	}
	for path, surface := range v.Surfaces {
		full := filepath.Join(m.home, homeRelative(c), filepath.FromSlash(path))
		raw, err := read(m.home, full)
		if err != nil {
			return err
		}
		if digest(raw) != surface.SHA256 || int64(len(raw)) != surface.Size {
			return fmt.Errorf("project generated surface drift: %s", path)
		}
		info, err := os.Lstat(full)
		if err != nil {
			return err
		}
		if err = privatedir.ValidateFile(full); err != nil {
			return err
		}
		if runtime.GOOS != "windows" && uint32(info.Mode().Perm()) != surface.Mode {
			return fmt.Errorf("project generated surface mode drift: %s", path)
		}
	}
	return nil
}

// PublishPrepared transactionally replaces only inventoried generated files.
// Native login/state files are neither copied nor replaced. The protected
// composition receipt, supplied by Authority, must match the complete output.
// Identical already-published bytes may be shared without writing beneath an
// active session; changed bytes always wait/refuse until all leases end.
func (m *Manager) PublishPrepared(ctx context.Context, id string, files map[string][]byte, wait bool) error {
	return m.publishView(ctx, id, "", "", "", files, wait)
}

// Update is the repair/profile/skill/CLI synchronization gate for a project
// home. All such callers must use this operation; legacy profile homes never
// route here. It refuses even an identical repair while the home is leased.
func (m *Manager) Update(ctx context.Context, ref, sha, owner string, files map[string][]byte, wait bool) error {
	return m.publishView(ctx, "", ref, sha, owner, files, wait)
}
func (m *Manager) publishView(ctx context.Context, id, ref, sha, owner string, files map[string][]byte, wait bool) error {
	var lastBusy error
	for {
		if ctx.Err() != nil && lastBusy != nil {
			return errors.Join(lastBusy, ctx.Err())
		}
		err := m.locked(ctx, func(l *ledger, lock *managerlock.HomeLock, engine *transaction.Engine) error {
			var c Composition
			var err error
			if id != "" {
				e, ok := l.Leases[id]
				if !ok || e.State != Prepared {
					return fmt.Errorf("prepared lease required before publication")
				}
				c, err = m.resolve(ctx, lock.Held(), e.Binding.ContextRef, e.Binding.ContextSHA256, e.Binding.OwnerRef)
				if err == nil && !reflect.DeepEqual(c, e.Composition) {
					err = fmt.Errorf("immutable composition projection changed")
				}
			} else {
				c, err = m.resolve(ctx, lock.Held(), ref, sha, owner)
			}
			if err != nil {
				return err
			}
			if err = consistentHome(l, c); err != nil {
				return err
			}
			expected := inventory(files)
			for path, file := range expected {
				if approved, ok := c.Generated[path]; ok {
					file.Mode = approved.Mode
					expected[path] = file
				}
			}
			if !reflect.DeepEqual(expected, c.Generated) {
				return fmt.Errorf("generated files do not match protected composition receipt")
			}
			key := homeRelative(c)
			old, exists := l.Views[key]
			if id != "" && exists && old.Composition.Ref == c.Ref && old.Composition.SHA256 == c.SHA256 && reflect.DeepEqual(old.Surfaces, expected) {
				return m.verifyView(c, old)
			}
			except := ""
			if id != "" {
				except = c.SHA256
			}
			if err = busy(l, key, except); err != nil {
				return err
			}
			writes := map[string][]byte{}
			removals := []string{}
			if exists {
				if err = m.verifyView(old.Composition, old); err != nil {
					return err
				}
				for path := range old.Surfaces {
					if _, ok := files[path]; !ok {
						removals = append(removals, filepath.Join(m.home, key, filepath.FromSlash(path)))
					}
				}
			}
			for path, data := range files {
				if !safeSurface(path) {
					return fmt.Errorf("invalid generated project surface")
				}
				target := filepath.Join(m.home, key, filepath.FromSlash(path))
				_, recorded := old.Surfaces[path]
				if !recorded {
					if _, err = os.Lstat(target); !os.IsNotExist(err) {
						return fmt.Errorf("unmanaged native state blocks generated surface: %s", path)
					}
				}
				writes[target] = append([]byte(nil), data...)
			}
			l.Views[key] = view{Composition: c, Surfaces: expected}
			return m.publish(l, lock, engine, writes, removals)
		})
		if err == nil || !wait || !strings.Contains(err.Error(), Busy) {
			if err != nil && ctx.Err() != nil && lastBusy != nil {
				return errors.Join(lastBusy, err)
			}
			return err
		}
		lastBusy = err
		timer := time.NewTimer(20 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return errors.Join(err, ctx.Err())
		case <-timer.C:
		}
	}
}

// Transfer persists runner ownership BEFORE provider start. A launching process
// may exit afterwards; neither detach nor descendants call a release path.
func (m *Manager) Transfer(ctx context.Context, id string) error {
	return m.transition(ctx, id, func(l *ledger, e *entry, lock *managerlock.HomeLock) error {
		if e.State != Prepared {
			return fmt.Errorf("transfer requires a prepared lease")
		}
		c, err := m.resolve(ctx, lock.Held(), e.Binding.ContextRef, e.Binding.ContextSHA256, e.Binding.OwnerRef)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(c, e.Composition) {
			return fmt.Errorf("immutable composition projection changed")
		}
		if err = m.verifyView(c, l.Views[homeRelative(c)]); err != nil {
			return err
		}
		observation, err := m.observe(ctx, e.Binding)
		if err != nil {
			return err
		}
		if !observation.Running || observation.ProcessGeneration == "" || !observation.NoChildStarted {
			return fmt.Errorf("reserved runner identity required before provider start")
		}
		e.State = Active
		e.ProcessGeneration = observation.ProcessGeneration
		return nil
	})
}
func (m *Manager) transition(ctx context.Context, id string, fn func(*ledger, *entry, *managerlock.HomeLock) error) error {
	return m.locked(ctx, func(l *ledger, lock *managerlock.HomeLock, engine *transaction.Engine) error {
		e, ok := l.Leases[id]
		if !ok {
			return fmt.Errorf("unknown project lease")
		}
		if err := fn(l, &e, lock); err != nil {
			return err
		}
		l.Leases[id] = e
		return m.publish(l, lock, engine, map[string][]byte{}, nil)
	})
}

// Reconcile is the ONLY release path. Unknown evidence retains roots. Active
// or recovered sessions require the recorded process incarnation plus proof
// that the complete session and descendants ended and authority reconciled.
func (m *Manager) Reconcile(ctx context.Context, id string) error {
	return m.transition(ctx, id, func(_ *ledger, e *entry, _ *managerlock.HomeLock) error {
		if e.State == Released {
			return nil
		}
		o, err := m.observe(ctx, e.Binding)
		if err != nil {
			e.State = Unknown
			return nil
		}
		if e.ProcessGeneration == "" {
			if o.NoChildStarted && !o.Running && o.OwnershipReconciled {
				e.State = Released
			} else {
				e.State = Unknown
			}
		} else if o.ProcessGeneration != e.ProcessGeneration {
			e.State = Unknown
		} else if o.Terminated && o.OwnershipReconciled && o.DescendantsTerminated {
			e.State = Released
		} else if o.Running {
			e.State = Active
		} else {
			e.State = Unknown
		}
		return nil
	})
}

// Recover treats every interrupted live lease as unknown. Reconciliation is
// explicit; recovery never infers death from PID reuse or elapsed time.
func (m *Manager) Recover(ctx context.Context) error {
	return m.locked(ctx, func(l *ledger, lock *managerlock.HomeLock, engine *transaction.Engine) error {
		for id, e := range l.Leases {
			if e.State != Released {
				e.State = Unknown
				l.Leases[id] = e
			}
		}
		return m.publish(l, lock, engine, map[string][]byte{}, nil)
	})
}

// Resume returns the original binding after current authority and exact
// session/process reconciliation. Failure never acquires a fresh composition.
func (m *Manager) Resume(ctx context.Context, id string) (Binding, error) {
	var result Binding
	err := m.locked(ctx, func(l *ledger, lock *managerlock.HomeLock, _ *transaction.Engine) error {
		e, ok := l.Leases[id]
		if !ok || e.State != Active {
			return fmt.Errorf("resume requires a reconciled active lease")
		}
		c, err := m.resolve(ctx, lock.Held(), e.Binding.ContextRef, e.Binding.ContextSHA256, e.Binding.OwnerRef)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(c, e.Composition) {
			return fmt.Errorf("immutable composition projection changed")
		}
		o, err := m.observe(ctx, e.Binding)
		if err != nil {
			return err
		}
		if !o.Running || o.ProcessGeneration == "" || o.ProcessGeneration != e.ProcessGeneration {
			return fmt.Errorf("resume session/process incarnation is not current")
		}
		if err = m.verifyView(c, l.Views[homeRelative(c)]); err != nil {
			return err
		}
		result = e.Binding
		return nil
	})
	return result, err
}

// Status performs no lock creation, journal recovery or state writes.
func (m *Manager) Status(ctx context.Context) ([]Status, error) {
	lm, err := managerlock.New(m.home)
	if err != nil {
		return nil, err
	}
	lock, err := lm.AcquireExistingHomeOnly(ctx)
	if err != nil {
		if !HasState(m.home) && errors.Is(err, os.ErrNotExist) {
			return []Status{}, nil
		}
		return nil, err
	}
	defer func() { _ = lock.Close() }()
	engine, err := transaction.New(m.home)
	if err != nil {
		return nil, err
	}
	pending, err := engine.PendingRecovery(lock)
	if err != nil {
		return nil, err
	}
	l, err := load(m.home)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for id := range l.Leases {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := []Status{}
	for _, id := range ids {
		e := l.Leases[id]
		s := Status{Binding: e.Binding, State: e.State, Busy: e.State != Released, Home: filepath.Join(m.home, homeRelative(e.Composition)), Liveness: "unknown"}
		if pending {
			s.State = Unknown
			s.Busy = true
		} else if e.State == Released {
			s.Liveness = "released"
		} else if o, er := m.observe(ctx, e.Binding); er == nil && o.ProcessGeneration == e.ProcessGeneration && o.Running && e.ProcessGeneration != "" {
			s.Liveness = "running"
		}
		if m.authority == nil {
			s.AuthorityUnknown = true
		} else if resolved, er := m.resolve(ctx, lock.Held(), e.Binding.ContextRef, e.Binding.ContextSHA256, e.Binding.OwnerRef); er != nil {
			s.Revoked = errors.Is(er, ErrRevoked)
			s.AuthorityUnknown = !s.Revoked
		} else if !reflect.DeepEqual(resolved, e.Composition) {
			s.AuthorityUnknown = true
		}
		result = append(result, s)
	}
	return result, nil
}

// ReferencesUnderLock is the GC boundary. A corrupt or unreadable ledger is an
// error, never an empty mark set. No observation can release a reference here.
func ReferencesUnderLock(home string, lock transaction.HomeLock) (Roots, error) {
	if lock == nil {
		return Roots{}, fmt.Errorf("project reference read requires manager lock")
	}
	if err := lock.AssertHeld(); err != nil {
		return Roots{}, err
	}
	if !HasState(home) {
		return Roots{}, nil
	}
	l, err := load(home)
	if err != nil {
		return Roots{}, err
	}
	r := Roots{}
	homes := map[string]bool{}
	comps := map[string]bool{}
	builds := map[string]bool{}
	runtime := map[string]bool{}
	add := func(c Composition) {
		homes[filepath.Join(home, homeRelative(c))] = true
		comps[c.Ref] = true
		for _, k := range c.BuildKeys {
			builds[k] = true
		}
		for _, k := range c.RuntimeEntries {
			runtime[k] = true
		}
	}
	for _, e := range l.Leases {
		add(e.Composition)
	}
	for _, v := range l.Views {
		add(v.Composition)
	}
	for k := range homes {
		r.Homes = append(r.Homes, k)
	}
	for k := range comps {
		r.Compositions = append(r.Compositions, k)
	}
	for k := range builds {
		r.BuildKeys = append(r.BuildKeys, k)
	}
	for k := range runtime {
		r.RuntimeEntries = append(r.RuntimeEntries, k)
	}
	sort.Strings(r.Homes)
	sort.Strings(r.Compositions)
	sort.Strings(r.BuildKeys)
	sort.Strings(r.RuntimeEntries)
	return r, nil
}

// HasState guards the historical unlocked runtime collector. Uncertain state
// is treated as present; project-aware callers must use locked collection.
func HasState(home string) bool {
	for _, p := range []string{filepath.Dir(statePath(home)), filepath.Join(home, "environments", "project-views")} {
		_, err := os.Lstat(p)
		if !errors.Is(err, os.ErrNotExist) {
			return true
		}
	}
	return false
}
