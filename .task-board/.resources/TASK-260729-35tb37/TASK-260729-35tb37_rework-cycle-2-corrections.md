# TASK-260729-35tb37 rework cycle 2 — corrections record

**Board task:** `TASK-260729-35tb37` (`refresh-csk-baseline-and-file-map`)
**Cycle date:** 2026-07-29
**Answering:** `TASK-260729-35tb37_review-verdict-cycle-2.md` (CHANGES REQUESTED)
**Revised artifact:** `TASK-260729-35tb37_cocoaskills-baseline-file-map.md`, now revision 3
**Revision 3 SHA-256:** `929098bec7303b5b7545e4ad469c47aa9afe484776f8d6f0cb870dd1b0b41b57` (846 lines)

## 1. What the reviewer accepted without change

The cycle-2 verdict independently supported every technical claim in revision 2
and required no correction to any of it:

- CocoaSkills provenance: clean local `main` at `edce8816…`, remote
  `origin/main` at `6fc2fd97…`, divergence `0 2`, no pull/fetch/checkout;
- the 19 distinct paths / 20 commit-level touch events upstream delta;
- the packaging, CLI, environment, PATH, and CI map against `6fc2fd97…`;
- both root file/function plans and their narrow pytest/mypy gates, including
  the added rc.3-rooted `test_skill_manifest_resolution_vectors` replay;
- the §2.3 historical regression evidence — rc.2 `98 passed` (exit 0) versus
  rc.5 `1 failed, 97 passed` (exit 1), the `scripts/golden-tool` cause, the
  semantic manifest equivalence, and the `deb971f` / `6fc2fd9` framing as a
  regression gate rather than product work or pin authorization;
- the local rc.2 versus upstream rc.3 conformance-pin boundary;
- the rc.5 candidate state and publication boundary;
- all other board-state rows, dependency edges, and the two stale-diagram
  findings.

None of that was edited in this cycle.

## 2. The single required correction

`TASK-260729-v5hqnv` completed its second independent review **after** revision 2
was written and is now reviewer-accepted `done`. Revision 2 therefore carried a
stale current-state claim in five places. All five are corrected:

| Location | Was | Now |
| --- | --- | --- |
| §1 executive finding 3 | `to-review`, "not accepted `done`" | `done`; seven brief-field retargets and two provenance dependency edges stated as reviewer-accepted |
| §6.2 drift table row | current state `to-review`; effect "review acceptance still pending" | current state `done`; effect "reviewer-accepted; brief wording and the two provenance edges are now binding" |
| §6.2 subsection | "is now `to-review`, awaiting a fresh reviewer"; bullet asserting the wording is "not yet reviewer-accepted" and may still change | `done` with the cycle-2 verdict cited; pending-review bullet removed and replaced with the accepted net-delta confirmation |
| §8 recommendation item 7 | "complete the pending `TASK-260729-v5hqnv` review cycle before treating any retargeted brief text as accepted" | pending-review clause removed; retains the board-`done`-is-not-an-upstream-ref/CI-pin distinction and applies it to both `3nx97g` and `v5hqnv` |
| §9 references | cycle-2 verdict absent | `TASK-260729-v5hqnv_review-verdict-cycle-2.md` cited, plus this task's own cycle-2 verdict |

## 3. Boundaries explicitly preserved

The verdict required these to survive the correction. Each is still stated in
revision 3:

- **No product change.** `TASK-260729-v5hqnv` touches no CocoaSkills or Curator
  source, test, pin, or dependency beyond the two board edges; acceptance does
  not change that. Revision 3 says so verbatim in §1 finding 3 and §6.2.
- **No tests run.** The retarget asserts no test result, and the cycle-2 review
  ran none. The §2.3 rc.2/rc.5 numbers remain attributed to
  `TASK-260729-1b9tc3` logs alone. This reconnaissance ran no test in any
  revision (§7.2 closing paragraph).
- **No root unblock.** `TASK-260720-z9j4c9` and `TASK-260720-z2z795` remain
  `backlog`, each `blockedBy = [TASK-260720-1pvfj5]`, which is itself `backlog`
  behind `done` `2qqq0w` and `development` `jrrgw9`. Acceptance of the retarget
  confers no start clearance.
- **Fail-closed `TASK-260720-3ag6pi`.** `TASK-260720-12r55p` still carries the
  hard edge to `3ag6pi`, which remains `blocked` and literal rc.4 with no rc.5
  replacement gate on the board.
- **Historical rc.2/rc.5 result unrewritten.** The red rc.5 run remains
  attributed as evidence against the stale local base, not as a green current
  product claim. §2.3 is byte-unchanged.

## 4. Re-queried live facts

All commands ran as standalone processes; real exit codes recorded.

| Check | Exit | Result |
| --- | ---: | --- |
| `task-board m 'set_status(TASK-260729-35tb37, status=analysis)'` | 0 | cycle re-entered `analysis` |
| `task-board q 'get(TASK-260729-v5hqnv) { id name status }'` | 0 | `retarget-csk-go-briefs-to-rc5`, **`done`** |
| `task-board q 'get(TASK-260720-z9j4c9) { id status blockedBy }'` | 0 | `backlog`, `[TASK-260720-1pvfj5]` |
| `task-board q 'get(TASK-260720-z2z795) { id status blockedBy }'` | 0 | `backlog`, `[TASK-260720-1pvfj5]` |
| `task-board q 'get(TASK-260720-1pvfj5) { id status blockedBy }'` | 0 | `backlog`, `[TASK-260720-2qqq0w, TASK-260720-jrrgw9]` |
| `task-board q 'get(TASK-260720-2dnqw2) { id status blockedBy }'` | 0 | `backlog`, `[TASK-260720-3c0ss2, TASK-260720-3j8pp5, TASK-260729-3nx97g]` |
| `task-board q 'get(TASK-260720-12r55p) { id status blockedBy }'` | 0 | `backlog`, `[TASK-260720-th0jdi, TASK-260720-3ag6pi, TASK-260729-3nx97g]` |
| scoped status projections for `2qqq0w`, `jrrgw9`, `3ag6pi`, `3nx97g`, `1b9tc3`, `2kaopg`, `3jku56`, `1nlmvv` | 0 each | `done`, `development`, `blocked`, `done`, `done`, `done`, `done`, `done` |
| `git status --porcelain=v2 --branch` in `/Users/iv/Developer/Wildberries/cocoaskills` | 0 | clean `main` at `edce8816…`, upstream `origin/main`, `+0 -2` |
| `shasum -a 256` revision 3 | 0 | `929098be…` |

One board query (`get([...])` batch form) returned exit 1 with
`element … not found`; the DSL has no multi-ID `get`, so the same facts were
re-read as individual scoped projections above. That failure is reported here
as a real non-zero exit, not suppressed.

## 5. What was not touched

No repository, task brief, dependency edge, diagram, pin, protocol artifact,
checkout, or historical rc.2/rc.5 test result was altered. No task other than
this one's outcome resources, checklist, notes, and status was mutated. No pull,
fetch into refs, checkout, install, or test suite was run.
