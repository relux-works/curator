# TASK-260818-3vfmjv: complete-portable-protected-execution-rework

## Description
Complete the portable and graph-integrity rework identified by the independent review of TASK-260811-27xisf while leaving the unavailable Darwin lossless observer boundary explicit and fail-closed.

## Scope
Implement atomic stale-permit rejection at the process-start seam; immutable admitted source snapshot/tree replay with containment and time-of-use identity checks; exact C4/C5/closure graph validation for publication; complete protected-cache entry/receipt/output reconciliation; canonical typed derivation permit/receipt evidence including resource-limit, evidence schema, manifests, digests, output paths, and next causal head; and a pluggable authoritative enforce-and-observe provider interface that refuses execution when no lossless provider is available. Preserve manager neutrality, global vendored compiled-binary denial, Kotlin exclusion, and all existing accepted identities. Do not claim Darwin Endpoint Security support.

## Acceptance Criteria
R2-R6 from TASK-260811-27xisf_review-verdict_RUN-260817-a83279.md are closed with security-negative tests. Competing stale permits start zero processes; only admitted immutable replay trees are readable; poisoned or wrong graph references never publish; tampered, duplicate, missing, extra, or substituted cache outputs never hit; manifest/vendor/mirror/metadata derivation records round-trip and drift fail canonically; unavailable observation providers fail closed. Focused, real-process/provider-contract, race, compatibility, full repository, vet, build, formatting, pinned lint, and canonical verifier gates pass. Independent reviewer accepts the task.
