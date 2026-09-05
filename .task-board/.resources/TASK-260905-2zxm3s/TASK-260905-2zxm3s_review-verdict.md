# Review findings cycle 1: `LaunchModeInteractive` in skill-agents-management

Task `TASK-260905-2zxm3s`. Subject: worktree
`/Users/iv/Developer/ReluxWorks/.worktrees/agents-management-interactive`, branch
`feat/launch-mode-interactive`, head `3edbde8`, base `91bf945`. Authority: curator-spec
Decision 0013 §5 at `83de1a5`. Reviewed 2026-09-05.

## Verdict: ACCEPT

No blocking or major finding. Two minor notes and one nit, none of which change the
verdict; they are recorded for the next cycle or the operator's discretion.

## Change Request `CR-TASK-260905-2zxm3s-1` rev 1 has `repository_delta=empty`, and that is correct

The leaf's deliverable is code in the Go module `skill-agents-management`, which is a
separate repository; the producer brief names its worktree and forbids writing into the
curator control root. curator-spec itself had nothing to change for this task: the
decision it implements was already landed at `83de1a5`. The delivered artefact is the two
signed commits in the external worktree plus the drafting report on the board. An empty
curator-spec delta is the expected shape, not an omission.

## Evidence I gathered myself (not accepted from the report)

### Commits (`git log --show-signature 91bf945..HEAD`)
- `cfd43f2` and `3edbde8`, both `Good "git" signature for oparin@me.com`, author
  `Ivan Oparin <oparin@me.com>`. Working tree clean. 18 files, +1305/-29. No `go.mod`/`go.sum` change.
- Not pushed, no tag (latest tag `v0.5.7` untouched).

### Gates rerun in the worktree (logs under its `.temp/review-agm-1/`)
| Command | Exit |
| --- | ---: |
| `make build vet` | 0 |
| `make test` | 0 (24 `ok` packages, no FAIL) |
| `make regress` | 0 (`ok internal/regress 0.669s`) |
| `gofmt -l .` | clean |

### Binary spellings (this machine, 2026-09-05)
| Binary | Version | Verified |
| --- | --- | --- |
| `claude` | 2.1.261 | `--model <model>` and `--effort <level>` are session flags; `--output-format` is documented "only works with --print"; `-p, --print` absent from the interactive argv |
| `codex` | codex-cli 0.153.2 | `-m, --model <MODEL>` and `-c, --config <key=value>` are top-level flags; `exec` is the non-interactive subcommand; `-p, --profile` exists and is not emitted |
| `pi` | 0.84.2 | `--model <pattern>`; `--print, -p` is non-interactive and absent |
| `agents-infra` | v1.6.1-128-gab60e0d | `agents-infra pi --help` passes through to raw pi help with `--model <pattern>` |

### My own attack test (`.temp/review-agm-1/attack/attack_test.go` in the worktree, through the real registry and `agentic.BuildPlan`)
Positive, per system, with an alias on the request (`ID: alias-name, AliasOf: the-real-model`):
| System | Argv | Stdin | Identity |
| --- | --- | --- | --- |
| claude-code | `--model the-real-model --effort xhigh` | detached | Requested=alias-name Launched=the-real-model |
| codex | `-m the-real-model -c model_reasoning_effort="xhigh"` | detached | same |
| pi | `pi --model the-real-model` (binary = `agents-infra` on PATH) | detached | same |

So alias substitution (`91bf945`) applies before argv in the new mode; `CLAUDECODE` stripped
and run context written in env; marker sweep over `-p --print --output-format
--dangerously-skip-permissions exec --dangerously-bypass-approvals-and-sandbox --yolo
--max-budget-usd --append-system-prompt-file --permission-mode --mcp-config --service-tier
--sandbox --ask-for-approval - --profile spawn /goal* *service_tier*` found nothing on any of
the three argvs.

Refusals confirmed: prefix-only and servers-only composition → `ErrCompositionNotInteractive`
on all three; prompt bytes and prompt path → `ErrParameterNotInteractive` on all three; missing
required effort → `ErrEffortMissing` (claude, codex); effort on pi → `ErrEffortNotTransportable`;
codex profile → plugin refusal naming the profile; agy and gemini → `ErrUnsupportedLaunchMode`;
iota values 0/1/2/3 unchanged, `LaunchMode(4).Valid()` false, `String()` = `interactive`.
Goal/budget/tier on a system that does not declare them are refused earlier by the capability
sentinels (`ErrServiceTierUnsupported` etc.), which is correct ordering, not a gap.

No model or effort literal appears in the non-test diff (grep of the diff for
`high|medium|low|xhigh|sonnet|opus|gpt|claude-` hit only a comment).

### Narrowing mutants on a scratch copy (reviewed worktree untouched)
| Mutant | Killed by |
| --- | --- |
| composition refusal narrowed to `Prefix && Servers` | `pkg/agentic` TestBuildPlanRefusesAnInteractiveLaunchCarryingAComposition; `pi` TestAnInteractiveLaunchRefusesWhatItsGrammarCannotCarry |
| stdin gate removed | `pkg/agentic` TestBuildPlanRefusesAnAttachedInteractiveStdinUnderANonStdinTransport |
| `--dangerously-bypass-approvals-and-sandbox` planted in codex interactive argv | codex TestTheInteractiveArgvCarriesNoExecMarker, TestTheInteractiveArgvIsModelAndEffortOnly; regress TestAnInteractivePlanCarriesNoExecMarkerForAnyMappedSystem |

### Production reachability
`pkg/vendorplugin/spawn.go:231` (`BuildLaunch`) is the production caller of `BuildPlan`, is
mode-generic, requires no prompt, and selects the model from the vendor declaration. The
interactive mode is reachable from the real entry point; no CLI flag is added, which the brief
did not ask for (the launcher is curator's).

## Findings

1. **minor** — `pkg/agentic/systems/claude/interactive_test.go:152`,
   `codex/interactive_test.go:382`, `internal/regress/interactive_test.go` composition case:
   the composition negative populates prefix AND servers together. Under my narrowing mutant
   (`Prefix && Servers`) these three suites stayed green; only the core suite and pi's caught
   it. Fix: split into prefix-only and servers-only subcases, as `pkg/agentic/interactive_test.go`
   already does. The gate is covered, but a per-plugin suite that cannot see the narrowing is
   a weaker net than it claims.
2. **minor** — `pkg/agentic/plan.go` `BuildPlan`, interactive branch: an empty `Model.ID`
   is admitted for claude (`--model ""`) and codex (`-m ""`) while pi refuses it in its
   plugin. This is pre-existing (exec mode admits `--model ""` too, verified) and the
   production path resolves the model upstream, so it is not this change's defect. The brief
   says "model required" for the mode; a core `ErrModelMissing` would make that one rule in
   one place. Operator's call whether to fold it into a follow-up.
3. **nit** — `ErrParameterNotInteractive` is a sentinel the decision did not name. The
   producer flagged it; I agree with the reasoning (refuse over drop, single source) and
   accept it. Decision 0013 §5 may want an erratum line naming it.

Deviations the producer declared (pi absent from regress, four systems undeclared, no
parity golden) are all correct by the cited package rules; verified `parity/doc.go` and the
regress smoke-case rule myself.
