package main

import (
	"bytes"
	"encoding/json"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/privatedir"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Empty CLI reads/previews and builtin removal leave an empty manager home
// untouched; the missing profile is a refusal, not implicit initialization.
func TestProfileReadPurityEmptyCLI(t *testing.T) {
	for _, tc := range []struct {
		name       string
		args       []string
		code       int
		diagnostic string
	}{
		{"list", []string{"profile", "list"}, exitOK, ""},
		{"status", []string{"env", "status", "--json"}, exitOK, ""},
		{"resolve", []string{"env", "resolve", "claude", "--profile", "default"}, exitFail, ""},
		{"preview", []string{"env", "resolve", "claude", "--profile", "default", "--repair", "--dry-run"}, exitFail, ""},
		{"remove-builtin", []string{"profile", "remove", "default"}, exitFail, "builtin_profile_remove_forbidden"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source, home := profileHome(t)
			code, _, stderr := runProfile(t, source, tc.args...)
			if code != tc.code || (tc.diagnostic != "" && !strings.Contains(stderr, tc.diagnostic)) {
				t.Fatalf("command exit=%d stderr=%q", code, stderr)
			}
			entries, err := os.ReadDir(home)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 0 {
				t.Fatalf("read-only command created manager state: %v", entries)
			}
		})
	}
}

// Leaf, parent, trailing-slash, and case-spelled aliases of the builtin are refused before the CLI changes any home entry or journal.
func TestProfileReadPurityBuiltinLeafAliasesCLI(t *testing.T) {
	type treeEntry struct {
		info    os.FileInfo
		content string
		link    string
	}
	snapshot := func(root string) (map[string]treeEntry, error) {
		out := map[string]treeEntry{}
		err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			entry := treeEntry{info: info}
			if info.Mode().IsRegular() {
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				entry.content = string(data)
			} else if info.Mode()&os.ModeSymlink != 0 {
				entry.link, walkErr = os.Readlink(path)
				if walkErr != nil {
					return walkErr
				}
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			out[rel] = entry
			return nil
		})
		return out, err
	}
	for _, tc := range []struct {
		name       string
		removeName string
		linkName   string
		parent     bool
	}{
		{name: "leaf", removeName: "linked", linkName: "linked"},
		{name: "leaf-trailing-slash", removeName: "linked/", linkName: "linked"},
		{name: "case-spelled-leaf", removeName: "DEFAULT", linkName: "DEFAULT"},
		{name: "parent-link", removeName: "alias/default", parent: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source, home := profileHome(t)
			if code, _, stderr := runProfile(t, source, "profile", "use", "default"); code != exitOK {
				t.Fatalf("initialize builtin exit=%d stderr=%q", code, stderr)
			}
			profiles := filepath.Join(home, "profiles")
			builtin := filepath.Join(profiles, "default")
			if tc.parent {
				if err := os.Symlink(profiles, filepath.Join(profiles, "alias")); err != nil {
					t.Skipf("parent-link fixture unavailable: %v", err)
				}
			} else {
				alias := filepath.Join(profiles, tc.linkName)
				if tc.name == "case-spelled-leaf" {
					if info, err := os.Stat(alias); err == nil {
						builtinInfo, statErr := os.Stat(builtin)
						if statErr != nil || !os.SameFile(info, builtinInfo) {
							t.Fatalf("case-spelled path already exists but is not the builtin: %v", statErr)
						}
					} else if !os.IsNotExist(err) {
						t.Fatal(err)
					} else if err := os.Symlink(builtin, alias); err != nil {
						t.Skipf("leaf-link fixture unavailable: %v", err)
					}
				} else if err := os.Symlink(builtin, alias); err != nil {
					t.Skipf("leaf-link fixture unavailable: %v", err)
				}
			}
			before, err := snapshot(home)
			if err != nil {
				t.Fatal(err)
			}
			code, _, stderr := runProfile(t, source, "profile", "remove", tc.removeName, "--purge")
			after, err := snapshot(home)
			if err != nil {
				t.Fatal(err)
			}
			unchanged := len(before) == len(after)
			for path, old := range before {
				current, ok := after[path]
				if !ok || !os.SameFile(old.info, current.info) || old.info.Mode() != current.info.Mode() || old.info.Size() != current.info.Size() || !old.info.ModTime().Equal(current.info.ModTime()) || old.content != current.content || old.link != current.link {
					unchanged = false
					break
				}
			}
			if !unchanged {
				t.Errorf("manager tree changed after builtin removal (entries %d -> %d)", len(before), len(after))
			}
			if code != exitFail || !strings.Contains(stderr, "builtin_profile_remove_forbidden") {
				t.Errorf("remove code=%d stderr=%q; want exact builtin_profile_remove_forbidden refusal", code, stderr)
			}
			if _, err := os.Stat(builtin); err != nil {
				t.Errorf("builtin target was not preserved: %v", err)
			}
		})
	}
}

