# THE ONLY CURRENT INSTRUCTION — TASK-261004-3s6ymq: rc.3 × launcher 0.2.0 compatibility smoke (researcher, validation)

Goal: a GO/NO-GO verdict for tagging curator-agent-launcher v0.2.0 at 2517d27 ("Prepare curator-agent-launcher 0.2.0") against the released curator v0.15.0-rc.3.

Setup, all in a scratch HOME and a scratch prefix; never touch the operator's homes or credentials:
- Download curator v0.15.0-rc.3 darwin_amd64 from the GitHub release. Verify it against checksums.txt and install it into the scratch prefix.
- Build curator-run from the launcher worktree at 2517d27 with `go build -work`.
- Check that `curator-run --version` reports 0.2.0.

Rows, each with its real exit code:
1. Fragment resolution through the launcher's real path: `curator env resolve --repair --format json` for claude_code, codex_cli and pi (v2), and for muse (v3, four XDG parents, HOME preserved). Use plan, dry-run or `--help`-level launches only, with no real model turns. Where a native tool is absent in the scratch env, record that as bounded rather than failed.
2. Permission transport: native and yolo, locked and headless refusals, and tracked-yolo fail-closed. Exercise the real curator fragment.
3. Negatives: a native arg that collides with prompt/MCP/profile flags is refused; a wrong fragment version or digest is refused.
4. Umbrella discovery: `curator run` finds curator-run in a trusted provider directory and refuses managed shim locations.
5. Check README and CHANGELOG at 2517d27: the install line says @v0.2.0, and the supported-environments sentence includes muse.

Host rules: GOFLAGS=-work; keep commands short; on a hang longer than 5 minutes, wait. Never edit LOGBOOK.md. Attach the report as outcome `TASK-261004-3s6ymq_smoke.md`. Then `task-board handoff TASK-261004-3s6ymq --role researcher` and END YOUR TURN.
