# TASK-261018-nd18ex: probe-user-runtime-discovery-paths

Result: both Codex0.162.0 and0.163.0 retain HOME/.agents/skills and HOME/.agents/plugins/marketplace.json discovery when only CODEX_HOME is overridden. Observed with valid sentinels, rendered prompt content, plugin availability and file-read controls. Claude Code2.1.302 follows CLAUDE_CONFIG_DIR for the tested skills, instructions, settings/hooks and registered local plugins.

Canonical report: [.research/261018-user-runtime-path-probe/README.md](../../../.research/261018-user-runtime-path-probe/README.md).
Reproducer: [probe.py](../../../.research/261018-user-runtime-path-probe/probe.py).
Evidence, controls, binary provenance and exact-source comparison JSON are adjacent to that report.

Validation: final normal matrix11/11 passed, exit0. Removing first selected skill causes the intended first-lane assertion failure, runs all ten remaining lanes green and exits1. Missing first binary reports four cannot_start lanes, then runs all four second-runtime lanes green and exits1. Sandbox write/network controls and atime controls passed. No model calls or host configuration mutation. Current macOS arm64 observations are not a cross-platform or security-isolation qualification. Codex hook execution/trust and alternate plugin marketplace layouts remain source-only or unqualified as stated in the report. Initial app-server configuration setup was inconclusive; final positive controls use supported local CLI entry points.

Execution: inline orch execution permitted by swarm keeper for this read-only probe; independent review = Ivan-lane PR review; no tracked producer/reviewer child. Self-review checks scope, host-path redaction and observed/source-only boundaries. Signed PR only; do not self-land. Refresh and other host-operation holds remain active.

State: ready for upstream PR review, not accepted or landed. No claim that BUG-261018-k2tn36 is fixed by this research-only delta.
