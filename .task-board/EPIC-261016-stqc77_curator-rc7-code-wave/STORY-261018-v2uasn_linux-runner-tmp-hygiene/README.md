# STORY-261018-v2uasn: linux-runner-tmp-hygiene

## Description
Self-hosted Linux runner relux-rise1-curator-4 fails every job in its job-completed cleanup hook because a pnpm conformance test left a read-only store tree in the runner tmp. Repository-side fix: the test restores write permission on what it created, and Linux self-hosted jobs make their own TMPDIR writable at the end so stale leftovers are healed.

## Scope
(define story scope)

## Acceptance Criteria
(define acceptance criteria)
