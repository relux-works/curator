# TASK-260916-1qfpu4: spec-nofollow-write-rule

## Description
curator-spec: add the normative rule that takeover and repair writes replace the directory entry and never follow a symlink at the target path (O_NOFOLLOW-class semantics) to environments §8.3/§9.5; add a vector with a symlinked target.

## Scope
protocol/environments.md §8.3/§9.5, conformance vectors

## Acceptance Criteria
Environments revision merged with the rule and the symlinked-target vector
