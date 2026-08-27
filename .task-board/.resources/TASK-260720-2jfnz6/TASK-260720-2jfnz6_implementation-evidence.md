# TASK-260720-2jfnz6 implementation evidence

Date: 2026-07-30

## Provenance

- Product repository: `/Users/iv/Developer/intranet/cocoaskills`
- Task worktree:
  `/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2jfnz6/worktree`
- Task branch:
  `task/TASK-260720-2jfnz6-protected-posix-build-cache`
- Recorded base and current HEAD:
  `495ad021847529ce5a544dba415ca2fe19949539`
- Dependency handoff `TASK-260720-2dnqw2` was present at the recorded base before
  worktree creation.
- Conformance root:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1`
- Conformance manifest SHA-256:
  `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`

## Delivered scope

- `src/csk/builds/cache.py`
  - Defines the platform-neutral cache protocol, immutable request/result
    records, stable statuses/errors, caller-held mutation-lock witness, and
    native backend factory.
- `src/csk/builds/cache_posix.py`
  - Stores csk build entries only below
    `<manager-home>/builds/go-v1/<hex-key>`.
  - Uses rooted file-descriptor traversal with `O_NOFOLLOW`, effective-UID,
    exact mode, type, ownership, single-link, exact-content, canonical-receipt,
    complete-input, derived-path, size, and SHA-256 validation before reuse.
  - Keeps inspection read-only and stages below `.builds-staging`, outside the
    live namespace.
  - Selects a complete directory winner with native Darwin
    `renameatx_np(RENAME_EXCL)` or Linux `renameat2(RENAME_NOREPLACE)`, then
    seals the winner for ordinary immutable use.
  - Quarantines corrupt/untrusted state under the caller-held guard without
    adopting it. Darwin read-only directories receive owner control only
    through verified no-follow state for the atomic move, then regain their
    original mode in quarantine.
  - Reuses only byte-identical concurrent winners and raises
    `CacheConflictError` for differing bytes under one logical key.
- `tests/test_build_cache_posix.py`
  - 33 POSIX-focused tests cover private layout/modes, effective UID, hard
    links, special files and symlinks at each boundary, parser ordering,
    canonical receipt/input/path/size/hash checks, dry-run immutability,
    untrusted and corrupt rebuilds, invalid publication read-only behavior,
    staging revalidation, independent concurrent publishers, conflict
    detection, and locked quarantine for later GC.

Windows ACLs, installer transactions, status integration, and GC are
intentionally outside this task.

## Final green gates

The isolated task environment used Python 3.14.4, pytest 9.1.1, mypy 2.3.0,
build 1.5.0, and twine 7.0.0.

1. POSIX-focused pytest:

   `PATH=/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2jfnz6/venv/bin:$PATH python -m pytest -q --basetemp=/private/tmp/TASK-260720-2jfnz6-pytest.jGUoAb tests/test_build_cache_posix.py`

   Exit 0: `33 passed in 0.40s`.

2. Full repository pytest with accepted conformance corpus:

   `PATH=/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2jfnz6/venv/bin:$PATH CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 python -m pytest -q --basetemp=/private/tmp/TASK-260720-2jfnz6-pytest.jGUoAb`

   Exit 0: `882 passed, 6 skipped in 71.47s`.

3. Strict typing:

   `PATH=/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2jfnz6/venv/bin:$PATH python -m mypy`

   Exit 0: `Success: no issues found in 63 source files`.

4. Whitespace/static validation:

   `GIT_INDEX_FILE=/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2jfnz6/lint.index git diff --check`

   Exit 0 with no findings. The alternate index makes the three new,
   intentionally untracked task files visible to `git diff --check` without
   changing the real index. The repository has no separately configured
   formatter or style linter.

   `PATH=/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2jfnz6/venv/bin:$PATH python -m compileall -q src/csk`

   Exit 0 with no findings.

5. Distribution build:

   `PATH=/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2jfnz6/venv/bin:$PATH python -m build`

   Exit 0; built
   `cocoaskills-0.12.6.dev6+g495ad0218.tar.gz` and
   `cocoaskills-0.12.6.dev6+g495ad0218-py3-none-any.whl`, including both new
   cache modules.

6. Distribution metadata:

   `PATH=/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2jfnz6/venv/bin:$PATH python -m twine check dist/cocoaskills-0.12.6.dev6+g495ad0218-py3-none-any.whl dist/cocoaskills-0.12.6.dev6+g495ad0218.tar.gz`

   Exit 0: both artifacts `PASSED`.

## Red and anomaly evidence

- The required pre-implementation focused pytest command exited 2 because
  `csk.builds.cache` did not yet exist. This was the expected red gate.
- A shared-environment `python -m mypy` attempt exited 1: that environment did
  not generate this worktree's `_version.py`, and it also found the initial
  `_effective_uid` return annotation. The task-specific editable environment
  removed the environment mismatch; the annotation was fixed before the
  strict green gates above.
- Darwin adversarial recovery iterations exited 1 while proving that
  `RENAME_EXCL` and ordinary directory rename both reject non-owner-writable
  source directories. Tests first caught mode `0550`, then mode `0000`. The
  implementation now performs guarded, verified, no-follow temporary owner
  unlock and mode restoration; the final focused suite proves `0700`, `0550`,
  `0000`, and sealed `0500` cases.
- One full-suite attempt exited 1 with `876 passed, 6 skipped, 1 failed`
  because its `--basetemp` was incorrectly placed under the Git worktree. The
  unrelated CLI test therefore saw a Git ancestor. The exact failed test
  passed in `/private/tmp` with exit 0, and the corrected full gate above
  passed with exit 0.

## Review state

The real worktree index remains untouched. `git status --short` reports only:

```text
?? src/csk/builds/cache.py
?? src/csk/builds/cache_posix.py
?? tests/test_build_cache_posix.py
```

The implementation is ready for review.
