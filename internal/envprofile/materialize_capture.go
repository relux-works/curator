package envprofile

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
)

// Snapshot capture diagnostics preserve the normative refusal names.
const (
	DiagMaterializeRecoveryRequired = "environment_materialize_recovery_required"
	DiagMaterializeCaptureDrift     = "environment_materialize_capture_drift"
	// These local capture bounds are not transport/archive limits.
	captureMaxBytes   int64 = 64 << 20
	captureMaxObjects       = 16384
	captureTimeout          = 30 * time.Second
)

// MaterializeCaptureRequest separates fresh machine selection from content
// capture. ReloadMachine must read effective policy anew on each call, without
// provisioning or executing providers. Capture is the bounded, local closure
// builder: it must use ReadObject for admitted objects, never publish/transport
// or wait on services. SNAP-CLI supplies ReloadMachine; SNAP-CLOSURE owns the
// inventory, protected source handles, relocation and secret admission.
// Neither callback may retain or mutate the capture after it returns.
type MaterializeCaptureRequest struct {
	Resolve       ResolveRequest
	ReloadMachine func() (envregistry.MachineConfig, Policy, error)
	Capture       func(*MaterializeCapture) ([]byte, error)
}

// CapturedMaterialization is returned only after revalidation and lock release.
// Publication and transport consume these private bytes outside this operation.
type CapturedMaterialization struct {
	Fragment []byte
	Content  []byte
}

// MaterializeCapture binds the closure to the verified inputs under the lock.
type MaterializeCapture struct {
	Request  ResolveRequest
	Fragment []byte
	ctx      context.Context
	objects  []captureObject
	total    int64
}

type captureObject struct {
	path   string
	file   *os.File
	info   os.FileInfo
	digest [32]byte
}

// ReadObject takes ownership of a protected regular-file handle supplied by
// the closure builder. The path must name that same object. It records identity,
// bytes, size, mode and timestamps for final revalidation; it never follows a
// source link or opens credential bytes to discover an inventory.
func (c *MaterializeCapture) ReadObject(path string, file *os.File) ([]byte, error) {
	if file == nil {
		return nil, captureDrift()
	}
	if err := c.ctx.Err(); err != nil {
		_ = file.Close()
		return nil, captureDrift()
	}
	before, err := file.Stat()
	if err != nil || !before.Mode().IsRegular() || before.Size() > captureMaxBytes-c.total || len(c.objects) >= captureMaxObjects {
		_ = file.Close()
		return nil, captureDrift()
	}
	named, err := os.Lstat(path)
	if err != nil || !sameCaptureInfo(before, named) {
		_ = file.Close()
		return nil, captureDrift()
	}
	data, err := io.ReadAll(io.LimitReader(file, captureMaxBytes-c.total+1))
	after, statErr := file.Stat()
	if err != nil || statErr != nil || int64(len(data)) != before.Size() || !sameCaptureInfo(before, after) {
		_ = file.Close()
		return nil, captureDrift()
	}
	c.total += int64(len(data))
	c.objects = append(c.objects, captureObject{path, file, before, sha256.Sum256(data)})
	return data, nil
}

func sameCaptureInfo(a, b os.FileInfo) bool {
	return os.SameFile(a, b) && a.Mode() == b.Mode() && a.Size() == b.Size() && a.ModTime().Equal(b.ModTime())
}

func captureDrift() error {
	return fmt.Errorf("%s: source generation changed or could not be revalidated", DiagMaterializeCaptureDrift)
}

func (c *MaterializeCapture) close() {
	for _, object := range c.objects {
		_ = object.file.Close()
	}
}

