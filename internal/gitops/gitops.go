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
// committed blob bytes. Entries are listed with "git ls-tree -r -l -z" and
// blob bytes are read with "git cat-file --batch", which returns raw objects
// and never applies core.autocrlf, text/eol attributes, clean/smudge filters,
// or export-subst; the snapshot is therefore a function of the commit alone.
// The former "git archive" path did apply those conversions and so produced
// machine-dependent bytes.
//
// Refused: symbolic links (mode 120000), gitlinks/submodules (160000), any
// non-blob entry, empty or escaping paths and ".git" components, two tree
// paths that map to one platform path (case-insensitive collisions on such
// filesystems), entries that already exist under destination, and blobs
// larger than maxSnapshotFileBytes. Every refusal that needs no blob bytes
// is decided from the listing before cat-file starts, so a refused
// extraction writes nothing; a failure while streaming removes what this
// call wrote. Mode 100755 is preserved as 0o755; every other regular file is
// written 0o644.
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
	created := false
	if _, err := os.Lstat(destRoot); os.IsNotExist(err) {
		created = true
	}
	if err := os.MkdirAll(destRoot, 0o755); err != nil {
		return err
	}
	plan, err := planWrites(destRoot, entries)
	if err != nil {
		if created {
			_ = os.RemoveAll(destRoot)
		}
		return err
	}
	if err := writeBlobs(repo, plan); err != nil {
		if created {
			_ = os.RemoveAll(destRoot)
		} else {
			for _, entry := range plan {
				_ = os.Remove(entry.target)
			}
		}
		return err
	}
	return nil
}

// treeEntry is one "git ls-tree -r -l -z" record plus its planned target.
type treeEntry struct {
	mode   string
	kind   string
	oid    string
	size   int64
	path   string
	target string
}

// listTree lists the recursive contents of tree and refuses every entry that
// is not a regular blob. Output framing (verified on git 2.50):
// "<mode> <type> <oid> <size>\t<path>\0" per entry under -l -z, the size
// space-padded, "-" for non-blobs and "BAD" for an object the repository
// cannot read; paths unquoted.
func listTree(repo, tree string) ([]treeEntry, error) {
	cmd := exec.Command("git", "-C", repo, "ls-tree", "-r", "-l", "-z", "--full-tree", tree) // #nosec G204 -- fixed binary and flags; tree is a resolved object id
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
		if len(fields) != 4 {
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
		if fields[3] == "BAD" {
			return nil, fmt.Errorf("missing or unreadable object in git snapshot: %s %q", entry.oid, entry.path)
		}
		size, err := strconv.ParseInt(fields[3], 10, 64)
		if err != nil || size < 0 {
			return nil, fmt.Errorf("malformed ls-tree record in %s: %q", repo, record)
		}
		entry.size = size
		entries = append(entries, entry)
	}
	return entries, nil
}

