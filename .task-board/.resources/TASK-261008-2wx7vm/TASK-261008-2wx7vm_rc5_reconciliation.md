# rc.5 reconciliation: v0.15.0-rc.4..origin/main

Range read in full after `git fetch origin main` (2026-10-08). 10 commits: 2 product, 8 board-record (no product change).

## Commits reconciled

- a8bf49cf 2026-10-08 STORY-260928-oflbe1: security-posture-hardened-flip (CR-TASK-260927-25hk87-3 rev 3) — PRODUCT
- 1de7dee6 2026-10-08 STORY-260928-2n2ii7: codex-seed-revision-b-flip (CR-TASK-260927-1e5qqm-3 rev 3) — PRODUCT
- 93e88447 Record STORY-260928-2n2ii7 board state — no product change
- 69956b36 Record STORY-261008-yt6lo4 board state — no product change
- 4b0ba5a8 Record STORY-261008-yt6lo4 board state — no product change
- 5b5b4d7b Record STORY-260928-oflbe1 board state — no product change
- 57be92aa Record STORY-261008-yt6lo4 board state — no product change
- 35cac659 Record STORY-260928-oflbe1 board state — no product change
- 75ab9a71 Record STORY-261002-2prz8d board state — no product change
- 5253fff5 Record STORY-261007-2tthu8 board state — no product change

## Entry -> commit mapping

- Security / posture revision B flip -> a8bf49cf (SecurityPostureRevision A->B, docs/environment-config.md Security posture section, ci-gates.md posture accounting 17/17, migration-hint reword, goldens). Operator-visible effects verified in code: defaultSecurityPosture schema-2 -> hardened (SchemaVersion=1 frozen permissive), SecurityPostureDiagnostics refusals (source_allowlist_empty, mcp_package_allowlist_empty, passable_env_names_unbounded_refused), transitive_system_modules default error-under-hardened; escape hatch explicit security_posture: permissive + once-per-operation warning. Test names: TestSecurityPostureRevisionBDefaultRefusesEmptySources, TestSecurityPostureExplicitPermissiveInstallsWithoutSourceAllowlist, TestSecurityPostureRevisionBDefaultEnvStatus, TestSecurityPostureRevisionBDefaultsAndPermissiveCompatibility, TestSecurityPostureRevisionBExplicitKnobsStillWin.
- Changed / Codex seed revision B flip -> 1de7dee6 (CodexSeedRevision A->B, 36 conformance-gap rows removed = 9 cases x 4 digests, ci-gates.md seed accounting 7/7 + 8/8). Operator-visible effects verified in code/tests: provisioning strips mcp_servers (inline + subtable forms) and keeps the rest; warning mcp_native_servers_not_inherited replaces mcp_native_servers_ungoverned (accept-the-loss hint gone); status rows revision B (shipped) / (not-inherited); revision-A homes preserved through repair with mcp_seed_unstripped re-provision warning (managed.go:904, status.go:778,790).
- Added: No additions in this release -> full range contains no feature commit (verified by stat review of both product commits: const flips + tests + docs + goldens + gap rows only).
- Fixed: No fixes in this release -> same verification.
- Known issues: Windows broker real-Git flake TASK-260930-fp8vx7 (carried from rc.4, still documented risk) and Go-qualification board-close gap spm#537 tooling-only (carried from rc.4); brief-directed, no product impact. The rc.4 B-flips-deferred item is dropped because both flips shipped above.

## Scope check

- `git status --short` shows only ` M CHANGELOG.md` (64 insertions). LOGBOOK.md untouched. No version constants touched.
- Docs-only change: no build or test command applies; none run.
