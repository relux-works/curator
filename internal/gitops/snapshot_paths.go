package gitops

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/identifiers"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

const maxPinnedPathBytes = 16 << 20
const maxPinnedPaths = 16384

// pinnedSnapshotSelection is shared by path-presence and tree-hash readers.
// It resolves no ref and reads no sibling of a selected directory.
func pinnedSnapshotSelection(repo, commit, directory string) (string, string, error) {
	if err := EnsureRepo(repo); err != nil {
		return "", "", err
	}
	env, cleanup, err := isolatedGitEnv()
	if err != nil {
		return "", "", err
	}
	defer cleanup()
	env = append(env, "GIT_NO_REPLACE_OBJECTS=1")
	for _, metadata := range []string{"objects/info/alternates", "info/grafts"} {
		path, err := runWithEnv(repo, env, "rev-parse", "--git-path", metadata)
		if err != nil {
			return "", "", err
		}
		if !filepath.IsAbs(path) {
			path = filepath.Join(repo, path)
		}
		if _, err := os.Lstat(path); err == nil {
			return "", "", fmt.Errorf("pinned snapshot refuses repository %s", metadata)
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", "", err
		}
	}
	tree, format, err := PinnedCommitTree(repo, commit)
	if err != nil {
		return "", "", err
	}
	if directory != "" {
		tree, err = directoryTree(repo, tree, directory, env)
	}
	return tree, format, err
}

type pinnedListingBuffer struct {
	buffer   bytes.Buffer
	exceeded bool
}

func (b *pinnedListingBuffer) Len() int      { return b.buffer.Len() }
func (b *pinnedListingBuffer) Bytes() []byte { return b.buffer.Bytes() }

func (b *pinnedListingBuffer) Write(raw []byte) (int, error) {
	if len(raw) > maxPinnedPathBytes-b.Len() {
		b.exceeded = true
		return 0, errors.New("pinned snapshot path metadata exceeds 16 MiB")
	}
	return b.buffer.Write(raw)
}

// PinnedDirectorySnapshotPaths returns the exact regular-blob roster of the
// selected local commit, retaining the same directory prefix as extraction
// and pin hashing. It fetches nothing and writes no snapshot or product state.
// Limits and unsafe metadata refuse with no partial roster.
func PinnedDirectorySnapshotPaths(repo, commit, directory string) ([]string, error) {
	tree, _, err := pinnedSnapshotSelection(repo, commit, directory)
	if err != nil {
		return nil, err
	}
	env, cleanup, err := isolatedGitEnv()
	if err != nil {
		return nil, err
	}
	defer cleanup()
	env = append(env, "GIT_NO_REPLACE_OBJECTS=1")
	cmd := gitCommand("-C", repo, "ls-tree", "-r", "-l", "-z", "--full-tree", tree)
	cmd.Env = env
	var stdout pinnedListingBuffer
	var stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	runErr := cmd.Run()
	// Git can exit on the closed output pipe before os/exec reports the
	// writer error. Preserve the named bound refusal in either ordering.
	if stdout.exceeded {
		return nil, errors.New("pinned snapshot path metadata exceeds 16 MiB")
	}
	if err := runErr; err != nil {
		return nil, fmt.Errorf("pinned snapshot path listing: %w", err)
	}
	entries, err := parseTreeListing(repo, stdout.Bytes())
	if err != nil {
		return nil, err
	}
	return pinnedPaths(entries, directory)
}

func pinnedPaths(entries []treeEntry, directory string) ([]string, error) {
	if len(entries) > maxPinnedPaths {
		return nil, errors.New("pinned snapshot path count exceeds 16384")
	}
	paths := make([]string, 0, len(entries))
	// Track directory prefixes as well: A/x and a/y collide on a folding
	// filesystem even though the full file names differ.
	folded := map[string]string{}
	files := map[string]bool{}
	directories := map[string]bool{}
	fold := cases.Fold()
	for _, entry := range entries {
		path := entry.path
		if directory != "" {
			path = directory + "/" + path
		}
		if !identifiers.PortablePath(path) {
			return nil, fmt.Errorf("unsafe pinned snapshot path %q", path)
		}
		prefix := ""
		parts := strings.Split(path, "/")
		for i, component := range parts {
			if strings.EqualFold(component, ".git") {
				return nil, fmt.Errorf("unsafe pinned snapshot Git metadata path %q", path)
			}
			if prefix != "" {
				prefix += "/"
			}
			prefix += component
			key := norm.NFC.String(fold.String(norm.NFC.String(prefix)))
			if original, ok := folded[key]; ok && original != prefix {
				return nil, fmt.Errorf("colliding pinned snapshot path %q", path)
			}
			if files[key] {
				return nil, fmt.Errorf("duplicate or colliding pinned snapshot path %q", path)
			}
			folded[key] = prefix
			if i == len(parts)-1 {
				if directories[key] {
					return nil, fmt.Errorf("colliding pinned snapshot path %q", path)
				}
				files[key] = true
			} else {
				directories[key] = true
			}
		}
		paths = append(paths, path)
	}
	return paths, nil
}
