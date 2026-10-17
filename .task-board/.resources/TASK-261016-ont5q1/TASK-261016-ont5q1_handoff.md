# M1 developer review handoff

Task: TASK-261016-ont5q1 — implement-rc7-muse-v4-producer.

Candidate is uncommitted on the managed Story checkpoint `34a89308a69d4863699887359cd122d64b1e716f`. Production policy/resolution already landed in `3113bad`; this delta adds the missing status evidence distinctions and strengthens producer tests. No second credential resolver, marker fields or consumer behavior were added. See `TASK-261016-ont5q1_gap-analysis.md` for the baseline mapping.

Normative source: curator-spec main `1059b94b4135089d842bf017070b8ab47d79a536`, CIP-0017 implementation leaf 5, environments §§5.4, 7.1, 7.4, 8.2–8.4.1, 10.1–10.3, 12, the v4 schema and Muse v4 conformance notes. The main conformance manifest equals the hosted gate pin's manifest (`6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5`). Frozen v4 schema/cases retain their existing source identity; no schema or vector bytes changed.

## Acceptance coverage

**4 of 4 AC rows driven through production entry points.** Runtime qualification remains outside this ratio.

| AC | Production call sites | Named behavioral tests |
|---|---|---|
| 1: current/repaired v4, file, four ordered XDG parents then literal, unchanged HOME, equal projections | `cmdEnvResolve` → `envprofile.Resolve` → `buildFragment` → `renderFragment` → `Fragment.JSON`/`EnvFormat`/`ShellFormat` | `TestMuseCLIFragmentAndFakeLaunch`, `TestMuseCLIStorePolicy`, `TestMuseCredentialPolicyMatrix`, `TestMuseFragmentV4PublishedCases` |
| 2: iff verified declared root, linked/copied/current/repaired/header-only, no fragment for unusable required root | `cmdEnvResolve` → `Resolve` → `verifyHome`/`checkSurfaces`, `repairUnderLock` → `applyPlan`/`copyLinkFallback` → `buildFragment`; `cmdEnvStatus` → `StatusOf` | `TestMuseRootContextVectors`, `TestMuseRootContextCopyFallback`, `TestMuseRootContextDrift`, `TestMuseRootContextRefusals` |
| 3: native/unset effective file without rewriting, explicit keychain refusal, native shared source, isolated no link | CLI config and `cmdEnvResolve` → operation-local `computeCredentialPolicy`/`inspectMuseAuth` | `TestMuseCLIStorePolicy`, `TestMuseCredentialPolicyMatrix`, `TestMuseMissingNativeFilePolicy`, `TestMusePolicySwitchPreservesCredentials` |
| 4: conflict/readability/seed preservation; distinct loader/backend/replay/residual/unknown status | `Resolve` → `inspectMuseAuth`/`applyPlan`; `StatusOf` → `homeState`/`verifyHomeContent` | `TestMusePublishedLinkStates`, `TestMuseAuthRepairAdditionalStates`, `TestMuseCLIFragmentAndFakeLaunch`, `TestMuseRootContextStatus`, `TestMuseCredentialPolicyStatus` |

## Coverage map

The brief provides no catalog surface table: this is an explicit brief gap. Its single attack surface is manager-owned verified state becoming launch policy. This map accounts for **6 of 6 mandatory paired rows**, in leaf order; it does not silently waive a missing catalog table.

