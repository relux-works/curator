# BUG-260923-krcm6m — Windows registry future-bound flake (THE ONLY CURRENT INSTRUCTION)

Control root: curator; your Story worktree only. Read `campaign-producer-rules.md`. Gate = hosted CI at handoff.
Evidence: hosted run 35901867517, `gh run download 35901867517 -n test-evidence-windows-latest` (run from your
worktree), `test/go-test.json`: `internal/registry TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew/0s/one_second_past_the_bound`
→ `registry_test.go:734: skew 0s offset 1s: refused=false want true ([])` (1.55 s). The test passes an explicit
`now` truncated to the second; find what is time- or platform-dependent (a hidden time.Now(), cache reuse across
subtests in the same TempDir, created_at rounding on Windows, …). Fix the CAUSE; never widen the bound or retry.
Keep the narrowing mutant (threshold +1 s) killed. Stress locally with -count=50. CHANGELOG. Attach results, check the
DoD item, `task-board handoff BUG-260923-krcm6m --role developer`. A `run_wrote_outside_worktree … policy warn` block is a warning.
