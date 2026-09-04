# Producer brief: ax integration — the single detailed PR content

## Setup
- Repo `~/Developer/ReluxWorks/agent-session-manager-spec` (origin relux-works, main 28bf96d). Create worktree `~/Developer/ReluxWorks/.temp/ax-curator-integration/worktree`, branch `draft/curator-environment-integration` from origin/main. Signed commits. Do not push (orchestrator opens the PR).

## Task — minimal, precise, additive edits to SPEC.md (v0.5.0)
Deliver the curator-integration contract Decision 0010 D10 and environments.md §10/§11 promise, as the smallest coherent additive delta (this is a proposal PR into an actively evolving spec — do not restructure anything):
1. Session Record extensions: recommend (SHOULD, optional capability) reverse-DNS extension keys recording at launch: profile name, profile effective pin (commit or state hash), and launch-env-fragment digest. Follow the spec's exact extensions conventions (find the extensions field rules in §5.1/§1.6 and match key grammar; pick a stable prefix like works.relux.curator.* consistent with any existing convention you find).
2. Resume fidelity: on resume/fork, re-resolve the same profile via `curator env resolve <env> --profile <p> --format json`; a resolved pin differing from the recorded one is drift — default warn-and-continue, strict flag refuses. Place where resume semantics live; keep it an OPTIONAL integration (ax without Curator unchanged).
3. SpawnPlan consumption: one paragraph noting fragment env vars merge into `env_literals` (values are non-secret paths, within existing limits), fragment names are a closed set from the curator-spec `launch-env-fragment-v1` surface (cite `relux-works/curator-spec` `protocol/environments.md` §10 by name, no vendored copy).
4. `curator-session` shim: a short informative note that the Curator umbrella may expose ax as `curator session` via PATH discovery; zero normative impact on ax.
Every edit marked with the spec's own conventions for optional integrations; no version bump (maintainers decide); no changes to normative invariants, matrices, or fixtures unless an existing fixture list demands a row — if it does, add the minimal row and say so in notes.

## Deliverables
Signed commit(s); board resource `ax-integration-notes.md` (exact sections touched, key grammar chosen and why, anything you deliberately did NOT touch); handoff to-review.
