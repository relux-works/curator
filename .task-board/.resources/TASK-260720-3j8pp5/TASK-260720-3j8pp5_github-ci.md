# TASK-260720-3j8pp5 GitHub CI evidence

Date: 2026-07-30

## Candidate

- Commit: `1d28910f5bb276ff58e2a102e06968bd7640abe3`
- Signature: good ECDSA signature for `oparin@me.com`
- PR: https://github.com/ivanopcode/cocoaskills/pull/9
- Workflow: https://github.com/ivanopcode/cocoaskills/actions/runs/30505740935

## Matrix result

- Strict mypy: passed.
- Ubuntu Python 3.11-3.14: passed.
- macOS Python 3.11-3.14: passed.
- Windows Python 3.11-3.14: the task-owned toolchain failures from run
  `30503926948` are closed. No `tests/test_builds_toolchain.py` test failed.
- Each Windows job reports the same eight failures exclusively in
  `tests/test_build_source.py`, with `722 passed, 39 skipped`.

The remaining Windows failures are the already tracked source-identity
physical-key defect owned by `TASK-260720-3c0ss2` / PR #8. They are outside
this task's two-file scope and reproduce identically on all four Windows Python
versions. PR #8 contains the source rework and must be rebased onto the landed
toolchain commit before its final full matrix.

The toolchain task may be accepted only if an independent exact-commit reviewer
confirms that all task-owned Windows failures are closed and the remaining
matrix failures map solely to the separately tracked source task.
