// Package gitops shells out to system git for clone, fetch, ref resolution,
// and snapshot extraction (Spec §8.2).
//
// Declared git URLs reach clone as untrusted input. Restricting the
// transport protocols blocks remote-helper URLs such as ext::sh -c ... which
// would otherwise execute arbitrary commands during installation.
package gitops

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AllowedProtocols is the GIT_ALLOW_PROTOCOL value for every network
// operation.
const AllowedProtocols = "file:git:http:https:ssh"

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

// Archive extracts the tree of a commit into destination using git archive.
// Path escapes are rejected; symbolic and hard links inside archives are
// unsupported (Spec §8.2).
func Archive(repo, commit, destination string) error {
	if err := EnsureRepo(repo); err != nil {
		return err
	}
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return err
	}
	cmd := exec.Command("git", "-C", repo, "archive", "--format=tar", commit) // #nosec G204 -- fixed binary and flags
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git archive failed in %s: %s", repo, strings.TrimSpace(stderr.String()))
	}
	return extractTar(&stdout, destination)
}

func extractTar(payload io.Reader, destination string) error {
	destRoot, err := filepath.Abs(destination)
	if err != nil {
		return err
	}
	reader := tar.NewReader(payload)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		name := header.Name
		if name == "" || strings.HasPrefix(name, "/") {
			return fmt.Errorf("unsafe path in git archive: %q", name)
		}
		target := filepath.Join(destRoot, filepath.FromSlash(name))
		rel, err := filepath.Rel(destRoot, target)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("unsafe path in git archive: %q", name)
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			mode := os.FileMode(header.Mode) & 0o777 // #nosec G115 -- tar mode bits fit
			if mode == 0 {
				mode = 0o644
			}
			file, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode) // #nosec G304 -- target is escape-checked above
			if err != nil {
				return err
			}
			if _, err := io.Copy(file, reader); err != nil { // #nosec G110 -- snapshot content is bounded by the repository
				_ = file.Close()
				return err
			}
			if err := file.Close(); err != nil {
				return err
			}
		case tar.TypeSymlink, tar.TypeLink:
			return fmt.Errorf("links in git archives are unsupported: %q", name)
		case tar.TypeXGlobalHeader:
			// pax global header (git writes one with the commit id); skip.
		default:
			return fmt.Errorf("unsupported entry type in git archive: %q", name)
		}
	}
}

// HasSubmodules reports whether a snapshot declares submodules, which are
// unsupported (Spec §8.2).
func HasSubmodules(snapshot string) bool {
	_, err := os.Stat(filepath.Join(snapshot, ".gitmodules"))
	return err == nil
}
