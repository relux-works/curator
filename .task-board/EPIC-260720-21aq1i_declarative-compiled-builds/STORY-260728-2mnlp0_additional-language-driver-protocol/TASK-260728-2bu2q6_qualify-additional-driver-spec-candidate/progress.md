## Status
closed

## Review
required

## Task Class
metadata

## Estimate
estimated(fibonacci(8))

## Blocked By
- TASK-260728-251p01

## Blocks
- TASK-260728-2gbtb9
- TASK-260728-1j72zq
- TASK-260728-q283m8
- TASK-260728-13ioo0
- TASK-260728-21x3yc
- TASK-260728-2lnhci
- TASK-260728-1koh5v
- TASK-260728-gmfxdg

## Checklist
- [ ] Candidate and generated artifact digests are independently recomputable from a clean tree
- [ ] Complete schema, vector, compatibility and deterministic double-regeneration gates pass
- [ ] Release metadata, downstream pins and platform claims accurately reflect only verified capabilities

## Notes
Linux qualification host now available temporarily via ssh lev. Add non-destructive Linux validation in private temporary directories when this task enters execution, after inventory, while preserving macOS-primary and Windows-via-ssh-win gates.
Closed 2026-09-15 as superseded: the Curator-side Rust and SwiftPM (and Node/TS) adapters were delivered in August 2026 under EPIC-260810-271m92 (STORY-260811-2epsp4; commits f8b7cc7 and 6f93b51; internal/rustsource, internal/swiftpmbuild, internal/swiftpmsource, docs/authoring-language-adapters.md); Kotlin was explicitly deferred in STORY-260811-1tybyr. The July 'driver pair' plan is not being implemented in that shape, and the csk (ivanopcode/cocoaskills) halves are outside this delivery (no push admission). Operator confirmed 2026-09-15.

## Precondition Resources
(none)

## Outcome Resources
(none)

## Created
2026-07-28T09:09:38Z

## Last Update
2026-09-15T19:24:58Z
