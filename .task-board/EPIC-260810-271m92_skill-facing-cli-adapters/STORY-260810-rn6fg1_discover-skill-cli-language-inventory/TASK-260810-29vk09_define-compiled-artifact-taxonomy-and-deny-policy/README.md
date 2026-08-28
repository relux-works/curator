# TASK-260810-29vk09: define-compiled-artifact-taxonomy-and-deny-policy

## Description
Define the shared artifact taxonomy and fail-closed policy that prohibits vendored compiled executable code.

## Scope
Classify source, generated source-like text, native executables, object files, static and dynamic libraries, frameworks, native extensions, JVM bytecode, WebAssembly, archives, trusted toolchains, and locally built outputs. Specify recursive archive inspection, detection limits, diagnostics, and the future verified-binary capability boundary.

## Acceptance Criteria
The policy gives deterministic admission or rejection for every artifact class, rejects compiled dependency payloads across all adapters, distinguishes trusted external toolchains and locally produced protected outputs, documents ambiguous cases and conservative defaults, and defines stable diagnostic and audit evidence requirements.
