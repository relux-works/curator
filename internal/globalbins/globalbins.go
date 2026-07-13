// Package globalbins publishes non-destructive forwarding shims for global
// commands into an existing user PATH directory. Canonical shims remain below
// the manager home; this package only owns entries recorded in its ledger.
package globalbins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/runtimestore"
)

const (
	// UserBinEnv explicitly selects a safe PATH-visible publication directory.
	UserBinEnv   = "CURATOR_GLOBAL_USER_BIN"
	managedFile  = ".curator-managed.json"
	ledgerSchema = 1
)

// Selection reports the selected user bin or why none can be used.
type Selection struct {
	Path    string
	Warning string
}

type ledger struct {
	SchemaVersion int      `json:"schema_version"`
	Entries       []string `json:"entries"`
}

// Refresh synchronizes forwarding shims for expected global commands.
// environment is injectable for tests; nil reads the current process.
func Refresh(home string, expected map[string]bool, platform string, environment map[string]string, userHome string) []string {
	if platform == "" {
		platform = runtimestore.Platform()
	}
	if userHome == "" {
		userHome, _ = os.UserHomeDir()
	}
	canonicalBin := filepath.Join(home, "global", "bin")
	selection := Select(home, platform, environment, userHome)
	if selection.Path == "" {
		if countExpected(expected) == 0 {
			return nil
		}
		if selection.Warning != "" {
			return []string{selection.Warning}
		}
		return []string{noSafeBinWarning(canonicalBin)}
	}
	target := selection.Path
	if err := os.MkdirAll(target, 0o755); err != nil {
		if countExpected(expected) == 0 {
			return nil
		}
		return []string{fmt.Sprintf(
			"global: command shims were installed in %s, but %s could not be created: %v; set %s to a writable PATH directory or use curator shell-init --install",
			canonicalBin, target, err, UserBinEnv,
		)}
	}

	managed := readLedger(target)
	nextManaged := map[string]bool{}
	var messages []string
	for name := range managed {
		if !expected[name] {
			published := shimPath(target, name, platform)
			canonical := shimPath(canonicalBin, name, platform)
			if !ownedTarget(published, canonical, platform) {
				messages = append(messages, fmt.Sprintf(
					"global: stale command %q was not removed from %s because the target no longer matches Curator ownership",
					name, target,
				))
				continue
			}
			if err := os.Remove(published); err != nil && !os.IsNotExist(err) {
				messages = append(messages, fmt.Sprintf("global: stale command %q could not be removed from %s: %v", name, target, err))
			}
		}
	}

	names := expectedNames(expected)
	for _, name := range names {
		canonical := shimPath(canonicalBin, name, platform)
		published := shimPath(target, name, platform)
		if unmanagedConflict(published, name, managed, canonical, platform) {
			messages = append(messages, fmt.Sprintf(
				"global: command %q was not published to %s; target exists and is not managed by Curator: %s",
				name, target, published,
			))
			continue
		}
		if _, err := runtimestore.WriteBinShim(target, name, canonical, platform); err != nil {
			messages = append(messages, fmt.Sprintf("global: command %q could not be published to %s: %v", name, target, err))
			continue
		}
		nextManaged[name] = true
	}
	if err := writeLedger(target, nextManaged); err != nil {
		messages = append(messages, fmt.Sprintf("global: user-bin ownership ledger could not be written in %s: %v", target, err))
	}
	if len(nextManaged) > 0 {
		messages = append(messages, "global: command shims published to "+target)
	}
	return messages
}

