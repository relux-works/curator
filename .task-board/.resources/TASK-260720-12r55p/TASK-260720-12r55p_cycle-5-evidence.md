# TASK-260720-12r55p cycle 5 developer evidence

Signed CocoaSkills PR19 head: 6e7742f0d28ad95ddd7d8e92364b84062571ad0b. PR: https://github.com/ivanopcode/cocoaskills/pull/19. Candidate source: curator-spec 432eb2ee1fe2d6b271e37269f867c8851c325539 with manifest sha256:12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071. No workflow/release pin, tag, Release, or claim changed.

Hosted run 30737293076 exposed 408 Windows failures cascading from one GC fixture os.utime follow_symlinks NotImplementedError. The repair centralizes the supported fallback, covers it without an OS skip, and reuses it at both timestamp call sites. ssh win was unavailable: bounded probe exit 255 and Tailscale reported the host offline.

Standalone gates: focused fallback plus representative GC vector exit 0, 2 passed; strict mypy exit 0, 68 files; exact-root conformance exit 0, 1028 passed and 1 expected POSIX-only skip; full pre-repair exact-root suite exit 0, 2297 passed and 55 expected platform skips; post-change and post-commit python -m build exit 0; exact versioned Twine checks exit 0; commit signature, diff whitespace, unchanged .github, clean tracked tree, and push each exit 0.

Evidence-honesty notes: two exploratory release-metadata reads exited 1 before the corrected nested-key check exited 0; several gh run watchers were manually interrupted and exited 1; an unbounded ssh probe was interrupted and exited 130. These were diagnostics, not green gates.

Fresh exact-head hosted run 30743353816 has mypy green and all 12 OS/Python test lanes active at handoff. Reviewer must require its terminal exact-head result before acceptance.