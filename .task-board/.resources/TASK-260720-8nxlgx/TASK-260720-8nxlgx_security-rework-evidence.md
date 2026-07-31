# TASK-260720-8nxlgx security rework evidence

## Scope and provenance

- Task: `TASK-260720-8nxlgx` — Implement protected Windows build cache.
- Reviewer source: `TASK-260720-8nxlgx_review-verdict_RUN-260730-e46794.md`.
- Rework run: `RUN-260730-72d4a3`.
- Worktree:
  `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-8nxlgx/worktree`.
- Branch: `task/TASK-260720-8nxlgx-protected-windows-build-cache`.
- Base/HEAD: `15860e3f309888845b9271a257fb95f7c2825b56`.
- The preserved worktree and accepted dependency handoff were already recorded
  before this rework. No branch, index, commit, remote, or PR mutation occurred.

Final product hashes, independently matched on native Windows:

```text
28b977d4bba4e2c75ab646344a95a0cb338cc2420aa332bd3195dffb1482e6f6  src/csk/builds/cache.py
43a413ab71944976f20009016a176d74853cd1262230ed618f67f159f706cd52  src/csk/builds/cache_windows.py
611aefdc0a4d2574d85b14b426e56ed0d090fecdde487f5e4d6f646c15874111  tests/test_build_cache_windows.py
```

## Reviewer findings closed

1. Receipt and artifact lookup now revalidate `NumberOfLinks == 1` at
   the final context-manager return boundary. The check runs on both the
   retained handle and the independently reopened selected path after
   identity, DACL, and attribute validation.
2. Publication source validation now performs the same current link-count
   validation on the retained source handle and selected source path at each
   final source-stability boundary. A late hard link raises
   `cache_publication_invalid`; no entry is accepted.
3. Manager-home validation now rejects an untrusted mutating allow ACE if it
   is effective on the manager home or inheritable by files or directories.
   `INHERIT_ONLY_ACE` no longer bypasses the mutation boundary.
4. Inherit-only manager ACEs no longer count toward the manager principal's
   effective manager-home grant.

The focused suite adds deterministic native regressions for late receipt,
artifact, and publication-source hard links; `(OI)(CI)(IO)`, `(OI)(IO)`, and
`(CI)(IO)` Everyone mutation ACEs; and synthetic proof that an inherit-only
manager ACE is not an effective home grant.

## Validation ledger

All green gates below ran as standalone commands and returned exit code 0.

### Native Windows

The remote files matched the hashes above before both final pytest runs.

- `python -m pytest -q tests\test_build_cache_windows.py`
  — 23 passed in 2.84s.
- `python -m pytest -q` with process-only
  `PSExecutionPolicyPreference=Bypass`
  — 790 passed, 144 skipped in 294.70s.
- `python -m mypy --strict --platform win32 --follow-imports skip
  src\csk\builds\cache_windows.py`
  — success, one source file.

### macOS host / non-Windows import surface

- Explicit import-safe test
  `python -m pytest -q
  tests/test_build_cache_windows.py::test_windows_backend_module_is_import_safe_on_every_host`
  — 1 passed.
- Focused POSIX plus Windows contract
  `python -m pytest -q tests/test_build_cache_posix.py
  tests/test_build_cache_windows.py`
  — 43 passed, 22 skipped.
- Full `python -m pytest -q`
  — 854 passed, 80 skipped in 106.71s.
- Strict `python -m mypy`
  — success, 65 source files.
- `uvx ruff check src/csk/builds/cache.py
  src/csk/builds/cache_windows.py tests/test_build_cache_windows.py`
  — all checks passed.
- `python -m compileall -q src tests`
  — exit 0.
- `python -m build --outdir .temp/TASK-260720-8nxlgx/dist-r1`
  — sdist and wheel built.
- `python -m twine check .temp/TASK-260720-8nxlgx/dist-r1/*`
  — both artifacts passed.
- Standalone `git diff --check`
  — exit 0.
- Standalone `test ! -e uv.lock`
  — exit 0.
- Final `git status --short` contains only the task-owned product delta:
  modified `src/csk/builds/cache.py`, new
  `src/csk/builds/cache_windows.py`, and new
  `tests/test_build_cache_windows.py`.

## Non-green command evidence

- Test-first local Windows-module pytest exited 1 before the implementation
  change: the synthetic inherit-only manager ACE test failed because no
  `_UntrustedState` was raised. This was the expected red gate.
- The reviewer's original standalone exploit script exited 1 after the fix.
  It printed `untrusted-provenance` for late artifact and receipt links, then
  stopped on the newly correct `cache_publication_invalid` exception for the
  publication source because the old script does not catch rejection as a
  passing outcome. It is not reported as a green gate; the integrated native
  regressions are green.
- One compound hygiene diagnostic exited 1 because its last intentional
  `git diff --no-index /dev/null <new-file>` reports content differences with
  exit 1. Every actual hygiene gate was rerun standalone and exited 0.
- Bare `python`, bare `ruff`, and shared-environment `python -m ruff`
  readiness probes were unavailable (127, 127, and 1 respectively). The
  successful project interpreter and `uvx ruff` entry points above were used
  for every real gate. Full readiness details are retained in the task-local
  `TASK-260720-8nxlgx_tool-readiness.md`.

## Durable finding

The shared `LOGBOOK.md` now records that Windows hard-link count is mutable
after byte verification and must be queried after security validation at the
actual handle-bound return boundary. It also records that inherit-only ACEs
can confer mutation rights on newly created descendants even when ineffective
on the manager home itself.
