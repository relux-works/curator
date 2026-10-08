# TASK-260927-25hk87 — flip-security-posture-default-hardened: revision-3 review

Verdict: **accepted**. No findings requiring rework. LANDING HELD until the operator schedules the B release. This review accepts the candidate; it does not authorize landing or release.

```json
{"revision":3,"verdict":"accepted","findings":[]}
```

Reviewed CR-TASK-260927-25hk87-3, base 75ab9a71a9b9049ec4242b4acdb60c3fead6aff1, candidate tree 922cb84700584e5e448aa18f9fb8f40896c1523f. All 19 changed paths swept. No repository files edited by this reviewer.

## Previous verdict and re-application

Read TASK-260927-25hk87_review-verdict-rev2.md in full: it accepts revision 2 and explicitly reports no findings requiring rework. The round brief's generic reference to a rejection is inconsistent with that artifact. There are no prior mechanisms to repeat or unresolved findings to answer; repeat-of is not applicable to the empty findings set. Review this round as the binding revision-3 note requests: re-application and collateral sweep, rather than a redesign.

The downloaded revision-2 and revision-3 CR patches have exactly the same 19 paths and are byte-identical after excluding index blob IDs and hunk line coordinates. This comparison retains every context, addition and deletion line. Of these 19 candidate files, 18 blobs are identical to the accepted revision-2 tree af82499fd08af44beeb05f50162fa107fafa2f4b. The only differing blob is cmd/curator/env_credential_marker_test.go, whose revision-3 versus fresh-main delta is exactly one fixture line adding explicit security_posture=permissive to TestEnvResolveCredentialRecordIsolatedKeychain.

All other content in that file equals fresh main: schema-3 markers require HashVersion=2; both schema-1 repair tests retain pinV1MarkerWriters and delete hash_version during downgrade, and all existing credential, byte-preservation and seed assertions remain intact. Compared with revision 2, the file changes are precisely the upstream schema/hash and v1-lane adjustments. No posture-B semantics change. The macOS hosted artifact proves the re-applied isolated-Keychain test and both frozen-v1 repair tests actually pass, rather than being skipped. Ubuntu skips only that macOS-specific Keychain test among these three, as its ledger declares.

Fresh origin HEAD was independently advertised as refs/heads/main at 75ab9a71a9b9049ec4242b4acdb60c3fead6aff1; a fresh exact-ref fetch yielded the same OID, equal to the CR base. Every one of the 19 working-file blob IDs matches the immutable candidate. git diff --check on the immutable delta exits 0.

## Swept surfaces

| Surface | Result and evidence |
| --- | --- |
| Released A prerequisite | Independently queried release v0.15.0-rc.3: published 2026-10-04T03:03:40Z, draft=false; its security_posture.go declares A. Prerequisite satisfied before B. |
| Re-application and NUL/hash lanes | Exactly one fixture change against fresh main; new writer/schema/hash assertions and pinned genuine-v1 repairs preserved. No new test suppressions. Current macOS artifact records all three affected credential/repair tests passing. |
| Hardened production default | SecurityPostureRevision=B remains the sole product-code selector change. config.Load -> parseConfig -> defaultSecurityPosture/applyHardenedDefaults selects hardened for omitted schema-2 posture, preserving schema-1 permissive and explicit-knob precedence. Current hosted default, explicit-knob and system-lock tests pass. |
| Permissive compatibility | Explicit permissive retains advisory audit/registry, drop, optional signers, existing empty allowlists and passthrough defaults. run() profile install succeeds with empty sources and warning once; status/warning tests pass. The unrelated legacy fixtures explicitly choose their former permissive policy. |
| Production refusals and controls | run() install/update/upgrade/profile install tests cover absent and explicit-empty source lists and assert no publication. Signed MCP fixtures reach MCP policy after real signature checks. Empty MCP without declarations remains admitted; declarations are refused under hardened. env resolve refuses explicit unbounded passthrough; env status --check reports actual MCP-bearing non-current state. Relevant current hosted tests pass. |
| Conformance and gaps | Hosted macOS and Ubuntu artifacts both measure config 12 driven + 5 bound + 0 known-gap + 0 skipped = 17/17; CLI 13 driven + 4 bound + 0 known-gap + 0 skipped = 17/17. All six B vectors remain driven; all four prior B deferrals are gone in both unchanged consumers. No security-posture/vectors gap ledger rows exist at the base or candidate; unrelated seed rows remain unchanged. |
| Other revisions and architecture | Only internal/config/security_posture.go changes production code, with selector, comments and warning hint. CodexSeedRevision remains A in internal/envregistry/envregistry.go, which has no delta; other revision selectors are unchanged by this CR. Tests and docs use the existing policy mechanism. |
| Documentation and stray changes | All 19 paths remain the accepted posture-B delta. Documentation states default/compatibility and exact counts; golden changes only remove the obsolete future-flip warning phrase. CHANGELOG.md, LOGBOOK.md and scripts/remote-gate.sh have zero delta. |

