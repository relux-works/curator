package envprofile
import (
 "os"
 "path/filepath"
 "strings"
 "testing"
 "github.com/relux-works/curator/internal/envregistry"
)
func TestReviewerUnreadableNativeConfig(t *testing.T) {
 for _, mode := range []string{"shared", "isolated"} { t.Run(mode, func(t *testing.T) {
  fx := writeManagedFixture(t, "acme")
  config := filepath.Join(fx.native["codex_cli"], "config.toml")
  if err := os.Mkdir(config, 0700); err != nil { t.Fatal(err) }
  req := fx.request("codex_cli"); req.Repair = true
  req.Machine = envregistry.DefaultMachineConfig()
  req.Machine.Isolation = map[string]map[string]string{"acme": {"codex_cli":mode}}
  _, err := Resolve(req)
  if err == nil || (!strings.Contains(err.Error(), "config.toml") || !strings.Contains(err.Error(), "unreadable")) { t.Fatalf("unreadable config must refuse, got %v", err) }
  if _, err := os.Lstat(ManagedHomeDir(fx.home,"acme","codex_cli")); !os.IsNotExist(err) { t.Fatalf("refused provisioning wrote home: %v",err) }
 }) }
}
