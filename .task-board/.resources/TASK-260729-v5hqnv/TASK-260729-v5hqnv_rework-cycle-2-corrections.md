# TASK-260729-v5hqnv — rework cycle 2 corrections

Answers `TASK-260729-v5hqnv_review-verdict-cycle-1.md`. Board-only. No product repository, spec, pin, git or test operation.

## Exact board delta of this cycle

Two fields, both on `TASK-260720-12r55p`. Verified by diffing a ten-field projection (`id`, `name`, `title`, `status`, `parent`, `description`, `scope`, `ac`, `blockedBy`, `blocks`, `notes`, `checklist`) of all seven briefs taken immediately before and immediately after this cycle.

| Task | Field | Before | After | Operation |
| --- | --- | ---: | ---: | --- |
| `TASK-260720-12r55p` | `notes` | 1727 chars | 515 chars | `set_notes(..., set=true)` — restore |
| `TASK-260720-12r55p` | `scope` | 1383 chars | 1979 chars | `set_details(..., scope=...)` — append retained-gate clause |

Nothing else changed on any of the seven. No status, title, name, parent, checklist, resource, or dependency edge was touched in this cycle. No element was created or deleted.

## Correction 1 — stale receipt-hash claim

**Was (false):** audit §5 — "`750f5f75…` (rc.4 receipt) | removed from the board entirely".

