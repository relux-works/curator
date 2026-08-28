# Reviewer verdict for TASK-260811-3kbf3l

Verdict: **changes requested -> to-dev**

Reviewer run: `RUN-260819-a16a5c`
Goal check immediately before verdict: run is not goal-bound.
Directives: none.
Reviewed producer outcome: `TASK-260811-3kbf3l_rework-and-verification.md`.

## Required changes

1. **Materialize and execute the mandatory named conformance families.** The task acceptance criteria require CGP05, R01-R09, RF09-RF12, RH01-RH10, and protected-cache regressions. CGP05 and protected-cache coverage exist, but the Rust suite does not prove several required Rust vectors. The sole positive build fixture has one direct registry dependency; it does not prove the transitive registry build in R01, a Git build with the original remote/cache unavailable in R02, transitive contained path packages in R03, two feature-selected builds and identities in R04, target-pruned builds in R05, exact multi-member selection in R06, include macro confinement in R07, or a second canonical clean rebuild and nondeterminism measurement in R09. RF12 custom target, unstable `-Z`, and artifact dependency cases are also absent. RH05 include escape and the full RH08-RH10 planted-output/toolchain-drift shapes are not materialized. Add explicit table-driven or named fixtures whose assertions match every required vector, including zero later starts/publications for negatives.

2. **Exercise forbidden attempts through the verified boundary instead of synthesizing only a returned audit mismatch.** In `internal/rustsource/build_test.go`, `rustBuildProvider.EnforceAndObserve` runs ordinary Cargo successfully and then `mutateAudit` appends `undeclared/child`, `undeclared/read`, `undeclared/write`, or changes network/evidence fields. This proves receipt reconciliation rejects a dishonest or drifted audit, but it does not prove attempted network, child process, undeclared read, or undeclared write is enforced and rejected before publication as required by the rework brief. Use an injected verified fixture provider/boundary that observes or models an actual attempted operation at the enforcement seam, assert the stable diagnostic, and prove no receipt/publication for each case.

3. **Complete stable Rust undeclared-input diagnostics and regression coverage.** `CodeUndeclaredInput` (`rust_undeclared_input`) is defined in `internal/rustsource/types.go` but has no production use or test assertion. The current verified negative loop expects only shared `closure_derivation_drift`. Implement the required deterministic Rust-facing diagnostic mapping for RH05/RH06 or document and test the exact accepted shared-code branch if the architecture explicitly supersedes it; the current accepted Rust contract names `rust_undeclared_input` for these cases.

4. **Make the mandatory pre-C0 regression observe ambient process-start attempts.** `TestBuildToolchainRegistrationStartsNoProcessBeforeC0` checks a verified-provider counter that cannot see a direct `os/exec` launch and supplements it with source-string scans of four files. Production inspection found no current registration subprocess, which is positive, but this is not the required instrumentation that counts every process-start attempt across NewManager, Cargo registration, Rust build-tool registration, and equivalent discovery paths through assurance preflight/C0/permit. Route all production launch seams through an injectable observer or equivalent testable boundary and assert zero until a committed permit crosses that boundary.

## Positive evidence

- Direct `osCargoBuildRunner` execution is gone; fresh metadata and build use committed `closureexec` permits and executor-issued receipts.
- Portable-default evidence remains honest (`network=not-observed`, empty lossless process/read/write claims); verified mode remains fail-closed without a provider.
- Cargo release identity is now bound to an exact approved executable SHA-256, and filesystem-only toolchain/SDK discovery contains no registration subprocess.
- Focused mandatory registration, binding, unit rejection, event validation, offline build/cache, and verified-unavailable tests passed.
- `go test -race ./internal/rustsource ./internal/closureexec -count=1` passed.
- `make lint` reported zero issues; `go vet ./...`, `go build ./...`, `git diff --check`, and `task-board validate` passed.
- Fresh `go test -count=1 ./...` passed; notable timings were `cmd/curator` 406.787s, `internal/rustsource` 108.289s, and `internal/closureexec` 38.558s.

No product code was modified. As a reviewer-archetype run, this verdict supplies no `commit_ack`.