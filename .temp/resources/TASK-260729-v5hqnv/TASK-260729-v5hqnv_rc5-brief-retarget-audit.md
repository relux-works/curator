# rc.5 brief retarget — exact before/after board audit

**Board task:** `TASK-260729-v5hqnv` (parent `STORY-260720-1uv5gi`)
**Date:** 2026-07-29
**Scope executed:** board `description`, `scope`, `ac`, `notes` and dependency edges for exactly the seven briefs named by the accepted parity map. No CocoaSkills, Curator, curator-spec, spec, pin, git or product changes.

---

## 1. Authoritative inputs re-resolved

| Input | Resolved value |
| --- | --- |
| Accepted parity map | `TASK-260729-1t1z2l_curator-go-to-csk-parity-delta.md` revision 2, §§3, 5, 7, 8 |
| rc.5 conformance root | `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1` |
| Root provenance | curator-spec worktree, base commit `57c1f56846d221ecc55786bd3c2467ec32f11730`, README `**Version:** 1.0.0-rc.5` |
| Manifest | `447` files, `protocol_version 1.0.0-rc.5`, `shasum -a 256 conformance/v1/manifest.json` = `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c` |
| Pin self-consistency | `release/1.0.0-rc.5.json` `candidate_protocol_pin.manifest_sha256` and `downstream_consumption.required_manifest_sha256` both equal `sha256:b6f56aac…`; `committed_release_pin_advanced: false` |
| Producing task | `TASK-260729-3nx97g` `regenerate-rc5-build-driver-goldens` — status **`done`**, reviewer verdict **ACCEPTED** |

### 1.1 The §3.3 build-driver gap is closed

The accepted parity map recorded that the then-canonical rc.5 snapshot (`TASK-260728-2kp3tv`, 422 files, manifest `9ba9b8ec…`) published **zero** `build-driver` entries, and that `TASK-260720-2dnqw2`, `TASK-260720-12r55p` and `TASK-260720-3pemm6` therefore had no shared byte-exact goldens.

That is no longer true. The root inspected here publishes `vectors/build-drivers.json` and a complete 11-file `expected/build-driver/` tree, and the manifest carries 12 `build-driver` entries.

Independently re-derived in this task, from the published bytes only:

| Artifact | Bytes | SHA-256 | Declared value |
| --- | ---: | --- | --- |
| `expected/build-driver/build-input.ccj.json` | 869 | `529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b` | matches `cache-key.txt` |
| `expected/build-driver/receipt.ccj.json` | 1120 | `919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd` | matches `receipt-sha256.txt` |

Receipt bytes end on `}` — no trailing newline. `vectors/build-drivers.json` `cache_identity` sets `aliases: false` and carries three numerically distinct keys: `portable` (`529370…`, `schema_valid: true`, `execution_policy: manager-worker-v1`), `legacy_rc4_without_execution_policy` (`3fcd714a…`, `schema_valid: false`, no `execution_policy`), `reserved_hardened` (`13736230…`, `schema_valid: false`, `hardened-worker-v1`, owner `STORY-260728-327soo`).

### 1.2 Vector inventory used to write the acceptance criteria

