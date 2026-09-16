## Status
closed

## Review
required

## Task Class
docs

## Estimate
estimated(fibonacci(5))

## Blocked By
- TASK-260728-r3j8ef
- TASK-260728-1aveb2
- TASK-260728-ypbuav

## Blocks
- (none)

## Checklist
- [ ] Author docs provide validated local build_roots and external skill-build.json examples for the selected Kotlin driver
- [ ] Operator docs cover trusted runtime and toolchain preflight, offline dependencies, errors, cache, repair and platform limits
- [ ] Examples contain no generic Gradle or arbitrary command escape and pass documentation or conformance validation

## Notes
Closed 2026-09-15 as superseded: the Curator-side Rust and SwiftPM (and Node/TS) adapters were delivered in August 2026 under EPIC-260810-271m92 (STORY-260811-2epsp4; commits f8b7cc7 and 6f93b51; internal/rustsource, internal/swiftpmbuild, internal/swiftpmsource, docs/authoring-language-adapters.md); Kotlin was explicitly deferred in STORY-260811-1tybyr. The July 'driver pair' plan is not being implemented in that shape, and the csk (ivanopcode/cocoaskills) halves are outside this delivery (no push admission). Operator confirmed 2026-09-15.

## Precondition Resources
(none)

## Outcome Resources
(none)

## Created
2026-07-28T09:16:06Z

## Last Update
2026-09-15T19:25:13Z
