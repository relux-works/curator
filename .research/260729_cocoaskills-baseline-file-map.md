# CocoaSkills baseline refresh and root-task file map

**Board task:** `TASK-260729-35tb37` (`refresh-csk-baseline-and-file-map`)  
**Research date:** 2026-07-29  
**Target repository:** `/Users/iv/Developer/Wildberries/cocoaskills`  
**Compared baseline:** accepted `TASK-260729-1t1z2l_curator-go-to-csk-parity-delta.md`, revision 2  
**Revision:** 2 — rework cycle 1, answering `TASK-260729-35tb37_review-verdict-cycle-1.md`  
**Scope:** read-only repository, protocol-candidate, and board analysis. No pull, fetch into local refs, checkout, worktree creation, dependency install, product edit, stage, commit, or broad test was performed.

### Revision 2 change log

| Verdict item | Correction applied |
| --- | --- |
| 1 | new §2.3 records the accepted rc.2 `98 passed` versus immutable rc.5 `1 failed, 97 passed` historical local-baseline regression, its exact `scripts/golden-tool` cause, the semantic manifest equivalence, the upstream `deb971f` fix and `6fc2fd9` rc.3 pin, and the explicit statement that this is a regression gate rather than a product change or pin authorization |
| 2 | §1, §6.2, §8, §9 now carry the current `TASK-260729-v5hqnv` state (`to-review` after two rework cycles) and its verified effect |
| 3 | §2.2 corrected from “20 files” to 19 distinct paths / 20 commit-level touch events |
| 4 | §3.5 makes the local rc.2 versus upstream rc.3 conformance-pin boundary explicit; §4.3 adds the focused `test_skill_manifest_resolution_vectors` replay to the schema-root gate with its justification |

All provenance, board, and protocol facts below were re-queried during this cycle. No CocoaSkills, Curator, spec, product, pin, dependency, or checkout state was changed, and no test suite was executed.

## 1. Executive findings

The accepted CocoaSkills repository baseline is still exact:

- local `main` is clean at `edce8816dda44bb121d661b7c4dea942558ce408`;
- `origin/main` resolves remotely and locally to `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`;
- local `main` is `+0/-2` relative to `origin/main`;
- the two missing commits are `deb971f` (`Adopt agent-skill manifest name`) and `6fc2fd9` (`Pin landed rc.3 protocol`);
- the repository has no target-local `AGENTS.md`.

The implementation base must therefore be a clean fast-forward of local `main` to the then-current verified `origin/main`, followed by one task-scoped worktree per root task. This reconnaissance did not perform that mutation.

The two root implementation tasks still have the same source ownership:

| Root task | Exact product files | Exact focused tests | Live integration deferred to |
| --- | --- | --- | --- |
| `TASK-260720-z9j4c9` schema-v6 build model | modify `src/csk/skillspec.py`; add `src/csk/builds/__init__.py`; modify validation-only logic in `src/csk/skillcheck.py` | modify `tests/test_skillspec.py`, `tests/test_skillcheck.py` | context exclusion, hashing, toolchain, driver, cache, and installer tasks |
| `TASK-260720-z2z795` install transaction engine | modify `src/csk/locking.py`; add `src/csk/transactions.py` | expand `tests/test_locking.py`; add `tests/test_transactions.py` | planner/dry-run routing, project/hybrid/global install, status/repair/GC tasks |

Neither root is startable yet. Both remain `backlog` and both are blocked only by `TASK-260720-1pvfj5` (`enforce-cross-platform-ci-gates`), which remains `backlog` while `TASK-260720-jrrgw9` is in `development`.

There are four material changes from the accepted parity map:

1. Four formerly in-flight prerequisites are now accepted: `TASK-260720-1nlmvv`, `TASK-260720-2qqq0w`, `TASK-260729-2kaopg`, and `TASK-260729-3jku56` are all `done`.
2. The rc.5 build-driver golden gap has an accepted candidate fix in `TASK-260729-3nx97g` (`done`). It adds the missing `build-drivers.json` and `expected/build-driver/` surface and makes Curator's candidate metadata test pass without skips. It is still uncommitted and unlanded on base `57c1f568`; it authorizes no protocol substitution, tag, publication, CI pin movement, or downstream release claim. The old accepted rc.5 snapshot `TASK-260728-2kp3tv` still lacks both paths.
3. `TASK-260729-v5hqnv` has applied the seven documented CocoaSkills rc.5 brief retargets and added satisfied golden-candidate dependency edges to `TASK-260720-2dnqw2` and `TASK-260720-12r55p`. That task is currently `to-review` after two producer cycles, not accepted `done`. The two upstream gates `TASK-260720-3ag6pi` and `TASK-260720-jrrgw9` remain rc.4-scoped and unresolved. Details in §6.2.
4. Accepted `TASK-260729-1b9tc3` (`done`) supplies a measured local-baseline conformance regression that the parity map did not carry: the stale local checkout is `98 passed` against its own pinned rc.2 conformance root but `1 failed, 97 passed` against the immutable rc.5 root, solely because of the `csk-skill.json` → `agent-skill.json` manifest rename. Upstream `deb971f` already fixes this and `6fc2fd9` already advances the CI pin to rc.3, so the failure is a regression gate for producers, not work to redo. Details and scope limits in §2.3.

One pre-existing root-task diagram is now semantically stale. `TASK-260720-z2z795_csk-build-lifecycle.puml` places a manager-home recovery pass before read-only planning. Accepted rc.5 `profiles/manager.md` §§2.5–2.6 requires recovery only after all private builds succeed, under the home lock in the serialized publication phase. Producers must follow the accepted text, not that diagram step.

## 2. Current repository provenance and cleanliness

### 2.1 Local and upstream identity

| Evidence | Result |
| --- | --- |
| Repository root | `/Users/iv/Developer/Wildberries/cocoaskills` |
| Branch | `main` |
| Local HEAD | `edce8816dda44bb121d661b7c4dea942558ce408` |
| Local commit | `Prepare changelog for v0.12.4`, 2026-07-13 |
| Origin URL | `git@github.com:ivanopcode/cocoaskills.git` |
| Read-only remote `refs/heads/main` | `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12` |
| Local `refs/remotes/origin/main` | same `6fc2fd97…` |
| Tracking divergence | `0 2` from `HEAD...origin/main` |
| Upstream tip | tag `v0.12.5`, `Pin landed rc.3 protocol`, 2026-07-14 |
| Worktree modifications | none |
| Staged modifications | none |
| Untracked non-ignored files | none |

`git status --porcelain=v2 --branch` reports only:

```text
# branch.oid edce8816dda44bb121d661b7c4dea942558ce408
# branch.head main
# branch.upstream origin/main
# branch.ab +0 -2
```

The working tree was checked independently with `git diff --quiet`, `git diff --cached --quiet`, and `git ls-files --others --exclude-standard`; all returned exit 0 and no paths.

### 2.2 Upstream-only delta that producers must inherit

