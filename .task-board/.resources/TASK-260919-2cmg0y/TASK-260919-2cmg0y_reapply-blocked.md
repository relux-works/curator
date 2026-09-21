# TASK-260919-2cmg0y re-apply blocked — base authority mismatch (worktree status outcome)

Per `2cmg0y-reapply-rev1.md` step 1 (binding): confirm the workspace is clean and the
Story tip descends from `d4fe8347` (`git merge-base --is-ancestor d4fe8347 HEAD`).
The check FAILED, so the run stops here without rebasing by hand and without applying
the rev1 candidate patch.

## Step 1 evidence (worktree: `.temp/STORY-260919-37szes/worktree`, branch `task-board/story/STORY-260919-37szes`)

```
=== git status ===
On branch task-board/story/STORY-260919-37szes
nothing to commit, working tree clean

=== rev-parse ===
HEAD=4f213e77a2115502365652232c0f13c4aa77dc63
d4fe8347=d4fe83475b78babfaa16e2c367363069328fbd86
merge-base(d4fe8347, HEAD)=7fa08e84bc85921dca450fa71e40740cc2052a6a

=== ancestor check ===
git merge-base --is-ancestor d4fe8347 HEAD -> exit 1 (NOT a descendant)

=== HEAD log 3 ===
4f213e77 BUG-260920-2d9gfv: attestation-evidence-unreadable-row-nondeterministic-under-race
e0f52c95 TASK-260920-3ccq6b: draft-literal-lane-env-overrides-and-ssh-config-residuals
680a2f47 TASK-260920-1sbj7o: install-external-refusal-masks-identity-invalid-class

=== main log 5 (d4fe8347) ===
d4fe8347 BUG-260921-30ycv0: wrap the gitops writeBlobs spawn error with sanitized operation context
98f8e633 BUG-260920-3vfwch: diagnose and bound the hosted macOS git spawn EACCES flake in the install test fixtures
c3f9eeea BUG-260916-2f3xbf: declare the Windows pnpm writable-store member and run the real-pnpm cases on Windows
7fa08e84 Record STORY-260910-1cnwwp board state: wave 3 landed
8a77829f STORY-260910-1cnwwp: source-cli-and-executable-conformance

=== d4fe8347..HEAD (7 commits, story-only line) ===
4f213e77 BUG-260920-2d9gfv
e0f52c95 TASK-260920-3ccq6b
680a2f47 TASK-260920-1sbj7o
6a6e2a14 BUG-260920-2eg8nv
9352a87a BUG-260920-3v7x6j
9be96486 BUG-260920-2wyzde
99cb352b BUG-260920-3ukdk4

=== HEAD..d4fe8347 (trunk-only, missing from story tip) ===
d4fe8347 BUG-260921-30ycv0
98f8e633 BUG-260920-3vfwch
c3f9eeea BUG-260916-2f3xbf
```

## Interpretation

- Workspace is clean (`nothing to commit, working tree clean`) — precondition for a replay holds.
- The Story tip does NOT descend from `d4fe8347` (exit 1; merge-base is `7fa08e84`).
  The expected orchestrator replay of Story checkpoints onto the fresh trunk did not
  happen in this worktree: the tip is 7 commits past `7fa08e84` on the old line, while
  trunk added `c3f9eeea`, `98f8e633`, `d4fe8347` after `7fa08e84`.
- Per the binding instruction, no `git apply`, no rebase, no other tree changes were made.
  Working tree left untouched and clean.

## What is needed (orchestrator)

- Replay the Story checkpoint stack onto `d4fe8347` (or respawn this task after the replay),
  then re-issue the re-apply steps 2–4 (apply rev1 candidate patch, verify patch-id,
  re-run fast checks, publish Change Request).
- Rev1 candidate patch and rev1 results.md remain preserved on the task per the
  re-apply brief; nothing here supersedes them.
