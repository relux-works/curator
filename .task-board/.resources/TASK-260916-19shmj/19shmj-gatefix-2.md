# TASK-260916-19shmj — gate fix 2 (THE ONLY CURRENT INSTRUCTION, with 19shmj-sec-brief.md)

Revision 2 fixed ubuntu/macos, but Test (windows-latest) still FAILS (run 36303519222; go-test.json of test-evidence-windows-latest).
Every failure is the same: repair/materialise returns `environment_repair_failed: link skills/myskill: environment_write_would_follow_link:
…\environments\acme\<tool>\skills\myskill` — TestResolveDriftRepair, TestResolvePassthroughLiveness, TestResolveClaudeProjectEntry,
TestSharedToIsolatedRemovesStaleLink, TestCredentialLinkRegularFileRefuses, TestStaleCredentialLinkRefusals, TestDanglingPiLinkReportedDetached,
TestCredentialLinkDirectory, TestCredentialLinkUnrecordedSymlinkRefuses, TestRepairNeverRefreshesSeeds, TestClaudeSeedMergePreservesToolState,
and your own TestEnvironmentWriteNofollowVectors (repair-planted-link-replaced, repair-manager-owned-link-replaced,
materialize-recorded-file-replaced). On Windows the skill entry `skills/myskill` is a directory link (junction / reparse point), and your
nofollow check refuses REPLACING that entry as if it were writing THROUGH it. The rule (cite rc.13) forbids writing through a link; replacing
or removing the link entry itself (Lstat semantics, no traversal) must work on every platform, including junctions/reparse points; a planted
link must be replaced, never followed. Fix the Windows link classification (os.Lstat + ModeSymlink|ModeIrregular for junctions, or the
repo's existing reparse helpers — look how internal/buildcache / transaction handle reparse points) and verify with `GOOS=windows go vet`
plus reasoning; the hosted Windows lane is the proof.
Also: trunk is d41da0fb — `git fetch origin main`, combine keeping both sides; VERIFY `git diff --name-only origin/main -- . ':!.task-board'`
lists only your paths. Set status development first; append "Revision 3 — Windows link entries replaced, never followed"; handoff and WAIT
for the gate (do not interrupt it); hand off only green. No CHANGELOG/LOGBOOK edit.
