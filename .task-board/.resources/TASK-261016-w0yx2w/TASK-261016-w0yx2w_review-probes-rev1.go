//go:build darwin || linux
package envsnapshot
import("context";"os";"path/filepath";"strings";"testing")
// A retry cannot adopt a competing snapshot; a cancelled absent publication can retry.
func TestReviewerPublicationRetry(t *testing.T){
 for _,kind:=range []string{"cancel","competitor","post-mismatch"}{t.Run(kind,func(t *testing.T){
 root,policy:=destinationSandbox(t);publishControl(t,root,policy);target:=filepath.Join(root,"target");d:=admitDestination(t,target,policy,DestinationAbsent)
 ctx,cancel:=context.WithCancel(context.Background());defer cancel()
 d.publicationHook=func(phase string,_ *os.File)error{
 if kind=="cancel"&&phase=="file-staged"{cancel()};if kind=="competitor"&&phase=="before-rename"{snapshotAtDestination(t,target)};if kind=="post-mismatch"&&phase=="after-rename"{write(t,target,"extra",[]byte("keep"),0600)};return nil}
 v,err:=d.Publish(ctx,publicationFixture(t));if err==nil||v!=nil{t.Fatal("first attack admitted")};d.publicationHook=nil
 v,err=d.Publish(context.Background(),publicationFixture(t));if kind=="cancel"{if err!=nil||v==nil{t.Fatal("valid retry refused",err)};verified(t,target)}else{requireDestinationError(t,err,DestinationInvalid,"target-changed");if v!=nil{t.Fatal("retry success")};if kind=="competitor"{verified(t,target)}else{b,e:=os.ReadFile(filepath.Join(target,"extra"));if e!=nil||string(b)!="keep"{t.Fatal("rollback",e)}}}
 })}
}
// Top-level stage substitution must refuse success and retain the published replacement.
func TestReviewerStageRootBeforeRename(t *testing.T){
 root,policy:=destinationSandbox(t);publishControl(t,root,policy);target:=filepath.Join(root,"target");d:=admitDestination(t,target,policy,DestinationAbsent)
 d.publicationHook=func(phase string,_ *os.File)error{if phase!="before-rename"{return nil};entries,e:=os.ReadDir(root);if e!=nil{t.Fatal(e)};for _,entry:=range entries{if strings.HasPrefix(entry.Name(),".curator-snapshot-"){stage:=filepath.Join(root,entry.Name());if e:=os.Rename(stage,stage+"-retained");e!=nil{t.Fatal(e)};if e:=os.Mkdir(stage,0700);e!=nil{t.Fatal(e)};write(t,stage,"sentinel",[]byte("keep"),0600)}};return nil}
 v,err:=d.Publish(context.Background(),publicationFixture(t));requireDestinationError(t,err,Mismatch,"published-identity");if v!=nil{t.Fatal("substitution admitted")};b,e:=os.ReadFile(filepath.Join(target,"sentinel"));if e!=nil||string(b)!="keep"{t.Fatal("rollback",e)};verified(t,filepath.Join(root,"control"))
}
