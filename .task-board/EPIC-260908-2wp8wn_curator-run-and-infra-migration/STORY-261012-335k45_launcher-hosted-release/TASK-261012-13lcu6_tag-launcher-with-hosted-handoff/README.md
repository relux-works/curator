# launcher-v030-muse-admission-and-failclosed-hosted

## Description
Owner go 2026-10-13 (native curator run muse on the operator host): bump skill-agents-management v0.5.53 to v0.5.62 (Muse 1.4.6 admitted with no release table, muse-wonder alias, fragment v3), adapt API drift; owner decision 2a (2026-10-12): --hosted keeps failing closed with session_host_missing and docs say it is unavailable until the swarma path; CHANGELOG 0.3.0 pending tag. The tag is cut by the orchestrator after landing.

## Scope
(define task scope)

## Acceptance Criteria
Every item of gw-LRM1-brief.md applied; agents-management v0.5.62 admits Muse 1.4.6 and muse-wonder through the launcher; --hosted fail-closed with existing refusal and documented unavailable; hosted gate green; public hygiene