The two upstream commits touch **19 distinct paths**. `git diff --name-only HEAD..origin/main | sort -u | wc -l` is `19`; `git log --name-only --format='' HEAD..origin/main` yields **20 commit-level touch events**, because `.github/workflows/ci.yml` is modified by both commits:

| Commit | Subject | Paths touched |
| --- | --- | ---: |
| `deb971f` | `Adopt agent-skill manifest name` | 19 (including `.github/workflows/ci.yml`) |
| `6fc2fd9` | `Pin landed rc.3 protocol` | 1 (`.github/workflows/ci.yml` only) |

The root-relevant delta is:

- `src/csk/skillspec.py`: `agent-skill.json` becomes canonical, `csk-skill.json` remains a legacy peer, equal dual manifests select canonical, conflicting/invalid peers fail closed, and source identity propagates into parsed command/dependency objects.
- `src/csk/skillcheck.py`: diagnostics select the actual canonical/legacy manifest path.
- `tests/test_skillspec.py`: adds canonical and dual-manifest resolution coverage.
- `tests/test_protocol_conformance.py`: consumes `vectors/skill-manifest-resolution.json`.
- `.github/workflows/ci.yml`: pins the released rc.3 conformance commit `00b1688a9b2457ca397a0bb550acf47cad8ee967`.

Therefore `TASK-260720-z9j4c9` must edit the upstream parser shape, not the stale local file. In particular, schema-v6 behavior must work identically through `agent-skill.json` and `csk-skill.json`, including equal-dual-manifest comparison and source-labelled errors.

### 2.3 Historical local-baseline conformance regression — why the stale checkout must not be the implementation base

This subsection exists to record measured evidence, not to authorize work. It is a **regression gate** for producers. It is **not** a product defect to fix in either root task, **not** authorization to move any committed CI protocol pin, and **not** a claim that upstream `6fc2fd97` fails.

#### 2.3.1 Measured result

Accepted, reviewer-accepted `TASK-260729-1b9tc3` (`design-csk-rc5-conformance-consumer`, status `done`) ran the same local CocoaSkills code at `edce8816…` against two conformance roots:

| Conformance root | Protocol | `manifest.json` SHA-256 | `pytest tests/test_protocol_conformance.py -q` | Exit |
| --- | --- | --- | --- | ---: |
| pinned released rc.2, `relux-works/curator-spec@cbe912d064e06275b0a1aa6762b7c31f687051c5` | `1.0.0-rc.2` | `728f772950414b9c3ddf38a8f1e9f2c7d2953bdca1d8c135c7e1a9abf40fff06` | **98 passed** | 0 |
| immutable rc.5 candidate root, `TASK-260729-3nx97g` worktree | `1.0.0-rc.5` | `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c` | **1 failed, 97 passed** | 1 |

Evidence files: `TASK-260729-1b9tc3_baseline-pin-rc2.log` (`exit=0`) and `TASK-260729-1b9tc3_candidate-rc5.log` (`exit=1`). Neither run was executed by this reconnaissance; both are quoted from the accepted artifact and independently corroborated below.

The single failure is `test_shared_fixture_context_hash_and_marker`:

```text
AssertionError: assert ['.skill_trig.../golden-tool'] == ['.skill_trig...ces/notes.md']
  Left contains one more item: 'scripts/golden-tool'
tests/test_protocol_conformance.py:50: AssertionError
```

#### 2.3.2 Exact cause — `scripts/golden-tool`

Local `edce8816…` `src/csk/skillspec.py` resolves exactly one modern manifest name:

```python
def load_skill_spec(snapshot: Path) -> SkillSpec:
    csk_path = snapshot / "csk-skill.json"
    if csk_path.exists():
        return _load_csk_skill(csk_path)
    runtime_path = snapshot / "agents" / "runtime.json"
    if runtime_path.exists():
        return _load_runtime_fallback(runtime_path)
    return SkillSpec(commands={}, source_file=None)
```

The rc.5 shared fixture `conformance/v1/fixtures/skill/` ships **only** `agent-skill.json`; there is no `csk-skill.json` and no `agents/runtime.json`. Local `load_skill_spec()` therefore falls through to the final branch and **silently returns an empty spec** — `commands={}`, `runtime_roots=()`, `source_file=None`. The conformance test then computes `include_scripts = not any(command.type == "script" …)` as `True` and passes `exclude_roots=()`, so `whitelist.copy_context()` copies `scripts/golden-tool` into the agent context. The expected list `[".skill_triggers/en.md", "SKILL.md", "references/notes.md"]` has no scripts entry, so the assertion fails by exactly one extra path.

#### 2.3.3 The two manifests are semantically identical

Independently verified in this cycle, not taken on trust:

| Check | Command | Exit | Result |
| --- | --- | ---: | --- |
| normalized JSON body equality | `diff <(json.dumps(sort_keys) rc.2 csk-skill.json) <(json.dumps(sort_keys) rc.5 agent-skill.json)` | 0 | identical: `schema_version: 5`, `runtime_roots: ["scripts"]`, one `script` command `golden-tool`, same capabilities, same empty dependency maps |
| `expected/marker.json` | `cmp rc.2 rc.5` | 0 | byte-identical |
| `expected/context_files.json` | `cmp rc.2 rc.5` | 0 | byte-identical |

Only the **filename** changed. No protocol semantics, schema shape, or expected golden changed between the two roots for this fixture. The regression is entirely a consumer-side manifest-resolution gap in the stale local checkout.

#### 2.3.4 Already fixed upstream — this is why the base matters

Upstream `deb971f` (`Adopt agent-skill manifest name`) replaces the single-name lookup with explicit canonical/legacy constants in `src/csk/skillspec.py`:

- `CANONICAL_MANIFEST = "agent-skill.json"`, `LEGACY_MANIFEST = "csk-skill.json"`;
- resolution order canonical → legacy → `RUNTIME_FALLBACK`;
- equal dual manifests select canonical; conflicting dual manifests fail closed with `conflicting_skill_manifests: …`;
- an invalid peer does not hide behind the other name;
- `source` / `source_file` carry the filename actually read.

The same commit adds `tests/test_protocol_conformance.py::test_skill_manifest_resolution_vectors`, parametrised over `vectors/skill-manifest-resolution.json`. Local `HEAD` has 16 `def test_*` functions in that module; upstream has 17.

`6fc2fd9` (`Pin landed rc.3 protocol`) then advances the CI conformance checkout from rc.2 `cbe912d0…` to rc.3 `00b1688a…` (`Adopt agent-skill.json as the canonical manifest (#11)`), whose root ships `fixtures/skill/agent-skill.json` and the 8-case `vectors/skill-manifest-resolution.json`.

#### 2.3.5 Scope limits — read this before acting on §2.3

- Both root tasks inherit the fix by basing on upstream `6fc2fd97…` or later. **Neither root re-implements it.**
- The failure is only reproducible on the stale local checkout. It must not be filed as a defect against `TASK-260720-z9j4c9`, `TASK-260720-z2z795`, or the rc.5 candidate.
- Nothing here authorizes moving `.github/workflows/ci.yml`'s conformance `ref`. That pin advances only through its own owning chain (`TASK-260720-25d05o` → `TASK-260720-1utsx8`).
- Nothing here authorizes landing, tagging, or publishing the rc.5 candidate.
- The operational use is a **regression gate**: after a schema-6 parser change, canonical/legacy manifest resolution must still hold. §4.3 encodes that gate.

