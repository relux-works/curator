package buildrepo

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/registry"
)

// DiskProtectedStore is an implementation-private protected snapshot/artifact
// store. Callers supply a manager-owned root, never a package-derived path.
type DiskProtectedStore struct {
	Root          string
	identityProof func(os.FileInfo, bool) bool
	proofHook     func(string)
}

func (s *DiskProtectedStore) proves(info os.FileInfo, directory bool) bool {
	if s.identityProof != nil {
		return s.identityProof(info, directory)
	}
	return protectedFileIdentity(info, directory)
}

func (s *DiskProtectedStore) prepare() error {
	return s.protectedDir(s.Root, true)
}

func cacheName(key string) (string, error) {
	value := strings.TrimPrefix(key, "sha256:")
	if len(value) != 64 {
		return "", fmt.Errorf("invalid protected key")
	}
	if _, err := hex.DecodeString(value); err != nil {
		return "", fmt.Errorf("invalid protected key")
	}
	return value, nil
}

// LoadSnapshot verifies and returns an exact protected snapshot entry.
func (s *DiskProtectedStore) LoadSnapshot(key string, mutate bool) (*Snapshot, error) {
	if err := s.prepareFor(mutate); err != nil {
		return nil, err
	}
	name, err := cacheName(key)
	if err != nil {
		return nil, err
	}
	parent := filepath.Join(s.Root, "snapshots")
	if err := s.protectedDir(parent, false); err != nil {
		return nil, err
	}
	entry := filepath.Join(parent, name)
	if err := s.protectedDir(entry, false); err != nil {
		return nil, err
	}
	filesRoot := filepath.Join(entry, "files")
	if err := s.protectedDir(filesRoot, false); err != nil {
		return nil, s.corrupt(entry, CodeObjectSemanticsInvalid, mutate, err)
	}
	guard, err := s.protectedDirGuard(s.Root, parent, entry)
	if err != nil {
		return nil, s.corrupt(entry, CodeObjectSemanticsInvalid, mutate, err)
	}
	defer guard.close()
	meta, err := s.readProtectedFile(filepath.Join(entry, "snapshot.json"), 4<<20)
	if err != nil {
		return nil, s.corrupt(entry, CodeObjectSemanticsInvalid, mutate, err)
	}
	var raw map[string]any
	decoder := json.NewDecoder(bytes.NewReader(meta))
	decoder.UseNumber()
	if protocoljson.Validate(meta) != nil || decoder.Decode(&raw) != nil {
		return nil, s.corrupt(entry, CodeObjectSemanticsInvalid, mutate, fmt.Errorf("snapshot metadata invalid"))
	}
	metadataCanonical, canonicalErr := registry.CanonicalBytesChecked(raw)
	if canonicalErr != nil || !bytes.Equal(metadataCanonical, meta) {
		return nil, s.corrupt(entry, CodeObjectSemanticsInvalid, mutate, fmt.Errorf("snapshot metadata is not exact CCJ-1"))
	}
	var record struct {
		Key          string `json:"key"`
		ObjectFormat string `json:"object_format"`
		Commit       string `json:"commit"`
		Digest       string `json:"digest"`
		Files        []struct {
			Path       string `json:"path"`
			Executable bool   `json:"executable"`
		} `json:"files"`
	}
	if json.Unmarshal(meta, &record) != nil || record.Key != key {
		return nil, s.corrupt(entry, CodeObjectSemanticsInvalid, mutate, fmt.Errorf("snapshot metadata invalid"))
	}
	files, err := s.readProtectedTree(filesRoot)
	if err != nil {
		return nil, s.corrupt(entry, CodeObjectSemanticsInvalid, mutate, err)
	}
	if len(files) != len(record.Files) {
		return nil, s.corrupt(entry, CodeObjectSemanticsInvalid, mutate, fmt.Errorf("snapshot file set differs"))
	}
	for i := range files {
		if files[i].Path != record.Files[i].Path || files[i].Executable != record.Files[i].Executable {
			return nil, s.corrupt(entry, CodeObjectSemanticsInvalid, mutate, fmt.Errorf("snapshot file metadata differs"))
		}
	}
	canonical := frameSnapshot(files)
	sum := sha256.Sum256(canonical)
	digest := "sha256:" + hex.EncodeToString(sum[:])
	if digest != record.Digest {
		return nil, s.corrupt(entry, CodeObjectSemanticsInvalid, mutate, fmt.Errorf("snapshot digest differs"))
	}
	if err := s.protectedDir(filesRoot, false); err != nil {
		return nil, s.corrupt(entry, CodeObjectSemanticsInvalid, mutate, err)
	}
	if err := guard.validate(); err != nil {
		return nil, s.corrupt(entry, CodeObjectSemanticsInvalid, mutate, err)
	}
	return &Snapshot{ObjectFormat: record.ObjectFormat, Commit: record.Commit, Digest: digest, Files: files, CanonicalBytes: canonical}, nil
}

