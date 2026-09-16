# TASK-260916-2fu85y integration results

Command (foreground, control root `/Users/administrator/Developer/ReluxWorks/curator/curator-spec`):

```text
task-board worktree complete STORY-260916-1on1d2 --cr TASK-260916-2fu85y --revision 2 --landed-commit 0f8f9073390fcef7296446b8b3bd5d5d78e37e3f
```

Exit code: 0

Full output:

```text
STORY-260916-1on1d2  cleanup_pending
  code landed:  0f8f9073390fcef7296446b8b3bd5d5d78e37e3f (proven on the code repository's protected default)
  board commit: 81fd85b2f721a81e4bd00f499834967f3779be73
  board published to refs/heads/main in /Users/administrator/Developer/ReluxWorks/curator/curator
  note: safe cleanup is now eligible; `worktree cleanup` removes the workspace and branch only after exact commit ancestry, Story done, a committed board record, a clean workspace, and no active lease or RUN
```
