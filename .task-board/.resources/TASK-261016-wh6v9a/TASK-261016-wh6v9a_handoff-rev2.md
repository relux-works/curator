# TASK-261016-wh6v9a \u2014 credential-store-policy-successors, revision 2

Ready for review for the accepted, narrowed Muse producer slice. This is a
revision-2 convergence fix and revalidation, not delivery of deferred consumers.

## Accepted-to-combined change and regression

Revision 1 was ACCEPTED with no findings in
TASK-261016-wh6v9a_review-verdict-rev1.md; the assignment's legacy rejection
sentence does not match the actual verdict. No repeat-of class was supplied.
SNAP-LOCK is now on the Story base. The clean three-way carry of all 88
revision-1 paths was checked against the accepted tree. The signed-snapshot
script fix is already on trunk and has no candidate delta.

Existing TestMaterializeCaptureAndStatusNeverProbeVendor reproduced a real
convergence regression before repair: Go exit 1, `vendor executable ran`.
The revision-1 fragment builder called resolveCredentialPolicy after content
verification restored credentialRelease. That escaped the supplied unknown
release observation and called req.detected, even for Codex captures.

The homePlan now retains the single effectiveCredentialPolicy computed by
assembleHomeObserved. buildFragment consumes that exact projection instead
of resolving again. Resolve, CaptureMaterialize and StatusOf share the existing
release-aware resolver; the release-keyed cache is unchanged. No new policy
selection or fallback path was introduced. The production route is
CaptureMaterialize \u2192 resolveVerified(verifyHomeContent) \u2192 verifyHomeObserved
\u2192 assembleHomeObserved \u2192 resolveCredentialPolicyObserved; buildFragment uses
plan.credentialPolicy. StatusOf uses verifyHomeContent and does not build a
fragment or observe the vendor.

The retained no-probe test now covers Codex plus Muse native/file \u00d7
shared/isolated, with Detect=nil and real PATH executable sentinels. Each
capture returns the v4 file projection, status retains a current home with
unknown detected release, and source/credential bytes, modes and links stay
unchanged. Export/manifest-only subtest labels exercise the common capture
API, not yet-delivered CLI commands. No CLI snapshot qualification is claimed.

Revision-2-specific AC: **1 of 1 interaction rows driven** through production
CaptureMaterialize and StatusOf, with the named no-probe test and the Muse-only
re-observation narrowing mutant. The original narrowed producer coverage
remains **7 of 8 AC rows driven**, with row 8 explicitly boundary/schema-only.

## Scope, authority and identity

Muse native/unset/file, shared/isolated admission, captured native sharing,
v4 composition and verified root-context fact are the delivered slice.
Readiness/login consumers, qualified Codex v5/application behavior, the real
hosted-override plant and adopted external-source policy remain in
TASK-261017-0pawkw \u2014 credential-store-policy-consumer-remainder. No runtime
login/refresh, hosted consumer, source-account or Codex qualification is claimed.

Fresh curator-spec main equals the unchanged revision-1 normative authority
1059b94b4135089d842bf017070b8ab47d79a536. The cited login-readiness,
environments and profile-codex-settings clauses and accepted decisions retain
the prior accepted scope. No conflicting normative requirement was found.
CI uses its separately frozen conformance pin
43bf0a2506d5c354a73bbc3ea4623d4653db10c7; focused Muse tests also ran with it.

- Story base/unchanged HEAD: `612960ee90bedc7a4a113eda13aaa6cf2a2cbe9f`.
- Signed detached gate snapshot: `178649841e0af556ff0827e8dded65194fb34ebc`; signature `G`.
- Exact candidate tree: `9bad791482638be26d3ab306653fd279791f449b`.
- All 3609 test function names from the accepted tree remain present.
- Story branch and real index remain unchanged; no producer commit.
- Local execution: Go 1.25.5, darwin/arm64.

Only three carried paths differ from their clean revision-1 merge:
internal/envprofile/managed.go (policy retained in plan and gofmt),
cmd/curator/muse_credential_mutants_test.go (new narrowing plant and its
Unix-only sentinel bound), and LOGBOOK.md (convergence evidence and correct
trunk attribution for gate signing). The only additional tracked path is
internal/envprofile/materialize_capture_test.go (Muse no-probe cases).
No product changes outside the declared module/seams. README, producer code,
corpus and other tests equal the clean accepted-to-trunk merge. No prior
reviewer test was removed or renamed; the full gate exercises retained tests.

## Revision-2 narrowing controls

All Go child commands must exit 1 at the named behavioral test to count as
kills. Compile/setup failures or absent named failures do not qualify.

