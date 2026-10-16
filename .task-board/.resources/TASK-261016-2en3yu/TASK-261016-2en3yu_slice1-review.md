# L15-1 slice 1 review handoff

## Scope and identity

The explicit 2026-10-16 scope decision narrows TASK-261016-2en3yu to machine-owned credential-store grammar, CLI, typed selector transport into production resolve and fail-closed refusal of selections without a qualified application. TASK-261016-hzbo83 owns the remaining L15-1b work. This report supersedes the earlier whole-leaf handoff constraint for this bounded slice only. Checklist evidence applies to the six scoped AC groups and their bounded plants below; it does not assert implementation of every original whole-leaf row.

Base: `dccac3199aa5c543c589c129f5ef2bc437523a2a`. Story HEAD and index remain unchanged; all candidate changes are uncommitted. Normative curator-spec main remains `15d970111277b1f68b3f7a4ddd63bc339b2c0a3f`, verified again through the repository API. Read login-readiness sections 2, 2.1, 5; profile-codex-settings sections 1, 3\u20136; and the cited environments sections. No normative conflict was found for the scoped slice. No vendor runtime qualification is claimed.

Hosted candidate: `8c55bff02936d64706bf97f34495aac758e11d2a`; full gate: https://github.com/relux-works/curator/actions/runs/38852843558 . Final workflow status: completed/success. All 20 executed jobs succeeded; two conditional jobs (an optional additional self-hosted test lane and the workflow-dispatch candidate suite) were skipped by workflow conditions. The candidate still matches all 16 source/doc files at handoff preflight; Story HEAD and index are unchanged.

## Behavior

- Manager-config 5 carries `environments.credential_store`; absent means native. Versions 1\u20133 reject the member; version 4 remains unallocated. Version 5 retains version-3 state-hash waiver semantics. Nulls, malformed maps, unknown values and noncanonical environment IDs refuse.
- The user-owned knob cannot be supplied or locked by system config; existing isolation locks still bind.
- `env config set/show/unset` accepts canonical IDs and the established CLI environment aliases. Canonical keys win alias collisions; profile names are not normalized. Set promotes the envelope atomically, preserving schema-1 permissive posture. Unset restores effective native. Editing policy changes no managed home, vendor config or credential.
- `machineFromConfig -> Resolve -> admitCredentialStore` transports the typed selector. Unsupported explicit mappings refuse before provisioning or fragment emission. Codex file uses `codex_settings_unsupported`; other unavailable mappings use `environment_credential_store_unsupported`. Native Codex read/parse/unknown-selector errors remain errors under explicit isolated/file. Native selection retains the existing producer behavior.

## Measured acceptance coverage

**6 of 6 slice AC groups driven** through production entry points:

| Slice AC group | Production call site | Named tests and observed contract |
|---|---|---|
| Successor grammar and frozen envelopes | config.Load/Parse -> parseEnvironmentsVersion -> parseCredentialStore | TestCredentialStoreConfigGrammar; TestCredentialStoreConfigVersionBoundary: valid machine entry beside invalid maps/values/IDs and old-envelope rejection; bytes unchanged on failed read. |
| Machine ownership and isolation lock separation | config.Load -> system merge validation | TestCredentialStoreConfigMachineOwnership; TestSystemV2LockableSubsetIsClassWide: locked/unlocked system selectors refuse; legitimate isolation lock survives. |
| CLI grammar, aliases and profile identity | cli.run -> cmdEnvConfigSet/Show/Unset -> WriteRaw | TestCredentialStoreCLIConfigRoundTrip; TestCredentialStoreCLIConfigRefusesMalformed: set/show/unset (3 of 3 command forms), canonical persistence/collision behavior, malformed edits refused atomically, unset returns native. |
| Envelope promotion without unrelated changes | cli.run -> cmdEnvConfigSet -> WriteRaw | TestCredentialStoreCLIPromotion: valid schema-1 promotion preserves posture; invalid selector leaves envelope bytes unchanged. |
| Typed transport and unsupported-selection refusal | machineFromConfig -> Resolve -> admitCredentialStore; checkIsolatedCredentialStore -> admitCredentialStore | TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery; TestCredentialStorePolicyActivation: CLI repair/non-repair emit no forbidden fragment; all five adapters \u00d7 two explicit selections refuse before provisioning. Native Codex produces a real fragment. |
| Failed native inspection never becomes absence; state preserved | Resolve -> admitCredentialStore -> codexCredentialStoreWith | TestCredentialStorePolicyNativeInspection: unreadable/malformed/unknown native selector refuses; readable valid/absent-key controls reach the intended delivery refusal. CLI tests preserve config, marker, native credential bytes and link identity; round-trip test also preserves a detached synthetic credential and vendor seed. |

