# Review brief — TASK-260910-2c7s0u (R4 service docs: proxy rate-limit bucketing and body-before-auth bounds, curator-skill-registry), review round 1 (Change Request revision 1)

You are the independent reviewer of a curator-skill-registry implementation
produced for `TASK-260910-2c7s0u` (story `STORY-260910-2xe3n2`, wave 4 of the 2026-09
security-audit remediation). Read, in this order:
`remediation-registry-producer-rules.md`, the producer brief
`TASK-260910-2c7s0u_brief.md`, the producer results `TASK-260910-2c7s0u_results.md`,
the published Change Request patch `TASK-260910-2c7s0u_change-request_rev1.patch`
and its validation log, the finding in
`docs/security-audit-2026-09.md` of the repository, and the service code the docs describe (`src/csk_registry/app.py` limiter, semaphore, body limits and forwarded-header handling).

## Where the candidate is
The managed Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-skill-registry/.temp/STORY-260910-2xe3n2/worktree`
on branch `task-board/story/STORY-260910-2xe3n2` (HEAD = the replayed R7 checkpoint `7a8627b`) holds the exact candidate the
runtime published as revision 1 (the hosted gate `scripts/remote-gate.sh`
ran green on it, see the validation log). Do not edit it and leave NO files
in it (no `.review/`, no venv, no caches you did not find): read, and run
tests from a disposable copy elsewhere (e.g. under `/tmp`) with a venv
outside the tree (`python3 -m venv /tmp/csk-review-venv && pip install -e '.[dev]'`
from the copy).

## What to verify
1. **Docs vs code** (this is a docs task; no code change is allowed): every documented limit and setting matches the code — the network limiter key, the forwarded-header trust setting/flag/env with its default (quote `app.py` file:line), 16 MiB / 15 s / 128 slots / 0.1 s acquire, `503 overloaded` semantics; a mismatch is a finding. The recommended proxy configuration and the worked nginx/Caddy/Traefik snippet are correct for the stated setting; the operator checklist is actionable; one place only (README or the deployment doc it delegates to). SECURITY.md paragraph and CHANGELOG "R4" entry present.
2. **Story-final state**: this is the last leaf of `STORY-260910-2xe3n2`; the Story branch was replayed by `refresh-candidate` onto `main` `bf5cac1` (three checkpoints `0b19c4a` R5, `b9847a1` R8, `7a8627b` R7 on top of P4/R6). Verify the replay is faithful: per-file patch-id of each replayed checkpoint vs the original checkpoints `d7f424c`/`bd27d32`/`5f3b028` (identical except `CHANGELOG.md`, whose entries must be the union of both sides, nothing dropped or reworded); the Change Request delta (this leaf) touches only the three documentation files; the whole story tree = `bf5cac1` + R5 + R8 + R7 + R4 docs; `python -m pytest -q` and `python -m mypy` green on that tree (`47c3c8c` root).
3. **Hygiene**: no results/logbook/build files anywhere in the candidate; nothing left in the worktree by you.
N. **Independent validation**: `python -m pytest -q` (with
   `CURATOR_CONFORMANCE_ROOT` set to the pinned `47c3c8c` root — the
   curator-spec checkout is at a later main, so use a detached worktree
   at `47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe` under `/tmp`) and
   `python -m mypy` from the disposable copy (`set -o pipefail`, exit
   codes, interpreter); at least two narrowing mutants caught by committed
   tests; scope and hygiene (brief's out-of-scope list respected, CHANGELOG
   entry present, nothing left in the worktree).

## Verdict
Record `TASK-260910-2c7s0u_review-verdict-rev1.md` (task outcome) with the
per-item table, transcripts, mutants and findings; then exactly one of
`task-board m 'accept_cr(TASK-260910-2c7s0u, revision=1, evidence=TASK-260910-2c7s0u_review-verdict-rev1.md)'`
or a changes-requested verdict routed with `set_status(TASK-260910-2c7s0u, status=to-dev)`
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.
