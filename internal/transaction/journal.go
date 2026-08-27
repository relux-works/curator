package transaction

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	pathpkg "path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"
)

func (engine *Engine) journalPath(transactionID string) string {
	return filepath.Join(engine.journalRoot, transactionID+".json")
}

func requireHomeLock(lock HomeLock) error {
	if lock == nil {
		return fmt.Errorf("manager-home lock is required")
	}
	value := reflect.ValueOf(lock)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		if value.IsNil() {
			return fmt.Errorf("manager-home lock is required")
		}
	}
	if err := lock.AssertHeld(); err != nil {
		return fmt.Errorf("manager-home lock is not held: %w", err)
	}
	return nil
}

func (engine *Engine) ensureJournalRoot() error {
	return makeDurableDirectories(engine.journalRoot)
}

func makeDurableDirectories(path string) error {
	info, err := os.Lstat(path)
	if err == nil {
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("transaction state path is not a safe directory: %s", path)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	parent := filepath.Dir(path)
	if parent == path {
		return fmt.Errorf("transaction state has no existing directory ancestor: %s", path)
	}
	if err := makeDurableDirectories(parent); err != nil {
		return err
	}
	if err := os.Mkdir(path, 0o700); err != nil && !os.IsExist(err) {
		return err
	}
	if err := syncDirectory(path); err != nil {
		return err
	}
	return syncDirectory(parent)
}

func (engine *Engine) saveJournal(journal *Journal) error {
	if err := engine.validateJournal(journal); err != nil {
		return err
	}
	if err := engine.ensureJournalRoot(); err != nil {
		return err
	}
	payload, err := marshalJournal(journal)
	if err != nil {
		return err
	}
	path := engine.journalPath(journal.TransactionID)
	if _, err := os.Lstat(path + ".delete"); err == nil {
		return corruptionf("journal cleanup tomb exists while saving transaction %s", journal.TransactionID)
	} else if !os.IsNotExist(err) {
		return err
	}
	temporary, err := os.CreateTemp(engine.journalRoot, ".journal-*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	ok := false
	defer func() {
		_ = temporary.Close()
		if !ok {
			_ = os.Remove(temporaryPath)
		}
	}()
	if err := temporary.Chmod(0o600); err != nil {
		return err
	}
	if _, err := temporary.Write(payload); err != nil {
		return err
	}
	if err := temporary.Sync(); err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if _, err := os.Lstat(path); err == nil {
		if err := durableReplaceFile(temporaryPath, path); err != nil {
			return err
		}
	} else if os.IsNotExist(err) {
		if err := durableRenameNoReplace(temporaryPath, path); err != nil {
			return err
		}
	} else {
		return err
	}
	ok = true
	return nil
}

func marshalJournal(journal *Journal) ([]byte, error) {
	payload, err := json.Marshal(journal)
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}

func (engine *Engine) loadJournal(transactionID string) (*Journal, error) {
	if !transactionIDPattern.MatchString(transactionID) {
		return nil, journalf("invalid transaction id")
	}
	path, tomb, err := engine.journalRecordPath(transactionID)
	if err != nil {
		return nil, err
	}
	before, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !before.Mode().IsRegular() || before.Mode()&os.ModeSymlink != 0 || before.Size() > 16*1024*1024 {
		return nil, journalf("journal %s is not a bounded regular file", transactionID)
	}
	file, err := os.Open(path) // #nosec G304 -- validated manager-owned journal path
	if err != nil {
		return nil, err
	}
	after, err := file.Stat()
	if err != nil || !os.SameFile(before, after) || !after.Mode().IsRegular() {
		_ = file.Close()
		return nil, journalf("journal %s changed while opening", transactionID)
	}
	payload, readErr := io.ReadAll(io.LimitReader(file, 16*1024*1024+1))
	closeErr := file.Close()
	if readErr != nil || closeErr != nil {
		return nil, errors.Join(readErr, closeErr)
	}
	if len(payload) > 16*1024*1024 {
		return nil, journalf("journal %s exceeds the size limit", transactionID)
	}
	decoder := json.NewDecoder(bufio.NewReader(bytes.NewReader(payload)))
	decoder.DisallowUnknownFields()
	var journal Journal
	if err := decoder.Decode(&journal); err != nil {
		return nil, journalf("decode %s: %v", transactionID, err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return nil, journalf("decode %s: %v", transactionID, err)
	}
	if err := engine.validateJournal(&journal); err != nil {
		return nil, err
	}
	if tomb && journal.Phase != PhaseCleanup && journal.Phase != PhaseRollbackCleanup {
		return nil, corruptionf("journal cleanup tomb for transaction %s has phase %q", transactionID, journal.Phase)
	}
	canonical, err := marshalJournal(&journal)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(payload, canonical) {
		return nil, journalf("journal %s is not canonical", transactionID)
	}
	return &journal, nil
}

func (engine *Engine) journalRecordPath(transactionID string) (string, bool, error) {
	path := engine.journalPath(transactionID)
	tomb := path + ".delete"
	_, pathErr := os.Lstat(path)
	_, tombErr := os.Lstat(tomb)
	pathExists := pathErr == nil
	tombExists := tombErr == nil
	if pathErr != nil && !os.IsNotExist(pathErr) {
		return "", false, pathErr
	}
	if tombErr != nil && !os.IsNotExist(tombErr) {
		return "", false, tombErr
	}
	if pathExists && tombExists {
		return "", false, corruptionf("transaction %s has both journal and cleanup tomb", transactionID)
	}
	if pathExists {
		return path, false, nil
	}
	if tombExists {
		return tomb, true, nil
	}
	return "", false, os.ErrNotExist
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return fmt.Errorf("multiple JSON values")
	}
	return err
}

func validateJournal(journal *Journal) error {
	if journal == nil || journal.Schema != journalSchema || !transactionIDPattern.MatchString(journal.TransactionID) {
		return journalf("schema or transaction id is invalid")
	}
	if journal.ProjectIdentity == "" || !validText(journal.ProjectIdentity) {
		return journalf("project identity is invalid")
	}
	switch journal.Phase {
	case PhasePreparing, PhasePrepared, PhaseCommitting, PhaseCleanup, PhaseRollingBack, PhaseRollbackCleanup:
	default:
		return journalf("phase %q is invalid", journal.Phase)
	}
	if journal.RemovalPath != "" {
		if journal.Phase != PhasePreparing && journal.Phase != PhaseCleanup && journal.Phase != PhaseRollbackCleanup {
			return journalf("removal progress is invalid during phase %q", journal.Phase)
		}
		candidate, found := findRemovalCandidate(journal, journal.RemovalPath)
		if !found {
			return journalf("removal path is not a canonical transaction sidecar")
		}
		if !validDigest(journal.RemovalDigest) || (journal.Phase != PhasePreparing && journal.RemovalDigest != candidate.digest) {
			return journalf("removal digest is invalid for phase %q", journal.Phase)
		}
		if err := validateRemovalEntries(journal.RemovalEntries); err != nil {
			return journalf("removal entries are not canonical: %v", err)
		}
	} else if journal.RemovalDigest != "" || len(journal.RemovalEntries) != 0 {
		return journalf("removal metadata exists without removal progress")
	}
	if !sortedUniqueStrings(journal.ReferencedBuildKeys) {
		return journalf("referenced build keys are not canonical")
	}
	if len(journal.Targets) == 0 {
		return journalf("target list is empty")
	}
	for index := range journal.Targets {
		target := &journal.Targets[index]
		if target.Class == "" || target.Identifier == "" || !validText(target.Class) || !validText(target.Identifier) {
			return journalf("target %d class or identifier is invalid", index)
		}
		if index > 0 && compareTargetRecords(journal.Targets[index-1], *target) >= 0 {
			return journalf("targets are not in strict canonical order")
		}
		if !filepath.IsAbs(target.LivePath) || !filepath.IsAbs(target.BackupPath) || !filepath.IsAbs(target.RollbackPath) || (target.StagedPath != "" && !filepath.IsAbs(target.StagedPath)) {
			return journalf("target %d paths are not absolute", index)
		}
		if filepath.Dir(target.LivePath) != filepath.Dir(target.BackupPath) || filepath.Dir(target.LivePath) != filepath.Dir(target.RollbackPath) || (target.StagedPath != "" && filepath.Dir(target.LivePath) != filepath.Dir(target.StagedPath)) {
			return journalf("target %d sidecars do not share the live parent", index)
		}
		expectedStaged := ""
		if target.DesiredDigest != DigestAbsent {
			expectedStaged = sidecarPath(target.LivePath, journal.TransactionID, index, "desired")
		}
		if target.StagedPath != expectedStaged || target.BackupPath != sidecarPath(target.LivePath, journal.TransactionID, index, "backup") || target.RollbackPath != sidecarPath(target.LivePath, journal.TransactionID, index, "rollback") {
			return journalf("target %d sidecars are not canonical", index)
		}
		if (target.PreimageDigest == "") == (target.ExpectedGeneration == "") {
			return journalf("target %d must have exactly one preimage expectation", index)
		}
		if target.PreimageDigest != "" && !validDigest(target.PreimageDigest) {
			return journalf("target %d preimage digest is invalid", index)
		}
		if target.ExpectedGeneration != "" && !validText(target.ExpectedGeneration) {
			return journalf("target %d expected generation is invalid", index)
		}
		if target.GenerationPath != "" && (filepath.IsAbs(target.GenerationPath) || filepath.Clean(target.GenerationPath) == ".." || strings.HasPrefix(filepath.Clean(target.GenerationPath), ".."+string(filepath.Separator))) {
			return journalf("target %d generation path escapes the live target", index)
		}
		if !validDigest(target.DesiredDigest) || (target.DesiredDigest == DigestAbsent) != (target.StagedPath == "") {
			return journalf("target %d desired digest or staging path is invalid", index)
		}
		if journal.Phase == PhasePreparing && target.StagedPath != "" {
			if !filepath.IsAbs(target.StagedSource) || filepath.Clean(target.StagedSource) != target.StagedSource {
				return journalf("target %d staged source is invalid", index)
			}
			if err := validateRemovalEntries(target.StagingEntries); err != nil {
				return journalf("target %d staging manifest is invalid: %v", index, err)
			}
			if target.StagingIndex < 0 || target.StagingIndex > len(target.StagingEntries) || (target.StagingActive && target.StagingIndex == len(target.StagingEntries)) || (target.StagingCreated && !target.StagingActive) || (target.StagingDiscarded && target.StagingActive) {
				return journalf("target %d staging progress is invalid", index)
			}
			if target.StagingBytes < 0 || (target.StagingBytes == 0) != (target.StagingPrefixDigest == "") {
				return journalf("target %d staging byte progress is invalid", index)
			}
			if target.StagingWriteBytes < 0 || (target.StagingWriteBytes == 0) != (target.StagingWriteDigest == "") || (target.StagingWriteBytes != 0 && (target.StagingWriteBytes <= target.StagingBytes || target.StagingWriteBytes-target.StagingBytes > stagingCopyChunkSize)) {
				return journalf("target %d staging write-ahead progress is invalid", index)
			}
			if target.StagingBytes > 0 {
				if !target.StagingActive || !target.StagingCreated || target.StagingPrefixDigest == DigestAbsent || !validDigest(target.StagingPrefixDigest) || target.StagingEntries[target.StagingIndex].Kind != "file" {
					return journalf("target %d staging prefix progress is invalid", index)
				}
			}
			if target.StagingWriteBytes > 0 {
				if !target.StagingActive || !target.StagingCreated || target.StagingWriteDigest == DigestAbsent || !validDigest(target.StagingWriteDigest) || target.StagingEntries[target.StagingIndex].Kind != "file" {
					return journalf("target %d staging write-ahead progress is invalid", index)
				}
			}
		} else if target.StagedSource != "" || len(target.StagingEntries) != 0 || target.StagingIndex != 0 || target.StagingActive || target.StagingCreated || target.StagingBytes != 0 || target.StagingPrefixDigest != "" || target.StagingWriteBytes != 0 || target.StagingWriteDigest != "" || target.StagingDiscarded {
			return journalf("target %d has staging progress outside preparation", index)
		}
		if target.BackupDigest != "" && !validDigest(target.BackupDigest) {
			return journalf("target %d backup digest is invalid", index)
		}
		switch target.State {
		case StatePending, StateBackedUp, StateCommitted, StateRolledBack:
		default:
			return journalf("target %d state %q is invalid", index, target.State)
		}
	}
	if err := validateIndependentTargetNamespaces(journal.Targets); err != nil {
		return journalf("target namespaces are not independent: %v", err)
	}
	return nil
}

func (engine *Engine) validateJournal(journal *Journal) error {
	if err := validateJournal(journal); err != nil {
		return err
	}
	if err := validateIndependentTargetNamespaces(journal.Targets, targetNamespacePath{
		owner: "engine",
		kind:  "journal namespace",
		path:  engine.journalRoot,
	}); err != nil {
		return journalf("target namespaces overlap manager transaction state: %v", err)
	}
	return nil
}

func validText(value string) bool {
	return utf8.ValidString(value) && !strings.ContainsRune(value, 0)
}

func validDigest(value string) bool {
	if value == DigestAbsent {
		return true
	}
	if len(value) != len("sha256:")+64 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	for _, value := range value[len("sha256:"):] {
		if (value < '0' || value > '9') && (value < 'a' || value > 'f') {
			return false
		}
	}
	return true
}

func sortedUniqueStrings(values []string) bool {
	for index, value := range values {
		if value == "" || !validText(value) || (index > 0 && strings.Compare(values[index-1], value) >= 0) {
			return false
		}
	}
	return true
}

func validateRemovalEntries(entries []RemovalEntry) error {
	if len(entries) == 0 || entries[0].RelativePath != "" {
		return fmt.Errorf("root entry is missing")
	}
	byPath := make(map[string]RemovalEntry, len(entries))
	for index, entry := range entries {
		if !validText(entry.RelativePath) || (index > 0 && bytes.Compare([]byte(entries[index-1].RelativePath), []byte(entry.RelativePath)) >= 0) {
			return fmt.Errorf("entry paths are not in strict unsigned-byte order")
		}
		if entry.RelativePath != "" {
			if pathpkg.IsAbs(entry.RelativePath) || pathpkg.Clean(entry.RelativePath) != entry.RelativePath || entry.RelativePath == ".." || strings.HasPrefix(entry.RelativePath, "../") {
				return fmt.Errorf("entry path %q is not a safe canonical relative path", entry.RelativePath)
			}
		}
		if entry.Mode > 0o777 {
			return fmt.Errorf("entry %q mode is invalid", entry.RelativePath)
		}
		switch entry.Kind {
		case "directory":
			if entry.Digest != "" {
				return fmt.Errorf("directory entry %q has a digest", entry.RelativePath)
			}
		case "file":
			if !validDigest(entry.Digest) || entry.Digest == DigestAbsent {
				return fmt.Errorf("file entry %q digest is invalid", entry.RelativePath)
			}
		default:
			return fmt.Errorf("entry %q kind %q is invalid", entry.RelativePath, entry.Kind)
		}
		byPath[entry.RelativePath] = entry
	}
	for _, entry := range entries[1:] {
		parent := pathpkg.Dir(entry.RelativePath)
		if parent == "." {
			parent = ""
		}
		parentEntry, found := byPath[parent]
		if !found || parentEntry.Kind != "directory" {
			return fmt.Errorf("entry %q has no recorded directory parent", entry.RelativePath)
		}
	}
	return nil
}

func (engine *Engine) journalIDs() ([]string, error) {
	entries, err := os.ReadDir(engine.journalRoot)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	seen := make(map[string]string, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		suffix := ""
		switch {
		case strings.HasSuffix(name, ".json.delete"):
			suffix = ".json.delete"
		case strings.HasSuffix(name, ".json"):
			suffix = ".json"
		default:
			continue
		}
		id := strings.TrimSuffix(name, suffix)
		if !transactionIDPattern.MatchString(id) {
			return nil, journalf("unexpected journal filename %q", name)
		}
		if prior, exists := seen[id]; exists {
			return nil, corruptionf("transaction %s has multiple journal records %q and %q", id, prior, name)
		}
		seen[id] = name
	}
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return bytes.Compare([]byte(ids[i]), []byte(ids[j])) < 0 })
	return ids, nil
}

