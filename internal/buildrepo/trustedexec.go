package buildrepo

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// SystemExecutableCandidates is a fixed manager policy, never inherited PATH.
// Windows uses literal system/install roots rather than environment redirects.
func SystemExecutableCandidates(name string) []string {
	if runtime.GOOS == "windows" {
		if name == "ssh" {
			return []string{`C:\Windows\System32\OpenSSH\ssh.exe`}
		}
		return []string{`C:\Program Files\Git\cmd\git.exe`, `C:\Program Files\Git\bin\git.exe`}
	}
	return []string{"/usr/bin/" + name, "/bin/" + name, "/usr/local/bin/" + name, "/opt/homebrew/bin/" + name}
}

// SelectTrustedExecutable resolves operator policy or the fixed search list.
// Managed command roots are forbidden even through an operator-selected alias.
func SelectTrustedExecutable(name, configured string, forbiddenRoots ...string) (string, error) {
	candidates := SystemExecutableCandidates(name)
	if configured != "" {
		if !filepath.IsAbs(configured) {
			return "", fmt.Errorf("%s executable must be absolute", name)
		}
		candidates = []string{configured}
	}
	for _, candidate := range candidates {
		resolved, err := filepath.EvalSymlinks(candidate)
		if err != nil {
			continue
		}
		info, err := os.Stat(resolved)
		if err != nil || !info.Mode().IsRegular() || (runtime.GOOS != "windows" && info.Mode().Perm()&0111 == 0) {
			continue
		}
		for _, root := range forbiddenRoots {
			if root == "" {
				continue
			}
			if canonical, err := filepath.EvalSymlinks(root); err == nil {
				root = canonical
			}
			rel, err := filepath.Rel(root, resolved)
			if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				return "", fmt.Errorf("%s executable is inside a managed command root", name)
			}
		}
		// Project roots are recognized anywhere in the resolved path, independently
		// of cwd, so selecting another project's command cannot bypass this check.
		normalized := "/" + strings.ToLower(filepath.ToSlash(resolved)) + "/"
		for _, part := range []string{"/.agents/bin/", "/.curator/bin/", "/global/bin/", "/environments/", "/runtime/"} {
			if strings.Contains(normalized, part) {
				return "", fmt.Errorf("%s executable is inside a managed command root", name)
			}
		}
		return resolved, nil
	}
	return "", fmt.Errorf("no %s executable found", name)
}

// ExecutableFingerprint pins executable bytes for revalidation before use.
func ExecutableFingerprint(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return "", fmt.Errorf("executable is not a regular file")
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// DiscoveryEnvironment removes Git and credential selectors before trusted-tool probes.
func DiscoveryEnvironment(environment []string) []string {
	var clean []string
	for _, item := range environment {
		key, _, _ := strings.Cut(item, "=")
		if strings.HasPrefix(key, "GIT_") || strings.HasPrefix(key, "SSH_") {
			continue
		}
		clean = append(clean, item)
	}
	return clean
}

// OperatorManagedExecutableRoots includes configurable manager command roots.
// The value comes from operator state, never package or repository data.
func OperatorManagedExecutableRoots() []string { return []string{os.Getenv("CURATOR_GLOBAL_USER_BIN")} }
