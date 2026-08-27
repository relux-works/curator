package closureexec

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/protocoljson"
)

// PublishedHit is returned only after an independently derived expected input
// and every protected blob and receipt have been revalidated.
type PublishedHit struct {
	Publication closuregraph.PublicationReceipt
	Paths       map[string]string
}

// ProtectedStore publishes generic multi-output closures. Blobs and the entry
// receipt are installed with no-replace hard links; the receipt is the single
// atomic visibility point, so a crash cannot expose a reusable partial entry.
type ProtectedStore struct{ root, blobs, receipts string }

// NewProtectedStore prepares an owner-only multi-output protected store.
func NewProtectedStore(root string) (*ProtectedStore, error) {
	abs, e := filepath.Abs(root)
	if e != nil {
		return nil, e
	}
	s := &ProtectedStore{root: abs, blobs: filepath.Join(abs, "blobs"), receipts: filepath.Join(abs, "receipts")}
	for _, d := range []string{s.root, s.blobs, s.receipts} {
		if e = preparePrivateDirectory(d); e != nil {
			return nil, e
		}
	}
	return s, nil
}

type storedOutput struct {
	Path             string          `json:"path"`
	ObservationBytes string          `json:"observation_bytes"`
	ObservationID    closuregraph.ID `json:"observation_id"`
	SHA256           closuregraph.ID `json:"sha256"`
	Size             int64           `json:"size"`
}
type storedEntry struct {
	SchemaID               string          `json:"schema_id"`
	AssuredCacheInputID    closuregraph.ID `json:"assured_cache_input_id"`
	AssuredCacheInputBytes string          `json:"assured_cache_input_bytes"`
	ExpectedCacheInputID   closuregraph.ID `json:"expected_cache_input_id"`
	ExecutionReceiptID     closuregraph.ID `json:"execution_receipt_id"`
	PublicationBytes       string          `json:"publication_bytes"`
	Outputs                []storedOutput  `json:"outputs"`
}

