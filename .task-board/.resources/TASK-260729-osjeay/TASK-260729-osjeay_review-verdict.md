# Review verdict — changes requested

Task: TASK-260729-osjeay
Reviewer run: RUN-260729-fffc3c
Verdict branch: changes requested -> analysis

## What is sound

- Current workflow inventory is source-accurate: .github/workflows/ci.yml has ubuntu-latest, macos-latest, and windows-latest running go vet ./... and go test ./..., uses Go from go.mod, retains golangci-lint, and has no race job.
- Current Makefile inventory is accurate: build, test, fmt, vet, lint, and check exist; race/check-ci/conformance do not.
- The rc.5 manifest digest independently matches b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c.
- The load-bearing platform finding is source-verified. Candidate internal/godriver/controls.go maps only darwin and windows; other hosts fail with build_execution_control_unavailable before worker launch. The accepted conformance-claim-v3 qualification vector explicitly marks linux excluded until TASK-260728-1skseh.
- TASK-260720-1pvfj5 still has stale rc.4 text and is still blocked by TASK-260720-jrrgw9, while TASK-260720-2qqq0w is done. The 23-file candidate-versus-accepted delta and the two-file CI/Make ownership boundary are independently reproducible.

## Blocking review findings

1. The proposed Linux matrix changes the target contract without identifying or resolving the full AC conflict. TASK-260720-1pvfj5 scope requires the full supplied candidate suite on Linux/macOS/Windows, and its AC requires both every compiled-build case on ubuntu/macos/windows and go test -race ./... on the selected runner. The submitted map instead prescribes a Linux package allowlist excluding internal/godriver and a scoped race command. That may be the right architecture, but it does not satisfy the current target scope/AC. The map flags only the ordinary Linux go test ./... conflict and does not list the separate stale go test -race ./... AC clause.

2. The candidate matrix is not executable as written. Rows say rc.5 path plus native runner only, while section 7.3 leaves the producer to choose between self-hosted workflow_dispatch and locally attached evidence. No exact runner label, candidate-root transport/materialization step, Windows-visible root, or selected input contract is given. The authoritative root is an uncommitted local curator-spec worktree, so hosted macOS/Windows runners cannot consume it and a Unix absolute path cannot execute on Windows. Implementation would still require a discovery/ownership decision, contrary to this refresh task.

3. The claimed exact YAML/Make plan stops at target names. It names race, check-ci, and conformance but does not define their exact dependencies, inputs, commands, or correspondence to workflow jobs. Option A also requires an explicit maintained Linux package list and a new-package guard, but neither list nor guard command is supplied.

4. Fact-check ledger row 2 is incorrect. The recorded command git status --short --untracked-files=all -- conformance/ currently yields 3 modified and 354 untracked paths, not 3 modified and 18 untracked files. Even collapsed default status yields 91 untracked entries. The digest is correct, but the provenance claim and command/result pair must be corrected.

5. No green test or CI evidence supports the final executable matrix. This audit correctly avoided Go and heavy tests under its scope, so it must describe these as producer gates, not satisfy or imply the reviewer DoD test-green condition.

## Required rework

- Add an explicit target-task contract-drift section covering both full Linux candidate execution and go test -race ./.... Supply the board-owner decision packet and exact proposed wording before treating scoped Linux gates as AC-compatible. This is research/decision rework, so route to analysis rather than to-dev or blocked.
- Select one real candidate delivery mechanism. Specify exact GitHub runner labels and candidate materialization/identity steps if self-hosted, or state an exact local/native evidence flow and remove non-executable hosted candidate-job rows. Include the Windows root mechanism and keep native Linux non-gating until the named prerequisite exists.
- Define the exact .github/workflows/ci.yml job/step changes and exact Makefile target recipes/dependencies, including the complete Linux safe-package inventory and drift guard if that option is selected.
- Correct the candidate-root status count and fact-check command/result. Preserve the verified manifest digest and all pin invariants.
- Reattach the revised task-scoped execution map and return through a new reviewer cycle.

No product, spec, CI, Makefile, pin, or target-task field was modified during review.