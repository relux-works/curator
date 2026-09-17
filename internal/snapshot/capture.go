package snapshot

// Draft local-snapshot capture, hashing and immutable store (revision 1).
//
// This file implements protocol/skillfile-sources.md section 3 for local
// acquisition: a local path always means admitted filesystem bytes,
// including dirty, staged and untracked files, whether or not .git
// exists. It never means Git HEAD: no Git object is read here and no
// commit is synthesized; the admitted file list from boundaries.go is
// the only input set.
//
// Capture copies the admitted regular files into the acquisition's
// private staging directory, records the local-snapshot-v1 inventory
// (package-relative portable path, SHA-256 of raw bytes, executable bit
// for every file), then revalidates: the admitted path set is
// re-enumerated, the live bytes are re-read, and the frozen copy is
// rehashed. Any concurrent mutation fails with source_snapshot_changed
// and requires a new explicit attempt. All downstream operations
// consume the frozen copy, never the mutable original.
//
// PublishLocal moves one frozen copy into the immutable snapshot store
// below the manager home, keyed by the inventory digest. OpenLocal
// resolves a locked digest back to its frozen tree; an absent snapshot
// fails with source_snapshot_unavailable and a tree that no longer
// matches its digest fails with source_snapshot_changed. Neither path
// recreates a snapshot from current bytes.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/protocoljson"
)

// InventoryAlgorithm is the local-snapshot-v1 algorithm identifier
// carried by every inventory.
const InventoryAlgorithm = "curator-local-snapshot-v1"

// InventorySchemaVersion is the local-snapshot-v1 schema version.
const InventorySchemaVersion = 1

// FileEntry is one local-snapshot-v1 files member: the package-relative
// portable path, the SHA-256 of the exact raw bytes (with the sha256:
// prefix), and whether any POSIX execute bit is set (false on
// filesystems without that metadata).
type FileEntry struct {
	Path       string
	SHA256     string
	Executable bool
}

// Inventory is the local-snapshot-v1 byte inventory of one captured
// package. Snapshot holds the digest over every other field and is
// excluded from its own preimage.
type Inventory struct {
	SchemaVersion int
	Algorithm     string
	Files         []FileEntry
	Snapshot      string
}

// LocalStoreDir is the immutable local-snapshot store below the manager
// home. It is a managed output: callers pass it in their outputs so
// admission prunes it and refuses packages selected inside it.
func LocalStoreDir(home string) string { return filepath.Join(home, "local-snapshots") }

// NormalizeSnapshotDigest validates a snapshot digest and returns its
// canonical lowercase form.
func NormalizeSnapshotDigest(digest string) (string, error) {
	lowered := strings.ToLower(strings.TrimSpace(digest))
	hexpart, ok := strings.CutPrefix(lowered, "sha256:")
	if !ok || len(hexpart) != 64 || !isHex(hexpart) {
		return "", fmt.Errorf("source_snapshot_unavailable: snapshot digest %q is malformed", digest)
	}
	return "sha256:" + hexpart, nil
}

// localSnapshotDir maps a canonical digest onto its frozen tree.
func localSnapshotDir(home, normalized string) string {
	return filepath.Join(LocalStoreDir(home), strings.TrimPrefix(normalized, "sha256:"), "snapshot")
}

