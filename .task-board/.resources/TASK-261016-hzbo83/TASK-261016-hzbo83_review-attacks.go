package envprofile

import (
 "os"
 "path/filepath"
 "reflect"
 "strings"
 "testing"
 "github.com/relux-works/curator/internal/envregistry"
 "github.com/relux-works/curator/internal/stateread"
)

// A transient first native-config read failure must not be retried into success
// by repair; a later public call must recover and preserve native credentials.
func TestReviewTransientPolicyRead(t *testing.T) {
 fx := writeManagedFixture(t, "review-transient")
 seedLiveNativeCredentials(t, fx)
 provision(t, fx, envregistry.CodexCLI, envregistry.DefaultMachineConfig())
 req := fx.request(envregistry.CodexCLI)
 req.Repair = true
 path := filepath.Join(fx.native[envregistry.CodexCLI], "config.toml")
 reads := 0
 req.readRegularFile = func(p string) (stateread.File,error) {
  if p == path { reads++; if reads == 1 { return stateread.File{}, os.ErrPermission } }
  return stateread.ReadRegularFile(p)
 }
 before := snapshotCredentialScope(t,fx)
 r,err := Resolve(req)
 if err == nil || !strings.Contains(err.Error(),envregistry.DiagSeedUnreadable) || (r != nil && len(r.Document)>0) || reads != 1 { t.Fatalf("transient failure admitted or retried: result=%+v err=%v reads=%d",r,err,reads) }
 if !reflect.DeepEqual(before,snapshotCredentialScope(t,fx)) { t.Fatal("failed repair changed protected state") }
 r,err = Resolve(req)
 if err != nil || r == nil || len(r.Document)==0 || reads != 2 { t.Fatalf("next call failed recovery: result=%+v err=%v reads=%d",r,err,reads) }
}

// Status must see valid backend drift and malformed selectors on successive
// calls, return current after restoration, and never modify credentials/links.
func TestReviewStatusBackendSwitch(t *testing.T) {
 fx := writeManagedFixture(t, "review-status")
 seedLiveNativeCredentials(t,fx)
 provision(t,fx,envregistry.CodexCLI,envregistry.DefaultMachineConfig())
 path := filepath.Join(fx.native[envregistry.CodexCLI],"config.toml")
 for _,store := range []string{"keyring","invalid","file"} {
  if err:=os.WriteFile(path,[]byte("cli_auth_credentials_store = '"+store+"'\n"),0600);err!=nil {t.Fatal(err)}
  before:=snapshotCredentialScope(t,fx)
  result,err:=StatusOf(statusRequest(fx))
  if err!=nil {t.Fatal(err)}
  row:=findHome(result,fx.profile,envregistry.CodexCLI)
  if row==nil || row.Current!=(store=="file") {t.Fatalf("backend drift misreported: %s %+v",store,row)}
  if store=="invalid" && !strings.Contains(strings.Join(row.Findings,";"),envregistry.DiagCredentialUnsupported) {t.Fatalf("lost invalid-selector diagnostic: %+v",row)}
  if !reflect.DeepEqual(before,snapshotCredentialScope(t,fx)) {t.Fatal("status modified protected state")}
 }
}

// Changing native file to keyring cannot make repair erase either a recorded
// native link or a detached vendor-owned regular credential. File is the
// positive control and restoring it re-admits the intact-link home.
func TestReviewDetachedBackendSwitch(t *testing.T) {
 for _, detached:=range []bool{false,true} { t.Run(map[bool]string{false:"link",true:"detached"}[detached],func(t *testing.T) {
  fx:=writeManagedFixture(t,"review-switch")
  seedLiveNativeCredentials(t,fx)
  provision(t,fx,envregistry.CodexCLI,envregistry.DefaultMachineConfig())
  req:=fx.request(envregistry.CodexCLI); req.Repair=true
  if r,err:=Resolve(req);err!=nil || r==nil || len(r.Document)==0 {t.Fatalf("file control failed: %v",err)}
  link:=filepath.Join(ManagedHomeDir(fx.home,fx.profile,envregistry.CodexCLI),"auth.json")
  if detached {
   if err:=os.Remove(link);err!=nil {t.Fatal(err)}
   if err:=os.WriteFile(link,[]byte("synthetic-detached-credential"),0600);err!=nil {t.Fatal(err)}
  }
  path:=filepath.Join(fx.native[envregistry.CodexCLI],"config.toml")
  if err:=os.WriteFile(path,[]byte("cli_auth_credentials_store = 'keyring'\n"),0600);err!=nil {t.Fatal(err)}
  before:=snapshotCredentialScope(t,fx)
  r,err:=Resolve(req)
  if err==nil || !strings.Contains(err.Error(),envregistry.DiagCredentialConflict) || (r!=nil && len(r.Document)>0) {t.Fatalf("backend switch silently migrated: %+v %v",r,err)}
  if !reflect.DeepEqual(before,snapshotCredentialScope(t,fx)) {t.Fatal("repair changed credentials, links or markers")}
  if !detached {
   if err:=os.WriteFile(path,[]byte("cli_auth_credentials_store = 'file'\n"),0600);err!=nil {t.Fatal(err)}
   if r,err:=Resolve(req);err!=nil || r==nil || len(r.Document)==0 {t.Fatalf("restored file control failed: %v",err)}
  }
 }) }
}
