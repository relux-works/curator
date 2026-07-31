# Curator Go → CocoaSkills Go parity delta (revision 2)

**Board task:** `TASK-260729-1t1z2l`

**Research date:** 2026-07-29

**Revision:** 2 — rework of the cycle-1 artifact after `TASK-260729-1t1z2l_review-verdict-cycle-1.md` returned CHANGES REQUESTED.

**Scope:** read-only reconnaissance across Curator, CocoaSkills/csk, the task board, the schema-6/`go-v1` wire contract, and the accepted rc.5 portable-execution amendment. No product, test, configuration, staging, commit, pin, or publication changes were made.

## 0. What changed in revision 2

The cycle-1 verdict accepted the 17-task coverage, the Curator package/test counterparts, the live dependency DAG, the repository provenance, and the no-publication/no-pin boundary. Those parts are retained and re-verified here. Five things are corrected, and two further defects found during this rework are added.

| # | Correction | Source |
| --- | --- | --- |
| 1 | rc.5 changes the local canonical identity. The claim that metadata/receipt/marker bytes are "frozen from rc.4" was false at the behavioral surface. §3 now separates frozen schema-6 *declaration* bytes from the rc.5 execution-policy-bound *canonical identity*. | verdict correction 1 |
| 2 | Only two of the 17 briefs contain literal rc.4 wording. §5 replaces the vague "across the 17 briefs" claim with an exact seven-task retarget table. | verdict correction 2 |
| 3 | Linux is not a current rc.5 gate. §7 makes macOS primary and `ssh win` the second current gate, removes Linux success from the current 17-task chain, and names later Linux ownership. | verdict correction 3 |
| 4 | Exact platform/toolchain readiness recorded, including that `win` has no Go on PATH, and the `TASK-260720-3j8pp5` vs `TASK-260728-1j72zq` boundary. | verdict correction 4 |
| 5 | Test/disk/signing attribution corrected and re-grounded on this run's own real exit codes. | verdict test evidence |
| **6** | **New — the cycle-1 artifact cited non-canonical candidate worktrees.** The canonical snapshots named by accepted `TASK-260729-1kq1rd` are `TASK-260720-q5oy3o` (rc.4) and `TASK-260728-2kp3tv` (rc.5), not `TASK-260720-3ag6pi` and `TASK-260728-zb2s4z`. §2 uses the canonical ones. | this rework |
| **7** | **New — the accepted rc.5 publication target publishes no `expected/build-driver` suite.** No candidate root today exercises byte-exact build-driver goldens under rc.5 semantics. §3.3 and §8 record this as a hard prerequisite. | this rework |

## 1. Executive finding

The 17 CocoaSkills Go tasks form a coherent independent Python implementation plan. Every task has a concrete Curator source/test analogue, but the reusable boundary is **protocol behavior and test vectors, not Go code**. CocoaSkills must implement its own Python domain models, process layer, protected filesystem backends, transaction engine, and CLI integration.

Implementation is not ready to start today:

1. The CocoaSkills root tasks `TASK-260720-z9j4c9` and `TASK-260720-z2z795` are both hard-blocked by Curator gate `TASK-260720-1pvfj5`.
2. That gate is blocked by `TASK-260720-jrrgw9` and `TASK-260720-2qqq0w`, which in turn wait on in-flight currentness/repair work `TASK-260720-1nlmvv` (`development`).
3. Two Curator follow-ups created from currentness review findings are now in flight: `TASK-260729-2kaopg` (`development`) and `TASK-260729-3jku56` (`reviewing`). Both must be integrated into or explicitly excluded from the gate.
4. `TASK-260720-12r55p` has an extra prerequisite, `TASK-260720-3ag6pi`, which is `blocked` because no authorized landed curator-spec candidate ref exists.
5. The CocoaSkills local `main` is clean but two commits behind `origin/main` and must be fast-forwarded before any task worktree is created.

Two contract decisions must be made by a board owner before implementation, not during it:

- **rc.5 supersession.** Accepted `TASK-260729-1kq1rd` recommends landing the accepted rc.5 snapshot rather than the stale rc.4 candidate. Seven of the 17 briefs then need explicit retargeting (§5). Only `TASK-260720-12r55p` and `TASK-260720-3s27te` name rc.4 literally; the rest encode pre-amendment assumptions.
- **Build-driver golden regeneration.** The accepted rc.5 publication target publishes **no** `conformance/v1/vectors/build-drivers.json` and **no** `conformance/v1/expected/build-driver/` tree (§3.3). Until those are regenerated under rc.5 semantics, `TASK-260720-2dnqw2`, `TASK-260720-12r55p`, and `TASK-260720-3pemm6` have no shared byte-exact goldens for the build-driver surface, and Curator's own candidate-artifact test silently skips on both candidate roots.

The fastest valid start after the Curator gate closes is to launch the two CocoaSkills roots in parallel:

- `TASK-260720-z9j4c9` — schema-v6 build model.
- `TASK-260720-z2z795` — lock/journal transaction engine.

## 2. Repository and evidence provenance

### 2.1 Repositories and canonical candidate snapshots

Accepted `TASK-260729-1kq1rd` names the canonical rc.4 and rc.5 snapshots. Revision 1 of this artifact cited two different worktrees (`TASK-260720-3ag6pi`, `TASK-260728-zb2s4z`); those are working checkouts of the *verification* and *amendment* tasks, not the publication targets, and `zb2s4z` has since advanced with v7/external-repository work. This revision uses the canonical snapshots.

| Repository / snapshot | Path | Provenance and caveat |
| --- | --- | --- |
| Curator accepted composite | `.temp/TASK-260720-1ljev5/worktree` | Uncommitted task worktree based on `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`. `git ls-remote origin refs/heads/main` re-resolved to that same commit on 2026-07-29. Contains the 14 reviewed `done` implementation tasks through safe build-cache GC. |
| Curator in-flight | `.temp/TASK-260720-1nlmvv/worktree` | Same base; adds currentness/status/repair diagnostics. Board status `development` after a changes-requested review. Evidence of unresolved delta only, **not** an accepted baseline. |
| curator-spec **canonical rc.4** | `curator/.temp/TASK-260720-q5oy3o/curator-spec-worktree` | Base `57c1f56846d221ecc55786bd3c2467ec32f11730`; 34 status entries, 0 staged. README `1.0.0-rc.4`. Manifest `70cd274150d629f16ca04f2073a8eae5a0f3fff4584f905d807775040af21eae`. |
| curator-spec **canonical releasable rc.5** | `ReluxWorks/.temp/TASK-260728-2kp3tv/curator-spec-worktree` | Base `57c1f568…`; 127 status entries, 0 staged. README `1.0.0-rc.5`. Manifest `9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf`; 422 manifest files. **This is the recommended publication target.** |
| curator-spec rc.5 amendment predecessor | `ReluxWorks/.temp/TASK-260728-zb2s4z/curator-spec-worktree` | Manifest `58f8d2299c8f4a5ed78546913e567f637f1cc905dc5212b460f4097be7ff2af9`. The accepted `manager-worker-v1` amendment. Cited for its review record only; it is not the publication target and the live worktree has moved on. |
| CocoaSkills/csk | `/Users/iv/Developer/Wildberries/cocoaskills` | Clean local `main` at `edce8816dda44bb121d661b7c4dea942558ce408`, two commits behind `origin/main` `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12` (re-resolved with `git ls-remote`). `git status --short` empty. |

No candidate is a release. `refs/tags/v1.0.0-rc.4` and `refs/tags/v1.0.0-rc.5` are both absent upstream; `refs/tags/v1.0.0-rc.3^{}` and `refs/heads/main` both peel to `57c1f568…`. The Curator composite likewise cannot be described as "landed on main": its audit identity is the base commit plus the accepted board handoff chain and the retained worktree. CocoaSkills implementation tasks must record the final accepted Curator composite provenance supplied at handoff, never silently substitute `origin/main`.

