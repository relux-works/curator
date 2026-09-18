# Integration run evidence — TASK-260910-2g5v17 rev 2 (CR-TASK-260910-2g5v17-2)

Run: RUN-260918-689770 (developer / implementer — matches the accepted revision's
immutable producer binding). Board left at `integrating`; no status change, no handoff
command per the integration assignment. All commands run from the control root
`/Users/administrator/Developer/ReluxWorks/curator/curator-skill-registry`
unless noted. Nothing was written to trunk, the board, or the Story worktree by this run.

## Outcome

Revision 2 **cannot be landed or closed as-is**: the P4 change already reached trunk
outside the board as PR #11 (`bf5cac1`), rebased over the R6 landing with a different
tree. The three integration exits were attempted/verified and each refused with a typed
code (exit 1 in every case — these refusals are the evidence, not a run failure).
The prescribed route is the exact-tree re-review in the attached
`TASK-260910-2g5v17_landed-rework-run.md`
(prepare-landed-review → start-landed-rework → apply update_patch → handoff rev3 →
independent review → accept → complete).

## Preflight facts

- Accepted record: kind `story_final`, base `c7ef32c75cc8dfda1abe38647af282a03175e8d3`,
  candidate tree `1dfd8513e036b3e2006dd6fc620e396ad2638ffb`, 8 changed paths including
  the stray `TASK-260910-2g5v17_results.md` artifact.
  (source: `.temp/changerequests/TASK-260910-2g5v17/rev-000002.json`)
- Protected trunk: `refs/heads/main` at `bf5cac1200cfa39dff0a6b0449072ff5b22f124d`
  ("Persist a per-upstream high-water and refuse rollback bundles on import (P4)"),
  on top of R6 `fb86420`. Control root on `main`, clean.
- Trunk advance `c7ef32c..bf5cac` touches 7 of the 8 CR paths
  (CHANGELOG.md, README.md, SECURITY.md, bundle.py, cli.py, store.py, test_registry.py).
- Landed tree check:

```text
$ git rev-parse 'bf5cac1200cfa39dff0a6b0449072ff5b22f124d^{tree}'
9b7d33115bd3ffe44c34c5a340de72aaab06d0cb
candidate_tree_oid: 1dfd8513e036b3e2006dd6fc620e396ad2638ffb
(exit 0 — the two trees differ)
```

- Landed commit is signed (`Good "git" signature for bot@relux.works`), authored by
  Relux Bot, and is the tip of `refs/heads/main` (exit 0 on all probes).

## Gate 1 — `worktree integrate` (the bound integration command)

```text
$ task-board worktree integrate STORY-260910-stz5f0 --cr TASK-260910-2g5v17 --revision 2
board_owner_separate: spawn.worktree_isolation.board_repository declares a separate board owner, so worktree integrate — which commits board state into the control root — is not this repository's delivery path; land the code through its own PR and run worktree complete
  board_repository_root: /Users/administrator/Developer/ReluxWorks/curator/curator
  control_root: /Users/administrator/Developer/ReluxWorks/curator/curator-skill-registry
  story_id: STORY-260910-stz5f0
INTEGRATE-EXIT:1
```

Expected refusal: this repository delivers code via its own PR and closes Stories with
`worktree complete`, not `worktree integrate`. Nothing written.

## Gate 2 — `worktree complete` with the observed landing commit

```text
$ task-board worktree complete STORY-260910-stz5f0 --cr TASK-260910-2g5v17 --revision 2 --landed-commit bf5cac1200cfa39dff0a6b0449072ff5b22f124d
code_landing_tree_mismatch: the landed tree differs from the accepted candidate; historical caller bases cannot authorize a different tree; use worktree prepare-landed-review, normal producer handoff and a new exact-tree accept_cr
  candidate_tree: 1dfd8513e036b3e2006dd6fc620e396ad2638ffb
  cr_id: CR-TASK-260910-2g5v17-2
  declared_commit: bf5cac1200cfa39dff0a6b0449072ff5b22f124d
  landed_tree: 9b7d33115bd3ffe44c34c5a340de72aaab06d0cb
  protected_oid: bf5cac1200cfa39dff0a6b0449072ff5b22f124d
  protected_ref: refs/heads/main
  remote_url: ssh://git@github.com/relux-works/curator-skill-registry
  resolved_commit: bf5cac1200cfa39dff0a6b0449072ff5b22f124d
COMPLETE-EXIT:1
```

Expected refusal, before anything was written: the landed tree (`9b7d331…`, rebased
over R6 with the CHANGELOG union and the results artifact dropped) is not the accepted
candidate tree (`1dfd851…`). The refusal itself prescribes the exact-tree re-review.

## Gate 3 — `worktree close-landed --dry-run` (read-only probe)

```text
$ task-board worktree close-landed TASK-260910-2g5v17 --dry-run --reason "integration-run evidence probe rev2"
landed_tree_not_on_trunk: TASK-260910-2g5v17 revision 2 is not on refs/heads/main at bf5cac1200cf: the tree proof failed (landed_tree_not_on_trunk) and the accepted patch does not reverse-apply to the tip
  apply_stderr: error: patch failed: src/csk_registry/cli.py:12
error: src/csk_registry/cli.py: patch does not apply
error: TASK-260910-2g5v17_results.md: No such file or directory
error: patch failed: CHANGELOG.md:33
error: CHANGELOG.md: patch does not apply
  ...
  cause: landed_tree_not_on_trunk: no commit in the repository carries candidate tree 1dfd8513e036 of TASK-260910-2g5v17 revision 2, so it is not on refs/heads/main
CLOSE-LANDED-DRY-EXIT:1
```

Confirms neither `tree_carried` (0 carriers) nor `patch_contained` (reverse-apply
fails: R6 context drift in cli.py, absent results artifact, CHANGELOG union) can close
rev 2. Nothing written (`--dry-run`).

## Post-run state (unchanged)

- `task-board worktree transaction show STORY-260910-stz5f0` →
  `No integration transaction is recorded for STORY-260910-stz5f0`
- `task-board worktree obligations` →
  `TASK-260910-2g5v17  2  accepted  checkpoint  …  integrating  STORY-260910-stz5f0`
- Control root still at `bf5cac1` on `main`, clean; Story worktree untouched
  (still the rev-2 candidate delta, uncommitted, as handed off).

## Routing recommendation (for the orchestrator, not executed here)

Do not integrate or close rev 2. Run the attached landed-rework instruction
(`TASK-260910-2g5v17_landed-rework-run.md`): `prepare-landed-review` for landed
commit `bf5cac1200cfa39dff0a6b0449072ff5b22f124d`, `start-landed-rework`, apply only
the export's `update_patch` in the Story worktree, verify the snapshot tree equals
`9b7d33115bd3ffe44c34c5a340de72aaab06d0cb`, attach rev-3 results, handoff, independent
review, accept, then `worktree complete`.