| Mutant | Narrowed behavior | Named failing test | Survival bound |
|---|---|---|---|
| fragment-reobserves-muse-release | re-enter policy observation only for Muse during fragment building | TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared | Killed on macOS; Unix-only sentinel, Windows skipped explicitly |
| M-policy-probe | ignore supplied credentialRelease only for Codex | TestMaterializeCaptureAndStatusNeverProbeVendor/codex | Killed |
| M-release-cache | allow unknown release to reuse a different release's cache | TestMaterializeCaptureCredentialPolicyReleaseObservation; TestMaterializeCaptureCredentialPolicyReleaseCache | Killed |
| M-probe | observe Codex vendor during content verification | TestMaterializeCaptureAndStatusNeverProbeVendor/codex | Killed |
| configured-form-ignored | ignore only Muse referenced machine form | TestMuseRootContextVectors/referenced-refused | Killed |
| referenced-fallback | admit Muse referenced form by replacing it with monolithic | TestMuseRootContextVectors/referenced-refused | Killed |
| no-context-admitted | materialize undeclared context only for Muse | TestMuseRootContextVectors/no-context | Killed |
| wrong-target | move Muse root to the wrong surface | TestMuseRootContextVectors/monolithic | Killed |
| wrong-bytes | add unexpected bytes only to Muse root header | TestMuseRootContextVectors/zero-applicable | Killed |
| changed-root-admitted | accept exactly the planted changed Muse root bytes | TestMuseRootContextDrift/changed | Killed |
| removed-root-admitted | accept an absent Muse root surface | TestMuseRootContextDrift/removed | Killed |

| M-codex-file-admission | admit only unqualified Codex explicit file | TestMaterializeCaptureCredentialStoreAdmission/file | Killed |
| M-denied-class | lose unusable classification only for permission-denied journal reads | TestMaterializeCaptureUnreadableJournal | Killed |
| M-lock-name | accept a replaced lock name only when its size is zero | TestMaterializeCaptureAuthorizingDrift/lock-identity | Killed |
| M-composition-policy | accept only an OverlaysAllowed change during capture | TestMaterializeCaptureSameFragmentDrift/composition-policy | Killed |
| M3-lock-only | revalidate only the retained lock object and omit other captured evidence | TestMaterializeCaptureSameFragmentDrift/aba | Killed |
| M-prepared | ignore only prepared pending journals | TestMaterializeCapturePendingDoesNotRecover/prepared | Killed |
| M-corrupt-absence | treat the planted broken journal as absence | TestMaterializeCaptureBadRecoveryEvidence/corrupt | Killed |
| M-executable-mode | ignore only execute-bit drift | TestMaterializeCaptureSameFragmentDrift/mode | Killed |
| M-provision-absent | provision only an absent capture home | TestMaterializeCaptureDoesNotProvision/absent | Killed |

Together with the eight retained policy controls below: **28 killed,
0 survivors locally**. All twelve SNAP-LOCK plants were rerun from the attached
revision-4 harness with task-local paths. The full sweep supplements the first
three-control release/probe run; its command and results are separately attached.
The new Muse plant preserves TBH_CREDENTIAL_BACKEND and every policy token
while changing behavior; it executes the behavioral capture/status suite.
No static token checker is used as the acceptance oracle.

The no-probe real-executable sentinel and its new Muse-only mutant are
Unix-only; Windows is explicitly skipped with the existing allowed reason.
Windows product compilation and its remaining Muse behavior are tested by
the hosted gate; no Windows vendor-execution claim is inferred from a skip.

## Production coverage

**7 of 8 narrowed AC rows driven through production entrypoints.** The eighth
row is explicitly structural/boundary-only, because this repository emits a
fragment; it does not provide the external launcher's fragment ingestion API.
These eight rows define the accepted Muse producer slice, not eight claims of
full-leaf delivery.

| Row | Requirement | Named tests | Production call site / bound |
|---|---|---|---|
| 1 | Native/unset/file effective file; unchanged selector; mapping reported | TestMuseCredentialPolicyMatrix; TestMuseCLIStorePolicy | cmdEnvConfig \u2192 config store; cmdEnvResolve \u2192 Resolve \u2192 resolveCredentialPolicy \u2192 buildFragment |
| 2 | Shared uses captured native XDG auth route | TestMuseCLIFragmentAndFakeLaunch; TestMuseCredentialPolicyMatrix | cmdEnvResolve \u2192 Resolve \u2192 nativeHome \u2192 inspectMuseAuth; fake child only observes projection |
| 3 | Missing shared native file refuses with guidance and no placeholder | TestMuseMissingNativeFilePolicy/shared | Resolve \u2192 inspectMuseAuth; unchanged credential-scope snapshot |
| 4 | Isolated does not share/copy/synthesize credentials, even without native auth | TestMuseCredentialPolicyMatrix/file/isolated; TestMuseMissingNativeFilePolicy/isolated | Resolve \u2192 resolveCredentialPolicy \u2192 home plan / repair |
| 5 | Explicit Muse keychain refuses with no fragment or state change | TestMuseCredentialPolicyMatrix/keychain/shared; TestMuseCLIStorePolicy | cmdEnvResolve \u2192 Resolve \u2192 admitCredentialStore |
| 6 | Repair preserves recorded link or detached file; explicit migration works | TestMusePolicySwitchPreservesCredentials; TestMusePolicyExplicitMigration | Resolve \u2192 repair; PlanMigration / ApplyMigration \u2192 existing transaction |
| 7 | V4 five-value projection in JSON/env/shell, root fact iff verified surface, including header-only | TestMuseCLIStorePolicy; TestMuseRootContextVectors; TestMuseRootContextDrift | cmdEnvResolve \u2192 Resolve \u2192 buildFragment \u2192 renderFragment |
| 8 | Closed successor fields, agreement, old-token rejection and one-parent paths | TestMuseFragmentV4PublishedCases; TestMuseFragmentBoundary | Structural oracle + direct production CheckBoundary; malformed wire ingestion by an external launcher is not exercised |

