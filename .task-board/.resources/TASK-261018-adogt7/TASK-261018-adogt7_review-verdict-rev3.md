# TASK-261018-adogt7 — implement CIP-0021 skill directory: revision 3 review

Changes requested: one reproduced numeric-type admission bypass.

```json
{
  "verdict": "changes_requested",
  "revision": 3,
  "candidate_tree": "d29170ea6bcef3309d81b7f79b9c856e3e596965",
  "findings": [
    {
      "id": "lock-numeric-discriminator-coercion",
      "row": "text grammar",
      "invariant": "The newly supported schema-4 lock must enforce the numeric discriminator and member weight types of the complete normative carrier (AC 3; rc7-carriers section 1).",
      "mechanism": "internal/contextlock/contextlock.go:438-439 and :452 decode numeric fields into json.Number, which accepts JSON strings; Atoi/ParseInt at :461, :467 and :491 erase that type distinction. ReadWith calls Parse at :537 after canonical JSON validation, which does not establish carrier types.",
      "reproductions": [
        {
          "test_file": "TASK-261018-adogt7_review-numeric-probe-rev3.py (test-only overlay: TestReviewSkillDirectoryNumericDiscriminants and TestReviewSkillDirectoryStringWeight)",
          "commands": [
            "python3 .temp/review-r3/TASK-261018-adogt7_review-numeric-probe-rev3.py",
            "go test -overlay .temp/review-r3/numeric-overlay.json -count=1 -v -run '^TestReviewSkillDirectory(NumericDiscriminants|StringWeight)$' ./internal/contextpkg"
          ],
          "expected_failure": "Exit 1 with three assertions: schema-4 Read admitted string schema_version, hash_version and weight. Each string negative has a passing numeric positive control. Actual candidate accepts canonical on-disk malformed carriers via contextlock.Read, not just a standalone helper.",
          "logs": "TASK-261018-adogt7_review-evidence-rev3.log: numeric.log; independent probe replay also exits 1",
          "actual_exit_codes": [
            0,
            1
          ]
        }
      ],
      "severity": "bypass",
      "repeat-of": "none"
    }
  ],
  "notes": [
    "Previous finding audit-markdown-cache-reuse from revision 2 is fixed. Genuine pre-Markdown cache regression passes both hash versions and public gates; restoring ruleset 4 is killed for the intended bypass.",
    "No supplied surface table or wall-clock/free-hunt budgets. Retained six-row revision-2 sweep, with explicit limited attack claims; no proof of absence.",
    "AC driving ratio: 8 of 9 rows have named behavioral tests. AC9 is a command-evidence obligation: exact-tree hosted full CI is green, but a full LOCAL suite on this exact tree remains unverified. Producer r7 evidence is scoped and explicitly does not reuse old full-suite receipts. Do not represent the post-handoff remote-only command as local full-suite execution. Supply the missing exact-tree local receipt before acceptance.",
    "Root README and CHANGELOG equal trunk per later round instructions; no real operator HOME changes. Production files and real Git index were not modified by review. All fixture homes are temporary.",
    "Fresh origin/main equals CR base e88fdf2d. All 21 candidate working-file blobs equal frozen tree d29170ea; signed hosted commit 4b6bab1476808f209ebd07b4edbe0f7f071d225e has that exact tree and signature G. Hosted run 39194686143 independently reports success, 20 jobs successful and two workflow-policy skips. No hosted result is synthesized.",
    "This is a new mechanism, not a repeat of revision-2 cache reuse. The numeric decoder shape predates schema 4, but this change newly admits schema 4 under the complete-carrier contract; fix the scoped reader without weakening historical guarantees."
  ]
}
```

# Revision 3 review sweep

Candidate d29170ea6bcef3309d81b7f79b9c856e3e596965; base and freshly fetched main e88fdf2dbec7f94c48b0232a350a0a9f9484fdeb. All 21 candidate blobs match working files. No supplied surface table or budgets: retain revision-2 six-row sweep; no exhaustive catalog claim. AC coverage: 8/9 named behavioral test rows; AC9 requires command evidence, exact-tree full local suite not yet established. Cache regression attacked first: TestReviewPriorMarkdownAllowance passes both public gates, both framings, genuine old cache and fresh positive controls.

