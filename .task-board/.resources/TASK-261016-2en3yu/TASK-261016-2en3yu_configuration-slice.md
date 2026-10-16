# L15-1 configuration slice: evidence and remaining scope

This candidate implements the first coherent part explicitly allowed by gw-L151-brief.md. It does not implement or claim acceptance of the whole L15-1 leaf.

## Identity and normative sources

- Curator base: `dccac3199aa5c543c589c129f5ef2bc437523a2a`, equal to refreshed origin/main at preflight; Story HEAD remains unchanged and the candidate is uncommitted.
- Normative curator-spec main: `15d9701` (fetched before implementation). Read login-readiness sections 2, 2.1, 5; profile-codex-settings sections 1, 3–6; and cited environments sections.
- Hosted candidate: `3786727d3d342e9e3b712db6466c3ed9248fa7f0`. All 16 changed/new source and documentation files were compared byte-for-byte with that snapshot.
- Full-gate run: https://github.com/relux-works/curator/actions/runs/38851138925 — terminal status completed/cancelled after the partial-scope handoff was refused; no passing full-gate verdict claimed.
- Local platform: darwin/arm64, Go 1.25.5; golangci-lint 2.12.2. Synthetic temporary homes only; no vendor login, model call, or runtime qualification performed.

## Implemented slice

`config.Load`/`Parse` accept the allocated manager-config version 5 selector subtree and reject the new member in frozen versions 1–3. Version 4 stays unallocated. Version 5 preserves version-3 hash-version waiver semantics. Nulls, malformed maps, unknown selectors and noncanonical environment IDs refuse. Neither locked nor unlocked system configuration may carry the selector; isolation locks remain binding.

`cli.run -> cmdEnvConfig` drives set/show/unset. CLI aliases normalize only environment positions, with canonical keys winning alias collisions. Profile names remain unchanged. Unset restores effective native and changes no home. Promotion from version 1 preserves its permissive posture. Failed edits preserve machine config bytes.

`machineFromConfig -> Resolve -> admitCredentialStore` carries a typed machine selector. Explicit stores cannot silently disappear into an old native fragment. Codex file refuses `codex_settings_unsupported` pending qualification and strict delivery; unavailable forced mappings refuse `environment_credential_store_unsupported`. Native Codex read/parse/unknown-selector failures remain failures even with explicit isolated/file. The same activation guard protects home assembly through `checkIsolatedCredentialStore`. This guard is not the eventual resolved effective policy object.

The tests preserve config, marker, vendor seed, link identity and synthetic credential bytes where those states exist. No credential migration, deletion or vendor config rewrite was added.

## Whole-leaf acceptance coverage

**2 of 8 paired AC rows driven** (rows 4 and 7); 6 rows remain partially exercised or deferred. This ratio is not whole-leaf acceptance.

| Leaf paired row | Production call site and named evidence | Status / bound |
|---|---|---|
| 1. Old envelope/system/package/overlay/malformed selector | `config.Load/Parse/WriteRaw`, `cmdEnvConfigSet`; TestCredentialStoreConfigGrammar, TestCredentialStoreConfigVersionBoundary, TestCredentialStoreConfigMachineOwnership, TestCredentialStoreCLIConfigRefusesMalformed, TestCredentialStoreCLIPromotion | Machine-config and system portions exercised; dedicated package/overlay controls not supplied by this slice. |
| 2. Shared explicit file resurrects stale native auth | `Resolve -> admitCredentialStore`; TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery | Explicit delivery refuses wholesale. Compatible explicit shared-file admission and stale-file matrix remain deferred. |
| 3. Native auto/keyring isolated; explicit keyring scoped per home | `Resolve -> checkIsolatedCredentialStore`; existing TestCodexIsolatedAdmission and new TestCredentialStorePolicyActivation | Existing native refusals retained. Qualified explicit isolated/file positive control and explicit-keychain isolation ordering remain deferred. |
| 4. Native read/parse failure becomes absence | `Resolve -> admitCredentialStore -> codexCredentialStoreWith`; TestCredentialStorePolicyNativeInspection; existing TestCodexIsolatedAdmission, TestCodexUnreadableConfigFails | Exercised. Permission-error seam is explicit; readable absent-key positive control and malformed/unknown neighbors retained. Does not qualify native runtime precedence. |
| 5. Muse native/unset omits literal or uses global Keychain readiness | No new production Muse transport or readiness consumer | Deferred; existing native producer behavior intentionally remains pending the subsequent slice. |
| 6. Selector switch/repair changes ownership | `cmdEnvConfigSet/Unset`, `Resolve`; TestCredentialStoreCLIConfigRoundTrip, TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery; existing credential/migration suites retained | Config-only preservation and activation refusal exercised. Full selector-driven backend migration matrix remains deferred. |
| 7. Alias persisted or profile normalized | `cli.run -> cmdEnvConfig -> normalizeEnvKnob/normalizeEnvKnobValue -> WriteRaw`; TestCredentialStoreCLIConfigRoundTrip, TestCredentialStoreCLIConfigRefusesMalformed | Exercised, including whole-map canonical-wins collision and profile named codex. |
| 8. Foreign source beside forced home store | No adopted node/helper/broker source consumer added | Deferred. Generic explicit-store refusal is not evidence of a source-specific ownership contract. |

