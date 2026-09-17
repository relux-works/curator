# TASK-260910-19w2aj — revision 8 review

Verdict: CHANGES_REQUESTED (to-dev).
Candidate: c2a6930eebc9c58c8b6144772b4fa10fbc9cf441; base aa46ecd80ad0b83853586454723ea99fb76977a9. All 14 changed paths independently byte-compared with candidate Git objects: 14/14 match. No production or test code modified. Run goal queried before verdict: not goal-bound.

## P1 — explicit refresh never fetches legacy Git roots or transitive repositories

Location: cmd/curator/project_resolve.go:123-129 (DraftResolveConfig omits Fetch), internal/closure/resolve.go:167-173 (FetchExisting receives false). The existing ensureRepo path at internal/closure/closure.go:364 fetches only when that boolean is true. Alias acquisition fetches separately, so the alias branch-refresh test misses this lane.

Reproduced through a newly built production cmd/curator binary with isolated temp configuration, a real local upstream and its configured-root clone, and an unchanged legacy branch root. Resolve and real install succeed. Advance only scripts/tool.sh on upstream main. `project refresh app` exits 0 but publishes the SAME lock and old commit. Manually fetching the clone then running the identical refresh command immediately selects the new commit. Reproduced 2/2 forms: configured-git `{name,branch}` and network-git `{name,branch,git}` (canonical fixture URL declared, preexisting local checkout with local fixture upstream; no external network or live credentials).

This violates the task's explicit refresh/runtime-only-change acceptance, with unchanged legacy entries retained by skillfile-sources section 1. Refresh presents success while the remote branch update is invisible. Transitive acquisition uses the same disabled option (inspection; separate transitive remote-change reproduction not run).

Required rework: make the explicit resolve/refresh operation refresh the applicable legacy/transitive repositories through the admitted acquisition path, with deduplication, policy and failure semantics preserved. Keep frozen install/launch offline. Do not blindly fetch origin-less configured repositories or double-fetch aliases. Add production CLI regressions with temp configuration for a remote branch advance while its checkout stays stale and for newly required transitive refs; verify updated complete lock/runtime bytes, pinned consumption before refresh, and unchanged prior lock/bindings/install on acquisition failure. A narrowing mutant that fetches aliases only must fail these checks.

## Independent validation

Shell: zsh, direct commands without pipelines; all listed exit codes observed.

- `go test -p 1 ./cmd/curator -run '^TestProjectResolveDraftAllowlist' -count=1 -timeout=75s`: exit 0, 8.407s; 4/4 subtests (legacy/transitive denied and allowed). This closes rev7's allowlist propagation finding.
- `go test -p 1 ./cmd/curator -run '^TestProjectResolve(LocalCreatesLockThroughCLI|LegacyUntouched|DraftOffRefuses|GitRuntimeTamperRefusedThroughCLI|GitMissingMemberRefusedThroughCLI)$' -count=1 -timeout=75s`: exit 0, 11.108s.
- `go test -p 1 ./internal/closure -run '^Test(RefreshCatchesRuntimeOnlyAndBuildOnly|RefreshDraftSecondWriteFailurePreservesLock|OpenDraftFrozenMissingSnapshotFails|ResolveDraftRefusals)$' -count=1 -timeout=75s`: exit 0, 3.260s. Helper-level bounds, not CLI coverage of remote refresh.
- `go build -o .temp/review-rev8/curator ./cmd/curator`: exit 0.
- Attached Python repro, default and `--network`: exit 0 each means stale-lock bug reproduced AND manual-fetch positive control passed. All subprocess commands and exits are in attached logs. Reproduce by placing script beside the built binary and running Python; it creates isolated fixtures under that directory. Python is review evidence only, not proposed shipping code.

Accepted attached evidence, not independently rerun: rev8 producer lint/vet and alias-only-policy mutant (two denied cases killed); prior regression claims. Hosted gate log reports run 35216056402 success, exit 0, 1/1 command shard. Independently verified gate commit b465b5347dd97740c7bf2e15826f88585043c28e has exactly candidate tree c2a6930eebc9c58c8b6144772b4fa10fbc9cf441. Hosted macOS/Linux/Windows results are attached evidence; rose-air and candidate suite were skipped, not passing. Full local suite and independent code mutants were not run. The independent attack here is the production refresh counterexample above, with a manual-fetch control.

Outcome resources: this verdict; TASK-260910-19w2aj_review-rev8-refresh-repro.py; TASK-260910-19w2aj_review-rev8-refresh-repro.log; TASK-260910-19w2aj_review-rev8-network-refresh.log.
