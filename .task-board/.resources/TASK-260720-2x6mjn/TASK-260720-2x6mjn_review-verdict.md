# TASK-260720-2x6mjn — reviewer verdict

**Verdict: changes requested → `to-dev`**

Reviewer: reviewer role, RUN-260730-6a9fa9 (not goal-bound).
Reviewed tree: `/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2x6mjn/worktree`
(branch `task/TASK-260720-2x6mjn-pure-build-planner`, uncommitted working tree on base
`82d1cfc`). Baseline for comparison: detached worktree at `82d1cfc`
(`.temp/TASK-260720-2x6mjn/review-baseline`). Nothing was staged or committed.

## What is solid

The planning core matches the AC and is well built.

- `src/csk/builds/planner.py` is genuinely side-effect free: provider-first + lexical
  command order, exact `GoBuildInput`, `cache_key`, target, artifact path, and all five
  outcomes (`cache-hit`, `would-preflight-and-build`, `would-rebuild-untrusted-cache`,
  `corrupt`, `unsupported`) via `CacheInspection.dry_run_outcome`.
- Gate ordering is asserted for both scopes
  (`test_build_planning_runs_only_after_validation_and_trust_gates`,
  `test_global_build_planning_runs_after_all_trust_gates`): freeze → collisions →
  system/skill deps → MCP → audit → registry → moved-tag → toolchain/cache.
- Dry-run acquires no mutation lock in either scope; asserted by the `ForbiddenLock`
  monkeypatches in `tests/test_cli.py` and `tests/test_global_install.py`.
- Before/after filesystem purity over project, `csk_home`, `skills_root` and `HOME`
  (`test_compiled_dry_run_preserves_every_persistent_surface`), plus `go_v1.build`,
  `consumers.record_consumer`, `gc.collect_runtime`, shim/env/adapter writers all
  monkeypatched to hard failures.
- Read-only registry paths hold: no `state/registry` creation, no catalog creation, no
  rollback-state advance, no HTTP cache creation or refresh.
- Generation recheck retries once then reports `concurrent_state_change`, at planner,
  project and global level.

Independently reproduced gates (my run, not the implementer's):

- `python -m pytest -q` → exit 0, **1009 passed, 92 skipped** in 99s
  (`.temp/TASK-260720-2x6mjn/review-pytest-full-01.log`)
- `python -m mypy` → exit 0, **Success: no issues found in 67 source files**
  (`.temp/TASK-260720-2x6mjn/review-mypy-01.log`)
- Focused suites `test_build_planner/test_audit_registry/test_install/test_global_install/test_cli`
  → exit 0, 182 passed, 1 skipped (`.temp/TASK-260720-2x6mjn/review-pytest-01.log`)

The green suite is exactly why the two blockers below matter: neither is covered by any
test, and both live in the *mutating* global-install path this task was supposed to leave
alone.

## Blocking findings

### B1 — `csk global install` now deletes every installed global skill when one declaration fails to resolve (data loss)

`src/csk/global_install.py:181` replaced the per-declaration `_build_plans` loop (which
recorded `f"{decl.name}: {exc}"` per bad decl and kept the good plans) with a single
`_build_nodes` → `closure.build_closure` call. On any resolution failure `_build_nodes`
records one error and returns `[]`, so `plans == []` and `build_providers == ()`.

`build_providers` being empty means the new `if result.errors and build_providers` guard
at `global_install.py:252` does **not** fire, so execution falls through into the
non-dry-run mutation path and reaches:

```python
installer._cleanup_removed_skills_root(
    global_skills_root(csk_home),
    {plan.decl.name for plan in plans},   # -> set()
)
```

`installer._cleanup_removed_skills_root` (`installer.py:1229`) `shutil.rmtree`s every
directory in the skills root not in the keep-set — so an empty keep-set wipes the whole
installed global skill tree. Adapter refresh then runs with an empty name list.

Failure scenario (probe:
`.temp/TASK-260720-2x6mjn/probe/test_probe_global_partial_failure.py`):
install `skill-a` globally, then add a second declaration `skill-missing` whose repo does
not exist, then run `csk global install` again.

| | status | errors | `global/skills/skill-a` |
|---|---|---|---|
| baseline `82d1cfc` | failed | `skill-missing: Skill repository not found …` | **present**, reported `up-to-date` |
| task branch | failed | `Skill repository not found for skill-missing … (via <project> -> skill-missing)` | **deleted** |

Direction: restore per-declaration error isolation, and independently make the keep-set
safe — a run that failed to build the closure must never narrow
`_cleanup_removed_skills_root`'s keep-set (union of declared manifest names and closure
plan names, or skip cleanup entirely when `result.errors` is non-empty). Note the
keep-set *did* need widening for transitive closure deps; the bug is the narrowing, not
the switch to plan names.

### B2 — one unrelated skill with a missing system dependency now blocks the whole global install

`global_install.py:252`:

```python
if result.errors and build_providers:
    result.status = "failed"
    return result
```

Any recorded error plus at least one surviving build provider aborts before a single
skill is installed. Baseline installed the healthy skills and still reported `failed`.
The `and build_providers` conjunct also makes failure semantics depend on whether some
*other* skill happens to declare a build command, which is not a defensible rule.

