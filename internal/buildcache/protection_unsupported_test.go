//go:build !unix && !windows

package buildcache

import (
	"os"
	"testing"
)

func protectTestHome(*testing.T, string) {}

func TestUnsupportedPlatformFailsClosed(t *testing.T) {
	home := t.TempDir()
	store, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	input := testInput("tool")
	if result := store.Inspect(Expectation{Input: input}); result.Status != Unsupported {
		t.Fatalf("unsupported inspection = %+v", result)
	}
	publication, _ := testPublication(t, home, input, []byte("artifact"))
	before, err := os.ReadDir(home)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Publish(publication, testHomeLock{}); err == nil {
		t.Fatal("unsupported platform published persistent cache state")
	}
	after, err := os.ReadDir(home)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before) {
		t.Fatal("unsupported publication created persistent cache state")
	}
}
