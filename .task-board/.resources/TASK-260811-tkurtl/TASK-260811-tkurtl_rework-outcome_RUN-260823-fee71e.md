# TASK-260811-tkurtl rework outcome — reviewer RUN-260823-fee71e

Scope: exactly the three CONFIRMED findings of
`TASK-260811-tkurtl_review-verdict_RUN-260823-fee71e.md`. Accepted items 1, 4,
6, 7, and 8 are preserved and re-proved by the unchanged vectors that cover
them. Nothing is staged or committed.

## Finding 1 — interop capture is now selection-neutral (scope item 5, CGP05)

**Defect.** `classifyTargets` walked only `state.selected` and
`directTargetDependencies` filtered declared edges through `conditionSelected`,
so per-target interop capture nodes and `interop_boundary` nodes existed only
for the destination-selected subset. A conditional edge made the republished
capture digest destination-dependent while the upstream `swiftpmsource` capture
stayed neutral.

**Fix.**

- `internal/swiftpminterop/interop.go` — `classifyTargets` now classifies the
  *condition-neutral* reach of the selected product. `declaredTargets` indexes
  every declared target that the accepted capture published as a target unit;
  `declaredReach` walks every declared target and product edge from the
  destination-selected seed set with every condition ignored. Each
  `TargetInterop` carries the new `Selected` flag.
- `directTargetDependencies` no longer filters by condition. It returns
  `targetEdge{index, condition}` so the declaring predicate travels with the
  edge into boundary derivation, include search roots, and module resolution.
- `internal/swiftpminterop/boundaries.go` — `deriveBoundaries` derives a
  boundary for every declared consumer→C-family edge regardless of condition.
  `Boundary` gains `Condition` (selection-neutral predicate) and `Selected`
  (exact destination verdict). The two destination-profile gates — direct
  Swift/C++ interoperation and the Objective-C runtime — now apply only to a
  selected boundary; the declaration-level gate (a C++ public interface without
  `interoperabilityMode(.Cxx)`) stays unconditional because it is a declaration
  defect, not a destination fact. Language-profile and unsafe-setting gates in
  `classifyTargets` are likewise applied only to the selected subset.
- `internal/swiftpminterop/graph.go` — the `consumes_interop` capture edge
  carries the declaring condition. Binding edges (`uses_tool`, `targets`,
  boundary `requires`) are emitted only for selected targets and boundaries:
  the binding overlay names exactly the identities the destination selects, and
  the shared projector rejects a binding edge with a pruned endpoint.
- `internal/swiftpminterop/boundaries.go` — `verifyReads` observes only selected
  targets; a pruned target is captured but never compiled, so it neither
  supplies nor weakens read evidence.

**Shared-contract change (fixed at source, not worked around).**
`ConsumesInteropPayload` in `internal/closuregraph/edge.go` gains an *optional*
`Condition`, the same shape `RequiresPayload` already carries. This was
necessary, not cosmetic: `selectionReachability` reaches an `interop_boundary`
forward from the consumer action and then traverses `provides_interop` in
reverse to its provider. With an unconditional consumer side, a boundary — and
through it its provider target — stayed reachable on every destination, so
`ProjectActive` could not record the pruned verdict the reviewer asked for. The
field is absent-by-default: an unconditional consumer side canonicalizes byte
for byte as before, and the 53 labeled canonical goldens still verify.

**New/changed tests.**

- `TestCGP05ConditionalEdgeKeepsInteropCaptureSelectionNeutral` — the reviewer's
  probe scenario: `.when(platforms: [.macOS])` on App→CLib, closed against the
  accepted Darwin and Linux destinations. Asserts byte-identical capture
  digests, equal target/boundary counts, opposite `Selected` verdicts, opposite
  `ActivationSelected`/`ActivationPruned` boundary activations, and differing
  active-projection and evidence identities.
- `TestConditionalProductDependencyIsProjectedExactlyOnce` — assertions moved to
  the active projection: the boundary carries the condition, is `Selected`, its
  node activation is `ActivationSelected`, and the conditional `consumes_interop`
  edge activation evaluates true.
- `TestPrunedConditionalDependencyDeclaresNoBoundary` →
  `TestPrunedConditionalDependencyIsCapturedAndProjectedPruned` — the pruned
  destination still captures the target and the boundary in full, reports
  `Selected=false`, projects the boundary and all four per-target interop nodes
  as `ActivationPruned`, and emits no exact binding edge for the pruned
  declaration.
- `TestConsumesInteropConditionIsProjected` (closuregraph) plus conditional
  `consumes_interop` cases in the codec round-trip and strict-optional-typing
  matrices.

## Finding 2 — the include scanner fails closed on non-literal operands

**Defect.** `includePattern` matched only literal operands, so
`#define CURATOR_SECRET </etc/passwd>` + `#include CURATOR_SECRET` produced no
reference and no diagnostic. In the only mode reachable today (`not-observed`,
no `ReadSetProvider` until `TASK-260811-2qfnai`) the declared static closure is
the entire header proof, so this was a fail-open inconsistent with the
fail-closed module-map parser.

**Fix.** `internal/swiftpminterop/headers.go` now isolates every
`#include`/`#import`/`#include_next` directive line with `directivePattern` and
requires `literalIncludeOperand` to resolve the operand to an exact quoted or
angled literal. A macro-computed operand, an empty spelling, an unterminated
delimiter, or any trailing token other than a `//` or `/*` comment is reported
as `swiftpm_header_input_undeclared` naming the target, source, and directive —
never dropped.

