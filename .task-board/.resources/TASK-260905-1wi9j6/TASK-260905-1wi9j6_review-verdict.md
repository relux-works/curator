# TASK-260905-1wi9j6 — review verdict (cycle 1): ACCEPT

Change Request `CR-TASK-260905-1wi9j6-1` revision 1. Subject: skill-agents-management worktree
`/Users/iv/Developer/ReluxWorks/.worktrees/agents-management-interactive-followups`, branch `feat/interactive-mode-follow-ups`,
head `93abeae`, base main `052dee1`. Diff reviewed: `git diff 052dee1..93abeae` (7 files, +171/-28, no go.mod/go.sum churn).

## Why the empty repository delta is the right outcome

The CR is recorded against the curator-spec story worktree (base `fcdb9ba`, candidate tree identical). All code for this leaf lives in
skill-agents-management, a separate repository, and is there at `93abeae`. The only spec-side item, the Decision 0013 erratum line for
`ErrParameterNotInteractive`, is explicitly scheduled by the task description for "when the decision is next touched (environments 1.1 batch)"
and is recorded in the drafting report. No curator-spec file should have changed for this leaf.

## Verified myself (not accepted from the report)

- Signature: `git log -1 --format=%G?` = `G`, author Ivan Oparin <oparin@me.com>. One commit, not pushed. Worktree clean.
- Gates at 93abeae, standalone processes, logs in `.temp/review-1wi9j6/` of that worktree:
  make build 0 · make vet 0 · make test 0 (24 ok packages) · make regress 0 · `gofmt -l .` empty.
- Production call site: `pkg/vendorplugin/spawn.go:231` → `agentic.BuildPlan`. The sentinel sits after `PrepareLaunchRequest` and before every
  contract check and plugin surface; pi's `PrepareLaunchRequest` does not touch `Model.ID`; the alias rewrite at plan.go:264 runs after the gate,
  so no default is injected ahead of the refusal.
- AC coverage: 5 of 5 rows driven through `agentic.BuildPlan` (table in the drafting report checked against the committed tests). Stated bound
  accepted: agy/gemini/muse/qwen covered by the core double only.

## Mutants applied on a scratch copy (suite `go test ./pkg/agentic/... ./internal/regress/...`, logs `mutant-M{1,2,3}.log`)

| Mutant | Narrowing | Failing tests | Survivor |
|---|---|---|---|
| M1 plan.go:153 `!IsZero()` → `len(Prefix)>0 && len(Servers)>0` | refuses only a full composition | core `TestBuildPlanRefusesAnInteractiveLaunchCarryingAComposition/{prefix_only,servers_only}`; claude `TestAnInteractiveLaunchRefusesWhatItsGrammarCannotCarry/a_composition_of_{prefix_only,servers_only}_is_refused…`; codex same; pi `…/a_composition_is_refused_with_the_decision's_sentinel,_before_the_grammar_refusal`; regress `TestAnInteractivePlanRefusesACompositionForEveryMappedSystem/{claude-code,codex}/{prefix_only,servers_only}` | no |
| M2 drop `TrimSpace` | admits a whitespace model id | core `TestBuildPlanRefusesAnEmptyModelInEveryMode`; claude `…/an_empty_model_is_refused_in_interactive_and_exec_mode_by_the_core_sentinel` | no |
| M3 gate only when `mode == LaunchModeInteractive` | exec admits `--model ""` again | core `TestBuildPlanRefusesAnEmptyModelInEveryMode`; claude + codex `…/an_empty_model_is_refused…`; regress `TestAPlanRefusesAnEmptyModelForEveryMappedSystemInEveryMode/{claude-code,codex}/exec` | no |

Finding (1) from review-findings-agm-1 is closed: the `Prefix && Servers` narrowing now fails per plugin in claude, codex and regress, not only in core.
No source-text gate is involved; no token-preserving mutant applies.

## Minor, non-blocking

- N1: codex's empty-model subcase uses `""` only; the whitespace class in codex is covered by core and claude, not by the codex package itself.
  Optional: use `"  "` like claude does. Not a bypass — the gate is one line in core.

## Decisions confirmed

- ErrModelMissing adopted in core, every mode, before dispatch; pi's own refusal retained deliberately with a note; docs/architecture.md invariant 7 consistent with the code.
- Decision 0013 erratum: scheduled for the environments 1.1 batch, spec side. Not part of this delta.

Verdict: ACCEPT. repeat-of: none. Landing (fast-forward of 93abeae onto skill-agents-management main) is the producer/integration side's step.