| Vector / schema | Facts used |
| --- | --- |
| `vectors/build-drivers.json` | 5 argv forms (`telemetry-off`, `version`, `env`, `list`, `build`); 8 `positive_cases`; **77** `rejection_cases`; 10 `build_source_cases`; 12 `toolchain_cases`; `fixed_environment`; `cache_identity` |
| `vectors/go-host-execution-policy.json` | `protocol_version 1.0.0-rc.5`; 18 `mandatory_controls`; 13 `session_states`; 4-node `process_graph`; 14 `identity_and_protocol_cases`; 8 `package_influence_cases`; 11 `capability_evidence_cases`; 6 `deferred_hardened_guarantees` + 6 `deferred_capability_rejection_guards`; `failure_boundary`; `capability_evidence_record` |
| `native_control_inventory` | `rc5-native-control-inventory-v1`, `exhaustive: true`, `platforms: ["macos","windows"]`, `probe_scope: per-operation`, `probe_timing: pre-worker-launch`, 5 controls with honest per-platform availability |
| `capability_evidence_record` | `record_version capability-evidence-v1`; `result_only: true`; `entry_cardinality exactly-one-per-inventory-control`; `exposed_in [dry-run-plan-result, install-result, status-result]`; `excluded_from [cache-key, conformance-claim, install-marker, receipt]`; 8 `consistency_rules` |
| `failure_boundary` | `missing_mandatory_portable_control` → `build_execution_control_unavailable`, `fails_before: worker-launch`, `rejects_build: true`, `published: false`; `unavailable_inventory_native_control` and `missing_deferred_hardened_capability` → do not reject, still publish |
| `schemas/v1/conformance-claim-v3.schema.json` | `protocol_version` const `1.0.0-rc.5`; required includes `build_drivers` |
| `vectors/conformance-claim-v3-qualification.json` | linux `excluded` `until_task TASK-260728-1skseh`; macos/windows `pending-downstream-native-evidence`; rules `schema-valid-is-not-qualified`, `driver-platform-subset`, `no-generic-driver` (`go-repository-v1`, `go-v1`), `no-unevidenced-platform` |
| `protocol/core.md` §4.2.1 | `manager-worker-v1` MUST be implemented **on macOS and Windows**; one worker session performs exactly one `go list` and exactly one `go build` |
| `profiles/manager.md` §—, `protocol/core.md` | Go **1.23** is the protocol floor; accepted Curator composite `internal/godriver/session.go:42` allows family **`1.25`** only |
| `expected/build-driver/marker.json` | `schema_version 2`, `skill_schema_version 6` — the compiled-build marker stays **install-marker-v2** |
| `schema-cases/build-receipt-v2/*`, `install-marker-v3/*` | driver `go-repository-v1` — external-repository line, **not** the go-v1 surface |

---

## 2. Per-brief before → after

### 2.1 `TASK-260720-2dnqw2` — canonical-build-metadata

| Item | Before | After |
| --- | --- | --- |
| Cache key in AC | `sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48` (asserted as **the** shared key) | `sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b`; `3fcd714a…` retained only as a required schema-invalid non-alias negative |
| Receipt hash in AC | `sha256:750f5f7557ffe1dac1ec06b7d47cb5976e390a7dc7d57ea616c9edda91013c11` | `sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd` (`750f5f75…` removed entirely — it does not exist in rc.5) |
| Execution policy | absent | `policy.execution_policy = manager-worker-v1` required; full rc.5 `goBuildPolicyV1` member list enumerated in scope |
| Reserved hardened | absent | `sha256:13736230d33ce59de7f7323dcd4cffd510655ad8dabd5ee9e8b6cb182ec70037` added as a required non-alias negative |
| Byte anchors | none | 869-byte `build-input.ccj.json`, 1120-byte `receipt.ccj.json` |
| Marker | "marker schema 2" | unchanged **by evidence**: golden `marker.json` is `schema_version 2`; `install-marker-v3`/`build-receipt-v2` explicitly declared out of scope (go-repository-v1 line) |
| Claim surface | undecided in the parity map | **decided**: claim emission is out of this task and stays with `TASK-260720-12r55p`, which targets **claim v3** |
| Root | implicit | caller-supplied root with manifest `sha256:b6f56aac…`, digest recorded as non-release evidence |
| `blockedBy` | `[3c0ss2, 3j8pp5]` | `[3c0ss2, 3j8pp5, TASK-260729-3nx97g]` |

### 2.2 `TASK-260720-2g21eg` — go-v1-compile-driver

