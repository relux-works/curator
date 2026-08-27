# TASK-260729-osjeay review verdict — cycle 2

**Verdict:** changes requested  
**Route:** `analysis` (the rejected deliverable is a research/execution-map artifact, not implementation)

Revision 2 closes the previous review's rc.4 wording, dependency, package-inventory, and
conformance-root count findings. The source facts below independently re-verified:

- `TASK-260720-jrrgw9` candidate and `TASK-260729-2kaopg` accepted comparison are both based on
  `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`.
- The conformance worktree is based on `57c1f56846d221ecc55786bd3c2467ec32f11730`,
  with 3 modified and 354 untracked paths under `conformance/`.
- `manifest.json` hashes to
  `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`.
- The documented whole-tree algorithm currently yields
  `e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae`
  over 448 files.
- The candidate contains 40 Go package directories, the named godriver rejection test exists at
  `internal/godriver/worker_test.go`, and the six hard-failing committed-pin reads are present.

No Go command, build, test, lint, network fetch, install, product edit, task-target edit, or pin
mutation was performed during this review.

## Blocking findings

### F1 — the selected archive transport command is not executable

Section 5.2 defines:

```text
DST=.../.temp/TASK-260720-1pvfj5/candidate/conformance/v1
```

but C3 archives with:

```bash
tar ... -C "$(dirname "$DST")" conformance
```

That resolves the archive operand to
`.../candidate/conformance/conformance`, which does not exist. The read-only reviewer check returned
`archive_C=.../candidate/conformance` and `archive_operand_exists=no`. This breaks the one selected
candidate delivery mechanism before transport begins.

Required correction: archive from the directory above `conformance`, for example
`-C "$(dirname "$(dirname "$DST")")" conformance`; add a fail-closed preflight and an archive
listing assertion that `conformance/v1/manifest.json` exists before upload. Keep the extraction root
and Windows-visible path consistent with that archive shape.

### F2 — immutable candidate identity is internally contradictory and under-enforced

Section 5.2 says that if the frozen manifest/tree digests differ, the producer should adopt the new
values as the candidate identity. That contradicts:

- the task's immutable rc.5 input, whose manifest digest is fixed at `b6f56aac...04c`; and
- invariant I2, which says a digest mismatch fails the run.

A moved live worktree is a different candidate and requires a refreshed accepted input/decision; it
must not be silently adopted by the CI producer.

The executable recipes also fail to implement I2:

- `candidate-digest` checks only `manifest.json`; it does not check or emit the whole-tree digest or
  the 448-file inventory.
- The Windows C2 command checks only `manifest.json`, despite C2 claiming two independent identity
  checks on every host before every candidate run.

Required correction: make any mismatch against the fixed accepted manifest and tree identities
abort. Add exact `CANDIDATE_TREE_SHA256` verification to the Make target and an equivalent
PowerShell whole-tree verification for Windows; emit both identities before every candidate gate.
If archive SHA-256 is part of the transport identity, define its expected-value handoff and verify it
before extraction, but do not substitute it for post-extraction tree verification.

### F3 — the Make/CI parity claim is false

The proposed `check-ci` recipe is labeled an exact mirror of the CI gate, but it:

- does not materialize or accept the committed conformance root;
- runs `go test ./...` without `CURATOR_CONFORMANCE_ROOT`, causing conformance tests to skip;
- adds `linux-package-guard`, which the macOS/Windows `test` job does not run.

Likewise, the local `make race-full` and `make race` commands in section 9.2 omit the committed
conformance root that the proposed CI race job exports. They therefore do not validate the same
suite as CI.

Required correction: either make the targets accept and require the exact pin/candidate root and
match the corresponding CI lane, or rename/document them as intentionally different local gates.
Update the target-to-job table so every claimed correspondence is executable and truthful.

### F4 — Windows transport is still an option set, not one exact mechanism

C3 selects tar plus `scp`, but C4 says that transport may be unavailable and then names a chunked
base64 fallback without executable commands. This leaves the producer choosing a mechanism again,
contrary to the rework requirement for one exact transport/materialization path. The map also lacks
the Windows whole-tree verification and exact candidate test command.

Required correction: select and specify one Windows path end to end, including preflight,
upload/materialization, archive digest, extraction, fixed native root, manifest/tree verification,
the exact `go test` command, exit capture, and cleanup. If the required transport or approved Go
root is absent, fail closed and record that named prerequisite rather than switching to an
unspecified fallback.

## Re-review gate

Revise only `TASK-260729-osjeay_final-ci-execution-map.md`. Preserve the verified rc.5 manifest
digest, current main pin, 3/354 status count, Linux non-gating prerequisite, and future-gate
evidence honesty. Do not edit product, CI, Makefile, target-task fields, spec, or pins. A new review
cycle can accept once F1-F4 are corrected and the selected mechanism plus recipes enforce the same
invariants the document states.
