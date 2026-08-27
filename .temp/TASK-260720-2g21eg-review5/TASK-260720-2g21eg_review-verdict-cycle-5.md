# TASK-260720-2g21eg review verdict — cycle 5

## Verdict

Accepted. Route to `done`.

All prior blocking findings (cycles 2 and 4) are closed with concrete code
plus native negatives, every independent gate this reviewer ran is green, and
the implementation matches the acceptance criteria and the accepted rc.5
vectors exactly.

## Candidate and provenance

- CocoaSkills worktree:
  `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-2g21eg/worktree`
- Branch: `task/TASK-260720-2g21eg-go-v1-compile-driver`,
  integrated on current `origin/main` `HEAD`
  `15860e3f309888845b9271a257fb95f7c2825b56`
- Candidate scope is exactly: new `src/csk/builds/go_v1.py`,
  new `tests/test_builds_go_v1.py`, new `tests/test_builds_go_v1_fixture.py`,
  and a 13-line hidden-dispatch addition to `src/csk/cli.py`
  (tracked diff: 1 file, 13 insertions)
- `go_v1.py`:
  `sha256:abdf349c3ff2ebad7f17c25480655e65134378b1a402c7de9e010de3a684fcd4`
  — byte-identical to the tester's candidate-8 native evidence
- `cli.py`:
  `sha256:ad25c28b8918c3e03d945f4eef2c52e214e9161d25d477ec92def6a22f1aa943`
  — unchanged since the cycle-4 reviewed candidate
