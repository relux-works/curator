# Rework brief for BUG-260810-2oxt8b

The first review requested a supported repair path for legacy raw statuses.

That path is now delivered and merged in `relux-works/skill-project-management` PR #4 at merge commit `8dc0b71490214fe5ead6bf9cfde9574df084fd91`; installed `task-board` reports `0.24.3-3-g8dc0b71`.

The orchestrator used the public mutation to normalize:

- `EPIC-260712-d77d32`: `in-progress` -> `backlog`
- `STORY-260713-72b914`: `\nin-progress` -> `backlog`
- `EPIC-260713-c12fbe`: `todo` -> `backlog`

It also removed the invalid prose-only `blockedBy` placeholder from `EPIC-260712-d77d32`, attached PR #5 delivery evidence, completed the remaining Story checklist item, and moved `STORY-260713-72b914` to `done`; its parent Epic aggregated to `done`. The dedicated website remains `backlog`. Adapter planning remains a separate backlog Epic with Swift/Kotlin/C baseline and a discovery Story for evidence-based additions.

Producer assignment:

1. Independently inspect these board states and attached evidence.
2. Run strict board validation and record the result.
3. Update or add a task-scoped outcome artifact describing the completed rework.
4. Complete only the applicable remaining producer checklist items and hand the Bug to `to-review`.
5. Do not implement the website or language adapters, and do not alter their backlog scope.
