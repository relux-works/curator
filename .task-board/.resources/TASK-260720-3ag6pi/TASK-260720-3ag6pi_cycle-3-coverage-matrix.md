# TASK-260720-3ag6pi — review cycle 3 coverage matrix

## Verified candidate and version resolution

- Product source: `.temp/TASK-260720-3ag6pi/rework-cycle-3`.
- Authoritative published baseline: remote `main` and annotated
  `v1.0.0-rc.5` both peel to
  `f5d7673039226ab81de2f4f87e2155ae995c4df3`.
- Verification checkout: unsigned, untagged, task-local commit
  `ddb181ca3b8e243f212e90ff26fcabe2234fb669`, parent
  `f5d7673039226ab81de2f4f87e2155ae995c4df3`, tree
  `fa0eb87cacf6cd9ec8510f7d26a91ea76b6d74fd`.
- A recursive comparison between the product source and verification checkout
  exits 0. The local commit exists only to satisfy the release gate's clean
  checkout invariant; it is not a tag, signature, pushed ref, implementation
  pin, claim, or publication record.
- Review cycle 2 and the current operator directive supersede the original
  literal rc.4 wording. The accepted target is rc.6 on the frozen published
  rc.5 base. The literal rc.4 release command is retained as expected-red
  evidence and exits 2 because README correctly identifies rc.6.

## Mechanical task acceptance

| Criterion | Result | Executable evidence |
| --- | --- | --- |
| `make validate` passes without skips | PASS | Exit 0. Validates 42 schemas and 447 manifest entries; 59 Python tests and Go tests pass. The output contains no skip. See `TASK-260720-3ag6pi_cycle-3-reverify-validate.log`. |
| First clean regeneration leaves no diff | PASS | Fresh clone 1: `make regenerate` exit 0, generated-path diff exit 0, porcelain status empty. |
| Second independent regeneration is byte-identical | PASS | Fresh clone 2: `make regenerate` exit 0, generated-path diff exit 0, porcelain status empty. Recursive `conformance/v1` comparison and digest comparison exit 0. Both trees have aggregate SHA-256 `8255898b37dd1f3b95423804bd0c35bd7ec48a16fbbe9b4d9e4cecc830900072`. |
| `make regenerate-check` passes | PASS | Exit 0 in the clean verification checkout. |
| Current release gate passes | PASS | `make release-check VERSION=1.0.0-rc.6` exits 0 at the exact clean local candidate SHA. `/usr/bin/git` is used; no alternate index, wrapper, Git directory, or Git worktree override is present. |
| Original literal rc.4 wording is not fabricated | PASS after reviewed supersession | `make release-check VERSION=1.0.0-rc.4` truthfully exits 2: `README version is not 1.0.0-rc.4`. This is expected-red evidence, not a passing rc.4 claim. |
| Manifest inventory and hashes are exact | PASS | 447 sorted unique entries equal every non-manifest file on disk and every SHA-256 matches. Manifest SHA-256 is `72c5d717027ca096b14bc32f5d60bb740676974e9429f3d09b730897e5fba89b`; both rc.6 pins match it. |
| Every schema case is indexed | PASS | The 376 index entries exactly equal the 376 case files. Required groups are agent-skill-v6 24, csk-skill-v6 24, build-receipt-v1 18, install-marker-v2 14, claim-v2 7, and claim-v3 13. Every path, validity expectation, and manifest hash is listed in the inventory audit. |
| Fixture, expected, build-driver, and lifecycle vectors are inventoried | PASS | All 13 Go build fixture files, all 11 expected build-driver files, `build-drivers.json`, and `manager-lifecycle.json` are present and hash-correct. |
| Restored lifecycle semantics match accepted evidence | PASS | All 22 compiled lifecycle cases are present. After normalizing only the intentional execution-policy/cache-key/receipt identity revision, the current vector is byte-identical to accepted `TASK-260720-cw39jh`, normalized SHA-256 `11fe66182e719e8e5c067d3ff2ab646dd889de7c97130881fcacdbf9694a7b6c`. |
| Frozen legacy semantics remain compatible | PASS | Twelve legacy schemas and 24 baseline cases are byte-identical to both rc.5 `origin/main` and pre-v6 baseline `57c1f568…`. Their index and manifest entries are preserved. Marker v1 stays schema 1; claim v1 stays schema 1 / protocol rc.3. |
| Failure evidence executes no package artifact | PASS | All 77 build-driver rejection cases require `result=reject`, `artifact_executed=false`, and `reuse=false`. Compiled dry-run forbids Go `list`/`build` and records `artifact_executed=false`. Generator and validator contain no process-execution API; release-gate subprocess use is Git-only. |
| Release evidence is not fabricated | PASS | Published rc.5 metadata is byte-identical to `origin/main` at SHA-256 `75ae17fc029b4f51ca40ce768d04fd72991ec3db2602b8fe59213bee6ac34583`. rc.6 emits no claim, advances no committed downstream pin, claims no hardened profile, and the verification commit is unsigned, untagged, local, and unpushed. |

