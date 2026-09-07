# Goal: ship `curator run` (untracked) and migrate relux-agents-infra onto Curator

Attach this file as the precondition resource of the epic you create for this goal.
The ax integration (stage d of the previous epic) is explicitly out of scope.

## Objective

Two deliverables, in this order:

1. **`curator-run` shipped.** The launcher of `relux-works/curator-agent-launcher`
   implemented to its SPEC (0.2.1-draft, revised only by recorded errata), with
   the untracked path complete on the operator's machine for `claude_code`,
   `codex_cli` and `pi`, and the tracked path implemented against the Decision
   0013 document and a fake `ax` binary in tests only.
2. **relux-agents-infra migrated.** Everything agents-infra installs that
   Curator can express — root-context instructions, skills, MCP declarations,
   model/effort defaults, launching — moves into Curator packages, profiles and
   `curator run`; agents-infra keeps only the residual Curator cannot express
   and says so in its README.

## Sources of truth (read first; never restate divergently)

- Launcher: `curator-agent-launcher` main (`SPEC.md` 0.2.1-draft, §3 CLI, §4
  composition, §5 system prompt, §6 diagnostics, §7 planned dependency).
- `curator-spec` main: `decisions/0013-execution-ownership-and-launch-plans.md`
  (D1 ownership, D3 `ax start --launch-plan` document, D4 stdin, D5
  `LaunchModeInteractive`, D6 launcher requirements, D7 extension keys),
  `decisions/0012-context-packages-and-semver-locks.md` (D6 MCP packages and
  launch channel, D8 fragment members), `protocol/environments.md` 1.1 (§7.1
  home variables, §7.3 channels, §7.4 seeds/passthrough, §7.6 Xcode targets,
  §7.8 MCP channels, §9.4 profile-scoped skills, §9.5/9.6 onboarding and
  takeover, §10 `env resolve --repair` and `launch-env-fragment-v1`, §11
  umbrella discovery, §12 machine configuration), `protocol/registry.md` §1
  (CCJ-1).
- `relux-works/curator` main: the shipped reference implementation (stages
  a–c): `cmd/curator` (`env resolve/status`, `profile *`, `umbrella.go`
  discovery of `curator-run`), `internal/envfragment`, `internal/mcp`,
  `internal/contextpkg`. Run the installed binary; do not infer its output
  from the spec.
- `relux-works/skill-agents-management` (`pkg/agentic`): `BuildPlan`,
  `LaunchModeInteractive` (landed 2026-09-05), `vendorplugin.Lineup`,
  provider-limits verdicts keyed by (provider, home). Import it at a tag.
- relux-agents-infra @ `dee5403` (`/Users/iv/Developer/IV/relux-agents-infra`)
  and the mapping `agents-infra-to-curator-mapping.md` (attach it beside this
  file): what it installs (§1), the 14 instruction modules and the `@` index
  (§2), configs (§3), launchers (§4), skills (§5), per-project policy (§7).
- Board resource `pre-implementation-review-v3.md` on STORY-260901-zddtn8 for
  the M-items that bind the launcher: M1, M2, M5 (channel spellings), M6
  (`path_prepend`, umbrella hardening), M7 (fragment before plan), M8
  (defaults), M16 (fragment digest over CCJ-1).

## Settled decisions (do not reopen)

Option A: `curator-run` is the single composer; ax is out of scope here.
Interactive plans carry no permission bypass; yolo exists only as
`--ax-profile yolo` and is therefore unavailable in this scope. Model/effort
defaults are the launcher's (`defaults.json`, operator over machine unless
locked). MCP only into managed homes, allowlist over source identities. csk
cleanup stays surface naming. No tags or GitHub Releases without the
operator's explicit command. Environments.md stays revision 1.1; launcher
SPEC bumps to 0.3.0-draft only for recorded errata found by implementation.

## Workstream A — launcher

A0. **Verification before code.** On this machine, record as board evidence:
the exact `curator env resolve <env> --repair --format json` output for a
real profile per env; the agents-management API at the tag you import
(`BuildPlan` request/plan shapes, `LaunchModeInteractive`, `Lineup`,
`ErrEffortMissing`, `ErrCompositionNotInteractive`); the argv boundary of
claude/codex/pi interactive launches (channel flags must precede native args;
verify what each tool treats as the user turn) at the pinned releases; codex
`-p` layer behaviour and pi `SYSTEM.md`/`APPEND_SYSTEM.md` detection. Any
SPEC statement the evidence falsifies becomes an erratum story first.

A1. **Implementation**, per SPEC section, each a story with goldens:
§3 parsing and usage errors; §4.1 fragment subprocess, closed parsing, CCJ-1
digest; §4.2 mapping; §4.3 `defaults.json` family with `locked` and the
per-member precedence, lineup fallback, the stderr line-group; §4.4
`BuildPlan` request (Home = managed home, empty composition); §4.5 argv
order, four env layers, `env_names` minus literals with warnings, stdin
mapping, codex layer stat; §4.6 untracked exec, and the tracked handoff
document + `ax start` subprocess exercised only against a fake `ax` in tests
(`ax.json` absent or `enabled:false` on this machine); §5 system-prompt
opt-in with the file-kind probe and warnings; §6 diagnostics and exit codes.
Parity bar as in agents-management: argv, environment, stdin bytes, side
effects, one golden per env per mode, negative goldens for forbidden flags.

