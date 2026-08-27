# TASK-260720-2jfnz6 Linux CI rework

Date: 2026-07-30

## Provenance

- Product repository: `/Users/iv/Developer/intranet/cocoaskills`
- Existing task worktree:
  `/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2jfnz6/worktree`
- Branch: `task/TASK-260720-2jfnz6-protected-posix-build-cache`
- Reviewed and published pre-rework HEAD:
  `0d6ad16fce35c1bd8854511e13766cd236908e3b`
- Original recorded base:
  `495ad021847529ce5a544dba415ca2fe19949539`
- Dependency `TASK-260720-2dnqw2` remains `done`.
- Pull request: `ivanopcode/cocoaskills#11`
- Red CI baseline: GitHub Actions run `30514706010`.

## CI finding and root cause

The red baseline passed strict mypy and every macOS and Windows test job. All
four Ubuntu jobs failed the same tests:

- `test_self_consistent_untrusted_candidate_is_read_only_then_rebuilt_fresh[360]`
- `test_self_consistent_untrusted_candidate_is_read_only_then_rebuilt_fresh[0]`
- `test_corrupt_candidate_is_read_only_then_rebuilt_from_verified_stage`
- `test_locked_quarantine_can_move_immutable_entry_for_later_gc`

Each failure reached `os.rename` in `_move_aside` with
`PermissionError: [Errno 13]`. The rooted verification and mode-restoration
helper was restricted to Darwin, so Linux attempted the quarantine rename while
the owned source directory remained sealed or otherwise lacked full owner
control.

## Rework

- Apply the existing owner-only temporary unlock to verified POSIX directories
  generally instead of Darwin only.
- Continue to require a real directory owned by the effective UID.
- Open with the existing rooted `O_NOFOLLOW` flags, compare device/inode/owner
  against the no-follow pre-stat, grant owner `rwx`, and fsync before rename.
- Skip the mode mutation only when all owner `rwx` bits are already present.
- Restore the exact original mode through the retained descriptor after success
  and on every rename failure path.
- Preserve all group/other bits exactly; the temporary change adds only missing
  owner-control bits and never adopts the candidate.

## Regression coverage

Four Linux-semantic cases were added:

- successful quarantine from sealed `0500`;
- successful quarantine from untrusted `0550`;
- successful quarantine from inaccessible `0000`;
- forced rename failure with exact `0500` restoration and empty reservation
  cleanup.

The seam raises Linux-style `EACCES` when `os.rename` sees a directory without
owner `rwx`. It therefore reproduces the Ubuntu boundary deterministically on
the Darwin development host and verifies the mode observed at the rename call,
not only the final result.

## Command ledger

1. Test-first Linux reproducer:

   `/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python -m pytest -q tests/test_build_cache_posix.py::test_linux_quarantine_temporarily_unlocks_and_restores_sealed_entry`

   Exit 1 as expected: the pre-fix Linux branch reached the seam with mode
   `0500`, received `EACCES`, and surfaced `cache_quarantine_failed`.

2. First post-fix Linux regression run:

   `/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python -m pytest -q tests/test_build_cache_posix.py -k linux_quarantine`

   Exit 1: 3 passed and 1 failed because the `0550` test incorrectly expected
   temporary mode `0700`; the implementation correctly preserved the original
   group bits and observed `0750`. The assertion was corrected to require
   `original_mode | 0700`.

3. Corrected Linux regression run:

   `/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python -m pytest -q tests/test_build_cache_posix.py -k linux_quarantine`

   Exit 0: `4 passed, 33 deselected`.

4. Intermediate focused suite:

   `/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python -m pytest -q tests/test_build_cache_posix.py`

   Exit 0: `37 passed in 0.34s`.

5. Final task-environment POSIX suite:

   `/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2jfnz6/venv/bin/python -m pytest -q --basetemp=/private/tmp/TASK-260720-2jfnz6-rework-pytest.XRws2T tests/test_build_cache_posix.py`

   Exit 0: `37 passed in 0.42s`.

6. Strict typing:

   `/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2jfnz6/venv/bin/python -m mypy`

   Exit 0: `Success: no issues found in 63 source files`.

7. Full repository suite with the accepted conformance corpus:

   `CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 /Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2jfnz6/venv/bin/python -m pytest -q --basetemp=/private/tmp/TASK-260720-2jfnz6-rework-pytest.XRws2T`

   Exit 0: `886 passed, 6 skipped in 83.99s`.

8. Whitespace/static validation:

   - `git diff --check` — exit 0.
   - `python -m compileall -q src/csk` — exit 0.

   The repository declares no separate formatter or style-linter command.

9. Distribution validation:

   - `python -m build` — exit 0; built
     `cocoaskills-0.12.6.dev7+g0d6ad16fc.d20260730.tar.gz` and
     `cocoaskills-0.12.6.dev7+g0d6ad16fc.d20260730-py3-none-any.whl`.
   - `python -m twine check` on those exact two artifacts — exit 0; both
     `PASSED`.

## Candidate state

- `src/csk/builds/cache_posix.py` SHA-256:
  `4acb413d2964296350e93e6be2e3ed193c063b95ff0ce6b619585b9d6a1e1112`
- `tests/test_build_cache_posix.py` SHA-256:
  `cd797ddce979611887076f6214c9c747eb50a7ac5103dfda509e84e5e79f4874`
- Diff: 2 tracked files, 108 insertions, 6 deletions.
- The real Git index is untouched. Both files remain modified and unstaged.
- No commit, push, PR mutation, or temporary Linux infrastructure was created
  by this developer run.

Docker, Podman, and Colima are unavailable on the development host (readiness
commands each exited 127). The deterministic Linux-semantic regressions cover
the reported boundary locally; a fresh Ubuntu GitHub matrix remains the
post-integration landing gate.

This rework is ready for review.
