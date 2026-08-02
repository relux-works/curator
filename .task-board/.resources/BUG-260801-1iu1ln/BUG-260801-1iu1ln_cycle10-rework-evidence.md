# BUG-260801-1iu1ln cycle-10 developer evidence

## Provenance

- Worktree: `/Users/iv/Developer/Wildberries/cocoaskills/.temp/BUG-260801-1iu1ln/worktree`
- Branch: `task/BUG-260801-1iu1ln-lifecycle-observed-traces`
- Exact signed merge base: `ba250bfc4dfe104a160eadd5b5f4e340693bf892`
- Parent: `dbd23cbf24602ccf5125ca78510d9eeb002d0bae`
- Signed commit: `80b5b1673db170e3db9be349c3649d9b4e03d520`
- `git verify-commit` for HEAD and base: exit 0, expected ECDSA key.
- Worktree status: clean; no tag points at HEAD.

## Repair

Cycle-9 review showed that the raw destination snapshot still ran after a
replaceable callable returned. A delegating callable supplied through
`cache_posix.ctypes.CDLL` could perform the real no-replace rename, mutate and
restore the live root with captured `os.fchmod`, then return with the complete
normative publication case unchanged.

The lifecycle publication observation now activates a `ContextVar`-scoped
CPython audit sink only around `backend.publish`. CPython's `os.chmod` audit
event observes descriptor-based `os.fchmod` even when captured before the test
observer patched `os`. Trusted audit paths below the live entry must correspond
exactly to the ordinary observed root-seal trace; any extra captured permission
event makes publication incomplete. The sink is reset in `finally`, leaving the
process audit hook inert outside the publication boundary.

A retained POSIX regression injects the mutating/restoring callable through
`ctypes.CDLL`, one layer inside the prior witness. A Windows-only equivalent
injects through `_api().kernel32.MoveFileExW` and mutates/restores via captured
`os.chmod`; Windows CI exercises that test, while this macOS run skipped it
explicitly. `LOGBOOK.md` records the finding and decision. No product source or
release surface changed in this cycle.

## Direct gates

- Authenticated POSIX regression plus 32 canonical lifecycle cases: exit 0,
  34 passed.
- Cycle-8/9 plus new native-boundary barrier: exit 0, 7 passed and 1 explicit
  Windows-only skip.
- Focused canonical/scalar/classification gate: exit 0, 414 passed, including
  all 378 scalar mutations rejected.
- Full authenticated exact-root protocol module: exit 0, 870 passed and 1
  Windows-only skip.
- Focused product regressions: exit 0, 3 passed.
- Install/global/currentness/installer-transaction suites: exit 0, 131 passed.
- Transaction/GC/status suites: exit 0, 111 passed and 1 expected platform skip.
- Strict project mypy: exit 0, no issues in 68 source files.
- `compileall`: exit 0.
- Uncommitted, staged, committed, and exact-base diff checks: exit 0.
- Cycle-parent product/release guard and exact-base pyproject/CI/changelog
  guard: exit 0.
- Clean signed-tree isolated PEP 517 build: exit 0; produced sdist and wheel
  version `0.12.6.dev47+g80b5b1673`.
- Twine check: exit 0 for both distributions.
- Sdist membership check: exit 0 for the logbook and both conformance files.
- HEAD/base signatures, exact merge base, clean tree, and no-tag checks: exit 0.

## Non-green and non-gate commands

- An initial pytest invocation without `CURATOR_CONFORMANCE_ROOT` exited 0 but
  skipped all 3 selected tests. It was not counted as coverage; the authenticated
  rerun passed 34 tests.
- A direct `mypy --strict` invocation against the test modules exited 1 with
  137 existing test-graph/import errors because the project type gate is
  configured for `src/csk`. It was not reported as passing. The canonical
  project command, `.venv/bin/python -m mypy`, ran independently after the
  signed commit and exited 0.
- The Windows native-API regression was not executed locally because this host
  is macOS; pytest reports the explicit skip. The equivalent test is retained
  for Windows CI.

No PR, main landing, tag, release, claim, pin, schema-v7, CI, changelog,
pyproject, or product-source action or change was made in cycle 10.
