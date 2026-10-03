# TASK-261003-1kcv6v — carry-rc14-lockstep integration preconditions

Bound run: RUN-261003-bcd3b2. Revision 1 remains accepted, story_final, integrating. Runner owns synchronous landing after producer exit. No integration, checkpoint, generic handoff, status mutation, source edit or LOGBOOK edit performed.

`git rev-parse HEAD`: exit 0
`git diff --name-only f17ea7331c258b59db84fc6ea83872cf638b51a0 bdfbfa2ded7f5ea868a8af642379d61c7bfb3a3c -- . :!.task-board`: exit 0
`git diff --cached --name-only`: exit 0
`git diff --quiet`: exit 0
`git diff --cached --check`: exit 0
`git diff --cached --quiet bdfbfa2ded7f5ea868a8af642379d61c7bfb3a3c -- . :!.task-board`: exit 0
`git rev-parse bdfbfa2ded7f5ea868a8af642379d61c7bfb3a3c:.github/ci/conformance-case-counts.tsv`: exit 0
`git rev-parse refs/campaign/2ipeqa-rev4-20261003:.github/ci/conformance-case-counts.tsv`: exit 0
`git hash-object .github/ci/conformance-case-counts.tsv`: exit 0
`git rev-parse bdfbfa2ded7f5ea868a8af642379d61c7bfb3a3c:.github/ci/conformance-gaps.tsv`: exit 0
`git rev-parse refs/campaign/2ipeqa-rev4-20261003:.github/ci/conformance-gaps.tsv`: exit 0
`git hash-object .github/ci/conformance-gaps.tsv`: exit 0
`git rev-parse bdfbfa2ded7f5ea868a8af642379d61c7bfb3a3c:CHANGELOG.md`: exit 0
`git rev-parse refs/campaign/2ipeqa-rev4-20261003:CHANGELOG.md`: exit 0
`git hash-object CHANGELOG.md`: exit 0
`git rev-parse bdfbfa2ded7f5ea868a8af642379d61c7bfb3a3c:cmd/curator/muse_test.go`: exit 0
`git rev-parse refs/campaign/2ipeqa-rev4-20261003:cmd/curator/muse_test.go`: exit 0
`git hash-object cmd/curator/muse_test.go`: exit 0
`git rev-parse bdfbfa2ded7f5ea868a8af642379d61c7bfb3a3c:internal/buildrepo/acquisition_conformance_test.go`: exit 0
`git rev-parse refs/campaign/2ipeqa-rev4-20261003:internal/buildrepo/acquisition_conformance_test.go`: exit 0
`git hash-object internal/buildrepo/acquisition_conformance_test.go`: exit 0
`git rev-parse bdfbfa2ded7f5ea868a8af642379d61c7bfb3a3c:internal/config/environments_conformance_test.go`: exit 0
`git rev-parse refs/campaign/2ipeqa-rev4-20261003:internal/config/environments_conformance_test.go`: exit 0
`git hash-object internal/config/environments_conformance_test.go`: exit 0
`git rev-parse bdfbfa2ded7f5ea868a8af642379d61c7bfb3a3c:internal/conformancecoverage/content_hash_v2_gaps_test.go`: exit 0
`git rev-parse refs/campaign/2ipeqa-rev4-20261003:internal/conformancecoverage/content_hash_v2_gaps_test.go`: exit 0
`git hash-object internal/conformancecoverage/content_hash_v2_gaps_test.go`: exit 0
`git rev-parse bdfbfa2ded7f5ea868a8af642379d61c7bfb3a3c:internal/conformancecoverage/coverage.go`: exit 0
`git rev-parse refs/campaign/2ipeqa-rev4-20261003:internal/conformancecoverage/coverage.go`: exit 0
`git hash-object internal/conformancecoverage/coverage.go`: exit 0
`git rev-parse bdfbfa2ded7f5ea868a8af642379d61c7bfb3a3c:internal/contextlock/schema_conformance_test.go`: exit 0
`git rev-parse refs/campaign/2ipeqa-rev4-20261003:internal/contextlock/schema_conformance_test.go`: exit 0
`git hash-object internal/contextlock/schema_conformance_test.go`: exit 0
`git rev-parse bdfbfa2ded7f5ea868a8af642379d61c7bfb3a3c:internal/envmarker/marker_env_schema_test.go`: exit 0
`git rev-parse refs/campaign/2ipeqa-rev4-20261003:internal/envmarker/marker_env_schema_test.go`: exit 0
`git hash-object internal/envmarker/marker_env_schema_test.go`: exit 0
`git rev-parse bdfbfa2ded7f5ea868a8af642379d61c7bfb3a3c:internal/envprofile/muse_test.go`: exit 0
`git rev-parse refs/campaign/2ipeqa-rev4-20261003:internal/envprofile/muse_test.go`: exit 0
`git hash-object internal/envprofile/muse_test.go`: exit 0
`git rev-parse bdfbfa2ded7f5ea868a8af642379d61c7bfb3a3c:internal/envprofile/read_failure_conformance_test.go`: exit 0
`git rev-parse refs/campaign/2ipeqa-rev4-20261003:internal/envprofile/read_failure_conformance_test.go`: exit 0
`git hash-object internal/envprofile/read_failure_conformance_test.go`: exit 0
`git rev-parse bdfbfa2ded7f5ea868a8af642379d61c7bfb3a3c:internal/interop/environments/snapshot_acquisition_test.go`: exit 0
`git rev-parse refs/campaign/2ipeqa-rev4-20261003:internal/interop/environments/snapshot_acquisition_test.go`: exit 0
`git hash-object internal/interop/environments/snapshot_acquisition_test.go`: exit 0
`git rev-parse bdfbfa2ded7f5ea868a8af642379d61c7bfb3a3c:internal/marker/schema_coverage_test.go`: exit 0
`git rev-parse refs/campaign/2ipeqa-rev4-20261003:internal/marker/schema_coverage_test.go`: exit 0
`git hash-object internal/marker/schema_coverage_test.go`: exit 0
`git rev-parse bdfbfa2ded7f5ea868a8af642379d61c7bfb3a3c:internal/registry/schema_conformance_test.go`: exit 0
`git rev-parse refs/campaign/2ipeqa-rev4-20261003:internal/registry/schema_conformance_test.go`: exit 0
`git hash-object internal/registry/schema_conformance_test.go`: exit 0
`git rev-parse bdfbfa2ded7f5ea868a8af642379d61c7bfb3a3c:internal/scriptpolicy/conformance_test.go`: exit 0
`git rev-parse refs/campaign/2ipeqa-rev4-20261003:internal/scriptpolicy/conformance_test.go`: exit 0
`git hash-object internal/scriptpolicy/conformance_test.go`: exit 0
`git merge-base --is-ancestor f17ea7331c258b59db84fc6ea83872cf638b51a0 62446680`: exit 0
`git merge-base --is-ancestor dfaa557f 62446680`: exit 1
`git merge-base --is-ancestor 64345d71 62446680`: exit 1
`git merge-base --is-ancestor 92ae3c19 62446680`: exit 1
`git rev-parse 62446680^{tree}`: exit 0
`git hash-object LOGBOOK.md`: exit 0
`git rev-parse HEAD:LOGBOOK.md`: exit 0
`git diff --name-only f17ea7331c258b59db84fc6ea83872cf638b51a0 c58d6556636f6b7e5c6cf6c421aa15117b33fb21 -- . :!.task-board`: exit 0

