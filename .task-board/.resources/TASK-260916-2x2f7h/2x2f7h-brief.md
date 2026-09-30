# TASK-260916-2x2f7h — spec notes: repair-as-persistence + MCP channel asymmetry (THE ONLY CURRENT INSTRUCTION)

curator-spec repository. Read `remediation-spec-producer-rules.md`, this task's README and the Story README (STORY-260928-3o6kp9).
1. In the S5 remediation text (environments.md: protected state / env resolve), add the repair-as-persistence residual:
   `env resolve --repair` re-applies store bytes on every launch, so a store entry that passes its checks is re-materialized each time.
   State what that implies for an attacker who can write the store (persistence), and state the bound the protocol accepts. Cite the
   S5 clause you amend.
2. In §7.8, the MCP channel rules shared with E3, add a row describing the asymmetry between the Claude and Codex MCP channels: what
   each runtime enforces and what it leaves to the manager. Keep it normative where the protocol constrains managers, and informative
   otherwise.
3. Change no vectors unless a MUST changes. If one does, add the vector.
4. Run the validators (`make` targets / tools/validate.py) and record the real exit codes. Add a CHANGELOG entry under Unreleased. No
   LOGBOOK.md. Never spell any employer name.

Update the results, then run `task-board handoff TASK-260916-2x2f7h --role developer`, then END YOUR TURN. Write only inside your Story
worktree.
