# BUG-260802-1s021p — Independent pre-acceptance review of CocoaSkills PR #19

**Reviewer run:** RUN-260802-e328bf (read-only; no code modified)
**Repository:** `ivanopcode/cocoaskills`
**PR:** [#19](https://github.com/ivanopcode/cocoaskills/pull/19) — `test: consume rc.6 candidate conformance vectors`
**Exact head reviewed:** `6e7742f0d28ad95ddd7d8e92364b84062571ad0b`
**Base / merge-base:** `main` @ `dacccaaf3ed18740a4d501fe8a3bfec64644c03e`
**Mergeable:** yes

## Verdict

**CHANGES REQUESTED.**

One confirmed blocking defect: the Windows portability fix in the head commit
repairs **one of five identical call sites**. The other four are on the same
straight-line path, in a function proven to execute on `windows-latest`, with
no platform guard and no exception handling. Windows CI at this exact head will
fail with the same `NotImplementedError` it failed with on the previous run,
approximately 66 lines later. Details in §3.

Everything else audited — correctness of the transaction-recovery relocation,
lock/`EXIT_LOCK` semantics, the new runtime build-root security guard,
signatures, release guards, suite integrity, and anti-vacuity of the new test
mass — is sound and would support acceptance once §3 is fixed. Findings §5–§11
are follow-ups, not rework.

## 1. Head and signature verification

Verified against the exact head, not the branch tip at some other time.

- `gh pr view 19 → headRefOid` = `6e7742f0d28ad95ddd7d8e92364b84062571ad0b` — matches task scope.
- Fetched `refs/pull/19/head` independently; `git rev-parse` confirms the same OID.
- All 22 commits `origin/main..6e7742f` verify with `%G? == G` (good signature),
  signer `oparin@me.com`. No unsigned, `U`, `B`, `E` or `N` commits.

## 2. Change surface

```
 LOGBOOK.md                               |  377 ++      (documentation)
 pyproject.toml                           |    1 +       (dev dep: jsonschema>=4.23)
 src/csk/cli.py                           |    4 +
 src/csk/global_install.py                |   12 +-
 src/csk/installer.py                     |   17 +-
 src/csk/status.py                        |   33 +
 tests/…                                  | 12419 +       (10 files)
 13 files changed, 12863 insertions(+), 37 deletions(-)
```

`.github/workflows/ci.yml` is **not** modified — relevant to the release guard (§7).

## 3. BLOCKING — Windows portability fix is incomplete (4 of 5 call sites unfixed)

### Root cause established from the previous run

Run 30737293076 (`9228054`) failed on `windows-latest` for all four Python
versions. Step-level inspection confirms genuine `Run tests` failures, not
cancellations. The log gives a single shared cause, repeated for every lifecycle
case:

```
>       os.utime(entry, (1, 1), follow_symlinks=False)
E       NotImplementedError: utime: follow_symlinks unavailable on this platform
tests\protocol_lifecycle_observations.py:2967
```

Because all 32 `test_rc6_manager_lifecycle_case[...]` cases draw from one shared
cached observation build, this single raise failed the entire lifecycle suite.

### What the head commit fixed

`6e7742f` adds `_utime_portably()` (a `NotImplementedError` → retry-without-
`follow_symlinks` shim) and rewrites exactly one call site — old line 2967,
new line 2989:

```python
-        os.utime(entry, (1, 1), follow_symlinks=False)
+        _utime_portably(os.utime, entry, times=(1, 1))
```

### What remains broken

Four identical unguarded call sites survive at head, all inside the same
function `_observe_gc` (defined at line 2751), all *after* the repaired one:

| head line | call |
|---|---|
| 3055 | `os.utime(rejected_entry, (1, 1), follow_symlinks=False)` |
| 3063 | `os.utime(entry, (young_mtime, young_mtime), follow_symlinks=False)` |
| 3071 | `os.utime(entry, (2, 2), follow_symlinks=False)` |
| 3112 | `os.utime(other_entry, (1, 1), follow_symlinks=False)` |

They were previously unreachable only because execution died at 2967 first.

Reachability and non-tolerance verified, not assumed:

- **No platform gate.** An `os.name` / `sys.platform` / `pytest.skip` scan across
  lines 2780–3120 returns nothing. Lines 2999–3055 are straight-line code — no
  early return, no `try`/`except` covering the span.
- **The function runs on Windows.** The previous run's traceback is inside
  `_observe_gc` itself, and `gc_cases:locked-mark-and-sweep-compiled-cache`
  appears in that run's failure list.
- **The active `os.utime` wrapper does not tolerate it.** `os.utime` *is*
  monkeypatched (line 638 → `observed_os_utime`), but that wrapper forwards
  `follow_symlinks=follow_symlinks` verbatim to `real_os_utime`
  (lines 582–605), so `NotImplementedError` propagates. This is consistent with
  the previous traceback, which shows the observer wrappers in scope and the
  error still surfacing at the call site.

### Falsifiable prediction

At the time of writing, run 30743353816 at head `6e7742f` has ubuntu×4,
macos×4 and `mypy strict` green, with the four `windows-latest` jobs still
running (~1h50m elapsed; the previous run's Windows jobs took ~2h56m). This
review predicts those four jobs fail with
`NotImplementedError: utime: follow_symlinks unavailable on this platform` at
`tests/protocol_lifecycle_observations.py:3055`. The static evidence above is
conclusive on its own and does not depend on that confirmation.

### Requested change

1. Route all four call sites through the existing `_utime_portably()` helper.
2. Add a regression guard so this does not require a sixth iteration. The suite
   already has the right pattern —
   `test_rc6_lifecycle_observer_rejects_known_lossy_proxy_forms` scans the
   observation module source for forbidden constructs. Extend that idea: fail
   if `follow_symlinks=False` appears in `protocol_lifecycle_observations.py`
   anywhere outside the body of `_utime_portably`. Fixing one call site per CI
   round on a ~3-hour Windows job is the actual cost driver here; a source-level
   guard collapses that loop.

## 4. Correctness — transaction-recovery relocation (sound)

Both `installer._install_project` and `global_install.install` lose their
pre-loop block:

```python
if not options.dry_run:
    with locking.ManagerHomeLock(csk_home) as home_lock:
        _transaction_engine(csk_home).recover(home_lock)
```

Recovery is not lost. It now runs at the pre-existing commit-phase call sites
(`installer.py:466-467`, `global_install.py:362-363`), inside
`with locking.ManagerHomeLock(...)`. Audited consequences:

- **Dry-run stays mutation-free.** The removed `if not options.dry_run` guard is
  subsumed by the existing dry-run early `return result` in both
  `_install_project_once` (installer.py:425-439) and `_install_once`, which
  fires *before* the `ManagerHomeLock`/`recover` block. `csk install --dry-run`
  therefore still performs no recovery write. This was the sharpest regression
  candidate and it is clean.
- **Plan-then-recover ordering converges.** `expected_generation` and
  `target_preimages` are captured before recovery; if recovery rolls back
  partial state, `_assert_generation_current` / `_assert_target_preimages_current`
  raise `BuildPlanningError("concurrent_state_change")`, which the retry loop
  explicitly retries (`attempts = 3` for non-dry-run). Attempt 2 re-plans against
  post-recovery state. No unsafe write ordering: `_publish_planned_builds` and
  `_commit_*` both run after `recover()` under the same held lock.
- **Error surfacing preserved.** `transactions.TransactionError` from recovery is
  not a `LockError`, so it falls to the broad boundary handler and still becomes
  a per-project / per-global `status="failed"` with the original message — same
  observable outcome as the removed block.
- **Covered.** `test_install.py::test_transaction_recovery_waits_for_private_builds_and_reports_per_project`
  asserts the new ordering directly: `home_lock.assert_held()`,
  `events == ["build:tool"]` at recovery time, and
  `result.errors == ["fixture corrupt journal"]`.

## 5. Lock semantics — EXIT_LOCK contract (sound)

`except locking.LockError: raise` added to `_install_project_once` and
`_install_once`, ahead of the `except Exception` boundary.

This **restores** a pre-existing contract rather than inventing one. Before the
relocation the home lock was acquired outside the broad handler, so a `LockError`
propagated to `cli.main`, which maps it to `EXIT_LOCK` (3) at `cli.py:85-87`.
Moving the home lock into the commit phase put it *inside* the `except Exception`
boundary, where it would have degraded to a generic per-project failure. The
re-raise is the correct minimal fix.

- `LockOrderError` subclasses `LockError` (`locking.py:40-45`), so lock-ordering
  violations — a coordination bug, not a data error — also propagate rather than
  being swallowed. Improvement.
- **Covered.** `test_global_install.py::test_global_install_lock_contention_returns_lock_exit`.
- **Observation (not a defect).** On a multi-project `csk install`, a `LockError`
  now aborts the whole run and discards the accumulated `ProjectResult` list, so
  the operator loses the partial report of projects that already succeeded. This
  matches pre-relocation behaviour (the old pre-loop home lock also propagated),
  so not a regression — but worth a line in the RC notes if partial-progress
  reporting is a documented property.

## 6. Security — new runtime build-root exposure guard (sound)

`status._runtime_exposes_build_roots` (status.py:775-789) fails closed when
excluded build input appears in an installed runtime tree.

- **Path layout is correct.** `csk_home / "runtime" / node.name / node.resolved.commit`
  is byte-for-byte the layout the installer materializes to (`installer.py:674`
  plus `identifier=f"{node.name}/{node.resolved.commit}"`, `global_install.py:858`,
  `gc.py:435`). A wrong path here would make the guard silently fail open; it
  does not.
- **Fail-closed semantics.** `FileNotFoundError → continue`; any other `OSError`
  → `True`; successful `lstat` → `True`. Skills with no runtime dir correctly
  return `False`.
- **No traversal reachable.** `build_root.split("/")` is safe because
  `skillspec._parse_build_roots` already enforces relative, strict-POSIX,
  link-free, unique, mutually disjoint roots that do not overlap runtime roots.
  A hostile `build_roots` entry cannot escape `runtime_dir`; worst case is a
  fail-closed false positive.
- **Mirrors the existing control.** Structurally identical to the second half of
  `installer._installed_context_exposes_build_roots`, wired next to it in
  `_inspect_node_marker` with a distinct `build-runtime-exposed` label.
- **Covered** by `test_build_currentness.py::test_runtime_build_root_exposure_is_non_current`.
  ⚠ That test carries `@POSIX_BUILD_VECTOR`, so **the new security guard is never
  exercised on Windows CI.** Consistent with suite-wide build-vector gating, but
  it means Windows regressions in this control would go uncaught. Follow-up.

## 7. Release guards (hold)

All three hold, and the substantive one was verified independently of the
assertions that claim it.

- **Committed spec pin not advanced.** `.github/workflows/ci.yml` is untouched
  and still pins `relux-works/curator-spec` @ `0c81c1f8d5321d822be2a2817b05aea03e656e15`.
- **PR-body / CI ref discrepancy is cosmetic.** The PR body cites curator-spec
  `432eb2ee1fe2d6b271e37269f867c8851c325539` while CI checks out `0c81c1f…`.
  Verified in the curator-spec clone that both commits carry an *identical*
  `conformance/v1` tree (`36287c9ae4cbb7e387b3d267f2736acd622f83e3`) and both
  manifests hash to the pinned
  `sha256:12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071`.
  No content divergence. `EXPECTED_MANIFEST_SHA256` is unchanged from `main`,
  consistent with "do not advance the committed ref".
- **No claim emitted.** `test_rc6_candidate_manifest_and_release_record_are_exact_non_release_evidence`
  asserts `committed_release_pin_advanced is False`, `claim_v3.claims_emitted == []`,
  `claim_v3.rc6_claim_schema is None`.

## 8. Suite integrity — caller-supplied conformance root (sound)

The threat is a caller-supplied `CURATOR_CONFORMANCE_ROOT` whose contents do not
match the pinned candidate. The chain holds:

- `_root()` pins `manifest.json` to a hardcoded digest constant.
- `_manifest_inventory()` asserts the exact manifest key set, asserts
  `protocol_version == "1.0.0-rc.6"`, rejects duplicate paths, rejects absolute
  paths and `..` components, validates each `sha256:` field shape.
- `_read_manifest_entry()` refuses to read any file absent from the inventory,
  re-checks absolute/`..`, and compares the file's actual digest to the inventory
  before returning bytes.
- `test_rc6_in_scope_vector_inventory_is_exhaustive` pins exact per-schema counts
  summing to 102, plus 8 positive / 77 rejection build-driver cases, 75
  condition-bound rejections, 10 build-source, 12 toolchain, and host-policy
  clusters — so the in-scope case set cannot silently shrink.

**Minor gap (low, non-blocking):** a few reads bypass `_authenticated_bytes` —
`expected/registry/pinned_key.txt`, `expected/context_sha256.txt`,
`expected/registry/record_audited.json`, and `release/1.0.0-rc.6.json` (the last
lives outside `conformance/v1`, so the manifest cannot cover it). Under a hostile
local root these could relax a guard assertion. Recommend routing the three
in-suite paths through `_authenticated_bytes`.

## 9. Anti-vacuity of the new test mass (sound)

12.4k lines of new test code is exactly where a review should look for tests that
pass because the harness avoided real behaviour. They do not.

- **Spies, not stubs.** The lifecycle observer captures `real_* = <production fn>`
  and installs wrappers that call `real_*(...)` then record. Every patched seam
  (`read_install_marker`, `build_closure`, `freeze_snapshot`, `plan_builds`,
  `cache_key`, `content_sha256`, `_installed_files`,
  `_installed_context_exposes_build_roots`, `_runtime_exposes_build_roots`,
  `cache_for_manager_home`, `go_v1.build`, `subprocess.run`/`Popen`) delegates to
  the real implementation.
- **The harness polices itself.**
  `test_rc6_lifecycle_literal_answers_are_explicitly_classified` AST-walks
  `protocol_lifecycle_observations.py` and fails on any hardcoded
  `observed["case"] = {...}` literal not registered in a 19-entry classification
  table — and fails equally on stale registrations. All 19 surviving literals are
  documented as *inputs* (exact CLI flags, fixture selections), not answers.
- **Known lossy proxies are banned by name** —
  `test_rc6_lifecycle_observer_rejects_known_lossy_proxy_forms` forbids
  `command[0]`, `.issubset(`, and three basename-proxy forms.
- **`mock` usage is legitimate** — `wraps=` spies plus targeted fault injection to
  drive documented rejection vectors, not stubbing of the system under test.
- Adapters validate schema cases through real `jsonschema` validators with a
  `referencing` registry and route executable vectors through real CocoaSkills
  parsers.

**Observation:** the whole conformance suite is
`pytest.mark.skipif(not CURATOR_CONFORMANCE_ROOT)`. If that env var were ever
dropped from `ci.yml`, ~1700 tests would silently skip and CI would stay green.
The pattern predates this PR, but its blast radius just grew substantially.
Recommend a guard test that hard-fails when the root is unset under CI.

## 10. `csk global upgrade` fetch fix (sound)

`fetch_existing=False` → `fetch_existing=options.fetch` in
`global_install._build_nodes`, gated in `cli.py` by
`global_command == "upgrade" and not dry_run`. `global_command` is a real
subparser dest declared `required=True` (`cli.py:422`), so the defensive
`getattr(..., None)` cannot silently disable the fetch. This makes global scope
symmetric with the project path (`cli.py:883`, `installer.py:349`) — previously
`csk global upgrade` never fetched its closure at all.

**Covered** by `test_global_upgrade_fetches_transitive_closure`, which asserts
direct *and* transitive repos are fetched and an unrelated repo is not.

## 11. Minor / hygiene findings

| # | Finding | Severity |
|---|---|---|
| 1 | Commit `9362cc8` is typed `test:` but carries three production behaviour changes (CLI fetch gate, recovery relocation in both installers, new `status.py` guard). Degrades changelog derivation and bisect readability for an RC. History is signed; cover it in the RC notes rather than rewriting. | low |
| 2 | New security guard is `@POSIX_BUILD_VECTOR`-gated → unexercised on Windows (§6). | low |
| 3 | Unauthenticated suite reads (§8). | low |
| 4 | `mypy` config is `files = ["src/csk"]`; the 12.4k new test lines are not type-checked in CI. Pre-existing config, not a regression. | info |
| 5 | `protocol_conformance_adapters.py` imports `referencing` directly while only `jsonschema>=4.23` is declared. Resolves transitively (jsonschema ≥4.18 hard-depends on it), but an explicit dev-dep would be cleaner. | info |
| 6 | `global_install._build_nodes` passes a fresh `fetched_repos=set()` per `resolve()`, so global scope emits no `"fetched X"` messages (project scope does) and can re-fetch on the per-decl isolation fallback path. Cosmetic/efficiency. | info |

## 12. CI evidence at exact head `6e7742f` (run 30743353816)

| Job | Result |
|---|---|
| Tests / Python 3.11–3.14 on ubuntu-latest | ✅ SUCCESS |
| Tests / Python 3.11–3.14 on macos-latest | ✅ SUCCESS |
| Type check / mypy strict | ✅ SUCCESS |
| Tests / Python 3.11–3.14 on windows-latest | ⏳ IN_PROGRESS — predicted FAIL, §3 |

Branch history: `b754bd7` ✅, `7915b49` ❌, `fed8276` ❌, `ba250bf` ✅,
`d0c2062` ❌, `041db70` ❌, `8cb7942` ❌, `2524c37` ❌, `3c3406b` ❌,
`9228054` ❌ (windows-only ×4), `6e7742f` ⏳.

## 13. Path to acceptance

1. Fix the four remaining `follow_symlinks=False` call sites (§3) and add the
   source-level regression guard.
2. All four `windows-latest` jobs terminal `success` at the new exact head.
3. Re-run this review at that new head — the verdict is head-bound and any new
   commit invalidates it.

§4–§10 need no rework. §8, §9 and §11 are follow-ups that should not block the
RC.
