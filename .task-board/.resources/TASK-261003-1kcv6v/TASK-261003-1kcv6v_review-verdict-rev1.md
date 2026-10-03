# TASK-261003-1kcv6v \u2014 carry-rc14-lockstep: revision 1 review

Verdict: accepted. No open findings against the binding 16-path carrier review scope. CR-TASK-261003-1kcv6v-1 revision 1; base f17ea7331c258b59db84fc6ea83872cf638b51a0; candidate tree bdfbfa2ded7f5ea868a8af642379d61c7bfb3a3c.

## Swept surfaces

| Surface | Result / bound |
| --- | --- |
| Exact carrier identity | Independently resolved both sides with git rev-parse TREE:path: 16/16 identical blobs. Exact changed-path sets also match 16/16. CHANGELOG requires no exception. Table below. Worktree matches candidate for all 16 paths. |
| Rewritten base / ancestry | Gate commit 62446680d914eeec2252cc1e67fb65ba3ca4c90c has exactly the CR tree and direct parent f17ea7331c258b59db84fc6ea83872cf638b51a0. merge-base --is-ancestor returns 0 for f17ea733, 1 for dfaa557f, 64345d71 and 92ae3c19. Intersection of gate ancestry with the 116 old-base ancestors excluded by rewritten main is empty (0/116). Shared pre-rewrite historical ancestors are necessarily retained; replaced history is not grafted. Old OIDs were used only as read-only comparison inputs. |
| Main freshness | Fresh ls-remote --symref origin HEAD advertised main f17ea733; exact refs/heads/main fetch returned the same full OID. No upstream delta at observation. |
| Standard gate | Hosted run 37135163093 completed success on gate commit 62446680. Fresh gh query confirmed head and conclusion; producer CR validation log read. Its candidate job was skipped and is not counted as candidate evidence. |
| Candidate gate | Hosted workflow_dispatch run 37137617861 completed success on the exact CR tree, head 62446680. Candidate jobs and actual Candidate suite + platform-case gate steps pass 3/3 (Ubuntu/macOS/Windows). Default tests pass 3/3, race passes 2/2, lint, naming, interop and gate self-tests pass. Only optional rose-air skipped. |
| Suite attribution | Downloaded all three candidate artifacts. Each records spec revision e3a88cedba7a844c594348ee89654471db086f98 and manifest SHA-256 6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5; each platform-case report ends with gate ok. Ubuntu plan confirms full-root enforcement, 82 served / 0 deferred / 1 excluded with a required refusal assertion. This is candidate-only evidence, not a release claim. |
| Prior substance / architecture | Read TASK-261002-2ipeqa_review-verdict-rev3.md and TASK-261002-2ipeqa_review-verdict-rev4.md in full. Rev3 found no product defect but requested scope reconciliation; rev4 explicitly resolved R3-F1 and accepted. Reuse that substance review for the identical carried files. |
| Reuse boundary | Prior accepted full tree 9706982a had 17 paths, including internal/conformancecoverage/rc14_test.go. The expressly designated campaign 92ae3c19 has 16 and omits that additional test. This carrier satisfies the latest exact campaign instruction; it is not claimed identical to the entire prior 17-path tree. Prior additional regression/mutant evidence is historical, not asserted present or rerun here. Fresh carrier CI supplies current test evidence. |
| Hygiene | git diff --check succeeds. Protected-path diff for LOGBOOK.md, CI workflow and internal/hashing is empty. No source changes, commit, cherry-pick, push or LOGBOOK edit by reviewer. |

## Independent blob identity

