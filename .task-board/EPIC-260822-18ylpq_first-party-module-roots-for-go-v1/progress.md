## Status
done

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
CURATOR QUALIFIED for schema-8 module roots: run 32689488293 (workflow_dispatch, main at a3abcf3) — conclusion success, zero non-green jobs, Candidate suite green on ubuntu/macos/windows against the final candidate 6001dc33281b94a4ec7442ab15278550dd0f51d9 (manifest sha256 803918bf...), WITH consumption assertions in place (family-removal proof: deleting a schema-8 family from the root fails the suite). Implementation trail: PR 33 parsing+bijection, PR 34 fail-closed script policy admission, PR 35 driver module roots, PR 36 suite-plan serve fix, PR 37 consumption coverage. Remaining before the skill switch: cocoaskills qualification (their board STORY-260822-27ze8z), the atomic spec landing with pin advances, rc.9 publish, SPEC_PIN bump, curator release.

## Precondition Resources
(none)

## Outcome Resources
(none)

## Created
2026-08-22T15:46:47Z

## Last Update
2026-08-25T02:47:21Z