Failure scenario (probe:
`.temp/TASK-260720-2x6mjn/probe/test_probe_global_partial_build.py`): global manifest
declares `skill-build` (schema v6, `go-v1` build command, healthy) and `skill-bad`
(system dependency `__csk_missing_global_dependency__`), then `csk global install`.

| | status | `global/skills/skill-build` |
|---|---|---|
| baseline `82d1cfc` | failed | **installed** (`global: skill-build tag v1 … installed`) |
| task branch | failed | **not installed** |

Direction: this guard exists to stop build planning when the closure is degraded. Scope
it to that — skip/short-circuit `plan_builds` on errors — instead of aborting the install.

### B3 — new non-dry-run global gates ship with no test coverage

`_install_once` also newly runs `installer._check_audit_registries(...)` and
`installer._check_mcp_servers(...)` for **`csk global install`**, not just dry-run. In the
non-dry-run case `_check_audit_registries` calls
`audit_registry.migrate_snapshot_states` and `_write_snapshot_state`, i.e. `csk global
install` now writes and migrates `csk_home/state/registry`, and can newly mark registries
unavailable and fail the install. That may well be correct (the AC does put registry and
attestation gates ahead of planning), but no test asserts any of it, so the change is
invisible to the suite.

Either add tests for the new non-dry-run global behavior, or restrict the new gates to
the planning path and land them in the task that owns global trust gates.

Related: B1/B2/B3 are all outside this task's stated scope — "the planning portions of
… `src/csk/global_install.py`", with compilation/publication/markers/shims/target swaps
deferred downstream. Rework should keep the mutating global path byte-for-byte
behaviourally identical unless a change is explicitly required by the AC and tested.

## Non-blocking findings (fix during rework)

1. **Dead code.** `global_install._build_plans` (`global_install.py:433`) and
   `installer._detect_command_collisions` (`installer.py:599`) have no remaining callers.
2. **Duplicated activation logic.** `installer._active_build_command_names` reimplements
   `closure.ClosureNode.active_commands()` with a `type == "build"` filter instead of
   `"script"`. Two copies of the full/runtime edge rules will drift; parameterize the
   `ClosureNode` method by command type.
3. **Tautological assertion.**
   `test_compiled_dry_run_preserves_every_persistent_surface` asserts
   `observed_argv == [("go","telemetry","off"), ("go","version"), ("go","env")]`, but the
   fake session appends those tuples itself — it proves nothing about real Go
   invocations. The real purity signal there is the `go_v1.build` → `unexpected`
   monkeypatch. Either record argv at the subprocess boundary with the real
   `establish_toolchain`, or drop the argv assertion so it stops reading as coverage it
   is not.
4. **Dry-run vs install trust divergence.** Read-only mode skips
   `audit_registry.migrate_snapshot_states`, so on a manager home whose rollback state
   has not been migrated yet, dry-run reads an empty `state/registry` and cannot detect a
   rollback/equivocation that the subsequent real install *would* catch. Purity is the
   right call per AC; consider a read-only legacy fallback that reads
   `cache/registry/snapshot-*.json` without moving it, so the dry-run verdict matches the
   install verdict.
5. **Late-binding closure.** `planner._plan_once`'s `inspect_current` captures the loop
   variable `provider`. Safe today because it is invoked in the same iteration, but bind
   it explicitly (default arg or `functools.partial`).

## Reproducing the probes

```bash
cd /Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2x6mjn/probe
# task branch — both fail
CSK_PROBE_ROOT=../worktree \
  /Users/iv/Developer/intranet/cocoaskills/.venv/bin/python -m pytest -q -s \
  -p no:cacheprovider -o addopts= .
# baseline 82d1cfc — both pass
CSK_PROBE_ROOT=../review-baseline \
  /Users/iv/Developer/intranet/cocoaskills/.venv/bin/python -m pytest -q -s \
  -p no:cacheprovider -o addopts= .
```

Reviewer scratch state (delete when done):
`.temp/TASK-260720-2x6mjn/probe/`, `.temp/TASK-260720-2x6mjn/review-*.log`, and the
baseline worktree — `git worktree remove .temp/TASK-260720-2x6mjn/review-baseline`.

## Definition-of-Done status

| DoD item | Status |
|---|---|
| Base SHA recorded, worktree from fast-forwarded main + dependency handoff | met (`11160f6` → handoff `82d1cfc`, verified ancestry) |
| Before/after purity tests for forbidden dry-run paths and audit-before-build ordering | met |
| Focused pytest + strict mypy with task-scoped evidence | met, independently reproduced |
| Code per description and AC | **not met** — B1/B2 change out-of-scope mutating behavior |
| Tests for new/changed behavior | **not met** — B3, no coverage for new non-dry-run global gates |
| Lint clean | met (no ruff in env; compileall/tabnanny/mypy/`git diff --check` clean) |
| Build/validation run, build not broken | met (`python -m build`, `twine check`) |
| Outcome artifact attached with task-scoped name | met |
| Implementation matches AC | partially — planning core yes, global install regressions no |
| Solution fits project architecture | **not met** — global install loses per-declaration error isolation |
| Tests green | met (1009 passed, 92 skipped) — but green over an uncovered regression |
