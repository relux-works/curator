# TASK-260823-czs1cx landing evidence (orchestrator, post-timeout)

Producer run RUN-260823-feedc0 delivered both halves and timed out before final bookkeeping.

- Verdict on the two Windows failures: candidate VECTOR bugs, not implementation bugs — fixed on the candidate side. New immutable candidate: revision edd0721, protocol 1.0.0-rc.9, manifest sha256 803918bf..., tree 9d5a10b6..., 692 files (identity file attached earlier as TASK-260823-czs1cx_candidate-suite-identity.txt; supersedes 859727b/782d686, which stays unrewritten).
- Curator-side companion: PR 29 (Preserve Windows candidate vector bytes, internal/godriver/builddriver_positive_conformance_test.go) merged as 351db49 with EVERY lane green including Test (windows-latest) — verified pre-merge.
- The superseding identity is recorded for the rerun task TASK-260823-1l1p8q and must be echoed on TASK-260822-c0rxj7.