// StoreSnapshot atomically publishes one proved snapshot.
func (s *DiskProtectedStore) StoreSnapshot(key string, snapshot *Snapshot) error {
	if err := s.prepare(); err != nil {
		return err
	}
	name, err := cacheName(key)
	if err != nil {
		return err
	}
	parent := filepath.Join(s.Root, "snapshots")
	if err = s.protectedDir(parent, true); err != nil {
		return err
	}
	final := filepath.Join(parent, name)
	if _, err = os.Lstat(final); err == nil {
		stored, loadErr := s.LoadSnapshot(key, true)
		if loadErr == nil && stored.Digest == snapshot.Digest && stored.Commit == snapshot.Commit && stored.ObjectFormat == snapshot.ObjectFormat {
			return nil
		}
		if loadErr == nil {
			_ = s.corrupt(final, CodeObjectSemanticsInvalid, true, fmt.Errorf("snapshot key collision"))
		}
	}
	stage, err := os.MkdirTemp(parent, ".stage-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(stage) }()
	if err = snapshot.Materialize(filepath.Join(stage, "files")); err != nil {
		return err
	}
	list := make([]any, len(snapshot.Files))
	for i, f := range snapshot.Files {
		list[i] = map[string]any{"path": f.Path, "executable": f.Executable}
	}
	meta, err := registry.CanonicalBytesChecked(map[string]any{"key": key, "object_format": snapshot.ObjectFormat, "commit": snapshot.Commit, "digest": snapshot.Digest, "files": list})
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(stage, "snapshot.json"), meta, 0o600); err != nil {
		return err
	}
	if err = secureProtectedTree(stage); err != nil {
		return err
	}
	return os.Rename(stage, final)
}

