# Review brief — TASK-260910-27yepb (P1 service half: cursor bound to the page boundary, curator-skill-registry), review round 1

You are the independent reviewer of a curator-skill-registry implementation
produced for `TASK-260910-27yepb` (story `STORY-260910-3rvvxh`, wave 2 of the
2026-09 security-audit remediation). Read, in this order:
`remediation-registry-producer-rules.md`, the producer brief
`TASK-260910-27yepb_brief.md`, the producer results `TASK-260910-27yepb_results.md`,
the published Change Request patch `TASK-260910-27yepb_change-request_rev1.patch`
and its validation log, and the landed spec sections the brief names
(curator-spec checkout `/Users/administrator/Developer/ReluxWorks/curator/curator-spec`
at `dced9b8`, read-only: `protocol/registry.md` §5/§9/§9.3,
`profiles/registry-service.md` §2/§5/§11, the v2 envelope schemas and the
`registry-service.json` `pagination` block).

## Where the candidate is
The managed Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-skill-registry/.temp/STORY-260910-3rvvxh/worktree`
on branch `task-board/story/STORY-260910-3rvvxh` holds the exact candidate the
runtime published as revision 1 (the hosted gate `scripts/remote-gate.sh`
ran green on it, see the validation log). Do not edit it and leave NO files
in it (no `.review/`, no venv, no caches you did not find): read, and run
tests from a disposable copy elsewhere (e.g. under `/tmp`) with a venv
outside the tree (`python3 -m venv /tmp/csk-review-venv && pip install -e '.[dev]'`
from the copy).

## What to verify
1. **Spec conformance** (profile §2/§5/§11, registry §9.3): a cursor page is
   served only at the cursor's carried boundary; the carried boundary's
   fields are verified against the store (head/merkle_root/log_size at that
   log_size) before serving; disagreement, a pruned/unavailable prefix, or a
   forged-but-acceptably-signed carried object → `404 invalid_cursor`; no
   fallback to `store.snapshot_boundary()` for a cursor page (never
   re-evaluated at a newer boundary); both endpoints. Quote file:line. The
   first leaf's guarantees (boundary on every page, byte-identical chain
   across rotation, real schema validation) still hold on this tree.
2. **Vectors are driven**: `registry-service.json` `pagination.cursor_boundary_cases`
   through the real endpoints; every existing `cursor_rejections` case
   green; the chain is not re-evaluated after an append.
3. **Independent validation**: `python -m pytest -q` (with the root set) and
   `python -m mypy` from a disposable copy (`set -o pipefail`, exit codes,
   interpreter); at least two narrowing mutants of your own (e.g. drop the
   store-field verification of the carried boundary; fall back to the
   current boundary when the carried one is unavailable) and confirm a
   committed test catches each; report survivors with bounds.
4. **Scope and hygiene**: no R2 memoization, no key-rotation policy change,
   no unrelated edits; CHANGELOG R1 entry extended with P1; README/SECURITY
   sentence; no writes outside the worktree and nothing left in it.
5. Producer-reported observations: check against the text.

## Verdict
Record `TASK-260910-27yepb_review-verdict-rev1.md` (task outcome) with the
per-item table, transcripts, mutants and findings; then exactly one of
`task-board m 'accept_cr(TASK-260910-27yepb, revision=1, evidence=TASK-260910-27yepb_review-verdict-rev1.md)'`
or a changes-requested verdict routed with `set_status(TASK-260910-27yepb, status=to-dev)`
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.
