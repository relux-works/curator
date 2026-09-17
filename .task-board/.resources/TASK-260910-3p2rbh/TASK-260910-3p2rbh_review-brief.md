# Review brief — TASK-260910-3p2rbh (R2 service: cached /health verdict with a background verifier, curator-skill-registry), review round 1

You are the independent reviewer of a curator-skill-registry implementation
produced for `TASK-260910-3p2rbh` (story `STORY-260910-1py4f3`, wave 3 of the
2026-09 security-audit remediation). Read, in this order:
`remediation-registry-producer-rules.md`, the producer brief
`TASK-260910-3p2rbh_brief.md`, the producer results `TASK-260910-3p2rbh_results.md`,
the published Change Request patch `TASK-260910-3p2rbh_change-request_rev1.patch`
and its validation log, and the landed spec sections the brief names
(curator-spec checkout `/Users/administrator/Developer/ReluxWorks/curator/curator-spec`
at `dced9b8`, read-only: `protocol/registry.md` §5/§6,
`profiles/registry-service.md` §2/§5/§11, the finding R2 in curator-skill-registry docs/security-audit-2026-09.md and the
`registry-service.json` recovery, concurrency and restore cases).

## Where the candidate is
The managed Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-skill-registry/.temp/STORY-260910-1py4f3/worktree`
on branch `task-board/story/STORY-260910-1py4f3` holds the exact candidate the
runtime published as revision 1 (the hosted gate `scripts/remote-gate.sh`
ran green on it, see the validation log). Do not edit it and leave NO files
in it (no `.review/`, no venv, no caches you did not find): read, and run
tests from a disposable copy elsewhere (e.g. under `/tmp`) with a venv
outside the tree (`python3 -m venv /tmp/csk-review-venv && pip install -e '.[dev]'`
from the copy).

## What to verify
1. **Fail-closed semantics** (profile §5/§9): `/health` serves the cached
   verdict; a failed refresh, a verifier that found corruption, or a stale
   verdict (no completed refresh within the staleness bound = multiplier ×
   interval) → `503` protocol error envelope AND writes disabled through the
   same integrity path a startup §5 mismatch uses; readiness stays down until
   a restart re-verifies; the first verdict is the startup verification; a
   hung verifier cannot leave a green verdict forever (probe it: freeze the
   clock / block the verifier and observe the flip at the bound).
2. **Zero full-chain work per probe**: the operation-count test proves
   `/health` performs no chain/Merkle recomputation between refreshes; the
   verifier's own walk is bounded and does not hold the write lock for the
   whole walk (state what it locks); appends update `verified_head`
   incrementally and cannot mark corrupted state verified.
3. **Corruption detection**: a mutation injected after startup (log row,
   canonical bytes, cached boundary row) is detected at the next explicitly
   driven refresh (no sleeps in tests), flips `/health` and refuses writes;
   restart after repair → ready; every existing `recovery_cases` and the §9
   conformance test green.
4. **Independent validation**: `python -m pytest -q` (with
   `CURATOR_CONFORMANCE_ROOT` set) and `python -m mypy` from a disposable
   copy (`set -o pipefail`, exit codes, interpreter); at least two narrowing
   mutants (e.g. staleness check removed; refresh failure ignored) caught by
   committed tests.
5. **Operations and scope**: flag/env var (`--health-verify-interval`,
   `CURATOR_SKILL_REGISTRY_HEALTH_VERIFY_INTERVAL`) documented in README /
   SECURITY / compose; the frozen `health-response-v1` schema untouched; no
   protocol/profile edits; CHANGELOG R2 entry completed; nothing left in the
   worktree.

## Verdict
Record `TASK-260910-3p2rbh_review-verdict-rev1.md` (task outcome) with the
per-item table, transcripts, mutants and findings; then exactly one of
`task-board m 'accept_cr(TASK-260910-3p2rbh, revision=1, evidence=TASK-260910-3p2rbh_review-verdict-rev1.md)'`
or a changes-requested verdict routed with `set_status(TASK-260910-3p2rbh, status=to-dev)`
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.
