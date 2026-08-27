# TASK-260720-2g21eg — go-v1 compile-driver cycle-4 evidence

## Provenance

- Base, `HEAD`, and `origin/main`:
  `495ad021847529ce5a544dba415ca2fe19949539`
- Worktree:
  `/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2g21eg/worktree`
- Branch: `task/TASK-260720-2g21eg-go-v1-compile-driver`
- Accepted rc.5 conformance root:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1`
- Native host and toolchain: macOS arm64, Go 1.25.5
- Native manager:
  task-private `csk` installed from the cycle-4 wheel; installed
  `go_v1.py` and `cli.py` hashes matched their source files exactly.

## Produced changes

- `src/csk/builds/go_v1.py`
  - fixed source-aware list/build argv and clean build environment
  - complete package-graph, vendor/module, directive, containment, and output
    validation
  - the 13-state, one-list/one-permit/one-build authenticated worker protocol
  - exhaustive macOS/Windows native-control inventory and closed capability
    evidence
  - pre-launch, retained, in-session, and post-exec manager identity checking
  - bounded identity for every importable standard-library tree component and
    every reachable `python*.zip` archive/slot
  - manager-created hidden launch authenticated with a fresh secret delivered
    only through an inherited anonymous pipe; it is absent from argv,
    environment, request data, and public runtime proof
  - staged native-executable validation, hashing, permissions, and never-run
    boundary
- `src/csk/cli.py`
  - hidden dispatch requires the consumed manager-created launch capability;
    a literal or programmatic hidden argument falls through to parser rejection
- `tests/test_builds_go_v1.py`
  - exact argv, environment, graph, state, evidence, identity, protocol, and
    never-run checks
  - deterministic launch-authentication and standard-library/archive
    mutation/replacement negatives
  - retained macOS identity polling after a transient mutate-and-restore event
- `tests/test_builds_go_v1_fixture.py`
  - wheel-installed real vendored Go build
  - poisoned-environment and failure-boundary scenarios
  - literal hidden-mode parser rejection with a fake-Go never-run marker
  - task-private Python runtime replacement after worker launch, with teardown
    and no-list/no-build/no-artifact assertions

## Validation evidence

Every command listed here ran as a standalone process without `tee` or a
status-masking pipeline.

| Validation | Exit | Evidence |
|---|---:|---|
| Pre-change focused driver baseline | 0 | 106 passed, 1 skipped |
| Two cycle-4 regression tests before implementation | 1 | expected red: runtime-tree identity and hidden-launch authentication were absent |
| Focused driver unit suite after implementation | 0 | 110 passed, 1 skipped |
| `python -m mypy` using project configuration | 0 | no issues in 62 source files |
| Deliberately over-broad diagnostic `python -m mypy src tests` | 1 | 759 pre-existing strict-annotation errors in 42 test files; this is not the configured project gate |
| `python -m build --wheel` | 0 | cycle-4 wheel built |
| `python -m twine check` on the exact wheel | 0 | passed |
| Fresh task-private venv creation | 0 | Python 3.14 environment created |
| Exact wheel installation into the fresh venv | 0 | installed successfully |
| Source/installed SHA-256 comparison and installed `csk --version` | 0 | both changed modules matched; installed entry point ran |
| Native macOS real-Go fixture | 0 | 8 passed in 43.86s |
| Focused source/toolchain/driver/fixture pytest | 0 | 295 passed, 4 skipped in 42.79s |
| Full repository pytest | 0 | 968 passed, 6 skipped in 125.13s |
| `python -m compileall -q src tests` | 0 | no syntax/import compilation errors |
| `python -m tabnanny` on all changed Python files | 0 | clean |
| task-file trailing-whitespace and conflict-marker checks | 0 | clean, including all untracked task files |
| `git diff --check` | 0 | clean |

The initial task-local tool-readiness probe also truthfully recorded exit 127
for its absent `.venv` interpreter and exit 1 for system-interpreter pytest,
mypy, and build imports. A temporary environment was provisioned, then all
authoritative gates above used the coordinator-provided canonical repository
environment. See `tool-readiness-cycle4.md`.

Windows native execution was not run because this producer host is macOS.
Platform-neutral tests and strict mypy cover the Windows inventory, evidence,
bounded inherited-handle launch capability, Job Object setup, and rejection
surfaces; native Windows CI/review remains appropriate.

## Artifact digests

- wheel:
  `sha256:ddf1ba4f7e779df818a1217bf5a0754fd627132838e39e6cc0c8a594679e4b36`
- installed/source `go_v1.py`:
  `sha256:4d5d6b33ed7b5597531271d37372d926cc961519abc59346792893e2f0d71283`
- installed/source `cli.py`:
  `sha256:ad25c28b8918c3e03d945f4eef2c52e214e9161d25d477ec92def6a22f1aa943`
- focused pytest XML:
  `sha256:c8349ea97d8a966c4971466aca3fa9c36435dca600bd9fb1968ec733195d39f7`
- full pytest XML:
  `sha256:cf500173e9ba365421a7bff25ace64e9d44c3e7b77e8b4df18d816e54e596c84`
- native fixture pytest XML:
  `sha256:bb87aa3371fdefcafc7377d3847a84ad09d1d5f90cf76ec8f917572ff4c39e88`

## Review focus

- Re-run `tests/test_builds_go_v1_fixture.py` on native Windows with the
  wheel-installed manager.
- Inspect the inherited anonymous-pipe launch capability on both platform
  paths and the retained identity checks around request/list/permit/results.
- Confirm downstream planner/install/status work consumes capability evidence
  only as a result value and never adds it to cache, receipt, marker, or claim
  inputs.
