# Review note — TASK-260916-2x2f7h spec notes (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev1 (curator-spec, base 4ad8042b, tree bd777646; CHANGELOG.md + protocol/environments.md; validators green, 641 tests OK) against
`2x2f7h-brief.md`, the task README and the Story README (STORY-260928-3o6kp9). Verify:
1. The repair-as-persistence residual is added to the S5 protected-state / env resolve text. It must be accurate: `env resolve --repair`
   re-applies store bytes on every launch, and it must state what an attacker who can write the store gains and which bound the protocol
   accepts. It must cite the amended clause and must not weaken any existing MUST.
2. The Claude/Codex MCP channel asymmetry row is in §7.8, shared with E3, and in the right place (check that §7.8 is in
   environments.md; if the brief's location is wrong, say where it belongs). It must say what each runtime enforces and what the
   manager must do; normative language only where the protocol constrains managers.
3. No vector changes were needed, meaning no MUST changed. Confirm, or name the missing vector.
4. The CHANGELOG entry is under Unreleased. No LOGBOOK.md. No stray files.
accept_cr, or changes requested with file:line. Never spell any employer name.
