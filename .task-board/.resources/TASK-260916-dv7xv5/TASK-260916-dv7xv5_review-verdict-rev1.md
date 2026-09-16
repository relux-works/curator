# Review verdict \u2014 TASK-260916-dv7xv5 \u2014 rev1

Final verdict: **rework / changes_requested**. Route to **analysis** (research corrections). No remediation work is requested.

Reviewed verify-e-findings.md against curator 80483355 and launcher b34e1e27 using git show with numbered lines. Re-ran all four quoted grep commands against temporary git-archive snapshots of those exact revisions (no checkout changes). This is static evidence, not dynamic reproduction. No tests executed or passing-suite evidence accepted; tests-green remains unverified, per the read-only brief.

## Per-finding checks

| Finding | Check | Evidence / required correction |
|---|---|---|
| E1 | Disagree with evidence strength; scoped concern remains plausible | profile.go:81 belongs to import, not update. Correct entry point is cmd/curator/profile.go:252, calling UpdateWithPolicy at :292 and reporting updated/unchanged at :301-305. pkgversion.go:286-289 supports latest. The claimed repository-wide empty signer grep is false, even excluding tests: internal/registry/registry.go:181,245 validates signature envelopes; internal/buildrepo/pipeline.go:33,177 mentions signer policy; internal/swiftpmsource/git.go:514 contains verify-commit. These do not establish tag-signature admission for context resolution, but prohibit the claim \u201cno verification anywhere.\u201d Trace and scope the context resolver/source path and record actual search output. |
| E2 | Agree with static verdict; disagree with empty-search evidence | internal/contextmaterialize/contextmaterialize.go:248-265 includes every emitted member; production call is internal/envprofile/managed.go:1785. internal/contextaudit/contextaudit.go:109-115 blocks Findings, not SystemModules. Quoted broad grep returns many matches, including internal/closure/closure.go:1 and internal/envprofile/gitsource.go:32. Replace absence assertion with scoped evidence. |
| E3 | Agree with seed behavior; disagree with empty-search evidence | internal/envregistry/envregistry.go:219 and internal/envprofile/managed.go:570-584 copy native config.toml bytes. Quoted mcp_servers grep DOES return internal/mcp/mcp.go:253 (and mcp_test.go:37). Explain this other path instead of claiming no matches. Runtime behavior is not dynamically reproduced. |
| E4 | Agree | cmd/curator/umbrella.go:30-38 resolves ambient PATH; :43-63 excludes selected user-bin/environments directories; :81-91 executes the result. No arbitrary project-directory ownership gate in this path. |
| E5 | Agree within stated bound | internal/envprofile/switch.go:523-530 removes before copied root write; :539-545 repeats removal on fallback; :694-696 removes before symlink. This mitigates the ordinary pre-existing foreign-manager link case, not race-free O_NOFOLLOW semantics or all managed/store writes (writeStoreDocument still uses WriteFile at :686). Preserve this limitation explicitly. |
| E6 | Agree with narrowed static split | internal/contextpkg/contextpkg.go:289-305 requires git dependency declarations; internal/envprofile/managed.go:163-170 loads MCP-kind members only. system class accepted at contextpkg.go:263; non-git marker identity empty at switch.go:760-777, path provenance at :793-797. Retain partially confirmed; do not imply git dependencies of a path root bypass MCP source admission (contextresolve.go:477-483). Directory-boundary absence must be bounded to inspected path-ingestion flow, not inferred solely from marker fields. |
| E7 | Agree with checked subclaims; incomplete residual coverage | Launcher internal/axconfig/config.go:58-74 and internal/defaults/defaults.go:108-118 follow links; defaults.go:90 reads bytes. Exact permission grep returns only internal/execution/execution.go:197. fragment/resolve.go:75-80 unconditionally adds --repair. defaults/lineup.go:271-290 prints model/effort. Missing: explicit verdict/evidence for the third Minor item, strict-MCP asymmetry. Curator envregistry.go:192 declares --strict-mcp-config; :215 declares Codex -p curator-mcp; :219 and managed.go:576-583 establish the seed half. Provider-path logging is ancillary, not a substitute for this residual. Store-tampering persistence is conditional on S5, not dynamically proved here. |

## Acceptance criteria

1. Outcome exists with 7/7 rows, both pins and file:line citations: structurally verified, checklist item 1 checked. Evidence accuracy needs the corrections above; this check does not accept those claims.
2. Sibling descriptions: FAIL. Read all seven current descriptions via authoritative task-board get queries (the CLI projection of README descriptions). None records the pinned implementation verification/consequences. E1 ioemse, E2 2d9coh, E3 1i1gfo, E4 2otjbn and E7 33vuzm retain specification findings only. E5 73a5zg mentions a historical local fix but does not record the verified main verdict or conformance/atomicity scope. E6 wgt8vz still leaves MCP admissibility unresolved rather than recording not-applicable for path MCP dependencies. Update all seven to record the final reviewed verdict, pins, evidence reference, and narrowed scope. Checklist item 2 remains unchecked.
3. Read-only discipline: verified current worktree and launcher status clean; curator control checkout changes are board metadata/resources only. No product code/test modifications made in this review. Item 3 checked. This is a present-tree observation, not historical proof of every producer operation.
4. Implementation matches AC: unchecked because criterion 2 fails and evidence needs correction.
5. Architecture: no product architecture changed; research uses the relevant existing resolver/materializer/launcher boundaries. Checked.
6. Tests green: unchecked/unverified; tests intentionally not run under brief.
7. Rework evidence and routing: this artifact plus analysis transition fulfills the explicit branch.

## Required producer rework

- Correct E1 command citation and bound signer claims to the context/profile path; correct E1/E2/E3 false zero-match assertions with exact reproducible commands and actual outputs, excluding tests explicitly if intended.
- Complete E7 with an explicit strict-MCP residual verdict and citations; preserve static/conditional limits for repair persistence and E5.
- Update seven sibling story descriptions to reflect verified implementation findings, especially E5 and E6.
- Resubmit corrected research for another independent review. Do not implement remediation or run tests merely to satisfy a generic checklist.

## Logbook \u2014 2026-09-16

Important review discovery: all three curator \u201creturns nothing\u201d searches produce matches at 80483355; the launcher permission search reproduces correctly. Missing sibling updates are a concrete AC failure, not an external blocker. Findings and disposition are persisted here and in task-scoped logbook outcome. Run goal queried: no active goal (run not goal-bound).
