# Review note — TASK-261002-1pif8m release readiness (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review the readiness checklist for truthfulness. It drives the next releases. Spot-check with real exit codes:
1. The 7 spec commits since v1.0.0-rc.13, with zero modified pre-existing schemas and 9 added: `git log v1.0.0-rc.13..origin/main`, and a diff of schemas.
2. **The rc.13 release record claim.** Main regenerates `release/1.0.0-rc.13.json` with the candidate digest bd03456…, while the tag pins be11bb1e…. Check whether main's rc.13 record bytes differ from the tagged ones. This is important: if true, published history was modified.
3. **The launcher v0.1.1 tag.** It exists and is signed, targets 1ac7eaf, equals main, and buildVersion is still 0.1.0.
4. `internal/hashing/hashing.go:47` is the single v2-writer switch and is OFF.
5. Codex seed A and posture A landed after rc.2 (`git tag --contains`).
6. The board states of the leaves in the open-leaves table.
7. The CR adds only `.research/` files. No LOGBOOK.md; never spell any employer name.

accept_cr, or changes requested with exact corrections.
