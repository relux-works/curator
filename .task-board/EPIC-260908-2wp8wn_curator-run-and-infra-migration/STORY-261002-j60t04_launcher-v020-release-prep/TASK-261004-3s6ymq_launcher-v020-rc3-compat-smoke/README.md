# TASK-261004-3s6ymq: launcher-v020-rc3-compat-smoke

## Description
Compatibility smoke: released curator v0.15.0-rc.3 (installed from the release tarball into a scratch prefix) x curator-agent-launcher 2517d27 (Prepare 0.2.0), at the real entry points, before tagging launcher v0.2.0.

## Scope
(define task scope)

## Acceptance Criteria
Smoke report with real exit codes in a scratch HOME: fragment resolve v2/v3 for claude_code, codex_cli, pi, muse (dry-run/plan only, no real model turns); permission transport native/yolo incl. locked/headless refusals; colliding prompt/MCP negatives; umbrella discovery of curator-run from a trusted provider dir and refusal of managed shim dirs; curator-run --version = 0.2.0; no real credentials read; verdict GO/NO-GO for the v0.2.0 tag.
