# TASK-260922-1t2w1q: envprofile-explicit-credential-migration

## Description
F-C2: Explicit credential migration step inspect then plan then apply under the manager lock, never inside resolve --repair: inventory the old marker, link targets and both Pi roots ~/.pi/auth.json and ~/.pi/agent/auth.json; preserve the effective mode on upgrade; the operator resolves isolated to shared conflicts out of band; no secret copies at any step; the plan is printed before applying and the apply is journaled and rollback-safe.

## Scope
internal/envprofile and the env CLI surface; docs/troubleshooting; CHANGELOG

## Acceptance Criteria
a Pi wrong-target home migrates to ~/.pi/agent with bytes preserved and mode intact, printing the plan before applying; conflicts refuse naming the exact operator choice needed; a row asserts that no secret bytes are copied; a narrowing mutant per refusal
