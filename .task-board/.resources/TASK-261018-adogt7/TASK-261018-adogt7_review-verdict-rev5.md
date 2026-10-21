# TASK-261018-adogt7 — implement CIP-0021 skill directory: revision 5 review

Accepted; not landed. Repository delta is present (22 paths).

```json
{
  "verdict": "accepted",
  "revision": 5,
  "candidate_tree": "e3e9a7107bf3576856e82546f41bb4c892bb9ca6",
  "findings": [],
  "notes": [
    "Previous revision-3 lock-numeric-discriminator-coercion is fixed. Raw JSON tokens preserve the distinction; all 77 numeric controls/refusals through Read pass. Schema-4 string-zero-weight narrowing mutant fails the named regression for the intended reason.",
    "Previous revision-2 audit-markdown-cache-reuse remains fixed: Gate and GateReadOnly refuse genuine prior-ruleset allowances; restoring ruleset 4 is killed.",
    "No supplied surface table or wall-clock/free-hunt budgets. Retained six-row prior-review sweep, each held only for named executed attacks; no exhaustive absence claim.",
    "8 of 9 AC rows driven by named behavioral tests; AC9 is the stated command-evidence obligation. Exact-tree full local receipt verified and reused; exact signed hosted gate independently confirmed. All nine AC obligations satisfied under these bounds.",
    "Revision 4 and 5 republish the unchanged round-8 tree after infrastructure recovery; this review addresses the three-file fix delta from last reviewed revision 3, not invented new rework.",
    "Root README and CHANGELOG equal fresh trunk under later briefs. Reviewer changed only scratch artifacts and board resources. No real HOME product state was touched; tests use temporary homes."
  ]
}
```

# Revision 5 review sweep

No supplied surface table or wall-clock/free-hunt budget. Retain the six catalog rows from previous review, with bounded named attacks.

| Row | Result | Evidence |
|---|---|---|
| read boundary | held | lifecycle.log exit 0: parent prefix/component, absolute, missing directory and inside/outside/component symlinks through Install; exact refusals and no-publication assertions. |
| selector paths | held | lifecycle.log exit 0: all 24 normative vectors through Install/UseWithPolicy/Resolve, including root omission, selected content absence, update/sync and unchanged migration. |
| text grammar | held | readers.log exit 0: all schema vectors, closed successors and numeric on-disk controls/refusals; previous numeric coercion finding no longer reproduces. |
| identity | held | readers.log and lifecycle.log: independent canonical identities, selected edits change hash, sibling edits preserve it. cache.log: old detector allowances refused through Gate/GateReadOnly, fresh controls accepted. |
| error surface | held | metadata.log: all six controls through Install; lifecycle.log absent content and unsafe paths refuse with specified diagnostics, preserving publication state. |
| destination conflicts | held | destinations.log: linked parent, backup collision and foreign symlink attacks through production materialization refuse as intended. |

Free hunt: neighbors.log, TestReviewNumericRepairNeighbors through contextlock.Read: ten malformed closed/type neighbors refused, numeric positive control accepted; input bytes preserved. No new finding. All twelve narrowing plants killed by intended behavioral assertions (mutants.log).


## Exact candidate and freshness

Fresh authenticated fetch confirms origin/main equals CR base `e88fdf2dbec7f94c48b0232a350a0a9f9484fdeb`; no upstream path overlap. All 22 changed working-file blobs equal candidate tree `e3e9a7107bf3576856e82546f41bb4c892bb9ca6`. Only three files differ from reviewed revision 3: numeric reader, named numeric regression suite, and narrowing mutant. These were inspected and reader tests executed first. No product edits, commits, branch switches or index changes were made by review.

## Coverage

**8 of 9 AC rows driven by named behavioral repository tests; 1 of 9 is a verified command obligation.** This is not a claim that a prose receipt is a driving test.

