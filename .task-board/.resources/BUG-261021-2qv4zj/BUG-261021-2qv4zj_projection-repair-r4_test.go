package envprofile
import("os";"path/filepath";"strings";"testing")
// Existing intact stores resolve; missing, extra and changed projected files
// must be diagnosed without writes and restored by explicit production repair.
func TestReviewerProjectionRepair(t *testing.T){
 for _,id:=range []string{"claude_code","codex_cli"}{for _,mode:=range []string{"intact","missing-directory","missing-file","extra-file","changed-file"}{t.Run(id+"/"+mode,func(t *testing.T){
  fx:=writeManagedFixture(t,"sample");req:=fx.request(id);req.Repair=true
  if _,err:=Resolve(req);err!=nil{t.Fatal(err)}
  live:=filepath.Join(ManagedHomeDir(fx.home,fx.profile,id),"skills/myskill");root,err:=os.Readlink(live);if err!=nil{t.Fatal(err)}
  original,err:=os.ReadFile(filepath.Join(root,"SKILL.md"));if err!=nil{t.Fatal(err)}
  switch mode{case "missing-directory":err=os.RemoveAll(root);case "missing-file":err=os.Remove(filepath.Join(root,"SKILL.md"));case "extra-file":err=os.WriteFile(filepath.Join(root,"extra.md"),[]byte("extra\n"),0600);case "changed-file":err=os.WriteFile(filepath.Join(root,"SKILL.md"),[]byte("changed\n"),0600)};if err!=nil{t.Fatal(err)}
  req.Repair=false;_,err=Resolve(req)
  if mode=="intact"{if err!=nil{t.Fatal(err)};return}
  if err==nil||!strings.Contains(err.Error(),DiagHomeStale){t.Fatalf("expected stale refusal: %v",err)}
  if mode=="missing-directory"{if _,e:=os.Stat(root);!os.IsNotExist(e){t.Fatal("status wrote store")}}
  req.Repair=true
  if _,err:=Resolve(req);err!=nil{t.Fatalf("projection repair did not converge for %s: %v",mode,err)}
  req.Repair=false;if _,err:=Resolve(req);err!=nil{t.Fatal(err)}
  got,err:=os.ReadFile(filepath.Join(root,"SKILL.md"));if err!=nil||string(got)!=string(original){t.Fatalf("restored bytes: %q %v",got,err)}
  if _,err:=os.Stat(filepath.Join(root,"extra.md"));!os.IsNotExist(err){t.Fatal("extra survived")}
 })}}
}
