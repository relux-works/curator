# TASK-260916-1zgucp results

Status: implementation is ready for review. Changes remain uncommitted on `task-board/story/STORY-260916-ioemse`.

Spec source: curator-spec v1.0.0-rc.13, pinned commit `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`, `protocol/environments.md`; configuration layering follows manager-config §1.

## Rules, production paths, and vectors

| Rule | Spec clauses | Production path exercised | Evidence |
|---|---|---|---|
| Enforce each configured source signer allowlist during Git resolution; accept either valid allowlisted tag or peeled-commit evidence, refuse unsigned/wrong signers, and do not fall back to a lower candidate after verification failure. | Environments §1.4; §12.1 signer configuration/defaults; manager-config §1 system/machine merge. | Config `Load` → `PolicyFromConfig` → `contextresolve.Resolve` / `gitManager.VerifyCandidate`; profile install uses the same resolver path. | All 19 pinned `verification_cases` through `Resolve`; all 6 `merge_cases` through `Load`; real-Git CLI tests for unsigned refusal, wrong SSH signer, accepted signed tag, empty allowlist, and exact-revision commit verification. Diagnostics checked: `context_source_unsigned`, `context_source_signer_rejected`, and required-list behavior. |
| Print the resolved-version delta and require a per-run confirmation when it touches a system-module inventory or MCP declaration. Refusal leaves the old lock and candidate store entries untouched. | Environments §9.2 (delta order/grammar, system and MCP trigger sets, update/reinstall/`--all`, revision B); §12 (no persistent confirmation posture). | `contextlock.ResolvedDelta` → `envprofile.resolvedDelta` → `updateLocked`; CLI update, `--all`, and same-source reinstall pass `--confirm-system-delta` through. | All 29 `delta_cases`; production CLI goldens for system-module and MCP deltas; both `all_cases` and both `reinstall_cases` through CLI; `UpdateWithOptions` refusal test checks delta output, unchanged lock, and absent candidate store entry. |
| Report signer and update-confirmation posture in status. An enforced pin with no local source material is `unknown` but current; a locally available pin that fails verification is non-current. Read failures remain non-current. | Environments §12 and §12.1; revision-B behavior from §9.2. | `StatusOf` plus human and JSON `env status`; status re-verifies cached sources without fetching. | All 5 `posture_cases` through `StatusOf`; current B-flip confirmation posture vector through `StatusOf`; real signed install followed by human and JSON CLI status. |

The pinned `environments-source-signers.json` families contain 65 rows: 19 verification, 6 merge, 5 posture, 29 delta, 2 all, 2 reinstall, and 2 confirmation-posture cases. The revision-B posture is shipped; the A-warning row is retained as historical protocol context, not simulated as the current manager behavior. The two deliberately nonconforming verification observations (`unsigned-accepted`, `fallback-selection`) are rejected.

## Conformance-gap ledger

| Count | Before | After | Result |
|---|---:|---:|---|
| All rows in `.github/ci/conformance-gaps.tsv` | 69 | 69 | No whole row became green because the affected config cases still fail on the separately owned `environments.permissions` field. |
| Rows associated with this Story/task | 39 | 0 | All 39 remain in the ledger with ownership transferred to `STORY-260922-1cenbr`, which owns the remaining permissions gap. |
| Whole gap rows removed | — | 0 | No passing whole row was left behind; the source-signer portions pass, while the full combined config cases remain blocked by permissions. |

## Mutation results

Every mutant was restored; each source restoration was checked with `diff -u` (exit 0).

| Rule narrowed or bypassed | Mutant-killing test | Real exit |
|---|---|---:|
| Signer allowlist | Only the first configured signer was considered; the second-allowlisted signature test rejected that mutant. | 1 |
| System-module trigger | Context inventory classification was disabled; pinned system-delta vectors failed on expected triggers/refusals. | 1 |
| MCP declaration trigger | MCP member classification was disabled; the MCP confirmation golden observed an unconfirmed update publish. | 1 |
| Confirmation gate | Update proceeded whenever a triggered delta lacked confirmation; `TestUpdateSystemDeltaRefusesBeforePublication` observed `moved=true` with no error. | 1 |
| Status posture | Missing cached source was incorrectly marked non-current; the `enforced-unknown-without-local-material` posture vector failed. | 1 |
| Closure refresh for update | MCP source refresh was skipped; the reinstall-without-flag vector observed an update proceed without the required refusal. | 1 |

