# Round 2 narrowing controls

Candidate `b32823ac920d8b0485c4a7a7111b48fff8e3ca2b`. Fresh standalone native suite exit 0; all twelve current children exited 1 with named intended behavioral failures. The CLI child retains actual run() routing via the existing helper lane.


| Mutant | What it narrows the gate or behavior to | Named failing test | Survival bound |
|---|---|---|---|
| pi-native-only | pi alone falls back to staging, omitting its native write while the other adapters stay native | TestNativeSkillsUseAndSync | Killed; no survivor |
| codex-link-drift | wrong-target links are ignored only for Codex native skills | TestNativeSkillsRepair | Killed; no survivor |
| dangling-foreign-consent | dangling relative links below the retired-manager spelling bypass consent | TestNativeSkillsTakeover/foreign | Killed; no survivor |
| plain-file-consent | unrecorded regular files alone bypass consent; other kinds retain the gate | TestNativeSkillsTakeover/file | Killed; no survivor |
| skills-parent-traversal | managedPath traverses a linked component named skills; other components retain refusal | TestNativeSkillsLinkedParent | Killed; no survivor |
| unnamed-takeover-directory | takeover removes one unrequested directory named unnamed | TestNativeSkillsConvergenceAndSwitch | Killed; no survivor |
| one-ledger-name | only skills/myskill loses ledger exclusion | TestNativeSkillsLedger | Killed; no survivor |
| one-old-staging-name | cleanup retains one old recorded staging name | TestNativeSkillsConvergenceAndSwitch | Killed; no survivor |
| one-retired-name | cleanup retains one obsolete native skill name | TestNativeSkillsConvergenceAndSwitch | Killed; no survivor |
| backup-directory-collision | existing directories alone bypass the next-generation collision check | TestNativeSkillsBackupCollision | Killed; no survivor |
| directory-marker-is-absence | a directory at the marker path alone is treated as absent | TestNativeSkillsUnreadableMarker | Killed; no survivor |
| claude-takeover-without-repair | bare resolve accepts takeover only for claude_code | TestNativeSkillsTakeoverFlagBoundary/env-resolve-claude_code | Killed; no survivor |

All mutation plants keep relevant source tokens and execute behavioral Go
suites. No new runtime gate attests behavior by scanning source text. Mutation
anchor checks only validate control construction; a token hit cannot produce a
passing kill verdict.


Historical Windows exception: dangling-foreign-consent at eb218081 had no intended named failure because its slash-only matcher did not apply; it is a survivor/invalid control there, bounding prior evidence to POSIX spelling. The fourth candidate uses filepath.ToSlash only in the mutant matcher; backup/refusal assertions compare unnormalized original Readlink text. Fresh hosted Windows job is green. Its downloaded test stream records all 12 controls passing, each with its named expected-red child failure; windows-native-current.json preserves the event names and output. The full workflow is green: 20 applicable jobs succeeded, with 2 documented conditional skips.

