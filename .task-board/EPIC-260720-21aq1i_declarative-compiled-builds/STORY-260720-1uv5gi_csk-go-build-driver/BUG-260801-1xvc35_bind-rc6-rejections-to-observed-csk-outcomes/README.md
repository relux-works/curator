# BUG-260801-1xvc35: bind-rc6-rejections-to-observed-csk-outcomes

## Description
Repair PR19 cycle-2 P1 rejection coverage. From exact signed CocoaSkills base ba250bfc4dfe104a160eadd5b5f4e340693bf892, make all 77 build-driver rejection cases execute the relevant CocoaSkills validator/backend with concrete vector data or independently materialized condition fixtures. Derive observed error, result, reuse, artifact execution, command and cache effects from execution; never return a hard-coded expected table value without observing that exact product outcome.

## Scope
Own rejection binding helpers and focused tests in the PR19 conformance harness. Preserve accepted env/argv, manifest, build-source/toolchain and Windows projection repairs. Add exhaustive mutation and sabotage probes proving altered condition, unrelated exceptions, wrong error codes and omitted backend inspection cannot pass. Work in a dedicated CocoaSkills worktree/branch based exactly on ba250bf; produce a signed commit for later integration into PR19, not a release or main landing.

## Acceptance Criteria
All 77 names are exhaustive and every expected vector field is compared to an observed CocoaSkills trace. The reviewer probes from TASK-260720-12r55p_review-verdict-cycle-2.md fail before the fix and pass after it: 75 condition mutations are rejected, unrelated SkillSpecError is not accepted for unknown-driver, untrusted_go_executable is not accepted for wrong-go-executable-path, and artifact-hash-mismatch reaches real cache inspection. Focused exact-root tests, strict mypy and diff checks pass; no pin, schema-v7, tag, release or claim change.
