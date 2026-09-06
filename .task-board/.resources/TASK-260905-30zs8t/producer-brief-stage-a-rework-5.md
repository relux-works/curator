# Producer brief: stage (a) core — rework 5 (F16, final cycle)

Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`, branch
`feat/agent-environments-stage-a`, head `b6f00e1a`. Findings:
`TASK-260905-30zs8t_review-findings-stage-a-5.md` — F13/F14/F15 verified fixed and held; one **major**
remains. New signed commits on top of `b6f00e1a`, no rewrite.

## Author decision — F16
A machine-scope switch currently overwrites a scoped adapter's surfaces while the scope record and
`profile list` keep claiming the scoped profile: the recorded state and the bytes disagree, which is
the same class cycle 4 ruled blocking as F13.

Take **option (a)**: the machine pass **skips adapters that carry a scope record**. `materializeScope`
consults `ScopedCurrents(home)` for the unnarrowed case, the way `SyncWithPolicy` already does in two
passes. Rationale: a scope record means "this adapter follows its scoped profile, not the machine
current" (§9.3), so the machine pass has no business writing that home; option (b) would write the same
home twice in one operation for no gain. The manager must never report `env:<id>=<profile>` for a home
it has just overwritten with a different profile.

Regression cover, driving the CLI and `Use`/`Install`: with `env:codex_cli` scoped to profile B and the
machine current on profile A, a machine-scope `profile use C` leaves the codex home on B, materializes
the other adapters to C, and `profile list` reports exactly that; the same through
`profile install <src> --use`; and a scoped `profile use --env codex_cli D` still switches only that
home. Add the narrowing mutant (machine pass ignores the scope record) and name the test it kills.

Also apply the reviewer's recorded follow-ups from this cycle where they are one-line corrections
inside the code you touch; anything larger goes into the report as a named follow-up rather than into
this commit.

## Gates and delivery
The full set as before, including `bash .github/ci/gate-selftest.sh`, the platform-case gate for the
three GOOS values, the vector families through
`CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1`, and
`go test -count=1 -timeout 30m ./cmd/curator` at the end. Signed commits; do not push, tag, or open a
PR. Attach `TASK-260905-30zs8t_rework-report-5.md`;
`task-board handoff TASK-260905-30zs8t --role developer`. Never write LOGBOOK.md or anything into the
control root.
