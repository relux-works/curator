# TASK-260924-5c0747 results — curator v0.15.0-rc.2

## Candidate and release-prep changes

Prepared the rc.2 release-prep candidate in the STORY-260924-2go2bz worktree, based at checkpoint e8620502. The last fetch before handoff kept origin/main at 316438cc3f830801c44adba75750a6b28a156933. The latest non-board trunk diff was incorporated, including the newly landed TASK-260923-2gt5f6 credential-record work and its CHANGELOG entry. No commit, tag, push, or release was made; integration and signed release remain with the orchestrator.

All Curator SPEC_PIN references in CI, conformance tests, docs/ci-gates.md, conformance-case-counts, and the gap ledger now point to curator-spec v1.0.0-rc.13 commit 23435129ebc4c29e5b7f75ec72a0aa0cd3f16065. The released skillfile-sources-v1 corpus pin is also rc.13. The annotated remote tag peels to that commit. The rc.13 core conformance/v1 inventory has the same 1,170 paths as the interim dcc7f015 pin; the 19 changed files are the manifest and 18 vectors whose only semantic change is protocol_version labeling. The rc.13 vectors are labelled 1.0.0-rc.13, accepted by scriptpolicy while retaining the legacy rc.9 label.

A byte comparison of the vendored skillfile-sources-v1 corpus and schemas against rc.13 checked 136 files: zero mismatches and zero extra copied files. Source corpus manifest SHA-256: 061ec05ddb1746b72168d157f81d2930047cd62371ffa46cf585844da61ee6be. The vendored MANIFEST.sha256 check passed when run from the corpus directory. The only source changes since interim pin 5746367 were the corpus README and the added manifest.

CHANGELOG.md now has the v0.15.0-rc.2 section grouped as Added, Changed, and Fixed. Its 27 entries are copied from qualifying landed leaf results: Added 10, Changed 10, Fixed 7, from 19 source tasks. I scanned all results resources with a CHANGELOG entry and matched eligibility against task status and the commits in v0.15.0-rc.1..origin/main. Excluded TASK-260907-2as5sx because its parent Story is outside the range; TASK-260918-bi6ouz because it is still integrating and its parent is outside the range; TASK-260925-h4syhu because it is still in development; and TASK-260926-1ddneb because its parent is not landed and its result has no Curator entry.

Entry-to-source mapping, in CHANGELOG order:

Added:
1. Draft schema-9 skill-manifest dependency directory selection — TASK-260924-1kpw4w
2. Default Skillfile schema-2 source support — TASK-260924-m28s6b
3. Replay missing schema-2 locked snapshots — TASK-260924-m28s6b
4. Draft registry evidence production-entry tests — TASK-260924-10d3l1
5. Counted published-case outcome harness — TASK-260922-18ex37
6. Skillfile collection directory/include/exclude — TASK-260924-11burj
7. Manifest dependency directory selection — TASK-260924-11burj
8. Decision 0017 credential-link production-entry tests — TASK-260922-cww1ov
9. Explicit credential migration — TASK-260922-cww1ov
10. Managed-home schema-2 credential metadata and provenance — TASK-260923-2gt5f6

Changed:
1. Marker-v5 raw shape and declared-tag validation — TASK-260924-2v4v2m
2. Go per-package test timeout — BUG-260922-3v8k23
3. Interim SPEC pin and gap ledger — TASK-260922-18ex37
4. Environment profile follow-ups — TASK-260906-1f2ng0
5. rc.13 script-worker label — TASK-260926-4hd81z
6. codex_cli isolation policy — TASK-260922-cww1ov
7. Pi native credential root — TASK-260922-cww1ov
8. Credential repair no longer migrates ownership — TASK-260922-cww1ov
9. System environment isolation locks — TASK-260923-2elcdc
10. Released Skillfile sources conformance/replay/refresh/revocation — TASK-260924-20o9dk

