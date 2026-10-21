# Brief: qualify project-native source exclusion for Codex CLI 0.162.1 (research harness and record)

**Repository:** curator. **Branch:** `research/project-source-exclusion-codex`, from `origin/main`.

## Why

CIP-0002 ("project context in managed launches", read it in curator-spec `cips/CIP-0002-project-context-in-managed-launches.md`) needs a certified way to start a managed session in a project checkout so that **nothing the repository ships configures the agent unless it was approved**. That covers instructions, rules, skills, hooks, plugins, MCP servers, permissions and settings. Approved copies must still load. Before any adapter is written we need evidence of what Codex CLI 0.162.1 actually discovers from a working directory, and which controls exclude it. The owner's first wave is Claude, Codex and Muse.

## Method (mirror the existing qualification harnesses)

Use `.research/muse1410/`, `docs/qualification/muse-1.4.10-R6427.1.md` and `.research/codex162/` as your templates:
- an exact binary with its digest pinned, through an opt-in env gate (`CODEX_EXCL_BINARY`);
- `env -i`, disposable HOME and XDG under task-local `.temp/`, a sandbox that denies network and the real native directories;
- no real credentials, no auth commands, no remote inference: use a loopback or echo provider, or synthetic credentials, as `.research/codex162` does;
- default `go test` skips everything.

Binary: `~/.codex/packages/standalone/releases/0.162.1-aarch64-apple-darwin/bin/codex` (exact release path, digest pinned; never the updating launcher).

## Rows to measure

For each row, build a synthetic project checkout that plants the source, and observe whether the runtime loads it. Use loader or debug output, the prompt-input dump or real tool behaviour, never the model's own claim. Then find the control that excludes it (flag, config, env, home layout) and show that it does.

- P1 root instructions (AGENTS.md / CLAUDE.md and the like, at the repository root and in ancestors);
- P2 nested and lazy instructions (subdirectory files loaded on entry);
- P3 project skills;
- P4 project hooks and plugins;
- P5 project MCP config: empty and non-empty;
- P6 project permission and settings files;
- P7 resume and child agents: does an excluded source come back?
- P8 late-created files during a session.

For each row, also show that an **approved** copy delivered through the managed home still loads (the positive control). Record a per-row verdict (excludable with control X / not excludable / unmeasured, with the reason) and the exact flags or config.

## Deliver

- `.research/project-exclusion-codex/`: the harness, fixtures and README.
- `docs/qualification/project-source-exclusion-codex.md`: the record.
- A narrowing control: a harness variant that drops one exclusion control must fail the matching row.
- Tests that state their claims. `go build` and `go vet` pass, the default suite stays green, the gated run is recorded with exit codes.
- No product code changes. Public-repo hygiene: no personal paths in committed files.
