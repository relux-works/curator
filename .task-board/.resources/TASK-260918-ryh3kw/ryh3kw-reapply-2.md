# TASK-260918-ryh3kw — re-apply accepted rev5 on trunk 6bd98d49 (THE ONLY CURRENT INSTRUCTION)

Revision 5 was ACCEPTED (review 2). Trunk moved to 6bd98d49 (E4 3oh0u8 provider trust roots, E3 33abdk Codex seed MCP) and the three-way
merge conflicts on 4 paths: .github/ci/root-artifacts.tsv, internal/envprofile/managed.go, internal/envprofile/status.go,
internal/envregistry/envregistry.go. The accepted content is `refs/campaign/1ll22r-rev5-20260927` (parent eca2bf27, tree b82d529f). Your
Story worktree is fresh on trunk.
1. `task-board m 'set_status(TASK-260918-ryh3kw, status=development)'`.
2. `git diff eca2bf27 refs/campaign/1ll22r-rev5-20260927 -- . ':!.task-board' > $TMPDIR/ryh3kw.patch; git apply --3way $TMPDIR/ryh3kw.patch`;
   resolve KEEPING BOTH SIDES: E3's codex_seed_record / CodexSeedRevision / seed-strip reads must ALSO go through internal/stateread where
   they are manager-state reads (the guard test enforces it) — if E3 added a raw os.Stat/ReadFile absence check, route it through the seam
   as part of the combine and say so. Add nothing else.
3. VERIFY `git diff --name-only HEAD -- . ':!.task-board'` = the 17 rev5 paths (plus only a file E3 needs routed, named in the results).
4. Focused runs (real exit codes): `go test ./internal/stateread`, `go test ./internal/envprofile -run
   'ReadFailure|Guarded|Restore|Unmanage|Seed|Codex|Status'`, `go test ./cmd/curator -run 'Unmanage|Status|EnvResolve'`.
5. Append "Revision 6 — re-apply on 6bd98d49"; handoff and WAIT for the gate; hand off only green. No CHANGELOG/LOGBOOK edit.
