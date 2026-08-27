package transaction

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
)

const stagingCopyChunkSize = 32 * 1024

var stagingPrefixDigestDomain = []byte("curator-transaction-staging-prefix-v1\x00")

func (engine *Engine) stageTarget(journal *Journal, targetIndex int) error {
	target := &journal.Targets[targetIndex]
	for target.StagingIndex < len(target.StagingEntries) {
		entry := target.StagingEntries[target.StagingIndex]
		target.StagingActive = true
		target.StagingCreated = false
		target.StagingBytes = 0
		target.StagingPrefixDigest = ""
		target.StagingWriteBytes = 0
		target.StagingWriteDigest = ""
		if err := engine.saveJournal(journal); err != nil {
			return err
		}
		if err := engine.createStagingEntry(journal, targetIndex, entry); err != nil {
			return err
		}
		target.StagingCreated = true
		if err := engine.saveJournal(journal); err != nil {
			return err
		}
		if entry.Kind == "file" {
			if err := engine.copyStagingFile(journal, targetIndex, entry); err != nil {
				return err
			}
		}
		path := stagingEntryPath(target.StagedPath, entry.RelativePath)
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if err := verifyRemovalEntry(path, info, entry); err != nil {
			return err
		}
		target.StagingIndex++
		target.StagingActive = false
		target.StagingCreated = false
		target.StagingBytes = 0
		target.StagingPrefixDigest = ""
		target.StagingWriteBytes = 0
		target.StagingWriteDigest = ""
		if err := engine.saveJournal(journal); err != nil {
			return err
		}
	}
	return nil
}

func (engine *Engine) createStagingEntry(journal *Journal, targetIndex int, entry RemovalEntry) error {
	target := &journal.Targets[targetIndex]
	destination := stagingEntryPath(target.StagedPath, entry.RelativePath)
	source := stagingEntryPath(target.StagedSource, entry.RelativePath)
	if _, err := os.Lstat(destination); err == nil {
		return corruptionf("staging entry already exists before creation: %s", destination)
	} else if !os.IsNotExist(err) {
		return err
	}
	switch entry.Kind {
	case "link":
		// A link is created complete in one step: its destination string is its
		// whole content, so there is no partial state to write ahead.
		if err := os.Symlink(entry.LinkTarget, destination); err != nil {
			return err
		}
	case "directory":
		if err := os.Mkdir(destination, 0); err != nil {
			return err
		}
		if err := os.Chmod(destination, os.FileMode(entry.Mode)); err != nil {
			return err
		}
		if err := syncDirectory(destination); err != nil {
			return err
		}
	case "file":
		info, err := os.Lstat(source)
		if err != nil {
			return err
		}
		if err := verifyRemovalEntry(source, info, entry); err != nil {
			return corruptionf("staging source entry changed before copy: %s", source)
		}
		file, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0) // #nosec G304 -- deterministic transaction sidecar
		if err != nil {
			return err
		}
		if err := file.Chmod(os.FileMode(entry.Mode)); err != nil {
			_ = file.Close()
			return err
		}
		if err := file.Sync(); err != nil {
			_ = file.Close()
			return err
		}
		if err := file.Close(); err != nil {
			return err
		}
	default:
		return journalf("unknown staging entry kind %q", entry.Kind)
	}
	return syncDirectory(filepath.Dir(destination))
}

