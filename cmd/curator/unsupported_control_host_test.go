//go:build curator_test_platform

package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/godriver"
)

// The real install entry reaches production native-control preflight, refuses
// before compilation/publication, and leaves status --check non-current.
func TestCompiledInstallRefusesUncoveredHost(t *testing.T) {
	t.Setenv("CURATOR_TEST_NATIVE_GOOS", "freebsd")
	_, home := compiledProject(t)
	config := filepath.Join(home, "config.json")
	code, _, stderr := capture(t, config, "install", "app")
	if code == exitOK || !strings.Contains(stderr, godriver.CodeControlUnavailable) {
		t.Fatalf("unsupported host was not refused by production install: code=%d stderr=%s", code, stderr)
	}
	if entries := publishedCacheEntries(t, home); len(entries) != 0 {
		t.Fatalf("refusal published %d entries", len(entries))
	}
	if code, _, _ := capture(t, config, "status", "app", "--check"); code == exitOK {
		t.Fatal("refused compiled command reported current")
	}
}
