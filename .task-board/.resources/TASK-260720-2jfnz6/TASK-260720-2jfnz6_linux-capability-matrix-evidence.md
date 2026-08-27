# TASK-260720-2jfnz6 Linux capability rework evidence

Date: 2026-07-30

## Provenance and scope

- Product repository: `/Users/iv/Developer/intranet/cocoaskills`
- Existing task worktree:
  `/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2jfnz6/worktree`
- Branch: `task/TASK-260720-2jfnz6-protected-posix-build-cache`
- Recorded base: `495ad021847529ce5a544dba415ca2fe19949539`
- Current HEAD: `0d6ad16fce35c1bd8854511e13766cd236908e3b`
- Dependency `TASK-260720-2dnqw2` (`canonical-build-metadata`) is `done`.
- Final product diff contains only:
  - `src/csk/builds/cache_posix.py`
  - `tests/test_build_cache_posix.py`
- `git diff --exit-code -- src/csk/builds/cache.py` exited 0.
- `test ! -e uv.lock` exited 0.
- The Git index was not changed. No commit, push, or PR mutation was made.
- The native-Linux capability result and cleanup decision were appended to the
  shared Curator `LOGBOOK.md`.

## Changes under review

- Quarantine now verifies and temporarily unlocks an owned sealed directory
  before allocating a destination reservation.
- Every unlock, reservation, rename, and retry failure restores the exact
  original mode and removes any reservation.
- Linux mode-`0000` recovery uses a rooted `O_PATH | O_NOFOLLOW | O_DIRECTORY`
  descriptor and `fchmodat2(AT_EMPTY_PATH)`. Unsupported kernel or architecture
  capability fails closed with `cache_protection_unsupported`.
- The reopened directory is matched to the original device, inode, owner, and
  mode before it can be moved.
- Darwin retains rooted no-follow chmod recovery and translates unavailable
  primitives into stable cache errors.
- Regression coverage proves sealed `0500`, untrusted `0550`, and inaccessible
  `0000` success, exact restoration on rename and fsync failures, no reservation
  leakage, real Linux descriptor restoration, and fail-closed capability
  handling.

## Exact green gates

### Native Linux CI matrix

A task-scoped Lima VM ran Ubuntu 26.04, Linux
`7.0.0-28-generic`, `aarch64`. Each command was a standalone process using the
exact mounted worktree bytes, `PYTHONPATH=src`, an isolated basetemp, and pytest
9.1.1:

- Python 3.11.15: exit 0, `40 passed in 0.47s`.
- Python 3.12.13: exit 0, `40 passed in 0.35s`.
- Python 3.13.12: exit 0, `40 passed in 0.35s`.
- Python 3.14.3: exit 0, `40 passed in 0.27s`.

This executes the real Linux `O_PATH` and `fchmodat2` regression on every
Python version in `.github/workflows/ci.yml`; there are no Linux skips.

### Host POSIX and repository gates

1. Focused POSIX pytest:

   `PATH=/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2jfnz6/venv/bin:$PATH python -m pytest -q --basetemp=/private/tmp/TASK-260720-2jfnz6-focused-final.oatJBX tests/test_build_cache_posix.py`

   Exit 0: `39 passed, 1 skipped in 0.36s`. The skip is the native-Linux-only
   test proven green in all four Linux matrix runs above.

2. Full repository pytest:

   `PATH=/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2jfnz6/venv/bin:$PATH CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 python -m pytest -q --basetemp=/private/tmp/TASK-260720-2jfnz6-full-final.4XWe0o`

   Exit 0: `888 passed, 7 skipped in 80.15s`.

3. Strict typing:

   `PATH=/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2jfnz6/venv/bin:$PATH python -m mypy`

   Exit 0: `Success: no issues found in 63 source files`.

4. Task-scoped Ruff:

   `uvx ruff check src/csk/builds/cache_posix.py tests/test_build_cache_posix.py`

   Exit 0: `All checks passed!`.

5. Static validation:

   - `python -m compileall -q src/csk`: exit 0.
   - `git diff --check`: exit 0.

6. Distribution build:

   `python -m build --outdir .temp/TASK-260720-2jfnz6/dist-final.kaT6Yd`

   Exit 0. Built:

   - `cocoaskills-0.12.6.dev7+g0d6ad16fc.d20260730.tar.gz`
   - `cocoaskills-0.12.6.dev7+g0d6ad16fc.d20260730-py3-none-any.whl`

7. Distribution metadata:

   `python -m twine check` on the two exact artifacts above exited 0; both
   artifacts passed.

## Tooling and anomaly ledger

- The host has no default `python` alias. Initial readiness invocations exited
  127 and were not treated as gates. The established task venv provided
  Python 3.14.4, pytest 9.1.1, mypy 2.3.0, build 1.5.0, and twine 7.0.0.
- Ruff was absent from the task venv. `uvx ruff` 0.16.0 was used without adding
  a project dependency. An initial broader three-file Ruff invocation exited 1
  with 13 findings, including an out-of-scope `cache.py` import-layout finding.
  The task-file findings were addressed, `cache.py` was restored byte-for-byte
  to HEAD, and the final two-file command above exited 0.
- `uv run` generated an untracked `uv.lock` and a scratch `.venv`. Both were
  moved under the ignored task readiness directory; neither remains in the
  product worktree root.
- Homebrew Lima 2.2.0 was installed with:

  `HOMEBREW_NO_AUTO_UPDATE=1 HOMEBREW_NO_ENV_HINTS=1 brew install lima`

  The task VM was created with:

  `limactl start --name=task-260720-2jfnz6-linux --tty=false --cpus=2 --memory=2 --disk=10 --containerd=none template:default`

  Guest venv creation first exited 1 because Ubuntu omitted `ensurepip`; an
  initial package install exited 100 before `apt-get update`. Installing the
  guest-only `python3-venv` package then succeeded. These setup failures were
  not reported as green gates.
- After the matrix passed, the VM was stopped and deleted with exit 0:

  - `limactl stop task-260720-2jfnz6-linux`
  - `limactl delete --tty=false task-260720-2jfnz6-linux`

  `limactl list` confirms that no instance remains. The disposable VM filesystem
  is not recoverable; all task evidence remains in host logs and this resource.
  The Homebrew Lima formula remains installed.
- A standalone `logbook` executable is unavailable (readiness exit 1); the
  established shared Curator `LOGBOOK.md` and task notes hold the durable
  finding.

## Candidate hashes

- `src/csk/builds/cache_posix.py`:
  `6068d2c772de0a2d9497bbc36def0f6ffe7d87ff26d21cf637208ec13d72a369`
- `tests/test_build_cache_posix.py`:
  `aedbe7fd6a03c31f30cbc48609bf6124b2686a7d50c90b429af6b251800b8f13`

The developer role work is ready for review.
