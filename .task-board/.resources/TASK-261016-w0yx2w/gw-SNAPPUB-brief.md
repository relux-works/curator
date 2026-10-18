# Brief SNAPPUB: rc.7 code wave, SNAP-PUBLISH — Retained-handle destination publication (TASK-261016-w0yx2w)

Implement the attached plan leaf file(s) exactly: `rc7-leaf-snapshot-destination-publication.md` 
Read rc7-code-common.md and rc7-decisions.md (attached) first.

- The leaf is the contract: its normative clauses, acceptance criteria, mandatory paired tests, plant controls, dependencies and paths. Read the cited normative sections on curator-spec main yourself; where the leaf and the normative text disagree, the normative text wins: stop and report.
- Stay inside the leaf's attack surface and paths. If the leaf is too big for one reviewable change, implement its first coherent part, say exactly what remains, and do not fake the rest.
- No runtime qualification is claimed unless you executed it. Record skipped or unavailable rows honestly.
- Leave the change uncommitted for the Change Request flow.

## Revision after the owner decision (2026-10-17)

Your blocked question is decided: option 1. curator-spec main 9dfc557 (TASK-261017-uyuj08) changed managed-home-snapshots §2 and CIP-0014: the destination is either **absent** under an existing caller-owned private parent, or an **intact prior snapshot of this version** (verified no-op). An existing empty directory refuses. Read the new §2 and the CIP-0014 rows on curator-spec main yourself; they win over the older leaf text, and the leaf now carries the same decision at its end.

Your Story workspace was converged onto current curator main (fcd785b) with your retained admission candidate carried (11 paths). Your old `UNRESOLVED_QUESTIONS.md` note was removed because the question is answered; a copy is kept by the orchestrator.

Now finish the leaf:
- Admission: an existing empty directory now refuses (change the empty-target branch of `OpenDestination` and its tests); absent with a valid private parent and intact prior stay admitted.
- Publication: stage privately beside the absent target, write the captured bytes with the required modes, fsync as needed, then publish with a no-replace rename on the retained parent (macOS `renamex_np(..., RENAME_EXCL)` / `renameatx_np`, Linux `renameat2(..., RENAME_NOREPLACE)`); a competing creation makes publication refuse, never overwrite. Unqualified platforms (including Windows) keep refusing explicitly.
- Exact-prior no-op through `MatchPrior` stays; a different generation refuses and preserves the target.
- Failure and interruption: no success is emitted; cleanup removes only this invocation's staging objects by retained handle; a substituted or linked staging entry is never followed or removed; no success manifest can name partial output.
- Tests: the leaf's paired table with the decision applied (an existing empty directory is now a negative row), real syscall race tests for the publish step (a competing destination created between validation and publication), staging-failure and cleanup-substitution tests, and narrowing mutants for no-replace publication and staging-only cleanup. Linux paths: run them in the remote gate (Linux CI runs on the self-hosted runners).
- No CLI integration in this leaf (SNAP-CLI is separate); say so in the handoff.
Run the full remote gate and hand off.
