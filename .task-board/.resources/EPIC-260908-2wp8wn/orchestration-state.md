# Launcher and infrastructure migration checkpoint — 2026-09-08

The attached full goal and mapping remain the scope. Operator policy since September 8 until further notice: no hosted CI; run necessary tests locally. Preserve signed commits, actual PR review and exact-head fast-forward landing. Do not create tags/releases or touch the real ax integration. The operator creates required tags.

## Delivered

- Task-board separate code/board ownership: PR187 merged in skill-project-management at signed416693dd. Installed binary/skill now0.24.3-317-g416693dd. Canonical setup installed successfully but final self-verification failed because it references a missing isolated config; source repair is tracked below. Existing daemon sessions were preserved.
- Launcher parser only: PR4 merged at signed25379f2. Full curator-run composition and installation are not delivered.
- Four accepted Curator Stories closed through actual worktree complete with signed scoped board commits in Curator main: A0 STORY-260908-3mcpz5 ->4f980b1a; CLI STORY-260908-xadoax ->22c0ce85; environment design STORY-260908-37tde1 ->a728e495; Pi design STORY-260908-3p1kfp ->324dc16a. Their task outcomes record commands, signatures and remote containment. CLI completion revalidated the exact landed tree with make check. Other closures had empty repository deltas and invented no code/test artifacts.

## Active implementation

- BUG-260908-2ovkto / STORY-260908-ps7x2v on the skill-project-management board, RUN-260908-fa9cf4: setup verification and safe reconciliation of an accepted delta after externally reviewed signed rebase, including the actual two-feature-commit case. Original source Story STORY-260908-3r282p remains integrating until supported reconciliation closes it.
- TASK-260908-2kapmh / STORY-260908-3lmnfs, RUN-260908-02aa71: native Pi in skill-agents-management, following accepted design TASK-260908-ggxfte rev2. Parent owns reviewed signed landing; operator tag follows it.
- TASK-260908-ranc5y / STORY-260908-15st55, RUN-260908-d1701c: launcher fragment subprocess, strict closed parser and CCJ-1 digest; A0 captured fragments/digests attached.
- Prepared next: TASK-260908-2zn8fu under STORY-260908-6gwnfc, curator-spec evidence corrections for Plan.Env/owned literals, Pi prompt precedence and argv_suffix. No work has started in that task yet.

## Roots and preservation

Authoritative product board: /Users/iv/Developer/ReluxWorks/curator/.task-board. Runtime/control worktrees under /Users/iv/Developer/ReluxWorks/.worktrees: launcher-control, agents-management-control, curator-spec-control and taskboard-cross-repo-control. Task-scoped config/briefs/observation logs in curator-agent-launcher/.temp/launcher-migration and .temp/cross-repo-delivery. Never locate a RUN from a different control root.

Curator duplicate origin-https config was disabled only after proving the same canonical URL as origin. All tracking refs and objects were preserved and exact configuration backed up in .temp/launcher-migration/curator-duplicate-remote-backup.json. Unrelated dirty board/LOGBOOK state in Curator and the original task-board source checkout remains preserved.

## Remaining scope

Finish task-board rollout and source Story closures; all remaining launcher sections and real managed-home launches; context/skill/MCP packages and umbrella, defaults, takeover with backup; agents-infra residual removal through its board/PR flow; three requested curator-spec gap proposals. No migration package/onboarding/residual work is claimed delivered.

## Pending operator intent

The full DoD says MCP for all three environments, while environments1.1 section7.8 and the implementation explicitly have no Pi MCP channel. The attached pi-mcp-scope-decision.md records this conflict. The operator question is pending: qualify MCP to Claude/Codex with native Pi without MCP, or add Pi MCP design/delivery scope. Unaffected work continues; no option is assumed from elapsed time.
