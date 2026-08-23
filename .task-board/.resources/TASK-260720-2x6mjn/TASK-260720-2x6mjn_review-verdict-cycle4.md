# TASK-260720-2x6mjn — reviewer verdict (cycle 4)

**Verdict: accepted**

Reviewer: reviewer role, RUN-260730-b68d14 (not goal-bound — `task-board spawn goal` reports
"run is not goal-bound").
Reviewed tree: `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-2x6mjn/worktree`
(branch `task/TASK-260720-2x6mjn-pure-build-planner`, uncommitted working tree on base
`82d1cfc`, 9 modified + 3 untracked files, `+2235/-50`).
Nothing was staged, committed, or modified by this review. No `commit_ack` supplied
(reviewer archetype — the commit-owning mover records it).
Prior verdicts: cycle 1 `TASK-260720-2x6mjn_review-verdict.md`, cycle 2
`..._review-verdict-cycle2.md`, cycle 3 `..._review-verdict-cycle3.md`.

Provenance note: the board lives in `/Users/iv/Developer/ReluxWorks/curator/.task-board`,
the code lives in `/Users/iv/Developer/Wildberries/cocoaskills`. Paths below are relative
to the code repo.

## Cycle-3 blocker F1: resolved

F1 was "dry-run now hard-requires Go, that failure mode has zero test coverage, and it
leaves a pre-existing test red on any machine without Go". All three sub-items are closed.

| F1 sub-item | Status | Evidence |
|---|---|---|
| No dry-run `go_toolchain_missing` coverage in either scope | **fixed** | `grep -rn go_toolchain_missing tests/` → 2 hits: `tests/test_install.py:384` (project), `tests/test_global_install.py:752` (global). Both assert the exact stable message `go-v1 go_toolchain_missing: captured operator PATH contains no Go executable`. |
| Blast radius unasserted (whole plan fails vs. unrelated skills still validate) | **fixed** | Both tests plan a build skill *plus* an unrelated `skill-plain`, then pin whole-plan failure: `result.status == "failed"`, `result.errors == [<single exact message>]`, `result.builds == []`, and `not any("(planned)" in m for m in result.messages)`. The chosen semantics (whole plan fails) is now explicit rather than an accident of `except Exception` placement. |
| Pre-existing test red without Go | **fixed** | `tests/test_install.py::test_schema_v6_build_root_stays_out_of_dry_run_real_and_up_to_date_context` (`tests/test_install.py:277`) now calls `_stub_trusted_toolchain(monkeypatch)` (`tests/test_install.py:56`), which patches both `capture_operator_search_path` and `establish_toolchain` with a deterministic fake `ToolchainIdentity`/`NativeTarget`. The test stays about build-root context filtering and no longer inherits the host's Go install. |
| Residue documented | **fixed** | `LOGBOOK.md:23-30` records the fail-closed decision, the whole-plan blast radius, preserved watched filesystem surface, and the host-independent test boundary. `results.md` has a "Review cycle 4 — F1 test-only rework" section. |

Both new tests use a byte-and-metadata filesystem snapshot (`_filesystem_state`,
`tests/test_install.py:23` — `lstat` mode, regular-file bytes, symlink targets, directory
entries) over `(project, csk_home, skills_root, Path.home())` and assert `before == after`,
so the fail-closed path is pinned as mutation-free, not just error-shaped.

**Non-vacuity check.** The tests are not no-ops. `installer.py:72` and `global_install.py:137`
both call `build_toolchain.capture_operator_search_path()` via module attribute, so the
monkeypatch binds. On this Go-equipped machine the full suite is green *including* these two
tests, which is only possible if the empty-`OperatorSearchPath` patch actually took effect
and drove the planner into the `go_toolchain_missing` branch.

## Cycle 4 is test-only, as claimed

Production sources are untouched since the cycle-3 review (verdict written 18:04); only
tests and the logbook moved afterwards:

```
17:44:46  src/csk/installer.py          17:45:44  src/csk/global_install.py
16:31:39  src/csk/audit_registry.py     16:32:13  src/csk/builds/planner.py
16:25:11  src/csk/cli.py                16:23:48  src/csk/builds/__init__.py
--- cycle-3 verdict 18:04 ---
18:11:26  tests/test_install.py         18:11:49  tests/test_global_install.py
18:11:57  LOGBOOK.md
```

Suite count corroborates: cycle 3 was `1020 passed`, cycle 4 is `1022 passed` — exactly the
two new scope tests; the schema-v6 test was edited in place, not added.

## Independently reproduced gates

Canonical interpreter `/Users/iv/Developer/Wildberries/cocoaskills/.venv/bin/python` (3.14.4),
rootdir = the task worktree so `pythonpath=["src"]` resolves to the worktree source.