### 2.2 Board-state provenance (re-resolved 2026-07-29)

CocoaSkills story `STORY-260720-1uv5gi` (`analysis`) contains exactly 17 pre-existing implementation/verification tasks, all `backlog`, plus this reconnaissance task. The story brief itself is version-neutral — it names neither rc.4 nor rc.5.

Curator gate chain: `TASK-260720-1nlmvv` `development`; `TASK-260720-jrrgw9` `backlog`; `TASK-260720-2qqq0w` `backlog`; `TASK-260720-1pvfj5` `backlog`, blocked by `2qqq0w` and `jrrgw9`; `TASK-260720-3ag6pi` `blocked`. Follow-ups: `TASK-260729-2kaopg` `development`, `TASK-260729-3jku56` `reviewing`. Both CocoaSkills roots list exactly `["TASK-260720-1pvfj5"]` as `blockedBy`.

## 3. Accepted protocol surfaces

The parity contract has **three** layers, and conflating the first two is what made revision 1 wrong.

### 3.1 Frozen schema-6 declaration bytes

These artifacts are byte-identical between the canonical rc.4 and rc.5 snapshots. Byte identity here is real and reusable:

| Artifact | SHA-256 in both rc.4 and rc.5 |
| --- | --- |
| `schemas/v1/agent-skill-v6.schema.json` | `982832e410f85e415e16e8f9104c3b9af23f6d846bbfbe5497ff170dde947f6f` |
| `schemas/v1/csk-skill-v6.schema.json` | `2148eafc4fa110311b52f528651424e2f53c69042235338fb2c8b414035eab9c` |
| `schemas/v1/build-receipt-v1.schema.json` | `f673a8815f5a5f752bc5b612f20c4ba63d9e8dcce61f5af6e7afe11b131c7ab9` |
| `schemas/v1/install-marker-v2.schema.json` | `6d7b65dbdf684272815fb0e61cc4eb02103d09dfdd397de948bd836293debeb2` |
| `schemas/v1/conformance-claim-v2.schema.json` | `4c05a97a1aa9f7dafe629a406a853239928413e79e95488ac2b20ebd0c52a38c` |

**Byte identity of these files is not semantic identity of canonical inputs, keys, receipts, or markers.** They `$ref` shared definitions in `common.schema.json`, and that file changed.

### 3.2 rc.5 execution-policy-bound canonical identity — the actual delta

`common.schema.json#/$defs/goBuildPolicyV1` gains a **required** `execution_policy` in rc.5:

- rc.4 required list: `[module_mode, network, workspace, cgo, compiler_directives, target_mode, link_mode, libgcc, package_assembly, host_objects, telemetry]`.
- rc.5 required list: the same **plus `execution_policy`**, with `"execution_policy": {"$ref": "#/$defs/goExecutionPolicyV1"}` and value `manager-worker-v1`. The source-aware variant additionally requires `source_kind`.

The consequence is three distinct, non-aliasing cache identities. All three were independently re-derived in this rework by canonicalizing the vector's `input` object with plain CCJ-1 rules (sorted keys, `,`/`:` separators, no whitespace, UTF-8, no trailing newline) and hashing — each matched its declared value exactly:

| Identity | Cache key | `schema_valid` | Meaning for CocoaSkills |
| --- | --- | ---: | --- |
| `portable` (rc.5, required) | `sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b` | `true` | The only positive case. Requires `policy.execution_policy = manager-worker-v1`. |
| `legacy_rc4_without_execution_policy` | `sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48` | `false` | **Negative, non-alias.** This is the value hard-coded in `TASK-260720-2dnqw2`'s current AC. |
| `reserved_hardened` | `sha256:13736230d33ce59de7f7323dcd4cffd510655ad8dabd5ee9e8b6cb182ec70037` | `false` | **Negative.** A reserved hardened policy must be rejected, not accepted as a future upgrade. |

The vector sets `cache_identity.aliases = false` explicitly. The three keys are numerically distinct, so a manager that ignores `execution_policy` cannot accidentally produce the right key.

The rc.5 receipt identity is `sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd`. The rc.4 receipt hash hard-coded in `TASK-260720-2dnqw2`'s AC is `sha256:750f5f7557ffe1dac1ec06b7d47cb5976e390a7dc7d57ea616c9edda91013c11`, which does not appear anywhere in the rc.5 snapshot.

**Both rc.5 identities are reproducible by an independent Python consumer today.** Canonicalizing `conformance/v1/schema-cases/build-receipt-v1/valid.json` from the rc.5 snapshot yields 1120 bytes hashing to `919fbbad…`; canonicalizing its `input` sub-object yields `529370…`, matching the file's own `cache_key` field. This is the concrete evidence that `TASK-260720-2dnqw2` and `TASK-260720-12r55p` can be built against rc.5 without copying Go code.

The claim schema also moves: `conformance-claim-v2` pins `protocol_version` to const `"1.0.0-rc.4"` in **both** snapshots, and rc.5 adds `conformance-claim-v3` with const `"1.0.0-rc.5"` plus a new required `build_drivers` field. "Claim v2" in the existing briefs is therefore an rc.4-era artifact.

### 3.3 Gap — rc.5 publishes no build-driver golden suite

The canonical rc.5 snapshot's 422-file manifest contains **zero** entries matching `build-driver`. Its 18 `vectors/` entries do not include `build-drivers.json`, and `conformance/v1/expected/build-driver/` does not exist. The canonical rc.4 snapshot has 12 manifest references and the full tree: `build-input.ccj.json`, `build-source.preimage.bin`, `build-source-sha256.txt`, `cache-key.txt`, `receipt.ccj.json`, `receipt-sha256.txt`, `toolchain.preimage.bin`, `toolchain-sha256.txt`, `marker.json`, `context_files.json`, `context_sha256.txt`.

Neither base commit has them: `57c1f568` tracks none of these paths, so both candidate lines added their own content as untracked work and the rc.4 build-driver suite was never carried into the rc.5 line.

Curator's own code makes the consequence observable. `internal/buildmeta/buildmeta_test.go` `TestCandidateBuildMetadataArtifacts` skips when `expected/build-driver` is absent, and skips again when the input lacks `execution_policy`. Running it against both candidate roots (real exit code 0 in each case, because a skip is not a failure):

| Conformance root | `TestCandidateBuildMetadataArtifacts` | `TestCandidateBuildReceiptSchemaCase` |
| --- | --- | --- |
| canonical rc.5 (`2kp3tv`) | **SKIP** — rc.5 publishes no `expected/build-driver` | **PASS** |
| canonical rc.4 (`q5oy3o`) | **SKIP** — pre-revision root without the portable execution policy | **SKIP** |

So today **no candidate root exercises the byte-exact build-driver artifact suite under rc.5 semantics.** This is a Curator-side coverage hole as well as a CocoaSkills prerequisite. It must be closed by regenerating `vectors/build-drivers.json` and `expected/build-driver/` under rc.5 before `TASK-260720-12r55p` can meet its "exact stored receipt bytes" and "CCJ-1 input and key" criteria from shared vectors rather than from a private reimplementation.

Separately, `internal/godriver/controls_test.go` hard-codes `portableVectorCacheKey` and reconstructs the vector input in Go rather than reading it from `CURATOR_CONFORMANCE_ROOT`. CocoaSkills `TASK-260720-12r55p` is specified as an *independent consumer* and must read the vector from the root instead of mirroring this shortcut.

### 3.4 Binding artifact index