// Publish validates the exact declared write set, hashes immutable output
// bytes, installs blobs, and atomically publishes one canonical receipt.
func (s *ProtectedStore) Publish(authority closuregraph.PublicationEvidence, cacheInput AssuredCacheInput, execution closuregraph.ExecutionReceipt, observations []closuregraph.ProducedArtifactObservation, staging string) (closuregraph.PublicationReceipt, error) {
	if s == nil {
		return closuregraph.PublicationReceipt{}, fmt.Errorf("protected store is absent")
	}
	for _, directory := range []string{s.root, s.blobs, s.receipts} {
		if err := validatePrivateDirectory(directory); err != nil {
			return closuregraph.PublicationReceipt{}, err
		}
	}
	expected := cacheInput.Expected
	cacheInputID, err := cacheInput.ID()
	if err != nil {
		return closuregraph.PublicationReceipt{}, err
	}
	cacheInputValue, err := cacheInput.canonicalValue()
	if err != nil {
		return closuregraph.PublicationReceipt{}, err
	}
	cacheInputBytes, err := protocoljson.MarshalCanonical(cacheInputValue)
	if err != nil {
		return closuregraph.PublicationReceipt{}, err
	}
	if err := authority.ValidateForPublication(expected, execution, observations); err != nil {
		return closuregraph.PublicationReceipt{}, err
	}
	actualWrites, err := walkOutputFiles(staging)
	if err != nil {
		return closuregraph.PublicationReceipt{}, err
	}
	if !reflect.DeepEqual(actualWrites, execution.WriteSet) {
		return closuregraph.PublicationReceipt{}, failure("closure_write_undeclared", "staging contains undeclared output paths")
	}
	expectedID, _ := expected.ID()
	executionID, _ := execution.ID()
	obs := append([]closuregraph.ProducedArtifactObservation(nil), observations...)
	sort.Slice(obs, func(i, j int) bool { a, _ := obs[i].ID(); b, _ := obs[j].ID(); return a < b })
	obsIDs := make([]closuregraph.ID, len(obs))
	paths := make([]string, len(obs))
	outputs := make([]storedOutput, len(obs))
	seenOutputs := map[closuregraph.ID]bool{}
	for i, o := range obs {
		oid, _ := o.ID()
		observationBytes, err := o.CanonicalBytes()
		if err != nil {
			return closuregraph.PublicationReceipt{}, err
		}
		obsIDs[i] = oid
		paths[i] = o.Path
		if seenOutputs[o.ExpectedOutputNodeID] {
			return closuregraph.PublicationReceipt{}, failure("artifact_local_output_drift", "duplicate expected output observation")
		}
		seenOutputs[o.ExpectedOutputNodeID] = true
		full, err := safeJoin(staging, o.Path)
		if err != nil {
			return closuregraph.PublicationReceipt{}, err
		}
		info, err := os.Lstat(full)
		if err != nil {
			return closuregraph.PublicationReceipt{}, err
		}
		if !info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink != 0 {
			return closuregraph.PublicationReceipt{}, failure("artifact_local_output_drift", "published output is not a regular file")
		}
		payload, err := os.ReadFile(full) // #nosec G304 -- full is resolved and containment-checked by safeJoin.
		if err != nil {
			return closuregraph.PublicationReceipt{}, err
		}
		if int64(len(payload)) != o.Size || digestBytes(payload) != o.SHA256 {
			return closuregraph.PublicationReceipt{}, failure("artifact_local_output_drift", "output bytes differ from observation")
		}
		if err = s.publishBlob(o.SHA256, payload, o.Class == "native.executable"); err != nil {
			return closuregraph.PublicationReceipt{}, err
		}
		outputs[i] = storedOutput{Path: o.Path, ObservationBytes: string(observationBytes), ObservationID: oid, SHA256: o.SHA256, Size: o.Size}
	}
	if !reflect.DeepEqual(sortedCopy(paths), execution.WriteSet) {
		return closuregraph.PublicationReceipt{}, failure("closure_write_undeclared", "observed outputs differ from execution write set")
	}
	if len(seenOutputs) != len(expected.ExpectedOutputNodeIDs) {
		return closuregraph.PublicationReceipt{}, failure("artifact_local_output_drift", "expected output set is incomplete")
	}
	for _, id := range expected.ExpectedOutputNodeIDs {
		if !seenOutputs[id] {
			return closuregraph.PublicationReceipt{}, failure("artifact_local_output_drift", "expected output has no observation")
		}
	}
	if !reflect.DeepEqual(obsIDs, execution.ProducedObservationIDs) {
		return closuregraph.PublicationReceipt{}, failure("artifact_local_output_drift", "execution observation set differs")
	}
	receipt := closuregraph.PublicationReceipt{SchemaID: closuregraph.SchemaPublicationReceipt, Decision: "published", ExecutionReceiptID: executionID, ExpectedCacheInputID: expectedID, ProtectedResult: "exact_write", PublishedObservationIDs: obsIDs}
	publicationBytes, err := receipt.CanonicalBytes()
	if err != nil {
		return closuregraph.PublicationReceipt{}, err
	}
	entry := storedEntry{SchemaID: "closure-protected-entry-v2", AssuredCacheInputID: cacheInputID, AssuredCacheInputBytes: string(cacheInputBytes), ExpectedCacheInputID: expectedID, ExecutionReceiptID: executionID, PublicationBytes: string(publicationBytes), Outputs: outputs}
	entryBytes, err := protocoljson.MarshalCanonical(entryValue(entry))
	if err != nil {
		return closuregraph.PublicationReceipt{}, err
	}
	if err = s.publishEntry(cacheInputID, entryBytes); err != nil {
		return closuregraph.PublicationReceipt{}, err
	}
	if _, err = s.Inspect(cacheInput); err != nil {
		return closuregraph.PublicationReceipt{}, err
	}
	return receipt, nil
}

