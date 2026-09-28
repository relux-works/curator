# TASK-260922-1s0zja — F-M1a: LaunchRequest permission-mode member + provider mapping (skill-agents-management)

Control root: /Users/administrator/Developer/ReluxWorks/skill-agents-management (work only in your assigned
Story worktree `.temp/STORY-260922-h3epwn/worktree`; the repository is shared with the operator's own session —
never write outside the worktree). Board: shared curator board. Read `campaign-producer-rules.md` first
(the control-root list there predates this repository; the same rules apply).

## Why
curator-spec Decision 0018 (ADOPTED 2026-09-21, corrected 2026-09-22; read
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/decisions/0018-curator-run-permission-interface.md`
choices 1, 3, 6 and the Compatibility section) makes agents-management the OWNER of the provider flag
spelling for the `yolo` permission mode: a `LaunchRequest` permission-mode member valid for
`LaunchModeInteractive`, mapped per pinned tool release, with positive and negative interactive goldens. The
launcher (curator run) only resolves and passes the mode (Decision 0013 D5); `argvguard` forbids spelling argv
grammar in two places — that is why the mapping lives here and nowhere else.

## What exists (anchors)
- `pkg/agentic/system.go`: `LaunchModeInteractive` (comment lines ~103–128 state the CLOSED interactive grammar:
  "no permission-bypass or unrestricted-mode flag") and `type LaunchRequest struct` (~line 383).
- Plugins: `pkg/agentic/systems/claude/{claude.go,args.go,argvguard_test.go,interactive_test.go}`,
  `pkg/agentic/systems/codex/...`, `pkg/agentic/systems/pi/interactive_test.go`; cross-cutting sweep
  `internal/regress/interactive_test.go` (`execModeMarkers` lists `--dangerously-skip-permissions` and
  `--dangerously-bypass-approvals-and-sandbox` as markers that must NOT appear on interactive argv today).
- `internal/argvguard` (shared scanner; each plugin's argvguard_test.go carries the mutant battery).
- README.md §(interactive mode) mentions the deletes of `--dangerously-skip-permissions`.

## Deliverable
1. Member: add a permission-mode member to `LaunchRequest` (name it, e.g. `PermissionMode` with values
   `native` (zero value: pass NOTHING — the provider's stored settings decide) and `yolo`). Validation in
   BuildPlan: any non-zero value outside `LaunchModeInteractive` is refused with a named sentinel
   (`ErrPermissionModeNotInteractive`-style); an unknown value is refused (`ErrPermissionModeUnknown`-style).
   Keep every existing golden byte-identical for `native`/zero value.
2. Mapping (per pinned tool release, spelled ONCE per plugin): claude_code `yolo` → `--dangerously-skip-permissions`;
   codex_cli `yolo` → `--dangerously-bypass-approvals-and-sandbox`; pi → its documented interactive bypass
   flag if the pinned pi release has one, otherwise an explicit refusal (`ErrPermissionModeUnsupported`) —
   read the pi plugin/docs and decide with evidence, do not guess. The flag is emitted exactly once, before
   any prompt text, in the interactive argv only.
3. Update the interactive grammar contract text in `system.go` (the closed set now admits exactly this one
   optional member under 0018; cite 0018 + 0013 D5) and the sweep in `internal/regress/interactive_test.go`:
   the marker sweep keeps rejecting the bypass flags for `native` and asserts exactly-one for `yolo`.
4. Goldens (positive + negative) per plugin, pinned to the tool release the repository already pins/tests
   against (state the release in results.md): positive native (no flag) and yolo (exact flag) argv for
   claude_code and codex_cli; negative: unknown value, yolo with a non-interactive LaunchMode, yolo combined
   with a raw bypass flag already present in the caller's native args (duplicate spelling refused, not
   de-duplicated). One narrowing mutant per refusal bound, executed and killed; table in results.md.
5. argvguard: the plugin guards must prove each bypass flag is spelled at exactly one site in the module.
6. README + CHANGELOG (unreleased) describing the member; do NOT tag a release in this leaf (F-M1b cuts it).

## Boundaries
No launcher, curator, ax, task-board changes. No tool-release capability table yet (F-M1b,
TASK-260922-1wvwc3). No real provider launches in tests (fake binaries/argv capture only). Local checks:
`go build ./... && go test ./... -count=1` is the board's validation command — run narrow package tests with
real exit codes yourself; the runtime runs the full command once at handoff. Hand off with
`task-board handoff TASK-260922-1s0zja --role developer` after attaching `TASK-260922-1s0zja_results.md`.
