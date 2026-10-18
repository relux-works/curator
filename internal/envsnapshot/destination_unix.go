//go:build darwin || linux

package envsnapshot

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

type retainedDirectory struct {
	file *os.File
	info os.FileInfo
	name string
}

// Destination retains every supplied ancestor, protected root and target until
// Close. It performs no writes and exposes no pathname or raw file descriptor.
// Revalidate is an observation, not an atomic publication primitive or a lease
// against another process with the same user's authority. Do not publish with
// os.Rename after calling it: publication needs its own no-replace backend.
type Destination struct {
	mu        sync.Mutex
	chain     []retainedDirectory
	protected [][]retainedDirectory
	lineage   []*os.File
	target    *os.File
	name      string
	kind      DestinationKind
	prior     []byte
	closed    bool
	// Per-destination fault/race seam; production leaves this nil.
	publicationHook func(string, *os.File) error
}

// OpenDestination validates without writes, retaining no-follow handles for
// every component as supplied (including components before '..'). The final
// parent must exist and belong to the caller. ProtectedRoots must be complete
// and nonempty; this API does not discover machine configuration or authorize
// source capture. An intact prior snapshot is fully verified before admission.
func OpenDestination(path string, policy DestinationPolicy) (_ *Destination, err error) {
	d := &Destination{}
	defer func() {
		if err != nil {
			_ = d.Close()
		}
	}()
	if len(policy.ProtectedRoots) == 0 {
		return nil, invalidDestination("protected-roots-required")
	}
	parts, err := destinationParts(path)
	if err != nil || len(parts) == 0 || parts[len(parts)-1] == "." || parts[len(parts)-1] == ".." {
		return nil, invalidDestination("path-invalid")
	}
	d.name = parts[len(parts)-1]
	d.chain, err = retainDirectories(parts[:len(parts)-1])
	if err != nil {
		return nil, err
	}
	parent := d.chain[len(d.chain)-1].file
	if err = destinationDirectory(parent, true, true); err != nil {
		return nil, err
	}
	for _, root := range policy.ProtectedRoots {
		rootParts, e := destinationParts(root)
		if e != nil {
			return nil, invalidDestination("protected-root-unreadable")
		}
		chain, e := retainDirectories(rootParts)
		if e != nil {
			return nil, invalidDestination("protected-root-unreadable")
		}
		d.protected = append(d.protected, chain)
	}
	d.target, err = openMember(parent, d.name, true)
	if err != nil && !errors.Is(err, syscall.ENOENT) {
		return nil, invalidDestination("target-unreadable")
	}
	if d.target != nil {
		if err = destinationDirectory(d.target, true, true); err != nil {
			return nil, err
		}
	}
	d.lineage, err = retainLineage(parent)
	if err != nil {
		return nil, invalidDestination("ancestry-unreadable")
	}
	if err = d.excludeProtected(); err != nil {
		return nil, err
	}
	d.kind, d.prior, err = inspectDestination(d.target)
	if err != nil {
		return nil, err
	}
	if err = d.revalidate(); err != nil {
		return nil, err
	}
	return d, nil
}

// Kind returns the admitted observation, including after Close.
func (d *Destination) Kind() DestinationKind { return d.kind }

// Revalidate rejects changed ancestor/target entries, permissions, protected
// roots and prior contents through retained handles. It never writes or repairs.
func (d *Destination) Revalidate() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.revalidate()
}

// MatchPrior verifies the actual tree again and compares exact manifest bytes.
// Portable equivalence is deliberately insufficient for idempotence. For an
// absent target this returns false, without granting publication rights.
func (d *Destination) MatchPrior(manifest []byte) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if err := d.revalidate(); err != nil {
		return false, err
	}
	if d.kind != DestinationSnapshot {
		return false, nil
	}
	if !bytes.Equal(d.prior, manifest) {
		return false, refuse("generation-mismatch")
	}
	return true, nil
}

