//go:build darwin || linux

package envsnapshot

import (
	"context"
	"crypto/rand"
	"errors"
	"os"
	"path"
	"sort"

	"golang.org/x/sys/unix"
)

type stagingObject struct {
	parent *os.File
	file   *os.File
	name   string
	info   os.FileInfo
}

// Publish writes independent private copies, syncs, publishes without replacing
// any destination entry, then verifies the published tree against the capture.
// A verified prior is a no-op. Errors return no success object. After rename,
// failure leaves the published entry for the operator, never rolling it back.
// Cancellation is cooperative; abrupt process death can leave private staging.
// Same-user mutation inside private staging is outside snapshot section 2's
// guarantee. Cleanup records objects, never recursively enumerates a pathname.
func (d *Destination) Publish(ctx context.Context, captured Publication) (result *Verified, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if err = d.revalidate(); err != nil {
		return nil, err
	}
	p, dirs, err := captured.freeze()
	if err != nil {
		return nil, err
	}
	if err = publicationContext(ctx); err != nil {
		return nil, err
	}
	if d.kind == DestinationSnapshot {
		if string(d.prior) != string(p.Manifest) {
			return nil, refuse("generation-mismatch")
		}
		return &Verified{p.Manifest, p.Launch}, nil
	}
	parent := d.chain[len(d.chain)-1].file
	stageName := ".curator-snapshot-" + rand.Text()
	var objects []stagingObject
	published := false
	defer func() {
		if !published {
			cleanupErr := cleanupStaging(objects)
			if cleanupErr != nil {
				result = nil
				err = errors.Join(err, invalidDestination("staging-cleanup"))
			}
		}
		for i := len(objects) - 1; i >= 0; i-- {
			_ = objects[i].file.Close()
		}
	}()
	stage, err := createStagingDirectory(parent, stageName, &objects)
	if err != nil {
		return nil, err
	}
	if err = d.publicationStep(ctx, "stage-created", stage); err != nil {
		return nil, err
	}
	handles := map[string]*os.File{"": stage}
	names := make([]string, 0, len(dirs))
	for name := range dirs {
		names = append(names, name)
	}
	sort.Strings(names) // Every parent precedes its children.
	for _, name := range names {
		if err = publicationContext(ctx); err != nil {
			return nil, err
		}
		child, e := createStagingDirectory(handles[publicationParent(name)], path.Base(name), &objects)
		if e != nil {
			return nil, e
		}
		handles[name] = child
	}
	m, _ := DecodeManifest(p.Manifest) // freeze already checked metadata.
	for _, f := range m.Files {
		mode := os.FileMode(0600)
		if f.Mode == "0700" {
			mode = 0700 // Only the capture coordinator supplies execute metadata.
		}
		if err = publicationContext(ctx); err != nil {
			return nil, err
		}
		if err = createStagingFile(handles[publicationParent(f.Path)], path.Base(f.Path), p.Files[f.Path], mode, &objects); err != nil {
			return nil, err
		}
		if err = d.publicationStep(ctx, "file-staged", stage); err != nil {
			return nil, err
		}
	}
	if err = createStagingFile(stage, "launch.json", p.Launch, 0600, &objects); err != nil {
		return nil, err
	}
	// Manifest last; private partial output never constitutes a success result.
	if err = createStagingFile(stage, "manifest.json", p.Manifest, 0600, &objects); err != nil {
		return nil, err
	}
	for i := len(objects) - 1; i >= 0; i-- {
		if objects[i].info.IsDir() {
			if err = objects[i].file.Sync(); err != nil {
				return nil, invalidDestination("staging-sync")
			}
		}
	}
	if err = d.revalidate(); err != nil {
		return nil, err
	}
	// Tests insert real kernel competition AFTER the last admission check.
	if err = d.publicationStep(ctx, "before-rename", stage); err != nil {
		return nil, err
	}
	if err = publishNoReplace(int(parent.Fd()), stageName, d.name); err != nil {
		return nil, invalidDestination("publication-no-replace")
	}
	published = true // No cleanup may touch this entry from this point onward.
	if err = parent.Sync(); err != nil {
		return nil, invalidDestination("publication-sync")
	}
	if err = d.publicationStep(ctx, "after-rename", stage); err != nil {
		return nil, err
	}
	if err = recheckDirectories(d.chain); err != nil {
		return nil, err
	}
	actual, err := openMember(parent, d.name, true)
	if err != nil {
		return nil, refuse("published-unreadable")
	}
	defer func() { _ = actual.Close() }()
	info, err := actual.Stat()
	if err != nil || !os.SameFile(objects[0].info, info) {
		return nil, refuse("published-identity")
	}
	v, err := verifyDirectory(actual, Digest(p.Manifest))
	if err != nil {
		return nil, err
	}
	if err = priorOwnershipACL(actual); err != nil {
		return nil, refuse("snapshot-owner-acl")
	}
	if err = publicationContext(ctx); err != nil {
		return nil, err
	}
	// Retain the published target so a second call is a verified no-op.
	d.target, err = openMember(parent, d.name, true)
	if err != nil {
		return nil, refuse("published-unreadable")
	}
	d.kind, d.prior = DestinationSnapshot, p.Manifest
	if err = d.revalidate(); err != nil {
		return nil, err
	}
	return v, nil
}

