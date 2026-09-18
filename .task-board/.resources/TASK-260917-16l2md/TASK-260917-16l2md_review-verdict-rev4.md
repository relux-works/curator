# TASK-260917-16l2md — revision 4 review verdict

Verdict: **changes_requested**. Route to **to-dev**; do not accept CR-TASK-260917-16l2md-4.

## Required correction

**Medium — remove the accidentally captured root executable `curator`.** Revision 4 adds a 18,798,096-byte executable, mode 100755, blob `39e528eb7c8ca06ed87d9c263cd3ed3469b540b7`. It is absent from revision 3 and the base. `file` identifies it as Mach-O x86_64; `go version -m` identifies GOOS=darwin, GOARCH=amd64, vcs.revision=a4a3bcee268b39edffc2f84de37bc9b869c1dd57, vcs.modified=true. This is a local build output, outside the requested source/spec-pin union, not a portable release artifact. The results §11.6 claims ten changed files, but the actual rev3→rev4 delta has eleven, including this binary. Remove it from the managed candidate before the next handoff, and direct any single-binary builds to /tmp or the existing ignored bin directory. Update the results to acknowledge the cleanup. Preserve the validated source/test/docs changes. No product redesign or human decision is needed.

## Identity and review scope

Base `3c45d4bbed81348dfc814bac98b87a58aadb02af`; candidate tree `f7335cbce8e7e24f8006fc069a9993a08cb4cbaa`. Downloaded revision-4 patch SHA256 `064fc07ac1fadc9e216cb0ff064c89ef40e5802e0f24cd41f8bc7ae73b1e4f21` matches the assignment. All 41 changed paths in the managed worktree match candidate blobs. Execution used a disposable archive of this exact tree under /tmp; no candidate code was edited. Spec worktree is exactly `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`.

Read pinned environments §§2.2/2.3/3/5.5/10.3/11/12.1/12.2. Compared revision 3 tree `17d166e7664eeaafa15c01078788f2ca12d75cbf` to revision 4: the ten documented source/test/docs changes are the requested corrections; the binary is the sole extra path. E2 admission, E4 identity/resolution, S4 passthrough, pin, CHANGELOG and ledger are byte-identical between these revisions. Independently re-ran the input-patch added-line comparison and inspected parser, lock maps, status and vector entry points; rewrites correspond to recorded union arms/formatting, obsolete skips, E4 published-directory handling and this round's emission/null fixes. This supplements rather than replaces the prior semantic review.

## Per-item assessment

| Item | Result |
|---|---|
| Revision-3 S4 ordering correction | Closed. CLI install/import/update supply stdout as SurfacingSink. Five carrySurfacing sites precede publication/activation/resync; fresh install at envprofile.go:762 precedes publish :772, changed reinstall :902 precedes :911/:920, changed update :1149 precedes :1161/:1165. Sink output removes Info.Surfacing, preventing double print. Write errors are intentionally non-fatal. |
| Independent ordering probe | Replayed the exact prior TestReviewerSurfacingBeforePublication overlay against revision 4: PASS, exit 0. New CLI install/update observations also pass. |
| Runtime order tests | Install, changed-path reinstall, changed and unchanged update, same-source git reinstall, failure after emission and refusing sink tests pass. Both order vectors now invoke live operation observations instead of the removed formatting proxy. Coverage observes lock existence/old bytes at emission; code inspection establishes subsequent resync order, not a separate filesystem-materialization probe. |
| Revision-3 E4 null correction | Closed. environments.go:319 parses every present value. Exact prior Load null probe now passes with omitted/empty controls. Committed user/system null regression passes and checks knob attribution. E2 null-waiver regression remains green. |
| Revision-3 docs correction | Closed. docs/environment-config.md documents all policies, waiver example, default profiles, absent/null semantics, migration and locks. Historical audit untouched. |
| Pin and skip accounting | One SPEC_PIN equals dced9b8, five suite checkouts use env.SPEC_PIN; pre-existing optional candidate_ref remains distinct. release.yml and CI ledger unchanged. Published E2/E4/S4 families fail rather than root-content-skip when missing. No vector suppression introduced. |
| E2 | Default drop, opt-in error, root/overlay directness and waiver admission, error-direction-only locking, posture and exact comparisons retained. 5/5 admission and 7/7 schema cases pass; no prunePostRevisionKnobs workaround. |
| E4 | Revision A is active. Pinned spec explicitly retains PATH selection in A with outside-root warning; B selects trust roots and refuses PATH-only matches. Managed-directory refusals and identity tests retained. Local case identity and umbrella tests pass; Windows-only 8.3/SameFile test read, Windows execution reused from exact rev4 hosted evidence. |
| S4 | s4-warn shipped, s4-enforce selectable; presence/null/list handling, production Resolve bound and allowlist/status warnings retained. All 25 passthrough-family cases run. |
| Vectors | 48/48 manager-config cases, 7/7 schema subset, 5/5 admission, 14/14 umbrella cases with both revisions, 25/25 S4 cases execute. Targeted logs have zero skips. |
| Shell coverage bound | Full shell package passes. As previously reported, shell-hook-trust has no vector driver in this candidate; do not equate package success with that family's execution. S6 remains explicitly outside scope. |
| Release/scope | E2/E4/S4 and rc.12 CHANGELOG entries unchanged; no E1/R1/S6 source additions. Root build executable violates candidate hygiene and prevents acceptance. |

