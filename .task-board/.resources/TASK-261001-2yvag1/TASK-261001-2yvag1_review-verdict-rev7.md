# TASK-261001-2yvag1 — curator-muse-environment: revision 7 review

Verdict: **accepted**. Record with `accept_cr` revision 7; route to integrating, not done. No unresolved blocking findings.

Candidate tree: `209d25f774cb07170e6190ac45eb0d4b98bf4cfc`; base: `e87d488b8fd892bc86c037ae5e325e596df8e745`. All 28 changed paths match the candidate byte-for-byte before and after reviewer overlay tests. No repository source was modified.

Normative source: spec `d373078a27a437071c2416d447221e77cec65725`, environments Muse sections, Decision 0018, v3 schema. Candidate manifest: `bd03456b92a7368d90ea74fe6953db10bc188588020683024a6a6b8735a40783`. The review first attacked the six-file delta from rejected revision 6, then combined three independent analyst surface reviews by mechanism. No new blocking finding; free hunt empty. Run goal query reports none.

## Prior findings

| Finding / repeat-of mechanism | Revision 7 result |
| --- | --- |
| R1 / live final credential link bypasses parent boundary | Fixed. `inspectMuseAuth` checks the entire managed parent route before metadata/liveness. Eight production Resolve rows cover config and config/muse, inside/outside symlinks, bare/repair. Two CLI rows require exit 1, empty stdout, preserved native/outside bytes and unchanged auth link. Boundary-drop mutant fails all eight rows. Immediate-parent-only narrowing mutant fails the ancestor-link row. |
| R2 / static corpus detached from production emission | Fixed. Permanent pinned-schema assertions validate actual Resolve.Document and actual CLI stdout. Four reachable Muse policy rows exercise both entries in both repair modes: 16 emitted-document schema assertions. All 36 corpus rows run; the other 32 are explicitly static/schema-reader rows. The previously surviving missing-permissions mutant now fails the smoke test and all four production corpus rows; a locked-only narrowing mutant also fails. |
| R3 / unrelated broker changes included | Fixed. `internal/testcli/cli.go` byte-identical to base; `cli_test.go` absent in both. No broker/askpass delta. |

## Full surface sweep

| Surface group | Result and measured bound |
| --- | --- |
| Adapter/layout | Held. Four XDG variables under one managed parent, HOME omitted; declared data/muse/skills location. No unsupported root-context, prompt or MCP target admitted. |
| Profile seeds | Held. Profile-root settings/trust seeds only, once copied then tool-owned; auth-content and secret refusals exercised by focused CLI checks. |
| Auth observation/repair | Held. 16/16 published states through production StatusOf/Resolve; zero gaps/bounds/skips. R1 adds 8 Resolve parent rows and 2 CLI parent rows. Stateread seam and E5 repairs retained. |
| Confidentiality/read failures | Held. Metadata/readlink and no-byte open/fstat only; unknown/error states refuse rather than become absence. No credential value emitted. Resolve-time boundary only; no refresh/path-swap race-freedom claim. |
| Fragment/fake child | Held. Actual production JSON schema assertions; fake child verifies env/argv and HOME. Coverage is 4/36 production corpus policies plus 32/36 static oracles, never 36 production cases. |
| Status semantics | Held. Never-enabled Muse with proven absent marker and known backup state does not make profiles non-current. Unreadable/malformed marker and unreadable backups stay non-current. |
| Conformance/pins | Held. Independently recomputed all 292 family pins: rc.13 92, hash-v2 candidate 97, Muse candidate 103. Prior pins preserved exactly. Muse counts 16 and 36. Historical candidate hash-v2 owned gaps remain explicit. |
| Scope/hygiene | Held. No LOGBOOK/CHANGELOG delta, no stray candidate files, diff --check clean. Launcher mapping/v3 reader is separately owned and not exercised here. |

## Six historical regressions

All six pass in the fresh reviewer focused command. MissingAndUnreadableKeepRecord and UnreadableApprovalStateSurfaced use the synthetic registry-driven provisioned control. RegistryBoundaryPostureAndCheck and ReportsShellHookTrustPosture use equivalent controls, so their specific registry/approval/hook postures determine the result. ManagerOwnedAbsenceReadsAreGuarded passes because Muse uses stateread; the guard and allowlist are unchanged. EmptyAllowlistWarningLeavesCurrentStatusCurrent provisions registered environments and keeps the warning advisory. These fixtures do not substitute for behavior: TestMuseStatusOptionalProvisioning and TestEnvStatusCheckCurrentScopeOnly separately exercise optional unprovisioned Muse and unknown-state refusal.

## Reused gate evidence and limits

Attached revision-7 gate log records run 36880226489, exit 0, all Linux/macOS/Windows tests, Linux/macOS race, lint, naming, interop and self-test lanes successful. Its commit `97dbcaa66c8e25eea9c31d8279e8cf490b59fa84` resolves exactly to the reviewed candidate tree. These broad checks are reused attached evidence, not reviewer reruns. Hosted candidate-suite lane skipped, so candidate counts/rows rely on the fresh local checks named here.

Raw Windows path schema compatibility is not claimed: the pinned schema accepts POSIX paths and the tests explicitly project Windows env paths for structure checks. This review ran on macOS. No real Muse session, real credential, native discovery assurance or launcher integration run. Native skills discovery, foreign personal-context isolation and refresh coordination retain spec bounds.

Three concurrent CLI attempts hit the package-wide host GOROOT test lock timeout before any test ran. They are not product failures, passes or killed mutants; their logs remain attached. The subsequent serialized checks are the claimed evidence.

