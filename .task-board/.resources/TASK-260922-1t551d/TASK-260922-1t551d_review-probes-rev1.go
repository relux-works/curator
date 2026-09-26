package envprofile
import (
 "os"
 "path/filepath"
 "strings"
 "testing"
 "github.com/relux-works/curator/internal/envregistry"
)
func TestReviewerDanglingExpectedTarget(t *testing.T) {
 fx := writeManagedFixture(t, "acme")
 provision(t, fx, "pi", envregistry.DefaultMachineConfig())
 status, err := StatusOf(statusRequest(fx)); if err != nil { t.Fatal(err) }
 row := findHome(status, "acme", "pi")
 if row == nil || row.Current || !strings.Contains(strings.Join(row.Findings,";"), envregistry.DiagCredentialConflict) { t.Errorf("dangling target must be detached with conflict: %+v", row) }
 _, err = Resolve(fx.request("pi"))
 if err == nil { t.Errorf("Resolve silently admits dangling target") }
}
func TestReviewerLiteralCodexSelector(t *testing.T) {
 for _, value := range []string{"keyring", "auto", "ephemeral"} { t.Run(value, func(t *testing.T) {
 fx := writeManagedFixture(t,"acme")
 config := "cli_auth_credentials_store = '"+value+"'\n"
 if err := os.WriteFile(filepath.Join(fx.native["codex_cli"],"config.toml"), []byte(config),0600); err != nil { t.Fatal(err) }
 req := fx.request("codex_cli"); req.Repair = true
 req.Machine = envregistry.DefaultMachineConfig()
 req.Machine.Isolation = map[string]map[string]string{"acme":{"codex_cli":"isolated"}}
 _, err := Resolve(req)
 want := envregistry.DiagIsolatedUnsupported; if value == "ephemeral" { want = envregistry.DiagCredentialUnsupported }
 if err == nil || !strings.Contains(err.Error(),want) { t.Fatalf("valid TOML literal selector must refuse with %s, got %v",want,err) }
 }) }
}
