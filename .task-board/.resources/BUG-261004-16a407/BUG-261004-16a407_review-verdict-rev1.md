# BUG-261004-16a407 — trust-pin-overrides-strict-findings: revision 1 re-review

Verdict: accepted. Scope: delta re-review per N5-rereview-note.md. Both prior findings R1 (missing CR) and R2 (older-tree hosted evidence) are resolved. No implementation finding remains. This is acceptance for integration, not a claim of landing.

## Revision identity and scope

The board worktree-status query confirms CR-BUG-261004-16a407-1 revision 1 is ready, with base a2df58875e8a7262923064b2b871a31127e155e4 and candidate tree c425f75c9785d6bc45543e8798c84698c7ac3f40. Its validation resource names hosted run 37203454130 and reports exit 0.

Read-only comparison executed:

    git diff --exit-code cc3ba67211bc533558c3a6d0fe8bb6bce2ad7a2f c425f75c9785d6bc45543e8798c84698c7ac3f40 -- internal/audit/audit.go internal/audit/audit_test.go internal/install/draftaudit_test.go

Result: exit 0, empty output. The three candidate files are byte-identical to the earlier reviewed snapshot. Base-to-candidate diff lists exactly those three paths (125 insertions, 9 deletions). No LOGBOOK.md, CHANGELOG.md, or unrelated paths changed. Diff whitespace check exited 0. Reviewer changed no product code.

Hosted commit d44b37fbfbb42e9a99a6ac052a07645ffb5a8d3b has tree c425f75c9785d6bc45543e8798c84698c7ac3f40 and parent a2df58875e8a7262923064b2b871a31127e155e4 (git show exit 0). Thus the hosted evidence binds exactly to this revision and its recorded current base, including the upstream changes missing from the former evidence.

## Independently downloaded hosted evidence

Run: https://github.com/relux-works/curator/actions/runs/37203454130

GitHub reports completed/success: 11/11 required jobs green, including lint and interop. Two optional jobs (Candidate suite and rose-air) were skipped; no coverage claim for them. gh run download and gh run view --log both exited 0.

Read the downloaded go-test.json for every Test/Race lane and asserted these leaf counts, parent-test passes, both audit/install package passes, and zero failing Go events:

| Lane | TestGatePinPolicyFreshAndCached | TestDraftAuditPinDoesNotWaiveStrictFindings | TestGateEmitsScriptAuditLabels | Go test / platform gate exit |
| --- | --- | --- | --- | --- |
| Test Ubuntu | 52/52 | 2/2 | 8/8 | 0 / 0 |
| Test macOS | 52/52 | 2/2 | 8/8 | 0 / 0 |
| Test Windows | 52/52 | 2/2 | 8/8 | 0 / 0 |
| Race Ubuntu | 52/52 | 2/2 | 8/8 | 0 / 0 |
| Race macOS | 52/52 | 2/2 | 8/8 | 0 / 0 |

Exit values were read from each hosted lane's explicit test-gate log line. The counting assertion script exited 0. An initial strict JSON reader exited 1 because the Ubuntu race artifact begins with three Go dependency-download diagnostic lines. Inspected all non-JSON lines, then reran allowing only the exact `go: downloading ` prefix; every other line remained subject to JSON parsing. This was an evidence-reader format issue, not a test failure or ignored malformed event.

## Prior substantive review retained

BUG-261004-16a407_review-verdict.md remains the detailed swept-surface, AC, static red-first, and mutation analysis. Its implementation conclusions apply unchanged because the three paths compare equal. Independently reread the complete candidate diff: both production audit entry variants and actual install.Project regression remain covered; shared decision logic preserves revocation and the pre-capability pin requirement while applying normal finding policy to pinned content. Advisory findings warn, threshold/off controls remain, and fresh/cache states are asserted explicitly.

Static red-first bound: the base's unconditional pinned allow contradicts 28/52 audit matrix rows and both install refusal assertions. The remaining 24 audit rows are controls. Restoring that bypass is predicted killed; retaining it on only one path is predicted killed by 14 corresponding fresh or cached audit rows. The earlier reviewer-added >= to > threshold mutant is predicted killed by exact-threshold rows. These are reading-based predictions, not locally executed mutants. The existing spec-clarification outcome remains the authorized draft-only deliverable.

No local Go build, test, vet, or lint was run, as required by HOSTED-EVIDENCE MODE. Fresh hosted measurements supersede the prior snapshot evidence; prior local execution claims are not independently reasserted. Coverage ratios describe the named regression/control matrices, not all possible audit detectors or schemas.

The reviewer goal query reported no active goal (run not goal-bound). This task-scoped artifact also records the review outcome and evidence-reader anomaly without editing LOGBOOK.md. Accept revision 1 via accept_cr; route to integrating for its tracked developer/implementer producer.
