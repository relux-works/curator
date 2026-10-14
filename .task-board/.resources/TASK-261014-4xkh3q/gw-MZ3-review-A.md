# TASK-261014-4xkh3q — Muse root-context materialization: reviewer A

Record-only review, as required by gw-MZ3-review-A-brief.md. This report is not CR acceptance or rejection; reviewer B owns that decision. No product code, branch, index, or commit was changed. No accept_cr/reject_cr/status verdict was submitted.

Candidate: CR revision 1, tree `716805ff8e27fa37f992d3658eb9d69942e9aea3`, base `5234cf38833554542e68afade15951c6e56ed89e`. All eight working-tree candidate blobs match their CR blobs. Local main is one commit ahead of this base, affecting only board files; this is not a fresh-authority assertion.

## Result

No reproduced blocking product finding. The implementation and named regression tests support points 1–5 of the developer brief within the explicit external argv ownership bound. This record does not attest to signed delivery or external launch behavior.

```json
{"findings": [], "notes": [
  {"id":"unsigned-gate-snapshot","text":"GitHub reports snapshot 599cf6b128a73f79c1d675fddbb3b6f03ce83986 unsigned. scripts/remote-gate.sh already creates unsigned commits with a remote-gate identity. This pre-existing workflow conflicts with the signed-commit requirement; signed delivery is not established by the green run. No delivery commit exists in this uncommitted Story candidate."},
  {"id":"external-launch-bound","text":"No production Muse exec argv composer was found. The gaps artifact quotes the exact section 10.2 requirement and assigns flag selection to curator-agent-launcher/task-board. External argv execution and same-release coordination remain unverified here."},
  {"id":"regular-file-consumer-bound","text":"The generic materializer links Muse root context to its managed store, like other non-Claude adapters. File reads follow that link. The external launcher must interpret the gate note consistently: a following regular-file check accepts this; an Lstat-only no-symlink check would not. No external consumer was inspected, so this is a compatibility note, not a reproduced finding."},
  {"id":"review-execution-bound","text":"R223 permits local compilation/static validation only. This reviewer independently checked hosted state, exact tree identity, complete uploaded behavioral evidence, source and mutation mechanisms; did not rerun tests or execute additional behavioral attacks. Reviewer A's specific brief requests reasoning about mutants and record-only delivery. No fresh attack sweep is claimed."}
]}
```

## AC coverage checked before implementation review

**7 of 8 AC rows driven**, with the eighth explicitly out of this repository's scope under developer brief point 4. The producer gaps resource contains the same ratio and call-site map. Tests are present in the candidate tree, not yet a signed delivery commit.

| AC | Named driving test | Production route |
| --- | --- | --- |
| Monolithic path, default, bytes, marker/hash | TestMuseRootContextVectors/monolithic | run → cmdEnvResolve → Resolve → homePlan.rootContext → Monolithic |
| Zero applicable modules/header only | TestMuseRootContextVectors/zero-applicable | same route |
| No context declaration/no root surface | TestMuseRootContextVectors/no-context | same route |
| Referenced refusal/no fallback | TestMuseRootContextVectors/referenced-refused | cmdEnvResolve → machineFromConfig → MachineConfig.Form → Adapter.ResolveForm |
| Changed root drift/refusal/repair | TestMuseRootContextDrift/changed | cmdEnvResolve → Resolve → verifyHome/checkSurfaces; cmdEnvStatus → StatusOf |
| Removed root drift/refusal/repair | TestMuseRootContextDrift/removed | same routes |
| Standing note/residuals | TestMuseRootContextStatus | cmdEnvStatus → StatusOf |
| Conditional exec flag | Stated external ownership bound | No Muse exec composer in curator |

Registry metadata is also checked by TestMuseHasNoAdmittedRootContextForm: target config/muse/AGENTS.md, monolithic only/default, version 1.4.7, advisory 32768. Its historical name is misleading but behavior is current.

