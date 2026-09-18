# Review brief — TASK-260910-28kmef (R5 service: pathological JSON returns 400 invalid_json, curator-skill-registry), review round 1 (Change Request revision 1)

You are the independent reviewer of a curator-skill-registry implementation
produced for `TASK-260910-28kmef` (story `STORY-260910-2xe3n2`, wave 4 of the 2026-09
security-audit remediation). Read, in this order:
`remediation-registry-producer-rules.md`, the producer brief
`TASK-260910-28kmef_brief.md`, the producer results `TASK-260910-28kmef_results.md`,
the published Change Request patch `TASK-260910-28kmef_change-request_rev1.patch`
and its validation log, the finding in
`docs/security-audit-2026-09.md` of the repository, and `profiles/registry-service.md` (error envelope and `invalid_json` semantics) at the pinned `dced9b8`.

## Where the candidate is
The managed Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-skill-registry/.temp/STORY-260910-2xe3n2/worktree`
on branch `task-board/story/STORY-260910-2xe3n2` holds the exact candidate the
runtime published as revision 1 (the hosted gate `scripts/remote-gate.sh`
ran green on it, see the validation log). Do not edit it and leave NO files
in it (no `.review/`, no venv, no caches you did not find): read, and run
tests from a disposable copy elsewhere (e.g. under `/tmp`) with a venv
outside the tree (`python3 -m venv /tmp/csk-review-venv && pip install -e '.[dev]'`
from the copy).

## What to verify
1. **Behaviour**: `RecursionError` from `protocol.load_json` and from the canonicalization / CCJ validation path maps to `ProtocolError` so every body-parsing endpoint answers `400 invalid_json` with the existing envelope, never `500`; the depth bound (if explicit) is stated and cheap; no other exception mapping changed; internal callers unchanged. Quote file:line.
2. **Tests**: an HTTP-level test posts a document nested beyond the bound to `submit` (and every other body-parsing endpoint) and asserts 400 + `invalid_json`; unit tests for `load_json` and the canonicalization path; the store's own JSON never trips the bound; the existing suite green. Probe: craft a 20 000-deep array and a 20 000-deep object yourself against a disposable server and confirm 400, then confirm a 500 cannot be produced with nested-inside-valid documents (e.g. deep nesting inside a signature envelope member).
3. **Docs**: README limits section / docstring and CHANGELOG "R5" entry.
N. **Independent validation**: `python -m pytest -q` (with
   `CURATOR_CONFORMANCE_ROOT` set to the pinned `dced9b8` root — the
   curator-spec checkout is at a later main, so use a detached worktree
   at `dced9b8317e0e8af79edf2d0539b32bd22b6c85b` under `/tmp`) and
   `python -m mypy` from the disposable copy (`set -o pipefail`, exit
   codes, interpreter); at least two narrowing mutants caught by committed
   tests; scope and hygiene (brief's out-of-scope list respected, CHANGELOG
   entry present, nothing left in the worktree).

## Verdict
Record `TASK-260910-28kmef_review-verdict-rev1.md` (task outcome) with the
per-item table, transcripts, mutants and findings; then exactly one of
`task-board m 'accept_cr(TASK-260910-28kmef, revision=1, evidence=TASK-260910-28kmef_review-verdict-rev1.md)'`
or a changes-requested verdict routed with `set_status(TASK-260910-28kmef, status=to-dev)`
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.
