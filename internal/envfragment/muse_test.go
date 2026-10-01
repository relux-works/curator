package envfragment

import (
	"github.com/relux-works/curator/internal/envregistry"
	"path/filepath"
	"strings"
	"testing"
)

func TestMuseFragmentBoundary(t *testing.T) {
	adapter, err := envregistry.ByID(envregistry.Muse)
	if err != nil {
		t.Fatal(err)
	}
	for _, state := range []string{"valid", "missing-data", "wrong-parent", "HOME", "channel"} {
		t.Run(state, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "environments")
			parent := filepath.Join(root, "acme", "muse")
			fragment := &Fragment{Environment: envregistry.Muse, Env: map[string]string{}}
			for _, variable := range adapter.HomeVariables {
				fragment.Env[variable[0]] = filepath.Join(parent, variable[1])
			}
			switch state {
			case "missing-data":
				delete(fragment.Env, "XDG_DATA_HOME")
			case "wrong-parent":
				fragment.Env["XDG_DATA_HOME"] = filepath.Join(parent, "config")
			case "HOME":
				fragment.Env["HOME"] = parent
			case "channel":
				fragment.SystemPrompt = &SystemPrompt{Path: filepath.Join(parent, "unverified.md")}
			}
			err := CheckBoundary(adapter, root, fragment)
			if (err == nil) != (state == "valid") {
				t.Fatalf("boundary=%v", err)
			}
			if state == "valid" {
				lines := strings.Split(strings.TrimSpace(string(fragment.EnvFormat(adapter))), "\n")
				for i, variable := range adapter.HomeVariables {
					if lines[i] != variable[0]+"="+filepath.Join(parent, variable[1]) {
						t.Fatalf("variable order: %v", lines)
					}
				}
			}
		})
	}
}