## STORY-260720-35dck7 acceptance criteria

| Story acceptance criterion | Result | Passing executable evidence |
| --- | --- | --- |
| A new schema version validates the agreed build declarations. | PASS | Agent-skill-v6 and csk-skill-v6 each have 24 indexed cases. Build-receipt-v1 has 18 and install-marker-v2 has 14. `make validate` executes every case; build-driver positives include `schema-6-mixed-script-and-build-commands`, while structural rejections cover forbidden args/env/output/toolchain/hooks and mixed command shapes. |
| Go driver semantics and install ordering are normative. | PASS | Decision 0004 and manager profile remain normative. Executable vectors enforce `all-source-and-trust-gates-before-build`, `provider-first-and-lexical-command-order`, `deterministic-lock-order`, and `deterministic-target-order-and-consumer-last`. Go tests pin portable identity reuse and all lifecycle ordering groups. |
| Build sources are excluded from agent context. | PASS | `build-root-excluded-from-agent-context` is positive evidence; `build-root-content-in-context` and `marker-embed-build-source-regression` reject leakage. The fixture context file/hash outputs are present and manifest-pinned. |
| Dry-run and audit-before-build are explicit. | PASS | `compiler-free-dry-run-miss`, `compiled-cache-miss-is-read-only`, and `all-source-and-trust-gates-before-build` enforce no compiler/cache/persistent mutation before accepted audit and trust gates. |
| Compatibility and security impact are recorded. | PASS | README, CHANGELOG, COMPATIBILITY, RELEASE, conformance docs, and external-repository docs identify rc.6 and preserve rc.5 history. The legacy audit proves schemas 1–5, marker v1, and claim v1 bytes/semantics unchanged. |
| Vectors cover valid builds and all key rejection cases. | PASS | `build-drivers.json` contains 8 positives, 77 rejections, 10 build-source cases, and 12 toolchain cases. `manager-lifecycle.json` contains 32 named cases, including all 22 accepted compiled transaction/concurrency/recovery/currentness cases. |
| Specification validation and deterministic regeneration pass. | PASS | `make validate`, two fresh `make regenerate` runs, recursive byte comparison, and `make regenerate-check` all exit 0. |

## Minimum rejection clusters

