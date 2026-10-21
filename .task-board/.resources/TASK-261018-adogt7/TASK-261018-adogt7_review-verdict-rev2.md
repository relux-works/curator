# TASK-261018-adogt7 — implement CIP-0021 skill directory: revision 2 review

Changes requested: one reproduced audit-cache bypass.

```json
{
  "verdict": "changes_requested",
  "revision": 2,
  "candidate_tree": "6c0d63c9f4f562e89b8164212f7dcd567a10f01f",
  "findings": [
    {
      "id": "audit-markdown-cache-reuse",
      "row": "identity",
      "invariant": "Changed Markdown detection must not reuse a pre-change allowance to bypass the selected-package network declaration gate (AC 4).",
      "mechanism": "internal/audit/audit.go:834 adds Markdown scanning but RulesetVersion at line 53 remains 4, identical to the CR base. The cache-hit early return at line 384 therefore skips the new detector. docs/profile-skill-directories.md incorrectly claims ruleset 4 prevents earlier allowances.",
      "reproductions": [
        {
          "test_file": "TASK-261018-adogt7_review-cache-probe-rev2.py (generates TestReviewPriorMarkdownAllowance in a test-only Go overlay)",
          "commands": [
            "python3 .temp/review-adogt7/cache-probe.py",
            "REVIEW_CACHE_ROOT=\"$PWD/.temp/review-adogt7/cache-fixture\" REVIEW_PHASE=seed go test -overlay .temp/review-adogt7/cache-old.json -count=1 -v -run ^TestReviewPriorMarkdownAllowance$ ./internal/audit",
            "REVIEW_CACHE_ROOT=\"$PWD/.temp/review-adogt7/cache-fixture\" REVIEW_PHASE=probe go test -overlay .temp/review-adogt7/cache-current.json -count=1 -v -run ^TestReviewPriorMarkdownAllowance$ ./internal/audit"
          ],
          "expected_failure": "Seed using the exact base detector exits 0 through public Gate. Exact candidate Gate on a fresh home blocks audit.capability.network-undeclared; the same candidate Gate with the genuinely generated old cache returns no error. Probe exits 1: prior ruleset-4 allowance bypassed current Markdown network gate.",
          "logs": "TASK-261018-adogt7_review-evidence-rev2.log: cache-seed.log and cache-current.log",
          "actual_exit_codes": [
            0,
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
    "No prior substantive reviewer verdict exists for revision 1; it became stale before review.",
    "No supplied catalog surface table or wall-clock budget; reviewer-derived sweep is documented separately and is not exhaustive.",
    "Full hosted gate reused and independently confirmed successful on signed commit 3662b2753cffc3b2c3a478b75d58fe3dd6fe9d99, whose tree equals the exact CR candidate. No full local suite replay is claimed: current producer local evidence is scoped; historical full-local evidence is from older trees.",
    "README.md and CHANGELOG.md equal the CR base. Fresh upstream path intersection is empty. Temporary fixture homes only; no product files were edited."
  ]
}
```

# Review sweep: TASK-261018-adogt7 revision 2

Candidate tree: `6c0d63c9f4f562e89b8164212f7dcd567a10f01f`.
Base: `5ecb68da8f8ceab4dda118f2263abfb3da63ab8c`.
Fresh upstream: `f9811b67257b0d8d1a81176346502e2f35dc3ac5`; changed-path intersection is empty.
No producer surface table or time budget was supplied. This reviewer uses the relevant catalog rows below, supplemented by the AC table; no complete catalog coverage is claimed.

| Row | Result | Executed attacks |
| --- | --- | --- |
| read boundary | held | Install lifecycle vectors: escaping parent components, absolute path, symlinked component, inside/outside symlinked leaf, missing directory. All named refusals and no-publication assertions pass. |
| selector paths | held | Install vectors: selected versus omitted root; missing selected manifest/SKILL.md; alternate directory; schema-1 directory refusal via ParseManifest. |
| text grammar | held | 56 schema vectors plus closed-successor negative controls through ParseManifest/Parse; malformed selected metadata through Install. |
| identity | broken | Published canonical directory identities and sibling/selected byte-change vectors pass; old-detector cache identity fails on upgrade through public audit.Gate (finding audit-markdown-cache-reuse). |
| error surface | held | Install vectors reject missing directories/content and symlink paths with the named diagnostics; missing content never falls back to repository root. |
| destination conflicts | held | Install agreement diamond refuses different sources, different directories and omitted-versus-selected selectors; agreeing requirers unify. UseWithPolicy and Resolve retain selected targets. |

Free hunt: pre-upgrade cached allowance bypass reproduced through public audit.Gate; no other confirmed finding. Every one of the ten candidate narrowing mutants was rerun and killed by its named behavioral assertion. These results do not prove absence of other defects or cover all families in any catalog row.


## Coverage and validation

AC driving map: **8 of 9 AC rows have named committed behavioral tests**; AC 9 is validation-command evidence, not a product test. All nine rows have an evidence route; full local replay on this exact tree remains unverified here.

| AC | Named driving tests / entry |
| --- | --- |
| 1 | TestSkillDirectorySchemaVectors and ClosedSuccessors -> ParseManifest |
| 2 | TestSkillDirectoryLifecycleVectors disagreement/agreeing cases -> Install -> Resolve; TestSkillDirectoryAgreement -> Resolve |
| 3 | TestSkillDirectoryLockIdentityVectors -> Parse / Canonical / Hash; lifecycle install emits schema 4 |
| 4 | Lifecycle selected-root-undeclared-contact, sibling-edit and selected-edit -> Install/Update -> auditAndStore -> strictAuditMember -> audit.Gate; reviewer warm-cache probe fails |
| 5 | Successful lifecycle vectors -> UseWithPolicy and managed Resolve; assert selected targets |
| 6 | Lifecycle parent, absolute, symlink and missing-directory cases -> Install -> gitManager.Manifest / Extract |
| 7 | Lifecycle legacy-migration-keeps-root -> EnsureDefaultWithPolicy |
| 8 | TestSkillDirectoryNarrowingMutants -> behavioral child suites; 10/10 intended kills freshly reproduced |
| 9 | Hosted run 39186903815 success on exact signed candidate; scoped local reruns below; full local exact-tree replay not asserted |

Reviewer reran: complete contextpkg/contextresolve/contextlock suites (exit 0); 24 lifecycle and 6 metadata cases (exit 0); all 10 narrowing plants (outer exit 0, all intended child exits 1). Four vendored JSON files equal the local pinned spec corpus. All 17 candidate paths equal the frozen candidate blobs. No existing product source or test file changed; test overlays and logs stay in the run worktree.

Hosted receipt: https://github.com/relux-works/curator/actions/runs/39186903815 . Signature verification reports G; its tree is 6c0d63c9f4f562e89b8164212f7dcd567a10f01f. Two workflow-policy jobs are skipped; they are not described as executed.

## Required rework

Invalidate allowances produced before Markdown detection (use a new detector ruleset identity or equivalent fail-closed cache versioning), update the documentation claim, and retain a named regression exercising public Gate with a real pre-upgrade cache plus fresh-cache and declared-host controls. Add a narrowing mutant that restores acceptance of only the old Markdown allowance and require that regression to fail. Do this in the same leaf; no separate harness/research task is needed. Preserve all existing negative tests.

## Logbook

The candidate correctly detects the undeclared Markdown contact on a cold cache. The upgrade path reuses a legitimate allowance from the current base because both detectors advertise ruleset 4. Existing ten mutants all die, but none tests this old-detector/new-detector transition. This is a bypass path around the check, not forged evidence and not a test setup failure.