The system CLI golden mutant attempt was interrupted by a concurrent host toolchain test process (exit 1); the system classifier mutant was killed by the pinned delta-vector test above, and the unmutated system CLI golden passed in the focused CLI suite.

## Verification

Final focused and task-relevant commands run directly:

| Command | Exit |
|---|---:|
| `CURATOR_CONFORMANCE_ROOT=/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T//curator-rc13.WNcrZu/conformance/v1 go test ./internal/contextresolve -run '^(TestSourceSignerVectorsAtResolve|TestResolveRejectsValidSignatureFromUnlistedSigner|TestResolveAcceptsAllowedSignatureWhenOtherVerificationErrors)$' -count=1` | 0 |
| `CURATOR_CONFORMANCE_ROOT=/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T//curator-rc13.WNcrZu/conformance/v1 go test ./internal/config -run '^(TestSourceSignerMergeVectorsAtLoad|TestManagerConfigV2SchemaCases|TestSystemConfigV2SchemaCases|TestManagerConfigV2Vectors)$' -count=1` | 0 |
| `CURATOR_CONFORMANCE_ROOT=/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T//curator-rc13.WNcrZu/conformance/v1 go test ./internal/envprofile -run '^(TestUpdateErrorLeavesLockUnchanged|TestSystemDeltaVectors|TestStatusSignerPostureUnknownWithoutCachedSource|TestStatusSignerPostureReadFailureIsNonCurrent|TestSignerPostureVectorsAtStatus|TestUpdateConfirmationPostureVectorAtStatus|TestUpdateEmitsSurfacingBeforePublication|TestReinstallEmitsSurfacingBeforePublication|TestUpdateSurfacesCandidateMCPSet|TestSurfacingUnreadableManifest)$' -count=1 -timeout=300s` | 0 |
| `CURATOR_CONFORMANCE_ROOT=/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T//curator-rc13.WNcrZu/conformance/v1 go test ./cmd/curator -run '^(TestSourceSignerVectorsAtProfileInstall|TestRevisionDoesNotBorrowTagSignature|TestProfileUpdateSystemDeltaConfirmationGolden|TestProfileUpdateMCPDeltaConfirmationGolden|TestProfileUpdateAllDeltaVectorsAtCLI|TestProfileUpdateAllStopsAtFirstUnconfirmedDelta|TestProfileInstallReinstallDeltaVectorsAtCLI|TestProfileUpdateListsNewDeclaration|TestProfileUpdateSurfacesBeforePublication)$' -count=1 -timeout=360s` | 0 |
| `CURATOR_CONFORMANCE_ROOT=/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T//curator-rc13.WNcrZu/conformance/v1 go test ./internal/envprofile -run '^(TestUpdateSystemDeltaRefusesBeforePublication|TestSystemDeltaVectors|TestStatusSignerPostureUnknownWithoutCachedSource|TestStatusSignerPostureReadFailureIsNonCurrent|TestSignerPostureVectorsAtStatus|TestUpdateConfirmationPostureVectorAtStatus)$' -count=1 -timeout=120s` | 0 |
| `go build ./...` | 0 |
| `make lint` | 0 (`0 issues`) |
| `git diff --check` | 0 |

A broader `go test ./internal/config ./internal/contextlock ./internal/contextresolve ./internal/envprofile -count=1 -timeout=360s` run was interrupted after 155.642 seconds while a concurrent host `envprofile.test` process was still active (overall exit 1). Its config, contextlock, and contextresolve package outputs were green; the full envprofile package was not completed. The focused envprofile suite above passed. An initial lint run exited 2 on two test-helper findings; both were fixed and the final lint run exited 0.

Real SSH signature verification was exercised through profile install. The GPG signer vectors exercised resolver evidence/allowlist decisions; real OpenPGP verification was not locally exercised.

## Files and handoff notes

No `CHANGELOG.md` or `LOGBOOK.md` edits. Important findings are recorded here. Work remains uncommitted in the assigned worktree.

## CHANGELOG entry (for release prep)

Configured SSH and OpenPGP signer allowlists now fail closed during context resolution. Profile updates print the resolved lock delta and require `--confirm-system-delta` for system-module or MCP declaration changes; `env status` reports signer and update-confirmation posture.