// Inspect is read-only and returns only exact protected hits.
func (s *ProtectedStore) Inspect(cacheInput AssuredCacheInput) (PublishedHit, error) {
	if s == nil {
		return PublishedHit{}, fmt.Errorf("protected store is absent")
	}
	for _, directory := range []string{s.root, s.blobs, s.receipts} {
		if err := validatePrivateDirectory(directory); err != nil {
			return PublishedHit{}, err
		}
	}
	if err := cacheInput.Expected.Validate(); err != nil {
		return PublishedHit{}, err
	}
	expected := cacheInput.Expected
	expectedID, _ := expected.ID()
	id, err := cacheInput.ID()
	if err != nil {
		return PublishedHit{}, err
	}
	cacheInputValue, err := cacheInput.canonicalValue()
	if err != nil {
		return PublishedHit{}, err
	}
	cacheInputBytes, err := protocoljson.MarshalCanonical(cacheInputValue)
	if err != nil {
		return PublishedHit{}, err
	}
	entryPath := filepath.Join(s.receipts, strings.TrimPrefix(string(id), "sha256:")+".ccj.json")
	if err := protectedRegular(entryPath); err != nil {
		return PublishedHit{}, err
	}
	payload, err := os.ReadFile(entryPath) // #nosec G304 -- entryPath is derived from a validated content identity below the protected root.
	if err != nil {
		return PublishedHit{}, err
	}
	var raw storedEntry
	if err = protocoljson.UnmarshalCanonical(payload, &raw); err != nil {
		return PublishedHit{}, err
	}
	if raw.SchemaID != "closure-protected-entry-v2" || raw.AssuredCacheInputID != id || raw.AssuredCacheInputBytes != string(cacheInputBytes) || raw.ExpectedCacheInputID != expectedID {
		return PublishedHit{}, failure("artifact_local_output_drift", "protected entry input differs")
	}
	publication, err := closuregraph.DecodePublicationReceipt([]byte(raw.PublicationBytes))
	if err != nil {
		return PublishedHit{}, err
	}
	if publication.ExpectedCacheInputID != expectedID || publication.ExecutionReceiptID != raw.ExecutionReceiptID {
		return PublishedHit{}, failure("artifact_local_output_drift", "protected receipt references differ")
	}
	if len(raw.Outputs) != len(publication.PublishedObservationIDs) || len(raw.Outputs) != len(expected.ExpectedOutputNodeIDs) {
		return PublishedHit{}, failure("artifact_local_output_drift", "protected output cardinality differs")
	}
	paths := map[string]string{}
	observationIDs := make([]closuregraph.ID, len(raw.Outputs))
	outputNodeIDs := make([]closuregraph.ID, len(raw.Outputs))
	for i, o := range raw.Outputs {
		if i > 0 && raw.Outputs[i-1].ObservationID >= o.ObservationID {
			return PublishedHit{}, failure("artifact_local_output_drift", "stored outputs are not sorted")
		}
		if _, duplicate := paths[o.Path]; duplicate {
			return PublishedHit{}, failure("artifact_local_output_drift", "stored output paths are duplicated")
		}
		observation, decodeErr := closuregraph.DecodeProducedArtifactObservation([]byte(o.ObservationBytes))
		if decodeErr != nil {
			return PublishedHit{}, failure("artifact_local_output_drift", "stored observation is invalid: %v", decodeErr)
		}
		observationID, _ := observation.ID()
		if observationID != o.ObservationID || observation.Path != o.Path || observation.SHA256 != o.SHA256 || observation.Size != o.Size {
			return PublishedHit{}, failure("artifact_local_output_drift", "stored observation and output record differ")
		}
		observationIDs[i] = observationID
		outputNodeIDs[i] = observation.ExpectedOutputNodeID
		blob := s.blobPath(o.SHA256)
		if err = protectedRegular(blob); err != nil {
			return PublishedHit{}, err
		}
		info, err := os.Stat(blob)
		if err != nil || !blobModeMatchesClass(info, observation.Class == "native.executable") {
			return PublishedHit{}, failure("artifact_local_output_drift", "protected blob mode differs from output class")
		}
		b, err := os.ReadFile(blob) // #nosec G304 -- blob is derived from a validated content identity below the protected root.
		if err != nil {
			return PublishedHit{}, err
		}
		if int64(len(b)) != o.Size || digestBytes(b) != o.SHA256 {
			return PublishedHit{}, failure("artifact_local_output_drift", "protected blob differs")
		}
		paths[o.Path] = blob
	}
	if !reflect.DeepEqual(observationIDs, publication.PublishedObservationIDs) {
		return PublishedHit{}, failure("artifact_local_output_drift", "publication observation set differs from stored outputs")
	}
	sort.Slice(outputNodeIDs, func(i, j int) bool { return outputNodeIDs[i] < outputNodeIDs[j] })
	if !reflect.DeepEqual(outputNodeIDs, expected.ExpectedOutputNodeIDs) {
		return PublishedHit{}, failure("artifact_local_output_drift", "stored observations differ from expected output set")
	}
	return PublishedHit{Publication: publication, Paths: paths}, nil
}

