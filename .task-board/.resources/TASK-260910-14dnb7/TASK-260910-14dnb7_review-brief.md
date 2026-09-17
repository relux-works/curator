# Review brief — TASK-260910-14dnb7 (R1 service half: boundary in page envelopes, curator-skill-registry), review round 1

You are the independent reviewer of a curator-skill-registry implementation
produced for `TASK-260910-14dnb7` (story `STORY-260910-3rvvxh`, wave 2 of the
2026-09 security-audit remediation). Read, in this order:
`remediation-registry-producer-rules.md`, the producer brief
`TASK-260910-14dnb7_brief.md`, the producer results `TASK-260910-14dnb7_results.md`,
the published Change Request patch `TASK-260910-14dnb7_change-request_rev1.patch`
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
1. **Spec conformance, item by item** against the landed text: the REQUIRED
   `boundary` member on every `/v1/records` and `/v1/log` success envelope is
   the full signed `registry-snapshot-v1` object of the evaluated boundary;
   byte-identical (CCJ-1) across a cursor chain, including after appends and
   across time (immutable `created_at` for one boundary, §5); equals
   `/v1/snapshot` for the same boundary; envelopes match
   `records-response-v2` / `log-response-v2` exactly (`additionalProperties:
   false`); error paths unchanged. Quote file:line.
2. **Vectors are driven**: the `registry-service.json` `pagination` flags
   (`boundary_emitted_on_every_page`, `chain_boundary_byte_identical`) are
   asserted through the real endpoints/store entry points, not a helper-only
   test; the v2 schema-cases are asserted on served envelopes. The CI pin
   moved to `dced9b8317e0e8af79edf2d0539b32bd22b6c85b` and every pre-existing
   conformance test is green at that pin (run the suite yourself with
   `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`).
3. **Independent validation**: `python -m pytest -q` (with the root set) and
   `python -m mypy` from the disposable copy (`set -o pipefail`, quote exit
   codes and interpreter version); attack the change with at least two
   narrowing mutants (e.g. drop `boundary` from cursor pages, refresh
   `created_at`, sign a different body) and confirm a committed test catches
   each; report survivors with bounds.
4. **Scope and hygiene**: no cursor/boundary disagreement refusal beyond what
   the sibling `TASK-260910-27yepb` owns (report if the producer pre-empted it,
   and whether that is harmful), no unrelated edits, no Docker/deploy changes;
   CHANGELOG Unreleased entry names R1, the member, the v2 schemas and spec
   `dced9b8`; README mention; no writes outside the worktree.
5. Anything the producer reports as a spec gap or observation: check it
   against the text; a real gap is a finding for the orchestrator, not a
   reason to accept divergent behaviour.

## Verdict
Record `TASK-260910-14dnb7_review-verdict-rev1.md` (task outcome) with the
per-item table, transcripts, mutants and findings; then exactly one of
`task-board m 'accept_cr(TASK-260910-14dnb7, revision=1, evidence=TASK-260910-14dnb7_review-verdict-rev1.md)'`
or a changes-requested verdict routed with `set_status(TASK-260910-14dnb7, status=to-dev)`
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.