**Now (verified):** the value is absent from the current `description`, `scope` and `ac` of all seven retargeted briefs — substring scan over all 21 fields, **0 hits**. It is deliberately not removed from the board: `task-board grep 750f5f75` returns **29** lines across historical outcome resources and progress records (`TASK-260720-12iigs`, `TASK-260720-1nvomm`, `TASK-260720-1s1vr6`, `TASK-260720-3mrm4z`, `TASK-260720-poa3ze`, `TASK-260720-cw39jh`, `TASK-260729-1t1z2l`, `TASK-260729-3nx97g`, and this task's own records). That historical evidence records the genuine rc.4 identity and is preserved untouched.

## Correction 2 — `TASK-260720-12r55p.notes` restored byte-for-byte

`notes` is outside this task's authorized scope (`description`, `scope`, `ac`, dependency edges). Cycle 1 appended a 1211-character rc.5 paragraph to it. Reverted.

Pre-retarget content reconstructed from two independent authoritative sources that agree byte-for-byte — not hand-reconstructed:

| Source | What it is | Bytes | SHA-256 |
| --- | --- | ---: | --- |
| `STORY-260720-1uv5gi_spawn-log_-analyst--solution-architect--codex-.log:11337` | raw `set_notes` mutation result that originally wrote these notes on 2026-07-20 | 515 | `3a04adc70cd2e5499608172203a7d3ed17b4ec33e281c4c7ca7b51f24ace2500` |
| `TASK-260729-1kq1rd_spawn-log_-analyst--researcher--codex-_RUN-260728-b72cf7.log:4720` | read-only compact board projection taken 2026-07-28, before this task ran | 515 | `3a04adc70cd2e5499608172203a7d3ed17b4ec33e281c4c7ca7b51f24ace2500` |

Live board field after restoration: **515 bytes**, SHA-256 `3a04adc70cd2e5499608172203a7d3ed17b4ec33e281c4c7ca7b51f24ace2500` — exact match to both sources.

Consistency check: the pre-restoration live value was exactly `<515 pre-retarget bytes>` + `\n` + the 1211-character appended paragraph, confirming the cycle-1 mutation was a pure append that disturbed nothing else in the field.

Other six briefs checked the same way against the 2026-07-28 pre-retarget projection: `3pemm6` (539 bytes) and `3s27te` (295 bytes) unchanged and matching; `2dnqw2`, `2g21eg`, `akf5kh`, `th0jdi` have empty `notes` before and after. `TASK-260720-12r55p` was the only brief whose `notes` this task ever mutated.

## The fail-closed gate explanation moved into `scope`

The substantive prerequisite had to survive the revert, so it was rewritten into the authorized `scope` field of `TASK-260720-12r55p` (596 characters appended, 1383 → 1979). Exact added text:

> Retained fail-closed prerequisite: the hard blockedBy edge to TASK-260720-3ag6pi is deliberately kept. TASK-260720-3ag6pi is protocol-v6-conformance-verification, is still status blocked, and is still scoped to the literal rc.4 line, and no rc.5 replacement verification gate exists on the board today, so no relink target was invented. Before this task starts, the Curator-side owner must either retarget TASK-260720-3ag6pi to this rc.5 candidate root and re-review it, or create the rc.5 verification gate and relink this edge to it; resolving that gate is Curator-side work outside this task.

The `TASK-260720-jrrgw9` observation was deliberately **not** copied into any brief field. It is a Curator-side gate that does not govern `12r55p`'s own start condition, so putting it in a CocoaSkills brief would be scope creep. It stays in audit §4.

This introduces the only literal `rc.4` string remaining in the 21 brief fields. It is intentional and describes the state of the still-rc.4 `3ag6pi` gate; it does not retarget `12r55p` to rc.4. Recorded explicitly in the revised audit §5.

## Evidence regenerated

- `TASK-260729-v5hqnv_after.jsonl` — regenerated from the live board; nine original fields plus `notes`, so the mutation inventory is complete. The nine shared fields still diff directly against `before.jsonl`.
- `TASK-260729-v5hqnv_before.jsonl` — left byte-identical. Review cycle 1 independently verified it; retroactively editing verified evidence would be worse than documenting the missing `notes` dimension here and in audit §7.2.
- `TASK-260729-v5hqnv_rc5-brief-retarget-audit.md` — revised: corrected §5 stale-hash row, corrected §5 residual-wording row, corrected header scope line, complete net mutation inventory table, updated §2.3 / §2.8 / §4.1, new §7 rework record.

## Checklist item 16 "Tests green" — no test was run and none is claimed

Read this before trusting the checkmark. Item 16 is a generic role-template item that does not apply to this task:

- This task changes **no code, test, doc or config file**. Its entire delta is board metadata, applied through `task-board m`.
- Both the task scope ("No CocoaSkills product or test code") and `TASK-260729-v5hqnv_review-instructions.md` ("do not ... run tests") forbid running tests.

Item 16 is therefore **vacuously satisfied** — there is no test-bearing change whose suite could be green or red. It is checked only because `task-board handoff` refuses the role transition while any checklist item is unchecked. **No test suite was executed at any point in either cycle, and no test result is asserted anywhere in this task's evidence.**

The related upstream fact is recorded elsewhere and is not this task's claim: `TASK-260729-1b9tc3` found CocoaSkills `tests/test_protocol_conformance.py` at 1 failed / 97 passed against the rc.5 root, caused by the `csk-skill.json` → `agent-skill.json` fixture rename. That is a product defect owned by `TASK-260720-z9j4c9`, not something this board-metadata task could or did fix.

## Re-verified after the corrections

- Seven direct board projections match `after.jsonl` exactly.
- `750f5f75` — 0 hits across all 21 brief fields.
- `3fcd714a` / `rc4` — only as the required negative-case identifier `legacy_rc4_without_execution_policy`.
- `claim v2` / `conformance-claim-v2` — 0 hits.
- rc.5 root `manifest.json` still hashes to `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`; conformance root holds 448 files (447 manifest-listed plus `manifest.json`).
- `plan(STORY-260720-1uv5gi, mode=related, active=true)` resolves over 83 elements; critical path shape unchanged through `th0jdi → 12r55p → 3pemm6 → 3s27te`.
- Curator checkout: no tracked modification, nothing staged. No commit, tag, pin or publication. No tests run, as the reviewer instruction requires.
