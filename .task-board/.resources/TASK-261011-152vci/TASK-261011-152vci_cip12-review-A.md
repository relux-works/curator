# cip12-review-A.md
Task: TASK-261011-152vci — cip-0012-symlinked-seed-files-rejected
Review: A of two, RECORD-ONLY, CR-TASK-261011-152vci-1 revision 1.
Base: 165d1ff45d1e0b74ccc67f932054c636bf770bf7
Candidate tree: 390ed4bd63a7e244e39e4bc606ed9e606a64dc8e

## Findings
- severity: note
  severity_reason: P3 informational; no actionable defect found in the six requested review surfaces.
  evidence: Exact candidate delta contains the new 196-line CIP and one README index row only. Reviewer B owns the CR decision.

## Swept surfaces
| Surface | Result and evidence |
| --- | --- |
| Template and disposition | Pass. All CIP-0001/TEMPLATE headings and metadata exist; operative summary covers purpose, boundary, authority, inputs/outputs and normative links. Exact requested provisional rejection and decision date present. |
| Normative facts and citations | Pass. Environments section 7.4 at lines 1610–1708 establishes the seed classes, Codex A/B behavior, one-time state and non-regular refusal; section 8.4.1 at lines 2193–2222 confirms environment_seed_unreadable before first write. Sections 8.2 and 10.1 support marker and stale/repair references. All four linked section anchors correspond to actual headings. |
| Options and decision | Pass. Options a–e include costs/security: migration propagation loss and installer recurrence; constrained same-owner/HOME admission and Nix limitation; insufficient caller harm reduction; silent-follow and absence fallback rejection. Operator decision is neutral and matches the supplied facts. Incident behavior is explicitly attributed to the supplied account, not represented as measured runtime evidence. |
| Revisit-only design | Pass. Implementation and test sections explicitly label future conditions, not scheduled work. Ownership, HOME containment, regularity, permissions, one-hop/no-follow, descriptor-bound race handling, proposed marker identity and later drift are covered. All seven required negative shapes appear, plus dangling-link and world-write cases. No new refusal codes or adopted marker fields. |
| Scope and index | Pass. README row says Rejected. Exact-tree diff changes only CIP-0012 and the index; protocol/environments.md and all existing CIPs unchanged. Specification changes explicitly says None. |
| Public hygiene | Pass with explicit brief interpretation: Owner uses the exact required project handle ivan-curator, also present in the binding template; no additional personal names appear. No host identifiers, absolute local paths, session links or board-resource links in the candidate delta. Home-relative examples are permitted. |

## Verification and limits
Read-only inspection of the exact Git delta and cited repository sources; current CIP blob matches candidate blob 5a36168524ece946b35d3149791e30d6f4ebccd8, and README has no difference from candidate. No tests or builds run, as instructed. No runtime incident reproduction or migration attestation; no validation-suite results inferred.
spawn goal reports no active goal for this run.
No repository or LOGBOOK edits. No status mutations and no accept_cr/reject_cr calls: the explicit reviewer-A hard rule overrides the generic lifecycle instructions. This report is evidence for reviewer B, not a persisted acceptance verdict.
