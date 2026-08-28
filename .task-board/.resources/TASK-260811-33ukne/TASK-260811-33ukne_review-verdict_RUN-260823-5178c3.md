# Reviewer verdict for TASK-260811-33ukne

Verdict: **changes requested -> to-dev**

## Review context

- Reviewer run: `RUN-260823-5178c3`
- Spawn goal: none; the run is not goal-bound
- Reviewed scope: the prior changes-requested verdict, developer rework outcome,
  `internal/swiftpmsource`, task README changes, focused/full-suite evidence, and
  the accepted SwiftPM plus graph/checkpoint contracts
- No product code was modified by this reviewer

## Blocking findings

1. **Git acquisition and mirror verification are not bound to the C0 Git tool.**
   `GitBroker` and `GitMirrorVerifier` each accept an arbitrary absolute
   `GitExecutable` and launch it directly with `exec.CommandContext`
   (`internal/swiftpmsource/git.go:22-25`, `147-160`, `165-168`, `207-220`).
   Neither type carries the C0 `ToolIdentity`, invokes its recheck callback, nor
   commits a protected process permit. `CaptureAndClose` rechecks
   `config.Toolchain.Git` once before calling the opaque broker, but there is no
   structural relationship between that identity and the executable the broker
   actually runs, and the multiple clone/archive/fsck/rev-parse processes do not
   receive immediate time-of-use rechecks. The production-path integration test
   demonstrates the hole: it records the Swift wrapper and its digest as the Git
   C0 component, then separately runs the ambient Git path
   (`internal/swiftpmsource/swift_integration_test.go:160-168`). A green fixture
   therefore proves that unbound Git can acquire and verify closure bytes while
   the checkpoints attest to a different executable. This violates exact C0
   binding/recheck and the requirement that every mirror process use the
   admitted toolchain.

2. **Generated-lock resolution records a permit for a process that never runs
   and substitutes a synthetic ID for an executor receipt.**
   `resolutionPermit` commits the exact argv `swift package
   --disable-experimental-prebuilts resolve` and the SwiftPM tool fingerprint
   (`internal/swiftpmsource/capture.go:437-446`). `BrokeredResolver.Resolve`
   never launches that argv and never uses `closureexec.Executor`; it performs
   direct broker Git operations and marshals a lock in-process
   (`internal/swiftpmsource/resolver.go:66-140`). Its purported resolution
   receipt is only `DomainID("swiftpm-brokered-resolution-receipt-v1", ...)`,
   not a verified `closureexec.DerivationReceipt` for the committed command
   (`resolver.go:124-140`). `CaptureAndClose` then inserts this arbitrary valid
   ID into C1/C3 as if it were causal derivation evidence. This fails the
   intake/permit/receipt journal contract and leaves every executable Git step
   used to derive the generated lock outside the pre-C5 process evidence.

3. **Mirror intake binds classifier evidence for different bytes.**
   `captureMirror` captures the complete bare Git repository, but calls
   `AdmitTree` with `ArtifactManifestID: packageManifestID`, where that ID is the
   artifact manifest derived from the extracted source snapshot
   (`internal/swiftpmsource/capture.go:259-284`; caller at `capture.go:192`).
   The shared admission API explicitly requires classifier evidence for the
   exact captured bytes (`internal/closureexec/intake.go:84-103`, `285-297`) and
   does not itself verify that correspondence. A bare mirror contains Git
   object, pack, index, ref, and configuration bytes absent from the checkout
   artifact manifest. The resulting `MirrorIntakeReceiptID` therefore attests
   that the mirror was classified under evidence belonging to another tree.
   Later byte-digest and Git-object rechecks detect mutation but cannot repair
   the invalid initial admission. The mirror needs its own canonical,
   policy-appropriate artifact/intake evidence or a separately specified
   mirror-container admission profile that binds these exact bytes.

## Required rework

- Make the acquisition broker and mirror verifier consume the exact C0 Git
  tool identity and execute each Git subprocess through a permit/receipt path
  with immediate exact recheck; reject any executable/config mismatch before
  start. Update the real fixture so the recorded Git node is the Git actually
  launched.
- Make generated-lock evidence truthful: either execute the committed SwiftPM
  resolution argv under the protected executor, or define and commit the real
  broker/in-process derivation contract and causally receipt every executable
  broker step. Do not journal a plain domain ID as an executor derivation
  receipt.
- Classify/admit the exact mirror repository bytes under a valid mirror
  artifact profile and bind that mirror-specific manifest to its intake
  receipt. Add a negative test proving checkout-manifest substitution cannot be
  accepted.
- Add regression tests asserting C0/broker executable mismatch has zero process
  starts and that every recorded resolution/mirror receipt resolves to the
  exact committed permit and input bytes.

## Independent verification

- `go test -count=1 ./internal/swiftpmsource`: pass
- `go test -count=1 -race ./internal/swiftpmsource`: pass
- `go vet ./internal/swiftpmsource`: pass
- `golangci-lint run ./internal/swiftpmsource`: pass, `0 issues.`
- `go build ./...`: pass
- `git diff --check -- README.md internal/swiftpmsource`: pass
- `task-board validate`: pass
- accepted canonical verifier: pass, 53 records and all references
- producer full-suite log digest verified as
  `05d75d722ac6077f5f64ce2e1ff71dd79d338f2981794a77dc8622635d5304f7`;
  recorded full suite is green

An initial reviewer canonical command used a non-materialized attachment path
and exited 1 with `LoadError`; the corrected authoritative `.research` paths
matched the accepted SHA-256 values and passed. The failed probe is not counted
as product failure.

These are ordinary security-contract implementation defects. No external or
human-only blocker exists, so the correct branch is `to-dev`, not `blocked`.
