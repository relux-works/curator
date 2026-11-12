package skillintegrity

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Missing members contradict only a complete roster. Observed additions and
// changed members contradict even a partial roster; an unchanged roster does not.
func TestObservedMismatchRequiresCompleteMissingRoster(t *testing.T) {
	a := File{Path: "a", Size: 1, SHA256: "original"}
	b := File{Path: "b", Size: 1, SHA256: "original"}
	changed := a
	changed.SHA256 = "changed"
	for _, tc := range []struct {
		name     string
		observed []File
		complete bool
		want     bool
	}{
		{"complete-unchanged", []File{a, b}, true, false},
		{"partial-unchanged", []File{a}, false, false},
		{"complete-missing", []File{a}, true, true},
		{"empty-complete", nil, true, true},
		{"empty-partial", nil, false, false},
		{"partial-changed", []File{changed}, false, true},
		{"partial-added", []File{{Path: "c"}}, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := observedMismatch(tc.observed, []File{a, b}, tc.complete); got != tc.want {
				t.Fatalf("roster contradiction: got=%v want=%v", got, tc.want)
			}
		})
	}
}

// A real completed walk retains its completeness across later timestamp-only
// instability. An interrupted walk stays partial and cannot establish absence.
// The Windows Verify removal regression separately exercises error aggregation.
func TestCaptureCompletenessSurvivesLateInstability(t *testing.T) {
	root, _ := fixture(t)
	budget := NewBudget()
	captured, err := captureTree(context.Background(), root, budget)
	if err != nil || !captured.complete {
		t.Fatalf("complete capture: %v", err)
	}
	expected := append([]File(nil), captured.files...)
	path := filepath.Join(root, "SKILL.md")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	changed := info.ModTime().Add(time.Hour)
	if err := os.Chtimes(path, changed, changed); err != nil {
		t.Fatal(err)
	}
	if err := captured.revalidate(context.Background(), root, budget); !IsCode(err, "skill_integrity_unavailable") || IsCode(err, "skill_integrity_mismatch") {
		t.Fatalf("timestamp-only instability: %v", err)
	}
	if !captured.complete || observedMismatch(captured.files, expected, captured.complete) {
		t.Fatal("late instability changed initial completeness or invented drift")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	partial, err := captureTree(ctx, root, NewBudget())
	if !IsCode(err, "skill_integrity_unavailable") || partial.complete || observedMismatch(partial.files, expected, partial.complete) {
		t.Fatalf("interrupted capture invented absence: %v", err)
	}
}
