# TASK-260720-2g21eg review rework, cycle 3

## Reviewed source state

- CocoaSkills worktree:
  `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-2g21eg/worktree`
- Branch: `task/TASK-260720-2g21eg-go-v1-compile-driver`
- Base SHA: `495ad021847529ce5a544dba415ca2fe19949539`
  (clean local `main`, fast-forwarded and equal to `origin/main`; it carries the
  `TASK-260720-2dnqw2` canonical-build-metadata handoff that cycle 2 predated)
- Accepted rc.5 vectors:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1`
- Host used for native validation: Darwin 25.5.0 arm64, Go 1.25.5 darwin/arm64,
  CPython 3.14.4

## Cycle-2 findings addressed

### 1. The worker's complete startup TCB is now closed and identity-bound

The worker is no longer started through the ordinary console-script shebang with
`site` processing enabled.

- **Nothing unverified can execute.** The launch is now the one fixed vector
  `<bound interpreter> -S -s -B -P <installed launcher> __csk-go-worker-v1`
  (`go_v1.WORKER_LAUNCH_FLAGS`, `go_v1.worker_argv`).  `-S` removes `site`
  entirely, so no `.pth` hook, `sitecustomize`, or `usercustomize` can run;
  `-s` removes the per-user site directory; `-P` keeps the launcher directory
  off `sys.path`; `-B` makes "no bytecode writes" an interpreter flag instead of
  an environment value.  `import csk` resolves from the single manager-built
  `PYTHONPATH` entry, which is the bound site root.
- **Every mutable startup component is bound.** `_ManagerIdentity` gained a
  `startup` component (`_StartupIdentity`): the site root and standard-library
  root identities plus a digest over every `*.pth`, `sitecustomize.py`,
  `usercustomize.py`, `pyvenv.cfg`, `python._pth`, and `pybuilddir.txt` in the
  launcher prefix, both interpreter directories, the site root, and the
  standard-library root.  It flows through pre-launch verification, the worker
  request, the worker proof, the macOS/Windows mutation guard watch set, and
  every post-exec re-verification, so inserting or mutating one of these
  components across the launch boundary is `build_execution_worker_identity_invalid`.
- **The worker proves its own runtime.** The ready proof now carries a
  `runtime` record (`worker_runtime_proof`) that the manager validates
  (`validate_worker_runtime`): the `no_site`, `no_user_site`, `safe_path`, and
  `dont_write_bytecode` flags must all be set, the interpreter and entry point
  must be the bound ones, every `sys.path` entry must be the bound site root or
  the bound standard library, and every loaded module must be the bound launcher,
  a bound package-tree entry, or a standard-library file outside the site root.
- **Shell-wrapper launchers fail closed.** A launcher whose shebang does not
  exec a Python interpreter directly is rejected, so a shell can never enter the
  fixed four-node graph or hide the real interpreter invocation.

On the reviewed host this closes more than the reported `.pth` case: Homebrew
CPython ships a real 3773-byte `sitecustomize.py` in its standard library, which
the cycle-2 launch executed unbound in every worker.  It is now bound and cannot
execute.

Named negatives added:

- installed-wheel `accepted-with-bound-startup-hook`: an executable `.pth` is
  installed beside the manager package before the operation; the real Go build is
  accepted, and the hook's marker is never written, proving the hook could not run.
- installed-wheel `manager-startup-hook-inserted-after-launch`: the same `.pth`
  is inserted across the launch boundary; the operation returns
  `build_execution_worker_identity_invalid`, reaches
  `SESSION_STATES[:3]` only, starts no compiler, publishes no staging result,
  leaves no hook marker, and the worker process is gone.
- `test_manager_identity_binds_every_mutable_startup_component`: `.pth`
  insertion, `.pth` mutation, `sitecustomize.py`, `usercustomize.py`, and
  `pyvenv.cfg` insertion each reject.
- `test_worker_runtime_proof_rejects_unbound_startup`: 11 named runtime-proof
  negatives, plus a worker-level negative that a site-enabled startup is refused
  before any Go process.

### 2. Every inventory control is now really probed, once per operation

`probe_native_controls` no longer copies a static label for entries the frozen
inventory marks unavailable.  All five controls are measured on the host for
every operation, and the measured result is reconciled with the frozen
inventory:

- measured available but inventory unavailable →
  `build_execution_capability_evidence_invalid`
- measured unavailable but inventory available →
  `build_execution_control_unavailable` before worker launch (this is the
  mandatory portable `inventory-native-controls-applied` control)
- a probe that cannot run is not an availability result; it is
  `build_execution_control_unavailable`

New real measurements:

- macOS `active-process-count-limit` / `aggregate-memory-limit`: the probe reads
  a bound the host does implement (`kern.maxprocperuid` / `hw.memsize`) through
  `sysctlbyname`, confirms the manager's descendant domain, then asks the host
  for the domain-scoped facility the control needs
  (`kern.procpergroup_max` / `kern.memorystatus_pergroup_limit`).  Darwin answers
  that no such name exists, which is the measured
  `no-private-aggregate-domain` result; a host that grew the facility would
  measure available.
- Windows `per-file-size-limit`: the probe creates and releases a real private
  worker job, then asks the platform C runtime for the `setrlimit` entry point
  the `rlimit-fsize` mechanism needs.

Measured on this host, in inventory order:
`descendant-domain-termination=True`, `active-process-count-limit=False`,
`aggregate-memory-limit=False`, `per-file-size-limit=True`,
`inherited-handle-restriction=True` — exactly the frozen macOS inventory.

Tests: `test_probe_measures_every_inventory_control_once_per_operation`
asserts exactly five probe invocations in inventory order for **both** the macOS
and Windows inventories, including the entries expected to be unavailable;
`test_native_probe_measurement_matches_the_frozen_inventory` performs the five
real host measurements; `test_probe_rejects_a_measurement_that_contradicts_the_inventory`
and `test_probe_failure_is_not_a_probed_availability_result` cover the two
rejection surfaces.

### 3. The order-sensitive wrong diagnostic is removed structurally

Root cause: `_WorkerClient.launch` called `client.teardown()` inside its
`except` block, so any teardown failure **replaced** the diagnostic that had
already rejected the launch, and `teardown()` itself kept whichever failure
arrived first rather than the strongest one.  A slow or failing domain join, a
retained-handle release error, or a non-empty private bytecode cache could
therefore publish `build_execution_control_unavailable` for an operation that had
already established `build_execution_worker_identity_invalid` — exactly the
observed instability, and load- and order-dependent by construction.

Fixes:

- `launch` now preserves the primary error and records the teardown failure as a
  note; teardown still runs to completion.
- `teardown` and `build` aggregate failures through `_dominant_failure`, an
  explicit precedence (`worker_identity_invalid` > `hardened_claim_forbidden` >
  `package_influence_forbidden` > `capability_evidence_invalid` >
  `worker_protocol_invalid` > `control_unavailable` > other), so a resource or
  join failure can never rewrite an execution-boundary diagnostic. The weaker
  failure is still reported as a note, and a teardown failure with no pending
  error still suppresses the result and removes staging.
- Two teardown inputs that were environment-dependent are now flag-enforced:
  bytecode writes (`-B`) and user-site imports (`-s`), removing the
  `_verify_empty_worker_cache` and user-site paths as sources of a spurious
  `control_unavailable` during teardown.

Tests: `test_teardown_failure_cannot_mask_an_identity_diagnostic`,
`test_launch_teardown_failure_preserves_the_launch_diagnostic`, and
`test_failure_precedence_is_deterministic`.  The native fixture, including the
two after-launch identity negatives and the injected teardown failure, was run
10 consecutive times with no failure.

## Gate ledger

Every command below ran directly as a standalone process, with no `tee` and no
status-masking pipeline.  Reported results are the real exit codes.

| Gate | Result |
| --- | --- |
| Focused accepted-root pytest (source, toolchain, driver, metadata, identity, native fixture) | exit 0; 289 passed, 4 skipped |
| Full accepted-root pytest | exit 0; 962 passed, 6 skipped |
| Native installed-wheel Go fixture, 6 scenarios | exit 0; 6 passed |
| Native fixture stability, 10 consecutive runs | 10/10 exit 0; 6 passed each |
| `python -m mypy src/csk` (strict) | exit 0; no issues in 62 source files |
| `python -m compileall -q src tests` | exit 0 |
| `python -m tabnanny` on the four task files | exit 0 |
| `git diff --check` | exit 0 |
| trailing-whitespace check on the four task files | exit 0 |
| `python -m build --wheel` | exit 0 |
| `python -m twine check` on the fresh wheel | exit 0; PASSED |
| source vs fresh installed-wheel `go_v1.py` / `cli.py` | exit 0; byte-identical |

The focused gate's 4 skips are the two after-launch identity negatives and the
two startup/teardown negatives that are macOS-only or Windows-only by design,
plus platform gates in the neighbouring suites; the native fixture was then run
explicitly as its own gate and all 6 scenarios passed.

## Candidate digests

- `src/csk/builds/go_v1.py`:
  `6e04f58712fd07a45a97d681838d02a89583636a94914bd5ae93bc4d415f0571`
- `src/csk/cli.py`:
  `9e0724b53e6fbcd86f967f611c19336130620d9a7fb7500b9dc4d879dc35a92c`
- `tests/test_builds_go_v1.py`:
  `dcdea7a9ae43c45b3801945940d9eb50fd85093a0624232115114c1388667a0c`
- `tests/test_builds_go_v1_fixture.py`:
  `07e33bb2661a36ef5b033d25bde25b6d0d5f2c7374e7cb43ea79bfc9a661c570`
- wheel used for every native gate:
  `8de8b5bb6152489957e171722f5bbd73c00eccdd2124872b6e929de8125bf5c9`

No commit or index mutation was made.  The worktree still contains only the four
task-owned source/test paths.

## Scope and non-claims

- Native Windows execution was not run: the producer host is macOS.  The Windows
  launch vector, Job Object paths, `per-file-size-limit` probe, and startup
  component set (`python._pth`, `Lib/site-packages`) are covered by
  platform-neutral tests, the two-platform probe test, and strict mypy.  A
  native `windows-latest` run remains an explicit reviewer/CI follow-up; no
  native Windows success is claimed here.
- The standard-library *file tree* is not re-hashed per operation.  Instead the
  standard-library root identity is bound, the launch cannot execute any
  `site` startup component, and the worker proves that every module it loaded
  lies inside that bound root (or the bound package tree or launcher).  A module
  loaded from anywhere else — including a planted `python3XX.zip` beside the
  standard library — is `build_execution_worker_identity_invalid`.
- No deferred hardened guarantee is claimed; the six
  `deferred_capability_rejection_guards` still refuse all six.