| Surface row (leaf paired row) | Attacking tests | Killed narrowing mutants | Out-of-contract inputs (AC clause) |
|---|---|---|---|
| Required root absence/drift/denial/failed repair / verified link/copy/current/repair | `TestMuseRootContextDrift`, `TestMuseRootContextRefusals`, `TestMuseRootContextVectors`, `TestMuseRootContextCopyFallback` | `changed-root-admitted`, `removed-root-admitted`, `denied-root-admitted`, `linked-root-fact-dropped` | None within M1 |
| Unrelated file cannot mint fact / declared root, including header-only | `TestMuseRootContextVectors/no-context`, `/monolithic`, `/zero-applicable` | `unrelated-root-mints-fact`, `header-only-root-fact-dropped`, `no-context-admitted` | No declared context intentionally omits fact (AC2); header-only does not |
| Closed names/literal/parent/suffix/adapter / exact object and order | `TestMuseFragmentBoundary`, `TestMuseFragmentV4PublishedCases`, `TestMuseCLIFragmentAndFakeLaunch` | `boundary-admits-HOME`, `boundary-admits-missing-literal`, `boundary-admits-false-root` | Invalid shapes refused per AC1 and paired row 3 |
| Fifth renderer value / all five parsed values equal | `TestMuseCLIStorePolicy` | `renderer-drops-only-backend` | Env/shell are projections, not typed policy transports (AC1) |
| Shared fork/retarget/dangling/nonempty/read denial / live link and admitted relink | `TestMusePublishedLinkStates`, `TestMuseForkRefusedRegression`, `TestMuseAuthRepairAdditionalStates`, `TestMuseMissingNativeFilePolicy` | `missing-shared-source-admitted`, `readlink-denial-admitted` | Concurrent/real vendor refresh qualification excluded by dependency/overlap clause |
| Preserve stored native, refuse keychain / effective file and no isolated link | `TestMuseCLIStorePolicy`, `TestMuseCredentialPolicyMatrix`, `TestMuseCredentialPolicyStatus` | `muse-keychain-admitted`, `isolated-muse-native-link`, `shared-status-misreports-isolation` | Explicit keychain refused per AC3 |

## Mutant evidence

Each retained narrowing harness requires subprocess `go test` exit **1** and a named behavioral failure. Exit 0, compilation failure, absent anchor, unrelated failure or a skipped test does not count. These expected-red child gates are failing commands; their containing test passes only when it verifies that failure. Copy fallback is a separate positive control, exit 0, taking the real fallback path and checking the copy record and fact on repair/current resolve.

| Mutant | Narrows gate to / forbidden effect admitted | Named failing test | Survivor bound |
|---|---|---|---|
| fragment-reobserves-muse-release | permits a second vendor observation while composing Muse only | `TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared` | Shell sentinel is Unix-only; explicitly skipped on Windows |
| muse-keychain-admitted | admits explicit keychain for Muse only | `TestMuseCredentialPolicyMatrix/keychain/shared` | — |
| missing-shared-source-admitted | accepts absent native shared file only | `TestMuseMissingNativeFilePolicy/shared` | — |
| isolated-muse-native-link | shares native credentials only for isolated Muse | `TestMuseCredentialPolicyMatrix/file/isolated` | — |
| renderer-drops-only-backend | keeps XDG rendering but omits fifth value | `TestMuseCLIStorePolicy` | — |
| header-only-root-fact-dropped | omits verified header-only fact | `TestMuseRootContextVectors/zero-applicable` | — |
| boundary-admits-HOME | permits HOME on Muse only | `TestMuseFragmentBoundary/HOME` | — |
| boundary-admits-missing-literal | permits missing backend but retains other file checks | `TestMuseFragmentBoundary/missing-literal` | — |
| boundary-admits-false-root | permits false exclusive Muse fact only | `TestMuseFragmentBoundary/false-root` | — |
| unrelated-root-mints-fact | gains fact from unrelated header-like AGENTS.md only | `TestMuseRootContextVectors/no-context` | — |
| linked-root-fact-dropped | omits only linked required roots | `TestMuseRootContextVectors/monolithic` | — |
| denied-root-admitted (repaired plant) | bypasses only required Muse root permission-denial verification | `TestMuseRootContextRefusals/denied` | Windows ACL denial is not established by chmod; skipped there explicitly |
| shared-status-misreports-isolation | misreports only shared policy as isolated | `TestMuseCredentialPolicyStatus/shared` | — |
| readlink-denial-admitted | admits unreadable final credential link while retaining metadata/target checks | `TestMuseAuthRepairAdditionalStates/readlink-error` | — |
| configured-form-ignored | ignores only configured Muse referenced form | `TestMuseRootContextVectors/referenced-refused` | — |
| referenced-fallback | maps Muse referenced form to monolithic | `TestMuseRootContextVectors/referenced-refused` | — |
| no-context-admitted | materializes only undeclared Muse context | `TestMuseRootContextVectors/no-context` | — |
| wrong-target | changes Muse's root target only | `TestMuseRootContextVectors/monolithic` | — |
| wrong-bytes | changes Muse header bytes only | `TestMuseRootContextVectors/zero-applicable` | — |
| changed-root-admitted | bypasses only the changed Muse root payload | `TestMuseRootContextDrift/changed` | — |
| removed-root-admitted | bypasses only missing Muse root verification | `TestMuseRootContextDrift/removed` | — |
| initial denied-root plant using `os.IsPermission` | intended denial exception never applied to wrapped `stateread.Error` | **No failing test; survivor**, child exit 0 | Proves no denial bound. Replaced with `errors.Is(err, os.ErrPermission)` and rerun; the initial result is retained as a failed instrument attempt. |