## Revision 3 — gate fix

The rev2 hosted gate exposed three actionable regressions. `deltaMemberRoot`, `gitManager.VerifyCandidateLocal`, and `gitManager.inspectEntry` now route manager-owned metadata reads through `internal/stateread`; only `KindAbsent` permits the Git snapshot fallback, while read failures and unusable present state fail closed. `UpdateWithOptions` now validates a path root's state pin before refreshing any locked source. The missing-pin test points at a local HTTP server and asserts that update refusal makes zero requests. The verified-signer status fixture supplies Git identity for both the commit and signed tag.

The manager-read audit now covers **364/364 (100.0%)** relevant readers: 250 guarded through the seam and 114 reviewed allowlist entries, across 409 production files.

### Revision 3 verification

| Command | Exit |
|---|---:|
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run '^(TestManagerOwnedAbsenceReadsAreGuarded|TestUpdatePathWithoutStatePinIsSourceInvalid|TestSignerPostureVectorsAtStatus)$' -count=1 -timeout=120s` (first pass, before routing the two signer-cache readers through the seam) | 1 |
| `go test ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$' -count=1 -timeout=120s` | 0 |
| `go test -v ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$' -count=1 -timeout=120s` | 0 |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run '^(TestUpdatePathWithoutStatePinIsSourceInvalid|TestSignerPostureVectorsAtStatus)$' -count=1 -timeout=120s` | 0 |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test -race ./internal/envprofile -run '^(TestManagerOwnedAbsenceReadsAreGuarded|TestUpdatePathWithoutStatePinIsSourceInvalid|TestSignerPostureVectorsAtStatus)$' -count=1 -timeout=180s` | 0 |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run '^(TestSystemDeltaVectors|TestUpdateSystemDeltaRefusesBeforePublication|TestUpdateConfirmationPostureVectorAtStatus)$' -count=1 -timeout=180s` | 0 |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./cmd/curator -run '^(TestProfileUpdateSystemDeltaConfirmationGolden|TestProfileUpdateMCPDeltaConfirmationGolden)$' -count=1 -timeout=180s` | 0 |
| `go test ./internal/envprofile -run '^(TestStatusSignerPostureUnknownWithoutCachedSource|TestStatusSignerPostureReadFailureIsNonCurrent)$' -count=1 -timeout=120s` | 0 |
| `go build ./...` | 0 |
| `make lint` | 0 (`0 issues`) |
| `git diff --check` | 0 |

The first targeted run exited 1 because the audit then identified the two signer-cache readers listed above; after routing them through the seam, the same three-test race run and the standalone guard passed. The signer resolver/config vectors and mutant probes recorded above were not repeated during this gate-fix pass. The pinned 29-case update-delta vector family and both CLI delta goldens were rerun. No full landing suite was run manually; handoff owns that suite.

Fetched `origin/main` at `97ca3370` (fetch exit 0). The trunk change in `internal/envprofile/managed.go` is not part of this task's diff, so the candidate does not revert that change. If publication refuses the stale candidate base, refresh the candidate through `task-board worktree refresh-candidate` before retrying handoff.

No `CHANGELOG.md` or `LOGBOOK.md` edits were made. The release-prep entry remains above.

## Revision 4 — re-apply on d41da0fb (carry-forward republish)

The assigned worktree was clean at `d41da0fb`, so I applied the accepted revision 3 tree from `refs/campaign/ioemse-full-20260927` using the requested three-way patch. The patch contains 31 paths. After resolution, `git diff --name-only HEAD -- . ':!.task-board'` lists exactly those 31 paths, with no unresolved entries or extra root task, test, or ledger files.

For all 24 paths trunk did not touch, the worktree bytes match the accepted revision 3 tree. The seven intersecting paths combine as follows:

| Path | Trunk change and resolution |
|---|---|
| `.github/ci/conformance-gaps.tsv` | Kept trunk's row set as the merge base. Removed the 39 E1-owned rows changed by revision 3. The pinned rc.13 coverage run also proved 16 pre-existing rows pass after both features are present, so those stale rows were removed too; the ledger is 69 → 14 data rows. No rows removed by trunk were re-added. |
| `internal/config/config.go` | Kept trunk's `environments.permissions` lock key and added the accepted `source_signers` and `require_source_signers` keys. |
| `internal/config/environments.go` | Combined trunk's permission parser/default/rendering with signer parsing, rendering, and lock mapping. Pinned valid OpenSSH fixtures omit base64 padding, so signer parsing now accepts both padded and unpadded standard base64. |
| `internal/config/environments_conformance_test.go` | Preserved trunk's system-module conformance drivers and signer merge coverage. The overlay regression now drives all 14 affected schema cases through `Load` and 14 pinned vectors through `Parse`, asserting exact environment output and absence of passing gap rows. The manager vector comparison projects the `environments` section, as the rc.13 vectors specify, while still rejecting extra environment knobs. |
| `internal/config/environments_test.go` | The system lockable set includes trunk permissions plus both signer knobs. Updated the deterministic unknown-field test because the signer fields are now supported. |
| `internal/envprofile/status.go` | Preserved trunk's profile diagnostics and combined them with signer posture and its non-current calculation. |
| `internal/envprofile/surfacing_test.go` | Kept the revision 3 confirmation-aware update call and trunk's empty-allowlist warning/current-status coverage. |

The ledger change is evidence-driven: besides the 39 revision 3 rows, the pinned conformance driver reported these 16 trunk rows as now passing after the merge: `valid.json`, `valid-system-module-waiver.json`, `invalid-transitive-system-modules-value.json`, `invalid-system-module-waiver-{package-grammar,missing-reason,unknown-field}.json`, `valid-provider-directories{,-windows-drive}.json`, `schema2-{minimal-defaults,empty-environments-defaults,partial-knobs-fill-defaults,permissions,provider-directories}.json` (vector IDs omit `.json`), and `schema2-passable-env-{absent-empty,explicit-null-unbounded,explicit-empty}`. Removing them leaves no passing row in the ledger; the 14 remaining rows belong to other unresolved gaps.

### Revision 4 verification

All commands below ran directly, with real exit codes captured. The conformance root was extracted from curator-spec commit `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065` into `$TMPDIR/curator-rc13-23435129`.

| Command | Exit |
|---|---:|
| `go build ./...` (final run after the merge) | 0 |
| `CURATOR_CONFORMANCE_ROOT=$TMPDIR/curator-rc13-23435129 go test ./internal/contextresolve ./internal/contextlock` (final full-package run) | 0 |
| `CURATOR_CONFORMANCE_ROOT=$TMPDIR/curator-rc13-23435129 go test ./internal/config` (final full-package run) | 0 |
| `CURATOR_CONFORMANCE_ROOT=$TMPDIR/curator-rc13-23435129 go test ./internal/envprofile -run 'Signer|Delta|Status|Guarded|StateRead|Surfacing'` | 0 (188.861s) |
| `CURATOR_CONFORMANCE_ROOT=$TMPDIR/curator-rc13-23435129 go test ./cmd/curator -run '^(TestSourceSignerVectorsAtProfileInstall|TestRevisionDoesNotBorrowTagSignature|TestProfileUpdateSystemDeltaConfirmationGolden|TestProfileUpdateMCPDeltaConfirmationGolden|TestProfileUpdateAllDeltaVectorsAtCLI|TestProfileUpdateAllStopsAtFirstUnconfirmedDelta|TestProfileInstallReinstallDeltaVectorsAtCLI|TestProfileUpdateListsNewDeclaration|TestProfileUpdateSurfacesBeforePublication)$' -count=1 -timeout=360s` | 0 (108.184s) |
| `CURATOR_CONFORMANCE_ROOT=$TMPDIR/curator-rc13-23435129 go test ./internal/contextresolve ./internal/contextlock ./internal/envprofile -run 'Signer|Delta|Status|StateRead|Guarded'` | 0 (envprofile 140.762s) |
| `CURATOR_CONFORMANCE_ROOT=$TMPDIR/curator-rc13-23435129 go test -v ./internal/config -run '^(TestManagerConfigV2SchemaCases|TestSourceSignerMergeVectorsAtLoad|TestSystemConfigV2SchemaCases|TestSystemConfigV2IsolationDirectionsFromPinnedCases|TestManagerConfigV2Vectors|TestOverlayConformancePassesWithPermissionsAndSourceSigners)$'` | 0 |
| `make lint` | 0 (`0 issues`) |
| `gofmt -d` on changed Go files | 0 |
| `git diff --check` | 0 |

