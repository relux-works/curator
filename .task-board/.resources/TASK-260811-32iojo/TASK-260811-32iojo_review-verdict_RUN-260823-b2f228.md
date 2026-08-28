# Reviewer verdict for TASK-260811-32iojo

Verdict: **changes requested -> to-dev**

Reviewer run: `RUN-260823-b2f228` (not goal-bound).

## Blocking findings

1. **PnP materialization and protected invocation still accept a nonfunctional loader.** `validateInstalledTree` checks only that `.pnp.cjs` is nonempty source text, then reports every selected graph package without reconciling the generated PnP package map (`internal/yarnmodernsource/materialize.go:726-745`). The claimed behaviorally faithful protected-executor test writes `module.exports = {};` as `.pnp.cjs`; its fake invocation checks only the `--require ./.pnp.cjs` argv prefix and never executes Node (`conformance_test.go:575-614`). The reviewer probe reproduced that `Invoke` accepts this state, while real Node 25.6.1 running the admitted `index.js` fails `MODULE_NOT_FOUND: is-number`. The separate manual Yarn/Node shell smoke does not exercise `Invoke` through the protected executor and therefore does not satisfy the prior required rework or the supported-graph offline invocation AC.

2. **Undeclared patch files are admitted.** `copyProjectSource` intentionally preserves every path below `.yarn/patches/` (`capture.go:745-759`), while `CaptureAndAdmit` verifies only paths already present in `Graph.patchBytes` (`capture.go:147-153`) and performs no discovered-versus-declared patch bijection. A fixture with no lock patch and no `ParseRequest.Patches`, but with `.yarn/patches/undeclared.patch` on disk, passed `CaptureAndAdmit`. This directly violates the explicit fail-closed requirement for undeclared patches.

3. **Unresolved required peers are silently omitted from the closed graph.** `buildEdges` permits an empty target for every peer edge (`lock.go:436-440`); `buildNodeCapture` then skips all empty-target edges (`capture.go:677-680`). The reviewer probe added required `react@npm:^18.0.0` peer metadata without a matching lock instance: `Parse` succeeded and emitted a peer edge with `To == ""`. The accepted profile requires unresolved peers to fail closed and peer context to remain explicit.

## Verification evidence

- Scratch overlay reproductions: `.temp/TASK-260811-32iojo-review-2/reproductions.log`; all three probes passed and logged the unsafe acceptance behavior.
- `go test -count=1 ./internal/yarnmodernsource`: exit 0.
- `go test -count=1 -race ./internal/yarnmodernsource`: exit 0.
- `gofmt -l internal/yarnmodernsource`: empty.
- `git diff --check`: exit 0.
- Source identities match the rework artifact: `capture.go cc9c16d4...`, `conformance_test.go 2475ffcf...`, `lock.go 61b65dda...`, and `materialize.go e2c5c599...`.
- The reworked condition evaluator's OR, conjunction, negation, and C4 tests pass; that prior finding is resolved.

The green focused suite cannot establish acceptance because it contains the PnP false positive and lacks the undeclared-patch and unresolved-required-peer negative vectors. No product code was modified by this reviewer.

## Required rework

- Reconcile regenerated PnP state against every selected/pruned package instance, and run a dependency-importing entrypoint through `Invoke` using the exact staged Node/Yarn 4.9.2 toolchain and real protected executor/network denial. A fake loader or argv-only assertion is insufficient.
- Discover `.yarn/patches` from the admitted tree and require an exact bijection with lock-declared, byte-bound patch inputs before manager execution.
- Reject every unresolved required peer before capture/C4; retain an explicitly optional unresolved peer only if the accepted canonical profile permits and records that exact pruned reason.
- Add regression vectors proving zero manager starts/publication for the two new negative cases and functional protected PnP dependency invocation for the positive case.
