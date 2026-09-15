# Pi MCP scope — explicit resolution (2026-09-15)

Context: the original DoD (goal-launcher-and-infra-migration.md) asks that `curator run` launch all three environments "with MCP". The authoritative contracts contradict that for Pi: environments.md 1.1 §7.8 declares no MCP channel for Pi 0.84.2 (no file, no fragment `mcp` section), Curator's `internal/envregistry` records `MCP: nil` for `pi`, launcher SPEC §4.1 repeats the boundary, and the accepted native-Pi design (TASK-260908-ggxfte rev2, landed as skill-agents-management v0.5.11) adds native launch plans only. The epic resource `pi-mcp-scope-decision.md` recorded the conflict and asked for an explicit choice between (a) narrowing the adapter requirement and (b) adding a separately designed Pi MCP delivery.

Decision (orchestrator, under the unified delivery goal's instruction "Resolve the Pi MCP capability/requirement mismatch explicitly ... do not silently waive either", and its boundary that generalized MCP endpoint/plugin imports remain deferred): **option (a) — the adapter requirement is narrowed.**

- `curator run claude_code` and `curator run codex_cli` MUST deliver the profile's MCP declarations through their declared launch channels (claude `--mcp-config <managed-home file> --strict-mcp-config`; codex `-p curator-mcp` layer), from the managed home only, under the machine allowlist.
- `curator run pi` launches native Pi from its managed home with context and defaults; it carries **no MCP**, and the launcher MUST NOT synthesize a Pi MCP flag, file or fragment section. `env status`/diagnostics must state that Pi has no MCP channel rather than reporting MCP as delivered.
- The DoD sentence "launches ... with MCP" is read as "with the MCP capability each adapter declares in environments 1.1 §7.8"; for Pi that capability is none.
- A Pi MCP channel remains a separate future protocol + adapter scope (environments.md revision, envregistry, launcher SPEC), not part of this delivery. It is filed with the deferred items, not silently dropped.

Rationale: implementing (b) would require inventing an unconsumed channel or a bridge no installed Pi consumes, which the goal forbids; (a) matches every landed contract and keeps assurance claims truthful.

Operator override: this decision stands unless the operator explicitly selects (b); an override reopens STORY-260908-1gywcb's MCP acceptance row for Pi only.
