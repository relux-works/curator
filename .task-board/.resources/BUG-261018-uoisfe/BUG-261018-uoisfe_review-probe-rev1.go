package main

import "testing"

// Unrelated built-in list/update must reject takeover; the same commands without it remain valid.
func TestReviewNativeSkillsUnrelatedTakeover(t *testing.T) {
 for _, command := range []string{"list", "update"} {
  t.Run(command, func(t *testing.T) {
   source, _ := profileHome(t)
   pinOperatorHome(t)
   if code, _, stderr := runProfile(t, source, command); code != exitOK { t.Fatalf("positive control = %d: %s", code, stderr) }
   if code, _, stderr := runProfile(t, source, command, "--takeover"); code != exitUsage { t.Fatalf("unrelated takeover admitted: exit %d, want %d; %s", code, exitUsage, stderr) }
  })
 }
}
