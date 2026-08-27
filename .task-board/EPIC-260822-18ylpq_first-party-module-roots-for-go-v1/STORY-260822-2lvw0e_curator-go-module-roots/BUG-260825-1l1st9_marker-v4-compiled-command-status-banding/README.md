# BUG-260825-1l1st9: marker-v4-compiled-command-status-banding

## Description
Conformance escape found by the first real schema-8 install (skill-project-management at 3958813, global scope): install succeeds, writes marker v4, publishes shims — but curator global status reports every compiled command needs-install with detail: install marker schema 4 cannot describe a compiled command; reinstall to record marker schema 2. This is exactly the not-current-after-successful-install class the marker-v4 spec rework (candidate 6001dc3, core.md section 10: managers supporting schema 8 MUST read marker schemas 1-4 and write v4 for schema-8 installation mutations) exists to prevent — some status/currentness reader still bands compiled commands on marker v2/v3 and rejects v4; the remedy text even instructs recording schema 2, contradicting the landed spec. Find the banding check, extend it to marker v4 for schema-8 installs per the spec, add a regression test (schema-8 skill with build commands: install then status must report installed/current), check whether the conformance suite lacks a vector for this path and record that gap for the suite owner if so. Land via PR with green lanes; verify on this machine that curator global status flips to installed for project-management without reinstalling. Maintainer pre-authorization 2026-08-22: fully autonomous. Executor: claude only.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
curator global status reports project-management and its three compiled commands installed/current on marker v4; regression test pins it; merged green.
