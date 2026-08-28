# Reviewer verdict for TASK-260811-32iojo

Verdict: **changes requested -> to-dev**

Reviewer run: `RUN-260823-66f288` (not goal-bound).

## Blocking findings

1. **PnP invocation does not load the generated PnP runtime state.** `Invoke` executes the selected Node binary with only `[entrypoint, args...]` (`internal/yarnmodernsource/materialize.go:330-349`). For a `nodeLinker: pnp` materialization, plain Node cannot resolve dependency packages unless `.pnp.cjs` is loaded (or an equivalent pinned Yarn/Node PnP launch path is used). Against the task's real Yarn 4.9.2 fixture, plain `node -e 'require("is-number")'` failed with `MODULE_NOT_FOUND`, while `node --require ./.pnp.cjs` succeeded. Evidence: `.temp/TASK-260811-32iojo-review-pnp-invoke-01.log`. The current fake runner never executes Node, so `TestS03S08N12N13ProtectedPrivateCachePnPReplayAndInvoke` masks this failure. This violates the AC that supported modern Yarn graphs materialize **and invoke** offline.

2. **Yarn platform conditions are parsed with incorrect semantics.** Pinned Yarn 4.9.2 emits multiple selectors as OR expressions such as `(os=linux | os=darwin)` and supports negated selectors. `conditionsMatch` strips `|`, parentheses, and `&`, then requires every remaining clause to match (`internal/yarnmodernsource/lock.go:576-590`), turning OR into AND and treating `!os` as an unknown field. A supported Linux target therefore fails closed for `(os=linux | os=darwin)`. Evidence: `.temp/TASK-260811-32iojo-review-condition-repro-01.log` reports `closure_graph_incomplete`; the pinned Yarn bundle's `Manifest.getConditions()`/`W8` implementation supplies the authoritative grammar. This breaks N10 exact selected/pruned condition behavior.

3. **The claimed S01-S08 / N01-N13 modern-Yarn conformance matrix is incomplete.** The package test names and assertions cover S01, S03, S06, S08 and parts of N01, N02, N04, N05, N10-N13. There are no modern-Yarn adapter vectors for S02, S04, S05, S07 or N03, N06, N07, N08. Common `nodesource` tests do not exercise the modern lock/cache/materializer wrapper, and the only PnP invoke test uses the non-executing fake described above. Add adapter-level negative/positive vectors, including a real or behaviorally faithful PnP dependency invocation.

4. **Several accepted `.yarnrc.yml` settings are neither closed to exact values nor represented in `ConfigurationDigest`.** `enableTelemetry`, `pnpEnableEsmLoader`, `defaultProtocol`, and `npmRegistryServer` are allowlisted (`internal/yarnmodernsource/lock.go:316`) but not validated, normalized into `Layout`, or included in the configuration ID (`lock.go:363-364`). These settings can affect manager behavior or generated state. The profile must bind their exact supported semantics or reject them.

## Validation

- Source SHA-256 values match the producer's outcome artifact.
- `go test -count=1 ./internal/yarnmodernsource`: pass.
- `go test -count=1 -race ./internal/yarnmodernsource`: pass.
- `go test -count=1 -cover ./internal/yarnmodernsource`: pass, 68.8% statements.
- `git diff --check`: pass before review.

The green focused suite is not acceptance evidence because it does not cover the two reproduced behavioral failures above. No product code was modified by this reviewer.

## Required rework

- Make the PnP runtime invocation explicit in C5 argv/environment and verify it through the protected executor with a dependency-importing fixture; keep node-modules invocation behavior separately tested.
- Implement the exact pinned Yarn condition grammar (OR groups, conjunction, negation) consistently in capture selection and C4 evaluation, with selected/pruned evidence tests.
- Close or reject every allowlisted rc setting and bind all accepted values into canonical configuration identity.
- Add the missing modern-Yarn S/N wrapper vectors and prove zero starts/publication on negative cases.
