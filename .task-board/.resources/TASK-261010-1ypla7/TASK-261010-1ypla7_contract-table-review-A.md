# Contract-table review A — record only

Task: TASK-261010-1ypla7 — research-platform-contract-table.
Reviewed CR-TASK-261010-1ypla7-1 revision 1, base `67f736f5e45f562df892674cb53c97b0e6d732e2`, candidate tree `645ad50d014a3fe443a9d6b9d50d35a39f6bc679`.

Recommendation to deciding reviewer: changes requested for incomplete P0 coverage. This is independent reviewer A evidence, not an accept/reject mutation. The specific record-only brief controls lifecycle; no acceptance, rejection, status routing, commit or handoff is performed here. Run goal query returned no active goal.

## Findings

### A-F1 — helper retirement contradiction is absent
- severity: robustness
- severity_reason: P1 — R-UM1 cannot use this as its single contract table to resolve an explicitly required closed-schema contradiction.
- Evidence: audit A03.9 and D1/R-UM1 require reconciling `keep`. No candidate row mentions that enum conflict. [UM L121–150](https://github.com/relux-works/swarma-user-manager/blob/92bdf1d59478b8fd28d2d98cfe9fb95716d10e8e/spec/helper.md#L121-L150) admits only archive/delete in the request example but allows operator-only keep in archive policy.
- Repair: add a cited row with the selected archive/delete/keep contract and operator boundary, or mark the unresolved disposition OWNER DECISION NEEDED with concrete options; include propagation to dispatcher cleanup and book ch03. Preserve existing retention bounds and state the applicable phase.

### A-F2 — SH1 evidence/publication and continuity contradictions lack rows
- severity: robustness
- severity_reason: P1 — R-SH1 explicitly includes these audit contradictions; ownership rows alone leave repair authors without the required canonical disposition.
- Evidence: A07.8 contrasts public module CI with private source/consumer lanes; A07.10 contrasts a public tagged module with a private-fork mandate; A07.5 contrasts exact-version native resume with qualified-version range. None is represented by a canonical row or audit ID in the candidate. [SG L3–20](https://github.com/relux-works/swarma-session-host/blob/2bb9fda8ed923266532f331c9e692ba88b5c4cb6/docs/sh1-gate.md#L3-L20); [SH L202–209](https://github.com/relux-works/swarma-session-host/blob/2bb9fda8ed923266532f331c9e692ba88b5c4cb6/spec/session-host.md#L202-L209); [AR L476–488](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/session-host/architecture.ru.md#L476-L488).
- Repair: add rows for lane ownership and synthetic-tooling evidence limits, public tagged-module consumption/provenance tense, and native-resume compatibility. Name all matching documents and label unpublished schemas separately from accepted direction.

### A-F3 — broker locator/adapter and dispatcher sweep work are incomplete
- severity: robustness
- severity_reason: P2 — named P0 repair subcontracts remain unspecified; the table's propagation claim exceeds its coverage.
- Evidence: R-CB1 requires the exact helper-ledger locator and key classes/adapter boundary. I3 names UM §4.5 but does not give its file locator or identify the stale CB §3.4 references. W2 distinguishes envelopes but omits the current software root versus later keeper verification adapter/migration. [UM L70–74](https://github.com/relux-works/swarma-user-manager/blob/92bdf1d59478b8fd28d2d98cfe9fb95716d10e8e/spec/helper.md#L70-L74); [UM L152–154](https://github.com/relux-works/swarma-user-manager/blob/92bdf1d59478b8fd28d2d98cfe9fb95716d10e8e/spec/helper.md#L152-L154); [CB L92–104](https://github.com/relux-works/swarma-credential-broker/blob/f910e68f882900f03ffd036ecc2481f35bf8988b/spec/broker.md#L92-L104). R-DP1 also requires sweep-entry clarification: [DP L209–232](https://github.com/relux-works/swarma-dispatcher/blob/9cec4115b8666aa1f3325a792359135a9df5a2f7/spec/dispatcher.md#L209-L232) specifies an every-minute v0 command but omits sweep from the public request operation table; the contract table gives no disposition.
- Repair: add concise rows for ledger ownership/path/reader failure rules and stale section references, key-class/versioned adapter migration, and internal sweep entry versus public request grammar. Mark any unresolved invocation decision instead of silently adding a public operation.

### A-N1 — undefined propagation alias
- severity: note
- severity_reason: P3 — LR appears repeatedly in must-match cells but is absent from the pinned source index.
- Repair: define LR with its pinned repository file (apparently launcher README), so every propagation target is unambiguous.

### A-N2 — explicit provenance exception is faithful
- severity: note
- severity_reason: P3 — the two owner overrides are board preconditions, not Git-pinned evidence; this is honestly disclosed rather than falsely attested.
- Independently retrieved owner precondition SHA-256 matches `0fab6869d4ae2fcace990f30133ad4e97797a405016b60b1d5844528b03e28bd`. G1/G2 preserve variant b now, optional board/goals and deferred SH2+ choice; R1/R2 preserve registration-time one-shot marking, post-retirement separate revoke and operator-only standing revoke. Do not require a prohibited commit to hide the exception. The deciding reviewer should explicitly retain it.

## Citation sample and swept surfaces

Fresh authenticated read-only `gh api repos/relux-works/<repo>/contents/<file>?ref=<full-commit>` reads: 15/15 successful. Read the exact candidate via git show, not an assumed working copy. These are document checks only; no runtime/vendor qualification is inferred.

| Rows sampled | Pinned text independently inspected | Result |
|---|---|---|
| N1–N5 | [B02 L29–40](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/02-sistema-celikom.md#L29-L40); [CB L465–473](https://github.com/relux-works/swarma-credential-broker/blob/f910e68f882900f03ffd036ecc2481f35bf8988b/spec/broker.md#L465-L473); [UM L30–36](https://github.com/relux-works/swarma-user-manager/blob/92bdf1d59478b8fd28d2d98cfe9fb95716d10e8e/spec/helper.md#L30-L36); [SH L103–123](https://github.com/relux-works/swarma-session-host/blob/2bb9fda8ed923266532f331c9e692ba88b5c4cb6/spec/session-host.md#L103-L123) | Consistent in inspected claims; goal/identity overrides checked against precondition. |
| W1–W3 | [DP L73–86](https://github.com/relux-works/swarma-dispatcher/blob/9cec4115b8666aa1f3325a792359135a9df5a2f7/spec/dispatcher.md#L73-L86); [CB L164–187](https://github.com/relux-works/swarma-credential-broker/blob/f910e68f882900f03ffd036ecc2481f35bf8988b/spec/broker.md#L164-L187); [CB L211–227](https://github.com/relux-works/swarma-credential-broker/blob/f910e68f882900f03ffd036ecc2481f35bf8988b/spec/broker.md#L211-L227); [CB L425–440](https://github.com/relux-works/swarma-credential-broker/blob/f910e68f882900f03ffd036ecc2481f35bf8988b/spec/broker.md#L425-L440); [UM L30–63](https://github.com/relux-works/swarma-user-manager/blob/92bdf1d59478b8fd28d2d98cfe9fb95716d10e8e/spec/helper.md#L30-L63) | Consistent in inspected claims; goal/identity overrides checked against precondition. |
| I1, I4 | [DP L145–151](https://github.com/relux-works/swarma-dispatcher/blob/9cec4115b8666aa1f3325a792359135a9df5a2f7/spec/dispatcher.md#L145-L151); [DP L270–276](https://github.com/relux-works/swarma-dispatcher/blob/9cec4115b8666aa1f3325a792359135a9df5a2f7/spec/dispatcher.md#L270-L276); [SH L125–139](https://github.com/relux-works/swarma-session-host/blob/2bb9fda8ed923266532f331c9e692ba88b5c4cb6/spec/session-host.md#L125-L139) | Consistent in inspected claims; goal/identity overrides checked against precondition. |
| L1–L2 | [UM L30–48](https://github.com/relux-works/swarma-user-manager/blob/92bdf1d59478b8fd28d2d98cfe9fb95716d10e8e/spec/helper.md#L30-L48); [UM L197–207](https://github.com/relux-works/swarma-user-manager/blob/92bdf1d59478b8fd28d2d98cfe9fb95716d10e8e/spec/helper.md#L197-L207); [DP L253–263](https://github.com/relux-works/swarma-dispatcher/blob/9cec4115b8666aa1f3325a792359135a9df5a2f7/spec/dispatcher.md#L253-L263); [CB L425–427](https://github.com/relux-works/swarma-credential-broker/blob/f910e68f882900f03ffd036ecc2481f35bf8988b/spec/broker.md#L425-L427) | Consistent in inspected claims; goal/identity overrides checked against precondition. |
| P2–P4 | [UM L200–209](https://github.com/relux-works/swarma-user-manager/blob/92bdf1d59478b8fd28d2d98cfe9fb95716d10e8e/spec/helper.md#L200-L209); [SH L12–20](https://github.com/relux-works/swarma-session-host/blob/2bb9fda8ed923266532f331c9e692ba88b5c4cb6/spec/session-host.md#L12-L20); [SH L103–123](https://github.com/relux-works/swarma-session-host/blob/2bb9fda8ed923266532f331c9e692ba88b5c4cb6/spec/session-host.md#L103-L123) | Consistent in inspected claims; goal/identity overrides checked against precondition. |
| C1, C3–C4 | [C10 L24–31](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0010-credentials-setup-token-and-inherited-auth.md#L24-L31); [CB L129–149](https://github.com/relux-works/swarma-credential-broker/blob/f910e68f882900f03ffd036ecc2481f35bf8988b/spec/broker.md#L129-L149); [CB L345–382](https://github.com/relux-works/swarma-credential-broker/blob/f910e68f882900f03ffd036ecc2481f35bf8988b/spec/broker.md#L345-L382) | Consistent in inspected claims; goal/identity overrides checked against precondition. |
| G1, R1–R2 | [SH L169–184](https://github.com/relux-works/swarma-session-host/blob/2bb9fda8ed923266532f331c9e692ba88b5c4cb6/spec/session-host.md#L169-L184); [DP L286–292](https://github.com/relux-works/swarma-dispatcher/blob/9cec4115b8666aa1f3325a792359135a9df5a2f7/spec/dispatcher.md#L286-L292); [AR L937–947](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/session-host/architecture.ru.md#L937-L947) | Consistent in inspected claims; goal/identity overrides checked against precondition. |
| Q2, X1–X2 | [DP L88–101](https://github.com/relux-works/swarma-dispatcher/blob/9cec4115b8666aa1f3325a792359135a9df5a2f7/spec/dispatcher.md#L88-L101); [C02 L161–169](https://github.com/relux-works/curator-spec/blob/2f0531b4edcc99c6118c277deb00e8736392c04d/cips/CIP-0002-project-context-in-managed-launches.md#L161-L169); [C07 L16–28](https://github.com/relux-works/curator-spec/blob/2f0531b4edcc99c6118c277deb00e8736392c04d/cips/CIP-0007-manager-provisioned-cli-tools.md#L16-L28); [C04 L96–104](https://github.com/relux-works/curator-spec/blob/2f0531b4edcc99c6118c277deb00e8736392c04d/cips/CIP-0004-shell-hook-without-sourcing-and-path-append.md#L96-L104) | Consistent in inspected claims; goal/identity overrides checked against precondition. |
| D1–D3 | [C09 L72–81](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0009-donor-side-deployment-and-bridge.md#L72-L81); [C09 L221–229](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0009-donor-side-deployment-and-bridge.md#L221-L229); [C09 L156–158](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0009-donor-side-deployment-and-bridge.md#L156-L158); [B11 L49–56](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/11-udalennoe-rabochee-mesto.md#L49-L56) | Consistent in inspected claims; goal/identity overrides checked against precondition. |

This sample covers 27 distinct rows across all nine requested categories, exceeding the minimum 12. It is not a claim that all 45 rows or every subclaim were independently verified. No sampled recommendation was promoted into an owner decision. Open rows supply actionable options, including retaining disabled phases; future signature grammars are not invented.

| Audit D1 P0 item | Coverage disposition |
|---|---|
| R-CS1 | D1–D5 cover variants, trust, CAs, pending schemas and absent helper extension. |
| R-LR1 | N6/P1/P2/L2/C9 identify versioned amendment and unresolved first consumer. |
| R-DP1 | N3/W1/I1–I4/P6/Q1–Q2 cover most work; sweep omission A-F3. |
| R-UM1 | N2/N7/W3/I1/L1/P2/P6/D5 cover most work; keep omission A-F1. |
| R-SH1 | N4/N5/P3/P4/G1/G2 cover ownership/goals; A-F2 identifies unrepresented contradictions. |
| R-CB1 | W2/C1–C9/N1 cover substantial scope; A-F3 identifies incomplete locator/adapter contracts. |
| R-CS2 | V1/C1–C9/X1/X2 preserve acceptance versus grammar and pending consumer/PATH choices. |

All seven P0 items were examined; four have identified coverage gaps. Audit input was the current board outcome (labelled rev3 hygiene update), sections A and D; the candidate explicitly records its earlier rev2 retrieval and does not falsely claim audit acceptance. Findings concern substantive contracts retained in that audit, not its historical hygiene edits.

## Hygiene and limits

Exact candidate delta: one added research file, 156 lines, 64,667 bytes, below 65,536 bytes by 869 bytes. Findings should be repaired with compression as needed to retain the bound. No LOGBOOK or code change in candidate. Reviewer changed no repository files; working status remains the producer's untracked research file. Text inspection found no personal paths, first names, credentials or session links in the candidate; GitHub source links and generic platform paths are legitimate. This is bounded inspection, not a secret-scanner completeness claim.

No tests, builds, commits, renderers or runtime gates run. Producer's reported 42/42 fetches and packaging verifier are not counted as independently rerun review evidence. Sources downloaded temporarily are not attached. This task-scoped report is the only reviewer outcome; important findings and provenance exception are also recorded in board notes. LOGBOOK is excluded by the explicit brief.
