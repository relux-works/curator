package envprofile

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/stateread"
)

func TestDotfileManagerTableMatchesSpecFixture(t *testing.T) {
	fixture, err := os.ReadFile("testdata/dotfile-manager-table-802caee.md")
	if err != nil {
		t.Fatal(err)
	}

	lines := []string{
		"| # | Manager | macOS | Linux | Windows | Source |",
		"|---|---|---|---|---|---|",
	}
	tick := string(rune(96))
	for index, row := range dotfileStateTable {
		lines = append(lines, fmt.Sprintf("| %d | %s%s%s | %s | %s | %s | %s |",
			index+1,
			tick, row.manager, tick,
			row.macOS.specText, row.linux.specText, row.windows.specText, row.source))
	}
	got := []byte(strings.Join(lines, "\n") + "\n")
	if !bytes.Equal(got, fixture) {
		limit := len(got)
		if len(fixture) < limit {
			limit = len(fixture)
		}
		for index := 0; index < limit; index++ {
			if got[index] != fixture[index] {
				t.Fatalf("dotfileStateTable differs from curator-spec 802caee fixture at byte %d: got %q, want %q", index, got[index:index+1], fixture[index:index+1])
			}
		}
		t.Fatalf("dotfileStateTable length %d differs from curator-spec 802caee fixture length %d", len(got), len(fixture))
	}
}

func TestDotfileStatePathResolution(t *testing.T) {
	platform, ok := dotfilePlatformForGOOS(runtime.GOOS)
	if !ok {
		t.Fatalf("unsupported test platform %q", runtime.GOOS)
	}
	home := t.TempDir()
	absoluteBase := filepath.Join(t.TempDir(), "xdg")
	for _, row := range dotfileStateTable {
		cell := row.cell(platform)
		if cell.baseEnv == "" {
			if _, resolved := resolveDotfileStatePath(cell, home, func(string) string { return "" }); resolved {
				t.Fatalf("%s unexpectedly resolved a none cell on %s", row.manager, platform)
			}
			continue
		}

		for _, envValue := range []string{"", "relative/xdg", absoluteBase} {
			t.Run(row.manager+"/"+valueName(envValue), func(t *testing.T) {
				path, resolved := resolveDotfileStatePath(cell, home, func(name string) string {
					if name != cell.baseEnv {
						t.Fatalf("looked up %q for %s", name, row.manager)
					}
					return envValue
				})
				if !resolved {
					t.Fatalf("%s on %s did not resolve", row.manager, platform)
				}
				base := filepath.Join(home, filepath.FromSlash(cell.defaultBase))
				if envValue != "" && filepath.IsAbs(envValue) {
					base = envValue
				}
				want := filepath.Join(base, cell.leaf)
				if path != want {
					t.Fatalf("resolved path %q, want %q", path, want)
				}
			})
		}
	}
	t.Logf("path resolution exercised on native platform %s with injected home and XDG values", platform)
}

func valueName(value string) string {
	switch {
	case value == "":
		return "empty"
	case filepath.IsAbs(value):
		return "absolute"
	default:
		return "relative"
	}
}

type dotfileTestFileInfo struct {
	name  string
	mode  os.FileMode
	isDir bool
}

func (info dotfileTestFileInfo) Name() string       { return info.name }
func (info dotfileTestFileInfo) Size() int64        { return 0 }
func (info dotfileTestFileInfo) Mode() os.FileMode  { return info.mode }
func (info dotfileTestFileInfo) ModTime() time.Time { return time.Time{} }
func (info dotfileTestFileInfo) IsDir() bool        { return info.isDir }
func (info dotfileTestFileInfo) Sys() any           { return nil }

// dotfileSeamLstat classifies an injected lstat result through the shared
// stateread seam, mirroring the production probe: absence stays quiet while
// inspection failures surface as errors.
func dotfileSeamLstat(readPath func(string) (os.FileInfo, error)) func(string) (stateread.Metadata, error) {
	return func(path string) (stateread.Metadata, error) {
		return stateread.LstatWith(path, readPath)
	}
}