Fixed:
1. Manager-lock expired-context test — BUG-260922-6chzf9
2. git check-ignore EACCES retry — BUG-260923-3mazfw
3. Captured SYSTEMROOT hard-link allowance — BUG-260924-5p8b0z
4. Audit-registry future-bound test — BUG-260923-krcm6m
5. Concurrent Windows snapshot retry — BUG-260923-11jgkt
6. Credential-link repairs fix-first — TASK-260922-cww1ov
7. rustup diagnostic block — BUG-260922-306v4m

Release prep follows v0.15.0-rc.1 commit d1bb0a4e92304526c782f77607aa30abc2f3ae07, “Keep a release candidate out of the install channels.” Comparing v0.15.0-rc.1..origin/main for .goreleaser.yml and .github/workflows/release.yml produced no diff. CI SPEC_PIN and conformance accounting are the release-prep changes; release.yml and goreleaser configuration remain as in rc.1.

## Validation evidence

Green bounded checks:
- CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test -count=1 -timeout 3m -run '^(TestSkillfileSourcesPin|TestSkillfileSourcesCorpusCounts|TestDraftSourcesSemanticCoverage|TestDraftSourcesSchemaCases)$' ./internal/crossconformance — exit 0.
- go test -count=1 -timeout 2m ./internal/conformancecoverage — exit 0.
- CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test -count=1 -timeout 3m -run '^(TestScriptWorkerProtocolVersionAcceptanceIsClosed|TestScriptHostExecutionPolicyProductionConsumersCoverAllCases)$' ./internal/scriptpolicy — exit 0.
- CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test -count=1 -timeout 3m -run '^TestExecutableIdentityCasesAtProductionEntry$' ./internal/scriptworker — exit 0.
- CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test -count=1 -timeout 3m -run '^TestConformanceSnapshotAcquisition$' ./internal/interop/environments — exit 0.
- CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test -count=1 -timeout 3m ./internal/envmarker — exit 0.
- go test -count=1 -timeout 2m ./tools/goreleaserconfig/ — exit 0.
- Focused schema-2 and credential migration tests: go test -count=1 -timeout 2m -run '^(TestMigrateSchema1UnlinkPublishesCompleteSchema2Marker|TestMigratePiWrongTargetToAgentRoot|TestMigrateNoSecretCopies|TestMigratePlanDriftRefuses|TestMigrateRollbackRestoresPriorState|TestMigratePublishFailureRevertsLinks)$' ./internal/envprofile — exit 0.
- Focused CLI migration and schema-2 credential-record tests: go test -count=1 -timeout 2m -run '^(TestEnvResolveCredentialRecord.*|TestEnvMigratePlanApplyPi|TestEnvMigrateConflictRefuses|TestEnvResolveRepairNeedsMigration|TestEnvMigratePrintBeforeWrite)$' ./cmd/curator — exit 0.
- go build -o /tmp/curator-rc2-build ./cmd/curator — exit 0.
- CI_REQUIRE_FULL_ROOT=1 bash .github/ci/suite-plan.sh /Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 /tmp/curator-rc2-evidence.oRyAiB/suite-plan — exit 0; 80 served, 0 deferred, 0 excluded.
- CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 bash .github/ci/ledger-consistency.sh /tmp/curator-rc2-evidence.oRyAiB/ledger — exit 0; 418 rows checked across linux/darwin/windows.
- bash .github/ci/release-source-gate.sh go.mod HEAD refs/remotes/origin/main — exit 0.
- golangci-lint run ./cmd/curator ./internal/crossconformance ./internal/envmarker ./internal/envprofile ./internal/interop/environments ./internal/scriptpolicy ./internal/scriptworker ./internal/conformancecoverage — exit 0, zero issues.
- go vet passed with exit 0 for ./cmd/curator, ./internal/crossconformance, ./internal/conformancecoverage, ./internal/envmarker, ./internal/envprofile, ./internal/interop/environments, ./internal/scriptpolicy, and ./internal/scriptworker.
- gofmt -l check over all modified/imported Go files — exit 0, no files listed. git diff --check — exit 0.
- shasum -a 256 -c --status ../MANIFEST.sha256 from internal/crossconformance/testdata/skillfile-sources-v1/corpus — exit 0.
- Final git fetch origin main — exit 0; origin/main remained 316438cc3f830801c44adba75750a6b28a156933.

