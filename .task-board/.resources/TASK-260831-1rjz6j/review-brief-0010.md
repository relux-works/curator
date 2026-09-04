# Review brief: Decision 0010 draft (agent environment profiles)

## Subject

Review the draft decision document, not the board metadata:

- Repo checkout (worktree): `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-draft-environments`
- Branch: `draft/agent-environment-profiles`, base `c9ea2ff` (== origin/main)
- Document: `decisions/0010-agent-environment-profiles.md`
- Commits under review: `b6b4ef9` (initial proposal), `3fd5617` (secondary fixed-home targets amendment)
- Supporting research: board resource `research-environment-support-matrix.md` on TASK-260831-hbq9n6 (read via task-board resource read, or accept the in-document matrix as the claim surface)

## Review dimensions (all required)

1. **House style and structure** — compare against existing `decisions/0004-compile-only-build-drivers.md` and `decisions/0009-first-party-module-roots.md` in the same worktree: section set, tone, density, normative-language usage appropriate for a Proposed decision.
2. **Internal consistency with the spec** — every cited section (§ references to `protocol/core.md`, `profiles/manager.md`) must exist and say what the decision claims it says. Check the adapter table against manager §5, MCP table §6, ledger §11, source identity §6.1–6.3, content hash §8, transaction §2.5.
3. **Factual accuracy of environment claims** — the isolation variables and home shapes for claude_code (CLAUDE_CONFIG_DIR), codex_cli (CODEX_HOME), opencode (XDG_CONFIG_HOME/opencode), pi (PI_CODING_AGENT_DIR), and the Xcode CodingAssistant secondary-target paths (ClaudeAgentConfig/, codex/). Where a claim is verifiable on this machine (installed binaries `claude`, `codex`, `pi`, `gemini`; `~/.curator`, `~/.claude`, `~/.codex`, `~/.pi/agent`), verify rather than trust the document. Do not install anything.
4. **Design soundness** — internal contradictions, holes in the three materialization modes, switching semantics, credential passthrough/isolation, launcher fragment boundary (profile data must not influence env injection), ax composition contract (SpawnPlan env_literals exists in agent-session-manager-spec SPEC.md v0.5.0 at ~/Developer/ReluxWorks/agent-session-manager-spec — read-only reference).
5. **Completeness vs story AC** (STORY-260831-6bbhow): profile model + git pins; context IR + materializer contract; symlink/copy/managed-home modes with recommendation; launcher + naming candidates; profile-scoped skills; inventory CLI; ax composition; phased rev-1 scope; open questions with recommendations.
6. **Security and compatibility sections** — do they hold against the decision content; anything claimed that is not delivered, anything delivered that is not covered.

## Constraints

- Read-only review: do NOT edit the decision document, do NOT commit, do NOT push. The producer will apply fixes.
- Use shell tools (`ls`, `git -C`, `grep`, file reads) — verify paths exist before citing them.
- The two reference checkouts are read-only context: `~/Developer/ReluxWorks/curator-spec` (main) and `~/Developer/ReluxWorks/agent-session-manager-spec`.

## Verdict contract

- Produce a findings list: each finding = severity (blocking | major | minor | nit), document section, exact quote or line anchor, what is wrong, suggested fix.
- Attach the full findings report as a board resource on TASK-260831-1rjz6j named `review-findings-<n>.md` (n = attempt number; use the next free number).
- Blocking or major findings present → set task status back to `development` and say so in the handoff.
- Only minor/nit or none → the review passes: leave status `to-review`, state ACCEPT explicitly in the handoff summary and in the findings resource.
- Do not mark anything `done`; the orchestrator owns closure.
