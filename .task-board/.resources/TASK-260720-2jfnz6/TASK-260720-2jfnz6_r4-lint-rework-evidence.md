# TASK-260720-2jfnz6 R4 lint rework evidence

Date: 2026-07-30
Developer run: `RUN-260730-ccb38d`

## Scope and provenance

- Product worktree:
  `/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2jfnz6/worktree`
- Branch: `task/TASK-260720-2jfnz6-protected-posix-build-cache`
- Recorded base and merge base: `495ad021847529ce5a544dba415ca2fe19949539`
- Task HEAD before this unstaged rework:
  `0d6ad16fce35c1bd8854511e13766cd236908e3b`
- Dependency `TASK-260720-2dnqw2` remains `done`.
- Reviewer finding R4-1 was reproduced exactly and closed by deleting the
  single extra blank line before `_SHA256_IDENTITY` in
  `src/csk/builds/cache.py`.
- The accepted R3 functional/security bytes are unchanged:
  - `src/csk/builds/cache_posix.py` SHA-256:
    `be370f6f4b63d355d7082b31d0f753ddcd8ad89449588130ae4edc2140d1aa36`
  - `tests/test_build_cache_posix.py` SHA-256:
    `c7724b619a8074d36f44d17aacea1124f0627d5130bcc8532e378dffe3f1d13c`
- Revised `src/csk/builds/cache.py` SHA-256:
  `a191769d14dca1b48d04e96b4f3c877b764ed98d1fc4935040932dd0015b87ee`
- The final product delta is exactly the three task-owned files above. The
  index, commit, branch, remote, and PR were not mutated.

## Command ledger

Every validation command ran as a standalone process.

1. Expected-red task-wide lint before the edit:

   `uvx ruff check src/csk/builds/cache.py src/csk/builds/cache_posix.py tests/test_build_cache_posix.py`

   Exit 1, expected failure: Ruff `I001` identified the reviewer's one extra
   blank line in `cache.py`.

2. Task-wide lint after the edit:

   `uvx ruff check src/csk/builds/cache.py src/csk/builds/cache_posix.py tests/test_build_cache_posix.py`

   Exit 0: `All checks passed!`

3. POSIX-focused pytest:

   `.venv/bin/python -m pytest -q --basetemp=/private/tmp/TASK-260720-2jfnz6-r4-focused-20260730 tests/test_build_cache_posix.py`

   Exit 0: `41 passed, 1 skipped in 0.51s`. The skip is the native-Linux-only
   `O_PATH`/`fchmodat2` test already covered by the exact-hash R3 Linux matrix.

4. Strict typing:

   `.venv/bin/python -m mypy`

   Exit 0: `Success: no issues found in 63 source files`.

5. Full repository regression suite:

   `CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 uv run --extra dev python -m pytest -q --basetemp=/private/tmp/TASK-260720-2jfnz6-r4-full-20260730`

   Exit 0: `890 passed, 7 skipped in 85.17s`.

6. Compile and diff hygiene:

   - `.venv/bin/python -m compileall -q src/csk`: exit 0.
   - `git diff --check`: exit 0.
   - `test ! -e uv.lock`: exit 0.

The full-suite `uv run` generated an untracked `uv.lock`; it was moved
recoverably into the task's ignored `.temp` area before the final scope check.
Distribution build/Twine and the native Linux matrix were not repeated because
the cooperative run directive bounded R4 to the one lint line and explicitly
requested no redundant native rerun. R3 evidence retains green build/Twine and
Linux 3.11-3.14 results for the unchanged functional and test bytes.

The R4 rework is ready for review.