| Item | Before | After |
| --- | --- | --- |
| Process boundary | direct `go list`/`go build` from the manager | fixed 4-node graph: manager parent → identity-verified manager-owned worker → fingerprinted `GOROOT/bin/go` → fingerprinted `GOROOT/pkg/tool/<GOHOSTOS>_<GOHOSTARCH>` children |
| argv ownership | "the five argv forms" run by the driver | **split by process**: `telemetry-off`/`version`/`env` stay parent-side and are owned by `TASK-260720-3j8pp5`; only `list` and `build` are issued by the worker, exactly one each per session |
| Session model | absent | 13 `session_states` driven end to end; no retry, second list, second build, replayed nonce, out-of-order, oversize or unknown message |
| Controls | ad hoc | all 18 `mandatory_controls` enforced always, named |
| Native controls | absent | per-operation pre-launch probe against `rc5-native-control-inventory-v1`, exhaustive over macos+windows, exactly 5 controls, no control outside the inventory, host label is not a probe |
| Capability evidence | absent | exactly one `capability-evidence-v1` record per operation, exact record and entry fields, `exposed_in`/`excluded_from` enumerated, all 8 `consistency_rules` with `build_execution_capability_evidence_invalid` / `build_execution_hardened_claim_forbidden` |
| Failure boundary | absent | missing mandatory control → `build_execution_control_unavailable` before `worker-launch`, rejects, publishes nothing; unavailable inventory control and missing deferred hardened capability → neither rejects nor blocks publication |
| Hardened guarantees | absent | all 6 deferred guarantees refused by their 6 rejection guards, never claimed |
| Negative clusters | prose list | + all 14 `identity_and_protocol_cases`, all 8 `package_influence_cases`, all 11 `capability_evidence_cases` |
| Platform | implicit any host | macOS and Windows; every other host fails closed with an unavailable-control error and never starts a worker |
| Fixture gate | "a valid real Go fixture build" | same, explicitly on a native macOS or Windows host |

### 2.3 `TASK-260720-12r55p` — shared-v6-vector-consumer

| Item | Before | After |
| --- | --- | --- |
| Protocol target | "protocol **rc.4**" (literal) | protocol **1.0.0-rc.5** |
| Root | unpinned candidate suite | `CURATOR_CONFORMANCE_ROOT` whose `manifest.json` must hash to `sha256:b6f56aac…`, cross-checked against `release/1.0.0-rc.5.json` |
| Claim coverage | **claim v2** | **conformance-claim-v3** (`protocol_version` const `1.0.0-rc.5`, required `build_drivers`); separation asserted across v1/v2/v3; no claim emitted by this task |
| New required vectors | `build-drivers.json`, `manager-lifecycle.json` | + `go-host-execution-policy.json`, `conformance-claim-v3-qualification.json`, the 11-file `expected/build-driver/` tree |
| Negative minimum | prose clusters | exact counts: 77 rejection cases, 10 build-source, 12 toolchain, 18 controls, 14 identity/protocol, 8 package-influence, 11 capability-evidence, 6 rejection guards, 3 failure-boundary outcomes |
| Identity assertions | none | `aliases false`, three distinct keys with `529370…` valid and `3fcd714a…`/`13736230…` invalid |
| Platform qualification | absent | linux `excluded until TASK-260728-1skseh`; macos/windows `pending-downstream-native-evidence`; 4 qualification rules; allowed drivers `go-repository-v1`, `go-v1` |
| Hard-coding shortcut | not addressed | explicitly forbidden — must not mirror Curator `internal/godriver/controls_test.go`, which reconstructs the vector input in code |
| Out of scope added | — | `build-receipt-v2`, `install-marker-v3`, schema v7, external-repository and registry surfaces |
| `blockedBy` | `[th0jdi, 3ag6pi]` | `[th0jdi, 3ag6pi, TASK-260729-3nx97g]` — `3ag6pi` **deliberately retained**, see §4 |

### 2.4 `TASK-260720-akf5kh` — schema-v6-user-docs

| Item | Before | After |
| --- | --- | --- |
| Go family claim | "Go **1.23-plus accepted-family**" | protocol floor is 1.23; csk accepts only the operator-trusted family recorded in handoff evidence, **family `1.25`** in the currently accepted Curator implementation |
| Execution policy | undocumented | portable `manager-worker-v1` as a normative cache/receipt/marker/claim input, never an option or operator preference; 4-node graph; pre-launch and post-exec identity verification; frozen snapshot integrity; worker-domain teardown; never-run artifact |
| Capability evidence | undocumented | one `capability-evidence-v1` record per operation, one entry per inventory control, honest per-platform availability incl. `active-process-count-limit`/`aggregate-memory-limit` available on Windows via job objects but unavailable on macOS, and `per-file-size-limit` available on macOS via `RLIMIT_FSIZE` but unavailable on Windows |
| Reject semantics | undocumented | unavailable inventory control does **not** reject; missing mandatory portable control **does** |
| Platform claim | absent | supported platforms are macOS and Windows; go-v1 fails closed elsewhere; Linux explicitly deferred with owners `TASK-260728-1skseh` and `TASK-260728-1e6811` |
| Hardened guarantees | absent | none of the six may be documented as provided |