All three configuration command forms are driven (3 of 3: set/show/unset); env resolve and the actual native fragment producer are driven by TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery. New tests remain in the uncommitted candidate for the Change Request snapshot.

## Coverage map and out-of-contract declarations

Brief gap: the supplied brief has adapter and paired-test tables but no attack-surface table drawn from the catalog. This does not waive whole-leaf coverage. The following map covers only the implemented slice.

| Surface row | Attacking tests | Killed narrowing mutants | Out-of-contract inputs (AC clause) |
|---|---|---|---|
| text grammar | TestCredentialStoreConfigGrammar, TestCredentialStoreConfigVersionBoundary | old-v3, unknown-auto, persist-alias | Nulls, malformed maps, unknown IDs/values: leaf says invalid; frozen versions: leaf requires allocated successor. |
| selector paths | TestCredentialStoreConfigMachineOwnership, TestCredentialStoreCLIConfigRoundTrip, TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery | system-store, profile-normalization, drop-codex-file | System store selection: leaf says packages/overlays/repositories/system locks never select policy. |
| error surface | TestCredentialStorePolicyNativeInspection | permission-as-absent | No read failure is treated as out of contract; permission/parse failures must refuse. |

Unsupported Claude forced stores, Muse keychain and Pi/OpenCode non-native mappings are refused per the leaf's adapter table. Qualified Codex explicit stores, Muse file/default, external sources and the eventual unified effective policy remain **in contract and deferred**, not classified as out of contract.

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

Logs are task-scoped under `.temp/TASK-261016-2en3yu/`; raw logs remain local because failing synthetic fixtures include temporary paths. Exact commands and real statuses:

| Command | Exit / evidence |
|---|---|
| `go test ./internal/config -run '^TestCredentialStoreConfig' -count=1` | 1 expected red before implementation (config-red-01.log). |
| `go test ./internal/config ./internal/envregistry -count=1` | 0 (config-registry-02.log). |
| `go test ./cmd/curator -run '^TestCredentialStore' -count=1` | Initial attempts 1 on incomplete fixture setup (cli-focused-03/04.log); corrected attempt 0 (cli-focused-06.log). |
| `go test ./internal/envprofile -run '^TestCredentialStore' -count=1` | 0 (policy-focused-05.log). |
| `go test ./internal/config ./internal/envregistry ./internal/envprofile ./cmd/curator -run '^TestCredentialStore' -count=1` | 0 (focused-07.log). |
| `go test ./internal/envprofile -count=1` | 0, 308.506s (envprofile-full-10.log); existing reviewer tests retained under their original names. |
| `go test ./cmd/curator -count=1` | 143, deliberately terminated after exceeding the useful local bounded scope; **unrun as a full verdict**, no pass claimed (cli-full-11.log). Hosted gate is the required whole-repository lane. |
| `go build ./...` | 0 both before and after lint changes (build-08/15.log). |
| `go vet ./internal/config ./internal/envregistry ./internal/envprofile ./cmd/curator` | 0 (vet-14.log). |
| `golangci-lint run` | 1 initially: two new style findings plus stale shared-cache references (lint-09.log). Both new findings corrected. |
| `GOLANGCI_LINT_CACHE="$PWD/.temp/TASK-261016-2en3yu/lint-cache" golangci-lint run` | 0, 0 issues (lint-12.log). |
| `git diff --check` | 0 (diff-check-16/17.log). |
| `sh scripts/remote-gate.sh` | 137: deliberately stopped while the run still had 3 active and 6 queued jobs after about 14 minutes; **unrun as a full verdict**. Its hosted workflow was cancelled and its owned gate branch deleted (remote-gate-13.log, remote-cancel-19.log, gate-cleanup-20.log). No failure cause or passing full-gate result is inferred from queued jobs. |

