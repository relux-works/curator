# Review brief — TASK-260910-2rsajv (R8 service: idempotency retention with slack, curator-skill-registry), review round 1 (Change Request revision 1)

You are the independent reviewer of a curator-skill-registry implementation
produced for `TASK-260910-2rsajv` (story `STORY-260910-2xe3n2`, wave 4 of the 2026-09
security-audit remediation). Read, in this order:
`remediation-registry-producer-rules.md`, the producer brief
`TASK-260910-2rsajv_brief.md`, the producer results `TASK-260910-2rsajv_results.md`,
the published Change Request patch `TASK-260910-2rsajv_change-request_rev1.patch`
and its validation log, the finding in
`docs/security-audit-2026-09.md` of the repository, and `profiles/registry-service.md` (the idempotency retention minimum) at the pinned `47c3c8c`.

## Where the candidate is
The managed Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-skill-registry/.temp/STORY-260910-2xe3n2/worktree`
on branch `task-board/story/STORY-260910-2xe3n2` (HEAD = the checkpointed R5 leaf) holds the exact candidate the
runtime published as revision 1 (the hosted gate `scripts/remote-gate.sh`
ran green on it, see the validation log). Do not edit it and leave NO files
in it (no `.review/`, no venv, no caches you did not find): read, and run
tests from a disposable copy elsewhere (e.g. under `/tmp`) with a venv
outside the tree (`python3 -m venv /tmp/csk-review-venv && pip install -e '.[dev]'`
from the copy).

## What to verify
1. **Behaviour**: `IDEMPOTENCY_TTL_SECONDS` is 26 h with a comment quoting the profile minimum and the slack rationale; README/SECURITY/docstrings that state the retention say 26 h; nothing else in the service changed. Quote file:line.
2. **Tests**: a retry between 24 h and 26 h after the first append is deduplicated and a retry after 26 h is treated as new (clock injected as the suite does); the existing idempotency tests adjusted and green. Probe: 25 h 59 min and 26 h 01 min boundaries.
3. **Story branch state**: the candidate sits on the Story branch that already carries the checkpointed R5 leaf (`d7f424c`); the Change Request delta is only this leaf's change (compare the published patch with `git diff HEAD` in the worktree).
4. **Docs**: CHANGELOG "R8" entry.
N. **Independent validation**: `python -m pytest -q` (with
   `CURATOR_CONFORMANCE_ROOT` set to the pinned `47c3c8c` root — the
   curator-spec checkout is at a later main, so use a detached worktree
   at `47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe` under `/tmp`) and
   `python -m mypy` from the disposable copy (`set -o pipefail`, exit
   codes, interpreter); at least two narrowing mutants caught by committed
   tests; scope and hygiene (brief's out-of-scope list respected, CHANGELOG
   entry present, nothing left in the worktree).

## Verdict
Record `TASK-260910-2rsajv_review-verdict-rev1.md` (task outcome) with the
per-item table, transcripts, mutants and findings; then exactly one of
`task-board m 'accept_cr(TASK-260910-2rsajv, revision=1, evidence=TASK-260910-2rsajv_review-verdict-rev1.md)'`
or a changes-requested verdict routed with `set_status(TASK-260910-2rsajv, status=to-dev)`
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.
