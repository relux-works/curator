## Status
closed

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(21))

## Blocked By
- TASK-260728-21x3yc
- TASK-260728-1j72zq
- TASK-260720-3s27te

## Blocks
- TASK-260728-3lqm4z

## Checklist
- [ ] csk compiles context-excluded vendored Swift sources with swift-v1 and publishes the same command contract as Curator
- [ ] Trusted toolchain, SDK, audit, offline, cache, rollback and repair semantics pass csk regression gates
- [ ] macOS and Windows native evidence matches the accepted protocol without implementation-specific escapes

## Notes
Closed 2026-09-15 as superseded: the Curator-side Rust and SwiftPM (and Node/TS) adapters were delivered in August 2026 under EPIC-260810-271m92 (STORY-260811-2epsp4; commits f8b7cc7 and 6f93b51; internal/rustsource, internal/swiftpmbuild, internal/swiftpmsource, docs/authoring-language-adapters.md); Kotlin was explicitly deferred in STORY-260811-1tybyr. The July 'driver pair' plan is not being implemented in that shape, and the csk (ivanopcode/cocoaskills) halves are outside this delivery (no push admission). Operator confirmed 2026-09-15.

## Precondition Resources
(none)

## Outcome Resources
(none)

## Created
2026-07-28T09:15:27Z

## Last Update
2026-09-15T19:25:52Z
