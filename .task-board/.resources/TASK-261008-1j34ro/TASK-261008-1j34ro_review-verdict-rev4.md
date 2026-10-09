# TASK-261008-1j34ro — N8 audit content identity review, revision 4

Verdict: accepted. CR-TASK-261008-1j34ro-4, revision 4.
Base: f2ed88a615f648f4c2fd8beca6a1c4d3a89c5ffc.
Candidate tree: 36d74898f16ed9dda196b4e12e6760676d0abcf7.
Route: accept_cr to integrating; acceptance does not claim integration or landing.

No current findings; no P0/P1 blocker or reproduced P2 defect. Revision-1 F1 and F2 are resolved. No new repeat-of mechanism remains. No code, branch, or LOGBOOK edits made by this reviewer.

```verdict-findings
{
  "findings": [],
  "notes": [],
  "surface_results": [
    {"row":"CLI supplied identity and refusal ordering","result":"held","detail":"Production run -> cmdAudit: 6/6 malformed-input rows exit usage with zero Load calls; 3/3 supported spellings pin successfully. Hosted ordering mutant fails all six no-Load assertions."},
    {"row":"Pin writer identity and namespace","result":"held","detail":"Production Pin -> PinAtVersion: 5/5 malformed inputs refused without audit/escape state; prefixed digest admitted. CLI traversal and nested-path attacks refused. ParseDigest precedes all writer filesystem calls; pinDir precedes MkdirAll and WriteFile."},
    {"row":"Merged carrier and version compatibility","result":"held","detail":"Hosted v1/v2 carrier, cross-version refusal, CLI writer-version and N9 creation-time regressions pass on the exact revision-4 tree."},
    {"row":"Trust decisions and revocation controls","result":"held","detail":"Hosted Gate/GateReadOnly fresh/cached policy matrix: 52/52 cases pass; old-schema, preserved pin, hash/source revocation, registry install refusal, draft strict findings, and 4/4 broken-report cases pass."}
  ],
  "free_hunt": []
}
```

## Prior verdict and swept surfaces

| Surface | Result and evidence |
| --- | --- |
| F1: explicit empty flag bypass | Resolved. flags.Visit tracks presence independently of value, and the same boolean selects validation and pinning. Both split-empty and equals-empty spellings are in TestAuditAllowRefusalPrecedesConfigLoad, with exitUsage and zero Load calls. An absent flag still enters the pre-existing ordinary audit branch. |
| F2: ordering proof and unreadable state | Resolved. loadCountingSource wraps the real configuration seam used by run; an admitted-digest control records nonzero loads. Six refused shapes prove zero loads. assertNoPinState returns WalkDir errors and checks the returned error, so failed reads cannot attest absence. The hosted ordering mutant is distinguished from a simple exit-code refusal. |
| Supported identity grammar | ParseDigest admits 64 ASCII hex characters, optional lowercase sha256: prefix, and surrounding whitespace consistent with existing Normalize. Canonical directory is lowercase bare hex. CLI hosted controls cover bare, uppercase hex and prefixed input (3/3), and malformed traversal, nested, short, nonhex and explicit empty input (6/6 ordering rows). These are enumerated cases, not exhaustive input-space coverage. |
| Writer callers and containment | CLI cmdAudit calls PinAtVersion; Pin delegates to PinAtVersion. Both reach ParseDigest before writer filesystem calls. pinDir checks cleaned lexical containment before directory creation and trust.json writing; its digest input is bare hex, so the parse subsumes path escape within the supported grammar. Direct-library regression covers five malformed values and one admitted prefix form. |
| N9 and 4mzun5 merge | The exact six-path base-to-candidate delta is N8 only. cmdAudit's validation/pin region and pinDir/write region equal revision 2; hashing and both new test files are byte-identical. Merged PinAtVersion retains N9 created_at and existing v1/v2 carrier rules; adjacent audit cache/package changes belong to current trunk. Exact-tree hosted green includes their existing regressions. |
| Negative controls | Inspected hosted event rows for versioned carriers, v2 refusing v1 pin, v1 refusing v2 pin, old-schema pin requirement, preserved source-audit pin, CLI writer versions, draft pin not waiving strict findings, Gate/GateReadOnly fresh/cached decisions, local revocations, registry revocation install refusal, draft revocation, and broken reports. All pass. Current-trunk strict-finding policy is preserved; this review does not assert the old report's pre-trunk pin-over-finding behavior remains current. |
| Documentation and scope | Unreleased/Fixed entry accurately states the supported identity and malformed-value namespace refusal. Six changed paths, no unrelated N8 product delta. |
| Free hunt | Examined split/equal empty values, absent versus supplied flag, parser normalization, alternate Pin entry, invalid-version refusal, CLI dispatch/config order and containment/write order. No additional reproduced finding. |

Containment here is lexical containment against operator-supplied identity. Physical symlink containment, directory exchange races, and exhaustive filesystem/platform schedules are outside the N8 path-input report and are not claimed proven. The containment guard is defense in depth behind a closed digest grammar; no independent claim that disabling that redundant guard alone would redden a test.

