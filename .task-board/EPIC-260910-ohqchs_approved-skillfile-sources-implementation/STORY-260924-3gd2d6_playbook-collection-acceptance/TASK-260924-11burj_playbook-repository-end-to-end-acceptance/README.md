# TASK-260924-11burj: playbook-repository-end-to-end-acceptance

## Description
End-to-end acceptance through the real CLI with schema 2 on by default: a local bare playbook repository with skills/orchestrator, skills/developer, ... installed by one collection entry {from, directory: skills, include: [*]} via install and update with lock and audit; one of those skills declares a manifest dependency with directory on a subfolder skill of a second repository; lock replay on a fresh machine; release-notes entries for both.

## Scope
(define task scope)

## Acceptance Criteria
both scenarios pass through install and update with lock and audit; fresh-machine replay from the committed lock; release notes cover both
