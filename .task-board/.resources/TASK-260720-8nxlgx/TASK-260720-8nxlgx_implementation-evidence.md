# TASK-260720-8nxlgx implementation evidence

## Provenance and scope

- Task: `TASK-260720-8nxlgx` — Implement protected Windows build cache.
- Clean accepted base SHA: `15860e3f309888845b9271a257fb95f7c2825b56`.
- Accepted dependency: `TASK-260720-2jfnz6`, signed commit
  `540af8ef` (verified before implementation).
- Worktree:
  `/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-8nxlgx/worktree`
- Branch: `task/TASK-260720-8nxlgx-protected-windows-build-cache`.
- Product scope:
  - `src/csk/builds/cache_windows.py` (new)
  - `src/csk/builds/cache.py` (Windows backend selection)
  - `tests/test_build_cache_windows.py` (new)
- No `uv.lock` was created. No file was staged or committed.

## Implemented security boundary

- Uses a protected DACL with exact trusted principals: the current manager SID,
  Local SYSTEM, and Builtin Administrators.
- Explicitly assigns the current manager SID as owner; an elevated Windows token
  otherwise creates objects owned by Builtin Administrators.
- Creates and seals a private same-volume staging tree before a non-replacing,
  write-through `MoveFileExW` publication.
- Opens objects with `CreateFileW` using
  `FILE_FLAG_OPEN_REPARSE_POINT`/backup semantics, checks object kind and reparse
  metadata, obtains final handle paths, and proves stable volume/file IDs before
  admitting a hit.
- Requires direct-child physical containment at every path component and rejects
  reparse escapes, alternate data streams, special objects, wrong ownership,
  DACL drift, and link counts other than one.
- Reopens receipt and artifact paths after validation and compares identities,
  preventing replacement/containment races from exposing candidate bytes.
- Verifies the exact receipt schema, input identity, artifact relative path, SHA-256,
  and size. Lookup performs no repair or mutation.
- Preserves POSIX-equivalent identical-winner, conflicting-winner, immutable-entry,
  dry-run, and quarantine behavior.
- Disables persistent reuse when the protected boundary cannot be established or
  verified.
- Keeps Win32 DLL access lazy so the module imports safely on non-Windows hosts.

## Test coverage

The 16 native Windows cases cover:

- backend selection, layout, publication, lookup, identical reuse, and immutability;
- exact owner/DACL profiles and manager-home permissive/owner drift rejection;
- receipt and artifact DACL drift;
- receipt/artifact hard links, source hard links, and guard-link defenses;
- receipt bytes/hash and artifact path/hash/size binding;
- special pre-existing boundaries and junction/reparse escapes;
- concurrent identical publication and conflicting winners;
- live artifact replacement while lookup has a pinned handle;
- quarantine and read-only miss behavior;
- import safety on non-Windows hosts.

## Green validation gates

All commands below were run directly as standalone processes.

1. `uvx ruff check src/csk/builds/cache.py src/csk/builds/cache_windows.py tests/test_build_cache_windows.py`
   - Exit `0`: `All checks passed!`
2. `python -m mypy`
   - Exit `0`: `Success: no issues found in 65 source files`
3. `python -m pytest -q tests/test_build_cache_posix.py tests/test_build_cache_windows.py`
   - Exit `0`: `42 passed, 16 skipped in 0.51s`
4. `python -m pytest -q`
   - Exit `0`: `853 passed, 74 skipped in 79.96s`
5. `python -m compileall -q src`
   - Exit `0`
6. `python -m build --outdir .temp/TASK-260720-8nxlgx/dist`
   - Exit `0`; sdist and wheel built successfully.
7. `python -m twine check .temp/TASK-260720-8nxlgx/dist/*`
   - Exit `0`; wheel and sdist both passed.
8. `git diff --check`
   - Exit `0`
9. `test ! -e uv.lock`
   - Exit `0`
10. Native Windows:
    `.venv\Scripts\python -m pytest -q tests\test_build_cache_windows.py`
    - Exit `0`: `16 passed in 9.79s`
11. Native Windows task-module typing:
    `.venv\Scripts\python -m mypy --platform win32 --follow-imports skip src\csk\builds\cache_windows.py`
    - Exit `0`: `Success: no issues found in 1 source file`
12. Native Windows full suite with process-only test prerequisite:
    `set PSExecutionPolicyPreference=Bypass&& .venv\Scripts\python -m pytest -q`
    - Exit `0`: `783 passed, 144 skipped in 306.56s`

The native Windows host was Windows 10 build 19045 with Python 3.14.4.
GitHub `windows-latest` remains the publication gate, while this native run
provides task-scoped platform evidence.

## Non-green attempts and environment anomalies

These results are retained as failures rather than represented as passes:

1. Native snapshot setup:
   `pip install -e .[dev]`
   - Exit `1`.
   - Expected setup-only failure: the transferred native test snapshot intentionally
     omitted `.git`, so setuptools-scm could not infer a version. Declared test,
     typing, and packaging dependencies were then installed directly.
2. First native full pytest without a process execution-policy preference:
   `.venv\Scripts\python -m pytest -q`
   - Exit `1`: `1 failed, 782 passed, 144 skipped in 307.54s`.
   - The only failure was the unrelated baseline
     `tests/test_shell_init.py::test_powershell_hook_activates_and_restores_on_every_prompt`;
     the host policy blocked its temporary PowerShell script.
   - The focused baseline test passed with a process-only
     `PSExecutionPolicyPreference=Bypass` (exit `0`, `1 passed in 3.46s`), followed
     by the green full native suite above. No machine or repository policy changed.
3. Native Windows full-repository `python -m mypy`
   - Exit `1` with 51 pre-existing Windows platform/stub errors in existing modules,
     including `_windows.py`, locking, source, config, transactions, and
     `cache_posix.py`.
   - The new backend has no remaining errors and passes the isolated strict
     Windows-platform command above; the required full strict local command passes
     all 65 source files.

## Durable platform findings

- Elevated Windows creation does not reliably imply current-user ownership; secure
  creation must explicitly set the manager owner SID.
- Sealed cleanup must update only the DACL. The sealed manager profile has implicit
  `WRITE_DAC` but intentionally does not retain `WRITE_OWNER`.
- Windows prevents renaming a whole entry while descendant handles are pinned even
  with delete sharing. The valid containment-race test is a path-level artifact
  replacement: the identity mismatch must force a miss before any path or bytes are
  returned.
