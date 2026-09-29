# TASK-260927-31gaka — ship Codex seed revision A first (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md`. curator-spec v1.0.0-rc.13 environments.md §7.4 "Codex seed MCP rule": "a manager MUST ship revision A before
revision B". TASK-260916-33abdk (landed, 6bd98d49) implemented revision B and set internal/envregistry CodexSeedRevision = B; curator has
never released A. Make the SHIPPED revision A, keeping the B implementation in place for the later flip leaf (TASK-260927-1e5qqm):
1. Revision A behaviour (cite §7.4, §7.7, §8.2, the diagnostics table rows): the codex_cli seed copies config.toml WHOLE; provisioning emits
   `mcp_native_servers_ungoverned` (warning) naming every inherited native `mcp_servers` entry, stating the next revision stops inheriting
   them, with the migration hint (declare the server in the profile's MCP set, or accept the loss); `codex_seed_record` {revision "A",
   snapshot of names}; `env status` lists them per managed codex_cli home as ungoverned. Warning fires exactly when the snapshot is non-empty.
2. One switch: the registry constant selects A or B; set it to A. Tests for BOTH revisions stay (B rows driven through the constant/seam, not
   deleted); the rc.13 revision-A provisioning/posture vectors that 33abdk bounded are now DRIVEN; the B-shipped cases become the bounded set
   attributed to TASK-260927-1e5qqm (gap rows with that owner).
3. Mutants: warning removed; whole copy replaced by strip under A — killed (real exit codes). New manager-state reads via internal/stateread.
4. No CHANGELOG/LOGBOOK edit (entry text in results under "## CHANGELOG entry (for release prep)"). Handoff and WAIT for the hosted gate;
   hand off only green. Write only inside your Story worktree.
