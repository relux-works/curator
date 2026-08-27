package closureexec

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/closuregraph"
)

// CaptureStore owns raw immutable bytes. Manager caches are never accepted as
// provenance; adapters may only receive a handle returned by Capture.
type CaptureStore struct{ root, trees string }

// CaptureHandle is an opaque reference to exact protected captured bytes.
type CaptureHandle struct {
	store    *CaptureStore
	id, path string
	digest   closuregraph.ID
	size     int64
}

// SnapshotFile is one immutable regular file in a captured source tree.
type SnapshotFile struct {
	Path       string
	SHA256     closuregraph.ID
	Size       int64
	Executable bool
}

// SourceTreeHandle is an opaque immutable source snapshot. Providers may map
// its protected path read-only only after VerifyAtUse succeeds.
type SourceTreeHandle struct {
	store    *CaptureStore
	id, path string
	digest   closuregraph.ID
	size     int64
	files    []SnapshotFile
}

// InputMount binds one admitted receipt to one logical read-only replay root.
type InputMount struct {
	ReceiptID closuregraph.ID `json:"receipt_id"`
	Path      string          `json:"path"`
}

func (mount InputMount) validate() error {
	if !mount.ReceiptID.Valid() {
		return fmt.Errorf("invalid input receipt identity")
	}
	return portablePath(mount.Path)
}
func (mount InputMount) canonicalValue() map[string]any {
	return map[string]any{"path": mount.Path, "receipt_id": string(mount.ReceiptID)}
}

// WorkCopy declares a writable operation-local copy seeded from one immutable
// admitted input. The admitted replay itself always remains read-only.
type WorkCopy struct {
	ReceiptID closuregraph.ID `json:"receipt_id"`
	Path      string          `json:"path"`
	// Retain leaves the writable derivative in the task-private execution root
	// after a successful process so the manager can reconcile and publish it.
	// Failed operations always remove it. The flag grants no additional write
	// authority: Path must still be an exact typed WriteRoot.
	Retain bool `json:"retain"`
}

func (work WorkCopy) validate() error {
	if !work.ReceiptID.Valid() {
		return fmt.Errorf("invalid work-copy receipt identity")
	}
	return portablePath(work.Path)
}

func (work WorkCopy) canonicalValue() map[string]any {
	return map[string]any{"path": work.Path, "receipt_id": string(work.ReceiptID), "retain": work.Retain}
}

// AdmissionEvidence is supplied only after the shared artifact classifier has
// admitted the exact captured bytes.
type AdmissionEvidence struct {
	PreviousCausalHead string
	ArtifactPolicyID   string
	SourceProfileID    string
	DetectorRegistryID string
	LimitVectorID      string
	ArtifactManifestID closuregraph.ID
}

// Admit binds classifier evidence to this exact protected handle.
func (s *CaptureStore) Admit(handle *CaptureHandle, originID string, evidence AdmissionEvidence) (IntakeAdmissionReceipt, error) {
	if s == nil || handle == nil || handle.store != s {
		return IntakeAdmissionReceipt{}, failure("closure_derivation_unauthorized", "capture handle is absent or foreign")
	}
	if err := handle.Recheck(); err != nil {
		return IntakeAdmissionReceipt{}, err
	}
	receipt := IntakeAdmissionReceipt{SchemaID: SchemaIntakeReceipt, PreviousCausalHead: evidence.PreviousCausalHead, OriginID: originID, ProtectedHandleID: handle.id, ContentSHA256: handle.digest, Size: handle.size, ArtifactPolicyID: evidence.ArtifactPolicyID, SourceProfileID: evidence.SourceProfileID, DetectorRegistryID: evidence.DetectorRegistryID, LimitVectorID: evidence.LimitVectorID, ArtifactManifestID: evidence.ArtifactManifestID, Decision: "ADMIT_INPUT"}
	if err := receipt.Validate(); err != nil {
		return IntakeAdmissionReceipt{}, err
	}
	return receipt, nil
}

// AdmittedInput pairs a canonical admission receipt with its opaque protected handle.
type AdmittedInput struct {
	Receipt IntakeAdmissionReceipt
	Handle  *CaptureHandle
	Tree    *SourceTreeHandle
}

