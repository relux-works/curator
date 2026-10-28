package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var detectorToolKeys = []string{"PATH", "GOCACHE", "GOPATH", "GOMODCACHE", "GOPROXY", "GOSUMDB", "GOTOOLCHAIN", "CGO_ENABLED", "LANG", "LC_ALL", "CURATOR_FIXTURE_PID_LOG", "SYSTEMROOT", "WINDIR"}

func isolateDetectorParent(t *testing.T) {
	t.Helper()
	saved := os.Environ()
	tools := map[string]string{}
	for _, key := range detectorToolKeys {
		if value, present := os.LookupEnv(key); present {
			tools[key] = value
		}
	}
	root := t.TempDir()
	paths := map[string]string{"HOME": "home", "USERPROFILE": "home", "TMPDIR": "tmp", "TMP": "tmp", "TEMP": "tmp", "GOTMPDIR": "tmp", "XDG_CACHE_HOME": "xdg-cache", "XDG_CONFIG_HOME": "xdg-config"}
	for _, rel := range paths {
		if err := os.MkdirAll(filepath.Join(root, rel), 0700); err != nil {
			t.Fatal(err)
		}
	}
	// Clear the actual test parent before constructing a child allowlist. No
	// GIT_*, selector, socket, askpass or ambient GOFLAGS reaches either child.
	os.Clearenv()
	t.Cleanup(func() {
		os.Clearenv()
		for _, entry := range saved {
			key, value, _ := strings.Cut(entry, "=")
			if err := os.Setenv(key, value); err != nil {
				t.Error(err)
			}
		}
	})
	for key, value := range tools {
		if err := os.Setenv(key, value); err != nil {
			t.Fatal(err)
		}
	}
	for key, rel := range paths {
		if err := os.Setenv(key, filepath.Join(root, rel)); err != nil {
			t.Fatal(err)
		}
	}
	for key, value := range map[string]string{"GOFLAGS": "-trimpath -buildvcs=false", "GOENV": "off", "GOTELEMETRY": "off", "GOPROXY": "off", "GOSUMDB": "off"} {
		// An explicit tool transport belongs to the caller. Local fixture runs
		// supply off; hosted drivers bind their effective module/cache context
		// before synthetic HOME changes. Never silently enable a network fallback.
		if (key == "GOPROXY" || key == "GOSUMDB") && tools[key] != "" {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			t.Fatal(err)
		}
	}
}
func detectorChildEnv() []string { return append([]string{}, os.Environ()...) }
