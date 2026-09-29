# Review note — TASK-260927-31gaka ship Codex seed revision A (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev1 (base 86552087 = trunk, tree ed98b5b1, 8 paths, gate green) against `31gaka-brief.md` and curator-spec rc.13 environments.md
§7.4 "Codex seed MCP rule" ("a manager MUST ship revision A before revision B"), §7.7 status, §8.2 codex_seed_record, the diagnostics rows
mcp_native_servers_ungoverned / mcp_seed_unstripped, and the seed vectors. Verify through the production entry:
1. The shipped registry revision is A (one switch constant); provisioning copies config.toml WHOLE (mcp_servers kept), writes
   codex_seed_record {revision "A", names snapshot}, and warns `mcp_native_servers_ungoverned` exactly when the snapshot is non-empty, naming
   every entry with the "next revision stops inheriting" statement and the migration hint (exact text per spec); empty snapshot → no warning,
   record still written.
2. env status lists the recorded names per managed codex_cli home as ungoverned; homes with a B record or no record behave per §7.4.
3. The B implementation from 33abdk is kept behind the switch and still tested; the revision-A vectors that 33abdk bounded are now driven;
   the B-shipped cases are bounded with owner TASK-260927-1e5qqm (gap rows) — not deleted.
4. Mutants (warning removed; strip applied under A) killed with real exit codes; manager-state reads via internal/stateread; marker bytes of
   existing homes unchanged ("an existing home keeps its bytes"). No trunk revert, no CHANGELOG/LOGBOOK, no stray files.
Bounded runs. accept_cr or changes requested with file:line. No LOGBOOK.md.
