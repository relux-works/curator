package install
import("testing";"os";"strings";"encoding/json";"path/filepath";"github.com/relux-works/curator/internal/audit")
func TestReviewRev9(t *testing.T){
 project,home,_:=strictLocalProject(t,""); cfg:=draftAuditConfig(home)
 assertGatePassed(t,draftProjectResult(cfg,project,home,false))
 op,rp:=auditBindingPaths(t,home); ob,_:=os.ReadFile(op); rb,_:=os.ReadFile(rp)
 for _,tc:=range []struct{name,target,prefix string}{
 {"object-escaped-duplicate","object",`"schema_vers\u0069on":999,`},
 {"report-escaped-duplicate","report",`"rev\u006fked":true,`},
 {"object-foreign-arm","object",""},{"report-null","report", ""},
 }{t.Run(tc.name,func(t *testing.T){
 os.WriteFile(op,ob,0600);os.WriteFile(rp,rb,0600)
 path,raw:=op,ob;if tc.target=="report"{path,raw=rp,rb}
 forged:=[]byte("{"+tc.prefix+string(raw[1:]))
 if tc.prefix=="" {var d map[string]any;json.Unmarshal(raw,&d);if tc.target=="object"{d["package"].(map[string]any)["commit"]=nil}else{d["revoked"]=nil};forged,_=json.Marshal(d)}
 os.WriteFile(path,forged,0600);if tc.target=="report"{rewriteAuditObject(t,op,map[string]any{"evidence_sha256":audit.EvidenceDigest(forged)})}
 beforeO,_:=os.ReadFile(op);beforeR,_:=os.ReadFile(rp)
 for _,dry:=range []bool{true,false}{r:=draftProjectResult(cfg,project,home,dry);if !strings.Contains(strings.Join(r.Errors,";"),"source_audit"){t.Errorf("dry=%v admitted: %+v",dry,r)}}
 afterO,_:=os.ReadFile(op);afterR,_:=os.ReadFile(rp);if string(beforeO)!=string(afterO)||string(beforeR)!=string(afterR){t.Error("records overwritten")}
 })}
 os.WriteFile(op,ob,0600);os.WriteFile(rp,rb,0600)
 t.Run("dangling-existing-binding",func(t *testing.T){
 if err:=os.Remove(op);err!=nil{t.Fatal(err)}
 target:=filepath.Join(home,"missing-binding-target.json");if err:=os.Symlink(target,op);err!=nil{t.Fatal(err)}
 for _,dry:=range []bool{true,false}{r:=draftProjectResult(cfg,project,home,dry);if !strings.Contains(strings.Join(r.Errors,";"),"source_audit"){t.Errorf("dry=%v broken binding admitted: %+v",dry,r)}}
 if _,err:=os.Stat(target);err==nil{t.Error("broken existing binding was re-established through symlink target")}
 })
 t.Run("unreadable-existing-binding",func(t *testing.T){
 os.Remove(op); if err:=os.Symlink(op,op);err!=nil{t.Fatal(err)}
 broken:=[]byte("{}");os.WriteFile(rp,broken,0600)
 r:=draftProjectResult(cfg,project,home,false)
 after,_:=os.ReadFile(rp)
 if string(after)!=string(broken){t.Errorf("unreadable binding caused corrupt existing report overwrite: %+v",r)}
 })
}