## 3. Packaging, CLI, environment, PATH, and CI map

### 3.1 Packaging and supported runtime

Source of truth: `origin/main@6fc2fd97`, `pyproject.toml`.

| Boundary | Current contract |
| --- | --- |
| Build backend | `setuptools.build_meta` |
| Build requirements | `setuptools>=80`, `setuptools-scm[simple]>=8` |
| Package layout | `src/`, discovered by `[tool.setuptools.packages.find]` |
| Runtime Python | `>=3.11` |
| Declared platforms | macOS, POSIX/Linux, Windows |
| Runtime dependencies | none |
| Dev dependencies | pytest, build, twine, mypy, cryptography |
| Console script | `csk = csk.cli:main` |
| Module entry | `python -m csk` via `src/csk/__main__.py` |
| Type checking | strict mypy; Python 3.11 semantics; all `src/csk` |
| Test discovery | `tests/`; `src` added to pytest import path |

Consequences for both roots:

- `src/csk/builds/__init__.py` and `src/csk/transactions.py` are included automatically by package discovery; neither root needs a `pyproject.toml` package list change.
- Both roots can and should remain standard-library-only; no dependency or lockfile change is justified.
- There is no `Makefile`, tox, nox, requirements file, or separate package manifest in the upstream tree.
- Adding a new importable module is covered by strict `python -m mypy` and the existing wheel/sdist build without workflow edits.

### 3.2 CLI and install/global/project flows

`src/csk/cli.py` owns `main()`, `build_parser()`, `_dispatch()`, and `_dispatch_global()`.

Current routing:

1. `csk install|update|upgrade` loads config and enters `GlobalLock(cfg.path.parent)`.
2. `_prepare_install_target()` resolves an alias/current checkout/path target.
3. `_cmd_install()` calls `installer.install()`.
4. `installer.install()` selects one/all configured projects, calls `_install_project()` for each, then calls `gc.collect_runtime()` after non-dry-run work.
5. `_install_project()` resolves closure, validates specs, audits, checks registries/MCP/tags, and only then returns for dry-run.
6. A real project install incrementally writes the consumer registry, runtime commands, context/marker directories, stale removals, shims, environment files, and adapters.
7. `csk global install|update|upgrade` follows the parallel `global_install.py` path under the same coarse `GlobalLock`.
8. `csk gc` also enters the same `GlobalLock` before `gc.collect_runtime()`.

Current dry-run still enters the coarse home lock. This conflicts with accepted rc.5's lock-free dry-run contract, but `TASK-260720-z2z795` explicitly does not own dry-run routing. `TASK-260720-2x6mjn` owns that integration.

### 3.3 Current persistent-write and durability patterns

Reusable patterns:

- `config._write_json_atomic()` uses a same-directory temporary file, restrictive permissions when available, `flush()`, `os.fsync()`, and `os.replace()`.
- `audit_registry._write_protected_state_file()` additionally syncs the containing directory on non-Windows platforms.
- `protocol_json.loads()` is the existing strict JSON reader and should back journal parsing rather than plain permissive `json.loads()`.
- tests use `tmp_path`, subprocess helpers, and an autouse isolated `HOME`/`USERPROFILE` fixture in `tests/conftest.py`.

Patterns that are evidence of the gap and must not be reused as transaction semantics:

- `consumers._write()` and `global_install._write_json()` write live JSON directly.
- `installer._replace_dir()` performs one local backup/rename/delete sequence without a durable multi-target journal.
- `snapshot.get_snapshot()` uses a fixed `.snapshot-<commit>.tmp` directory and rename, with no transaction journal.
- `gc.sweep_orphans()` infers old temp/backup ownership from PID-bearing names.

### 3.4 Environment variables and PATH boundaries

| Surface | Current integration point |
| --- | --- |
| `CSK_CONFIG` | overrides user config path in `config.config_path()` |
| `CSK_SYSTEM_CONFIG` | overrides system config path |
| `CSK_LOCK_TIMEOUT` | controls current `GlobalLock` timeout |
| `CSK_GLOBAL_USER_BIN` | explicit safe global forwarding-bin choice |
| `CSK_REGISTRY_TOKEN` | audit publication credential |
| `CSK_PROJECT_ROOT` | written into project env files |
| `CSK_GLOBAL_ROOT` | written into global env files |
| `CSK_AUTO_ENV`, `CSK_ACTIVE_ENV`, `CSK_OLD_PATH`, `CSK_ACTIVE_GLOBAL_ENV` | optional shell-hook activation state |
| `PATH` | command dependency probing, project-bin warning, global user-bin selection, wrappers/shims |
| `HOME`, `USERPROFILE` | user-scoped config/adapters/bin paths; isolated by tests |

PATH order and ownership:

- project commands live under `<project>/.agents/bin`;
- global commands live under `<csk-home>/global/bin`;
- `global_bins.select_user_bin_with_warning()` may publish forwarding shims only into a safe writable PATH-visible user bin;
- `CSK_GLOBAL_USER_BIN` may select a bin but protected tool-manager/CocoaSkills directories are rejected;
- runtime wrappers prepend command directory, implementation runtime directories, and resolved system dependency directories while preserving inherited PATH;
- optional `shell-init` changes human shell PATH, but agent execution is designed to use explicit project/global shim locations.

Root boundaries:

- schema-v6 parsing and validation must not inspect PATH, resolve `go`, or read toolchain environment;
- the transaction engine must not own command resolution or PATH activation;
- lock timeout compatibility may continue to read `CSK_LOCK_TIMEOUT`;
- transaction tests must isolate `HOME`, `USERPROFILE`, PATH, and lock roots through existing fixtures.

### 3.5 Tests and CI

Current CI:

- test matrix: Ubuntu, macOS, Windows × Python 3.11, 3.12, 3.13, 3.14;
- install: `python -m pip install -e ".[dev]"`;
- tests: `python -m pytest -v`;
- conformance env: `CURATOR_CONFORMANCE_ROOT=<checked-out protocol-spec>/conformance/v1`;
- typecheck: `python -m mypy` on Ubuntu/Python 3.12;
- build: `python -m build`, then `twine check dist/*`;
- release repeats build/twine and publishes tag-driven artifacts;
- distribution smoke verifies pipx, uv-tool, mise, install.sh, optional Homebrew, CLI version, minimal install/status, and cross-channel content hash stability.

#### Conformance-pin boundary — local rc.2 versus upstream rc.3

The committed conformance pin differs between the stale local checkout and the implementation base. Producers must read the upstream value, not the local one.

