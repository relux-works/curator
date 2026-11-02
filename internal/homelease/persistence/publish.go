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
func Publish(p Publication, lock *managerlock.HomeLock, engine *transaction.Engine) error {
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
	if err := privatedir.MakeAll(p.StageDir); err != nil {
		return err
	}
	plan := transaction.Plan{TransactionID: p.TransactionID, ProjectIdentity: p.Owner}
	var temps []string
	defer func() {
		for _, path := range temps {
			_ = os.Remove(path)
		}
	}()
	for _, path := range paths {
		if err := lock.Held().Validate(p.Home); err != nil {
			return err
		}
		if err := privatedir.MakeAll(filepath.Dir(path)); err != nil {
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