// Digest returns the local-snapshot-v1 digest: SHA-256, prefixed
// sha256:, over the CCJ-1 bytes of exactly {schema_version, algorithm,
// files} with entries sorted by UTF-8 path bytes. Line endings,
// Unicode, file contents and executable bits enter unnormalized;
// absolute source paths, timestamps, owner IDs, .git, pruned output and
// traversal order do not.
func (inventory Inventory) Digest() (string, error) {
	if inventory.SchemaVersion != InventorySchemaVersion {
		return "", fmt.Errorf("source_member_invalid: snapshot schema_version must be %d", InventorySchemaVersion)
	}
	if inventory.Algorithm != InventoryAlgorithm {
		return "", fmt.Errorf("source_member_invalid: snapshot algorithm must be %q", InventoryAlgorithm)
	}
	ordered := make([]FileEntry, len(inventory.Files))
	copy(ordered, inventory.Files)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Path < ordered[j].Path })
	files := make([]any, 0, len(ordered))
	for _, entry := range ordered {
		files = append(files, map[string]any{
			"path": entry.Path, "sha256": entry.SHA256, "executable": entry.Executable,
		})
	}
	payload, err := protocoljson.MarshalCanonical(map[string]any{
		"schema_version": inventory.SchemaVersion,
		"algorithm":      inventory.Algorithm,
		"files":          files,
	})
	if err != nil {
		return "", fmt.Errorf("source_member_invalid: snapshot inventory is not encodable: %v", err)
	}
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// BuildInventory hashes the exact byte inventory of files below
// packageRoot: one entry per absolute path with its package-relative
// portable path, raw-byte SHA-256 and executable bit. Entries sort by
// UTF-8 path bytes; duplicates and filesystem-equivalent path
// collisions fail. Links and special files in admitted inputs fail.
// Absence reports source_member_missing; every other defect reports
// source_member_invalid.
func BuildInventory(packageRoot string, files []string) (Inventory, error) {
	entries := make([]FileEntry, 0, len(files))
	folded := map[string]string{}
	for _, file := range files {
		info, err := os.Lstat(file) // #nosec G304 -- admitted inventory member under hashing
		if err != nil {
			if os.IsNotExist(err) {
				return Inventory{}, fmt.Errorf("source_member_missing: snapshot member %s: %v", file, err)
			}
			return Inventory{}, fmt.Errorf("source_member_invalid: snapshot member %s: %v", file, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return Inventory{}, fmt.Errorf("source_member_invalid: %s is a link", file)
		}
		if !info.Mode().IsRegular() {
			return Inventory{}, fmt.Errorf("source_member_invalid: %s is not a regular file", file)
		}
		rel, err := filepath.Rel(packageRoot, file)
		if err != nil {
			return Inventory{}, fmt.Errorf("source_member_invalid: %s escapes its package: %v", file, err)
		}
		posix := filepath.ToSlash(rel)
		if posix == "." || posix == ".." || strings.HasPrefix(posix, "../") || filepath.IsAbs(posix) {
			return Inventory{}, fmt.Errorf("source_member_invalid: %s escapes its package", file)
		}
		payload, err := os.ReadFile(file) // #nosec G304 -- admitted inventory member under hashing
		if err != nil {
			if os.IsNotExist(err) {
				return Inventory{}, fmt.Errorf("source_member_missing: snapshot member %s: %v", file, err)
			}
			return Inventory{}, fmt.Errorf("source_member_invalid: snapshot member %s: %v", file, err)
		}
		sum := sha256.Sum256(payload)
		entries = append(entries, FileEntry{
			Path:       posix,
			SHA256:     "sha256:" + hex.EncodeToString(sum[:]),
			Executable: info.Mode().Perm()&0o111 != 0,
		})
		if prior, ok := folded[strings.ToLower(posix)]; ok && prior != posix {
			return Inventory{}, fmt.Errorf("source_member_invalid: %s and %s map to one platform path", prior, posix)
		}
		folded[strings.ToLower(posix)] = posix
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	for index := 1; index < len(entries); index++ {
		if entries[index].Path == entries[index-1].Path {
			return Inventory{}, fmt.Errorf("source_member_invalid: duplicate snapshot member %q", entries[index].Path)
		}
	}
	inventory := Inventory{SchemaVersion: InventorySchemaVersion, Algorithm: InventoryAlgorithm, Files: entries}
	digest, err := inventory.Digest()
	if err != nil {
		return Inventory{}, err
	}
	inventory.Snapshot = digest
	return inventory, nil
}

// captureAfterCopyHook is a test-only deterministic scheduling seam: when
// non-nil it runs once after the copy loop so a test can race the live
// tree inside one Capture. Production code leaves it nil.
var captureAfterCopyHook func()

// publishAfterAuditHook is the same seam for publication: when non-nil it
// runs after the pre-publication staging audit so a test can mutate the
// frozen copy before the publication copy is made. Production code
// leaves it nil.
var publishAfterAuditHook func()

// Capture freezes one prepared acquisition: it copies the admitted
// dirty/untracked bytes into the acquisition staging directory and
// returns the byte inventory. The admitted set comes from the boundary
// enumeration, so .git metadata and managed outputs never enter while
// every other live byte does, including files Git reports as dirty or
// untracked.
//
// After the copy Capture re-enumerates the admitted path set,
// revalidates the physical identity of every admitted file (device and
// inode on unix, volume serial plus file index on Windows),
// re-reads the live bytes and rehashes the frozen copy. Any difference —
// a file added, removed or replaced even with identical bytes, bytes or
// executable bits changed live, or the staged copy differing from what
// was read — fails with source_snapshot_changed. Staging I/O failures
// unrelated to the tree are returned without a snapshot class.
func Capture(acquisition *LocalAcquisition) (Inventory, error) {
	if acquisition == nil || acquisition.Physical == "" || acquisition.Staging == "" {
		return Inventory{}, fmt.Errorf("source_member_invalid: local acquisition is not prepared")
	}
	live, err := BuildInventory(acquisition.Physical, acquisition.Files)
	if err != nil {
		return Inventory{}, fmt.Errorf("source_snapshot_changed: capture raced the admitted tree: %v", err)
	}
	identities, err := statIdentities(acquisition.Files)
	if err != nil {
		return Inventory{}, fmt.Errorf("source_snapshot_changed: capture raced the admitted tree: %v", err)
	}
	byPath := make(map[string]FileEntry, len(live.Files))
	for _, entry := range live.Files {
		byPath[entry.Path] = entry
	}
	for _, file := range acquisition.Files {
		rel, err := filepath.Rel(acquisition.Physical, file)
		if err != nil {
			return Inventory{}, fmt.Errorf("source_snapshot_changed: capture raced the admitted tree: %v", err)
		}
		posix := filepath.ToSlash(rel)
		entry, ok := byPath[posix]
		if !ok {
			return Inventory{}, fmt.Errorf("source_snapshot_changed: capture raced the admitted tree: %s vanished", file)
		}
		payload, err := os.ReadFile(file) // #nosec G304 -- admitted member under capture
		if err != nil {
			return Inventory{}, fmt.Errorf("source_snapshot_changed: capture raced %s: %v", posix, err)
		}
		destination := filepath.Join(acquisition.Staging, filepath.FromSlash(posix))
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return Inventory{}, fmt.Errorf("capture staging for %s: %w", posix, err)
		}
		mode := os.FileMode(0o644)
		if entry.Executable {
			mode = 0o755
		}
		if err := os.WriteFile(destination, payload, mode); err != nil {
			return Inventory{}, fmt.Errorf("capture staging for %s: %w", posix, err)
		}
	}
	if captureAfterCopyHook != nil {
		captureAfterCopyHook()
	}
	reenumerated, err := EnumerateInputs(acquisition.Physical, acquisition.Outputs, acquisition.RootEntries)
	if err != nil {
		return Inventory{}, fmt.Errorf("source_snapshot_changed: admitted inputs changed during capture: %v", err)
	}
	if !equalPaths(reenumerated, acquisition.Files) {
		return Inventory{}, fmt.Errorf("source_snapshot_changed: admitted inputs changed during capture: membership changed")
	}
	if err := revalidateIdentities(acquisition.Files, identities); err != nil {
		return Inventory{}, err
	}
	again, err := BuildInventory(acquisition.Physical, acquisition.Files)
	if err != nil {
		return Inventory{}, fmt.Errorf("source_snapshot_changed: admitted inputs changed during capture: %v", err)
	}
	if again.Snapshot != live.Snapshot {
		return Inventory{}, fmt.Errorf("source_snapshot_changed: admitted inputs changed during capture: bytes changed")
	}
	staged := make([]string, 0, len(acquisition.Files))
	for _, file := range acquisition.Files {
		rel, err := filepath.Rel(acquisition.Physical, file)
		if err != nil {
			return Inventory{}, fmt.Errorf("source_snapshot_changed: frozen copy diverged: %v", err)
		}
		staged = append(staged, filepath.Join(acquisition.Staging, rel))
	}
	frozen, err := BuildInventory(acquisition.Staging, staged)
	if err != nil {
		return Inventory{}, fmt.Errorf("source_snapshot_changed: frozen copy diverged: %v", err)
	}
	if frozen.Snapshot != live.Snapshot {
		return Inventory{}, fmt.Errorf("source_snapshot_changed: frozen copy diverged before publication")
	}
	return live, nil
}

// PublishLocal publishes one frozen copy into the immutable snapshot
// store below home, keyed by the inventory digest, and returns the
// store tree. The frozen copy is rehashed before publication, and the
// actual publication copy is hashed again immediately before the rename:
// a staged tree that no longer matches the audited inventory, or a
// publication copy mutated after the audit, fails with
// source_snapshot_changed and leaves no store entry. A concurrent
// publisher of the same digest wins harmlessly only when its complete
// tree authenticates; anything else conflicting fails the same way. The
// caller keeps owning staging.
func PublishLocal(home, staging string, inventory Inventory) (string, error) {
	if inventory.Snapshot == "" {
		return "", fmt.Errorf("source_member_invalid: snapshot has no digest; capture it first")
	}
	normalized, err := NormalizeSnapshotDigest(inventory.Snapshot)
	if err != nil {
		return "", err
	}
	staged := make([]string, 0, len(inventory.Files))
	for _, entry := range inventory.Files {
		staged = append(staged, filepath.Join(staging, filepath.FromSlash(entry.Path)))
	}
	frozen, err := BuildInventory(staging, staged)
	if err != nil {
		return "", fmt.Errorf("source_snapshot_changed: frozen copy changed between audit and publication: %v", err)
	}
	if frozen.Snapshot != normalized {
		return "", fmt.Errorf("source_snapshot_changed: frozen copy changed between audit and publication")
	}
	if publishAfterAuditHook != nil {
		publishAfterAuditHook()
	}
	target := localSnapshotDir(home, normalized)
	if _, err := os.Lstat(target); err == nil {
		if err := authenticateLocalSnapshot(target, normalized); err != nil {
			return "", err
		}
		return target, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", err
	}
	tmp, err := os.MkdirTemp(filepath.Dir(target), ".snapshot-*.tmp")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	for _, entry := range inventory.Files {
		payload, err := os.ReadFile(filepath.Join(staging, filepath.FromSlash(entry.Path))) // #nosec G304 -- frozen member under publication
		if err != nil {
			return "", fmt.Errorf("source_snapshot_changed: frozen copy changed between audit and publication: %v", err)
		}
		destination := filepath.Join(tmp, filepath.FromSlash(entry.Path))
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return "", err
		}
		mode := os.FileMode(0o644)
		if entry.Executable {
			mode = 0o755
		}
		if err := os.WriteFile(destination, payload, mode); err != nil {
			return "", err
		}
	}
	published := make([]string, 0, len(inventory.Files))
	for _, entry := range inventory.Files {
		published = append(published, filepath.Join(tmp, filepath.FromSlash(entry.Path)))
	}
	actual, err := BuildInventory(tmp, published)
	if err != nil {
		return "", fmt.Errorf("source_snapshot_changed: publication copy diverged before rename: %v", err)
	}
	if actual.Snapshot != normalized {
		return "", fmt.Errorf("source_snapshot_changed: publication copy diverged before rename")
	}
	if err := renameNoReplace(tmp, target); err != nil {
		authErr := authenticateLocalSnapshot(target, normalized)
		if authErr == nil {
			return target, nil
		}
		return "", fmt.Errorf("source_snapshot_changed: concurrent snapshot publication conflicts: %v", authErr)
	}
	return target, nil
}

