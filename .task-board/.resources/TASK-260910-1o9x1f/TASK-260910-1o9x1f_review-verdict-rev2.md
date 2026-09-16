# TASK-260910-1o9x1f review verdict — revision 2

Verdict: ACCEPTED for the scoped draft loader/planner leaf. R1 is resolved; no additional rework found. Integration remains producer-owned.

## Exact candidate and review

CR-TASK-260910-1o9x1f-2; base 12f1287ee0fb538f9ca004dd53b870e824e5baf2; candidate tree 109e50d9d559ba04ec73329365eb16f5e9b7a96d. Independently compared all 13 changed files against candidate blobs: 13/13 byte-identical. Hosted gate commit 9e857e0bf6d41390089ae8b581236a3eac3257ed has exactly this tree. Revision-1 to revision-2 delta is only the two presence guards and regression tests (two files, 37 insertions, 2 deletions).

Read the producer results, prior rejection, revision-2 validation log, local accepted repository-transport revision-1 contract, source-policy-v1 schema and relevant Skillfile diagnostics. Diff stays within config, identity, gitcred and tests/docs, builds on existing parser, labels support draft, rejects revision 2, and changes no frozen schemas.

R1: present null root_inputs now reaches object validation; present null pin reaches string validation. TestLoadSourcePolicyRejectsNullOptionalFields loads both exact prior-verdict documents from real temporary files and requires repository_policy_invalid. Omission and valid-value controls also load through LoadSourcePolicy. All 4/4 subtests pass independently.

## Independent checks

Shell: zsh; direct tool return codes captured. Python invoked Go directly for overlays.

- go test -count=1 ./internal/config ./internal/identity ./internal/gitcred: exit 0, all three packages pass, including legacy checks.
- go vet ./internal/config ./internal/identity ./internal/gitcred: exit 0.
- gofmt -l internal/config internal/identity internal/gitcred: exit 0, no output.
- go test -count=1 ./internal/config -run TestLoadSourcePolicyRejectsNullOptionalFields -v: exit 0, 4/4 pass after attacks.
- M1: narrow root_inputs validation to present && rawInputs != nil using a temporary Go overlay. go test -count=1 -overlay <overlay> ./internal/config -run TestLoadSourcePolicyRejectsNullOptionalFields -v: exit 1, only null_root_inputs fails with nil error. KILLED.
- M2: narrow pin validation to present && rawPin != nil using an independent temporary overlay. Same command: exit 1, only null_pin fails with nil error. KILLED.
- Measured attacks: 2/2 killed, no survivors. This is not exhaustive schema-clause mutation coverage. Product bytes were never modified; after overlays source SHA256 remains b755e77db500db3e0438a1ba3e118c2beadb985dd56d10151b7d0741944a5963.

Accepted hosted evidence, not independently replayed: TASK-260910-1o9x1f_change-request_rev2-validation.log, exit 0, run https://github.com/relux-works/curator/actions/runs/35078813943. Windows/macOS/Ubuntu tests, macOS/Ubuntu race, all three gate self-tests, lint, naming and interop conformance report success. Candidate suite and rose-air skipped; ARM portability is unverified. No independent full-suite or golangci-lint run claimed.

## Acceptance traceability and bounds

| Acceptance | Candidate tests and entry point |
|---|---|
| Schema | TestDraftPolicyPublishedSchemaCases (5/5 vectors), TestParseSourcePolicyDocumentShape, TestParseSourcePolicyRootInputs, TestParseSourcePolicyRejectsRevision2Closed through ParseSourcePolicy; null regressions through LoadSourcePolicy |
| Exact canonical identity | TestParseSourcePolicyEntryKeysAreExactCanonical through ParseSourcePolicy; TestDraftCanonicalKey and TestPolicyKeyAndEndpointIdentityAgree bound helpers |
| One/two distinct endpoints and provider refs | TestParseSourcePolicyEndpoints through ParseSourcePolicy; TestValidProviderAdmitsOpaqueIdentifiersOnly bounds provider syntax |
| Order and pin | TestParseSourcePolicyPinAndFallback and TestResolveRepositoryEndpoints through parser/planner |
| URL without entry once | TestResolveRepositoryEndpoints: one planned attempt, empty provider and no fallback |
| Logical identity without entry | TestResolveRepositoryEndpoints: repository_endpoint_unavailable |
| Invalid/unreadable policy | TestLoadSourcePolicy and TestLoadSourcePolicyRejectsNullOptionalFields: real files/directory error, malformed/null documents, omission and valid controls |
| Package inputs cannot introduce providers/commands | TestParseSourcePolicyEndpoints, TestValidProviderAdmitsOpaqueIdentifiersOnly and TestResolveRepositoryEndpoints establish closed identifier syntax and planner interface only |

As in review 1, caller search finds no acquisition callers of LoadSourcePolicy/ResolveRepositoryEndpoints. This accepts library loader/planner behavior, not enabled acquisition: 0 acquisition entry points driven here. Configured-provider existence and broker resolution, actual network attempt counts/fallback/deadlines and package-to-acquisition isolation remain TASK-260910-5nrmtt integration responsibilities. Root-input existence and filesystem case equivalence remain acquisition-time checks. Cross-platform gate success is not proof of every untested path/mode edge. Tests are captured in the candidate tree, not a producer-authored commit.

Goal query immediately before verdict: not goal-bound. No directives pending. No code edits, commits, runtime-home changes or commit_ack. No LOGBOOK.md edits under campaign prohibition; review evidence is persisted in this task outcome. Acceptance must be recorded with accept_cr revision=2, routing to integrating rather than done.
