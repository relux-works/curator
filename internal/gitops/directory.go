package gitops

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/relux-works/curator/internal/identifiers"
)

// ExtractDirectory extracts only a selected Git tree, retaining its relative
// directory prefix. No sibling blob is read or materialized. Empty selection
// retains Extract's strict whole-snapshot behavior.
func ExtractDirectory(repo, commit, directory, destination string) error {
	if directory == "" {
		return Extract(repo, commit, destination)
	}
	if !identifiers.PortablePath(directory) {
		return fmt.Errorf("invalid selected directory %q", directory)
	}
	if err := EnsureRepo(repo); err != nil {
		return err
	}
	if strings.TrimSpace(commit) == "" || strings.HasPrefix(commit, "-") {
		return fmt.Errorf("refusing suspicious commit operand: %q", commit)
	}
	tree, err := revParse(repo, commit+"^{tree}")
	if err != nil {
		return err
	}
	tree, err = directoryTree(repo, tree, directory, nil)
	if err != nil {
		return err
	}
	entries, err := listTree(repo, tree)
	if err != nil {
		if link, ok := err.(*UnsupportedLinkError); ok {
			return &UnsupportedLinkError{Path: directory + "/" + link.Path}
		}
		return err
	}
	for i := range entries {
		entries[i].path = directory + "/" + entries[i].path
	}
	return extractEntries(repo, entries, destination)
}

// directoryTree walks exact tree components, never following a Git symlink or
// accepting a blob/submodule as an ancestor. The optional environment keeps
// pin verification on its isolated no-replacement object-database route.
func directoryTree(repo, tree, directory string, env []string) (string, error) {
	if !identifiers.PortablePath(directory) {
		return "", fmt.Errorf("invalid selected directory %q", directory)
	}
	walked := ""
	for _, component := range strings.Split(directory, "/") {
		if walked != "" {
			walked += "/"
		}
		walked += component
		cmd := gitCommand("-C", repo, "ls-tree", "-z", "--full-tree", tree)
		if env != nil {
			cmd.Env = env
		}
		out, err := cmd.Output()
		if err != nil {
			return "", objectUnavailable("tree", err)
		}
		found := false
		for _, record := range bytes.Split(out, []byte{0}) {
			if len(record) == 0 {
				continue
			}
			meta, name, ok := bytes.Cut(record, []byte{'\t'})
			fields := strings.Fields(string(meta))
			if !ok || len(fields) != 3 {
				return "", fmt.Errorf("malformed selected directory tree")
			}
			if string(name) != component {
				continue
			}
			if fields[0] == "120000" {
				return "", &UnsupportedLinkError{Path: walked}
			}
			if fields[0] != "040000" || fields[1] != "tree" {
				return "", fmt.Errorf("selected directory %q is not a Git tree", walked)
			}
			tree = fields[2]
			found = true
			break
		}
		if !found {
			return "", fmt.Errorf("selected directory %q does not exist", walked)
		}
	}
	return tree, nil
}