func publicationParent(name string) string {
	parent := path.Dir(name)
	if parent == "." {
		return ""
	}
	return parent
}

func publicationContext(ctx context.Context) error {
	if ctx.Err() != nil {
		return invalidDestination("publication-interrupted")
	}
	return nil
}

func (d *Destination) publicationStep(ctx context.Context, phase string, stage *os.File) error {
	if d.publicationHook != nil {
		if err := d.publicationHook(phase, stage); err != nil {
			return invalidDestination("staging-io")
		}
	}
	return publicationContext(ctx)
}

func recordStaging(parent, f *os.File, name string, objects *[]stagingObject) error {
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return invalidDestination("staging-stat")
	}
	*objects = append(*objects, stagingObject{parent, f, name, info})
	return nil
}

func createStagingDirectory(parent *os.File, name string, objects *[]stagingObject) (*os.File, error) {
	if err := unix.Mkdirat(int(parent.Fd()), name, 0700); err != nil {
		return nil, invalidDestination("staging-create")
	}
	f, err := openMember(parent, name, true)
	if err != nil {
		return nil, invalidDestination("staging-open")
	}
	if err := recordStaging(parent, f, name, objects); err != nil {
		return nil, err
	}
	if err := f.Chmod(0700); err != nil {
		return nil, invalidDestination("staging-mode")
	}
	if err := destinationDirectory(f, true, true); err != nil {
		return nil, err
	}
	return f, nil
}

func createStagingFile(parent *os.File, name string, data []byte, mode os.FileMode, objects *[]stagingObject) error {
	fd, err := unix.Openat(int(parent.Fd()), name, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0600)
	if err != nil {
		return invalidDestination("staging-create")
	}
	f := os.NewFile(uintptr(fd), name)
	if err := recordStaging(parent, f, name, objects); err != nil {
		return err
	}
	if err := f.Chmod(mode); err != nil {
		return invalidDestination("staging-mode")
	}
	if _, err := f.Write(data); err != nil {
		return invalidDestination("staging-write")
	}
	if err := f.Sync(); err != nil {
		return invalidDestination("staging-sync")
	}
	return nil
}

func cleanupStaging(objects []stagingObject) error {
	var failure error
	for i := len(objects) - 1; i >= 0; i-- {
		obj := objects[i]
		fresh, err := openMember(obj.parent, obj.name, obj.info.IsDir())
		if err != nil {
			failure = errors.Join(failure, err)
			continue
		}
		info, statErr := fresh.Stat()
		_ = fresh.Close()
		if statErr != nil || !os.SameFile(obj.info, info) {
			failure = errors.Join(failure, os.ErrInvalid)
			continue
		}
		flags := 0
		if obj.info.IsDir() {
			flags = unix.AT_REMOVEDIR
		}
		if err := unix.Unlinkat(int(obj.parent.Fd()), obj.name, flags); err != nil {
			failure = errors.Join(failure, err)
		}
	}
	return failure
}
