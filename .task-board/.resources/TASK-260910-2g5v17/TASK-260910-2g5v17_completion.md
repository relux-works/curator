# TASK-260910-2g5v17 completion (integration run RUN-260918-ac7956)

Revision 3 = exact landed tree. `worktree complete` proved the landing and published the board record.

## Pre-complete state
```
$ task-board worktree obligations
ELEMENT                REV  STATE     NEEDS       AGE      BOARD        SCOPE
TASK-260910-2g5v17     3    accepted  checkpoint  1m       integrating  STORY-260910-stz5f0
EXIT:0
$ task-board q 'get(TASK-260910-2g5v17) { status }' -> {"status":"integrating"}
$ task-board q 'get(STORY-260910-stz5f0) { status children }' -> {"children":["TASK-260910-2g5v17"],"status":"integrating"}
```

## Landing proof (quoted from this run)
```
$ git -C <control-root> rev-parse bf5cac1200cfa39dff0a6b0449072ff5b22f124d^{tree}
9b7d33115bd3ffe44c34c5a340de72aaab06d0cb
$ git -C <control-root> log --oneline -5
bf5cac1 Persist a per-upstream high-water and refuse rollback bundles on import (P4)
fb86420 Protect the signing key with an optional passphrase behind a key-provider seam (R6)
c7ef32c Compare live state against the operator checkpoint at startup (R3/P2)
...
$ git verify-commit bf5cac1... -> Good "git" signature for bot@relux.works with ED25519 key
$ git rev-parse origin/main -> bf5cac1200cfa39dff0a6b0449072ff5b22f124d
$ worktree snapshot tree (temporary index, remote-gate.sh method) -> 9b7d33115bd3ffe44c34c5a340de72aaab06d0cb
$ ls TASK-260910-2g5v17_results.md in worktree -> No such file or directory
```

## Complete transaction
```
$ task-board worktree complete STORY-260910-stz5f0 --cr TASK-260910-2g5v17 --revision 3 --landed-commit bf5cac1200cfa39dff0a6b0449072ff5b22f124d
(from /Users/administrator/Developer/ReluxWorks/curator/curator-skill-registry)
STORY-260910-stz5f0  cleanup_pending
  code landed:  bf5cac1200cfa39dff0a6b0449072ff5b22f124d (proven on the code repository's protected default)
  board commit: d69863d4bd286db56c65647526255d5227ac8850
  board published to refs/heads/main in /Users/administrator/Developer/ReluxWorks/curator/curator
  note: safe cleanup is now eligible; `worktree cleanup` removes the workspace and branch only after exact commit ancestry, Story done, a committed board record, a clean workspace, and no active lease or RUN
EXIT:0
```
No resume needed; single invocation completed. No repair, reset, manual commit, or push performed.

## Post-complete state
```
$ task-board worktree obligations -> No Change Request revision is waiting on a next step without a live run.
$ task-board q 'get(STORY-260910-stz5f0) { status children }' -> {"children":["TASK-260910-2g5v17"],"status":"done"}
$ task-board q 'get(TASK-260910-2g5v17) { status }' -> {"status":"done"}
$ git -C ../curator log --oneline -2
d69863d Record STORY-260910-stz5f0 board state
b56089e Record STORY-260910-9484i4 board state
$ git -C ../curator show -s --format='%H %s' d69863d... -> d69863d4bd286db56c65647526255d5227ac8850 Record STORY-260910-stz5f0 board state
```

Board commit id: `d69863d4bd286db56c65647526255d5227ac8850`
Landed code commit: `bf5cac1200cfa39dff0a6b0449072ff5b22f124d` (tree `9b7d33115bd3ffe44c34c5a340de72aaab06d0cb`)
