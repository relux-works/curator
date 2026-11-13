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
