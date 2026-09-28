# TASK-260922-2u5jzw — F-L1b resume 2 (THE ONLY CURRENT INSTRUCTION; `2u5jzw-brief.md` still defines the deliverable)

Both of your Stop-The-Line findings are RESOLVED by skill-agents-management **v0.5.21** (signed tag, commit 850ac39):
- `vendorplugin.SpawnRequest` now carries `PermissionMode`, `ToolRelease`, `NativeArgs` through `Vendor.Spawn` into the plan
  (fidelity-checked; `BuildLaunchWithEnvironment` stays the only path). v0.5.20's conflict refusal
  (`ErrNativePolicyConflict`, `*NativePolicyConflictError{Selector, Placement}`) and v0.5.18's drift refusal apply on it.
  Grammar token: `permission-grammar-v2` for Claude/Codex, v1 for Pi.
- `agentic.StoredPolicyInspection` (+ capability dispatch from an admitted plan): best-effort Claude/Codex stored-policy
  inspection with relaxations, sources inspected and sources NOT inspected; Pi → unsupported. Read its README section.

1. `task-board m 'set_status(TASK-260922-2u5jzw, status=development)'`.
2. Pin v0.5.21 (`go get …@v0.5.21`, `go mod tidy`).
3. Implement exactly `2u5jzw-brief.md` on the vendor path: flag/alias parsing, precedence with provenance, the resolved
   mode into `SpawnRequest.PermissionMode` (never a provider flag), tracked refusal, transport refusal, lock, headless
   detector; map the module's typed errors to SPEC §6 exits via errors.Is/As; choice 4: call the inspector under a
   `native` request and print the SPEC's effective-native-policy stderr line (summarise long allow-lists by count) and, when
   tracked, record the launcher-SPEC key — never claim a source it reports as not inspected.
4. The launcher SPEC 0.5.0-draft names `permission-grammar-v1` (§4.4) and does not list the inspected sources; do NOT edit
   the SPEC — record both as SPEC follow-ups in results.
5. Rows, mutants, `make check`, CHANGELOG, results, checklist, `task-board handoff TASK-260922-2u5jzw --role developer`
   exactly as the brief says. A `run_wrote_outside_worktree … policy warn` block is a warning — verify status `to-review`.
