# Checkpoint — TASK-260917-16l2md (rc.12 pin-promotion union, CR revision 5 accepted)

Accepted revision 5 checkpointed per `TASK-260917-16l2md_checkpoint-rev5.md`. This leaf is NOT the story's final leaf (qualification `TASK-260917-2ecpjv` follows), so checkpointed, not integrated. No code changes, no handoff, no `done`.

- Checkpoint commit: `73fc8a4b264a4af6174cc1a9e487c82eedac8f2c` on `task-board/story/STORY-260917-3w3lvj`
- Board status after checkpoint: `integrating` (unchanged, per checkpoint output)
- Obligation row for TASK-260917-16l2md: present before, gone after (other leaves' rows untouched)

## 1. Obligations before checkpoint

```
$ task-board worktree obligations
ELEMENT                REV  STATE     NEEDS       AGE      BOARD        SCOPE
TASK-260916-2bwfli     3    accepted  checkpoint  1d1h     done         STORY-260910-1bhj0g
TASK-260910-16k7xy     6    accepted  checkpoint  1d       done         STORY-260910-24nyb1
TASK-260916-hxr6qv     4    accepted  checkpoint  17h49m   done         STORY-260916-v58b5y
TASK-260910-19w2aj     11   accepted  checkpoint  9h45m    done         STORY-260910-3vxe3y
TASK-260910-3ungjy     5    accepted  checkpoint  4h25m    done         STORY-260910-2awkzu
TASK-260910-2qtiho     3    accepted  checkpoint  26m      integrating  STORY-260910-2qmrb8
TASK-260910-1tvf2t     2    accepted  checkpoint  18m      integrating  STORY-260910-6bo7ej
TASK-260917-16l2md     5    accepted  checkpoint  1m       integrating  STORY-260917-3w3lvj
```

Row confirmed: `TASK-260917-16l2md  5  accepted  checkpoint`. Exit 0.

## 2. Checkpoint

```
$ task-board worktree checkpoint TASK-260917-16l2md
TASK-260917-16l2md: checkpointed as 73fc8a4b264a4af6174cc1a9e487c82eedac8f2c on task-board/story/STORY-260917-3w3lvj
TASK-260917-16l2md: status integrating
```

Exit 0. No typed error.

## 3. Verification after checkpoint

```
$ git -C <story worktree> log --oneline -3
73fc8a4 TASK-260917-16l2md: TASK-260917-16l2md: promote-spec-pin-rc12-with-wave1-manager-union
3c45d4b Record STORY-260910-1py4f3 board state
9b185d2 Record campaign board state: wave 2 complete (lock story landed)

$ task-board worktree obligations
ELEMENT                REV  STATE     NEEDS       AGE      BOARD        SCOPE
TASK-260916-2bwfli     3    accepted  checkpoint  1d1h     done         STORY-260910-1bhj0g
TASK-260910-16k7xy     6    accepted  checkpoint  1d       done         STORY-260910-24nyb1
TASK-260916-hxr6qv     4    accepted  checkpoint  17h49m   done         STORY-260916-v58b5y
TASK-260910-19w2aj     11   accepted  checkpoint  9h45m    done         STORY-260910-3vxe3y
TASK-260910-3ungjy     5    accepted  checkpoint  4h26m    done         STORY-260910-2awkzu
TASK-260910-2qtiho     3    accepted  checkpoint  26m      integrating  STORY-260910-2qmrb8
TASK-260910-1tvf2t     2    accepted  checkpoint  19m      integrating  STORY-260910-6bo7ej

$ git -C <story worktree> status --short | head -20
(empty — worktree clean)

$ git -C <story worktree> rev-parse HEAD
73fc8a4b264a4af6174cc1a9e487c82eedac8f2c
```

Leaf stays `integrating`; obligation row gone; worktree clean. All commands exit 0.
