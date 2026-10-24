package projectlease

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/relux-works/curator/internal/transaction"
)

// Actual Acquire subprocess exits at durable preparation or binding install.
// Built CLI GC must recover to durable unknown, retain the leased runtime and
// remove an orphan. A prior ledger is a neighboring case, not a prerequisite.
func TestInterruptedAcquireGCBecomesUnknown(t *testing.T) {
	binary := os.Getenv("CURATOR_TEST_LEASE_GC_BINARY")
	if binary == "" {
		binary = filepath.Join(t.TempDir(), "curator")
		if runtime.GOOS == "windows" {
			binary += ".exe"
		}
		goTool := os.Getenv("CURATOR_GO")
		if goTool == "" {
			goTool = "go"
		}
		build := exec.Command(goTool, "build", "-o", binary, "../../cmd/curator")
		if out, err := build.CombinedOutput(); err != nil {
			t.Fatalf("build CLI: %v\n%s", err, out)
		}
	}
	for _, prior := range []bool{false, true} {
		for _, point := range []transaction.Point{transaction.PointPrepared, transaction.PointAfterInstall} {
			t.Run(string(point)+map[bool]string{false: "/first", true: "/prior"}[prior], func(t *testing.T) {
				f := setup(t)
				if prior {
					acquire(t, f, "first")
				}
				_ = nextAdmission(t, &f)
				child := exec.Command(os.Args[0], "-test.run=^TestLeasePrepareExitHelper$")
				child.Env = append(os.Environ(), "CURATOR_TEST_PREPARE_EXIT_HOME="+f.home, "CURATOR_TEST_PREPARE_EXIT_POINT="+string(point))
				out, err := child.CombinedOutput()
				exit, ok := err.(*exec.ExitError)
				if !ok || exit.ExitCode() != 57 {
					t.Fatalf("admission did not reach crash boundary: %v\n%s", err, out)
				}
				config := filepath.Join(f.home, "config.json")
				raw, err := json.Marshal(map[string]any{"schema_version": 1, "skills_root": filepath.Join(f.home, "skills"), "projects": map[string]any{}})
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(config, raw, 0600); err != nil {
					t.Fatal(err)
				}
				gc := exec.Command(binary, "gc")
				gc.Env = append(os.Environ(), "CURATOR_CONFIG="+config)
				if out, err = gc.CombinedOutput(); err != nil {
					t.Fatalf("GC recovery: %v\n%s", err, out)
				}
				l, err := load(f.home)
				if err != nil {
					t.Fatal(err)
				}
				if e, ok := l.Leases["next"]; !ok || e.State != Unknown {
					t.Fatalf("CLI recovery left durable phase %s, want unknown", e.State)
				}
				statuses, err := f.m.Status(context.Background())
				if err != nil || len(statuses) != 1+map[bool]int{false: 0, true: 1}[prior] {
					t.Fatalf("recovered status: %v %v", statuses, err)
				}
				if _, err = os.Stat(filepath.Join(f.home, "runtime", "tool", "next")); err != nil {
					t.Fatal("GC removed recovered runtime", err)
				}
				if _, err = os.Stat(filepath.Join(f.home, "runtime", "tool", "orphan")); !os.IsNotExist(err) {
					t.Fatal("GC failed orphan control", err)
				}
			})
		}
	}
}
