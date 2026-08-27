# TASK-260729-iuepxx: audit-language-story-dependency-cycle

## Description
Independently reconstruct and diagnose the dependency cycle among Kotlin, Swift, Rust, toolchain-preflight, and additional-language-protocol stories that prevents the compiled-build epic plan from being computed.

## Scope
Board graph and the requirements encoded by existing story/task blockers only. Produce the smallest semantically correct unlink/link proposal that preserves real implementation and qualification ordering. Do not change dependencies, task status, product artifacts, release state, or worktrees.

## Acceptance Criteria
The outcome enumerates every edge participating in the cycle with its rationale/source, distinguishes hard prerequisites from sequencing/advisory relationships, proves the proposed graph is acyclic, preserves the product priority of finishing language specifications before Curator/CocoaSkills implementations, and gives exact board mutations for orchestrator review.
