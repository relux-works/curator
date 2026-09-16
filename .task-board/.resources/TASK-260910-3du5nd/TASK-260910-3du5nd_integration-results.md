# TASK-260910-3du5nd integration results

Command (control root `/Users/administrator/Developer/ReluxWorks/curator/curator-spec`):

```bash
task-board worktree complete STORY-260916-2txa8v --cr TASK-260910-3du5nd --revision 2 --landed-commit 8ba9c235ec5be00d52378479516c82386fd0c178
```

Exit code: 0

Full output:

```text
STORY-260916-2txa8v  cleanup_pending
  code landed:  8ba9c235ec5be00d52378479516c82386fd0c178 (proven on the code repository's protected default)
  board commit: e40d00f713c30fc9203644807c7a8c0a9aaf66ee
  board published to refs/heads/main in /Users/administrator/Developer/ReluxWorks/curator/curator
  note: safe cleanup is now eligible; `worktree cleanup` removes the workspace and branch only after exact commit ancestry, Story done, a committed board record, a clean workspace, and no active lease or RUN
```
