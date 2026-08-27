# TASK-260728-1yhuqi — review verdict cycle 6

## Verdict

**ACCEPTED**

The cycle-5 blocking finding is closed. The task satisfies its acceptance
criteria and definition of done. Route to `done`.

This reviewer run is not goal-bound (`task-board spawn goal
RUN-260729-4a7b86` reported no active run goal).

## Closure of the cycle-5 finding

The authoritative probe session now implements the normative order:

1. `AdmitSources`
2. one `swiftc -###` graph command
3. `VerifyPlan`
4. `Reverify` over every plan binding plus
   `ReverifyAdmittedSources` over every admitted source
5. one `swiftc compile_argv` command

`runSession` returns before step 5 when either re-verification function reports
a finding. The outcome is
`build_execution_control_unavailable` /
`swift_permit_binding_changed`, exactly one manager-started graph command, no
compile child, and no artifact.

The binding repairs are internally consistent:

- a binding records the plan token as `raw` and the actually identified path as
  `checked`;
- a correctly absent output records and re-checks its operation-private parent;
- an absent plugin remains absent only when `Lstat` returns `ENOENT`; existence,
  permission/I/O failure, dangling resolution, or any other result fails closed;
- Stage B records a SHA-256 digest of every admitted source and the permit
  re-reads and compares those bytes, so an in-place modification is detected
  even if size, mode, and mtime are restored.

The decision and implementation reference state the same session order,
binding shape, failure outcome, and honest residual: the permit closes the
graph-to-permit interval, while ownership of the manager-distribution roots and
operation-private staging bounds the permit-to-open interval. The documents do
not claim that the latter interval is eliminated.

## Independent replay

Submitted probe archive SHA-256:
`4465da8689f68031ef7c1908369c6cdf44aa8e18ebdc3f6db5de9dbfef37f5b9`,
matching the producer gate log.

Standalone module:

- `gofmt -l .`: exit 0, no paths
- `go vet ./...`: exit 0
- `go test ./... -count=1`: exit 0
- `go build ./...`: exit 0

Native replay on Apple Swift 6.3.2 / macOS 26.5 arm64:

- 23/23 cases matched, 0 divergences
- 32 closure checks, 0 yielded a verdict
- 17/17 controls failed as required
- 70/70 structural checks matched
- executed P2 native admission held
- aggregate `green: true`, exit 0

Cycle-5 integrated cases independently reproduced:

- `S65`: permit ran over 35 bindings (33 plan + 2 sources), 0 findings,
  2 commands, artifact produced
- `S66`: absent plugin appeared; 2 findings, 1 command, no compile, no artifact
- `S67`: source gained `@`; stable source-admission detail present, 1 command,
  no compile, no artifact
- `S68`: source replaced by rename; digest change named, 1 command, no compile,
  no artifact
- `S69`: SDK presentation re-pointed; 1 command, no compile, no artifact
- `S70`: output parent replaced; 1 finding, 1 command, no compile, no artifact
- `S71`: synthetic bound executable changed; the control reached 2 commands and
  an artifact, while the mutated session stopped at 1 command with no artifact
- `S72`: on the same happy-path bindings the live re-check reported 0 findings
  and the retired raw-token model reported 5; on an unknowable absent plugin
  path the live check reported 1 and the retired check reported 0

Every expected-red control `C1` through `C17` was replayed individually and
exited 1. In particular, `C16` showed that removing permit-time re-binding lets
the appeared plugin reach command 2, while the live session stops after command
1; `C17` reproduced both retired binding-model defects.

The degraded replay used unresolved toolchain/platform roots, returned 23
`not_run` cases with the reason recorded, exited 0, and installed nothing.

## Contract and traceability review

The reviewed decision/reference retain the previously accepted closures:

- closed local `swift-v1` and external `swift-repository-v1` source recipes;
- direct `swiftc` pipeline with SwiftPM rejected because metadata queries that
  expose package declarations execute `Package.swift`;
- exact trusted toolchain, data-only SDK presentation, native target,
  runtime-library, linker, and plan-derived process closure identities;
- exhaustive reject/bound/inert handling for dependencies, Package.resolved,
  plugins, macros, scripts, response files, unsafe/compiler/linker flags,
  binary/system/native inputs, environment, network, target/SDK selection,
  configurations, and non-compiler-visible files;
- deterministic source enumeration, module-name derivation, artifact path,
  cache/receipt/marker/claim inputs, and deferred post-build signing;
- measured macOS arm64 qualification only, implementation-ready Windows
  qualification rules without a false claim, and explicit Linux follow-up.

All three declared dependencies are linked and `done`:
`TASK-260728-2spy93`, `TASK-260728-1g0z69`, and `TASK-260729-rhjxtx`.

## Focused spec and hygiene checks

The task-worktree decision and reference are byte-identical to the board
artifacts:

- decision SHA-256:
  `be75f586b2dc15c6c8d7361909119c682eed020f600cff4795e4922a3144f238`
- reference SHA-256:
  `6350a7aa37a106a8dac001fe2d6c446edcc76cd079d826617e035e6605ee3447`

The focused spec validator reproduced the producer's expected-red link failure
only in copied `docs/external-build-repositories.md`. The six links in the two
task-authored documents all resolve. The task spec worktree has no staged or
tracked modifications.

The reviewer did not edit product/spec/probe code, stage, commit, publish, pin,
install, or widen a platform claim.
