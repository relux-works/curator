# TASK-260823-3fnobk results (recorded by the orchestrator after producer timeout)

Producer run RUN-260823-2b7b32 delivered the fix and landed it before timing out at 45m; this artifact records the verifiable outcome.

- Change: internal/buildsource admits case- and Unicode-normalization-distinct encoded paths and rejects only exact encoded-path duplicates, matching the rc.9 candidate case duplicate-build-source-path. Derived from the reviewed patch attached to TASK-260823-1l1p8q (sha256 4d62e862...).
- Landing: PR 26, merge commit c6092af on curator main, branch fix/TASK-260823-3fnobk-buildsource-encoded-path.
- Gates: PR required checks green (Lint, Test x3, Race, Gate self-test x3, Interop conformance gate, Naming gate) — see the PR 26 checks tab; lint and build ran in that pipeline.
- Context: the candidate suite was superseded in parallel by TASK-260823-czs1cx (new identity edd0721, manifest sha256 803918bf...); the rerun task TASK-260823-1l1p8q carries the updated dispatch identity.