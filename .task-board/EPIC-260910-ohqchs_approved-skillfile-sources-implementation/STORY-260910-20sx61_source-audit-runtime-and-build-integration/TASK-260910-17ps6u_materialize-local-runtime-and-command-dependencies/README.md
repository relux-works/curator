# TASK-260910-17ps6u: materialize-local-runtime-and-command-dependencies

## Description
Materialize local runtime and command dependencies. Extend current implementation after inspecting existing outcomes; this task is not authorization to start work.

## Scope
internal/runtimestore, globalbins, capabilities, install, adapters.

## Acceptance Criteria
Install scripts/assets and dependency runtimes from immutable local snapshots with existing command, capability, protected runtime-store and shim semantics. Exercise an actual local skill with runnable script and dependency; refresh uses frozen replacement, never live links or arbitrary new hooks.
