// Package gitops shells out to system git for clone, fetch, ref resolution,
// and snapshot extraction (Spec §8.2). Snapshots are extracted from the
// object database so their bytes never depend on the acquiring machine's git
// configuration or the repository's attributes (environments §1.2).
//
// Declared git URLs reach clone as untrusted input. Restricting the
// transport protocols blocks remote-helper URLs such as ext::sh -c ... which
// would otherwise execute arbitrary commands during installation.
package gitops

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// AllowedProtocols is the GIT_ALLOW_PROTOCOL value for every network
// operation.
const AllowedProtocols = "file:git:http:https:ssh"

// maxSnapshotFileBytes bounds one extracted file (decompression bomb guard).
var maxSnapshotFileBytes int64 = 512 << 20

// ResolvedRef is a reference resolved to a commit.
type ResolvedRef struct {
	Kind   string
	Ref    string
	Commit string
}

func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...) // #nosec G204 -- fixed binary, arguments are built by this package
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(), "GIT_ALLOW_PROTOCOL="+AllowedProtocols)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = strings.TrimSpace(stdout.String())
		}
		return "", fmt.Errorf("git %s failed: %s", strings.Join(args, " "), detail)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// Clone clones a remote URL into destination. It refuses suspicious URLs
// (empty or dash-prefixed) and passes the URL positionally after "--".
func Clone(remoteURL, destination string) error {
	trimmed := strings.TrimSpace(remoteURL)
	if trimmed == "" || strings.HasPrefix(trimmed, "-") {
		return fmt.Errorf("refusing to clone suspicious git URL: %q", remoteURL)
	}
	if _, err := os.Stat(destination); err == nil {
		return fmt.Errorf("clone destination already exists: %s", destination)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	if _, err := run("", "clone", "--", trimmed, destination); err != nil {
		_ = os.RemoveAll(destination)
		return fmt.Errorf("git clone failed for %s -> %s: %w", remoteURL, destination, err)
	}
	return nil
}

// EnsureRepo verifies dir is a git repository.
func EnsureRepo(dir string) error {
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		return fmt.Errorf("not a git repository: %s", dir)
	}
	return nil
}

// Resolve resolves a ref of the given kind to a commit (Spec §8.2): tag via
// refs/tags/<v>^{commit}, revision via <v>^{commit}, branch preferring
// refs/remotes/origin/<v> and falling back to a local head.
func Resolve(repo, kind, value string) (ResolvedRef, error) {
	if err := EnsureRepo(repo); err != nil {
		return ResolvedRef{}, err
	}
	var commit string
	var err error
	switch kind {
	case "tag":
		commit, err = revParse(repo, "refs/tags/"+value+"^{commit}")
	case "revision":
		commit, err = revParse(repo, value+"^{commit}")
	case "branch":
		commit, err = revParse(repo, "refs/remotes/origin/"+value)
		if err != nil {
			commit, err = revParse(repo, "refs/heads/"+value)
		}
	default:
		return ResolvedRef{}, fmt.Errorf("unknown ref kind: %s", kind)
	}
	if err != nil {
		return ResolvedRef{}, err
	}
	return ResolvedRef{Kind: kind, Ref: value, Commit: commit}, nil
}

func revParse(repo, spec string) (string, error) {
	out, err := run(repo, "rev-parse", "--verify", spec)
	if err != nil {
		return "", fmt.Errorf("could not resolve %s in %s: %w", spec, repo, err)
	}
	if out == "" {
		return "", fmt.Errorf("could not resolve %s in %s", spec, repo)
	}
	return out, nil
}

// Fetch updates a repository: all remotes, tags, prune.
func Fetch(repo string) error {
	if err := EnsureRepo(repo); err != nil {
		return err
	}
	_, err := run(repo, "fetch", "--all", "--tags", "--prune")
	return err
}

// Extract materializes the tree of a commit into destination from the git
// object database (Spec core §6.2, §6.5; environments §1.2).
//
// Every regular-file entry of the commit's tree is written with exactly its
// committed blob bytes. Entries are listed with "git ls-tree -r -z" and blob
// bytes are read with "git cat-file --batch", which returns raw objects and
// never applies core.autocrlf, text/eol attributes, clean/smudge filters, or
// export-subst; the snapshot is therefore a function of the commit alone.
// The former "git archive" path did apply those conversions and so produced
// machine-dependent bytes.
//
// Refused: symbolic links (mode 120000), gitlinks/submodules (160000), any
// non-blob entry, empty or escaping paths, two tree paths that map to one
// platform path (case-insensitive collisions on such filesystems), and blobs
// larger than maxSnapshotFileBytes. Mode 100755 is preserved as 0o755; every
// other regular file is written 0o644.
func Extract(repo, commit, destination string) error {
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
	entries, err := listTree(repo, tree)
	if err != nil {
		return err
	}
	destRoot, err := filepath.Abs(destination)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(destRoot, 0o755); err != nil {
		return err
	}
	return writeBlobs(repo, destRoot, entries)
}

// treeEntry is one "git ls-tree -r -z" record.
type treeEntry struct {
	mode string
	kind string
	oid  string
	path string
}

