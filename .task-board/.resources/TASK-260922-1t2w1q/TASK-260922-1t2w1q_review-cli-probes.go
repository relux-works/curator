package main
import("os";"strings";"testing")
type reviewPlanWriter struct { t *testing.T; link, old string; checked bool }
func(w *reviewPlanWriter) Write(p []byte)(int,error){ if !w.checked {w.checked=true; got,e:=os.Readlink(w.link); if e!=nil || got!=w.old {w.t.Errorf("first stdout write occurs AFTER mutation: target=%q err=%v; wanted prior %q",got,e,w.old)} };return len(p),nil }
func TestReviewerMigrationCLI(t *testing.T){
 for _,scenario:=range []string{"missing-plan","print-before-write"}{t.Run(scenario,func(t *testing.T){
 source,_:=profileHome(t); link,old,_:=installPiProfile(t,source);os.Remove(link);if e:=os.Symlink(old,link);e!=nil{t.Fatal(e)}
 if scenario=="missing-plan" {code,_,errout:=runProfile(t,source,"env","migrate","--apply"); got,_:=os.Readlink(link);if code==exitOK || got!=old {t.Fatalf("apply without prior plan succeeded/mutated: code=%d target=%q stderr=%s",code,got,errout)};return}
 _,plan,_:=runProfile(t,source,"env","migrate","--plan");w:=&reviewPlanWriter{t:t,link:link,old:old};var stderr strings.Builder
 code:=run([]string{"env","migrate","--apply","--expect",planHash(t,plan)},source,w,&stderr);if code!=exitOK {t.Fatal(stderr.String())};if !w.checked{t.Fatal("no printed plan")}
 })}
}