No source-text gate was introduced: overlays preserve the original checks/tokens and execute behavioral suites. There is no static-token-only claim.

## Scope and limits

Changes: `internal/envprofile/status.go`, its existing Muse tests, `docs/environment-config.md`, and three existing CLI Muse test files. The CLI files are the explicitly justified public resolve/status and existing mutant-harness seams beyond the leaf's literal file list. Shared resolver/registry/marker/credential logic stays unchanged. All 19 original test names in modified test files are retained; three new named tests cover refusals, policy status and copy fallback.

No out-of-contract mandatory producer rows are dropped. The leaf's dependency/overlap clause excludes readiness/login wrappers, consumer launch/seal/hosted merge, broker delivery, release qualification and real token/concurrent refresh. No real Muse login, secret read, native credentials, model call, runtime refresh or 1.4.8 exclusion replay was performed. Synthetic permission denial is platform bounded; Windows requires ACL qualification separately. The standing status note reports historical normative facts, not new observations.

## Validation ledger

Local environment: Go 1.25.5, darwin/arm64. No test/config/source changes after the current focused run. Candidate patch SHA-256: `eb413e419ba66f354c68b478258dad939f20388adf9d4e84e4dc0827d8fbee51`.

| Command | Real exit | Evidence / meaning |
|---|---:|---|
| `go test ./internal/envfragment ./internal/envprofile ./cmd/curator -run 'TestMuse(CLIStorePolicy\|RootContextVectors\|RootContextDrift\|RootContextStatus\|FragmentBoundary\|CredentialPolicyMatrix\|MissingNativeFilePolicy\|ForkRefusedRegression\|AuthRepairAdditionalStates)$' -count=1 -timeout=5m` | 0 | Initial focused baseline; no external corpus selected |
| `go test ./cmd/curator -run '^TestMuseRootContextStatus$' -count=1` | 1 | Expected-red regression witness before status implementation; attached sanitized log |
| `go test ./cmd/curator -run '^TestMuse(RootContextRefusals\|CredentialPolicyStatus\|RootContextDrift\|RootContextStatus\|CLIStorePolicy\|RootContextVectors)$' -count=1 -timeout=5m` | 1 | Test-development failure: JSON whitespace assumption and absent on-disk stub config; corrected assertions, no product failure concealed |
| `go test ./internal/envfragment ./internal/envprofile ./cmd/curator -run '^TestMuse(RootContextRefusals\|CredentialPolicyStatus\|RootContextDrift\|RootContextStatus\|CLIStorePolicy\|RootContextVectors\|AuthRepairAdditionalStates\|RootContextCopyFallback)$' -count=1 -timeout=5m -v` | 1 | First copy-fallback control had a nonunique overlay anchor; not counted, corrected to the unique applyPlan loop |
| `go test ./internal/envfragment ./internal/envprofile ./cmd/curator -run '^TestMuse' -count=1 -timeout=8m -v` | 1 | Initial denial plant survivor; no external corpus selected, published rows skipped. Superseded by current run, not passing evidence |
| `CURATOR_CONFORMANCE_ROOT="$PWD/.temp/TASK-261016-ont5q1/spec/conformance/v1" go test ./internal/envfragment ./internal/envprofile ./cmd/curator -run '^TestMuse' -count=1 -timeout=8m -v` | 0 | Current identity, main spec checkout at the SHA above; all Muse tests, 16 published liveness vectors, v3/v4 schemas, 21 expected-red named mutant kills, positive copy fallback; no skipped local Muse rows. Attached `TASK-261016-ont5q1_focused.log` |
| `go build ./...` | 0 | Whole module compiled after production changes |
| `go vet ./internal/envfragment ./internal/envprofile ./cmd/curator` | 0 | Focused static validation, last run after all candidate edits |
| `git diff --check` | 0 | Whitespace check, rerun after skip-message fix |
| `CI_GATE_GOOS=windows bash .github/ci/platform-case-gate.sh .temp/TASK-261016-ont5q1/windows-evidence/test/go-test.json .temp/TASK-261016-ont5q1/windows-replay-red` | 1 | Captured hosted stream; intended reproduction of unclassified skips |
| `CI_GATE_GOOS=windows bash .github/ci/platform-case-gate.sh .temp/TASK-261016-ont5q1/windows-reclassified-fixture.json .temp/TASK-261016-ont5q1/windows-replay-green` | 0 | Classifier-only fixture; exactly two skip-message replacements, not a new Windows run |
| `git diff d9821926293a38ca5fd3f310753e4e24bd77c4db --exit-code` | 0 | Worktree equals signed hosted snapshot; untracked candidate files absent |
| `git log -1 --format='%G?' d9821926293a38ca5fd3f310753e4e24bd77c4db` | 0 | Output `G`: detached gate snapshot signature verified |
| `sh scripts/remote-gate.sh` (initial candidate) | 1 | Real Windows skip-class failure; attached initial gate log |
| `sh scripts/remote-gate.sh` (corrected candidate) | 1 | Real Linux CGO link failure before its script-policy tests; 19 other jobs green |
| `gh run rerun 39029607579 --job 117387246885` (early request) | 1 | Provider refused while workflow was still running |
| `gh run rerun 39029607579 --job 117387246885` (after terminal state) | 0 | Only failed Linux job retried, same candidate head |
| `gh run watch 39029607579 --exit-status --interval 20` | 0 | Real full hosted aggregate success, 20 jobs green and 2 conditional skips |
| `cmp .temp/TASK-261016-ont5q1/spec/schemas/v1/launch-env-fragment-v4.schema.json internal/envfragment/testdata/curator-spec-muse/schemas/v1/launch-env-fragment-v4.schema.json` | 0 | Published current-main schema equals retained frozen schema |

