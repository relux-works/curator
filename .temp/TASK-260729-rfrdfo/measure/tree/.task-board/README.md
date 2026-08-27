# Task board

The in-repo task board for Curator. Work items are directories:

```text
.task-board/
  EPIC-<yymmdd>-<id>_<slug>/
    README.md            # description, scope, acceptance criteria
    progress.md          # status, assignee, blockers, checklist
    STORY-<yymmdd>-<id>_<slug>/
      TASK-<yymmdd>-<id>_<slug>/
  .resources/<ID>/       # artifacts attached to an item
```

Rules:

- Every change starts from a task on this board; status moves in progress.md (todo, in-progress, review, done).
- Epics map one to one onto the phases of docs/implementation-plan.md; the dependency order lives there.
- Research and results are written into .resources/<ID>/ so context survives sessions.