// LookupArtifact verifies an exact receipt-v2 artifact entry.
func (s *DiskProtectedStore) LookupArtifact(key string, input map[string]any, mutate bool) (*ArtifactHit, error) {
	if err := s.prepareFor(mutate); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	name, err := cacheName(key)
	if err != nil {
		return nil, err
	}
	entry := filepath.Join(s.Root, "artifacts", name)
	if err := s.protectedDir(filepath.Join(s.Root, "artifacts"), false); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if err := s.protectedDir(entry, false); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	guard, err := s.protectedDirGuard(s.Root, filepath.Join(s.Root, "artifacts"), entry)
	if err != nil {
		return nil, s.corrupt(entry, CodeArtifactInvalid, mutate, err)
	}
	defer guard.close()
	receipt, err := s.readProtectedFile(filepath.Join(entry, "receipt.json"), 4<<20)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, s.corrupt(entry, CodeReceiptInvalid, mutate, err)
	}
	if err = protocoljson.Validate(receipt); err != nil {
		return nil, s.corrupt(entry, CodeReceiptInvalid, mutate, err)
	}
	var obj map[string]any
	decoder := json.NewDecoder(bytes.NewReader(receipt))
	decoder.UseNumber()
	if decoder.Decode(&obj) != nil {
		return nil, s.corrupt(entry, CodeReceiptInvalid, mutate, fmt.Errorf("receipt JSON invalid"))
	}
	canonical, err := registry.CanonicalBytesChecked(obj)
	if err != nil || !bytes.Equal(canonical, receipt) {
		return nil, s.corrupt(entry, CodeReceiptInvalid, mutate, fmt.Errorf("receipt is not exact CCJ-1"))
	}
	wantInput, err := registry.CanonicalBytesChecked(input)
	if err != nil {
		return nil, err
	}
	gotInput, ok := obj["input"].(map[string]any)
	if !ok {
		return nil, s.corrupt(entry, CodeReceiptInvalid, mutate, fmt.Errorf("receipt input absent"))
	}
	gotInputBytes, _ := registry.CanonicalBytesChecked(gotInput)
	if obj["cache_key"] != key || !bytes.Equal(wantInput, gotInputBytes) {
		return nil, s.corrupt(entry, CodeReceiptInvalid, mutate, fmt.Errorf("receipt input mismatch"))
	}
	if len(obj) != 4 || !numberEquals(obj["schema_version"], 2) {
		return nil, s.corrupt(entry, CodeReceiptInvalid, mutate, fmt.Errorf("receipt closed shape mismatch"))
	}
	artifact, err := s.readProtectedFile(filepath.Join(entry, "artifact"), 1<<30)
	if err != nil {
		return nil, s.corrupt(entry, CodeArtifactInvalid, mutate, err)
	}
	artifactInfo, statErr := os.Lstat(filepath.Join(entry, "artifact"))
	if statErr != nil || runtime.GOOS != "windows" && artifactInfo.Mode().Perm()&0o111 == 0 {
		return nil, s.corrupt(entry, CodeArtifactInvalid, mutate, fmt.Errorf("artifact is not executable"))
	}
	sum := sha256.Sum256(artifact)
	meta, ok := obj["artifact"].(map[string]any)
	artifactPath, pathOK := artifactPathFromInput(input)
	if !ok || !pathOK || len(meta) != 3 || meta["path"] != artifactPath || meta["sha256"] != "sha256:"+hex.EncodeToString(sum[:]) || !numberEquals(meta["size"], len(artifact)) {
		return nil, s.corrupt(entry, CodeArtifactInvalid, mutate, fmt.Errorf("artifact metadata mismatch"))
	}
	executionReceipt, err := s.readProtectedFile(filepath.Join(entry, "execution-receipt.ccj.json"), 1<<20)
	if err != nil {
		return nil, s.corrupt(entry, CodeReceiptInvalid, mutate, err)
	}
	if _, err := closureexec.DecodeBuildSessionReceipt(executionReceipt); err != nil {
		return nil, s.corrupt(entry, CodeReceiptInvalid, mutate, err)
	}
	if err := guard.validate(); err != nil {
		return nil, s.corrupt(entry, CodeArtifactInvalid, mutate, err)
	}
	return &ArtifactHit{Bytes: artifact, Receipt: receipt, ExecutionReceipt: executionReceipt}, nil
}