func (input AdmittedInput) recheck(expected closuregraph.ID) error {
	if (input.Handle == nil) == (input.Tree == nil) {
		return failure("closure_derivation_unauthorized", "admitted input handle is absent")
	}
	id, err := input.Receipt.ID()
	if err != nil {
		return err
	}
	handleID, digest, size := "", closuregraph.ID(""), int64(0)
	if input.Handle != nil {
		handleID, digest, size = input.Handle.id, input.Handle.digest, input.Handle.size
	}
	if input.Tree != nil {
		handleID, digest, size = input.Tree.id, input.Tree.digest, input.Tree.size
	}
	if id != expected || input.Receipt.ProtectedHandleID != handleID || input.Receipt.ContentSHA256 != digest || input.Receipt.Size != size {
		return failure("closure_derivation_unauthorized", "intake receipt does not bind protected handle")
	}
	if input.Handle != nil {
		return input.Handle.Recheck()
	}
	return input.Tree.VerifyAtUse()
}

// ID returns the origin-and-content-bound protected handle identity.
func (h *CaptureHandle) ID() string {
	if h == nil {
		return ""
	}
	return h.id
}

// Digest returns the captured content digest.
func (h *CaptureHandle) Digest() closuregraph.ID {
	if h == nil {
		return ""
	}
	return h.digest
}

// Size returns the exact captured byte size.
func (h *CaptureHandle) Size() int64 {
	if h == nil {
		return 0
	}
	return h.size
}

// NewCaptureStore prepares an owner-only immutable capture root.
func NewCaptureStore(root string) (*CaptureStore, error) {
	abs, e := filepath.Abs(root)
	if e != nil {
		return nil, e
	}
	if filepath.Clean(root) != root && filepath.IsAbs(root) {
		return nil, fmt.Errorf("capture root must be clean")
	}
	if e = preparePrivateDirectory(abs); e != nil {
		return nil, e
	}
	trees := filepath.Join(abs, "trees")
	if e = preparePrivateDirectory(trees); e != nil {
		return nil, e
	}
	return &CaptureStore{root: abs, trees: trees}, nil
}

