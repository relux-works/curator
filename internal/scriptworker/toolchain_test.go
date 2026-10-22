package scriptworker

import (
	"path/filepath"
	"testing"
)

// Every dependent call skips when Go is absent, including repeated calls to a
// cached fixture builder. Goexit inside sync.Once would poison its cache and
// let the next call return an empty or previously built path without a skip.
// The normal worker/protocol tests exercise the builders with a real Go tool.
func TestFixtureBuildersSkipWithoutGo(t *testing.T) {
	t.Setenv("CURATOR_GO", filepath.Join(t.TempDir(), "absent-go"))
	for _, builder := range []struct {
		name  string
		build func(*testing.T) string
	}{
		{"interpreter", stubInterpreterBinary},
		{"forged-worker", forgeWorkerBinary},
		{"production", builtCuratorBinary},
	} {
		t.Run(builder.name, func(t *testing.T) {
			for _, name := range []string{"first", "second"} {
				skipped := false
				t.Run(name, func(t *testing.T) {
					defer func() { skipped = t.Skipped() }()
					path := builder.build(t)
					t.Errorf("missing toolchain returned a fixture path %q instead of skipping", path)
				})
				if !skipped {
					t.Error("dependent fixture call did not skip")
				}
			}
		})
	}
}
