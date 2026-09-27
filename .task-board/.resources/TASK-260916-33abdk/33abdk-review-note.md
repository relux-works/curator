# Review note — TASK-260916-33abdk E3 Codex seeded MCP tables (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev2 (base 55b94af2, tree a1a14805, 11 paths, gate green) against `33abdk-sec-brief.md`, `33abdk-gatefix-1.md` and curator-spec
v1.0.0-rc.13 (cite clauses). Verify through the production entry: undeclared `[mcp_servers.*]` tables in a Codex seed are stripped at
provisioning and reported (sorted names) in env status — no silent pass-through into the launched environment; declared MCP servers (gocke2
surfacing, on trunk) still flow; the marker field `codex_seed_record` is written only when there is something to record, so a metadata-only
resolve keeps schema-1 marker bytes (TestEnvResolveKeepsSchema1BytesForMetadataOnly) — check the rc.13 marker schema allows the field at all
(v5 closed shapes!). The producer drove 9/15 seed vectors with "six explicit Revision A bounds": judge whether each bound is a legitimate
separate surface with an owner (attributed gap rows) or work this leaf should do. Mutants (strip removed / report removed / inline-table
form) killed with real exit codes. No trunk revert, no CHANGELOG/LOGBOOK, no stray files; stateread guard passes. Bounded runs. accept_cr
or changes requested with file:line. No LOGBOOK.md.
