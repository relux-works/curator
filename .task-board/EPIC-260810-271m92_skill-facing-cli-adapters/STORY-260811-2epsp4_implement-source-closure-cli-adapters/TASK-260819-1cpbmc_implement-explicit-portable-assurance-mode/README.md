# TASK-260819-1cpbmc: implement-explicit-portable-assurance-mode

## Description
Turn the accepted closure execution substrate into an explicit portable assurance mode while preserving the future verified provider seam.

## Scope
Add configuration parsing and defaults for execution.mode portable and explicit verified selection; portable enforcement and declared-output verification; typed actual-capability and receipt evidence; cache and checkpoint identity separation by mode and provider contract; policy rejection when verified is required; stable diagnostics for missing or unhealthy providers; no silent upgrade or downgrade. Keep provider interfaces platform-neutral. Preserve closure graph, TOCTOU, publication, cache reconciliation, binary denial, and Kotlin exclusion.

## Acceptance Criteria
Default CLI execution succeeds in portable mode using only guarantees it actually establishes. Receipts and cache keys encode portable mode and exact capabilities and cannot be accepted as verified. Explicit verified mode starts no process without a compatible healthy provider. Unknown modes, providers, capability drift, claim inflation, and cross-mode cache reuse fail closed. Focused, race, compatibility, full Go, lint, build, verifier, binary-deny, and Kotlin-exclusion gates pass; independent review accepts.
