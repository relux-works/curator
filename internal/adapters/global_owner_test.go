package adapters

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Verified profile entries are neither replaced nor claimed by the legacy
// ledger, including a previously double-claimed name retired from Skillfile.
// Other names still stage normally, and ownership errors refuse the plan.
func TestGlobalProfileOwnerPreservesEntries(t *testing.T) {
	for _, retired := range []bool{false, true} {
		t.Run(map[bool]string{false: "selected", true: "retired"}[retired], func(t *testing.T) {
			home, userHome, stage := t.TempDir(), t.TempDir(), t.TempDir()
			canonical := filepath.Join(home, "global", "skills")
			makeSkill(t, canonical, "shared-skill")
			makeSkill(t, canonical, "legacy-skill")
			root := filepath.Join(userHome, ".claude", "skills")
			writeLedgerFor(t, root, "shared-skill")
			shared := filepath.Join(root, "shared-skill")
			if err := os.WriteFile(shared, []byte("profile-owned preimage"), 0600); err != nil {
				t.Fatal(err)
			}
			skills := []string{"legacy-skill"}
			if !retired {
				skills = append(skills, "shared-skill")
			}
			owner := func(adapterRoot, name string) (bool, error) {
				return adapterRoot == root && name == "shared-skill", nil
			}
			mirror, err := StageGlobalWithOwnership(stage, home, userHome, []string{"claude_code"}, skills, "copy", nil, nil, owner)
			if err != nil {
				t.Fatal(err)
			}
			if _, found := find(mirror.Plan(), shared); found {
				t.Fatal("profile entry was replaced or removed")
			}
			if _, found := find(mirror.Plan(), filepath.Join(root, "legacy-skill")); !found {
				t.Fatal("legacy-only positive control did not stage")
			}
			ledger, found := find(mirror.Plan(), filepath.Join(root, LedgerName))
			if !found {
				t.Fatal("legacy ledger missing")
			}
			payload, err := os.ReadFile(ledger.StagedPath)
			if err != nil || strings.Contains(string(payload), "shared-skill") || !strings.Contains(string(payload), "legacy-skill") {
				t.Fatalf("profile entry claimed by legacy ledger: %s %v", payload, err)
			}
			if got, err := os.ReadFile(shared); err != nil || string(got) != "profile-owned preimage" {
				t.Fatal("staging changed profile entry")
			}
			sentinel := errors.New("ownership state unreadable")
			_, err = StageGlobalWithOwnership(t.TempDir(), home, userHome, []string{"claude_code"}, skills, "copy", nil, nil, func(string, string) (bool, error) { return false, sentinel })
			if !errors.Is(err, sentinel) {
				t.Fatalf("ownership error was ignored: %v", err)
			}
		})
	}
}

// Without verified profile ownership, a foreign directory refuses before
// publication and retains its bytes. The same request with no live entry is
// admitted, so rejecting everything cannot satisfy this negative control.
func TestGlobalProfileOwnerForeignRefuses(t *testing.T) {
	home, userHome := t.TempDir(), t.TempDir()
	canonical := filepath.Join(home, "global", "skills")
	makeSkill(t, canonical, "shared-skill")
	owner := func(string, string) (bool, error) { return false, nil }
	if _, err := StageGlobalWithOwnership(t.TempDir(), home, userHome, []string{"claude_code"}, []string{"shared-skill"}, "copy", nil, nil, owner); err != nil {
		t.Fatalf("absent-entry positive control refused: %v", err)
	}
	root := filepath.Join(userHome, ".claude", "skills")
	target := filepath.Join(root, "shared-skill")
	if err := os.MkdirAll(target, 0700); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(target, "SKILL.md")
	if err := os.WriteFile(file, []byte("foreign preimage"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := StageGlobalWithOwnership(t.TempDir(), home, userHome, []string{"claude_code"}, []string{"shared-skill"}, "copy", nil, nil, owner)
	if err == nil || !strings.Contains(err.Error(), "adapter target already exists and is not managed") {
		t.Fatalf("foreign adapter entry admitted: %v", err)
	}
	if got, err := os.ReadFile(file); err != nil || string(got) != "foreign preimage" {
		t.Fatal("foreign refusal changed its preimage")
	}
	if _, err := os.Lstat(filepath.Join(root, LedgerName)); !os.IsNotExist(err) {
		t.Fatalf("foreign refusal wrote a ledger: %v", err)
	}
}
