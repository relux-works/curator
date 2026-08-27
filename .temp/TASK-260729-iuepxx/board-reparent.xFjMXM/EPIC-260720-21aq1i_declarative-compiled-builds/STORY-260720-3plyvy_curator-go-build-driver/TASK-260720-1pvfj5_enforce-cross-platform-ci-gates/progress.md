## Status
backlog

## Assigned To
(none)

## Created
2026-07-20T02:11:49Z

## Last Update
2026-07-20T03:01:20Z

## Blocked By
- TASK-260720-2qqq0w
- TASK-260720-jrrgw9

## Blocks
- TASK-260720-38l1sy
- TASK-260720-3pvihp
- TASK-260720-z2z795
- TASK-260720-z9j4c9

## Checklist
- [ ] CI pins the reviewed immutable rc.4 protocol commit
- [ ] Linux, macOS, and Windows exercise their platform-specific behavior
- [ ] Race, vet, formatting, lint, and acceptance evidence are required
- [ ] Candidate suite input is explicit and never advances or impersonates the qualified released pin
- [ ] Keep every default committed protocol pin on the previous release; supply the candidate suite only through a non-default immutable input

## Notes
Cross-story release boundary from STORY-260720-21bsr2: candidate rc.4 tests may use an explicitly supplied CURATOR_CONFORMANCE_ROOT, but the committed curator-spec release pin must not move until TASK-260720-25d05o qualifies the actual published protocol release. TASK-260720-38l1sy audits the resulting pin and no-skip gate. Do not substitute a merely landed or reviewed but unreleased commit.
Cross-story checklist clarification 2026-07-20: inherited checklist item 1 uses stale pin language. It is superseded by the current description, scope, AC, and checklist items 4-5: candidate qualification uses an explicit immutable non-default input, while the committed suite pin remains on the previous release until TASK-260720-25d05o and TASK-260720-38l1sy.

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260720-1pvfj5_candidate-release-ci-gates.puml](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_candidate-release-ci-gates.puml) — PlantUML source separating candidate CI evidence from official released-suite pin promotion
- [TASK-260720-1pvfj5_candidate-release-ci-gates.svg](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_candidate-release-ci-gates.svg) — Rendered candidate CI and released-suite evidence gates