The full hosted gate is **green** on signed detached snapshot `d9821926293a38ca5fd3f310753e4e24bd77c4db`: https://github.com/relux-works/curator/actions/runs/39029607579/attempts/2 . All 20 executable jobs succeeded; the optional self-hosted macOS job and dispatch-only candidate-suite job were conditionally skipped, not executed. Lint, naming, driver matrix, platform gate self-tests, interop, all three platform tests and both race jobs are green. The final worktree equals this snapshot and the Story HEAD/index remain unchanged. No prior owner's evidence substitutes for this run.

Both launcher invocations of `sh scripts/remote-gate.sh` exited **1**; they are retained as failures. The corrected candidate's failed Linux job alone was retried through GitHub: the retry request exited **0**, `gh run watch --exit-status` exited **0**, and GitHub's real latest workflow conclusion is `success`. The 19 unchanged-head green jobs were reused. This is a real hosted retry, not a local mirror or a status override.

The first hosted run (39026822677, snapshot `726285604baae987a8c0fb6273639c29ae881807`) terminated with `sh scripts/remote-gate.sh` exit 1 (19 successful jobs, one failed, two conditional jobs skipped): a repository-caused Windows platform-case failure: Windows `go test` exit 0, classifier exit 1 for two unrecognized denial skip messages. The messages now name the already-declared `POSIX mode-bit unreadability` class; neither the skip guards nor class policy changed. The original captured stream replay exited 1; a fixture changing exactly those two messages exited 0. That fixture establishes only classifier compatibility, not Windows runtime success. A new full focused run and vet on the corrected candidate both exited 0. The corrected candidate later passed real Windows tests and the skip classifier, followed by the successful hosted aggregate above. This failed gate is not converted to passing evidence.

