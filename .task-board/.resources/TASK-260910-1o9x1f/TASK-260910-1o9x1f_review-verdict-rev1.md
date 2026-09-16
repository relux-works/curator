# TASK-260910-1o9x1f review verdict — revision 1

Verdict: CHANGES_REQUESTED. Route to `to-dev`. Ordinary implementation rework; no human decision required.

## Candidate and scope

Reviewed CR-TASK-260910-1o9x1f-1 revision 1, base `12f1287ee0fb538f9ca004dd53b870e824e5baf2`, candidate tree `881b978bfebd4c6e9179dda6fb27b30aaa23cfba`. Compared all 13 changed files byte-for-byte against the tree: 13/13 match. Hosted gate commit `3ffe194d7cefa35f7a19dff4dd5bc7c3ba303697` has exactly this tree. No product code was edited. Mutation/reproducer runs used temporary Go overlays; original bytes never changed. Source SHA256 after attacks: `e5a2d89dd3597367c27a9849c9ecbb12c15e7e493c831e668a4a0b7424457c1f`.

Diff is confined to config, identity, gitcred, tests and documentation. It builds on the existing parser, labels support draft, rejects revision 2, and leaves frozen schemas untouched. Read producer results and revision-1 validation log; normative source is local curator-spec repository-transport sections 1–3 and source-policy-v1.schema.json.

## Required fix R1 — invalid optional nulls silently become absence

`internal/config/sourcepolicy.go:168` skips `root_inputs` validation when its value is null. Line 273 similarly skips `pin` validation for null. The accepted schema permits omission, but a present root_inputs must be an object and a present pin must be a repository URL string. Transport section 2 requires invalid policy to fail repository_policy_invalid, never be treated as absent. In particular a null pin silently becomes unpinned list/fallback behavior.

Independent overlay regression `TestReviewRejectNullOptionalPolicyFields` expected rejection for these exact documents; both returned nil error (go test exit 1):

```json
{"schema_version":1,"repositories":{},"root_inputs":null}
```

```json
{"schema_version":1,"repositories":{"example.org/kit":{"endpoints":[{"url":"https://example.org/kit.git","authentication":"team"}],"fallback":"none","pin":null}}}
```

Fix presence/type handling for both optional fields. Add candidate regression tests through `LoadSourcePolicy` using real temporary files, requiring repository_policy_invalid; preserve positive omission and valid-value controls. The existing tests lack both null cases (0/2 detected). Re-run narrow package checks and hand off a new revision for independent review.

## Independent checks and gate attacks

Commands ran under zsh; Go overlay subprocesses invoked directly by Python, with actual return codes captured.

- `go test -count=1 ./internal/config ./internal/identity ./internal/gitcred`: exit 0, all three packages green, including legacy tests.
- `go vet ./internal/config ./internal/identity ./internal/gitcred`: exit 0.
- `gofmt -l internal/config internal/identity internal/gitcred`: exit 0, no output.
- M1: narrowed duplicate-URL refusal to only `git@` endpoints (`seen[endpoint.URL] && strings.HasPrefix(endpoint.URL, "git@")`). `go test -count=1 -overlay <temporary-overlay> ./internal/config -run TestParseSourcePolicyEndpoints`: exit 1; killed by repeated HTTPS URL with different providers.
- M2: narrowed endpoint-count refusal from `len(list) > 2` to `len(list) > 3`. Same command/test: exit 1; killed by three distinct endpoints.
- Narrowing attacks killed 2/2. This is measured coverage of those two attacks, not all schema clauses.
- R1 reproducer: `go test -count=1 -overlay <temporary-overlay> ./internal/config -run TestReviewRejectNullOptionalPolicyFields`: exit 1, two invalid documents accepted. Added test body loops over the documents above and calls ParseSourcePolicy, failing if err == nil.

Hosted evidence accepted without rerunning the full suite: run https://github.com/relux-works/curator/actions/runs/35073777209, revision-1 validation log exit 0. Windows/Ubuntu/macOS tests, Ubuntu/macOS race, all three gate self-tests, conformance, naming and lint report success. rose-air and Candidate suite skipped; ARM portability remains unverified. No independent hosted platform rerun or independent golangci-lint run claimed.

## Acceptance traceability and bounds

| Acceptance row | Candidate tests | What is established |
| --- | --- | --- |
| Policy schema | TestDraftPolicyPublishedSchemaCases, TestParseSourcePolicyDocumentShape, TestParseSourcePolicyRootInputs, TestParseSourcePolicyRejectsRevision2Closed | Parser boundary, including 5/5 vendored schema vectors; R1 invalid-null gap remains |
| Exact canonical identity | TestDraftCanonicalKey, TestParseSourcePolicyEntryKeysAreExactCanonical, TestPolicyKeyAndEndpointIdentityAgree | Exact key and endpoint equality at helper/parser boundaries |
| One/two distinct endpoints, provider refs | TestParseSourcePolicyEndpoints, TestValidProviderAdmitsOpaqueIdentifiersOnly | Count, uniqueness, identity mismatch and opaque syntax; 2/2 reviewer mutants killed |
| Order and pin | TestParseSourcePolicyPinAndFallback, TestResolveRepositoryEndpoints | Planning order and exact listed pin; null-pin refusal missing |
| URL without entry once | TestResolveRepositoryEndpoints | One planned attempt, no provider/fallback; no actual fetch driven |
| Logical identity without entry | TestResolveRepositoryEndpoints | Planner returns repository_endpoint_unavailable |
| Invalid/unreadable policy | TestLoadSourcePolicy | Real file loading: malformed file and directory error, absence and valid file; no network caller driven |
| Package inputs cannot supply providers/commands | TestParseSourcePolicyEndpoints, TestValidProviderAdmitsOpaqueIdentifiersOnly, TestResolveRepositoryEndpoints | Identifier syntax and planner interface only; package-to-acquisition isolation not driven here |

`rg` confirms LoadSourcePolicy and ResolveRepositoryEndpoints have no non-test callers outside their definitions. Therefore this leaf establishes loader/planner APIs, not enforced acquisition behavior: 0 acquisition entry points are driven by these tests. The producer explicitly assigns fetch execution, failure classification and provider-to-broker resolution to TASK-260910-5nrmtt; retain that bound and do not claim these as completed end-to-end properties. Actual configured-provider existence, network attempts, fallback deadlines, root-input existence and package isolation still require sibling production integration. This verdict does not demand out-of-scope transport implementation in this leaf.

No LOGBOOK.md edit made because campaign rules prohibit it. Findings are persisted here and in task notes. Run goal query reports not goal-bound. No acceptance mutation or commit acknowledgement issued.

Lifecycle: outcome attached before routing; set_status(to-dev) succeeded. Requested reviewer handoff was attempted and returned exit 1: role reviewer has no end_status and cannot use handoff. Explicit changes-requested status routing is persisted; no acceptance/done transition attempted.
