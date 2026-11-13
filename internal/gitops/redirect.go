package gitops

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RedirectError reports that git followed an HTTP redirect away from the
// endpoint the caller named. The transfer itself succeeded, but the server
// answered for a different location, so the named endpoint is not the
// repository's canonical address.
type RedirectError struct {
	Endpoint string // the endpoint the caller named
	Target   string // the location git reports it was redirected to
}

func (e *RedirectError) Error() string {
	return fmt.Sprintf("git was redirected from %s to %s", e.Endpoint, e.Target)
}

// redirectWarning is the line git's HTTP transport writes to standard error
// when it follows a redirect (remote-curl.c discover_refs). The operator lanes
// below pin the message locale so the line is never translated.
const redirectWarning = "warning: redirecting to "

// RedirectTarget returns the location named by git's "warning: redirecting
// to <url>" line in stderr, with a trailing slash removed. ok is false when
// stderr carries no such line or the line names no location.
func RedirectTarget(stderr string) (target string, ok bool) {
	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimRight(line, "\r")
		rest, found := strings.CutPrefix(line, redirectWarning)
		if !found {
			continue
		}
		target = strings.TrimRight(strings.TrimSpace(rest), "/")
		if target != "" {
			return target, true
		}
	}
	return "", false
}

// runCapturingStderr runs one git invocation on the operator transport, the
// environment run uses, and returns git's standard error from a successful
// run. The message locale is pinned to C so a redirect warning is matched in
// git's untranslated wording; locale does not change transport behaviour.
func runCapturingStderr(dir string, args ...string) (string, string, error) {
	cmd := gitCommand(args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(), "GIT_ALLOW_PROTOCOL="+AllowedProtocols, "LC_ALL=C")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = strings.TrimSpace(stdout.String())
		}
		return "", "", fmt.Errorf("git %s failed: %w", strings.Join(args, " "), &processError{cause: err, detail: detail})
	}
	return strings.TrimSpace(stdout.String()), stderr.String(), nil
}

// CloneRefusingRedirect is Clone, on the same operator transport, except
// that a clone git reports as redirected is removed again and refused with a
// *RedirectError naming the redirect target.
func CloneRefusingRedirect(remoteURL, destination string) error {
	var warnings string
	err := cloneWith(remoteURL, destination, func(dir string, args ...string) (string, error) {
		out, stderr, err := runCapturingStderr(dir, args...)
		warnings = stderr
		return out, err
	})
	if err != nil {
		return err
	}
	if target, redirected := RedirectTarget(warnings); redirected {
		_ = os.RemoveAll(destination)
		return &RedirectError{Endpoint: remoteURL, Target: target}
	}
	return nil
}

// FetchURLRefusingRedirect is FetchURL, on the same operator transport and
// with the same endpoint contract, except that a fetch git reports as
// redirected returns a *RedirectError naming the redirect target. The refs
// the redirected fetch already wrote are not rolled back; the caller refuses
// to use them.
func FetchURLRefusingRedirect(repo, remoteURL string) error {
	if err := ValidateFetchEndpoint(remoteURL); err != nil {
		return err
	}
	args, err := fetchURLArgs(repo, lowerURLScheme(remoteURL))
	if err != nil {
		return err
	}
	_, warnings, err := runCapturingStderr(repo, args...)
	if err != nil {
		return err
	}
	if target, redirected := RedirectTarget(warnings); redirected {
		return &RedirectError{Endpoint: remoteURL, Target: target}
	}
	return nil
}

// OriginURL reads the checkout's stored origin URL without network access.
// Values are read like the profile lane's cache binding: repository-local
// configuration with its includes expanded, NUL-delimited, before any
// url.<base>.insteadOf rewrite. present is false when no origin URL is
// configured. Several values, or a blank one, are refused, because git would
// fetch from only one of them.
func OriginURL(repo string) (originURL string, present bool, err error) {
	if err := EnsureRepo(repo); err != nil {
		return "", false, err
	}
	out, err := run(repo, "config", "--local", "--includes", "--null", "--get-all", "remote.origin.url")
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return "", false, nil
		}
		return "", false, err
	}
	values := strings.Split(strings.TrimSuffix(out, "\x00"), "\x00")
	if len(values) != 1 {
		return "", false, fmt.Errorf("checkout %s has %d origin URLs; exactly one is required", repo, len(values))
	}
	if strings.TrimSpace(values[0]) == "" {
		return "", false, fmt.Errorf("checkout %s has an empty origin URL", repo)
	}
	return values[0], true, nil
}

// RemoteConfigured reports, without network access, whether the checkout's
// repository-local configuration (includes expanded) defines any variable of
// a remote named exactly name. git reads a fetch argument that names such a
// remote as that remote, not as a location.
func RemoteConfigured(repo, name string) (bool, error) {
	if err := EnsureRepo(repo); err != nil {
		return false, err
	}
	out, err := run(repo, "config", "--local", "--includes", "--null", "--name-only", "--get-regexp", `^remote\.`)
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, err
	}
	for _, key := range strings.Split(out, "\x00") {
		// key is remote.<name>.<variable>; the subsection keeps its case.
		rest, ok := strings.CutPrefix(key, "remote.")
		if !ok {
			continue
		}
		if dot := strings.LastIndexByte(rest, '.'); dot > 0 && rest[:dot] == name {
			return true, nil
		}
	}
	return false, nil
}
