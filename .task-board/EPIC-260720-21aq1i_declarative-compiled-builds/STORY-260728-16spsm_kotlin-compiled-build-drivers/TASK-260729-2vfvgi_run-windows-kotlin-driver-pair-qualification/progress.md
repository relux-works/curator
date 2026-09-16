## Status
closed

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(13))

## Blocked By
- TASK-260728-168smo
- TASK-260728-1koh5v
- TASK-260728-gmfxdg

## Blocks
- (none)

## Checklist
(empty)

## Notes
JUSTIFIED GAP (created by TASK-260728-168smo, solution-architect, rework cycle 2).

Missing piece: a per-tuple Kotlin/Native qualification run on windows/amd64. The story contained a Linux qualification task (TASK-260728-3u1nho) and no Windows one.

Requirement that would otherwise stay unimplemented: decision 0008 section 9 admits a (driver, os, arch) tuple into a claim only with immutable native evidence for that exact tuple; decision 0007 section 1.3 and section 4 allow the reserved kotlin entry to carry primary_relpath, probe, platforms and compatibility only once a host is qualified; STORY-260728-16spsm AC requires platform and shared gates to pass.

Consequence of leaving it open: decision 0010 measured macOS as permanently unsupported for this pair (seven host executables outside every curatable root, unavoidable by any manager-fixed input, plus an unfingerprinted Xcode SDK in the cache key), and Linux is outside the protocol platform set until TASK-260728-1skseh. windows/amd64 is therefore the only tuple that can admit the pair. With no task owning it, TASK-260728-251p01 would retire both Kotlin identifiers by default rather than by evidence, and retirement is one-way for the whole of Protocol 1.0 because decision 0008 section 2 closes the identifier space.

How this element closes it: a Windows-host qualification sibling to TASK-260728-3u1nho, running decision 0010 section 12 A1-A9 on windows/amd64 and supplying the registry and allow-list values from that evidence only.

Self-verification before creation: decision 0007 sections 1, 1.3, 3, 4; decision 0008 sections 3, 6 item 3, 9, 10; STORY-260728-16spsm scope and AC; the decision 0008 section 9 out-of-scope list (macOS and Windows remain the portable-policy platforms, only Linux is excluded until TASK-260728-1skseh). Result: Windows is in scope and unaddressed, so the addition is required rather than invented. TASK-260728-2bu2q6 was checked and does not cover it: it qualifies the spec candidate (digests, gates, honest empty native claims), not a per-tuple host run.

Dependency note: a link TASK-260728-251p01 blocked_by this task was attempted and rejected by the board as a cycle. The obligation is instead carried normatively in decision 0010 section 14.

Host note: TASK-260729-rhjxtx measured that the reachable Windows host carries no Kotlin toolchain of any backend. That is not a blocker: decision 0010 section 3 makes the toolchain an operator-curated bundle, and section 1.3 of the reference gives the curation procedure.
Closed 2026-09-15 as superseded: the Curator-side Rust and SwiftPM (and Node/TS) adapters were delivered in August 2026 under EPIC-260810-271m92 (STORY-260811-2epsp4; commits f8b7cc7 and 6f93b51; internal/rustsource, internal/swiftpmbuild, internal/swiftpmsource, docs/authoring-language-adapters.md); Kotlin was explicitly deferred in STORY-260811-1tybyr. The July 'driver pair' plan is not being implemented in that shape, and the csk (ivanopcode/cocoaskills) halves are outside this delivery (no push admission). Operator confirmed 2026-09-15.

## Precondition Resources
(none)

## Outcome Resources
(none)

## Created
2026-07-29T00:04:43Z

## Last Update
2026-09-15T19:25:28Z
