// Package gitignore enforces the managed .gitignore block (Spec §6.3).
//
// Generated paths must be ignored by git before installation proceeds; the
// check probes git check-ignore so any ignore mechanism counts.
package gitignore

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// BlockComment heads the appended entries.
const BlockComment = "# Curator"

const checkIgnoreEACCESRetryDelay = 100 * time.Millisecond

// checkIgnoreCommandRunner is the production command seam. Tests replace it
// to inject child-start errors without depending on runner filesystem state.
var checkIgnoreCommandRunner = func(cmd *exec.Cmd) error { return cmd.Run() }

// NotIgnoredError reports the policy outcome that entries are not ignored.
// It is distinct from a tool failure: a git that cannot be executed (a
// spawn error or a missing git binary) is returned as a plain wrapped
// error, never as this type. Git's own verdicts, including "not a
// repository", are policy outcomes as before.
type NotIgnoredError struct {
	Missing []string
}

func (e *NotIgnoredError) Error() string {
	return fmt.Sprintf("generated paths are not ignored by git; missing entries: %s", strings.Join(e.Missing, ", "))
}

// IsNotIgnored reports whether err is the not-ignored policy outcome as
// opposed to a git tool failure.
func IsNotIgnored(err error) bool {
	var target *NotIgnoredError
	return errors.As(err, &target)
}

// Missing returns the entries not currently ignored in projectRoot.
//
// A git that cannot be executed — a spawn error or exec.ErrNotFound, i.e.
// any cmd.Run() failure that is not an *exec.ExitError — is returned as an
// error carrying the tool diagnostic, so callers never mistake a broken git
// invocation for a policy outcome. Every exit status git itself reports
// keeps the pre-existing behaviour: the entry is reported as not ignored,
// whether the status is 1 (not ignored) or 128/other (not a repository, a
// fatal git error). The diagnostic names only the entry, never the full
// probe path, environment, or project root.
func Missing(projectRoot string, entries []string) ([]string, error) {
	var missing []string
	for _, entry := range entries {
		probe := strings.TrimSuffix(entry, "/") + "/.curator-probe"
		stderr, err := runCheckIgnoreWithEACCESRetry(projectRoot, probe)
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				missing = append(missing, entry)
				continue
			}
			if detail := strings.TrimSpace(stderr); detail != "" {
				return nil, fmt.Errorf("git check-ignore failed for %q: %s: %w", entry, detail, err)
			}
			return nil, fmt.Errorf("git check-ignore failed for %q: %w", entry, err)
		}
	}
	return missing, nil
}

// runCheckIgnoreWithEACCESRetry retries one child-start EACCES after a short
// delay. It does not retry Git exit statuses or other spawn errors; a second
// failure is returned so persistent permission errors still fail closed.
func runCheckIgnoreWithEACCESRetry(projectRoot, probe string) (string, error) {
	stderr, err := runCheckIgnoreAttempt(projectRoot, probe)
	if !isCheckIgnoreSpawnEACCES(err) {
		return stderr, err
	}

	firstErr := err
	time.Sleep(checkIgnoreEACCESRetryDelay)
	stderr, err = runCheckIgnoreAttempt(projectRoot, probe)
	if err != nil {
		return stderr, fmt.Errorf("git check-ignore still failed after one EACCES spawn retry (first attempt: %v): %w", firstErr, err)
	}
	return stderr, nil
}

func runCheckIgnoreAttempt(projectRoot, probe string) (string, error) {
	cmd := exec.Command("git", "-C", projectRoot, "check-ignore", "-q", probe) // #nosec G204 -- fixed binary and flags
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := checkIgnoreCommandRunner(cmd)
	return stderr.String(), err
}

func isCheckIgnoreSpawnEACCES(err error) bool {
	if err == nil {
		return false
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return false
	}
	var pathErr *os.PathError
	return errors.As(err, &pathErr) && pathErr.Op == "fork/exec" && errors.Is(pathErr.Err, syscall.EACCES)
}

// Ensure verifies the entries are ignored; with fix it appends the missing
// ones under the managed comment and re-checks.
func Ensure(projectRoot string, entries []string, fix bool) error {
	missing, err := Missing(projectRoot, entries)
	if err != nil {
		return err
	}
	if len(missing) == 0 {
		return nil
	}
	if fix {
		if err := Append(filepath.Join(projectRoot, ".gitignore"), missing); err != nil {
			return err
		}
		missing, err = Missing(projectRoot, entries)
		if err != nil {
			return err
		}
		if len(missing) == 0 {
			return nil
		}
	}
	return &NotIgnoredError{Missing: missing}
}

// Append adds entries to the .gitignore under the managed comment, skipping
// lines already present.
func Append(path string, entries []string) error {
	existing := ""
	// #nosec G304 -- the path is the caller's own project .gitignore.
	payload, err := os.ReadFile(path)
	if err == nil {
		existing = string(payload)
	}
	present := map[string]bool{}
	for _, line := range strings.Split(existing, "\n") {
		present[strings.TrimSpace(line)] = true
	}
	var toAdd []string
	for _, entry := range entries {
		if !present[entry] {
			toAdd = append(toAdd, entry)
		}
	}
	if len(toAdd) == 0 {
		return nil
	}
	prefix := ""
	if existing != "" && !strings.HasSuffix(existing, "\n") {
		prefix = "\n"
	}
	block := prefix + BlockComment + "\n" + strings.Join(toAdd, "\n") + "\n"
	return os.WriteFile(path, []byte(existing+block), 0o644)
}
