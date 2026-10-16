# Coverage map

The brief has no attack-surface catalog table. This is a brief gap; its adapter table and paired-test table are mapped below without pretending they are catalog row IDs. No source-text gate is added or changed; token-preserving static-checker controls are inapplicable. All seven plants execute behavioral Go tests.

## Scoped acceptance and production coverage

**2 of 3 scoped AC rows fully driven through public production calls; 1 internal-seam bound.** All three have executable evidence, but internal staging is not counted as a public-operation drive. **0 of 5 original remainder workstreams fully delivered**; the first workstream is advanced. The brief expressly allows a next coherent part. The original complete checklist and deferred requirements are in TASK-261016-hzbo83_scope-adjustment.md.

| Scoped row | Production call site and tests | Narrowing controls | Bound |
|---|---|---|---|
| One native managed-home projection across resolution and repair; fresh next operation | Resolve → assembleHome / finalizeMarker / checkPassthrough → resolveCredentialPolicy; TestCredentialProjectionSingleObservation; TestCredentialProjectionIsolationRefusals; TestCredentialStorePolicyActivation; TestCredentialStorePolicyNativeInspection | M1, M2, M3, M4, M5, M6 | Existing native source roles only; no external-source adoption; one-time provisioning seed is a separate vendor-config read |
| Status shares the projection; failed observation is non-current and preserves state | StatusOf → homeState → verifyHome → assembleHome → resolveCredentialPolicy; TestCredentialProjectionStatusObservation | M3 (read-failure distinction), M5/M6 attack the shared resolver through Resolve | Metadata/policy verification only, not login readiness; no token validity claim |
| Migration and identity-subrequest integration | PlanMigration / ApplyMigration → inventoryHome → resolveCredentialPolicy; TestSharedToIsolatedRemovesStaleLink, existing RC14 identity tests. TestCredentialProjectionIdentitySubrequest drives prepareIdentityMigration directly | M7 | Public migration integration is exercised, but priming a different cached caller policy is an internal staging seam; this entire scoped row is conservatively not counted as fully public-driven |

## Every adapter-table row

| Leaf row | Evidence in this candidate | Remaining / out-of-contract classification |
|---|---|---|
| Codex native | TestCodexKeyringAmbient, TestCodexIsolatedAdmission, TestCodexUnreadableConfigFails; new projection tests; native read failure and absence remain distinct; native emits no forced override | Preserved and consolidated; no supported-release qualification claim |
| Codex explicit file | TestCredentialStorePolicyActivation/codex_cli/file; TestCredentialStorePolicyNativeInspection; TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery | Required activation and typed transport deferred under brief next-coherent-part clause; codex_settings_unsupported remains; not claimed out of contract |
| Codex explicit keychain | Activation refusal under TestCredentialStorePolicyActivation/codex_cli/keychain | Qualified macOS mapping deferred; no native fallback; isolated mapping remains refused; not claimed out of contract |
| Managed Muse absent/native/file | Existing Muse tests stay unchanged; forced file activation refusal remains | Effective-file literal and successor fragment/login consumption deferred; not claimed out of contract |
| Managed Muse shared | Existing Muse link/ownership tests remain green when the full gate passes | New pre-override native-file qualification/admission deferred; not claimed out of contract |
| Claude native | Existing registry/isolation and credential-record tests retained | Session-qualified readiness derivation absent; deferred, not claimed out of contract |
| Claude forced file/keychain; Muse keychain; Pi/OpenCode non-native | TestCredentialStorePolicyActivation table and CLI refusal test | Unsupported mappings are explicitly out of contract for activation per leaf required result "Refuse ... never silent native fallback"; valid native Codex is the adjacent positive control |
| Adopted external source with forced home-store override | No adopted source contract exists in this baseline | External-source integration deferred; neither replacement nor local login is introduced; not claimed out of contract |

## Every paired-test-table row

| Leaf row | Tests / control in current scope | What remains |
|---|---|---|
| Old envelope/system/package/overlay/malformed selector | Existing TestCredentialStoreConfigGrammar, TestCredentialStoreConfigVersionBoundary, TestCredentialStoreConfigMachineOwnership and CLI round-trip/refusal tests | Unchanged slice-1 coverage; no newly claimed grammar gate or mutant |
| Shared explicit file cannot use stale native keyring/auto auth.json | Explicit file remains refused; M4 / TestCodexKeyringAmbient/keyring protect native keyring no-file policy | Explicit-file delivery positive control deferred |
| Native keyring/auto isolation and explicit keyring scope | TestCredentialProjectionIsolationRefusals; TestCodexIsolatedAdmission; TestReviewerLiteralCodexSelector; M1 | Explicit isolated-file positive control deferred because activation is not qualified |
| Native read/parse failure never becomes absence/file | TestCredentialStorePolicyNativeInspection; TestCredentialProjectionStatusObservation; TestCodexUnreadableConfigFails; M3; valid absent-key controls retained | No new parser subset; existing TOML parser remains unchanged |
| Muse native/unset literal and readiness | Existing Muse behavior untouched | Entire new literal/readiness row deferred |
| Selector switch + repair cannot migrate/delete a credential | TestCredentialProjectionSingleObservation snapshots config/profile/markers/links/synthetic auth after changed invalid selector; TestSharedToIsolatedRemovesStaleLink exercises explicit migration after repair refusal | Forced-store switch activation remains deferred; M4 narrows the keyring/file confusion bound; no broader migration-delete guarantee newly claimed |
| Alias/canonical and profile-name preservation | Existing CLI tests, including reviewer dotted-profile tests, retained unchanged | Unchanged slice-1 behavior; full gate executes reviewer tests; no new alias mutant claimed |
| Foreign source + forced home store | No new source enrollment or login path | Entire external-source row deferred |

No owed feature is reclassified as unsupported product behavior. Unsupported forced mappings remain conservative refusals, and the next slices must implement the qualified rows before this parent objective can be called delivered.
