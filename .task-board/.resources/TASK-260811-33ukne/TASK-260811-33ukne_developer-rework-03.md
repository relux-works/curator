# TASK-260811-33ukne developer rework 03

Status: ready for independent review

## Review blockers closed

1. `GitBroker` and `GitMirrorVerifier` now share one manager-owned
   `swiftpm-git-process-v1` authority created from the exact C0 Git tool.
   `GitExecutionRoot + ExecutableRelativePath` must equal the executable each
   component launches. Every Git invocation commits a unique causal permit,
   performs immediate exact fingerprint and executable-byte rechecks before
   and after execution, and records an authority-issued receipt. A mismatch
   rejects before the process-start observer.
2. Generated locks now commit the real
   `swiftpm-brokered-resolution-v1` in-process algorithm. The resolver retains
   and verifies the issued result, exact root intake, lock digest, Git
   permit/receipt pairs, manifest derivations, and C0 Git fingerprint. C1 and
   C3 contain the corresponding permits and receipts; no unexecuted
   `swift package resolve` argv or synthetic process receipt remains.
3. Each protected bare repository now has an independent canonical
   `swiftpm-git-mirror-artifact-manifest-v1`. It covers every regular node,
   path, byte digest, size, executable bit, exact pin revision/tree, policy,
   and verifier permit/receipt. Admission recomputes the protected mirror
   inventory and rejects a checkout artifact-manifest substitution.

The real production fixture records and launches the staged Git executable and
its own byte digest, rather than relabeling the Swift wrapper as Git.

## Regression evidence

- `TestGitC0ExecutableMismatchRejectsBeforeAnyProcessStart` proves exact C0
  executable mismatch with zero Git process starts.
- `TestBrokeredResolverGeneratesTransitiveLockBeforeMainCaptureR02` proves
  every generated-lock Git receipt resolves to an authority-retained permit
  and that every pair is present in C1/C3.
- `TestMirrorAdmissionRejectsCheckoutManifestSubstitution` proves package-tree
  artifact evidence cannot authorize bare-repository intake.
- `TestRealSwiftPMManifestRunsThroughProductionManagerAndExecutor` exercises
  production Git acquisition, mirror admission, controlled Swift manifests,
  and forced-lock offline replay.

## Validation

Every command ran as a standalone process. Exit codes below are the real gate
statuses.

| Command | Exit | Evidence |
| --- | ---: | --- |
| Four focused blocker regressions | 0 | `focused-regressions-03.log` |
| `go test -count=1 ./internal/swiftpmsource` | 0 | `go-test-focused-03.log` |
| `go test -count=1 -race ./internal/swiftpmsource` | 0 | `go-test-race-03.log` |
| Focused coverage | 0 | 80.2%; `swiftpmsource-cover-03.out` |
| `go vet ./internal/swiftpmsource` | 0 | `go-vet-03.log` |
| `golangci-lint run ./internal/swiftpmsource` | 0 | `0 issues.` |
| `go build ./...` | 0 | `go-build-03.log` |
| `make no-broad-suppression` | 0 | `no-broad-suppression-03.log` |
| Accepted Ruby canonical verifier | 0 | 53 records; all references resolve |
| Binary-target P01-P08 guard | 0 | `binary-guard-03.log` |
| `rg --files internal/swiftpmsource -g '*.kt'` | 1 | Expected no-match failure: no Kotlin files in scope; not reported as a passing command |
| `git diff --check -- README.md internal/swiftpmsource` | 0 | `git-diff-check-03.log` |
| `task-board validate` | 0 | Board valid |
| `go test -count=1 ./...` | 0 | Full uncached repository suite; `go-test-all-03.log` |

Full-suite log SHA-256:
`9c2f56e5ea8a7b321ca32d01d0ca73cb5c77d40d13db6d7d2da59de369b22a71`.

Coverage profile SHA-256:
`3300b790f7ca4f6028f0bf5dcf1580ceaaaa49b98bd0c9ac1096c3f4ef821fc1`.

No files were staged or committed.