Mutant commands have the common form `go test -overlay=.temp/TASK-261016-2en3yu/mutants/<name>/overlay.json <package> -run '<test>' -count=1`; each definition outcome supplies package/test. Each executed kill exited 1 as expected, never presented as green. No previous attached green evidence was substituted for a command run here.

## Paths, self-review and continuation

Changed paths stay within the named config/registry/managed/CLI modules, their focused tests, and docs/environment-config.md. New credential_store.go/credential_policy.go files keep policy-related additions separate from shared managed.go. The only additional path is README.md: required tools documentation links the focused command and runtime limits. No previously committed reviewer test was removed or renamed. This is the first implementation, not rework; no repeat-of class was supplied.

Remaining requirements: a single resolved source/store/isolation policy used by verification/readiness/login/producer; managed Muse effective-file mapping with four XDG paths and literal at login/launch; shared native-file compatibility and guidance; qualified Codex file/keychain admission with one v5 typed payload and strict consumers; source-specific foreign-source refusals; complete backend/link transition evidence; the hosted override-drop plant. Runtime qualification has not been performed or inferred from schema support.

## Logbook

- The task initially lacked an estimate; the required first status mutation refused. The leaf's size 5 estimate was applied, then development started.
- No disagreement with the normative main text was found. It explicitly gates Codex enablement on runtime qualification and strict consumers; this candidate keeps that boundary closed.
- The brief explicitly allows a first coherent part, but the task still carries the unchanged whole-leaf DoD. In particular the all-eight-rows and hosted-plant items cannot truthfully be checked. No acceptance criterion was removed or silently weakened. A scope decision is needed before an honest partial handoff can satisfy that guard: split the remaining leaf work into tracked scope and explicitly scope this candidate's checklist, or retain the original leaf and continue the remaining implementation/dependencies before handoff. Recommended: split and review the bounded config slice. Candidate code remains preserved and uncommitted.

## Handoff refusal and exact decision

`task-board handoff TASK-261016-2en3yu --role developer` exited **1**: `unchecked checklist items [1 2 4 5 10]`; the tool also warned that it could not compare board content for the run and recorded no progress for that command. No handoff or Change Request acceptance is claimed. Items 1, 2, 5 and 10 still require whole-leaf implementation/controls; item 4 lacks the required green full gate. Checked items describe only facts evidenced above; no original checklist item was deleted or rewritten.

Constraint: the existing task's unchanged whole-leaf handoff gate cannot admit the brief-permitted first coherent slice. Failed attempt: the real handoff rejected the five unchecked items. Implementation tests and seven bounded controls do not establish the remaining runtime or hosted behavior.

Options: (1) explicitly split the remaining in-contract leaf work and scope the current task/DoD to the configuration slice, then run its required gate and review this candidate; or (2) retain the full leaf scope and continue the listed policy/producer/login/qualification dependencies before review. Recommendation: option 1 keeps the change reviewable without weakening runtime qualification. Exact input needed: the scope owner's choice and corresponding tracked scope/DoD update. No criteria will be removed merely to make handoff green.

The full-gate wait was stopped only after this concrete handoff refusal; no successful-looking background validation is left to be attached after the session. Local CLI full-suite interruption and all lock-timeout attempts remain unverified, even though bounded positive controls and mutants subsequently ran.

Hosted cancellation readback: status `completed`, conclusion `cancelled`. Cancellation request and owned-branch cleanup each exited 0. Story HEAD and index remain unchanged.