| Surface | Binding artifacts |
| --- | --- |
| Schema-6 declaration | `schemas/v1/agent-skill-v6.schema.json`, `schemas/v1/csk-skill-v6.schema.json`, generated cases under `conformance/v1/schema-cases/agent-skill-v6/` and `csk-skill-v6/`, `protocol/core.md` §4.2 |
| Build-root context and raw source identity | `protocol/core.md` §§3.1, 8.1; `conformance/v1/expected/context_files.json`, `context_sha256.txt`, `snapshot_sha256.txt` (rc.5); rc.4 `expected/build-driver/build-source.preimage.bin` and `build-source-sha256.txt` pending rc.5 regeneration |
| Trusted Go identity and closed driver | `protocol/core.md` §8.2; `profiles/manager.md` §2.2; `decisions/0004-compile-only-build-drivers.md` |
| **rc.5 portable execution policy** | `protocol/core.md` §4.2.1; `profiles/manager.md` §2.2.1; `decisions/0006-portable-manager-worker-execution.md`; `docs/portable-go-execution-policy.md`; `conformance/v1/vectors/go-host-execution-policy.json` |
| Canonical input, key, receipt, marker | `protocol/core.md` §§9–10; `schemas/v1/build-receipt-v1.schema.json`; `install-marker-v2.schema.json`; `conformance-claim-v2` (rc.4) and `conformance-claim-v3` (rc.5); `schema-cases/build-receipt-v1/valid.json` (the only rc.5 positive byte golden today) |
| Planning, transaction, rollback, recovery | `profiles/manager.md` §§2.4–2.6; `conformance/v1/vectors/manager-lifecycle.json`; `cli/curator.md` |
| Activation, status, repair, GC | `profiles/manager.md` §§8, 10; `cli/curator.md` |
| Security and platform trust | `SECURITY.md`; `decisions/0004`; `decisions/0006`; negative clusters in `manager-lifecycle.json` and `go-host-execution-policy.json` |

Candidate tests must continue to use an explicit `CURATOR_CONFORMANCE_ROOT` and record its digest. `release/1.0.0-rc.5.json` records `committed_release_pin_advanced: false`. The committed default pin remains on the released prior suite until `TASK-260720-25d05o` qualifies the chosen release and the pin-audit tasks accept promotion.

### 3.5 The rc.5 execution contract CocoaSkills must implement

`protocol/core.md` §4.2.1 defines exactly one execution policy for protocol 1.0 and states that every conforming manager MUST implement `manager-worker-v1` **on macOS and Windows**. The policy identity is a normative cache, receipt, marker, and claim input, never a package-visible option or operator preference. The fixed process graph is: manager parent → identity-verified manager-owned worker → fingerprinted `<GOROOT>/bin/go` → fingerprinted regular executables below `<GOROOT>/pkg/tool/<GOHOSTOS>_<GOHOSTARCH>/`.

`go-host-execution-policy.json` carries the closed vocabularies CocoaSkills must consume:

- 18 `mandatory_controls`, from `fixed-offline-vendored-go` and `fixed-empty-environment` through `pre-launch-worker-identity-verification`, `post-exec-identity-reverification`, `worker-domain-teardown`, `no-artifact-execution`, `inventory-native-controls-applied`, and `closed-capability-evidence-record`.
- `native_control_inventory`, version `rc5-native-control-inventory-v1`, `exhaustive: true`, `probe_scope: per-operation`, `probe_timing: pre-worker-launch`, covering **exactly `["macos", "windows"]`** with five controls: `descendant-domain-termination`, `active-process-count-limit`, `aggregate-memory-limit`, `per-file-size-limit`, `inherited-handle-restriction`. Availability is per-platform and honest — e.g. `active-process-count-limit` is `available` on Windows via `job-object-active-process-limit` but `unavailable` on macOS with reason `no-private-aggregate-domain`.
- 6 `deferred_hardened_guarantees` with matching `deferred_capability_rejection_guards`, 11 `capability_evidence_cases`, 14 `identity_and_protocol_cases`, and 8 `package_influence_cases`.

## 4. Task-by-task parity map

Legend: **Accepted** = present in the reviewer-accepted `TASK-260720-1ljev5` composite. **In flight** = exists only in `TASK-260720-1nlmvv` after a changes-requested review. **Pending** = Curator task still backlog; a handoff prerequisite, not reusable evidence.