// OpenLocal resolves a locked snapshot digest to its immutable frozen
// tree. It never falls back to current bytes: an absent snapshot fails
// with source_snapshot_unavailable, and a store tree that no longer
// matches its digest fails with source_snapshot_changed.
func OpenLocal(home, digest string) (string, error) {
	normalized, err := NormalizeSnapshotDigest(digest)
	if err != nil {
		return "", err
	}
	target := localSnapshotDir(home, normalized)
	if _, err := os.Lstat(target); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("source_snapshot_unavailable: snapshot %s is not in the store; capture it with an explicit attempt", normalized)
		}
		return "", err
	}
	if err := authenticateLocalSnapshot(target, normalized); err != nil {
		return "", err
	}
	return target, nil
}

// authenticateLocalSnapshot proves a store tree is exactly the snapshot
// digest: a real link-free directory whose complete byte inventory
// digests to normalized. Anything else fails with
// source_snapshot_changed and is never passed to installers.
func authenticateLocalSnapshot(target, normalized string) error {
	info, err := os.Lstat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("source_snapshot_unavailable: snapshot %s is not in the store; capture it with an explicit attempt", normalized)
		}
		return err
	}
	if !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
		return fmt.Errorf("source_snapshot_changed: stored snapshot %s is not a directory", normalized)
	}
	files, err := listStoreFiles(target)
	if err != nil {
		return err
	}
	actual, err := BuildInventory(target, files)
	if err != nil {
		return fmt.Errorf("source_snapshot_changed: stored snapshot %s failed revalidation: %v", normalized, err)
	}
	if actual.Snapshot != normalized {
		return fmt.Errorf("source_snapshot_changed: stored snapshot %s does not match its recorded bytes", normalized)
	}
	return nil
}

