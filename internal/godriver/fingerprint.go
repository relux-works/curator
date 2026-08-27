package godriver

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
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
	path string
	kind byte
	link string
	info fs.FileInfo
}

func fingerprintToolchain(ctx context.Context, rootPath, version string) (string, []toolchainState, error) {
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return "", nil, diagnosticErr("toolchain_unreadable", err, "cannot open GOROOT")
	}
	defer func() { _ = root.Close() }()

	records, err := collectToolchainRecords(ctx, root)
	if err != nil {
		return "", nil, err
	}
	return digestToolchainRecords(ctx, root, records, version)
}

// collectToolchainRecords produces the canonical record set in canonical order.
// It is split from the digest phase so a test can mutate the tree exactly at the
// boundary between the two and observe which side of the traversal notices.
func collectToolchainRecords(ctx context.Context, root *os.Root) ([]treeRecord, error) {
	walk := toolchainWalk{root: root, encoded: make(map[string]struct{})}
	if err := walk.descend(ctx, root, ""); err != nil {
		return nil, err
	}
	records := walk.records
	sort.Slice(records, func(left, right int) bool { return records[left].path < records[right].path })
	return records, nil
}

// digestToolchainRecords writes the canonical byte stream and hashes it.
//
// Every file is re-opened from GOROOT by its root-relative path, so the bytes
// that reach the digest always come from whatever is reachable at that
// canonical path at read time, under the containment os.Root enforces on each
// component. Reusing an ancestor handle here instead would keep reading through
// a directory that had been renamed out of the tree, which is why this phase
// resolves paths and the collection phase does not.
func digestToolchainRecords(ctx context.Context, root *os.Root, records []treeRecord, version string) (string, []toolchainState, error) {
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
			file, openErr := root.Open(record.path)
			if openErr != nil {
				return "", nil, diagnosticErr("toolchain_unreadable", openErr, "cannot open toolchain file %q", record.path)
			}
			opened, statErr := file.Stat()
			if statErr != nil || !opened.Mode().IsRegular() || !os.SameFile(record.info, opened) {
				_ = file.Close()
				return "", nil, diagnosticErr("toolchain_mutated", statErr, "toolchain file %q changed while opening", record.path)
			}
			if opened.Size() < 0 {
				return "", nil, diagnostic("toolchain_mutated", "toolchain file %q reports a negative size", record.path)
			}
			writeLength(digest, uint64(opened.Size())) // #nosec G115 -- the size was just proved non-negative
			contentDigest := sha256.New()
			written, copyErr := copyWithContext(ctx, io.MultiWriter(digest, contentDigest), file)
			closeErr := file.Close()
			if err := digestCopyDiagnostic(record.path, written, opened.Size(), copyErr, closeErr); err != nil {
				return "", nil, err
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

// toolchainWalk collects the canonical record set. It reproduces exactly what
// fs.WalkDir over os.Root.FS() produced — the same entries, the same
// name-sorted visit order, and therefore the same error precedence — but it
// takes each entry's Lstat through a handle scoped to the directory the entry
// lives in instead of re-resolving the entry's whole root-relative path one
// component at a time. That Lstat is the only step this saves; every
// step whose result nothing later re-resolves stays anchored at GOROOT:
//
//   - each directory is listed through a handle opened by its full root-relative
//     path, exactly as fs.ReadDir opened it;
//   - an entry the listing types as a directory is Lstat-ed by its full path,
//     because the digest phase never revisits a directory record, and it is
//     descended by its full path, because fs.WalkDir took both decisions from
//     the listed entry and failed closed when the path stopped being one;
//   - a link's target and resolvability are read from the root, because a link
//     target is validated against GOROOT and not against its own directory —
//     the accepted RC4 vector contains "pkg/tool-link" -> "../bin/go", which
//     resolves inside the root but outside its own directory.
//
// What remains scoped is the Lstat of files and links, and both are re-checked
// later: a file is re-opened from GOROOT and matched with os.SameFile against
// this record, and a link is re-read and re-resolved from GOROOT below.
type toolchainWalk struct {
	root    *os.Root
	records []treeRecord
	encoded map[string]struct{}
}

func (walk *toolchainWalk) descend(ctx context.Context, dir *os.Root, prefix string) error {
	if err := ctx.Err(); err != nil {
		return diagnosticErr("toolchain_timeout", err, "toolchain fingerprint deadline exceeded")
	}
	entries, readErr := readScopedDir(dir)
	if readErr != nil {
		return diagnosticErr("toolchain_unreadable", readErr, "cannot walk GOROOT")
	}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return diagnosticErr("toolchain_timeout", err, "toolchain fingerprint deadline exceeded")
		}
		name := entry.Name()
		protocolPath := filepath.ToSlash(name)
		if prefix != "" {
			protocolPath = prefix + "/" + protocolPath
		}
		if !validToolchainPath(protocolPath) {
			return diagnostic("invalid_unicode", "toolchain path is not valid protocol UTF-8")
		}
		if err := claimEncodedPath(walk.encoded, protocolPath); err != nil {
			return err
		}

		// fs.WalkDir treats the listed entry's own type as the decision that
		// selects between "leaf" and "directory to read", so the same listed
		// type selects here which handle resolves the entry. An entry the
		// listing types as a directory is resolved from GOROOT, because
		// nothing re-resolves a directory record later and because the descent
		// below must see whatever occupies that path in the root; every other
		// entry is resolved through the scoped handle, and each thing its
		// record asserts is re-resolved from GOROOT before it is trusted.
		var info fs.FileInfo
		var statErr error
		if entry.IsDir() {
			info, statErr = walk.root.Lstat(protocolPath)
		} else {
			info, statErr = dir.Lstat(name)
		}
		if statErr != nil {
			return diagnosticErr("toolchain_unreadable", statErr, "cannot inspect toolchain path %q", protocolPath)
		}
		record := treeRecord{path: protocolPath, info: info}
		switch {
		case info.Mode().IsDir():
			record.kind = 'D'
		case info.Mode().IsRegular():
			record.kind = 'F'
		case info.Mode()&fs.ModeSymlink != 0:
			record.kind = 'L'
			target, readErr := walk.root.Readlink(protocolPath)
			if readErr != nil {
				return diagnosticErr("toolchain_link_dangling", readErr, "cannot read toolchain link %q", protocolPath)
			}
			// Readlink hands back the separators the host stored, which on
			// Windows are the ones the platform substituted for the protocol's.
			// Converting here, before anything validates or hashes the target,
			// is what keeps the toolchain identity a property of the tree
			// rather than of the host that walked it.
			target = protocolLinkTarget(target)
			if !utf8.ValidString(target) || strings.ContainsRune(target, 0) {
				return diagnostic("invalid_unicode", "toolchain link %q has an invalid target", protocolPath)
			}
			// A rooted target is absolute regardless of host path conventions:
			// Windows treats a leading separator as drive-relative rather than
			// absolute, so filepath.IsAbs alone would admit "/etc/passwd" there.
			if filepath.IsAbs(target) || filepath.VolumeName(target) != "" ||
				strings.HasPrefix(target, "/") || strings.HasPrefix(target, `\`) {
				return diagnostic("toolchain_link_absolute", "toolchain link %q is absolute", protocolPath)
			}
			resolvedName := filepath.Clean(filepath.Join(filepath.Dir(protocolPath), target))
			if resolvedName == ".." || strings.HasPrefix(resolvedName, ".."+string(filepath.Separator)) || filepath.IsAbs(resolvedName) {
				return diagnostic("toolchain_link_escape", "toolchain link %q escapes GOROOT", protocolPath)
			}
			if _, statErr := walk.root.Stat(protocolPath); statErr != nil {
				return diagnosticErr("toolchain_link_dangling", statErr, "toolchain link %q does not resolve safely", protocolPath)
			}
			record.link = target
		default:
			return diagnostic("special_file_forbidden", "toolchain path %q is not a directory, regular file, or link", protocolPath)
		}
		walk.records = append(walk.records, record)

		// fs.WalkDir descends on the type of the entry it listed, not on a
		// later lstat, and that entry type is lstat-backed, so a symlinked
		// directory is never followed. Taking the same decision from the same
		// listed entry keeps the fail-closed outcome when a listed directory
		// stops being one: opening it as a root fails exactly where fs.ReadDir
		// failed, with the same code and the same detail.
		if !entry.IsDir() {
			continue
		}
		child, openErr := walk.root.OpenRoot(protocolPath)
		if openErr != nil {
			return diagnosticErr("toolchain_unreadable", openErr, "cannot walk GOROOT")
		}
		descendErr := walk.descend(ctx, child, protocolPath)
		_ = child.Close()
		if descendErr != nil {
			return descendErr
		}
	}
	return nil
}

// readScopedDir lists one directory through its own handle. os.Root.FS() sorts
// entries by name before fs.WalkDir sees them; the canonical record set is
// sorted again later and does not depend on this order, but which offending
// entry reports the first error does, so the same order is reproduced here.
func readScopedDir(dir *os.Root) ([]fs.DirEntry, error) {
	handle, err := dir.Open(".")
	if err != nil {
		return nil, err
	}
	entries, readErr := handle.ReadDir(-1)
	closeErr := handle.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	sort.Slice(entries, func(left, right int) bool { return entries[left].Name() < entries[right].Name() })
	return entries, nil
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

// errCopyAbandoned marks the copy loop's own cancellation check, so the digest
// phase can tell "the caller stopped reading" from "the read failed" without
// having to recognise an error that came back from the filesystem.
var errCopyAbandoned = errors.New("toolchain copy abandoned")

// digestCopyDiagnostic classifies a record copy that did not finish cleanly.
//
// A copy carrying errCopyAbandoned stopped because the caller's context ended
// mid-file, and the short length that comes with it is that abort rather than
// the file moving underneath the read; cancellation is reported as the same
// deadline diagnostic every other cancellation check in this file reports. The
// precedence is exactly that narrow: only the abort this package raised is read
// as cancellation, so a genuine concurrent write — a read or close failure, or a
// length that moved on its own — still fails closed as toolchain_mutated even
// when the context happens to be done.
func digestCopyDiagnostic(path string, written, size int64, copyErr, closeErr error) error {
	if errors.Is(copyErr, errCopyAbandoned) {
		return diagnosticErr("toolchain_timeout", errors.Join(copyErr, closeErr), "toolchain fingerprint deadline exceeded")
	}
	if copyErr != nil || closeErr != nil || written != size {
		var lengthErr error
		if written != size {
			lengthErr = errors.New("file length changed")
		}
		return diagnosticErr("toolchain_mutated", errors.Join(copyErr, closeErr, lengthErr), "toolchain file %q changed while reading", path)
	}
	return nil
}

func copyWithContext(ctx context.Context, destination io.Writer, source io.Reader) (int64, error) {
	buffer := make([]byte, 128*1024)
	var total int64
	for {
		if err := ctx.Err(); err != nil {
			return total, fmt.Errorf("%w: %w", errCopyAbandoned, err)
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
