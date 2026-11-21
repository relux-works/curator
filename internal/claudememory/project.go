package claudememory

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/stateread"
)

// ProjectRoot resolves the actual launch checkout, ignoring ambient Git
// redirection. Outside Git the launch directory itself is the project root.
// Broken repository metadata refuses instead of widening the ancestor policy.
func ProjectRoot(launchDir string) (string, error) {
	if !filepath.IsAbs(launchDir) {
		return "", fmt.Errorf("claude launch directory must be absolute")
	}
	launchDir = filepath.Clean(launchDir)
	canonical, err := filepath.EvalSymlinks(launchDir)
	if err != nil {
		return "", fmt.Errorf("resolve Claude launch directory: %w", err)
	}
	info, err := os.Stat(canonical)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("claude launch directory must be an existing directory")
	}
	gitPresent, gitEntryPresent := false, false
	for dir := canonical; ; dir = filepath.Dir(dir) {
		gitEntry, err := stateread.Lstat(filepath.Join(dir, ".git"))
		if err != nil {
			return "", fmt.Errorf("inspect launch repository metadata: %w", err)
		}
		if gitEntry.Kind == stateread.KindPresent {
			gitPresent = true
			gitEntryPresent = true
			break
		}
		// Git also discovers bare repositories. They have no working-tree
		// toplevel and must refuse, not masquerade as an outside-Git launch.
		head, err := stateread.Lstat(filepath.Join(dir, "HEAD"))
		if err != nil {
			return "", fmt.Errorf("inspect bare Git HEAD: %w", err)
		}
		if head.Kind == stateread.KindPresent && head.Info != nil && head.Info.Mode().IsRegular() {
			objects, err := stateread.Stat(filepath.Join(dir, "objects"))
			if err != nil {
				return "", fmt.Errorf("inspect bare Git objects: %w", err)
			}
			if objects.Kind == stateread.KindPresent && objects.Info != nil && objects.Info.IsDir() {
				gitPresent = true
				break
			}
		}
		if filepath.Dir(dir) == dir {
			break
		}
	}
	if !gitPresent {
		return launchDir, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "--no-optional-locks", "--no-replace-objects", "-c", "core.fsmonitor=false", "-c", "core.hooksPath="+os.DevNull, "-C", canonical, "rev-parse", "--show-toplevel") // #nosec G204 -- fixed executable and read-only arguments, no shell.
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if !strings.HasPrefix(strings.ToUpper(key), "GIT_") && key != "SSH_AUTH_SOCK" {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	cmd.Env = append(cmd.Env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_TERMINAL_PROMPT=0", "GIT_NO_LAZY_FETCH=1")
	var output bytes.Buffer
	cmd.Stdout = &memoryRootOutput{buffer: &output}
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if !gitEntryPresent && ctx.Err() == nil && errors.As(err, &exit) && exit.ExitCode() == 128 {
			// HEAD/objects names alone do not prove a repository. Let Git
			// distinguish an actual bare repository from an ordinary directory
			// with those legitimate names, without parsing localized stderr.
			args := append([]string{}, cmd.Args[1:]...)
			args[len(args)-1] = "--is-bare-repository"
			probe := exec.CommandContext(ctx, cmd.Path, args...) // #nosec G204 -- same resolved Git executable and fixed read-only query.
			probe.Env, probe.Stderr = cmd.Env, io.Discard
			probe.Stdout = &memoryRootOutput{buffer: &bytes.Buffer{}}
			probeErr := probe.Run()
			if ctx.Err() == nil && errors.As(probeErr, &exit) && exit.ExitCode() == 128 {
				return launchDir, nil
			}
		}
		return "", fmt.Errorf("resolve launch Git toplevel failed")
	}
	root := strings.TrimSuffix(output.String(), "\n")
	if !filepath.IsAbs(root) {
		return "", fmt.Errorf("launch Git toplevel is not absolute")
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil || !withinMemoryRoot(root, canonical) {
		return "", fmt.Errorf("launch Git toplevel cannot be verified")
	}
	// Retain a lexical root alias when one exists on the launch path. An alias
	// directly into a nested directory need not provide a spelling for its
	// containing Git root; in that case use the verified canonical root.
	for dir := launchDir; ; dir = filepath.Dir(dir) {
		resolved, err := filepath.EvalSymlinks(dir)
		if err != nil {
			return "", fmt.Errorf("verify lexical project root: %w", err)
		}
		if resolved == root {
			return dir, nil
		}
		if filepath.Dir(dir) == dir {
			break
		}
	}
	return root, nil
}

type memoryRootOutput struct{ buffer *bytes.Buffer }

func (out *memoryRootOutput) Write(payload []byte) (int, error) {
	if out.buffer.Len()+len(payload) > 65536 {
		return 0, fmt.Errorf("git toplevel output exceeds bound")
	}
	return out.buffer.Write(payload)
}