| CocoaSkills task | Curator counterpart and state | Concrete Curator source and tests | Required Python/csk adaptation | Protocol surface |
| --- | --- | --- | --- | --- |
| `TASK-260720-z9j4c9` schema-v6 build model | `2g0e3b` **Accepted** | `internal/skillspec/{types.go,parse.go,build_test.go,conformance_test.go}`; `internal/skillcheck/{skillcheck.go,skillcheck_test.go}` | Extend `src/csk/skillspec.py` dataclasses/parser and `skillcheck.py`; preserve schemas 1–5 and both manifest names; reproduce static path/module/root validation without invoking Go. Tests: `tests/test_skillspec.py`, `test_skillcheck.py`, protocol cases. | Schema-6 schemas/cases; core §4.2 |
| `TASK-260720-3c0ss2` build-source/context boundary | `256kj1`, exclusion part of `11pfex` **Accepted** | `internal/buildsource/*`; `internal/whitelist/{whitelist.go,whitelist_test.go,conformance_test.go}`; `internal/install/install_test.go`; `internal/skillcheck/*` | Add a Python frozen-snapshot token and framed digest under `src/csk/builds/`; keep `hashing.content_sha256` marker-excluding; pass `build_roots` into whitelist/runtime exclusion in real and dry-run paths. | Core §§3.1, 8.1; `expected/context_files.json`, `context_sha256.txt`, `snapshot_sha256.txt` |
| `TASK-260720-3j8pp5` Go toolchain identity | `6i3cya` **Accepted** | `internal/godriver/{session.go,fingerprint.go,identity.go}`; `session_test.go`, `fingerprint_test.go`, Unix/Windows helpers | New `src/csk/builds/toolchain.py`; direct executable resolution and subprocess calls, clean environment, private telemetry state, frozen target/tuning, byte-identical GOROOT framing. Accepted Curator allowlist is Go family `1.25`; `1.23+` is only the protocol floor. **Go-specific — do not merge with `TASK-260728-1j72zq` (§6.3).** | Core §8.2; manager §2.2 |
| `TASK-260720-2dnqw2` canonical build metadata | `3mrm4z`, `4bd0it` **Accepted** | `internal/buildmeta/{models.go,codec.go,buildmeta_test.go}`; `internal/protocoljson/ccj.go`; `internal/marker/{marker.go,marker_v2_test.go}` | Typed Python input/receipt/marker modules; reuse only the strict JSON loader/CCJ primitive; keep physical cache paths csk-specific; strict readers duplicate-key/noncanonical aware. **AC hard-codes the rc.4 key and receipt — must be retargeted (§5).** | rc.5 §3.2 identities; core §§9–10; `schema-cases/build-receipt-v1/valid.json` |
| `TASK-260720-2g21eg` go-v1 compile driver | `1zntv0` **Accepted** after portable-worker review cycle 2 | `internal/godriver/{build.go,graph.go,executor.go,workerclient.go,workerserver.go,workerproto.go,controls.go,controls_darwin.go,controls_windows.go,controls_other.go}`; `build_test.go`, `build_conformance_test.go`, `graph_test.go`, `boundary_test.go`, `controls_test.go`, worker tests, real fixture | Implement a Python subprocess/worker boundary rather than porting Go internals. Preserve the five argv forms, empty/fixed environment, graph rejection, native controls, staged-output validation, no output launch. **The brief specifies a direct `go list`/`go build` boundary; rc.5 requires hidden identity-verified worker re-execution (§5).** | Decisions 0004, 0006; core §4.2.1; manager §§2.2–2.2.1; `go-host-execution-policy.json` |
| `TASK-260720-2jfnz6` protected cache POSIX | `3pwg2w` **Accepted** | `internal/buildcache/{cache.go,publish.go,protection_unix.go}` plus cache, publication, conformance, validation, Unix tests | Backend-neutral cache API plus POSIX backend using fd/rooted no-follow operations and ownership/mode/link checks. csk may choose a different physical namespace but must preserve logical key/receipt behavior. | Core §9; manager §2.4 |
| `TASK-260720-8nxlgx` protected cache Windows | `3pwg2w` **Accepted** | `internal/buildcache/protection_windows.go`, `protection_windows_test.go`, `collect_windows_test.go` | DACL/owner/reparse/file-ID/hard-link checks via Python `ctypes` or standard APIs; module import-safe elsewhere. Cross-compilation is insufficient — native Windows negative tests required. | Same logical cache contract; Windows platform policy |
| `TASK-260720-z2z795` transaction engine | `1zl1cj`, `31nl14` **Accepted** | `internal/managerlock/*`; `internal/transaction/*`; `internal/staging/*`; identity, durability, recovery, namespace, rollback, subprocess, Darwin and Windows tests | Extend `locking.py` with canonical project/home lock hierarchy; add Python journal/transaction modules with deterministic target ordering, preimage/generation digests, crash recovery, reverse rollback, consumer-last durability. | Manager §§2.5–2.6; `manager-lifecycle.json` |
| `TASK-260720-11yhth` command runtime activation | `11pfex`, `29hi1h` **Accepted** | `internal/closure/*`; `internal/runtimestore/{scripts.go,targets.go,*_test.go}`; staged helpers in `globalbins`, `adapters`, `envfiles`; closure and shim tests | Split script-runtime completeness checking from compiled-target activation in `shims.py`; point compiled shims at immutable artifacts; preserve csk project/global/user-bin conventions and mixed-command collision rules. | Core §4; manager §8; `closures.json` |
| `TASK-260720-2x6mjn` side-effect-free planner | `3itlly` **Accepted** | `internal/install/{plan.go,private.go,stage.go,builddeps.go}`; `private_test.go`, `stage_test.go`, `revalidation_test.go`; snapshot locks | Pure Python planner plus explicit read-only audit/registry paths. Do not reuse the coarse `GlobalLock` routing for dry-run. Model every persistent write as a later staged target. | Manager §§2.4–2.5; compiler-free dry-run vectors |
| `TASK-260720-3t8nr3` project/hybrid transaction | `2284br` **Accepted** | `internal/install/{install.go,commit.go,generation.go,targets.go,atomicity/*}`; `internal/scopes/{stage.go,hybrid.go}`; adapter/env/global-bin staging; commit, ABA, atomicity, revalidation tests | Refactor `installer.py` from per-node/per-target mutation to plan → private builds → home-lock revalidation/publication → one journaled commit. Express project, hybrid, runtime, context, marker, shims, adapters, env, stale removals, and consumer ledger as transaction targets. | Manager §§2.5–2.6 |
| `TASK-260720-g7kgox` global transaction | Global half of `2284br` **Accepted** | `internal/install/global.go`; `internal/globalbins/*`; `internal/adapters/*`; `internal/envfiles/*`; global cases in `install_test.go` and atomicity tests | Reuse the same planner/transaction implementation in `global_install.py`; retain csk's global manifest and user-bin selection but remove partial materialization. | Same lifecycle contract |
| `TASK-260720-th0jdi` currentness/repair/GC | Marker and GC parts `4bd0it` + `1ljev5` **Accepted**; diagnostic/repair layer `1nlmvv` **In flight** | Accepted: `internal/marker/*`, `internal/buildcache/collect.go`, `internal/scopes/{gc.go,gc_*test.go}`, `cmd/curator/gc_test.go`. In flight: `cmd/curator/{builds.go,builds_test.go,status_test.go}`, `internal/install/{diagnostics.go,diagnostics_test.go}` | Implement only after the reworked Curator vocabulary/semantics are accepted. Extend `status.py` and `global_install.py` status; make install/upgrade the repair path; keep GC conservative across corrupt/unknown markers, consumers, journals, redirected roots. **Add execution-policy mismatch as a currentness/rebuild dimension (§5).** | Marker v2; manager §§2.6, 10; CLI status/repair/GC |
| `TASK-260720-12r55p` shared v6 vector consumer | Package conformance tests **Accepted**; integrated Curator `jrrgw9` **Pending**; literal rc.4 protocol `3ag6pi` **Blocked** | `internal/{skillspec,buildsource,buildmeta,buildcache,godriver,closure,whitelist,runtimestore}/…conformance_test.go`; `internal/interop/golden_test.go` | Independent Python assertions in `tests/test_protocol_conformance.py`; adapters may parse shared fixtures but must not duplicate product logic or copy Go code. Read vectors from `CURATOR_CONFORMANCE_ROOT` — do not hard-code them as `controls_test.go` does. Keep legacy rc.3 green. **Literal rc.4 wording; blocked on the rc.5 build-driver golden gap (§3.3, §5, §8).** | All schema-6 cases, lifecycle vectors, `go-host-execution-policy.json`, claim-version separation |
| `TASK-260720-akf5kh` user documentation | Protocol author/CLI docs `3lo9jc` **Accepted**; Curator product docs `2qqq0w` **Pending** | Candidate `README.md`, `SECURITY.md`, `cli/curator.md`, `profiles/manager.md`; accepted Curator composite `README.md` | Adapt terminology and paths to csk (`README*`, `ARCHITECTURE*`, `SECURITY*`, authoring docs) including maintained Russian mirrors. Do not copy Curator CLI claims before final diagnostics/currentness behavior is accepted. **Must document the portable manager-worker boundary, honest capability evidence, the macOS/Windows support claim, and the deferred Linux boundary (§5).** | Authoring, CLI, security, lifecycle, activation, status/GC docs |
| `TASK-260720-3pemm6` cross-platform Go E2E | Platform-specific Curator tests **Accepted** throughout packages; Curator `1pvfj5`, `jrrgw9` **Pending** | `internal/godriver/*_{darwin,windows,unix}_test.go`; buildcache Windows/POSIX tests; managerlock/transaction platform tests; runtimestore/shim tests; install/atomicity tests | Add a real vendored Go skill and black-box Python subprocess E2E. **Current AC requires a green real-fixture Go build on ubuntu; rc.5 forbids it (§5, §7).** CI must set up the accepted Go family, accept an explicit candidate suite, and retain the old default pin. | Full driver/lifecycle/activation vectors |
| `TASK-260720-3s27te` integrated verification | Whole accepted composite plus `jrrgw9`, `2qqq0w`, `1pvfj5` once accepted | Entire Curator package/test graph and final platform evidence | Verification only: clean CocoaSkills worktree; full pytest, strict mypy, build/twine, diff check, candidate conformance, CI matrix. Route semantic failures back to owning tasks. **Literal rc.4 root and three-OS final gate — both must be retargeted (§5).** | Every story criterion and required vector cluster |

All 17 pre-existing CocoaSkills tasks are represented exactly once above; the exact-ID and row-count assertions are in §10.

## 5. Exact brief and dependency retargets

A scoped board search over all 17 briefs (`description`, `scope`, `ac`) finds literal `rc.4`/`rc4` wording in **exactly two**: `TASK-260720-12r55p` (1 hit) and `TASK-260720-3s27te` (2 hits). The parent story `STORY-260720-1uv5gi` is version-neutral. The other fifteen are version-neutral in wording but seven of them encode pre-amendment assumptions that a supersession decision must correct.

The following retargets are required **after** a board owner approves rc.5 supersession, and must not be applied unilaterally.

