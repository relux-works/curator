# Delta review — TASK-260918-ryh3kw rev6 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Rev5 ACCEPTED (your review 2). Rev6 re-applies it on trunk 6bd98d49 (E4 provider roots + E3 Codex seed MCP landed): tree 3229dac8, base
6bd98d49, gate green, same 17 paths. Orchestrator line check vs (trunk ∪ rev5): 15 paths have 0 lines in neither side and 0 trunk lines
dropped. Review ONLY:
1. internal/envprofile/managed.go — 3 lines in neither side and 5 trunk (E3) lines dropped, e.g. `switch file.Kind {`: the re-apply brief
   asked E3's new manager-state reads (codex seed config.toml read / seed record) to be routed through internal/stateread. Confirm the
   E3 behaviour is unchanged (absent native config → no seed/empty snapshot; unreadable → environment_seed_unreadable fail-closed; strip +
   mcp_native_servers_not_inherited, and now revision A per trunk if 31gaka landed — it has not; trunk ships B) and that the routing is the
   only change. Run `go test ./internal/envprofile -run 'Seed|Codex|Mcp|ReadFailure|Guarded'` with a real exit code.
2. internal/envregistry/envregistry.go — 2 lines in neither side (e.g. `DiagPassthroughUnreadable = "environment_passthrough_unreadable"`
   realignment): confirm it is gofmt alignment of rev5's own constant only.
accept_cr or changes requested with file:line. No LOGBOOK.md.