func (d *Destination) revalidate() error {
	if d.closed || len(d.chain) == 0 {
		return invalidDestination("destination-closed")
	}
	if err := recheckDirectories(d.chain); err != nil {
		return err
	}
	for _, chain := range d.protected {
		if err := recheckDirectories(chain); err != nil {
			return invalidDestination("protected-root-changed")
		}
	}
	parent := d.chain[len(d.chain)-1].file
	if err := destinationDirectory(parent, true, true); err != nil {
		return err
	}
	fresh, err := openMember(parent, d.name, true)
	if d.target == nil {
		if fresh != nil {
			_ = fresh.Close()
		}
		if !errors.Is(err, syscall.ENOENT) {
			return invalidDestination("target-changed")
		}
		return nil
	}
	if err != nil {
		return invalidDestination("target-changed")
	}
	defer func() { _ = fresh.Close() }()
	a, aerr := d.target.Stat()
	b, berr := fresh.Stat()
	if aerr != nil || berr != nil || !os.SameFile(a, b) {
		return invalidDestination("target-changed")
	}
	if err := destinationDirectory(fresh, true, true); err != nil {
		return err
	}
	kind, prior, err := inspectDestination(fresh)
	if err != nil {
		var refusal *Error
		if d.kind == DestinationSnapshot && errors.As(err, &refusal) && refusal.Code == DestinationInvalid {
			return refuse("prior-unreadable")
		}
		return err
	}
	if kind != d.kind || !bytes.Equal(prior, d.prior) {
		return refuse("generation-mismatch")
	}
	return nil
}

func inspectDestination(target *os.File) (DestinationKind, []byte, error) {
	if target == nil {
		return DestinationAbsent, nil, nil
	}
	// A fresh open file description avoids sharing ReadDir's cursor with a
	// previous inspection or SNAP-ID traversal.
	fresh, err := openMember(target, ".", true)
	if err != nil {
		return "", nil, invalidDestination("target-unreadable")
	}
	defer func() { _ = fresh.Close() }()
	entries, err := fresh.ReadDir(-1)
	if err != nil {
		return "", nil, invalidDestination("target-unreadable")
	}
	if len(entries) == 0 {
		return "", nil, invalidDestination("empty-target")
	}
	hasManifest := false
	for _, entry := range entries {
		if entry.Name() == "manifest.json" {
			hasManifest = true
		}
	}
	if !hasManifest {
		return "", nil, invalidDestination("unrelated-target")
	}
	verifyHandle, err := openMember(target, ".", true)
	if err != nil {
		return "", nil, refuse("tree-unreadable")
	}
	defer func() { _ = verifyHandle.Close() }()
	v, err := verifyDirectory(verifyHandle, "")
	if err != nil {
		return "", nil, err
	}
	if err := priorOwnershipACL(target); err != nil {
		return "", nil, refuse("snapshot-owner-acl")
	}
	return DestinationSnapshot, v.ManifestBytes(), nil
}

