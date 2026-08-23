// Package buildsource validates and identifies immutable raw skill snapshots.
package buildsource

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/relux-works/curator/internal/identifiers"
)

const (
	// Algorithm is the protocol identifier for immutable raw snapshot content.
	Algorithm = "curator-build-source-v1"
	domain    = Algorithm + "\x00"
)

var (
	// ErrInvalidSnapshot reports a raw snapshot that is not a portable tree of
	// directories and regular files.
	ErrInvalidSnapshot = errors.New("invalid build-source snapshot")
	// ErrSnapshotMutated reports that a validated snapshot changed while it was
	// frozen for cache lookup or build use.
	ErrSnapshotMutated = errors.New("build-source snapshot mutated")
)

// Identity is the domain-separated content identity of a raw snapshot.
type Identity struct {
	Algorithm     string `json:"algorithm"`
	ContentSHA256 string `json:"content_sha256"`
}

type stateEntry struct {
	path    string
	kind    byte
	size    int64
	content [sha256.Size]byte
}

type fileRecord struct {
	path string
	info fs.FileInfo
}

// Token binds an identity and tree state to one validated directory instance.
// Call Recheck after the last build child exits, or use Use to enforce checks
// immediately before and after a cache/build callback.
type Token struct {
	path     string
	root     *os.Root
	rootInfo fs.FileInfo
	identity Identity
	state    []stateEntry

	mu     sync.Mutex
	closed bool
}

// Validate rejects unsafe entries and returns a token for the exact directory
// instance that was hashed. Walks never follow links.
func Validate(path string) (*Token, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("%w: resolve root: %v", ErrInvalidSnapshot, err)
	}
	rootInfo, err := os.Lstat(abs)
	if err != nil {
		return nil, fmt.Errorf("%w: root: %v", ErrInvalidSnapshot, err)
	}
	if !rootInfo.IsDir() || rootInfo.Mode()&fs.ModeSymlink != 0 {
		return nil, fmt.Errorf("%w: root is not a directory", ErrInvalidSnapshot)
	}
	root, err := os.OpenRoot(abs)
	if err != nil {
		return nil, fmt.Errorf("%w: open root: %v", ErrInvalidSnapshot, err)
	}
	openedInfo, err := root.Lstat(".")
	if err != nil || !os.SameFile(rootInfo, openedInfo) {
		_ = root.Close()
		return nil, fmt.Errorf("%w: root changed while opening", ErrInvalidSnapshot)
	}

	identity, state, err := scan(root)
	if err != nil {
		_ = root.Close()
		return nil, err
	}
	return &Token{
		path:     abs,
		root:     root,
		rootInfo: openedInfo,
		identity: identity,
		state:    state,
	}, nil
}

// WithValidated validates before invoking use and rechecks after it returns.
// Cache lookup and every build child must remain inside the callback.
func WithValidated(path string, use func(*Token) error) error {
	token, err := Validate(path)
	if err != nil {
		return err
	}
	defer func() { _ = token.Close() }()
	return token.Use(use)
}

// Path returns the absolute path bound to the validated directory instance.
func (token *Token) Path() string { return token.path }

// Identity returns the immutable build-source identity computed by Validate.
func (token *Token) Identity() Identity { return token.identity }

// Equivalent reports whether two validated snapshots contain the same full
// tree. Unlike Identity, this comparison includes directory structure so an
// added or removed empty directory is not accepted as an unchanged instance.
func (token *Token) Equivalent(other *Token) bool {
	if token == nil || other == nil || token.identity != other.identity {
		return false
	}
	return slices.Equal(token.state, other.state)
}

// Use checks the token before invoking use and again after it returns. This
// places validation and digesting before a cache callback and detects mutation
// through the last child owned by that callback.
func (token *Token) Use(use func(*Token) error) error {
	if use == nil {
		return errors.New("build-source callback is nil")
	}
	if err := token.Recheck(); err != nil {
		return err
	}
	useErr := use(token)
	recheckErr := token.Recheck()
	return errors.Join(useErr, recheckErr)
}

// Recheck verifies both the directory instance and every path, file type, and
// file byte against the state frozen by Validate. Modes and timestamps remain
// deliberately outside the identity and frozen state.
func (token *Token) Recheck() error {
	token.mu.Lock()
	defer token.mu.Unlock()
	if token.closed {
		return fmt.Errorf("%w: validation token is closed", ErrSnapshotMutated)
	}
	currentRoot, err := os.Lstat(token.path)
	if err != nil || !os.SameFile(token.rootInfo, currentRoot) {
		return fmt.Errorf("%w: validated root was replaced", ErrSnapshotMutated)
	}
	identity, state, err := scan(token.root)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSnapshotMutated, err)
	}
	if identity != token.identity || !slices.Equal(state, token.state) {
		return ErrSnapshotMutated
	}
	return nil
}

// Close releases the directory handle retained by the validation token.
func (token *Token) Close() error {
	if token == nil {
		return nil
	}
	token.mu.Lock()
	defer token.mu.Unlock()
	if token.closed {
		return nil
	}
	token.closed = true
	return token.root.Close()
}

