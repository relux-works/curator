# TASK-260916-5aqozl: pin-pnpm-via-corepack-in-ci-lanes

## Description
Operator decision 2026-09-16: every CI lane (hosted ubuntu/macos/windows Test and Race, and the self-hosted rose-air lane) installs the pinned pnpm (internal/pnpmsource SupportedPNPMVersion = 10.33.0) through corepack after setup-node and puts it first on PATH, so the real-pnpm integration tests run everywhere instead of skipping, and a broken pnpm shim on a self-hosted runner PATH (macbook-iv: /Users/iv/.cache/codex-runtimes/.../fallback/pnpm, node SyntaxError on set -euo pipefail) can no longer fail the lane.

## Scope
(define task scope)

## Acceptance Criteria
.github/workflows/ci.yml: after actions/setup-node in each lane a step `corepack enable --install-directory <lane temp dir>` + `corepack prepare pnpm@10.33.0 --activate` and that directory prepended via GITHUB_PATH; the pinned version is read from one place (a workflow env or the Go constant via a small script) so it cannot drift; hosted gate green with the three internal/pnpmsource real-pnpm tests EXECUTED (not skipped) on all three OSes; skip-classes untouched; a comment explains the macbook-iv shim incident.