The verbose conformance run drove 104/107 manager schema cases, 55/56 manager vectors, and 41/42 system schema cases; the remaining 3, 1, and 1 rows are still known gaps. The overlay regression's 14 schema cases and 14 vectors all passed. The six source-signer merge vectors also passed through `Load`.

The first nine `go test ./internal/config` attempts exited 1 while fixing the newly merged parser/test projection and removing rows the pinned runner reported as now passing; the final full-package run above exited 0. The first `gofmt -d` check exited 1 on one alignment difference; after formatting, the final check exited 0. The broad `go test ./cmd/curator -run 'Profile|Signer|Delta|Status'` pattern was interrupted after more than eight minutes to keep the run bounded (real exit 1); the focused named suite immediately afterward passed with exit 0.

Revision 3's accepted reviewer evidence remains applicable to the unchanged signer and confirmation rules: its verdict and prior results record narrowing mutants killed with real exit 1, the real CLI golden coverage, and the hosted revision 3 validation log. No `CHANGELOG.md` or `LOGBOOK.md` file was edited; the release-prep entry above remains the proposed text.

## Revision 5 — requested re-apply fixes; blocked by pinned conformance conflict

Applied the two requested rev5 corrections only: removed the `base64.RawStdEncoding` fallback from `parseOpenSSHPublicKey`, and restored trunk's full-object `exactManagerEffectiveJSON` plus `TestManagerEffectiveJSONComparisonRejectsExtraKnobs`. The restored named test was attacked with a narrowing mutant that compared only expected keys and ignored unexpected actual keys; `go test ./internal/config -run '^TestManagerEffectiveJSONComparisonRejectsExtraKnobs$' -count=1` failed as expected (exit 1). The file was restored from a saved copy and `cmp` returned 0.

Relative to rev4 tree `3a680a1e`, `git diff --name-only 3a680a1e -- . ':!.task-board'` lists only `internal/config/environments.go` and `internal/config/environments_conformance_test.go`. `git diff --check` exited 0. No CHANGELOG or LOGBOOK file was changed.

### Revision 5 verification

| Command | Exit |
|---|---:|
| `CURATOR_CONFORMANCE_ROOT="$TMPDIR/curator-rc13-23435129" go test ./internal/config` | 1 |
| `CURATOR_CONFORMANCE_ROOT="$TMPDIR/curator-rc13-23435129" go test ./internal/envprofile -run 'Status|Surfacing|Signer|Delta|Guarded'` | 0 (165.574s) |
| `go build ./...` | 0 |
| `make lint` | 0 (`0 issues`) |
| `git diff --check` | 0 |
| Narrowing mutant run of `TestManagerEffectiveJSONComparisonRejectsExtraKnobs` | 1 (expected: the narrowed comparator admitted the extra actual key) |
| `cmp` of restored conformance test against its saved pre-mutant copy | 0 |

The required config package run is red for two independent contract conflicts:

1. The pinned rc.13 `schema-cases/manager-config-v2/valid.json` and `schema-cases/system-config-v2/valid.json` contain the same unpadded SSH key material, `AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2A5GK` (66 base64 characters, no `=`). With the rev5 fallback removed, `base64.StdEncoding.DecodeString` rejects it. The package also reports 14 valid overlay schema cases rejected for the same reason. The pinned protocol text at `protocol/environments.md` §12.1 describes the key as an OpenSSH public-key line containing `<base64>` but does not state a padding requirement. Rev4's F1 finding says to reject this widening based on the interpretation that OpenSSH authorized_keys material is padded; the pinned valid fixtures require the opposite behavior.
2. The pinned `vectors/manager-config-v2.json` expected object for `schema2-empty-environments-defaults` contains only the `environments` section, while `EffectiveJSON()` includes top-level `adapter_mode` and `default_agents`. Restoring trunk's exact full-object comparator therefore fails `TestManagerConfigV2Vectors` on this row. The re-apply instruction requires the comparator and its regression test verbatim and says to change nothing else; reconciling the expected shape would exceed that boundary.

### Decision needed

Please resolve these two contract points before another code change:

- Does environments §12.1 admit valid unpadded OpenSSH key material as used in the pinned rc.13 valid fixtures? If yes, the review's F1 rejection needs to be superseded and the accepted implementation should retain an unpadded decode path with a named regression test. If no, the pinned rc.13 fixtures/cases need an upstream protocol correction or a different pinned conformance version; changing them in this leaf would invalidate the required conformance evidence.
- For manager-config vectors whose expected shape is the `environments` projection, may the conformance driver reconcile that projection while keeping the general comparator and extra-key regression strict, or must the pinned vectors provide the full normalized manager object?