// SNAP-ID covers bytes, modes and the complete entry set. Admission additionally
// checks each prior object's owner/ACL through fresh no-follow handles.
func priorOwnershipACL(dir *os.File) error {
	fresh, err := openMember(dir, ".", true)
	if err != nil {
		return err
	}
	defer func() { _ = fresh.Close() }()
	entries, err := fresh.ReadDir(-1)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		f, err := openMember(dir, entry.Name(), entry.IsDir())
		if err != nil {
			return err
		}
		info, err := f.Stat()
		if err == nil {
			stat, ok := info.Sys().(*syscall.Stat_t)
			if !ok || int64(stat.Uid) != int64(os.Geteuid()) {
				err = os.ErrPermission
			}
		}
		if err == nil {
			err = destinationACL(f)
		}
		if err == nil && info.IsDir() {
			err = priorOwnershipACL(f)
		}
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func destinationParts(path string) ([]string, error) {
	if path == "" || strings.ContainsRune(path, 0) {
		return nil, os.ErrInvalid
	}
	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		path = cwd + "/" + path
	}
	var parts []string
	for _, part := range strings.Split(path, "/") {
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts, nil
}

func retainDirectories(parts []string) (_ []retainedDirectory, err error) {
	root, err := os.Open("/")
	if err != nil {
		return nil, invalidDestination("ancestor-unreadable")
	}
	chain := []retainedDirectory{{file: root}}
	defer func() {
		if err != nil {
			closeDirectories(chain)
		}
	}()
	for i := 0; ; i++ {
		current := &chain[len(chain)-1]
		if err = destinationDirectory(current.file, false, false); err != nil {
			return nil, err
		}
		current.info, err = current.file.Stat()
		if err != nil {
			return nil, invalidDestination("ancestor-unreadable")
		}
		if i == len(parts) {
			return chain, nil
		}
		child, e := openMember(current.file, parts[i], true)
		if e != nil {
			return nil, invalidDestination("ancestor-unreadable")
		}
		chain = append(chain, retainedDirectory{file: child, name: parts[i]})
	}
}

func recheckDirectories(chain []retainedDirectory) error {
	for i, retained := range chain {
		if err := destinationDirectory(retained.file, false, false); err != nil {
			return err
		}
		if i == 0 {
			continue
		}
		fresh, err := openMember(chain[i-1].file, retained.name, true)
		if err != nil {
			return invalidDestination("ancestor-changed")
		}
		info, err := fresh.Stat()
		closeErr := fresh.Close()
		if err != nil || closeErr != nil || !os.SameFile(retained.info, info) {
			return invalidDestination("ancestor-changed")
		}
	}
	return nil
}

func destinationDirectory(f *os.File, caller, private bool) error {
	info, err := f.Stat()
	if err != nil || !info.IsDir() {
		return invalidDestination("directory-unreadable")
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || (int64(stat.Uid) != int64(os.Geteuid()) && (caller || stat.Uid != 0)) {
		return invalidDestination("directory-owner")
	}
	if info.Mode().Perm()&0022 != 0 || info.Mode()&(os.ModeSetuid|os.ModeSetgid|os.ModeSticky) != 0 || (private && info.Mode().Perm() != 0700) {
		return invalidDestination("directory-permissions")
	}
	if err := destinationACL(f); err != nil {
		return invalidDestination("directory-acl")
	}
	return nil
}

// retainLineage follows physical parents from the retained handle. It does not
// mistake a spelling visited before '..' for a physical ancestor of the target.
func retainLineage(start *os.File) (_ []*os.File, err error) {
	current, err := openMember(start, ".", true)
	if err != nil {
		return nil, err
	}
	lineage := []*os.File{current}
	defer func() {
		if err != nil {
			for _, f := range lineage {
				_ = f.Close()
			}
		}
	}()
	for len(lineage) < 4096 {
		parent, e := openMember(current, "..", true)
		if e != nil {
			return nil, e
		}
		a, ae := current.Stat()
		b, be := parent.Stat()
		if ae != nil || be != nil {
			_ = parent.Close()
			return nil, os.ErrInvalid
		}
		if os.SameFile(a, b) {
			_ = parent.Close()
			return lineage, nil
		}
		lineage = append(lineage, parent)
		current = parent
	}
	return nil, os.ErrInvalid
}

func (d *Destination) excludeProtected() error {
	for _, chain := range d.protected {
		protected := chain[len(chain)-1].info
		for _, ancestor := range d.lineage {
			info, err := ancestor.Stat()
			if err != nil || os.SameFile(info, protected) {
				return invalidDestination("protected-destination")
			}
		}
		if d.target != nil {
			info, err := d.target.Stat()
			if err != nil {
				return invalidDestination("target-unreadable")
			}
			// Equal/containing checks use the protected root's physical lineage,
			// not string prefixes (including case-folded filesystem aliases).
			lineage, err := retainLineage(chain[len(chain)-1].file)
			if err != nil {
				return invalidDestination("ancestry-unreadable")
			}
			found := false
			for _, ancestor := range lineage {
				other, e := ancestor.Stat()
				if e != nil || os.SameFile(info, other) {
					found = true
				}
				_ = ancestor.Close()
			}
			if found {
				return invalidDestination("protected-destination")
			}
		}
	}
	return nil
}

func closeDirectories(chain []retainedDirectory) {
	for i := len(chain) - 1; i >= 0; i-- {
		_ = chain[i].file.Close()
	}
}

// Close releases all retained handles, without cleanup or any filesystem write.
func (d *Destination) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return nil
	}
	d.closed = true
	closeDirectories(d.chain)
	for _, chain := range d.protected {
		closeDirectories(chain)
	}
	for _, f := range d.lineage {
		_ = f.Close()
	}
	if d.target != nil {
		if err := d.target.Close(); err != nil {
			return invalidDestination("handle-close")
		}
	}
	return nil
}
