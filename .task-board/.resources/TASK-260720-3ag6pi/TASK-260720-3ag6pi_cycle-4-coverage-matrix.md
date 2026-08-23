# TASK-260720-3ag6pi review cycle 4 coverage matrix

## Story acceptance criteria

| STORY-260720-35dck7 criterion | Result | Executable conformance evidence |
| --- | --- | --- |
| New schema version validates the agreed build declarations | PASS | `make validate` executes the schema-6 agent-skill/csk-skill cases, receipt-v1 and marker-v2 cases, and build-driver structural positives/rejections. |
| Go driver semantics and install ordering are normative | PASS | Decision 0004 and manager profile bytes are preserved; `manager-lifecycle.json` executes audit-first, provider-first, lexical-command, deterministic-lock/target, consumer-last, and reverse-rollback cases. |
| Build sources are excluded from agent context | PASS | `build-drivers.json` executes the positive context-exclusion case and the build-root content, marker-embedding, and context-leak rejections. |
| Dry-run and audit-before-build are explicit | PASS | `manager-lifecycle.json` executes compiler-free dry-run, `compiled-cache-miss-is-read-only`, and `all-source-and-trust-gates-before-build`. |
| Compatibility and security impact are recorded | PASS | Cycle-3 manifest/legacy/safety audits remain applicable because the cycle-4 product bytes are preserved exactly; current validation rechecks all linked documents. |
| Valid builds and key rejection cases are covered | PASS | Current `make validate` executes 8 build-driver positives, 77 build-driver rejections, 10 build-source identity cases, 12 toolchain identity cases, and 32 manager lifecycle cases. |
| Validation and deterministic regeneration pass | PASS | Cycle-4 validate, two independent regenerations, recursive byte comparisons, regenerate-check, rc.6 release-check, workflow parity regression, and rc.6 metadata-drift expected-red evidence all passed their stated expectations. |

## Minimum rejection clusters

| Minimum cluster | Result | Passing schema case or vector evidence |
| --- | --- | --- |
| Structural manifest | PASS | Schema-6 invalid cases plus forbidden args/env/output/toolchain/hooks and mixed-shape build-driver rejections under `make validate`. |
| Build-root/source/context paths | PASS | Build-driver missing, unused, overlapping, root, escaped, outside, link, special, non-directory, context-leak, and marker-embedding rejections. |
| Build-source identity | PASS | All 10 `build_source_cases` cover framing, path, metadata, mutation, collision, and root-marker boundaries. |
| Toolchain identity/release boundary | PASS | All 12 `toolchain_cases` plus switch, family, executable, tuning, and digest rejections. |
| Module/dependency/compiler graph | PASS | Module, package, vendor, workspace, cgo/native/SWIG/syso/assembly/embed/generate/PGO build-driver rejections. |
| Process/host isolation | PASS | PATH, GOFLAGS/GOENV/GOWORK, VCS, fake-Go, telemetry, external-link, libgcc, child-tool, and capability-evidence cases. |
| Cache/receipt/protected trust | PASS | Cache identity, receipt/artifact, canonical form, link/special, race, forged receipt, and protected-state cases. |
| Cache-hit/dry-run/marker-context | PASS | Protected hit, compiler-free miss, compiled read-only miss, marker embedding, and context leakage cases. |
| Claim transition/stale suite | PASS | Frozen claim-v1/v2 identities, claim-v3 rc.5 transition, no rc.6 claim, and stale/duplicate rc.6 pin rejections. |
| Private build/cache publication | PASS | Private staging, second-build isolation, protected atomic publication, race, corruption, and trust lifecycle cases. |
| Commit/swap/rollback | PASS | Deterministic lock order, target order/consumer-last, and reverse-under-lock rollback lifecycle cases. |
| Concurrent projects/recovery | PASS | Two-project success, isolated rollback, transaction-ID recovery, and post-build recovery lifecycle cases. |
| Currentness/repair/GC | PASS | Current/failure matrix, rebuild-only repair, locked mark/sweep, and post-commit warning lifecycle cases. |
| Fail-closed inventory/release | PASS | Release negatives cover missing/renamed/lifecycle/fixture/history/pin/claim defects; the new workflow regression requires exact Makefile parity, and rc.6-only drift now makes both workflow commands exit 1. |

## Evidence resources

- `TASK-260720-3ag6pi_cycle-4-validate.log`
- `TASK-260720-3ag6pi_cycle-4-regenerate-1.log`
- `TASK-260720-3ag6pi_cycle-4-regenerate-2.log`
- `TASK-260720-3ag6pi_cycle-4-regeneration-compare.log`
- `TASK-260720-3ag6pi_cycle-4-regenerate-check.log`
- `TASK-260720-3ag6pi_cycle-4-release-check-rc6.log`
- `TASK-260720-3ag6pi_cycle-4-workflow-drift.log`
- `TASK-260720-3ag6pi_review-cycle-3-inventory-legacy-safety.log`
- `TASK-260720-3ag6pi_review-cycle-3-focused-negative-and-lifecycle.log`