Entry text kept in this artifact per the no-LOGBOOK instruction: revision-7 re-review verifies all three prior findings, tests production boundary refusal and schema emission with independent mutants, preserves explicit corpus and platform bounds, and keeps launcher and broker work outside this resolver leaf.

## Fresh verification results

Every test invocation uses `-count=1`; commands, sources and full logs are in `TASK-261001-2yvag1_review-evidence-rev7.zip`.

| Fresh reviewer check | Real exit | Result |
| --- | --- | --- |
| Root: requested two-package TestMuse / TestEnvStatus / seam guard / empty-allowlist scope | 0 | envprofile 171.187s; CLI 501.051s; all matching tests pass |
| Auth analyst: parent/fork/extra-state/seam suite | 0 | 240.620s; eight Resolve parent rows, seven extra auth states, both fork modes and guard |
| Auth analyst: candidate link-state corpus | 0 | 16/16 driven, zero skips/gaps/bounds; 229.825s |
| Fragment analyst: all three exact count inventories | 0 | 292/292 family pins match actual corpus; all base pins preserved |
| Root: boundary-drop overlay | 1 | Killed; eight of eight parent-link cases detect the removed guard |
| Root: immediate-parent-only boundary narrowing overlay | 1 | Killed; config/outside bare Resolve catches ancestor traversal |
| Root: missing-permissions emission overlay | 1 | Killed; actual CLI smoke and four production policy rows fail required-member schema assertion; 13.107s |
| Root: locked-only missing-permissions narrowing overlay | 1 | Killed by valid-permissions-native-locked row; other 35 oracle rows pass; 7.616s |
| Root: final unmodified candidate CLI smoke + fragment corpus + auth-parent probe | 0 | 19.920s; 36/36 oracle rows, four production policies; two CLI refusal probes each exit 1 with empty stdout and preserved bytes |
| Root: source identity / hygiene | 0 | 28/28 paths unchanged; diff --check clean; testcli base-equal |

The two root CLI boundary probes independently repeat the rev6 outside-parent scenario through the real command parser, retaining a live expected final auth link; the passing command verifies deliberate refusal exits rather than treating refusal as test failure. Root ran all four mutants through external Go overlays and retained no mutant in the candidate. The original HOME/fork/XDG_DATA_HOME mutant results are historical attached evidence; this round freshly exercises their positive/refusal behaviors, but does not claim rerunning those three original mutant implementations.

Analyst fan-out artifacts are auth-surface.md, fragment-surface-summary.md and regression-surface.md in the archive, grouped into the full surface table above. No formal surface-table precondition exists on this legacy leaf. Prior R1/R2/R3 mechanisms are all resolved; repeat-of entries are closure records, not new findings.

## Recovery run RUN-261001-4871bc

Acceptance verdict reaffirmed on the same immutable tree. The prior run attached this verdict and then its accept_cr command was interrupted with exit 130 during a host failure; this run completes that pending mutation. Prior review evidence above remains identified as that run's evidence. This recovery also independently attacked the rev6-to-rev7 delta and repeated three analyst surface groups; no new blocker or free-hunt finding. Findings set: `[]`. R1/R2/R3 are closed mechanisms, with no recurring unresolved finding.

Fresh recovery results (all tests use count=1):

| Check | Exit | Result |
| --- | --- | --- |
| Production link-state/optional-status/empty-allowlist tests | 0 | 16/16 published rows, no skips; optional status and advisory warning pass |
| Four historical CLI status failures plus current-scope control | 0 | All five tests pass; 200.49s command wall time |
| Auth analyst's parent/fork/extra-state/seam tests | 0 | Eight parent rows, seven extra states, two fork modes and unchanged seam guard pass |
| Fragment analyst's exact-candidate corpus | 0 | 36/36 rows, no skips; four production policies, 32 explicitly static oracles |
| Boundary check disabled | 1 | All eight parent rows detect wrongly emitted fragments |
| Corrected immediate-parent-only boundary narrowing | 1 | Four ancestor config-link rows fail; four immediate config/muse-link rows remain green |
| Corrected missing-permissions emission | 1 | Actual CLI smoke and four production policies fail the required-member schema assertion |
| Corrected locked-only missing-permissions narrowing | 1 | Locked policy fails; other 35 oracle rows remain green |
| Raw Lstat seam bypass | 1 | Guard drops to 474/475 and names inspectMuseAuth |
| HOME emitted | 1 | Production Resolve rejects registry-undeclared HOME |
| Final restored CLI smoke and outside-parent probes | 0 | 9.79s; both CLI probes exit 1, no fragment, unchanged native/outside bytes and final link |

Exactly six compiling mutants are counted as killed. Two preliminary permissions edits failed compilation because they hit the wrong return statement, and the preliminary boundary-narrowing edit changed the refusal diagnostic; those attempts are retained but excluded from the six. Corrected sources and full logs are in `TASK-261001-2yvag1_review-recovery-RUN-261001-4871bc.zip`.

The original workspace still matches all 28 candidate paths. All three mutated production files in the disposable copy were restored byte-for-byte. No product code was modified. testcli remains base-identical, and whitespace checks pass. The hosted gate commit was independently resolved to tree 209d25f774cb07170e6190ac45eb0d4b98bf4cfc; its broad green matrix is attached evidence, not freshly rerun here. No launcher integration, native Muse session, or raw Windows path-schema claim is added. The current run goal query reports none. Evidence is attached before acceptance; the board must move to integrating, never done from this reviewer.
