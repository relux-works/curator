# TASK-260916-33abdk — review verdict rev3 (delta) — ACCEPTED

Candidate: CR rev3, base eca2bf27, tree 84fdbc5b (working tree == tree, `git diff --stat 84fdbc5b` empty). Rev2 accepted on content (TASK-260916-33abdk_review-verdict-rev2.md).

## Checks
- Byte-identical to refs/campaign/1i1gfo-rev2-20260927 (blob OIDs): envmarker.go, managed.go, codex_seed_test.go, envregistry.go, root-artifacts.tsv.
- `git diff <merge-tree(55b94af2→rev2 onto eca2bf27)> 84fdbc5b`: only conformance-gaps.tsv (merge-tree conflict text) and the two test files. No production code beyond rev2; envstatus.go/status.go/case-counts equal the merge-tree.
- conformance-gaps.tsv vs trunk eca2bf27: exactly two `-` lines (agent-environment-marker-v1 valid-codex-seed-record{,-empty-snapshot}.json, owner STORY-260916-1i1gfo); zero `+` lines; trunk's STORY-260916-ioemse rows kept. Same two rows rev2 removed.
- No CHANGELOG/LOGBOOK change.

## New tests (rc.13 §7.4 seed strip/report, §8.2 marker schema-1 stability)
- TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome (env_credential_marker_test.go:196): pre-rule schema-1 marker without codex_seed_record + seed carrying [mcp_servers.legacy]; `env resolve --repair` (production entry) keeps marker and config bytes (metadata-only repair must not rewrite) and `env status` reports `mcp_seed_unstripped … re-provision`. Correct: no silent pass-through — the retained table is explicitly reported (strip-or-report).
- TestEnvResolveStripsAndReportsInlineNativeCodexMCPTable (envstatus_test.go:70): inline `mcp_servers = {…}` stripped from managed config.toml, other keys kept, warning on resolve and name in env status, no command leak. Correct.

## Runs (zsh, pipefail, -count=1)
- go test ./cmd/curator -run 'TestEnvResolvePreservesPreRule…|TestEnvResolveStripsAndReportsInline…|EnvResolve|Marker|Seed|Mcp|EnvStatus' → ok 346s, exit=0
- go test ./internal/envprofile -run 'Seed|Mcp|MCP|Codex|Status|Guarded' → ok 85s, exit=0 (includes TestManagerOwnedAbsenceReadsAreGuarded)
Hosted gate on rev3: green per orchestrator note (not rerun here).

Verdict: ACCEPTED.
