# BUG-261009-2s6t3y: legacy-movedtag-false-refusal-after-old-tag-deleted

## Description
From the TASK-260924-4mzun5 revision-4 review (P2-BUG-1, non-blocking under tb-R226): install a schema-9 consumer at v1 under a schema-1 Skillfile; create v2 at another commit; delete the old v1 tag; change the Skillfile declaration to v2. v2 never moved, yet both production install entries (project and global) refuse with: moved tag for consumer: v2 old -> new. Cause: the moved-tag check infers declaration continuity from live tags on the old commit. Reproduction: reviewer probe TestReviewR4ChangedTagAfterOldTagDeleted (resource TASK-260924-4mzun5_review-install-probes-rev4.go). Expected to be fixed by the F4 rework of TASK-260924-4mzun5 (comparing the validated previous declared ref instead of live tags); close this BUG with that landing if the probe becomes a passing permanent regression, otherwise fix it separately.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
An explicit declaration change to a tag that never moved installs on project and global entries after the old tag is deleted; a permanent regression covers it; same-tag movement under StrictTags is still refused.
