package install

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/stateread"
)

type generationErrorReader struct{ err error }

func (r generationErrorReader) InstalledMarker(string) (*marker.Marker, error) {
	return nil, r.err
}

func TestMarkerGenerationDistinguishesAbsentUnreadableAndInvalid(t *testing.T) {
	dir := t.TempDir()
	reader := markerGeneration{}
	if got, err := reader.InstalledMarker(dir); err != nil || got != nil {
		t.Fatalf("absent marker generation = (%v, %v), want nil, nil", got, err)
	}

	path := filepath.Join(dir, marker.Name)
	if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := reader.InstalledMarker(dir)
	var invalid *marker.InvalidError
	if got != nil || !errors.As(err, &invalid) {
		t.Fatalf("invalid marker generation = (%v, %v), want typed invalid marker so the caller can re-derive it", got, err)
	}

	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if got, err := reader.InstalledMarker(dir); got != nil || err == nil || !strings.Contains(err.Error(), stateread.DiagUnreadable) {
		t.Fatalf("unreadable marker generation = (%v, %v), want nil and typed %s error", got, err, stateread.DiagUnreadable)
	}
}

func TestMovedTagReaderTreatsInvalidMarkerAsStaleButKeepsReadFailure(t *testing.T) {
	node := &closure.Node{Name: "skill-a", Resolved: gitops.ResolvedRef{Kind: "tag", Ref: "v1", Commit: "new"}}

	invalidDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(invalidDir, marker.Name), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, invalidErr := marker.ReadState(invalidDir)
	var invalid *marker.InvalidError
	if !errors.As(invalidErr, &invalid) {
		t.Fatalf("invalid marker read error = %v, want typed invalid marker", invalidErr)
	}
	warnings, err := detectMovedTagsIn(t.TempDir(), []*closure.Node{node}, generationErrorReader{err: invalidErr})
	if err != nil || len(warnings) != 0 {
		t.Fatalf("invalid marker moved-tag result = (%v, %v), want no warning and re-derivation", warnings, err)
	}

	unreadableDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(unreadableDir, marker.Name), 0o755); err != nil {
		t.Fatal(err)
	}
	_, _, unreadableErr := marker.ReadState(unreadableDir)
	warnings, err = detectMovedTagsIn(t.TempDir(), []*closure.Node{node}, generationErrorReader{err: unreadableErr})
	if len(warnings) != 0 || err == nil || !strings.Contains(err.Error(), stateread.DiagUnreadable) {
		t.Fatalf("unreadable marker moved-tag result = (%v, %v), want typed %s refusal", warnings, err, stateread.DiagUnreadable)
	}
}
