# TASK-260720-3ag6pi review cycle 2 coverage matrix

## Reviewed candidate

- Authorized rc.5 supersession: `TASK-260729-1kq1rd` recommended the exact
  `TASK-260728-2kp3tv` candidate; the newer human release order recorded on
  `TASK-260730-1fsbqd` authorized its landing and publication.
- Remote `main` and annotated tag `v1.0.0-rc.5` resolve to
  `f5d7673039226ab81de2f4f87e2155ae995c4df3`.
- The landed tree is `78210085727ec33b79a050a807f51da253ffb0c8`, exactly
  the independently accepted candidate tree.
- Baseline for frozen legacy comparison:
  `57c1f56846d221ecc55786bd3c2467ec32f11730`, which is also the landed
  commit's sole parent.

## Mechanical task acceptance criteria

| Criterion | Result | Evidence |
|---|---|---|
| `make validate` passes without skips | PASS | 42 schemas, 447 manifest entries, 41 Python tests, and Go tests pass. No Python skip is reported. See `TASK-260720-3ag6pi_review-cycle-2-validate.log`. |
| First regeneration and regenerate-check leave no diff | PASS | `make regenerate` and `make regenerate-check` both exit 0; Git status remains empty. |
| Second independent regeneration is byte-identical | PASS | Two independent clean clones produce 448 conformance files with aggregate SHA-256 `e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae`; recursive comparison exits 0. |
| Release gate passes on the authorized landed candidate | PASS after approved supersession | Unwrapped `make release-check VERSION=1.0.0-rc.5` exits 0 at exact `HEAD` `f5d7673`; `/usr/bin/git` is used with no alternate index, worktree, or Git directory. Literal rc.4 is correctly rejected as the superseded version. |
| Manifest inventory and hashes are exact | PASS for the inventory actually generated | All 447 listed files equal the on-disk inventory and all hashes match. The manifest SHA-256 is `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`; both rc.5 release pins match it. |
| Every schema case is indexed | PASS | The 376 schema-case index entries exactly equal all case files. Required v6 groups are present: agent-skill-v6 24, csk-skill-v6 24, build-receipt-v1 18, install-marker-v2 14, and conformance-claim-v2 7. |
| Fixture, expected build-driver, and vector files are inventoried | PASS | 13 `fixtures/go-build-skill` files, 11 `expected/build-driver` files, `build-drivers.json`, and `manager-lifecycle.json` are present and correctly hashed. Content coverage of `manager-lifecycle.json` fails separately below. |
| Frozen legacy semantics remain compatible | PASS | Twelve legacy schemas and 24 baseline cases are byte-identical. The 24 baseline index/manifest entries are preserved. Seventy added schema-7 guard cases for manifest schemas 1–5 are all indexed invalid. Marker v1 is unchanged; claim v1 remains schema 1 / protocol rc.3 with all three frozen SHA-256 values unchanged. |
| Failure evidence executes no package artifact | PASS | All 77 rejection cases require `artifact_executed=false` and `reuse=false`; the fixed driver uses no shell and executes no output; cache-hit and dry-run execute no source-aware Go command. Generator and validator contain no process-execution API. |
| Release evidence is not fabricated | PASS | Clean committed checkout, immutable commit/tree identities, authoritative remote refs, no Git wrapper, and no alternate index. |
| Every story criterion and minimum rejection cluster has executable evidence | **FAIL** | The landed rc.5 suite dropped 22 accepted schema-6 compiled lifecycle cases and the fail-closed guards that required them. |

## STORY-260720-35dck7 acceptance criteria

| Story acceptance criterion | Result | Landed evidence or gap |
|---|---|---|
| A new schema version validates the agreed build declarations. | PASS | Frozen v6 schemas are byte-identical to the accepted rc.4 composite. Required v6 schema cases pass. |
| Go driver semantics and install ordering are normative. | **FAIL — executable ordering evidence missing** | Normative prose and all accepted build-driver names remain, but `all-source-and-trust-gates-before-build`, `provider-first-and-lexical-command-order`, `deterministic-lock-order`, and `deterministic-target-order-and-consumer-last` are absent from all landed conformance files. |
| Build sources are excluded from agent context. | PASS | `build-root-excluded-from-agent-context`, `build-root-content-in-context`, and `marker-embed-build-source-regression` remain enforced; fixture context hashes and exclusions validate. |
| Dry-run and audit-before-build are explicit. | **FAIL — executable lifecycle evidence missing** | Normative prose remains and build-driver dry-run/context cases pass, but `compiled-cache-miss-is-read-only` and `all-source-and-trust-gates-before-build` are absent. |
| Compatibility and security impact are recorded. | PASS | Compatibility/security prose is present; schemas 1–5, marker v1, and claim v1 are frozen as described above. |
| Vectors cover valid builds and all key rejection cases. | **FAIL** | Build-driver coverage is preserved and extended from 7/75/10/12 to 8/77/10/12, but all 22 accepted compiled transaction/concurrency/recovery lifecycle cases are missing. |
| Specification validation and deterministic regeneration pass. | PASS mechanically, insufficient for acceptance | All supported commands pass and regenerate deterministically, but the current generator and validator deterministically reproduce and accept the incomplete lifecycle inventory. |

