package artifactpolicy

import (
	"strings"
	"testing"
)

func TestEffectiveLabels(t *testing.T) {
	labels := EffectiveLabels()
	if len(labels) == 0 {
		t.Fatalf("no effective labels")
	}
	joined := strings.Join(labels, "\n")
	for _, want := range []string{PolicyID, DetectorRegistryID, LimitVectorID} {
		if !strings.Contains(joined, want) {
			t.Fatalf("labels %q miss %q", labels, want)
		}
	}
	seen := map[string]bool{}
	for _, label := range labels {
		if label == "" || seen[label] {
			t.Fatalf("labels must be non-empty and unique: %q", labels)
		}
		seen[label] = true
	}
}
