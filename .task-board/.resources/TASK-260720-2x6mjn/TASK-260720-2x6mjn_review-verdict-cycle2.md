# TASK-260720-2x6mjn — reviewer verdict (cycle 2)

**Verdict: changes requested → `to-dev`**

Reviewer: reviewer role, RUN-260730-819f46 (not goal-bound).
Reviewed tree: `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-2x6mjn/worktree`
(branch `task/TASK-260720-2x6mjn-pure-build-planner`, uncommitted working tree on base
`82d1cfc`). Baseline for comparison: detached worktree at `82d1cfc`
(`.temp/TASK-260720-2x6mjn/review-baseline`). Nothing was staged or committed.
No `commit_ack` supplied (reviewer archetype).
Cycle-1 verdict: `TASK-260720-2x6mjn_review-verdict.md`.

## Cycle-1 blockers: all three resolved

| Cycle-1 blocker | Status | Evidence |
|---|---|---|
| B1 — one unresolvable global declaration wiped every installed global skill | **fixed** | cycle-1 probe now passes; `global_install._build_nodes` (`global_install.py:380`) isolates per-declaration failures, and `_cleanup_removed_skills_root` is suppressed whenever `result.errors` is non-empty (`global_install.py:310`). Keep-set widened to closure plan names. |
| B2 — one skill with a missing system dependency blocked the whole global install | **fixed** | cycle-1 probe now passes; the `if result.errors and build_providers: return failed` guard is gone, replaced by "skip `plan_builds`, keep installing healthy skills" (`global_install.py:252`). |
| B3 — new non-dry-run global gates had no coverage | **addressed** | `test_global_real_install_runs_mcp_and_registry_before_build_planning` and `test_global_real_install_requirement_gate_failure_blocks_planning_and_writes[mcp\|registry]` assert order, `alias="global"`, `read_only is False`, and that a failed gate blocks planning and all writes. |

Cycle-1 probes, re-run against the reworked tree (`review2-probe-worktree-01.log`):
`2 passed`.

Independently reproduced gates (my run):

- `python -m pytest -q` → exit 0, **1016 passed, 92 skipped** in 101s
  (`review2-pytest-full-01.log`)
- `python -m mypy` → exit 0, **Success: no issues found in 67 source files**
  (`review2-mypy-01.log`)

The planning core remains solid and unchanged in character: `planner.py` is side-effect
free, provider-first + command-lexical, all five outcomes, gate ordering asserted for both
scopes, dry-run takes no mutation lock, read-only registry/HTTP paths hold, generation
recheck retries then reports `concurrent_state_change`.

## New blocking findings

