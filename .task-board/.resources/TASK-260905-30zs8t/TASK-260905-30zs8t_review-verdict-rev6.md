# TASK-260905-30zs8t — review verdict, revision 6: **ACCEPT**

Change Request `CR-TASK-260905-30zs8t-6`, revision 6, element `TASK-260905-30zs8t`, integration scope
`STORY-260905-1n0iy8`. Reviewer run `RUN-260906-64c717`. Findings:
`TASK-260905-30zs8t_review-findings-stage-a-6.md`.

## What was reviewed

Not the story workspace tree — the curator branch it describes. Head **`834b40f6`**
(`feat/agent-environments-stage-a`, worktree `/Users/iv/Developer/ReluxWorks/.worktrees/
curator-stage-a-core`), twelve signed commits on curator main `a2406dfe`, every one `G` and authored
and committed by `Ivan Oparin <oparin@me.com>`. This cycle's commit is 3 files, +377/−10.

## Why `repository_delta=empty` is the right outcome here

Verified, not accepted on the producer's word:

```
git diff f39f4a9309f41a9208da817eba9129cf5a9f8dc0 c3e47989fa1a04a2e4c78172d947b4d2fbb47985  -> zero paths
git rev-parse HEAD^{tree}                                                                   -> c3e47989…
wc -c < TASK-260905-30zs8t_change-request_rev6.patch                                         -> 0
sha256 of the patch resource                                                                 -> e3b0c442…  (the empty-input digest)
```

Every brief attached to this leaf — the producer brief and all five rework briefs — places the
deliverable in the curator repository and says, verbatim, "Never write LOGBOOK.md or anything into the
control root"; rework brief 1 adds "the story workspace carries an empty delta by design". The leaf's
job was to implement stage (a) of the agent-environments protocol **in curator** against the
curator-spec authority at `f39f4a9`, and to attach the drafting, review and rework artifacts to the
board. A non-empty curator-spec delta would have meant the producer wrote into the control root, which
is the thing the briefs forbid. So an empty delta here is compliance, not absence of work — and the
work itself is real and was reviewed line by line: a twelve-commit, ~8500-line implementation across
`internal/{envprofile,pkgversion,contextpkg,contextresolve,contextlock,contextstore,contextaudit,
contextmaterialize,envmarker,interop}` and `cmd/curator`, plus the twelve outcome resources this CR
lists.

## Verdict basis

F16 — the cycle-5 major finding, a machine-scope switch overwriting a scoped adapter's surfaces while
the scope record and `profile list` kept claiming the scoped profile — is **fixed and holds under
attack**. I reproduced my predecessor's evidence with my own scripts against a binary built from this
head, and extended it to the shapes the brief named separately:

- machine `profile use`, `profile install --use`, `profile update` (both the scoped-only and the
  machine-current variants) and `profile sync` all leave a scoped adapter's home on its scoped profile,
  and `profile list` reports exactly that; `profile sync` now writes each home once, so the
  backup-generation churn is gone;
- a scoped `profile use --env` still switches only that home, and `--clear` re-materializes it from the
  machine default and removes the record;
- six SIGKILLs into a machine switch with a scope record present left the scoped home untouched, the
  current unmoved and no journal behind;
- 17 locks and 16 markers produced by these paths are schema-VALID against `context-lock-v1` and
  `agent-environment-marker-v1` at `f39f4a9`.

The producer's deletion mutant kills four named tests when I run it. A narrowing mutant of my own
survives the suite (FU-4) — so I established the property the mutant failed to establish, by driving
the four-scoped-adapter case through the CLI: the shipped code is correct, the committed fixture is one
adapter short. That is a coverage follow-up, not a defect.

The cumulative cycle-1…5 "verified, held" set was re-run at this head — F1 detector scope, F2 fresh
home, F3 interrupted switch, F8 three spellings, F9 gates with `audit.enabled: false`, F10 system-locked
migration, F12 `file://` rejection, F13 partial activation, F14 syntactic classification, F4 skip
honesty, and the 19-row CLI sweep — and every item held. All gates are green, including
`golangci-lint` 0 issues, the vector families (95 PASS / 0 FAIL / 7 classified `stage-deferred` skips),
`gate-selftest.sh` 81/0, `ledger-consistency.sh` 103 rows, the platform-case gate on linux, darwin and
windows, and the full test suite: **70 packages, 0 FAIL** (`cmd/curator` `ok 261.222s`).

AC coverage as measured: rework-5 author decision **1 of 1**; producer-brief items **8 of 8** (item 7
closes with F16); the machine-switch-over-a-scope CLI row **1 of 1**, from 0 of 1 last cycle.

## What is not accepted silently

Three follow-ups are recorded in the findings for TASK-260906-1f2ng0, with reproductions:

- **FU-3** — §9.3's "a scope record equal to the machine default is never kept" is not enforced when the
  *machine default moves onto* an already-scoped profile, and the F16 fix makes that stale record
  load-bearing (the adapter then stops following the machine default). Not blocking: the state stays
  self-consistent and accurately reported, recovery is one command, and the spec sentence is genuinely
  ambiguous about whether it binds this third path — an author call, not a defect I can prove.
- **FU-4** — the F16 mutant table is delete-only; a narrowing mutant survives because every fixture
  scopes exactly one adapter.
- **FU-5** — a machine switch in which every adapter is scoped prints nothing and exits 0.

Plus cycle 5's FU-1 (`profile_source_path_missing`/`_unreadable` diagnostic names absent, both shapes
still fail closed) and FU-2, carried forward unchanged.

None of these ships a wrong artifact, bypasses a gate, or leaves recorded state contradicting the
bytes — the bar this acceptance cycle's brief set. Stage (a) is ready to land.

## Recorded

`task-board m 'accept_cr(TASK-260905-30zs8t, revision=6,
evidence=TASK-260905-30zs8t_review-verdict-rev6.md)'` — the element routes to `integrating`; landing
belongs to a new tracked producer run, not to this review.