| Task | Literal rc.4? | Required retarget |
| --- | --- | --- |
| `TASK-260720-2dnqw2` | no | Replace the rc.4 goldens in its AC. Cache key `3fcd714a…` → `sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b`; receipt `750f5f75…` → `sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd`. Require `policy.execution_policy = manager-worker-v1` in the canonical input, and add `3fcd714a…` (missing policy) and `13736230d3…` (reserved hardened) as negative non-alias cases. Decide whether the marker/claim surface targets claim v2 (rc.4) or claim v3 (rc.5, adds required `build_drivers`). |
| `TASK-260720-2g21eg` | no | Replace the direct source-aware `go list`/`go build` process boundary with the hidden identity-verified `manager-worker-v1` re-execution: authenticated one-list/one-build session, pre-launch worker identity verification, per-operation native-control preflight against `rc5-native-control-inventory-v1`, worker-domain teardown, post-exec identity re-verification, frozen-snapshot integrity re-check, and a closed `capability-evidence-v1` record. Its 18 mandatory controls and 14 identity/protocol negative cases become AC. Deferred hardened guarantees must be rejected, not silently claimed. |
| `TASK-260720-12r55p` | **yes** | Change literal rc.4 candidate consumption to the landed rc.5 root and digest. Add `go-host-execution-policy.json`, the native-control inventory, capability-evidence cases, and the legacy/reserved non-alias cases. Move claim coverage from claim v2 to claim v3 if rc.5 lands. **Hard prerequisite: the rc.5 root must publish `vectors/build-drivers.json` and `expected/build-driver/` (§3.3), or this task cannot assert byte-exact receipt/key/preimage goldens from shared vectors.** Keep its hard edge to `TASK-260720-3ag6pi` only if the owner retargets and re-reviews that task as the rc.5 verification gate; otherwise relink it to the replacement rc.5 verification task. |
| `TASK-260720-akf5kh` | no | Document the portable manager-worker boundary, honest capability evidence (available vs unavailable controls per platform, with reasons), the macOS/Windows support claim, and the explicitly deferred Linux boundary. Do not document hardened guarantees the implementation does not provide. |
| `TASK-260720-3pemm6` | no | Replace the "Linux, macOS, and Windows" real-fixture claim with current rc.5 native macOS and Windows gates. Linux CI may still run portable non-driver coverage but must not be required to produce a green `go-v1` build; route Linux driver qualification to later work unless the protocol owner first extends the inventory and accepts a new contract. |
| `TASK-260720-3s27te` | **yes** | Replace the literal "immutable curator-spec rc.4 conformance root" and the "Linux, macOS, and Windows CI matrix with the real Go fixture is green" AC with the landed rc.5 root/digest and current macOS/Windows qualification. Keep Linux explicitly deferred. |
| `TASK-260720-th0jdi` | no | State that execution-policy mismatch is a currentness/rebuild dimension. Its complete-key/receipt checks already cover this transitively, so this is a clarification, not new behavior. |

Curator-side note, outside the 17: `TASK-260720-jrrgw9` is literally named `verify-rc4-build-conformance`. It gates `TASK-260720-1pvfj5`, which gates both CocoaSkills roots. A supersession decision must state whether that task verifies rc.4 or rc.5, because CocoaSkills inherits whichever suite it accepts.

## 6. Adaptation and reuse boundaries

### 6.1 Reuse directly

- Canonical JSON/CCJ-1 rules and the exact expected bytes that exist in the chosen landed root.
- Schema files, generated positive/negative cases, fixtures, and vector names.
- The rc.5 closed vocabularies: mandatory controls, native-control inventory, capability-evidence record shape, deferred-guarantee rejection guards.
- Stable diagnostic codes — only after `TASK-260720-1nlmvv` is accepted.
- Ordering rules: provider-first, unsigned UTF-8 lexical command order, deterministic target-class order, consumer last.
- Security invariants: audit before compiler; immutable/protected cache; no package-controlled args/env/output name; no output execution during install; compiler-free dry-run.
- Platform acceptance matrices and negative cases.

### 6.2 Reimplement independently

- Python dataclasses/readers and error types.
- Subprocess/worker execution, worker identity proof, and resource-control binding.
- POSIX fd-rooted and Windows handle/DACL/reparse filesystem operations.
- Lock identity, journaling, durability, recovery, and rollback.
- csk-specific cache paths, global manifest, user-bin selection, adapter/environment materialization, and CLI rendering.

### 6.3 Do not reuse or merge

- Go package structure or copied Go algorithms.
- Curator physical cache/lock/journal path names.
- Test-only Curator worker dispatch, internal dependency injection, or `controls_test.go`'s hard-coded vector input.
- Unaccepted `TASK-260720-1nlmvv` diagnostics and status semantics.
- A working-tree candidate treated as a published protocol release or manager pin.
- **`TASK-260720-3j8pp5` and `TASK-260728-1j72zq` must stay separate.** `3j8pp5` is the Go-specific trusted-toolchain identity inside this 17-task chain, blocked by `z9j4c9` and blocking `2dnqw2`/`2g21eg`. `1j72zq` (`implement-trusted-toolchain-preflight-in-csk`, parent `STORY-260728-2fsqtv`) is the later generic declarative toolchain-preflight implementation, blocked by `TASK-260728-2bu2q6`, `TASK-260728-2gbtb9`, and `TASK-260720-3s27te` — i.e. downstream of this entire chain. Neither may absorb the other's scope.

## 7. Platform gates

### 7.1 Current rc.5 gates — macOS primary, Windows second

`protocol/core.md` §4.2.1 requires `manager-worker-v1` on macOS and Windows. `rc5-native-control-inventory-v1` is marked `exhaustive: true` and covers exactly `["macos", "windows"]`. Accepted Curator `internal/godriver/controls_other.go` (build tag `!darwin && !windows`) fails closed on every other host: `prepareControlDomain` returns `CodeControlUnavailable` with "the portable execution policy is specified for macOS and Windows only", and `probeNativeControl` rejects before the worker starts. A source-aware `go-v1` build therefore **cannot succeed on Linux** under the accepted contract.

| Gate class | macOS (primary, current) | Windows via `ssh win` (current) |
| --- | --- | --- |
| Static models/bytes | Native pytest/mypy for schema, framing, CCJ-1, receipt, marker | Import/path/case vectors; exact bytes must match |
| Toolchain/driver | Native Go 1.25-family probe, real vendored build, worker identity proof, per-operation control preflight, poisoned-env negatives | Native executable/GOROOT identity, job-object controls, path/DACL/reparse behavior, real vendored build |
| Cache | Native POSIX ownership/mode/no-follow/link/race/publication tests | Native DACL/owner/reparse/hard-link/file-ID/race/publication tests; cross-build alone does not qualify |
| Locks/transactions | Native file locking, Darwin case/Unicode namespace, fsync/rename/recovery/rollback | Native canonical identity, first-use case aliases, durability, replace/journal cleanup, recovery/rollback |
| Shims/activation | POSIX project/global argv and exit propagation | `.cmd` quoting, `%*`, `ERRORLEVEL`, injection rejection, project/global/user-bin paths |
| Status/repair/GC | Native protected-state inspection, conservative two-pass GC | Native protected boundary, reparse/redirect, lock serialization, repair and GC |
| Final story gate | Python 3.11–3.14 macOS CI plus strict mypy/build metadata | Python 3.11–3.14 Windows CI, no unexpected skip/xfail in task-owned cases |

Curator review history shows why native Windows is mandatory: cross-compilation passed while later native runs found lock identity, DACL, journal durability, and GC containment defects. CocoaSkills should budget native Windows review cycles from the start.

### 7.2 Linux is deferred, and it is owned

Linux is not silently dropped. The accepted rc.5 snapshot states the exclusion normatively. `conformance/v1/vectors/conformance-claim-v3-qualification.json` declares:

```json
"platforms": [
  {"name": "linux",   "status": "excluded", "until_task": "TASK-260728-1skseh"},
  {"name": "macos",   "status": "pending-downstream-native-evidence"},
  {"name": "windows", "status": "pending-downstream-native-evidence"}
]
```

with qualification rules `schema-valid-is-not-qualified`, `driver-platform-subset`, `no-generic-driver` (allowed drivers `go-repository-v1`, `go-v1`), and `no-unevidenced-platform`.

Named later ownership:

- `TASK-260728-1skseh` — `run-linux-native-external-repository-qualification`, `backlog`, parent `STORY-260728-1eye8p` (`linux-native-external-build-validation`). Provisions the dedicated Linux host and runs the rc.5 external-build-repository suite before proposing Linux claim evidence. This is the exact task the rc.5 vector names as the exclusion unblocker.
- `TASK-260728-1e6811` — `qualify-linux-toolchain-preflight`, `backlog`, the deferred clean-host Linux toolchain-preflight qualification. It is already `blockedBy` `TASK-260720-3pemm6`, so it sits downstream of this chain by construction.

**Gap to flag for the owner:** neither of those tasks is a full Linux lifecycle gate for local `go-v1`. `1skseh` covers external build repositories; `1e6811` covers toolchain preflight. If local `go-v1` needs its own Linux lifecycle qualification, the owner must create or designate that task — it cannot be inferred from a macOS/Windows-only inventory.

CocoaSkills CI today declares `os: [ubuntu-latest, macos-latest, windows-latest]` across Python 3.11–3.14. Ubuntu jobs may keep running portable non-driver coverage (models, CCJ, POSIX cache, locks/transactions, shims), but no `go-v1` build/E2E success may be required or claimed there while the rc.5 inventory excludes Linux.

A Linux host is temporarily reachable at SSH alias `lev`. Per the operator directive on this run it is recorded here only as an **available later qualification surface**, not as a reason to expand the current chain or delay the Go critical path. The read-only inventory captured in the 2026-07-29 0440 logbook entry found Ubuntu 26.04 LTS x86_64 with no Go on PATH, no conventional Go root, and no installed `golang*` package; the distribution offers Go 1.26, which is outside the accepted `1.25` family. Any future Linux work through `lev` would therefore still require a manually installed, manager-approved official Go 1.25 linux-amd64 tree, and it remains owned by `TASK-260728-1skseh` / `TASK-260728-1e6811` rather than by any of the 17 tasks.

### 7.3 Verified host readiness

| Host | Status | Evidence |
| --- | --- | --- |
| macOS primary | **Ready.** macOS 26.5, arm64, `go version go1.25.5 darwin/arm64`, Python 3.14.4. Satisfies the currently accepted Curator Go-family allowlist. | `sw_vers`, `uname -m`, `go version`, `python3 --version` |
| Windows via `ssh win` | **Reachable, not build-ready.** `Microsoft Windows NT 10.0.19045.0` (exit 0). `go` is **not on PATH**: `Get-Command go` returns an empty `Source` and `ssh win 'go version'` exits **1**. No host install was performed. | see §10.4 |

Before any Windows native driver/cache/transaction gate, the handoff must supply **either** an approved installation of a supported Go family on `win`, **or** an exact operator-trusted absolute Go path plus its fingerprinting/tuning evidence. `TASK-260720-3j8pp5` explicitly rejects repository- or project-managed Go candidates, so an ad-hoc unpacked toolchain on `PATH` is not acceptable evidence.

Local disk is a real constraint: the data volume ran at 98–99% capacity during this rework (9.3–18 GiB free, moving as concurrent gates ran). Full-repository Go runs must preflight free space and use task-local temp roots; CI remains the authoritative full platform gate.

## 8. Exact handoff prerequisites

Before the first CocoaSkills implementation task:

- [ ] A board owner records the rc.5 supersession decision, then applies the seven retargets in §5. Recommended per accepted `TASK-260729-1kq1rd`: land the accepted `TASK-260728-2kp3tv` rc.5 snapshot. The alternative — publishing literal rc.4 — diverges from the already accepted Curator driver boundary.
- [ ] **The landed rc.5 root regenerates `conformance/v1/vectors/build-drivers.json` and `conformance/v1/expected/build-driver/` under `execution_policy = manager-worker-v1`.** Until then `TASK-260720-2dnqw2`, `TASK-260720-12r55p`, and `TASK-260720-3pemm6` have no shared byte-exact build-driver goldens, and Curator's `TestCandidateBuildMetadataArtifacts` skips on every available root (§3.3).
- [ ] `TASK-260720-1nlmvv` is reviewer-accepted, including production-reachable stable diagnostics, non-attributing input drift, atomic repair of corrupt cache state, and bounded cross-platform redaction.
- [ ] The coordinator resolves or explicitly scopes `TASK-260729-2kaopg` (`development`) and `TASK-260729-3jku56` (`reviewing`) before Curator conformance/CI acceptance.
- [ ] `TASK-260720-jrrgw9`, `TASK-260720-2qqq0w`, and `TASK-260720-1pvfj5` are reviewer-accepted; `1pvfj5` is the actual hard blocker of both CocoaSkills roots. The supersession decision must state which suite `jrrgw9` verifies.
- [ ] The Curator handoff records the exact accepted composite identity: base commit, task outcome chain, worktree/archive digest, candidate conformance-root digest, and macOS/Windows gate evidence.
- [ ] `/Users/iv/Developer/Wildberries/cocoaskills` is still clean, is fast-forwarded from `edce8816…` to `6fc2fd97…` (or the then-current verified `origin/main`), and each task worktree records that base before edits.
- [ ] The final accepted Go-family allowlist and native tuning contract are copied from protocol/Curator handoff evidence. Current Curator code allows `1.25`; `1.23+` is only the protocol floor.
- [ ] A macOS native runner is available (verified ready) and `win` has an approved supported Go install or an operator-trusted absolute Go path with identity evidence (§7.3). Windows security/durability tasks must not be accepted from cross-compile evidence alone.
- [ ] Candidate conformance input remains explicit and immutable; no committed suite pin moves merely because a working-tree candidate passes. `release/1.0.0-rc.5.json` records `committed_release_pin_advanced: false`.

Before `TASK-260720-12r55p`:

- [ ] An authorized curator-spec commit/ref exists for the owner-selected release line. No `v1.0.0-rc.4` or `v1.0.0-rc.5` tag exists upstream today.
- [ ] Literal rc.4 path: `TASK-260720-3ag6pi` is rerun against that real clean ref without virtual-index or status wrappers and reaches reviewer acceptance.
- [ ] Recommended rc.5 path: the brief and its `blockedBy` edge are retargeted to the landed rc.5 snapshot; conformance covers schema-6 declarations plus `manager-worker-v1` identity, the native-control inventory, capability evidence, non-alias negatives, and failure-boundary vectors.
- [ ] `CURATOR_CONFORMANCE_ROOT`, repository SHA, complete manifest/suite digest, and the previous released default pin are recorded.

Before the final CocoaSkills gate (`TASK-260720-3s27te`):

- [ ] Every predecessor is accepted and landed in the CocoaSkills integration base.
- [ ] macOS and native Windows real-fixture CI is green across Python 3.11–3.14 with no unexpected task-owned skip/xfail. Linux driver success is **not** required and must not be claimed (§7.2).
- [ ] Full pytest, strict mypy, `python -m build`, `python -m twine check dist/*`, candidate conformance, and `git diff --check` run as standalone gates with real exit codes.
- [ ] Release/pin promotion remains owned by `TASK-260720-25d05o` and `TASK-260720-1utsx8`; candidate qualification alone must not fabricate release evidence.

## 9. Critical path and parallel order

### 9.1 External gate to CocoaSkills start

```text
TASK-260720-1nlmvv currentness/repair rework accepted   (development)
  ├─> TASK-260720-jrrgw9 Curator build conformance      (backlog)
  └─> TASK-260720-2qqq0w Curator compiled-build docs     (backlog)
          \                 /
           TASK-260720-1pvfj5 cross-platform CI gate      (backlog)
                    |
          +---------+---------+
          |                   |
  z9j4c9 schema model   z2z795 transaction engine
```

`TASK-260729-2kaopg` (global status, `development`) and `TASK-260729-3jku56` (repeated-install idempotence, `reviewing`) were created from currentness review findings. Before `jrrgw9` or `1pvfj5` closes, the coordinator must either integrate their accepted behavior or explicitly prove they are outside the gate; otherwise CocoaSkills inherits a known parity gap.