Ancestry exit 1 results above are expected negative checks: replaced commits are not ancestors; they are not green exit-0 commands. Old objects were comparison inputs only. Identity 16/16, no CHANGELOG exception.

| Path | Candidate = campaign = working file blob | Equal |
| --- | --- | --- |
| .github/ci/conformance-case-counts.tsv | f8f1eb69a7d4e8f1632ed844c8aefdafaf3d6b7c | yes |
| .github/ci/conformance-gaps.tsv | d66d43ca540980269e64530f1713602f2dbe518e | yes |
| CHANGELOG.md | e3d440076ae48ea805b08d763f78e88cf960796a | yes |
| cmd/curator/muse_test.go | 95412d346794dc43747388405563771d99d44d26 | yes |
| internal/buildrepo/acquisition_conformance_test.go | 752408f33abda26424f2c1b8cd7a46e094339348 | yes |
| internal/config/environments_conformance_test.go | 1668897bcdda9f5d2c8933cf799f2c4eec9faeb0 | yes |
| internal/conformancecoverage/content_hash_v2_gaps_test.go | 3a1c7453d5a20b960c53839aeab065e0ba2f0a89 | yes |
| internal/conformancecoverage/coverage.go | 4d7ca1eda8e61740a68b934e4cc3edec709c085b | yes |
| internal/contextlock/schema_conformance_test.go | b8355867fffed12aac3a07d885b453590aab08b5 | yes |
| internal/envmarker/marker_env_schema_test.go | f41e75061d7b463e006da63b88ec1c65d6ee1fc5 | yes |
| internal/envprofile/muse_test.go | d2fc911797f3e36ef7270781d5482b6a3921558c | yes |
| internal/envprofile/read_failure_conformance_test.go | d85c33a1d5cdb56d305b8aba3073f0d9487a2a0d | yes |
| internal/interop/environments/snapshot_acquisition_test.go | 2ad8a7ad5bb11311f8045d6ad5fa185df0dbd52a | yes |
| internal/marker/schema_coverage_test.go | a2eb1c84333ecd8cdc5cff4784302120f496e929 | yes |
| internal/registry/schema_conformance_test.go | 3a3d2e5b20307e65ece37ae7e58792955d2220f0 | yes |
| internal/scriptpolicy/conformance_test.go | f3f2726d635f31e3bcefc1ae7ea9923c2471a29a | yes |

Read-only task-board worktree integrating exited 0: accepted revision 1, story_final, awaiting_landing; fresh protected main c58d6556636f6b7e5c6cf6c421aa15117b33fb21. Its non-board delta against the base is empty. This is a precondition observation, not an integration success claim. Runner must enforce authority freshness and validation reuse at landing.

Evidence reused, not rerun: TASK-261003-1kcv6v_change-request_rev1-validation.log records sh scripts/remote-gate.sh exit 0 and standard gate run 37135163093 success. TASK-261003-1kcv6v_review-verdict-rev1.md accepts revision 1 and records candidate run 37137617861 green on 3/3 OSes plus prior rev3/rev4 substance verdicts. No local Go build/test/vet/run or lint executed in this integration-only run; unchanged candidate bytes and existing accepted evidence are reused. No host crash/duration measurements claimed.

Exploratory schema(operation=change_request) returned exit 1 (unsupported operation); corrected CLI discovery and resource reads exited 0. No failure counted as a pass.
