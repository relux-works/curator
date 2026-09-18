# Review brief — TASK-260910-2g5v17 (P4 service: per-upstream high-water for import-bundle, curator-skill-registry), review round 1 (Change Request revision 1)

You are the independent reviewer of a curator-skill-registry implementation
produced for `TASK-260910-2g5v17` (story `STORY-260910-stz5f0`, wave 4 of the 2026-09
security-audit remediation). Read, in this order:
`remediation-registry-producer-rules.md`, the producer brief
`TASK-260910-2g5v17_brief.md`, the producer results `TASK-260910-2g5v17_results.md`,
the published Change Request patch `TASK-260910-2g5v17_change-request_rev1.patch`
and its validation log, the finding in
`docs/security-audit-2026-09.md` of the repository, and `protocol/registry.md` §5 (rollback rules the import comparison mirrors) and `profiles/registry-service.md` (import/bundle sections) at the pinned `dced9b8`.

## Where the candidate is
The managed Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-skill-registry/.temp/STORY-260910-stz5f0/worktree`
on branch `task-board/story/STORY-260910-stz5f0` holds the exact candidate the
runtime published as revision 1 (the hosted gate `scripts/remote-gate.sh`
ran green on it, see the validation log). Do not edit it and leave NO files
in it (no `.review/`, no venv, no caches you did not find): read, and run
tests from a disposable copy elsewhere (e.g. under `/tmp`) with a venv
outside the tree (`python3 -m venv /tmp/csk-review-venv && pip install -e '.[dev]'`
from the copy).

## What to verify
1. **Behaviour** (settled decisions in the brief): per upstream `key_id` the store persists the highest accepted upstream boundary (version, log_size, head, merkle_root) transactionally in `registry.db`; `import-bundle` refuses version-below and same-version-different-body bundles with closed diagnostics naming the `key_id` and both boundaries, an identical re-import is a no-op, a newer bundle imports and advances the high-water in the SAME transaction as the records (never before, never partially); `--accept-older-upstream` imports an older bundle with a warning WITHOUT lowering the high-water and the inconsistent case is never overridable; a failed import (chain break) leaves the high-water untouched; first import of an unknown upstream establishes it. Quote file:line; check the diagnostic spellings against the profile (reused if one exists there).
2. **Transactionality probe**: from the disposable copy, inject a failure after the records are written but before the high-water update inside the same transaction (or read the code path) and show both roll back together; confirm the persisted state is protected like the rest of the database (permissions) and included/excluded from `backup`/`verify-backup` as the results state.
3. **Tests and docs**: the seven scenarios of the brief have committed tests; README import section, SECURITY.md paragraph, CHANGELOG "P4" entry.
N. **Independent validation**: `python -m pytest -q` (with
   `CURATOR_CONFORMANCE_ROOT` set to the pinned `dced9b8` root — the
   curator-spec checkout is at a later main, so use a detached worktree
   at `dced9b8317e0e8af79edf2d0539b32bd22b6c85b` under `/tmp`) and
   `python -m mypy` from the disposable copy (`set -o pipefail`, exit
   codes, interpreter); at least two narrowing mutants caught by committed
   tests; scope and hygiene (brief's out-of-scope list respected, CHANGELOG
   entry present, nothing left in the worktree).

## Verdict
Record `TASK-260910-2g5v17_review-verdict-rev1.md` (task outcome) with the
per-item table, transcripts, mutants and findings; then exactly one of
`task-board m 'accept_cr(TASK-260910-2g5v17, revision=1, evidence=TASK-260910-2g5v17_review-verdict-rev1.md)'`
or a changes-requested verdict routed with `set_status(TASK-260910-2g5v17, status=to-dev)`
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.