Red or unavailable checks, recorded without masking:
- CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test -count=1 -timeout 5m ./internal/crossconformance — exit 1 at 300.666s; TestDraftSourcesSemanticCasesBatch3 reached the 5-minute Go test timeout. The bounded pin, count, semantic-coverage, and schema-case subset above passed.
- go test -count=1 -timeout 2m ./internal/envprofile — exit 1 at 120.446s in TestRemovePurgeCleansHomes; the stack was in Darwin transaction.syncDirectory while opening/syncing the transaction directory. The focused migration tests above passed.
- goreleaser check — exit 127 because the goreleaser executable is not installed locally. The repository's tools/goreleaserconfig test passed; the hosted release lane remains authoritative.
- An initial shasum manifest check from the parent directory exited 1 because MANIFEST.sha256 paths are relative to corpus/. Rerunning the same check from corpus/ exited 0 as recorded above.
- The previous Change Request revision 1 hosted run 36245916469 had one Windows failure, job 108414964500: TestEnforcedInstallAndLaunchAtCLIEntry reported worker Job Object flags 0x0, expected at least 0x2308. The same Windows Job Object symptom is recorded in TASK-260910-hwx26 results with direction to retry once unchanged and, if repeated, report it to the orchestrator. No workaround or test weakening was introduced. Handoff runs the configured landing suite once; its attached validation log is the next hosted result.

The full-module test gate was not run locally, per the bounded-run and memory constraints. Handoff runs the configured landing suite once. The signed tag and hosted release workflow are post-integration orchestrator actions; neither has been run by this producer.

## Orchestrator release commands

After the reviewed candidate is integrated into main, run:

~~~
git fetch origin main
git tag -s v0.15.0-rc.2 -m "v0.15.0-rc.2" origin/main
git push origin refs/tags/v0.15.0-rc.2
~~~

Watch .github/workflows/release.yml (workflow name Release) on the tag:

~~~
gh run list --workflow release.yml --ref v0.15.0-rc.2 --limit 1 --json databaseId --jq '.[0].databaseId'
gh run watch <run-id> --exit-status
~~~

No LOGBOOK.md edit was made; findings and anomalies are recorded here as the task-scoped outcome resource.


## Revision 3 — onto f02ba39e (2026-09-26)

Revision 2 was withdrawn because it was based on `e8620502`, before the later main landings. This section supersedes its candidate-base statement, changelog exclusions, and verification snapshot; the earlier command results remain historical evidence only.

### Refreshed candidate and path scope

- Started from clean `f02ba39e4c6f6d12af950c5138e8ffd7bbe8bb95`, matching `origin/main`.
- Verified `git diff 316438cc refs/campaign/5c0747-rev2-full-20260926 -- . ':!.task-board'` contained exactly these 14 paths, then applied it with `git apply --3way /tmp/TASK-260924-5c0747-rev3.patch` — exit 0. There were **no conflicts**. `git reset` exited 0 and left everything unstaged:
  - `.github/ci/conformance-case-counts.tsv`
  - `.github/ci/conformance-gaps.tsv`
  - `.github/workflows/ci.yml`
  - `CHANGELOG.md`
  - `docs/ci-gates.md`
  - `internal/crossconformance/draftsources_corpus_test.go`
  - `internal/crossconformance/testdata/skillfile-sources-v1/MANIFEST.sha256`
  - `internal/crossconformance/testdata/skillfile-sources-v1/README.md`
  - `internal/crossconformance/testdata/skillfile-sources-v1/SKILLFILE_SOURCES_PIN`
  - `internal/crossconformance/testdata/skillfile-sources-v1/corpus/conformance/skillfile-sources-v1/README.md`
  - `internal/crossconformance/testdata/skillfile-sources-v1/corpus/conformance/skillfile-sources-v1/manifest.json`
  - `internal/interop/environments/snapshot_acquisition_test.go`
  - `internal/scriptpolicy/conformance_test.go`
  - `internal/scriptworker/executable_identity_conformance_test.go`