| Row | Result | Evidence |
|---|---|---|
| read boundary | held | boundary.log: seven escape/missing/symlink lifecycle attacks through Install, exact diagnostics and preserved publication state. |
| selector paths | held | lifecycle.log: 24/24 vectors, selected/root omission, missing content no fallback, update/sync and migration; Install/UseWithPolicy/Resolve. |
| text grammar | broken | readers.log published 56 vectors pass; numeric.log public contextlock.Read admits string schema/hash discriminants. Finding lock-numeric-discriminator-coercion. |
| identity | held | readers.log: both independent lock identities; lifecycle.log: selected edits change identity, sibling edits do not. cache.log and cache-mutant.log verify prior finding fixed and intended mutant failure. |
| error surface | held | metadata.log: six valid/invalid metadata controls; lifecycle.log: absent directory/content and symlink refusal through Install, no fallback or publication. |
| destination conflicts | held | destinations.log: NativeSkillsLinkedParent, NativeSkillsBackupCollision, ForeignSymlinkStopsSwitch through production materialization; lifecycle agreement conflicts also refuse. |

Free hunt: neighboring member weight uses the same json.Number coercion and is also admitted as a string (numeric.log). Merged into the same mechanism; no second finding. All 11 narrowing mutants rerun, intended child exit 1 and outer exit 0. Numeric probes change tests only through a Go overlay; exact production candidate blobs remain unchanged.


## Coverage and evidence

| AC | Named test / production call |
|---|---|
| 1 | TestSkillDirectorySchemaVectors / ClosedSuccessors -> ParseManifest -> parseRequirements |
| 2 | Lifecycle agreement cases and TestSkillDirectoryAgreement -> Install -> contextresolve.Resolve |
| 3 | Schema and LockIdentity vectors -> contextlock.Parse/Hash; new negative probes -> contextlock.Read -> ReadWith -> Parse (broken) |
| 4 | Lifecycle selected contact and sibling/selected edits -> Install/Update -> auditAndStore -> strictAuditMember -> audit.Gate; TestReviewPriorMarkdownAllowance -> Gate/GateReadOnly |
| 5 | Successful lifecycle cases -> UseWithPolicy and managed Resolve -> skillsOf -> packageRoot |
| 6 | Lifecycle traversal and symlinks -> Install -> gitManager.Manifest / Extract / selectedPackageRoot |
| 7 | legacy-migration-keeps-root -> EnsureDefaultWithPolicy |
| 8 | TestSkillDirectoryNarrowingMutants -> production behavioral child tests |
| 9 | Exact-tree hosted receipt independently confirmed; exact-tree full local suite remains unverified, as explained above |

Fresh reviewer runs: complete contextpkg/contextlock/contextresolve suites; 24/24 lifecycle vectors; 6/6 metadata cases; 11/11 intended mutant kills; cache upgrade regression and its controls; destination-link/conflict tests. Every listed unmutated suite exited 0. Numeric negative probes exited 1 with the intended assertions. Four vendored JSON files match the locally pinned normative corpus. The 56 structural cases and two identity cases pass; the new probes expose a class those cases omit.

Hosted receipt: https://github.com/relux-works/curator/actions/runs/39194686143 . Reviewer reused this exact-tree hosted result, not historical hosted runs. No full local replay is claimed. No lint/build replay was needed to reproduce the reader defect; r7 producer's scoped build/vet/lint receipt remains separate evidence.

## Required rework

Reject JSON strings in numeric lock fields before converting their values. Preserve numeric positive controls and existing schema-1/2 compatibility. Commit named regression tests for schema_version, hash_version and member weight through the production lock reader, paired with a narrowing mutant admitting only a string-valued member of this class. Keep the warm-cache regression and all existing tests. Complete the missing exact-candidate full local validation evidence, then hand off for a new review cycle.

## Logbook

Ruleset 5 repairs the previous upgrade-only bypass. The sweep found a distinct numeric-looking-scalar coercion: canonical JSON is not schema validation, and Go json.Number accepts quoted digits. ReadWith returns a usable lock and raw-byte hash for these malformed carriers. One mechanism, three reproduced inputs, repeat-of none. Review writes only scratch artifacts and task-board resources; no control-root product files are edited.
