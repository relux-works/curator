# TASK-260910-gocke2 review verdict — CR rev2 — ACCEPTED

Reviewer: claude-opus-5-5 (low). Candidate tree 7da18bfc (base 0ffe2e1d); worktree diff vs candidate is empty. Delta: internal/envprofile/surfacing_test.go only (+71/-1).

## Production code on main per rule (spec rc.13 environments §2.2/§2.3/§10.3/§12)
- Empty-allowlist warning `mcp_package_allowlist_empty` (§2.2): internal/envprofile/surfacing.go:16 AllowlistEmptyWarning; emitted at install envprofile.go:772, update envprofile.go:890/919/1160/1180, env status status.go:312.
- Env-name passthrough warning (s4-warn `mcp_env_passthrough_unlisted` + migration hint; s4-enforce `mcp_env_passthrough_dropped`) (§10.3): internal/envfragment/envfragment.go:226-313 ResolvePassthrough, ActiveS4Profile = s4-warn; production call site internal/envprofile/managed.go:1945 (once per resolution). The spec's form of the "secret-looking intersection" warning is this unlisted/dropped rule; no separate heuristic is specified.
- §2.3 surfacing rows: contextmaterialize/mcp.go:146-163 FormatDeclarationRows; envprofile/surfacing.go surfacingRows/emitSurfacing at install/update before publication.
- Posture: TestStatusS4Posture; vectors: TestEnvironmentsEnvPassthroughVectors (root-artifacts.tsv:73; case counts .github/ci/conformance-case-counts.tsv:68-72). There are no conformance-gaps.tsv rows for this surface, so nothing is left to remove.

## Candidate rows
TestAllowlistWarningOperations now asserts the warning is PRESENT with an empty policy through Install, UpdateWithPolicy and StatusOf (production entries). Before this, only the absent path was pinned. TestEmptyAllowlistWarningLeavesCurrentStatusCurrent proves the §12 rule "warnings never make a row non-current".

## Independent reruns (zsh, pipefail, CURATOR_CONFORMANCE_ROOT=curator-spec checkout)
go vet ./internal/envprofile/ ok; gofmt -l clean; go test ./internal/envprofile/ ./internal/envfragment/ -run 'Allowlist|Surfacing|Passthrough|EnvPassthrough|S4' → ok (120s) / ok, 17 tests PASS incl. vectors.

## Mutants (applied, then reverted)
- M1 AllowlistEmptyWarning always "" → KILLED by the new rows (surfacing_test.go:238). It would survive the base rows, which only asserted absence.
- M2 drop the s4-warn unlisted warning → KILLED in envfragment (TestResolvePassthroughProfiles) and at the production entry (surfacing_test.go:301, via Resolve).

Hosted gate: the runtime ran it at handoff; I relied on that and did not rerun the full suite locally.
Residual: the test fixtures cover the "operator-secret-looking" aspect only as the spec's generic unlisted-name rule.