- Fetched main before handoff — exit 0. `HEAD` and `origin/main` still equal `f02ba39e4c6f6d12af950c5138e8ffd7bbe8bb95`; no post-fetch trunk delta appeared. `git diff --name-only origin/main -- . ':!.task-board'` lists only expected tracked release-prep paths; the new manifest is the one expected untracked corpus file, making 14 total changed paths. There is no production-code change.
- `STORY-260906-1a2i5a` is done and landed as `7de564c4` after `v0.15.0-rc.1`. Its three relevant result sources (`TASK-260907-2as5sx`, `TASK-260925-h4syhu`, and `TASK-260907-187z6x`) are closed/done. The two other done children, `TASK-260906-19gjyw` and `TASK-260906-1uf713`, have no `*_results.md` resources and no release-prep entry text to copy.

### rc.13 pin and corpus

- `.github/workflows/ci.yml` `SPEC_PIN` and `SKILLFILE_SOURCES_PIN` both contain `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`. The curator-spec checkout is clean at that exact commit. Its local tag ref is absent, so byte comparisons used the checked-out commit hash directly; the commit tree contains the released corpus manifest.
- `diff -rq` between the rc.13 `conformance/skillfile-sources-v1` and its vendored copy exited 0. The rc.13 `schemas/skillfile-sources-v1` comparison also exited 0. Counts are 126 conformance files plus 10 schema files, 136 total, with no differing or extra copied files. The source `manifest.json` digest is `061ec05ddb1746b72168d157f81d2930047cd62371ffa46cf585844da61ee6be`; `shasum -a 256 -c --status ../MANIFEST.sha256` from the vendored corpus directory exited 0.
- The search for interim `dcc7f015`, vendored `5746367`, and rc.9 assumptions found no remaining current pin. `internal/scriptpolicy` intentionally accepts rc.9 as a legacy label and rc.13 as the current label, with rc.14 rejected. `docs/security-audit-2026-09.md` refers to rc.9 as the historical audit baseline. The interop comment now calls rc.9 historical rather than pinned.

### CHANGELOG source audit

The prior 27 bullets remain intact. Three newly required task sources contribute two distinct texts: `TASK-260907-2as5sx` contributes the absence-sensitive reader entry; `TASK-260925-h4syhu` and `TASK-260907-187z6x` carry identical same-source reinstall text, so the rc.2 section contains that text once and the mapping cites both sources. The updated rc.2 section has 10 Added, 10 Changed, and 9 Fixed bullets (29 unique bullets), from 22 source tasks.

The board query showed all 22 mapped source tasks terminal (`done`, except `TASK-260907-2as5sx` `closed`). A separate commit-subject check found all 20 distinct parent Story IDs in `v0.15.0-rc.1..origin/main` (exit 0). The latest Story commit `7de564c4` accounts for the new three source tasks. The current `TASK-260907-187z6x_results.md` has no exact `## CHANGELOG entry (for release prep)` heading; its accepted Change Request revision 2 contains the same text as the h4syhu results section. The corrected source comparator verified that equality and that the text occurs once in rc.2. It also verified the h4syhu and 2as5sx results entries each occur verbatim once.

Entry-to-source mapping, in CHANGELOG order:

Added:
1. Draft schema-9 skill-manifest dependency directory selection — `TASK-260924-1kpw4w`
2. Default Skillfile schema-2 source support — `TASK-260924-m28s6b`
3. Replay missing schema-2 locked snapshots — `TASK-260924-m28s6b`
4. Draft registry evidence production-entry tests — `TASK-260924-10d3l1`
5. Counted published-case outcome harness — `TASK-260922-18ex37`
6. Skillfile collection directory/include/exclude — `TASK-260924-11burj`
7. Manifest dependency directory selection — `TASK-260924-11burj`
8. Decision 0017 credential-link production-entry tests — `TASK-260922-cww1ov`
9. Explicit credential migration — `TASK-260922-cww1ov`
10. Managed-home schema-2 credential metadata and provenance — `TASK-260923-2gt5f6`