// Select chooses a safe writable user directory that is already on PATH.
func Select(home, platform string, environment map[string]string, userHome string) Selection {
	if userHome == "" {
		userHome, _ = os.UserHomeDir()
	}
	pathDirs := pathDirectories(envValue(environment, "PATH"), platform)
	explicit := strings.TrimSpace(envValue(environment, UserBinEnv))
	if explicit != "" {
		explicit = expandHome(explicit, userHome)
		switch {
		case disallowed(explicit, home, platform):
			return Selection{Warning: fmt.Sprintf(
				"global: %s points to a protected tool-manager or Curator directory: %s; choose a normal user bin or use curator shell-init --install",
				UserBinEnv, explicit,
			)}
		case !underHome(explicit, userHome, platform):
			return Selection{Warning: fmt.Sprintf(
				"global: %s must select a directory below the user home: %s",
				UserBinEnv, explicit,
			)}
		case !containsPath(pathDirs, explicit, platform):
			return Selection{Warning: fmt.Sprintf(
				"global: %s is not on PATH: %s; add it to PATH or choose an existing user bin",
				UserBinEnv, explicit,
			)}
		case !writableOrCreatable(explicit):
			return Selection{Warning: fmt.Sprintf(
				"global: %s is not writable: %s; choose a writable PATH directory or use curator shell-init --install",
				UserBinEnv, explicit,
			)}
		default:
			return Selection{Path: explicit}
		}
	}

	for _, preferred := range []string{
		filepath.Join(userHome, ".local", "bin"),
		filepath.Join(userHome, "bin"),
	} {
		if containsPath(pathDirs, preferred, platform) && safeExistingUserBin(preferred, userHome, home, platform) {
			return Selection{Path: preferred}
		}
	}
	for _, candidate := range pathDirs {
		if safeExistingUserBin(candidate, userHome, home, platform) {
			return Selection{Path: candidate}
		}
	}
	return Selection{Warning: noSafeBinWarning(filepath.Join(home, "global", "bin"))}
}

func envValue(environment map[string]string, name string) string {
	if environment == nil {
		return os.Getenv(name)
	}
	return environment[name]
}

func pathDirectories(value, platform string) []string {
	separator := string(os.PathListSeparator)
	if platform == "windows" {
		separator = ";"
	}
	var paths []string
	for _, raw := range strings.Split(value, separator) {
		if trimmed := strings.TrimSpace(raw); trimmed != "" {
			paths = append(paths, trimmed)
		}
	}
	return paths
}

