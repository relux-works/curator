# Review verdict: accepted

Task: TASK-260910-1xph2y
CR: CR-TASK-260910-1xph2y-2, revision 2
Base/HEAD: d019f0e7179520b5c8dcde321c4fe51e04552f58
Candidate tree: 4087f02f5459a96d1a03d78ddb343d82608df0e6
repeat-of: none (prior F1 resolved)

## Findings and decision

No material findings remain. Accept the exact current CR. F1 is resolved in
protocol/skillfile-sources.md section 4 and install-marker-v5.schema.json:
attestation and legacy substituted are retained through frozen v4 references,
with explicit Git/local applicability, non-authorizing summaries, effective-plan
comparison, and refusal of missing/unreadable/stale/mismatching required evidence.
Strict substitution refusal precedes cache/compiler/publication and cannot be
bypassed by omitting marker evidence. New source selectors do not acquire legacy
substitution behavior. External substitution remains independently represented,
including for local packages.

The complete 25-field v4 migration and both build-record arms were checked.
External records retain declared/effective identities, locked commits, typed
substitution, source/target, execution policy and artifact/receipt evidence;
receipt version becomes 3 with explicit package and input equality rules.
Currentness, repair and refresh retain trust checks and atomic failure behavior.
The source-audit object does not replace registry evidence or authorize itself.

Reviewed the current full delta and revision-1-to-2 changes, approved brief and
its four scoped design inputs, canonical source/transport contracts, versioned
schemas, guide and compatibility/navigation changes. Local input/output physical
boundaries, frozen complete context/runtime/build inputs, root selection,
collection membership, dependency closure, and machine-owned bounded transport
remain consistent with the approved design. Existing external-build security
and unsupported implementation boundaries are preserved. No diagram changed.

## Independently rerun evidence

Every validation pass below was rerun in this reviewer session; none is accepted
solely from producer logs. Reused the already-installed task-local Python venv.
The project is specification/tooling, with a platform-neutral Python/Go lane.

| Check | Exit | Result |
| --- | ---: | --- |
| Tool and venv readiness | 0 | task-board, Git, Python, Make, Go, jsonschema/referencing available |
| PATH="$PWD/.temp/source-contract/venv/bin:$PATH" make validate | 0 | 60 schemas, 1047 vectors, 227 Python tests and Go tooling tests pass |
| Exact documented draft command, freshly extracted into review scratch | 0 | 102/102 cases; 82/82 negatives; 7/7 wire schemas; 3/3 snapshot vectors |
| Documented frozen-v4 migration/mutation checks | 0 | 25/25 fields, 2/2 build arms, 18/18 narrowed refusals detected |
| Independent adversarial.py | 0 | 16/16 marker checks; 11/11 selector attacks rejected; root selector admitted; endpoint limits and narrowed mutant; 10/10 author examples schema-valid |
| Candidate byte comparison | 0 | 126/126 changed paths match CR tree; HEAD remains base; HEAD..main=0 |
| Exact-delta whitespace check | 0 | no whitespace errors |
| Frozen schemas/v1 JSON, conformance/v1, release and tools diff | 0 | unchanged |
| Real index diff | 0 | unchanged |

The initial optional skill/file discovery command exited 2 because queried
project skill directories did not exist; it was discovery, not a validation
pass. A later bounded reference grep returned 1 for no matching logbook helper
name. Neither was used to infer a passing gate or missing runtime capability.
Raw validation logs and independent reproduction are attached separately.

## Negative evidence bounds

The exercised specification entrypoint is the documented
Draft202012Validator(..., registry=registry).is_valid(instance) loop in
conformance/draft-sources-v1/README.md. This actually loads the shipped schemas
and fixtures. Applicability mutation narrows local refusal to require both
forbidden fields, while each single-field case catches it. Sixteen external
required-field mutations isolate each missing-field refusal. Independently,
changing endpoint maxItems from 2 to 3 admits the shipped three-endpoint negative;
the expected-negative assertion detects that narrowing. Counts 0/1/2/3 were
also checked against the original endpoint schema.

Independent marker probes now admit retained attestation/substitution on both
Git arms, reject them on local snapshots, and reject malformed summaries. The
old revision-1 reproduction was read as historical evidence; its assertions
intentionally expect the old defect and were not misrepresented as a current pass.

Manager semantic execution is 0/73, unverified. The new semantic vectors cover
planning/status/repair/refresh, missing/unreadable/stale/revoked/mismatching trust
records and strict substitution with marker omission, but are downstream
requirements. No manager production path, local v2 installation, physical race,
credential fallback, compiler containment or real attestation was executed.
Synthetic marker hashes are structural fixtures, not verified linked receipts.
These explicitly documented limits fit this specification-only assignment.

## Logbook and lifecycle

F1 is closed by retained evidence and exhaustive migration coverage. No new
material regression was found. This outcome records the review decision and
validation limits in the task documentation flow.

Review was read-only except task-scoped scratch and board resources/status.
No source/schema/example repair, commit, tag, push, PR, checkpoint, integration,
switch, rebase or merge was performed. Files remain uncommitted. Invoke only
accept_cr for revision 2; the resulting integrating state means accepted and
not landed. No commit acknowledgement or done transition is authorized here.
