package envprofile

import (
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

func TestPermissionsForProfileProperties(t *testing.T) {
	for _, locked := range []bool{false, true} {
		for _, mode := range []string{"", "native", "yolo"} {
			machine := envregistry.DefaultMachineConfig()
			machine.PermissionsLocked = locked
			if mode != "" {
				machine.Permissions["configured"] = mode
			}
			for _, profile := range []string{"configured", "absent"} {
				got := PermissionsForProfile(machine, profile)
				if got.Locked != (got.Source == "global") || got.Locked != locked {
					t.Fatalf("lock/source invariant: %+v", got)
				}
				wantMode, wantSource := "native", "default"
				if mode != "" && profile == "configured" {
					wantMode, wantSource = mode, "profile"
				}
				if locked {
					wantMode, wantSource = "native", "global"
				}
				if got.Mode != wantMode || got.Source != wantSource {
					t.Fatalf("locked=%t mode=%q profile=%q: %+v, want %s/%s", locked, mode, profile, got, wantMode, wantSource)
				}
			}
		}
	}
}
