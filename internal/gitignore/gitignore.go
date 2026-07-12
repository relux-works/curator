// Package gitignore enforces the managed .gitignore block (Spec §6.3).
//
// Generated paths must be ignored by git before installation proceeds; the
// check probes git check-ignore so any ignore mechanism counts.
package gitignore

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BlockComment heads the appended entries.
const BlockComment = "# Curator"

// Missing returns the entries not currently ignored in projectRoot.
func Missing(projectRoot string, entries []string) ([]string, error) {
	var missing []string
	for _, entry := range entries {
		probe := strings.TrimSuffix(entry, "/") + "/.curator-probe"
		cmd := exec.Command("git", "-C", projectRoot, "check-ignore", "-q", probe) // #nosec G204 -- fixed binary and flags
		if err := cmd.Run(); err != nil {
			missing = append(missing, entry)
		}
	}
	return missing, nil
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
	return fmt.Errorf("generated paths are not ignored by git; missing entries: %s", strings.Join(missing, ", "))
}

// Append adds entries to the .gitignore under the managed comment, skipping
// lines already present.
func Append(path string, entries []string) error {
	existing := ""
	if payload, err := os.ReadFile(path); err == nil { // #nosec G304 -- project .gitignore
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
	return os.WriteFile(path, []byte(existing+block), 0o644) // #nosec G306 -- .gitignore is not a secret
}
