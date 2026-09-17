# Review brief — TASK-260910-1ny7yl (R3+P2 service: serve --checkpoint startup comparison, curator-skill-registry), review round 2 (Change Request revision 3)

You are the independent reviewer of a curator-skill-registry implementation
produced for `TASK-260910-1ny7yl` (story `STORY-260910-35tbgb`, wave 3 of the
2026-09 security-audit remediation). Read, in this order:
`remediation-registry-producer-rules.md`, the producer brief
`TASK-260910-1ny7yl_brief.md`, the producer results `TASK-260910-1ny7yl_results.md`,
the published Change Request patch `TASK-260910-1ny7yl_change-request_rev3.patch`
and its validation log, and the landed spec sections the brief names
(curator-spec checkout `/Users/administrator/Developer/ReluxWorks/curator/curator-spec`
at `47c3c8c`, read-only: `profiles/registry-service.md` §6 (startup checkpoint comparison), §5/§9/§10/§11,
`protocol/registry.md` §5 and the
`registry-service.json` `checkpoint_cases` block).

## Where the candidate is
The managed Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-skill-registry/.temp/STORY-260910-35tbgb/worktree`
on branch `task-board/story/STORY-260910-35tbgb` holds the exact candidate the
runtime published as revision 1 (the hosted gate `scripts/remote-gate.sh`
ran green on it, see the validation log). Do not edit it and leave NO files
in it (no `.review/`, no venv, no caches you did not find): read, and run
tests from a disposable copy elsewhere (e.g. under `/tmp`) with a venv
outside the tree (`python3 -m venv /tmp/csk-review-venv && pip install -e '.[dev]'`
from the copy).

## What to verify
1. **Spec conformance** (profile §6 as landed in `47c3c8c`): with
   `serve --checkpoint`, the signed `registry-snapshot-v1` file is verified
   against the accepted signing keys and compared with the live boundary
   AFTER the §5 startup verification (incl. the R2 boundary/frontier
   rebuild) and BEFORE the listener binds / readiness; the four closed
   outcomes (`restore_below_checkpoint`, `restore_inconsistent_with_checkpoint`
   for equal-different and above-without-prefix-reproduction,
   `checkpoint_signature_invalid`) refuse as non-ready with writes disabled
   through the common integrity path, no truncation; without a checkpoint
   the service starts and records `checkpoint_not_configured`; the compared
   boundary and outcome are in the structured startup log; frozen
   `health-response-v1` untouched; `verify-backup` kept and the docs say
   which is normative. Quote file:line.
2. **Vectors are driven**: `registry-service.json` `checkpoint_cases` through
   the real startup path (every case), the AC scenario "restore an older
   database and prove serve refuses" through the CLI entry, CI ref moved to
   `47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe`, all pre-existing conformance
   tests green there.
3. **The rev-1 gate repair is harness-only**: revision 1 failed once on Py
   3.14/ubuntu in the R2 test `test_health_stale_verifier_fails_closed_and_refresh_recovers`
   (background verifier racing the manual clock); revision 2 pauses the
   verifier for that test (`TASK-260910-1ny7yl_gate-failure-rev1.md`).
   Confirm the production path is unchanged by that repair, the test-only
   hook cannot be reached in production, and the exact-boundary assertions
   still hold.
4. **Independent validation**: `python -m pytest -q` (with the root set) and
   `python -m mypy` from a disposable copy (`set -o pipefail`, exit codes,
   interpreter); at least two narrowing mutants (e.g. skip the
   above-checkpoint prefix reproduction; accept an unverified checkpoint
   signature) caught by committed tests.
5. **Scope and hygiene**: no P4/R4–R8 content, no key-rotation change,
   CHANGELOG R3/P2 entry, README/SECURITY/compose checkpoint docs, nothing
   left in the worktree.

## Verdict
Record `TASK-260910-1ny7yl_review-verdict-rev3.md` (task outcome) with the
per-item table, transcripts, mutants and findings; then exactly one of
`task-board m 'accept_cr(TASK-260910-1ny7yl, revision=3, evidence=TASK-260910-1ny7yl_review-verdict-rev3.md)'`
or a changes-requested verdict routed with `set_status(TASK-260910-1ny7yl, status=to-dev)`
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.

## Round 2 specifics (revision 3)
Revision 2 was rejected with three P2 corrections
(`TASK-260910-1ny7yl_review-verdict-rev2.md`): the startup posture/outcome
events never reached a sink in the real CLI (logging configured after
enforcement); no discriminating negative for the above-checkpoint
`merkle_root` comparison; the restore scenario did not traverse the CLI.
Verify each closure with your own reproductions: a fresh-process console
entry point with no checkpoint and with a valid checkpoint emits the
`startup_checkpoint` events (posture and successful comparison); the
genuine-prefix-head/different-merkle-root checkpoint above the live store
→ `restore_inconsistent_with_checkpoint`, non-ready, writes disabled, and
the mutant dropping the `merkle_root` comparison now fails a committed
test; the end-to-end restore test invokes the real `main([... "serve",
"--checkpoint" ...])` with factory and enforcement real (only Uvicorn's run
boundary intercepted) and observes 503, write refusal, unchanged history;
env-fallback/flag-precedence coverage kept; the rev-2→rev-3 diff stays
within those corrections.
