# TASK-260811-13xlp0: implement-shared-source-closure-foundation

## Description
Implement the language-neutral adapter contracts and source-closure policy foundation selected by the accepted architecture decision.

## Scope
Add the shared closure graph and manifest model, immutable source and toolchain identity inputs, compiled-artifact classification and fail-closed diagnostics, offline build boundary, checkpoint serialization, and adapter registration seams. Preserve existing Go behavior and keep ecosystem-specific dependency resolution out of this task.

## Acceptance Criteria
The common API represents recursive and mixed-language closure edges; checkpoint inputs are deterministic and include source, dependency, build, target, and toolchain identity; prohibited compiled artifacts and undeclared inputs fail closed with stable diagnostics; existing Go behavior remains compatible; focused unit and regression tests pass.
