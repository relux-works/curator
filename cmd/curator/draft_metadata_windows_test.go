//go:build windows

package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/testgitenv"
)

// The draft resolve/refresh caller preserves the clone's typed bound refusal
// without retry or cache-parent writes. The adjacent admitted roots complete
// real Git acquisition; endpoint exhaustion cannot replace the local error.
func TestWindowsDraftAcquisitionMetadataBoundary(t *testing.T) {
	source := testgitenv.ShortTempDir(t)
	runGit(t, source, "init", "-q")
	if err := os.WriteFile(filepath.Join(source, "payload"), []byte("locked"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-qm", "fixture")
	resolution := config.Resolution{Attempts: []config.Attempt{{URL: source}, {URL: source}}, Fallback: "availability-auth"}
	for _, length := range []int{132, 133, 134} {
		t.Run(fmt.Sprint(length), func(t *testing.T) {
			base := testgitenv.ShortTempDir(t)
			parent := filepath.Join(base, "new")
			repo := filepath.Join(parent, strings.Repeat("r", length-len(utf16.Encode([]rune(parent)))-1))
			var diagnostics bytes.Buffer
			err := fetchDraftRepoAllowingFallback(repo, resolution, "s", &diagnostics)
			if length > 133 {
				var refusal *gitops.MetadataPathError
				if !errors.As(err, &refusal) || refusal.RootLength != 134 || refusal.Actual != 260 || !strings.HasPrefix(err.Error(), "git_metadata_path_too_long:") || diagnostics.Len() != 0 {
					t.Fatalf("draft acquisition lost typed metadata refusal: %v diagnostics=%q", err, diagnostics.String())
				}
				if _, err := os.Lstat(parent); !os.IsNotExist(err) {
					t.Fatalf("draft refusal created cache parent: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if body, err := os.ReadFile(filepath.Join(repo, "payload")); err != nil || string(body) != "locked" {
				t.Fatalf("admitted draft bytes=%q error=%v", body, err)
			}
		})
	}
}
