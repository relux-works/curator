// Package persistence supplies the shared project/home lease journal publisher.
// Namespace schemas, immutable identity and lifecycle authority stay with the
// caller; this primitive creates no lock or second journal protocol.
package persistence

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/transaction"
)

type Publication struct {
	Home, StageDir, TransactionID, Owner, Class string
	Files                                       map[string][]byte
	Removals                                    []string
	PrivateFile                                 bool
	Route                                       func(string, string) error
	Immutable                                   func(string) bool
	Mode                                        func(string) os.FileMode
}

// Publish retains and validates the caller's acquisition through staging and
// the existing engine's atomic forward/rollback lifecycle. Every route and
// immutable preimage is checked before creating staging directories or files.
func Publish(p Publication, lock *managerlock.HomeLock, engine *transaction.Engine) (retErr error) {
	if err := lock.Held().Validate(p.Home); err != nil {
		return err
	}
	paths := append([]string{}, p.Removals...)
	for path := range p.Files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	preimages := map[string]string{}
	for i, path := range paths {
		if i > 0 && paths[i-1] == path {
			return fmt.Errorf("duplicate lease transaction target")
		}
		if err := p.Route(p.Home, path); err != nil {
			return err
		}
		pre, err := transaction.DigestTarget(transaction.KindBytes, path)
		if err != nil {
			return err
		}
		if p.Immutable != nil && p.Immutable(path) && pre != transaction.DigestAbsent {
			return fmt.Errorf("immutable lease evidence already exists")
		}
		preimages[path] = pre
	}
	if err := p.Route(p.Home, p.StageDir); err != nil {
		return err
	}
	directories := publicationDirectories{root: p.StageDir}
	var temps []string
	defer func() {
		for _, path := range temps {
			_ = os.Remove(path)
		}
		if retErr == nil || len(directories.created) == 0 {
			return
		}
		// A pending or unreadable journal owns its directories until explicit
		// recovery. Only a completed rollback permits restoring prior absence.
		pending, err := engine.PendingRecovery(lock)
		if err != nil {
			retErr = errors.Join(retErr, err)
			return
		}
		if !pending {
			retErr = errors.Join(retErr, directories.removeEmpty(p, lock))
		}
	}()
	if err := directories.makeAll(p.StageDir); err != nil {
		return err
	}
	plan := transaction.Plan{TransactionID: p.TransactionID, ProjectIdentity: p.Owner}
	for _, path := range paths {
		if err := lock.Held().Validate(p.Home); err != nil {
			return err
		}
		if err := directories.makeAll(filepath.Dir(path)); err != nil {
			return err
		}
		target := transaction.Target{Class: p.Class, Identifier: path, LivePath: path, PreimageDigest: preimages[path], PrivateFile: p.PrivateFile}
		if payload, ok := p.Files[path]; ok {
			f, err := privatedir.CreateTemp(p.StageDir, "stage-*")
			if err != nil {
				return err
			}
			temps = append(temps, f.Name())
			if p.Mode != nil {
				if mode := p.Mode(path); mode != 0 {
					err = f.Chmod(mode)
				}
			}
			if err == nil {
				_, err = f.Write(payload)
			}
			err = errors.Join(err, f.Close())
			if err != nil {
				return err
			}
			target.StagedSource = f.Name()
		}
		plan.Targets = append(plan.Targets, target)
	}
	if err := lock.AssertCurrent(); err != nil {
		return err
	}
	if _, err := engine.Prepare(lock, plan); err != nil {
		return err
	}
	return engine.Commit(lock, plan.TransactionID)
}

// Track only directories this publication exclusively created inside its state
// namespace. Other target parents and pre-existing directories retain their
// prior lifecycle. No cleanup recursively removes content or repairs drift.
type publicationDirectory struct {
	path     string
	identity os.FileInfo
}
type publicationDirectories struct {
	root    string
	created []publicationDirectory
}

func (d *publicationDirectories) makeAll(path string) error {
	rel, err := filepath.Rel(d.root, path)
	if err != nil {
		return err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return privatedir.MakeAll(path)
	}
	info, err := os.Lstat(path)
	if err == nil {
		// The caller already admitted this route. Preserve existing directory
		// permissions; only directories created here acquire a private mode.
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("publication parent is not a real directory")
		}
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := d.makeAll(filepath.Dir(path)); err != nil {
		return err
	}
	if err := privatedir.Make(path); err != nil {
		return err
	}
	info, err = os.Lstat(path)
	if err != nil {
		return err
	}
	// Complete Windows' deferred file identity while the created name is current.
	if !os.SameFile(info, info) {
		return fmt.Errorf("publication directory identity unavailable")
	}
	d.created = append(d.created, publicationDirectory{path, info})
	return nil
}
func (d *publicationDirectories) removeEmpty(p Publication, lock *managerlock.HomeLock) error {
	for i := len(d.created) - 1; i >= 0; i-- {
		dir := d.created[i]
		if err := lock.Held().Validate(p.Home); err != nil {
			return err
		}
		if err := p.Route(p.Home, dir.path); err != nil {
			return err
		}
		current, err := os.Lstat(dir.path)
		if err != nil {
			return err
		}
		if !current.IsDir() || !os.SameFile(dir.identity, current) || current.Mode() != dir.identity.Mode() {
			return fmt.Errorf("publication directory identity changed")
		}
		if err := privatedir.Validate(dir.path); err != nil {
			return err
		}
		// Remove only an empty directory; unknown bytes always survive refusal.
		if err := os.Remove(dir.path); err != nil {
			return err
		}
	}
	return nil
}
