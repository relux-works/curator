# EPIC-260712-78cc4d: bootstrap-and-ci

## Description
Phase 0 of docs/implementation-plan.md. Go module, curator CLI skeleton with --version, Makefile, GitHub Actions matrix (ubuntu, macos, windows) with gofmt, go vet, golangci-lint, and a case-insensitive CI grep enforcing the naming rule from the plan. CONTRIBUTING.md carries the working agreements.

## Scope
See docs/implementation-plan.md, Phase 0. Stories and tasks are created under this epic as work starts.

## Acceptance Criteria
- Green CI on all three OS targets
- curator --version prints a version
- Naming-rule grep wired into CI
