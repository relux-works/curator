# TASK-260916-3bc4fo: structured-error-injection-property

## Description
Inject PermissionError, OSError, RecursionError, RuntimeError and UnicodeDecodeError at each filesystem call site and assert the public entry point returns a structured refusal, never a raw exception.

## Scope
error surface of source expansion and readers

## Acceptance Criteria
Property green for every injected type at every site
