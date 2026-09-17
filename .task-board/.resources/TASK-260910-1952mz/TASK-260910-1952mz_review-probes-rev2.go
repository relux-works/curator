package shell
import("os";"os/exec";"path/filepath";"strings";"testing";"time";"github.com/relux-works/curator/internal/hookapproval")
func TestReviewerMalformedApproval(t *testing.T) {
 for _, row:=range []string{"two-fields","bad-approver","bad-timestamp"} {t.Run(row,func(t *testing.T){
 base,_:=filepath.EvalSymlinks(t.TempDir()); project:=filepath.Join(base,"project"); home:=filepath.Join(base,"home"); os.MkdirAll(filepath.Join(project,".agents"),0700); os.MkdirAll(home,0700)
 candidate:=filepath.Join(project,".agents","env.sh"); payload:=[]byte("export CURATOR_PROJECT_ENV=1\n"); os.WriteFile(candidate,payload,0600)
 record:=candidate+"\t"+hookapproval.Digest(payload)
 switch row {case "bad-approver":record+="\tproject\t2026-09-17T00:00:00Z";case "bad-timestamp":record+="\tmanager\tnot-a-time"}
 os.WriteFile(hookapproval.ApprovalsPath(home),[]byte(record+"\n"),0600)
 _,err:=hookapproval.List(home);if err==nil {t.Fatal("Go reader unexpectedly accepted")}
 out,diag:=runPosixTrustActivation(t,project,home,TrustProfileBEnforcing)
 if strings.Contains(out,"sourced1=1") {t.Fatalf("INVALID RECORD SOURCED under B: %s stderr=%q",out,diag)}
 })}
}
func TestReviewerAlias(t *testing.T){
 base,_:=filepath.EvalSymlinks(t.TempDir()); project:=filepath.Join(base,"project"); home:=filepath.Join(base,"home"); os.MkdirAll(filepath.Join(project,".agents"),0700); os.MkdirAll(home,0700)
 candidate:=filepath.Join(project,".agents","env.sh");os.WriteFile(candidate,[]byte("export CURATOR_PROJECT_ENV=1\n"),0600)
 alias:=filepath.Join(base,"alias");os.Symlink(project,alias)
 _,err:=hookapproval.ApproveFile(home,candidate,hookapproval.ApprovedByManager,time.Now());if err!=nil{t.Fatal(err)}
 _,found,err:=hookapproval.Lookup(home,filepath.Join(alias,".agents","env.sh"));t.Logf("alias lookup: found=%v err=%v",found,err)
 out,diag:=runPosixTrustActivation(t,alias,home,TrustProfileBEnforcing)
 if !strings.Contains(out,"sourced1=1"){t.Fatalf("APPROVED ALIAS REFUSED: %s stderr=%q",out,diag)}
}
func TestReviewerPOSIX(t *testing.T){for _,name:=range []string{"sh","dash"}{t.Run(name,func(t *testing.T){exe,err:=exec.LookPath(name);if err!=nil{t.Skip(err)};hook,_:=HookWithProfile("bash",false,TrustProfileBEnforcing);p:=filepath.Join(t.TempDir(),"hook");os.WriteFile(p,[]byte(hook),0600);out,err:=exec.Command(exe,"-n",p).CombinedOutput();if err!=nil{t.Fatalf("%s: %v %s",exe,err,out)}})}}