### 2.5 `TASK-260720-3pemm6` — cross-platform-go-build-e2e

| Item | Before | After |
| --- | --- | --- |
| Platform claim | "install … on **Linux, macOS, and Windows**"; "CI runs the real fixture on **ubuntu, macOS, and Windows**" | real-fixture go-v1 runs on **native macOS and native Windows**; ubuntu keeps portable non-driver coverage only |
| Linux assertion | green go-v1 build required | ubuntu must **prove the source-aware path fails closed** with an unavailable-control error and no worker launch; Linux driver success neither required nor claimed anywhere |
| Linux ownership | absent | routed to `TASK-260728-1skseh` and `TASK-260728-1e6811`, explicitly out of scope |
| Go setup | "accepted Go **1.23-or-newer** release family" | operator-trusted accepted family from handoff evidence (`1.25` today); 1.23 is the protocol floor only |
| Rebuild dimensions | source, toolchain, target, policy, receipt, artifact | same, with **execution-policy** named explicitly |
| Evidence | "records the suite digest" | records the **full manifest digest** |
| Capability evidence | absent | per-platform evidence must match `rc5-native-control-inventory-v1` exactly on each native runner |

### 2.6 `TASK-260720-3s27te` — integrated-csk-v6-verification

| Item | Before | After |
| --- | --- | --- |
| Conformance root | "the immutable curator-spec **rc.4** conformance root" (literal) | immutable curator-spec **1.0.0-rc.5** candidate root via `CURATOR_CONFORMANCE_ROOT`, manifest `sha256:b6f56aac…` |
| pytest gate | "passes with the **rc.4** conformance root" | "passes against the **rc.5** candidate conformance root" |
| CI gate | "the **Linux, macOS, and Windows** CI matrix with the real Go fixture is green" | native macOS and native Windows matrix green with the real fixture; ubuntu green for portable non-driver coverage and the fail-closed assertion; Linux driver success neither required nor claimed |
| Report contents | SHAs, commands, logs, platform evidence | + full rc.5 manifest digest; + execution-policy identity, native-control-inventory and capability-evidence cases in the criterion matrix |
| Claim emission | not stated | explicitly forbidden — no conformance claim emitted, no committed pin moved |

### 2.7 `TASK-260720-th0jdi` — build-currentness-repair-gc

| Item | Before | After |
| --- | --- | --- |
| Currentness dimensions | ref, content, roots, source, toolchain, target, cache key, boundary, receipt, artifact, marker, shim | + **execution policy**, stated explicitly as a currentness/rebuild dimension |
| Legacy keys | absent | a build under `sha256:3fcd714a…` (no `execution_policy`) or `sha256:13736230…` (reserved hardened) is non-current and is **rebuilt, not adopted** |
| Non-current triggers | missing, corrupt, wrong-target, wrong-toolchain, unsupported, context-leaking, untrusted | + **wrong-policy** |
| Capability evidence | absent | surfaced in status output but **never** a currentness input (`result_only`, excluded from all hashed identities) |
| Mechanism note | — | clarification only: already covered transitively by the complete cache-key/receipt comparison; no separate mechanism added |

### 2.8 Field-level delta table