Recommendation: keep the rev5 fixes as requested and leave this task blocked until the spec owner/orchestrator resolves the key grammar and the expected-object shape. No hosted gate or handoff was run because the required local conformance package is red; a red candidate must not be published as ready for review.


## Revision 5 — grammar-level signer keys, projected manager vectors

This section resolves the earlier revision 5 block using `1zgucp-decision-1.md`; that decision supersedes the two open questions in the preceding section.

### Findings addressed

- **F1, SSH signer grammar (§12.1):** manager config now accepts padded and unpadded standard Base64 material by checking the encoded form without decoding or validating the SSH wire blob. The published valid vectors carry unpadded material, and §12.1 defines the key as a public-key line whose identity is key type plus Base64 material. Cryptographic candidate verification remains the resolver's admission step, so malformed blobs cannot produce valid signature evidence. Config duplicate detection and resolver identity comparison normalize padding by removing trailing `=` from both sides; the comments cite §12.1. `FormatSigner` handles both forms while keeping key material out of status output.
- **F2, manager vector projections:** the trunk `exactManagerEffectiveJSON` comparator and `TestManagerEffectiveJSONComparisonRejectsExtraKnobs` are preserved unchanged for full-object comparisons. Vectors declaring a subset of top-level keys now compare exactly those canonical JSON values. `TestProjectedManagerEffectiveJSONRejectsExtraEnvironmentKnob` proves an extra key inside projected `environments` still fails.

Named production-entry regressions: `TestSourceSignerOpenSSHKeyMaterialGrammar` drives `config.Parse`, accepting opaque padded/unpadded Base64 and refusing non-Base64 characters and an unknown key type; `TestResolveMatchesUnpaddedAllowlistToPaddedVerifiedSSHKey` drives `contextresolve.Resolve` and proves padding/comment normalization.

### Conformance accounting

The carried conformance ledger moved from **69 data rows to 14** in revision 4; it remains at 14 data rows here. No passing row remains in the ledger. The pinned manager vector consumer reports **55 driven, 1 known gap, 0 bound, 0 skipped of 56 total**. The sole row is `schema2-registry-bootstrap-members` (owned by `STORY-260910-6bo7ej`); all signer cases and `schema2-empty-environments-defaults` pass. The pinned source-signer merge family remains six of six through `Load`; the resolver and CLI suites below exercise the production signer path and posture.

### Verification