func (c *MaterializeCapture) revalidateObjects() error {
	for _, object := range c.objects {
		if c.ctx.Err() != nil {
			return captureDrift()
		}
		info, err := os.Lstat(object.path)
		if err != nil || !sameCaptureInfo(object.info, info) {
			return captureDrift()
		}
		info, err = object.file.Stat()
		if err != nil || !sameCaptureInfo(object.info, info) {
			return captureDrift()
		}
		if _, err := object.file.Seek(0, io.SeekStart); err != nil {
			return captureDrift()
		}
		h := sha256.New()
		size, err := io.Copy(h, io.LimitReader(object.file, object.info.Size()+1))
		if err != nil || size != object.info.Size() || !bytes.Equal(h.Sum(nil), object.digest[:]) {
			return captureDrift()
		}
		info, err = object.file.Stat()
		if err != nil || !sameCaptureInfo(object.info, info) {
			return captureDrift()
		}
		info, err = os.Lstat(object.path)
		if err != nil || !sameCaptureInfo(object.info, info) {
			return captureDrift()
		}
	}
	return nil
}

// CaptureMaterialize is the non-recovering capture entry point, shared by
// export and manifest-only. It cannot publish a destination or repair a source.
func CaptureMaterialize(req MaterializeCaptureRequest) (result *CapturedMaterialization, resultErr error) {
	if req.ReloadMachine == nil || req.Capture == nil || req.Resolve.Repair || req.Resolve.DryRun || req.Resolve.PreserveLock {
		return nil, fmt.Errorf("usage: capture requires read-only inputs and local capture callbacks")
	}
	op, err := beginCaptureOperation(req.Resolve.Home, req.Resolve.transactionOptions...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := op.close(); resultErr == nil {
			resultErr = err
			if err != nil {
				result = nil
			}
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), captureTimeout)
	defer cancel()
	selectInputs := func() (ResolveRequest, error) {
		selected := req.Resolve
		selected.borrowedOperation = op
		selected.Machine, selected.Policy, err = req.ReloadMachine()
		if err != nil {
			return ResolveRequest{}, err
		}
		selected.Profile, err = currentProfileFor(selected.Home, selected.EnvID, req.Resolve.Profile)
		selected.Format = "json"
		return selected, err
	}
	selected, err := selectInputs()
	if err != nil {
		return nil, err
	}
	verified, err := resolveVerified(selected, verifyHomeContent)
	if err != nil {
		return nil, err
	}
	c := &MaterializeCapture{Request: selected, Fragment: append([]byte(nil), verified.Document...), ctx: ctx}
	defer c.close()
	// Retain identity as well as bytes for the records that authorize capture.
	for _, path := range []string{sourcePath(selected.Home, selected.Profile), lockPath(selected.Home, selected.Profile), filepath.Join(ManagedHomeDir(selected.Home, selected.Profile, selected.EnvID), envmarker.Name)} {
		file, err := openReadNoFollow(path)
		if err != nil {
			return nil, captureDrift()
		}
		if _, err := c.ReadObject(path, file); err != nil {
			return nil, err
		}
	}
	// Copy policy values: a callback must not make a changed map compare equal by
	// mutating the original machine object in place.
	baseline, err := json.Marshal(struct {
		Machine envregistry.MachineConfig
		Policy  Policy
	}{selected.Machine, selected.Policy})
	if err != nil {
		return nil, err
	}
	content, err := req.Capture(c)
	if err != nil {
		return nil, err
	}
	next, err := selectInputs()
	if err != nil || next.Profile != selected.Profile {
		return nil, captureDrift()
	}
	nextPolicy, err := json.Marshal(struct {
		Machine envregistry.MachineConfig
		Policy  Policy
	}{next.Machine, next.Policy})
	if err != nil || !bytes.Equal(baseline, nextPolicy) {
		return nil, captureDrift()
	}
	if err := op.lock.AssertCurrent(); err != nil {
		return nil, captureDrift()
	}
	pending, err := op.engine.PendingRecovery(op.lock)
	if err != nil || pending {
		return nil, captureDrift()
	}
	after, err := resolveVerified(next, verifyHomeContent)
	if err != nil || !bytes.Equal(verified.Document, after.Document) {
		return nil, captureDrift()
	}
	if err := c.revalidateObjects(); err != nil {
		return nil, err
	}
	if ctx.Err() != nil {
		return nil, captureDrift()
	}
	return &CapturedMaterialization{Fragment: append([]byte(nil), verified.Document...), Content: append([]byte(nil), content...)}, nil
}
