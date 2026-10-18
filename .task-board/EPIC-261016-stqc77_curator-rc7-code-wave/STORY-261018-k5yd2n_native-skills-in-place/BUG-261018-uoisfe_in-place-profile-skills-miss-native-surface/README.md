# BUG-261018-uoisfe: in-place-profile-skills-miss-native-surface

## Description
In-place profile skills are written only to <native home>/.agent-context/skills/<name> (internal/envprofile/switch.go:65), which no harness discovers; the native discovery directories (internal/adapters/adapters.go:22-35) are written only by the legacy global scope, and no operation takes over foreign links there. Direct native sessions keep discovering a retired manager's skills, so the infra exit needs manual steps.

## Scope
internal/envprofile switch, status and repair paths, the registry lookups they need, tests and a short docs note. Opencode and managed homes unchanged. No spec text; a separate curator-spec clarification follows.

## Acceptance Criteria
1. For the current profile of each scope, use, sync, install --use, update and repair materialize the profile's skills as recorded manager-owned links at <native home>/skills/<name> for claude_code, codex_cli and pi. Muse, opencode and managed homes keep today's behaviour. 2. Unrecorded entries follow the closed takeover rule of environments §9.5 and §8.3.1: a foreign symlink stops without --takeover and is backed up with identical link text with it; other unrecorded entries give environment_surface_unmanaged_conflict; entries the profile does not name are never touched; a symlinked component refuses with environment_write_would_follow_link. 3. Adapter-ledger entries are left in place and reported. 4. Recorded .agent-context/skills paths converge away. 5. The twelve attack-surface rows of gw-NATSKILL-r2-brief.md are driven by named tests with narrowing mutants; the real HOME is never touched.
