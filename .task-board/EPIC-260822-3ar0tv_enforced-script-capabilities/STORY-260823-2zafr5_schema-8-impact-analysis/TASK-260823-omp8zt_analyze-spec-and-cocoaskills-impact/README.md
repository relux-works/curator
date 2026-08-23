# TASK-260823-omp8zt: analyze-spec-and-cocoaskills-impact

## Description
Research, no code changes. Answer three questions with evidence: (1) SEQUENCING — curator-spec CI runs pinned implementations against the conformance suite (Implementations jobs; see Pin landed agent-skill implementations #12, and the candidate-suite mechanism: a new-schema suite enters only through the candidate-conformance workflow_dispatch and is never a default) — determine whether landing schema-8 vectors on main breaks the pinned-implementation jobs, and what the correct landing order is: vector staging, candidate suite, pin bumps, SPEC_PIN in curator CI, release/rc versioning, COMPATIBILITY.md. (2) COCOASKILLS — enumerate the concrete changes it needs for schema 8: manifest parsing of execution_policy on script commands, script-worker-v1 containment obligations, module-roots bijection validation (its go_v1.py:980 currently raises vendor_metadata_inconsistent on any Module.Replace), audit labeling; give per-item scope estimates against the existing stories (cocoaskills board STORY-260822-2evh3p and STORY-260822-27ze8z). (3) CURATOR-SPEC RESIDUALS — anything the in-flight tasks do not cover (registry/audit record surface, profiles, release process). Deliverable: add_resource impact-analysis.md on this task. Maintainer pre-authorization 2026-08-22: fully autonomous, no human approval gates.

## Scope
(define task scope)

## Acceptance Criteria
impact-analysis.md resource on this task answers all three questions with evidence and concrete change lists; the landing task and implementation stories can cite it directly.
