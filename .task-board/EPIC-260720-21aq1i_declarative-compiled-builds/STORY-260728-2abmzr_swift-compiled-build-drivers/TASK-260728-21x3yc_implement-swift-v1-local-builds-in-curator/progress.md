## Status
closed

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(21))

## Blocked By
- TASK-260728-1yhuqi
- TASK-260728-2bu2q6
- TASK-260728-2gbtb9
- TASK-260720-jrrgw9

## Blocks
- TASK-260728-3j60e3
- TASK-260728-3lqm4z

## Checklist
- [ ] Curator compiles context-excluded vendored Swift sources through only swift-v1 and atomically installs the declared executable
- [ ] Trusted Swift toolchain, SDK, audit, offline dependency, cache, rollback and repair failures have no-mutation tests
- [ ] macOS-primary and Windows-native build and invocation gates pass with exact toolchain evidence

## Notes
Closed 2026-09-15 as superseded: the Curator-side Rust and SwiftPM (and Node/TS) adapters were delivered in August 2026 under EPIC-260810-271m92 (STORY-260811-2epsp4; commits f8b7cc7 and 6f93b51; internal/rustsource, internal/swiftpmbuild, internal/swiftpmsource, docs/authoring-language-adapters.md); Kotlin was explicitly deferred in STORY-260811-1tybyr. The July 'driver pair' plan is not being implemented in that shape, and the csk (ivanopcode/cocoaskills) halves are outside this delivery (no push admission). Operator confirmed 2026-09-15.

## Precondition Resources
(none)

## Outcome Resources
(none)

## Created
2026-07-28T09:15:27Z

## Last Update
2026-09-15T19:25:43Z
