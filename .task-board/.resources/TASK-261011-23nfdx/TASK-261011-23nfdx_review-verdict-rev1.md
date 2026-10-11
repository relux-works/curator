# TASK-261011-23nfdx — R-CS3 diagrams — reviewer B verdict, CR rev1

Verdict: **ACCEPTED**. No P0/P1 remains. P2 items are notes under tb-R226 for a follow-up pass.

Target: CR-TASK-261011-23nfdx-1 revision 1, base f61feedffbddba8b9123cb9051547c5013da71cb, candidate tree cf9b9d78ef2fa80ca9ba2fc8b482b0a74a71ac1e (26 docs/diagram paths, no code). Sources: contract table 177cb554, audit 48789217 (B4 M-CIP1/M-CIP2, D1 R-CS3), owner decisions precondition (including D-PATH, D-WIRE-NAMES, D-DONOR-PROVISIONING, D-EXECUTOR, D-IDENTITY).

## Order of work
Checks 1-5 were done first, independently, before reviewer A's record was opened; A was then read and each of its findings re-checked against the spec text.

## Independent checks
1. **Figures exist and match spec.** All 12 PlantUML sources and the retained C4 source exist; M-CIP1 (donor-side, first-deployment, join-handshake, turn, retire extended; variants, lifecycle, layout new) and M-CIP2 (credential-source-lifecycle, configuration-publication, tool-admission, audit-transport) are all present. Read each against CIP-0009 D0-D12, CIP-0002/3/4/5/6/7/10/11. Spot-verified distinctive claims against the text (CIP-0005 65,536/65,537 byte cap, 1 MiB / 64 KiB bounds, env deny list; CIP-0011 lease.renew within the response deadline; CIP-0010 revoke semantics and expiry; CIP-0006 / CIP-0002 error names; CIP-0004 no-source-v1). Each figure carries a phase line, process/UID identity, state-owner note, refusal/recovery branches and spec links.
2. **Owner decisions applied exactly.** D-PATH order and the X1 dedupe/position rule restated (F1); D-DONOR-PROVISIONING: B drawn disabled, provisioner separate from swarma-user-manager, no donor op added to its schema; D-WIRE-NAMES names used only where the table/owner gave them; rows derived by rule are labelled "derived by rule, pending matching module vectors" (F3); D-IDENTITY: donor identity does not inherit the one-shot revoke exception; D-EXECUTOR: agent-account curator-run shown as first broker consumer with board/session host not embedding a broker client. No recommendation shown as a decision (A selection after B/C, storage UIDs, filenames, retention all marked unspecified).
3. **No behaviour beyond spec.** Gaps are visible notes ("unspecified in spec ..."). Minor inferences listed as notes below.
4. **Outcome file** TASK-261011-23nfdx_figures.md: 12 PlantUML + 1 C4, figure/file/section/audit-row mapping agrees with the tree; validation statement is honest (author ran nothing; handoff CR gate: exit 0, 73 schemas, 1294 vectors, 689 tests).
5. **Docs only.** git status shows only README, cips/* paths. No local paths, session links, first names or LOGBOOK writes in the added lines (grep).

## Mechanical checks I ran (no renderer on this host)
- Link/anchor check over README, cips/diagrams/README, all .puml `[[...]]` links and every CIP + disposition: 0 missing files, 0 bad anchors (checker self-tested with a deliberately bad anchor). All protocol/remote-worker.md anchors cited exist.
- PlantUML structural balance (alt/loop/break/box/note vs end, braces, @startuml/@enduml): balanced in all 12 files; `!theme` removed from the four extended sequences, so syntax is conservative. Rendering itself is NOT verified (bound: no renderer).
- 10/10 operative summaries (CIP-0002..0011) present with purpose/boundary/authority/I-O/phase/links; disposition blocks retained.

## Reconciliation with reviewer A
A raised four findings, all classified P2 by A. I re-checked each against the spec text; all four are confirmed as real figure-vs-prose gaps, none rises to P0/P1:
- A-F1 (confirmed, P2): configuration-publication and tool-admission show the D-PATH transform / absent-PATH refusal unconditionally; CIP-0002 separates context admission from command transport, and the script-worker-v1 lane discards inherited PATH (CIP-0007).
- A-F2 (confirmed, P2): hook cache publication is drawn before the unknown/custom-cache classification (CIP-0004 compat item 3).
- A-F3 (confirmed, P2): tool-admission says store/publication recovery is "unspecified" although CIP-0007 Specification changes item 8 specifies staging, journaled store/marker targets and rollback distinction. The note under-claims; nothing invented.
- A-F4 (confirmed, P2): the retained actor label "operator or orchestrator with onboard grant" omits CIP-0009 D4 step 5's "plus operator confirmation". Pre-existing label; the figure's admission step still names operator checks.

## Findings
```yaml
findings:
  - id: B-N1
    severity: note
    severity_reason: P2 — small inference in the lifecycle figure
    where: cips/diagrams/donor-lifecycle.puml (Left --> Stopped on init --rejoin; "new invitation")
    detail: CIP-0009 D7 says rejoin returns the pinned identity to a new workplace/package unless revoked; the target execution state (Stopped, wait for run) and "new invitation" wording are not in the spec. "new invitation" is pre-existing in remote-worker-retire.puml.
  - id: B-N2
    severity: note
    severity_reason: P2 — wording beyond the literal X1 rule
    where: cips/platform-contract-disposition.md D-PATH item 3; CIP-0002 D-PATH paragraph
    detail: Adds "otherwise append it" to the X1 rule (dedupe by identity; preserve existing position). It follows from the owner's order (dispatcher third) so it is not a contradiction, but it is an interpretation; the grammar stays marked pending.
  - id: A-F1
    severity: regression
    severity_reason: P2 — unconditional PATH transform/refusal in two figures
  - id: A-F2
    severity: regression
    severity_reason: P2 — cache publication drawn before classification
  - id: A-F3
    severity: robustness
    severity_reason: P2 — specified publication transaction shown as "unspecified"
  - id: A-F4
    severity: bypass
    severity_reason: P2 — retained label omits operator confirmation for orchestrator admission
```

Recommended follow-up (not blocking): fold A-F1..A-F4, B-N1, B-N2 into one small figure-polish pass.

## Bounds
Rendering, native behaviour and runtime qualification were not exercised; "matches the spec" means documentary correspondence. Gate evidence is the attached handoff validation log (not rerun by me).
