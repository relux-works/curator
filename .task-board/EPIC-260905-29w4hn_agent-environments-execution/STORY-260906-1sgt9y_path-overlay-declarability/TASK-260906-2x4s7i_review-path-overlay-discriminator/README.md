# TASK-260906-2x4s7i: review-path-overlay-discriminator

## Description
Adversarially review the repaired source-kind discriminator that makes the section 6 path overlay declarable, on PR #47 (branch feat/path-overlay-declarable at bd39adb, one signed commit past curator-spec main). Cycle 1 on TASK-260906-3x0w4y found the discriminator classified a Windows absolute path as git and had no case that could fail a wrong discriminator; both were addressed and the fix itself is what this review attacks. The review artefact is the PR, not a story-branch candidate: the previous story workspace drifted and its Change Request record is unusable, so no CR is captured for this element.

## Scope
Read-only review of git diff origin/main..bd39adb in curator-spec. Authority: environments.md sections 1, 6 and 12.1, core.md 6.1, profiles/manager.md, schemas/v1/manager-config-v2.schema.json.

## Acceptance Criteria
Every dimension of the attached review brief is driven against the committed schema rather than against the pattern alone, the residual edges are decided with the sentence that decides them, the new cases are tested as a gate by mutating the discriminator and observing a named case flip, and the verdict states plainly whether PR 47 is safe to land.
