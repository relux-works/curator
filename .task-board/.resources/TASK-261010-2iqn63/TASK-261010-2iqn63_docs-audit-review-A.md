# TASK-261010-2iqn63 — platform-docs-audit — reviewer A record-only verdict

CR-TASK-261010-2iqn63-1 revision 1. Candidate e819ceae5110373530a1489029f452239690982b; base c53ba4b95ff38ef832caffe45009e7fd3160e61d.

Recommendation: changes_requested. RECORD-ONLY reviewer A; reviewer B owns the deciding mutation. No product tests/builds/renderers were run. No repository file or LOGBOOK was changed.

## Findings

```yaml
findings:
  - id: A-F1
    severity: robustness
    severity_reason: P1 — reversed capacity claim would generate an incorrect P0 repair
    rows: [A09.9]
    required_change: Describe book capacity waiting separately from quota refusal; compare unqualified waiting to v0 refusal, not to v1 queuing. Update R-DP1 and A-F1 references accordingly.
  - id: A-F2
    severity: robustness
    severity_reason: P1 — alleged trust contradiction is not established by the cited text
    rows: [A09.4]
    required_change: Reclassify built-in-registry wording as terminology ambiguity unless evidence proves compiled-in keys. A default registry and separately shipped default trust bundle can coexist. Do not commission a security-contract change on this inference.
  - id: A-F3
    severity: robustness
    severity_reason: P1 — normative clause citations identify unrelated sections and defeat actionable fact checking
    rows: [A06.1, A08.4, A11.7, A11.8]
    required_change: Repair CIP-0009 D9/D10/D11 locators and sweep all CIP-0009 citations semantically, not merely for valid line bounds. D9 is line 160, D10 line 162, D11 lines 164–170 at the pinned revision; lines 149–154 concern lifecycle/control and 155–156 are blank/CA heading.
  - id: A-N1
    severity: note
    severity_reason: P2 — goal wording is ambiguous but the board-goal context already supplies a scope
    rows: [A07.6, A08.3]
    required_change: Prefer scoped clarification over asserting a proven incompatible product decision; SH section 3.8 begins with a board-goal flow, while architecture explicitly separates agent and board namespaces.
```

## Independently swept surfaces