| AC | Driving tests / production entry |
|---|---|
| 1 | TestSkillDirectorySchemaVectors, TestSkillDirectoryClosedSuccessors -> ParseManifest -> parseRequirements |
| 2 | TestSkillDirectoryAgreement and lifecycle conflict/unification cases -> Install -> contextresolve.Resolve; selected metadata -> gitManager.Manifest -> manifest.ReadSkillPackage |
| 3 | TestSkillDirectoryLockIdentityVectors, schema vectors, TestReviewSkillDirectoryNumericDiscriminants/StringWeight and TestLegacyLockNumericCarrierTypes -> contextlock.Read/Parse/Hash |
| 4 | Lifecycle selected contact and sibling/selected edits -> Install/Update -> auditAndStore -> strictAuditMember -> packageRoot -> audit.Gate; TestReviewPriorMarkdownAllowance -> Gate/GateReadOnly |
| 5 | Lifecycle successful selected installs/sync -> UseWithPolicy/Resolve -> skillsOf -> packageRoot; both native and managed targets asserted |
| 6 | Lifecycle parent-prefix/component, absolute, missing-directory and three symlink variants -> Install -> gitManager.Manifest -> Extract/selectedPackageRoot |
| 7 | Lifecycle legacy-migration-keeps-root -> EnsureDefaultWithPolicy |
| 8 | TestSkillDirectoryNarrowingMutants -> behavioral child suites, 12 intended failures; temporary homes; four vendored JSON files match pinned CIP-0021 corpus |
| 9 | Complete unchanged round-8 local receipt plus fresh exact-signed-candidate hosted gate, below |

Fresh reviewer execution: complete contextpkg/contextlock/contextresolve suites; 24 lifecycle vectors; 6 metadata cases; cache upgrade regression; 3 destination safety tests; all 12 narrowing mutants; 11 neighboring malformed-carrier/positive-control cases. All unmutated commands exit 0. Mutant harness exits 0 only after each child exits 1 with its intended named assertion. No static-check-only mutant result is accepted. Production call sites were inspected: selected audit roots in envprofile.go and managed skill targets in managed.go use packageRoot(entry, directory).

Commands (logs in TASK-261018-adogt7_review-evidence-rev5.log):

- `go test -count=1 -timeout=2m -v ./internal/contextpkg ./internal/contextlock ./internal/contextresolve`
- `go test -count=1 -timeout=5m -v -run '^TestSkillDirectoryLifecycleVectors$' ./internal/envprofile`
- `go test -count=1 -timeout=2m -v -run '^TestReviewPriorMarkdownAllowance$' ./internal/audit`
- `go test -count=1 -timeout=3m -v -run '^TestSkillDirectoryMetadata$' ./internal/envprofile`
- `go test -count=1 -timeout=3m -v -run '^Test(NativeSkillsLinkedParent|NativeSkillsBackupCollision|ForeignSymlinkStopsSwitch)$' ./internal/envprofile`
- `go test -count=1 -timeout=8m -v -run '^TestSkillDirectoryNarrowingMutants$' ./internal/envprofile`
- `go test -overlay .temp/review-r5/neighbors-overlay.json -count=1 -timeout=2m -v -run '^TestReviewNumericRepairNeighbors$' ./internal/contextpkg`

The neighbor overlay replaces only lock_numeric_carrier_test.go with the attached test file; production blobs are unchanged. Its ten negatives cover unknown top/member/precedence fields and null required fields through on-disk Read, with a valid positive control and preserved input bytes. Fixture-dependent commands use producer's documented temporary tool/conformance paths and child-only Git/PATH isolation; no host installation changes.

## Reused validation, not reviewer replay

Attached `TASK-261018-adogt7_r8-validation.json` and `.log` were downloaded afresh. All 1985 file hashes match current source/test/configuration bytes; scope hash `52db2b21092c2969c7f0f2a1279358e62f9b7600d48784b4a6fa6a96db72a8f4`. Every one of 71 exit-0 command receipts has its complete raw log present with matching SHA-256; complete sanitized attached log hash also matches. That suite covers 85 unique packages, with heavy-package top-level masks partitioned as recorded; platform/capability skips retain their stated bounds. Build/vet/lint receipts remain separate recorded evidence. The reviewer did not replay the full suite or claim remote execution was local. Toolchain remains Go 1.25.5 on darwin/arm64.

Hosted run https://github.com/relux-works/curator/actions/runs/39507549530 independently reports success: 20 successful jobs, 2 workflow-policy skips. Its head `e5d5d5c6bc7b34bc68faffa232f7bc40c205cc68` has the exact CR tree and verified signature `G`. The revision-5 validation log terminates with `[exit 0]`; previous runner-capacity failures are not reused as passes. Fresh hosted JSON is attached as TASK-261018-adogt7_review-hosted-rev5.json.

## Logbook

The previously rejected numeric decoder now rejects quoted digits without conflating them with numbers. The narrowed schema-4 zero-weight mutant demonstrates the regression test reaches the real Read path. Neighboring closed-carrier attacks found no additional issue. Prior Markdown cache invalidation still holds. Review is acceptance evidence for this exact revision only, not a landing or release claim. Run goal read returns no bound goal; no directives were recorded. Record acceptance through accept_cr revision 5.
