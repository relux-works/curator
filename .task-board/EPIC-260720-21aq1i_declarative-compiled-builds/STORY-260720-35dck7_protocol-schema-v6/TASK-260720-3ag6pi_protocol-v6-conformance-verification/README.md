# Verify integrated protocol v6 conformance

## Description
Perform the final curator-spec integration verification after all protocol rc.4 changes land. Regenerate from a clean state, prove byte stability, run every validation and release gate, and capture a coverage matrix showing that the accepted build contract and legacy compatibility are represented by executable conformance evidence.

## Scope
Work in curator-spec after all implementation, vector, validation, documentation, and release-metadata tasks. This is a conformance verification task, not a new feature-design task. Run the repository-supported commands and attach task-scoped logs and a concise outcome report. Make only narrowly scoped corrections to generated inventory or test expectations; route substantive contract defects back to the owning task.

## Acceptance Criteria
make validate passes with all Python and Go tests and no skips introduced; make regenerate followed by make regenerate-check leaves no diff, and a second independent regeneration produces byte-identical conformance/v1 output; make release-check VERSION=1.0.0-rc.4 passes; the generated manifest contains every new schema case, fixture expected file, build-driver vector, and lifecycle vector with correct hashes; a baseline comparison proves agent-skill and csk-skill schemas 1 through 5, install-marker-v1, and conformance-claim-v1 semantics remain unchanged; the outcome report maps every STORY-260720-35dck7 acceptance criterion and every minimum rejection cluster to a passing schema case or vector; failure logs contain no package-provided code execution and no release evidence is fabricated.