| Task | Field | Before chars | After chars | Changed |
| --- | --- | ---: | ---: | --- |
| `TASK-260720-2dnqw2` | description | 192 | 244 | yes |
| `TASK-260720-2dnqw2` | scope | 523 | 1218 | yes |
| `TASK-260720-2dnqw2` | ac | 719 | 1935 | yes |
| `TASK-260720-2dnqw2` | blockedBy | 2 | 3 | yes |
| `TASK-260720-2dnqw2` | blocks | 1 | 1 | no |
| `TASK-260720-2g21eg` | description | 188 | 344 | yes |
| `TASK-260720-2g21eg` | scope | 479 | 1672 | yes |
| `TASK-260720-2g21eg` | ac | 939 | 3999 | yes |
| `TASK-260720-2g21eg` | blockedBy | 2 | 2 | no |
| `TASK-260720-2g21eg` | blocks | 1 | 1 | no |
| `TASK-260720-12r55p` | description | 164 | 197 | yes |
| `TASK-260720-12r55p` | scope | 704 | 1383 | yes |
| `TASK-260720-12r55p` | ac | 928 | 2401 | yes |
| `TASK-260720-12r55p` | blockedBy | 2 | 3 | yes |
| `TASK-260720-12r55p` | blocks | 2 | 2 | no |
| `TASK-260720-akf5kh` | description | 204 | 273 | yes |
| `TASK-260720-akf5kh` | scope | 652 | 882 | yes |
| `TASK-260720-akf5kh` | ac | 732 | 2302 | yes |
| `TASK-260720-akf5kh` | blockedBy | 2 | 2 | no |
| `TASK-260720-akf5kh` | blocks | 1 | 1 | no |
| `TASK-260720-3pemm6` | description | 179 | 257 | yes |
| `TASK-260720-3pemm6` | scope | 825 | 1474 | yes |
| `TASK-260720-3pemm6` | ac | 918 | 1383 | yes |
| `TASK-260720-3pemm6` | blockedBy | 1 | 1 | no |
| `TASK-260720-3pemm6` | blocks | 3 | 3 | no |
| `TASK-260720-3s27te` | description | 203 | 257 | yes |
| `TASK-260720-3s27te` | scope | 529 | 879 | yes |
| `TASK-260720-3s27te` | ac | 662 | 1124 | yes |
| `TASK-260720-3s27te` | blockedBy | 2 | 2 | no |
| `TASK-260720-3s27te` | blocks | 9 | 9 | no |
| `TASK-260720-th0jdi` | description | 234 | 252 | yes |
| `TASK-260720-th0jdi` | scope | 498 | 832 | yes |
| `TASK-260720-th0jdi` | ac | 896 | 1376 | yes |
| `TASK-260720-th0jdi` | blockedBy | 1 | 1 | no |
| `TASK-260720-th0jdi` | blocks | 2 | 2 | no |

---

## 3. Dependency changes

| Edge | Action | Reason |
| --- | --- | --- |
| `TASK-260720-2dnqw2` blockedBy `TASK-260729-3nx97g` | **added** | Its AC now hard-codes the rc.5 goldens `529370…` / `919fbbad…`, which exist only because `3nx97g` regenerated them. `3nx97g` is `done`, so the edge adds provenance without changing scheduling. |
| `TASK-260720-12r55p` blockedBy `TASK-260729-3nx97g` | **added** | It is the byte-exact consumer of `vectors/build-drivers.json` and `expected/build-driver/`. Same rationale. |
| `TASK-260720-12r55p` blockedBy `TASK-260720-3ag6pi` | **retained deliberately** | See §4. |
| all other edges on the seven briefs | **unchanged** | The parity map's §12 recommendation stands: retargets are contract corrections, not restructuring. |

No edge was removed. No task was created, renamed, deleted, re-parented or status-changed. Checklists were left untouched — the task scope is description, scope, AC and dependencies, and the AC governs.

---

## 4. Deliberately fail-closed — remaining prerequisites

Per the task instruction "if a claimed artifact or decision is not actually present, leave the brief fail-closed and document the remaining prerequisite rather than inventing it":

1. **`TASK-260720-3ag6pi` is still an rc.4 gate.** It is `protocol-v6-conformance-verification`, status **`blocked`**, and still scoped to the literal rc.4 line. The parity map §5 permits relinking `12r55p` to "the replacement rc.5 verification task" — **no such task exists on the board today**. One was not invented. The hard edge is retained and the condition is recorded in `TASK-260720-12r55p` notes: before that task starts, the Curator-side owner must either retarget and re-review `3ag6pi` against the rc.5 candidate root, or create the rc.5 verification gate and relink the edge.
2. **`TASK-260720-jrrgw9` is still named `verify-rc4-build-conformance`** and is `development`. It gates `TASK-260720-1pvfj5`, which hard-blocks both CocoaSkills roots `z9j4c9` and `z2z795`. Whichever suite it verifies is what CocoaSkills inherits. Retargeting it is Curator-side and outside this task's stated scope.
3. **The rc.5 root is a candidate, not a release.** `release/1.0.0-rc.5.json` records `committed_release_pin_advanced: false`, and `TASK-260729-3nx97g` authorizes no landing, tagging, signing, publication or pin change. Every retargeted brief therefore consumes the root through a caller-supplied `CURATOR_CONFORMANCE_ROOT` and records the digest as non-release evidence.
4. **Windows is reachable but not build-ready.** The parity map §7.3 recorded `ssh win 'go version'` exit 1 with no Go on PATH. `3pemm6` and `3s27te` now require native Windows gates, so that prerequisite must be closed by the runner-readiness work (`TASK-260729-1bf72u`, in flight) before those tasks can pass.