func expandHome(path, home string) string {
	if path == "~" {
		return home
	}
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, `~\`) {
		return filepath.Join(home, path[2:])
	}
	return path
}

func safeExistingUserBin(path, userHome, managerHome, platform string) bool {
	if disallowed(path, managerHome, platform) || !underHome(path, userHome, platform) {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir() && info.Mode().Perm()&0o200 != 0
}

func writableOrCreatable(path string) bool {
	if info, err := os.Stat(path); err == nil {
		return info.IsDir() && info.Mode().Perm()&0o200 != 0
	}
	parent := filepath.Dir(path)
	for {
		info, err := os.Stat(parent)
		if err == nil {
			return info.IsDir() && info.Mode().Perm()&0o200 != 0
		}
		next := filepath.Dir(parent)
		if next == parent {
			return false
		}
		parent = next
	}
}

func disallowed(path, managerHome, platform string) bool {
	if samePath(path, filepath.Join(managerHome, "global", "bin"), platform) {
		return true
	}
	resolved := resolvedPath(path)
	parts := splitPathParts(resolved)
	has := func(value string) bool {
		for _, part := range parts {
			if strings.EqualFold(part, value) {
				return true
			}
		}
		return false
	}
	for _, protected := range []string{".agents", ".curator", ".venv", "venv", "venvs"} {
		if has(protected) {
			return true
		}
	}
	if has("mise") && (has("installs") || has("shims")) {
		return true
	}
	if has("shims") && (has(".asdf") || has(".pyenv") || has(".rbenv") || has("scoop")) {
		return true
	}
	return false
}

func splitPathParts(path string) []string {
	path = strings.ReplaceAll(path, `\`, "/")
	var result []string
	for _, part := range strings.Split(path, "/") {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func underHome(path, home, platform string) bool {
	resolved := resolvedPath(path)
	resolvedHome := resolvedPath(home)
	if samePath(resolved, resolvedHome, platform) {
		return true
	}
	relative, err := filepath.Rel(resolvedHome, resolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return false
	}
	return !filepath.IsAbs(relative)
}

func containsPath(paths []string, candidate, platform string) bool {
	for _, path := range paths {
		if samePath(path, candidate, platform) {
			return true
		}
	}
	return false
}

func samePath(left, right, platform string) bool {
	left = resolvedPath(left)
	right = resolvedPath(right)
	if platform == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}

func resolvedPath(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return filepath.Clean(resolved)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	absolute = filepath.Clean(absolute)
	current := absolute
	var suffix []string
	for {
		if resolved, err := filepath.EvalSymlinks(current); err == nil {
			for index := len(suffix) - 1; index >= 0; index-- {
				resolved = filepath.Join(resolved, suffix[index])
			}
			return filepath.Clean(resolved)
		}
		parent := filepath.Dir(current)
		if parent == current {
			return absolute
		}
		suffix = append(suffix, filepath.Base(current))
		current = parent
	}
}

func shimPath(binDir, name, platform string) string {
	if platform == "windows" && !strings.HasSuffix(strings.ToLower(name), ".cmd") {
		return filepath.Join(binDir, name+".cmd")
	}
	return filepath.Join(binDir, name)
}

func unmanagedConflict(path, name string, managed map[string]bool, canonical, platform string) bool {
	_, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return true
	}
	if managed[name] && ownedTarget(path, canonical, platform) {
		return false
	}
	// A ledger entry is necessary but not sufficient. This avoids adopting a
	// matching shim that a user created, and detects replacement of a formerly
	// managed path.
	return true
}

func ownedTarget(path, canonical, platform string) bool {
	if platform == "windows" {
		payload, err := os.ReadFile(path) // #nosec G304 -- path is a validated command below the selected user bin
		if err != nil {
			return false
		}
		expected := "@echo off\r\n\"" + canonical + "\" %*\r\n"
		return string(payload) == expected
	}
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		return false
	}
	target, err := os.Readlink(path)
	if err != nil {
		return false
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(path), target)
	}
	return samePath(target, canonical, platform)
}

func readLedger(binDir string) map[string]bool {
	payload, err := os.ReadFile(filepath.Join(binDir, managedFile)) // #nosec G304 -- binDir is selected from trusted manager state
	if err != nil {
		return map[string]bool{}
	}
	var value ledger
	if json.Unmarshal(payload, &value) != nil || value.SchemaVersion != ledgerSchema {
		return map[string]bool{}
	}
	managed := map[string]bool{}
	for _, entry := range value.Entries {
		if !identifiers.Valid(entry) || managed[entry] {
			return map[string]bool{}
		}
		managed[entry] = true
	}
	return managed
}

func writeLedger(binDir string, entries map[string]bool) error {
	names := expectedNames(entries)
	payload, err := json.MarshalIndent(ledger{SchemaVersion: ledgerSchema, Entries: names}, "", "  ")
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(binDir, ".curator-managed.*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if _, err := temporary.Write(append(payload, '\n')); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	target := filepath.Join(binDir, managedFile)
	if err := os.Rename(temporaryPath, target); err != nil {
		if removeErr := os.Remove(target); removeErr != nil && !os.IsNotExist(removeErr) {
			return err
		}
		return os.Rename(temporaryPath, target)
	}
	return nil
}

func expectedNames(expected map[string]bool) []string {
	names := make([]string, 0, len(expected))
	for name, active := range expected {
		if active {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func countExpected(expected map[string]bool) int {
	return len(expectedNames(expected))
}

func noSafeBinWarning(canonicalBin string) string {
	return fmt.Sprintf(
		"global: command shims were installed in %s, but no safe PATH-visible user bin was found; set %s to a writable PATH directory or use curator shell-init --install",
		canonicalBin, UserBinEnv,
	)
}
