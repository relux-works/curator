# Review brief — TASK-260910-2g5v17 (P4, curator-skill-registry), review round 3: exact landed tree (Change Request revision 3)

You are the independent reviewer of revision 3 of `TASK-260910-2g5v17`
(story `STORY-260910-stz5f0`). Revision 2 (candidate tree
`1dfd8513e036b3e2006dd6fc620e396ad2638ffb` on base `c7ef32c`) was accepted
(`TASK-260910-2g5v17_review-verdict-rev2.md`). It was then landed on
curator-skill-registry `main` as the signed commit `bf5cac1200cfa39dff0a6b0449072ff5b22f124d` (PR #11) with a
DIFFERENT tree, `9b7d33115bd3ffe44c34c5a340de72aaab06d0cb`: the candidate
rebased onto the R6 landing `fb86420` (CHANGELOG entries unioned) minus the
board artifact `TASK-260910-2g5v17_results.md` the accepted tree carried at
the repository root. Per the board canon a changed landed tree needs an
exact-tree re-review: `worktree prepare-landed-review` exported the delta
(`TASK-260910-2g5v17_landed-review-export.json`), the landed rework applied
it to the Story worktree, and revision 3 republishes exactly that tree.
Read: the export, `TASK-260910-2g5v17_results-rev3.md`, the rev-2 verdict,
the published `TASK-260910-2g5v17_change-request_rev3.patch` and its
validation log, and `remediation-registry-producer-rules.md`.

## Where the candidate is
The managed Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-skill-registry/.temp/STORY-260910-stz5f0/worktree`
(HEAD `c7ef32c`, uncommitted delta). Read only; leave NO files in it; run
anything from a disposable copy under `/tmp` with a venv outside the tree.

## What to verify (all read-only)
1. **Exact tree**: compute the worktree snapshot tree with a temporary index
   (`GIT_INDEX_FILE=$tmp git read-tree HEAD; git add -A .; git write-tree`)
   — it MUST equal `9b7d33115bd3ffe44c34c5a340de72aaab06d0cb`, which MUST
   equal `git rev-parse bf5cac1200cfa39dff0a6b0449072ff5b22f124d^{tree}` and the tree of `origin/main`
   (`git fetch` in the disposable copy; `git verify-commit bf5cac1200cfa39dff0a6b0449072ff5b22f124d` Good by the
   bot key). Quote all three.
2. **Delta accounting**: `git diff-tree -r 1dfd8513e036b3e2006dd6fc620e396ad2638ffb 9b7d33115bd3ffe44c34c5a340de72aaab06d0cb`
   lists exactly: the R6 landing's files (`README.md`, `SECURITY.md`,
   `compose.yaml`, `src/csk_registry/__init__.py`, `src/csk_registry/cli.py`,
   `src/csk_registry/keys.py`, `src/csk_registry/signing.py`,
   `tests/test_key_passphrase.py`), `CHANGELOG.md` (union of the P4 and R6
   entries, both verbatim) and the deletion of `TASK-260910-2g5v17_results.md`.
   For every file the R6 landing touched, the content in the landed tree
   equals `fb86420`'s content OR (for `CHANGELOG.md`, `README.md`,
   `SECURITY.md`, `cli.py`) is the faithful union of the R6 landing and the
   accepted P4 candidate — show the three-way evidence (`git diff fb86420 9b7d331 -- <file>`
   must equal the P4 candidate's change to that file, hunk by hunk / patch-id).
   Nothing else may differ.
3. **Validation** on the exact tree from the disposable copy: `python -m pytest -q`
   with `CURATOR_CONFORMANCE_ROOT` at a detached curator-spec worktree
   `47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe`, and `python -m mypy` (exit
   codes, interpreter); the hosted gate log for revision 3 read.
4. **Hygiene**: no results/logbook/build/coverage files anywhere in the
   candidate; `git status --short --untracked-files=all` of the worktree
   lists only source, tests, docs and packaging paths.

## Verdict
Record `TASK-260910-2g5v17_review-verdict-rev3.md` (task outcome) with the
per-item table and transcripts, then exactly one of
`task-board m 'accept_cr(TASK-260910-2g5v17, revision=3, evidence=TASK-260910-2g5v17_review-verdict-rev3.md)'`
or a changes-requested verdict routed with `set_status(TASK-260910-2g5v17, status=to-dev)`
naming the exact discrepancy. Never accept on the producer's or the
orchestrator's word alone.
