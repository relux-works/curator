# TASK-260720-g7kgox — Atomic global builds

## Outcome

Implemented atomic schema-6 global installs on base
`c7dbd6daf6562a264275fca06b50a527bce236d4`. Global install now uses the
shared closure/build/cache/marker/shim/transaction pipeline, performs private
cache misses without the manager-home lock, revalidates and publishes under
that lock, and commits all global activation surfaces through one durable
transaction.

The transaction owns global contexts and marker-only nodes, script runtimes,
canonical launchers, PATH-visible user-bin launchers and ledger, environment
files, adapter mirrors and ledgers, and stale removals. Source and dependency
diagnostics remain per declaration, but any error now stops before build or
materialization instead of partially replacing the install.

Global dry-run remains read-only: it constructs no project/build/home mutation
lock and leaves the manager home and user home byte-identical. Script-only and
system-only flows remain covered by the existing global suite.

## Source evidence

- Lock lifecycle, shared build planning, publication, retry, and GC handoff:
  `src/csk/global_install.py:154`.
- Transaction target enumeration for contexts, runtime, canonical/user bins,
  env, adapters, and stale removals: `src/csk/global_install.py:419`.
- Private staged global materialization and marker-v2/build activation:
  `src/csk/global_install.py:679`.
- Concurrent runtime-reference inputs in the generation probe:
  `src/csk/global_install.py:1018`.
- Transactional user-bin planning/staging, including Windows wrappers and
  POSIX live-relative links: `src/csk/global_bins.py:36` and
  `src/csk/global_bins.py:147`.
- Deduplicated global adapter target planning:
  `src/csk/adapters.py:202`.
- Shared generic durable target commit used by project and global installs:
  `src/csk/installer.py:1299`.
- Build ordering, home-lock exclusion, marker v2, build-root filtering, and
  canonical/user-bin execution test: `tests/test_global_install_transactions.py:246`.
- Prior-install preservation for build and publication failures:
  `tests/test_global_install_transactions.py:353`.
- Reverse rollback at every global target class:
  `tests/test_global_install_transactions.py:449`.
- Lock-free byte-pure dry-run and Windows user-bin staging:
  `tests/test_global_install_transactions.py:527` and
  `tests/test_global_install_transactions.py:570`.
- Upgrade lock release plus option/exit preservation:
  `tests/test_global_install.py:1715`.
- Design decisions and non-obvious staging/link behavior:
  `LOGBOOK.md`, section “2026-08-01 — TASK-260720-g7kgox atomic global
  builds”.

## Validation evidence

All commands below ran as standalone processes; the recorded exit codes are
their actual process results.

- `.venv/bin/python -m pytest tests/test_global_install.py tests/test_global_install_transactions.py tests/test_installer_transactions.py tests/test_adapters.py tests/test_env_files.py tests/test_build_activation.py tests/test_shims.py tests/test_cli.py -q`
  — exit 0; 226 passed, 6 skipped.
- `.venv/bin/python -m pytest -q` — exit 0; 1166 passed, 100 skipped.
- `.venv/bin/python -m mypy` — exit 0; strict mode, 67 source files.
- `.venv/bin/python -m ruff check src/csk/adapters.py src/csk/cli.py src/csk/env_files.py src/csk/global_bins.py src/csk/global_install.py src/csk/installer.py tests/test_env_files.py tests/test_global_install.py tests/test_global_install_transactions.py`
  — exit 0.
- `.venv/bin/python -m build` — exit 0; sdist and wheel built.
- `.venv/bin/python -m twine check dist/*` — exit 0; both artifacts passed.
- `git diff --check` — exit 0.
- `python3 -m py_compile src/csk/global_install.py src/csk/global_bins.py src/csk/adapters.py src/csk/env_files.py src/csk/installer.py src/csk/cli.py`
  — exit 0.

Development diagnostics that did not pass are not counted as green gates:

- `python -m py_compile ...` — exit 127 because this host exposes `python3`,
  not a bare `python` command.
- `python3 -m mypy ...` — exit 1 before the task virtual environment was
  provisioned because the system interpreter had no `mypy` module.
- The first focused global run — exit 1 with six tests still asserting the
  retired declaration-only/partial-install behavior; those expectations were
  updated and the post-edit focused gate above passed.
- The first default Ruff run — exit 1; it found import/style findings in the
  touched files. They were corrected, and the exact post-edit Ruff gate above
  passed.

## Review/CI evidence

- Signed commit: `ea64669df0fa58b776ffd67842d40d85c32f4857`
  (`feat: make global installs atomic`). `git verify-commit` reports a good
  ECDSA signature for `oparin@me.com`, fingerprint
  `SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`.
- Pull request: https://github.com/ivanopcode/cocoaskills/pull/17
- CI workflow: https://github.com/ivanopcode/cocoaskills/actions/runs/30671637640
- The first `gh pr checks 17` observation exited 8 because checks were still
  pending; it is recorded as a pending diagnostic rather than a passing gate.
- Final `gh pr checks 17` — exit 0; all 14 checks passed: strict mypy, artifact
  build, and Python 3.11–3.14 on Ubuntu, macOS, and Windows.
