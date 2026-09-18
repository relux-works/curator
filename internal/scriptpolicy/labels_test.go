package scriptpolicy

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
	if !strings.Contains(joined, "script-worker-v1") {
		t.Fatalf("labels %q miss the script-worker-v1 posture", labels)
	}
	seen := map[string]bool{}
	for _, label := range labels {
		if label == "" || seen[label] {
			t.Fatalf("labels must be non-empty and unique: %q", labels)
		}
		seen[label] = true
	}
}