| Base | `.github/workflows/ci.yml` conformance `ref` | Protocol | `manifest.json` SHA-256 | Manifest files |
| --- | --- | --- | --- | ---: |
| local `edce8816…` | `cbe912d064e06275b0a1aa6762b7c31f687051c5` | `1.0.0-rc.2` | `728f7729…` | 81 |
| upstream `6fc2fd97…` (implementation base) | `00b1688a9b2457ca397a0bb550acf47cad8ee967` | `1.0.0-rc.3` | `7951cda1711d34d2a9dd9a873cf9d537c41ca4e9527e94f138f38743610a379e` | 93 |

The only difference between the two workflow files is that one `ref:` line; `6fc2fd9` changes nothing else in CI.

Content difference that matters to the schema root:

| Conformance path | rc.2 pin | rc.3 pin (upstream CI) | rc.5 candidate |
| --- | --- | --- | --- |
| `fixtures/skill/csk-skill.json` | present | absent | absent |
| `fixtures/skill/agent-skill.json` | absent | present | present |
| `vectors/skill-manifest-resolution.json` | absent (`test -e` exit 1) | present, 8 cases | present |
| `schema-cases/{agent,csk}-skill-v6/` | absent | absent | present, 24 files each |
| `vectors/build-drivers.json`, `expected/build-driver/` | absent | absent | present (candidate only) |

The 8 rc.3 manifest-resolution cases are `canonical-only`, `legacy-only`, `equal-dual-manifests`, `conflicting-dual-manifests`, `invalid-canonical-does-not-fallback`, `invalid-legacy-does-not-hide-behind-canonical`, `runtime-fallback-without-modern-manifest`, and `pure-context-without-manifest`.

Consequences:

- The upstream CI conformance checkout is rc.3 and it **does not** supply schema-v6 cases. A root producer must not make ordinary CI collection fail merely because v6 candidate files are absent, and must not move the committed protocol pin. Candidate-root qualification is explicit evidence, not a root-task pin change.
- `tests/test_protocol_conformance.py` carries `pytestmark = pytest.mark.skipif(not ROOT_TEXT, reason="CURATOR_CONFORMANCE_ROOT is not set")`. Any local replay of that module is **vacuous unless `CURATOR_CONFORMANCE_ROOT` is exported**; a bare `pytest tests/test_protocol_conformance.py` on a developer machine skips everything and proves nothing.
- Because the manifest-resolution vector exists only at rc.3 and later, a schema-root replay of `test_skill_manifest_resolution_vectors` requires an rc.3-or-later root. The upstream-pinned rc.3 checkout is the correct and sufficient root for that gate. See §4.3.

## 4. Root plan — `TASK-260720-z9j4c9` schema-v6 build model

### 4.1 Current upstream seam

`src/csk/skillspec.py` currently has:

- `SUPPORTED_SCHEMA_VERSIONS = {1, 2, 3, 4, 5}`;
- one frozen `CommandSpec` with script/system optional fields;
- one frozen `SkillSpec` with `runtime_roots`, capabilities, dependencies, skill requirements, and MCP requirements;
- canonical/legacy/fallback selection in `load_skill_spec()` — this is the post-`deb971f` shape and must be preserved exactly; see §2.3 for the regression it fixed and §4.3 for the gate that protects it;
- manifest parsing in `_load_skill_manifest()`;
- path validation in `_validate_relative_path()`, `_parse_runtime_roots()`, `_validate_v2_script_path()`, and `_path_contains()`;
- a regression test that currently expects schema 6 to fail.

`src/csk/skillcheck.py` currently:

- reports parse failures as `skill.spec_invalid`;
- warns about prompt-visible runtime roots;
- warns when script/provider commands lack a shell-neutral resolver;
- scans prompt-visible Markdown without a build-root exclusion.

The accepted schema files declare:

- top-level `build_roots` as a closed path set;
- `commands.<name>` may be script, system, or build in schema 6;
- a build command has exactly `type`, `driver`, and `source_dir`;
- `type` is `build`;
- `driver` is exactly `go-v1`;
- `source_dir` is a portable non-dot relative path;
- schemas 1–5 and the legacy runtime fallback do not gain build surfaces.

Both `agent-skill-v6` and `csk-skill-v6` candidate directories contain 24 cases each: one positive and closed-shape negatives including unsupported/missing driver, missing/dot source, args/env/flags/output/toolchain/hooks/scripts/tags, mixed script/system shapes, dot build root, generic invalid fields, and reserved schema-7 repository/driver/target fields.

### 4.2 Exact producer file/function plan

#### Add `src/csk/builds/__init__.py`

- Establish an import-safe, platform-neutral build domain package.
- Define only schema-level constants/types needed by schema 6, such as the closed `go-v1` driver identity.
- Do not import subprocess, toolchain, cache, platform control, or installer modules.
- Keep the initializer stable so later tasks can add `source.py`, `toolchain.py`, metadata, driver, and cache modules without moving the schema identity again.

#### Modify `src/csk/skillspec.py`

- Add schema 6 to `SUPPORTED_SCHEMA_VERSIONS`.
- Extend `CommandSpec` with optional `driver` and `source_dir` fields while preserving all existing script/system attributes and construction semantics.
- Extend `SkillSpec` with `build_roots: tuple[str, ...] = ()`.
- In `_load_skill_manifest()`:
  - allow `build_roots` only for schema 6;
  - explicitly reject `build_roots` and build commands in schemas 1–5, including schema 1's otherwise permissive top level;
  - preserve schema-5 capabilities/dependencies parsing unchanged;
  - parse build roots before build commands;
  - dispatch a closed build-command parser only when `schema == 6`;
  - preserve canonical/legacy equal-manifest behavior and source labels.
- Add narrowly named helpers:
  - `_parse_build_roots(...)`;
  - `_parse_build_command(...)` or an equivalent closed branch in the command parser;
  - `_validate_link_free_directory(...)`;
  - `_overlapping_roots(...)`;
  - `_validate_build_layout(...)`;
  - `_validate_nearest_go_module(...)`.
- Static validation must:
  - reject missing, file, dot, escaped, linked, duplicate, and overlapping build roots;
  - reject any overlap in either direction with `runtime_roots`;
  - require every build root to be used;
  - require every source directory to be a real link-free directory below exactly one build root;
  - allow `source_dir == build_root`;
  - require the build root's direct `go.mod` to be the nearest module file;
  - reject a missing, linked, special, or intervening `go.mod`;
  - order multi-command diagnostics deterministically.
- Keep `_load_runtime_fallback()` string-only; a legacy fallback build object must fail.
- Do not invoke `go`, hash content, exclude context during install, or mutate any install state.

#### Modify `src/csk/skillcheck.py`

- Add build-root prompt-reference warnings with a stable `skill.build_root_in_prompt_context` code.
- Treat build commands as managed commands for shell-neutral resolver guidance.
- Ensure Markdown inside a declared build root is not itself scanned as prompt-visible context.
- Preserve existing runtime-root, provider-runtime, localization, and script-command warnings.
- Do not modify `whitelist.py` here; install-time build-root exclusion belongs to `TASK-260720-3c0ss2`.

#### Modify `tests/test_skillspec.py`

Add focused positive and negative groups covering:

- both manifest names and equal dual manifests with schema 6;
- mixed build/script/system positive parsing;
- preservation of schema-5 capabilities and dependencies;
- all 15 closed build-command field/shape rejections;
- build surfaces rejected for every schema 1–5 and runtime fallback;
- missing/file/dot/escape/link build roots;
- duplicate, nested, and runtime-overlapping roots;
- missing roots declaration and unused roots;
- source outside root, missing, file, linked component;
- direct root `go.mod`, missing/linked `go.mod`, and intervening nested module;
- source equal to root;
- deterministic diagnostics;
- replace the old `schema_version=6 is unsupported` assertion with a future-version rejection.

#### Modify `tests/test_skillcheck.py`

Add:

- prompt-visible build-root warning;
- no warning from Markdown physically inside the excluded build root;
- build-command shell-neutral resolver warning;
- accepted resolver documentation case;
- preservation of existing runtime/localization behavior.

#### Do not modify in this root

`src/csk/whitelist.py`, `hashing.py`, `installer.py`, `closure.py`, `shims.py`, `status.py`, `global_install.py`, CLI files, packaging metadata, workflows, or protocol pins.

### 4.3 Narrow initial producer gates

Run these as standalone commands after the producer adds the tests:

```bash
python -m pytest -q tests/test_skillspec.py tests/test_skillcheck.py
python -m mypy src/csk/skillspec.py src/csk/skillcheck.py src/csk/builds
```

#### Required manifest-resolution regression gate

Because this root rewrites `_load_skill_manifest()` and the command parser — the exact code path that `deb971f` fixed — the producer must also replay the upstream manifest-resolution conformance test against the **upstream-pinned rc.3 root**, with `CURATOR_CONFORMANCE_ROOT` set explicitly:

```bash
CURATOR_CONFORMANCE_ROOT=<rc.3 checkout>/conformance/v1 \
  python -m pytest -q tests/test_protocol_conformance.py::test_skill_manifest_resolution_vectors
```

`<rc.3 checkout>` is `relux-works/curator-spec@00b1688a9b2457ca397a0bb550acf47cad8ee967`, exactly the ref the committed workflow already checks out. This adds no pin movement, no new dependency, and no candidate-root substitution.

Why direct `tests/test_skillspec.py` coverage is **not** an acceptable substitute:

- The eight rc.3 vectors are owner-published protocol expectations. `test_skillspec.py` cases are producer-authored in the same commit as the parser change, so a schema-6 refactor that regresses canonical/legacy resolution can regress its own assertions in lockstep and stay green.
- `§2.3` shows the concrete failure mode this guards: the regression surfaced only when a real conformance fixture that ships one manifest name met a parser that resolved the other, and it surfaced as a *silent empty spec*, not an exception. `_load_skill_manifest()` gaining a schema-6 branch is precisely where that silence can reappear.
- The gate is cheap: one already-existing test function, one already-pinned checkout, no new files.

If the producer proposes to drop this gate, the substitution must be justified in the task record with equivalent evidence that all eight named cases are covered — not merely that `test_skillspec.py` passes.

When an accepted candidate conformance root is explicitly supplied, add a task-owned schema-v6 case replay (24 `agent-skill-v6` plus 24 `csk-skill-v6` files, rc.5-only) as separate explicit evidence, again without changing the committed CI pin.

The root task's final required gates are:

```bash
python -m pytest -q tests/test_skillspec.py tests/test_skillcheck.py
CURATOR_CONFORMANCE_ROOT=<rc.3 checkout>/conformance/v1 \
  python -m pytest -q tests/test_protocol_conformance.py::test_skill_manifest_resolution_vectors
python -m mypy
git diff --check
```

These commands are a producer plan; this reconnaissance did not run them.

## 5. Root plan — `TASK-260720-z2z795` install transaction engine

### 5.1 Current upstream seam

`src/csk/locking.py` currently provides only `GlobalLock`:

- lock path `<csk-home>/.lock`;
- exclusive creation with `O_CREAT|O_EXCL`;
- JSON `{pid, created_at}`;
- PID-based stale break by rename;
- timeout from `CSK_LOCK_TIMEOUT`;
- four tests: held timeout, dead-PID break, live-PID refusal, corrupt-lock refusal.

There is no `src/csk/transactions.py`.

The live CLI uses `GlobalLock` around project/global install, update, upgrade, and GC. Persistent targets then mutate incrementally. There is no project operation lock, optional build-key lock, target-class plan, durable journal, shared recovery scan, desired-digest rollback guard, or consumer-last commit.

Accepted rc.5 ordering is:

1. project operation locks by canonical identity in unsigned UTF-8 byte order;
2. optionally one build-key lock at a time, released before home lock;
3. manager-home mutation lock;
4. no project/build lock acquisition while home lock is held;
5. recover all journals, revalidate, publish, prepare one journal;
6. commit target classes in fixed order with consumer ledger last;
7. rollback in exact reverse order only if current bytes still equal the journal's desired digest;
8. keep backups until success/rollback is durable;
9. release home lock before project locks.

Recovery inside install occurs only after private builds succeed. Standalone recovery/GC acquires only the home lock.

### 5.2 Exact producer file/function plan

#### Modify `src/csk/locking.py`

- Preserve `LockError` and `CSK_LOCK_TIMEOUT` compatibility.
- Replace crash-persistent sentinel ownership as the core primitive with a process-owned native file lock whose kernel ownership releases on abnormal process exit:
  - POSIX `fcntl` path;
  - Windows `msvcrt`/Win32-compatible path;
  - import-safe conditional platform code.
- Keep stable manager-local lock files; do not delete the shared lock inode on ordinary release.
- Add canonical identity helpers:
  - `canonical_project_identity(path: Path) -> str`;
  - `canonical_project_identities(paths: Iterable[Path]) -> tuple[str, ...]`;
  - sort with `identity.encode("utf-8")`, not locale or Python display order.
- Add lock-domain APIs:
  - project operation lock set;
  - optional per-key build lock;
  - manager-home mutation lock;
  - held-lock assertions for transaction APIs.
- Enforce order in process-local acquisition state:
  - reject project-order inversion before waiting;
  - reject home acquisition while a build lock is held;
  - reject project/build acquisition while home is held;
  - release partial multi-project acquisition on timeout/cancellation/error.
- Use full collision-resistant lock names derived from canonical identities. Do not reuse `project_resolver.stable_path_hash()`, whose four-character SHA-1 suffix is only a display alias.
- Retain a `GlobalLock` compatibility wrapper for current CLI imports until later integration tasks replace coarse routing. Do not edit `cli.py` in this root.

#### Add `src/csk/transactions.py`

Add one standard-library-only generic engine with explicit immutable models:

- transaction/journal version and phase;
- target class, identifier, live path, staged path/removal intent, expected preimage or generation digest, desired digest, backup path, and commit state;
- plan and journal objects;
- engine requiring a held home mutation lock.

Required functions/methods:

