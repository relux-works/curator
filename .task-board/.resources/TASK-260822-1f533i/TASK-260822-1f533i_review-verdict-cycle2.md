# TASK-260822-1f533i — review verdict cycle 2: ACCEPTED

**Run:** RUN-260822-dd89f5, reviewer archetype, read-only, not goal-bound.
**Subject:** `spec/sw-core-prose` @ `78d544d` = `origin/spec/sw-core-prose`, two commits on
`origin/main` `3dc9ca6`. Whole-branch diff: `protocol/core.md` only, +407/-2.
**Answers:** `TASK-260822-1f533i_review-verdict.md` (RUN-260822-ecf868, *changes requested* on
`41cf556`) and `TASK-260822-1f533i_rework-78d544d.md`.
**Siblings at review time:** `origin/spec/sw-manager-security` = `c2371d3`,
`origin/spec/sw-schema` = `ebfed81`.

## Provenance

`78d544d` signature Good, ECDSA `SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`,
`oparin@me.com`. No AI attribution anywhere in `3dc9ca6..78d544d`. Pushed, no PR, as instructed.
Working tree clean except untracked `tools/__pycache__/` (the known no-`.gitignore` trap; nothing
staged).

## Cycle-1 items — all four verified fixed in the file, not taken from the notes

**1 (blocking) — Linux `active-process-count-limit` mandated a false `applied` claim. FIXED.**
`core.md:432` now reads `host-conditional: delegated cgroup v2 pids.max`. This is the recommended
fix verbatim and it closes the defect completely: with `host-conditional`, the section's own
"an `available` control MUST report `applied`" no longer forces a Linux manager to assert a
private aggregate descendant bound, and "a `host-conditional` control that probes unavailable MUST
NOT reject the invocation" (`core.md:472-473`) keeps a non-delegated host running. The row now
matches `aggregate-memory-limit` directly beneath it and stops contradicting its own macOS cell and
the deferred `script-hard-aggregate-descendant-resource-bounds`.

The added normative paragraph (`core.md:446-453`) goes past the literal ask and is the right call:
it states that `RLIMIT_NPROC`/`RLIMIT_AS` are not private aggregate domains and that a manager MUST
NOT back either aggregate control with them on any platform. That converts three previously
asserted verdicts (macOS `unavailable` on both aggregate rows, Linux `host-conditional`) into a
stated argument, which is what stops the next editor reverting the cell.

The record the cycle-1 verdict actually asked for exists:
`TASK-260822-1f533i_results.md` cites `TASK-260822-1l4r4f_analysis.md:560` by line as the source of
the wrong cell and lists the deviation as a recorded divergence, not a silent correction.

**2 — "complete diagnostic set" overclaim. FIXED.** `core.md:558` now reads "The policy-level
diagnostic set of this policy is". Consistent with `profiles/manager.md` @ `c2371d3`, which
pre-reconciles with "The first four are the policy-level set of Protocol Core section 4.1.1. The
last three are worker-session codes of this profile". The merge contradiction is gone. Four-row
table unchanged.

**3 — manifest fields never named. FIXED.** `execution_policy` named in the opening paragraph with
its single admitted value (`core.md:210-212`), `interpreter` named in the interpreter paragraph
(`core.md:261-262`), plus a co-requirement paragraph (`core.md:216-219`) making one-without-the-other
an invalid manifest with no default and no declared-only install. Verified against the schema branch
rather than assumed: `schemas/v1/common.schema.json` @ `ebfed81` lines 160-167 carry
`execution_policy` -> `scriptExecutionPolicyV1`, `interpreter` -> `scriptInterpreterV1`, and
`dependentRequired` in both directions. Prose and schema agree exactly.

**4 (nit) — wrong prefix on the non-aliasing sentence. FIXED.** `core.md:455-461` now names
`script-total-network-denial` as this policy's deferred guarantee and `total-network-denial` as
section 4.2.1's build-policy guarantee that may not appear on a script surface at all. Consistent
with the deferred-guarantee list at `core.md:545-552`.

## Cross-branch coherence re-verified at current heads

Extracted every `script_execution_*`, `script-command-*`, `script-*-v1` and deferred-guarantee
identifier from `profiles/manager.md` @ `c2371d3` and compared with `core.md` @ `78d544d`: the seven
deferred guarantees, the two warning classes (`script-command-declared-only`,
`script-command-unfiltered-declared-network`), `script-capability-evidence-v1` and
`script-worker-v1` are identical. manager.md's seven diagnostics are core.md's four plus three
worker-session codes, exactly as manager.md states. The stale identifier list in the `3fkfmf`
coordination note describes neither file.

The `first five rows carry the same macOS and Windows verdicts` claim is accurate: 4.2.1's
`rc5-native-control-inventory-v1` table (`core.md:801-807`) is macOS/Windows only and its five rows
match cell for cell.

