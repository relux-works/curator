package main

import "testing"

// Resolve without repair must refuse takeover even when help short-circuits dispatch.
func TestReviewNativeTakeoverHelpBoundary(t *testing.T) {
 for _, extra := range [][]string{nil, {"--repair=false"}, {"--repair=0"}} {
  source, _ := profileHome(t)
  pinOperatorHome(t)
  base := append([]string{"env", "resolve", "claude_code"}, extra...)
  if code, _, stderr := runProfile(t, source, append(append([]string{}, base...), "--help")...); code != exitOK { t.Fatalf("positive help control: %d %s", code, stderr) }
  args := append(append([]string{}, base...), "--takeover", "--help")
  if code, _, stderr := runProfile(t, source, args...); code != exitUsage { t.Errorf("%v: takeover without repair admitted: exit %d, want 2; %s", args, code, stderr) }
 }
}
