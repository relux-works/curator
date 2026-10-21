# Rework brief r4: publish the core §3.1 projection for profile skills (BUG-261021-2qv4zj, curator)

Review r3 (`BUG-261021-2qv4zj_review-r3.md`, plus two probe sources as outcome resources) accepted F1 and F2 and requested changes for a P1, F3. Read the whole review and run both probes first. Continue on branch `fix/audit-markdown-scope`.

## F3

Skill Markdown is audited through `WalkContext`, the core §3.1 projection. But profile skill publication exposes the **raw selected package root**:
- `internal/envprofile/managed.go:279-290` assigns the raw root;
- `switch.go:744` uses it for native skill links;
- `managed.go:3065` uses it for managed-home links;
- the copy fallback copies the raw tree.

So `.research/notes.md`, `references/README.md` and `assets/tests/a.md` are admitted unaudited and remain readable under `skills/<name>/`.

## Decision (orchestrator, spec-based): fix publication, keep the narrowed audit

core §3.1 is normative:
- only the listed root entries are eligible for agent-facing context;
- the exclusion patterns apply at every depth;
- "Build roots MUST NOT be copied into the commit-keyed runtime store or any installed context";
- generated artifacts stay outside agent-facing context.

A raw-root link is an installed context that contains excluded files and build roots, so publication is what is non-conformant. Do not widen the audit back. That would re-block task-board install, which is the original bug.

Required:
1. **Inventory every publisher of skill bytes to an agent surface:** profile native links, managed-home links, the copy fallback, status, repair, surface identity/markers, and any legacy global-skill or project install path. Post it in the `audit-md-scope` topic as a short table (publisher, current source, new source) before the large edits. Continue without waiting unless I object.
2. **Publish the §3.1 projection.** Each publisher links or copies a projection produced by the same `WalkContext`/`CopyContext` the audit uses, never the raw root. Use one projected store directory per package identity (content-addressed or derived from the existing store entry). Link and copy modes must yield byte-identical trees.
3. **Runtime and commands.** Script commands and `runtime_roots` must still resolve from the runtime store exactly as today. Check that no command or runtime path depended on the raw-root link, and say so with evidence.
4. **Status and repair.** A home that still links the raw root is reported as stale and is repaired to the projection, with no takeover prompt for Curator's own links. This matters because the rc.8 host swap migrates ivmbp homes. Surface identity must change, so old markers don't claim currency.
5. **Cache.** Keep ruleset 8, unless publication identity enters the cache key. Explain your choice either way.

## Tests (each states its claim)

- Both r3 probes pass, the native and the managed one. Excluded repository files, fixed excluded names and directories, runtime/build roots and exported-script Markdown are absent from `skills/<name>/` in link and copy modes, for Claude and Codex.
- Visible and declared controls are still published and readable.
- Repair moves a raw-root home to the projection.
- All F1/F2/r1-r3 regressions stay green.
- Mutant: publishing the raw root again fails the projection tests.
- task-board-shaped package (a repo-root `SKILL.md`, `.research/` with URLs, a vendor build root): it installs, and none of those files appear on any surface.

## Run and hand-off

Run build, vet, focused tests, pinned conformance and the full default suite, with a private GOCACHE deleted afterwards. Attach `BUG-261021-2qv4zj_results-r4.md` (with the publisher table and the changelog line), push signed, hand over in `audit-md-scope`, and delete scratch. Write notes, test names and the report in neutral QA terms (coverage, publication, refusal). Reviewer runs on this task have been stopped by the provider's content filter.
