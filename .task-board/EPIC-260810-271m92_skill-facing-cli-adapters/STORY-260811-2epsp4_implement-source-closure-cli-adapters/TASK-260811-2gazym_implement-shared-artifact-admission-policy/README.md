# TASK-260811-2gazym: implement-shared-artifact-admission-policy

## Description
Implement the adapter-independent recursive artifact admission service. Spec trace: .spec/skill-facing-cli-source-closure.md Source closure invariant items 1, 2, 4, and 6; Vendored compiled artifact prohibition; Required research artifact taxonomy. Accepted input: TASK-260810-29vk09.

## Scope
Add closed artifact class, trust role, decision, and diagnostic enums; byte and structure detectors; recursive ZIP, ZIP64, tar, gzip, and native-archive traversal; safe virtual paths and entries; closed resource limits; canonical artifact-manifest-v1 encoding; and pre-execution admission APIs. Keep verified-binary-v1 unavailable and preserve Go behavior.

## Acceptance Criteria
The shared service deterministically classifies and manifests every admitted node; recursively rejects compiled, opaque, unsafe, encrypted, malformed, unsupported, ambiguous, or incompletely inspected dependency content before manager execution; enforces toolchain and local-output roles; passes the accepted A01-A08, C01-C12, F01-F14, T01-T05, and current-capability V01 vectors plus Go regressions.
