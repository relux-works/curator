# TASK-260908-1bpra2 resumed producer evidence

The preserved candidate was inspected without changing repository code. This run
read the epic B3 goal, current task brief, Decision 0012 D6 and environments 2.2.
The task brief supersedes the old LLDB and home-install requirements.

All commands below were rerun directly in zsh without pipelines; no prior test
result is substituted for these observations. Curator checkout remains
f750344f5e6fe3ef539485981c12e24240cca1ed.

| Command | Actual exit | Observation |
| --- | --- | --- |
| scripts/validate.sh | 0 | 2/2 manifests valid |
| python3 -m unittest discover -s tests -v | 0 | 6 tests; 3/3 narrowing mutants killed |
| sh -n bin/safaridriver-mcp scripts/validate.sh | 0 | POSIX syntax |
| git diff --check | 0 | whitespace clean |
| go run -mod=mod . ../../packages/figma/agent-mcp.json ../../packages/safari/agent-mcp.json | 0 | both ACCEPT |
| go run -mod=mod . absolute.json | 1 | expected bare-command refusal |
| go run -mod=mod . token.json | 0 | structural parser accepts literal argument text |
| go run -mod=mod . unknown.json | 1 | expected unknown-field refusal |

Go commands ran in .temp/oracle. Output:

```text
../../packages/figma/agent-mcp.json: ACCEPT
../../packages/safari/agent-mcp.json: ACCEPT
absolute.json: REJECT mcp_declaration_invalid: server.command must be a bare executable name without a path separator
exit status 1
token.json: ACCEPT
unknown.json: REJECT mcp_declaration_invalid: unknown field server.token
exit status 1
```

| Narrowing attack | Original gate exit | Narrowed gate exit |
| --- | --- | --- |
| absolute command | 1 | 0 |
| literal token argument | 1 | 0 |
| unknown server field | 1 | 0 |

The release gate covers the two pinned declarations, not arbitrary secret
recognition. Original producer results contain package fields and precise guard
mutations. Wrapper tests use a sandbox path substitution; real Safari execution
and other platforms remain unverified. No full landing suite run manually.
No tags, commits, installs, home changes or LOGBOOK edits performed by this run.
Independent review and signed integration remain downstream responsibilities.
