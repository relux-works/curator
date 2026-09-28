# Review note — TASK-260924-2y5w1w F-M1d, CR revision 1 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review against `2y5w1w-brief.md`, the AC and the results. Read-only; disposable clone of skill-agents-management.
1. `vendorplugin.SpawnRequest` PermissionMode/ToolRelease/NativeArgs reach the plan THROUGH vendor admission
   (Vendor.Spawn → PrepareLaunchRequest → BuildPlanWithEnvironment); no bypass path; the v0.5.18 grammar/drift refusal and
   the v0.5.20 conflict refusal apply on the vendor path (rows through BuildLaunchWithEnvironment).
2. Stored-policy inspector (curator-spec decisions/0018 choice 4 "best-effort detection over known selectors; never claims
   beyond what it inspected"): sources and selectors per plugin are documented; unreadable/malformed/absent sources are
   reported as NOT inspected, never as clean; inputs only from the supplied environment/home/workdir (no ambient reads —
   grep for os.UserHomeDir/os.Getenv in the new code). Pi → unsupported.
   Judge the selector choice: is `permissions.allow` / `additionalDirectories` a "relaxation beneath native" or noise?
   Flag over-reporting only if it would mislead the operator's stderr line.
3. Provider spellings stay inside plugins (argvguard/spelling guard green). CHANGELOG: one new bullet, released entries
   untouched, heading `## Unreleased`.
4. Mutant table; re-apply one yourself (report unreadable as clean) → killed. `go test ./...`, `go vet ./...` exit 0.
accept_cr or changes requested with file:line. No LOGBOOK.md.