### 9.2 CocoaSkills implementation critical path

Verified against live board edges; the only two roots are `z9j4c9` and `z2z795`.

```text
z9j4c9
  ├─> 3c0ss2 ─┐
  └─> 3j8pp5 ─┴─> 2dnqw2 ─> 2jfnz6 ─> 8nxlgx ─> 11yhth ─┐
                 └─> 2g21eg ────────────────────> 2x6mjn ─┤
z2z795 ─────────────────────────────────────────> 2x6mjn ─┘
                                                          |
                       3t8nr3 -> g7kgox -> th0jdi
                                      ├─> 12r55p -> 3pemm6 ─┐
                                      └─> akf5kh ───────────┴─> 3s27te
```

Additional joins: `12r55p` currently waits on `TASK-260720-3ag6pi` (`blocked`), which must reach acceptance on a real ref or be retargeted (§5). `akf5kh` also waits on already accepted protocol docs `TASK-260720-3lo9jc`. Downstream of the chain and outside the 17: `TASK-260720-31zeo2` (`csk-shared-build-suite-consumer`, parent `STORY-260720-21bsr2`) and `TASK-260728-1j72zq`/`TASK-260728-1e6811` (§6.3, §7.2).

Recommended execution:

1. Start `z9j4c9` and `z2z795` in parallel after `1pvfj5` is `done`.
2. After the model, run `3c0ss2` and `3j8pp5` in parallel.
3. After both, run `2dnqw2` and `2g21eg` in parallel — but only once the rc.5 retargets in §5 are applied, since both hard-code pre-amendment expectations.
4. Keep cache backends serial: POSIX `2jfnz6`, then Windows `8nxlgx`.
5. Start `11yhth` after Windows cache; start `2x6mjn` once driver, Windows cache, and transaction engine are accepted.
6. Join both at `3t8nr3`, then serialize global integration and maintenance: `g7kgox` → `th0jdi`.
7. Run shared vectors and docs in parallel; then E2E; then integrated verification.

## 10. Fact-check record

Every command below was run as a standalone process with its real exit code captured; none was piped through `tee` or a status-swallowing chain. No checklist item is checked on the basis of an unrun or assumed-green command.

### 10.1 Provenance and board

| Command | Exit | Finding |
| --- | ---: | --- |
| `git ls-remote origin refs/heads/main` in Curator | 0 | `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8` |
| `git ls-remote origin refs/heads/main` in CocoaSkills | 0 | `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12` |
| `git rev-parse HEAD` in CocoaSkills | 0 | `edce8816dda44bb121d661b7c4dea942558ce408` — two commits behind |
| `git status --short` in CocoaSkills | 0 | empty; clean |
| `git rev-parse HEAD` in both candidate worktrees | 0 | `57c1f56846d221ecc55786bd3c2467ec32f11730` for each |
| `git status --short \| wc -l` on rc.4 / rc.5 snapshots | 0 | 34 and 127 entries, matching accepted `TASK-260729-1kq1rd` |
| `task-board q 'list(type=task, parent=STORY-260720-1uv5gi)'` | 0 | 17 pre-existing tasks + this one; exact-ID set matches §4 with no omission or duplicate |
| Per-task `rc.4`/`rc4` grep over all 17 briefs | 0 | literal hits in exactly two: `12r55p` (1), `3s27te` (2) |
| `task-board q` on gate tasks | 0 | statuses as recorded in §2.2 |

### 10.2 Protocol identity

| Check | Exit | Finding |
| --- | ---: | --- |
| `shasum -a 256` on v6 schemas across rc.4/rc.5 | 0 | `982832e4…` and `2148eafc…` byte-identical in both |
| `grep` `goBuildPolicyV1` required lists | 0 | rc.5 adds `execution_policy`; rc.4 has none |
| Python CCJ-1 re-derivation of all three `cache_identity` inputs | 0 | `529370…`, `3fcd714a…`, `13736230…` each match their declared `cache_key`; `aliases: false` |
| Python CCJ-1 hash of `schema-cases/build-receipt-v1/valid.json` | 0 | 1120 bytes → `919fbbad…`, matching Curator's `wantReceiptHash`; its `input` → `529370…`, matching the file's own `cache_key` |
| rc.5 manifest scan for `build-driver` | 0 | 0 of 422 entries; 18 `vectors/` entries, none `build-drivers.json`; `expected/build-driver/` absent |
| rc.4 manifest scan for `build-driver` | 0 | 12 references; full `expected/build-driver/` tree present |
| `git cat-file -e 57c1f568:<path>` for v6 schema, build-drivers vector, cache-key | non-zero (expected) | none tracked in the shared base — both candidate lines added their own untracked content |
| `grep` claim schemas | 0 | `conformance-claim-v2` pins `1.0.0-rc.4` in both snapshots; rc.5 adds `conformance-claim-v3` pinning `1.0.0-rc.5` with required `build_drivers` |
| Read `conformance-claim-v3-qualification.json` | 0 | `linux: excluded until_task TASK-260728-1skseh`; macos/windows `pending-downstream-native-evidence` |
| Read `native_control_inventory` | 0 | `rc5-native-control-inventory-v1`, `exhaustive: true`, `platforms: ["macos","windows"]`, 5 controls |
| Read `internal/godriver/controls_other.go` | 0 | non-darwin/non-windows hosts fail closed with `CodeControlUnavailable` |

### 10.3 Curator tests run in this rework

`TMPDIR` was set to `/private/tmp/TASK-260729-1t1z2l`, outside any Git worktree, after the cycle-1 review found that a `TMPDIR` inside a Git worktree causes environmental failures.

| Command | Exit | Finding |
| --- | ---: | --- |
| `go test ./internal/buildmeta/ -run TestCandidate -v` with `CURATOR_CONFORMANCE_ROOT` = rc.5 root | **0** | `TestCandidateBuildMetadataArtifacts` **SKIP** (rc.5 publishes no `expected/build-driver`); `TestCandidateBuildReceiptSchemaCase` **PASS** |
| same with `CURATOR_CONFORMANCE_ROOT` = rc.4 root | **0** | **both SKIP** — rc.4 is a pre-revision root without the portable execution policy |
| `go test` over `buildmeta buildsource skillspec skillcheck marker whitelist godriver buildcache` | **0** | all `ok` |
| `go test` over `managerlock transaction staging runtimestore scopes closure protocoljson interop` (first attempt) | **1** | `internal/transaction` **build failed** — the linker could not open a `go-build` cache object. Attributable to my own preceding run being killed mid-link (below), not to product code. All seven other packages `ok`. |
| `go test ./internal/transaction/` (retry) | **0** | `ok … 14.721s` — confirms the previous failure was a transient build-cache artifact |

Honest attribution of the exit-143 events:

- The cycle-1 producer reported `go test ./...` exit 143 with `internal/install` hitting `no space left on device`. The cycle-1 reviewer's controlled replays did **not** reproduce that as a product failure: the first replay failed on host Git SSH-signing contamination, and the second was deliberately stopped when a concurrent gate consumed disk headroom.
- In this rework I reproduced an exit 143 of my own — it was **my tool's 120 s timeout killing the run**, not a storage error. Its only lasting effect was the truncated build-cache entry that made `internal/transaction` fail once and pass on retry.
- Neither 143 is evidence of a Curator defect. Equally, **no repository-wide `go test ./...` green run is claimed by this task**; a clean serialized full replay was not completed here, and CI remains the authoritative full gate.

### 10.4 Platform probes

| Command | Exit | Finding |
| --- | ---: | --- |
| `sw_vers` / `uname -m` | 0 | macOS 26.5, build 25F71, arm64 |
| `go version` | 0 | `go1.25.5 darwin/arm64` |
| `python3 --version` | 0 | 3.14.4 |
| `df -h /System/Volumes/Data` | 0 | 98–99% capacity; 9.3–18 GiB free during the run |
| `ssh win 'powershell … OSVersion.VersionString'` | **0** | `Microsoft Windows NT 10.0.19045.0` — host reachable |
| `ssh win 'powershell … (Get-Command go).Source'` | 0 | empty output — Go not found |
| `ssh win 'go version'` | **1** | Go is not on PATH. No install performed. |