## Hosted evidence independently verified

Tests were not run locally. The reviewer downloaded hosted Ubuntu test-evidence artifacts, parsed their go-test.json events, checked API conclusions, and compared frozen Git trees and mutation diffs. The reviewer accepts the producer's disposable-clone hosted executions; no new mutant or full suite was executed by this reviewer.

- Revision-4 green: https://github.com/relux-works/curator/actions/runs/37916848969, head 5a6b389c0a13d96a63453da32f572c02f992e396. GitHub commit API tree equals the candidate exactly. Conclusion success; full hosted matrix green. The trigger-inapplicable rose-air and candidate-suite jobs are skipped, not claimed exercised. Ubuntu artifact 11610853384, go-test.json sha256 bfae5b0fdbf3ae635c24441f4e9debc75838448a4ebc393e2dba15d4bee4a880. Zero fail events. TestAuditAllowPinsOnlySupportedContentIdentity: 8/8 pass events (parent + 7 cases); TestAuditAllowRefusalPrecedesConfigLoad: 8/8 (parent + 6 refused + admitted control); TestPinRefusesNonDigest: 1/1 test event with all five refusal loop iterations and admitted control. Creation-time CLI test: 3/3 (parent + two writers). Gate policy matrix: 53/53 (parent + 52 cases). Draft broken reports: 5/5 (parent + four cases).
- Revision-4 handoff gate: https://github.com/relux-works/curator/actions/runs/37921921508, head 603cdf4662251e5fe7169a635d2248ea1ba61a28. GitHub tree also equals the candidate. API success and attached validation log exit 0, 20 successful jobs, two trigger-inapplicable skips. Its command-shard coverage is 1/1; case evidence above comes from the downloaded green artifact, not from the shard summary.
- Revision-2 green reference: https://github.com/relux-works/curator/actions/runs/37891286118, head 3573df19e68172be9ca903bad0574a8c4cb3769f, API success. Used to verify the following mutant diffs and unchanged mutation-relevant regions. Revision-4 green supplies current merged-candidate validation; revision-2 green is not substituted for it.
- Hosted red mutant, original-unvalidated-pin-writer-and-CLI: https://github.com/relux-works/curator/actions/runs/37891285927, head 2deff1c753cd5ede3baa96c0c9c21dfdcbc0dcd4, API failure. Independently fetched and diffed against revision-2 green: only cmd/curator/main.go, internal/audit/audit.go and internal/hashing/hashing.go production fixes are reverted; tests and documentation retained. Artifact 11599197055, go-test.json sha256 8089b32bccdf94af52d45129ecd6d7aea2595f91308b341d70427f59673e72ed. Fail events: TestAuditAllowPinsOnlySupportedContentIdentity and its four refusal cases; TestAuditAllowRefusalPrecedesConfigLoad and its six refusal cases, including both empty spellings; TestPinRefusesNonDigest; corresponding two package summaries. Exactly 15 fail events, no other failed test. All three original positive CLI forms and admitted-load control pass.
- Hosted narrowing mutant, configuration-load-before-digest-validation: https://github.com/relux-works/curator/actions/runs/37891285902, head 1eada2904ee3b0f1a986f219f6c9f78f33b2712c, API failure. Independently fetched and diffed: only loadConfig's four-line block moves ahead of validation; digest refusal remains enforced. Artifact 11599287151, go-test.json sha256 ca23a123c685c6c6fb9a4f7c2e9fc309ca33e2fce7e672ec547eb9313d7d7942. TestAuditAllowRefusalPrecedesConfigLoad fails for 6/6 refusal cases with `refused --allow reached configuration (1 Load calls), want zero`. Parent and package summary give eight fail events total. Original CLI refusal regression and library regression pass; admitted-load control passes. This distinguishes ordering from gate existence.

The older red/ordering runs are retained as targeted mechanism evidence because both test blobs, ParseDigest, cmdAudit's validation/pin branch, pinDir and writer path are unchanged. Current merged carrier behavior is separately validated by revision-4 green. No claim that the full revision-2 source/environment identity equals revision 4.

## Reviewer checks and lifecycle

Personally reran go vet ./... and go build ./...: both exit 0. gofmt -l on all five changed Go files: no output. git diff --check for the exact base/candidate delta: exit 0. No local go test, go run or compiled test binary execution.

Read the live checklist: 14/14 items checked. No LOGBOOK edits per the task-specific brief; this task-scoped verdict records the review decisions. Queried spawn goal: this run is not goal-bound. No directives recorded at the review checkpoint.

Fresh origin main advertised 0cbdd9a6c350eb9f9d1934bc415b49a083ff724f and fetched main was compared with the CR base: upstream delta contains board records only, no source or validation-input overlap. No convergence needed for this review.

Acceptance is for revision 4 only. The next tracked developer/implementer run owns checkpoint/integration; this reviewer supplies no commit acknowledgement and does not set done.