| Path | CR candidate blob | Designated accepted campaign blob | Equal |
| --- | --- | --- | --- |
| .github/ci/conformance-case-counts.tsv | f8f1eb69a7d4e8f1632ed844c8aefdafaf3d6b7c | f8f1eb69a7d4e8f1632ed844c8aefdafaf3d6b7c | yes |
| .github/ci/conformance-gaps.tsv | d66d43ca540980269e64530f1713602f2dbe518e | d66d43ca540980269e64530f1713602f2dbe518e | yes |
| CHANGELOG.md | e3d440076ae48ea805b08d763f78e88cf960796a | e3d440076ae48ea805b08d763f78e88cf960796a | yes |
| cmd/curator/muse_test.go | 95412d346794dc43747388405563771d99d44d26 | 95412d346794dc43747388405563771d99d44d26 | yes |
| internal/buildrepo/acquisition_conformance_test.go | 752408f33abda26424f2c1b8cd7a46e094339348 | 752408f33abda26424f2c1b8cd7a46e094339348 | yes |
| internal/config/environments_conformance_test.go | 1668897bcdda9f5d2c8933cf799f2c4eec9faeb0 | 1668897bcdda9f5d2c8933cf799f2c4eec9faeb0 | yes |
| internal/conformancecoverage/content_hash_v2_gaps_test.go | 3a1c7453d5a20b960c53839aeab065e0ba2f0a89 | 3a1c7453d5a20b960c53839aeab065e0ba2f0a89 | yes |
| internal/conformancecoverage/coverage.go | 4d7ca1eda8e61740a68b934e4cc3edec709c085b | 4d7ca1eda8e61740a68b934e4cc3edec709c085b | yes |
| internal/contextlock/schema_conformance_test.go | b8355867fffed12aac3a07d885b453590aab08b5 | b8355867fffed12aac3a07d885b453590aab08b5 | yes |
| internal/envmarker/marker_env_schema_test.go | f41e75061d7b463e006da63b88ec1c65d6ee1fc5 | f41e75061d7b463e006da63b88ec1c65d6ee1fc5 | yes |
| internal/envprofile/muse_test.go | d2fc911797f3e36ef7270781d5482b6a3921558c | d2fc911797f3e36ef7270781d5482b6a3921558c | yes |
| internal/envprofile/read_failure_conformance_test.go | d85c33a1d5cdb56d305b8aba3073f0d9487a2a0d | d85c33a1d5cdb56d305b8aba3073f0d9487a2a0d | yes |
| internal/interop/environments/snapshot_acquisition_test.go | 2ad8a7ad5bb11311f8045d6ad5fa185df0dbd52a | 2ad8a7ad5bb11311f8045d6ad5fa185df0dbd52a | yes |
| internal/marker/schema_coverage_test.go | a2eb1c84333ecd8cdc5cff4784302120f496e929 | a2eb1c84333ecd8cdc5cff4784302120f496e929 | yes |
| internal/registry/schema_conformance_test.go | 3a3d2e5b20307e65ece37ae7e58792955d2220f0 | 3a3d2e5b20307e65ece37ae7e58792955d2220f0 | yes |
| internal/scriptpolicy/conformance_test.go | f3f2726d635f31e3bcefc1ae7ea9923c2471a29a | f3f2726d635f31e3bcefc1ae7ea9923c2471a29a | yes |

## Executed versus reused

Reviewer executed read-only Git comparisons/ancestry checks, board resource reads, fresh GitHub job queries, downloaded candidate artifacts and verified their identity/gate reports. No local Go build/test/vet/run, lint or mutation tests were rerun; no reviewer syspolicyd duration/crash pair is claimed. Producer TASK-261003-1kcv6v_results.md records fresh scoped 11-package candidate/default passes, exact six-package implementation command, 8/8 consumption gate, build/vet/lint and 301/301 gate self-test assertions, with -work, host lock and stable crash count 374. Those local results are accepted supporting evidence, not reviewer executions.

Exploratory failures were not counted as passes: an unsupported board schema query, a missing optional reviewer reference, trying spec OIDs in the curator repository, and a mixed JSON/download-output stream parse. Corrected queries succeeded; mixed stream parsing explicitly separated three dependency-download lines on Ubuntu and found zero failed JSON test events on Ubuntu/macOS. No failed read was interpreted as absence.

Sources: [carrier standard gate](https://github.com/relux-works/curator/actions/runs/37135163093), [carrier candidate/full matrix](https://github.com/relux-works/curator/actions/runs/37137617861), prior task-scoped rev3/rev4 verdict resources, and producer carrier results. Prior [accepted hosted evidence](https://github.com/relux-works/curator/actions/runs/37017428885) is historical only.

Queried spawn goal before verdict: run is not goal-bound. Sole branch: accept_cr revision 1, routing to integrating; no done transition or commit_ack. Findings recorded in board notes and this outcome; LOGBOOK unchanged per binding host rules.
