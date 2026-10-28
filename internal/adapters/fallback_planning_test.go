package adapters

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Planning uses only its private staging directory. In particular, a file at
// the old predictable live-parent probe path must survive unchanged. Auto
// supplies a journal alternative, while explicit link/copy preserve their mode.
// This host supports symlinks; destination capability refusal is tested at the
// transaction syscall boundary, not simulated by probing this live parent.
func TestMirrorFallbackPrivatePlanning(t *testing.T) {
	for _, mode := range []string{"auto", "symlink", "copy"} {
		t.Run(mode, func(t *testing.T) {
			manager, user, stage := t.TempDir(), t.TempDir(), t.TempDir()
			canonical := filepath.Join(manager, "global", "skills")
			makeSkill(t, canonical, "skill-g")
			parent := filepath.Join(user, ".claude")
			if err := os.MkdirAll(parent, 0750); err != nil {
				t.Fatal(err)
			}
			sentinel := filepath.Join(parent, fmt.Sprintf(".curator-symlink-probe-%d", os.Getpid()))
			if err := os.WriteFile(sentinel, []byte("operator-owned sentinel"), 0640); err != nil {
				t.Fatal(err)
			}
			before, err := os.Lstat(parent)
			if err != nil {
				t.Fatal(err)
			}
			mirror, err := StageGlobal(stage, manager, user, []string{"claude_code"}, []string{"skill-g"}, mode, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			after, err := os.Lstat(parent)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(sentinel)
			if err != nil || string(raw) != "operator-owned sentinel" || !os.SameFile(before, after) || before.Mode() != after.Mode() || !before.ModTime().Equal(after.ModTime()) {
				t.Fatal("planning changed live parent or probe sentinel")
			}
			live := filepath.Join(parent, "skills", "skill-g")
			if _, err := os.Lstat(filepath.Dir(live)); !os.IsNotExist(err) {
				t.Fatalf("planning materialized live adapter: %v", err)
			}
			target, ok := find(mirror.Plan(), live)
			if !ok {
				t.Fatal("entry not journaled")
			}
			info, err := os.Lstat(target.StagedPath)
			if err != nil {
				t.Fatal(err)
			}
			if (info.Mode()&os.ModeSymlink != 0) != (mode != "copy") {
				t.Fatal("explicit mode lost")
			}
			if (target.CopyFallback != "") != (mode == "auto") {
				t.Fatal("copy alternative presence differs")
			}
			if target.CopyFallback != "" {
				rel, err := filepath.Rel(stage, target.CopyFallback)
				if err != nil || rel == ".." || filepath.IsAbs(rel) {
					t.Fatal("alternative escaped private staging")
				}
				if raw, err := os.ReadFile(filepath.Join(target.CopyFallback, "SKILL.md")); err != nil || string(raw) != "# skill-g" {
					t.Fatal("prepared alternative missing content")
				}
			}
		})
	}
}
