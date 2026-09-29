package gitops

import (
	"bytes"
	"crypto/sha1" // #nosec G505 -- Git SHA-1 object IDs must retain Git's required object format.
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/stateread"
)

// PinnedCommitTree resolves a full commit ID to its tree ID using only the
// repository's existing object database. It never fetches or consults refs.
func PinnedCommitTree(repo, commit string) (treeID, objectFormat string, err error) {
	if err := EnsureRepo(repo); err != nil {
		return "", "", err
	}
	if err := validateLockedCommit(commit); err != nil {
		return "", "", err
	}
	env, cleanup, err := isolatedGitEnv()
	if err != nil {
		return "", "", err
	}
	defer cleanup()
	// Do not let caller Git overrides or replacement refs supply objects
	// outside the source repository's own object database.
	env = append(env, "GIT_NO_REPLACE_OBJECTS=1")
	objectFormat, err = runWithEnv(repo, env, "rev-parse", "--show-object-format")
	if err != nil {
		return "", "", fmt.Errorf("read pinned repository object format: %w", err)
	}
	if (objectFormat == "sha1" && len(commit) != 40) || (objectFormat == "sha256" && len(commit) != 64) {
		return "", "", fmt.Errorf("commit object ID does not match repository object format %s", objectFormat)
	}
	resolved, err := runWithEnv(repo, env, "rev-parse", "--verify", commit+"^{commit}")
	if err != nil {
		return "", "", fmt.Errorf("pinned commit object unavailable: could not resolve %s^{commit}: %w", commit, err)
	}
	if !strings.EqualFold(resolved, commit) {
		return "", "", fmt.Errorf("pinned commit object unavailable: repository resolved %s as %s", commit, resolved)
	}
	treeID, err = runWithEnv(repo, env, "rev-parse", "--verify", commit+"^{tree}")
	if err != nil {
		return "", "", fmt.Errorf("pinned commit tree unavailable: could not resolve %s^{tree}: %w", commit, err)
	}
	return treeID, objectFormat, nil
}

// TreeObjectIDFromDir computes a Git tree object ID from a regular-file tree.
// The caller establishes the store boundary first; this function still uses
// lstat and the shared state-read seam so links and failed reads cannot be
// mistaken for bytes.
func TreeObjectIDFromDir(root, objectFormat string) (string, error) {
	newHash, err := gitHasher(objectFormat)
	if err != nil {
		return "", err
	}
	metadata, err := stateread.Lstat(root)
	if err != nil {
		return "", err
	}
	if metadata.Kind != stateread.KindPresent || metadata.Info == nil || !metadata.Info.IsDir() {
		return "", fmt.Errorf("tree root is not a present directory")
	}
	return treeObjectID(root, objectFormat, newHash)
}

type treeObjectEntry struct {
	name   string
	mode   string
	isDir  bool
	object []byte
}

func treeObjectID(dir, objectFormat string, newHash func() hash.Hash) (string, error) {
	listing, err := stateread.ReadDir(dir)
	if err != nil {
		return "", err
	}
	if listing.Kind != stateread.KindPresent {
		return "", fmt.Errorf("tree directory is not present")
	}
	entries := make([]treeObjectEntry, 0, len(listing.Entries))
	for _, child := range listing.Entries {
		name := child.Name()
		path := filepath.Join(dir, name)
		metadata, err := stateread.Lstat(path)
		if err != nil {
			return "", err
		}
		if metadata.Kind != stateread.KindPresent || metadata.Info == nil {
			return "", fmt.Errorf("tree entry %q is not present", name)
		}
		entry := treeObjectEntry{name: name}
		switch {
		case metadata.Info.IsDir():
			entry.mode = "40000"
			entry.isDir = true
			childTree, err := treeObjectID(path, objectFormat, newHash)
			if err != nil {
				return "", fmt.Errorf("hash directory %q: %w", name, err)
			}
			entry.object, err = hex.DecodeString(childTree)
			if err != nil {
				return "", fmt.Errorf("decode tree ID for %q: %w", name, err)
			}
		case metadata.Info.Mode().IsRegular():
			if metadata.Info.Size() < 0 || metadata.Info.Size() > maxSnapshotFileBytes {
				return "", fmt.Errorf("tree file %q exceeds the supported snapshot size", name)
			}
			file, err := stateread.ReadRegularFile(path)
			if err != nil {
				return "", err
			}
			if file.Kind != stateread.KindPresent {
				return "", fmt.Errorf("tree file %q is not present", name)
			}
			entry.mode = "100644"
			if metadata.Info.Mode().Perm()&0o111 != 0 {
				entry.mode = "100755"
			}
			entry.object = gitObjectID("blob", file.Bytes, newHash)
		default:
			return "", fmt.Errorf("tree entry %q is not a regular file or directory", name)
		}
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return bytes.Compare(treeSortKey(entries[i]), treeSortKey(entries[j])) < 0
	})
	var body bytes.Buffer
	for _, entry := range entries {
		body.WriteString(entry.mode)
		body.WriteByte(' ')
		body.WriteString(entry.name)
		body.WriteByte(0)
		body.Write(entry.object)
	}
	return hex.EncodeToString(gitObjectID("tree", body.Bytes(), newHash)), nil
}

func treeSortKey(entry treeObjectEntry) []byte {
	key := make([]byte, 0, len(entry.name)+1)
	key = append(key, entry.name...)
	if entry.isDir {
		key = append(key, '/')
	} else {
		key = append(key, 0)
	}
	return key
}

func gitObjectID(kind string, content []byte, newHash func() hash.Hash) []byte {
	digest := newHash()
	_, _ = digest.Write([]byte(kind + " " + strconv.Itoa(len(content)) + "\x00"))
	_, _ = digest.Write(content)
	return digest.Sum(nil)
}

func gitHasher(objectFormat string) (func() hash.Hash, error) {
	switch objectFormat {
	case "sha1":
		return func() hash.Hash { return sha1.New() }, nil // #nosec G401 -- required to reproduce legacy Git tree object IDs.
	case "sha256":
		return func() hash.Hash { return sha256.New() }, nil
	default:
		return nil, fmt.Errorf("unsupported Git object format %q", objectFormat)
	}
}