## Current hosted CR gate

[Hosted gate 37710630256](https://github.com/relux-works/curator/actions/runs/37710630256) is completed/success. Independently queried GitHub run and jobs. Its head snapshot 5e23dbe77f9a68ef35e8e06a0e171bc0f78959ae has tree 922cb84700584e5e448aa18f9fb8f40896c1523f and parent 75ab9a71a9b9049ec4242b4acdb60c3fead6aff1: exactly this candidate and base.

20/20 executed jobs passed: Ubuntu/macOS/Windows tests, Ubuntu/macOS race, lint, naming, interop, three gate self-tests and nine Go driver jobs. Two declared job skips: rose-air and alternate candidate-suite lane. These job skips are distinct from posture-vector accounting, which has zero skipped cases. The producer CR validation resource records remote-gate exit 0. This new gate qualifies revision 3; revision-2 full-suite evidence is not substituted for the changed candidate.

Downloaded current macOS artifact 11522580382 and Ubuntu artifact 11522064332; independently parsed complete test/go-test.json streams. Both contain successful config and CLI posture-vector parents, exact measured coverage lines, and passing hardened-default/permissive/negative-control tests. The condensed task-scoped TASK-260927-25hk87_review-hosted-rev3.json attaches run/job results, artifact member digests and sizes, coverage lines and relevant test outcomes without private machine paths or duplicated corpora.

## Reruns, reused evidence and bounds

Reran independently: immutable-delta whitespace check; all-path patch/content/working-tree comparisons; fresh advertised/fetched main comparison; release and hosted run/job queries; complete hosted test-artifact parsing. No local cmd/curator tests were run, per R194. No product/test files were mutated. Full test/build/lint qualification is accepted from the current exact-tree hosted gate. The producer's revision-3 results additionally record build-lock-protected GOFLAGS=-work go vet for cmd/curator and internal/envprofile, exit 0; that command was not rerun here.

Accepted from earlier attached evidence, rather than claimed rerun: revision-2 review reports actual expected-red mutation exits for reverting B to A and narrowing the source-list gate to nil-only. The relevant product/default/refusal test files are byte-identical in revision 3; those results establish the previously reviewed test sensitivity, while today's unmutated candidate is independently qualified by the new hosted gate. No whole-repository mutation-coverage claim.

Bounds remain explicit: historical implicit-A vectors cannot describe the shipped B default; config-level registry I/O is additionally bounded and exercised by production CLI tests. The coverage ratio reports accounted cases, not that every bound was executed as historical behavior.

## Logbook and lifecycle

The stale revision-2 conflict is repaired with no posture-B semantic difference; the generic prior-rejection wording is resolved by reading the actual accepted verdict. This section records the review findings/anomalies without changing the prohibited LOGBOOK.md file. No remaining finding or external blocker was discovered.

Queried task-board spawn goal: run is not goal-bound. All 11 live checklist items are checked. Attach this verdict and current hosted verification before accept_cr revision=3, which routes to integrating. No done transition, commit_ack, checkpoint, integration, branch switch or release. The operator's B-release hold persists.