### 10.5 No-change verification

| Command | Exit | Finding |
| --- | ---: | --- |
| `git status --short -- cmd internal go.mod go.sum README.md Makefile .github` in Curator root | 0 | empty; no Curator product/config surface touched |
| `git status --short` in CocoaSkills | 0 | empty |
| `git diff --cached --quiet` in Curator and CocoaSkills | 0 each | nothing staged |

Prior green evidence is cited as **reviewed historical evidence**, not as a rerun in this task: `TASK-260720-1ljev5_review-verdict-cycle-3.md`, `TASK-260720-1zntv0_portable-review-cycle-2-verdict.md`, `TASK-260720-31nl14_review-cycle-11-verdict.md`, `TASK-260720-3ag6pi_reviewer-verdict.md`, `TASK-260728-zb2s4z_review-cycle-2-verdict.md`, `TASK-260720-1nlmvv_review-verdict-cycle-1.md`, and `TASK-260729-1kq1rd_protocol-publication-gate-for-parity.md`.

## 11. Risks and mitigations

| Risk | Evidence | Required mitigation |
| --- | --- | --- |
| rc.5 changes canonical identity while briefs assume rc.4 | `2dnqw2` AC hard-codes `3fcd714a…`/`750f5f75…`, both non-alias or absent in rc.5 | Apply the §5 retargets after an explicit owner decision; add the two negative keys as required rejection cases |
| **rc.5 publishes no build-driver golden suite** | rc.5 manifest has 0 of 422 `build-driver` entries; Curator's candidate-artifact test skips on both roots | Regenerate `vectors/build-drivers.json` and `expected/build-driver/` under rc.5 before `12r55p`; treat the current skip as a Curator coverage hole, not as passing evidence |
| Linux required by briefs but excluded by contract | `controls_other.go` fails closed; inventory is macOS/Windows only; claim-v3 vector marks Linux `excluded` | Retarget `3pemm6`/`3s27te`; keep Ubuntu CI for portable non-driver coverage only; designate a Linux lifecycle owner if one is wanted |
| Windows gate cannot run | `ssh win 'go version'` exits 1; no Go on PATH | Approved supported Go install or operator-trusted absolute path plus identity evidence before any Windows driver/cache/transaction acceptance |
| Treating an uncommitted composite as a release | No `v1.0.0-rc.4`/`rc.5` tags upstream; `main` and `rc.3` both peel to `57c1f568` | Require exact artifact/archive digests now and a real immutable ref before released-suite claims |
| Cycle-1 provenance drift | Revision 1 cited `3ag6pi`/`zb2s4z`; `zb2s4z` has since gained v7/external-repository content | Pin all future citations to the canonical `q5oy3o`/`2kp3tv` snapshots and their recorded manifest digests |
| Copying Curator internals into Python | Story explicitly requires an independent consumer; `controls_test.go` hard-codes its vector | Share only schemas, vectors, bytes, codes, and black-box cases; read vectors from `CURATOR_CONFORMANCE_ROOT` |
| Dry-run mutates registry/cache/state | `installer.py` performs `_check_audit_registries` before dry-run return | Explicit read-only planner/audit path plus whole-tree before/after tests |
| Existing coarse lock is the wrong lifecycle | `GlobalLock` lacks the accepted lock hierarchy, journal, and revalidation semantics | Project locks + home mutation lock + durable generic target journal before installer integration |
| Partial per-node mutation breaks rollback | `installer.py` materializes consumers/runtime/context incrementally | Every live surface becomes a staged transaction target; consumer last |
| Physical cache trust confused with receipt consistency | Protocol treats hashes as consistency metadata, not provenance | Validate the owner-protected boundary before parsing or adopting receipt/artifact bytes |
| Native Windows bugs hidden by cross-build | Curator review found native-only DACL, identity, durability, and GC issues | Require native Windows positive/negative/race evidence on the owning tasks |
| In-flight status vocabulary copied too early | `1nlmvv` review found unreachable codes, misclassification, repair refusal, path leaks | Wait for accepted production-path semantics; consume the final code list as a handoff artifact |
| Space-heavy verification destabilizes local gates | Data volume at 98–99% throughout this rework; one interrupted run corrupted a build-cache object | Preflight free space, use task-local temp roots outside Git worktrees, preserve standalone exits, keep CI authoritative |

## 12. Recommendation

Do not collapse or renumber the 17 CocoaSkills tasks. The decomposition matches Curator's actual subsystem boundaries and correctly isolates the two platform cache backends, transaction infrastructure, project/global integration, maintenance, vectors, docs, E2E, and final verification. The required changes are contract retargets (§5) and platform-gate corrections (§7), not restructuring.

Two owner decisions gate the start, and both are cheap to make now and expensive to make late:

1. **Approve rc.5 supersession** and apply the seven §5 retargets. Without it, `2dnqw2` and `2g21eg` would be implemented against a contract the accepted Curator driver no longer honors.
2. **Require rc.5 build-driver golden regeneration** as part of landing the rc.5 snapshot. Without it, `12r55p` cannot satisfy its byte-exact criteria from shared vectors, and the gap is currently invisible because the corresponding Curator test skips rather than fails.

At Curator gate closure, hand CocoaSkills two immutable packets:

1. the accepted Curator composite provenance plus the task-to-file/test inventory and native macOS/Windows platform evidence;
2. the owner-selected landed protocol ref and complete conformance digest, explicitly reconciling the frozen schema-6 declaration bytes with the rc.5 execution-policy-bound canonical identity, and confirming the build-driver golden suite is present.

Then start only `z9j4c9` and `z2z795` in parallel. This preserves the no-bypass dependency rule while removing reconnaissance latency from the first implementation cycle.

## 13. References

### Board outcomes

- `.task-board/.resources/STORY-260720-1uv5gi/STORY-260720-1uv5gi_17-task-audit.md`
- `.task-board/.resources/STORY-260720-1uv5gi/STORY-260720-1uv5gi_decomposition-plan.md`
- `.task-board/.resources/STORY-260720-1uv5gi/STORY-260720-1uv5gi_planning-validation.md`
- `.task-board/.resources/TASK-260720-1ljev5/TASK-260720-1ljev5_review-verdict-cycle-3.md`
- `.task-board/.resources/TASK-260720-1zntv0/TASK-260720-1zntv0_portable-review-cycle-2-verdict.md`
- `.task-board/.resources/TASK-260720-31nl14/TASK-260720-31nl14_review-cycle-11-verdict.md`
- `.task-board/.resources/TASK-260720-3ag6pi/TASK-260720-3ag6pi_reviewer-verdict.md`
- `.task-board/.resources/TASK-260728-zb2s4z/TASK-260728-zb2s4z_review-cycle-2-verdict.md`
- `.task-board/.resources/TASK-260720-1nlmvv/TASK-260720-1nlmvv_review-verdict-cycle-1.md`
- `.task-board/.resources/TASK-260729-1kq1rd/TASK-260729-1kq1rd_protocol-publication-gate-for-parity.md`
- `.task-board/.resources/TASK-260729-1t1z2l/TASK-260729-1t1z2l_review-verdict-cycle-1.md`

### Source roots

- Curator accepted composite: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1ljev5/worktree`
- Curator currentness/repair rework: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1nlmvv/worktree`
- Canonical rc.4 snapshot: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-q5oy3o/curator-spec-worktree`
- Canonical releasable rc.5 snapshot: `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-2kp3tv/curator-spec-worktree`
- rc.5 amendment predecessor: `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-zb2s4z/curator-spec-worktree`
- CocoaSkills baseline: `/Users/iv/Developer/Wildberries/cocoaskills`
