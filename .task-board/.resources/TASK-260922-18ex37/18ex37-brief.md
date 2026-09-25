# TASK-260922-18ex37 — move SPEC_PIN to current curator-spec + fill the gap ledger (THE ONLY CURRENT INSTRUCTION)

Control root: curator; your Story worktree (STORY-260922-2goxjs; it carries the checkpointed gap ledger TASK-260922-3bbvrs).
Read `campaign-producer-rules.md` and this task's description (exact). Landing gate = hosted CI at handoff.

## Target pin (orchestrator decision — do not pick another)
`dcc7f015e2d97edf2d52928afb6fd79ec8129e8b` in relux-works/curator-spec: current spec main (`3d2c611`, incl. draft decisions 0019/0021) + the accepted erratum
TASK-260924-mcmova (PR #88, "hard-link substitution" + 8 `executable_identity_cases` in script-host-execution-policy). That
commit is NOT on spec main yet: spec main will be fast-forwarded to it right after this leaf lands (its Implementations
check needs curator to classify the new sections first). Fetch it from the PR branch `land/TASK-260924-mcmova`. Also update
the CI's pin mechanics exactly as the previous pin promotion (48da2690, rc.12) did — find every place SPEC_PIN / the
conformance root revision is pinned (ci.yml, scripts, tests, docs) and move them together.

## Work
1. Run the whole suite against the new root (bounded per-package calls). Every NEW failure is either (a) implemented now
   when it is a pure classification/bookkeeping change (e.g. `scriptHostExecutionPolicySections` must classify the new
   sections — `executable_identity_cases` and its definition — and any other section/family the new root adds; classify
   = consume it or record why unreachable), or (b) a ROW in `.github/ci/conformance-gaps.tsv` attributed to the owning
   board element per this task's description, verifying each mapping against the spec commit that published the case.
   Never weaken an assertion; never skip a driven row.
2. Coordinate with in-flight work that assumed the old pin: `TASK-260923-em42lw` (fragment v2) is parked blocked_by this
   task — list in results which of its failures the new pin removes. R5 (`TASK-260916-2ok97n`) is landing first; do not
   touch scriptworker behaviour — only classification.
3. Results: table of every new-root failure → classified | gap row (owner) | fixed; tally before/after; the exact pin lines.
Bounded runs; gate scripts (ledger-consistency, gate-selftest, no-broad-suppression). CHANGELOG. Attach results, `resource
update` them, check DoD, `task-board handoff TASK-260922-18ex37 --role developer`. A `run_wrote_outside_worktree … policy warn`
block is a warning — verify status `to-review`.
