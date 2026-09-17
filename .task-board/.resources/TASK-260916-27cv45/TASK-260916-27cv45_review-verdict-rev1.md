# TASK-260916-27cv45 review verdict revision 1

ACCEPTED. No blocking findings for this loader/identity leaf. Integration and done remain producer-owned.

Exact candidate: 4dfd5b3f629dfb7193e1f5c48ad8176f525475e1, base c64ceafc046828535e7acf94b2250ac49b984e03. Independently compared 17/17 changed file bytes to candidate blobs, including untracked test/fixture files. No product code edits or commits. Read repository-transport \u00a7\u00a74-7 from curator-spec main, producer results and validation, and predecessor loader/wiring acceptance evidence.

Implementation: sourcepolicy.go:194 dispatches schema 1/2 and refuses unknown versions; schema-1 parsing retains prior semantics. sourcepolicy.go:615 calls the shared identity guard from the production ParseSourcePolicy entry. Lines 635-679 enforce path equality, embedded alias refusal, exact lookup, mirror URL plus alias refusal, double-port/auth equality and the resolved-host mirror predicate. Port parsing preserves the existing closed URL grammar through buildrepo.ParseSource after stripping the port. Identity remains the canonical map key. Executor and frozen protocol files are unchanged; docs describe draft scope and the sibling-owned resolution/provenance work.

Independent verification (zsh; direct command exit codes; no full landing-suite replay):
- go test -p 1 ./internal/config -run 'SourcePolicy|DraftPolicy|Schema1Golden|Resolve.*' -count=1 -timeout=90s: exit 0, 0.798s.
- go test -p 1 ./internal/install ./internal/buildrepo -run 'TestDraftTransport|TestAcquireDraft|TestDraftPlan|TestParse|TestCanonical|TestTransportPlan|TestValidateTransport' -count=1 -timeout=90s: exit 0; install 34.059s, buildrepo 0.514s.
- go vet ./internal/config: exit 0. gofmt -l on three changed Go files and git diff --check: exit 0, no output.
- Focused post-attack JSON test run: TestParseSourcePolicyV2Refusals 36/36 pass; TestDraftPolicySchema2Corpus 13/13 pass, exit 0. Producer prose says 34 refusal rows; actual candidate has 36.

Coverage: ParseSourcePolicy is the actual validator called by LoadSourcePolicy, not a helper-only entry. Refusal cases exercise undeclared mirrors (direct and aliased), unknown alias (three table states), attestation mismatch/spurious attestation, mirror plus alias, embedded alias host, auth mismatch, chaining, double port, exact pin, invalid URL/alias ports and grammar, and path mismatch. Existing tests cover unknown versions, exact keys, malformed input and file-loading behavior. Schema-1 golden checks unchanged projected endpoint/fallback bytes and absent v2 properties; it is not a serialization golden for every exported Go struct field. Portable receipt/lock behavior and network execution of v2 properties remain sibling scope.

Independent narrowing mutations via temporary Go overlays (candidate bytes unchanged), each run with -p 1 -count=1 -timeout=60s against TestParseSourcePolicyV2Refusals:
1. Restrict resolved-host mirror guard to aliasName == empty: exit 1, alias_to_another_host_without_mirror_of fails because ParseSourcePolicy wrongly accepts it.
2. Restrict path-equality guard to URL host == key host: exit 1, mirror_path_must_equal_the_key_path fails because ParseSourcePolicy wrongly accepts it.
Measured 2/2 selected mutations killed by behavioral assertions, zero survivors. This is not exhaustive clause/boundary mutation coverage. Initial overlay setup failed because the optional .temp directory was absent; rerun used an automatically removed temporary directory inside the worktree. No failed setup counted as a mutation result.

Hosted gate evidence reused: TASK-260916-27cv45_change-request_rev1-validation.log, exit 0. Independently queried https://github.com/relux-works/curator/actions/runs/35173336753: success, head 3679aa8977a86a60a7a4866949185c59a3e128bc; git confirms its tree equals the exact reviewed candidate. 11/11 executed jobs successful (Ubuntu/macOS/Windows tests, Ubuntu/macOS race, three gate self-tests, lint, naming, interop). rose-air and Candidate suite skipped; ARM runtime behavior is unverified.

No pending directives; spawn goal reports not goal-bound. This task outcome and board notes record the review; no LOGBOOK.md edit under campaign prohibition. Accept through accept_cr revision=1, no commit_ack.
