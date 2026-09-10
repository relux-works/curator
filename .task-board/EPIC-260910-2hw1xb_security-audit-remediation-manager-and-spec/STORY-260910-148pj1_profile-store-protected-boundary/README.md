# STORY-260910-148pj1: profile-store-protected-boundary

## Description
Finding S5 (Medium): the environments store, locks and markers lack the protected-boundary discipline the build cache has; lock-free resolve trusts link-target identity alone, so a same-user swap of store bytes passes without detection.

## Scope
curator-spec environments + internal/envprofile/contextstore

## Acceptance Criteria
env resolve validates ownership/permissions/containment of the environments root and store entries like core 9.3; surface hashes of system-prompt and root-context files are verified at resolve; drift is reported