- deterministic `sort_targets()` using fixed target-class order and unsigned UTF-8 identifier order;
- strict target-plan validation, including duplicate/overlapping/aliased live targets;
- file/tree/entry digest functions that bind the intended target kind and reject links/special files where unsupported;
- durable same-filesystem journal write using restrictive temp file, file fsync, replace, and directory sync where supported;
- `prepare()` that records the complete ordered plan before live mutation;
- `commit()` that journals each state transition durably before the next swap;
- `rollback()` in exact reverse committed order;
- desired-digest recheck before restoration, preserving unknown concurrent bytes on mismatch;
- `recover_all()` scanning journals by transaction id, not initiating project;
- durable backup/journal cleanup only after consumer-last success or successful rollback;
- recovery/reference helpers needed later by GC, without invoking GC now.

Target classes must match accepted order:

1. project/global contexts and markers;
2. runtime, shim, and environment targets;
3. adapter ledgers and hybrid/global mirrors;
4. stale managed removals;
5. machine-wide consumer ledger last.

Use `protocol_json.loads()` for strict journal reads. Reuse the durability shape from `audit_registry._write_protected_state_file()` and `_sync_directory()` conceptually, but keep transaction-owned error types and paths. Do not use direct-write helpers from `consumers.py`, `global_install.py`, or `manifest.py` as the journal implementation.

#### Expand `tests/test_locking.py`

Keep the four existing compatibility tests and add:

- canonical absolute/symlink identity;
- deterministic unsigned UTF-8 multi-project order;
- same project contention and independent project concurrency;
- missing-home/first-use identity;
- timeout releases partial acquisition;
- order inversion fails before waiting;
- build lock must be released before home lock;
- project/build acquisition while home-held fails;
- abnormal subprocess exit releases native ownership;
- compatibility `GlobalLock` remains usable by current CLI.

Native Windows identity/case tests must run on Windows CI; Darwin normalization/case tests should be guarded by actual filesystem behavior, not hard-coded OS assumptions.

#### Add `tests/test_transactions.py`

Add focused fault-injection and subprocess coverage:

- exact canonical journal shape and strict unknown/duplicate/trailing JSON rejection;
- target-class and unsigned identifier ordering, consumer last;
- expected preimage/generation mismatch before commit;
- durable journal transition before each swap;
- crash at preparation and every target boundary;
- recovery scans all transaction IDs independently of project;
- reverse rollback order;
- desired-digest mismatch preserves unknown state;
- backups retained until durable success/rollback;
- concurrent project successes preserve both results and merge the consumer ledger;
- one project rollback cannot overwrite another project's successful target;
- APIs reject calls without the home lock;
- recovery/standalone maintenance uses only the home lock;
- link/special/overlapping target rejection;
- platform-native durability cases.

#### Do not modify in this root

`src/csk/cli.py`, `installer.py`, `global_install.py`, `gc.py`, `consumers.py`, `manifest.py`, `snapshot.py`, shims/adapters/env writers, build/cache modules, packaging metadata, workflows, or protocol pins.

Live routing belongs to:

- `TASK-260720-2x6mjn`: planner and lock-free dry-run;
- `TASK-260720-3t8nr3`: project/hybrid journaled commit;
- `TASK-260720-g7kgox`: global journaled commit;
- `TASK-260720-th0jdi`: recovery/status/repair/locked GC.

### 5.3 Narrow initial producer gates

Run as standalone commands:

```bash
python -m pytest -q tests/test_locking.py tests/test_transactions.py
python -m mypy src/csk/locking.py src/csk/transactions.py
```

The final task gates are:

```bash
python -m pytest -q tests/test_locking.py tests/test_transactions.py
python -m mypy
git diff --check
```

Native Windows and macOS focused replays are required before handoff for platform-owned lock/identity/durability assertions. Linux may exercise generic transaction infrastructure, but it is not evidence for the rc.5 `go-v1` driver platform claim. These commands are a producer plan; this reconnaissance did not run them.

## 6. Drift from the accepted parity map

### 6.1 Repository drift

No CocoaSkills provenance drift:

- accepted local HEAD remains local HEAD;
- accepted remote HEAD remains current `origin/main`;
- accepted `+0/-2` divergence remains exact;
- CocoaSkills remains clean and unstaged;
- current upstream file layout remains the one mapped above.

### 6.2 Board-state drift

| Task | Accepted-map state | Current state | Effect |
| --- | --- | --- | --- |
| `TASK-260720-1nlmvv` currentness/repair | `development` | `done` | prerequisite converged |
| `TASK-260720-2qqq0w` compiled-build docs | `backlog` | `done` | one `1pvfj5` blocker resolved |
| `TASK-260729-2kaopg` global status | `development` | `done` | formerly in-flight delta accepted |
| `TASK-260729-3jku56` idempotent compiled install | `reviewing` | `done` | formerly in-flight delta accepted |
| `TASK-260720-jrrgw9` rc.4 conformance | `backlog` | `development` | active remaining predecessor |
| `TASK-260720-1pvfj5` cross-platform CI | `backlog` | `backlog` | roots remain blocked |
| `TASK-260720-3ag6pi` protocol-v6 verification | `blocked` | `blocked` | immutable-ref/publication gap remains |
| `TASK-260729-3nx97g` rc.5 goldens | absent/open gap | `done` | technical golden set accepted, not landed |
| `TASK-260729-v5hqnv` seven brief retargets | required future action | `to-review` | board text changed and corrected across two cycles; review acceptance still pending |
| `TASK-260729-1b9tc3` rc.5 conformance consumer design | absent | `done` | supplies the §2.3 rc.2/rc.5 regression evidence and the consumer test design |
| both CocoaSkills roots | `backlog` | `backlog` | no start clearance |

The two satisfied golden edges are now explicit, and both were re-read live in this cycle:

- `TASK-260720-2dnqw2.blockedBy` = `TASK-260720-3c0ss2`, `TASK-260720-3j8pp5`, `TASK-260729-3nx97g`;
- `TASK-260720-12r55p.blockedBy` = `TASK-260720-th0jdi`, `TASK-260720-3ag6pi`, `TASK-260729-3nx97g` — still fail-closed on blocked `TASK-260720-3ag6pi`.

Root edges are unchanged: `TASK-260720-z9j4c9.blockedBy` and `TASK-260720-z2z795.blockedBy` are each exactly `[TASK-260720-1pvfj5]`, and `TASK-260720-1pvfj5.blockedBy` is `[TASK-260720-2qqq0w, TASK-260720-jrrgw9]` with `2qqq0w` now `done` and `jrrgw9` in `development`.

#### `TASK-260729-v5hqnv` — current state and effect

The accepted parity map listed the retarget as a required future action. It has since run two producer cycles and is now `to-review`, awaiting a fresh reviewer:

| Cycle | Reviewer finding | Producer response |
| --- | --- | --- |
| 1 | audit §5 falsely claimed the stale rc.4 receipt hash `750f5f75…` was “removed from the board entirely”; `task-board grep` still finds it in historical outcome resources and progress records | claim narrowed to the verified statement — absent from the current `description`/`scope`/`ac` of all seven briefs, 0 hits across all 21 fields, 29 historical lines deliberately preserved |
| 1 | an rc.5 paragraph was appended to `TASK-260720-12r55p.notes`, outside the task's authorized `description`/`scope`/`ac`/dependency-edge scope, so the attached audit was not a complete mutation inventory | `notes` restored byte-for-byte; the fail-closed explanation rewritten into the authorized `scope` field (1383 → 1979 chars); `after.jsonl` regenerated with `notes` included; audit revised |