16 conflict samples across chapters 02–11 (chapter 06 is checked through A07.1's direct comparison):

| Row | Result |
|---|---|
| A02.2 | Secret-free overview lacks the broker's explicit readable-lease exception; supported scoping concern. |
| A03.2 | Confirmed 16 KiB book versus 80 KiB launcher / 48 KiB decoded plan. |
| A03.5 | Confirmed book receipt follows privilege drop; helper receipt precedes drop. |
| A04.2 | Confirmed dispatcher revocation versus operator-only keeper revocation. |
| A05.6 | Confirmed different signed domain constants. |
| A07.1 | Confirmed chapter 06 process owns PTY versus chapter 07 separate runner. |
| A07.2 | Shared app-server versus current private-per-session app-server is real wording drift; label current and target phases rather than infer an implementation defect. |
| A07.5 | Confirmed same-version restriction is narrower than qualified version range. |
| A07.6 | Board context limits strength of conflict conclusion; clarification note A-N1. |
| A08.4 | Old onboarding delivers carrier tokens; later donor design excludes them. Underlying finding holds, but one citation is wrong (A-F3). |
| A09.4 | Unsupported contradiction: A-F2. |
| A09.9 | Reversed book claim: A-F1. |
| A10.3 | Confirmed unqualified human Keychain description versus first-release file-only decision. |
| A10.9 | Confirmed first-consumer inconsistency between CIP header/plan and later Curator executor recommendation. |
| A11.5 | Confirmed mandatory DNSSEC residual prose versus explicitly allowed no-DNSSEC invitation fallback. |
| A11.9 | Confirmed raw authorized-key/host-pin recipes versus host/user CA decision. |

Seven gap samples: A04.6, A05.4, A05.5, A06.2, A08.1, A11.6, A11.7. Architecture defines keeper risk, general action registry and local bridge, but the four module contracts do not provide their standalone contracts/runbooks. CIP-0009 explicitly lists protocol and invitation/transcript/package schema publication as future specification changes (187–195); the complete tree lacks the proposed remote-worker protocol/schema artifacts. A11.7's gap is real but its D9 citation is wrong. These absence conclusions are bounded to the selected pinned repositories, not external trust repositories. No unavailable read was counted as absence.

Diagram checks (more than five): B1 user-manager launch contains UID 612 and sudo signal-relay note, inconsistent with helper UID range 30000–59999 and exec.stop ownership; B1 session-host has no diagram files; B1 launcher has no diagram files; B2 donor-side models old helper/raw-key/DNSSEC recipe; B2 deployment figure calls absent create-worker operation and lacks invitation fallback; B2 remote turn uses raw host pin/AuthorizedKeysFile; B3 complete wiki tree contains 26 book PNGs and 7 same-stem local PUML sources, so 19/26 local-source gaps holds. B4 proposed new paths are absent; existing dispatcher components/lifecycle/states and helper lifecycle/launch are correctly marked extensions, not wholly absent diagrams. Broker has five sequence files and no component/state/layout file. This reviewer did not re-render PNGs or reproduce the producer's metadata-equality checks.

D1/D2 review: contract reconciliation before diagrams is sound; sizes explicitly exclude implementation and live qualification. Helper extension, launcher amendment and goal authority dependencies are identified rather than silently inventing behavior. However A-F1/A-F2 invalidate part of the conflict basis for R-DP1/R-CS2/A-F1 and must be corrected before task decomposition. Rows generally name files and acceptance outcomes; some large grouped book tasks require their A/B cross-references, an acceptable P2 usability limitation rather than a new blocker.

Scope/hygiene: exact base-to-candidate diff contains only the 517-line research addition. LOGBOOK absent from delta. Read-only repository inspection and temporary private source cache only. No personal/local path or session-link match in the candidate; no secret material observed. Pattern scan is bounded, not universal secret certification.

Evidence acquisition: 66 unique cited textual/source files fetched independently via gh contents API at exact commit IDs, all successful; 8 recursive trees fetched, all untruncated. I did not accept producer test results as current passes. No runtime tests are applicable or authorized for this read-only research review.

## Pinned evidence for sampled rows

- A02.2: [B02:L19–25](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/02-sistema-celikom.md#L19-L25), [LR:L111–114](https://github.com/relux-works/curator-agent-launcher/blob/68be5ffe2bc9ec0fc9e8a1bb8777368c55069ded/README.md#L111-L114), [SH:L115–119](https://github.com/relux-works/swarma-session-host/blob/2bb9fda8ed923266532f331c9e692ba88b5c4cb6/spec/session-host.md#L115-L119), [CB:L403–413](https://github.com/relux-works/swarma-credential-broker/blob/f910e68f882900f03ffd036ecc2481f35bf8988b/spec/broker.md#L403-L413), [AR:L347–348](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/session-host/architecture.ru.md#L347-L348).

- A03.2: [B03:L17–22](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/03-polzovatel-na-agenta.md#L17-L22), [UM:L32–48](https://github.com/relux-works/swarma-user-manager/blob/92bdf1d59478b8fd28d2d98cfe9fb95716d10e8e/spec/helper.md#L32-L48), [DP:L255–263](https://github.com/relux-works/swarma-dispatcher/blob/9cec4115b8666aa1f3325a792359135a9df5a2f7/spec/dispatcher.md#L255-L263).

- A03.5: [B03:L40–48](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/03-polzovatel-na-agenta.md#L40-L48), [UM:L166–170](https://github.com/relux-works/swarma-user-manager/blob/92bdf1d59478b8fd28d2d98cfe9fb95716d10e8e/spec/helper.md#L166-L170), [UM:L200–209](https://github.com/relux-works/swarma-user-manager/blob/92bdf1d59478b8fd28d2d98cfe9fb95716d10e8e/spec/helper.md#L200-L209).

- A04.2: [B04:L21–31](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/04-klyuchi.md#L21-L31), [AR:L937–971](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/session-host/architecture.ru.md#L937-L971), [DP:L290–292](https://github.com/relux-works/swarma-dispatcher/blob/9cec4115b8666aa1f3325a792359135a9df5a2f7/spec/dispatcher.md#L290-L292).

- A05.6: [B05:L95–100](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/05-granty.md#L95-L100), [DP:L75–86](https://github.com/relux-works/swarma-dispatcher/blob/9cec4115b8666aa1f3325a792359135a9df5a2f7/spec/dispatcher.md#L75-L86), [CB:L193–227](https://github.com/relux-works/swarma-credential-broker/blob/f910e68f882900f03ffd036ecc2481f35bf8988b/spec/broker.md#L193-L227).

- A07.1: [B07:L3–33](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/07-host-sessiy.md#L3-L33), [SH:L105–123](https://github.com/relux-works/swarma-session-host/blob/2bb9fda8ed923266532f331c9e692ba88b5c4cb6/spec/session-host.md#L105-L123), [B06:L11–15](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/06-put-soobscheniya.md#L11-L15).

- A07.2: [B07:L37–45](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/07-host-sessiy.md#L37-L45), [SH:L14–20](https://github.com/relux-works/swarma-session-host/blob/2bb9fda8ed923266532f331c9e692ba88b5c4cb6/spec/session-host.md#L14-L20), [SH:L115–123](https://github.com/relux-works/swarma-session-host/blob/2bb9fda8ed923266532f331c9e692ba88b5c4cb6/spec/session-host.md#L115-L123).

- A07.5: [B07:L71–77](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/07-host-sessiy.md#L71-L77), [AR:L476–488](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/session-host/architecture.ru.md#L476-L488), [AR:L509–580](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/session-host/architecture.ru.md#L509-L580), [AR:L1345–1362](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/session-host/architecture.ru.md#L1345-L1362).

- A07.6: [B07:L81–90](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/07-host-sessiy.md#L81-L90), [SH:L171–184](https://github.com/relux-works/swarma-session-host/blob/2bb9fda8ed923266532f331c9e692ba88b5c4cb6/spec/session-host.md#L171-L184), [AR:L586–613](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/session-host/architecture.ru.md#L586-L613), [B08:L26–28](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/08-most.md#L26-L28).

- A08.4: [B08:L36–47](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/08-most.md#L36-L47), [C09:L155–156](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0009-donor-side-deployment-and-bridge.md#L155-L156), [C09:L175–178](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0009-donor-side-deployment-and-bridge.md#L175-L178), [AR:L1417–1426](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/session-host/architecture.ru.md#L1417-L1426).

- A09.4: [B09:L53–60](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/09-zapusk.md#L53-L60), [C07:L16–28](https://github.com/relux-works/curator-spec/blob/2f0531b4edcc99c6118c277deb00e8736392c04d/cips/CIP-0007-manager-provisioned-cli-tools.md#L16-L28), [C07:L216–244](https://github.com/relux-works/curator-spec/blob/2f0531b4edcc99c6118c277deb00e8736392c04d/cips/CIP-0007-manager-provisioned-cli-tools.md#L216-L244).

- A09.9: [B09:L95–105](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/09-zapusk.md#L95-L105), [DP:L94–99](https://github.com/relux-works/swarma-dispatcher/blob/9cec4115b8666aa1f3325a792359135a9df5a2f7/spec/dispatcher.md#L94-L99), [DP:L159–163](https://github.com/relux-works/swarma-dispatcher/blob/9cec4115b8666aa1f3325a792359135a9df5a2f7/spec/dispatcher.md#L159-L163), [DP:L248–251](https://github.com/relux-works/swarma-dispatcher/blob/9cec4115b8666aa1f3325a792359135a9df5a2f7/spec/dispatcher.md#L248-L251).

- A10.3: [B10:L31–35](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/10-uchetnye-dannye.md#L31-L35), [C10:L22–27](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0010-credentials-setup-token-and-inherited-auth.md#L22-L27), [C10:L73–93](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0010-credentials-setup-token-and-inherited-auth.md#L73-L93).

- A10.9: [B10:L65–67](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/10-uchetnye-dannye.md#L65-L67), [C10:L145–149](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0010-credentials-setup-token-and-inherited-auth.md#L145-L149), [C11:L7](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0011-credential-broker-and-agent-users.md#L7), [C11:L99–104](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0011-credential-broker-and-agent-users.md#L99-L104), [C11:L117–119](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0011-credential-broker-and-agent-users.md#L117-L119), [DP:L49–50](https://github.com/relux-works/swarma-dispatcher/blob/9cec4115b8666aa1f3325a792359135a9df5a2f7/spec/dispatcher.md#L49-L50).

- A11.5: [B11:L49–56](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/11-udalennoe-rabochee-mesto.md#L49-L56), [C09:L129–147](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0009-donor-side-deployment-and-bridge.md#L129-L147), [C09:L207](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0009-donor-side-deployment-and-bridge.md#L207), [C09:L222](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0009-donor-side-deployment-and-bridge.md#L222), [CIP-join-handshake:L11–29](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/diagrams/join-handshake.puml#L11-L29).

- A11.9: [B11:L93–95](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/11-udalennoe-rabochee-mesto.md#L93-L95), [C09:L133–145](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0009-donor-side-deployment-and-bridge.md#L133-L145), [C09:L157–158](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0009-donor-side-deployment-and-bridge.md#L157-L158), [C09:L191–195](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0009-donor-side-deployment-and-bridge.md#L191-L195), [CIP-remote-worker-turn:L21–25](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/diagrams/remote-worker-turn.puml#L21-L25).

- A04.6: [B04:L61](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/04-klyuchi.md#L61), [AR:L929–954](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/session-host/architecture.ru.md#L929-L954).

- A05.4: [B05:L54–56](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/05-granty.md#L54-L56), [AR:L1148–1161](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/session-host/architecture.ru.md#L1148-L1161), [CB:L231–235](https://github.com/relux-works/swarma-credential-broker/blob/f910e68f882900f03ffd036ecc2481f35bf8988b/spec/broker.md#L231-L235).

- A05.5: [B05:L60–91](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/05-granty.md#L60-L91), [AR:L1181–1275](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/session-host/architecture.ru.md#L1181-L1275).

- A06.2: [B06:L23–29](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/06-put-soobscheniya.md#L23-L29), [AR:L330–394](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/session-host/architecture.ru.md#L330-L394), [AR:L1366–1380](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/session-host/architecture.ru.md#L1366-L1380).

- A08.1: [B08:L3–14](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/08-most.md#L3-L14), [AR:L86–125](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/session-host/architecture.ru.md#L86-L125), [AR:L284–411](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/session-host/architecture.ru.md#L284-L411), [W-bridge-components:L1–48](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/src/bridge-components.puml#L1-L48).

- A11.6: [B11:L58–69](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/11-udalennoe-rabochee-mesto.md#L58-L69), [C09:L129–132](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0009-donor-side-deployment-and-bridge.md#L129-L132), [C09:L149–156](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0009-donor-side-deployment-and-bridge.md#L149-L156), [C09:L189–195](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0009-donor-side-deployment-and-bridge.md#L189-L195), [CIP-join-handshake:L20–46](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/diagrams/join-handshake.puml#L20-L46).

- A11.7: [B11:L73–81](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/11-udalennoe-rabochee-mesto.md#L73-L81), [C09:L149–150](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0009-donor-side-deployment-and-bridge.md#L149-L150), [C09:L191–195](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0009-donor-side-deployment-and-bridge.md#L191-L195).

Corrected CIP locators: [D9/D10/D11](https://github.com/relux-works/curator-spec/blob/1604d4022f97d6f0751e17a37c549d2a10581829/cips/CIP-0009-donor-side-deployment-and-bridge.md#L160-L170).

## Role correction and handoff

The injected prompt contained both A and B briefs. I initially followed B, but the task run metadata identifies this run as R138 reviewer A (record-only). This record supersedes the earlier B-labelled artifacts. Findings remain independently obtained. No acceptance/rejection or verdict-status mutation has been made; reviewer B must perform its own review and reconcile this record. The absent A record was this run's pending deliverable, not an external blocker. No cross-provider reconciliation is claimed.

Reviewer B should verify every P1 above and decide acceptance or research rework. No product tests/builds were run and no source or LOGBOOK edits were made.
