# TASK-260728-zb2s4z: amend-portable-go-v1-host-execution-contract

## Description
Amend the unreleased rc.5 candidate so go-v1 normatively supports the maximum autonomously enforceable portable manager-owned execution profile on macOS and Windows without claiming the separately tracked hardened guarantees.

## Scope
Start from the independently accepted TASK-260728-3b8qym candidate state. Update protocol core, manager profile, security/threat text, decision record, go-v1 driver policy, cache/receipt/marker/claim identity, conformance vectors, compatibility guards, authoring guidance and rc.5 release metadata. Permit one identity-verified manager-owned worker boundary with package influence excluded. Define exact portable mandatory controls and capability evidence; explicitly exclude full network/filesystem/resource/executable-allowlist guarantees and point to STORY-260728-327soo. Keep schemas 1-5 byte compatible, schema-6/7 declaration shapes closed, external audit-before-cache/compiler unchanged, candidate platform claims empty and all release provenance honest. No Curator or csk code.

## Acceptance Criteria
Normative text and executable vectors distinguish portable manager-worker-v1 from future hardened execution; package-controlled executable, argv, environment, output path, flags, hooks, plugins and generators remain impossible; fixed offline vendored Go behavior, private staging, bounded time/output/artifact, pre/post identity verification and available native controls are mandatory; unavailable hardened capabilities do not reject portable builds and cannot appear as hardened claims. Execution-policy identity is included in cache, receipt, marker and claim semantics so aliases are impossible. All schema, Python, Go, deterministic regeneration, compatibility and clean rc.5 release gates pass; a new exact downstream candidate digest is independently recomputed.
