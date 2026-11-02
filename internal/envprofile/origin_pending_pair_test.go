package envprofile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/transaction"
)

// Both journals come from actual native publishers. Temporarily hiding J1 in
// this owned fixture permits planting an independently valid overlapping J2;
// restoring J1 must refuse that pair without effects. A distinct profile/root
// pair must instead recover both publishers and reach both context aliases.
// This tests pending inventory admission, not hostile same-UID tamper resistance.
func TestProvisioningOriginTwoPendingUnits(t *testing.T) {
	for _, overlap := range []bool{false, true} {
		t.Run(map[bool]string{false: "disjoint", true: "overlap"}[overlap], func(t *testing.T) {
			fx, req, _ := recordedOriginFixture(t, "a")
			root := t.TempDir()
			writePackage(t, root, "pending-q-context", "1.0.0", "Q context\n")
			if _, _, _, err := Install(fx.home, InstallOptions{Operand: root, As: "pending-q"}); err != nil {
				t.Fatal(err)
			}
			a := fx.native["codex_cli"]
			b := t.TempDir()
			stop := func(profile, native string) string {
				t.Helper()
				id := ""
				op, err := beginOperation(fx.home, transaction.WithHooks(transaction.Hooks{Fault: func(e transaction.Event) error {
					if e.Point != transaction.PointPrepared {
						return nil
					}
					raw, err := os.ReadFile(filepath.Join(fx.home, "state", "transactions", "v1", e.TransactionID+".json"))
					if err != nil {
						t.Fatal(err)
					}
					var j transaction.Journal
					if err = json.Unmarshal(raw, &j); err != nil {
						t.Fatal(err)
					}
					if j.Unit != nil && len(j.Unit.Native) == 1 && j.Unit.Native[0].Root == native {
						id = e.TransactionID
						panic("two-pending-stop")
					}
					return nil
				}}))
				if err != nil {
					t.Fatal(err)
				}
				func() {
					defer op.close()
					defer func() {
						if got := recover(); got != "two-pending-stop" {
							t.Fatalf("wrong publisher stop: %v", got)
						}
					}()
					rows, err := materializeScopeWithNativeHome(op, fx.home, profile, "codex_cli", Policy{}, func(string) (string, error) { return native, nil })
					t.Fatalf("native publisher did not reach Prepare: %v %+v", err, rows)
				}()
				if id == "" {
					t.Fatal("no actual native unit")
				}
				return filepath.Join(fx.home, "state", "transactions", "v1", id+".json")
			}
			j1 := stop(fx.profile, a)
			hidden := filepath.Join(t.TempDir(), "j1.private")
			if err := os.Rename(j1, hidden); err != nil {
				t.Fatal(err)
			}
			target := b
			if overlap {
				target = a
			}
			t.Setenv("CODEX_HOME", b)
			_ = req // Recorded managed origin remains separate from native surface custody.
			j2 := stop("pending-q", target)
			if err := os.Rename(hidden, j1); err != nil {
				t.Fatal(err)
			}
			raw1, err := os.ReadFile(j1)
			if err != nil {
				t.Fatal(err)
			}
			raw2, err := os.ReadFile(j2)
			if err != nil {
				t.Fatal(err)
			}
			preA, preB := nativeSnapshot(t, a), nativeSnapshot(t, b)
			pending, err := envauthority.PendingUnits(fx.home)
			if overlap {
				if err == nil || !strings.Contains(err.Error(), envorigin.DiagInvalid) || !strings.Contains(err.Error(), "overlapping pending unit effects") {
					t.Fatalf("overlapping real journals admitted: %v", err)
				}
				after1, e1 := os.ReadFile(j1)
				after2, e2 := os.ReadFile(j2)
				if e1 != nil || e2 != nil || string(after1) != string(raw1) || string(after2) != string(raw2) || nativeSnapshot(t, a) != preA || nativeSnapshot(t, b) != preB {
					t.Fatal("pending-pair refusal changed preimages")
				}
				return
			}
			if err != nil || len(pending.Units) != 2 {
				t.Fatalf("disjoint actual pending units refused: %v", err)
			}
			op, err := beginOperation(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			op.close()
			for _, native := range []string{a, b} {
				if _, err := os.Readlink(filepath.Join(native, "AGENTS.md")); err != nil {
					t.Fatalf("disjoint recovery did not reach context alias: %v", err)
				}
			}
			pending, err = envauthority.PendingUnits(fx.home)
			if err != nil || len(pending.Units) != 0 {
				t.Fatalf("disjoint cleanup not completed: %v", err)
			}
		})
	}
}
