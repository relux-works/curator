# TASK-260916-33abdk — re-apply accepted rev2 on current trunk (THE ONLY CURRENT INSTRUCTION)

Revision 2 was ACCEPTED on content. Trunk moved to eca2bf27 and `worktree converge` conflicted on .github/ci/conformance-gaps.tsv only.
The exact accepted content is `refs/campaign/1i1gfo-rev2-20260927` (parent 55b94af2, tree a1a14805 == CR rev2). Your Story worktree is fresh.
1. `task-board m 'set_status(TASK-260916-33abdk, status=development)'`.
2. `git diff 55b94af2 refs/campaign/1i1gfo-rev2-20260927 -- . ':!.task-board' > $TMPDIR/33abdk.patch; git apply --3way $TMPDIR/33abdk.patch`;
   resolve conformance-gaps.tsv keeping trunk's rows and removing only the rows rev2 removed (the 2 STORY-260916-1i1gfo rows); add
   NOTHING in neither side. conformance-case-counts.tsv: if trunk changed counts for the same suites, recompute from the pinned suite, and say how.
3. VERIFY `git diff --name-only HEAD -- . ':!.task-board'` lists exactly the 11 rev2 paths; non-conflicting paths byte-identical to rev2.
4. Focused runs (real exit codes): `go test ./internal/envprofile -run 'Seed|Mcp|MCP|Codex|Status|Guarded'`, `go test ./cmd/curator -run
   'EnvResolve|Marker|Seed|Mcp|EnvStatus'`.
5. Append "Revision 3 — re-apply on eca2bf27"; handoff and WAIT for the gate; hand off only green. No CHANGELOG/LOGBOOK edit.
