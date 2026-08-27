# TASK-260720-2g21eg review verdict

## Verdict

Changes requested. Route the task to `to-dev`.

The fixed argv, session ordering, graph parser, environment, evidence-record
shape, artifact verifier, and native macOS fixture are green. The candidate
does not yet enforce three mandatory execution boundaries: the complete mutable
worker identity, the fingerprinted Go/tool process graph at actual use, and
fail-closed worker-domain teardown.

## Reviewed state

- CocoaSkills worktree:
  `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-2g21eg/worktree`
- Branch: `task/TASK-260720-2g21eg-go-v1-compile-driver`
- Base and current `HEAD`:
  `97a0ed870782b48eebc5a9c25a9cfa8fea5ff245`
- Dependency handoffs `TASK-260720-3c0ss2` and `TASK-260720-3j8pp5`:
  accepted/done with task-scoped outcomes.
- Reviewer run `RUN-260730-2d9409`: not goal-bound.
- Accepted rc.5 source:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree`
- Candidate scope:
  `src/csk/cli.py`, new `src/csk/builds/go_v1.py`, new
  `tests/test_builds_go_v1.py`, and new
  `tests/test_builds_go_v1_fixture.py`.
- Candidate SHA-256:
  - `src/csk/builds/go_v1.py`:
    `6920201335468881e3b58b07308da0bd38aa7e04aa899ac7c36b0857fd9ce325`
  - `src/csk/cli.py`:
    `9e0724b53e6fbcd86f967f611c19336130620d9a7fb7500b9dc4d879dc35a92c`
  - `tests/test_builds_go_v1.py`:
    `7fbda2d3075200209c42705a7acf48e939c250b29984cc82a3c84008982809df`
  - `tests/test_builds_go_v1_fixture.py`:
    `c46e7619f353013ab22be95e5e3a117369e273d6d1c35667705cf43adee447e3`

## Findings

### 1. The identity proof omits the installed worker's interpreter and package tree

Severity: high. Mandatory controls affected:
`identity-verified-manager-owned-worker`,
`pre-launch-worker-identity-verification`, and
`post-exec-identity-reverification`.

`pyproject.toml` installs `csk = csk.cli:main`, and the exact wheel-installed
manager used by the native fixture is a Python entry-point script whose first
line selects the task-private Python interpreter and whose body imports
`csk.cli`. The worker therefore depends on at least the launcher, interpreter,
and imported installed `csk` package tree.

The candidate identity at `go_v1.py:1198-1345` hashes only the launcher file.
Its protocol proof contains exactly `path`, `sha256`, and `size` for that file.
`run_worker` at `go_v1.py:1778-1810` proves only `sys.argv[0]`. Neither
`sys.executable` nor the installed package tree is recorded in the request,
proved in-session, or reverified after execution.

Independent probe result:

```text
manager_identity_ignores_package_tree= True
manager_identity_fields= ['path', 'sha256', 'size']
```

Changing the package code that defines the worker leaves the claimed worker
identity unchanged. A replacement worker can read the request secret, produce
the expected launcher proof, and control the protocol while every implemented
launcher check still passes.

Required rework:

- Bind every mutable worker TCB component required by the installed execution
  form, including the interpreter and installed package tree, into the
  pre-launch identity, request, in-session proof, and post-exec recheck; or
  distribute a self-contained manager-owned worker whose complete executable
  identity can be proved.
- Add an integrated installed-wheel negative that replaces package/interpreter
  state after the parent has loaded its code and before the worker proves
  identity. It must return
  `build_execution_worker_identity_invalid`, start no compiler, and publish no
  result.

### 2. Go and tool identities are paths, not execution-bound identities

Severity: high. Mandatory control affected:
`fixed-manager-selected-process-graph`.

`_WorkerPlan` at `go_v1.py:1518-1533` carries only
`go_executable`, `goroot`, and `tool_directory` paths. The worker validation at
`go_v1.py:1642-1659` checks path equality only. `ProcessRequest` has no expected
file identity, and `SubprocessProcessExecutor.run` at
`go_v1.py:1047-1099` invokes the current bytes at the path without a content or
file-identity check. It also observes only its direct `Popen`; it cannot see a
Go descendant outside `GOROOT/pkg/tool/<GOHOSTOS>_<GOHOSTARCH>`.

The parent fingerprints GOROOT before and after phases, but a
replace/execute/restore sequence is invisible. The independent probe replaced
the selected Go path after a simulated preflight, ran an outside executable,
restored the original bytes, and obtained:

```text
go_pre_post_digest_equal= True
outside_tool_child_ran= True
executor_started_count= 1
executor_returncode= 0
executor_stdout= maliciousn
```

This also shows why the current named tests do not establish the vector cases:

- `worker-executable-replaced-between-checks` only mutates a file and calls
  `identity.verify()` without launching a worker, although the vector requires
  `worker_started=true`.
- `unexpected-program-started-below-the-worker` injects a synthetic
  `ProcessResult(started=2)`. The real executor always reports its one direct
  `Popen`, even when that program starts an outside descendant.

Required rework:

- Carry the frozen Go/tool identity, not just paths, in the worker request and
  bind it to the accepted session.
- Recheck the selected Go image at the actual execution boundary using a
  race-resistant mechanism or an explicitly verified equivalent.
- Enforce and observe the manager-selected descendant surface so an executable
  outside the fingerprinted host tool directory is rejected as
  `build_execution_worker_identity_invalid`.
- Replace the two label-only cases with integrated negatives matching the
  vector's worker/compiler-start and publication assertions. Include a
  replacement/restore case.

### 3. Complete-domain teardown failures are suppressed

Severity: high. Mandatory control affected: `worker-domain-teardown`.

`_NativeControlDomain.terminate` at `go_v1.py:2507-2540` catches and discards a
Windows Job Object `CloseHandle` failure, sets `job_handle = 0`, and therefore
prevents `close()` from retrying or reporting it. On macOS, a non-ESRCH
`killpg` error can fall back to killing only the worker; kill errors are also
discarded. `_WorkerClient.teardown` then returns, allowing the already-created
`BuildResult` to be returned without proof that the complete domain terminated
and joined.

Independent seam result:

```text
windows_teardown_failure_propagated= False
windows_job_handle_after_failed_close= 0
```

Required rework:

- Propagate teardown failures and prevent the operation/result from returning
  when complete-domain termination or join cannot be proved.
- Preserve sufficient handle/domain state to retry or diagnose cleanup instead
  of discarding it after a failed close.
- Add Windows Job-close and macOS process-group-kill failure tests asserting
  no accepted `BuildResult`/publication, plus successful complete-domain join
  cases.

## Independent gate ledger

Project tools:

```text
task-board 0.23.0
Python 3.14.4
pytest 9.0.3
mypy 2.1.0
Go 1.25.5 darwin/arm64
```

Results:

| Gate | Result |
| --- | --- |
| Focused source/toolchain/driver/fixture suite with accepted rc.5 root | exit 0; 160 passed, 5 skipped |
| Full pytest with accepted rc.5 root | exit 0; 847 passed, 7 skipped |
| Strict `python -m mypy` | exit 0; 60 source files |
| Wheel-installed native macOS Go fixture | exit 0; 1 passed in 8.18s |
| Source vs installed-wheel `go_v1.py` and `cli.py` hashes | exact match |
| `python -m compileall -q src tests` | exit 0 |
| `python -m tabnanny` on the four candidate files | exit 0 |
| `git diff --check` | exit 0 |

The initial focused command used two wrong reviewer path assumptions
(`task-board` from the CocoaSkills worktree and `tests/test_builds_source.py`);
it exited before collecting tests. The corrected board-root checkpoint and
correct file `tests/test_build_source.py` produced the focused result above.

## Re-review entry conditions

1. Close all three findings with integrated regression tests, not synthetic
   status injection.
2. Rerun the accepted-root focused and full pytest suites, strict mypy,
   compileall/tabnanny/diff hygiene, and the source-matching wheel-installed
   native fixture.
3. Exercise the reworked Job Object path on native Windows CI or a qualified
   Windows host; the current macOS run makes no Windows success claim.
4. Preserve the exact two source-aware argv forms, one-list/one-permit/one-build
   protocol, fixed environment, result-only capability evidence, artifact
   never-run invariant, and unsupported-host fail-closed behavior.

