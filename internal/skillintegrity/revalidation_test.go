package skillintegrity

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// Revalidation applies the same hard path/kind admission after the initial
// capture. A root timestamp change must not hide a later forbidden entry.
func TestRevalidationKeepsLateBoundariesFatal(t *testing.T) {
	for _, kind := range []string{"symlink", "hardlink", "git-metadata", "oversize"} {
		t.Run(kind, func(t *testing.T) {
			root, _ := fixture(t)
			budget := NewBudget()
			captured, err := captureTree(context.Background(), root, budget)
			if err != nil {
				t.Fatal(err)
			}
			if err := captured.revalidate(context.Background(), root, budget); err != nil {
				t.Fatal(err)
			}
			want := "skill_integrity_boundary_invalid"
			switch kind {
			case "symlink":
				if err := os.Symlink("SKILL.md", filepath.Join(root, "late")); err != nil {
					t.Fatal(err)
				}
			case "hardlink":
				if err := os.Link(filepath.Join(root, "SKILL.md"), filepath.Join(root, "late")); err != nil {
					t.Fatal(err)
				}
			case "git-metadata":
				writeFixture(t, root, ".git/late", []byte("metadata"), 0o644)
			case "oversize":
				path := filepath.Join(root, "late")
				writeFixture(t, root, "late", nil, 0o644)
				if err := os.Truncate(path, MaxBytes+1); err != nil {
					t.Fatal(err)
				}
				want = "skill_integrity_manifest_invalid"
			}
			if err := captured.revalidate(context.Background(), root, budget); !IsCode(err, want) {
				t.Fatalf("late %s: %v", kind, err)
			}
		})
	}
}

// An admitted file addition/removal is a definite roster contradiction; a
// timestamp-only change is unavailable, since timestamps are not frozen identity.
func TestRevalidationDistinguishesRosterAndInstability(t *testing.T) {
	for _, kind := range []string{"add", "remove", "timestamp"} {
		t.Run(kind, func(t *testing.T) {
			root, _ := fixture(t)
			budget := NewBudget()
			captured, err := captureTree(context.Background(), root, budget)
			if err != nil {
				t.Fatal(err)
			}
			want := "skill_integrity_mismatch"
			switch kind {
			case "add":
				writeFixture(t, root, "late", []byte("new"), 0o644)
			case "remove":
				if err := os.Remove(filepath.Join(root, "references/guide.md")); err != nil {
					t.Fatal(err)
				}
			case "timestamp":
				path := filepath.Join(root, "SKILL.md")
				info, err := os.Stat(path)
				if err != nil {
					t.Fatal(err)
				}
				changed := info.ModTime().Add(time.Hour)
				if err := os.Chtimes(path, changed, changed); err != nil {
					t.Fatal(err)
				}
				want = "skill_integrity_unavailable"
			}
			err = captured.revalidate(context.Background(), root, budget)
			if !IsCode(err, want) || kind == "timestamp" && IsCode(err, "skill_integrity_mismatch") {
				t.Fatalf("%s: %v", kind, err)
			}
		})
	}
}

// A final-path link introduced after an earlier regular-file observation is
// a hard refusal, rather than an unavailable read that a lower level can allow.
func TestCapturedReadRejectsLateFinalLink(t *testing.T) {
	root, _ := fixture(t)
	path := filepath.Join(root, "references/guide.md")
	before, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../SKILL.md", path); err != nil {
		t.Fatal(err)
	}
	if _, _, err := readCapturedFile(path, before, MaxBytes); !IsCode(err, "skill_integrity_boundary_invalid") {
		t.Fatalf("late link: %v", err)
	}
}

// An unavailable early read does not suppress a later observed hard boundary or
// byte contradiction. This fixture needs actual Unix read-permission denial.
func TestCaptureContinuesAfterUnavailableRead(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("fixture requires Unix permission denial without root bypass")
	}
	for _, kind := range []string{"link", "changed-file"} {
		t.Run(kind, func(t *testing.T) {
			root, id := fixture(t)
			writeFixture(t, root, "000-unreadable", []byte("denied"), 0o600)
			expected, err := Build(context.Background(), root, id, nil)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(root, "000-unreadable")
			if err := os.Chmod(path, 0); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(path, 0o600) })
			if _, err := os.ReadFile(path); !os.IsPermission(err) {
				t.Fatalf("permission plant did not apply: %v", err)
			}
			want := "skill_integrity_boundary_invalid"
			if kind == "link" {
				if err := os.Symlink("SKILL.md", filepath.Join(root, "zzz-link")); err != nil {
					t.Fatal(err)
				}
			} else {
				writeFixture(t, root, "references/guide.md", []byte("changed"), 0o644)
				want = "skill_integrity_mismatch"
			}
			err = Verify(context.Background(), root, id, expected, nil)
			if !IsCode(err, "skill_integrity_unavailable") || !IsCode(err, want) {
				t.Fatalf("%s: %v", kind, err)
			}
		})
	}
}

// Valid expected evidence distinguishes installed parse-breaking drift from bad
// supplied evidence. Missing SKILL and invalid native bytes are mismatches;
// the unmodified native neighbor verifies and malformed expectations refuse.
func TestExpectedEvidenceClassifiesParseBreakingDrift(t *testing.T) {
	for _, mode := range []string{"intact", "missing-skill", "invalid-native", "invalid-expectation"} {
		t.Run(mode, func(t *testing.T) {
			root, id := fixture(t)
			writeFixture(t, root, "agent-skill.json", []byte(syntheticDeclaration), 0644)
			expected, err := Build(context.Background(), root, id, nil)
			if err != nil {
				t.Fatal(err)
			}
			want := ""
			switch mode {
			case "missing-skill":
				if err := os.Remove(filepath.Join(root, "SKILL.md")); err != nil {
					t.Fatal(err)
				}
				want = "skill_integrity_mismatch"
			case "invalid-native":
				writeFixture(t, root, "agent-skill.json", []byte("invalid declaration"), 0644)
				want = "skill_integrity_mismatch"
			case "invalid-expectation":
				expected.Manifest.Schema = "unsupported"
				want = "skill_integrity_unsupported_version"
			}
			err = Verify(context.Background(), root, id, expected, nil)
			if want == "" {
				if err != nil {
					t.Fatalf("intact neighbor: %v", err)
				}
			} else if !IsCode(err, want) {
				t.Fatalf("installed drift classification: want %s: %v", want, err)
			}
		})
	}
}
