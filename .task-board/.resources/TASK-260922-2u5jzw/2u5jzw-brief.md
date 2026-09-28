# TASK-260922-2u5jzw — F-L1b: implement the 0018 permission interface in curator-run

Control root: /Users/administrator/Developer/ReluxWorks/curator/curator-agent-launcher; work only in your
Story worktree `.temp/STORY-260922-39hxog/worktree`. Read `campaign-producer-rules.md` first.
Landing gate = hosted CI (the runtime runs it once at handoff).

## Source of truth
- The launcher SPEC `0.5.0-draft` already on your Story branch (F-L1a, TASK-260922-1zfqq0 rev2,
  checkpointed): §3 flag table, §4.1 fragment members + transport precondition, §4.3 precedence,
  §4.5 composition placement, §4.6 tracked refusal + headless detector {CI, GITHUB_ACTIONS} + lock,
  §4.7 file-family knob, §6 diagnostics (`permission_policy_unsupported`,
  `permission_mode_tracked_unsupported`, `permission_mode_unsupported` — already declared in
  `internal/diagnostics`, you give them behaviour), choice-4 effective-native-policy stderr line and key.
- curator-spec `protocol/environments.md` §10.1/§10.2 and fragment revision `launch-env-fragment-v2`
  (`permissions {mode, locked, source}`; `locked iff source==global`).
- skill-agents-management **v0.5.18**: `LaunchRequest.PermissionMode`, `permission-grammar-v1`,
  `ErrPermissionModeUnverifiedRelease`. Bump go.mod from v0.5.13 to v0.5.18 (`go get …@v0.5.18`,
  `go mod tidy`); if the bump breaks unrelated call sites, adapt them minimally and list them.

## Deliverable
Exactly the task's description and checklist: cli parsing (`--permissions native|yolo`, `--yolo`
alias in `=` and separate forms, conflicts refused, `-d/--danger` rejected), precedence with
provenance source, composition into `LaunchRequest.PermissionMode` (NEVER a provider flag string —
decision 0013 D5), execution refusals, the choice-4 line, CHANGELOG (unreleased).

## Tests (the substance of this leaf)
- Every choice-5 row family is driven through the REAL `curator-run` entry (the same function
  `main` calls) with a FAKE tool binary/script — never a real agent CLI, never real `ax`, no network.
  Report the count of executed rows per family in results.
- A module-level test that scans the launcher's non-test Go sources and fails on any provider bypass
  spelling (`--dangerously`, `skip-permissions`, `bypassPermissions`, `--full-auto`, `--yolo` passed
  to a provider argv, `--approval`, `--sandbox`).
- Narrowing mutants: one per refusal bound (tracked+yolo from each level, lock-engaged, transport not
  established, unknown value, conflicting flags, -d/--danger). Apply each in a DISPOSABLE copy
  (`git worktree add` a detached scratch under `.temp/` of your Story or `cp -R` to $TMPDIR), run the
  focused test, record killed/survived; never mutate your Story worktree. Table in results.
- `make check` (build, fmt-check, vet, test, race) with its real exit code. Long runs: split into
  bounded calls (≤10 min each), never background past your turn.

## Handoff
Attach `TASK-260922-2u5jzw_results.md` (row counts per family, mutant table, grep/scan row, make
check exit, go.mod diff), check each checklist item citing that resource, then
`task-board handoff TASK-260922-2u5jzw --role developer`. A `run_wrote_outside_worktree … policy
warn` block is a warning, not a refusal — verify the status moved to `to-review`.

## Boundaries
No SPEC edits (F-L1a is landed on the branch; a genuine SPEC defect → Stop-The-Line with evidence).
No edits to skill-agents-management. No LOGBOOK.md. Fake tools only.
