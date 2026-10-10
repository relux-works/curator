# Delta review A — platform contract table, revision 2

Task: TASK-261010-1ypla7 — research-platform-contract-table.
Mode: RECORD-ONLY under contract-table-delta-A-brief.md. Recommendation: no remaining P0/P1 in the reviewed delta; deciding reviewer retains disposition authority. No accept_cr/reject_cr or terminal status mutation performed.

Reviewed revision 1 tree `645ad50d014a3fe443a9d6b9d50d35a39f6bc679` against revision 2 tree `3a9b0d864af6b1398d07759277d90597344050f0`; base `67f736f5e45f562df892674cb53c97b0e6d732e2`. Read the prior deciding verdict resource in full. Run goal query returned no active goal.

## Swept surfaces

| Surface | Result | Prior mechanism / repeat-of |
|---|---|---|
| F1, retirement fate, P7 | Fixed. Request archive/delete, policy operator-only keep, caller roles, defaults and propagation are present. Unresolved unified enum is explicitly an owner decision with both real options. | rev1 F1: missing canonical retirement contract; resolved |
| F2, resume, P8 | Fixed. Qualified range plus adapter probe and other preconditions, no transport conversion or inferred fallback; exact equality is not substituted for qualification. | rev1 F2 / A07.5: missing compatibility contract; resolved |
| F2, gate lanes, P9 | Fixed. Private projection and consumer versus public module lanes, runner ownership and public-evidence limits match sources. | rev1 F2 / A07.8: missing CI ownership contract; resolved |
| F2, public module, P10 | Fixed. Explicit supersession of private-fork cutover; serving library versus board executable; future tag consumption distinguished from tooling already present. | rev1 F2 / A07.10: missing consumption disposition; resolved |
| P2 helper ledger and key classes, I5/W4 | Addressed. Exact default locator, read refusal, separate caller/software-root/keeper identity classes and versioned migration boundary. W4 owner cell calls the migration section 6.5, but the linked platform-forms subsection is 6.6; see note below. | rev1 A-F3: coverage gaps; resolved substantively |
| P2 sweep entry, P11 | Addressed. v0 timer/v1 internal sweep and missing public operation are accurately separated; no public authorization invented. | rev1 A-F3: missing sweep boundary; resolved |
| P2 alias and goals | LR now maps to launcher README at the LS commit. G1 adds the requested explicit board-binding rewrite while keeping its canonical value unchanged. | rev1 A-N1 and B additional G1 note; resolved |
| Other semantic changes | Seven contract rows added. All prior contract rows are byte-identical except G1's requested mismatch-cell clarification. Source-index compression retains scope; baseline/evidence prose now distinguishes historical reads from current rework. | No new normative drift found |
| Hygiene and size | Only requested research file differs from CR base. 76,851 bytes of 81,920 allowed (5,069 spare). No LOGBOOK changes. Manual review and bounded scans found no personal paths, first names, secrets or session links. Platform installation paths are contract values. | No repeat |

## Fresh source evidence

Eight of eight authenticated, read-only `gh api repos/<repo>/contents/<file>?ref=<full commit>` reads succeeded. Exact canonical ranges and contradiction ranges inspected:

- UM `92bdf1d59478b8fd28d2d98cfe9fb95716d10e8e`, `spec/helper.md`: L55–74, L118–154 (P7/I5).
- AR/wiki `2e30edbd54977b306ab53497663fcd94a5ca6cb6`, `session-host/architecture.ru.md`: L474–488, L929–954 (P8/W4).
- SH `2bb9fda8ed923266532f331c9e692ba88b5c4cb6`, `docs/sh1-gate.md`: L3–20; `spec/session-host.md`: L3–5, L50–52, L69–97, L202–209 (P9/P10).
- DP `9cec4115b8666aa1f3325a792359135a9df5a2f7`, `spec/dispatcher.md`: L75–81, L209–232 (W4/P11).
- CB `f910e68f882900f03ffd036ecc2481f35bf8988b`, `spec/broker.md`: L92–104, L231–235 (I5/W4).
- Wiki at the same pin: `book/swarma/ru/07-host-sessiy.md` L71–77, L119–121; `book/swarma/ru/12-doska.md` L49 (P8–P10 mismatches).

Coverage: 7/7 added canonical rows semantically checked against their pinned sources; 4/4 mandatory F1/F2 contract additions verified. This is a delta review, not a fresh whole-table audit. Earlier whole-table findings are retained only as prior review evidence. LR index addition was compared to the existing LS repository/commit, not independently fetched. The refreshed audit retrieval status is producer provenance, not independently re-established here. The disclosed D-GOALS/D-IDENTITY resource-versus-commit exception is unchanged.

## Findings

```yaml
findings:
  - id: DA-N1
    severity: note
    severity_reason: "P2: formatting only; canonical content remains readable in source and no contract is missing."
    repeat_of: null
    location: ".research/261010_platform-contract-table.md:88-90,104-109"
    mechanism: "Blank lines terminate the preceding Markdown tables; the added I5/W4 and P7-P11 blocks have no header/delimiter, so GFM renders them as paragraphs rather than six-column table rows."
    recommendation: "Remove the separating blank lines or repeat the table header and delimiter for each added block."
  - id: DA-N2
    severity: note
    severity_reason: "P2: owner subsection label is stale; pinned line citation and canonical value are correct."
    repeat_of: null
    location: ".research/261010_platform-contract-table.md:90 (W4)"
    mechanism: "Owner cell says CB 6.5 migration; platform-forms/versioned-adapter text at cited L233-L235 is section 6.6, with migration procedure in 4.3."
    recommendation: "Use CB sections 4.3/6.6 in the owner cell."
```

No tests, builds, commits, source edits or LOGBOOK edits were performed. Only read-only source/document inspection and local temporary review artifacts were used. Record-only brief overrides generic terminal-verdict routing; this report is for the deciding review, not acceptance.
