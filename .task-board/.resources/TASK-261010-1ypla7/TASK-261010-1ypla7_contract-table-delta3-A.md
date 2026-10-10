# Delta reviewer A — revision 3 — record only

Task: TASK-261010-1ypla7 — research-platform-contract-table.
Recommendation to deciding reviewer: no remaining P0/P1 in the reviewed delta; one P2 consistency note. This record does not accept or reject the Change Request.

Reviewed candidate tree `9c4b56dc03d120fca857d2b57b0a19af19b6dac8` against revision 2 tree `3a9b0d864af6b1398d07759277d90597344050f0`; also checked changed paths against base `67f736f5e45f562df892674cb53c97b0e6d732e2`. Candidate file SHA-256: `b50becef762053c8b40b6b11853546ec1ba3deee2652ceca97f1e94a4ecb29b6`.

## Swept surfaces and prior findings

| Surface / previous mechanism | Result and evidence |
|---|---|
| DB-F1, P8 proposal promoted to canon | Resolved. Both same-version-only and qualified-range/probe options explicitly remain OWNER DECISION NEEDED; neither is canonical. AR proposed/open status, recommendation register and book alternative now cited. Common preconditions explicitly remain from the proposal. |
| DB-N1, broken GFM tables | Resolved. Removed both blank lines. I5/W4 remain contiguous with preceding table; P7–P11 remain contiguous with bounds/ownership table. Each affected row has six cells; headings and separator rows retained. Structural inspection, no renderer run. |
| DB-N2, W4 owner section | Resolved: CB §§4.3/6.6. CB L233 names §6.6 and L235 supplies versioned adapter/platform precedence. |
| DB-N3, P9 exposure decision provenance | Resolved per newer owner brief. Real 2026-10-10 acceptance is explicit, attributed to B07 and rework3 precondition, without claiming a pinned session record. Private lane 1/2 layout remains unchanged. Acceptance of possible log exposure does not move CI lanes. |
| DB-N4, P11 unnecessary owner question | Resolved. Internal service-only sweep is the default reading of DP §7.2; only a new public operation requires a decision. DP L211 timer/account/cadence and L215–232 closed caller operation table support this distinction. |
| Other changes in meaning | None outside requested corrections: open-decision register follows P8/P11, and a rev3 changes section records the fixes. Remaining rows unchanged byte-for-byte by Git delta inspection. |
| Size / scope / hygiene | 78,890 bytes ≤ 81,920. Exactly one candidate path: `.research/261010_platform-contract-table.md`. No LOGBOOK or code changes. Reviewed added prose and bounded whole-file scans found no personal paths, first names or credential tokens. URL matches containing “session” were repository source links, not session links. Source personal names were not copied into this artifact. |

## Independent pinned-source reads

Six of six read-only GitHub contents requests succeeded (exit 0), using `gh api repos/relux-works/<repository>/contents/<file>?ref=<full commit>` with raw content acceptance. Read the following ranges directly, not merely link-checked:

- [AR §5.3](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/session-host/architecture.ru.md#L413), L474–488, L1814–1825 and L1973–1977: proposed continuity rules, open recommendations and status definitions.
- [B07](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/book/swarma/ru/07-host-sessiy.md#L71-L77), L119–121: same-version alternative; owner-attributed public-log acceptance and conflicting lane placement.
- [CB](https://github.com/relux-works/swarma-credential-broker/blob/f910e68f882900f03ffd036ecc2481f35bf8988b/spec/broker.md#L229-L235): section boundary and trust adapter contract.
- [DP](https://github.com/relux-works/swarma-dispatcher/blob/9cec4115b8666aa1f3325a792359135a9df5a2f7/spec/dispatcher.md#L209-L232), also L75–81: sweep ownership/cadence, closed public operations and caller-key context.
- [SG](https://github.com/relux-works/swarma-session-host/blob/2bb9fda8ed923266532f331c9e692ba88b5c4cb6/docs/sh1-gate.md#L3-L20): private/public lanes and tooling proof limits.
- [SH](https://github.com/relux-works/swarma-session-host/blob/2bb9fda8ed923266532f331c9e692ba88b5c4cb6/spec/session-host.md#L69-L97): lane contracts and immutable-tag cutover conditions.

Session date/authority derives from `contract-table-rework3-brief.md`; B07 corroborates acceptance, not the session date. D-GOALS/D-IDENTITY remain unchanged. Prior wholesale citation checks are historical evidence only; this review independently checked the rev2→rev3 delta, not all 52 rows anew. No tests, builds, producer check scripts, commits or LOGBOOK edits were run. No runtime capability is claimed. Run goal query returned no active goal.

## Findings

```yaml
findings:
  - id: D3A-N1
    severity: note
    severity_reason: "P2: provenance-summary consistency only; P9 itself discloses its authority correctly and no contract choice is hidden."
    repeat_of: null
    location: ".research/261010_platform-contract-table.md:11 and row P9 at line 105"
    mechanism: "The introduction says only D-GOALS and D-IDENTITY cannot be cited at a Git commit. P9 now also expressly cites a session decision/date confirmed by the rev3 owner precondition and states that no Git-pinned session record is claimed. The global exclusivity wording is therefore stale."
    recommendation: "Qualify the introductory sentence as referring to those two overrides, or enumerate the P9 session-date/confirmation exception alongside them. Retain B07 as the pinned corroboration of acceptance."
```

Repeat-of audit: DB-F1 and DB-N1–DB-N4 are resolved, as detailed above. D3A-N1 is a new summary-versus-row inconsistency introduced by the newly explicit P9 provenance, not recurrence of the missing P9 owner attribution mechanism.

## Routing

Latest delta3-A brief explicitly requires RECORD-ONLY and forbids accept_cr/reject_cr. Attached this task-scoped outcome for the deciding review; no acceptance/rejection or final-status mutation performed. This specific instruction supersedes the generic reviewer lifecycle branch for this A pass.
