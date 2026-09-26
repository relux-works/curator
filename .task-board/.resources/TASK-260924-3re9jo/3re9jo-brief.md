# TASK-260924-3re9jo — Skillfile sources: clause-by-clause gap matrix (THE ONLY CURRENT INSTRUCTION; research, read-only)

Operator priority: HIGHEST. Source of truth is the specification — curator-spec main (local checkout
/Users/administrator/Developer/ReluxWorks/curator/curator-spec, `git fetch` then read `origin/main`):
`protocol/skillfile-sources.md` (skillfile-sources-v1, Skillfile schema 2), `protocol/repository-transport.md` (revisions 1
and 2), `schemas/draft-sources-v1/*` and every conformance case/vector that exercises them (search `conformance/` for
skillfile-v2, skillfile-lock-v1, source-policy, install-marker-v5, build-receipt-v3, source-audit, repository-transport).
Implementation under test: Curator `origin/main` (read-only; do not edit anything).

Background you must verify, not trust: EPIC-260910-ohqchs implemented the extension as an opt-in "draft" surface
(`cmd/curator/draft_*.go`, `internal/crossconformance/draftsources_semantic_test.go`), and past reviews recorded residuals:
known-gap xfail rows (insteadOf/security, alias loader-only, registry exact match), marker v5 local go-v1 arm nulls and
declared_tag grammar, revocation breadth, raw-snapshot vs lock context hash, repository/commit-only rows helper-level only,
scp+alias-port ssh://~/ rendering, checkout-origin persistence. Search `.task-board/.resources/` review verdicts for more.

Deliver `TASK-260924-3re9jo_results.md`:
1. Matrix: every normative MUST / MUST NOT / SHOULD and every conformance case → status
   {implemented+driven | implemented-not-driven | known-gap(owner) | missing | draft-gated} with spec line and Curator
   file:line or test name.
2. How Curator gates schema 2 today (opt-in flag? draft commands? refusal?) and exactly what the spec requires of a reader that
   "explicitly supports the extension" — i.e. what first-class support means per the spec.
3. Whether the spec text itself must change to make the extension released (it currently says "unreleased working
   specification … opt-in"), with the precise edits — a separate curator-spec leaf, not Curator code.
4. The implementation leaf list, directly creatable: title, scope, AC, dependency order, which can run in parallel (they will be
   run in parallel), and which existing board elements they supersede.
Do not name any other product in task titles; cite the spec. Attach results, check the DoD item, then
`task-board handoff TASK-260924-3re9jo --role researcher`. A `run_wrote_outside_worktree … policy warn` block is a warning.
