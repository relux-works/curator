package envprofile

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/envregistry"
)

// These native HOME values are references only; the managed fixture supplies
// separate temporary native adapter files. Native memory is never opened.
// Explicit expected literals avoid depending on production quoting or temp roots.
type claudeMemoryHomeCase struct{ form, name, first, next, firstPattern, nextPattern string }

func claudeMemoryHomeCases(t *testing.T) []claudeMemoryHomeCase {
	t.Helper()
	base := filepath.ToSlash(filepath.VolumeName(t.TempDir())+string(filepath.Separator)) + "curator-memory-fixture/"
	var cases []claudeMemoryHomeCase
	for _, form := range []string{envregistry.FormReferenced, envregistry.FormMonolithic} {
		cases = append(cases,
			claudeMemoryHomeCase{form, "plain", filepath.FromSlash(base + "operator"), filepath.FromSlash(base + "next-operator"), base + "operator/.claude/CLAUDE.md", base + "next-operator/.claude/CLAUDE.md"},
			claudeMemoryHomeCase{form, "metacharacter", filepath.FromSlash(base + "operator[1]"), filepath.FromSlash(base + "next-operator[2]"), base + "operator[[]1[]]/.claude/CLAUDE.md", base + "next-operator[[]2[]]/.claude/CLAUDE.md"},
		)
	}
	return cases
}

// Both home forms provision the native exclusion, repair its removal without
// dropping other settings/exclusions, and leave satisfied settings untouched.
func TestClaudeNativeMemorySettings(t *testing.T) {
	for _, tc := range claudeMemoryHomeCases(t) {
		t.Run(tc.form+"/"+tc.name, func(t *testing.T) {
			t.Setenv("HOME", tc.first)
			t.Setenv("USERPROFILE", tc.first)
			_, req, seed := claudeRepairFixture(t, tc.form)
			path := filepath.Join(filepath.Dir(seed), "settings.json")
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
					found = found || s == tc.firstPattern
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

// Changing the operator home makes its old exclusion stale without read-only
// writes; repair preserves that exclusion and admits the new home in both forms.
func TestClaudeNativeMemoryChangedHome(t *testing.T) {
	for _, tc := range claudeMemoryHomeCases(t) {
		t.Run(tc.form+"/"+tc.name, func(t *testing.T) {
			firstHome := tc.first
			t.Setenv("HOME", firstHome)
			t.Setenv("USERPROFILE", firstHome)
			_, req, seed := claudeRepairFixture(t, tc.form)
			req.Repair = false
			if _, err := Resolve(req); err != nil {
				t.Fatalf("original home positive: %v", err)
			}
			root := filepath.Join(filepath.Dir(seed), "CLAUDE.md")
			settings := filepath.Join(filepath.Dir(seed), "settings.json")
			beforeRoot, beforeSeed, beforeSettings := readClaudeState(t, root), readClaudeState(t, seed), readClaudeState(t, settings)
			nextHome := tc.next
			t.Setenv("HOME", nextHome)
			t.Setenv("USERPROFILE", nextHome)
			if _, err := Resolve(req); err == nil || !strings.Contains(err.Error(), DiagHomeStale) || !strings.Contains(err.Error(), "native memory exclusion") {
				t.Fatalf("changed home must refuse the stale exclusion: %v", err)
			}
			if !bytes.Equal(beforeRoot, readClaudeState(t, root)) || !bytes.Equal(beforeSeed, readClaudeState(t, seed)) || !bytes.Equal(beforeSettings, readClaudeState(t, settings)) {
				t.Fatal("read-only changed-home refusal wrote context, seed or settings")
			}
			req.Repair = true
			if _, err := Resolve(req); err != nil {
				t.Fatal(err)
			}
			req.Repair = false
			if _, err := Resolve(req); err != nil {
				t.Fatalf("repaired home positive: %v", err)
			}
			var got struct {
				Excludes []string `json:"claudeMdExcludes"`
			}
			if err := json.Unmarshal(readClaudeState(t, settings), &got); err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{tc.firstPattern, tc.nextPattern} {
				found := false
				for _, exclude := range got.Excludes {
					found = found || exclude == want
				}
				if !found {
					t.Fatalf("repair did not retain exclusion %q", want)
				}
			}
		})
	}
}