Restoration independently re-verified in this cycle: live `TASK-260720-12r55p.notes` is **515 bytes**, SHA-256 `3a04adc70cd2e5499608172203a7d3ed17b4ec33e281c4c7ca7b51f24ace2500`, matching both authoritative pre-retarget sources cited in the rework record.

Effect on this reconnaissance and on the two roots:

- The seven CocoaSkills brief texts on the board **are** rc.5-aligned today, and the two golden-candidate dependency edges **are** live. Producers reading those briefs will see current rc.5 wording.
- That wording is **not yet reviewer-accepted**. A second review cycle can still change brief `description`/`scope`/`ac`, so a producer must re-read the brief at start time rather than caching text from this artifact.
- `TASK-260729-v5hqnv` touches **no** CocoaSkills or Curator source, test, pin, or dependency. It cannot unblock, start, or alter either root task.
- `TASK-260720-12r55p` remains fail-closed on blocked `TASK-260720-3ag6pi` by deliberate design; that gate is still literal rc.4 and no rc.5 replacement gate exists on the board.
- The retarget explicitly ran **no** tests and asserts no test result. The rc.2/rc.5 numbers in §2.3 come from `TASK-260729-1b9tc3`, not from `TASK-260729-v5hqnv`.

### 6.3 Protocol-candidate drift

Accepted `TASK-260729-3nx97g` closes the technical golden-generation defect reported by the parity map:

- accepted candidate adds `conformance/v1/vectors/build-drivers.json`;
- accepted candidate adds `conformance/v1/expected/build-driver/`;
- portable cache key remains `sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b`;
- portable receipt remains `sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd`;
- Curator candidate build-metadata coverage passes with no skip.

The publication boundary has not moved:

- candidate base is still `57c1f56846d221ecc55786bd3c2467ec32f11730`;
- the candidate has unstaged/uncommitted working-tree content and no staged files;
- no tag, immutable upstream ref, publication, CocoaSkills CI pin, or downstream release claim is authorized;
- the old accepted `TASK-260728-2kp3tv` snapshot still returns exit 1 for existence checks of both golden paths and exit 1 for `rg build-driver` in its manifest.

Thus the parity-map wording should change from “goldens must be regenerated” to “the accepted regenerated candidate must be landed in the owner-selected immutable rc.5 root before downstream shared-vector qualification.” This does not unblock either root by itself.

### 6.4 Stale root diagrams

- `TASK-260720-z9j4c9_csk-build-components.puml` still labels the protocol input rc.4. Its component/file ownership remains useful, but protocol-version wording is stale.
- `TASK-260720-z2z795_csk-build-lifecycle.puml` performs a brief home-lock recovery before planning. That step contradicts accepted rc.5 `profiles/manager.md` §§2.5–2.6, which permits in-install recovery only after private builds succeed in the serialized home-lock phase.

No diagram was edited by this read-only task. Producers and reviewers should cite the accepted textual contract until those resources are explicitly revised.

## 7. Fact-check record

Every command below ran as a standalone process, except line-count reads that explicitly used `zsh -o pipefail`; no gate was piped through `tee`. Real exit codes are recorded.

| Check | Exit | Finding |
| --- | ---: | --- |
| required initial `task-board m set_status(...)` | 0 | task moved `backlog` → `analysis` |
| `git --version` | 0 | git 2.50.1 |
| `rg --version` | 0 | ripgrep 15.2.0 |
| `task-board --help` | 0 | required q/m/resource/handoff surfaces available |
| `rg --files -g AGENTS.md` in CocoaSkills | 1 | expected absence; no target-local instructions |
| `git status --porcelain=v2 --branch` in CocoaSkills | 0 | clean `main`, `+0/-2` |
| `git rev-parse HEAD` | 0 | `edce8816…` |
| `git ls-remote origin refs/heads/main` | 0 | `6fc2fd97…` |
| `git rev-parse refs/remotes/origin/main` | 0 | same `6fc2fd97…` |
| `git rev-list --left-right --count HEAD...origin/main` | 0 | `0 2` |
| `git diff --quiet` / `git diff --cached --quiet` | 0 each | no tracked or staged delta |
| `git ls-files --others --exclude-standard` | 0 | no untracked files |
| scoped board queries for roots/gates/retarget task | 0 | statuses and blockers in §6.2 |
| schema-v6 case count, agent/csk | 0 each | 24 files in each directory |
| old rc.5 `test -e vectors/build-drivers.json` | 1 | expected absence, still missing |
| old rc.5 `test -e expected/build-driver` | 1 | expected absence, still missing |
| old rc.5 `rg build-driver manifest.json` | 1 | expected absence, zero manifest hits |
| accepted golden candidate `git rev-parse HEAD` | 0 | base `57c1f568…` |
| accepted golden candidate `git diff --cached --quiet` | 0 | nothing staged |

### 7.1 Additional checks run in revision 2

