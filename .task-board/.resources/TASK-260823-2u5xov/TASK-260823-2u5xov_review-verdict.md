# TASK-260823-2u5xov review verdict — CHANGES REQUESTED (to-dev)

Reviewed delta: PR #37, merged as `a3abcf3` (tree identical to head `a0e1557`,
verified: both `f7f4588e…`). Diff base `b00836c`. 7 files, +348/-12.

The engineering is correct, and the consumption claim is not taken on trust —
it was independently reproduced and mutation-tested below. The verdict is
`to-dev` for one unmet AC clause only: **evidence routed**. No code change is
requested.

---

## 1. Consumption assertions — VERIFIED, and verified to actually bite

Ran the new `internal/scriptpolicy` cases against the real candidate root
(`relux-works/curator-spec` @ `6001dc33281b94a4ec7442ab15278550dd0f51d9`,
cloned fresh for this review, not the producer's copy). All pass:

    TestScriptHostExecutionPolicySectionsAreAllClassified   PASS
    TestScriptExecutionPolicyIdentityMatchesTheSuite        PASS
    TestScriptExecutionOptInCases                           PASS (6 subtests)
    TestARefusalPrecedesEveryWorkerSurface                  PASS (2 subtests)

A passing test proves nothing about consumption, so each assertion was mutated
against a scratch copy of the root. Every mutation fails, by the right message:

| # | Mutation of the published family | Result |
| --- | --- | --- |
| M1 | add top-level section `brand_new_section` | FAIL — "publishes sections this build does not classify" |
| M2 | delete section `audit_label_cases` | FAIL — "classifies sections the root no longer publishes" |
| M3 | `execution_policy` → `script-worker-v2` | FAIL — suite names v2, build hard-codes v1 |
| M4 | append interpreter `ruby-v1` | FAIL — published set ≠ accepted set |
| M5 | `unknown-policy`: `accepted` false→true | FAIL — accepted=false, want true |
| M6 | `schema8-explicit-opt-in`: mode enforced→declared-only | FAIL — `Enforced` = true, want false |
| M7 | `schema8-absent-policy`: accepted true→false | FAIL — accepted=true, want false |
| M8 | `opt_in_cases` emptied | FAIL — "published no opt-in cases" |

That is a real consumer: the suite's own bytes decide the outcome in both
directions, on the section set, the two closed identities, manifest acceptance
and the enforced/declared-only classification. The surface driven is production
(`skillspec.Load`, `skillspec.ScriptExecutionPolicy`, `skillspec.ScriptInterpreters`,
`scriptpolicy.Enforced`, `scriptpolicy.Admit`, `scriptpolicy.Code`), not a test
helper.

The `refusedBeforeReached` classification holds structurally, not just by
sampling: `skillspec` parses exactly one policy value, `Enforced` is
`ExecutionPolicy != ""`, and `Admit` refuses the first enforced command in
lexical order — so no enforced command of any shape can reach a worker surface
from this build. `TestARefusalPrecedesEveryWorkerSurface` pins that for every
interpreter the suite publishes.

`audit_label_cases` is classified `notImplementedYet` with owner
STORY-260822-2h0v9j. Correct call: it is real curator surface, it is named so it
cannot read as covered, and `script-command-declared-only` genuinely changes
audit decision semantics — out of scope for a coverage commit. Not a forced fit.
Declining to declare `conformance-claim-v5` is likewise right: curator publishes
no claim, and a forced consumer would be a fake gate.

## 2. Family-removal negative proof — INDEPENDENTLY REPRODUCED

Ran `suite-plan.sh` with `CI_REQUIRE_FULL_ROOT=1` against the 6001dc3 root with
one family removed at a time. Baseline (full root) exits 0; all five removals
exit 1 and name the deferred package:

| Removed from the root | suite-plan | deferred package |
| --- | ---: | --- |
| `vectors/script-host-execution-policy.json` | exit 1 | `internal/scriptpolicy` |
| `schema-cases/csk-skill-v8` | exit 1 | `internal/skillspec` |
| `schema-cases/agent-skill-v8` | exit 1 | `internal/skillspec` |
| `vectors/module-roots.json` | exit 1 | `internal/moduleroots` |
| `schema-cases/install-marker-v4` | exit 1 | `internal/marker` |

Checked the tolerance surface for a hole and found none. `root-unset` is
`deferred-only` in `skip-classes.tsv`, so with a fully serving root no tolerated
skip survives — the seven rows carrying it are required-in-fact. The one
`root-content` row (`internal/godriver :: TestModuleRootVectorsDriveTheWholeBuild`,
class `allow`) looked like a possible escape, but that case's only non-`root-unset`
skip path is file-not-exist on `vectors/module-roots.json`, which
`internal/moduleroots`' artefact row already makes fatal in the candidate lane;
an empty case list is `t.Fatal`. Not exploitable.

## 3. Candidate qualification — VERIFIED AT SOURCE

Run 32689488293, `workflow_dispatch` on `main` at `a3abcf3` (i.e. WITH this
delta). All 14 jobs SUCCESS, zero non-green. Candidate identity read out of the
job logs, not the summary:

- `CANDIDATE_REF: 6001dc33281b94a4ec7442ab15278550dd0f51d9`, accepted as
  immutable full 40-hex, checked out at `6001dc3 Band the external repository
  manager profile on marker v4`
- `manifest_sha256 sha256:803918bf8672f76cf990985e51db213b826674cd5bb54fbf47731b8404b44403`,
  digest matched the supplied expectation
- `CI_REQUIRE_FULL_ROOT: 1`; `suite-plan: served=42 deferred=0 excluded=1`, ok

All eight new ledger rows were observed by the platform-case gate on all three
runners — `ok` everywhere, except `internal/godriver ::
TestModuleRootVectorsDriveTheWholeBuild` which is `excl` on linux because the
root's own qualification vector excludes that package there (same treatment as
the pre-existing `TestProbeRejectsAnUncoveredPlatformBeforeTheWorker` row). The
gate matching every row to a live case is itself the proof the rows are
well-formed: a wrong package or case name would have failed by name.