## Implementation and mutation review

The registry redefines RootTarget as a portable home-relative path; existing generic materialization, storage, marker and drift code handles the nested target. No Muse special case was added to production callers. The generic CLI Forms binding repairs the real production route; direct registry tests alone would have missed it.

TestMuseRootContextVectors assembles the expected full document independently and compares bytes. Its expected value begins with `<!--` then `curator-root-context-v2` and the complete normal generation header; a prefix, reordered header, missing first line, or changed body cannot pass. The zero-applicable case checks header-only bytes. This is stronger than a first-line-only assertion and meets MZ3-gate-note.md's provenance-prefix requirement.

The status note includes delivered root context and every required residual: serve exclusion, observed workspace precedence, unverified model-visible foreign removal, skills and system-prompt discovery. Tests use temporary config/native homes and an inert Muse executable; the new paths do not touch a live home.

No catalog surface table or review/free-hunt budget was supplied. The concrete candidate behavior table below records reused hosted attacks; these are not newly executed reviewer attacks.

| Surface | Result | Attack evidence |
| --- | --- | --- |
| Configured form/refusal | held (hosted reuse) | configured-form-ignored and referenced-fallback selectively admit Muse referenced; both fail TestMuseRootContextVectors/referenced-refused |
| No-context gate | held (hosted reuse) | no-context-admitted admits only Muse without declaration; fails no-context |
| Output identity/path and bytes | held (hosted reuse) | wrong-target changes only Muse target; wrong-bytes preserves header and adds bytes; named monolithic/zero-applicable tests fail |
| Drift verification | held (hosted reuse) | changed-root-admitted and removed-root-admitted selectively skip exactly one Muse root class; named changed/removed CLI tests fail |
| Status residuals | not-attacked | Named positive/negative text assertions verified in hosted evidence; no independent behavior-changing status mutant executed |
| External argv | not-attacked | Explicit owner bound; consumer absent from this repository |
| Destination link behavior | not-attacked | No new fixture executed; consumer compatibility question recorded above |

Seven mutants retain the relevant production behavior and selectively weaken or alter it; they are not a delete-only table. The harness executes behavioral CLI tests with -count=1 and requires the named failing subtest, rejecting compilation errors or unrelated failures. Source tokens merely anchor overlay construction; this is not a static-only checker. The byte mutant preserves the generation header while changing output.

## Evidence independently inspected

- Pinned curator-spec eb1f08c environments §§7.1, 7.2, 7.9, 10.2, 12, 13; generic materializer and its callers; candidate patch; producer TASK-261014-4xkh3q_gaps.md.
- GitHub API: [run 38489868868](https://github.com/relux-works/curator/actions/runs/38489868868) completed/success, snapshot `599cf6b128a73f79c1d675fddbb3b6f03ce83986`, exact tree equal to CR. Twenty jobs succeeded, two conditional jobs skipped. No skipped job is claimed passed.
- Downloaded all five Test/Race evidence artifacts. Each complete CLI event stream ends in package pass, includes 18 matching Muse test/subtest pass events and seven named expected-red mutant kills (35 total). Ubuntu Race starts with three Go dependency-download lines, followed by valid JSON events; those lines were explicitly accounted for, not mistaken for truncated test output.
- Attached TASK-261014-4xkh3q_review-A-evidence.json records lane source hashes, named tests, kills and terminal results. Hosted source/test/workflow identity is the exact candidate, not a different producer attempt.
- git diff --check passed. No new local test, full remote-gate rerun, live Muse probe, or signature repair was performed. The producer's successful remote-gate exit is cited from its gaps resource; hosted success and tree identity were independently verified via GitHub.

Reviewer B should use this as record-only input, resolve delivery-signature policy at the owning workflow, and avoid claiming that curator's tests establish external argv behavior. The task remains reviewing intentionally under the specific reviewer-A instruction.
