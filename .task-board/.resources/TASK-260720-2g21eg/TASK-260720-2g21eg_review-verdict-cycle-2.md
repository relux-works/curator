# TASK-260720-2g21eg review verdict — cycle 2

## Verdict

Changes requested. Route to `to-dev`.

The cycle-2 changes repair the earlier launcher/package-tree checks and teardown
propagation, but the implementation still does not satisfy the closed rc.5
worker identity boundary or the per-operation native-control availability probe
requirement. These are implementation defects with viable in-scope fixes, not a
human-only or external stop-the-line boundary.

## Candidate reviewed

- Base/branch HEAD: `97a0ed362631cd1d977151534fab0642a4f8baab`
- Worktree:
  `/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2g21eg/worktree`
- `src/csk/builds/go_v1.py`:
  `1a395b52b8bc7c0caf35565f06f1e9a7fe71faa97865955309be657ff9f6766f`
- `src/csk/cli.py`:
  `9e0724b53e6fbcd86f967f611c19336130620d9a7fb7500b9dc4d879dc35a92c`
- `tests/test_builds_go_v1.py`:
  `5acbcac78d935de7f7a796ae1e553a89ea2c96057fbeaa14ffbf37b9c8a3a015`
- `tests/test_builds_go_v1_fixture.py`:
  `ee47bf895deea53920a99ff8ee4d4d33ad897c71d0fcc2729c4fb1c47780e4d2`

## Blocking findings

### 1. The installed worker's complete Python startup TCB is not identity-bound

`_resolve_manager_identity` binds the launcher, interpreter executable/link
chain, and only the installed `site-packages/csk` subtree. The worker is then
started through the ordinary console-script shebang with Python `site`
processing enabled. Mutable startup components beside that subtree, including
executable `.pth` hooks, can therefore run before `csk.cli` and before the worker
can present its identity proof without changing the manager identity digest.

Installed-wheel reproduction:

1. Install the reviewed final wheel in a fresh venv.
2. Add an executable `.pth` hook as a sibling of `site-packages/csk`.
3. Resolve the manager identity before and after mutating the hook.
4. Run the real accepted native Go fixture through that installed manager.

Observed:

- identity before:
  `4436da493c3bc885ba8a8fc4c7f6f2fd31bf0b69f4d99412efe2768cc521c967`
- identity after:
  `4436da493c3bc885ba8a8fc4c7f6f2fd31bf0b69f4d99412efe2768cc521c967`
- hook is outside the bound package tree (`startup_hook_bound=False`)
- hook marker was written during the accepted build
- accepted fixture result: `1 passed, 3 deselected`

Thus an otherwise accepted operation executes mutable, unverified manager-worker
startup code. This violates the mandatory
`identity-verified-manager-owned-worker`,
`pre-launch-worker-identity-verification`, and
`post-exec-identity-reverification` controls.

Required rework: make the launched worker's complete executable/startup/runtime
TCB immutable and identity-bound before any worker code can execute (for example,
a genuinely self-contained manager-owned executable, or an equivalently closed
Python launch whose every mutable startup component is covered). Add an
installed-wheel negative that inserts or mutates a `.pth`/`sitecustomize` startup
component across the launch boundary and proves
`build_execution_worker_identity_invalid`, no compiler start, and no result.

### 2. Inventory entries marked unavailable are not actually probed

`probe_native_controls` calls `_native_probe` only when the frozen inventory
already says a control is available. For entries marked unavailable it copies
the static label directly into evidence and still reports
`probed_at=pre-worker-launch`.

On the reviewed macOS host:

- inventory entries: 5
- real probe calls: 3
- skipped calls:
  `active-process-count-limit`, `aggregate-memory-limit`
- both skipped entries were nevertheless emitted with
  `probed_at=pre-worker-launch`

The acceptance contract requires availability to be probed once per operation
before worker launch and expressly excludes a cached, configured, host-label, or
build-time-constant result from being a probe.

Required rework: perform one genuine availability probe for every one of the
five inventory controls on each supported host, reconcile the measured result
with the exhaustive frozen inventory, and reject contradictory evidence with
the declared error. Update tests so macOS and Windows inventories each assert
exactly five real probe invocations, including controls expected to be
unavailable.

### 3. The required focused gate produced an order-sensitive wrong diagnostic

With the authoritative accepted conformance root, the focused command collected
178 cases and produced:

- 173 passed
- 4 skipped
- 1 failed

The failing native case was
`manager-package-replaced-after-launch`; it expected
`build_execution_worker_identity_invalid` but received
`build_execution_control_unavailable`. The same case passed in ten isolated
runs, and the subsequent full suite passed, so the observed gate is unstable or
order-sensitive rather than a consistently reproducible product failure.

Required rework: determine and remove the race/error-precedence instability, then
provide repeatably green focused and native-fixture gates. A required gate that
has already emitted the wrong public diagnostic cannot be accepted solely
because a later, differently ordered suite passes.

## Verification evidence

- Correct accepted-root full pytest: `861 passed, 6 skipped` (867 total);
  JUnit attached.
- Correct accepted-root focused pytest: `173 passed, 4 skipped, 1 failed`;
  JUnit attached.
- Isolated native identity negative: 10/10 passed; stress log attached.
- Real accepted Go fixture through the unbound startup hook: passed while the
  hook executed; reproduction logs attached.
- Strict mypy: `Success: no issues found in 60 source files`.
- Hygiene: compileall, tabnanny, `git diff --check`, and trailing-whitespace
  check passed.
- Candidate hashes remained unchanged after review.

No product code was modified during review.
