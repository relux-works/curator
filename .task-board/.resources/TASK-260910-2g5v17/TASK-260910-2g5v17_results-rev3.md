# TASK-260910-2g5v17 results rev3 — landed-tree re-publish (P4, mechanical)

Run `RUN-260918-a36f78`, role `developer`. This is the landed-rework run
prescribed by `TASK-260910-2g5v17_landed-rework-run.md`: re-publish the exact
landed tree as revision 3 for re-review. NOT an integration run — no
`integrate` / `complete` / `close-landed` was attempted (the predecessor run
already proved all three refuse revision 2).

## Accepted tree vs landed tree

- Accepted candidate rev2 tree: `1dfd8513e036b3e2006dd6fc620e396ad2638ffb`
- Landed commit (PR #11, signed): `bf5cac1200cfa39dff0a6b0449072ff5b22f124d`,
  tree `9b7d33115bd3ffe44c34c5a340de72aaab06d0cb`
- Delta (the export's `update_patch`, sha256
  `c636f9eaa861db8e0efbb048c2eee6368f4c762ac12cda8b83a524c3deaad781`,
  verified before apply): rebase of the P4 change over the R6 landing
  (`fb86420`), CHANGELOG entries unioned (P4 entry kept verbatim, R6 entry
  added), and the stray board artifact `TASK-260910-2g5v17_results.md`
  removed from the repository tree (it is a task resource, not source).
- `update_patch` file list: `CHANGELOG.md`, `README.md`, `SECURITY.md`,
  `TASK-260910-2g5v17_results.md` (deleted), `compose.yaml`,
  `src/csk_registry/__init__.py`, `src/csk_registry/cli.py`,
  `src/csk_registry/keys.py`, `src/csk_registry/signing.py`,
  `tests/test_key_passphrase.py` (new). The `cli.py` hunks are R6-only
  (`KeyPassphraseError` import + `except` in `main`).
- P4 implementation is byte-identical to accepted rev2: `bundle.py`,
  `store.py`, `tests/test_registry.py` are NOT in `update_patch`. All rev2
  behaviour evidence, transcripts and the rev2 review verdict carry over;
  only the R6 rebase + artifact removal are new.

## Commands and outputs (abridged only for the 171 KiB export JSON)

1. `task-board worktree obligations` → exit 0:
   `TASK-260910-2g5v17  2  accepted  checkpoint  18m  integrating  STORY-260910-stz5f0`
   (matches the expected pre-state; no `set_status` needed).
2. `task-board worktree prepare-landed-review STORY-260910-stz5f0 --cr
   TASK-260910-2g5v17 --revision 2 --landed-commit
   bf5cac1200cfa39dff0a6b0449072ff5b22f124d --json >
   /tmp/2g5v17-landed-review.json` → exit 0. Export summary:
   `comparison_base_oid=c7ef32c75cc8dfda1abe38647af282a03175e8d3`,
   `previous_tree_oid=1dfd8513e036b3e2006dd6fc620e396ad2638ffb`,
   `observed_landing.tree_oid=9b7d33115bd3ffe44c34c5a340de72aaab06d0cb`
   (commit `bf5cac1…`, `refs/heads/main`, author Relux Bot, signature `G`,
   observed `2026-09-18T12:38:08Z`), 12 `changed_paths`,
   `review_patch` 110538 chars / 2554 lines (sha256 `49ec78c9…`),
   `update_patch` 56482 chars / 1237 lines (sha256 `c636f9ea…`).
   Full JSON attached as `TASK-260910-2g5v17_landed-review-export.json`.
3. `task-board m 'add_resource(TASK-260910-2g5v17,
   name="TASK-260910-2g5v17_landed-review-export.json",
   path="/tmp/2g5v17-landed-review.json", type=outcome, …)'` → exit 0,
   `{"ok":true,…}`.
4. `task-board worktree start-landed-rework STORY-260910-stz5f0 --cr
   TASK-260910-2g5v17 --revision 2 --landed-commit bf5cac1… --export
   /tmp/2g5v17-landed-review.json` → exit 0 (re-emits the export JSON).
   `obligations` after: `…  accepted  checkpoint  20m  to-dev  …` —
   leaf released to `to-dev` with recovery intent.
5. Update patch apply in the Story worktree
   (`.temp/STORY-260910-stz5f0/worktree`, base `c7ef32c`, pre-state: 7
   modified files + untracked `TASK-260910-2g5v17_results.md`):
   - extracted `update_patch` to `/tmp/2g5v17-update.patch`, sha256
     `c636f9ea…` matched `update_sha256`;
   - `git apply --check /tmp/2g5v17-update.patch` → exit 0;
   - `git apply /tmp/2g5v17-update.patch` → exit 0.
   - post-state `git status --short --untracked-files=all`: 11 modified
     (`CHANGELOG.md`, `README.md`, `SECURITY.md`, `compose.yaml`,
     `__init__.py`, `bundle.py`, `cli.py`, `keys.py`, `signing.py`,
     `store.py`, `tests/test_registry.py`) + `?? tests/test_key_passphrase.py`;
     `TASK-260910-2g5v17_results.md` gone (confirmed: `ls` → No such file).
6. Tree verification (temp-index method from `scripts/remote-gate.sh`):
   `tree=9b7d33115bd3ffe44c34c5a340de72aaab06d0cb`, `TREE_MATCH`, exit 0.

## Validation on the exact rev3 tree (extra, read-only; hosted gate still runs at handoff)

Interpreter `/tmp/csk-venv/bin/python` 3.14.6, conformance root
`/tmp/spec-47c3c8c/conformance/v1` (detached worktree at `47c3c8c…`, the CI pin).

- `CURATOR_CONFORMANCE_ROOT=… /tmp/csk-venv/bin/python -m pytest -q` → exit 1:
  `1 failed, 205 passed` in 213.20 s; the single failure was
  `tests/test_protocol_conformance.py::test_shared_service_concurrent_writer_vector`
  (`OperationalError` at `store.py:704`, sqlite lock under heavy load).
- Isolated re-run of that test → exit 0: `1 passed` in 13.58 s.
- Identical full re-run → exit 0: `206 passed, 2 warnings` in 55.96 s.
  Verdict: transient timing flake in a concurrent-writer test, not a tree
  defect; the second full run is green.
- `/tmp/csk-venv/bin/python -m mypy` → exit 0:
  `Success: no issues found in 14 source files`.
- Post-run `git status --short --untracked-files=all` lists only the 12
  source/test/doc paths above (`.pytest_cache` / `__pycache__` are ignored).

## Checklist / handoff

All 16 checklist items remain `done:true`; `start-landed-rework` unchecked
nothing, so no `check_item` was run. This file is attached as
`TASK-260910-2g5v17_results-rev3.md` (type `outcome`); handoff follows and
the hosted gate runs once there.
