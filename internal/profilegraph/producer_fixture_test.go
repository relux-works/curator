package profilegraph_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/privatedir"
)

func TestProducerFixtureWriterProtectsFiles(t *testing.T) {
	for _, rewrite := range []bool{false, true} {
		name := "create"
		if rewrite {
			name = "rewrite"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			const rel = "lock.json"
			full := filepath.Join(root, rel)
			if rewrite {
				// Seed an unprotected test-owned file. Writing it again with
				// mode 0600 alone cannot repair an existing mode or Windows DACL.
				if err := os.WriteFile(full, []byte("old"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(full, 0644); err != nil {
					t.Fatal(err)
				}
				if err := privatedir.ValidateFile(full); err == nil {
					t.Fatal("unprotected control unexpectedly validated")
				}
			}
			want := []byte("{\"fixture\":true}\n")
			writeFixture(t, root, rel, want)
			if err := privatedir.ValidateFile(full); err != nil {
				t.Fatalf("fixture writer did not establish native protection: %v", err)
			}
			raw, err := os.ReadFile(full)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(raw, want) {
				t.Fatalf("fixture bytes changed: got %q, want %q", raw, want)
			}
		})
	}
}
