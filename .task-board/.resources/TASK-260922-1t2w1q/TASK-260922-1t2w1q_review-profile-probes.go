package envprofile
import("os";"path/filepath";"strings";"testing";"github.com/relux-works/curator/internal/envmarker";"github.com/relux-works/curator/internal/envregistry")
func TestReviewerMarkerDrift(t *testing.T){
 fx:=writeManagedFixture(t,"acme");provision(t,fx,"pi",envregistry.DefaultMachineConfig());link:=filepath.Join(ManagedHomeDir(fx.home,"acme","pi"),"auth.json");os.Remove(link);os.Symlink(filepath.Join(fx.native["pi"],"auth.json"),link)
 req:=fx.migrateRequest();before,e:=PlanMigration(req);if e!=nil{t.Fatal(e)};mp:=filepath.Join(ManagedHomeDir(fx.home,"acme","pi"),envmarker.Name);raw,e:=os.ReadFile(mp);if e!=nil{t.Fatal(e)};marker,e:=envmarker.Parse(raw);if e!=nil{t.Fatal(e)};marker.Profile.LockSHA256=strings.Repeat("a",64);updated,e:=marker.Marshal();if e!=nil{t.Fatal(e)};if e=os.WriteFile(mp,updated,0600);e!=nil{t.Fatal(e)};req.Expect=before.Hash
 _,e=ApplyMigration(req);if e==nil{t.Fatal("changed old marker profile.lock_sha256 accepted with stale hash")}
}
func TestReviewerFailedRelinkRollback(t *testing.T){
 home:=t.TempDir();full:=filepath.Join(ManagedHomeDir(home,"acme","pi"),"auth.json");os.MkdirAll(filepath.Dir(full),0755);os.Symlink("old-target",full)
 ops,e:=executeMigrationOps(MigrateRequest{Home:home},[]MigrateOp{{Profile:"acme",EnvID:"pi",Kind:MigrateOpRelink,Path:"auth.json",From:"old-target",To:strings.Repeat("x",100000)}});if e==nil{t.Fatal("expected OS rejection of oversized target")};rb:=rollbackMigrationLinks(home,ops);got,re:=os.Readlink(full);if rb!=nil || re!=nil || got!="old-target"{t.Fatalf("failed relink lost old link despite rollback: applied=%d rollback=%v readlink=%v target=%q",len(ops),rb,re,got)}
}
