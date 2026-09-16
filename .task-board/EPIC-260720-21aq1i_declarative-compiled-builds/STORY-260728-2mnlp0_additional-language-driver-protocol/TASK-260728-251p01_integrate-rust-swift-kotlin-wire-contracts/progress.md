## Status
closed

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(21))

## Blocked By
- TASK-260728-12pnm1
- TASK-260728-1yhuqi
- TASK-260728-168smo
- TASK-260728-2jaw7h

## Blocks
- TASK-260728-2bu2q6

## Checklist
- [ ] Schemas, docs and vectors encode Rust, Swift and selected Kotlin local plus repository drivers with structured toolchain requirements
- [ ] Negative cases reject command, path, environment, hook, plugin and generic build-system escape fields while retaining audit-before-build semantics
- [ ] Legacy schema and Go behavior remain compatible and deterministic full-generation gates pass

## Notes
Closed 2026-09-15 as superseded: the Curator-side Rust and SwiftPM (and Node/TS) adapters were delivered in August 2026 under EPIC-260810-271m92 (STORY-260811-2epsp4; commits f8b7cc7 and 6f93b51; internal/rustsource, internal/swiftpmbuild, internal/swiftpmsource, docs/authoring-language-adapters.md); Kotlin was explicitly deferred in STORY-260811-1tybyr. The July 'driver pair' plan is not being implemented in that shape, and the csk (ivanopcode/cocoaskills) halves are outside this delivery (no push admission). Operator confirmed 2026-09-15.

## Precondition Resources
(none)

## Outcome Resources
(none)

## Created
2026-07-28T09:09:38Z

## Last Update
2026-09-15T19:24:55Z
