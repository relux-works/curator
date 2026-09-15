# Unresolved product decisions

Date: 2026-09-10. Task: TASK-260910-nys3f2.
All recommendations below are discussion proposals, not accepted requirements.

| Decision | Recommendation | Consequence |
| --- | --- | --- |
| Is ordinary local installation a snapshot or a live link? | Snapshot at explicit install/sync/update, with content-hash identity. | Reproducible installed state; local edits require an explicit refresh. A later link mode has weaker guarantees. |
| How broad is collection discovery? | Immediate child skill directories, all or explicit subset, plus exclusions. | Simple P1 and compatible Agent Plugins directory shape; nested arbitrary collections can be added explicitly later. |
| When does the project lock arrive? | With source/collection work; fix exact member/hash recording first. | Ranges can wait, but membership must not silently change during launch. Exact filename, schema and private-overlay rules need a decision. |
| Are rules and knowledge separate package ecosystems? | Distinct component types within the existing context-package/resolver machinery. | Visible user semantics without duplicated dependency resolution. |
| What is the portable activation floor? | Always-on rules and indexed on-demand knowledge first; file activation only where preserved or explicitly reported unsupported. | No promise that every host implements every trigger. |
| Does managed project context require curator-run? | Decide explicitly; support sync-generated project files for direct agent launches when portable. | A launch-only channel is insufficient for users who launch agents directly. |
| What happens to existing tracked AGENTS.md/CLAUDE.md? | Explicit future ownership migration; public source can stay committed. | Prevents automatic overwrite/untracking and duplicate ingestion. |
| Should generated combined instructions be committed? | Opt-in generated uncommitted mode; private inputs/effective lock stay local. | Requires defined generation, cleanup and concurrency behavior. |
| Does MCP mean profile launch, project config, or native user-home edits? | Existing managed launch first, project wiring as a real additional surface; native-home edits separately owned. | Keeps read-only dependency checks distinct from installation. |
| How broad is initial plugin support? | Agent Plugins skills and representable MCP metadata; report unsupported required content. | Does not claim a universal plugin runtime or enable arbitrary extensions/hooks. |
| Which CLI binaries are meant? | Curator bootstrap already downloads binaries; scope new work to skill-exported CLI artifacts if that is the intended request. | Different repos/contracts and different provenance obligations. |
| Is source fallback allowed? | Explicit policy after missing compatible artifact; never after an integrity failure. | Predictable installation behavior and no hidden toolchain requirement. |

Detailed evidence, recommended manifest fragments and the specification impact
matrix are in research-and-proposal.md in this task's resources.
