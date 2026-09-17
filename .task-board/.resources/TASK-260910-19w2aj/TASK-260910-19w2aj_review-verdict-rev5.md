# TASK-260910-19w2aj revision 5 independent review

Verdict: CHANGES_REQUESTED; route to to-dev. No human decision needed.

Candidate 15962fa5c82c04ae3367d6909dfc3c352db3c871; base 62ea2d2ced3fac7c3c5f7a2ff4e887e0f92f4540. Product code unchanged. Run is not goal-bound.

## Findings

1. P1: unchanged legacy network-Git root entries in schema 2 cannot consume their successfully resolved lock. internal/install/draftsources.go:122-125 treats every network-Git root as an alias selector; declaringGitSource at lines 178-180 rejects a valid legacy declaration without Selector. Independent production CLI fixture: schema_version 2, skills [{name:"review",tag:"v1",git:"https://fixture.test/review.git"}], tagged repository already under configured skills root. project resolve exits 0; install --dry-run and real install both exit 1 with `source_member_invalid: locked review has no declaring source`. The spec §1 explicitly retains unchanged legacy named entries. Recover the indexed manifest declaration with separate legacy/selector handling. Also reconcile gitCacheKey (internal/closure/resolve.go:328-329): legacy network roots are snapshotted under the legacy source key, not necessarily the canonical repository key. Do not merely remove the selector refusal and leave the subsequent cache lookup broken. Cover default and custom legacy source locations, plus mixed selector/legacy manifests, through production CLI resolve and frozen install.

2. P1: configured-Git legacy root entries lose declared ref during frozen consumption. internal/install/draftsources.go:139-152 finds their repository but stores FrozenRef with Source only; LoadDraftFrozenNodes consequently leaves Resolved.Kind/Ref empty. Independent fixture with no origin, schema_version 2 and skills [{name:"review",tag:"v1"}]: resolve exits 0; dry-run exits 0 but prints a blank ref; real install exits 1 with `review: install marker is invalid for schema 2`. Recover the legacy root's accepted Source/Git/Ref from its indexed declaration, preserving configured-git package semantics. Add real-install and pinned-reinstall CLI coverage, including custom source, without changing frozen v1 marker validation.

Both cases use an isolated temp config/repository, no network acquisition and a lock created only by project resolve. The repro script and log are attached. These are ordinary implementation regressions in frozen identity recovery, not requests to reopen accepted lock or transport decisions.

## Verification and bounds

Independent commands in zsh, observed exit 0:
- go build -o /tmp/TASK-260910-19w2aj-review-cli ./cmd/curator
- go test -p 1 ./cmd/curator -run '^TestProjectResolve' -count=1 -timeout=90s: PASS, 21.634s (all 10 top-level tests).
- go test -p 1 ./internal/closure ./internal/install -run 'Draft|Refresh|LegacyInstallUntouched' -count=1 -timeout=90s: PASS, 7.722s and 13.037s.
- Python production CLI reproduction completed exit 0; individual CLI exit codes above and in attached log. Both valid legacy forms fail real installation: 2/2 probed forms, not exhaustive conformance coverage.

Read producer rev5 results and attached rev5-validation.log. Hosted run 35197218253 reports success and remote-gate exit 0, with Ubuntu/macOS/Windows tests, race lanes and lint green; rose-air skipped. These are accepted attached results, not independently rerun. Producer's two mutants are reported killed; not independently replayed. Actual production failures establish these findings without code mutation. No full local suite, independent platform run, or exhaustive refusal-clause claim.

Rev4 requested SkillsRoot and exact-alias fixes are present and their production tests pass independently. Preserve these and the complete Git inventory authentication, real Git selector marker and refresh rollback fixes. Rework the two legacy entry paths, add production CLI regressions, and hand off a new revision.

Exact revision verification: independently compared all 5,539 candidate blobs against worktree bytes, including symlink targets: zero mismatches. Gate commit 95ce43c4dc3e4961d6603d0bd4dc4425d9a5849f resolves to the exact candidate tree 15962fa5c82c04ae3367d6909dfc3c352db3c871. Verification command exited 0. Final spawn-goal query again reports no goal bound.
