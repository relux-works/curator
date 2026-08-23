# TASK-260811-3kbf3l rework evidence

Implemented every change requested by reviewer run `RUN-260819-a16a5c`.

## Enforcement and diagnostics

- All production Cargo launches remain routed through committed `closureexec` permits and executor-issued receipts.
- `ManagerConfig.ProcessStartObserver` is installed before C0 registration and observes the sole portable launch seam. `TestBuildToolchainRegistrationStartsNoProcessBeforeC0` proves zero launches through construction, toolchain registration, assurance preflight, and C0, then observes the first launch only after a committed permit.
- The verified fixture provider now models attempted network, child-process, undeclared-read, and undeclared-write operations at `EnforceAndObserve`; it rejects before Cargo execution, receipt issuance, or publication.
- Rust build execution maps those verified boundary violations deterministically to `rust_undeclared_input` for RH05/RH06.
- Pre-existing build output roots now fail with `artifact_local_output_unreceipted` instead of being cleared.

## Named conformance map

- CGP05: `TestCGP05ReconcileReusesSelectionNeutralCaptureAndSeparatesTargetBinding`.
- R01/R08/R09 and protected cache: `TestRustConformanceR01R08R09RH05RH06RH08AndProtectedCache`; this uses a two-level registry closure, frozen fresh homes, executable validation, an independent second clean manager rebuild with matching logical identities and bytes, exact cache reuse, cache tamper rejection, verified attempt rejection, and planted-output rejection.
- R02: `TestRustConformanceR02GitBuildWithoutOriginalRemoteOrCache`; exact commit, recursive submodule receipts, original repository renamed away before metadata/build, frozen build publishes.
- R03/R05/R06/R07: `TestRustConformanceR03R05R06R07PathWorkspaceBuild`; transitive contained paths, all-target lock superset/native pruning, exact workspace package/bin, and in-closure `include_str!`/`include_bytes!`.
- R04: `TestRustConformanceR04FeatureSelectionKeepsCaptureNeutral` and `TestRustConformanceR04FeatureSelectedBuildsHaveDistinctIdentities`; same capture with disabled/enabled optional dependency, distinct active/command/output identities.
- RF09-RF11: `TestRustConformanceRF09RF10RF11FailBeforeCompilation`; unknown metadata kind/outside path, feature/target drift, config/wrapper rejection.
- RF12: `TestRustConformanceRF12ClosedTargetAndUnstableArtifactInputs`; custom/cross/multiple targets, unstable configuration, artifact dependency.
- RH01-RH04: named table cases in `TestValidateBuildUnitsRejectsEveryUnsupportedActiveUnit`.
- RH05/RH06/RH08: named verified-attempt and planted-output cases in the R01/R08/R09 protected-cache test.
- RH07: `TestRustConformanceRH07CapturePreservesCompiledGitAndPathDiagnostics`.
- RH09: `TestRustConformanceRH09ToolchainCopyIsDependencyArtifact`.
- RH10: `TestRustConformanceRH10NativeToolchainDriftStartsNoProcess`.

## Additional defects corrected

- Multi-package registry staging now reuses and verifies the immutable per-source `config.json` instead of attempting to overwrite its read-only first copy.
- Cargo event validation now accepts the canonical short path-package ID form only when its URL basename, package name, and version agree.

## Validation

- Targeted named Rust conformance command: exit 0, 62.801s.
- `go test -count=1 ./internal/rustsource ./internal/closureexec`: exit 0; Rust 91.209s, closureexec 3.245s.
- `go test -race -count=1 ./internal/rustsource ./internal/closureexec`: exit 0; Rust 143.072s, closureexec 7.445s.
- `make lint`: exit 0, 0 issues.
- `go vet ./...`: exit 0.
- `go build ./...`: exit 0.
- `go test -count=1 ./...`: exit 0; notable timings `cmd/curator` 459.915s, `internal/rustsource` 162.695s, `internal/closureexec` 23.163s.
- `git diff --check`: exit 0.
- `task-board validate`: exit 0, board valid.