package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/relux-works/curator/internal/globalbins"
)

// Umbrella subcommand discovery (environments §11): a CLI subcommand the
// manager does not implement resolves to an executable named curator-<name>
// on PATH and is executed with the remaining arguments verbatim — the git,
// kubectl, and docker external-subcommand convention. The manager carries
// no knowledge of any provider: no provider registry, no provider-specific
// flags, no version coupling. The first providers are curator-run (the
// launcher, its own specification) and curator-session (a shim to the
// agent session manager); neither ships here.

// findProvider resolves curator-<name> on PATH. Profile, marker, and
// fragment data never influence the dispatched name, the resolved path,
// or the argument vector: dispatch input is operator argv and PATH alone.
// A provider inside a manager-published or managed directory — the
// user-bin shim directory, or any directory below the environments root —
// is refused with subcommand_provider_untrusted, so profile-materialized
// files cannot poison the PATH the dispatch trusts.
func findProvider(home, name string) (string, error) {
	path, err := exec.LookPath("curator-" + name)
	if err != nil {
		return "", fmt.Errorf("subcommand_provider_missing: no curator-%s on PATH: install the %s provider; nothing is downloaded or installed implicitly", name, name)
	}
	if untrusted, dir := providerUntrustedDir(home, path); untrusted {
		return "", fmt.Errorf("subcommand_provider_untrusted: curator-%s resolved inside %s (%s)", name, dir, path)
	}
	return path, nil
}

// providerUntrustedDir reports whether the resolved provider lives in a
// directory the manager itself publishes onto PATH.
func providerUntrustedDir(home, path string) (bool, string) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return false, ""
	}
	// Resolve through the link: a provider reached via a shim symlink
	// into a managed directory is still a managed directory.
	if resolved, err := filepath.EvalSymlinks(absolute); err == nil {
		absolute = resolved
	}
	dir := filepath.Dir(absolute)
	userHome, _ := os.UserHomeDir()
	selection := globalbins.Select(home, runtime.GOOS, pathEnvironment(), userHome)
	if selection.Path != "" && sameDir(dir, selection.Path) {
		return true, "the user-bin shim directory"
	}
	if underDir(absolute, filepath.Join(home, "environments")) {
		return true, "the environments root"
	}
	return false, ""
}

// pathEnvironment carries the process PATH to the selector.
func pathEnvironment() map[string]string {
	return map[string]string{"PATH": os.Getenv("PATH")}
}

func sameDir(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}

func underDir(path, root string) bool {
	clean := filepath.Clean(root)
	return path == clean || strings.HasPrefix(path, clean+string(filepath.Separator))
}

// cmdUmbrella executes the resolved provider with the remaining arguments
// verbatim and propagates its exit code.
func (c cli) cmdUmbrella(home string, args []string) int {
	path, err := findProvider(home, args[0])
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	command := exec.Command(path, args[1:]...) // #nosec G204,G702 -- §11 dispatches the LookPath-resolved provider with operator argv verbatim
	command.Stdin = os.Stdin
	command.Stdout = c.stdout
	command.Stderr = c.stderr
	if err := command.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			return exit.ExitCode()
		}
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	return exitOK
}
