# TASK-261016-wh6v9a — credential-store-policy-successors, revision 3

Ready for review. Revision 3 is validation rework after convergence onto the Landlock probe
recursion fix. No product code, test or configuration was changed in this run.
The two existing task logbook entries moved verbatim to the top, newest first.
An independent clean three-way merge of revision 2 onto the new Story base
matches every candidate path except that exact logbook move.

## Identity and bounded delta

- Story base/unchanged HEAD: a17e58156ac30a1629e3ff36efa259bab1555a15.
- Included upstream probe fix: 6e85a912.
- Previous revision-2 candidate tree: 9bad791482638be26d3ab306653fd279791f449b.
- Signed detached gate snapshot: 809e3154f350986be78944503a9182a1d5d09940.
- Candidate tree: da51dd68abb173fc894d8369e28662d6a17886d7; signature G.
- All 3637 revision-2 test functions remain under their original paths/names.
- scripts/remote-gate.sh has no candidate delta; its signing fix is on trunk.
- Product changes outside the declared module: none. LOGBOOK.md is the
  expressly requested documentation move. The upstream scriptworker changes
  are convergence content, not new producer changes.
- The Story branch and real index remain at the recorded checkpoint. No
  producer commit, switch, rebase, merge or manual integration was performed.
  git merge-tree only computed comparison objects without moving any ref.

## Review findings answered

Review B of revision 2 found no Muse/SNAP-LOCK defect; its only open findings
were the two hosted Ubuntu test failures. The revision-3 brief attributes the
load to the probe recursion bug, now fixed upstream. This candidate carries
that fix. The fresh hosted verdict below, rather than a local substitute,
is the evidence for the current validation outcome. No existing assertion
was relaxed, skipped or renamed. No repeat-of class was assigned.

The narrowed delivery remains Muse managed file policy, shared/isolated
admission, v4 fragment composition and verified root-context fact. The single
operation-local credential-policy projection is retained; the plan feeds the
fragment without vendor discovery. The fresh no-probe command exercises
CaptureMaterialize and StatusOf for Codex and Muse native/file x shared/isolated.
Its executable sentinel and associated control remain Unix-only; no Windows
vendor-execution claim follows from that explicit skip.

Readiness/login consumers, qualified Codex v5/application behavior, the real
hosted override plant and adopted external-source policy remain preserved in
TASK-261017-0pawkw — credential-store-policy-consumer-remainder. No runtime
qualification or deferred consumer delivery is claimed. Normative authority
and scope are inherited unchanged from the accepted revision-1 and reviewed
revision-2 evidence; this validation-only run performed no new spec research.

## Evidence provenance

Fresh commands and real exits are listed below; the hosted waiter terminated with exit 0.
The full gate re-executes committed tests and their committed Muse/root
narrowing harnesses on the exact signed snapshot. The twelve separate
SNAP-LOCK overlay controls are inherited revision-2 evidence, not claimed as
fresh revision-3 execution. Their test and product definitions are retained.
Original focused/build/lint results remain recorded in
TASK-261016-wh6v9a_handoff-rev2.md and its validation resources; they are not
represented as new local commands here. Fresh hosted jobs supply current-tree
build, test, race, lint and conformance evidence.

## Production coverage

**7 of 8 narrowed AC rows driven through production entrypoints.** The eighth
row is explicitly structural/boundary-only, because this repository emits a
fragment; it does not provide the external launcher's fragment ingestion API.
These eight rows define the accepted Muse producer slice, not eight claims of
full-leaf delivery.

| Row | Requirement | Named tests | Production call site / bound |
|---|---|---|---|
| 1 | Native/unset/file effective file; unchanged selector; mapping reported | TestMuseCredentialPolicyMatrix; TestMuseCLIStorePolicy | cmdEnvConfig → config store; cmdEnvResolve → Resolve → resolveCredentialPolicy → buildFragment |
| 2 | Shared uses captured native XDG auth route | TestMuseCLIFragmentAndFakeLaunch; TestMuseCredentialPolicyMatrix | cmdEnvResolve → Resolve → nativeHome → inspectMuseAuth; fake child only observes projection |
| 3 | Missing shared native file refuses with guidance and no placeholder | TestMuseMissingNativeFilePolicy/shared | Resolve → inspectMuseAuth; unchanged credential-scope snapshot |
| 4 | Isolated does not share/copy/synthesize credentials, even without native auth | TestMuseCredentialPolicyMatrix/file/isolated; TestMuseMissingNativeFilePolicy/isolated | Resolve → resolveCredentialPolicy → home plan / repair |
| 5 | Explicit Muse keychain refuses with no fragment or state change | TestMuseCredentialPolicyMatrix/keychain/shared; TestMuseCLIStorePolicy | cmdEnvResolve → Resolve → admitCredentialStore |
| 6 | Repair preserves recorded link or detached file; explicit migration works | TestMusePolicySwitchPreservesCredentials; TestMusePolicyExplicitMigration | Resolve → repair; PlanMigration / ApplyMigration → existing transaction |
| 7 | V4 five-value projection in JSON/env/shell, root fact iff verified surface, including header-only | TestMuseCLIStorePolicy; TestMuseRootContextVectors; TestMuseRootContextDrift | cmdEnvResolve → Resolve → buildFragment → renderFragment |
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

