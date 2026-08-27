package godriver

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

const toolchainDomain = "curator-go-toolchain-v1\x00"

type toolchainState struct {
	path    string
	kind    byte
	size    int64
	payload [sha256.Size]byte
}

type treeRecord struct {
	path       string
	filesystem string
	kind       byte
	link       string
	info       fs.FileInfo
}

func fingerprintToolchain(ctx context.Context, rootPath, version string) (string, []toolchainState, error) {
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return "", nil, diagnosticErr("toolchain_unreadable", err, "cannot open GOROOT")
	}
	defer root.Close()

	var records []treeRecord
	encoded := make(map[string]struct{})
	err = fs.WalkDir(root.FS(), ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return diagnosticErr("toolchain_unreadable", walkErr, "cannot walk GOROOT")
		}
		if err := ctx.Err(); err != nil {
			return diagnosticErr("toolchain_timeout", err, "toolchain fingerprint deadline exceeded")
		}
		if name == "." {
			return nil
		}
		protocolPath := filepath.ToSlash(name)
		if !validToolchainPath(protocolPath) {
			return diagnostic("invalid_unicode", "toolchain path is not valid protocol UTF-8")
		}
		if err := claimEncodedPath(encoded, protocolPath); err != nil {
			return err
		}

		info, statErr := root.Lstat(name)
		if statErr != nil {
			return diagnosticErr("toolchain_unreadable", statErr, "cannot inspect toolchain path %q", protocolPath)
		}
		record := treeRecord{path: protocolPath, filesystem: name, info: info}
		switch {
		case info.Mode().IsDir():
			record.kind = 'D'
		case info.Mode().IsRegular():
			record.kind = 'F'
		case info.Mode()&fs.ModeSymlink != 0:
			record.kind = 'L'
			target, readErr := root.Readlink(name)
			if readErr != nil {
				return diagnosticErr("toolchain_link_dangling", readErr, "cannot read toolchain link %q", protocolPath)
			}
			if !utf8.ValidString(target) || strings.ContainsRune(target, 0) {
				return diagnostic("invalid_unicode", "toolchain link %q has an invalid target", protocolPath)
			}
			if filepath.IsAbs(target) || filepath.VolumeName(target) != "" {
				return diagnostic("toolchain_link_absolute", "toolchain link %q is absolute", protocolPath)
			}
			resolvedName := filepath.Clean(filepath.Join(filepath.Dir(name), target))
			if resolvedName == ".." || strings.HasPrefix(resolvedName, ".."+string(filepath.Separator)) || filepath.IsAbs(resolvedName) {
				return diagnostic("toolchain_link_escape", "toolchain link %q escapes GOROOT", protocolPath)
			}
			if _, statErr := root.Stat(name); statErr != nil {
				return diagnosticErr("toolchain_link_dangling", statErr, "toolchain link %q does not resolve safely", protocolPath)
			}
			record.link = target
		default:
			return diagnostic("special_file_forbidden", "toolchain path %q is not a directory, regular file, or link", protocolPath)
		}
		records = append(records, record)
		return nil
	})
	if err != nil {
		return "", nil, err
	}
	sort.Slice(records, func(left, right int) bool { return records[left].path < records[right].path })

	digest := sha256.New()
	_, _ = io.WriteString(digest, toolchainDomain)
	state := make([]toolchainState, 0, len(records))
	for _, record := range records {
		if err := ctx.Err(); err != nil {
			return "", nil, diagnosticErr("toolchain_timeout", err, "toolchain fingerprint deadline exceeded")
		}
		pathBytes := []byte(record.path)
		writeRecordHeader(digest, record.kind, pathBytes)
		item := toolchainState{path: record.path, kind: record.kind}
		switch record.kind {
		case 'D':
			writeLength(digest, 0)
		case 'L':
			payload := []byte(record.link)
			writeLength(digest, uint64(len(payload)))
			_, _ = digest.Write(payload)
			item.size = int64(len(payload))
			item.payload = sha256.Sum256(payload)
		case 'F':
			file, openErr := root.Open(record.filesystem)
			if openErr != nil {
				return "", nil, diagnosticErr("toolchain_unreadable", openErr, "cannot open toolchain file %q", record.path)
			}
			opened, statErr := file.Stat()
			if statErr != nil || !opened.Mode().IsRegular() || !os.SameFile(record.info, opened) {
				_ = file.Close()
				return "", nil, diagnosticErr("toolchain_mutated", statErr, "toolchain file %q changed while opening", record.path)
			}
			writeLength(digest, uint64(opened.Size()))
			contentDigest := sha256.New()
			written, copyErr := copyWithContext(ctx, io.MultiWriter(digest, contentDigest), file)
			closeErr := file.Close()
			if copyErr != nil || closeErr != nil || written != opened.Size() {
				var lengthErr error
				if written != opened.Size() {
					lengthErr = errors.New("file length changed")
				}
				return "", nil, diagnosticErr("toolchain_mutated", errors.Join(copyErr, closeErr, lengthErr), "toolchain file %q changed while reading", record.path)
			}
			item.size = written
			copy(item.payload[:], contentDigest.Sum(nil))
		}
		state = append(state, item)
	}
	writeRecordHeader(digest, 'V', nil)
	writeLength(digest, uint64(len([]byte(version))))
	_, _ = io.WriteString(digest, version)
	return "sha256:" + hex.EncodeToString(digest.Sum(nil)), state, nil
}

func claimEncodedPath(encoded map[string]struct{}, path string) error {
	if _, duplicate := encoded[path]; duplicate {
		return diagnostic("duplicate_path", "toolchain contains duplicate encoded path %q", path)
	}
	encoded[path] = struct{}{}
	return nil
}

func validToolchainPath(path string) bool {
	if path == "" || path == "." || !utf8.ValidString(path) || strings.ContainsRune(path, 0) {
		return false
	}
	for _, component := range strings.Split(path, "/") {
		if component == "" || component == "." || component == ".." {
			return false
		}
	}
	return true
}

func writeRecordHeader(writer io.Writer, kind byte, path []byte) {
	_, _ = writer.Write([]byte{kind})
	writeLength(writer, uint64(len(path)))
	_, _ = writer.Write(path)
}

func writeLength(writer io.Writer, length uint64) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], length)
	_, _ = writer.Write(encoded[:])
}

func copyWithContext(ctx context.Context, destination io.Writer, source io.Reader) (int64, error) {
	buffer := make([]byte, 128*1024)
	var total int64
	for {
		if err := ctx.Err(); err != nil {
			return total, err
		}
		read, readErr := source.Read(buffer)
		if read > 0 {
			written, writeErr := destination.Write(buffer[:read])
			total += int64(written)
			if writeErr != nil {
				return total, writeErr
			}
			if written != read {
				return total, io.ErrShortWrite
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				return total, nil
			}
			return total, readErr
		}
	}
}