// listTree lists the recursive contents of tree and refuses every entry that
// is not a regular blob. Output framing (verified on git 2.50):
// "<mode> <type> <oid>\t<path>\0" per entry, paths unquoted under -z.
func listTree(repo, tree string) ([]treeEntry, error) {
	cmd := exec.Command("git", "-C", repo, "ls-tree", "-r", "-z", "--full-tree", tree) // #nosec G204 -- fixed binary and flags; tree is a resolved object id
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git ls-tree failed in %s: %s", repo, strings.TrimSpace(stderr.String()))
	}
	var entries []treeEntry
	for _, record := range bytes.Split(stdout.Bytes(), []byte{0}) {
		if len(record) == 0 {
			continue
		}
		meta, path, ok := bytes.Cut(record, []byte{'\t'})
		if !ok {
			return nil, fmt.Errorf("malformed ls-tree record in %s: %q", repo, record)
		}
		fields := strings.Fields(string(meta))
		if len(fields) != 3 {
			return nil, fmt.Errorf("malformed ls-tree record in %s: %q", repo, record)
		}
		entry := treeEntry{mode: fields[0], kind: fields[1], oid: fields[2], path: string(path)}
		switch {
		case entry.mode == "120000":
			return nil, fmt.Errorf("links in git snapshots are unsupported: %q", entry.path)
		case entry.mode == "160000" || entry.kind == "commit":
			return nil, fmt.Errorf("unsupported entry type in git snapshot (submodule): %q", entry.path)
		case entry.kind != "blob" || (entry.mode != "100644" && entry.mode != "100755"):
			return nil, fmt.Errorf("unsupported entry type in git snapshot: %s %s %q", entry.mode, entry.kind, entry.path)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// safeTarget maps a tree path to a destination path, rejecting empty names,
// absolute paths, and escapes.
func safeTarget(destRoot, name string) (string, error) {
	if name == "" || strings.HasPrefix(name, "/") || strings.ContainsRune(name, '\\') {
		return "", fmt.Errorf("unsafe path in git snapshot: %q", name)
	}
	for _, component := range strings.Split(name, "/") {
		if component == "" || component == "." || component == ".." {
			return "", fmt.Errorf("unsafe path in git snapshot: %q", name)
		}
	}
	target := filepath.Join(destRoot, filepath.FromSlash(name))
	rel, err := filepath.Rel(destRoot, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe path in git snapshot: %q", name)
	}
	return target, nil
}

// writeBlobs streams every entry's blob through one "git cat-file --batch"
// process and writes it under destRoot. Batch framing (verified on git 2.50):
// "<oid> blob <size>\n<bytes>\n" per requested object; a missing object
// answers "<oid> missing\n". cat-file reads raw objects: --filters and
// --textconv are deliberately not passed.
func writeBlobs(repo, destRoot string, entries []treeEntry) error {
	if len(entries) == 0 {
		return nil
	}
	cmd := exec.Command("git", "-C", repo, "cat-file", "--batch") // #nosec G204 -- fixed binary and flags
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	defer func() {
		_ = stdin.Close()
		_ = cmd.Wait()
	}()
	go func() {
		defer func() { _ = stdin.Close() }()
		for _, entry := range entries {
			if _, err := io.WriteString(stdin, entry.oid+"\n"); err != nil {
				return
			}
		}
	}()
	reader := bufio.NewReader(stdout)
	// Tree paths differing only in case (or other platform folding) map to
	// one platform path on case-insensitive filesystems. Remember what this
	// extraction wrote and refuse a second entry that lands on an existing
	// file: on a case-sensitive filesystem the distinct paths coexist; on a
	// case-insensitive one the second write would have silently replaced the
	// first (Spec core §2, §6.2).
	written := map[string]bool{}
	for _, entry := range entries {
		target, err := safeTarget(destRoot, entry.path)
		if err != nil {
			return err
		}
		header, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("git cat-file --batch failed in %s: %v %s", repo, err, strings.TrimSpace(stderr.String()))
		}
		fields := strings.Fields(header)
		if len(fields) != 3 || fields[0] != entry.oid || fields[1] != "blob" {
			return fmt.Errorf("unexpected git cat-file --batch response for %q: %q", entry.path, strings.TrimSpace(header))
		}
		size, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil || size < 0 {
			return fmt.Errorf("unexpected git cat-file --batch response for %q: %q", entry.path, strings.TrimSpace(header))
		}
		if size > maxSnapshotFileBytes {
			return fmt.Errorf("file too large in git snapshot: %q", entry.path)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if written[target] {
			return fmt.Errorf("duplicate path in git snapshot: %q", entry.path)
		}
		if _, err := os.Lstat(target); err == nil {
			return fmt.Errorf("duplicate platform path in git snapshot: %q", entry.path)
		} else if !os.IsNotExist(err) {
			return err
		}
		mode := os.FileMode(0o644)
		if entry.mode == "100755" {
			mode = 0o755
		}
		file, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode) // #nosec G304 -- target is escape-checked above
		if err != nil {
			return err
		}
		if _, err := io.CopyN(file, reader, size); err != nil {
			_ = file.Close()
			return fmt.Errorf("reading blob %s for %q: %w", entry.oid, entry.path, err)
		}
		if err := file.Close(); err != nil {
			return err
		}
		written[target] = true
		// Trailing LF after the object bytes.
		if b, err := reader.ReadByte(); err != nil || b != '\n' {
			return fmt.Errorf("unexpected git cat-file --batch framing after %q", entry.path)
		}
	}
	return nil
}

// HasSubmodules reports whether a snapshot declares submodules, which are
// unsupported (Spec §8.2).
func HasSubmodules(snapshot string) bool {
	_, err := os.Stat(filepath.Join(snapshot, ".gitmodules"))
	return err == nil
}
