# TASK-261012-2gkqm3 — apply-review-findings-cs: revision 2 evidence

Review defect D1 (P1) answered in cips/diagrams/join-handshake.puml: a revision comment and the replay note now align with CIP-0009 D5/D6. Later frame effects rely on durable binding-generation-scoped operation-identity/canonical-argument-digest claims and authenticated admission floors. Connection IDs/output cache cannot authorize retry mutation. Tombstones and pending collection-proof/retry wording are retained. No other repository change was added in this rework; revision-1 changes remain uncommitted.

Updated TASK-261012-2gkqm3_gaps.md adds the omitted F-7 figure, acknowledges the revision-1 sweep error, reports nine changed figure sources across both revisions, and documents named regression F7-later-frame-replay-figure-consistency with narrowing mutant F7-mailbox-only-frame-claims. Those are test specifications, not executed test/mutation evidence, because the brief prohibits local tests and builds.

Verification performed directly:
- git diff --check: exit 0 (whitespace only).
- Required replay-term rg sweep: exit 0 before and after the correction. Scope: 13 tracked PlantUML/DSL figure sources; before the edit 3/3 matching sources manually inspected; after it 2 sources still match and are consistent. The corrected join-handshake note was inspected separately in git diff. No further figure edit was required. Differently worded contradictions, rendering and runtime semantics are outside this search evidence.
- Public hygiene of the two added lines inspected: no personal name, deployment hostname, absolute path, session link, board-resource link, field or refusal code added.

No tests, builds or rendering run locally. No revision-1 result is claimed as a revision-2 pass. Configured CR validation is sanctioned at handoff and must capture its own result. Runtime qualification remains open. LOGBOOK.md untouched. Handoff is for review, not acceptance.

Lifecycle qualification: first revision-2 handoff exited 1 before CR validation because inherited checklist items 7–11 were unchecked. The local “Tests green” item was replaced by the explicit no-local-tests/builds and mandatory configured-CR-validation requirement; the inherited LOGBOOK item was removed because the current brief forbids that file and existing item 6 already requires board notes/outcome. Documentation AC, architecture fit and the attached revision-1 verdict/route were verified and checked. No unrun test was marked green. Retrying handoff must still run the configured CR gate.

Successor recovery: the configured revision-2 suite subsequently exited 1 on a broken local link in the materialized gaps report. No test/build step ran after that validator failure. The successor corrected only the report's source-reference formatting, retained the public CIP/README figure links, reread D1 and its diagram correction, inspected all replay-term matching sources, and reran `git diff --check` directly (exit 0). Named regression/mutant specifications remain unexecuted. No new repository edit or LOGBOOK write. The sanctioned handoff retry must record current validation evidence.
