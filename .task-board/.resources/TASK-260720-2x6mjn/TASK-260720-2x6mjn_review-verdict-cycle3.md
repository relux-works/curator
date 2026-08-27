# TASK-260720-2x6mjn — reviewer verdict (cycle 3)

**Verdict: changes requested → `to-dev`**

Reviewer: reviewer role, RUN-260730-ccb11a (not goal-bound).
Reviewed tree: `/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2x6mjn/worktree`
(branch `task/TASK-260720-2x6mjn-pure-build-planner`, uncommitted working tree on base
`82d1cfc`, 9 modified + 3 untracked files, `+2040/-49`).
Baseline for comparison: detached worktree at `82d1cfc`
(`.temp/TASK-260720-2x6mjn/review-baseline`).
Nothing was staged or committed. No `commit_ack` supplied (reviewer archetype).
Prior verdicts: `TASK-260720-2x6mjn_review-verdict.md` (cycle 1),
`TASK-260720-2x6mjn_review-verdict-cycle2.md` (cycle 2).

Note on provenance: the board lives in `/Users/iv/Developer/ReluxWorks/curator/.task-board`,
but the code for this task lives in the `cocoaskills` repo
(`/Users/iv/Developer/intranet/cocoaskills`). All paths below are relative to that repo.

## Cycle-2 blockers: both resolved

| Cycle-2 blocker | Status | Evidence |
|---|---|---|
| C1 — `csk global install` materialized the whole dependency closure (undeclared transitive skills, their runtime roots, all their commands into `global/bin`, silent shim shadowing) | **fixed** | Real global install is declaration-driven again: `global_install.install` splits dry-run (`_build_nodes` → closure → planning) from the real path (`_build_plans` per decl, `installer._detect_command_collisions`) at `global_install.py:185`. Both cycle-2 probes now match baseline exactly (table below). |
| C2 — real `csk install` / `csk global install` hard-failed with `go_toolchain_missing` when no Go was on PATH | **fixed** | `operator_search_path` is captured only when `options.dry_run` (`installer.py:69`, `global_install.py:136`); `_freeze_build_providers` + `plan_builds` are both behind `if options.dry_run` (`installer.py:252`, `global_install.py:283`). Both cycle-2 probes pass; Go-less suite failures dropped from 4 to 1. |

Cycle-2 probe archive (`TASK-260720-2x6mjn_review2-probes.tar.gz`), extracted fresh and
re-run unmodified against the reworked tree (`review3-probe2-worktree-01.log`):
`1 failed, 5 passed`.

| probe | baseline `82d1cfc` | task branch (cycle 3) |
|---|---|---|
| `test_probe_global_transitive_leak` | ok, `global/skills=['consumer']`, `global/bin=[]` | ok, `global/skills=['consumer']`, `global/bin=[]` |
| `test_probe_global_inactive_shim_collision` | ok, `['consumer-one','consumer-two']`, `global/bin=[]` | ok, `['consumer-one','consumer-two']`, `global/bin=[]` |
| `test_probe_global_requires_go` | ok, `['skill-build','skill-plain']` | ok, `['skill-build','skill-plain']` |
| `test_probe_real_install_requires_go` | pass | pass |

The single failure is `test_probe_global_validation_gate.py::test_probe_global_install_with_skill_missing_skill_md`,
documented in cycle 2 as a **non-finding**. Independently re-confirmed against baseline
`82d1cfc` this cycle (`review3-probe2-baseline-validationgate-01.log`): `1 failed, 1 passed`,
`STATUS=failed`, `skill-ok installed: False` — identical on base, not introduced here.

Independently reproduced gates (my run, canonical `.venv/bin/python` 3.14.4, rootdir =
worktree so `pythonpath=["src"]` resolves to the worktree source):

- `python -m pytest -q` → exit 0, **1020 passed, 92 skipped** in 168s (`review3-pytest-full-01.log`)
- `python -m mypy` → exit 0, **Success: no issues found in 67 source files** (`review3-mypy-01.log`)

