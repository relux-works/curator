# BUG-261021-4xrg5d: projection-atomic-link-swap

## Description
Follow-up P2 F5 from BUG-261021-2qv4zj review r5: during repair of a damaged skill projection, shared skill links briefly return ENOENT between the backup and install renames (never a mixed tree). Stage the projection at a new path and atomically swap the link (rename a fresh symlink over the old one).

## Scope
(define bug scope / affected area)

## Acceptance Criteria
(define fix acceptance criteria)