// Explicit use creates default; listing that initialized profile succeeds
// without replacing its lock, and the profile record parents stay private.
func TestProfileReadPurityExplicitInitializationCLI(t *testing.T) {
	source, home := profileHome(t)
	if code, _, stderr := runProfile(t, source, "profile", "use", "default"); code != exitOK {
		t.Fatalf("explicit initialization exit=%d stderr=%q", code, stderr)
	}
	lock := filepath.Join(home, "profiles", "default", "lock.json")
	before, err := os.Stat(lock)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{filepath.Join(home, "profiles"), filepath.Join(home, "profiles", "default")} {
		if err := privatedir.Validate(path); err != nil {
			t.Fatalf("record parent is not private: %v", err)
		}
	}
	if code, out, stderr := runProfile(t, source, "profile", "list"); code != exitOK || !strings.Contains(out, "default") {
		t.Fatalf("initialized read exit=%d stdout=%q stderr=%q", code, out, stderr)
	}
	after, err := os.Stat(lock)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(before, after) || !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("listing replaced the initialized lock")
	}
}

// Unavailable verification is visibly unknown with its repair hint, and JSON
// keeps the existing current bool for every cryptographic verdict.
func TestProfileReadPuritySignerOutput(t *testing.T) {
	for _, verdict := range []string{"verified", "failed", "unavailable"} {
		row := envprofile.SourceSignerPosture{Profile: "sample", State: "enforced", Signer: "unknown", Current: verdict != "failed", Verification: verdict}
		if verdict == "unavailable" {
			row.Diagnostic = "signer_verification_unavailable"
			row.Hint = "OpenPGP backend unavailable; curator profile update sample"
		}
		var out bytes.Buffer
		printEnvStatus(&out, &envprofile.Status{SourceSignerPosture: []envprofile.SourceSignerPosture{row}})
		if verdict == "unavailable" && (!strings.Contains(out.String(), "verification=unknown") || !strings.Contains(out.String(), row.Hint)) {
			t.Fatalf("unknown status output = %q", out.String())
		}
		encoded, err := json.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		var wire map[string]json.RawMessage
		if err := json.Unmarshal(encoded, &wire); err != nil {
			t.Fatal(err)
		}
		if _, ok := wire["current"]; !ok {
			t.Fatal("current field omitted")
		}
		if string(wire["verification"]) != `"`+verdict+`"` {
			t.Fatal("verification enum changed")
		}
	}
}

// Valid update-all retains explicit builtin initialization; rejected update
// arguments fail before that mutation can create any manager state.
func TestProfileReadPurityUpdateAllInitialization(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		source, home := profileHome(t)
		if code, _, stderr := runProfile(t, source, "profile", "update", "--all"); code != exitOK {
			t.Fatalf("update all exit=%d stderr=%q", code, stderr)
		}
		if _, err := os.Stat(filepath.Join(home, "profiles", "default", "lock.json")); err != nil {
			t.Fatalf("explicit update did not initialize default: %v", err)
		}
	})
	t.Run("invalid", func(t *testing.T) {
		source, home := profileHome(t)
		if code, _, _ := runProfile(t, source, "profile", "update", "--all", "--revision", "invalid"); code != exitUsage {
			t.Fatalf("invalid update exit=%d", code)
		}
		entries, err := os.ReadDir(home)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 0 {
			t.Fatal("rejected update initialized manager state")
		}
	})
}
