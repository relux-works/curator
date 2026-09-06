# Review brief: path overlay declarability, cycle 5 (narrow, landing decision)

## Subject

- Branch `feat/path-overlay-declarable` at `407424e`, four signed commits past curator-spec main.
  PR https://github.com/relux-works/curator-spec/pull/47, all eight checks green. **The only delta
  under review is `git diff 18dca85..407424e`** — three files.
- Read `TASK-260906-2x4s7i_review-findings-4.md` (cycle 4: F16 major, F17 and F18 minor) and the PR
  description at head.
- Authority: `protocol/environments.md` §1, §6, §12.1; `protocol/core.md` §6.1;
  `schemas/v1/manager-config-v2.schema.json`; `CHANGELOG.md`.

## Scope, deliberately narrow

Four cycles have now attacked this schema. Cycle 4 found **no behavioural defect**: F16 was the
in-repo changelog still describing the cycle-1 behaviour, F17 three claims in the PR description that
did not reproduce, F18 a coverage regression with behaviour unchanged. All three repairs were made by
the **orchestrator**, who also wrote this brief — review them accordingly.

Do not re-derive the earlier cycles. Attack this delta and only what it touches, then answer the
landing question.

## Review dimensions

1. **F16 — the changelog now.** Read the entry at head against the committed schema, clause by clause.
   It must describe what the schema does, not what an earlier head did: the allowed scheme set, the
   refusal of an SCP-shaped spelling whose host is outside core §6.1, and the `file:` URL refused as
   neither §1 kind. Check every classification claim in it resolves against a published case. This is
   the third document in this change to have carried a stale claim; assume the fourth is possible.

2. **F18 — the bare drive letter.** `C:` is now pinned as **refused**, on the reasoning that §1 wants
   an absolute or project-relative directory and a bare drive letter is neither, while as a network
   form it is a host with an empty path that core §6.1 rejects. Verify that reasoning holds under both
   readings rather than only one, and that the case kills `M-drive-wide` (widening the drive pattern
   to `^[A-Za-z]:`). Then ask whether `C:` refused is the behaviour an operator on Windows should get,
   or whether it is merely the behaviour that is easy to pin — if the latter, say so.

3. **F17 — the description.** Every number and claim in the PR body must reproduce at head: the case
   count, the vector count, the enumerated spellings, the mutant list, and the bounds paragraph, which
   now names `branch` being refused by `additionalProperties` and the three cycle-3 survivors. Check
   nothing is overstated and nothing that belongs in the bounds is missing.

4. **Regression surface of the delta.** The generator gained one case and the changelog one paragraph.
   Confirm the regenerated artefacts are the generator's own output, that `git ls-files` still carries
   no build artefact, and that no other case's verdict moved. Re-run `make validate` and
   `make regenerate-check` yourself.

5. **The landing question.** Say plainly whether PR #47 is safe to land. If it is not, name the one
   thing that must change; if it is, say so without hedging — four cycles of "not yet" on a change
   whose behaviour is now settled would itself be a finding about the review loop.

## Constraints

Read-only outside your scratch. Never write into the control root, and do not push or modify the
branch.

## Verdict contract

Attach `TASK-260906-2x4s7i_review-findings-5.md` with a `repeat-of:` line naming any class that
recurs. Blocking or major → set the task to `development`. Otherwise an explicit ACCEPT at
`to-review`. Do not mark the task done. Then `task-board handoff TASK-260906-2x4s7i --role reviewer`.
