# Owner decisions, 2026-10-10, from the docs-audit questions

## D-GOALS: agent self-goals outside a board, variant (b) for now
- Now: only the board/orchestrator (board goals) or the launching caller at start (a goal passed through the launcher) sets a session's goal. The agent account gets no goal capability on the session host socket.
- The docs must say WHY this is the choice for now, and that the final choice between (a) agent self-goals in a separate namespace and (b) is deferred to a separate SH2+ item.
- Hard requirement (owner): no strict coupling to goal functionality, to our orchestrator or to its toolchain. In the module, task-board is an OPTIONAL component. The session host must work without a board and without goals. Board goals are one optional client of a generic goal interface, not a dependency.
- Targets: swarma-session-host spec §3.8 and README (R-SH1), wiki architecture goal-namespace section (R-WA1), book ch07/08.

## D-IDENTITY: identity revocation at worker retirement
- One-shot workers: the dispatcher revokes the identity itself at cleanup, as a separate operation after the account is retired (dispatcher §14). The dispatcher's right covers only identities it registered as one-shot, and that flag is recorded at registration.
- Standing agents (orchestrator, fixed team roles): retiring the account does not touch the identity. Only the operator revokes it, with a signed record.
- Targets: swarma-dispatcher spec (R-DP1), key-keeper/architecture (R-WA1), swarma-user-manager retire wording (R-UM1), book ch04 and ch13.
