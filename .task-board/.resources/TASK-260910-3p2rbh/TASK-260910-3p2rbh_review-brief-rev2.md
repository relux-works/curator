# Review brief — TASK-260910-3p2rbh (R2 service: cached /health verdict with a background verifier, curator-skill-registry), review round 2

You are the independent reviewer of a curator-skill-registry implementation
produced for `TASK-260910-3p2rbh` (story `STORY-260910-1py4f3`, wave 3 of the
2026-09 security-audit remediation). Read, in this order:
`remediation-registry-producer-rules.md`, the producer brief
`TASK-260910-3p2rbh_brief.md`, the producer results `TASK-260910-3p2rbh_results.md`,
the published Change Request patch `TASK-260910-3p2rbh_change-request_rev2.patch`
and its validation log, and the landed spec sections the brief names
(curator-spec checkout `/Users/administrator/Developer/ReluxWorks/curator/curator-spec`
at `dced9b8`, read-only: `protocol/registry.md` §5/§6,
`profiles/registry-service.md` §2/§5/§11, the finding R2 in curator-skill-registry docs/security-audit-2026-09.md and the
`registry-service.json` recovery, concurrency and restore cases).

## Where the candidate is
The managed Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-skill-registry/.temp/STORY-260910-1py4f3/worktree`
on branch `task-board/story/STORY-260910-1py4f3` holds the exact candidate the
runtime published as revision 2 (the hosted gate `scripts/remote-gate.sh`
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
Record `TASK-260910-3p2rbh_review-verdict-rev2.md` (task outcome) with the
per-item table, transcripts, mutants and findings; then exactly one of
`task-board m 'accept_cr(TASK-260910-3p2rbh, revision=2, evidence=TASK-260910-3p2rbh_review-verdict-rev2.md)'`
or a changes-requested verdict routed with `set_status(TASK-260910-3p2rbh, status=to-dev)`
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.

## Round 2 specifics
Revision 1 was rejected with four corrections
(`TASK-260910-3p2rbh_review-verdict-rev1.md`); the orchestrator re-decided
correction 4 in `TASK-260910-3p2rbh_rework-rev2.md`: staleness is TRANSIENT
(503 + write refusal while stale; a later successful refresh restores
readiness with a logged event), while a failed refresh or detected corruption
latches until restart — review item 1 above is superseded by that wording.
Verify the closures with your own reproductions from round 1: (1) a real
append interposed before the verifier's frontier load stays green (one WAL
read snapshot across the pass); (2) corrupted length-preserving level-0 tails
+ an append through the production entry → append refused and `/health` 503
via the common integrity path, no per-request full-chain hashing; (3) the
BEFORE INSERT ABORT trigger on `imported_records` → cached verdict never
runs ahead of committed state (also for idempotent replays); (4)
deterministic clock tests at the exact boundary, late completion of a
stalled verifier restoring readiness, and a failing late completion staying
latched; README/SECURITY wording matches the transient/latched split.
