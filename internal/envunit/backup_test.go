package envunit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// This metadata grammar test covers fixed managed seed/config slots and their
// native-custody exclusions. It is not an account or publication-owner witness.
func TestProvisioningOriginBackupSlots(t *testing.T) {
	for _, tc := range []struct {
		adapter, path  string
		managed, valid bool
	}{
		{"codex_cli", "config.toml", true, true}, {"codex_cli", "curator-mcp.config.toml", true, true},
		{"codex_cli", "config.toml", false, false}, {"codex_cli", "auth.json", true, false},
		{"codex_cli", "sessions/transcript", true, false}, {"claude_code", ".claude.json", true, true},
		{"claude_code", ".credentials.json", true, false}, {"pi", "settings.json", true, true},
		{"pi", "models.json", false, false}, {"opencode", "opencode.json", true, true},
		{"muse", "config/muse/settings.json", true, true}, {"muse", "config/muse/auth.json", true, false},
		{"muse", "config/muse/AGENTS.md", true, true}, {"muse", "data/muse/skills/fixture-skill", true, true},
		{"claude_code", ".agent-context/modules/fixture/member.md", true, true},
	} {
		t.Run(tc.adapter+"/"+tc.path+"/"+map[bool]string{true: "managed", false: "native"}[tc.managed], func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, filepath.FromSlash(tc.path))
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink("unread-absent-synthetic-target", path); err != nil {
				t.Fatal(err)
			}
			err := CheckBackupGeneration(root, tc.adapter, tc.managed)
			if tc.valid && err != nil {
				t.Fatal(err)
			}
			if !tc.valid && (err == nil || !strings.Contains(err.Error(), "backup contains an unowned slot")) {
				t.Fatalf("unowned backup slot admitted: %v", err)
			}
		})
	}
}
