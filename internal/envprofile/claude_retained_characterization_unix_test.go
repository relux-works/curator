//go:build unix

package envprofile

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
)

func TestClaudeReadOnlySettingsBytesAndWriteBitRefusal(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	req := fx.request(envregistry.ClaudeCode)
	enableClaudeReadOnly(&req)
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}

	home := ManagedHomeDir(fx.home, fx.profile, req.EnvID)
	settings := filepath.Join(home, "settings.json")
	keepPolicy := claudeSettingsFixtureBytes(t, home)
	want := keepPolicy("settings.json", []byte(`{"operator":"preserve","model":"opaque"}`))
	if err := os.Chmod(settings, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settings, want, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(settings, 0o400); err != nil {
		t.Fatal(err)
	}
	req.Repair = false
	if _, err := Resolve(req); err != nil {
		t.Fatalf("read-only operator settings were rejected: %v", err)
	}
	got, err := os.ReadFile(settings)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("read-only settings bytes changed: error=%v bytes=%q want=%q", err, got, want)
	}
	info, err := os.Stat(settings)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o400 {
		t.Fatalf("read-only mode = %v, want 0400", info.Mode().Perm())
	}

	if err := os.Chmod(settings, 0o600); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(settings)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(req); err == nil || !strings.Contains(err.Error(), "write bits") {
		t.Fatalf("writable read-only settings were not refused: %v", err)
	}
	after, err := os.ReadFile(settings)
	if err != nil || !bytes.Equal(after, before) {
		t.Fatalf("refusal changed settings bytes: error=%v bytes=%q want=%q", err, after, before)
	}
	info, err = os.Stat(settings)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("refusal changed writable mode: mode=%v", info.Mode().Perm())
	}
}

func TestClaudeSwitchCopiesRootAndReferencedFormKeepsOnlyRootIncludes(t *testing.T) {
	t.Run("switch copies regular root file", func(t *testing.T) {
		fx, homes := nativeSkillsFixture(t)
		req := fx.request(envregistry.ClaudeCode)
		req.Repair = true
		if _, err := Resolve(req); err != nil {
			t.Fatal(err)
		}
		managedRoot := filepath.Join(ManagedHomeDir(fx.home, fx.profile, req.EnvID), "CLAUDE.md")
		want, err := os.ReadFile(managedRoot)
		if err != nil {
			t.Fatal(err)
		}

		results, err := UseWithPolicy(fx.home, fx.profile, envregistry.ClaudeCode, "", false, Policy{})
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 1 || !results[0].OK {
			t.Fatalf("Claude Code switch results = %#v", results)
		}
		root := filepath.Join(homes[envregistry.ClaudeCode], "CLAUDE.md")
		info, err := os.Lstat(root)
		if err != nil {
			t.Fatal(err)
		}
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			t.Fatalf("native CLAUDE.md is not a regular copied file: info=%v", info)
		}
		got, err := os.ReadFile(root)
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("native root bytes differ from rendered managed root: error=%v bytes=%q want=%q", err, got, want)
		}
		if !bytes.Contains(got, []byte("hello\n")) || bytes.Contains(got, []byte("Be terse.\n")) {
			t.Fatalf("native root has wrong root/system module content: %q", got)
		}
	})

	t.Run("referenced root imports exact root module bytes only", func(t *testing.T) {
		fx := writeManagedFixture(t, "acme")
		seedLiveNativeCredentials(t, fx)
		req := fx.request(envregistry.ClaudeCode)
		req.Machine.Forms = map[string]string{envregistry.ClaudeCode: envregistry.FormReferenced}
		req.Repair = true
		if _, err := Resolve(req); err != nil {
			t.Fatal(err)
		}

		home := ManagedHomeDir(fx.home, fx.profile, req.EnvID)
		root := filepath.Join(home, "CLAUDE.md")
		info, err := os.Lstat(root)
		if err != nil {
			t.Fatal(err)
		}
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			t.Fatalf("managed CLAUDE.md is not a regular copied file: info=%v", info)
		}
		data, err := os.ReadFile(root)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(data, []byte("@.agent-context/modules/acme/a.md\n")) || bytes.Contains(data, []byte("Be terse.\n")) {
			t.Fatalf("referenced root has unexpected include/system content: %q", data)
		}
		module := filepath.Join(home, ".agent-context", "modules", "acme", "a.md")
		if got, err := os.ReadFile(module); err != nil || !bytes.Equal(got, []byte("hello\n")) {
			t.Fatalf("referenced module bytes = %q, error %v", got, err)
		}
		system := filepath.Join(home, ".agent-context", "modules", "acme", "s.md")
		if _, err := os.Lstat(system); !os.IsNotExist(err) {
			t.Fatalf("system module was materialized as an include: %v", err)
		}
		marker := readManagedMarker(t, fx, req.EnvID)
		surface, ok := marker.Surfaces["root-context"]
		if !ok || surface.Copies == nil {
			t.Fatalf("root-context marker has no copy records: %#v", surface)
		}
		var copied bool
		for _, copy := range *surface.Copies {
			if copy.Path == "CLAUDE.md" && copy.Reason == envmarker.ReasonClaudeCodeRootContext {
				copied = true
			}
		}
		if !copied {
			t.Fatal("root-context marker does not record the copied CLAUDE.md")
		}
	})
}

func TestClaudeVersionTokenRemainsInEnvprofile(t *testing.T) {
	if got := versionToken.FindString("Claude Code 2.1.261 (Claude Code)\n"); got != "2.1.261" {
		t.Fatalf("version token = %q, want %q", got, "2.1.261")
	}
	if got := versionToken.FindString("Claude Code unknown\n"); got != "" {
		t.Fatalf("unknown release produced version token %q", got)
	}
}