```text
=== RUN   TestNativeSkillsNarrowingMutants
=== RUN   TestNativeSkillsNarrowingMutants/pi-native-only
    native_skills_mutants_test.go:79: expected-red exit 1; killed by TestNativeSkillsUseAndSync:
        === RUN   TestNativeSkillsUseAndSync
            native_skills_test.go:61: native skill link missing or outside store: "" readlink <temporary-path> no such file or directory
        --- FAIL: TestNativeSkillsUseAndSync (1.03s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/envprofile	1.359s
        FAIL
--- PASS: TestNativeSkillsNarrowingMutants/pi-native-only (3.07s)
=== RUN   TestNativeSkillsNarrowingMutants/codex-link-drift
    native_skills_mutants_test.go:79: expected-red exit 1; killed by TestNativeSkillsRepair:
        === RUN   TestNativeSkillsRepair
            native_skills_test.go:283: drift status = &{Homes:[{markerState:absent Profile:default Environment:claude_code Mode:managed-home Form: Home:<temporary-path> Provisioned:false Current:false LockHash:sha256:06fb1543fb69055047d8f34bbac2901f06b4da1374cdbbbc47a26d22126362a4 MarkerHash: Surfaces:[] Passthrough:[] Seeds:[] SeedLinks:[] SeededProjects:[] CodexSeedRecord:<nil> NativeMCPServers:[] NativeMCPServersDisposition:none BackupsKnown:true Backups:0 BackupsOldest:- BackupsNewest:- Findings:[home unprovisioned] Warnings:[]} {markerState:absent Profile:default Environment:codex_cli Mode:managed-home Form: Home:<temporary-path> Provisioned:false Current:false LockHash:sha256:06fb1543fb69055047d8f34bbac2901f06b4da1374cdbbbc47a26d22126362a4 MarkerHash: Surfaces:[] Passthrough:[] Seeds:[] SeedLinks:[] SeededProjects:[] CodexSeedRecord:<nil> NativeMCPServers:[] NativeMCPServersDisposition: BackupsKnown:true Backups:0 BackupsOldest:- BackupsNewest:- Findings:[home unprovisioned] Warnings:[]} {markerState:absent Profile:default Environment:muse Mode:managed-home Form: Home:<temporary-path> Provisioned:false Current:false LockHash:sha256:06fb1543fb69055047d8f34bbac2901f06b4da1374cdbbbc47a26d22126362a4 MarkerHash: Surfaces:[] Passthrough:[] Seeds:[] SeedLinks:[] SeededProjects:[] CodexSeedRecord:<nil> NativeMCPServers:[] NativeMCPServersDisposition:none BackupsKnown:true Backups:0 BackupsOldest:- BackupsNewest:- Findings:[home unprovisioned] Warnings:[]} {markerState:absent Profile:default Environment:opencode Mode:managed-home Form: Home:<temporary-path> Provisioned:false Current:false LockHash:sha256:06fb1543fb69055047d8f34bbac2901f06b4da1374cdbbbc47a26d22126362a4 MarkerHash: Surfaces:[] Passthrough:[] Seeds:[] SeedLinks:[] SeededProjects:[] CodexSeedRecord:<nil> NativeMCPServers:[] NativeMCPServersDisposition:none BackupsKnown:true Backups:0 BackupsOldest:- BackupsNewest:- Findings:[home unprovisioned] Warnings:[]} {markerState:absent Profile:default Environment:pi Mode:managed-home Form: Home:<temporary-path> Provisioned:false Current:false LockHash:sha256:06fb1543fb69055047d8f34bbac2901f06b4da1374cdbbbc47a26d22126362a4 MarkerHash: Surfaces:[] Passthrough:[] Seeds:[] SeedLinks:[] SeededProjects:[] CodexSeedRecord:<nil> NativeMCPServers:[] NativeMCPServersDisposition:none BackupsKnown:true Backups:0 BackupsOldest:- BackupsNewest:- Findings:[home unprovisioned] Warnings:[]} {markerState:absent Profile:native-skills Environment:claude_code Mode:managed-home Form: Home:<temporary-path> Provisioned:false Current:false LockHash:sha256:e5b14a8c09f1ba5760894ad228e7e65c0aacf18de03f60c362802f1d72288c63 MarkerHash: Surfaces:[] Passthrough:[] Seeds:[] SeedLinks:[] SeededProjects:[] CodexSeedRecord:<nil> NativeMCPServers:[] NativeMCPServersDisposition:none BackupsKnown:true Backups:0 BackupsOldest:- BackupsNewest:- Findings:[home unprovisioned] Warnings:[]} {markerState:present Profile:native-skills Environment:codex_cli Mode:managed-home Form:monolithic Home:<temporary-path> Provisioned:true Current:true LockHash:sha256:e5b14a8c09f1ba5760894ad228e7e65c0aacf18de03f60c362802f1d72288c63 MarkerHash:e5b14a8c09f1ba5760894ad228e7e65c0aacf18de03f60c362802f1d72288c63 Surfaces:[{Key:mcp Paths:[curator-mcp.config.toml] Form: State:current Detail:current} {Key:root-context Paths:[AGENTS.md] Form:monolithic State:current Detail:current} {Key:skills Paths:[skills/myskill] Form: State:current Detail:current} {Key:system-prompt Paths:[.agent-context/system-prompt.md] Form: State:current Detail:current}] Passthrough:[auth.json (keyring-preferred)] Seeds:[] SeedLinks:[] SeededProjects:[] CodexSeedRecord:0x1400036c660 NativeMCPServers:[] NativeMCPServersDisposition:none BackupsKnown:true Backups:0 BackupsOldest:- BackupsNewest:- Findings:[] Warnings:[]} {markerState:absent Profile:native-skills Environment:muse Mode:managed-home Form: Home:<temporary-path> Provisioned:false Current:false LockHash:sha256:e5b14a8c09f1ba5760894ad228e7e65c0aacf18de03f60c362802f1d72288c63 MarkerHash: Surfaces:[] Passthrough:[] Seeds:[] SeedLinks:[] SeededProjects:[] CodexSeedRecord:<nil> NativeMCPServers:[] NativeMCPServersDisposition:none BackupsKnown:true Backups:0 BackupsOldest:- BackupsNewest:- Findings:[home unprovisioned] Warnings:[]} {markerState:absent Profile:native-skills Environment:opencode Mode:managed-home Form: Home:<temporary-path> Provisioned:false Current:false LockHash:sha256:e5b14a8c09f1ba5760894ad228e7e65c0aacf18de03f60c362802f1d72288c63 MarkerHash: Surfaces:[] Passthrough:[] Seeds:[] SeedLinks:[] SeededProjects:[] CodexSeedRecord:<nil> NativeMCPServers:[] NativeMCPServersDisposition:none BackupsKnown:true Backups:0 BackupsOldest:- BackupsNewest:- Findings:[home unprovisioned] Warnings:[]} {markerState:absent Profile:native-skills Environment:pi Mode:managed-home Form: Home:<temporary-path> Provisioned:false Current:false LockHash:sha256:e5b14a8c09f1ba5760894ad228e7e65c0aacf18de03f60c362802f1d72288c63 MarkerHash: Surfaces:[] Passthrough:[] Seeds:[] SeedLinks:[] SeededProjects:[] CodexSeedRecord:<nil> NativeMCPServers:[] NativeMCPServersDisposition:none BackupsKnown:true Backups:0 BackupsOldest:- BackupsNewest:- Findings:[home unprovisioned] Warnings:[]}] Scopes:[{NativeSkills:0x14000115ec0 Scope:machine Profile:native-skills Environment:claude_code Native:<temporary-path> Managed:<temporary-path> Provisioned:false ProvisionedKnown:true Diagnostic:0x1400036d470} {NativeSkills:0x140004521e0 Scope:machine Profile:native-skills Environment:codex_cli Native:<temporary-path> Managed:<temporary-path> Provisioned:true ProvisionedKnown:true Diagnostic:<nil>} {NativeSkills:<nil> Scope:machine Profile:native-skills Environment:muse Native: Managed:<temporary-path> Provisioned:false ProvisionedKnown:true Diagnostic:<nil>} {NativeSkills:<nil> Scope:machine Profile:native-skills Environment:opencode Native:<temporary-path> Managed:<temporary-path> Provisioned:false ProvisionedKnown:true Diagnostic:<nil>} {NativeSkills:0x14000452480 Scope:machine Profile:native-skills Environment:pi Native:<temporary-path> Managed:<temporary-path> Provisioned:false ProvisionedKnown:true Diagnostic:0x140004de1b0}] Adapters:[{ID:muse Recorded:1.4.7 Detected:unknown Supported:[monolithic]} {ID:claude_code Recorded:2.1.261 Detected:unknown Supported:[monolithic referenced]} {ID:codex_cli Recorded:0.153.2 Detected:unknown Supported:[monolithic]} {ID:opencode Recorded:unrecorded Detected:unknown Supported:[monolithic referenced]} {ID:pi Recorded:0.84.2 Detected:unknown Supported:[monolithic]}] Targets:[{ID:xcode-coding-assistant Adapter:claude_code Participating:false Consented:false Detail:auto: probe path absent, nothing materialized Ungoverned:the embedded host's MCP configuration and commands/ are present, unaudited, and outside this capability} {ID:xcode-coding-assistant Adapter:codex_cli Participating:false Consented:false Detail:auto: probe path absent, nothing materialized Ungoverned:the embedded host's MCP configuration and commands/ are present, unaudited, and outside this capability}] Profiles:[{Profile:default LockHash:sha256:06fb1543fb69055047d8f34bbac2901f06b4da1374cdbbbc47a26d22126362a4 Members:[{Kind:context Name:default Weight:0}] Precedence:{Winner:higher-weight Placement:winner-last} TransitiveSystemModules:drop DroppedSystemModules:[]} {Profile:native-skills LockHash:sha256:e5b14a8c09f1ba5760894ad228e7e65c0aacf18de03f60c362802f1d72288c63 Members:[{Kind:context Name:native-skills Weight:100} {Kind:mcp Name:figma-devmode Weight:0} {Kind:skill Name:myskill Weight:0}] Precedence:{Winner:higher-weight Placement:winner-last} TransitiveSystemModules:drop DroppedSystemModules:[]}] Providers:[] UnregisteredEnvironments:[] Orphans:[] Diagnostics:[] Notes:[Muse: monolithic root context delivered at config/muse/AGENTS.md (verified on 1.4.7, 1.4.7-R6101.1); native root suppresses foreign personal rules fallback, but foreign personal skills persist; managed exec requires --no-foreign-personal-context when the target is materialized; interactive requires -c context.foreign_personal_skills=false (JSON boolean) when the target is materialized; serve cannot exclude foreign personal context (foreign-personal-context isolation gap), so required serve exclusion refuses; workspace AGENTS.md prevailed in observed conflicts (recorded residual, not a general precedence guarantee); 1.4.8 on macOS 27: file backend login, subsequent file backend with zero Keychain access, and shared file link verified; 1.4.8 exclusion qualification unknown until the exact build replays P1–P4 on both forms, root present/absent, foreign rules/skills separately, each exclusion spelling, Curator link and regular-file snapshot, and workspace conflict; record loader observations separately from model-visible behavior; model-visible foreign removal, skills discovery, system-prompt discovery, refresh semantics, and state/cache internal layouts remain unverified; Linux and Windows native Muse login and refresh remain unqualified opencode skills come from the machine-current profile, split-brain by construction (§7.1)] NonCurrent:true ShellHookTrust:[] ShellHookTrustWarnings:[] S4Profile:s4-warn PassableEnvNames:[] Warnings:[mcp_package_allowlist_empty: the MCP package allowlist is empty; every declaration package in the closure is admitted] CodexSeedRule:{Revision:B Provenance:shipped} MCPDeclarations:[{Scope:machine Profile:native-skills Rows:[mcp-declaration figma-devmode 1.2.0 stdio command=npx args=["-y","figma-developer-mcp","--stdio"] env_names=["FIGMA_API_KEY"]]}] RequireCurrentProfile:<nil> RequireCurrentLocked:false RegistryPosture:[] RequireSourceSigners:false SourceSignerPosture:[{Profile:native-skills Kind:mcp Name:figma-devmode Source:github.com/example/mcp-figma State:unconfigured Signer:unknown Current:true} {Profile:native-skills Kind:skill Name:myskill Source:github.com/example/skill-my State:unconfigured Signer:unknown Current:true}] UpdateConfirmationRevision:B-flip UpdateConfirmationBehavior:a triggered update delta refuses profile_update_confirmation_required unless --confirm-system-delta is given SecurityPostureRows:[]} <nil>
        --- FAIL: TestNativeSkillsRepair (0.76s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/envprofile	1.082s
        FAIL
--- PASS: TestNativeSkillsNarrowingMutants/codex-link-drift (2.67s)
=== RUN   TestNativeSkillsNarrowingMutants/dangling-foreign-consent
    native_skills_mutants_test.go:79: expected-red exit 1; killed by TestNativeSkillsTakeover/foreign:
        === RUN   TestNativeSkillsTakeover
        === RUN   TestNativeSkillsTakeover/foreign
            native_skills_test.go:118: switch error = nil, want diagnostic environment_foreign_manager_detected
        --- FAIL: TestNativeSkillsTakeover (0.50s)
            --- FAIL: TestNativeSkillsTakeover/foreign (0.50s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/envprofile	0.833s
        FAIL
--- PASS: TestNativeSkillsNarrowingMutants/dangling-foreign-consent (2.49s)
=== RUN   TestNativeSkillsNarrowingMutants/plain-file-consent
    native_skills_mutants_test.go:79: expected-red exit 1; killed by TestNativeSkillsTakeover/file:
        === RUN   TestNativeSkillsTakeover
        === RUN   TestNativeSkillsTakeover/file
            native_skills_test.go:118: switch error = nil, want diagnostic environment_surface_unmanaged_conflict
        --- FAIL: TestNativeSkillsTakeover (0.58s)
            --- FAIL: TestNativeSkillsTakeover/file (0.58s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/envprofile	0.914s
        FAIL
--- PASS: TestNativeSkillsNarrowingMutants/plain-file-consent (2.62s)
=== RUN   TestNativeSkillsNarrowingMutants/skills-parent-traversal
    native_skills_mutants_test.go:79: expected-red exit 1; killed by TestNativeSkillsLinkedParent:
        === RUN   TestNativeSkillsLinkedParent
        === RUN   TestNativeSkillsLinkedParent/ordinary
            native_skills_test.go:191: switch error = nil, want diagnostic environment_write_would_follow_link
        === RUN   TestNativeSkillsLinkedParent/takeover
            native_skills_test.go:191: switch error = nil, want diagnostic environment_write_would_follow_link
        --- FAIL: TestNativeSkillsLinkedParent (1.08s)
            --- FAIL: TestNativeSkillsLinkedParent/ordinary (0.51s)
            --- FAIL: TestNativeSkillsLinkedParent/takeover (0.58s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/envprofile	1.421s
        FAIL
--- PASS: TestNativeSkillsNarrowingMutants/skills-parent-traversal (3.00s)
=== RUN   TestNativeSkillsNarrowingMutants/unnamed-takeover-directory
    native_skills_mutants_test.go:79: expected-red exit 1; killed by TestNativeSkillsConvergenceAndSwitch:
        === RUN   TestNativeSkillsConvergenceAndSwitch
            native_skills_test.go:364: unnamed skill changed
        --- FAIL: TestNativeSkillsConvergenceAndSwitch (0.63s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/envprofile	0.966s
        FAIL
--- PASS: TestNativeSkillsNarrowingMutants/unnamed-takeover-directory (2.56s)
=== RUN   TestNativeSkillsNarrowingMutants/one-ledger-name
    native_skills_mutants_test.go:79: expected-red exit 1; killed by TestNativeSkillsLedger:
        === RUN   TestNativeSkillsLedger
            native_skills_test.go:219: ledger entry not reported: [{Adapter:claude_code Home:<temporary-path> OK:true Detail: Notice: Warnings:[]}]
        --- FAIL: TestNativeSkillsLedger (2.50s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/envprofile	2.857s
        FAIL
--- PASS: TestNativeSkillsNarrowingMutants/one-ledger-name (7.07s)
=== RUN   TestNativeSkillsNarrowingMutants/one-old-staging-name
    native_skills_mutants_test.go:79: expected-red exit 1; killed by TestNativeSkillsConvergenceAndSwitch:
        === RUN   TestNativeSkillsConvergenceAndSwitch
            native_skills_test.go:355: old staging entry retained
        --- FAIL: TestNativeSkillsConvergenceAndSwitch (3.70s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/envprofile	4.119s
        FAIL
--- PASS: TestNativeSkillsNarrowingMutants/one-old-staging-name (9.13s)
=== RUN   TestNativeSkillsNarrowingMutants/one-retired-name
    native_skills_mutants_test.go:79: expected-red exit 1; killed by TestNativeSkillsConvergenceAndSwitch:
        === RUN   TestNativeSkillsConvergenceAndSwitch
            native_skills_test.go:361: retired recorded skill retained
        --- FAIL: TestNativeSkillsConvergenceAndSwitch (5.56s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/envprofile	5.932s
        FAIL
--- PASS: TestNativeSkillsNarrowingMutants/one-retired-name (11.89s)
=== RUN   TestNativeSkillsNarrowingMutants/backup-directory-collision
    native_skills_mutants_test.go:79: expected-red exit 1; killed by TestNativeSkillsBackupCollision:
        === RUN   TestNativeSkillsBackupCollision
            native_skills_test.go:437: switch error = nil, want diagnostic environment_backup_exists
        --- FAIL: TestNativeSkillsBackupCollision (3.59s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/envprofile	3.964s
        FAIL
--- PASS: TestNativeSkillsNarrowingMutants/backup-directory-collision (9.96s)
=== RUN   TestNativeSkillsNarrowingMutants/directory-marker-is-absence
    native_skills_mutants_test.go:79: expected-red exit 1; killed by TestNativeSkillsUnreadableMarker:
        === RUN   TestNativeSkillsUnreadableMarker
            native_skills_test.go:259: switch error = nil, want diagnostic environment_marker_unreadable
        --- FAIL: TestNativeSkillsUnreadableMarker (3.62s)
        FAIL
        FAIL	github.com/relux-works/curator/internal/envprofile	3.982s
        FAIL
--- PASS: TestNativeSkillsNarrowingMutants/directory-marker-is-absence (8.63s)
=== RUN   TestNativeSkillsNarrowingMutants/claude-takeover-without-repair
    native_skills_mutants_test.go:79: expected-red exit 1; killed by TestNativeSkillsTakeoverFlagBoundary/env-resolve-claude_code:
        === RUN   TestNativeSkillsTakeoverFlagBoundary
        === RUN   TestNativeSkillsTakeoverFlagBoundary/env-resolve-claude_code
            native_skills_test.go:123: takeover admitted or unrelated failure = 1: warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults
                curator: profile_unknown: no profile is current
        --- FAIL: TestNativeSkillsTakeoverFlagBoundary (0.07s)
            --- FAIL: TestNativeSkillsTakeoverFlagBoundary/env-resolve-claude_code (0.07s)
        FAIL
        FAIL	github.com/relux-works/curator/cmd/curator	0.483s
        FAIL
--- PASS: TestNativeSkillsNarrowingMutants/claude-takeover-without-repair (5.41s)
--- PASS: TestNativeSkillsNarrowingMutants (68.52s)
```
