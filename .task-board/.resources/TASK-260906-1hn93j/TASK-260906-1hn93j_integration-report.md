# TASK-260906-1hn93j integration report (RUN-260906-c1d421)

Assignment: integration run for accepted `CR-TASK-260906-1hn93j-2` revision 2
(producer binding: role `developer`, archetype `implementer`).
Board status before and after this run: `integrating` (unchanged).
No trunk commit was made. No board mutation was made except this report's
attachment. The story worktree and branch were left untouched (clean tree).

## Decision: integration withheld

`task-board worktree integrate` was deliberately NOT executed, because
executing it now would land the wrong tree. Evidence below. The branch tip
contains reviewed-and-accepted content PLUS unreviewed post-acceptance work;
the accepted revision's recorded candidate is empty. Landing either operand
loses work or closes the story on an unreviewed tree.

## Evidence

1. Accepted revision (from `task-board worktree status --json`):
   - `CR-TASK-260906-1hn93j-2`, revision 2, state `accepted`, kind `task_delta`
   - `base_oid = e01de3f5731555457c8d3c7de6bec58b8e768f32`
   - `candidate_tree_oid = f3b96cd32716842b02eb09b5c0f627010cfa3678`
   - `repository_delta = empty`, `changed_paths = null`
   - Patch resource `TASK-260906-1hn93j_change-request_rev2.patch` is 0 bytes
     (measured with `wc -c` on the board resource file).
   - `git rev-parse e01de3f^{tree}` returns `f3b96cd...`, i.e. the candidate
     tree equals the tree of its own base commit. The CR snapshot was cut
     after the producer had already committed, so base tree == candidate tree
     and the recorded delta is empty. This is the same base-recording artifact
     the cycle-2 review already documented for rev 1 and rev 2.

2. Branch reality (story branch `task-board/story/STORY-260905-2z9pw4`):
   - Tip `c25d78e` ("Publish the takeover flag and onboarding import ..."),
     exactly one commit past trunk `f39f4a9`, tree clean, signed (ECDSA,
     `git verify-commit HEAD` reports a good signature).
   - `git rev-parse HEAD^{tree}` returns `3fbbb7e569b68901a9fc4d9c3e40403bfb71675e`,
     which differs from the accepted candidate `f3b96cd...`.
   - `git diff e01de3f..c25d78e --stat`: 2 files, 2 insertions, 2 deletions —
     the rework-2 deletion of `, and an explicit takeover` /
     `, and an explicit takeover.` from the two onboarding-trigger lists
     (`protocol/environments.md:1774`, `profiles/manager.md:2285`).
     That change is post-acceptance and unreviewed.
   - `git diff f39f4a9..HEAD --stat`: 4 files, 51 insertions, 15 deletions
     (`CHANGELOG.md`, `cli/curator.md`, `profiles/manager.md`,
     `protocol/environments.md`). This is the real deliverable, and no
     accepted revision records it as a non-empty delta.

3. Trunk (control root `/Users/iv/Developer/ReluxWorks/curator-spec`):
   - `HEAD = f39f4a9`, `status --short --branch` clean (`## main...origin/main`).
   - Trunk has not advanced past the story base, so there is no
     `integration_base_moved` staleness on trunk grounds. The staleness is
     branch-side: tip `c25d78e` != accepted candidate tree `f3b96cd`.

4. What landing rev 2 would do:
   - Per the tracked-spawn contract, `integrate` lands the accepted candidate
     tree, not the branch tip. With `repository_delta = empty` and
     base == candidate, the three-way merge resolves to trunk unchanged, so
     integration would produce NO story commit and still produce the
     board-only commit — closing board state while the 4-file / 51-insertion
     deliverable (plus the 2-line rework-2 fix) never reaches trunk.
   - Landing the tip instead would land unreviewed rework-2 content. Both
     options are wrong, so neither was executed. No integration transaction
     exists for the story (`transaction show` reports none); none was opened.