| Gate | Result | Log |
|---|---|---|
| `python -m pytest -q` | exit 0 — **1022 passed, 92 skipped** in 94.94s | `review4/pytest-full-01.log` |
| `python -m mypy` | exit 0 — **Success: no issues found in 67 source files** | (stdout) |
| Go-less focused suite (`test_install.py`, `test_global_install.py`, `test_cli.py`, `test_build_planner.py`) | exit 0 — **149 passed** in 42.88s | `review4/nogo-worktree-01.log` |
| Cycle-2 probe archive, extracted fresh and run unmodified | exit 1 — **1 failed, 5 passed** | `review4/probe2-worktree-01.log` |

The Go-less run used `.temp/TASK-260720-2x6mjn/nogo-bin`, a 3081-entry symlink farm of the
full `PATH` with `go`/`gofmt` removed (verified: `command -v go gofmt` finds nothing under it).
Cycle 3 measured `1 failed, 146 passed` there; it is now `149 passed, 0 failed` — the last of
the original 4 Go-less failures is gone and the 2 new tests are included.

The single probe failure is
`test_probe_global_validation_gate.py::test_probe_global_install_with_skill_missing_skill_md`,
documented in cycle 2 as a non-finding and re-confirmed in cycle 3 against detached base
`82d1cfc` (`1 failed, 1 passed` there too). Not introduced by this task.

C1/C2 probe outcomes are byte-identical to the cycle-3 accepted state:

| probe | cycle 3 (accepted) | cycle 4 |
|---|---|---|
| `test_probe_global_transitive_leak` | `global/skills=['consumer']`, `global/bin=[]` | same |
| `test_probe_global_inactive_shim_collision` | `['consumer-one','consumer-two']`, `global/bin=[]` | same |
| `test_probe_global_requires_go` | `['skill-build','skill-plain']` | same |
| `test_probe_real_install_requires_go` | pass | pass |

## Acceptance-criteria re-verification

Spot-checked the load-bearing AC clauses against the current tree (all still hold):

- **Gates before any toolchain probe or cache lookup.** Project: validation → closure/collision
  → dependencies → MCP → migration warnings → audit gate → registry/attestation → moved-tag,
  and only then the `if options.dry_run:` `plan_builds` block (`installer.py:249-289`). Global:
  the same ordering at `global_install.py:205-300`. `capture_operator_search_path`
  (`builds/toolchain.py:113`) is a pure `os.environ["PATH"]` split with no filesystem access,
  so hoisting it to `install()` entry is not a toolchain probe — it is exactly the
  "capture at process entry" contract its own docstring states.
- **Dry-run takes no mutation/project lock.** `cli.py:564-571` (project) and `cli.py:617-621`
  (global) return before `config.validate_skills_root_for_work` / `GlobalLock`; covered by
  `tests/test_cli.py:945,965`.
- **Real installs defer planning and do not require Go.** `operator_search_path` is `None`
  unless `options.dry_run` (`installer.py:71`, `global_install.py:136`); `_freeze_build_providers`
  and `plan_builds` are both dry-run-gated. Asserted by
  `test_real_install_does_not_plan_builds_or_require_go` (`tests/test_install.py:391`) and
  `test_global_real_install_defers_build_planning_and_planning_only_gates`
  (`tests/test_global_install.py:759`).
- **Declaration-driven real global install** (cycle-2 C1) still holds via the
  `_build_plans` / `_detect_command_collisions` branch at `global_install.py:228-238`.
- Planner purity, provider-first + command-lexical ordering, all five outcomes, read-only
  registry/HTTP paths, no `go list` / `go build`, and generation recheck →
  `concurrent_state_change` were verified in cycle 3 against production code that has not
  changed since.

## Non-blocking findings carried forward

None of these block acceptance. Items 1 and F1 from earlier cycles are now closed; the rest
are unchanged from cycle 3 and are recorded here so they are not lost.

1. **Duplicated activation logic.** `installer._active_build_command_names` reimplements
   `closure.ClosureNode.active_commands()` with a `type == "build"` filter. Parameterize the
   `ClosureNode` method by command type.
2. **Tautological assertion.** `tests/test_install.py` asserts `observed_argv == [("go","telemetry","off"),
   ("go","version"), ("go","env")]` against a list the fake session appends itself.
3. **Dry-run vs install trust divergence.** Read-only mode skips
   `audit_registry.migrate_snapshot_states` (`installer.py:691`), so on an unmigrated manager
   home dry-run cannot see a rollback the real install would catch. Purity is correct per AC;
   a read-only legacy fallback would align the verdicts.
4. **Late-binding closure.** `planner._plan_once`'s `inspect_current` (`planner.py:371`) captures
   the loop variable `provider`. Harmless today (invoked synchronously in the same iteration)
   but a trap for anyone who later defers the call.
5. **Read-only registry state is stricter than the install.**
   `audit_registry._validate_read_only_state_directory` (`audit_registry.py:1013`) raises on broad
   permissions or a foreign owner where the non-read-only path `mkdir`s + `chmod 0o700`s and
   proceeds. Also: `check_snapshots`' second positional is a `state_dir` but the parameter is
   named `cache_dir` — pre-existing misnomer, now load-bearing for the read-only branch.
