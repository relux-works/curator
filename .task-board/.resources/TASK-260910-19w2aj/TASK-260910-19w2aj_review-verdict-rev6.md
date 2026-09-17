# TASK-260910-19w2aj revision 6 independent review

Verdict: CHANGES_REQUESTED. Route to to-dev; ordinary implementation rework, no human decision required.

Candidate tree: 52902724a34d7710bc8a3944680300aa7ff32fd7. Base: 62ea2d2ced3fac7c3c5f7a2ff4e887e0f92f4540. Independently compared all 5,539 candidate blobs with worktree bytes (including symlink targets): zero mismatches. No production or test code modified. Goal queried at start and before verdict: run is not goal-bound.

## P1 — Git selection and collection membership come from the checkout instead of the declared ref

cmd/curator/project_resolve.go:165 passes the acquired working checkout as the expansion root. internal/closure/resolve.go:111 expands it before acquireGit resolves the declared tag/branch/revision (resolve.go:591). Thus selection names, validation and collection membership come from default-branch HEAD on initial clone, or the old checkout on subsequent fetches, while the lock commit and materialized bytes come from the separately resolved ref. These must describe the same immutable tree.

Two independent production CLI reproductions with isolated config, local bare Git fixture and fake URL transport:

1. Tag v1 contains a valid review package; later default-branch HEAD removes SKILL.md. Skillfile explicitly selects v1. `project resolve app` exits 1 with source_member_invalid referring to home/draft-git/.../SKILL.md. The valid selected tag is rejected based on unrelated HEAD bytes.
2. Tag v1 contains collection members alpha and beta; later HEAD removes beta. Skillfile selects tag v1, directory skills, include ["*"]. `project resolve app` exits 0 and publishes a lock with only alpha. Measured membership: 1/2 expected members. This silently omits a valid selected package from the deterministic locked plan.

Attached review-rev6-git-ref-repro.py/.log and review-rev6-collection-repro.py/.log contain full commands and exit codes. Both reproduction scripts exit 0 after asserting the incorrect behavior. No external network needed.

Required correction: resolve each Git alias once and authenticate/capture that exact commit before any selection or collection expansion. Run baseline expansion, closure and membership recheck over that same frozen tree; retain the proving repository separately for Git identity/authentication. Do not repair this by adopting checkout HEAD or supplying preselected roots only in tests. Add CLI regressions for individual and collection selections whose declared tag/revision differs from default HEAD, plus a branch refresh whose newly fetched commit changes membership while the checkout is unchanged. Assert complete expected membership and pinned consumption, and kill a narrowing mutant that uses the checkout only for collection expansion. Preserve all previous fixes.

## Revision-6 rework and verification

The two rev5 legacy identity defects are fixed in the reviewed code. Independent zsh commands:

- `go build -o /tmp/TASK-260910-19w2aj-review6-curator ./cmd/curator`: exit 0.
- `go test -p 1 ./cmd/curator -run '^TestProjectResolveLegacy(ConfiguredGit|NetworkGit)ThroughCLI$' -count=1 -timeout=90s`: exit 0, 78.583s. Both forms and both default/custom source locations pass (4/4 subcases).
- `go test -p 1 ./internal/closure ./internal/install -run 'Draft|Refresh|LegacyInstallUntouched' -count=1 -timeout=90s`: exit 0; closure 42.014s, install 55.294s.
- Broader `go test -p 1 ./cmd/curator -run 'TestProjectResolve' -count=1 -timeout 90s`: exit 1, timeout at 90.515s while TestProjectResolveLegacyMixedThroughCLI was running a fixture Git command. Not claimed passing; narrowed rerun above verifies the rework. This timeout is not the rejection reason.
- Three independent Python CLI reproductions: exit 0; constituent CLI results recorded in attached logs.

Accepted attached hosted evidence, not rerun: TASK-260910-19w2aj_change-request_rev6-validation.log reports run 35202909339 success, remote gate exit 0. Gate commit c86175fd2d44f1445af56686d282b22960b44927 resolves locally to the exact reviewed candidate tree. Ubuntu/macOS/Windows tests, race lanes, lint and other configured gates green; rose-air and candidate suite skipped. Producer reports two killed rev6 mutants; not independently replayed. Actual production counterexamples establish the missing coverage without changing code. No exhaustive conformance, platform or mutation claim.

## Integration limitation, separately owned

A local source resolves and dry-runs successfully but real install exits 1: `review: install marker is invalid for schema 2`. Reproduced with the production binary; attached review-rev6-repro.py/.log. install.go:386 discards the returned lock and the existing marker builder at install.go:1023 still requires Git ref identity. The new mixed CLI test explicitly stops at dry-run (draft_sources_test.go:739-744). Do not claim local real-install or launch coverage from it.

Board scope identifies marker migration and atomic source installation as STORY-260910-1s75e1 / TASK-260910-1xs0pj and TASK-260910-3eu4cy. Record this as an integration limitation for those owners, not an instruction to fabricate legacy Git refs or expand this leaf into marker migration. The Git membership P1 alone requires rework here. No acceptance of complete local install/launch behavior is implied.

Findings are persisted in the task-scoped review logbook resource instead of editing forbidden LOGBOOK.md. Preserve the accepted lock-model decisions, transport and capture/store work, SkillsRoot and exact-alias recovery, complete Git inventory authentication, legacy identity recovery and rollback checks.
