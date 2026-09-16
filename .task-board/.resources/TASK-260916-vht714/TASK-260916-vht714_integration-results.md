# TASK-260916-vht714 integration results

Command (control root /Users/administrator/Developer/ReluxWorks/curator/curator-spec):

```text
task-board worktree complete STORY-260916-1on1d2 --cr TASK-260916-vht714 --revision 1 --landed-commit 871d11bcdfd240a6260d0722503bdd1642a8fce8
```

Exit code: 0

Full output:

```text
STORY-260916-1on1d2  cleanup_pending
  code landed:  871d11bcdfd240a6260d0722503bdd1642a8fce8 (proven on the code repository's protected default)
  board commit: c1aa0d2f4d18f2838282ff255fb81d4543b959d7
  board published to refs/heads/main in /Users/administrator/Developer/ReluxWorks/curator/curator
  note: safe cleanup is now eligible; `worktree cleanup` removes the workspace and branch only after exact commit ancestry, Story done, a committed board record, a clean workspace, and no active lease or RUN
```