Published corpus: **67 cases (12 valid, 55 invalid)**, copied verbatim and
selecting their exact frozen v1/v2/v3/v4 schemas. The one-parent/suffix relation
also runs through CheckBoundary. Windows uses the existing explicit POSIX-path
projection for schema structure; raw Windows path schema qualification is not
claimed. No prior test function was removed or renamed. Existing
TestMuseDirectoriesAndIsolation now checks stale-home refusal when switching
an already shared home; new isolated provisioning is separately tested.

## Leaf paired-table coverage and explicit bounds

The brief does not contain an attack-surface catalog table: this is a brief
coverage-map gap, not an implicit waiver. Its eight negative/positive rows are
accounted for below. A structural oracle is never called a launched consumer.

| Leaf paired row | Evidence in this slice | Remaining bound |
|---|---|---|
| Old config/system/package/overlay/malformed selectors | Retained TestCredentialStoreConfigGrammar, TestCredentialStoreConfigVersionBoundary, TestCredentialStoreConfigMachineOwnership and CLI config tests | No new source carrier adoption; reused established config contract |
| Codex shared explicit file resurrects stale auth under keyring/auto | TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery remains green | Explicit Codex application refuses before delivery; qualified mismatch/application row remains in successor task |
| Codex native auto/keyring isolated / per-home keyring confusion | Retained TestCredentialProjectionIsolationRefusals and activation tests | Isolated explicit-file positive application remains disabled pending qualification |
| Native config read/parse failure becomes absence | Retained TestCredentialStorePolicyNativeInspection and projection observation tests | No runtime qualification inferred from config parsing |
| Muse file literal omitted or Keychain readiness substituted | Matrix, CLI five-value rendering, isolated/shared controls | Login/readiness and real hosted composition remain deferred; fake executable is not qualification |
| Selector switch plus repair erases vendor file / ownership | TestMusePolicySwitchPreservesCredentials paired with TestMusePolicyExplicitMigration | Concurrent vendor refresh remains unqualified |
| Alias persisted or profile name normalized | Retained CLI round-trip/routing and reviewer dotted-profile regressions | No new source aliases/carriers |
| Foreign source admitted beside forced home store | Not implemented or claimed | Preserved in TASK-261017-0pawkw; native home sources only in this slice |

## Narrowing mutant evidence

Every listed subprocess actually exits **1** and prints the named failing
test. The harness itself exits **0** only after confirming those failures.
No compile/setup error is counted as a kill. **Revision-1 eight controls retained; revision-2 execution adds the controls below.**

| Mutant | Gate narrowed to admit / omit | Named failing test | Survival bound |
|---|---|---|---|
| muse-keychain-admitted | additionally admit Muse keychain as file | TestMuseCredentialPolicyMatrix/keychain/shared | No survivor |
| missing-shared-source-admitted | treat only absent native shared auth as live | TestMuseMissingNativeFilePolicy/shared | No survivor |
| isolated-muse-native-link | let only Muse isolated use shared native-link construction | TestMuseCredentialPolicyMatrix/file/isolated | No survivor |
| renderer-drops-only-backend | omit only TBH_CREDENTIAL_BACKEND during rendering; all policy tokens remain in source | TestMuseCLIStorePolicy | No survivor; producer renderer plant, not hosted-consumer plant |
| header-only-root-fact-dropped | omit root fact only for header-only verified root | TestMuseRootContextVectors/zero-applicable | No survivor |
| boundary-admits-HOME | additionally admit Muse HOME while other variable checks remain | TestMuseFragmentBoundary/HOME | No survivor; direct boundary scope |
| boundary-admits-missing-literal | admit absent backend literal while other policy checks remain | TestMuseFragmentBoundary/missing-literal | No survivor; direct boundary scope |
| boundary-admits-false-root | admit false Muse root fact while non-Muse refusal remains | TestMuseFragmentBoundary/false-root | No survivor; direct boundary scope |

