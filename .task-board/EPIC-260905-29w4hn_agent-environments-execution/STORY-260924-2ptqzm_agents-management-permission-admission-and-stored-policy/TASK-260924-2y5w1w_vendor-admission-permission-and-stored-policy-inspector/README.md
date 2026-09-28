# TASK-260924-2y5w1w: vendor-admission-permission-and-stored-policy-inspector

## Description
(a) vendorplugin.SpawnRequest gains PermissionMode, ToolRelease and NativeArgs, carried unchanged into the agentic.LaunchRequest that Vendor.Spawn/PrepareLaunchRequest/BuildPlanWithEnvironment use, so vendor admission stays the only path; (b) a best-effort stored-policy inspector per plugin (Claude, Codex; Pi reports unsupported) that reads the provider's known stored settings for the known selectors and returns the relaxations found plus the exact list of sources inspected and sources not inspected (unreadable, absent) — never claiming beyond what it read; (c) release v0.5.21.

## Scope
(define task scope)

## Acceptance Criteria
1) SpawnRequest permission members reach the plan through vendor admission (rows through BuildLaunchWithEnvironment); 2) inspector rows per known selector and source, unreadable source reported as not inspected, never as clean; 3) no provider spelling outside plugins (argvguard); 4) narrowing mutants killed; 5) go test/vet exit 0; 6) README + CHANGELOG new bullet (released entries untouched)
