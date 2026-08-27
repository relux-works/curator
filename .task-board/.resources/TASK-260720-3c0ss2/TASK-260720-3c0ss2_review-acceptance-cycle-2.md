# TASK-260720-3c0ss2 — review acceptance cycle 2

## Verdict

Accepted. Verdict branch: `done`.

`task-board spawn goal RUN-260729-f11446` reported `Active Goal: none`; this reviewer supplied no `commit_ack`.

## Rework closure

The prior P1 is closed. Before accepting an installed marker as current, `_marker_is_current` now rejects both marker `files` entries at or below any declared build root and any physical declared build-root path observable with `lstat`. Either form forces the ordinary context replacement path, which rewrites a build-root-free tree and marker. Malformed marker file lists fail closed.

The regression matrix covers physical-only contamination, marker-entry-only contamination, and the combined pre-exclusion tree. Every variant proves the contaminated second install is not reported up-to-date, the stale root and marker entry are removed, and a third clean install is up-to-date.

## Independent evidence

- Original reviewer stale-tree probe replay: `second_errors=[]`; result message changed from `up-to-date` to `installed`; `stale_build_root_exists=False`. The legacy probe exited 1 only because its final assertions intentionally encode the old buggy outcome.
- Focused stale-currentness regression: 3 passed, 43 deselected, exit 0.
- Accepted rc.5 plus task-focused pytest: 196 passed, exit 0.
- Strict `python -m mypy`: success, 57 source files, exit 0.
- `git diff --check`: exit 0.
- Final worktree status contains only the producer task paths; test execution introduced no candidate change.
- The prior reviewer already independently accepted the exact framed digest, frozen snapshot boundary, shared fixture context selection, and legacy-hash separation. This rework changes only installed-context currentness and its focused regression, so that positive evidence remains applicable. The producer also records a full suite result of 705 passed and 1 platform skip.
- No Go command was run.

No unresolved finding or stop-the-line boundary remains.