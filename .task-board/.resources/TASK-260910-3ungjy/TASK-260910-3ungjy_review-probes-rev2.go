package main
import (
 "os"
 "path/filepath"
 "strings"
 "testing"
 "github.com/relux-works/curator/internal/hookapproval"
)
func TestReviewRecordedMissingAndUnreadable(t *testing.T) {
 for _, scenario := range []string{"missing", "unreadable"} { t.Run(scenario, func(t *testing.T) {
  configPath, project := hookStatusProject(t)
  envPath := writeHookEnvFile(t, project, "export REVIEW=1\n")
  if code, _, stderr := capture(t, configPath, "hook", "approve", envPath); code != exitOK { t.Fatal(stderr) }
  if err := os.Remove(envPath); err != nil { t.Fatal(err) }
  if scenario == "unreadable" { if err := os.Mkdir(envPath, 0700); err != nil { t.Fatal(err) } }
  for _, args := range [][]string{{"status", "app", "--check"}, {"status", "app", "--check", "--json"}, {"env", "status", "--check"}, {"env", "status", "--check", "--json"}} {
   code, out, stderr := capture(t, configPath, args...)
   t.Logf("%v exit=%d stdout=%s stderr=%s", args, code, out, stderr)
   if !strings.Contains(out, envPath) { t.Errorf("recorded path omitted by %v", args) }
   if !strings.Contains(out, "operator") { t.Errorf("recorded approved_by omitted by %v", args) }
   if scenario == "unreadable" && !strings.Contains(out, "unreadable") { t.Errorf("unreadability hidden by %v", args) }
  }
 }) }
}
func TestReviewUnreadableStateJSON(t *testing.T) {
 configPath, project := hookStatusProject(t)
 writeHookEnvFile(t, project, "export REVIEW=1\n")
 if err := os.Mkdir(hookapproval.ApprovalsPath(filepath.Dir(configPath)),0700); err != nil { t.Fatal(err) }
 for _, args := range [][]string{{"status", "app", "--check", "--json"}, {"env", "status", "--check", "--json"}} {
  code,out,stderr := capture(t,configPath,args...)
  t.Logf("%v exit=%d stdout=%s stderr=%s",args,code,out,stderr)
  if !strings.Contains(out+stderr,"cannot read") && !strings.Contains(out+stderr,"unreadable") { t.Errorf("read error disappeared from %v",args) }
 }
}
