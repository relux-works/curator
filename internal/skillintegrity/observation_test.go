package skillintegrity

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Identity observation accepts the hardlink, then admission gives the named
// fatal refusal. Removing it restores the intact positive neighbor.
func TestSkillHardlinkObservationKeepsNamedRefusal(t *testing.T) {
	root, _ := fixture(t)
	path := filepath.Join(root, "SKILL.md")
	if err := ValidateLegacyBoundary(root); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "linked.md")
	if err := os.Link(path, link); err != nil {
		t.Fatal(err)
	}
	entry, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range entry {
		if e.Name() != "SKILL.md" {
			continue
		}
		found = true
		info, err := captureEntryInfo(path, e)
		if err != nil {
			t.Fatalf("identity observation prematurely rejected hardlink: %v", err)
		}
		if !info.Mode().IsRegular() {
			t.Fatal("observed hardlink is not regular")
		}
	}
	if !found {
		t.Fatal("fixture entry missing")
	}
	err = ValidateLegacyBoundary(root)
	if !IsCode(err, "skill_integrity_boundary_invalid") || !strings.Contains(err.Error(), path) || !strings.Contains(err.Error(), "single-link regular files") {
		t.Fatalf("hardlink missed named admission refusal: %v", err)
	}
	t.Logf("hardlink observation reached named admission refusal on %s: %v", runtime.GOOS, err)
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	if err := ValidateLegacyBoundary(root); err != nil {
		t.Fatalf("intact neighbor refused: %v", err)
	}
}
