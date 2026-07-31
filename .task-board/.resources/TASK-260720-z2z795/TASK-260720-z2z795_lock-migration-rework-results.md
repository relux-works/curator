# TASK-260720-z2z795 lock-migration rework outcome

## Scope and provenance

Reused the required worktree without reset or byte loss:

`/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/worktree`

- Original accepted base:
  `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`
- Current committed task head:
  `5f8bfbbd6e0c3a00dd0a3de0295e74f05990a0ec`
- The accepted base remains an ancestor of current head (`git merge-base
  --is-ancestor`, exit 0).
- The clean canonical main checkout was read-only inspected at
  `97a0ed870782b48eebc5a9c25a9cfa8fea5ff245`, equal to its local
  `origin/main`; it was not fetched, reset, or modified.
- Current changes remain unstaged and limited to:
  - `src/csk/locking.py`
  - `src/csk/transactions.py`
  - `tests/test_locking.py`
  - `tests/test_transactions.py`

Current SHA-256:

- `src/csk/locking.py`:
  `c272ff852f44a965c8da6e15afb396b43cd4ce7f6fbcc9522350fbf8f4760bb0`
- `src/csk/transactions.py`:
  `415a2418474582fbc9094c7f6e157212fd1e67543ccd77c905182f587e77cb52`
- `tests/test_locking.py`:
  `b668a726eb685d5a4882790c2283cb035697adbe189ed17cb73c572bcd42eb11`
- `tests/test_transactions.py`:
  `f578bbc4f400b6003e754e2ab58401af2875a287312ef055567fe9dd7b67606c`

## Reviewer finding closed

The legacy-to-stable transition is now fail-closed:

- A legacy record, including one whose PID is provably dead, is never rewritten
  or adopted online.
- A previously written unguarded v1 record is also rejected for offline
  migration, because a pre-upgrade breaker could already have read its integer
  PID.
- Stable v1 records carry an intentionally non-integer legacy `pid` guard and a
  separate diagnostic `owner_pid`. Pre-upgrade code therefore treats the
  persistent stable record as non-breakable instead of renaming it after the
  owner exits.
- Active `.lock.stale-*` side witnesses are detected before and after durable
  publication. Their namespace and descendants are also reserved from
  transaction targets, aliases, and sidecars before mutation.
- Acquisition rechecks canonical path/descriptor identity and the published
  token after `fsync` and before recording ownership.
- Stable release still drops only the OS lock and descriptor. It never unlinks
  or renames the canonical v1 file, and timeout diagnostics explicitly say the
  persistent v1 file must not be removed.

The deterministic migration regression runs a real pre-upgrade stale breaker
in a separate process, pauses it at the rename boundary, and attempts two
current owners around that boundary. Both current owners fail closed while the
legacy witness is active; after the breaker exits, a current stable owner
acquires normally. The test uses platform-neutral filesystem and subprocess
operations and is written to run unchanged on macOS, Linux, and Windows.

## Validation

Final current-byte gates:

- New migration and legacy-side-witness regressions: 14 passed, exit 0.
- Prior lock-integrity regressions: 14 passed, exit 0.
- Earlier reviewer regressions: 14 passed, exit 0.
- Contract-targeted regressions: 13 passed, exit 0.
- Focused pytest: 88 passed, 1 Windows-only skipped, exit 0.
- Strict `python -m mypy`: 57 source files clean, exit 0.
- Ruff lint: clean, exit 0.
- Ruff format check: 4 files already formatted, exit 0.
- Full pytest: 658 passed, 20 skipped, exit 0.
- Package build: sdist and wheel built, exit 0.
- `git diff --check`: clean, exit 0.

The initial two-test migration gate failed as expected with 2 failures and exit
1 before implementation. A second expected-red namespace gate failed with 2
failures, 1 pass, and exit 1 before the legacy side namespace was reserved.
The first full-suite run failed honestly with 657 passed, 20 skipped, 1 failed,
and exit 1 because the CLI contention test required its established live-owner
wording. The fail-closed error retained that wording, the focused CLI test
passed, and the full suite then passed as reported above.

Exact commands, real exit codes, diagnostic failures, hashes, and scope are in
`TASK-260720-z2z795_lock-migration-validation.log`. Tool readiness is in
`TASK-260720-z2z795_lock-migration-tool-readiness.md`.

This run was on macOS. No current-byte real-Windows run was available, and
earlier Windows CI is not claimed for these unstaged bytes. No SSH, host
management, staging, commit, push, tag, release, pin update, Go UX work, or
installer-policy integration was performed.