## Retained narrowing evidence (revision 2)

Historical executions, with actual child exits 1 at the named failing tests.
No survivors were reported in revision 2; this does not claim exhaustive mutation coverage.

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


## Revision-3 terminal validation

The standalone full gate actually exited **0**. Provider readback reports
**20 jobs success and 2 configured optional skips**, on the exact signed
snapshot above. No skipped job is claimed as executed. All enabled Linux,
macOS and Windows Test jobs, both Race jobs, lint, driver qualification,
interop, naming and gate self-tests passed.

Hosted run: https://github.com/relux-works/curator/actions/runs/39010003372

| Exact command | Real exit | Evidence |
|---|---:|---|
| `go test ./internal/envprofile -run '^TestMaterializeCaptureAndStatusNeverProbeVendor$' -count=1 -v -timeout=3m` | 0 | TASK-261016-wh6v9a_validation-rev3.md, full local output |
| `sh scripts/remote-gate.sh` (configured agent Git environment) | 0 | TASK-261016-wh6v9a_gate-rev3.log and gate-rev3.json |
| `git diff --check` | 0 | Executed before the gate and again after its terminal success; no output |
| `gh run download 39010003372 --name test-evidence-ubuntu-latest --dir .temp/TASK-261016-wh6v9a/revision-3/ubuntu-evidence` | 0 | Uploaded Ubuntu go-test.json, extracted named events and SHA-256 in validation-rev3.md |

Every command ran directly as its own process, without tee or a verdict pipe.
The full gate stayed active in this turn and was observed through bounded
exec-session calls; no waiter or delayed evidence attachment is left pending.
The temporary remote gate branch was removed, independently confirmed by an
empty successful exact-ref read.

The fresh Ubuntu artifact explicitly records pass events for
`TestAuthoritativeDryRunCasesMutateNothingPersistent/compiled-cache-miss-is-read-only`
and `TestReviewProbeLegacyLaneIdentityInvalid`. These are the two prior review
findings; both existing assertions are retained and pass on revision 3.
No product repair was needed after the fresh full gate.

The same artifact records **16 of 16 Muse/root narrowing controls killed**,
zero survivors. Each parent control checks a child exit **1** and its named
behavioral failure; all sixteen corresponding expected-red log lines are
included in validation-rev3.md. The real renderer control remains a producer
plant, not the deferred hosted-consumer qualification. The retained tables
above identify each weakened class and killer; the three release/probe and
nine other separate SNAP-LOCK controls there remain historical revision-2
executions, not newly claimed executions.

Convergence interaction coverage remains **1 of 1 rows driven**, through
`CaptureMaterialize` and `StatusOf` by
`TestMaterializeCaptureAndStatusNeverProbeVendor`, including all four Muse
selector/isolation combinations. Producer coverage remains **7 of 8 rows
driven**, with the structural-only eighth row and deferred leaf rows declared
above. The missing attack-surface catalog table remains an explicit brief gap.

After CI, an independent temporary index reconstructed the working-tree tree:
it equals signed gate tree `da51dd68abb173fc894d8369e28662d6a17886d7` exactly.
The real index and Story HEAD are unchanged; signature verification still
returns G. Comparing the complete clean-convergence tree to that gate tree
returns exactly `LOGBOOK.md`, whose two original task blocks were moved
verbatim. No additional product, test or configuration edits were made.

Persisted resources: TASK-261016-wh6v9a_validation-rev3.md,
TASK-261016-wh6v9a_identity-rev3.json,
TASK-261016-wh6v9a_gate-rev3.json,
TASK-261016-wh6v9a_gate-rev3.log and this handoff.
The optional self-hosted job label is sanitized in the public gate resources;
actual conclusions, job identifiers and URLs are unchanged.

The candidate remains uncommitted for the managed Change Request handoff.
Review and integration remain subsequent role obligations.
