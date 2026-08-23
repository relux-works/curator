# Reviewer verdict for TASK-260811-33ukne

Verdict: **changes requested -> to-dev**

## Review context

- Reviewer run: `RUN-260823-5da0b5`
- Spawn goal: none; the run is not goal-bound
- Reviewed scope: developer rework 03, both earlier reviewer verdicts, the
  accepted SwiftPM and graph/checkpoint contracts, `internal/swiftpmsource`,
  README profile documentation, and producer validation artifacts
- No product code was modified by this reviewer

## Findings

1. **The new Git permit/receipt IDs do not implement the accepted pre-C5
   derivation contract, and Git still executes outside protected execution.**
   The accepted contract requires every `derivation_permit` to bind admitted
   inputs, the exact C0 executable, argv/cwd/environment, host and target,
   process/read/write/network policies, expected evidence, resource limits,
   and immediate recheck; the receipt must bind the actual process/read/
   environment/write/network audit, exit result, evidence outputs, decision,
   and next causal head. `gitProcessPermit` contains only C0/tool IDs,
   executable, cwd, phase, a network string, argv, predecessor and sequence;
   `gitProcessReceipt` contains only permit ID, stdout digest/size and before/
   after fingerprints (`internal/swiftpmsource/git.go:36-47`). The broker and
   verifier launch direct `exec.CommandContext` processes
   (`git.go:319-335`, `408-428`) rather than the shared protected executor.
   Consequently `network="none"` for mirror verification is an unaudited
   label, not OS-level denial, and there is no enforced read/write/process or
   environment reconciliation. The in-memory `committed`/`issued` maps prove
   correspondence between locally invented IDs, but not the required security
   predicate. Inserting these IDs into C1/C3 therefore cannot satisfy the
   accepted derivation permit/receipt contract.

2. **The generated-lock receipt remains a custom domain ID, not a complete
   causally audited derivation receipt.** `BrokeredResolver.Resolve` hashes a
   lock digest, journal IDs and Git ID lists into
   `swiftpm-brokered-resolution-receipt-v1`, then stores the result only in its
   in-memory `issued` map (`internal/swiftpmsource/resolver.go:145-171`).
   `VerifyResult` proves equality with that map and re-hashes the same fields
   (`resolver.go:190-213`), but it cannot recover the missing executable
   audits from finding 1 and the receipt itself has none of the accepted
   receipt fields. `CaptureAndClose` nevertheless adds this ID and the custom
   Git receipt IDs to C3 `DerivationReceiptIDs`
   (`internal/swiftpmsource/capture.go:245-261`). The generated lock is thus
   better described than in rework 02, but still is not derived under the
   normative pre-C5 evidence contract.

3. **Mirror byte hashing is not shared recursive artifact admission.** The
   new mirror manifest walks regular files and records path, SHA-256, size and
   executable bit, then unconditionally declares `ADMIT_MIRROR_CONTAINER`
   (`internal/swiftpmsource/capture.go:316-354`). `admitMirrorTree` submits that
   hand-built ID straight to `CaptureStore.AdmitTree` without invoking the
   shared `artifactpolicy.Service` or proving a policy-approved transform for
   Git object/pack bytes (`capture.go:357-370`). This correctly prevents the
   checkout manifest ID from being substituted, but does not establish the
   required recursive compiled/opaque/ambiguous-content decision for the exact
   bare-repository bytes used as an admitted offline input. A new schema name
   and byte inventory cannot by themselves create an artifact admission
   capability.

4. **The C0 Git executable can escape its declared execution root.**
   `newGitProcessAuthority` joins `Git.ExecutableRelativePath` to
   `GitExecutionRoot` but never proves the relative path is clean and contained
   (`internal/swiftpmsource/git.go:65-74`); `validateToolchain` checks only that
   the string is nonempty. A clean `../outside/git` therefore resolves to an
   absolute regular file outside the selected root and can be accepted by
   `cleanAbsoluteFile`. This contradicts the type's stated below-root contract
   and the exact contained C0 toolchain boundary. The current mismatch test
   covers two absolute executable values, not path traversal from the recorded
   C0 relative path.

## Required rework

- Route every Git broker/verifier subprocess through the shared protected
  pre-C5 executor, or implement the complete accepted permit and receipt
  schemas with real environment/process/read/write/network enforcement and
  audits. Preserve the exact C0 executable binding and zero-start mismatch
  behavior already added.
- Make the generated-lock receipt resolve to those complete issued derivation
  receipts and bind the full manifest/Git evidence used by the in-process
  algorithm; do not place partial custom IDs in C3 as derivation receipts.
- Define and enforce a policy-approved mirror admission/transform that covers
  the exact Git repository bytes through the shared artifact boundary, or
  derive a deterministic same-kind mirror solely from already admitted source
  under a receipted transform. Keep the checkout-manifest substitution
  regression.
- Reject absolute, escaping, non-clean, or symlink-resolved-outside C0
  executable relative paths before any process start, and add the corresponding
  zero-start regression.

## Verification evidence

- Named four rework regressions: pass, including the real SwiftPM production
  fixture.
- `go test -count=1 ./internal/swiftpmsource`: pass.
- `go test -count=1 -race ./internal/swiftpmsource`: pass.
- `go vet ./internal/swiftpmsource`: pass.
- `golangci-lint run ./internal/swiftpmsource`: pass, `0 issues.`
- `go build ./...`: pass.
- Focused P01-P08-named extension/binary test: pass.
- Kotlin scope check: no Kotlin files under `internal/swiftpmsource`.
- Accepted canonical verifier: pass, 53 records and all references resolve.
- `git diff --check -- README.md internal/swiftpmsource`: pass.
- `task-board validate`: pass.
- Producer full-suite artifact digest verified as
  `9c2f56e5ea8a7b321ca32d01d0ca73cb5c77d40d13db6d7d2da59de369b22a71`;
  its contents report a green repository suite.
- Independent `go test -count=1 ./...`: pass across 53 package/result lines;
  `internal/swiftpmsource` passed in 35.061s and the slowest package was
  `cmd/curator` at 303.242s.

These defects are ordinary implementation rework. No external or human-only
blocker exists, so the correct verdict branch is `to-dev`, not `blocked`.
