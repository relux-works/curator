package audit

import ("bytes"; "fmt"; "os"; "strings"; "testing"; "github.com/relux-works/curator/internal/capabilities"; "github.com/relux-works/curator/internal/hashing")

// Reviewer probe: different nonempty declarations must not cross-hit in dry-run;
// dry-run leaves cached bytes unchanged and writable retry recomputes.
func TestReviewerNonemptyReadonlyCrossover(t *testing.T) {
 for _, version := range []hashing.Version{hashing.VersionV1, hashing.VersionV2} {
  for _, allowFirst := range []bool{false,true} {
   t.Run(fmt.Sprintf("v%d/allow-first=%v",version,allowFirst),func(t *testing.T){
    cfg:=newCfg(t,"strict","medium")
    s:=subjectWith(t,"curl https://api.example.com/data\nsubprocess(\"git\")\n",capabilities.ImplicitNone(),3)
    s.HashVersion=version
    set:=func(allow bool){s.Capabilities=capabilities.Manifest{Network:[]string{"other.example.com"},Exec:[]string{"other"}};if allow{s.Capabilities=capabilities.Manifest{Network:[]string{"api.example.com"},Exec:[]string{"git"}}}}
    set(allowFirst)
    Gate(cfg,[]Subject{s})
    digest,err:=hashing.ContentSHA256WithVersion(s.Snapshot,nil,version);if err!=nil{t.Fatal(err)}
    path:=verdictPath(cfg,digest)
    old,err:=os.ReadFile(path);if err!=nil{t.Fatal(err)}
    set(!allowFirst)
    for retry:=0;retry<2;retry++ {
     _,errs:=GateReadOnly(cfg,[]Subject{s})
     if !allowFirst && len(errs)!=0{t.Fatalf("readonly false refusal: %v",errs)}
     if allowFirst && (!strings.Contains(strings.Join(errs,";"),"network-undeclared") || !strings.Contains(strings.Join(errs,";"),"exec-undeclared")){t.Fatalf("readonly stale allow: %v",errs)}
     now,err:=os.ReadFile(path);if err!=nil || !bytes.Equal(old,now){t.Fatal("readonly changed cached bytes")}
    }
    _,errs:=Gate(cfg,[]Subject{s});if (len(errs)==0)==allowFirst{t.Fatalf("retry wrong decision: %v",errs)}
    report,err:=auditSubject(cfg,s,false);if err!=nil || !report.CacheHit{t.Fatalf("retry did not cache: %+v %v",report,err)}
   })
  }
 }
}
