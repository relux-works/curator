package gitops

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FetchLocalPathIsolated refreshes repo from the repository at the absolute
// filesystem location path, on the isolated lane FetchURLIsolated uses, with
// the same refspecs. Unlike FetchURLIsolated it hands git the path byte for
// byte: nothing is trimmed, cleaned, case-folded or decoded, because a
// significant byte (a final space, a tab, a newline) names another directory,
// and the sibling without it may be a different repository. A relative path
// is refused, since git reads a relative word as a configured remote name
// before a location; an absolute POSIX path cannot be a remote name.
func FetchLocalPathIsolated(repo, path string) error {
	if path == "" || !filepath.IsAbs(path) || strings.HasPrefix(path, "-") {
		return fmt.Errorf("refusing to fetch from local path %q: an absolute filesystem path is required", path)
	}
	if err := EnsureRepo(repo); err != nil {
		return err
	}
	env, cleanup, err := isolatedGitEnv()
	if err != nil {
		return err
	}
	defer cleanup()
	_, err = runWithEnv(repo, env, withIsolatedConfigArgs([]string{
		"fetch", "--force", "--prune", "--no-write-fetch-head", "--", path,
		"+refs/heads/*:refs/remotes/origin/*",
		"+refs/tags/*:refs/tags/*",
	})...)
	return err
}

// CopyCommitFromLocalPath transfers one exact commit from an existing source
// cache. A private ref makes an otherwise unreferenced object fetchable from
// the local object store; the same ref in repo keeps the copied pin reachable
// through Git maintenance.
func CopyCommitFromLocalPath(repo, path, commit string) error {
	if err := validateLockedCommit(commit); err != nil {
		return err
	}
	if path == "" || !filepath.IsAbs(path) || strings.HasPrefix(path, "-") {
		return fmt.Errorf("refusing to copy a locked commit from a non-absolute local path")
	}
	if err := EnsureRepo(repo); err != nil {
		return err
	}
	if err := EnsureRepo(path); err != nil {
		return fmt.Errorf("refusing to copy a locked commit from an invalid local source cache: %w", err)
	}
	ref := profilePinRef(commit)
	if err := RetainProfilePin(path, commit); err != nil {
		return fmt.Errorf("retain local source pin: %w", err)
	}
	args := []string{
		"fetch", "--depth=1", "--no-tags", "--no-write-fetch-head", "--no-auto-maintenance",
		"--recurse-submodules=no", "--refmap=", "--", path, "+" + ref + ":" + ref,
	}
	if _, err := run(repo, args...); err != nil {
		return fmt.Errorf("fetch locked commit from local source cache: %w", err)
	}
	return RetainProfilePin(repo, commit)
}

// RetainProfilePin anchors one exact commit under its own private ref so
// later pruning cannot discard an installed profile lock's object.
func RetainProfilePin(repo, commit string) error {
	if err := validateLockedCommit(commit); err != nil {
		return err
	}
	if err := EnsureRepo(repo); err != nil {
		return err
	}
	ref := profilePinRef(commit)
	zero := strings.Repeat("0", len(commit))
	if _, err := run(repo, "update-ref", ref, commit, zero); err == nil {
		return nil
	} else {
		current, checkErr := run(repo, "rev-parse", "--verify", ref+"^{commit}")
		if checkErr == nil && current == commit {
			return nil
		}
		if checkErr == nil {
			return fmt.Errorf("refusing to retarget profile pin ref %q", ref)
		}
		return fmt.Errorf("retain profile pin %s: %w", commit, err)
	}
}

func profilePinRef(commit string) string {
	return "refs/curator/profile-pins/" + commit
}

// runUntrimmed is run with git's standard output returned exactly as written.
// run trims it, which removes significant leading whitespace from a value
// read with --null; readers of configuration values must use this instead.
func runUntrimmed(dir string, args ...string) (string, error) {
	cmd := gitCommand(args...)
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
		return "", fmt.Errorf("git %s failed: %w", strings.Join(args, " "), &processError{cause: err, detail: detail})
	}
	return stdout.String(), nil
}
