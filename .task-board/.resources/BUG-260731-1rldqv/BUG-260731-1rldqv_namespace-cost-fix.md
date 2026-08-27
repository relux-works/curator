# BUG-260731-1rldqv — Windows transactional install regression

Developer run `RUN-260731-d6e3c5`, resumed after the host migration.
Repository `ivanopcode/cocoaskills`, PR 16, branch
`task/TASK-260720-3t8nr3-transactional-project-hybrid`.

## State on arrival

Commit `7a66c73`, produced by the migration-cancelled `RUN-260731-2dec82`, had
already fixed the four Windows platform causes and was published to PR 16.
Windows 3.11, 3.13 and 3.14 of run `30624304158` report `SUCCESS`, so all four
original failure families (digest-mode, WinError 5, publication-owner, stub
toolchain) are gone and stayed gone.

One blocker remained: Windows Python 3.12 had been running 2h54m without
finishing.

## The GitHub API timestamps are wrong; the logs are not

`gh run view --json jobs` reported `completedAt` 11:13:18Z for the three green
Windows cells, i.e. about 34 minutes. Their own logs disagree:

| lane | suite result | duration |
|---|---|---|
| `main` `b3a5031`, run `30556125542`, win 3.11 | 1190 passed, 146 skipped | **349.43s (5m49s)** |
| PR 16 `7a66c73`, run `30624304158`, win 3.11 | 1206 passed, 152 skipped | **8323.00s (2h18m43s)** |

So the branch was **23.8x slower** than `main` on the same platform, every
Windows cell was affected rather than only 3.12, and individual install tests
took 76-247 seconds each. A matrix in that state is not reliably green: the
job timeout is the only thing bounding it.

## Diagnosis

`ssh win` was unreachable (Tailscale reports the host offline, last seen 1d
ago), so reproduction used a disposable probe branch running the suite on
`windows-latest` under `pytest-timeout` and `faulthandler`. Probe run
`30634563988`, branch deleted afterwards.

Two independent dumps, Python 3.11 and Python 3.12, landed on the identical
frame:

```
_canonical_target_path        transactions.py:3222   <- Path.resolve -> ntpath.realpath
_namespace_parts              transactions.py:3358
_namespaces_overlap           transactions.py:3314
_validate_namespace_independence transactions.py:3294
_validate_journal             transactions.py:1066
_save_journal                 transactions.py:809
prepare / _stage_target       transactions.py:193
_commit_materialization       installer.py:1331
```

Same hot path on both interpreters, so this was one cost problem, not a hang.

## Root cause

`_validate_namespace_independence` compares every declared namespace with every
other one, and each comparison canonicalised **both** of its operands through
`Path.resolve()`. Filesystem work therefore grew with the square of the
namespace count, and the whole pass re-runs on every `_save_journal` —
including twice per 32 KiB staging chunk.

Measured on macOS, where `realpath` is cheap, for one install test
(`test_full_requirement_is_the_default_and_activates_both_surfaces`):

```
canonical_target_path: 750620
canonical_seconds:     182.478
validate_journal:      134
save_journal:          133
namespace passes:      135          (~74 namespaces each)
1 passed in 210.00s
```

182 of 210 seconds inside path canonicalisation. Windows opens a handle per
path component to answer the same question, which is why only Windows became
unusable while POSIX merely looked slow.

Reachable only under `c4131bd`: `main`'s installer plans builds but does not
materialise them transactionally.

## Fix — signed commit `98ab7a2` (parent `7a66c73`)

A namespace is now a `_NamespaceProbe` that resolves its path, and reads its
physical identity, at most once per pass; every comparison in the pass reads
that one answer. The redundant re-derivation of manager-home parts once per
added namespace is gone too.

**The guard is unchanged.** For every pair the pass still runs parts equality,
prefix containment, and the `samestat` physical-alias check. Identity is still
`lstat` for entry targets and `stat` otherwise. A real `OSError` is still never
cached, so it cannot be hidden from a later reader. Asking the filesystem one
question once per pass is if anything more internally consistent than asking it
once per comparison, and the pass runs under the exclusive manager-home lock.

Same test after the fix:

```
canonical_target_path: 12575        (59.7x fewer)
canonical_seconds:     10.442       (17.5x less)
1 passed in 27.88s                  (7.5x faster wall clock)
```

## Tests

Five new tests in `tests/test_transactions.py`:

- `test_namespace_independence_canonicalises_each_namespace_once` — one
  canonicalisation per namespace plus the manager home.
- `test_namespace_independence_cost_stays_linear_in_targets` — growth tracks
  the namespace count, not its square.
- `test_namespace_probe_reads_one_identity_for_repeated_comparisons`
- `test_namespace_probe_reports_absence_without_raising`
- `test_namespace_independence_still_detects_a_physical_alias` — a hardlink
  alias between two targets is still rejected as a namespace overlap.

