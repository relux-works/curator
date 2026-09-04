# Review brief: ax integration delta

Subject: worktree `/Users/iv/Developer/ReluxWorks/.temp/ax-curator-integration/worktree`, branch `draft/curator-environment-integration`, head `d7075e1`, base `28bf96d` (= ax main v0.5.0). Producer notes on TASK-260901-14atb7. Read-only; do not push.

Checks:
1. **Additivity**: prose-only, no heading/numbering changes, no normative invariant of ax touched; OPTIONAL-integration framing everywhere; ax-without-Curator behavior byte-identical.
2. **Why is `scripts/validate_spec.py` in the delta (±1 line)?** Determine exactly what changed and whether it is a legitimate necessity (e.g. a link/section checker needing the new reference) or scope creep. Anything beyond the minimal necessity is a finding.
3. **Extensions conformance**: run the repo's own validator (`scripts/validate_spec.py` or the repo's documented gate) at head; the `works.relux.curator.*` key grammar must satisfy the SPEC's §1.6/§5.1 extensions rules exactly (reverse-DNS, opacity, no core-semantics influence); pin-shape claim (commit XOR state hash) must match curator-spec environments.md §10 fragment reality — verify against the landed main of curator-spec.
4. **Resume/fork drift paragraph**: placed where resume semantics live; warn-and-continue default + strict refuse consistent with Decision 0010 D10; fragment-digest re-resolution story coherent (digest over exact consumed JSON bytes — is that well-defined given resolve output? flag ambiguity if not).
5. **Citations**: repo/doc/§ references to curator-spec exact; no vendored copies; `curator session` note informative-only.
6. Signed commit; delta = SPEC.md + the justified script line only.

Verdict: `review-findings-ax-delta-1.md` on TASK-260901-14atb7; blocking/major -> development; else ACCEPT + accept_cr. The PR itself is opened by the orchestrator afterwards and stays open for the ax maintainer — landing is not in scope.