A2. **Integration on the operator's machine.** `make install` to
`~/.local/bin/curator-run`; `curator run <env> --profile <p> -- <args>`
launches through umbrella discovery for the three envs; `curator run` with a
stale home repairs it (`--repair` always); MCP flags reach the tool; the
model/effort line-group prints. Record transcripts as evidence.

A3. **Release readiness**: README, `--help`, CHANGELOG section for 0.1.0;
CI on the launcher repo (lint, test, race, goldens). No tag.

## Workstream B — migration

B1. **Context packages.** New repository `relux-works/relux-root-context`
(private by default; the operator flips visibility). Packages with
`agent-context.json`, strict `v`-tags, weights per Decision 0012:
`relux-root-context-core` (STRUCTURE, TOOLS, SKILLS, SKILL_TRIGGERS, DOCS,
DIAGRAMS, PLATFORM), `-workflow` (WORKFLOW, TESTING), `-style` (STYLE, lowest
weight), `-claude` (EXTERNAL_RESOURCES, REMOTE_AGENTS with
`environments: ["claude_code"]`), `-attachments` (ATTACHMENTS + the
attachments CLI skill), and the umbrella `relux-root-context-ivan` requiring
them plus skills and MCP. Module bytes come from `.instructions/*.md` with
runtime-specific text split into per-env modules rather than rewritten;
rewrite only references to agents-infra commands that this migration
retires (`agents-infra codex|claude` → `curator run`).

B2. **Skills.** `pdf`, `skill-creator` (or its standalone successor per
agents-infra STORY-260903-2pxzvb) and an `agents-attachments` skill that
ships the CLI as a skill utility get Curator skill manifests in their own
repositories and are declared in the umbrella's `requires.skills` with
ranges. `relux-agents-infra`'s own SKILL.md stays with the residual.

B3. **MCP declaration packages.** Repository `relux-works/relux-mcp`
with `agent-mcp.json` packages `figma` (https url), `lldb` (bare `lldb-mcp`),
`safari` (bare command: ship a `safaridriver-mcp` wrapper on PATH since
absolute commands are rejected). Machine allowlist configured; `env_names`
for bearer tokens. Declared in the umbrella's `requires.mcp`.

B4. **Defaults.** `~/.config/curator-run/defaults.json` carrying the model
and effort agents-infra set today (codex `model`/`model_reasoning_effort`,
claude `model`).

B5. **Onboarding on the operator's machine.** `curator profile install
<umbrella> --use --takeover` after a recorded backup; `env status` current
for `claude_code`, `codex_cli`, `pi`; Xcode target consent only if the
operator says so. Before takeover, snapshot what agents-infra installed
(`agents-infra doctor global`) as evidence.

B6. **agents-infra residual.** Through agents-infra's own board and PR
canon: remove instruction sync and `@` rendering, skills fan-out, the MCP
registry and the MCP composition in `agents-infra claude|codex` (retire
those launchers in favour of `curator run`, or reduce them to a printed
deprecation pointing at `curator run`); keep `claude-settings.json`
linking, the codex `config.toml` merge, `.rules`, the pi local-model runtime
(broker, profiles, targets, harness), `lldb-mcp` wrapper, attachments
manifest contract. README rewritten to state exactly what remains and why.

B7. **Spec gaps, filed not worked around.** Open curator-spec issues (or
decision drafts, no landing) for: tool-configuration surfaces (claude
`settings.json`, codex `rules/`), per-project policy (MCP opt-in, model,
yolo) versus profile-level, skill command roots in managed homes
(`path_prepend`). Do not build agents-infra workarounds for them.

## Operating rules

- Board first: one epic on the curator board (`../curator/.task-board`),
  stories with worktrees under `~/Developer/ReluxWorks/.worktrees/`, producer
  and reviewer subagents on `claude-fable-5-1` at LOW reasoning effort,
  briefs as precondition resources, findings as outcome resources.
- Hygiene learned the hard way: close every story with a board-state commit
  on curator `main`; runs never write LOGBOOK.md or anything else into the
  control root; `git pull --ff-only` in the control root before every
  board-state push; stage by explicit pathspec, never `git add -A`; land by
  fast-forward push of the exact reviewed head spelled `${SHA}:refs/heads/main`
  with hard-fail checks (local = remote = PR head, origin/main is ancestor,
  signature Good); delete a branch only after the PR reads MERGED.
- Delivery canon: branch → PR → comment-review verdict with evidence → green
  checks → fast-forward push; signed commits by the configured human identity,
  never rewritten (rebase with `-S`, prove identity by `git range-diff`).
- Verify facts on installed binaries before recording them; label
  docs-confidence claims.
- Never create tags or GitHub Releases; never merge the ax PR; never touch
  `~/.agents`, `~/.claude`, `~/.codex` by hand — only through `curator` or
  `agents-infra` commands, with backups recorded.
- Escalate only for product decisions (repository visibility, Xcode consent,
  retiring versus deprecating the agents-infra launchers) or human-only access.

## Definition of done

`curator run claude_code|codex_cli|pi --profile relux-root-context-ivan`
launches on this machine from Curator-managed homes with MCP and the
model/effort line-group, goldens green in the launcher repo's CI, install
documented; the umbrella profile installed and `env status` current for the
three adapters; the context, skill and MCP repositories tagged only by the
operator; agents-infra reduced to its residual with README updated and the
retired code removed through landed PRs; the three spec-gap proposals filed;
ax untouched.