No new source-text acceptance checker is introduced. The renderer control
preserves searched policy literals and executes the real CLI behavioral test,
so static token agreement cannot substitute for its observed output.

## Revision-2 command results

All commands run directly, with logs redirected, no tee/pipeline verdicts.
No previously attached green command is claimed as new execution.

| Exact command | Real exit | Log |
|---|---:|---|
| go test ./internal/envprofile -run '^TestMaterializeCaptureAndStatusNeverProbeVendor$' -count=1 -v -timeout=3m (before fix) | 1 | no-probe-before-01.log; regression, not green evidence |
| go test ./internal/envprofile -run '^TestMaterializeCaptureAndStatusNeverProbeVendor$' -count=1 -v -timeout=3m (after fix) | 0 | no-probe-after-01.log |
| go test ./internal/envprofile ./internal/envfragment ./internal/envregistry ./internal/config -run 'Test(Muse\|CredentialStore\|CredentialProjection\|MaterializeCapture)' -count=1 -timeout=6m | 0 | focused-01.log |
| go test ./cmd/curator -run 'TestMuse(CredentialPolicyNarrowingMutants\|CLI\|Fragment\|RootContext)\|TestCredential(Store\|Projection)\|TestReview.*Credential' -count=1 -v -timeout=8m | 0 | cli-mutants-01.log; before Windows skip annotation |
| python3 .temp/TASK-261016-wh6v9a/revision-2/snap-all-harness.py | 0 | snap-all-plants-01.log; all twelve Go children exit 1 at named assertions |
| python3 .temp/TASK-261016-wh6v9a/revision-2/snap-harness.py | 0 | snap-plants-01.log; each of three Go children exits 1 at named assertions |
| CURATOR_CONFORMANCE_ROOT="$PWD/.temp/TASK-261016-wh6v9a/ci-spec/conformance/v1" go test ./cmd/curator ./internal/envprofile -run '^TestMuse\|^TestMaterializeCapture(CredentialPolicyRelease\|AndStatusNeverProbeVendor)\|^TestReview.*Credential' -count=1 -v -timeout=6m | 0 | pinned-muse-01.log; exact candidate, includes all 16 Muse/root plants |
| CURATOR_CONFORMANCE_ROOT="$PWD/.temp/TASK-261016-wh6v9a/ci-spec/conformance/v1" go test ./internal/envfragment ./internal/envregistry ./internal/config -count=1 -timeout=5m | 0 | packages-01.log |
| go build ./... | 0 | build-01.log |
| go vet ./internal/envprofile ./internal/envregistry ./internal/envfragment ./cmd/curator | 0 | vet-01.log |
| GOLANGCI_LINT_CACHE="$PWD/.temp/TASK-261016-wh6v9a/revision-2/lint-cache" golangci-lint run | 0 | lint-01.log; 0 issues |
| git diff --check | 0 | whitespace-01.log |

| sh scripts/remote-gate.sh (configured agent Git environment) | 0 | remote-gate-01.log; exact signed candidate, 20 enabled jobs success, 2 configured optional skips |

Full hosted gate: **success**, actual standalone script exit **0**.
https://github.com/relux-works/curator/actions/runs/38983284153
All Linux/macOS Test and Race jobs, Windows Test, driver qualification,
interop, naming, lint and gate self-tests passed. The configured optional
self-hosted and candidate-suite jobs were skipped, not executed.
The waiter stayed active in this turn and was observed through bounded tool
calls; no pending process or delayed evidence write is left behind.

Post-gate identity was independently recomputed: working tree equals signed
gate tree 9bad791482638be26d3ab306653fd279791f449b, signature G; Story HEAD and
real index are unchanged. Temporary gate branch is removed. Fresh protected
HEAD still equals the Story base. No committed producer candidate.

Four cited normative documents were independently fetched from immutable
curator-spec revision 1059b94b and verified byte-for-byte equal to the source
read here; SHA-256 evidence is in TASK-261016-wh6v9a_source-identity-rev2.json.
The pinned local conformance root's manifest digest equals the CI pin:
6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5.

Attached evidence: TASK-261016-wh6v9a_validation-rev2.md (full local logs),
TASK-261016-wh6v9a_snap-controls-rev2.md and snap-harness-rev2.py (all twelve
SNAP-LOCK controls), TASK-261016-wh6v9a_gate-rev2.json and gate-rev2.log,
TASK-261016-wh6v9a_identity-rev2.json and source-identity-rev2.json.
Earlier progress resources report their then-pending state; this handoff and
the terminal gate resources supersede that pending status.

