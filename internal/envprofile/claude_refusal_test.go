package envprofile

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

type claudeRefusalEntry struct {
	info fs.FileInfo
	data []byte
	link string
}

// Snapshot every entry, not only .claude.json. Reading can update atime;
// the contract here covers bytes, inode, mode and mtime, not atime/ctime.
func claudeRefusalSnapshot(t *testing.T, root string) map[string]claudeRefusalEntry {
	t.Helper()
	result := map[string]claudeRefusalEntry{}
	err := filepath.WalkDir(root, func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		entry := claudeRefusalEntry{info: info}
		if info.Mode()&os.ModeSymlink != 0 {
			entry.link, err = os.Readlink(path)
		} else if info.Mode().IsRegular() && info.Mode().Perm()&0400 != 0 {
			entry.data, err = os.ReadFile(path)
		}
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		result[rel] = entry
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func assertClaudeRefusalSnapshot(t *testing.T, before, after map[string]claudeRefusalEntry) {
	t.Helper()
	for path, old := range before {
		now, ok := after[path]
		if !ok || !os.SameFile(old.info, now.info) || old.info.Mode() != now.info.Mode() || !old.info.ModTime().Equal(now.info.ModTime()) || !bytes.Equal(old.data, now.data) || old.link != now.link {
			t.Errorf("refusal changed managed entry %s", path)
		}
	}
	for path := range after {
		if _, ok := before[path]; !ok {
			t.Errorf("refusal left new managed entry %s", path)
		}
	}
}

// Real Resolve on synthetic homes: invalid seed state must refuse before
// any managed entry changes. The valid neighbor must actually repair drift.
func TestClaudePreserveLockRefusalIsZeroWrite(t *testing.T) {
	for _, form := range []string{envregistry.FormReferenced, envregistry.FormMonolithic} {
		for _, variant := range []string{"malformed", "array", "null", "scalar", "symlink", "directory", "unreadable", "trailing-object", "trailing-garbage", "valid", "valid-whitespace"} {
			t.Run(form+"/"+variant, func(t *testing.T) {
				if variant == "unreadable" && (runtime.GOOS == "windows" || os.Geteuid() == 0) {
					t.Skip("POSIX mode-bit unreadability")
				}
				fx, req, path := claudeRepairFixture(t, form)
				req.PreserveLock = true
				root := ManagedParent(fx.home, fx.profile, envregistry.ClaudeCode)
				contextPath := filepath.Join(filepath.Dir(path), "CLAUDE.md")
				if err := os.WriteFile(contextPath, []byte("deliberate context drift\n"), 0600); err != nil {
					t.Fatal(err)
				}
				payload := []byte(`{"unknown":1e-7}`)
				diagnostic := envregistry.DiagSeedUnreadable
				switch variant {
				case "trailing-object":
					payload = []byte(`{} {}`)
				case "trailing-garbage":
					payload = []byte(`{}x`)
				case "valid-whitespace":
					payload = []byte("{\"unknown\":1e-7} \t\r\n")
				case "malformed":
					payload = []byte(`{"broken":`)
				case "array":
					payload = []byte(`[]`)
				case "null":
					payload = []byte(`null`)
				case "scalar":
					payload = []byte(`42`)
				}
				if err := os.WriteFile(path, payload, 0600); err != nil {
					t.Fatal(err)
				}
				if variant == "symlink" || variant == "directory" {
					if err := os.Remove(path); err != nil {
						t.Fatal(err)
					}
					if variant == "symlink" {
						target := filepath.Join(t.TempDir(), "state.json")
						if err := os.WriteFile(target, payload, 0600); err != nil {
							t.Fatal(err)
						}
						if err := os.Symlink(target, path); err != nil {
							t.Fatal(err)
						}
						diagnostic = DiagWriteWouldFollowLink
					} else if err := os.Mkdir(path, 0700); err != nil {
						t.Fatal(err)
					}
				}
				if variant == "unreadable" {
					if err := os.Chmod(path, 0); err != nil {
						t.Fatal(err)
					}
					defer func() { _ = os.Chmod(path, 0600) }()
				}
				before := claudeRefusalSnapshot(t, root)
				lockBefore := readClaudeState(t, lockPath(fx.home, fx.profile))
				_, err := Resolve(req)
				if strings.HasPrefix(variant, "valid") {
					if err != nil {
						t.Fatal(err)
					}
					if bytes.Equal(readClaudeState(t, contextPath), []byte("deliberate context drift\n")) {
						t.Fatal("valid control did not repair context")
					}
					req.Repair = false
					req.PreserveLock = false
					if _, err := Resolve(req); err != nil {
						t.Fatalf("verify valid repair: %v", err)
					}
				} else {
					if err == nil || !strings.Contains(err.Error(), diagnostic) {
						t.Errorf("Resolve = %v, want %s refusal", err, diagnostic)
					}
					assertClaudeRefusalSnapshot(t, before, claudeRefusalSnapshot(t, root))
					if variant == "unreadable" {
						if err := os.Chmod(path, 0600); err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(readClaudeState(t, path), payload) {
							t.Error("refusal changed unreadable bytes")
						}
					}
				}
				if !bytes.Equal(readClaudeState(t, lockPath(fx.home, fx.profile)), lockBefore) {
					t.Error("repair changed preserved lock")
				}
			})
		}
	}
}

// A real rename failure is an I/O failure after validation, not a validation
// refusal. Prior direct writes/backups may remain; the failed destination and
// configuration must survive, and owned staging entries must be removed.
func TestClaudeRepairCommitFailureBoundary(t *testing.T) {
	for _, form := range []string{envregistry.FormReferenced, envregistry.FormMonolithic} {
		t.Run(form, func(t *testing.T) {
			fx, req, config := claudeRepairFixture(t, form)
			req.PreserveLock = true
			root := ManagedParent(fx.home, fx.profile, envregistry.ClaudeCode)
			context := filepath.Join(filepath.Dir(config), "CLAUDE.md")
			if err := os.Remove(context); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(context, 0700); err != nil {
				t.Fatal(err)
			}
			child := filepath.Join(context, "keep")
			if err := os.WriteFile(child, []byte("keep"), 0600); err != nil {
				t.Fatal(err)
			}
			before := claudeRefusalSnapshot(t, root)
			_, err := Resolve(req)
			if err == nil || !strings.Contains(err.Error(), DiagRepairFailed) {
				t.Fatalf("Resolve = %v, want commit failure", err)
			}
			after := claudeRefusalSnapshot(t, root)
			for _, path := range []string{"CLAUDE.md", filepath.Join("CLAUDE.md", "keep"), ".claude.json", ".agent-environment.json"} {
				old, ok := before[path]
				if !ok {
					t.Fatalf("fixture missing required managed entry %s", path)
				}
				assertClaudeRefusalSnapshot(t, map[string]claudeRefusalEntry{path: old}, map[string]claudeRefusalEntry{path: after[path]})
			}
			for path := range after {
				if strings.Contains(path, ".curator-write-") {
					t.Errorf("commit failure left temporary entry %s", path)
				}
			}
			// Neighboring successful control uses the same repaired request.
			if err := os.Remove(child); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(context); err != nil {
				t.Fatal(err)
			}
			if _, err := Resolve(req); err != nil {
				t.Fatalf("valid control: %v", err)
			}
		})
	}
}