6. **`result.builds` has no consumer on the real path.** `ProjectResult.builds` / `GlobalResult.builds`
   are read only by the dry-run print loops. `LOGBOOK.md` names TASK-260720-3t8nr3 as the owner.
7. **Global dry-run is stricter than global install.** `_check_mcp_servers` (`global_install.py:249`)
   and `_check_audit_registries` (`global_install.py:271`) run on the global dry-run path only, so
   `csk global install --dry-run` can fail on a registry/MCP condition that `csk global install`
   never checks — the inverse of the usual "dry-run previews the real thing" contract. Deliberate
   per the AC (MCP/system requirements are planning gates) and per the deferral to
   TASK-260720-3t8nr3, but the divergence is asserted in neither direction. Worth an explicit
   decision when 3t8nr3 wires planning into the real path.
8. **`csk update --dry-run` now validates the skills root.** `cli.py:564` narrows `dry_run_operation`
   to `{install, upgrade}`, so `update` falls through to `config.validate_skills_root_for_work(cfg)`;
   baseline skipped it for any `--dry-run`. `--dry-run` is accepted on `update` (shared `_add_install`
   parser) but `_cmd_update` ignores it. Arguably the correct fix, still an unmentioned CLI behavior
   change with no test either way.
9. **Generation probe cost.** `FilesystemGenerationProbe` content-hashes whole directory trees
   (`csk_home/builds`, `audit`, `cache/registry`, `state/registry`, the project's `.agents/skills`,
   the hybrid skills root, `global/skills`) twice per planning attempt. Correct per AC; worth a
   cheap-stat fast path before this lands on real installs in TASK-260720-3t8nr3.
10. **Go is an undeclared test dependency in CI.** No `setup-go` / `go-version` step in
    `.github/workflows/`. This is now benign for the four files covered by the Go-less run above,
    but the *full* suite has not been measured without Go this cycle; other test files may still
    depend on a host Go. Cheap follow-up: add `actions/setup-go` or a Go-availability `skipif`.

## Definition-of-Done status

| DoD item | Status |
|---|---|
| Base SHA recorded, worktree from fast-forwarded main + dependency handoff | met — `origin/main` `11160f6`, task base `82d1cfc`, branch `task/TASK-260720-2x6mjn-pure-build-planner` |
| Before/after purity tests for forbidden dry-run paths and audit-before-build ordering | met |
| Focused pytest + strict mypy with task-scoped evidence | met — independently reproduced |
| Code per description and AC | met |
| Tests for new/changed behavior and passing | **met** — F1 closed; both scopes cover `go_toolchain_missing` |
| Lint clean | met — no ruff installed and no ruff config in repo; strict mypy clean, `compileall`/`tabnanny`/`git diff --check` all exit 0 per results.md |
| Build/validation run, build not broken | met — 1022 passed / 92 skipped, mypy 0, `python -m build` + `twine check` green |
| Outcome artifact attached with task-scoped name | met — `results.md` updated with the cycle-4 section |
| Important findings recorded in logbook | met — `LOGBOOK.md` records the fail-closed blast radius and host-independent test boundary |
| Implementation matches AC | met |
| Solution fits project architecture | met |
| Tests green | **met** — green with and without Go on `PATH` |

## Reproducing

```bash
REPO=/Users/iv/Developer/Wildberries/cocoaskills
VENV=$REPO/.venv/bin/python
WT=$REPO/.temp/TASK-260720-2x6mjn/worktree

# gates
(cd $WT && $VENV -m pytest -q && $VENV -m mypy)

# F1 — Go-less focused suite (nogo-bin = full PATH symlink farm minus go/gofmt)
(cd $WT && PATH=$REPO/.temp/TASK-260720-2x6mjn/nogo-bin $VENV -m pytest -q \
   tests/test_install.py tests/test_global_install.py tests/test_cli.py tests/test_build_planner.py)

# C1/C2 probes, extracted fresh from the cycle-2 archive
mkdir -p $REPO/.temp/TASK-260720-2x6mjn/review4 && cd $_
tar xzf <board>/.resources/TASK-260720-2x6mjn/TASK-260720-2x6mjn_review2-probes.tar.gz
(cd probe2 && CSK_PROBE_ROOT=../../worktree $VENV -m pytest -q -s -p no:cacheprovider -o addopts= .)
```

## Handoff to the commit-owning mover

The work is accepted on content. This review is read-only and supplied no `commit_ack`; the
worktree is still unstaged and uncommitted on branch
`task/TASK-260720-2x6mjn-pure-build-planner` (base `82d1cfc`). The commit-owning mover should
commit the 9 modified + 3 untracked files (`LOGBOOK.md`, `src/csk/builds/planner.py`,
`tests/test_build_planner.py` are new) and make the final `done` transition with
`commit_ack=scope_committed`.

Reviewer scratch to delete when the task closes: `.temp/TASK-260720-2x6mjn/probe/`, `probe2/`,
`review3/`, `review4/`, `nogo-bin/`, `review2-*.log`, and the baseline worktree
(`git worktree remove .temp/TASK-260720-2x6mjn/review-baseline`).
