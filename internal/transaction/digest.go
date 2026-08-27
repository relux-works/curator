package transaction

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

var digestDomain = []byte("curator-transaction-target-v1\x00")

// DigestPath returns a deterministic digest for one regular file or safe
// directory tree. Links and special files are rejected. Absence is explicit.
func DigestPath(path string) (string, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return DigestAbsent, nil
	}
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	_, _ = hash.Write(digestDomain)
	if info.Mode().IsRegular() {
		if err := digestFile(hash, path, "", info); err != nil {
			return "", err
		}
	} else if info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
		if err := digestDirectory(hash, path, info); err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("unsafe transaction target type at %s", path)
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}

func digestRegularPayload(payload []byte, mode os.FileMode) (string, error) {
	hash := sha256.New()
	_, _ = hash.Write(digestDomain)
	if err := writeDigestEntry(hash, 'f', "", mode.Perm(), uint64(len(payload))); err != nil {
		return "", err
	}
	_, _ = hash.Write(payload)
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}

func digestDirectory(hash io.Writer, root string, rootInfo os.FileInfo) error {
	if err := writeDigestEntry(hash, 'd', "", rootInfo.Mode().Perm(), 0); err != nil {
		return err
	}
	var paths []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 || (!info.IsDir() && !info.Mode().IsRegular()) {
			return fmt.Errorf("unsafe transaction tree entry at %s", path)
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return err
	}
	sort.Slice(paths, func(i, j int) bool { return bytes.Compare([]byte(paths[i]), []byte(paths[j])) < 0 })
	for _, rel := range paths {
		path := filepath.Join(root, filepath.FromSlash(rel))
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			if err := writeDigestEntry(hash, 'd', rel, info.Mode().Perm(), 0); err != nil {
				return err
			}
			continue
		}
		if err := digestFile(hash, path, rel, info); err != nil {
			return err
		}
	}
	after, err := os.Lstat(root)
	if err != nil {
		return err
	}
	if !after.IsDir() || after.Mode()&os.ModeSymlink != 0 || !os.SameFile(rootInfo, after) || after.Mode().Perm() != rootInfo.Mode().Perm() {
		return fmt.Errorf("transaction target changed while digesting: %s", root)
	}
	return nil
}

func digestFile(hash io.Writer, path, rel string, info os.FileInfo) (result error) {
	file, err := os.Open(path) // #nosec G304 -- caller-selected transaction target
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, file.Close()) }()
	if info.Size() < 0 {
		return fmt.Errorf("transaction target has a negative size: %s", path)
	}
	// #nosec G115 -- the negative range was rejected immediately above.
	if err := writeDigestEntry(hash, 'f', rel, info.Mode().Perm(), uint64(info.Size())); err != nil {
		return err
	}
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}
	after, err := file.Stat()
	if err != nil {
		return err
	}
	if !after.Mode().IsRegular() || after.Size() != info.Size() || after.Mode().Perm() != info.Mode().Perm() {
		return fmt.Errorf("transaction target changed while digesting: %s", path)
	}
	return nil
}

func writeDigestEntry(writer io.Writer, kind byte, rel string, mode os.FileMode, size uint64) error {
	if _, err := writer.Write([]byte{kind}); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.BigEndian, uint64(len([]byte(rel)))); err != nil {
		return err
	}
	if _, err := writer.Write([]byte(rel)); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.BigEndian, uint32(mode.Perm())); err != nil {
		return err
	}
	return binary.Write(writer, binary.BigEndian, size)
}

func shortIdentity(value string) string {
	digest := sha256.Sum256([]byte("curator-transaction-sidecar-v1\x00" + value))
	return hex.EncodeToString(digest[:16])
}