## Minimum rejection clusters

| Rejection cluster | Result | Passing evidence or exact gap |
|---|---|---|
| Structural manifest | PASS | V6 schema invalid cases and build-driver `schema-5-build-command`, `unknown-driver`, forbidden field, and mixed-shape cases remain. |
| Build-root, source, and context paths | PASS | Missing/overlapping/root/symlink/special-file build roots, escaped source, nested module, non-main package, and context-leak cases remain. |
| Build-source identity algorithm | PASS | Ten build-source cases remain with no accepted-name loss, including framing, duplicate/path/link/mutation, root-marker, NUL collision, and preimage anchors. |
| Toolchain identity and release boundary | PASS | Twelve toolchain cases remain with no accepted-name loss, including version framing, path/link/tree mutation, wrong executable/digest, and unsupported family cases. |
| Module, dependency, and compiler-input graph | PASS | Vendoring, workspace, toolchain-switch, cgo/native object/assembly/embed/generate/PGO rejection names remain. |
| Process and host isolation | PASS | PATH/GOENV/GOWORK/VCS/telemetry/external-link/libgcc/tool escape cases remain; fixed driver records no shell or artifact execution. |
| Cache, receipt, and protected-state trust | PASS | Cache/receipt/artifact mismatch, noncanonical receipt, partial/link/special entry, forged receipt, and protected-boundary cases remain. |
| Cache-hit, dry-run, and marker/context regression | **FAIL** | Build-driver `protected-cache-hit`, `compiler-free-dry-run-miss`, and context/marker cases remain, but lifecycle case `compiled-cache-miss-is-read-only` is missing. |
| Claim transition and stale-suite evidence | **FAIL** | Claim v1 bytes and claim-v2 schema cases remain, but the accepted release-gate tests for redefined claim v1, stale rc.3 suite identity, and duplicate rc.4 claim identity were removed; the current release gate no longer contains the frozen-claim guard. |
| Private build failure and cache publication | **FAIL** | Missing: `all-misses-stage-and-verify-before-home-lock`, `second-build-failure-preserves-persistent-state`, `publish-complete-immutable-entry-under-home-lock`, `concurrent-identical-winner`, `concurrent-determinism-mismatch`, `corrupt-live-entry`, and `untrusted-cache-boundary`. |
| Commit, target swap, and reverse rollback | **FAIL** | Missing: `deterministic-lock-order`, `deterministic-target-order-and-consumer-last`, and `reverse-rollback-under-home-lock`. |
| Concurrent projects and recovery | **FAIL** | Missing: `two-project-success-preserves-both-consumers`, `successful-project-survives-other-project-rollback`, `interrupted-global-journal-recovered-by-transaction-id`, and `install-recovery-runs-after-private-builds`. |
| Currentness, repair, and GC | **FAIL** | Missing: `compiled-installation-current`, `compiled-currentness-failure-matrix`, `repair-rebuilds-invalid-compiled-entry`, `locked-mark-and-sweep-compiled-cache`, and `post-commit-gc-failure-is-maintenance-warning`. |
| Fail-closed inventory/release evidence | **FAIL** | Landed `manager-lifecycle.json` is byte-identical to the rc.3 baseline, yet current validation and rc.5 release gates pass. The accepted lifecycle name guards, compiled fixture identity guard, v6 required-artifact release guard, and frozen claim-v1 release guard are absent. |

## Lifecycle regression identity

The independently accepted `TASK-260720-cw39jh` lifecycle vector has SHA-256
`676e617a0e0a6d575310f38e1de740eab583d709e2351be9eaa818c9882d78d4`
and 32 named cases. The landed rc.5 vector has SHA-256
`2ddbd2665a63f44dc0e03e060f4cd34bfde219a56b3192511fe1ef81047feedf`,
exactly equal to the rc.3 baseline, and only 10 named legacy cases. It also
lacks `schema_version`, `compiled_build_fixture`, and ten accepted top-level
compiled lifecycle groups.

The manifest correctly hashes this incomplete file. Therefore manifest
integrity, deterministic regeneration, and a green release gate do not satisfy
the integrated protocol-v6 coverage contract.

