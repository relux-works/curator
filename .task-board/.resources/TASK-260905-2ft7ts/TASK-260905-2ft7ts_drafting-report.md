# Drafting report: Decision 0013 — execution ownership and launch plans

- File: `decisions/0013-execution-ownership-and-launch-plans.md` (704 lines, the only change)
- Worktree: `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-decision-0013`
- Branch: `draft/decision-0013-execution-ownership`, base `b4f29cd`
- Commit: `71ac9d13a29187db04ebc23be7fecc4af5ce8924`
- `git log --show-signature -1`: `Good "git" signature with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`
  (followed by `No principal matched.` — `gpg.ssh.allowedSignersFile` points at a
  stale `/private/tmp/curator-spec-rc8-verify.*/maintainers.allowed_signers`; a local
  verification-config artifact, not a signing defect. Author `Ivan Oparin <oparin@me.com>`.)
- Not pushed, no tag, no PR. `git status` clean after commit.

## Contract items → sections

| Brief item | Section(s) in the decision |
|---|---|
| 1 Numbering and status | Status (numbering paragraph, acceptance paragraph, references paragraph) |
| 2 Ownership model (Option A), why A over B | Decision 1, Decision 2, Rejected alternatives (Option B, verbatim review reasoning) |
| 3 `ax start --launch-plan` | Decision 3.1 surface + exclusivity; 3.2 document shape (schema, argv/argv_suffix, env_names, env_literals, stdin, extensions); 3.3 validation (`launch_plan_invalid` + `field`); 3.4 record (final argv, `ax.launch-plan-request` key); 3.5 plugin contract (`caller_launch_plan`, resume); 3.6 execution profile |
| 4 `SpawnPlan.stdin` / Launch Plan `stdin` | Decision 4 |
| 5 `LaunchModeInteractive` | Decision 5 |
| 6 Launcher SPEC 0.2 | Decision 6.1 (M7 ordering/home), 6.2 (M8 defaults, `Lineup`), 6.3 (composition rule), 6.4 (tracked delegate, session name, extension keys, untracked exec, `ax_handoff_failed`), 6.5 |
| 7 ax PR #1 revision items | Decision 7 (items 1–8), Decision 8 (pin reconciliation) |
| 8 Compatibility, security, consequences, open questions | the four closing sections |

## Decisions taken where the brief left a choice

- Schema id: `urn:ax:schema:launch-plan-request` / `1.0.0` (ax §1.6 convention) instead of `ax-launch-plan-request-v1`.
- `argv` complete form: element 0 is the executable as the plugin resolves it; `argv` + `--profile yolo` refused; composer always uses `argv_suffix`.
- Record: final argv in `launch_plan.argv`; caller form/suffix/digest under ax-generic key `ax.launch-plan-request` (not the whole document — 65,536-byte extension bound). Final argv obtained by a planning-role `launch` before persistence (ax §13.1 step order).
- Capability names: `caller_launch_plan` (8th) and `stdin_resume_replay` (9th).
- Stdin: `{encoding: utf-8|base64, bytes}`, 65,536 decoded bytes, no replay on resume by default.
- System-modules fact: fourth key `works.relux.curator.system-modules` (boolean), not folded into fragment data.
- Session name: `<env-id>-<profile-name>-<utc-stamp>` with `--name` override.
- Refuse-on-drift for system modules uses existing `policy_refused`; capability refusal uses existing `capability_unavailable`.

## Facts verified (source + commit)

