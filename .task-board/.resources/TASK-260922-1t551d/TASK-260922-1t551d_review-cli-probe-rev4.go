package main
import (
 "strings"
 "testing"
)
func TestReviewerCredentialCLI(t *testing.T) {
 source, _ := profileHome(t)
 pkg := t.TempDir()
 writeContextPackage(t,pkg,"acme","1.0.0","hello\n")
 if code,_,errout:=runProfile(t,source,"profile","install",pkg);code!=exitOK{t.Fatalf("install: %s",errout)}
 if code,_,errout:=runProfile(t,source,"env","resolve","pi","--profile","acme","--repair");code!=exitOK || !strings.Contains(errout,"detached-pending"){t.Fatalf("pending resolve code=%d: %s",code,errout)}
 if _,out,errout:=runProfile(t,source,"env","status","--json");!strings.Contains(out,"detached-pending"){t.Fatalf("pending status: %s %s",out,errout)}
}
