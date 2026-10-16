package envprofile

import (
 "bytes"
 "os"
 "path/filepath"
 "testing"
)

// Late corrupt recovery evidence must refuse without cleaning the injected
// journal; the identical callback with no injection must still capture.
func TestReviewCaptureLateJournal(t *testing.T) {
 fx, req := captureFixture(t)
 if _, err := CaptureMaterialize(req); err != nil { t.Fatal(err) }
 original := req.Capture
 var expected []byte
 req.Capture = func(c *MaterializeCapture) ([]byte,error) {
  b,e := original(c); if e != nil { return nil,e }
  root := filepath.Join(fx.home,"state","transactions","v1")
  if e=os.MkdirAll(root,0700); e != nil { return nil,e }
  if e=os.WriteFile(filepath.Join(root,"late.json"),[]byte("{"),0600); e != nil { return nil,e }
  expected,e=captureFixtureState(t,fx)
  return b,e
 }
 requireCaptureRefusal(t,req,DiagMaterializeCaptureDrift)
 after,e:=captureFixtureState(t,fx); if e != nil { t.Fatal(e) }
 if !bytes.Equal(expected,after) { t.Fatal("late evidence mutated") }
}

// Each mutation-capable input flag is refused before callbacks or any fixture
// write; a valid request succeeds immediately before those attacks.
func TestReviewCaptureRejectsMutationFlags(t *testing.T) {
 for _, flag := range []string{"repair","dry-run","preserve-lock"} {
  t.Run(flag,func(t *testing.T){
   fx,req:=captureFixture(t)
   if _,e:=CaptureMaterialize(req); e != nil { t.Fatal(e) }
   before,e:=captureFixtureState(t,fx); if e != nil { t.Fatal(e) }
   switch flag {case "repair":req.Resolve.Repair=true;case "dry-run":req.Resolve.DryRun=true;case "preserve-lock":req.Resolve.PreserveLock=true}
   req.Capture=func(*MaterializeCapture)([]byte,error){t.Fatal("capture reached");return nil,nil}
   requireCaptureRefusal(t,req,"usage")
   after,e:=captureFixtureState(t,fx);if e != nil {t.Fatal(e)}
   if !bytes.Equal(before,after) {t.Fatal("usage refusal mutated source")}
  })
 }
}
