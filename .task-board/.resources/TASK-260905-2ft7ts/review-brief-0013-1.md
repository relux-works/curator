# Review brief: Decision 0013 draft (execution ownership and launch plans)

## Subject
- Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-decision-0013`,
  branch `draft/decision-0013-execution-ownership`, head `71ac9d1`, base `b4f29cd` (= main).
- Document: `decisions/0013-execution-ownership-and-launch-plans.md`. Also read the
  producer's drafting report (outcome resource on TASK-260905-2ft7ts) and the
  producer brief `producer-brief-0013.md` (precondition resource): the brief's
  contract items 1–8 are the acceptance bar.
- Inputs to judge against (read-only): the same sources the producer brief lists —
  curator-spec `decisions/0010` D6/D10, `decisions/0012` Context/D6/D8,
  `protocol/environments.md` §7.3/§10/§11, `protocol/registry.md` §1;
  `pre-implementation-review-v3.md` (STORY-260901-zddtn8) M1/M2/M7/M8/M15/M16;
  curator-agent-launcher `SPEC.md` 0.1.2-draft; agent-session-manager-spec main
  `SPEC.md` §5.1/§7.3/§7.5/§13.1/§13.10/§14.1 and PR #1 diff
  (`git diff main...origin/draft/curator-environment-integration`);
  skill-agents-management main (`pkg/agentic/system.go`, `plan.go`,
  `systems/claude/args.go`, `systems/codex/args.go`, `systems/pi/`, `docs/architecture.md`).

## Review dimensions
1. **Settled decisions honored, none reopened**: Option A exactly as the brief states;
   `--launch-plan FILE|-`; `SpawnPlan.stdin` grown (not refused); `LaunchModeInteractive`
   with no print mode / output format / permission bypass / goal machinery; launcher
   fragment-before-plan with `LaunchRequest.Home`; default model/effort ownership in the
   launcher; tracked mode delegating to ax; PR #1 revision items (operation section,
   refuse-on-drift for `class: system`, CCJ-1 digest). Flag any item missing, weakened,
   or turned back into an open question.
2. **Closedness and validatability of the plan document**: every member typed and
   bounded; `argv` vs `argv_suffix` exclusivity; env_names/env_literals disjointness and
   §5.1 limits; stdin bound and encoding; unknown-member rejection; a typed refusal
   before any Session Record; resume replay semantics unambiguous; plugin capability
   declaration named. Attack: can a caller smuggle a permission-bypass or a secret; can
   two readings of the argv[0] rule produce different launches; is anything left to
   "the implementation".
3. **agents-management fit**: the mode respects the module invariants (plans as values,
   single construction site per system, no injected default, vendor owns vocabulary);
   `EffortTransportStdin` interaction with "empty stdin" stated correctly; nothing here
   requires the launcher to spell provider flags.
4. **Launcher composition rule**: the argv concatenation order and the three env layers
   are stated once, closed, and consistent with launcher §4.4/§5 and 0012 D6 (MCP
   channel flags, `env_names` allowlist); `ax_handoff_failed` still terminal; the
   `Lineup` fallback identifier verified against the module (grep it).
5. **ax consistency**: record immutability (§5.1) preserved; no change to lease, fencing,
   checkpoint, materialization semantics; task-board launches untouched; profile-pin
   re-keyed to the 0012 lock identity without contradicting 0012 D8; `fragment-digest`
   over CCJ-1 (registry §1) and the `system-modules` signal defined.
6. **Citations and facts**: every § and identifier exists at the cited commit; the
   number-reconciliation statement (0011 on the swift-driver draft branch, 0012 landed)
   is true; every "verified" claim in the decision is backed by the report or by your own
   check on the installed binaries/sources — anything else is docs-confidence and must be
   labeled or removed.
7. **House style**: matches 0012 (sections, density, English); Rejected alternatives
   records Option B with the review's reasoning; Compatibility/Security/Consequences/
   Open questions present and non-empty.

## Constraints
Read-only: no edits, commits, pushes. Shell tooling; all checkouts read-only. Never
write LOGBOOK.md or anything into the control root.

## Verdict contract
Attach `TASK-260905-2ft7ts_review-findings-0013-1.md` (outcome resource): per finding —
severity (blocking|major|minor|nit), section, quote, what is wrong, fix. Blocking or
major → route to `development`; otherwise an explicit ACCEPT and leave at `to-review`
(the orchestrator lands). Do not mark done. Then `task-board handoff TASK-260905-2ft7ts --role reviewer`.
