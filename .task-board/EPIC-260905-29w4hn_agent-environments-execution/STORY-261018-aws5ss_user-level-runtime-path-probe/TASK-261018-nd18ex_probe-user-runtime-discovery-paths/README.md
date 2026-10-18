# TASK-261018-nd18ex: probe-user-runtime-discovery-paths

## Description
Probe user-level discovery for Codex 0.162.0, Codex 0.163.0 and the installed Claude Code version with fixed scratch HOME and overridden runtime config directory. Record exact binary/source provenance; skills, instructions, plugins and hooks paths; paired positive and negative sentinels; no model requests, real-home writes or runtime behavior changes.

## Scope
One bounded read-only discovery matrix and reproducible scratch harness. Initial budget 90 minutes. Upstream owner decides implementation and landing.

## Acceptance Criteria
Report every runtime/surface verdict with positive-control status and explicit static-versus-observed evidence; include fixed-HOME config override and shifted-HOME control; record unsupported/unavailable paths without claiming negatives; preserve real-home state; produce public-safe signed PR for upstream owner review.
