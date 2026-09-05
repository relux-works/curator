# Review brief: `LaunchModeInteractive` in skill-agents-management (cycle 1)

## Subject
- Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/agents-management-interactive`, branch
  `feat/launch-mode-interactive`, head `3edbde8`, base `91bf945` (= main). Diff:
  `git diff 91bf945..3edbde8`.
- Read the producer brief `producer-brief-agm-interactive.md` and the drafting report
  (resources on TASK-260905-2zxm3s). Authority: curator-spec Decision 0013 Decision 5
  (`decisions/0013-execution-ownership-and-launch-plans.md` on curator-spec main `83de1a5`).

## Review dimensions
1. **Grammar constraints hold, attacked not read**: for each system declaring the mode, build
   an interactive plan in a Go test of your own under the worktree's `.temp/` (or run the
   producer's tests with `-v`) and assert the argv contains only model selection and effort
   transport; grep the argv for every exec marker (`-p`, `--output-format`,
   `--dangerously-skip-permissions`, `exec`, `--dangerously-bypass-approvals-and-sandbox`,
   `--yolo`, `--max-budget-usd`, `--append-system-prompt-file`, `/goal`, service-tier flags);
   confirm stdin `Attached:false` for claude/codex/pi; confirm a non-empty `Composition`
   is refused with `ErrCompositionNotInteractive`; confirm an undeclared system refuses with
   `ErrUnsupportedLaunchMode`; confirm `ErrEffortMissing` still fires for a required-effort
   model without effort; confirm no default is injected anywhere (grep for a model/effort
   literal in the mode's code paths).
2. **Module invariants** (`docs/architecture.md`): plans as values, single construction site
   per system (the interactive argv must come from the same `Args`/construction function,
   not a parallel spelling), vendor owns vocabulary, alias resolution (`91bf945`) still
   applied before argv in the new mode; iota order of existing modes unchanged.
3. **Spellings verified on binaries**: rerun `claude --help`, `codex --help`, `pi --help` on
   this machine; confirm each interactive flag the plugins emit exists for an interactive
   invocation (not only `exec`/`-p`); record versions.
4. **Gates**: `make build vet test regress` in the worktree (record tails); `gofmt -l` clean;
   no `go.mod`/`go.sum` churn beyond need.
5. **Docs**: architecture/README/consuming-the-module updated accurately; no claim of a
   tagged release.
6. **Commits**: signed, human identity, scoped; no push, no tag.

## Constraints
Read-only on the worktree (scratch only under its `.temp/`). Never write into the control root.

## Verdict contract
Attach `TASK-260905-2zxm3s_review-findings-agm-1.md`: severity (blocking|major|minor|nit),
file:line, quote, what is wrong, fix, plus your own test evidence. Blocking/major →
`development`; else explicit ACCEPT at `to-review`. Do not mark done.
`task-board handoff TASK-260905-2zxm3s --role reviewer`.