func scan(root *os.Root) (Identity, []stateEntry, error) {
	var files []fileRecord
	var state []stateEntry
	paths := newPathSet()
	err := fs.WalkDir(root.FS(), ".", func(path string, _ fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == "." {
			return nil
		}
		if err := paths.add(path); err != nil {
			return err
		}
		local := filepath.FromSlash(path)
		info, err := root.Lstat(local)
		if err != nil {
			return err
		}
		switch {
		case info.IsDir():
			state = append(state, stateEntry{path: path, kind: 'D'})
		case info.Mode().IsRegular():
			files = append(files, fileRecord{path: path, info: info})
		case info.Mode()&fs.ModeSymlink != 0:
			return fmt.Errorf("%w: link forbidden at %q", ErrInvalidSnapshot, path)
		default:
			return fmt.Errorf("%w: special file forbidden at %q", ErrInvalidSnapshot, path)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrInvalidSnapshot) {
			return Identity{}, nil, err
		}
		return Identity{}, nil, fmt.Errorf("%w: walk: %v", ErrInvalidSnapshot, err)
	}

	slices.SortFunc(files, func(left, right fileRecord) int {
		return bytes.Compare([]byte(left.path), []byte(right.path))
	})
	slices.SortFunc(state, compareState)

	digest := sha256.New()
	_, _ = digest.Write([]byte(domain))
	for _, record := range files {
		contentDigest, size, err := hashFile(root, record, digest)
		if err != nil {
			return Identity{}, nil, err
		}
		state = append(state, stateEntry{
			path: record.path, kind: 'F', size: size, content: contentDigest,
		})
	}
	slices.SortFunc(state, compareState)
	return Identity{
		Algorithm:     Algorithm,
		ContentSHA256: "sha256:" + hex.EncodeToString(digest.Sum(nil)),
	}, state, nil
}

func hashFile(root *os.Root, record fileRecord, digest io.Writer) ([sha256.Size]byte, int64, error) {
	var zero [sha256.Size]byte
	if record.info.Size() < 0 {
		return zero, 0, fmt.Errorf("%w: negative size at %q", ErrInvalidSnapshot, record.path)
	}
	file, err := root.Open(filepath.FromSlash(record.path))
	if err != nil {
		return zero, 0, fmt.Errorf("%w: open %q: %v", ErrInvalidSnapshot, record.path, err)
	}
	defer func() { _ = file.Close() }()
	openedInfo, err := file.Stat()
	if err != nil || !openedInfo.Mode().IsRegular() || !os.SameFile(record.info, openedInfo) {
		return zero, 0, fmt.Errorf("%w: file changed while opening %q", ErrInvalidSnapshot, record.path)
	}

	_, _ = digest.Write([]byte{'F'})
	writeUint64(digest, uint64(len([]byte(record.path))))
	_, _ = digest.Write([]byte(record.path))
	// #nosec G115 -- the negative range was rejected before the file was opened
	// and os.SameFile above proves this is the same regular file.
	writeUint64(digest, uint64(openedInfo.Size()))
	contentHash := sha256.New()
	written, copyErr := io.CopyN(io.MultiWriter(digest, contentHash), file, openedInfo.Size())
	if copyErr != nil {
		return zero, 0, fmt.Errorf("%w: read %q: %v", ErrInvalidSnapshot, record.path, copyErr)
	}
	var extra [1]byte
	count, extraErr := file.Read(extra[:])
	if count != 0 || (extraErr != nil && extraErr != io.EOF) {
		return zero, 0, fmt.Errorf("%w: file grew while reading %q", ErrInvalidSnapshot, record.path)
	}
	afterInfo, err := file.Stat()
	if err != nil || afterInfo.Size() != written || !os.SameFile(openedInfo, afterInfo) {
		return zero, 0, fmt.Errorf("%w: file changed while reading %q", ErrInvalidSnapshot, record.path)
	}
	var sum [sha256.Size]byte
	copy(sum[:], contentHash.Sum(nil))
	return sum, written, nil
}

func writeUint64(writer io.Writer, value uint64) {
	var framed [8]byte
	binary.BigEndian.PutUint64(framed[:], value)
	_, _ = writer.Write(framed[:])
}

func compareState(left, right stateEntry) int {
	if pathOrder := bytes.Compare([]byte(left.path), []byte(right.path)); pathOrder != 0 {
		return pathOrder
	}
	return int(left.kind) - int(right.kind)
}

type pathSet struct {
	exact map[string]struct{}
}

func newPathSet() *pathSet {
	return &pathSet{
		exact: make(map[string]struct{}),
	}
}

func (set *pathSet) add(path string) error {
	if !identifiers.PortablePath(path) {
		return fmt.Errorf("%w: invalid protocol path %q", ErrInvalidSnapshot, path)
	}
	if _, exists := set.exact[path]; exists {
		return fmt.Errorf("%w: duplicate protocol path %q", ErrInvalidSnapshot, path)
	}
	set.exact[path] = struct{}{}
	return nil
}