// CaptureTree copies a directory into protected immutable storage, rejecting
// links and special nodes. The canonical tree digest covers every path, size,
// and file digest.
func (s *CaptureStore) CaptureTree(origin, source string) (*SourceTreeHandle, error) {
	if s == nil || origin == "" {
		return nil, fmt.Errorf("invalid tree capture request")
	}
	abs, err := filepath.Abs(source)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(abs)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, failure("closure_input_undeclared", "source snapshot root is not a real directory")
	}
	tmp, err := os.MkdirTemp(s.trees, ".tree-*.tmp")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	files := []SnapshotFile{}
	var total int64
	err = filepath.WalkDir(abs, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == abs {
			return nil
		}
		rel, relErr := filepath.Rel(abs, current)
		if relErr != nil {
			return relErr
		}
		logical := filepath.ToSlash(rel)
		if err := portablePath(logical); err != nil {
			return err
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return failure("closure_input_undeclared", "source snapshot contains a link")
		}
		target := filepath.Join(tmp, rel)
		if entry.IsDir() {
			return os.Mkdir(target, 0o700)
		}
		entryInfo, infoErr := entry.Info()
		if infoErr != nil {
			return infoErr
		}
		if !entryInfo.Mode().IsRegular() {
			return failure("closure_input_undeclared", "source snapshot contains a special node")
		}
		payload, readErr := os.ReadFile(current) // #nosec G304 -- WalkDir supplies a containment-checked member below the admitted source root.
		if readErr != nil {
			return readErr
		}
		if int64(len(payload)) != entryInfo.Size() {
			return failure("closure_derivation_drift", "source changed during capture")
		}
		mode := fs.FileMode(0o400)
		if entryInfo.Mode().Perm()&0o111 != 0 {
			mode = 0o500
		}
		if writeErr := os.WriteFile(target, payload, mode); writeErr != nil {
			return writeErr
		}
		files = append(files, SnapshotFile{Path: logical, SHA256: digestBytes(payload), Size: int64(len(payload)), Executable: mode&0o100 != 0})
		total += int64(len(payload))
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err = filepath.WalkDir(tmp, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return markTreeDirImmutable(current)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	digest, err := snapshotDigest(files)
	if err != nil {
		return nil, err
	}
	idValue, err := closuregraph.DomainID("curator-protected-source-tree-v1", map[string]any{"content_sha256": string(digest), "origin_id": origin, "size": total})
	if err != nil {
		return nil, err
	}
	target := filepath.Join(s.trees, strings.TrimPrefix(string(idValue), "sha256:"))
	if err = os.Rename(tmp, target); err != nil {
		// A digest-named tree that already exists is a reuse, and VerifyAtUse
		// below re-proves its exact content either way. Windows reports the
		// occupied target as access-denied rather than fs.ErrExist, so the
		// reuse condition is read from the target itself, not the error kind.
		if info, statErr := os.Lstat(target); statErr != nil || !info.IsDir() {
			return nil, err
		}
	}
	handle := &SourceTreeHandle{store: s, id: string(idValue), path: target, digest: digest, size: total, files: files}
	if err = handle.VerifyAtUse(); err != nil {
		return nil, err
	}
	return handle, nil
}

// AdmitTree binds classifier evidence to an immutable source snapshot tree.
func (s *CaptureStore) AdmitTree(handle *SourceTreeHandle, originID string, evidence AdmissionEvidence) (IntakeAdmissionReceipt, error) {
	if s == nil || handle == nil || handle.store != s {
		return IntakeAdmissionReceipt{}, failure("closure_derivation_unauthorized", "source tree handle is absent or foreign")
	}
	if err := handle.VerifyAtUse(); err != nil {
		return IntakeAdmissionReceipt{}, err
	}
	receipt := IntakeAdmissionReceipt{SchemaID: SchemaIntakeReceipt, PreviousCausalHead: evidence.PreviousCausalHead, OriginID: originID, ProtectedHandleID: handle.id, ContentSHA256: handle.digest, Size: handle.size, ArtifactPolicyID: evidence.ArtifactPolicyID, SourceProfileID: evidence.SourceProfileID, DetectorRegistryID: evidence.DetectorRegistryID, LimitVectorID: evidence.LimitVectorID, ArtifactManifestID: evidence.ArtifactManifestID, Decision: "ADMIT_INPUT"}
	if err := receipt.Validate(); err != nil {
		return IntakeAdmissionReceipt{}, err
	}
	return receipt, nil
}

// VerifyAtUse rewalks the protected tree and checks containment, node type,
// permissions, paths, sizes, and digests immediately before provider use.
func (h *SourceTreeHandle) VerifyAtUse() error {
	if h == nil || h.store == nil {
		return failure("closure_derivation_unauthorized", "source tree handle is absent")
	}
	if err := validatePrivateDirectory(h.store.root); err != nil {
		return err
	}
	if err := validatePrivateDirectory(h.store.trees); err != nil {
		return err
	}
	rel, err := filepath.Rel(h.store.trees, h.path)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.Base(h.path) != strings.TrimPrefix(h.id, "sha256:") {
		return failure("closure_derivation_drift", "source tree handle escapes protected containment")
	}
	rootInfo, err := os.Lstat(h.path)
	if err != nil {
		return err
	}
	if !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 || !treeDirIsImmutable(rootInfo) {
		return failure("closure_derivation_drift", "source tree root is mutable or linked")
	}
	files := []SnapshotFile{}
	var total int64
	err = filepath.WalkDir(h.path, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == h.path {
			return nil
		}
		rel, relErr := filepath.Rel(h.path, current)
		if relErr != nil {
			return relErr
		}
		logical := filepath.ToSlash(rel)
		if err := portablePath(logical); err != nil {
			return err
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return failure("closure_derivation_drift", "source tree contains a link")
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			return infoErr
		}
		if entry.IsDir() {
			if !treeDirIsImmutable(info) {
				return failure("closure_derivation_drift", "source tree member is writable")
			}
			return nil
		}
		if info.Mode().Perm()&0o222 != 0 {
			return failure("closure_derivation_drift", "source tree member is writable")
		}
		if !info.Mode().IsRegular() {
			return failure("closure_derivation_drift", "source tree contains a special node")
		}
		payload, readErr := os.ReadFile(current) // #nosec G304 -- WalkDir supplies a protected member below the rechecked snapshot root.
		if readErr != nil {
			return readErr
		}
		files = append(files, SnapshotFile{Path: logical, SHA256: digestBytes(payload), Size: int64(len(payload)), Executable: info.Mode().Perm()&0o100 != 0})
		total += int64(len(payload))
		return nil
	})
	if err != nil {
		return err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	digest, err := snapshotDigest(files)
	if err != nil {
		return err
	}
	if total != h.size || digest != h.digest || !snapshotFilesEqual(files, h.files) {
		return failure("closure_derivation_drift", "source snapshot identity changed")
	}
	return nil
}

// ProtectedPath returns the contained snapshot path after a fresh identity check.
func (h *SourceTreeHandle) ProtectedPath() (string, error) {
	if err := h.VerifyAtUse(); err != nil {
		return "", err
	}
	return h.path, nil
}

func snapshotDigest(files []SnapshotFile) (closuregraph.ID, error) {
	values := make([]any, len(files))
	for i, f := range files {
		if err := portablePath(f.Path); err != nil || !f.SHA256.Valid() || f.Size < 0 {
			return "", fmt.Errorf("invalid snapshot file")
		}
		values[i] = map[string]any{"executable": f.Executable, "path": f.Path, "sha256": string(f.SHA256), "size": f.Size}
	}
	return closuregraph.DomainID("curator-source-tree-content-v1", map[string]any{"files": values})
}
func snapshotFilesEqual(a, b []SnapshotFile) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Capture streams exact bytes to an absent destination and makes them read-only.
func (s *CaptureStore) Capture(origin string, size int64, reader io.Reader) (*CaptureHandle, error) {
	if s == nil || origin == "" || size < 0 || reader == nil {
		return nil, fmt.Errorf("invalid capture request")
	}
	tmp, e := os.CreateTemp(s.root, ".capture-*.tmp")
	if e != nil {
		return nil, e
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	h := shaWriter()
	n, e := io.Copy(io.MultiWriter(tmp, h), io.LimitReader(reader, size+1))
	closeErr := tmp.Close()
	if e != nil {
		return nil, e
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if n != size {
		return nil, fmt.Errorf("capture size mismatch: expected %d, observed %d", size, n)
	}
	digest := h.id()
	id := captureID(origin, digest, size)
	target := filepath.Join(s.root, strings.TrimPrefix(id, "sha256:"))
	if e = publishByLink(tmpPath, target, 0o400); e != nil {
		return nil, e
	}
	handle := &CaptureHandle{store: s, id: id, path: target, digest: digest, size: size}
	if e = handle.Recheck(); e != nil {
		return nil, e
	}
	return handle, nil
}

// Recheck hashes the protected bytes immediately before each use.
func (h *CaptureHandle) Recheck() error {
	if h == nil || h.store == nil {
		return failure("closure_derivation_unauthorized", "capture handle is absent")
	}
	if err := validatePrivateDirectory(h.store.root); err != nil {
		return err
	}
	info, e := os.Lstat(h.path)
	if e != nil {
		return e
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0o222 != 0 {
		return failure("closure_derivation_drift", "capture is mutable or not regular")
	}
	f, e := os.Open(h.path)
	if e != nil {
		return e
	}
	defer func() { _ = f.Close() }()
	w := shaWriter()
	n, e := io.Copy(w, f)
	if e != nil {
		return e
	}
	if n != h.size || w.id() != h.digest {
		return failure("closure_derivation_drift", "captured bytes changed")
	}
	return nil
}

// Open rechecks and opens the immutable capture for read-only consumption.
func (h *CaptureHandle) Open() (io.ReadCloser, error) {
	if e := h.Recheck(); e != nil {
		return nil, e
	}
	return os.Open(h.path)
}

type digestWriter struct{ h hashHash }
type hashHash interface {
	Write([]byte) (int, error)
	Sum([]byte) []byte
}

func shaWriter() *digestWriter                      { return &digestWriter{h: newSHA256()} }
func (w *digestWriter) Write(b []byte) (int, error) { return w.h.Write(b) }
func (w *digestWriter) id() closuregraph.ID {
	return closuregraph.ID("sha256:" + fmt.Sprintf("%x", w.h.Sum(nil)))
}

func captureID(origin string, d closuregraph.ID, size int64) string {
	id, _ := closuregraph.DomainID("curator-protected-capture-v1", map[string]any{"content_sha256": string(d), "origin_id": origin, "size": size})
	return string(id)
}