Changed:
1. Marker-v5 raw-shape and declared-tag validation — `TASK-260924-2v4v2m`
2. Go per-package test timeout — `BUG-260922-3v8k23`
3. Interim SPEC pin and gap ledger — `TASK-260922-18ex37`
4. Environment profile follow-ups — `TASK-260906-1f2ng0`
5. rc.13 script-worker label — `TASK-260926-4hd81z`
6. codex_cli isolation policy — `TASK-260922-cww1ov`
7. Pi native credential root — `TASK-260922-cww1ov`
8. Credential repair no longer migrates ownership — `TASK-260922-cww1ov`
9. System environment isolation locks — `TASK-260923-2elcdc`
10. Released Skillfile sources conformance/replay/refresh/revocation — `TASK-260924-20o9dk`

Fixed:
1. Manager-lock expired-context test — `BUG-260922-6chzf9`
2. git check-ignore EACCES retry — `BUG-260923-3mazfw`
3. Captured SYSTEMROOT hard-link allowance — `BUG-260924-5p8b0z`
4. Audit-registry future-bound test — `BUG-260923-krcm6m`
5. Concurrent Windows snapshot retry — `BUG-260923-11jgkt`
6. Credential-link repairs fix-first — `TASK-260922-cww1ov`
7. Manager-owned absence-sensitive reads — `TASK-260907-2as5sx`
8. Same-source git reinstall flag handling — `TASK-260925-h4syhu`, `TASK-260907-187z6x` (identical text; one bullet)
9. rustup diagnostic block — `BUG-260922-306v4m`

One initial ad-hoc comparator used an overbroad patch regex and raised a Python assertion; the surrounding shell continued to later `wc` commands and returned 0, so that invocation is **not** counted as passing evidence. A corrected `set -e` comparator exited 0 and verified the two result sections, the identical accepted-CR text, and the single occurrence in rc.2.

### rc.1 release-path comparison

The signed `v0.15.0-rc.1` tag points to `d1bb0a4e92304526c782f77607aa30abc2f3ae07` (“Keep a release candidate out of the install channels”). `git diff --exit-code v0.15.0-rc.1..origin/main -- .github/workflows/release.yml .goreleaser.yml` exited 0. The release workflow still runs on `v*`, checks release-source ancestry, then invokes the pinned GoReleaser v2 action; the binary version is derived from the tag. No extra version-file or release-workflow change was introduced.

`goreleaser check` was attempted and exited 127 (`goreleaser: command not found`). The Go config test and the self-test covering the rc channel values passed; the hosted release workflow remains the release arbiter.

### Revision 3 validation

Each command below ran directly in the worktree; exit codes are the process results.

