# TASK-260923-3e4i3n — F-M1c: refuse Decision 0018 known conflicting native selectors (skill-agents-management)

Control root: /Users/administrator/Developer/ReluxWorks/skill-agents-management (shared with the operator's
own session — never write outside your assigned Story worktree). Read `campaign-producer-rules.md` first.
Board validation runs once at handoff.

## Why
F-L1b (TASK-260922-2u5jzw, launcher) stopped correctly: v0.5.18 forwards known provider policy selectors
(`codex/policy.go scanNativePolicy`, `claude/policy.go`; README "Permission-mode" calls conflict refusal
"a later leaf"). Decision 0013 D5 keeps every provider spelling inside this module, so the refusal belongs here.

## Normative source
curator-spec `decisions/0018-curator-run-permission-interface.md` choice 4 "Refusal rules" (origin/main):
effective `yolo` plus a conflicting native selector is refused, in `=` and separate forms, including
aliases and `exec` placement:
- Claude: `--permission-mode`, `--allow-dangerously-skip-permissions`, `--restricted`, duplicate
  `--dangerously-skip-permissions`;
- Codex: `-a/--ask-for-approval`, `-s/--sandbox`, `--approve-for-me`, any `--dangerously-bypass-*`, and
  `-c`/`--config` keys `approval_policy`, `sandbox_mode`, `sandbox_permissions`.
Invalid value, repeated flag, unknown env/version fail closed. **Native mode performs no argv inspection**
(raw contract unchanged) — do not add refusals to `native`.

## Deliverable
- The refusal inside the existing per-plugin native-argument grammar (`permission-grammar-v1` or a bumped
  token if the grammar contract changes — say which and why), both placements, `exec` subcommand placement
  for codex.
- Exported, stable typed errors (sentinel or typed error with the selector and placement) the launcher can
  map with `errors.Is/As`; document them in README "Permission-mode" (drop the "later leaf" wording) and
  CHANGELOG.
- Tests: one row per selector × placement for yolo (refused) plus native rows proving no inspection, plus
  forwarded non-conflicting selectors; one narrowing mutant per refusal family applied in a DISPOSABLE copy
  (never your worktree), executed, killed — table in results. `go test ./...` and `go vet ./...` exit 0 in
  bounded calls.
- Release: the next free tag is **v0.5.20** (v0.5.19 is the operator's `4e229cc`). Bump the module's version
  constant/CHANGELOG heading to 0.5.20 as the repo's release convention requires; the orchestrator lands
  and tags after review.

## Handoff
Attach `TASK-260923-3e4i3n_results.md` (selector table, row counts, mutant table, exit codes, exported error
names), check each checklist item citing it, `task-board handoff TASK-260923-3e4i3n --role developer`.
A `run_wrote_outside_worktree … policy warn` block is a warning — verify the status moved to `to-review`.
No real provider CLI calls; fake binaries only. No LOGBOOK.md.
