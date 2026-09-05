# TASK-260905-1wi9j6 — interactive-mode review follow-ups: drafting report

Repo: skill-agents-management, worktree `/Users/iv/Developer/ReluxWorks/.worktrees/agents-management-interactive-followups`,
branch `feat/interactive-mode-follow-ups`, base main `052dee1` (main had moved past `3edbde8` by one commit; branch is rebased on it by construction).
Head: `93abeae` (one signed commit, `git log --format=%G?` = `G`, author key oparin@me.com). Not pushed, no tag, no PR (brief: no push).

## Items

1. **Composition negatives split** — `pkg/agentic/systems/claude/interactive_test.go`, `pkg/agentic/systems/codex/interactive_test.go`,
   `internal/regress/interactive_test.go` now drive three shapes per plugin: prefix+servers, prefix only, servers only.
2. **ErrModelMissing — adopted.** Core sentinel `agentic.ErrModelMissing` in `pkg/agentic/plan.go`, checked in `BuildPlan` right after
   `PrepareLaunchRequest` and before every contract check and plugin surface, for every mode. `strings.TrimSpace(req.Model.ID) == ""` → refusal
   naming the system and the mode. pi's own interactive refusal in `pi/args.go` is kept as a second line of defence (tested directly).
   Invariant 7 added to `docs/architecture.md`.
3. **Decision 0013 erratum** (`ErrParameterNotInteractive`) — left to the spec side, recorded here only: file the erratum line when 0013 is next touched
   in the environments 1.1 batch. Not a code change in this repo.

## AC coverage — 5 of 5 rows driven through the production entry point (`agentic.BuildPlan`, `pkg/agentic/plan.go`)

| AC row | Named committed test | Production call site |
|---|---|---|
| prefix-only refused, claude | `claude.TestAnInteractiveLaunchRefusesWhatItsGrammarCannotCarry/a_composition_of_prefix_only…` | `BuildPlan` → `refuseNonInteractiveParameters` |
| servers-only refused, claude/codex/regress | same test in codex; `regress.TestAnInteractivePlanRefusesACompositionForEveryMappedSystem/{claude-code,codex}/{prefix_only,servers_only}` | same |
| empty model refused, interactive | `agentic.TestBuildPlanRefusesAnEmptyModelInEveryMode`; `pi…/a_missing_model_is_refused_by_the_core_sentinel…`; `regress.TestAPlanRefusesAnEmptyModelForEveryMappedSystemInEveryMode/*/interactive` | `BuildPlan` ErrModelMissing check |
| empty model refused, exec | `claude…/an_empty_model_is_refused_in_interactive_and_exec_mode…`, codex same, `regress…/*/exec`, core (exec, dry-run, managed-session) | same |
| refusal before plugin dispatch | `agentic.TestBuildPlanRefusesAnEmptyModelInEveryMode` asserts ResolveBinary/Argv/ChildEnv/Stdin call counts = 0 | same |

Stated bound: agy, gemini, muse and qwen are covered by the core double test only (they do not declare interactive mode and have no per-plugin empty-model test).

## Mutants (scratch copy `/tmp/agm-mutant-1wi9j6`, suite `go test ./pkg/agentic/... ./internal/regress/...`; logs `.temp/TASK-260905-1wi9j6/mutant-*.log` in the story worktree)

| Mutant | Narrows the gate to | Named failing tests | Survivor? |
|---|---|---|---|
| M1 `!IsZero()` → `len(Prefix)>0 && len(Servers)>0` | refuses only a full composition | pkg/agentic: `TestBuildPlanRefusesAnInteractiveLaunchCarryingAComposition/{prefix_only,servers_only}`; claude: `…/a_composition_of_{prefix_only,servers_only}_is_refused…`; codex: same names; pi: `…/a_composition_is_refused_with_the_decision's_sentinel,_before_the_grammar_refusal`; regress: `TestAnInteractivePlanRefusesACompositionForEveryMappedSystem/{claude-code,codex}/{prefix_only,servers_only}` | no (exit 1) |
| M2 drop `TrimSpace` on Model.ID | admits a whitespace model id | pkg/agentic: `TestBuildPlanRefusesAnEmptyModelInEveryMode`; claude: `…/an_empty_model_is_refused_in_interactive_and_exec_mode_by_the_core_sentinel` | no (exit 1) |
| M3 gate only when `mode == LaunchModeInteractive` | admits `--model ""` in exec again | pkg/agentic: `TestBuildPlanRefusesAnEmptyModelInEveryMode`; claude + codex: `…/an_empty_model_is_refused_in_interactive_and_exec_mode…`; regress: `TestAPlanRefusesAnEmptyModelForEveryMappedSystemInEveryMode/{claude-code,codex}/exec` | no (exit 1) |
| M4 delete the check | gate absent (existence only) | all M3 tests plus `pi…/a_missing_model_is_refused_by_the_core_sentinel…` and regress `/*/interactive` | no (exit 1) |

No source-text gate is involved, so no token-preserving mutant applies.

## Gates at head 93abeae (each run as a standalone process, real exit codes; logs in `.temp/TASK-260905-1wi9j6/`)

| Command | Exit | Log |
|---|---:|---|
| `make build` | 0 | make-build-02.log |
| `make vet` | 0 | make-vet-02.log |
| `make test` | 0 (24 ok packages) | make-test-02.log |
| `make regress` | 0 | make-regress-02.log |
| `gofmt -l .` | 0, no output | — |

## Notes
- Commit history: first attempt cut two commits, but the first one did not compile alone (per-plugin tests referenced the sentinel). Re-cut as one signed commit.
- Managed story worktree (curator-spec) untouched except `.temp/` logs; control root untouched.
