# Brief — CLI aliases claude / codex (implementation)

Spec: curator-spec main 2d6d497, profiles/manager.md CLI section ("environment operand accepts claude and codex as aliases of claude_code and codex_cli, normalized before validation; outputs/markers/fragments/config/locks keep the canonical id; aliases never persisted"). Wire ids unchanged: do NOT touch envregistry ids, markers, fragment schemas, defaults.json keys, or the spec.

Curator (TASK-260916-11lwua, control root curator): add one normalization helper (e.g. internal/envregistry.NormalizeEnvID or in cmd/curator) applied to every CLI environment operand: `curator env resolve|status|unmanage`, `curator run <env>` umbrella dispatch, `curator profile use --env`, config knob values given on the command line that name an environment. Print canonical ids in all output; refuse unknown ids as today. Tests: both spellings through the real CLI entry points (cmd/curator tests), and a test that persisted config/markers never contain the alias. Help text lists the aliases.

Launcher (the sibling task, control root curator-agent-launcher): `curator-run <env-id>` accepts the aliases, normalizes before the environment lookup / `curator env resolve` call; provenance lines print canonical ids; README SPEC + help; goldens for both spellings.

Narrow tests only (host stalls on big suites); the remote gate runs at handoff. Evidence resource with exit codes; tick the checklist; handoff.
