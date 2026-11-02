package envprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
)

// A selected healthy Codex home cannot let migration plan reach backend
// observation while a different adapter's authority is missing/invalid or an
// unsealed legacy journal stands. Healthy control reaches the real inventory.
func TestProvisioningOriginMigrationCommonBarrier(t *testing.T) {
	for _, mode := range []string{"healthy", "legacy-neighbor", "orphan-neighbor", "unsealed-journal"} {
		t.Run(mode, func(t *testing.T) {
			fx, _, _ := recordedOriginFixture(t, "a")
			agent := filepath.Join(fx.native["pi"], "agent")
			if err := os.MkdirAll(agent, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(agent, "auth.json"), nil, 0600); err != nil {
				t.Fatal(err)
			}
			pi := fx.request("pi")
			pi.Repair = true
			if _, err := Resolve(pi); err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "legacy-neighbor":
				if err := os.Remove(filepath.Join(ManagedHomeDir(fx.home, fx.profile, "pi"), envorigin.Name)); err != nil {
					t.Fatal(err)
				}
			case "orphan-neighbor":
				if err := os.Remove(filepath.Join(ManagedHomeDir(fx.home, fx.profile, "pi"), envmarker.Name)); err != nil {
					t.Fatal(err)
				}
			case "unsealed-journal":
				if err := writeMigrationJournal(fx.home, &migrationJournal{Version: 1, Plan: strings.Repeat("a", 64), Ops: []migrationJournalOp{}, Markers: []migrationJournalMarker{}}); err != nil {
					t.Fatal(err)
				}
			}
			observations := 0
			req := fx.migrateRequest()
			req.Profile, req.EnvID = fx.profile, "codex_cli"
			req.Detect = func(envregistry.Adapter) string { observations++; return "unknown" }
			req.NativeHomeOf = func(string) (string, error) {
				t.Error("migration used current native selector")
				return "", fmt.Errorf("forbidden native lookup")
			}
			before := nativeSnapshot(t, fx.home)
			report, err := PlanMigration(req)
			if mode == "healthy" {
				if err != nil || report == nil || len(report.Homes) != 1 || observations == 0 {
					t.Fatalf("healthy real inventory not reached: observed=%d report-present=%t err=%v", observations, report != nil, err)
				}
				return
			}
			code := envorigin.DiagInvalid
			if mode == "legacy-neighbor" {
				code = envorigin.DiagRequired
			}
			if err == nil || !strings.Contains(err.Error(), code) || observations != 0 {
				t.Fatalf("migration common barrier before backend: observed=%d report-present=%t err=%v", observations, report != nil, err)
			}
			if nativeSnapshot(t, fx.home) != before {
				t.Fatal("read-only migration refusal changed state")
			}
		})
	}
}