The planning core is unchanged in character and still holds: `planner.py` is side-effect
free, provider-first + command-lexical, produces all five outcomes, gates run before any
toolchain probe or cache lookup in both scopes, dry-run takes no `GlobalLock`
(`cli.py:565` project, `cli.py:617` global; both covered by `tests/test_cli.py:945,965`),
read-only registry/HTTP paths hold, `plan_builds` never reaches `go list` / `go build`
(`go_v1.build` asserted unexpected at `tests/test_install.py:560`), and generation recheck
retries then reports `concurrent_state_change`.

Verified the two things the C1 fix could plausibly have gotten wrong, both clean:
`installer._build_plans` (`installer.py:550`) iterates only `project_manifest.skills` and
does not expand dependencies, so the real global `plans` list is exactly the declared set;
and `_cleanup_removed_skills_root`'s widened keep-set (`{plan.decl.name for plan in plans}`,
`global_install.py:342`) is equivalent to baseline's `{decl.name for decl in
global_manifest.skills}` whenever cleanup runs, because every plan dropped by
`_plans_with_available_dependencies` appends to `result.errors` (`global_install.py:533`)
and cleanup is skipped when `result.errors` is non-empty.

## Blocking finding

### F1 — dry-run now hard-requires Go, that failure mode has zero test coverage, and it leaves a pre-existing test red on any machine without Go

This is the unaddressed remainder of C2. Cycle 2 offered two directions and the rework took
direction (a) — restrict `plan_builds` to dry-run — which is the right call and fixed the
real-install regression. But dry-run still probes Go, so cycle-2 item (a) of the *other*
direction still applies to the dry-run path: **cover missing/unusable toolchain for project
and global scope.** It was not done, and `results.md` and `LOGBOOK.md` both report only the
half that was ("real project/global installs succeed without Go on `PATH`"), so the residue
reads as closed when it is not.

Concretely:

1. **No coverage.** `grep -rn go_toolchain_missing tests/` → **no hits**. Neither project
   nor global dry-run has a test for an absent or unusable Go toolchain, in a task whose
   deliverable is the dry-run Go probe. `ToolchainError` is not a `BuildPlanningError`, so
   it lands in the generic `except Exception` (`installer.py:406`, `global_install.py:368`)
   and fails the **entire** project/global dry-run — the operator loses the validation
   verdict for every unrelated skill too. That whole-plan-failure shape is exactly the rule
   the B2 fix established for the real path, and it is now unasserted on the dry-run path.

2. **A pre-existing test is red without Go.** Same machine, `go`/`gofmt` removed from PATH
   only (`.temp/TASK-260720-2x6mjn/nogo-bin`, a symlink farm of the full PATH):

   ```
   PATH=.temp/TASK-260720-2x6mjn/nogo-bin .venv/bin/python -m pytest -q \
     tests/test_install.py tests/test_global_install.py tests/test_cli.py tests/test_build_planner.py
   → 1 failed, 146 passed          (review3-nogo-worktree-01.log)
   ```

   `tests/test_install.py::test_schema_v6_build_root_stays_out_of_dry_run_real_and_up_to_date_context`
   (`tests/test_install.py:246`, pre-existing, untouched by this task) fails on
   `assert not dry_run.errors` with
   `['go-v1 go_toolchain_missing: captured operator PATH contains no Go executable']`.
   That test is about build roots staying out of dry-run/real/up-to-date context — it has
   nothing to do with toolchains, and it does not stub one.

   Cycle 2 measured 4 such failures; this is 1 of the original 4, still open. Baseline
   `82d1cfc` was 125 passed / 0 failed on the same three files.

3. **The dependency is real but undeclared.** No `setup-go` or `go-version` step anywhere in
   `.github/workflows/` (`ci.yml`, `distribution-smoke.yml`, `release.yml`), and no
   Go-availability guard in the suite — `tests/` has `skipif` only for `sys.platform ==
   "win32"`. CI probably stays green because GitHub-hosted runners ship Go preinstalled, so
   this fails silently for contributors, not in CI. Worth weighing against the immediate
   base commit `82d1cfc`, "fix(builds): make Go driver CI portable".

I checked whether the in-model alternative exists and it does not: `unsupported` is a
`CacheEntryStatus` mapped in `cache.CacheInspection.dry_run_outcome` (`cache.py:92`) and is
only reachable *after* a toolchain identity exists, so a missing toolchain cannot be
reported as an `unsupported` plan row. Failing the plan is the correct design — this is a
coverage and portability gap, not an architecture problem, and no production change is
needed.

Direction (test-only, and the pattern already exists in this tree):

- Make the pre-existing test toolchain-independent the same way this task's own purity test
  already does — `monkeypatch.setattr(build_toolchain, "establish_toolchain", ...)` plus a
  fake `capture_operator_search_path` (`tests/test_install.py:550`) — or guard it on Go
  availability with an explicit reason.
- Add dry-run `go_toolchain_missing` coverage for **both** scopes: assert the error surfaces
  with a stable message, that no mutation occurs, and pin the intended blast radius (whole
  plan fails vs. unrelated skills still validate). Pick one and assert it; today it is
  whole-plan-fails by accident of `except Exception` placement.
- If any of this is deliberately deferred, say so explicitly in `results.md` and
  `LOGBOOK.md` rather than leaving the cycle-2 item looking closed.

## Non-blocking findings

Carried over from cycles 1–2, still open:

1. **Dead code.** `installer._detect_command_collisions` (`installer.py`) — still live, now
   used by the real global path again, so cycle 1's item is **resolved** by the C1 fix.
2. **Duplicated activation logic.** `installer._active_build_command_names`
   (`installer.py:473`) reimplements `closure.ClosureNode.active_commands()` with a
   `type == "build"` filter. Parameterize the `ClosureNode` method by command type.
3. **Tautological assertion.** `tests/test_install.py` still asserts `observed_argv ==
   [("go","telemetry","off"), ("go","version"), ("go","env")]` against a list the fake
   session appends itself.
4. **Dry-run vs install trust divergence.** Read-only mode skips
   `audit_registry.migrate_snapshot_states` (`installer.py:691`), so on an unmigrated
   manager home dry-run cannot see a rollback the real install would catch. Purity is right
   per AC; a read-only legacy fallback would align the verdicts.
5. **Late-binding closure.** `planner._plan_once`'s `inspect_current` (`planner.py:371`)
   captures the loop variable `provider`. Currently harmless — it is invoked synchronously
   in the same iteration via `provider.snapshot.use(...)` — but it is a trap for anyone who
   later defers the call.
6. **Read-only registry state is stricter than the install.**
   `audit_registry._validate_read_only_state_directory` (`audit_registry.py:1013`) raises on
   broad permissions or a foreign owner, where the non-read-only path `mkdir`s + `chmod
   0o700`s and proceeds. On a mis-permissioned `state/registry`, dry-run reports registries
   unavailable while the real install succeeds. (Also: `check_snapshots`' second positional
   is `state_dir` but the parameter is named `cache_dir` — pre-existing misnomer, now load-
   bearing for the read-only branch. Worth a rename.)
7. **`result.builds` has no consumer on the real path** — now moot for the real path (it is
   never computed there), but `ProjectResult.builds` / `GlobalResult.builds` are still only
   read by the dry-run print loops. `LOGBOOK.md` names TASK-260720-3t8nr3 as the owner, which
   closes cycle 2's ask.

New this cycle:

8. **Global dry-run is stricter than global install.** `_check_mcp_servers` and
   `_check_audit_registries` now run on the global dry-run path only
   (`global_install.py:249`, `:271`). `csk global install --dry-run` can therefore fail on a
   registry/MCP condition that `csk global install` never checks — the inverse of the usual
   "dry-run is a preview of the real thing" contract. Cycle 2's B3 tests that asserted these
   gates on the real global path were converted to dry-run tests
   (`test_global_dry_run_requirement_gate_failure_blocks_planning_and_writes`,
   `tests/test_global_install.py:729`), so nothing asserts the divergence either way.
9. **`csk update --dry-run` now validates the skills root.** `cli.py:565` narrows
   `dry_run_operation` to `{install, upgrade}`, so `update` falls through to
   `config.validate_skills_root_for_work(cfg)`; baseline skipped it for any `--dry-run`.
   `--dry-run` is accepted on `update` (shared `_add_install` parser, `cli.py:373`) and
   `_cmd_update` ignores it, so `update --dry-run` always fetched for real — validating is
   arguably the correct fix, but it is an unmentioned CLI behavior change. No test asserts
   either behavior.
10. **Generation probe cost.** `FilesystemGenerationProbe` content-hashes whole directory
    trees (`csk_home/builds`, `audit`, `cache/registry`, `state/registry`, the project's
    `.agents/skills`, the hybrid skills root, `global/skills`) and does so twice per
    planning attempt. On a large install the dry-run pays two full-tree hashes. Correct per
    AC, but worth a cheap-stat fast path before this lands on real installs in
    TASK-260720-3t8nr3.

## Reproducing

```bash
cd /Users/iv/Developer/intranet/cocoaskills
VENV=.venv/bin/python
WT=.temp/TASK-260720-2x6mjn/worktree

