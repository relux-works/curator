# Review brief: launcher SPEC `0.2.0-draft` (cycle 1)

## Subject
- Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-agent-launcher-spec-0.2`, branch
  `draft/spec-0.2`, head `ffe9b68`, base `6de42d8` (= main). Diff: `git diff 6de42d8..ffe9b68`.
- Read `producer-brief-launcher-0.2.md` and the drafting report (resources on
  TASK-260905-3ewdq0). Authority: curator-spec Decision 0013 Decisions 1, 2, 5, 6.1–6.5, 3.2,
  3.6, 4, 7 (curator-spec main `83de1a5`), Decision 0012 D6/D8, environments.md §10, and the
  0.1.2-draft SPEC as the baseline.

## Review dimensions
1. **Every Decision 0013 D6 item present and exact**: fragment-first ordering with
   `LaunchRequest.Home` and `WorkDir`; `LaunchModeInteractive` requested by name with empty
   `Composition`; default precedence (flags → machine config → `Lineup` fallback with
   `Effort.Recommended`), per-member resolution, stderr print; composition rule (argv order,
   three env layers plus variable channel, `env_names`, stdin, the literal-vs-lookup collision
   rule); tracked mode: exact `ax start … --launch-plan -` invocation, document members with
   their derivations, `argv_suffix` = composed argv minus element 0, `env_literals` = composer's
   own names only, provider-id column, ax-profile flag, session-name default and `--name`,
   ax §2.1 grammar, `ax_handoff_failed` terminal; untracked exec unchanged. Compare each
   sentence against the decision text — flag any weakening, any restatement that diverges,
   and any provider flag the launcher now spells itself (must be none).
2. **Internal consistency**: §3 flag table vs parsing rules vs §6 codes vs §9 items; §5
   unchanged where the decision says it stands; §7 dependency statement; §8/§8.1 version row;
   README and stub `specVersion` say `0.2.0-draft`; `make check` passes (run it).
3. **Facts**: verify `vendorplugin.Lineup`, `Effort.Recommended`, `LaunchRequest.Home` in
   skill-agents-management `91bf945`; verify ax §2.1 name grammar and §14.1 in
   agent-session-manager-spec main; verify Decision 0012 D6 channel spellings quoted.
4. **Commit**: one signed commit, human identity; nothing outside SPEC.md, README.md, the stub
   version constant and its test.

## Constraints
Read-only. Never write into the control root.

## Verdict contract
Attach `TASK-260905-3ewdq0_review-findings-launcher-1.md` (severity, section, quote, what is
wrong, fix). Blocking/major → `development`; else explicit ACCEPT at `to-review`. Do not
mark done. `task-board handoff TASK-260905-3ewdq0 --role reviewer`.