func (engine *Engine) copyStagingFile(journal *Journal, targetIndex int, entry RemovalEntry) (result error) {
	target := &journal.Targets[targetIndex]
	sourcePath := stagingEntryPath(target.StagedSource, entry.RelativePath)
	destinationPath := stagingEntryPath(target.StagedPath, entry.RelativePath)
	source, err := os.Open(sourcePath) // #nosec G304 -- caller-selected private staging source
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, source.Close()) }()
	destination, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_APPEND, 0) // #nosec G304 -- transaction-owned sidecar
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, destination.Close()) }()
	buffer := make([]byte, stagingCopyChunkSize)
	prefix := sha256.New()
	_, _ = prefix.Write(stagingPrefixDigestDomain)
	for {
		count, readErr := source.Read(buffer)
		if count > 0 {
			_, _ = prefix.Write(buffer[:count])
			target.StagingWriteBytes = target.StagingBytes + int64(count)
			target.StagingWriteDigest = "sha256:" + hex.EncodeToString(prefix.Sum(nil))
			if err := engine.saveJournal(journal); err != nil {
				return err
			}
			written, err := destination.Write(buffer[:count])
			if err != nil {
				return err
			}
			if written != count {
				return io.ErrShortWrite
			}
			if err := destination.Sync(); err != nil {
				return err
			}
			if err := engine.emit(journal, targetIndex, PointAfterStagingChunkSync); err != nil {
				return err
			}
			target.StagingBytes = target.StagingWriteBytes
			target.StagingPrefixDigest = target.StagingWriteDigest
			target.StagingWriteBytes = 0
			target.StagingWriteDigest = ""
			if err := engine.saveJournal(journal); err != nil {
				return err
			}
			if err := engine.emit(journal, targetIndex, PointDuringStagingCopy); err != nil {
				return err
			}
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	return nil
}

func (engine *Engine) discardPreparing(journal *Journal) error {
	if journal.Phase != PhasePreparing {
		return journalf("discard preparation in phase %q", journal.Phase)
	}
	if err := engine.resumeRecordedRemoval(journal); err != nil {
		return err
	}
	for index := len(journal.Targets) - 1; index >= 0; index-- {
		target := &journal.Targets[index]
		if target.StagedPath == "" {
			continue
		}
		if target.StagingDiscarded {
			if digest, err := target.digest(target.StagedPath); err != nil {
				return err
			} else if digest != DigestAbsent {
				return corruptionf("discarded preparation staging reappeared: %s", target.StagedPath)
			}
			continue
		}
		if err := validatePreparingTarget(target); err != nil {
			return err
		}
		digest, err := target.digest(target.StagedPath)
		if err != nil {
			return err
		}
		if digest == DigestAbsent {
			continue
		}
		entries, err := captureRemovalEntries(target.Kind, target.StagedPath, digest)
		if err != nil {
			return err
		}
		if err := validatePreparingTarget(target); err != nil {
			return err
		}
		journal.RemovalPath = target.StagedPath
		journal.RemovalDigest = digest
		journal.RemovalEntries = entries
		if err := engine.saveJournal(journal); err != nil {
			journal.RemovalPath = ""
			journal.RemovalDigest = ""
			journal.RemovalEntries = nil
			return err
		}
		if err := engine.finishRecordedRemoval(journal, removalCandidate{
			targetIndex: index, kind: target.Kind, path: target.StagedPath, digest: digest,
		}); err != nil {
			return err
		}
	}
	for index := range journal.Targets {
		target := &journal.Targets[index]
		target.StagedSource = ""
		target.StagingEntries = nil
		target.StagingIndex = 0
		target.StagingActive = false
		target.StagingCreated = false
		target.StagingBytes = 0
		target.StagingPrefixDigest = ""
		target.StagingWriteBytes = 0
		target.StagingWriteDigest = ""
		target.StagingDiscarded = false
	}
	journal.Phase = PhaseRollbackCleanup
	if err := engine.saveJournal(journal); err != nil {
		return err
	}
	return engine.rollback(journal)
}

func validatePreparingTarget(target *TargetRecord) error {
	if target.StagedPath == "" {
		return nil
	}
	if target.StagingIndex < 0 || target.StagingIndex > len(target.StagingEntries) || (target.StagingActive && target.StagingIndex == len(target.StagingEntries)) {
		return journalf("staging progress for %s/%s is invalid", target.Class, target.Identifier)
	}
	positions := make(map[string]int, len(target.StagingEntries))
	for index, entry := range target.StagingEntries {
		positions[entry.RelativePath] = index
	}
	seen := make(map[int]struct{}, target.StagingIndex+1)
	_, rootErr := os.Lstat(target.StagedPath)
	if os.IsNotExist(rootErr) {
		if target.StagingIndex != 0 || target.StagingCreated || target.StagingBytes != 0 || target.StagingPrefixDigest != "" || target.StagingWriteBytes != 0 || target.StagingWriteDigest != "" {
			return corruptionf("preparing target %s/%s lost recorded staging", target.Class, target.Identifier)
		}
		return nil
	}
	if rootErr != nil {
		return rootErr
	}
	err := filepath.Walk(target.StagedPath, func(current string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(target.StagedPath, current)
		if err != nil {
			return err
		}
		if relative == "." {
			relative = ""
		} else {
			relative = filepath.ToSlash(relative)
		}
		position, found := positions[relative]
		if !found {
			return corruptionf("preparing target %s/%s contains unrecorded staging entry %s", target.Class, target.Identifier, current)
		}
		entry := target.StagingEntries[position]
		switch {
		case position < target.StagingIndex:
			if err := verifyRemovalEntry(current, info, entry); err != nil {
				return err
			}
		case position == target.StagingIndex && target.StagingActive:
			if err := validateActiveStagingEntry(target, current, info, entry); err != nil {
				return err
			}
		default:
			return corruptionf("preparing target %s/%s contains staging beyond durable progress: %s", target.Class, target.Identifier, current)
		}
		seen[position] = struct{}{}
		return nil
	})
	if err != nil {
		return err
	}
	for index := 0; index < target.StagingIndex; index++ {
		if _, found := seen[index]; !found {
			return corruptionf("preparing target %s/%s is missing completed staging entry %q", target.Class, target.Identifier, target.StagingEntries[index].RelativePath)
		}
	}
	if target.StagingActive && target.StagingCreated {
		if _, found := seen[target.StagingIndex]; !found {
			return corruptionf("preparing target %s/%s is missing active staging entry", target.Class, target.Identifier)
		}
	}
	return nil
}

func validateActiveStagingEntry(target *TargetRecord, path string, info os.FileInfo, entry RemovalEntry) (result error) {
	if entry.Kind == "link" {
		// A link staging entry is created atomically, so the only state it can
		// durably be in is exactly the recorded one, with no byte progress.
		if err := verifyRemovalEntry(path, info, entry); err != nil {
			return corruptionf("preparing target contains changed active staging link %s", path)
		}
		if target.StagingBytes != 0 || target.StagingPrefixDigest != "" ||
			target.StagingWriteBytes != 0 || target.StagingWriteDigest != "" {
			return corruptionf("preparing target records byte progress for staging link %s", path)
		}
		return nil
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return corruptionf("preparing target contains unsafe active staging entry %s", path)
	}
	if entry.Kind == "directory" {
		if !info.IsDir() || (target.StagingCreated && uint32(info.Mode().Perm()) != entry.Mode) {
			return corruptionf("preparing target contains changed active staging directory %s", path)
		}
		return nil
	}
	if entry.Kind != "file" || !info.Mode().IsRegular() {
		return corruptionf("preparing target contains changed active staging entry %s", path)
	}
	if !target.StagingCreated {
		if info.Size() != 0 || target.StagingBytes != 0 || target.StagingPrefixDigest != "" {
			return corruptionf("preparing target contains bytes before active staging ownership %s", path)
		}
		return nil
	}
	if uint32(info.Mode().Perm()) != entry.Mode {
		return corruptionf("preparing target contains changed active staging mode %s", path)
	}
	source := stagingEntryPath(target.StagedSource, entry.RelativePath)
	sourceInfo, err := os.Lstat(source)
	if err != nil {
		return corruptionf("cannot verify partial staging source %s: %v", source, err)
	}
	if err := verifyRemovalEntry(source, sourceInfo, entry); err != nil {
		return corruptionf("partial staging source changed: %s", source)
	}
	if target.StagingBytes < 0 || target.StagingBytes > sourceInfo.Size() {
		return corruptionf("recorded partial staging size is invalid: %s", path)
	}
	if target.StagingWriteBytes > sourceInfo.Size() {
		return corruptionf("recorded staging write-ahead size is invalid: %s", path)
	}
	if target.StagingWriteBytes != 0 {
		expectedWriteBytes := target.StagingBytes + stagingCopyChunkSize
		if expectedWriteBytes > sourceInfo.Size() {
			expectedWriteBytes = sourceInfo.Size()
		}
		if target.StagingWriteBytes != expectedWriteBytes {
			return corruptionf("staging write-ahead range is not canonical: %s", path)
		}
	}
	maximumSize := target.StagingBytes
	if target.StagingWriteBytes != 0 {
		maximumSize = target.StagingWriteBytes
	}
	if info.Size() < target.StagingBytes || info.Size() > maximumSize {
		return corruptionf("partial staging size is outside durable ownership: %s", path)
	}
	if target.StagingWriteBytes != 0 {
		pendingSource, err := os.Open(source) // #nosec G304 -- recorded staging source
		if err != nil {
			return err
		}
		pendingDigest, digestErr := stagingPrefixDigest(io.LimitReader(pendingSource, target.StagingWriteBytes))
		closeErr := pendingSource.Close()
		if err := errors.Join(digestErr, closeErr); err != nil {
			return err
		}
		if pendingDigest != target.StagingWriteDigest {
			return corruptionf("staging write-ahead digest changed from source: %s", path)
		}
	}
	if info.Size() == 0 {
		if target.StagingBytes != 0 || target.StagingPrefixDigest != "" {
			return corruptionf("empty partial staging disagrees with durable progress: %s", path)
		}
		return nil
	}
	stagedFile, err := os.Open(path) // #nosec G304 -- validated transaction sidecar
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, stagedFile.Close()) }()
	sourceFile, err := os.Open(source) // #nosec G304 -- recorded staging source
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, sourceFile.Close()) }()
	stagedDigest, err := stagingPrefixDigest(stagedFile)
	if err != nil {
		return err
	}
	sourceDigest, err := stagingPrefixDigest(io.LimitReader(sourceFile, info.Size()))
	if err != nil {
		return err
	}
	if stagedDigest != sourceDigest {
		return corruptionf("partial staging bytes changed from durable source prefix: %s", path)
	}
	if info.Size() == target.StagingBytes && stagedDigest != target.StagingPrefixDigest {
		return corruptionf("partial staging bytes changed from durable progress: %s", path)
	}
	return nil
}

func stagingPrefixDigest(reader io.Reader) (string, error) {
	hash := sha256.New()
	_, _ = hash.Write(stagingPrefixDigestDomain)
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}

func stagingEntryPath(root, relative string) string {
	if relative == "" {
		return root
	}
	return filepath.Join(root, filepath.FromSlash(relative))
}
