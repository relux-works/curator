package skillruntime

import (
	"os"
	"path/filepath"
	"testing"
)

// Every default home/adapter location is inside the package fixture. An outer
// HOME is not sufficient when the session supplies CODEX_HOME or XDG roots.
func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "curator-runtime-test-home-*")
	if err != nil {
		panic(err)
	}
	for key, value := range map[string]string{
		"HOME": home, "USERPROFILE": home, "CODEX_HOME": filepath.Join(home, ".codex"),
		"CLAUDE_CONFIG_DIR": filepath.Join(home, ".claude"), "XDG_CONFIG_HOME": filepath.Join(home, ".config"),
		"XDG_STATE_HOME": filepath.Join(home, ".local", "state"), "PI_CODING_AGENT_DIR": filepath.Join(home, ".pi"),
		"CURATOR_CONFIG": filepath.Join(home, "config.json"), "CURATOR_SYSTEM_CONFIG": "",
		"CURATOR_GLOBAL_USER_BIN": filepath.Join(home, "bin"),
	} {
		if err := os.Setenv(key, value); err != nil {
			panic(err)
		}
	}
	code := m.Run()
	_ = os.RemoveAll(home)
	os.Exit(code)
}
