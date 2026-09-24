package audit
import (
 "encoding/json"
 "os"
 "path/filepath"
 "testing"
 "github.com/relux-works/curator/internal/capabilities"
)
func TestReviewerExistingVerdictGetsPolicyRecord(t *testing.T) {
 cfg := newCfg(t, "advisory", "high")
 subject := scriptSubject(t, enforcedCommands(), capabilities.ImplicitNone(), 8)
 if _, errs := Gate(cfg, []Subject{subject}); len(errs) != 0 { t.Fatal(errs) }
 paths, _ := filepath.Glob(filepath.Join(cfg.Home(), "audit", "*", "verdict-*.json"))
 data, _ := os.ReadFile(paths[0])
 var stored map[string]json.RawMessage
 if err := json.Unmarshal(data, &stored); err != nil { t.Fatal(err) }
 delete(stored, "script_policies") // valid pre-R4 cached verdict
 data, _ = json.Marshal(stored)
 if err := os.WriteFile(paths[0], data, 0600); err != nil { t.Fatal(err) }
 if _, errs := Gate(cfg, []Subject{subject}); len(errs) != 0 { t.Fatal(errs) }
 data, _ = os.ReadFile(paths[0])
 if err := json.Unmarshal(data, &stored); err != nil { t.Fatal(err) }
 if _, ok := stored["script_policies"]; !ok { t.Fatalf("production Gate left existing verdict without per-command identity: %s", data) }
}
