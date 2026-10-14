# BUG-261014-351fcr — migrated-global-skill-audited-under-v1-hash — reviewer A

Recommendation: accept the scoped BF1 fix. RECORD-ONLY: reviewer B owns the CR verdict. No accept_cr, reject_cr, done transition, commit, or repository source edit was performed. The specific gw-BF1-review-A brief overrides the generic verdict-routing instructions for this run. The run goal query reported no bound goal.

Candidate: CR-BUG-261014-351fcr-1 revision 1; tree `3524e8ad403eb248181e8cd526033fdeb9bc46ca`; base `84cf54b94670ae9881f91b1c147feb4b75afddee`. Fresh remote main advertisement equals the base. HEAD is zero commits behind local main. Both changed worktree files matched candidate blobs before and after execution. No production files were modified: independent mutants ran via Go overlays under ignored task scratch.

```json
{
  "mode": "record_only",
  "recommendation": "accept",
  "findings": [],
  "notes": [
    "The fixture parses a branch declaration, then explicitly pins its commit as revision before migration. It does not demonstrate direct live-branch migration. Existing production deliberately skips branch declarations; source-selection changes are outside this two-call-site BF1 fix. This bound is explicit in the gaps artifact.",
    "No surface table or wall-clock budget was supplied. This review derived the table below from all four AC rows and the five reviewer checks; it does not claim catalog-wide completeness.",
    "The network-undeclared outcome is confirmed and precisely recorded for a separate producer leaf; it is intentionally not remediated here.",
    "Hosted Candidate suite and Test (rose-air) jobs were skipped, not passed."
  ]
}
```

## Coverage checked before implementation inspection

**4 of 4 AC rows driven**, within the explicitly stated revision-pinned legacy-source bound. `TestMigratedGlobalSkillBinaryWriteVersion` drives AC1, AC2 and AC3. `TestMigratedGlobalSkillNetworkUndeclaredFollowUp` drives AC4. Production entry is `EnsureDefaultWithPolicy -> ensureDefault -> migrateGlobalSkills`; both direct `contextaudit.DetectAtVersion` and `strictAuditMember -> audit.Gate` are reached. `internal/envprofile/global.go` also invokes that public migration entry. Lock assertions consume the published lock through `readLock` and validate schema/hash 2, source/commit identity and unchanged binary bytes.

## Swept surface table

Results were appended to the scratch sweep artifact before advancing to each next row. Held means only the named attacks did not expose a defect.

| Row | Result | Attack and observed result |
| --- | --- | --- |
| AC1: both audit version selections | held | Independently changed direct detector and migrated Resolved.HashVersion to V1. Each overlay failed the production migration test under v2: opaque-file and audit.opaque.nul-byte respectively. |
| AC1: serialized lock version | held | Removed hash_version from serialized lock, retaining schema 2. Test failed reading the published lock: schema_version 2 requires hash_version 2. |
| AC2/AC3: forced-v1 opaque refusal regardless of directory | held | Baseline exercised gzip-only, binary-only and combined files in two directories, both versions. Narrowed scanner to exempt only .task-board/attachments/attachment.gz. Gzip-only/v1 failed with unexpected nil migration error. The remaining binary still blocked combined/v1, demonstrating the importance of the isolated-file assertion. |
| AC4: network follow-up | held | Baseline reproduces the exact network refusal. Narrowed the detector to admit only downloads.example.com while preserving URL detection and the network-undeclared token. Behavioral migration test failed with unexpected nil error. |

Free hunt: repeated the behavioral matrix with the supported tag declaration instead of revision, via a test-only overlay. All cases passed, including forced-v1 refusal and network refusal. No additional findings.

## Implementation fit and isolation

The production diff has exactly two changes: direct migration detector selects `hashing.WriteVersion()` and migrated `Resolved.HashVersion` carries the same selection into the strict audit. The normal resolver uses the existing `hashing.WriteVersion()` rule. No alternate rule, path exemption or capability relaxation is added. Lock canonicalization already supplies schema/hash 2 for current writers; the regression test checks the serialized outcome.

Reviewed local spec core section 8: the v1 NUL rule applies at every depth regardless of name; v2 treats NUL as ordinary data. Environments section 9.1 requires strict audit of every member. Both remain enforced.

Fixtures use temporary repository/home directories, local bare remotes, test-owned Git configuration, explicit Git identities and example.com identities rewritten to file remotes. Package TestMain isolates machine/Git configuration. Tests call EnsureDefaultWithPolicy with an explicit policy. No live manager home is accessed by the changed path; no personal paths or private identifiers occur in the two-file delta.

## Own execution and reused evidence

Own execution on Go 1.25.5, darwin/arm64:

- `go test ./internal/envprofile/... -run '^TestMigratedGlobalSkill(BinaryWriteVersion|NetworkUndeclaredFollowUp)$' -count=1 -v`: exit 0; all 12 binary/path/version cases plus network follow-up passed.
- Five independent overlay mutants: each exited 1 with the intended behavioral assertion failure, never a compile failure. `attacks.py` in the evidence bundle constructs each overlay from the exact candidate tree and records its command.
- Tag-fixture overlay of the same focused command: exit 0.
- `git diff --check`: exit 0. Final candidate-byte equality assertions: passed.

The public CLI binary itself was not launched. Full package/build/vet were not redundantly rerun; their producer evidence is recorded in BUG-261014-351fcr_gaps.md. No local full-suite execution occurred (R223).

Hosted gate was independently read through `gh run view 38517800551 --repo relux-works/curator --json headSha,status,conclusion,url,jobs`: completed/success, 20 successful applicable jobs, 2 explicit skips. Hosted commit `f34a0db8be5b13282f7bb4742b5d8a21ead874cf` resolves locally to the exact candidate tree above, so code/tests/workflow configuration identities match. Hosted platform execution is provider evidence, not claimed as local replay. The complete producer remote-gate log terminates in the reported results and is cited by the gaps file.

Hosted evidence: https://github.com/relux-works/curator/actions/runs/38517800551

## Follow-up / task-scoped logbook record

Independently reproduced:

`profile_source_invalid: migrated global skill "binary-skill" audit blocked: audit blocked: binary-skill: high audit.capability.network-undeclared - contacts downloads.example.com, which no capabilities.network entry covers (file: scripts/fetch.sh)`

The gaps artifact names `TestMigratedGlobalSkillNetworkUndeclaredFollowUp` and the exact finding. `strictAuditMember` still passes `capabilities.ImplicitNone()`. The separate capability-contract follow-up remains for the orchestrator; BF1 does not silently fix it.

Evidence bundle: `BUG-261014-351fcr_review-A-evidence.txt`. It contains raw baseline/mutant/tag output, incremental sweep, reproducible overlay driver and provider result. Host-local temporary path prefixes are redacted; source and failure text are retained.