## DoD evidence by checklist item

| Item | Evidence |
|---|---|
| 1. Current/repaired v4 | AC1 mapping; public CLI fragment/policy tests |
| 2. Derived verified root fact | AC2 mapping; drift/denial/failed repair and link/copy/current controls |
| 3. Native/unset effective file | AC3 mapping; stored selector preserved, explicit keychain refused |
| 4. Conflict/readability/config and status | AC4 mapping; store snapshots, status evidence distinctions |
| 5. Paired rows | 6 of 6 mapped above, including forbidden-state assertions and controls |
| 6. Required checker control | `renderer-drops-only-backend` killed by `TestMuseCLIStorePolicy`, child exit 1 |
| 7. Scope | Six paths listed above; CLI seams explicitly justified |
| 8. Focused and full gate | Focused exit 0; full hosted workflow success after job-only retry, watch exit 0; original launcher exits 1 retained |
| 9. Code and AC | Existing implementation reused with missing status behavior added |
| 10. Relevant tests | Current focused run exit 0; existing names retained |
| 11. Uncommitted Story candidate | Story HEAD remains the recorded checkpoint; detached signed CI snapshot only |
| 12. Production entry-point ratio | 4 of 4 AC rows driven; call sites and runtime bounds above |
| 13. Negative gates | Denial/drift/conflict/keychain/boundary tests and unchanged forbidden state |
| 14. Narrowing mutants | 21 retained named kills; initial failed instrument separately declared |
| 15. Source-text gate | Not applicable: no source-text gate introduced; overlay mutations execute behavioral tests |
| 16. Lint | Hosted Lint job green on current signed snapshot; full hosted workflow success |
| 17. Build/validation | `go build ./...` and focused `go vet` exit 0 |
| 18. Outcome resources | Task-prefixed gap analysis, focused log, diagnostic logs and handoff attached |
| 19. Logbook | `TASK-261016-ont5q1_logbook.md` attached; no control-root file edits |

## Corrected candidate hosted retry diagnostic

Second run: https://github.com/relux-works/curator/actions/runs/39029607579 . Its Linux test job failed to link unchanged `internal/scriptpolicy.test` before any test in that package ran: three `net` CGO relocation targets were not defined. The downstream required-case failures correctly refuse the missing execution; they are not waived. The uploaded JSON has one package build-fail and no failed named test. Cache restore also reported file-exists warnings, but no cache-cause conclusion is established. A real hosted job-only retry passed both `go test` and the platform-case gate, with no source/test/config changes. The original failure is retained; its underlying cause is not established. No local substitute or product workaround is claimed.

`gh run rerun 39029607579 --job 117387246885` first exited 1 while the workflow was still running (`job cannot be rerun`). After terminal state, that exact command exited 0. `gh run watch 39029607579 --exit-status --interval 20` then exited 0. Attempt 2's Linux job `117393897040` reports served tests exit 0, the uncovered-platform assertion exit 0, overall tests exit 0 and platform-case gate exit 0. The full hosted aggregate is success at the unchanged signed snapshot. The full gate checklist can now be checked against that real terminal result.

## Attached terminal evidence

`TASK-261016-ont5q1_hosted-result.json` records the exact signed head and each real hosted job conclusion. `TASK-261016-ont5q1_hosted-retry.log` records retry commands, real exits and selected original Linux retry log lines. Both unsuccessful launcher logs, the Windows classification diagnostic, current focused output, regression red, initial failed mutant diagnostic, baseline gap analysis and task-scoped logbook remain attached. All 19 original test names in the changed files are retained; the final name/hash check was rerun after the two skip-message edits. Source/test/config bytes are unchanged since the last focused run and match the hosted snapshot.

Developer handoff: ready for review. Changes remain uncommitted for Change Request snapshotting.