| Runner | platform-case gate | schema-8 rows observed |
| --- | --- | --- |
| ubuntu-latest | ok | 7 ok, 1 excl (godriver, package not run on linux) |
| macos-latest | ok | 8 ok |
| windows-latest | ok | 8 ok |

## 4. Fit and documentation

The two-table split (presence via `root-artifacts.tsv`, consumption via
`platform-cases.tsv`) is the project's existing mechanism, used as designed —
nothing new was invented to make this pass. The `internal/scriptpolicy` artefact
row is correctly required: the loader `t.Fatal`s on a read error rather than
guarding with a `publishes no …` skip, so the table's own "only unguarded
packages need a row" rule applies. The README subsection states the two halves
accurately, and the schema-v6→version-agnostic prose generalisation in `ci.yml`,
`candidate-suite.sh`, `suite-plan.sh` and `root-artifacts.tsv` changes no check.

Nit (non-blocking, no rework required): this delta is the first user of the
`root-unset` skip class in `platform-cases.tsv` (0 rows before, 7 now), but the
class list in that file's own header comment still names only
`platform-control`, `host-capability`, `root-content` and `opt-in`. The class is
defined in `skip-classes.tsv`, so the gate is unaffected — it is a legibility
gap in a table whose whole purpose is legibility. Worth folding into whatever
touches that header next.

---

## Unmet: evidence routed

AC: "Consumption assertions in place …; candidate-conformance green on all three
OSes with the new coverage; **evidence routed**." The first two are verified
above. The third is not done:

- EPIC-260822-18ylpq notes carry nothing about the qualification — no run id, no
  candidate identity, no green matrix. The epic still reads as if curator
  qualification is pending.
- STORY-260822-2lvw0e checklist item 1 ("After landing: note unblock on
  skill-project-management board task TASK-260822-hje0ya, branch
  task/go-v1-switch") is unchecked, and nothing on this board references the
  auto-return pointer.
- Task checklist items 6, 12 and 13 are unchecked. `grep` over the whole board
  for `32689488293` / `6001dc3` finds the evidence only inside this task's own
  resources.

The producer run exited 124 (timeout) after the PR landed, which is where this
fell out. Nothing is wrong with the work — it just did not reach the two
consumers the task exists to unblock.

### What the next producer needs to do (board-only, no code)

1. Note the qualification on EPIC-260822-18ylpq: candidate
   `6001dc33281b94a4ec7442ab15278550dd0f51d9`, manifest `sha256:803918bf…`, run
   32689488293 green on ubuntu/macos/windows at `a3abcf3`, with the consumption
   assertions in place — i.e. step 4 of the impact-analysis landing order is
   satisfied for the module-roots family; the script-worker family is
   consumption-covered but behaviorally fail-closed
   (`script_execution_policy_unsupported`) pending STORY-260822-2h0v9j.
2. Route the unblock to the skill-project-management auto-return pointer
   TASK-260822-hje0ya (branch `task/go-v1-switch` at origin) and tick
   STORY-260822-2lvw0e checklist item 1.
3. Optional but cheap, and it closes checklist item 12 as literally worded:
   attach the 3-OS job matrix for run 32689488293 as its own task-scoped
   artifact, in the shape TASK-260822-c0rxj7 already uses
   (`*_6001dc3-green-matrix.md`), rather than leaving it as prose inside
   `_qualification-evidence.md`.
4. Tick task checklist items 6, 12, 13 and return to `to-review`.

No re-review of the code delta is needed on the next cycle — sections 1–4 above
stand. Verify the routing only.

### Reviewer-archetype note

PR #37 is already merged as `a3abcf3`; no `commit_ack` is supplied or needed
from this run. When the routing lands, the commit-owning mover makes the final
`done` transition.

## Commands run for this review

    git diff --stat b00836c..a0e1557
    git diff a0e1557 a3abcf3                       # empty — merge tree == PR head tree
    gh run view 32689488293 --json …               # 14/14 SUCCESS
    gh run view --job 97320486810|…814|…849 --log  # candidate identity + all 8 rows
    go build ./...                                 # exit 0 (after submodule init)
    CURATOR_CONFORMANCE_ROOT=<6001dc3 root> go test ./internal/scriptpolicy/... -v
    8 × mutated-root reruns (M1–M8 above)          # all FAIL as designed
    5 × CI_REQUIRE_FULL_ROOT=1 suite-plan.sh with one family removed  # all exit 1