func TestForeignManagerHintLstatDiscipline(t *testing.T) {
	platform, ok := dotfilePlatformForGOOS(runtime.GOOS)
	if !ok {
		t.Fatalf("unsupported test platform %q", runtime.GOOS)
	}
	home := t.TempDir()
	var paths []string
	var managers []string
	for _, row := range dotfileStateTable {
		path, resolved := resolveDotfileStatePath(row.cell(platform), home, func(string) string { return "" })
		if resolved {
			paths = append(paths, path)
			managers = append(managers, row.manager)
		}
	}
	if len(paths) == 0 {
		t.Fatalf("no state paths for %s", platform)
	}
	emptyEnv := func(string) string { return "" }

	t.Run("absence advances in table order", func(t *testing.T) {
		hint, err := foreignManagerHintAt(home, runtime.GOOS, emptyEnv, dotfileSeamLstat(func(path string) (os.FileInfo, error) {
			if path == paths[0] {
				return nil, os.ErrNotExist
			}
			if len(paths) > 1 && path == paths[1] {
				return dotfileTestFileInfo{name: path, mode: os.ModeDir, isDir: true}, nil
			}
			return nil, os.ErrNotExist
		}))
		if err != nil {
			t.Fatal(err)
		}
		if len(paths) > 1 && hint != managers[1] {
			t.Fatalf("hint %q, want next present manager %q", hint, managers[1])
		}
		if len(paths) == 1 && hint != "" {
			t.Fatalf("hint %q with no later present manager, want none", hint)
		}
	})

	t.Run("symlink to directory is not present", func(t *testing.T) {
		hint, err := foreignManagerHintAt(home, runtime.GOOS, emptyEnv, dotfileSeamLstat(func(path string) (os.FileInfo, error) {
			if path == paths[0] {
				// Lstat returns the link's mode even when its target is a directory.
				return dotfileTestFileInfo{name: path, mode: os.ModeSymlink, isDir: true}, nil
			}
			if len(paths) > 1 && path == paths[1] {
				return dotfileTestFileInfo{name: path, mode: os.ModeDir, isDir: true}, nil
			}
			return nil, os.ErrNotExist
		}))
		if err != nil {
			t.Fatal(err)
		}
		if len(paths) > 1 && hint != managers[1] {
			t.Fatalf("symlink counted as %q; want the next present manager %q", hint, managers[1])
		}
		if len(paths) == 1 && hint != "" {
			t.Fatalf("symlink counted as %q on %s", hint, platform)
		}
	})

	t.Run("failed inspection continues to later yadm directory", func(t *testing.T) {
		sentinel := os.ErrPermission
		var yadmPath string
		for _, row := range dotfileStateTable {
			if row.manager != "yadm" {
				continue
			}
			var resolved bool
			yadmPath, resolved = resolveDotfileStatePath(row.cell(platform), home, emptyEnv)
			if !resolved {
				yadmPath = ""
			}
		}
		if yadmPath != "" {
			if err := os.MkdirAll(yadmPath, 0o755); err != nil {
				t.Fatal(err)
			}
		}
		calls := make(map[string]int)
		hint, err := foreignManagerHintAt(home, runtime.GOOS, emptyEnv, dotfileSeamLstat(func(path string) (os.FileInfo, error) {
			calls[path]++
			if path == paths[0] {
				return nil, sentinel
			}
			return os.Lstat(path)
		}))
		if yadmPath == "" {
			if hint != "" {
				t.Fatalf("hint %q on %s, whose yadm cell is none", hint, platform)
			}
		} else if hint != "yadm" {
			t.Fatalf("later present yadm row returned hint %q, want yadm", hint)
		}
		if err == nil || !errors.Is(err, sentinel) {
			t.Fatalf("inspection error %v, want wrapped permission error", err)
		}
		if calls[paths[0]] != 1 {
			t.Fatalf("first candidate lstat calls %d, want 1", calls[paths[0]])
		}
		if yadmPath != "" && calls[yadmPath] != 1 {
			t.Fatalf("yadm row lstat calls %d, want 1 after earlier permission error", calls[yadmPath])
		}
	})

	t.Run("unreadable only candidate yields unknown without a hint", func(t *testing.T) {
		sentinel := os.ErrPermission
		calls := make(map[string]int)
		hint, err := foreignManagerHintAt(home, runtime.GOOS, emptyEnv, dotfileSeamLstat(func(path string) (os.FileInfo, error) {
			calls[path]++
			if path == paths[0] {
				return nil, sentinel
			}
			return nil, os.ErrNotExist
		}))
		if hint != "" {
			t.Fatalf("unreadable candidate produced manager hint %q", hint)
		}
		if err == nil || !errors.Is(err, sentinel) {
			t.Fatalf("inspection error %v, want wrapped permission error", err)
		}
		for _, path := range paths {
			if calls[path] != 1 {
				t.Fatalf("candidate %q lstat calls %d, want 1 (continue after unreadable row)", path, calls[path])
			}
		}
	})

	t.Run("regular file is not present", func(t *testing.T) {
		hint, err := foreignManagerHintAt(home, runtime.GOOS, emptyEnv, dotfileSeamLstat(func(path string) (os.FileInfo, error) {
			if path == paths[0] {
				return dotfileTestFileInfo{name: path, mode: 0, isDir: false}, nil
			}
			if len(paths) > 1 && path == paths[1] {
				return dotfileTestFileInfo{name: path, mode: os.ModeDir, isDir: true}, nil
			}
			return nil, os.ErrNotExist
		}))
		if err != nil {
			t.Fatal(err)
		}
		if len(paths) > 1 && hint != managers[1] {
			t.Fatalf("regular file counted as %q; want %q", hint, managers[1])
		}
		if len(paths) == 1 && hint != "" {
			t.Fatalf("regular file counted as %q on %s", hint, platform)
		}
	})
}

func TestDotfileManagerStateAloneDoesNotWarn(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	installIdleProfile(t, home, "dotfile-state-without-unmanaged-surface")
	claudeHome(t)

	operatorHome := t.TempDir()
	pinOperatorHome(t, operatorHome)
	dataHome := filepath.Join(operatorHome, "xdg-data")
	t.Setenv("XDG_DATA_HOME", dataHome)
	if err := os.MkdirAll(filepath.Join(dataHome, "chezmoi"), 0o755); err != nil {
		t.Fatal(err)
	}

	results, err := UseWithPolicy(home, "dotfile-state-without-unmanaged-surface", "", "", false, Policy{Takeover: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range results {
		for _, warning := range result.Warnings {
			if strings.HasPrefix(warning, DiagForeignSuspect+":") {
				t.Fatalf("manager state without unmanaged surfaces produced warning: %+v", result)
			}
		}
	}
}