Previously committed reviewer tests retain their original names; no test was removed or renamed. Existing native isolation and credential/migration suites remain present. The unchanged-source full envprofile suite was green in earlier evidence; the full hosted gate reruns repository suites.

## Coverage map and declared bounds

Brief gap: the supplied brief contains adapter and paired-test tables but no catalog attack-surface table. This gap is recorded explicitly. The following supplemental catalog map applies to slice 1; it does not claim whole-leaf runtime coverage.

| Surface row | Attacking tests | Killed narrowing mutants | Out-of-contract inputs (AC clause) |
|---|---|---|---|
| text grammar | TestCredentialStoreConfigGrammar, TestCredentialStoreConfigVersionBoundary | old-v3, unknown-auto, persist-alias | Nulls, malformed maps, unknown IDs/values: leaf says invalid; frozen versions: leaf requires allocated successor. |
| selector paths | TestCredentialStoreConfigMachineOwnership, TestCredentialStoreCLIConfigRoundTrip, TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery | system-store, profile-normalization, drop-codex-file | System store selection: leaf says packages/overlays/repositories/system locks never select policy. |
| error surface | TestCredentialStorePolicyNativeInspection | permission-as-absent | No read failure is treated as out of contract; permission/parse failures must refuse. |


## Original leaf rows outside this slice

The acceptance-criteria clause limiting this task is the explicit scope decision: \u201cconfig grammar, CLI, typed transport into resolve, and fail-closed unsupported selections.\u201d The following portions of the original leaf remain in contract for L15-1b, outside acceptance for this slice:

| Original paired row | Slice evidence / follow-up bound |
|---|---|
| 1. Old/system/package/overlay/malformed selector | Config/system/CLI portions exercised here; new package and overlay contribution controls belong to the successor carrier/settings owner. No new package vocabulary is admitted. |
| 2. Shared explicit file versus stale auth | Explicit selection refuses before a fragment or write here. Qualified shared-file admission and stale-auth compatibility matrix are deferred. |
| 3. Native keyring/auto isolation; explicit keychain scope | Existing native isolation tests retained; all new explicit application refuses. Qualified isolated/file and explicit-keychain scope handling deferred. |
| 4. Native failure versus absence | Driven here by TestCredentialStorePolicyNativeInspection, with the explicit permission-read seam stated. |
| 5. Muse native/file literal and readiness | Effective-file policy, four-XDG-plus-literal login/launch and readiness consumers deferred. Existing native behavior is retained, not claimed as rc.7 Muse conformance. |
| 6. Selector/repair ownership migration | Config edits and explicit refusal preserve existing state here. Qualified backend transitions and inspect/plan/apply migration matrix deferred. |
| 7. Alias/profile normalization | Driven here by the CLI round-trip and malformed/collision tests. |
| 8. Foreign source with forced home store | Generic explicit refusal is exercised; source-specific node/helper/broker policy and positive source controls deferred. |

Full original leaf acceptance remains 2 of 8 fully driven paired rows; the task's acceptance is the 6 of 6 slice groups above. No runtime capability, hosted override application, unified effective policy, readiness or login implementation is inferred from schema support.

## Narrowing controls


Every command below runs the behavioral Go suite using a Go source overlay. No source-text checker was added or changed; the extra source-token-preserving static-checker rule is inapplicable. The JSON outcome records relative source paths and exact replacements for replay.

| Mutant | What the weakened gate admits | Named failing test | Result / survivor bound |
|---|---|---|---|
| old-v3 | Only version 3 may carry the new selector | TestCredentialStoreConfigVersionBoundary | Exit 1; killed. |
| unknown-auto | Only the extra machine selector auto | TestCredentialStoreConfigGrammar/unknown_selector | Exit 1; killed. |
| persist-alias | CLI alias accepted directly in persisted machine config | TestCredentialStoreConfigGrammar/alias_id | Exit 1; killed. |
| system-store | Only credential_store bypasses the system subset gate | TestCredentialStoreConfigMachineOwnership | Exit 1; killed. |
| profile-normalization | Profile segment normalized like an environment alias | TestCredentialStoreCLIConfigRoundTrip | Exit 1 on changed profile identity; killed. |
| drop-codex-file | Only Codex file selector dropped from CLI-to-resolve transport | TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery | Exit 1 because a forbidden native fragment is emitted; killed. |
| permission-as-absent | Only native permission failure becomes absent config | TestCredentialStorePolicyNativeInspection/permission | Exit 1 on the wrong refusal class; killed. |