// StoreArtifact atomically publishes an artifact and canonical receipt v2.
func (s *DiskProtectedStore) StoreArtifact(key string, input map[string]any, _ string, artifact, executionReceipt []byte) ([]byte, error) {
	if _, err := closureexec.DecodeBuildSessionReceipt(executionReceipt); err != nil {
		return nil, fmt.Errorf("execution receipt is invalid: %w", err)
	}
	if err := s.prepare(); err != nil {
		return nil, err
	}
	name, err := cacheName(key)
	if err != nil {
		return nil, err
	}
	parent := filepath.Join(s.Root, "artifacts")
	if err = s.protectedDir(parent, true); err != nil {
		return nil, err
	}
	final := filepath.Join(parent, name)
	if _, err = os.Lstat(final); err == nil {
		hit, e := s.LookupArtifact(key, input, true)
		if e == nil && hit != nil && bytes.Equal(hit.Bytes, artifact) && bytes.Equal(hit.ExecutionReceipt, executionReceipt) {
			return hit.Receipt, nil
		}
		if e == nil && hit != nil {
			_ = s.corrupt(final, CodeReceiptInvalid, true, fmt.Errorf("execution receipt or artifact differs for the same assured key"))
		}
	}
	stage, err := os.MkdirTemp(parent, ".stage-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(stage) }()
	if err = os.WriteFile(filepath.Join(stage, "artifact"), artifact, 0o700); err != nil {
		return nil, err
	}
	sum := sha256.Sum256(artifact)
	artifactPath, ok := artifactPathFromInput(input)
	if !ok {
		return nil, fmt.Errorf("receipt input does not identify a portable artifact path")
	}
	record := map[string]any{"schema_version": 2, "cache_key": key, "input": input, "artifact": map[string]any{"path": artifactPath, "sha256": "sha256:" + hex.EncodeToString(sum[:]), "size": len(artifact)}}
	receipt, err := registry.CanonicalBytesChecked(record)
	if err != nil {
		return nil, err
	}
	if err = os.WriteFile(filepath.Join(stage, "receipt.json"), receipt, 0o600); err != nil {
		return nil, err
	}
	if err = os.WriteFile(filepath.Join(stage, "execution-receipt.ccj.json"), executionReceipt, 0o600); err != nil {
		return nil, err
	}
	if err = secureProtectedTree(stage); err != nil {
		return nil, err
	}
	if err = os.Rename(stage, final); err != nil {
		return nil, err
	}
	return receipt, nil
}

func (s *DiskProtectedStore) readProtectedFile(name string, maxBytes int64) ([]byte, error) {
	if s.identityProof == nil {
		return readNativeProtectedFile(name, maxBytes, s.proofHook)
	}
	info, err := os.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() > maxBytes || !s.proves(info, false) {
		return nil, fmt.Errorf("protected file shape invalid")
	}
	return os.ReadFile(name) // #nosec G304 -- manager-derived path was shape- and ownership-proved above.
}

func (s *DiskProtectedStore) readProtectedTree(root string) ([]File, error) {
	var files []File
	err := filepath.WalkDir(root, func(name string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return s.protectedDir(name, false)
		}
		info, err := os.Lstat(name)
		if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("non-regular protected snapshot entry")
		}
		payload, err := s.readProtectedFile(name, 1<<30)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		files = append(files, File{Path: filepath.ToSlash(rel), Content: payload, Executable: info.Mode()&0o100 != 0})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

func (s *DiskProtectedStore) protectedDir(name string, create bool) error {
	if s.identityProof == nil {
		return nativeProtectedDir(name, create, s.proofHook)
	}
	if create {
		if err := os.MkdirAll(name, 0o700); err != nil {
			return err
		}
	}
	info, err := os.Lstat(name)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm()&0o022 != 0 || !s.proves(info, true) {
		return admissionError(CodeProtectedBoundaryUntrusted, "protected directory cannot be proved private")
	}
	return nil
}

func (s *DiskProtectedStore) protectedDirGuard(names ...string) (*nativeProtectedDirGuard, error) {
	if s.identityProof != nil {
		return &nativeProtectedDirGuard{}, nil
	}
	return openNativeProtectedDirGuard(names, s.proofHook)
}

func (s *DiskProtectedStore) prepareFor(mutate bool) error {
	if mutate {
		return s.prepare()
	}
	return s.protectedDir(s.Root, false)
}
func numberEquals(value any, want int) bool {
	switch n := value.(type) {
	case json.Number:
		return string(n) == fmt.Sprint(want)
	case float64:
		return n == float64(want)
	case int:
		return n == want
	}
	return false
}

func artifactPathFromInput(input map[string]any) (string, bool) {
	command, ok := input["command"].(string)
	if !ok || command == "" || strings.ContainsAny(command, "/\\") {
		return "", false
	}
	target, ok := input["target"].(map[string]any)
	if !ok {
		return "", false
	}
	goos, ok := target["goos"].(string)
	if !ok {
		return "", false
	}
	path, err := buildmeta.ArtifactPath(command, goos)
	return path, err == nil
}
func (s *DiskProtectedStore) corrupt(entry, code string, mutate bool, cause error) error {
	if mutate {
		q := filepath.Join(s.Root, "quarantine")
		if s.protectedDir(q, true) == nil {
			_ = os.Rename(entry, filepath.Join(q, filepath.Base(entry)+fmt.Sprintf("-%d", time.Now().UnixNano())))
		}
	}
	return admissionError(code, "protected entry invalid: %v", cause)
}
