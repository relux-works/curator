# Producer brief: `LaunchModeInteractive` in skill-agents-management

## Where and what

- Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/agents-management-interactive`,
  branch `feat/launch-mode-interactive`, base `91bf945` (= main). Go 1.25.5 module
  `github.com/relux-works/skill-agents-management`.
- Authority: curator-spec Decision 0013, Decision 5 (`decisions/0013-execution-ownership-and-launch-plans.md`
  at `83de1a5` on `~/Developer/ReluxWorks/curator-spec` main) — read it verbatim before
  writing; its grammar constraints are the acceptance bar. Read `AGENTS.md`,
  `docs/architecture.md` (invariants: plans as values, parity bar, single source per fact,
  no injected default, vendor owns vocabulary), `pkg/agentic/system.go`, `plan.go`,
  `pkg/agentic/parity/doc.go`, and every `pkg/agentic/systems/*/` plugin.
- Deliverable: signed commits (`git commit -S`, one or few, each verifying with
  `git log --show-signature`), `make build vet test regress` green in the worktree (record
  the tails), a drafting report `TASK-TASK-260905-2zxm3s` … see Board below. Do not push, tag, or
  open a PR; the orchestrator lands. Never write LOGBOOK.md or anything into the control root.

## The change (Decision 0013 §5, closed)

1. `pkg/agentic/system.go`: add `LaunchModeInteractive` after `LaunchModeManagedSession`
   (append — never renumber existing iota values); extend `Valid()` and `String()`
   (`"interactive"`); doc comment states the grammar constraints from the decision:
   argv = model selection + effort transport only; no print/headless mode, no output-format
   flag, no permission-bypass or unrestricted-mode flag, no goal/assignment-prompt machinery,
   no budget flag, no service-tier flag; stdin `Attached: false` unless the system's
   `EffortTransport` is `EffortTransportStdin`, then exactly the effort encoding.
2. `pkg/agentic/plan.go`: `BuildPlan` in interactive mode refuses a request carrying a
   non-empty `Composition` (prefix or any member) with a new sentinel
   `ErrCompositionNotInteractive` — the MCP composition prefix is the composer's plane.
   `Home`/`WorkDir` carry as today; model required; effort follows `EffortSupport` with
   `ErrEffortMissing` unchanged; undeclared mode → `ErrUnsupportedLaunchMode`.
3. Systems: declare and implement the mode in `claude`, `codex`, and `pi` (the launcher
   §4.2 mapped systems). For each, a single-construction-site argv for the mode: claude
   `--model <id> [--effort <e>]`; codex the interactive (non-`exec`) grammar with `-m`/`--model`
   and `-c model_reasoning_effort=…` as the plugin already spells the transport — verify
   against the installed binaries (`claude --help`, `codex --help`) that the spellings are
   accepted in interactive invocation and record versions; pi per its plugin
   (`EffortTransportNone`: model flag only — verify `pi --help`). Do NOT add the mode to
   `agy`, `gemini`, `muse`, `qwen` in this change unless it is a one-line declaration with
   a trivially correct argv; if you skip them, say so in the report — they refuse with
   `ErrUnsupportedLaunchMode`, which is the decision's rule for an undeclared mode.
4. Goldens: the decision requires one argv-parity golden per system for the new mode under
   the existing parity bar and one negative golden per system proving the exec-mode markers
   (`-p`, `--output-format`, `--dangerously-skip-permissions`, `exec`,
   `--dangerously-bypass-approvals-and-sandbox`, `--max-budget-usd`, goal machinery) are
   absent. Read `pkg/agentic/parity/doc.go` first: existing goldens were captured by the
   SOURCE repository and this package "captures nothing". Interactive mode has no source
   capture, so add the interactive goldens as plugin-local tests (in each system's
   `*_test.go`, in the style of `systems/claude/args_test.go`) that assert the exact argv,
   env delta, and stdin, plus the negative marker assertions — and say in the report why
   they are not `parity/testdata/goldens` entries. If the parity package has a declared way
   to add non-source goldens, use it instead and cite it.
5. `internal/regress`: add the cross-cutting negative (an interactive plan containing any
   exec marker fails) if the package's pattern admits it; otherwise report why not.
6. Docs: `docs/architecture.md` (launch modes list + the new invariant sentence),
   `README.md` line ~72 area, `docs/shipped-state.md` if it enumerates modes,
   `docs/consuming-the-module.md` "Launch and availability entry points" (one paragraph:
   who calls interactive mode and what it must not contain). Keep prose density of the repo.
7. Version: do not tag. Note in the report that a compatible minor (next after the latest
   `v0.5.x` tag) is the operator's release call.

## Board

Task `TASK-260905-2zxm3s` on the curator board (`TASK_BOARD_DIR` is set). Attach
`TASK-260905-2zxm3s_drafting-report.md` as an outcome resource: commits with signature lines, gate
outputs, per-system argv table (mode → exact argv), verified binary versions, skipped
systems with reasons, docs touched. Then `task-board handoff TASK-260905-2zxm3s --role developer`.
