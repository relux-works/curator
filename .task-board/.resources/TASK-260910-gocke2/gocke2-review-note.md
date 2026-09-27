# Review note — TASK-260910-gocke2 MCP surfacing (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review against `gocke2-sec-brief.md` and the task description (surface the resolved MCP set — command, args, env_names — at profile
install/update; warn when the allowlist is empty or env_names intersect operator-secret-looking variables; spec rc.13, cite clauses). The
candidate changes ONLY internal/envprofile/surfacing_test.go: establish which production code on main already implements each rule
(file:line) and that the new rows drive it through the production entry; every rule has a row; kill at least two mutants yourself (drop the
empty-allowlist warning / drop the secret-name intersection warning). If a rule is NOT implemented on main, that is a finding. Gate green;
gap rows for this surface removed. accept_cr or changes requested with file:line. No LOGBOOK.md.
