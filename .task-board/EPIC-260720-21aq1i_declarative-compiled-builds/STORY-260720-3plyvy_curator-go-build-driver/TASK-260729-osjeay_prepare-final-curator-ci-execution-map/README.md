# TASK-260729-osjeay: prepare-final-curator-ci-execution-map

## Description
Read-only refresh of the final Curator compiled-build CI task against the accepted rc.5 candidate, current workflow, Makefile, platform runners, and current task dependencies so implementation can start without another discovery phase.

## Scope
Inspect TASK-260720-1pvfj5, current Curator candidate and accepted comparison, .github/workflows/ci.yml, Makefile, Go/tool versions, macOS/Windows/Linux runner constraints, and existing task evidence. Produce an exact file-level producer plan, conflict-free edit ownership, candidate-input/pin invariants, platform matrix, and narrow/full validation commands. Do not edit product/spec files, CI, pins, task 1pvfj5, or run heavy tests.

## Acceptance Criteria
Outcome identifies all stale rc.4 wording and dependency drift, exact files and YAML/Make targets, immutable rc.5 candidate evidence inputs, Linux/macOS/Windows and race/vet/gofmt/lint gates, native prerequisite handling, and the smallest producer/reviewer sequence. Every claim is source-verified and no product or pin mutation occurs.
