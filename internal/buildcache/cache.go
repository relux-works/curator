// Package buildcache stores immutable, manager-protected go-v1 build outputs.
// Logical identities remain in buildmeta; this package owns only filesystem
// layout, protected-state validation, quarantine, and atomic publication.
package buildcache

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/buildmeta"
)

const (
	// ReceiptFilename is deliberately Curator-local. It is not part of the
	// portable receipt schema or logical cache identity.
	ReceiptFilename = "curator-receipt.ccj.json"
	maxReceiptBytes = 1 << 20
)

// Status is the reusable-cache outcome. Every non-Hit value fails closed.
type Status string

const (
	Hit                 Status = "hit"
	Miss                Status = "miss"
	Corrupt             Status = "corrupt"
	UntrustedProvenance Status = "untrusted-provenance"
	Unsupported         Status = "unsupported"
)

// Expectation is the independently derived logical state required for reuse.
// ReceiptHash may be empty during a cold cache lookup. Marker/currentness
// callers set it to require an exact previously recorded receipt identity.
type Expectation struct {
	Input       buildmeta.Input
	ReceiptHash buildmeta.ReceiptHash
}

// Result describes a read-only cache inspection. ArtifactPath is returned only
// for a protected, exact Hit and is never executed by this package.
type Result struct {
	Status       Status
	Reason       string
	Receipt      buildmeta.Receipt
	ReceiptBytes []byte
	ReceiptHash  buildmeta.ReceiptHash
	ArtifactPath string
}

// DryRunOutcome maps a read-only inspection to the stable planner vocabulary.
// It performs no I/O and, in particular, never repairs untrusted state.
func (result Result) DryRunOutcome() string {
	switch result.Status {
	case Hit:
		return "cache-hit"
	case Miss:
		return "would-preflight-and-build"
	case UntrustedProvenance:
		return "would-rebuild-untrusted-cache"
	case Corrupt:
		return "corrupt"
	default:
		return "unsupported"
	}
}

// Store addresses the Curator-local go-v1 cache below one manager home.
type Store struct {
	home      string
	supported func() bool
}

// New constructs a store without creating or changing any filesystem state.
func New(home string) (*Store, error) {
	if home == "" {
		return nil, fmt.Errorf("manager home is empty")
	}
	abs, err := filepath.Abs(home)
	if err != nil {
		return nil, fmt.Errorf("resolve manager home: %w", err)
	}
	if filepath.Clean(home) != home && filepath.IsAbs(home) {
		return nil, fmt.Errorf("manager home must be a clean absolute path")
	}
	return &Store{home: abs, supported: protectionSupported}, nil
}

// Home returns the clean absolute manager home used by this store.
func (store *Store) Home() string { return store.home }

// Inspect is strictly read-only. It validates the complete protected boundary,
// exact canonical receipt, complete expected input, optional receipt hash, and
// artifact path, bytes, size, and hash.
func (store *Store) Inspect(expect Expectation) Result {
	if store == nil || store.supported == nil || !store.supported() {
		return Result{Status: Unsupported, Reason: "platform protection is unavailable"}
	}
	key, err := expect.Input.CacheKey()
	if err != nil {
		return Result{Status: Corrupt, Reason: "invalid expected input: " + err.Error()}
	}
	entryPath, _, err := store.paths(key)
	if err != nil {
		return Result{Status: Corrupt, Reason: err.Error()}
	}
	if _, err := os.Lstat(entryPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Result{Status: Miss, Reason: "cache entry is absent"}
		}
		return Result{Status: UntrustedProvenance, Reason: "inspect cache boundary: " + err.Error()}
	}

	return store.inspectEntry(entryPath, expect, key)
}

