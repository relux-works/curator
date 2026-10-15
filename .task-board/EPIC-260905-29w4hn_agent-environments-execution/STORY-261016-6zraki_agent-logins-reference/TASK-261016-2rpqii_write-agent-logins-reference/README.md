# TASK-261016-2rpqii: write-agent-logins-reference

## Description
Write docs/agent-logins.md: how Claude Code, Codex CLI and Muse store logins (measured behaviour with versions), Curator isolation and the credential_store machine knob (rc.7), env login and profile prepare (rc.7), the setup-token source (planned), recipes per platform, with PlantUML diagrams; link it prominently from README and docs/environment-config.md.

## Scope
(define task scope)

## Acceptance Criteria
1. docs/agent-logins.md exists and covers the three tools, isolation, credential_store with exact CLI spelling and default, login routes, and a status column (works today / rc.7 / planned) that matches the code on main. 2. Diagrams in diagrams/plantuml with rendered output following the repository convention. 3. README has a visible pointer; environment-config.md links to it. 4. No host, person, account, seat or incident names; facts carry tool versions. 5. Markdown link check passes.
