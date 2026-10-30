package envprofile

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/claudememory"
	"github.com/relux-works/curator/internal/envregistry"
)

// Both home forms provision the native exclusion, repair its removal without
// dropping other settings/exclusions, and leave satisfied settings untouched.
func TestClaudeNativeMemorySettings(t *testing.T) {
	for _, form := range []string{envregistry.FormReferenced, envregistry.FormMonolithic} {
		t.Run(form, func(t *testing.T) {
			_, req, seed := claudeRepairFixture(t, form)
			path := filepath.Join(filepath.Dir(seed), "settings.json")
			native, err := claudememory.NativeFile()
			if err != nil {
				t.Fatal(err)
			}
			assert := func() {
				t.Helper()
				var settings struct {
					Excludes []string `json:"claudeMdExcludes"`
				}
				if err := json.Unmarshal(readClaudeState(t, path), &settings); err != nil {
					t.Fatal(err)
				}
				found := false
				for _, s := range settings.Excludes {
					found = found || s == filepath.ToSlash(native)
				}
				if !found {
					t.Fatal("native parent memory is not excluded")
				}
			}
			assert()
			before := []byte(`{"claudeMdExcludes":["**/vendor/CLAUDE.md"],"hooks":{"SessionStart":[]},"number":1e999}`)
			if err := os.WriteFile(path, before, 0600); err != nil {
				t.Fatal(err)
			}
			req.Repair = false
			if _, err := Resolve(req); err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
				t.Fatalf("missing exclusion did not make home stale: %v", err)
			}
			if !bytes.Equal(readClaudeState(t, path), before) {
				t.Fatal("read-only verification changed settings")
			}
			req.Repair = true
			if _, err := Resolve(req); err != nil {
				t.Fatal(err)
			}
			assert()
			after := readClaudeState(t, path)
			if !bytes.Contains(after, []byte(`"**/vendor/CLAUDE.md"`)) || !bytes.Contains(after, []byte(`"number":1e999`)) || !bytes.Contains(after, []byte(`"SessionStart":[]`)) {
				t.Fatal("repair changed unrelated settings")
			}
			stamp := time.Unix(1234567890, 0)
			if err := os.Chtimes(path, stamp, stamp); err != nil {
				t.Fatal(err)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(filepath.Join(filepath.Dir(seed), "CLAUDE.md")); err != nil {
				t.Fatal(err)
			}
			if _, err := Resolve(req); err != nil {
				t.Fatal(err)
			}
			next, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if !os.SameFile(info, next) || !info.ModTime().Equal(next.ModTime()) || !bytes.Equal(after, readClaudeState(t, path)) {
				t.Fatal("satisfied settings were rewritten")
			}
		})
	}
}

// Malformed existing settings stop repair before changing the root or seed;
// the valid provisioning and repair controls are in the neighboring test.
func TestClaudeNativeMemoryMalformedSettings(t *testing.T) {
	for _, input := range []string{`{`, `null`, `{"claudeMdExcludes":null}`, `{"claudeMdExcludes":[1]}`, ""} {
		t.Run(input, func(t *testing.T) {
			_, req, seed := claudeRepairFixture(t, envregistry.FormReferenced)
			root := filepath.Join(filepath.Dir(seed), "CLAUDE.md")
			beforeRoot, beforeSeed := readClaudeState(t, root), readClaudeState(t, seed)
			path := filepath.Join(filepath.Dir(seed), "settings.json")
			if err := os.WriteFile(path, []byte(input), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := Resolve(req); err == nil || !strings.Contains(err.Error(), envregistry.DiagSeedUnreadable) {
				t.Fatalf("malformed settings did not refuse: %v", err)
			}
			if string(readClaudeState(t, path)) != input || !bytes.Equal(beforeRoot, readClaudeState(t, root)) || !bytes.Equal(beforeSeed, readClaudeState(t, seed)) {
				t.Fatal("refusal changed existing settings or context")
			}
		})
	}
}

// An absent file can be created, while directory and symlink destinations must
// survive a refused publication without changing their target or staging files.
func TestClaudeNativeMemorySettingsFileKinds(t *testing.T) {
	for _, kind := range []string{"absent", "directory", "symlink", "dangling-symlink"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "settings.json")
			target := filepath.Join(t.TempDir(), "target.json")
			if err := os.WriteFile(target, []byte(`{"unrelated":true}`), 0600); err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "directory":
				if err := os.Mkdir(path, 0700); err != nil {
					t.Fatal(err)
				}
			case "symlink", "dangling-symlink":
				requireLinkCapability(t)
				link := target
				if kind == "dangling-symlink" {
					link += ".absent"
				}
				if err := os.Symlink(link, path); err != nil {
					t.Fatal(err)
				}
			}
			before, _ := os.Lstat(path)
			err := publishClaudeSettings(root, ".")
			if kind == "absent" {
				if err != nil {
					t.Fatal(err)
				}
				if plan, err := prepareClaudeSettings(root, "."); err != nil || plan.payload != nil {
					t.Fatalf("created settings do not satisfy exclusion: %v", err)
				}
			} else {
				after, statErr := os.Lstat(path)
				if err == nil || statErr != nil || !os.SameFile(before, after) || before.Mode() != after.Mode() {
					t.Fatalf("nonregular settings were not refused intact: %v, %v", err, statErr)
				}
			}
			if string(readClaudeState(t, target)) != `{"unrelated":true}` {
				t.Fatal("publication changed symlink target")
			}
			if staged, err := filepath.Glob(filepath.Join(root, ".curator-write-*.tmp")); err != nil || len(staged) != 0 {
				t.Fatalf("publication left staging files: %v, %v", staged, err)
			}
		})
	}
}