// safeTarget maps a tree path to a destination path, rejecting empty names,
// absolute paths, escapes, and ".git" components (git's own verify_path
// refuses that name, case-insensitively).
func safeTarget(destRoot, name string) (string, error) {
	if name == "" || strings.HasPrefix(name, "/") || strings.ContainsRune(name, '\\') {
		return "", fmt.Errorf("unsafe path in git snapshot: %q", name)
	}
	for _, component := range strings.Split(name, "/") {
		if component == "" || component == "." || component == ".." || strings.EqualFold(component, ".git") {
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

// planWrites decides every refusal that needs no blob bytes before cat-file
// starts: path safety, the size bound from the listing, two tree paths that
// map to one platform path, and entries already present under destRoot.
//
// Tree paths differing only in case map to one platform path on
// case-insensitive filesystems. When two planned targets fold to the same
// key the destination filesystem is probed once; on a case-sensitive
// filesystem the distinct paths coexist, on a case-insensitive one the pair
// is refused (Spec core §2, §6.2). Any other folding a filesystem applies is
// caught by the existence check writeBlobs repeats before each file.
func planWrites(destRoot string, entries []treeEntry) ([]treeEntry, error) {
	plan := make([]treeEntry, 0, len(entries))
	exact := map[string]bool{}
	folded := map[string]bool{}
	caseInsensitive, probed := false, false
	for _, entry := range entries {
		target, err := safeTarget(destRoot, entry.path)
		if err != nil {
			return nil, err
		}
		if entry.size > maxSnapshotFileBytes {
			return nil, fmt.Errorf("file too large in git snapshot: %q", entry.path)
		}
		if exact[target] {
			return nil, fmt.Errorf("duplicate path in git snapshot: %q", entry.path)
		}
		exact[target] = true
		key := strings.ToLower(target)
		if folded[key] {
			if !probed {
				caseInsensitive, err = destinationFoldsCase(destRoot)
				if err != nil {
					return nil, err
				}
				probed = true
			}
			if caseInsensitive {
				return nil, fmt.Errorf("duplicate platform path in git snapshot: %q", entry.path)
			}
		}
		folded[key] = true
		if _, err := os.Lstat(target); err == nil {
			return nil, fmt.Errorf("duplicate platform path in git snapshot: %q", entry.path)
		} else if !os.IsNotExist(err) {
			return nil, err
		}
		entry.target = target
		plan = append(plan, entry)
	}
	return plan, nil
}

// destinationFoldsCase reports whether destRoot's filesystem maps two names
// differing only in case to one file, by creating and removing a probe file.
func destinationFoldsCase(destRoot string) (bool, error) {
	probe, err := os.CreateTemp(destRoot, ".curator-case-probe-*a")
	if err != nil {
		return false, err
	}
	name := probe.Name()
	_ = probe.Close()
	defer func() { _ = os.Remove(name) }()
	_, err = os.Lstat(strings.TrimSuffix(name, "a") + "A")
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// drainBound bounds the bytes discarded from a terminated cat-file child.
const drainBound = 64 << 20

// writeBlobs streams every planned entry's blob through one
// "git cat-file --batch" process and writes it to its target. Batch framing
// (verified on git 2.50): "<oid> blob <size>\n<bytes>\n" per requested
// object; a missing object answers "<oid> missing\n". cat-file reads raw
// objects: --filters and --textconv are deliberately not passed.
//
// A mid-stream error (framing, I/O, a size that disagrees with the listing)
// must not leave git blocked on a full stdout pipe: the child is terminated
// deterministically -- stdin closed, process killed, stdout drained to
// io.Discard within drainBound -- before Wait, and the partial target is
// removed. No path reaches Wait with undrained stdout.
func writeBlobs(repo string, entries []treeEntry) error {
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
	go func() {
		defer func() { _ = stdin.Close() }()
		for _, entry := range entries {
			if _, err := io.WriteString(stdin, entry.oid+"\n"); err != nil {
				return
			}
		}
	}()
	reader := bufio.NewReader(stdout)
	abort := func(partial string, err error) error {
		_ = stdin.Close()
		_ = cmd.Process.Kill()
		_, _ = io.Copy(io.Discard, io.LimitReader(reader, drainBound))
		_ = cmd.Wait()
		if partial != "" {
			_ = os.Remove(partial)
		}
		return err
	}
	for _, entry := range entries {
		header, err := reader.ReadString('\n')
		if err != nil {
			return abort("", fmt.Errorf("git cat-file --batch failed in %s: %v %s", repo, err, strings.TrimSpace(stderr.String())))
		}
		fields := strings.Fields(header)
		if len(fields) != 3 || fields[0] != entry.oid || fields[1] != "blob" {
			return abort("", fmt.Errorf("unexpected git cat-file --batch response for %q: %q", entry.path, strings.TrimSpace(header)))
		}
		size, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil || size != entry.size {
			return abort("", fmt.Errorf("unexpected git cat-file --batch response for %q: %q (listed size %d)", entry.path, strings.TrimSpace(header), entry.size))
		}
		if err := os.MkdirAll(filepath.Dir(entry.target), 0o755); err != nil {
			return abort("", err)
		}
		// The planned existence check is repeated here so a folding the
		// plan's probe did not model still refuses rather than replaces.
		if _, err := os.Lstat(entry.target); err == nil {
			return abort("", fmt.Errorf("duplicate platform path in git snapshot: %q", entry.path))
		} else if !os.IsNotExist(err) {
			return abort("", err)
		}
		mode := os.FileMode(0o644)
		if entry.mode == "100755" {
			mode = 0o755
		}
		file, err := os.OpenFile(entry.target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode) // #nosec G304 -- target is escape-checked by planWrites
		if err != nil {
			return abort("", err)
		}
		if _, err := io.CopyN(file, reader, size); err != nil {
			_ = file.Close()
			return abort(entry.target, fmt.Errorf("reading blob %s for %q: %w", entry.oid, entry.path, err))
		}
		if err := file.Close(); err != nil {
			return abort(entry.target, err)
		}
		// Trailing LF after the object bytes.
		if b, err := reader.ReadByte(); err != nil || b != '\n' {
			return abort(entry.target, fmt.Errorf("unexpected git cat-file --batch framing after %q", entry.path))
		}
	}
	_ = stdin.Close()
	// Every requested object has been consumed; the child has nothing left
	// to write, so Wait cannot block on the pipe.
	_, _ = io.Copy(io.Discard, io.LimitReader(reader, drainBound))
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("git cat-file --batch failed in %s: %v %s", repo, err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

// HasSubmodules reports whether a snapshot declares submodules, which are
// unsupported (Spec §8.2).
func HasSubmodules(snapshot string) bool {
	_, err := os.Stat(filepath.Join(snapshot, ".gitmodules"))
	return err == nil
}
