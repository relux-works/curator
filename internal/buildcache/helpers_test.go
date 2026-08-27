package buildcache

import (
	"os"
	"testing"
)

func TestClosedHandlesFailReadHelpers(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "closed-")
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := readExactFile(file); err == nil {
		t.Fatal("receipt reader accepted a closed handle")
	}
	if _, _, err := hashOpenFile(file); err == nil {
		t.Fatal("artifact hasher accepted a closed handle")
	}
	if _, err := directoryNames(file); err == nil {
		t.Fatal("directory reader accepted a closed handle")
	}
	var opened *openedEntry
	opened.close()
}
