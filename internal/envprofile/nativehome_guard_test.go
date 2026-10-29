package envprofile

import (
	"github.com/relux-works/curator/internal/testhomeguard"
	"testing"
)

// Both child controls exercise this package's actual TestMain exit boundary.
func TestNativeHomeGuard(t *testing.T) { testhomeguard.Control(t, "TestNativeHomeGuardChild") }

// This named child resolves native defaults only within synthetic fixture homes.
func TestNativeHomeGuardChild(t *testing.T) {
	testhomeguard.Child(t, func() (string, error) { adapter, _ := adapterByID("claude_code"); return NativeHome(adapter) })
}
