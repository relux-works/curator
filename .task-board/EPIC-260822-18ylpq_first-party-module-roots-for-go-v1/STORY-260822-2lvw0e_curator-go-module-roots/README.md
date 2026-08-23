# STORY-260822-2lvw0e: curator-go-module-roots

## Description
curator (Go): implement declared module roots in the go-v1 driver once the spec story lands — manifest parsing for the modules list, bijection validation against go.mod replace directives, path containment and disjointness checks, scan-surface extension over declared directories, diagnostics, and the shared conformance vectors green on all platform lanes. Fully autonomous per the maintainer pre-authorization of 2026-08-22. Final step: unblock skill-project-management — its board task TASK-260822-hje0ya (auto-return) switches that manifest to go-v1 module roots; the prepared diff waits on branch task/go-v1-switch at origin.

## Scope
(define story scope)

## Acceptance Criteria
Module-roots conformance vectors green on ubuntu, macos, and windows lanes; cross-implementation CI green; skill-project-management auto-return task explicitly unblocked in its notes.