| Command | Exit | Result |
|---|---:|---|
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/config` (final run) | 0 | Full config package and pinned subset green. |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test -v ./internal/config -run '^TestManagerConfigV2Vectors$' -count=1` | 0 | 55 driven, 1 known gap, 56 total. |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/contextresolve ./internal/contextlock` | 0 | Both packages green. |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run 'Status|Surfacing|Signer|Delta|Guarded'` | 0 | 207.555s. |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./cmd/curator -run '^(TestSourceSignerVectorsAtProfileInstall|TestRevisionDoesNotBorrowTagSignature|TestProfileUpdateSystemDeltaConfirmationGolden|TestProfileUpdateMCPDeltaConfirmationGolden|TestProfileUpdateAllDeltaVectorsAtCLI|TestProfileUpdateAllStopsAtFirstUnconfirmedDelta|TestProfileInstallReinstallDeltaVectorsAtCLI|TestProfileUpdateListsNewDeclaration|TestProfileUpdateSurfacesBeforePublication)$' -count=1 -timeout=360s` | 0 | 127.555s; includes the system and MCP confirmation goldens and `env status` posture assertions. |
| `go build ./...` | 0 | Build green. |
| `make lint` | 0 | `0 issues`. |
| `gofmt -d` on the five revision 5 Go files | 0 | No formatting diff. |
| `git diff --check` | 0 | Clean. |

The first full config run during this revision exited **1**: `base64.*Strict()` rejected the pinned unpadded key material, and the new grammar test used a noncanonical source identity. I replaced strict decoding with the spec-compatible standard/raw Base64 check and corrected the test identity; the final full config run above exited **0**.

### Narrowing mutants

All mutant source files were isolated under `/tmp/curator-1zgucp-mutants`, outside the worktree. Real exit codes:

| Narrowed gate | Test and result | Exit |
|---|---|---:|
| Removed Base64 grammar validation | `TestSourceSignerOpenSSHKeyMaterialGrammar` failed because the non-Base64 character was admitted. The first overlay attempt exited 1 at compile time due to an unused import and was not counted; after removing that import, this test-level mutant was killed. | 1 |
| Removed padding normalization | `TestResolveMatchesUnpaddedAllowlistToPaddedVerifiedSSHKey` failed with `context_source_signer_rejected`. | 1 |
| Projected comparison checked only top-level key presence | `TestProjectedManagerEffectiveJSONRejectsExtraEnvironmentKnob` failed because the narrowed comparator admitted the extra nested knob. | 1 |
| Full-object comparison ignored unexpected top-level keys | `TestManagerEffectiveJSONComparisonRejectsExtraKnobs` failed because the narrowed comparator admitted the extra manager knob. | 1 |

The previously accepted revision 3 mutation evidence for signer authorization, system/MCP delta classification, confirmation refusal, status posture, and update closure remains in the earlier results above; the prior record also notes that one attempted system CLI-golden mutant was interrupted and not counted (the system classifier mutant was killed). No `CHANGELOG.md` or `LOGBOOK.md` was edited. The release-prep entry remains in the section above.

The worktree still contains only the original 31 accepted revision 3 paths relative to trunk, with the five-file revision 5 delta above. It remains uncommitted. Hosted validation for this revision has not run yet; handoff will publish it and run that gate.


## Revision 6 — re-apply on trunk 86552087

Reapplied accepted rev5 (`refs/campaign/ioemse-rev5-20260927`, tree `1b69fd0c`, parent `d41da0fb`) onto the Story worktree at `86552087`. The candidate remains uncommitted. The final diff contains exactly the same 31 paths as rev5: no missing or extra paths. All 25 paths trunk did not touch remain byte-identical to rev5:

- `cmd/curator/profile.go`, `cmd/curator/profile_delta_confirmation_test.go`, `cmd/curator/profile_signers_test.go`, `cmd/curator/profile_surfacing_test.go`, `cmd/curator/testdata/profile-update-mcp-delta.golden`, `cmd/curator/testdata/profile-update-system-delta.golden`
- `internal/config/config.go`, `internal/config/environments.go`, `internal/config/environments_conformance_test.go`, `internal/config/environments_test.go`
- `internal/contextlock/contextlock_test.go`
- `internal/contextresolve/contextresolve.go`, `internal/contextresolve/contextresolve_test.go`, `internal/contextresolve/source_signers_conformance_test.go`
- `internal/envprofile/delta_confirmation_test.go`, `internal/envprofile/envprofile_policy_test.go`, `internal/envprofile/gitsource.go`, `internal/envprofile/overlays.go`, `internal/envprofile/pathkind_test.go`, `internal/envprofile/profiledelta.go`, `internal/envprofile/profiledelta_conformance_test.go`, `internal/envprofile/status_test.go`, `internal/envprofile/surfacing.go`, `internal/envprofile/surfacing_order_test.go`, `internal/envprofile/surfacing_test.go`

The six intersecting paths combine as follows:

- `.github/ci/conformance-gaps.tsv`: kept trunk's surviving rows and rev5's accepted removals. The manager-config consumer showed 20 carried gap rows now pass, so those rows were removed; trunk's two Codex seed gap removals were also kept. The ledger has 12 data rows after merge cleanup.
- `.github/ci/root-artifacts.tsv`: combined the signer/update-confirmation vectors with trunk's lifecycle and read-failure vectors; kept the Codex seed, passthrough, source-signer, and read-failure declarations for `internal/envprofile`.
- `cmd/curator/envstatus.go`: retained trunk's registry boundary posture output and rev5's signer allowlist and update-confirmation posture.
- `internal/contextlock/contextlock.go`: retained the trunk lock/state handling together with rev5's resolved lock delta support.
- `internal/envprofile/envprofile.go`: retained trunk's `stateread`-backed lock reads and read-before-write ordering, and rev5's path state-pin check, locked-source refresh, signer resolution, delta output, and confirmation gate. `updateLocked` carries both the file-read seam and confirmation options.
- `internal/envprofile/status.go`: retained trunk's unreadable-profile visibility and registry posture while adding rev5 signer and update-confirmation fields. `statusProfiles` loads an available lock through the existing `readLock`/`stateread` seam so source signer posture can be reported without dropping profiles whose lock cannot be read.

The rev5 review-verdict resource named in the appended Review Round Brief is misattributed: its contents are titled `TASK-260918-ryh3kw review verdict — rev4`, while the board identifies this task's rev5 verdict as accepted. The backup-record regression named in that text is present on the carried trunk and was verified here; no unrelated product paths were changed.

### Revision 6 verification

| Command | Exit | Result |
|---|---:|---|
| `go build ./...` | 0 | Build passed after the merge. |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/config` | 0 | Final full config package passed. Seven earlier invocations exited 1 while surfacing passing carried ledger rows; those exact rows were removed before the final green run. |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/config -run '^(TestManagerConfigV2SchemaCases|TestManagerConfigV2Vectors|TestSourceSignerMergeVectorsAtLoad)$' -count=1 -v` | 0 | Manager schema cases: 104 driven, 3 known gaps of 107; manager vectors: 55 driven, 1 known gap of 56; signer merge vectors: 6/6. Remaining gaps are owned by other stories. |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/contextresolve ./internal/contextlock` | 0 | Resolver and lock packages passed. |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run 'Status|Surfacing|Signer|Delta|Guarded'` | 0 | Passed in 86.623s, including the manager-owned read guard. |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./cmd/curator -run '^(TestSourceSignerVectorsAtProfileInstall|TestRevisionDoesNotBorrowTagSignature|TestProfileUpdateSystemDeltaConfirmationGolden|TestProfileUpdateMCPDeltaConfirmationGolden|TestProfileUpdateAllDeltaVectorsAtCLI|TestProfileUpdateAllStopsAtFirstUnconfirmedDelta|TestProfileInstallReinstallDeltaVectorsAtCLI|TestProfileUpdateListsNewDeclaration|TestProfileUpdateSurfacesBeforePublication)$' -count=1 -timeout=360s -parallel=1` | 0 | Named signer, status-posture, delta-golden, and production-entry tests passed in 71.612s. |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envmarker -run '^TestParseAuthoritativeEnvMarkerSchemaCases$' -count=1` | 0 | Trunk's two Codex seed marker gap removals remain valid. |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./cmd/curator -run '^TestEnvUnmanageBackupRecordVectors$' -count=1 -timeout=180s -parallel=1` | 0 | Both backup-record restore vectors run through `env unmanage --restore-backups`. |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test -overlay=/tmp/TASK-260916-1zgucp-mutants/absent-overlay.json ./cmd/curator -run '^TestEnvUnmanageBackupRecordVectors$' -count=1 -timeout=180s -parallel=1` | 1 | Expected mutant kill: narrowing absent-inventory handling caused the absent-record no-op vector to refuse. |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test -overlay=/tmp/TASK-260916-1zgucp-mutants/unreadable-overlay.json ./cmd/curator -run '^TestEnvUnmanageBackupRecordVectors$' -count=1 -timeout=180s -parallel=1` | 1 | Expected mutant kill: admitting an unreadable inventory made the unreadable-record vector mutate state instead of refusing. Mutant files stayed in `/tmp`. |
| `make lint` | 0 | `0 issues`. |
| `git diff --check HEAD` | 0 | Clean. |
| `git diff --name-only HEAD -- . ':!.task-board'` | 0 | Exactly the 31 accepted rev5 paths; no `CHANGELOG.md` or `LOGBOOK.md` path. |

The broad `go test ./cmd/curator -run 'Profile|Signer|Delta|Status'` command exited 1 after the 10-minute Go test timeout. Its broad mask selected many parallel package tests and stalled in `TestProfileInstallReinstallUsePreservesProjectScope` during a journal rename; the exact signer/delta subset above passed. The broad run is not claimed as green evidence.

The final conformance-gap ledger contains 12 data rows. The merged ledger had 34 rows before cleanup; 20 manager-config rows were removed after the production config suite proved those published cases pass, and the two Codex seed rows removed by trunk stayed removed after their schema consumer passed. No `CHANGELOG.md` or `LOGBOOK.md` was edited. The release-prep text remains in the section above. Hosted validation for this reapply is pending handoff; local results are not treated as hosted-gate evidence.
