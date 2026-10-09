package main

// N8 regression: `audit --allow` addresses a content identity. It pins a
// supported digest and refuses everything else before any filesystem
// access, so pin state can never escape the audit namespace. This test
// drives the production run() entry point.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
)

// TestAuditAllowPinsOnlySupportedContentIdentity pins supported digests
// through the CLI and refuses path-shaped and malformed values without
// writing any pin state. A mutant that restores the direct --allow join
// pins the traversal input with exit 0 and fails the refusal rows.
func TestAuditAllowPinsOnlySupportedContentIdentity(t *testing.T) {
	t.Setenv("USER", "fixture")
	bare := strings.Repeat("a", 64)
	upper := strings.Repeat("B", 64)
	for _, valid := range []struct{ name, allow, dir string }{
		{"bare-hex", bare, bare},
		{"upper-hex", upper, strings.ToLower(upper)},
		{"sha256-prefixed", "sha256:" + bare, bare},
	} {
		t.Run("allow/"+valid.name, func(t *testing.T) {
			source, home := profileHome(t)
			if code, _, stderr := runProfile(t, source, "audit", "--allow", valid.allow, "--reason", "synthetic approval"); code != exitOK {
				t.Fatalf("audit --allow %q = %d, want %d\nstderr:\n%s", valid.allow, code, exitOK, stderr)
			}
			if _, err := os.Stat(filepath.Join(home, "audit", valid.dir, "trust.json")); err != nil {
				t.Fatalf("pin record for %q: %v", valid.allow, err)
			}
		})
	}
	for _, refused := range []struct{ name, allow, escape string }{
		{"parent-traversal", "../outside-audit", "outside-audit"},
		{"nested-path", "sub/dir", "sub"},
		{"short-digest", strings.Repeat("a", 63), ""},
		{"non-hex-digest", strings.Repeat("z", 64), ""},
	} {
		t.Run("refuse/"+refused.name, func(t *testing.T) {
			source, home := profileHome(t)
			code, _, _ := runProfile(t, source, "audit", "--allow", refused.allow, "--reason", "synthetic approval")
			if code == exitOK {
				t.Fatalf("audit --allow %q = %d, want a refusal", refused.allow, code)
			}
			if refused.escape != "" {
				if _, err := os.Stat(filepath.Join(home, refused.escape, "trust.json")); err == nil {
					t.Fatalf("non-digest hash escaped the audit namespace")
				}
			}
			assertNoPinState(t, home)
		})
	}
}

// loadCountingSource wraps a stubConfigSource and records configuration
// Load calls, proving refusal happens before any configuration access.
type loadCountingSource struct {
	stubConfigSource
	loads *int
}

func (s loadCountingSource) Load(warn func(string)) (*config.Config, error) {
	*s.loads++
	return s.stubConfigSource.Load(warn)
}

func runWithSource(t *testing.T, source configSource, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr strings.Builder
	code := run(args, source, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// TestAuditAllowRefusalPrecedesConfigLoad proves every refused --allow
// value exits usage before any configuration access: the config source
// records zero Load calls. The admitted-digest control row pins with a
// nonzero Load count, proving the counter observes real loads. An ordering
// mutant that loads configuration ahead of the CLI parse fails the no-Load
// assertion. This test drives the production run() entry point.
func TestAuditAllowRefusalPrecedesConfigLoad(t *testing.T) {
	t.Setenv("USER", "fixture")
	for _, refused := range []struct {
		name string
		argv []string
	}{
		{"empty-split", []string{"audit", "--allow", "", "--reason", "synthetic approval"}},
		{"empty-equals", []string{"audit", "--allow=", "--reason", "synthetic approval"}},
		{"parent-traversal", []string{"audit", "--allow", "../outside-audit", "--reason", "synthetic approval"}},
		{"nested-path", []string{"audit", "--allow", "sub/dir", "--reason", "synthetic approval"}},
		{"short-digest", []string{"audit", "--allow", strings.Repeat("a", 63), "--reason", "synthetic approval"}},
		{"non-hex-digest", []string{"audit", "--allow", strings.Repeat("z", 64), "--reason", "synthetic approval"}},
	} {
		t.Run("refuse/"+refused.name, func(t *testing.T) {
			base, home := profileHome(t)
			var loads int
			source := loadCountingSource{stubConfigSource: base, loads: &loads}
			code, _, stderr := runWithSource(t, source, refused.argv...)
			if code != exitUsage {
				t.Fatalf("audit %s = %d, want usage refusal (%d)\nstderr:\n%s", refused.name, code, exitUsage, stderr)
			}
			if loads != 0 {
				t.Fatalf("refused --allow reached configuration (%d Load calls), want zero: refusal must precede config access", loads)
			}
			assertNoPinState(t, home)
		})
	}
	t.Run("allow/admitted-loads-config", func(t *testing.T) {
		base, home := profileHome(t)
		var loads int
		source := loadCountingSource{stubConfigSource: base, loads: &loads}
		digest := strings.Repeat("a", 64)
		code, _, stderr := runWithSource(t, source, "audit", "--allow", digest, "--reason", "synthetic approval")
		if code != exitOK {
			t.Fatalf("audit --allow %q = %d, want %d\nstderr:\n%s", digest, code, exitOK, stderr)
		}
		if loads == 0 {
			t.Fatalf("admitted pin bypassed configuration; the Load counter cannot distinguish refusal from a dead probe")
		}
		if _, err := os.Stat(filepath.Join(home, "audit", digest, "trust.json")); err != nil {
			t.Fatalf("pin record for %q: %v", digest, err)
		}
	})
}

// assertNoPinState fails when any pin record exists under home: a refused
// --allow value must not write trust state anywhere. A failed read is
// reported, never treated as an absence of records.
func assertNoPinState(t *testing.T, home string) {
	t.Helper()
	err := filepath.WalkDir(home, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && entry.Name() == "trust.json" {
			t.Fatalf("refused --allow wrote pin state at %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
