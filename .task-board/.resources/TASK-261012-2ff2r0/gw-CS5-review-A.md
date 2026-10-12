# TASK-261012-2ff2r0 — apply-f23-operator-custody-cs: reviewer A

Record-only review of CR-TASK-261012-2ff2r0-1 revision 1. Reviewer B owns the decision. No status, accept_cr or reject_cr mutation was made, following the specific gw-CS5-review-A-brief.md hard rule over generic lifecycle instructions.

Base: `365c687f457d587a2da9892613c9bda97e7baf84`.
Candidate tree: `e6352bc4f8b421cfd243e8d327530a5c22ab71eb`.

## Findings

```yaml
findings: []
```

No actionable bypass, regression or robustness defect found in the reviewed documentation delta. This is a source-review result, not runtime acceptance or a CR verdict.

## Swept surfaces

| Surface | Evidence and result |
|---|---|
| Dedicated custody and principal separation | CIP-0011 §5.1 pins swarma-operator as broker operator/account owner, helper operator UID and operator signing authority; forbids harness sessions and separates interactive accounts. Protected software root and direct-UID mutations are covered. |
| Owner phone amendment | §5.2 specifies secure-element, non-exportable, biometric operator device key; key-only SSH over tailnet, forced operator command/no shell, exact displayed canonical act, pinned verification and no cross-act reuse. Revocation uses one biometric confirmation. |
| Terminal fallback | §5.2 requires password sudo, timestamp_timeout=0 and no NOPASSWD; fallback works without a phone. Dispatcher helper/launcher exception remains explicit. |
| Command/read boundary | §§5.1–5.3 route enrol/grant through authentication and content confirmation, revoke/disable through authentication, and status/grants/accounts/bindings through mandatory swarma-observer without act confirmation. Agent-session widening refusal covers signing and direct-UID paths. |
| Exact owner widening list | §5.3 includes all eight categories, including standing retirement/re-creation and standing renewal; preserves bounded service authority without inferred infinite re-grants. |
| Slice 0 and unresolved contracts | Implementation plan item 2, Test plan operator row and Open question 9 require publication/negative qualification; no runtime success inferred from a dedicated account alone. Command binary/wire, signed-record domain/canonical bytes, key admission and phone lifecycle remain unpublished. |
| Shared names and cross-CIP table | Disposition Revision 4, R-OPERATOR schema row, Operator authority boundary and open questions agree with the owner amendment; earlier shared domains are retained and no new wire fields/refusal values were added. |
| Figures and links | Reviewed all 12 changed PlantUML sources, including new operator-custody; each is linked in diagrams/README.md. The audit view removes analyzer execution under operator UID. Configuration/tool views separate caller execution from operator authority. Existing donor variants and C4 source do not assert a contrary operator account/signing path. |
| Revision and scope | CIP-0011 remains Draft revision 6; disposition revision 4 and figure notes are present. All 15 candidate paths are documentation. |
| Gap outcome | TASK-261012-2ff2r0_gaps.md maps F-23 and owner amendment to actual sections/figures and accurately lists unpublished work. No false runtime qualification claim found. Author-reported checks are not treated as reviewer execution evidence. |
| Public hygiene | Read the complete added delta: no personal first names, real host names, machine-local absolute paths, session URLs or board-resource links introduced. Fixed public service names and existing normative installation paths are not private machine identifiers. |

## Neighbouring specifications

Compared current wave-4b working drafts of swarma-credential-broker §§4.2–4.3, 14 and relevant operator/open-question rows; swarma-user-manager §§2–3, 4.3, 11 and question 17; swarma-dispatcher §§3.2, 3.4, 5.4–5.5, 8.1, 11 and related publication/qualification rows. They agree on separate operator custody, phone-first records, terminal fallback, observer reads, exact confirmation list and unpublished integration. These companion worktree drafts are contextual cross-checks, not evidence of landed implementations.

## Verification and bounds

- Reviewed exact base-to-candidate Git delta, all 15 paths; read CIP context, figure index and attached gap report.
- Independently ran `git diff --check` for the exact base/candidate pair: exit 0.
- No tests, builds or renderer run, per task instruction. PlantUML inspected as source only; rendered output and runtime behavior remain unverified. No hosted validation result is attested by this review.
- `task-board spawn goal` reports no active run goal; directive checkpoint reports none.
- No repository modifications, commits or LOGBOOK.md writes. Findings persisted as this board outcome.

Recommendation to reviewer B: no source-review changes requested from reviewer A; make the CR decision using the full required evidence and reviewer B's independent review.