## Gates — all eight re-run by the reviewer at `78d544d`, real exit codes

Logs in `curator/.temp/TASK-260822-1f533i/review2/`. Python through the task venv
(`tools/validate.py` needs `jsonschema==4.25.1`).

| Gate | Exit | Detail |
|---|---|---|
| `python tools/validate.py` | 0 | 49 schemas, 471 vector files |
| `python -m unittest discover -s tools` | 0 | 91 tests, OK |
| `go test ./tools/...` | 0 | ok generate-vectors |
| `go run ./tools/generate-vectors -root .` | 0 | — |
| `git diff --exit-code` over `conformance/v1` + `release/` | 0 | regeneration clean |
| `git diff --check` | 0 | formatting job |
| `lychee --offline '**/*.md'` | 0 | 36 OK, 0 errors, 6 excluded |
| `gofmt -l tools` | 0 | empty |

Gate set matches `.github/workflows/ci.yml` (validate.py, gofmt, `git diff --check`, lychee-action).
AC clause "formatting and link gates pass" is confirmed independently, not from the notes.

## AC and DoD

- Subsection committed to the story branch — yes, `spec/sw-core-prose` @ `78d544d`, pushed.
- Formatting and link gates pass — yes, re-run above.
- Structure and honesty mirror 4.2.1 — yes. Mechanism and kernel guarantee named separately in the
  seven-row table and in every derivation bullet; the three divergences from 4.2.1 written as
  positive controls; the seven deferred script guarantees disjoint from 4.2.1's six in both
  directions; the `4.3` edit scopes "audit surface, not a runtime sandbox" to commands that have
  not opted in instead of leaving it false. No false hardened claim survives — the one that existed
  is the cell fixed in item 1.
- Follows the `1l4r4f` analysis or records an explicit justified deviation — yes, the one deviation
  is now recorded with the analysis line cited.
- Outcome resources — `TASK-260822-1f533i_rework-78d544d.md` is new for this cycle;
  `TASK-260822-1f533i_results.md` updated. Logbook entry 2026-08-22 2115 present.

## Non-blocking, recorded not charged

**N1 — `per-file-size-limit` Windows reason is now locally self-refuting.** `core.md:434` keeps
`unavailable: no-private-aggregate-domain` for that cell while the new paragraph at `core.md:453`
says the control "bounds one file write, not an aggregate over a domain". The reason term therefore
does not describe why the control is missing on Windows. This is inherited verbatim from 4.2.1
(`core.md:806`), the closed five-entry reason vocabulary offers nothing better, and the same cell is
frozen in the rc5 vectors — so fixing it here would diverge one of several copies for a wording
defect that produces no false `applied` claim. Belongs to a future inventory revision across 4.2.1,
4.1.1 and both vector files, not to this branch. Do NOT fix unilaterally.

**N2 — schema-8 reachability from section 4 is owned, contrary to the cycle-1 note.** 4.1.1 says
"manifest schema 8 or later" while `core.md:156-157` still says v1 through v7, the Added-behavior
table stops at row 7, `core.md:185-186` says "schemas 2 through 7 reject unknown fields", and
section 10 still names marker schemas 1-3. The cycle-1 verdict guessed `TASK-260822-1mwy10`; that
task is now `done` and its accepted review confirmed the deltas are still owed and named this file.
The real owner is `TASK-260822-c0rxj7`, whose notes already carry the four required core.md edits
(a)-(d) as landing input from `1mwy10`. Not chargeable to this task's AC, and not a merge risk that
this branch can close: a single branch cannot renumber the schema series without the schema branch.
Reviewer added a checklist row on `c0rxj7` so it is a tracked row rather than a paragraph in notes.

**N3 — lockstep still open, already tracked.** `profiles/manager.md:710` @ `c2371d3` still reads
`available: RLIMIT_NPROC`. `TASK-260822-3fkfmf` is in `development` with a rework live; checklist
rows exist on `c0rxj7` and `f4qv7w`. Correctly handed off, nothing owed by this branch.

**N4 — forward reference is intentional.** 4.1.1 cites
`conformance/v1/vectors/script-host-execution-policy.json` (`TASK-260822-f4qv7w`, not yet written),
mirroring how 4.2.1 cites `go-host-execution-policy.json`. Inline code, not a markdown link, so the
link gate is not fooled — it genuinely has nothing to resolve. This branch must land together with
the vectors branch, which is `c0rxj7`'s job.

## Verdict

**ACCEPTED.** Every cycle-1 finding is fixed in the file, the blocking one exactly as recommended
plus a justified normative addition; the divergence record the cycle-1 verdict said was missing now
exists and cites its source by line; all eight gates re-run green at `78d544d`; the sibling branches
agree on every identifier. Scope is already committed, signed and pushed at `78d544d`, so this
reviewer supplies no `commit_ack`.