Red/green check: with per-access re-resolution restored, 60 namespaces produce
3597 canonicalisations instead of 61 and both cost tests fail. All 13
pre-existing namespace, overlap and lock-alias guard tests pass unchanged.

## Local evidence (macOS, CPython 3.11.6)

| command | exit | result |
|---|---|---|
| `python -m mypy` | **0** | no issues in 67 source files |
| `pytest tests/test_transactions.py tests/test_installer_transactions.py tests/test_locking.py tests/test_adapters.py` | **0** | 156 passed, 4 skipped in 38.63s |
| `pytest tests/test_transactions.py -k namespace` | **0** | 18 passed, 82 deselected |

No dedicated linter is configured in `pyproject.toml` (no ruff/flake8/black);
`mypy` strict is the project's static gate and is green.

## Second stage — signed commit `32737a8`

Resolving once per namespace left the pairwise scan itself as the remaining
square term: 364,635 comparisons for one install, which became the dominant
cost on Windows once the canonicalisations were gone. The pass now asks its
question of an index rather than of every pair:

- **naming** — a dictionary keyed by the normalized parts;
- **containment** — a lookup of each namespace's proper prefixes over that same
  dictionary, path depth being bounded;
- **physical aliasing** — a dictionary keyed by `st_dev` with `st_ino`, which is
  exactly what `os.path.samestat` compares.

Same predicate, same rejections, same message shape, and the reported pair is
still a pair that genuinely collides. The install test measured above:

```
canonical_target_path: 12575
canonical_seconds:     1.596
1 passed in 5.49s
```

210s -> 27.9s -> **5.49s**, a 38x reduction overall.

## CI

**Run `30636027978`, head `98ab7a2`: all 14 jobs green**, including
`Build artifacts` and `mypy strict`.

| lane | Windows suite duration |
|---|---|
| `main` `b3a5031` baseline | 5m49s |
| PR 16 `7a66c73`, before this fix | 2h18m43s; 3.12 never finished |
| PR 16 `98ab7a2` win 3.11 | 1211 passed, 152 skipped in **1069.66s (17m49s)** |
| PR 16 `98ab7a2` win 3.12 | 1211 passed, 152 skipped in **703.08s (11m43s)** |
| PR 16 `98ab7a2` win 3.13 | 1211 passed, 152 skipped in **1057.67s (17m37s)** |
| PR 16 `98ab7a2` win 3.14 | 1211 passed, 152 skipped in **827.69s (13m47s)** |

Python 3.12 on Windows — the cell that had run three hours without finishing —
now completes in under twelve minutes.

Full local macOS suite, exit 0 on both commits, 1145 passed / 100 skipped
each. The recorded pre-fix baseline was 1140 passed / 100 skipped, so the
delta is exactly the five new tests.

| commit | full local suite |
|---|---|
| `98ab7a2` | 1145 passed, 100 skipped, **626.80s (10m26s)**, exit 0 |
| `32737a8` | 1145 passed, 100 skipped, **574.42s (9m34s)**, exit 0 |

PR 16 head is `32737a8`; run `30637483316` verifies the second stage on the
full matrix.

## Final CI — run `30637483316`, head `32737a8`

`completed/success`. All 14 jobs green: 12 test cells (Ubuntu, macOS and
Windows across Python 3.11, 3.12, 3.13, 3.14), `Type check / mypy strict`, and
`Build artifacts`.

| Windows cell | before the fix | at `98ab7a2` | at `32737a8` |
|---|---|---|---|
| 3.11 | 2h18m43s | 17m49s | **10m27s** |
| 3.12 | never finished (3h+) | 11m43s | **7m55s** |
| 3.13 | ~2h18m | 17m37s | **15m20s** |
| 3.14 | ~2h18m | 13m47s | **16m17s** |

All four cells report 1211 passed, 152 skipped. Windows Python 3.12 — the cell
that had run more than three hours without finishing — now completes in under
eight minutes, against a `main` baseline of 5m49s.

## Residual and open items

- The `_save_journal` -> `_validate_journal` path still re-validates the full
  journal on every save, including twice per 32 KiB staging chunk. That was
  left alone deliberately: skipping passes changes *when* corruption is
  detected, which is a guard question rather than a cost question. Both fixes
  here only make each pass cheaper, so every pass that ran before still runs.
- The product decision recorded by the previous run is unchanged and still
  open: a Windows manager home created by an elevated shell before `7a66c73`
  stays Administrators-owned and fails closed. Adopting it automatically would
  blunt the drift guard, so explicit opt-in repair versus documented manual
  re-provisioning remains unresolved. Not blocking this fix.
- `ssh win` was offline for the whole run, so all Windows reproduction and
  verification used `windows-latest` runners.
