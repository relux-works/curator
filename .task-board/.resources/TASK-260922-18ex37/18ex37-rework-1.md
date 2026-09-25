# TASK-260922-18ex37 — rework 1 + refresh onto trunk (THE ONLY CURRENT INSTRUCTION). Unblocks curator-spec v1.0.0-rc.13.

Revision 2 was CHANGES_REQUESTED for F1 (`TASK-260922-18ex37_review-verdict-rev2*`); everything else (pin dcc7f015 everywhere, ledger via
RunOutcomes, gate green, classification) is verified — keep it.
F1: the 28 overlay rows in `.github/ci/conformance-gaps.tsv` (14 `valid-overlay-*` manager-config-v2 schema cases + 14 `schema2-overlay-*`
vectors) name owner STORY-260916-wgt8vz with a wrong reason. Those cases existed at dced9b8 and passed at rc.12; they now fail on NEW
fields: `environments: has unsupported field "source_signers"` (E1 → STORY-260916-ioemse) or `unsupported field "permissions"`
(0017/0018 → STORY-260922-1cenbr). For EVERY gap row, re-derive the owner from the actual `config.Load`/Parse error at the new pin
(first blocker, or both if both), fix owner + reason in the ledger, the results table and docs; wgt8vz keeps a row only if a path-kind
failure really owns it. Report the owner histogram before/after.
CHANGELOG POLICY (2026-09-24): do not edit CHANGELOG.md — revert this Story's CHANGELOG hunk to trunk's bytes and put the entry text
in the results resource under "## CHANGELOG entry (for release prep)".
Then refresh onto current trunk `948ae7c9` (R5 landed): combine `git diff 1511b345 948ae7c9 -- . ':!.task-board' ':!CHANGELOG.md'` into the
worktree (`git apply --3way`; keep both sides), leave nothing staged, run `task-board worktree refresh-candidate TASK-260922-18ex37`
(checkpoint replay conflicts only via its `--replay-resolutions` template). Bounded re-runs of the conformance packages + gate scripts.
Results stay a board resource (no repo files). Append "Revision 3" + `resource update`, then `task-board handoff TASK-260922-18ex37
--role developer`. A `run_wrote_outside_worktree … policy warn` block is a warning — verify status `to-review`.
