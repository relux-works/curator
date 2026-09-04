# Review brief: Decision 0012 cycle 2

`review-brief-0012.md` applies with these updates:
- Head under review: `8444706` (rework of `a25dc67`) on `draft/decision-0012-context-packages`, worktree `~/Developer/ReluxWorks/.worktrees/curator-spec-decision-0012`.
- Inputs: `review-findings-0012-1.md` (1 blocking, 12 major, 8 minor, 2 nit), `producer-brief-0012-rework-1.md` (the orchestrator's author decisions per finding), `rework-report-0012-1.md` (producer dispositions).

Primary: verify each of the 24 findings is genuinely resolved in the document text and that the resolution matches the author decision in the producer brief (deviations must be recorded in the rework report with rationale — judge each). Attack the new text specifically:
1. **F2 algorithm** — walk it on a small graph with a re-selection (a later constraint excluding a chosen version) and on a `||` disjunction; confirm termination argument and the final-check rule are stated and correct; confirm "no backtracking across names" cannot produce a lock that violates a requirement.
2. **F1 impact table** — recheck every environments §1–§13 row and the manager/schemas/vectors rows against the landed revision 1 text; any section marked "unchanged" that the new model actually touches is a finding.
3. **F8 grammar** — recheck coercions and 0.x caret/tilde against node-semver README; confirm hyphen ranges excluded, `v` excluded in ranges, `-0` upper bound, `latest` labeled a Curator spelling.
4. **F9/F10 MCP** — allowlist over MCP package source identities; bare `command`; https-only `url` grammar; args/url under `context-secret-material`; `env_names` grammar + reserved-name exclusion + lockable allowlist; fragment `mcp` carries the env_names union and curator-run adds them to the plan allowlist.
5. **F13 worked example** — internally consistent: lock members (kind,name order, pins, weights, chain, overlay flag), header lines under default policy in emitted order matching Decision 4/8 rules, fragment with `lock_sha256` + `mcp`, one materialized MCP file's bytes.
6. **F11/F12** — "revision 1" consistent everywhere; no "Decision 0011" citation; numbering note present.
7. Regression sweep of untouched sections; house style; § citations.

Verdict: `review-findings-0012-2.md` on TASK-260902-3cnbwa; blocking/major → development; else ACCEPT explicit + accept_cr on the current CR revision, leave to-review. Read-only; do not push or mark done.
