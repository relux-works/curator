# TASK-260729-osjeay — review verdict, cycle 7

**Role:** reviewer  
**Artifact reviewed:** `TASK-260729-osjeay_final-ci-execution-map-rev7.md`  
**Artifact SHA-256:** `d6e2c6a92f8c1a7da62ed0a79ddf0959541e3a0e5296650907ebbd2f838ba1f3`  
**Verdict:** **changes requested → `analysis`**

Revision 7 closes the three cycle-6 executable-contract findings. The independently materialized
55-case no-Go/no-Windows/no-network harness has SHA-256
`c2391ab755af5c0cb4163012eed0f690e7800fcc1228dc1d7fd71f85612e2a41`, printed
`ALL 55 EXPECTATIONS MET`, and exited 0. The execution map still identifies itself as the prior
revision and cycle in its controlling header metadata. That makes the final artifact identity
internally contradictory and requires bounded document rework before acceptance.

## Independently verified

- Cycle-6 F1 is closed: the rev7 T-P1, local Make wrapper, and remote `ssh lev` body use explicit
  environment words or a quoted shell function. Literal zsh and `/bin/sh` cases AR-AS, AW-AX, and
  BB-BC all exited 0; the historical bad forms reproduced their expected failures.
- Cycle-6 F2 is closed: section 7.4 and ledger rows 41/53/58 name the correct harness generations,
  and the current rev7 resource name, digest, and 55-case contract match the board resource.
- Cycle-6 F3 is closed: section 5.3 requires an approved absolute launcher reporting exactly
  `go1.25.5`; all ten `1.25.x` occurrences are rejected-case or meta wording, not prerequisites.
- Local source truth remeasured without Go:
  - `main` and `origin/main` are `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`;
    current `HEAD` is the divergent `c06aa1a15e4093410a686ff0ce4f579fba59dec1`.
  - `origin/main` still uses `checkout@v4`, `setup-go@v5`,
    `golangci-lint-action@v7`, mutable `version: latest`, and no race target.
  - `go.mod` requires Go 1.25.5; the current Makefile remains the six-target baseline.
  - Candidate manifest is
    `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`;
    relative-path whole-tree digest is
    `e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae`
    over 448 files; conformance status is 3 modified plus 354 untracked paths.
  - Accepted-to-candidate product delta is 23 files, all `_test.go`, with zero production drift.
- Current upstream drift was rechecked against primary GitHub sources:
  - checkout v7.0.1:
    https://github.com/actions/checkout/releases/tag/v7.0.1
  - setup-go v7.0.0:
    https://github.com/actions/setup-go/releases/tag/v7.0.0
  - golangci-lint-action v9.3.0:
    https://github.com/golangci/golangci-lint-action/releases/tag/v9.3.0
  - golangci-lint v2.12.2:
    https://github.com/golangci/golangci-lint/releases/tag/v2.12.2
  - current runner labels:
    https://github.com/actions/runner-images/blob/main/README.md
    confirms Ubuntu 24.04 x64, macOS 26 arm64, and Windows Server 2025 x64.

No Go command, test, build, vet, format, lint, install, dependency fetch, product/spec/CI/Make/pin
edit, or `TASK-260720-1pvfj5` mutation was performed. The reviewer made read-only network requests
only to the official upstream pages listed above.

## Finding

### F1 — the rev7 outcome still self-identifies as revision 6 / cycle 5

The board resource name, resource description, digest, section 1.2e, section 1.3, gate-status
artifact, and harness all identify the current deliverable as revision 7 / rework cycle 6.
The controlling metadata at the top of the same outcome says otherwise:

- line 1: `Final Curator compiled-build CI execution map — revision 6`
- line 3: `read-only audit, rework cycle 5`
- line 5: `Supersedes: revision 5`, referring readers to section 1.2d
- line 25: `Cycle 5 (this revision) made no network read`

Those are not harmless historical rows: they define the identity and provenance of the artifact a
producer or reviewer opens first. They contradict section 1.2e, which correctly says revision 7
fixes revision 6, and section 1.3, which correctly says cycle 6 is the current revision. The outcome
therefore fails the independently-reviewable artifact requirement even though its executable
contracts now pass.

## Required bounded rework

1. Update the title and task metadata to revision 7 / rework cycle 6.
2. State that revision 7 supersedes revision 6 and that the cycle-6 findings are corrected in
   section 1.2e. Preserve older correction history as historical content.
3. Recast line 25 as a historical cycle-5/revision-6 statement or remove the stale
   `this revision` qualifier. Keep the existing cycle-6 no-network statement in section 1.3.
4. Recompute the document SHA-256, update the existing rev7 outcome resource and gate-status digest,
   and rerun the 55-case harness once from the board resource. Do not create a misleading second
   rev7 map alongside the stale one.
5. Preserve every substantive contract and identity already verified above. No product, CI,
   Makefile, pin, target-task, or Go execution is needed for this correction.

This is recoverable research/document rework, not a stop-the-line boundary, so the explicit verdict
branch is `analysis`.
