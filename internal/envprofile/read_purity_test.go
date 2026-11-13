package envprofile

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/transaction"
)

type readWitness struct {
	info    fs.FileInfo
	content string
	link    string
}

// readWitnesses excludes kernel atime; it records names, types, identity,
// permissions, mtime, file bytes and symlink targets without following links.
func readWitnesses(t *testing.T, root string) map[string]readWitness {
	t.Helper()
	out := map[string]readWitness{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		item := readWitness{info: info}
		if info.Mode().IsRegular() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			item.content = string(data)
		} else if info.Mode()&os.ModeSymlink != 0 {
			item.link, err = os.Readlink(path)
			if err != nil {
				return err
			}
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		out[rel] = item
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func checkReadWitnesses(before, after map[string]readWitness) error {
	if len(before) != len(after) {
		return fmt.Errorf("read-only witnesses changed: entries %d -> %d", len(before), len(after))
	}
	for path, old := range before {
		now, ok := after[path]
		if !ok || !os.SameFile(old.info, now.info) || old.info.Mode() != now.info.Mode() || !old.info.ModTime().Equal(now.info.ModTime()) || old.content != now.content || old.link != now.link {
			return fmt.Errorf("read-only witnesses changed: %s", path)
		}
	}
	return nil
}

func assertReadWitnesses(t *testing.T, root string, before map[string]readWitness) {
	t.Helper()
	if err := checkReadWitnesses(before, readWitnesses(t, root)); err != nil {
		t.Fatal(err)
	}
}

// Empty reads report absence without making manager state or temporary source
// snapshots. Initialized reads list the existing default and remain unchanged.
func TestReadPurityManagerReads(t *testing.T) {
	for _, initialized := range []bool{false, true} {
		t.Run(fmt.Sprintf("initialized=%v", initialized), func(t *testing.T) {
			base := t.TempDir()
			home := filepath.Join(base, "manager")
			scratch := filepath.Join(base, "scratch")
			if err := os.Mkdir(scratch, 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("TMPDIR", scratch)
			pinHomes(t)
			if initialized {
				if err := EnsureDefaultWithPolicy(home, Policy{}); err != nil {
					t.Fatal(err)
				}
			}
			originalEngine := newOperationEngine
			newOperationEngine = func(string, ...transaction.Option) (*transaction.Engine, error) {
				t.Fatal("read attempted journal/recovery engine")
				return nil, nil
			}
			t.Cleanup(func() { newOperationEngine = originalEngine })
			before := readWitnesses(t, base)
			infos, err := ListWithPolicy(home, Policy{})
			if err != nil {
				t.Fatal(err)
			}
			if initialized {
				if len(infos) != 1 || infos[0].Name != DefaultProfile {
					t.Fatalf("initialized listing = %+v", infos)
				}
			} else if len(infos) != 0 {
				t.Fatalf("empty listing = %+v", infos)
			}
			status, err := StatusOf(StatusRequest{Home: home, Machine: envregistry.DefaultMachineConfig(), Detect: func(envregistry.Adapter) string { return "unknown" }})
			if err != nil || status == nil {
				t.Fatalf("status = %v, %v", status, err)
			}
			for _, preview := range []bool{false, true} {
				_, err := Resolve(ResolveRequest{Home: home, Profile: DefaultProfile, EnvID: envregistry.ClaudeCode, Machine: envregistry.DefaultMachineConfig(), Repair: preview, DryRun: preview, Detect: func(envregistry.Adapter) string { return "unknown" }})
				if err == nil {
					t.Fatal("unprovisioned resolve must report missing/stale state")
				}
			}
			assertReadWitnesses(t, base, before)
		})
	}
}

// Builtin removal refuses before lock, recovery or store creation on both an
// absent manager and a valid initialized one, with purge enabled or disabled.
func TestReadPurityBuiltinRemove(t *testing.T) {
	for _, initialized := range []bool{false, true} {
		for _, purge := range []bool{false, true} {
			t.Run(fmt.Sprintf("initialized=%v/purge=%v", initialized, purge), func(t *testing.T) {
				base := t.TempDir()
				home := filepath.Join(base, "manager")
				if initialized {
					if err := EnsureDefaultWithPolicy(home, Policy{}); err != nil {
						t.Fatal(err)
					}
				}
				before := readWitnesses(t, base)
				err := Remove(home, DefaultProfile, purge)
				if err == nil || !strings.Contains(err.Error(), "builtin_profile_remove_forbidden") {
					t.Fatalf("builtin removal = %v", err)
				}
				assertReadWitnesses(t, base, before)
			})
		}
	}
}

// The checker accepts identical state and detects a planted mkdir in the
// immediate changed witness. The syscall observer covers transient attempts.
func TestReadPurityWitnessControl(t *testing.T) {
	base := t.TempDir()
	before := readWitnesses(t, base)
	if err := checkReadWitnesses(before, readWitnesses(t, base)); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(base, "forbidden-lock"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := checkReadWitnesses(before, readWitnesses(t, base)); err == nil || !strings.Contains(err.Error(), "read-only witnesses changed") {
		t.Fatalf("mkdir plant = %v", err)
	}
}

// Read-only managers refuse acquisition/extraction/publication seams before
// creating a cache or snapshot; existing local inspection remains usable.
func TestReadPurityGitManager(t *testing.T) {
	base := t.TempDir()
	scratch := filepath.Join(base, "scratch")
	if err := os.Mkdir(scratch, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TMPDIR", scratch)
	home := filepath.Join(base, "manager")
	manager := newReadOnlyGitManager(home)
	resolved := contextresolve.Resolved{Kind: contextlock.KindContext, Name: "sample", Source: "example.com/sample", Commit: strings.Repeat("1", 40)}
	operations := map[string]func() error{
		"ensure-repo": func() error { _, err := manager.ensureRepo(resolved.Source); return err },
		"fetch":       func() error { return manager.fetch(resolved.Source) },
		"manifest": func() error {
			_, err := manager.Manifest(resolved.Kind, resolved.Name, resolved.Source, "", resolved.Commit)
			return err
		},
		"context-manifest": func() error { _, err := manager.contextManifestAt("missing", resolved.Commit, ""); return err },
		"ensure-entry":     func() error { _, err := manager.ensureEntry(home, resolved); return err },
		"inspect-entry":    func() error { _, cleanup, err := manager.inspectEntry(home, resolved); cleanup(); return err },
	}
	for name, run := range operations {
		t.Run(name, func(t *testing.T) {
			before := readWitnesses(t, base)
			err := run()
			if err == nil || !strings.Contains(err.Error(), "source_resolution_required") {
				t.Fatalf("read-only acquisition = %v", err)
			}
			assertReadWitnesses(t, base, before)
		})
	}
	entry := manager.entryPath(home, resolved)
	if err := os.MkdirAll(entry, 0700); err != nil {
		t.Fatal(err)
	}
	before := readWitnesses(t, base)
	got, cleanup, err := manager.inspectEntry(home, resolved)
	cleanup()
	if err != nil || got != entry {
		t.Fatalf("local inspection = %q, %v", got, err)
	}
	got, err = manager.ensureEntry(home, resolved)
	if err != nil || got != entry {
		t.Fatalf("local entry = %q, %v", got, err)
	}
	assertReadWitnesses(t, base, before)
}

// Missing MCP material is reported locally even when its Git source is cached;
// status must not extract a temporary fallback snapshot to improve its output.
func TestReadPurityStatusMissingStore(t *testing.T) {
	fx := writeManagedFixture(t, "sample")
	lock, _, err := readLock(fx.home, "sample")
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find(contextlock.KindMCP, "figma-devmode")
	if !ok {
		t.Fatal("missing fixture member")
	}
	if err := os.RemoveAll(memberEntryPath(fx.home, member)); err != nil {
		t.Fatal(err)
	}
	scratch := t.TempDir()
	t.Setenv("TMPDIR", scratch)
	before := readWitnesses(t, fx.home)
	tempBefore := readWitnesses(t, scratch)
	status, err := StatusOf(statusRequest(fx))
	if err != nil {
		t.Fatal(err)
	}
	assertReadWitnesses(t, fx.home, before)
	assertReadWitnesses(t, scratch, tempBefore)
	if len(status.MCPDeclarations) != 0 || !strings.Contains(strings.Join(status.Warnings, "\n"), "source_resolution_required") {
		t.Fatalf("missing local declaration = %+v, warnings %v", status.MCPDeclarations, status.Warnings)
	}
}

// Explicit initialization creates new private record parents without silently
// changing a pre-existing broader parent. That retained mode is a finding.
func TestReadPurityExistingRecordParentNotRepaired(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix mode witness; Windows DACL behavior requires native qualification")
	}
	base := t.TempDir()
	home := filepath.Join(base, "manager")
	if err := os.MkdirAll(ProfilesDir(home), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(ProfilesDir(home), 0755); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(ProfilesDir(home))
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureDefaultWithPolicy(home, Policy{}); err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(ProfilesDir(home))
	if err != nil {
		t.Fatal(err)
	}
	if before.Mode() != after.Mode() {
		t.Fatal("explicit initialization silently changed an existing record parent mode")
	}
	readBefore := readWitnesses(t, base)
	if _, err := ListWithPolicy(home, Policy{}); err != nil {
		t.Fatal(err)
	}
	assertReadWitnesses(t, base, readBefore)
}

// Provisioned status, nonrepair resolve and a repair preview do not run the
// implicit vendor probe. Supplied observations and explicit mutation probing
// remain usable; the tripwire is this test binary in an isolated fixture.
func TestReadPurityVendorObservation(t *testing.T) {
	fx := writeManagedFixture(t, "sample")
	seedLiveNativeCredentials(t, fx)
	provision(t, fx, envregistry.ClaudeCode, envregistry.DefaultMachineConfig())
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	name := "mcp-qualification-tripwire"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	probe := filepath.Join(t.TempDir(), name)
	if err := os.Link(executable, probe); err != nil {
		raw, err := os.ReadFile(executable)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(probe, raw, 0700); err != nil {
			t.Fatal(err)
		}
	}
	marker := filepath.Join(t.TempDir(), "vendor-ran")
	t.Setenv("CURATOR_TEST_MCP_QUALIFICATION_TRIPWIRE", marker)
	index := -1
	for i := range envregistry.Registry {
		if envregistry.Registry[i].ID == envregistry.ClaudeCode {
			index = i
			break
		}
	}
	if index < 0 {
		t.Fatal("missing fixture adapter")
	}
	original := envregistry.Registry[index]
	envregistry.Registry[index].Probe = []string{probe}
	t.Cleanup(func() { envregistry.Registry[index] = original })
	req := statusRequest(fx)
	req.Detect = nil
	if _, err := StatusOf(req); err != nil {
		t.Fatal(err)
	}
	for _, preview := range []bool{false, true} {
		read := fx.request(envregistry.ClaudeCode)
		read.Detect = nil
		read.Repair = preview
		read.DryRun = preview
		_, _ = Resolve(read)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("read started vendor probe: %v", err)
	}
	observed := ResolveRequest{readOnly: true, Detect: func(envregistry.Adapter) string { return "observed" }}
	if got := observed.detected(envregistry.Registry[index]); got != "observed" {
		t.Fatalf("existing observation = %q", got)
	}
	explicit := ResolveRequest{}
	_ = explicit.detected(envregistry.Registry[index])
	data, err := os.ReadFile(marker)
	if err != nil || string(data) != "called" {
		t.Fatalf("explicit probe positive did not execute: %v", err)
	}
}
