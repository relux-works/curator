# TASK-260919-2cmg0y re-apply STOPPED at step 1 (base mismatch)

Per 2cmg0y-reapply-rev1.md step 1, the re-apply must stop when the Story tip does not descend from d4fe8347. No patch applied, no rebase attempted, tree left clean.

## Worktree status (verbatim)

```
On branch task-board/story/STORY-260919-37szes
nothing to commit, working tree clean
```

## Ancestry evidence

- `git merge-base --is-ancestor d4fe8347 HEAD` exit=1 (NOT descendant; required by step 1)
- HEAD=4f213e77a2115502365652232c0f13c4aa77dc63 (BUG-260920-2d9gfv tip)
- main=d4fe83475b78babfaa16e2c367363069328fbd86 (BUG-260921-30ycv0)
- `git merge-base HEAD d4fe8347` = 7fa08e84bc85921dca450fa71e40740cc2052a6a
- `git merge-base --is-ancestor HEAD d4fe8347` exit=1 (diverged both ways)
- `git merge-base --is-ancestor c3f9eeea HEAD` exit=1 (old base c3f9eeea also not an ancestor; branch forked at 7fa08e84)
- HEAD log (8): 4f213e77, e0f52c95, 680a2f47, 6a6e2a14, 9352a87a, 9be96486, 99cb352b, 7fa08e84
- main log (3): d4fe8347, 98f8e633, c3f9eeea

## Patch precondition (unapplied)

- `/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/resources/TASK-260919-2cmg0y/TASK-260919-2cmg0y_rev1-candidate.patch` present (27139 bytes), not applied per stop rule.
- rev1 results.md snapshot remains on the task; no results.md changes made.

## Needed from orchestrator

Replay the Story checkpoints onto fresh trunk d4fe8347 (or re-spawn with the replay done), then re-issue the re-apply steps 2-4.
