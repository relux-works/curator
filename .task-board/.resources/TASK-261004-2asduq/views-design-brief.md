# THE ONLY CURRENT INSTRUCTION — TASK-261004-2asduq: project × profile launch views design options (researcher)

DESIGN PENDING. Scenario: work on project X with profile Y via curator run; X's admitted skills and commands are usable; nothing is sourced from the repository. Design input is attached (launch-roots-design-input.md). The operator added: a project can also bring its own MCP servers, root context and permissions — cover them. Note: the launcher passes --strict-mcp-config to Claude, so a repository .mcp.json is ignored today; establish what Claude and Codex natively read from the project directory (AGENTS.md/CLAUDE.md, .mcp.json, project config) and what a managed launch keeps or suppresses.
Compare: per project×profile homes vs a composed view in the profile home vs a launch-time overlay; how project MCP/context/permissions are admitted (manager-admitted declarations, never raw repo files); the command dispatcher + PATH append (fragment v4); leases; login cost per new home on macOS (see TASK-261004-34brhn).
## Operator decision (2026-10-04, binding)
Project-level context MUST be supportable when needed, under manager control: project MCP servers, project AGENTS.md / CLAUDE.md and their per-environment equivalents, per-environment rules (e.g. Claude rules/memory files, Cursor-style rules, Codex project config), knowledge files, permissions, skills and commands. Think this through properly: an inventory of what each supported environment (claude_code, codex_cli, opencode, pi, muse) natively reads from a project directory; which of it a managed launch keeps, suppresses (e.g. --strict-mcp-config) or must re-admit; an admission model (manager-admitted declarations with digests and operator approval, never raw repository bytes executed or trusted implicitly); per-project opt-in/opt-out and per-item control; precedence between profile and project layers; and the security model for a hostile repository.

## Output format (operator decision 2026-10-04)
Write the result as a **Curator Improvement Proposal draft** using the attached `cip-template.md`, saved as `.research/261004_CIP-NNNN-<slug>.md` (NNNN given below). CIPs will live in curator-spec `cips/`; the orchestrator publishes the draft there after review. Put raw evidence in a companion `.research/261004_CIP-NNNN-<slug>_evidence.md`.

CIP number: CIP-0002 (slug: project-context-in-managed-launches).

## Research rules (binding)
Read-only research: no product code changes. Output ONE document under .research/ (named <YYMMDD>_<slug>.md) plus evidence, attached as task outcome. Cite file:line on curator main, the curator-spec rc.14 text, or a measured probe in a scratch HOME. Never read, print or copy any real credential, token or Keychain secret; never log in or out of the operator account. Never edit LOGBOOK.md. Board ids always with their titles. The public board is visible to anyone: no internal hostnames, personal paths or employer names.
End with: options (2-4), tradeoffs, a recommendation, open questions for the operator, and the spec/implementation leaves the recommendation implies.

Then `task-board handoff TASK-261004-2asduq --role researcher` and END YOUR TURN.