## Independent validation

Shell bash, set -o pipefail; CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc12-review/conformance/v1; every Go test uses -count=1. Transcripts in TASK-260917-16l2md_review-evidence-rev4.zip.

- go build ./...: exit 0; go vet ./...: exit 0; gofmt -l internal cmd: empty, exit 0.
- Full config, contextresolve, contextmaterialize, contextaudit, shell, interop/environments, envfragment, globalbins suites: exit 0.
- envprofile targeted mask: Test(EnvironmentsEnvPassthroughVectors|InstallEmits|UpdateEmits|ReinstallEmits|SurfacingEmitted|SurfacingSink|InstallError|UpdateError|Drop|StatusReportsPolicy|StatusErrorReports|PolicyFromConfigCarries): exit 0, 40 PASS entries, zero skips.
- cmd/curator targeted mask: Test(Umbrella|ProviderInputs|ActiveRevision|EnvStatus|ProfileInstall.*Surfac|ProfileUpdate.*Surfac): exit 0, 39 PASS entries, zero skips.
- Config vector/subset/null tests: exit 0, 59 PASS entries; admission vectors: exit 0, 6 PASS entries; zero skips.
- Exact prior publication and null reviewer probes: both exit 0.
- Gate self-test: 180 passed, 0 failed, exit 0.
- golangci-lint run ./internal/config/... ./internal/envprofile/... ./cmd/curator/...: 0 issues, exit 0.

Full envprofile/CLI suites were not independently replayed: targeted validation establishes correction closure and finds concrete rework. Full-package/Linux/Windows/race evidence is reused only from the exact revision-4 validation log: hosted run 35295586973 reports all required lanes successful, exit 0; optional rose-air/candidate lanes skipped. No full remote gate rerun. This is not a claim of exhaustive local execution.

## Narrowing mutants: 3/3 killed

Replayed prior mutations using realpath Go overlays against the disposable revision-4 tree; verified each target file is unchanged from revision 3 and each overlay differs by only the intended predicate. Candidate code untouched.

| Gate narrowing | Committed failing tests | Exit |
|---|---|---|
| E2 error refusal only when len(dropped)>1 | TestSystemPromptErrorRefusesFirst; system-module-transitive-error vector through SystemPrompt | 1 |
| E4 managed-directory refusal restricted to revision B | Three umbrella vectors, TestUmbrellaRefusedDirectoriesBothRevisions, symlink and case-identity tests | 1 |
| S4 allow unlisted variables when effective list is empty | Enforce-absent/empty-list tests and production Resolve TestResolvePassthroughKnobs/empty-list-bounds-all | 1 |

Bound: three targeted predicates, not exhaustive mutation coverage. Both ordering vector cases now observe production emission; no independent emission-order mutant was added this round.

## Lifecycle

No active run goal binding; directives checked, none pending. Campaign prohibits LOGBOOK.md edits, so this artifact and board note preserve the finding. Evidence is attached before routing to-dev. No accept_cr, commit_ack, code edits, commits, pushes or branch changes. Ordinary cleanup rework, not blocked.
