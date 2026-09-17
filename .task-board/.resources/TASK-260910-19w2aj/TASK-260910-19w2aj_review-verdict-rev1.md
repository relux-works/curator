# TASK-260910-19w2aj revision 1 — CHANGES_REQUESTED

Candidate: 6af80797582172a2787e27b1f986f73cc643f3a3; base: 62ea2d2ced3fac7c3c5f7a2ff4e887e0f92f4540. Reviewed exact tree delta and confirmed tracked source matches candidate; four untracked leaf files match candidate blob hashes. No production or test source modified.

## Blocking findings

1. P1 — No production resolve/refresh entry point (internal/closure/resolve.go:85,182; internal/install/draftsources.go:64). Repository-wide Go caller search finds ResolveDraft only in tests and RefreshDraft, and RefreshDraft only in tests. The install path requires a preexisting lock and tells the user to run explicit resolve, but no command creates it. cmd/curator/main.go:1100 project resolve only prints project paths. Consequently users cannot resolve selected local/Git packages or explicitly refresh runtime/build changes through the delivered CLI. Wire the approved explicit production operation, including existing resolved transport acquisition, gates and publication. Add cmd/curator tests with fake transport and captured snapshots; do not seed the lock by calling the helper under test.

2. P1 — Git frozen consumption ignores selected directory (internal/closure/resolve.go:597-615). acquireGit selects the package subtree, but openGitFrozen returns cache/<repository>/<commit>/snapshot regardless of member.Package.Directory. LoadDraftFrozenNodes loads the spec from that returned repository root. A valid Git selection directory skills/review therefore loads the wrong package or fails when the repository root has no spec. Return the authenticated locked subtree with safe containment and test individual and collection Git selections through install/launch.

3. P1 — Git cache bytes are unauthenticated on frozen reads (internal/closure/resolve.go:603-615). Only Lstat and directory type are checked. A modified regular file under a commit-shaped cache directory is admitted to LoadDraftFrozenNodes and downstream consumers. No content/inventory check occurs in this path; local OpenLocal does perform verification, and existing snapshot.Get authenticates Git cache hits against extracted bytes. Preserve full Git snapshot integrity without network or live source adoption, including runtime/build files excluded from the context hash. Add negative tests for tampered Git snapshots and symlink/partial cache content at the production entry.

4. P1 — Failed refresh can replace the previous lock (internal/closure/resolve.go:187-192). sourcelock.Write publishes the lock before WriteBindings. If bindingsPath cannot be written, RefreshDraft returns an error after changing the lock, with no rollback. The existing atomicity test fails resolution before either write, so it cannot catch this publication failure. Stage and transactionally publish the required state after gates; inject failure at the second publication step and assert old lock/bindings/installed state remain unchanged. Contract skillfile-sources sections 3 and 5 requires failure preservation.

## Evidence and bounds

Applied project-management reviewer and negative-evidence rules. Read accepted skillfile-sources and repository-transport revision 1 requirements, producer results, lock leaf revision 3 acceptance bounds, and hosted validation resource. The lock leaf explicitly left CLI/manager consumption to this work; these findings do not reopen its accepted model decisions.

Independent read-only checks: candidate blob comparison and git diff --check base candidate: exit 0. Repository-wide caller search establishes missing production resolve/refresh wiring. Findings 2-4 are direct control-flow inspections, not claimed executable reproductions.

Independent zsh test attempt:
go test -p 1 ./internal/closure ./internal/install -run 'Test(ResolveDraft|RefreshDraft|Draft|OpenDraft|LoadDraft)' -count=1 -timeout=90s
Closure test process stalled, reported 'Test killed with quit: ran too long (2m30s)' and FAIL; invocation terminated with exit 143. No passing result.

Retried by compiling dedicated /tmp/TASK-260910-19w2aj-review-{closure,install}.test binaries with go test -c, ad-hoc signing, and running masks Test(ResolveDraft|OpenDraft|Refresh) and Test(Draft|LegacyInstallUntouched), count=1, timeout=90s. Both executions stalled without test output and were explicitly terminated: exit 143 each. No local test pass claimed. No processes intentionally left running.

Hosted evidence reused, not rerun: TASK-260910-19w2aj_change-request_rev1-validation.log reports run 35178513692 success and exit 0. Gate commit e24d1e1a2330093627cb6cbc14d00743a4192ef4 resolves to exact candidate tree 6af80797582172a2787e27b1f986f73cc643f3a3. Log reports Ubuntu/macOS/Windows test lanes, race/lint/conformance/naming/self-tests successful; rose-air and candidate suite skipped. This is attached gate evidence, not an independent platform replay.

Coverage bounds: 0 new cmd/curator tests in this candidate; new install tests use DryRun and helper-generated locks. No executable independent mutants completed; kill ratio unknown. Producer's 2 binding mutants concern local missing snapshot/context hashing and do not cover Git integrity/subdirectory or second-write rollback. Legacy test asserts status rather than byte-identical output. Require production positive/negative checks and narrowing mutants for repaired gates in the next revision.

Run goal queried via the spawning runtime binary: not goal-bound; no directives. Wrapper/current binary calls stalled; used already-installed task-board-main-6cb09a23-curatorlike, the binary hosting this reviewer, for board operations. No installation/runtime changes.

Verdict: changes_requested; route TASK-260910-19w2aj to to-dev for ordinary implementation rework and another independent review. This artifact also records findings in lieu of the campaign-prohibited LOGBOOK.md edit.