| Check | Exit | Finding |
| --- | ---: | --- |
| `task-board m set_status(TASK-260729-35tb37, status=analysis)` | 0 | rework cycle re-entered `analysis` |
| `git status --porcelain=v2 --branch` in CocoaSkills | 0 | still clean `main`, `+0 -2`, unchanged |
| `git rev-parse HEAD` / `git ls-remote origin refs/heads/main` / `git rev-parse refs/remotes/origin/main` | 0 each | `edce8816…` and `6fc2fd97…` re-confirmed |
| `git diff --name-only HEAD..origin/main \| sort -u \| wc -l` | 0 | **19** distinct paths |
| `git log --name-only --format='' HEAD..origin/main \| sed '/^$/d' \| wc -l` | 0 | **20** commit-level touch events |
| `git show --name-only deb971f` / `6fc2fd9` | 0 each | 19 and 1 paths; `.github/workflows/ci.yml` in both |
| `git show HEAD:.github/workflows/ci.yml` conformance ref | 0 | local pin `cbe912d064e06275b0a1aa6762b7c31f687051c5` (rc.2) |
| `git show origin/main:.github/workflows/ci.yml` conformance ref | 0 | upstream pin `00b1688a9b2457ca397a0bb550acf47cad8ee967` (rc.3) |
| `git diff HEAD..origin/main -- .github/workflows/ci.yml` | 0 | one changed line — the `ref:` value only |
| `git show HEAD:src/csk/skillspec.py` manifest names | 0 | only literal `csk-skill.json`; no `agent-skill.json` |
| `git show origin/main:src/csk/skillspec.py` manifest names | 0 | `CANONICAL_MANIFEST = "agent-skill.json"`, `LEGACY_MANIFEST = "csk-skill.json"` |
| `git show deb971f -- tests/test_protocol_conformance.py` | 0 | added lines introduce `test_skill_manifest_resolution_vectors` and `vectors/skill-manifest-resolution.json` |
| `def test_*` count, local vs upstream conformance module | 0 each | 16 vs 17 |
| `ls` rc.5 candidate `conformance/v1/fixtures/skill/` | 0 | `agent-skill.json` only; no `csk-skill.json` |
| `ls` rc.2 pin `conformance/v1/fixtures/skill/` | 0 | `csk-skill.json` only; no `agent-skill.json` |
| normalized JSON `diff` of rc.2 `csk-skill.json` vs rc.5 `agent-skill.json` | 0 | semantically identical bodies |
| `cmp` rc.2 vs rc.5 `expected/marker.json` | 0 | byte-identical |
| `cmp` rc.2 vs rc.5 `expected/context_files.json` | 0 | byte-identical |
| `shasum -a 256` rc.2 pin `manifest.json` | 0 | `728f7729…` |
| `shasum -a 256` rc.5 candidate `manifest.json` | 0 | `b6f56aac…` |
| `git -C <spec worktree> show 00b1688a…:conformance/v1/manifest.json` | 0 | `protocol_version 1.0.0-rc.3`, 93 files, digest `7951cda1…` |
| rc.3 `vectors/skill-manifest-resolution.json` case list | 0 | 8 named cases |
| `test -e` rc.2 `vectors/skill-manifest-resolution.json` | 1 | expected absence |
| `git ls-tree` rc.3 `schema-cases/{agent,csk}-skill-v6` | 0 | empty result — expected absence at rc.3 |
| `git ls-tree` rc.3 `vectors/build-drivers.json` | 0 | empty result — expected absence at rc.3 |
| `git ls-tree` rc.3 `fixtures/skill` | 0 | ships `agent-skill.json`, no `csk-skill.json` |
| `find` rc.5 candidate `schema-cases/{agent,csk}-skill-v6` | 0 each | 24 files each, unchanged |
| `test -e` old rc.5 snapshot `vectors/build-drivers.json` / `expected/build-driver` | 1 each | expected absence, still missing |
| rc.5 candidate `git rev-parse HEAD` / `git diff --cached --quiet` | 0 each | still `57c1f568…`, nothing staged |
| board projections for the 12 tracked tasks and 5 dependency edges | 0 each | statuses and edges in §6.2 |
| live `TASK-260720-12r55p.notes` byte length and SHA-256 | 0 | 515 bytes, `3a04adc7…` — restoration confirmed |

Two commands intentionally used a pipe for counting (`wc -l`); their upstream `git` stages produce no meaningful failure mode here and the counted output is reproduced above. Every provenance, existence, and comparison gate ran as a standalone process with its real exit code recorded. Expected-red checks (`test -e` on absent paths, `rg` with no match) are reported as failing with their real exit code 1 and an explicit expected-absence rationale.

No pytest, mypy, build, twine, pull, checkout, install, pin change, or broad repository test command was run by this reconnaissance in either revision. The `98 passed` and `1 failed, 97 passed` figures in §2.3 are quoted from accepted `TASK-260729-1b9tc3` logs; this task did not re-execute them.

## 8. Recommendation

Keep both root tasks blocked until `TASK-260720-1pvfj5` is reviewer-accepted. At handoff:

1. record the current CocoaSkills remote head;
2. require a clean fast-forward of local `main` — never start a root on `edce8816…`, whose single-manifest parser is the §2.3 regression;
3. record the resulting base SHA before creating each task worktree;
4. hand `TASK-260720-z9j4c9` the landed schema-v6 root plus canonical/legacy manifest-resolution expectations, and require the §4.3 `test_skill_manifest_resolution_vectors` gate against the already-pinned rc.3 checkout;
5. hand `TASK-260720-z2z795` the accepted rc.5 manager §§2.5–2.6 text and explicitly retire the stale pre-build recovery step from its diagram;
6. keep the root changes within the exact files above, and leave the committed conformance `ref:` untouched in both roots;
7. complete the pending `TASK-260729-v5hqnv` review cycle before treating any retargeted brief text as accepted, then land/publish the accepted `TASK-260729-3nx97g` golden candidate through its owning protocol workflow without treating board `done` as an upstream ref or CI pin;
8. resolve or replace the remaining literal rc.4 gates `TASK-260720-3ag6pi` and `TASK-260720-jrrgw9`.

## 9. References

### CocoaSkills upstream at `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`

- `pyproject.toml`
- `.github/workflows/ci.yml`
- `.github/workflows/distribution-smoke.yml`
- `.github/workflows/release.yml`
- `src/csk/{__main__,cli,config,skillspec,skillcheck,locking,installer,global_install,global_bins,env_files,shell_init,gc,consumers,snapshot,audit_registry,protocol_json,project_resolver}.py`
- `tests/{conftest,test_skillspec,test_skillcheck,test_locking,test_install,test_global_install,test_hybrid_scope,test_protocol_conformance,test_gc,test_shims}.py`

### Accepted Curator, protocol, and board evidence

- `.research/260729_curator-go-to-csk-parity-delta.md`
- `.task-board/.resources/TASK-260729-3nx97g/TASK-260729-3nx97g_results.md`
- `.task-board/.resources/TASK-260729-3nx97g/TASK-260729-3nx97g_review-verdict.md`
- `.task-board/.resources/TASK-260729-v5hqnv/TASK-260729-v5hqnv_rc5-brief-retarget-audit.md`
- `.task-board/.resources/TASK-260729-v5hqnv/TASK-260729-v5hqnv_review-verdict-cycle-1.md`
- `.task-board/.resources/TASK-260729-v5hqnv/TASK-260729-v5hqnv_rework-cycle-2-corrections.md`
- `.task-board/.resources/TASK-260729-1b9tc3/TASK-260729-1b9tc3_rc5-conformance-consumer-design.md` §§2.2–2.3, §9, §12
- `.task-board/.resources/TASK-260729-1b9tc3/TASK-260729-1b9tc3_baseline-pin-rc2.log` (`98 passed`, `exit=0`)
- `.task-board/.resources/TASK-260729-1b9tc3/TASK-260729-1b9tc3_candidate-rc5.log` (`1 failed, 97 passed`, `exit=1`)
- `.task-board/.resources/TASK-260729-35tb37/TASK-260729-35tb37_review-verdict-cycle-1.md`
- pinned rc.2 conformance export: `.temp/TASK-260729-1b9tc3/pin-cbe912d0/conformance/v1`
- upstream-pinned rc.3 conformance ref: `relux-works/curator-spec@00b1688a9b2457ca397a0bb550acf47cad8ee967`
- accepted Curator composite: `.temp/TASK-260720-1ljev5/worktree/internal/{skillspec,skillcheck,managerlock,transaction,staging}`
- accepted old rc.5 snapshot: `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-2kp3tv/curator-spec-worktree`
- accepted regenerated golden candidate: `.temp/TASK-260729-3nx97g/worktree`
- protocol text: `profiles/manager.md` §§2.5–2.6
- schema declarations: `schemas/v1/{agent-skill-v6,csk-skill-v6,common}.schema.json`
- root diagrams: `.task-board/.resources/TASK-260720-{z9j4c9,z2z795}/`
