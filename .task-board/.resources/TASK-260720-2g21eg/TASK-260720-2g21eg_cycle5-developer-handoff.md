# TASK-260720-2g21eg — cycle-5 developer handoff

## Provenance and scope

- Preserved worktree:
  `/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2g21eg/worktree`
- Branch: `task/TASK-260720-2g21eg-go-v1-compile-driver`
- Base, `HEAD`, `origin/main`, and merge-base:
  `15860e3f309888845b9271a257fb95f7c2825b56`
- Accepted rc.5 conformance root:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1`
- The unstaged/uncommitted CocoaSkills task delta is exactly:
  - `src/csk/cli.py`
  - `src/csk/builds/go_v1.py`
  - `tests/test_builds_go_v1.py`
  - `tests/test_builds_go_v1_fixture.py`
- No stage, commit, push, tag, release, cache publication, install marker,
  shim, receipt, conformance claim, or artifact execution was performed.

## Implementation delivered for review

- The source-aware manager/worker engine owns the fixed list/build argv,
  manager-derived staging output, empty fixed environment, complete list JSON
  parsing, graph/vendor/module/path/directive validation, authenticated build
  permit, output hashing/permissioning, and the 13-state session.
- The installed manager re-executes in one hidden fixed mode authenticated only
  by a manager-created inherited anonymous-pipe capability. Literal and
  programmatic user selection fall through to parser rejection.
- macOS and Windows use the exhaustive five-control native inventory and emit
  exactly one closed `capability-evidence-v1` result record. Unsupported hosts
  fail closed before worker launch. Deferred hardened guarantees remain
  rejected and unclaimed.
- The worker identity binds the launcher, interpreter, package tree, complete
  importable runtime trees, startup components, process image, loaded Python
  runtime image, Go executable, and regular tool executables. Identity is
  checked before launch, proved by the worker, retained during execution, and
  reverified after execution.
- macOS binds the actual loaded framework/process Mach-O images. Its kqueue
  identity guard retains every bounded manager/runtime/tool path, serializes
  monitor lifetimes, raises manager descriptor capacity only for the bounded
  monitor plus fixed launch-capability FD set, and restores the exact prior
  `RLIMIT_NOFILE` during teardown.
- Windows parses the identity-bound `pyvenv.cfg`, resolves the base Python
  installation, binds `python.exe` and the loaded `python314.dll`, and launches
  the base interpreter directly so the authenticated worker is the manager's
  direct child rather than the grandchild of `venvlauncher.exe`.
- Windows restores only distlib's fixed stripped `.exe` suffix for `argv[0]`;
  no `PATH`, environment, descriptor, package value, manifest, shell, or user
  option contributes to launcher selection.
- Windows CPython necessarily includes `python_home` in `sys.path`, so the
  entire bounded importable Python-home tree and exact versioned archive slot
  are fingerprinted. Disabled `site-packages`/`dist-packages` trees remain
  excluded. Root-level import insertion and runtime-image replacement are
  covered as fail-closed negatives.
- Tests cover exact argv and environment, every named identity/protocol,
  package-influence, capability-evidence and consistency case, one-list /
  one-permit / one-build ordering and replay rejection, complete graph
  rejection, output verification, never-run behavior, poisoned environments,
  runtime replacement, native teardown, and unsupported-host fail-closed
  behavior.

## Exact handoff artifact

- Wheel:
  `candidate8/cocoaskills-0.12.6.dev13+g15860e3f3.d20260730-py3-none-any.whl`
- Wheel SHA-256:
  `8cc70ef4ef574088c0c017905b6a5609820a6f2bdc3e984b6b1669674ba9af4f`
- Source and installed-wheel `go_v1.py` SHA-256:
  `abdf349c3ff2ebad7f17c25480655e65134378b1a402c7de9e010de3a684fcd4`
- Source and installed-wheel `cli.py` SHA-256:
  `ad25c28b8918c3e03d945f4eef2c52e214e9161d25d477ec92def6a22f1aa943`
- `cmp` of both changed source modules against extracted wheel content:
  exit 0 for each.
- Installed `csk --version` on both macOS and Windows:
  exit 0, `csk 0.12.6.dev13+g15860e3f3.d20260730`.
- Local and remote Windows wheel SHA-256 matched exactly. Both macOS and
  Windows installed module hashes matched source exactly.

## Authoritative green gates

Every gate below ran directly as a standalone process without `tee` or a
status-masking pipeline.

| Gate | Exit | Result |
| --- | ---: | --- |
| Focused go-v1 unit pytest | 0 | 139 passed, 1 platform skip |
| Cycle-5 changed line/branch coverage | 0 | 93.13% lines, 88.68% branches, 91.94% combined |
| Accepted-root focused source identity/source/toolchain/metadata/go-v1/fixture pytest | 0 | 316 passed, 14 host/environment-selected skips |
| Full accepted-root repository pytest | 0 | 1158 passed, 21 skips in 79.40 s |
| Native macOS arm64 wheel-in-venv fixture, Go 1.25.5 | 0 | 10 passed in 63.99 s |
| Native Windows amd64 wheel-in-venv fixture, Go 1.25.5 | 0 | 5 passed, 5 host-specific skips in 290.27 s |
| Project-configured `python -m mypy` | 0 | no issues in 65 source files |
| Final `python -m build --wheel` | 0 | exact candidate-8 wheel built |
| `python -m twine check` on exact candidate-8 wheel | 0 | passed |
| `python -m compileall -q src tests` | 0 | passed |
| `python -m tabnanny` on all four task files | 0 | passed |
| `git diff --check` | 0 | passed |
| Explicit task-file trailing-whitespace check | 0 | no matches |
| Explicit task-file conflict-marker check | 0 | no matches |

The native macOS fixture exercised the accepted build, private runtime,
pre-bound startup hook, startup/stdlib/runtime/package/launcher mutation
negatives, and teardown failure. The native Windows fixture exercised the
literal hidden-mode negative, all three accepted build variants, loaded
runtime-DLL replacement and teardown, while macOS-only mutation injections
were skipped by their declared platform guards. Every accepted build verified
the staged executable and proved it was never launched.

## Non-green and superseded attempts

These are reported as failures and were not treated as passed gates.

- Initial remote command spelling: exit 1; corrected standalone
  `cmd.exe /c ver` probe exited 0.
- One intermediate local edit produced an invalid mixed dict
  comprehension: focused pytest exit 2 and mypy exit 2. The edit was corrected
  immediately; the authoritative gates above are post-correction.
- Windows attempt 1: exit 1, 10 failed. The fixture placed the task-private Go
  tree below a forbidden repository root and omitted indispensable Windows
  bootstrap variables.
- Windows attempt 2: exit 1, 3 failed / 2 passed / 5 skipped. It exposed a
  cold fingerprint timeout plus the venv-launcher grandchild boundary.
- Windows attempt 3: exit 1, 3 failed / 2 passed / 5 skipped. It reached the
  direct base interpreter and exposed distlib's stripped `.exe` suffix.
- Windows attempt 4: exit 1, 3 failed / 2 passed / 5 skipped. It exposed the
  unavoidable unbound `python_home` import root.
- Windows attempt 5: exit 1, 3 failed / 2 passed / 5 skipped. It exposed stale
  root/archive consistency validation after the Python-home tree was bound.
- Windows candidate 6 then exited 0 with 5 passed / 5 skipped. The exact final
  candidate 8 was rerun after the macOS-only guard changes and independently
  exited 0 as recorded above.
- The first macOS setup used `venv --copies`: exit 1, 9 failed / 1 passed.
  A copied launcher incorrectly appeared to be a private Python home without a
  stdlib; the canonical framework-symlink venv was used thereafter.
- Canonical macOS attempt before descriptor-capacity support: exit 1,
  8 failed / 2 passed with `EMFILE` before worker launch.
- Canonical macOS attempt with capacity below the fixed protocol FD range:
  exit 1, 8 failed / 2 passed. Candidate 8 covered both bounded descriptor sets
  and passed all 10 cases.
- An initial focused command used the wrong filename
  `tests/test_builds_source.py`: exit 4, no tests collected. The corrected
  filename was used for all focused results.
- A corrected focused diagnostic without the accepted conformance root exited
  0 with 235 passed / 46 skipped. It was superseded by the authoritative
  accepted-root run above.

## Evidence artifacts

- Native macOS JUnit:
  `TASK-260720-2g21eg_cycle5-macos-native-final.xml`,
  SHA-256
  `e532d2bfca8fb56986a44c70688f17c02a91afec82dc52c89ede26f1f304fb70`
- Native Windows candidate-8 JUnit:
  `TASK-260720-2g21eg_cycle5-windows-native-candidate8-final.xml`,
  SHA-256
  `849d22110c12a6eea6f30b36bf3f13d911f311fd73fe6a4c2cf11e2353bdfeb2`
- Accepted-root focused JUnit:
  `TASK-260720-2g21eg_cycle5-focused-final.xml`,
  SHA-256
  `dbdefe27127afb110751a0cc3900e9b12a0dbc4627c4cd7f84e809bd93fa4ec0`
- Full repository JUnit:
  `TASK-260720-2g21eg_cycle5-full-final.xml`,
  SHA-256
  `c203aaefee8235e707fcab8fd9537a05273307a5e93595b70ccc552250c66ccc`
- Changed-coverage report:
  `TASK-260720-2g21eg_cycle5-changed-coverage.md`, with exact reviewed
  baseline, commands, calculation method, initial below-target result, and
  post-test measurement.
- Earlier red JUnits are retained under the cycle-5 run directory rather than
  overwritten or represented as green.

## Review focus

- Confirm the macOS descriptor-limit raise is manager housekeeping for the
  retained identity monitor and fixed inherited capability, is serialized,
  and restores the exact previous limit on every teardown path.
- Confirm the Windows direct-base-interpreter launch preserves exact installed
  manager selection while eliminating the venv redirector's extra process.
- Confirm accepting Windows `python_home` as an import root is coupled to the
  complete bounded tree identity and exact archive-slot consistency checks.
- Recheck that only list/build are worker-side, each occurs exactly once, and
  capability evidence remains result-only and absent from cache, receipt,
  marker, or claim inputs.

This developer candidate is ready for review; board acceptance remains a
reviewer decision.
