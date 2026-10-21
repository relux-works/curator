# Rework brief r3: audit Markdown scope per subject kind (BUG-261021-2qv4zj, curator)

Review r2 (`BUG-261021-2qv4zj_review-r2.md`, board outcome resource) confirmed that F1 (case aliases) is fixed and requested changes for a new P1, F2. Read its F2 section and reproduction in full. Continue on branch `fix/audit-markdown-scope`.

## F2: context members are not covered

The shared skill `WalkContext` is applied to every audit subject, unconditionally. Context members (`KindContext`, e.g. a profile whose `agent-context.json` declares `context.modules: [{path: a.md}]`) publish their declared modules through `internal/contextmaterialize` into the native `CLAUDE.md` / `AGENTS.md`. Those files are outside the skill roots, so the walk never reaches them. A strict profile install exits 0, places the undeclared URL into both native documents, and caches an empty ruleset-7 verdict. The base detector refused it.

The spec requires the unchanged strict source audit for every closure member: environments §9.1 at the pinned spec.

## Required invariant (generalises r2)

**For every subject kind, the audited Markdown set is a superset of what that kind can publish to an agent.** The selection is driven by the subject's actual publication contract:
- **Skills:** keep the r2 `WalkContext`, which is shared with `CopyContext`.
- **Context members:** use the declared modules, resolved by the same code `contextmaterialize` uses to read them, not a copy of its path logic.
- **Every other kind the audit Gate accepts:** enumerate them all from the code. For each, state what it publishes and how the audit covers it. Where a kind publishes nothing to agent context, say so and keep its previous behaviour. If unsure, fall back to the base behaviour (scan everything), which is fail-closed.

Write the per-kind table in the results file.

## Cache

Bump the ruleset to 8, so that ruleset-7 allowances created by the r2 head (for example the F2 fixture's `findings: []`) are not reused. Prove it with a test seeded from a real ruleset-7 record.

## Tests (each test states its claim)

- **F2 regression:** reproduce the r2 profile-install fixture exactly. The URL module refuses with `audit.capability.network-undeclared`, and nothing is published: no `source.json`, no `lock.json`, no URL in native `CLAUDE.md`/`AGENTS.md`. The clean module and a declared host still install.
- **Per-kind coverage:** for each other subject kind, a positive and a negative case.
- **Unchanged:** all r1/r2 regressions stay green (nine aliases, exclusion parity, `.research` noise, CIP-0021 witness).
- **Mutants:** a mutant that drops the context-module branch fails the F2 regression, and the r2 mutants still fail.

## Run and hand-off

Build, vet, focused packages, pinned conformance, and the full default suite in a normal-.git clone with a private GOCACHE deleted afterwards. Attach `BUG-261021-2qv4zj_results-r3.md` as an outcome resource, push signed commits, hand over in `audit-md-scope`, and delete task scratch.