**New tests.** `TestH09ComputedIncludeDirectiveIsRejected`, with subtests for a
macro operand in a source, a macro operand in a public header, an unterminated
literal, an empty spelling, a trailing token after the literal, and a positive
case proving a trailing comment is still resolved exactly.

## Finding 3 — publicHeadersPath is honored; non-representable layouts fail closed

**Defect.** `swiftpmsource.Target` dropped SwiftPM's `publicHeadersPath` and
`publicHeaderRoot(target.Path, "")` hardcoded the `include` default, so a
package with a custom public-header directory got a *generated* module map for
the wrong directory while the real escaping map was never parsed — a defeated
H03.

**Fix.**

- `internal/swiftpmsource` (additive): `Target.PublicHeadersPath`, decoded from
  `dump-package`'s `publicHeadersPath` in `executor_runtime.go`, and bound into
  `manifestDigest` under `public_headers_path`. Because
  `targetDeclarationDigest` already includes `manifest_digest`, the declaration
  is now bound into the target-unit node identity as well, so a silent
  public-header relocation is manifest-replay drift. No existing field, record,
  or behavior changed.
- `internal/swiftpminterop/headers.go`: `publicHeaderRoot` takes the declared
  value and applies SwiftPM's `include` default only when the target declared
  none. A declaration this profile cannot represent exactly — absolute, a
  Windows drive path, containing a backslash or control character, or resolving
  outside the target path — is `swiftpm_target_platform_unsupported` rather than
  a silent substitution.
- `confineModuleMapLayout` rejects a C-family target whose admitted tree holds a
  `module.modulemap` or `module.private.modulemap` anywhere outside the resolved
  public-header root with `swiftpm_modulemap_escape`. This is the reviewer's
  conservative guard: such a map is one SwiftPM or Clang may honour while this
  stage would never parse, confine, or digest it.

**New tests.** `TestH10DeclaredPublicHeadersPathIsHonored` (custom directory
inventoried and its custom map parsed; the reviewer's `Sources/CLib/pub`
escaping map now rejected as `swiftpm_modulemap_escape`) and
`TestH11NonRepresentablePublicHeaderLayoutFailsClosed` (absolute, parent escape,
escaping suffix, Windows drive; plus a module map beside and nested below the
public-header root).

## Negative controls

Each new vector was proved non-vacuous by reverting its fix in the working tree,
running the vector, and restoring the file:

| Reverted fix | Vector run | Result |
| --- | --- | --- |
| `literalIncludeOperand` result ignored (old `continue`) | `TestH09…` | FAIL — 5 subtests, all `err=<nil>` |
| `publicHeaderRoot(..., "")` + guard removed | `TestH10…`, `TestH11…` | FAIL — 8 subtests, including both silent-pass layouts |
| condition filter restored in `directTargetDependencies` + selected-only classification | `TestCGP05Conditional…`, `TestPrunedConditional…` | FAIL — capture digests diverged (`b1c1e85c` vs `d907f1c2`), boundary dropped |

## Preserved accepted behavior

Items 1, 4, 6, 7, and 8 of the verdict are unchanged. `S02`–`S09`, `H01`–`H08`,
`P01`–`P09`, `CGN03`, `CGN09`, `CGN15`, the exec-seam guard test, and the
portable-honesty (`not-observed` / fail-closed verified) contract all still pass
without modification. `P01` still captures dormant plugin and macro declarations
without classifying them: the condition-neutral walk skips and does not traverse
extension targets, and still rejects one the destination actually selects
(`P02`–`P04`, `P06`, `P08`).

## Files changed

| File | Change |
| --- | --- |
| `internal/closuregraph/edge.go` | optional `Condition` on `ConsumesInteropPayload` (field, `condition()`, `validate`, `value`, decoder) |
| `internal/closuregraph/codec_test.go` | conditional consumer side in the round-trip and strict-optional matrices |
| `internal/closuregraph/condition_projection_test.go` | `TestConsumesInteropConditionIsProjected` |
| `internal/swiftpmsource/types.go` | `Target.PublicHeadersPath` |
| `internal/swiftpmsource/executor_runtime.go` | decode `publicHeadersPath` |
| `internal/swiftpmsource/manifest.go` | bind `public_headers_path` into the normalized manifest digest |
| `internal/swiftpminterop/interop.go` | condition-neutral classification, declared reach, `targetEdge`, honored public-header root, layout guard |
| `internal/swiftpminterop/boundaries.go` | neutral boundary derivation, condition/selected verdicts, selection-gated profile checks, selected-only read observation |
| `internal/swiftpminterop/graph.go` | conditional `consumes_interop`, selection-gated binding edges, evidence record fields |
| `internal/swiftpminterop/headers.go` | fail-closed include grammar, honored/validated public-header root, `confineModuleMapLayout` |
| `internal/swiftpminterop/{parser,conformance,modulemap}_test.go` | reworked and new vectors, activation helpers |

`internal/swiftpmsource/manager.go` retains the three additive accessors from
the previous run; it is unchanged by this rework.

## Not run

The monolithic `go test ./...` including the full `cmd/curator` package exceeds
one bounded call and remains the Orchestrator's gate. `cmd/curator` references
only `closuregraph.ID` (a string type) and no SwiftPM package, so this delta
cannot reach it; `go vet ./...` type-checked its test files and a bounded subset
of its closure/assurance tests ran green.