# gates
(cd $WT && ../../../.venv/bin/python -m pytest -q && ../../../.venv/bin/python -m mypy)

# cycle-2 probes, extracted fresh from the attached archive
mkdir -p .temp/TASK-260720-2x6mjn/review3 && cd .temp/TASK-260720-2x6mjn/review3
tar xzf <board>/.resources/TASK-260720-2x6mjn/TASK-260720-2x6mjn_review2-probes.tar.gz
cd probe2
CSK_PROBE_ROOT=../../worktree         $VENV -m pytest -q -s -p no:cacheprovider -o addopts= .
CSK_PROBE_ROOT=../../review-baseline  $VENV -m pytest -q -s -p no:cacheprovider -o addopts= .

# F1 — Go-less suite
cd $WT
PATH=/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2x6mjn/nogo-bin \
  $VENV -m pytest -q tests/test_install.py tests/test_global_install.py \
  tests/test_cli.py tests/test_build_planner.py
```

`nogo-bin` is a symlink farm of the whole PATH minus `go`/`gofmt`.
The C1 probes print state rather than asserting; compare the two runs' `STATUS=` /
`installed dirs:` / `global bins:` lines.

Reviewer scratch to delete when the task closes: `.temp/TASK-260720-2x6mjn/probe/`,
`probe2/`, `review3/`, `nogo-bin/`, `review2-*.log`, and the baseline worktree
(`git worktree remove .temp/TASK-260720-2x6mjn/review-baseline`).

F1 was not written to the repository `LOGBOOK.md`: this review is read-only and must not add
uncommitted changes to the tree under review. Record it there during rework.

## Definition-of-Done status

| DoD item | Status |
|---|---|
| Base SHA recorded, worktree from fast-forwarded main + dependency handoff | met |
| Before/after purity tests for forbidden dry-run paths and audit-before-build ordering | met |
| Focused pytest + strict mypy with task-scoped evidence | met, independently reproduced |
| Code per description and AC | met |
| Tests for new/changed behavior and passing | **not met** — F1: dry-run `go_toolchain_missing` uncovered in both scopes; one pre-existing test red without Go |
| Lint clean | met (no ruff in env / no ruff config; mypy strict clean) |
| Build/validation run, build not broken | met (1020 passed, 92 skipped; mypy 0) |
| Outcome artifact attached with task-scoped name | met |
| Important findings recorded in logbook | met for C1/C2; F1 still to record |
| Implementation matches AC | met |
| Solution fits project architecture | met — C1's declared/undeclared and activation boundary is restored |
| Tests green | met on this machine (Go present); **red without Go** |
