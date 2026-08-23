## Status
backlog

## Review
required

## Task Class
code

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
(empty)

## Notes
Driver: third-party adoption (tool repos with lockstep local modules are common; one-module-per-build-root is a packaging shape requirement that costs adoption). First local consumer: skill-project-management. CORRECTED PREMISE (was wrong, caught by RUN-260822-c8550c review): its agent-skill.json has always declared build/go-v1 commands — it never shipped a type=system manifest. The previously packaged revision ca5c4fd3 was installable because its snapshot shipped with the replace directives STRIPPED and the first-party modules pre-vendored as ordinary v0.0.0 requirements (zero replace lines in go.mod and modules.txt). Current main diverged from that shape: directory-form replaces restored (2ed3acd) and tools/*/vendor gitignored (TASK-260819-3vr8j3), so main is unbuildable under go-v1 for two independent reasons — no vendor in the snapshot, and replaced modules rejected. Fresh vendor trees WITH replace directives (the shape decision 0009 admits) are parked on skill-project-management branch task/go-v1-switch at origin (36c1e02), together with a make vendor/vendor-check drift gate and CI job; the auto-return task there is TASK-260822-hje0ya. Note per-module reality: tools/board-cli requires all three pkg modules; tools/board-tui requires pkg/remoteconfig only. Decision 0009 landed (b92b105); review requested a follow-up decision-only amendment — see TASK-260822-1yz9ug review-verdict (F1 factual corrections, F2 laundering-path residual). Implementation stories follow after the amendment and the normative story.

## Precondition Resources
(none)

## Outcome Resources
(none)

## Created
2026-08-22T15:46:47Z

## Last Update
2026-08-23T10:07:11Z