func (store *Store) inspectEntry(entryPath string, expect Expectation, key buildmeta.CacheKey) Result {
	artifactRel, err := buildmeta.ArtifactPath(expect.Input.Command, expect.Input.Target.GOOS)
	if err != nil {
		return Result{Status: Corrupt, Reason: "derive artifact path: " + err.Error()}
	}
	opened, err := openProtectedEntry(store.home, entryPath, artifactRel)
	if err != nil {
		status := Corrupt
		if errors.Is(err, errUntrusted) {
			status = UntrustedProvenance
		}
		return Result{Status: status, Reason: err.Error()}
	}
	defer opened.close()

	if err := exactEntryContents(opened, artifactRel); err != nil {
		return Result{Status: Corrupt, Reason: err.Error()}
	}
	receiptBytes, err := readExactFile(opened.receipt)
	if err != nil {
		return Result{Status: Corrupt, Reason: "read receipt: " + err.Error()}
	}
	receipt, err := buildmeta.DecodeExpectedReceipt(receiptBytes, expect.Input)
	if err != nil {
		return Result{Status: Corrupt, Reason: "invalid receipt: " + err.Error()}
	}
	if receipt.CacheKey != key {
		return Result{Status: Corrupt, Reason: "receipt cache key does not match entry key"}
	}
	receiptHash, err := buildmeta.HashReceiptBytes(receiptBytes)
	if err != nil {
		return Result{Status: Corrupt, Reason: "hash receipt: " + err.Error()}
	}
	if expect.ReceiptHash != "" && receiptHash != expect.ReceiptHash {
		return Result{Status: Corrupt, Reason: "receipt hash mismatch"}
	}

	artifactHash, artifactSize, err := hashOpenFile(opened.artifact)
	if err != nil {
		return Result{Status: Corrupt, Reason: "read artifact: " + err.Error()}
	}
	if artifactSize != receipt.Artifact.Size {
		return Result{Status: Corrupt, Reason: "artifact size mismatch"}
	}
	if artifactHash != receipt.Artifact.SHA256 {
		return Result{Status: Corrupt, Reason: "artifact hash mismatch"}
	}
	return Result{
		Status:       Hit,
		Receipt:      receipt,
		ReceiptBytes: append([]byte(nil), receiptBytes...),
		ReceiptHash:  receiptHash,
		ArtifactPath: filepath.Join(entryPath, filepath.FromSlash(artifactRel)),
	}
}

func (store *Store) paths(key buildmeta.CacheKey) (entry, base string, err error) {
	keyText := string(key)
	if !strings.HasPrefix(keyText, "sha256:") || len(keyText) != len("sha256:")+64 {
		return "", "", fmt.Errorf("cache key is malformed")
	}
	hexKey := strings.TrimPrefix(keyText, "sha256:")
	if _, err := hex.DecodeString(hexKey); err != nil || strings.ToLower(hexKey) != hexKey {
		return "", "", fmt.Errorf("cache key is malformed")
	}
	base = filepath.Join(store.home, "cache", "build", buildmeta.DriverGoV1)
	entry = filepath.Join(base, hexKey)
	if !pathWithin(store.home, base) || !pathWithin(base, entry) {
		return "", "", fmt.Errorf("cache path crosses the manager-home boundary")
	}
	return entry, base, nil
}

func pathWithin(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

type openedEntry struct {
	entryDir *os.File
	binDir   *os.File
	receipt  *os.File
	artifact *os.File
	extra    []io.Closer
}

func (opened *openedEntry) close() {
	if opened == nil {
		return
	}
	for _, closer := range []io.Closer{opened.artifact, opened.receipt, opened.binDir, opened.entryDir} {
		if !isNilCloser(closer) {
			_ = closer.Close()
		}
	}
	for i := len(opened.extra) - 1; i >= 0; i-- {
		if opened.extra[i] != nil {
			_ = opened.extra[i].Close()
		}
	}
}

func isNilCloser(closer io.Closer) bool {
	if closer == nil {
		return true
	}
	value := reflect.ValueOf(closer)
	return value.Kind() == reflect.Pointer && value.IsNil()
}

func exactEntryContents(opened *openedEntry, artifactRel string) error {
	entryNames, err := directoryNames(opened.entryDir)
	if err != nil {
		return fmt.Errorf("list cache entry: %w", err)
	}
	if len(entryNames) != 2 || entryNames[0] != "bin" || entryNames[1] != ReceiptFilename {
		return fmt.Errorf("cache entry has unexpected contents")
	}
	binNames, err := directoryNames(opened.binDir)
	if err != nil {
		return fmt.Errorf("list artifact directory: %w", err)
	}
	wantName := filepath.Base(filepath.FromSlash(artifactRel))
	if len(binNames) != 1 || binNames[0] != wantName {
		return fmt.Errorf("artifact directory has unexpected contents")
	}
	return nil
}

func directoryNames(dir *os.File) ([]string, error) {
	if _, err := dir.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	entries, err := dir.ReadDir(-1)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	sort.Strings(names)
	return names, nil
}

func readExactFile(file *os.File) ([]byte, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() < 0 || info.Size() > maxReceiptBytes {
		return nil, fmt.Errorf("file size is outside the supported range")
	}
	payload, err := io.ReadAll(io.LimitReader(file, info.Size()+1))
	if err != nil {
		return nil, err
	}
	if int64(len(payload)) != info.Size() {
		return nil, fmt.Errorf("file changed while reading")
	}
	return payload, nil
}

func hashOpenFile(file *os.File) (string, int64, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", 0, err
	}
	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		return "", 0, err
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), size, nil
}

var errUntrusted = errors.New("untrusted cache provenance")

func untrustedf(format string, args ...any) error {
	return fmt.Errorf("%w: %s", errUntrusted, fmt.Sprintf(format, args...))
}
