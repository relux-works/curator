# TASK-260910-19w2aj revision 7 independent review

Verdict: CHANGES_REQUESTED. Route to to-dev. One P1 policy propagation regression; no human decision or external blocker.

Candidate tree: 6ddeb407a982ffc28c65045689946d4c5fc9a009. Base: aa46ecd80ad0b83853586454723ea99fb76977a9. Independently compared all 5,638 candidate blobs against working-tree bytes (including symlink targets): zero mismatches. No production/test code modified. Run goal queried at start and before verdict: not goal-bound.

## P1 — draft closure drops the configured acquisition allowlist

Locations: cmd/curator/project_resolve.go:113-125 and internal/closure/resolve.go:163-171. DraftResolveConfig has no AllowedSources input; the Options passed to BuildExpanded omit it. Existing ensureRepo/gateSource therefore sees an empty allowlist, which identity.Allowed explicitly defines as allow-all. Alias acquisition correctly uses cfg.AllowedSources, but legacy named roots and transitive dependencies take the unguarded closure lane.

Contract: skillfile-sources §1 retains existing dependency/trust rules across the full closure; repository-transport §3 requires stable-identity allowlist/trust checks before acquisition. The unchanged legacy install path passes cfg.AllowedSources at internal/install/install.go:396.

Independent production binary reproduction (attached Python fixture and complete log): isolated temp config allows only trusted.test/team; fake Git transport maps https://denied.test/provider.git to a local fixture, records every clone, and needs no real network.

- Schema-1 install --dry-run refuses the denied provider before clone (exit 1).
- Schema-2 alias selection project resolve refuses the same provider before clone (exit 1).
- Schema-2 unchanged legacy root project resolve clones the denied provider and publishes a one-member lock (exit 0).
- Schema-2 local root with that transitive provider clones it and publishes a two-member lock containing repository denied.test/provider (exit 0).
- Positive control explicitly allows denied.test and resolves the transitive dependency successfully (exit 0).

Measured denied-acquisition enforcement among these draft shapes: 1/3 (alias root), with legacy root and transitive acquisition bypassing the gate. This is a concrete production counterexample, not an inference from missing tests. Fixture script exits 0 after asserting those observations.

Required correction: carry the machine AllowedSources policy from the production CLI through DraftResolveConfig into closure.Options for both legacy and transitive acquisitions, preserving existing semantics. Add CLI negative tests with a temp restrictive config and a clone-attempt recorder: denied root/dependency must fail before acquisition and preserve prior lock/bindings/installed state. Include allowed positive controls and a narrowing mutant that propagates policy only for alias roots. Do not supply the omitted configuration directly to the helper under test.

## Revision-7 rework and independent validation

The previous checkout-versus-declared-ref defect is fixed: pinGitAliases resolves/captures the selected commit before baseline expansion, BuildExpanded, and membership recheck. The proving repository remains separate. All commands below ran in zsh; real exits observed:

- go build -o /tmp/curator-review-19w2aj-rev7 ./cmd/curator: exit 0.
- go test -p 1 ./cmd/curator -run 'TestProject(ResolveGitTagDiffersFromHEADThroughCLI|ResolveGitCollectionPinnedToTagThroughCLI|RefreshGitBranchMembershipChangeThroughCLI)$' -count=1 -timeout=90s: exit 0, 43.631s. Three production regressions pass, including complete tag membership and pinned branch consumption.
- go test -p 1 ./internal/closure -run 'Test(ResolveDraftExpandsOverResolvedCommitNotCheckout|ResolveDraftResolvesEachGitAliasOnce|RefreshDraftSecondWriteFailurePreservesLock|RefreshCatchesRuntimeOnlyAndBuildOnly|ResolveDraftRefusals)$' -count=1 -timeout=90s: exit 0, 8.857s.
- go test -p 1 ./cmd/curator -run '^TestProjectResolveGit(RuntimeTamperRefused|MissingMemberRefused)ThroughCLI$' -count=1 -timeout=90s: exit 0, 5.250s.
- python3 /tmp/TASK-260910-19w2aj_review-rev7-allowlist-repro.py: exit 0; each CLI exit and attempted clone recorded in attached log. Initial fixture draft omitted schema_version in its legacy control; corrected before the recorded final run. That initial harness failure is not product evidence.

Accepted attached hosted evidence, not rerun: change-request_rev7-validation.log reports run 35210509374 success, remote gate exit 0. Gate commit 17bf718305f0e0808d4d30b477e8e384c8a64f72 resolves locally to the exact candidate tree. Hosted Ubuntu/macOS/Windows tests, race lanes, lint and configured gates green; rose-air and candidate suite skipped. Producer reports checkout-expansion mutants killed; not independently replayed because reviewer code access is read-only. Independent policy attack demonstrates missing coverage despite the green gate. No exhaustive conformance/platform/mutation claim.

## Scope and integration bounds

Preserve all accepted fixes, especially declared-commit expansion, SkillsRoot, exact-alias/legacy identity recovery, complete Git inventory authentication and refresh rollback. Local real-install/launch marker integration remains the separately owned limitation recorded in rev6 (STORY-260910-1s75e1 / TASK-260910-1xs0pj and TASK-260910-3eu4cy); no claim of complete local install/launch coverage and no request to fabricate legacy Git identities. The new P1 concerns this leaf's own production resolve policy propagation.

Review checklist: exact candidate verified; relevant contracts/prior verdict/results read; production regressions independently rerun; policy gate attacked with denied and allowed controls; finding and evidence attached; explicit changes-requested routing. Logbook recorded as a task-scoped resource instead of editing forbidden LOGBOOK.md.