- Accepted rc.5 conformance root:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1`
- Review wheel built by this reviewer from the worktree source:
  `sha256:fac160fe64a0118458d3acf7509378fd223ba09a536d9f46c022c5add73405d1`;
  installed `go_v1.py` and `cli.py` matched the source hashes exactly.

No product or test file was modified during review.

## Independent gates run by this reviewer (macOS arm64, Go 1.25.5)

| Gate | Result |
| --- | --- |
| Focused `tests/test_builds_go_v1.py` with `CURATOR_CONFORMANCE_ROOT` set | 140 passed, 0 skipped; exit 0 |
| Project-configured strict `python -m mypy` | no issues in 65 source files; exit 0 |
| `python -m build --wheel` from the worktree | exit 0 |
| Fresh-venv wheel install + source/installed hash comparison | exit 0; byte-identical |
| Native wheel-installed real-Go fixture `tests/test_builds_go_v1_fixture.py` | 10 passed in 71.74s; exit 0 |
| `git diff --check`, tabnanny, trailing-whitespace scan on all task files | clean |

JUnit SHA-256:

- focused:
  `ce42cab877b3e5b706165a1994cd6e1e83aaaee26e47c2d53d35c07a9f0f12bc`
- native fixture:
  `99beeb8ddc1a08d169cd843039edfd1b9b74c3cd4ad3b349e0c3d83bca9e4d1f`

The native run covered all three accepted variants (plain, task-private
runtime, bound startup hook) and every negative: startup hook inserted after
launch, stdlib module replaced, runtime image replaced, manager package
replaced, worker executable replaced between checks, teardown failure, and the
poisoned-PATH hidden-literal parser rejection with the fake-Go never-run
marker absent. Poisoned environment (`GOFLAGS=-toolexec`, `GOENV`, `GOWORK`,
`GOTOOLCHAIN=auto`, attacker `GOPROXY`/`CC`/`HTTP_PROXY`, poisoned `PATH`) was
active on the accepted paths and did not influence the build.

## Prior blocking findings — verified closed

- Cycle-2: launcher-only identity → the manager identity now binds launcher,
  interpreter link chain, base executable, process image, loaded runtime image,
  bounded runtime trees, archives, startup hooks, and configuration, and
  polls them pre-launch, per protocol step, and post-exec
  (`_resolve_macos_interpreter_runtime`, `_resolve_windows_interpreter_runtime`,
  `_verify_retained_identity`, post-exec `identity.verify()`).
- Cycle-2: 3-of-5 native probes → `probe_native_controls` measures every
  inventory control once per operation, including expected-unavailable ones,
  and rejects probe/inventory contradictions with the declared errors;
  `test_probe_measures_every_inventory_control_once_per_operation` pins it.
- Cycle-4 finding 1: unbound macOS `Python.framework` runtime image → the
  framework image (`<home>/Python`) and `Python.app` process image are resolved
  as first-class executable identities; ambiguity fails closed; the native
  `manager-runtime-image-replaced-after-launch` negative passed in this
  reviewer's run.
- Cycle-4 finding 2: wrong Windows venv runtime root → `pyvenv.cfg` home is
  parsed under bounded read, base `python.exe` and exactly one `pythonXY.dll`
  are bound, `worker_argv` launches the identity-bound base executable
  (closing the venv-redirector ancestry defect), distlib `argv[0]` suffix
  recovery is fixed-form only, and the complete bounded `python_home` import
  tree excludes site trees disabled by `-S -s`.

## Acceptance-criteria verification

- Source-aware argv: `LIST_ARGUMENTS`/`BUILD_ARGUMENT_PREFIX` match the
  vector's `list`/`build` forms byte-for-byte and are issued only inside the
  worker; parent keeps only the three package-independent probe forms
  (TASK-260720-3j8pp5). Asserted against the vector file by a passing test.
- Session protocol: 13 states driven in order; one list, parent-side complete
  graph validation, one nonce-bound authenticated permit, one build; retry,
  second list/build, replayed nonce, out-of-order, oversize, and unknown
  messages tear down without a compiler start (named protocol cases passed).
- Environment: starts empty and fixes the vector `fixed_environment` exactly
  (private caches, `GOPROXY`/VCS off, `GOWORK` off, `GOTOOLCHAIN` local, CGO
  disabled, `GO_EXTLINK_ENABLED` 0, native target/tuning, locale, temp roots,
  empty `PATH`).
- 18 mandatory controls enforced; 5-control inventory exhaustive over exactly
  macos/windows; per-operation genuine probes; applied set matches inventory
  availability; nothing outside the inventory reported.
- Capability evidence: exactly one `capability-evidence-v1` record per
  operation with exactly `record_version`/`execution_policy`/`platform`/
  `controls`, one entry per control with exactly
  `name`/`availability`/`status`/`probed_at`; result-only value on
  `BuildResult`; all 8 consistency rules enforced with their declared errors
  (6× `build_execution_capability_evidence_invalid`,
  2× `build_execution_hardened_claim_forbidden`).
- Failure boundary matches the vector: missing mandatory portable control →
  `build_execution_control_unavailable` before worker-launch, nothing
  published; unavailable inventory native control and missing deferred
  hardened capability do not reject or block.
- All 6 deferred hardened guarantees refused by guards and never claimed.
- Named coverage: 14 identity/protocol cases, 8 package-influence cases, 11
  evidence cases — lists asserted equal to the vector case names by a passing
  conformance test; `worker-executable-replaced-between-checks` runs as a
  native fixture scenario.
- Graph rules: exactly one non-DepOnly main package; vendor/module
  consistency, workspace/toolchain switching, containment, load errors,
  cgo/native fields, `SysoFiles`, nonstandard `SFiles`, escaped embeds, and
  active nonstandard `cgo_import_dynamic` all fail before the permit.
- Output: exactly one bounded single-link regular executable in staging,
  native-header checked, hashed, permissioned 0700, TOCTOU-guarded, never run.
- Non-macOS/Windows hosts fail closed with `build_execution_control_unavailable`
  at the probe state, before any worker exists; no Linux control path.

## Caveats (non-blocking)

- Native Windows could not be executed by this reviewer (macOS host). The
  tester's candidate-8 evidence — same `go_v1.py` bytes as this candidate —
  records the native Windows fixture at 5 passed/5 host skips plus an
  independent accepted case at 1 passed. Native Windows re-verification in CI
  remains the standing follow-up already noted on the board.
- Whole-module branch coverage is 57%; the coverage checklist item was
  operator-clarified to the cycle-affected scope, measured independently by
  the tester at 91.94% combined changed lines/branches (>80%).
- Exposure of the evidence record in dry-run-plan-result/install-result/
  status-result and its exclusion from cache-key/receipt/install-marker/
  conformance-claim is emitted here as a result-only value; the consuming
  surfaces are downstream scope (TASK-260720-2x6mjn), consistent with this
  task's "without cache or install concerns" boundary.