func removeDurably(path, expectedDigest string) error {
	return removeDurablyUnrecorded(path, expectedDigest, nil, nil)
}

func captureRemovalEntries(path, expectedDigest string) ([]RemovalEntry, error) {
	if err := verifyRemovalDigest(path, expectedDigest); err != nil {
		return nil, err
	}
	var entries []RemovalEntry
	err := filepath.Walk(path, func(current string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.Mode()&os.ModeSymlink != 0 || (!info.Mode().IsRegular() && !info.IsDir()) {
			return corruptionf("durable removal refused unsafe entry %s", current)
		}
		relative, err := filepath.Rel(path, current)
		if err != nil {
			return err
		}
		if relative == "." {
			relative = ""
		} else {
			relative = filepath.ToSlash(relative)
		}
		entry := RemovalEntry{RelativePath: relative, Mode: uint32(info.Mode().Perm())}
		if info.IsDir() {
			entry.Kind = "directory"
		} else {
			entry.Kind = "file"
			entry.Digest, err = DigestPath(current)
			if err != nil {
				return err
			}
		}
		entries = append(entries, entry)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool {
		return bytes.Compare([]byte(entries[i].RelativePath), []byte(entries[j].RelativePath)) < 0
	})
	if err := validateRemovalEntries(entries); err != nil {
		return nil, corruptionf("durable removal manifest for %s is invalid: %v", path, err)
	}
	if err := verifyRemovalDigest(path, expectedDigest); err != nil {
		return nil, err
	}
	return entries, nil
}

func validateRemovalStart(path, expectedDigest string) (bool, error) {
	if expectedDigest == "" {
		expectedDigest = DigestAbsent
	}
	if !validDigest(expectedDigest) {
		return false, corruptionf("durable removal for %s has invalid expected digest", path)
	}
	tomb := path + ".delete"
	_, pathErr := os.Lstat(path)
	_, tombErr := os.Lstat(tomb)
	pathExists := pathErr == nil
	tombExists := tombErr == nil
	if pathErr != nil && !os.IsNotExist(pathErr) {
		return false, pathErr
	}
	if tombErr != nil && !os.IsNotExist(tombErr) {
		return false, tombErr
	}
	if pathExists && tombExists {
		return false, corruptionf("durable removal refused simultaneous original and tomb for %s", path)
	}
	if tombExists {
		return false, corruptionf("durable removal refused unowned tomb for %s", path)
	}
	if !pathExists {
		return false, nil
	}
	if err := verifyRemovalDigest(path, expectedDigest); err != nil {
		return false, err
	}
	return true, nil
}

func removeDurablyUnrecorded(path, expectedDigest string, afterRename, afterRemoval func() error) error {
	if expectedDigest == "" {
		expectedDigest = DigestAbsent
	}
	if !validDigest(expectedDigest) {
		return corruptionf("durable removal for %s has invalid expected digest", path)
	}
	tomb := path + ".delete"
	_, pathErr := os.Lstat(path)
	_, tombErr := os.Lstat(tomb)
	pathExists := pathErr == nil
	tombExists := tombErr == nil
	if pathErr != nil && !os.IsNotExist(pathErr) {
		return pathErr
	}
	if tombErr != nil && !os.IsNotExist(tombErr) {
		return tombErr
	}
	if pathExists && tombExists {
		return corruptionf("durable removal refused simultaneous original and tomb for %s", path)
	}
	if tombExists {
		if err := verifyRemovalDigest(tomb, expectedDigest); err != nil {
			return err
		}
	}
	if pathExists {
		if err := verifyRemovalDigest(path, expectedDigest); err != nil {
			return err
		}
		if err := durableRenameNoReplace(path, tomb); err != nil {
			return err
		}
		if err := verifyRemovalDigest(tomb, expectedDigest); err != nil {
			return err
		}
		if afterRename != nil {
			if err := afterRename(); err != nil {
				return err
			}
		}
		tombExists = true
	}
	if !tombExists {
		return nil
	}
	if _, err := os.Lstat(path); err == nil {
		return corruptionf("durable removal refused concurrent original and tomb for %s", path)
	} else if !os.IsNotExist(err) {
		return err
	}
	return removeTreeDurably(tomb, afterRemoval)
}

func removeTreeDurably(path string, afterRemoval func() error) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || (!info.Mode().IsRegular() && !info.IsDir()) {
		return corruptionf("durable removal refused unsafe tomb entry %s", path)
	}
	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if err := removeTreeDurably(filepath.Join(path, entry.Name()), afterRemoval); err != nil {
				return err
			}
		}
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	if err := syncDirectory(filepath.Dir(path)); err != nil {
		return err
	}
	if afterRemoval != nil {
		return afterRemoval()
	}
	return nil
}

