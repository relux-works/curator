# TASK-260720-z2z795 post-migration rework outcome

Run: `RUN-260730-9c8f83`  
Date: 2026-07-30  
Role: developer  
Worktree: `/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/worktree`

## Review findings addressed

1. Physical lock identity
   - macOS case and Unicode-normalization aliases now resolve to the stored
     physical path spelling before project/home lock identity is derived;
   - Windows identity is derived one component at a time using the parent
     directory's case-sensitivity flag, including deterministic missing-path
     suffixes;
   - project, build, and manager-home process ordering uses the same canonical
     manager-home identity.

2. Journal filename binding
   - every `.json` and `.json.delete` record is rejected when the filename
     transaction ID differs from the embedded ID;
   - the check occurs immediately after decoding and before validation,
     recovery, target mutation, sidecar mutation, or journal removal.

3. Read-only target durability
   - staging uses writable construction modes and journaled reverse-order mode
     finalization to preserve exact `0444` files and `0555` trees;
   - sidecar cleanup journals top-down writable transitions before bottom-up
     removal, so commit cleanup and reverse rollback are crash-recoverable;
   - staging and cleanup recovery validate contents and the exact modes allowed
     by durable progress, rejecting foreign mode changes;
   - Windows regular-file durability falls back from a write handle to a read
     handle for read-only files.

## Scope and provenance

- Original accepted base:
  `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`
- Reused task branch HEAD and matching remote task branch:
  `5f8bfbbd6e0c3a00dd0a3de0295e74f05990a0ec`
- Clean canonical `main` and local `origin/main`:
  `97a0ed870782b48eebc5a9c25a9cfa8fea5ff245`
- Modified task-owned paths only:
  - `src/csk/locking.py`
  - `src/csk/transactions.py`
  - `tests/test_locking.py`
  - `tests/test_transactions.py`
- No files were staged or committed. No SSH, push, tag, release, pin, compiler
  policy, installer policy, or Go UX work was performed.

## Current-byte validation

- New post-migration regressions: `18 passed` (exit 0).
- Preserved migration regressions: `14 passed` (exit 0).
- Corrected prior lock-integrity set: `14 passed` (exit 0).
- Earlier reviewer regressions: `14 passed` (exit 0).
- Contract-targeted regressions: `13 passed` (exit 0).
- Focused pytest: `105 passed, 1 skipped` (exit 0).
- Full pytest: `675 passed, 20 skipped` (exit 0).
- Strict mypy: no issues in 57 source files (exit 0).
- Ruff lint and format check: green (exit 0 each).
- Package build: sdist and wheel built (exit 0).
- `git diff --check`: clean (exit 0).

The focused skip is the real Windows file/tree flush test. macOS case and
Unicode alias contention ran on the host. Windows per-directory
case-sensitivity and read-only flush fallback have deterministic routing tests,
but no current-unstaged-byte Windows runner result is claimed.

See `TASK-260720-z2z795_post-migration-rework-validation.log` for exact commands,
non-zero diagnostic invocations, exit codes, hashes, and build artifacts.