| Fact | Source |
|---|---|
| 0011 exists only on `draft/TASK-260728-1yhuqi-swift-driver` head `604d525`; main `b4f29cd` has 0012 and no 0011 | `git ls-remote --heads origin`; `ls decisions/` in both checkouts |
| 0012 Context names M1 Option A and "next free number after reconciliation with the swift-driver draft's 0011"; Decision 6 "Under Option A the launcher, `curator-run`, is the single composer"; Decision 3 lock hash is the pin for `works.relux.curator.profile-pin`; Decision 8 fragment `profile.lock_sha256`, `precedence` object, `mcp` section | curator-spec `b4f29cd` |
| 0010 Decision 6 (four-plane launcher boundary, spawn/context/session planes), Decision 10 (table, drift warn-and-continue recommendation, PR to ax) | curator-spec `b4f29cd` |
| environments §10.1 `env resolve`, §10.2 fragment (`system_prompt` present exactly when chain carries an applicable system module), §10.3 boundary, §7.3 channels, §11 discovery | curator-spec `b4f29cd` |
| registry §1 CCJ-1 rules | curator-spec `b4f29cd` |
| Review v3 M1 (Option A/B text, recommendation), M2, M7, M8, M15, M16, action plan item 1 | `.task-board/.resources/STORY-260901-zddtn8/pre-implementation-review-v3.md` |
| Launcher SPEC `0.1.2-draft`: §1 non-goals, §3 CLI, §4.1–4.5, §6 codes, §7, §8, §9 | curator-agent-launcher `6de42d8` |
| ax §5.1 Launch Plan limits (argv 1..128, 1–4,096 B each, 65,536 total; env_names 0..64 grammar; env_literals 0..64 × 4,096, disjoint; contains_secrets false); §1.6 extension rules (reverse-DNS, ≤64, 65,536 B, depth 4); §2.1 session-name grammar; §2.4 profiles; §7.3 seven capability names; §7.5 `launch` request/`SpawnPlan` row/`resume` request; §13.1 steps; §13.10; §14.1 `ax start` row; §15.1 error shape (`details`), §15.3 codes (`invalid_arguments` 2, `capability_unavailable` 6, `policy_refused` 16, `provider_protocol_error` 13); §16.2 exclusions | agent-session-manager-spec `28bf96d` (tag v0.5.0) |
| PR #1 diff: three keys, §7.5 "launcher merges … into env_literals" paragraph and "adds no member to SpawnPlan", §13.10 drift paragraph, §14 informative note | `git diff main...origin/draft/curator-environment-integration -- SPEC.md`, head `d7075e1` |
| `LaunchMode` = Exec/DryRun/ManagedSession, `Valid()`, `String()`; `LaunchRequest.Home` "load-bearing" comment, `WorkDir`; `EffortTransport` None/Argv/Stdin; `StdinPayload{Attached, Bytes}`; `EffortSupport` None/Required; `Composition`/`Composition.Prefix`; `Capabilities.LaunchModes` | skill-agents-management `944c7b4` `pkg/agentic/system.go` |
| `Plan{Binary, Argv, Env, Stdin, WorkDir}`, `BuildPlan(r, req, mode)`, `ErrUnsupportedLaunchMode`, `ErrEffortMissing` | `pkg/agentic/plan.go` |
| claude argv `-p --output-format json --model … [--effort …] [--max-budget-usd …] --dangerously-skip-permissions`, `compositionArgvPrefix` | `pkg/agentic/systems/claude/args.go` |
| codex argv `exec -m …`, `--dangerously-bypass-approvals-and-sandbox`, `-c model_reasoning_effort=…` | `pkg/agentic/systems/codex/args.go` |
| `vendorplugin.Lineup(models []Model) []RankedModel` | `pkg/vendorplugin/lineup.go` |
| `Effort.Recommended` declaration | `pkg/vendorplugin/vendor.go`, `spawn.go` |
| architecture invariants: parity bar, no injected default, single source per fact, one golden set per system | `docs/architecture.md` |

## Divergence from the brief (verified)

- The brief says `pkg/agentic/systems/pi/` is the stdin effort transport. At `944c7b4`
  `pi` declares `EffortTransportNone` and `stdin()` returns an empty `StdinPayload`;
  the only `EffortTransportStdin` system is `qwen` (`pkg/agentic/systems/qwen/qwen.go`,
  `stdin.go`). Decision 4 is written against the verified fact and names qwen.

## Not verified (docs-confidence; kept out of the decision text or hedged)

- Whether ax's `launch` plugin operation is deterministic/idempotent today — the
  decision *requires* it for `caller_launch_plan` plugins rather than asserting it.
- Whether any pinned tool treats everything after the last flag as the user turn
  (the composition-order rationale) — left to launcher SPEC 0.2 per-tool verification,
  as the brief instructed.
- Core §2 identifier grammar fitting inside the ax session-name grammar — asserted from
  environments.md's description ("no separators, no traversal"), core §2 itself not re-read.
- Exact schema-version consequence of growing the closed `capability_names` registry —
  explicitly left to the ax maintainer.
- `ErrCompositionNotInteractive` is a proposed new sentinel, not an existing identifier.
