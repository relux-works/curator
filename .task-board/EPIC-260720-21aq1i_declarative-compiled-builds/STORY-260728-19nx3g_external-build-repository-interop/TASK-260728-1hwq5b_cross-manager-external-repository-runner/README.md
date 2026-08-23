# TASK-260728-1hwq5b: cross-manager-external-repository-runner

## Description
Extend the black-box interop runner to execute the shared rc.5 corpus against released Curator and csk binaries as independent processes and compare protocol-required bytes, state transitions, filesystem effects, diagnostics, and rollback outcomes.

## Scope
Runner adapters, isolated homes/repos, deterministic network/local fixture serving, command execution and output normalization, filesystem/receipt/marker/cache assertions, failure injection, macOS/Windows execution, and machine-readable reports. Do not import either manager implementation as a library.

## Acceptance Criteria
The same cases run against both released binaries with isolated state; canonical identities, digests, receipts, markers, and stable typed errors compare exactly where normative, while implementation-private physical paths remain unconstrained; runner detects extra processes/network/writes and mutation on failed install; positive and adversarial cases pass on macOS and Windows; reports include exact binary/spec/toolchain/OS revisions and preserve failure artifacts.
