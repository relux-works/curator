package audit
import (
 "os"
 "encoding/json"
 "reflect"
 "testing"
 "github.com/relux-works/curator/internal/capabilities"
 "github.com/relux-works/curator/internal/skillspec"
)
func TestReviewerCurrentManifestBackfill(t *testing.T) {
 for _, dirty := range []bool{false,true} { t.Run(map[bool]string{false:"clean",true:"cached-finding"}[dirty],func(t *testing.T){
 cfg:=newCfg(t,"advisory","high")
 subject:=scriptSubject(t,enforcedCommands(),capabilities.ImplicitNone(),8)
 Gate(cfg,[]Subject{subject})
 path:=soleVerdictPath(t,cfg)
 old:=readStoredVerdict(t,path)
 delete(old,"script_policies")
 if dirty { old["findings"]=[]byte(`[{"id":"cached-sentinel","severity":"high","file":"sentinel","evidence":"cached only","verifiable":true}]`) }
 writeStoredVerdict(t,path,old)
 subject.Commands=map[string]skillspec.Command{"current":{Name:"current",Type:"script"},"guarded":{Name:"guarded",Type:"script",ExecutionPolicy:"script-worker-v1"}}
 before,_:=os.ReadFile(path)
 w,e:=GateReadOnly(cfg,[]Subject{subject}); if len(e)!=0 {t.Fatal(e)}
 after,_:=os.ReadFile(path); if string(before)!=string(after){t.Fatal("read-only wrote")}
 report,err:=auditSubject(cfg,subject,false); if err!=nil||!report.CacheHit||len(report.ScriptPolicies)!=2 {t.Fatalf("report: %+v %v",report,err)}
 records:=ScriptPolicyRecords([]Subject{subject}); if len(records)!=2||records[0].Command!="current"||records[0].Policy!=""||records[1].Policy!="script-worker-v1" {t.Fatal(records)}
 w2,e2:=Gate(cfg,[]Subject{subject}); if !reflect.DeepEqual(w,w2)||!reflect.DeepEqual(e,e2){t.Fatalf("decision changed %v %v -> %v %v",w,e,w2,e2)}
 got:=readStoredVerdict(t,path); var a,b any
 _=a;_=b
 if string(got["findings"])=="" {t.Fatal("lost findings")}
 policies:=storedPolicyEntries(t,got); if len(policies)!=2||policies["guarded"]!="script-worker-v1" {t.Fatal(policies)}
 if policy,ok:=policies["current"];!ok||policy!="" {t.Fatal(policies)}
 // JSON formatting may change; semantic findings must not.
 var oldFindings,newFindings []Finding
 decodeReviewer(t,old["findings"],&oldFindings);decodeReviewer(t,got["findings"],&newFindings)
 if !reflect.DeepEqual(oldFindings,newFindings){t.Fatalf("cached findings changed: %v -> %v",oldFindings,newFindings)}
 }) }
}

func decodeReviewer(t *testing.T,b []byte,v any){t.Helper();if err:=json.Unmarshal(b,v);err!=nil{t.Fatal(err)}}
