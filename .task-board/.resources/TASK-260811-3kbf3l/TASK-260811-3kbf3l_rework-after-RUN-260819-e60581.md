# Mandatory Rust rework after RUN-260819-e60581

Preserve every currently green behavior and the accepted architecture: portable mode remains functional by default with honest `network=not-observed` receipts; verified mode remains provider-backed and fail-closed; exact Cargo bytes remain approved at C0; compiled binary vendoring remains forbidden.

Implement both bounded reviewer requirements before handoff:

1. Complete R07 with a real in-closure `include!` source fragment in addition to the existing `include_str!` and `include_bytes!` leaves. Run the fresh offline build and assert the `include!` leaf is present in the captured closure/receipt input, not merely that compilation succeeds.
2. Replace the direct in-memory RH10 `tools.recheck` unit probe with an end-to-end Build/Executor time-of-use drift rejection. Inject drift after accepted C4/C0 identity but before executor process start for at least one physical Rust toolchain component. Assert stable `artifact_toolchain_identity_changed`, zero process starts for the affected action, no executor/publication receipt, and no protected output.

Run focused R07/RH10 tests, Rust/closureexec package tests, race tests, `make lint`, `go vet ./...`, `go build ./...`, fresh `go test -count=1 ./...`, `git diff --check`, and `task-board validate`. Update outcome evidence with exact test-to-requirement mapping. Do not commit, stage, reset, clean, or overwrite unrelated user changes.
