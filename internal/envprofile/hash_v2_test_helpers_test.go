package envprofile

import (
	"testing"

	"github.com/relux-works/curator/internal/hashing"
)

func enableV2WritersForTest(t *testing.T) {
	t.Helper()
	prior := hashing.EnableV2Writers
	hashing.EnableV2Writers = true
	t.Cleanup(func() { hashing.EnableV2Writers = prior })
}

func enableV1WritersForTest(t *testing.T) {
	t.Helper()
	prior := hashing.EnableV2Writers
	hashing.EnableV2Writers = false
	t.Cleanup(func() { hashing.EnableV2Writers = prior })
}
