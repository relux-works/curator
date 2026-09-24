# BUG-260924-5p8b0z — reject hard-linked System32 exec when SystemRoot is not manager-captured (THE ONLY CURRENT INSTRUCTION)

Control root: curator; your Story worktree only. Read `campaign-producer-rules.md` (no CHANGELOG edits — entry text in results; results are
board resources; artifacts in $TMPDIR). Landing gate = hosted CI at handoff.
Source of truth: the accepted curator-spec erratum TASK-260924-mcmova (commit dcc7f015 on the PR #88 branch; core.md exec-capability bullet +
`script-host-execution-policy.json` `executable_identity_cases`, 8 cases: 1 accept, 7 rejections). Evidence: hosted windows run
35977701956 → `internal/scriptworker TestExecutableIdentityCasesAtProductionEntry` case `windows-exec-uncaptured-systemroot-hardlinks`:
production resolver accepted = true, want false (reason systemroot-not-manager-captured). Fix `internal/scriptworker/exec.go` (the R5 allowance
~66-95, 229-285) so the allowance requires System32 derived from the MANAGER-CAPTURED SYSTEMROOT, exactly the erratum's conjunctive
conditions; drive ALL 8 executable_identity_cases at the production entry (copy the vector bytes from dcc7f015 into fixtures with sha256 if
the curator pin does not yet contain them). Narrowing mutant (drop the captured-SystemRoot condition) killed; cross-platform row via the
existing GOOS/env seam so it is provable off-Windows. Bounded runs; attach results; check DoD; `task-board handoff BUG-260924-5p8b0z --role
developer`. A `run_wrote_outside_worktree … policy warn` block is a warning.