Both `3nx97g`'s own notes and this audit agree that retargeting `3ag6pi` and `jrrgw9` is the next authority-owned step; it was intentionally not performed here.

---

## 5. Completeness verification

| Check | Result |
| --- | --- |
| All seven named briefs updated | yes — `2dnqw2`, `2g21eg`, `12r55p`, `akf5kh`, `3pemm6`, `3s27te`, `th0jdi`; `description`, `scope`, `ac` changed on all seven |
| Task IDs preserved | yes — no create/delete/rename/re-parent; all seven remain `backlog` under `STORY-260720-1uv5gi` |
| Residual stale-target wording | none. Remaining `rc4`/`3fcd714a` strings appear only as required negative-case identifiers (`legacy_rc4_without_execution_policy`); remaining `ubuntu` strings appear only in the new portable-non-driver / fail-closed clauses. Verified by regex scan over all three fields of all seven briefs. |
| `750f5f75…` (rc.4 receipt) | removed from the board entirely |
| Other ten briefs + parent story scanned | `z9j4c9`, `3c0ss2`, `2jfnz6`, `8nxlgx`, `z2z795`, `11yhth`, `3t8nr3`, `g7kgox` and `STORY-260720-1uv5gi` are clean of `rc.4`, stale hashes, `claim v2`, ubuntu/Linux platform claims and Go-version claims. Two hits inspected and confirmed **correct as-is**: `TASK-260720-3j8pp5` ("run no `go list` or `go build`"; "Pre-Go-1.23 … fail closed") matches the rc.5 session model, where the parent runs the package-independent probe and the worker runs the source-aware forms; `TASK-260720-2x6mjn` ("never runs `go list` or `go build`") is the compiler-free dry-run invariant, unchanged by rc.5. Neither needed an edit, and neither was edited. |
| Sibling in-flight tasks checked for conflict | `TASK-260729-1b9tc3`, `TASK-260729-1bf72u`, `TASK-260729-35tb37` are all read-only design/inventory tasks under the same story — no mutation overlap; they consume these retargeted briefs |
| Board DAG | `plan(STORY-260720-1uv5gi, mode=related, active=true)` resolves; critical path unchanged in shape: `jrrgw9 → 1pvfj5 → z9j4c9 → 3c0ss2 → 2dnqw2 → 2jfnz6 → 8nxlgx → 11yhth → 3t8nr3 → g7kgox → th0jdi → 12r55p → 3pemm6 → 3s27te → …` |
| Product/source/git state | untouched — no CocoaSkills, Curator, curator-spec, spec, pin, tag, commit or stage operation was performed |

### 5.1 Corrections made during this task

One self-correction: the first `2g21eg` revision stated that **all five** argv forms are issued by the worker. `protocol/core.md` §4.2.1 and `go-host-execution-policy.json` `session_states` show the parent runs `parent-package-independent-toolchain-probe` (telemetry-off, version, env) and the worker runs `worker-fixed-go-list` and `worker-fixed-go-build`. The brief was re-issued with the split stated explicitly and the three probe forms attributed to `TASK-260720-3j8pp5`, which already owns them. The final board text is the corrected version.

---

## 6. Raw evidence

- `TASK-260729-v5hqnv_before.jsonl` — exact pre-change board projection of the seven briefs (`id`, `name`, `title`, `description`, `scope`, `ac`, `blockedBy`, `blocks`, `status`).
- `TASK-260729-v5hqnv_after.jsonl` — exact post-change projection, same fields, machine-diffable against the before file.
