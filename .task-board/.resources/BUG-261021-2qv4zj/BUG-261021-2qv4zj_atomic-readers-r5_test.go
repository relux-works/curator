package envprofile
import (
 "os"
 "path/filepath"
 "testing"
 "github.com/relux-works/curator/internal/contextlock"
 "github.com/relux-works/curator/internal/transaction"
)
// An atomic rebuild must keep both existing consumer links readable at every
// swap boundary. Old and new complete bytes are valid; ENOENT is not either.
func TestReviewerR5AtomicReaders(t *testing.T){
 for _,id:=range []string{"claude_code","codex_cli"}{t.Run(id,func(t *testing.T){
 fx:=writeManagedFixture(t,"sample");req:=fx.request(id);req.Repair=true
 if _,err:=Resolve(req);err!=nil{t.Fatal(err)}
 lock,_,err:=readLock(fx.home,fx.profile);if err!=nil{t.Fatal(err)}
 source,err:=os.ReadFile(sourcePath(fx.home,fx.profile));if err!=nil{t.Fatal(err)}
 if _,err:=contextlock.Write(lockPath(fx.home,"sibling"),lock);err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(sourcePath(fx.home,"sibling"),source,0600);err!=nil{t.Fatal(err)}
 sibling:=req;sibling.Profile="sibling";if _,err:=Resolve(sibling);err!=nil{t.Fatal(err)}
 first:=filepath.Join(ManagedHomeDir(fx.home,fx.profile,id),"skills/myskill")
 second:=filepath.Join(ManagedHomeDir(fx.home,"sibling",id),"skills/myskill")
 root,err:=os.Readlink(first);if err!=nil{t.Fatal(err)};other,err:=os.Readlink(second);if err!=nil||other!=root{t.Fatal("fixture does not share projection")}
 good,err:=os.ReadFile(filepath.Join(root,"SKILL.md"));if err!=nil{t.Fatal(err)}
 old:="old complete guide\n";writeGitFile(t,root,"SKILL.md",old);writeGitFile(t,root,"extra.md","old extra\n")
 seen:=map[transaction.Point]int{}
 req.transactionOptions=[]transaction.Option{transaction.WithHooks(transaction.Hooks{Fault:func(e transaction.Event)error{
 if e.Class!="skill-projection"{return nil}
 if e.Point!=transaction.PointBeforeBackup&&e.Point!=transaction.PointAfterBackup&&e.Point!=transaction.PointAfterInstall{return nil}
 seen[e.Point]++
 for n,live:=range []string{first,second}{
 b,err:=os.ReadFile(filepath.Join(live,"SKILL.md"))
 if err!=nil{t.Errorf("atomic reader unavailable: consumer=%d point=%s missing=%t",n,e.Point,os.IsNotExist(err));continue}
 if string(b)!=old&&string(b)!=string(good){t.Errorf("partial tree at %s",e.Point)}
 }
 return nil
 }})}
 if _,err:=Resolve(req);err!=nil{t.Fatalf("repair: %v",err)}
 for _,point:=range []transaction.Point{transaction.PointBeforeBackup,transaction.PointAfterBackup,transaction.PointAfterInstall}{if seen[point]!=1{t.Fatalf("boundary %s observed %d times",point,seen[point])}}
 sibling.Repair=false;if _,err:=Resolve(sibling);err!=nil{t.Fatal(err)}
 })}
}
