package main
import("testing";"reflect")
// Repeated activation aliases must use the final boolean value before any help return; neither help nor rejection may write fixture state.
func TestReviewR7RepeatedActivation(t *testing.T) {
 source,home:=profileHome(t);pinOperatorHome(t); before:=snapshotHelpTree(t,home)
 cases:=[]struct{args []string; want int}{
 {[]string{"env","resolve","--repair","-repair=0","--takeover","--help"},2},
 {[]string{"env","resolve","--repair=false","-repair=1","--takeover","-h"},0},
 {[]string{"help","env","resolve","-repair=1","--dry-run=false","-dry-run=1","--takeover"},2},
 {[]string{"help","env","resolve","-repair=1","--dry-run=true","-dry-run=0","--takeover"},0},
 {[]string{"profile","install","--use","-use=false","--takeover","--help"},2},
 {[]string{"help","profile","install","--use=false","-use=true","--takeover"},0},
 {[]string{"--help","env","resolve","--repair=0","--takeover"},2},
 {[]string{"--help","env","resolve","-repair=1","--takeover"},0},
 }
 for _,c:=range cases {code,_,err:=runProfile(t,source,c.args...);if code!=c.want {t.Errorf("%v: exit %d want %d: %s",c.args,code,c.want,err)}}
 if !reflect.DeepEqual(before,snapshotHelpTree(t,home)){t.Fatal("fixture state changed")}
}
