package gitops

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

// The fixed metadata budget includes SHA-256 loose objects and the longer
// incremental MIDX bitmap. UTF-16 surrogate pairs count twice, not as UTF-8
// bytes or a single rune; the adjacent accepted/refused lengths are explicit.
func TestGitMetadataPathBudget(t *testing.T) {
	for _, tc := range []struct {
		root string
		want int
	}{
		{"root", 130},
		{"root\U0001f600", 132},
	} {
		actual, overhead := metadataPathLength(tc.root)
		if actual != tc.want || overhead != 126 {
			t.Fatalf("metadata budget for %q = (%d,%d), want (%d,126)", tc.root, actual, overhead, tc.want)
		}
		if err := checkMetadataPathLength(tc.root, tc.want); err != nil {
			t.Fatalf("exact bound refused: %v", err)
		}
		var refusal *MetadataPathError
		if err := checkMetadataPathLength(tc.root, tc.want-1); !errors.As(err, &refusal) || refusal.Actual != tc.want || refusal.Limit != tc.want-1 {
			t.Fatalf("want typed adjacent metadata refusal: %v", err)
		}
	}
	paths := strings.Join(metadataPaths(), "\n")
	for _, required := range []string{
		".git/objects/00/" + strings.Repeat("0", 62),
		".git/objects/pack/pack-" + strings.Repeat("0", 64) + ".keep",
		".git/objects/pack/multi-pack-index.d/multi-pack-index-" + strings.Repeat("0", 64) + ".bitmap",
	} {
		if !strings.Contains(paths, filepath.ToSlash(required)) {
			t.Fatalf("required metadata layout omitted: %s", required)
		}
	}
}
