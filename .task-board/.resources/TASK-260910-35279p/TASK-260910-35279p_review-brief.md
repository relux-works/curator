# Review brief — TASK-260910-35279p (R2 service: memoized snapshot boundary, curator-skill-registry), review round 1

You are the independent reviewer of a curator-skill-registry implementation
produced for `TASK-260910-35279p` (story `STORY-260910-1py4f3`, wave 3 of the
2026-09 security-audit remediation). Read, in this order:
`remediation-registry-producer-rules.md`, the producer brief
`TASK-260910-35279p_brief.md`, the producer results `TASK-260910-35279p_results.md`,
the published Change Request patch `TASK-260910-35279p_change-request_rev1.patch`
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
1. **Correctness over the log**: the durable `boundaries` / `merkle_frontier`
   tables (schema 3) are written in the same transaction as the append,
   rebuilt (never trusted) at startup and after unclean shutdown, and a
   mismatch between the recomputed chain/tree and the cached rows fails
   readiness like any §5 mismatch; `snapshot_boundary(max_seq)`,
   `boundary_available()` and the R1 carried-boundary check read cached rows
   for historical boundaries (root-at-size correct for every size, including
   odd sizes and sizes across power-of-two boundaries — probe several sizes
   against a from-scratch Merkle computation). Migration from schema 2 is
   idempotent and crash-safe (interrupt mid-backfill in a test if feasible).
2. **The proof is by operation count, not timing**: a committed test shows
   zero Merkle hash recomputation for repeated `/v1/records`, `/v1/log`,
   `/v1/snapshot` and cursor validations after an append, and O(log n) per
   append; confirm the counter is real (monkeypatch or instrumented hasher)
   and cannot be satisfied by an implementation that just skips verification.
3. **Independent validation**: `python -m pytest -q` (with
   `CURATOR_CONFORMANCE_ROOT` set) and `python -m mypy` from a disposable copy
   (`set -o pipefail`, exit codes, interpreter); every existing concurrency,
   idempotency, recovery, restore and rotation conformance test green; at
   least two narrowing mutants (e.g. trust the cached row without startup
   recomputation; skip frontier update on append) caught by committed tests.
4. **Lock discipline**: what moved outside the write lock and why it is safe
   across separate `Store` instances/processes (the producer's justification
   for durable tables over an in-memory frontier); no weakening of the
   single-writer serialization.
5. **Scope and hygiene**: no `/health` changes (sibling), no protocol/profile
   changes, CHANGELOG R2 entry (this half), README operations note; nothing
   left in the worktree.

## Verdict
Record `TASK-260910-35279p_review-verdict-rev1.md` (task outcome) with the
per-item table, transcripts, mutants and findings; then exactly one of
`task-board m 'accept_cr(TASK-260910-35279p, revision=1, evidence=TASK-260910-35279p_review-verdict-rev1.md)'`
or a changes-requested verdict routed with `set_status(TASK-260910-35279p, status=to-dev)`
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.