- `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test -count=1 -timeout 3m -run '^(TestSkillfileSourcesPin|TestSkillfileSourcesCorpusCounts|TestDraftSourcesSemanticCoverage|TestDraftSourcesSchemaCases)$' ./internal/crossconformance` — exit 0.
- `go test -count=1 -timeout 2m ./internal/conformancecoverage` — exit 0.
- `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test -count=1 -timeout 3m -run '^(TestScriptWorkerProtocolVersionAcceptanceIsClosed|TestScriptHostExecutionPolicyProductionConsumersCoverAllCases)$' ./internal/scriptpolicy` — exit 0.
- `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test -count=1 -timeout 3m -run '^TestExecutableIdentityCasesAtProductionEntry$' ./internal/scriptworker` — exit 0.
- `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test -count=1 -timeout 3m -run '^TestConformanceSnapshotAcquisition$' ./internal/interop/environments` — exit 0.
- `bash .github/ci/gate-selftest.sh` — exit 0; 251 passed, 0 failed.
- `suite_plan_dir="$(mktemp -d /tmp/TASK-260924-5c0747-suite-plan.XXXXXX)" && CI_REQUIRE_FULL_ROOT=1 bash .github/ci/suite-plan.sh /Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 "$suite_plan_dir"` — exit 0; 80 served, 0 deferred, 0 excluded.
- `ledger_output_dir="$(mktemp -d /tmp/TASK-260924-5c0747-ledger.XXXXXX)" && CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 bash .github/ci/ledger-consistency.sh "$ledger_output_dir"` — exit 0; 450 rows checked across linux, darwin, and windows.
- `go test -count=1 -timeout 2m ./tools/goreleaserconfig` — exit 0.
- `golangci-lint run --concurrency 2 ./internal/crossconformance ./internal/conformancecoverage ./internal/interop/environments ./internal/scriptpolicy ./internal/scriptworker ./tools/goreleaserconfig` — exit 0, 0 issues.
- `go vet ./internal/crossconformance ./internal/conformancecoverage ./internal/interop/environments ./internal/scriptpolicy ./internal/scriptworker ./tools/goreleaserconfig` — exit 0.
- `go build -o /tmp/TASK-260924-5c0747-rev3-curator ./cmd/curator` — exit 0.
- `gofmt -l` over all four modified Go test files — exit 0, no output. `git diff --check` — exit 0.
- `diff -rq /Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/skillfile-sources-v1 internal/crossconformance/testdata/skillfile-sources-v1/corpus/conformance/skillfile-sources-v1` — exit 0; `diff -rq /Users/administrator/Developer/ReluxWorks/curator/curator-spec/schemas/skillfile-sources-v1 internal/crossconformance/testdata/skillfile-sources-v1/corpus/schemas/skillfile-sources-v1` — exit 0. `git -C /Users/administrator/Developer/ReluxWorks/curator/curator-spec cat-file -e '23435129ebc4c29e5b7f75ec72a0aa0cd3f16065:conformance/skillfile-sources-v1/manifest.json'` and `(cd internal/crossconformance/testdata/skillfile-sources-v1/corpus && shasum -a 256 -c --status ../MANIFEST.sha256)` — exit 0.
- Release workflow / GoReleaser config unchanged comparison — exit 0.

The full `internal/crossconformance` package was not rerun; this revision used bounded test masks. The earlier revision-2 full-package attempt remains an exit-1 timeout, not a pass. The hosted landing suite will run once through handoff; no full-module test run is claimed here.

### Orchestrator signed-tag command and workflow to watch

After the reviewed candidate is integrated into main, run:

~~~sh
git fetch origin main
git tag -s v0.15.0-rc.2 -m "v0.15.0-rc.2" origin/main
git push origin refs/tags/v0.15.0-rc.2
~~~

Watch `.github/workflows/release.yml` (workflow name `Release`) on that tag:

~~~sh
gh run list --workflow release.yml --ref v0.15.0-rc.2 --limit 1 --json databaseId --jq '.[0].databaseId'
gh run watch <run-id> --exit-status
~~~

No tag, push, or release workflow was run in the developer worktree. The signed tag and hosted green release remain post-integration orchestrator steps. No `LOGBOOK.md` edit was made; task findings are recorded in this outcome resource.

## Handoff correction — stale recorded checkpoint (2026-09-26)

The first developer handoff attempt returned exit 0 and set the task to `to-review`, but it attached the wrong candidate. Inspection of its remote gate snapshot showed commit `b2371cad9cb8818ac6f874d68d32f5df60d7416d` parented by stale checkpoint `e8620502`, not the current worktree base `f02ba39e`. The attached rev1 patch contained the 14-path release-prep delta but omitted the two new Fixed changelog entries; the previous rev2 patch also remained the withdrawn 31-path candidate. This first handoff is superseded and does not count as the revision-3 handoff.

The reused validation log names run `36245916469`; its Windows `test-gate` and platform-case gate exited 1 at required case `internal/install::TestEnforcedInstallAndLaunchAtCLIEntry`. The attached log reports `failure_class=unknown`; I do not infer a cause from the case name. This red result belongs to the stale `e8620502` snapshot and is not evidence for the current release-prep candidate.

I reopened the task to `development` (exit 0) and ran `task-board worktree refresh-candidate TASK-260924-5c0747` (exit 0). It returned `refresh_already_current` with `TrunkOID`, `ReviewedTrunkOID`, and `BranchOID` all equal to `f02ba39e4c6f6d12af950c5138e8ffd7bbe8bb95`. The worktree content is still the uncommitted 14-path release-prep delta described above. A fresh handoff of this checkpoint is required; only its own attached revision and validation log establish the current candidate result.
