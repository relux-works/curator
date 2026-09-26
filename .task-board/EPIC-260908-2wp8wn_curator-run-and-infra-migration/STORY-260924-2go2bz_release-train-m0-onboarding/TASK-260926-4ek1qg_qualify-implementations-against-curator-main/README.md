# TASK-260926-4ek1qg: qualify-implementations-against-curator-main

## Description
The spec Implementations workflow pins relux-works/curator at a3abcf34 (PR 37). PR #88 (hard-link erratum, lockstep) needs curator main 0a628621 (TASK-260922-18ex37 classifies executable_identity_cases and hard_link_substitution_definition). With that pin (PR #88 head 94e8c78, run 36202537423): (1) tools/implementation_coverage.py FAILS — it declares go: internal/scriptpolicy.TestARefusalPrecedesEveryWorkerSurface which curator main no longer has; (2) windows-latest checkout of curator main fails: Filename too long under .task-board/ (deep board paths). Fix both so the Implementations job is green against curator 0a628621 with the #88 content.

## Scope
(define task scope)

## Acceptance Criteria
implementations.yml pins curator 0a628621 with an accurate comment; Windows checkout succeeds (sparse checkout excluding .task-board or core.longpaths, stated); every declared Go consumption case exists in curator 0a628621 (verified by go test -list) and is observed; the Implementations job steps pass locally/against a candidate that includes #88 content (b202b5d); CHANGELOG entry under unreleased.
