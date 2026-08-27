# TASK-260823-1vleh5: schema8-manifest-parsing-module-roots

## Description
curator (Go): implement schema-8 manifest parsing for the modules declaration on go-v1 build commands per the landed candidate prose (candidate/schema-8-rc.9 at 6001dc3: protocol/core.md section 4.2 module-roots rules, agent-skill-v8.schema.json): parse the modules list, validate portability/disjointness against build and runtime roots, and the bijection against go.mod replace directives — directory form only, strictly inside the snapshot, no versions, no module-to-module redirects; closed diagnostics per the spec. Include the execution_policy field admission on script commands (schema-8 parse surface is one bump) with declared-only behavior for now if the worker containment is out of this task scope — parsing must not reject valid schema-8 manifests. Unit tests against the schema-cases of the candidate suite. Worktree from origin/main. Maintainer pre-authorization 2026-08-22: fully autonomous; land via PR with every lane verified green pre-merge. Executor: claude only.

## Scope
(define task scope)

## Acceptance Criteria
Schema-8 manifests parse and validate per the candidate rules; schema-case families for agent-skill-v8 consumed by tests; merged to main green.