5. Rework-3 is still pending and was deliberately NOT applied here:
   - `protocol/environments.md:1771-1777` line lengths 75, 78, 42, 79, 72, 48;
     `profiles/manager.md:2282-2287` line lengths 73, 77, 47, 74, 72 —
     the ragged short lines left by the rework-2 deletion.
   - Applying the whitespace re-flow would be a hand commit on the managed
     story branch, which the CR lifecycle forbids (it would widen tip-vs-
     candidate drift further). It belongs in the next producer revision, not
     in this integration run.

## Gates run in this run (standalone processes, honest exit codes)

- `make validate` (with `.temp/venv/bin` on `PATH`, output redirected to a
  file, exit captured via `$?` before any `tail`): exit 0.
  Tail: `validated 60 schemas and 1017 vector files`, `Ran 227 tests ... OK`,
  `go test ./tools/... ok`.
- `make regenerate-check` (same method): exit 0.
  Tail: `go run ./tools/generate-vectors -root .`,
  `git diff --exit-code -- conformance/v1 release/...` with no diff
  (byte-clean).
- Prior piped `make validate | tail` output in this same run is superseded by
  the redirected re-run above; only the redirected runs are quoted as evidence.

## Verification bounds (honestly established)

- CR emptiness: established by reading the board JSON (`repository_delta`,
  `changed_paths`, `candidate_tree_oid`) AND by `wc -c` on the stored patch
  (0 bytes) AND by `git rev-parse e01de3f^{tree}` equaling the recorded
  candidate tree. Three independent methods, all agree.
- Tip-vs-candidate drift: established by `git rev-parse HEAD^{tree}` vs
  `e01de3f^{tree}` (differ) and `git diff e01de3f..HEAD` (exactly the two
  rework-2 hunks). Direct object comparison, not a proxy.
- Trunk state: established by `git -C <control-root> rev-parse HEAD` and
  `status --short --branch` in this run, not from cached worktree-status text.
- Link/coverage bound (unchanged from cycle 2): no committed test asserts
  `cli/curator.md` row text, flag spellings, or the 9.5 enumeration; the
  cycle-2 reviewer's mutant harness (widening + token-preserving rename green,
  broken-link control red) is accepted as attached evidence and was not
  re-run here. Stated bound: gates prove link integrity and schema/vector
  health on the landing tree, not prose correctness.
- Mutants for THIS run: none — this run changed no file, added no gate, and
  opened no transaction, so there is no gate to narrow. The DoD mutant items
  remain covered by the cycle-2 reviewer's attached mutant evidence for the
  prose deliverable, which this report does not re-claim as its own.

## Needed orchestrator routing

1. Publish revision 3 of `CR-TASK-260906-1hn93j` from tip `c25d78e` PLUS the
   rework-3 whitespace re-flow (producer brief `producer-brief-cli-takeover-rework-3.md`),
   so the recorded candidate tree equals the reviewed tree.
2. Fresh review of rev 3 (the rework-2 two-line normative deletion and the
   whitespace-only re-flow both post-date the rev-2 ACCEPT).
3. Then `task-board worktree integrate STORY-260905-2z9pw4 --cr
   TASK-260906-1hn93j --revision 3` from the bound integration run.
   Note the lifecycle ordering constraint: rev 3 must be accepted first, and
   if it is the story's last open leaf the close path is `integrate`, not
   `checkpoint` (`change_request_final_leaf_checkpoint`).

Two follow-up tasks already exist for the carried-forward cycle-2 minors and
are unaffected by this report: `TASK-260906-1xbrz6` (closed-set scope) and
`TASK-260906-3o75d6` (clause editorial). Rework-3 must not be folded into them.

## Files and refs

- Story branch tip: `c25d78e` (one commit past `f39f4a9`, clean, signed).
- Accepted but NOT landed: `CR-TASK-260906-1hn93j-2` rev 2 (empty delta).
- Control-root trunk: `f39f4a9` (clean, unmoved by this run).
- Board element `TASK-260906-1hn93j` left at `integrating`; story
  `STORY-260905-2z9pw4` left at `integrating`.
