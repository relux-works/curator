# Promote curator-spec implementation pins after manager releases

## Description
Advance curator-spec implementation conformance pins only to the qualified released Curator and csk revisions, then make the shared black-box parity gate part of the specification CI matrix.

## Scope
Work in curator-spec after TASK-260720-3pvihp. This step precedes and must not wait for the manager-side released-suite pin changes audited by TASK-260720-38l1sy and TASK-260720-1utsx8. Own only the Curator and csk refs and their compiled-build invocations in the implementation-conformance workflow plus the cross-manager runner step. Preserve the registry-service pin unless its own released contract requires a change. Do not use branches, mutable tags, candidate worktree commits, manager suite-pin changes, release claims, or implementation-owned expected values.

## Acceptance Criteria
The workflow pins Curator and csk by the exact full commits recorded in release qualification and each commit is traceable to a published manager release. On Linux, macOS, and Windows it supplies the exact in-tree candidate suite digest qualified by TASK-260720-3pvihp to both consumers, installs the supported Go and Python toolchains, rejects any skip or xfail, and runs the black-box runner against isolated manager state. The gate fails for missing case IDs, result-contract violations, suite-digest mismatch, normalized outcome divergence, launch stdout or stderr or exit mismatch, or changed negative-case rejection boundaries. Existing registry-service coverage remains intact, action pins remain immutable, no protocol-release claim is emitted, no manager suite pin changes here, and no implementation pin is advanced before its recorded release evidence.
