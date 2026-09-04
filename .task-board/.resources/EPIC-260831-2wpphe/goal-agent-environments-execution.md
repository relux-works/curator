# Goal: execute the agent-environments capability (Curator + launcher + ax)

## Objective

Turn the accepted specifications into a shipped, conformance-tested
capability: Curator becomes a full agent environment manager — versioned
context packages resolved into semver-locked profiles, materialized into
managed and in-place agent homes, launched through `curator run` and, when
configured, tracked by `ax`. Work spec-first: close the remaining
specification gaps in the recorded order, then implement in stages, each
stage gated by its own conformance subset and a producer/reviewer cycle.

## Sources of truth (read before acting; never restate divergently)

- `relux-works/curator-spec` main `b4f29cd`:
  `decisions/0010-agent-environment-profiles.md` (profiles, adapters, modes,
  onboarding, launcher boundary), `decisions/0012-context-packages-and-semver-locks.md`
  (context/MCP packages, semver lock, weights, overlays, launch-channel
  MCP — supersedes 0010's repository shape), `protocol/environments.md`
  revision 1 (normative surface to be rewritten in place per the 0012
  impact table), `schemas/v1/*environment*|*fragment*|*profilefile*|*context-manifest*`,
  `conformance/v1/expected/environments/*`, `profiles/manager.md` §12,
  `cli/curator.md`.
- Board resource `pre-implementation-review-v3.md` on STORY-260901-zddtn8
  (curator board `~/Developer/ReluxWorks/curator/.task-board`): 16 MUST
  items (M1–M16), 14 NEXT, LATER — all still binding, re-targeted onto the
  0012 model.
- `relux-works/curator-agent-launcher` main `6de42d8` (SPEC 0.1.2-draft).
- `relux-works/agent-session-manager-spec` PR #1 (open, `d7075e1`) and
  main v0.5.0; `relux-works/skill-agents-management` (`pkg/agentic`,
  `BuildPlan`/`BuildLaunch`).
- `relux-works/curator` main `66e34a23` (reference implementation;
  `internal/transaction`, `marker`, `adapters`, `scopes`, `audit`,
  `gitops`, `globalbins`, `config`).

## Settled operator decisions (do not reopen)

Execution ownership = Option A: `curator-run` is the single composer;
`ax start --launch-plan FILE|-` accepts a closed `{argv|argv_suffix,
env_literals, extensions}`; `SpawnPlan` gains `stdin`. Semver only where a
lock exists; core stays frozen. Precedence = `winner` × `placement`.
Weights: manifest → direct-requirer edge (must agree) → root `weights`.
Overlays from git or `path`. MCP in the first shipped revision as a launch
channel, managed homes only, allowlist over package source identities.
Strict SemVer 2.0 `v`-tags without build metadata; npm-shaped ranges plus
`latest`. environments.md stays revision 1; only the header type line bumps
to `curator-root-context-v2`. csk cleanup is surface naming only.

## Ordered work plan

1. **Execution-ownership decision** (review M1/M2/M7/M8): record Option A
   under the next free decision number (0011 is taken by the swift-driver
   draft — reconcile); specify the `ax start --launch-plan` operation and
   the `stdin` SpawnPlan member; add `LaunchModeInteractive` to
   agents-management (per-system argv with model/effort only — no print
   mode, no permission bypass); launcher SPEC 0.2: fragment-before-plan
   ordering with `LaunchRequest.Home` = managed home, default model/effort
   ownership, tracked mode delegates; revise ax PR #1 (operation section,
   refuse-on-drift when the chain carries `class: system` modules, CCJ-1
   fragment digest). Deliver the ax change as a revision of PR #1 and leave
   landing to the ax maintainer.
2. **Snapshot acquisition byte-exactness** (review M3), its own PR first:
   normative rule (core §6.2 or environments §1) that snapshots reproduce
   exact committed blob bytes — no autocrlf, no `export-subst`; vector with
   `* text=auto` + `export-subst`; reference implementation moves from
   `git archive` to object-database extraction. Fixes the shipped skills
   pipeline too.
3. **Decision 0010 erratum**: pi flag "verified in 0.84.2" claim, the
   superseded sequencing sentence, the credentials claim — a recorded
   amendment, never a silent edit.
4. **environments.md revision 1.1 batch on the 0012 model** (one
   producer/reviewer cycle, then schemas/vectors/manager/CLI): the 0012
   impact table section by section; review M4 provisioning seeds +
   passthrough strategies + `environment_passthrough_detached` +
   `environment_isolated_unsupported` (claude_code/macOS, opencode); M5
   channel `argument` + verified spellings + size advisory; M6 command
   roots + fragment `path_prepend` + umbrella provider hardening; M9
   unpinnable detector + scoped waiver + `context-system-module-present`;
   M10 read-only resolve + explicit `--repair` + lock classes; M11
   update/remove/unmanage/switch-failure/versioned backups; M12 config
   schema 2 + CLI rows; M13 shadow acknowledgment; M14 XDG seed
   reconciliation; M15 lockability or explicit deferral; M16; N1–N14.
5. **Verification sprint** before any adapter freezes: Keychain
   `oauth.claude.profile.*` keying; `.claude.json` seed shape; codex global
   `AGENTS.md` cap; codex/pi `auth.json` write mode; fresh-home first run
   per tool with seeds; Xcode embedded agents honoring root context;
   opencode `XDG_CONFIG_HOME` on Windows; claude referenced-form approval;
   codex `-p curator-mcp` layer composition. Record as board evidence;
   adjust the registry; only then freeze vectors.
6. **Implementation, staged** (curator, Go), each stage a conformance
   subset and a producer/reviewer cycle: (a) acquisition fix, package
   store, `git`/`local` kinds, context/MCP manifests, semver resolution +
   lock, monolithic form, `linked` switching with the M11 transactional
   shape, global-scope migration; (b) managed homes with seeds and
   passthrough strategies, read-only resolve, fragment with `path_prepend`
   and `mcp`, untracked `curator-run` with interactive plans; (c)
   composition/weights/overlays, `path` kind and onboarding import, config
   schema 2 + CLI; (d) ax integration once the execution-ownership
   decision and the ax operation land.

## Operating rules

- Board first: every change tracked on the curator board (epic
  EPIC-260831-2wpphe or its successor); stories with worktrees from fresh
  main; producer/reviewer cycles on `claude-fable-5-1`; briefs as
  precondition resources; findings and reports as outcome resources.
- Delivery canon: branch → PR → comment-review verdict with evidence →
  green checks → fast-forward push of the exact reviewed head; signed
  commits by the configured human identity; never rewrite reviewed commits;
  rebase with `-S` and prove identity by `git range-diff` when main
  advances.
- Never create tags or GitHub Releases without the operator's explicit
  release command. Never merge the ax proposal PR — the ax maintainer lands
  it.
- Verify facts on the installed binaries and sources before recording them
  in a normative table; label docs-confidence claims as such.
- Escalate to the operator only for genuine product/architecture decisions
  or human-only access; otherwise finish the stage.

## Definition of done

The execution-ownership decision, the acquisition fix, the 0010 erratum,
and environments.md revision 1.1 with schemas, vectors, manager, and CLI
are landed on curator-spec main; the launcher SPEC 0.2 is landed; the ax
proposal PR is revised and open; the verification-sprint evidence is
recorded; implementation stages (a) through (c) pass their conformance
subsets in curator main; stage (d) is landed once the ax operation exists.
