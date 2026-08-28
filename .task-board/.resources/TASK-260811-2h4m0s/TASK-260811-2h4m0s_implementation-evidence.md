# TASK-260811-2h4m0s implementation evidence

Date: 2026-08-19

## Delivered behavior

- Added the sealed `rustsource.Manager` and raw-data-only capture API. The manager owns the Cargo/oracle selectors, protected intake, one causal `closureexec.Executor`, private session roots, vendor and metadata execution, receipts, and lifecycle.
- Added closed Cargo 1.91.0 selection through a Curator-owned startup registration, complete external tool-root fingerprinting, exact executable hashing, and immediate full-root/executable rechecks. Capture requests expose no tool, runner, executor, provider, permit, receipt, config, destination, projection, or normalized-manifest authority.
- Added closed Cargo manifest/lock parsing, selection-neutral lock-superset capture, registry index/archive identity binding, workspace-contained path origins, exact Git lock/source matching, and target/feature selection binding through the retained active-graph implementation.
- Added complete pre-Cargo artifact admission. Registry transform parsing now follows archive admission; Git source bytes admit before Git inspection; rejected compiled/dirty/missing/mismatched origins leave oracle/vendor/metadata starts at zero and do not create Cargo home, vendor, or target state.
- Added a manager-owned Go implementation of pinned Cargo 0.92 Git repository/projection behavior, including loose and packed objects, index checksum/tree reconstruction, tracked/untracked validation, filter rejection, recursive contained submodules, include/no-include modes, workspace inheritance, normalized manifests, and sealed derivation receipts.
- Git worktree plus `.git` index/object/submodule administration is copied once into protected intake. All later inspection, oracle inputs, Cargo Git-cache staging, and vendor expectations derive only from the protected snapshot; raw repository paths are no longer consulted after admission.
- Added physical `cargo vendor --locked --offline --versioned-dirs` execution into an absent destination through the manager executor. Exact per-leaf registry/Git dispositions, normalized manifests, canonical checksum bytes, unique lock-directory mapping, containment, artifact readmission, and post-vendor comparison are enforced.
- Added two ordered physical Cargo metadata derivations (unfiltered lock-superset and active selection) through the same executor and causal head, with typed stdout evidence and issued receipt verification.
- Preserved immutable admitted replay mounts. Cargo receives a separately declared writable operation work copy seeded by the trusted runner from an immutable admitted Cargo-home tree. Original replay bytes and read-only permissions are rechecked after execution; no mutable admitted-mount semantic remains.

## Tests added or extended

- External-package production positives for registry and packed Git repositories, recursive submodules, physical vendor output, and both metadata receipts.
- Exported API reflection audit and zero/foreign manager/capture rejection.
- Pre-admission compiled Git zero-start and absent-state regression.
- Protected Git snapshot mutation-after-admission zero-oracle regression.
- Portable runner regression proving an executable cannot mutate an admitted replay.
- Retained parser, transform, origin, containment, tamper, graph, selection, permit, toolchain-drift, receipt, and lifecycle fixtures.

## Validation evidence

- `go test -count=1 ./internal/rustsource ./internal/closureexec`: exit 0 (`rustsource` 17.484s, `closureexec` 4.012s) after the last code changes.
- `go test -count=1 ./internal/artifactpolicy`: exit 0 (`artifactpolicy` 32.514s).
- `go test -race -count=1 ./internal/rustsource ./internal/closureexec`: exit 0 (`rustsource` 27.716s, `closureexec` 6.715s).
- `go test -count=1 -timeout 30m ./...`: exit 0. Notable packages: `cmd/curator` 355.923s, `artifactpolicy` 126.542s, `install/atomicity` 110.364s, `install` 109.428s, `godriver` 90.272s, `rustsource` 34.952s.
- `go vet ./...`: exit 0 after the last code changes.
- `go build ./...`: exit 0 after the last code changes.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run ./internal/rustsource/... ./internal/closureexec/... ./internal/artifactpolicy/...`: exit 0, `0 issues.` The first attempt to invoke an uninstalled local `golangci-lint` binary exited 127; CI's exact pinned version was then run directly. Intermediate pinned-lint findings exited 1 and were corrected before the green run.
- `make no-broad-suppression`: exit 0, `no-broad-suppression: ok`.
- `git diff --check`: exit 0 after the last code changes.

## Review focus

Review the generic `closureexec.WorkCopy` contract alongside the Cargo calls: the seed receipt remains an immutable `InputMount`; the work copy is a separate, declared writable root and is never accepted as origin evidence. Also verify that Git inspection and Cargo Git staging consume only the protected combined repository snapshot and that `GitObjectReceipts` and oracle permits bind that snapshot.