Both are the same failure mode as B1/B2 — the switch to `closure.build_closure` and to
running the planner on the **real** install path keeps changing the mutating install
behavior this task was scoped to leave alone ("Own … the *planning* portions of
`installer.py` and `global_install.py`"; "Defer compilation, publication, markers, shims,
and target swaps to downstream tasks"). Neither is covered by any test — the 1016-green
suite passes over both.

### C1 — `csk global install` now materializes the whole dependency closure: undeclared transitive skills, their runtime roots, and *all* their commands land in `global/bin`

Baseline global install built one plan per **declared** decl (`installer._build_plans` per
decl) and installed exactly those. The rework installs every `closure.build_closure` node:

```python
plans = [installer.SkillPlan(...) for node in nodes]      # global_install.py:188 — full closure
...
for plan in plans:                                        # global_install.py:290
    command_names = installer.install_runtime_commands(csk_home, csk_home / "global" / "bin", plan)
    installer._install_skill_context_to_root(global_skills_root(csk_home), plan, ...)
```

Two problems compound:

1. **Undeclared closure members are installed as global skills.** No declared/undeclared
   distinction — compare the project path, which routes non-declared closure nodes to the
   hybrid store (`installer.py:239`, `installer.py:322`) and only renders context when
   `node.context_active`.
2. **`install_runtime_commands` is called with no `only=` filter.** The project path passes
   `only=active` (`installer.py:316`) from `node.active_commands()`. Global passes nothing,
   so *inactive* commands of a *context-mode* dependency become `global/bin` shims — and
   `global/bin` is exported onto the operator PATH via
   `global_bins.refresh_user_bin_shims` / `env_files.write_global_env_files`.
   `closure.detect_active_command_collisions` only inspects **active** commands, so two
   inactive providers exporting the same name collide silently.

Failure scenario A (probe `probe2/test_probe_global_transitive_leak.py`): global Skillfile
declares only `consumer`; `consumer` declares `provider` with `"mode": "context"`;
`provider` exports script command `provider-tool`.

| | status | `global/skills` | `global/bin` |
|---|---|---|---|
| baseline `82d1cfc` | ok | `['consumer']` | `[]` |
| task branch | ok | `['consumer', 'provider']` | `['provider-tool']` |

Failure scenario B — silent shadowing (probe
`probe2/test_probe_global_inactive_shim_collision.py`): `consumer-one` → `provider-one`
and `consumer-two` → `provider-two`, both context-mode, both providers exporting a command
named `tool`.

| | status | `global/skills` | `global/bin` |
|---|---|---|---|
| baseline `82d1cfc` | ok | `['consumer-one', 'consumer-two']` | `[]` |
| task branch | **ok** (no collision error) | all four | `['tool']` |

The surviving `global/bin/tool` execs
`.cocoaskills/runtime/provider-two/<commit>/bin/tool` — `provider-one`'s identically named
command is silently shadowed, and runtime roots were materialized for a skill the operator
never declared. `status` is `ok`.

Direction: keep the closure for **planning** (it is genuinely needed — a build command may
be activated on a transitive provider), but drive the global materialization loop from the
declared set, exactly as baseline did. If global closure materialization is actually
wanted, it belongs in the task that owns global materialization, with activation filtering
(`only=node.active_commands()`), a declared/undeclared distinction, and collision
detection that covers installed-but-inactive commands.

### C2 — real `csk install` / `csk global install` now hard-fails with `go_toolchain_missing` when no Go is on PATH, installing nothing at all

`plan_builds` is called on the **non-dry-run** path too (`installer.py:272`,
`global_install.py:258`). `toolchain.establish_toolchain` raises `ToolchainError`
(`go_toolchain_missing`) when the captured operator PATH has no Go; that is not a
`BuildPlanningError`, so it falls to the generic `except Exception` and fails the entire
result. Nothing is installed — not the build skill, not unrelated healthy skills.

Failure scenario (probes `probe2/test_probe_real_install_requires_go.py` and
`probe2/test_probe_global_requires_go.py`): PATH holds `git`/`sh`/`env`/`uname` but no
`go`; one skill declares a schema-v6 `go-v1` build command, plus (global case) one
unrelated plain skill.

| | project install | global install |
|---|---|---|
| baseline `82d1cfc` | ok, `.agents/skills/skill-build` installed | ok, `['skill-build', 'skill-plain']` installed |
| task branch | failed — `go-v1 go_toolchain_missing: captured operator PATH contains no Go executable`, nothing installed | failed — same error, `global/skills` does not exist |

This also breaks **pre-existing** tests in a Go-less environment. Same machine, `go`
removed from PATH only (`.temp/TASK-260720-2x6mjn/nogo-bin`):

```
baseline    tests/test_install.py test_global_install.py test_cli.py  → 125 passed
task branch same three files + test_build_planner.py                  → 4 failed, 139 passed
```

Failures: `test_schema_v6_build_root_stays_out_of_dry_run_real_and_up_to_date_context`,
`test_schema_v6_stale_build_root_forces_context_reinstall[physical-root]`,
`[marker-entry]`, `[pre-exclusion-tree]` — all pre-existing, none touched by this task.
`grep -rn go_toolchain_missing tests/` returns nothing, so the missing/unusable-toolchain
case is unasserted on every path. No workflow in `.github/workflows/` runs `setup-go`;
GitHub-hosted runners ship Go preinstalled, so CI probably stays green — the dependency is
now real but undeclared. Worth weighing against the immediate base commit `82d1cfc`,
"fix(builds): make Go driver CI portable".

Note the plan is then **discarded** on the real path: nothing reads `result.builds` outside
the dry-run print loops. The real install currently pays the toolchain probe and cache
inspection, gains nothing, and acquires a new total-failure mode.

Direction — pick one:
- restrict `plan_builds` to the dry-run path in this task and let
  TASK-260720-3t8nr3 ("Own the project and hybrid materialization path") wire the real
  path together with the compile step that consumes the plan; or
- keep it on both paths and (a) cover missing/unusable-toolchain for project and global
  scope, (b) isolate the failure so unrelated healthy skills still install — the same rule
  B2's fix established, and (c) declare the Go dependency for the four pre-existing tests
  that now need it.

## Non-blocking findings

Carried over from cycle 1, still open:

1. **Dead code.** `installer._detect_command_collisions` (`installer.py:599`) now has no
   callers anywhere in `src/` or `tests/` — global install was its last one.
   (`global_install._build_plans` is gone; that half is resolved.)
2. **Duplicated activation logic.** `installer._active_build_command_names`
   (`installer.py:462`) still reimplements `closure.ClosureNode.active_commands()` with a
   `type == "build"` filter. Parameterize the `ClosureNode` method by command type.
3. **Tautological assertion.** `tests/test_install.py:533` still asserts
   `observed_argv == [("go","telemetry","off"), ("go","version"), ("go","env")]` against a
   list the fake session appends itself.
4. **Dry-run vs install trust divergence.** Read-only mode skips
   `audit_registry.migrate_snapshot_states`, so on an unmigrated manager home dry-run
   cannot see a rollback the real install would catch. Purity is right per AC; a read-only
   legacy fallback would align the verdicts.
5. **Late-binding closure.** `planner._plan_once`'s `inspect_current` (`planner.py:371`)
   still captures the loop variable `provider`.

New this cycle:

6. **Read-only registry state is stricter than the install.**
   `audit_registry._validate_read_only_state_directory` raises on broad permissions or a
   foreign owner, where the non-read-only path `mkdir`s + `chmod 0o700`s and proceeds. On a
   mis-permissioned `state/registry`, dry-run reports registries unavailable while the real
   install succeeds.
7. **`result.builds` has no consumer on the real path** (see C2). If the plan is meant to
   be computed and dropped until TASK-260720-3t8nr3, say so in a comment.

## Reproducing

```bash
cd /Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-2x6mjn/probe2
# task branch
CSK_PROBE_ROOT=../worktree /Users/iv/Developer/Wildberries/cocoaskills/.venv/bin/python \
  -m pytest -q -s -p no:cacheprovider -o addopts= .
# baseline 82d1cfc
CSK_PROBE_ROOT=../review-baseline /Users/iv/Developer/Wildberries/cocoaskills/.venv/bin/python \
  -m pytest -q -s -p no:cacheprovider -o addopts= .
```

The probes print state rather than asserting a verdict (except
`test_probe_real_install_requires_go`) — compare the two runs' `STATUS` / `installed dirs`
/ `global bins` lines.

Go-less suite comparison:

```bash
PATH=/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-2x6mjn/nogo-bin \
  .venv/bin/python -m pytest -q tests/test_install.py tests/test_global_install.py tests/test_cli.py
```

`nogo-bin` is a symlink farm of the whole PATH minus `go`/`gofmt`.

`probe2/test_probe_global_validation_gate.py::test_probe_global_install_with_skill_missing_skill_md`
is a **non-finding** kept for the record: a global skill snapshot without `SKILL.md` fails
the whole install on baseline too (the task branch only prefixes the skill name onto the
error, which is an improvement).

Reviewer scratch state to delete when done: `.temp/TASK-260720-2x6mjn/probe/`,
`probe2/`, `nogo-bin/`, `review2-*.log`, and the baseline worktree
(`git worktree remove .temp/TASK-260720-2x6mjn/review-baseline`).

Findings were not written to the repository `LOGBOOK.md`: this review is read-only and must
not add uncommitted changes to the tree under review. Record C1/C2 there during rework.

## Definition-of-Done status

| DoD item | Status |
|---|---|
| Base SHA recorded, worktree from fast-forwarded main + dependency handoff | met |
| Before/after purity tests for forbidden dry-run paths and audit-before-build ordering | met |
| Focused pytest + strict mypy with task-scoped evidence | met, independently reproduced |
| Code per description and AC | **not met** — C1/C2 change out-of-scope mutating behavior |
| Tests for new/changed behavior | **not met** — C1 and C2 uncovered; C2 breaks 4 pre-existing tests without Go |
| Lint clean | met (no ruff in env; compileall/tabnanny/mypy/`git diff --check` clean) |
| Build/validation run, build not broken | met |
| Outcome artifact attached with task-scoped name | met |
| Implementation matches AC | partially — planning core yes, install-path side effects no |
| Solution fits project architecture | **not met** — global materialization ignores the declared/undeclared and activation model the project path enforces |
| Tests green | met (1016 passed, 92 skipped) — green over both new findings |
