# TASK-260922-cww1ov results — revision 6 (base refresh)

Status: refresh-only; no product behavior changed. This result records the refreshed candidate and local verification before the required developer handoff. The hosted validation result is not yet available here; it is run by the handoff workflow.

## Review record

The attached review-round text says revision 5 was rejected and asks for a regression test plus a narrowing mutant. The actual `TASK-260922-cww1ov_review-verdict-rev5.md` says **ACCEPTED** for identity/tree-bound evidence and carries the content acceptance from revision 3. `TASK-260922-cww1ov_results-rev5.md` also records that there was no content finding. No new test or mutant was added for a finding that does not exist. The accepted coverage and mutant tables remain in `TASK-260922-cww1ov_results-rev3.md`; revision 5 accepted that content. This refresh did not alter those test files.

## Refresh and candidate identity

- `task-board worktree refresh-candidate TASK-260922-cww1ov` — exit 0, `refresh_advanced`; refreshed trunk `fad881368b632f43a18b01681a8fc7def110cbae`, branch tip `84e2fb272834e55cdc61cce6a2052facdfd4d39d`.
- Candidate tree, computed with a temporary Git index from `HEAD` plus the uncommitted candidate: `8cfe34fb470add983e81f7ef4053e5604b116026`.
- The refresh initially left old-base content in paths changed by newer trunk commits. I restored unrelated stale worktree content to refreshed `HEAD` and reapplied the F-C3 additions to the two intended merge paths. This retained the trunk GitOps ledger rows and Changelog entry. Worktree `.task-board` snapshot files were restored to refreshed `HEAD`; the authoritative board was changed only through `task-board` CLI.
- `CHANGELOG.md` now equals refreshed `HEAD` plus the same 13 F-C3 lines in the existing top `### Added` section. Its diff is 14 insertions (13 content lines plus spacing), zero deletions.
- `.github/ci/platform-cases.tsv` retains refreshed-trunk rows and adds all 46 F-C3 rows. Its diff has 63 insertions and zero deletions; the four newer trunk GitOps rows are retained.
- F-C3 test/doc blobs are byte-identical to the revision 5 candidate (`768bacfa2c52a0906b80e233e8af14a275137ede`):

| Path | SHA-256 | Rev5 identical |
| --- | --- | --- |
| `cmd/curator/env_migrate_test.go` | `d3a04674657c7b1e2a1de67aa5740b2901e4dc848860d3938cde33ca47512908` | yes |
| `docs/troubleshooting.md` | `457ccdb9247cc62400b58447587dfd573934a32de027f582b6334561a569b5cb` | yes |
| `internal/envprofile/credential_link_test.go` | `2c6209317ae5eda911f4d1a737e5ca93a9a6e18ec046abbc0b7c83b7ec338e7a` | yes |
| `internal/envprofile/migrate_test.go` | `2b4c885502b60588efef9e63760377ebfb2bc46e22b5f9e2cc23db629fbc82` | yes |
| `internal/envprofile/reviewer_recovery_test.go` | `66fb9b5a2ed1be0f3aea9f0d444b4e683766e94fb9dcd67f51bb09d9cc863244` | yes |
| `internal/envprofile/credential_production_test.go` | `30b1c2d181e6050476bd38760d568a2de16edb4942a294ffb641ac9bdeda80ed` | yes |

Current `git status --short` is the expected eight task paths: seven modified tracked files (`.github/ci/platform-cases.tsv`, `CHANGELOG.md`, `cmd/curator/env_migrate_test.go`, `docs/troubleshooting.md`, and the three envprofile test files) plus untracked `internal/envprofile/credential_production_test.go`. Per-file diff: ledger +63/−0; Changelog +14/−0; CLI test +61/−0; troubleshooting +20/−0; credential-link test +18/−0; migration test +141/−4; reviewer recovery test +2/−0. The new production test is untracked and unchanged from revision 5. `git diff --check` passes.

## Coverage and ledger

| Coverage area | Production-entry rows | Revision 6 result |
| --- | --- | --- |
| 0017 stale shared-to-isolated link; regular file at link path | `TestSharedToIsolatedRemovesStaleLink`; `TestCredentialLinkRegularFileRefuses` | unchanged from accepted rev5; rerun in API mask, pass |
| Dangling native target: mis-targeted and dangling-to-declared | `TestDanglingPiLinkReportedDetached`; `TestReviewerDanglingExpectedTarget`; `TestDanglingExpectedTargetRepairSucceeds` | unchanged; rerun in API mask, pass |
| Migration, drift, journal/recovery, foreign temp preservation, no-copy | `TestMigrateNoSecretCopies`; `TestMigrateInterruptedApplyRecovers`; `TestMigrateRecoveryRefusesUnexpectedTarget`; `TestMigrateRecoveryCleansOwnedTemp`; related migration rows | unchanged; rerun in API mask, pass |
| Repair refusal classes and codex admission/TOML spellings | `TestCredentialLinkInspectionRepairFails`; `TestCodexIsolatedAdmission`; `TestCodexUnknownStoreSharedRefuses`; `TestCodexCredentialStoreTOMLSpellings`; `TestCodexMalformedStoreSharedRefuses` | unchanged; rerun in API mask, pass |
| CLI repair does not migrate silently | `TestEnvResolveRepairNeedsMigration` | unchanged; rerun in CLI mask, pass |
| Narrowing mutants | Full row-to-mutant table and bounds in accepted revision 3 results | carried unchanged from accepted revision 3; its mutant driver killed 47/47 mutants; no new mutant run for this refresh-only revision |

The F-C3 ledger block contains 46 rows: 37 all-platform rows with the existing Windows `host-capability` tolerance, 5 all-platform rows with no skip tolerance, and 4 Linux/macOS rows with the existing Windows `host-capability` tolerance. `host-capability` is the only nonempty skip class. Ledger consistency checks 291 total rows across Linux, Darwin, and Windows. The refreshed `SPEC_PIN` remains `dced9b8317e0e8af79edf2d0539b32bd22b6c85b` (rc.12); the accepted prior run recorded no F-C3 vector-consuming row whose skip/execute state changed.

## Verification (zsh; standalone commands)

| Command | Result |
| --- | --- |
| `go test -count=1 -timeout 9m -run 'Test(CredentialLink|SharedToIsolated|StaleCredential|Dangling|Codex|PiProvision|Reviewer|Migrate)' ./internal/envprofile/` | exit 0; package passed in 46.541s |
| `go test -count=1 -timeout 9m -run 'TestEnv(Migrate|ResolveRepair)' ./cmd/curator/` | exit 0; package passed in 123.044s |
| `sh .github/ci/ledger-consistency.sh` | exit 2; usage error because this script requires an evidence-directory argument |
| `sh .github/ci/ledger-consistency.sh .temp/TASK-260922-cww1ov-ledger-evidence` | exit 0; 291 rows checked across Linux, Darwin, Windows |
| `sh .github/ci/gate-selftest.sh` | exit 0; 185 passed, 0 failed |
| `go vet ./internal/envprofile/ ./cmd/curator/` | exit 0 |
| `go build ./...` | exit 0 |
| `gofmt -l` over all five touched Go files | exit 0; no paths printed |
| `git diff --check` | exit 0 |

The complete `go test ./internal/envprofile/...` suite was not rerun in this refresh. The refresh brief requested the owned narrow rows; the focused API and CLI masks above were rerun, and the accepted revision 5 evidence covers the unchanged broader suite. The configured hosted gate remains the cross-platform arbiter and is triggered by the task-board handoff; it is not claimed green in this artifact.
