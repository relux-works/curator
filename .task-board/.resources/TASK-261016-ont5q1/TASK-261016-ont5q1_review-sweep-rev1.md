# M1 revision 1 review surface sweep

Candidate tree: dc879a2d8126093d786c98a99c18e531829407a2.
Base: 34a89308a69d4863699887359cd122d64b1e716f.
The brief supplies no catalog table or time budgets. Use the six mandatory paired rows as the surface floor; bounded review targets 20 minutes including hosted gate, with up to five minutes additional free hunt. No source modifications.

Coverage checked before code review: producer names tests for 4 of 4 AC rows and 6 of 6 paired rows. Inspected the named tests and actual production callers. Consumer readiness and runtime qualification are explicitly outside this leaf.

## R1 — Required root verification

Result: held. Public resolve/status tests `TestMuseRootContextDrift`, `TestMuseRootContextRefusals`, `TestMuseRootContextVectors`, `TestMuseRootContextCopyFallback` passed, including required denial, failed repair, unchanged marker/foreign bytes, link/copy/current/repair controls. Narrowed changed/removed/denied-root gates all generated named behavioral failures. Reviewer test `TestM1ReviewRootForgery` removed both surface record and root, then retried twice; same-byte retarget was also refused. Status remained non-current and marker unchanged. Logs: focused-01.log and attacks-01.log. Claims are bounded to executed synthetic fixtures on darwin/arm64.

## R2 — Undeclared and header-only roots

Result: held. `TestMuseRootContextVectors/no-context` places unrelated native/workspace header-like files and still receives no fact; monolithic and zero-applicable controls emit the fact. `unrelated-root-mints-fact`, `header-only-root-fact-dropped`, and `no-context-admitted` all killed through the CLI tests. Logs: focused-01.log; row2-01.log is a targeted replay.

## R3 — Closed fragment boundary

Result: held. Existing schema corpus and `TestMuseFragmentBoundary` passed. Reviewer `TestM1ReviewBoundary` injected six corrupt proposed objects immediately before the real `CheckBoundary` call in `buildFragment`, while leaving the boundary intact: extra name, keychain literal, wrong suffix, outside parent, HOME, wrong adapter. Public `env resolve` rejected all six with the expected diagnostic text, empty stdout and unchanged marker. Initial unpoisoned provisioning succeeded. This proves production wiring as well as helper behavior. Logs: boundary-01.log, focused-01.log. The overlay changes only disposable source copies, not the candidate.

## R4 — Renderer parity

Result: held. `TestMuseCLIStorePolicy` compared all five parsed JSON/env/shell values and independent order. `renderer-drops-only-backend` kept the other four values and was killed by that named production CLI test, child exit 1. Log: focused-01.log.

## R5 — Shared-store liveness and non-destructive refusal

Result: held. Published liveness vectors and `TestMuseAuthRepairAdditionalStates`, `TestMuseForkRefusedRegression`, `TestMuseMissingNativeFilePolicy` passed. Fork, retarget, absent/dangling target, nonempty directory, metadata/readlink denial exercise public `Resolve`/`StatusOf` with preservation assertions and live/absent-link/empty-directory controls. Missing-source and readlink-denial narrowing mutants generated the expected named failures. Log: focused-01.log. No real credentials or live refresh were exercised.

## R6 — Machine selector and isolation

Result: held. `TestMuseCLIStorePolicy`, `TestMuseCredentialPolicyMatrix`, and `TestMuseCredentialPolicyStatus` passed through production config/resolve/status entry points. Unset/native/file map to file, stored native remains native, keychain refuses without writes, isolated creates no native link. Keychain-only admission, isolated-link and shared-status narrowing mutants were killed. Reviewer `TestM1ReviewStatusUnsupported` changed a provisioned selector to keychain: non-current status reported the unsupported-store diagnostic without retaining a valid file-policy warning. Logs: focused-01.log, attacks-01.log.

## Free hunt

No blocking finding reproduced. The additional marker-forgery/retry, identical-byte retarget, selector-change and six production-boundary poisoning probes are recorded above. No proof of absence is claimed. Runtime exclusion qualification, login/refresh and Windows ACL behavior are not established by this review.
