# TASK-260729-1b9tc3: design-csk-rc5-conformance-consumer

## Description
Design the exact independent Python conformance consumer and fixture strategy for CocoaSkills against the immutable rc.5 build-driver, lifecycle, execution-policy, schema, receipt, marker, and claim artifacts.

## Scope
Read-only design and executable test blueprint. No CocoaSkills product/test edits, no copied Go logic, no pin or publication changes, no broad test execution.

## Acceptance Criteria
Outcome maps every relevant rc.5 vector cluster to Python test modules and product boundaries, defines fixture loading and digest provenance, negative/non-alias coverage, platform gating and skip policy, and literal first-wave pytest commands for implementation.
