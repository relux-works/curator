# TASK-260720-17llva review verdict cycle 2

## Verdict

Changes requested; route to `to-dev`.

## Findings

1. P1 legacy-currentness and GC mismatch: `profiles/manager.md:574-576` says GC marks entries referenced by every existing consumer, valid marker v2, and journals. Protocol Core section 10 at `protocol/core.md:575-578` requires managers to read marker schemas 1 and 2 and permits a valid marker-v1 installation for schema 1 through 5 to remain current. The new mark set therefore omits valid marker-v1 references. After the grace period, locked GC can classify runtime or snapshot entries used only by a still-current marker-v1 installation as unreferenced and sweep them. Required rework: include every supported valid marker schema in the mark traversal, while limiting logical compiled-artifact references to marker v2 entries that actually carry them. Preserve the existing fail-safe behavior for unreadable markers.

2. P2 stale normative cross-reference: `profiles/manager.md:463` says the adapter ledger is in Protocol Core section 10, but the accepted core now defines install markers in section 10 and the adapter ledger in section 11 at `protocol/core.md:619`. Required rework: change the manager-profile reference to Protocol Core section 11.

## Lifecycle rework verification

The prior recovery-ordering finding is resolved. The profile now stages and verifies every miss before home-lock recovery or shared mutation, restarts affected work after recovery drift, and preserves the operation-entry installation, consumers, and live caches on private build failure. The five fixed Go argv forms, clean environment, closed process graph, source gates, go-list rejection rules, compiler-free dry-run, protected publication, one journal, reverse rollback, consumer-last commit, repair, and locked GC are otherwise covered.

Independent commands passed on 2026-07-20:

- `PATH=.temp/TASK-260720-1nvomm/venv/bin:$PATH python3 tools/validate.py`: validated 30 schemas and 93 vector files.
- `PATH=.temp/TASK-260720-1nvomm/venv/bin:$PATH make validate`: 8 Python tests and `go test ./tools/...` passed.
- `git diff --check -- profiles/manager.md`: passed.