func (s *ProtectedStore) publishBlob(id closuregraph.ID, payload []byte, executable bool) error {
	if !id.Valid() {
		return fmt.Errorf("invalid blob identity")
	}
	target := s.blobPath(id)
	mode := fs.FileMode(0o400)
	if executable {
		mode = 0o500
	}
	if err := writeNoReplace(s.blobs, target, payload, mode); err != nil {
		return err
	}
	if err := protectedRegular(target); err != nil {
		return err
	}
	existing, err := os.ReadFile(target) // #nosec G304 -- target is a validated digest path below the protected blob root.
	if err != nil {
		return err
	}
	if digestBytes(existing) != id {
		return failure("artifact_local_output_drift", "poisoned protected blob")
	}
	return nil
}
func (s *ProtectedStore) publishEntry(id closuregraph.ID, payload []byte) error {
	target := filepath.Join(s.receipts, strings.TrimPrefix(string(id), "sha256:")+".ccj.json")
	if err := writeNoReplace(s.receipts, target, payload, 0o400); err != nil {
		return err
	}
	existing, err := os.ReadFile(target) // #nosec G304 -- target is a validated expected-input path below the protected receipt root.
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(existing, payload) {
		return failure("artifact_local_output_drift", "poisoned protected cache entry")
	}
	return nil
}
func (s *ProtectedStore) blobPath(id closuregraph.ID) string {
	return filepath.Join(s.blobs, strings.TrimPrefix(string(id), "sha256:"))
}
func writeNoReplace(dir, target string, payload []byte, mode fs.FileMode) error {
	tmp, err := os.CreateTemp(dir, ".publish-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer func() { _ = os.Remove(name) }()
	if _, err = tmp.Write(payload); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return publishByLink(name, target, mode)
}
func protectedRegular(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0o222 != 0 {
		return failure("artifact_local_output_drift", "protected cache member is mutable or not regular")
	}
	return nil
}

func preparePrivateDirectory(path string) error {
	if _, err := os.Lstat(path); errors.Is(err, fs.ErrNotExist) {
		return privatedir.Make(path)
	} else if err != nil {
		return err
	}
	return validatePrivateDirectory(path)
}
func validatePrivateDirectory(path string) error {
	if _, err := os.Lstat(path); err != nil {
		return err
	}
	if privatedir.Validate(path) != nil {
		return failure("artifact_local_output_drift", "protected directory is pre-existing, mutable, or not owner-only")
	}
	return nil
}
func safeJoin(root, rel string) (string, error) {
	if err := portablePath(rel); err != nil {
		return "", err
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	full := filepath.Join(abs, filepath.FromSlash(rel))
	r, err := filepath.Rel(abs, full)
	if err != nil || r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes staging")
	}
	return full, nil
}
func walkOutputFiles(root string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(root, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == root {
			return nil
		}
		rel, err := filepath.Rel(root, current)
		if err != nil {
			return err
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return failure("artifact_local_output_drift", "output tree contains a link")
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return failure("artifact_local_output_drift", "output tree contains a special node")
		}
		paths = append(paths, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	return paths, nil
}
func entryValue(e storedEntry) map[string]any {
	outputs := make([]any, len(e.Outputs))
	for i, o := range e.Outputs {
		outputs[i] = map[string]any{"observation_bytes": o.ObservationBytes, "observation_id": string(o.ObservationID), "path": o.Path, "sha256": string(o.SHA256), "size": o.Size}
	}
	return map[string]any{"assured_cache_input_bytes": e.AssuredCacheInputBytes, "assured_cache_input_id": string(e.AssuredCacheInputID), "execution_receipt_id": string(e.ExecutionReceiptID), "expected_cache_input_id": string(e.ExpectedCacheInputID), "outputs": outputs, "publication_bytes": e.PublicationBytes, "schema_id": e.SchemaID}
}
