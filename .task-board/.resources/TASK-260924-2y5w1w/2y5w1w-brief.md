# TASK-260924-2y5w1w — F-M1d: permission request through vendor admission + stored-policy inspector (THE ONLY CURRENT INSTRUCTION)

Control root: /Users/administrator/Developer/ReluxWorks/skill-agents-management (shared with the operator's session —
never write outside your assigned Story worktree). Read `campaign-producer-rules.md`. Board validation runs at handoff.

## Why
F-L1b (launcher, TASK-260922-2u5jzw) stopped correctly on v0.5.20 (see its results): the launcher MUST use the vendor
admission path (`vendorplugin.BuildLaunchWithEnvironment(SpawnRequest)` → private `agentic.LaunchRequest` from
`Vendor.Spawn` → `PrepareLaunchRequest` → `BuildPlanWithEnvironment`, `pkg/vendorplugin/spawn.go:67-288`), but
`SpawnRequest` has no `PermissionMode`, `ToolRelease` or `NativeArgs`; and Decision 0018 choice 4 needs a best-effort
stored-policy inspection that only this module may spell (Decision 0013 D5).

## Deliverable
(a) `vendorplugin.SpawnRequest` gains `PermissionMode`, `ToolRelease`, `NativeArgs`, carried unchanged into the
    `agentic.LaunchRequest` so the existing release-bound grammar, mapping, drift refusal and v0.5.20 conflict refusal
    all apply on the vendor path. Rows through `BuildLaunchWithEnvironment` with a fake vendor/tool: native, yolo,
    conflict refusal, unverified-release refusal — admission never bypassed.
(b) Stored-policy inspector (curator-spec `decisions/0018-curator-run-permission-interface.md` choice 4: "best-effort
    detection over known selectors; the launcher never claims beyond what it inspected"). Per plugin: Claude — its
    stored settings files and the known permission selectors (e.g. `permissions.defaultMode`, allow/bypass settings);
    Codex — its config file keys `approval_policy`, `sandbox_mode`, `sandbox_permissions` (+ profiles if the config
    selects one); Pi → unsupported. Return a typed result: relaxations found (selector, value, source path), sources
    inspected, sources NOT inspected (absent / unreadable / unparseable — never reported as clean). Home/config roots
    are injected (the launcher passes the launch environment); no ambient reads. Document which sources each plugin
    inspects in README "Permission-mode".
(c) CHANGELOG: one NEW bullet (released entries untouched; heading stays `## Unreleased`), release v0.5.21 cut by the
    orchestrator.

## Tests
Rows per selector × source; unreadable (chmod 000 on POSIX) → not-inspected; malformed → not-inspected; absent →
not-inspected-absent; argvguard/spelling guard still green (no provider spelling outside plugins). Narrowing mutants in a
DISPOSABLE copy: report unreadable as clean; drop one selector; bypass admission for permission members → each killed.
`go test ./...`, `go vet ./...` bounded. Fake files only; never read the real user's provider settings.

## Handoff
Attach `TASK-260924-2y5w1w_results.md`, check each DoD item citing it, `task-board handoff TASK-260924-2y5w1w --role
developer`. A `run_wrote_outside_worktree … policy warn` block is a warning — verify status `to-review`. No LOGBOOK.md.