func removeDurablyWithEntries(path, expectedDigest string, entries []RemovalEntry, afterRename, afterRemoval func() error) error {
	if err := validateRemovalEntries(entries); err != nil {
		return corruptionf("durable removal for %s has invalid manifest: %v", path, err)
	}
	tomb := path + ".delete"
	_, pathErr := os.Lstat(path)
	_, tombErr := os.Lstat(tomb)
	pathExists := pathErr == nil
	tombExists := tombErr == nil
	if pathErr != nil && !os.IsNotExist(pathErr) {
		return pathErr
	}
	if tombErr != nil && !os.IsNotExist(tombErr) {
		return tombErr
	}
	if pathExists && tombExists {
		return corruptionf("durable removal refused simultaneous original and tomb for %s", path)
	}
	if pathExists {
		if err := verifyRemovalDigest(path, expectedDigest); err != nil {
			return err
		}
		if err := validateRemovalTree(path, entries, true); err != nil {
			return err
		}
		if err := durableRenameNoReplace(path, tomb); err != nil {
			return err
		}
		if err := validateRemovalTree(tomb, entries, true); err != nil {
			return err
		}
		if afterRename != nil {
			if err := afterRename(); err != nil {
				return err
			}
		}
		tombExists = true
	}
	if !tombExists {
		return nil
	}
	if _, err := os.Lstat(path); err == nil {
		return corruptionf("durable removal refused concurrent original and tomb for %s", path)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := validateRemovalTree(tomb, entries, false); err != nil {
		return err
	}
	for index := len(entries) - 1; index >= 0; index-- {
		entry := entries[index]
		current := tomb
		if entry.RelativePath != "" {
			current = filepath.Join(tomb, filepath.FromSlash(entry.RelativePath))
		}
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if err := verifyRemovalEntry(current, info, entry); err != nil {
			return err
		}
		if err := os.Remove(current); err != nil {
			if entry.Kind == "directory" {
				return corruptionf("durable removal refused nonempty recorded directory %s: %v", current, err)
			}
			return err
		}
		if err := syncDirectory(filepath.Dir(current)); err != nil {
			return err
		}
		if afterRemoval != nil {
			if err := afterRemoval(); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateRemovalTree(root string, entries []RemovalEntry, requireComplete bool) error {
	expected := make(map[string]RemovalEntry, len(entries))
	for _, entry := range entries {
		expected[entry.RelativePath] = entry
	}
	seen := make(map[string]struct{}, len(entries))
	err := filepath.Walk(root, func(current string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, current)
		if err != nil {
			return err
		}
		if relative == "." {
			relative = ""
		} else {
			relative = filepath.ToSlash(relative)
		}
		entry, found := expected[relative]
		if !found {
			return corruptionf("durable removal refused unrecorded tomb entry %s", current)
		}
		if err := verifyRemovalEntry(current, info, entry); err != nil {
			return err
		}
		seen[relative] = struct{}{}
		return nil
	})
	if err != nil {
		return err
	}
	if requireComplete && len(seen) != len(entries) {
		return corruptionf("durable removal refused incomplete original tree %s", root)
	}
	return nil
}

func verifyRemovalEntry(path string, info os.FileInfo, entry RemovalEntry) error {
	if info.Mode()&os.ModeSymlink != 0 || uint32(info.Mode().Perm()) != entry.Mode {
		return corruptionf("durable removal refused changed tomb entry %s", path)
	}
	switch entry.Kind {
	case "directory":
		if !info.IsDir() {
			return corruptionf("durable removal refused changed tomb entry %s", path)
		}
	case "file":
		if !info.Mode().IsRegular() {
			return corruptionf("durable removal refused changed tomb entry %s", path)
		}
		digest, err := DigestPath(path)
		if err != nil || digest != entry.Digest {
			return corruptionf("durable removal refused changed tomb entry %s", path)
		}
	default:
		return corruptionf("durable removal refused unknown tomb entry kind %q", entry.Kind)
	}
	return nil
}

func verifyRemovalDigest(path, expectedDigest string) error {
	digest, err := DigestPath(path)
	if err != nil {
		return corruptionf("durable removal cannot verify %s: %v", path, err)
	}
	if digest != expectedDigest {
		return corruptionf("durable removal refused %s digest %s, want %s", path, digest, expectedDigest)
	}
	return nil
}

func journalDigest(journal *Journal) (string, error) {
	payload, err := marshalJournal(journal)
	if err != nil {
		return "", err
	}
	return digestRegularPayload(payload, journalMode())
}

func (engine *Engine) removeJournalDurably(journal *Journal) error {
	digest, err := journalDigest(journal)
	if err != nil {
		return err
	}
	return removeDurably(engine.journalPath(journal.TransactionID), digest)
}

func compareTargetRecords(left, right TargetRecord) int {
	if compared := bytes.Compare([]byte(left.Class), []byte(right.Class)); compared != 0 {
		return compared
	}
	return bytes.Compare([]byte(left.Identifier), []byte(right.Identifier))
}