No surviving executed mutants. The two initial CLI mutant attempts exited 1 solely on the host GOROOT lock deadline; those attempts are **unrun controls**, not kills. Sequential retries failed in the named tests above. The specifically required hosted-override-drop plant is **not run**: this slice adds no hosted override transport; drop-codex-file proves only the current CLI composition boundary.


## Validation ledger

Local environment: darwin/arm64, Go 1.25.5, golangci-lint 2.12.2. Synthetic temporary homes only. Source/test/config bytes match `source-identity.json`; every stored mutant overlay was checked against its exact replacement on the same bytes. No changed identity was used to reuse evidence.

| Command | Exit | Evidence / execution |
|---|---:|---|
| go test ./internal/config ./internal/envregistry ./internal/envprofile ./cmd/curator -run '^TestCredentialStore' -count=1 | 0 | Reran this run; focused-22.log. |
| go build ./... | 0 | Reran this run; build-23.log. |
| git diff --check | 0 | Reran this run; diff-check-25.log. |
| go test ./internal/config ./internal/envregistry -count=1 | 0 | Reused exact unchanged-source green evidence, config-registry-02.log. |
| go test ./internal/envprofile -count=1 | 0 | Reused exact unchanged-source green evidence, envprofile-full-10.log (308.506s). |
| go vet ./internal/config ./internal/envregistry ./internal/envprofile ./cmd/curator | 0 | Reused exact unchanged-source green evidence, vet-14.log. |
| GOLANGCI_LINT_CACHE="$PWD/.temp/TASK-261016-2en3yu/lint-cache" golangci-lint run | 0 | Reused exact unchanged-source green evidence, lint-12.log: 0 issues. |
| sh scripts/remote-gate.sh | 0 | Executed directly in this run; remote-gate-21.log; completed/success, 20 successful jobs and two conditional skips. Candidate identity bound above and in TASK-261016-2en3yu_hosted-gate.json. |

Earlier diagnostics remain honest bounds: the initial grammar test exited 1 expected red before implementation; early CLI fixture attempts exited 1 before correction; two mutant attempts hit host toolchain lock deadlines and were unrun, then sequential retries failed in named tests. The earlier full local CLI suite exited 143 after intentional termination and supplies no full-suite verdict. The earlier remote gate exited 137 and its workflow was cancelled; it supplies no green verdict and is replaced only after this run's real gate result. None of these failed commands are presented as passing.

## Paths and self-review

All code changes are inside the declared config, registry, managed-home and CLI modules plus corresponding focused tests and docs/environment-config.md. README.md is the sole extra path: it adds the required focused tool command and links the documented runtime bounds. No existing reviewer test is removed or renamed. This is continuation of the first implementation under an explicit scope split, not review rework; no repeat-of class was supplied.

## Logbook

- Scope ownership was resolved explicitly on 2026-10-16: this task owns slice 1 and TASK-261016-hzbo83 owns the remaining L15-1b work. The earlier whole-leaf handoff refusal is preserved as historical evidence, not treated as a test failure or erased.
- Normative main and all 16 candidate source/doc files match the earlier tested identity. Focused selector tests, build and whitespace validation were rerun in this run; exact unchanged evidence was reused elsewhere as required.
- The hosted-drop-override control belongs to deferred hosted transport. This slice executes the nearest real composition narrowing plant, drop-codex-file: weakening only Codex file transport makes TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery fail because the old native fragment would otherwise escape.
- No signing/landing/branch operation is performed on the managed Story branch. The repository's required remote gate produces its detached throwaway snapshot; integration is the orchestrator's responsibility.

- The full remote gate returned exit 0 in this run. All 20 executed workflow jobs succeeded. The gate snapshot and all candidate bytes were verified again immediately before attaching review evidence. Required runtime qualification remains deferred; a green repository gate is not vendor qualification.