| Minimum rejection cluster | Result | Passing schema case or vector |
| --- | --- | --- |
| Structural manifest | PASS | V6 invalid schema cases plus `schema-5-build-command`, `unknown-driver`, `forbidden-args`, `forbidden-env`, `forbidden-output`, `forbidden-toolchain`, `forbidden-hooks`, and `mixed-script-build-shape`. |
| Build-root, source, and context paths | PASS | `missing-build-roots`, missing/unused/overlapping/root/symlink/special build roots, root/escaped/outside/link/special/non-directory sources, `build-root-content-in-context`, and `marker-embed-build-source-regression`. |
| Build-source identity algorithm | PASS | All 10 named `build_source_cases`: exact fixture identity, framing/order/empty/binary/root marker, non-input metadata, invalid Unicode, duplicate paths, link/special entries, mutation, NUL collision, and root-marker input. |
| Toolchain identity and release boundary | PASS | All 12 named `toolchain_cases`, plus toolchain switch, unsupported family, wrong executable, and digest mismatch rejections. |
| Module, dependency, and compiler-input graph | PASS | Missing/nested modules; package/vendor/workspace graph failures; cgo/native/SWIG/syso/assembly/embed failures; `cgo-import-dynamic`, `attempted-go-generate`, and `default-pgo`. |
| Process and host isolation | PASS | Poisoned PATH, inherited GOFLAGS/GOENV/GOWORK, VCS metadata, fake Go, telemetry failure/escape, external link, libgcc fallback, and child tool escape rejections. Fixed-driver vectors contain no shell or artifact execution. |
| Cache, receipt, and protected-state trust | PASS | Cache key/target/toolchain/policy/source mismatches; receipt/artifact/path/size mismatches; noncanonical receipt; partial/link/special entries; concurrent mismatch; forged receipt; and protected publication/trust lifecycle cases. |
| Cache-hit, dry-run, and marker/context regression | PASS | `protected-cache-hit`, `compiler-free-dry-run-miss`, `compiled-cache-miss-is-read-only`, `marker-embed-build-source-regression`, and `build-root-content-in-context`. |
| Claim transition and stale-suite evidence | PASS | Frozen claim-v1 rc.3, frozen claim-v2 rc.4, claim-v3 rc.5, and no rc.6 claim. Release tests reject redefined claim history, claim-v3 transition mismatch, stale/duplicate rc.6 suite pins, and fabricated rc.6 claims. |
| Private build failure and cache publication | PASS | `all-misses-stage-and-verify-before-home-lock`, `second-build-failure-preserves-persistent-state`, `publish-complete-immutable-entry-under-home-lock`, `concurrent-identical-winner`, `concurrent-determinism-mismatch`, `corrupt-live-entry`, and `untrusted-cache-boundary`. |
| Commit, target swap, and reverse rollback | PASS | `deterministic-lock-order`, `deterministic-target-order-and-consumer-last`, and `reverse-rollback-under-home-lock`. |
| Concurrent projects and recovery | PASS | `two-project-success-preserves-both-consumers`, `successful-project-survives-other-project-rollback`, `interrupted-global-journal-recovered-by-transaction-id`, and `install-recovery-runs-after-private-builds`. |
| Currentness, repair, and GC | PASS | `compiled-installation-current`, `compiled-currentness-failure-matrix`, `repair-rebuilds-invalid-compiled-entry`, `locked-mark-and-sweep-compiled-cache`, and `post-commit-gc-failure-is-maintenance-warning`. |
| Fail-closed inventory and release evidence | PASS | Validator tests require all 22 cases, every lifecycle group, schema version, and current portable identity. Release tests reject each missing required artifact, renamed v6 schema, missing manifest entry, removed lifecycle case even after hash refresh, stale compiled fixture, changed rc.5 metadata, and stale/duplicate rc.6 pins. |

## Test implementation coverage

The rework adds executable regression tests rather than relying on the matrix
alone:

- Go: frozen rc.5 metadata, honest rc.6 metadata, portable lifecycle identity,
  planning/order/dry-run, publication/cross-project isolation,
  transaction/recovery/status/repair/GC, and deterministic generation.
- Python validator: exact 22-case inventory, every lifecycle group
  fail-closed, compiled dry-run, lifecycle schema/portable identity, rc.6 pin,
  and frozen rc.5 metadata.
- Python release gate: complete rc.6 acceptance; every missing artifact;
  renamed schema; missing manifest entry; removed lifecycle case with refreshed
  hash; stale fixture with refreshed hash; frozen claim history; claim
  transition; frozen rc.5 metadata; no fabricated rc.6 claim; and stale or
  duplicate suite pins.