// listStoreFiles enumerates the regular files of one store tree. Links
// and special files fail with source_snapshot_changed: the tree is no
// longer the snapshot it claims to be.
func listStoreFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("source_snapshot_changed: stored snapshot carries a link at %s", path)
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("source_snapshot_changed: stored snapshot carries a special file at %s", path)
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		if strings.HasPrefix(err.Error(), "source_snapshot_changed") {
			return nil, err
		}
		return nil, fmt.Errorf("source_snapshot_changed: stored snapshot cannot be read: %v", err)
	}
	sort.Strings(files)
	return files, nil
}

func isHex(value string) bool {
	for _, character := range value {
		digit := character >= '0' && character <= '9'
		lower := character >= 'a' && character <= 'f'
		if !digit && !lower {
			return false
		}
	}
	return true
}

// statIdentities records the physical identity of every admitted file at
// inventory time: the file the digest covers, not just its bytes. Each
// token comes from one eager inspection (see captureFileToken), never
// from a lazily-resolved FileInfo: a record left lazy would resolve
// against whatever object the path names later and miss a same-byte
// replacement.
func statIdentities(files []string) (map[string]string, error) {
	out := make(map[string]string, len(files))
	for _, file := range files {
		token, err := captureFileToken(file)
		if err != nil {
			return nil, err
		}
		out[file] = token
	}
	return out, nil
}

// revalidateIdentities refuses an admitted file whose physical identity
// no longer matches the inventory-time record: a replaced file with
// identical bytes and mode is still a different source, so the fresh
// token (device and inode on unix, volume serial plus file index on
// Windows) must equal the pinned one for every member. Anything else
// fails with source_snapshot_changed.
func revalidateIdentities(files []string, baseline map[string]string) error {
	for _, file := range files {
		token, err := captureFileToken(file)
		if err != nil {
			return fmt.Errorf("source_snapshot_changed: admitted inputs changed during capture: %s: %v", file, err)
		}
		prior, ok := baseline[file]
		if !ok || prior != token {
			return fmt.Errorf("source_snapshot_changed: admitted inputs changed during capture: %s was replaced", file)
		}
	}
	return nil
}

func equalPaths(first, second []string) bool {
	if len(first) != len(second) {
		return false
	}
	for index := range first {
		if first[index] != second[index] {
			return false
		}
	}
	return true
}
