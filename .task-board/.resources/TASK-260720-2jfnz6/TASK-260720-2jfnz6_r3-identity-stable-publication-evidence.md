# TASK-260720-2jfnz6 R3 identity-stable publication evidence

Date: 2026-07-30
Developer run: `RUN-260730-7f45a4`

## Provenance and scope

- Product worktree:
  `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-2jfnz6/worktree`
- Branch: `task/TASK-260720-2jfnz6-protected-posix-build-cache`
- Recorded base and merge base with `main`:
  `495ad021847529ce5a544dba415ca2fe19949539`
- Published task HEAD:
  `0d6ad16fce35c1bd8854511e13766cd236908e3b`
- Dependency `TASK-260720-2dnqw2` is `done`.
- The unstaged product delta remains exactly:
  - `src/csk/builds/cache_posix.py`
  - `tests/test_build_cache_posix.py`
- `src/csk/builds/cache.py` is byte-identical to HEAD. The Git index was not
  changed, and no commit, push, or PR mutation was made.

## Reviewer finding closed

Reviewer cycle R3 proved that publisher A could be paused after atomic rename
and before sealing. Publisher B then quarantined A's mode-`0700` directory,
published and sealed a replacement, after which A reopened the logical key and
attempted to seal B's mode-`0500` winner. A returned
`cache_boundary_untrusted` instead of resolving the concurrent winner.

Publication now opens and revalidates the staged entry immediately before
rename and retains that rooted no-follow descriptor across rename. Sealing uses
only the retained descriptor, so it always targets the inode that actually won
that publisher's rename, even if the live name is moved afterward. The sealed
descriptor is re-inspected against the complete input, receipt hash, artifact
path, size, and hash.

After sealing, publication opens the selected live name read-only and pins that
descriptor before resolving the result:

- same directory identity: `published`;
- different identity with byte-identical canonical receipt and artifact:
  `reused-winner`;
- different receipt or artifact bytes for the same logical key:
  `CacheConflictError`.

The post-rename resolver never seals or quarantines through the logical name.
Transient missing or private-mode selections are observed briefly without
mutation; a persistent untrusted selection still fails closed.

## Deterministic regression proof

Two regressions pause publisher A inside `_seal_published_entry`, which is
called only after its successful atomic rename. Publisher B runs synchronously
until it has exceeded the existing retry window, quarantined A's private-mode
entry, atomically published its own distinct inode, sealed it, and returned.
Only then is A resumed.

- Identical bytes: B returns `published`; A seals its retained quarantined
  inode, detects B's different live identity, compares both pinned artifacts,
  and returns `reused-winner`.
- Different bytes: B returns `published`; A detects the different live
  identity and raises `CacheConflictError`.
- Both cases prove two different sealed inode identities, one protected
  mode-`0500` live hit, one protected mode-`0500` quarantined loser, and an
  empty staging namespace.

The focused two-test command exited 0 with `2 passed in 0.73s`.

## Exact green gates

All commands below were standalone processes and ran on the final candidate
hashes recorded below.

1. Final host POSIX-focused pytest:

   `.venv/bin/python -m pytest -q --basetemp=/private/tmp/TASK-260720-2jfnz6-r3-focused-final tests/test_build_cache_posix.py`

   Exit 0: `41 passed, 1 skipped in 0.51s`. The skip is the native-Linux-only
   `O_PATH`/`fchmodat2` case exercised by every Linux matrix run below.

2. Full repository pytest with the pinned accepted conformance corpus:

   `CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 uv run --extra dev python -m pytest -q --basetemp=/private/tmp/TASK-260720-2jfnz6-r3-full`

   Exit 0: `890 passed, 7 skipped in 81.60s`.

3. Strict typing:

   `.venv/bin/python -m mypy`

   Exit 0: `Success: no issues found in 63 source files`.

4. Task-scoped lint:

   `uvx ruff check src/csk/builds/cache_posix.py tests/test_build_cache_posix.py`

   Exit 0: `All checks passed!`.

5. Static and scope validation:

   - `.venv/bin/python -m compileall -q src/csk`: exit 0.
   - `git diff --check`: exit 0.
   - `git diff --exit-code -- src/csk/builds/cache.py`: exit 0.
   - `test ! -e uv.lock`: exit 0.

6. Distribution build:

   `.venv/bin/python -m build --outdir .temp/TASK-260720-2jfnz6/dist-r3.HLUP5f`

   Exit 0. Built:

   - `cocoaskills-0.12.6.dev7+g0d6ad16fc.d20260730.tar.gz`
   - `cocoaskills-0.12.6.dev7+g0d6ad16fc.d20260730-py3-none-any.whl`

   `.venv/bin/python -m twine check .temp/TASK-260720-2jfnz6/dist-r3.HLUP5f/*`
   exited 0; both artifacts passed.

## Native Linux matrix

A disposable Lima VM ran Ubuntu 26.04, Linux `7.0.0-28-generic`, `aarch64`.
Guest-only `uv` 0.12.0 installed managed CPython and pytest 9.1.1 into isolated
`/tmp` environments. Each focused pytest command was a standalone process,
used `PYTHONPATH=src`, disabled pytest's cache provider because the host mount
is read-only, and ran against the exact mounted worktree bytes:

- CPython 3.11.15: exit 0, `42 passed in 0.44s`.
- CPython 3.12.13: exit 0, `42 passed in 0.40s`.
- CPython 3.13.14: exit 0, `42 passed in 0.45s`.
- CPython 3.14.6: exit 0, `42 passed in 0.46s`.

Host and guest SHA-256 values matched exactly. The VM
`task-260720-2jfnz6-r3-linux` was then stopped and deleted, both commands
exiting 0. `limactl list` confirms that no instance remains. Its disposable
filesystem is not recoverable; the product files and this evidence remain on
the host.

## Non-green setup diagnostics

- The first host readiness chain exited 1 because this worktree initially had
  no `.venv`; `uv run --extra dev` established the ignored scratch environment.
- That setup generated an untracked `uv.lock`. A scope check truthfully exited
  1, and the generated lock was moved to the task's ignored readiness area.
  The final `test ! -e uv.lock` gate exited 0.
- Two initial guest `uv run` attempts exited 2 because guest `uv` discovered
  the mounted host macOS `.venv` and received `Exec format error`. They were
  setup attempts, not reported as test gates. Explicit guest `/tmp` virtual
  environments were then used for the four green standalone gates.
- An exploratory `uvx ruff format --check` exited 1 and changed no files. The
  repository does not configure Ruff formatting; the required task-scoped
  `ruff check` lint gate above exited 0.

## Candidate hashes

- `src/csk/builds/cache_posix.py`:
  `be370f6f4b63d355d7082b31d0f753ddcd8ad89449588130ae4edc2140d1aa36`
- `tests/test_build_cache_posix.py`:
  `c7724b619a8074d36f44d17aacea1124f0627d5130bcc8532e378dffe3f1d13c`

The developer rework is ready for review.
