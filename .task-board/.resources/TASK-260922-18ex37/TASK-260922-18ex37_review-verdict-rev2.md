# TASK-260922-18ex37 review verdict — CR rev2 (tree a1448c09): CHANGES REQUESTED

## Checks that passed
- Item 0: I split rev1 and rev2 by file and compared them. The only file that differs is the removed stray `TASK-260922-3bbvrs_results.md`; the other 58 file diffs are byte-identical. There are no stray root TASK-/BUG-/STORY- files in a1448c09.
- Item 1: the pin is dcc7f015 everywhere: ci.yml:44 SPEC_PIN, DRAFT_SOURCES_PIN, draftsources_corpus_test.go:19, draft README, conformance-case-counts.tsv header and CHANGELOG:9. `dced9b8` now appears only in a CHANGELOG history line.
- Item 4: gate run 35962642338 pushed commit 5d934556, whose tree is a1448c09 (the candidate tree). Every required lane passed: Test ubuntu/macos/windows, Race, Lint, Interop, Gate self-test x3, Naming. Candidate-suite and rose-air were skipped and remain unverified.
- The ledger is enforced through `conformancecoverage.RunOutcomes`. I re-ran `TestManagerConfigV2SchemaCases` myself against spec dcc7f015; it passes and reports "78 driven, 29 known-gap, 0 skipped, 107 total".

## F1 (BLOCKING, AC2 / brief "verify each mapping against the spec commit that published the case") — wrong owner on the path-kind overlay rows
Location: `.github/ci/conformance-gaps.tsv`, all `valid-overlay-*` manager-config-v2 schema-case rows (14 rows) and all `schema2-overlay-*` vector rows (14 rows), owner `STORY-260916-wgt8vz`, reason "Load rejects the valid path-kind and overlay source case published by the path-admission spec change".

Evidence:
- The overlay cases were published by spec bd39adb, which is before the old pin. They already existed at dced9b8 (`git cat-file -e dced9b8:conformance/v1/schema-cases/manager-config-v2/valid-overlay-path-source.json`).
- The base driver `TestManagerConfigV2SchemaCases` (at 1511b345) had no skip list, so curator already drove these cases green at rc.12. Path-kind overlays are therefore already implemented.
- Between the two pins, only 684c9f1 (E1: `source_signers` / `require_source_signers`) and 937e795 (0017/0018: `permissions`) touched these files; `git diff dced9b8 dcc7f015` shows exactly those added fields.
- I probed the real `config.Load` in a disposable clone (non-test config code is identical to the candidate) against spec dcc7f015:
  - valid-overlay-path-source.json: `environments: has unsupported field "source_signers"`
  - valid-overlay-git-scp.json: `unsupported field "permissions"`
  - valid-overlay-path-windows-slash.json: `unsupported field "permissions"`

Implementing wgt8vz would clear none of these rows. Each row is owed by STORY-260916-ioemse (E1) or STORY-260922-1cenbr (0017/0018), whichever unsupported field fails first; a row needing both should name both, or name the first blocker. The reason text is also factually wrong. Apply the same check to the `schema2-overlay-*` vector rows: their expected EffectiveJSON changed under 937e795 / 684c9f1, not under a path-kind change.

Fix: for each overlay row, re-derive the owner from the actual Load/Parse error. Correct the owner and reason in conformance-gaps.tsv, the results table and the CHANGELOG/docs attribution. After the fix, STORY-260916-wgt8vz should own no rows unless a path-kind failure actually occurs. Also sweep the other rows the same way (probe the real error, compare it with the commit that last changed the case). I spot-checked E3 (4a2fa3e), S2 (1e73c03), E1 (684c9f1) and S1+S3 (5146c7b); those publish commits match their owners.

## Not blocking, noted
- Five marker-v4 rows belong to BUG-260923-2afgyq (backlog). They predate this leaf and are the same cases recorded in LOGBOOK line 430 → acceptable.
- scriptpolicy conformance_test.go:63 classifies `executable_identity_cases` as refusedBeforeReached (verified in rev1; unchanged in rev2).
