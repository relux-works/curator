# Implement canonical build metadata models

## Description
Implement portable logical build input, CCJ-1 cache keys, exact canonical receipt bytes and hashes, and install-marker v2 models while keeping physical csk cache paths implementation-specific.

## Scope
Own typed metadata modules under src/csk/builds, a dedicated install-marker model module, narrowly shared CCJ-1 support in src/csk/protocol_json.py, and focused tests. Model the complete go-v1 input including build source, root, command, source directory, native target, toolchain, and fixed policy. Parse and canonicalize receipt schema 1 and marker schema 2. Keep marker schema 1 readable for schema 1 through 5 installs. Do not implement filesystem trust, compiler execution, cache layout, status, or installer mutation.

## Acceptance Criteria
CCJ-1 input bytes derive the shared cache key sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48. Exact stored receipt bytes contain no BOM, whitespace, or trailing newline and derive the shared receipt hash sha256:750f5f7557ffe1dac1ec06b7d47cb5976e390a7dc7d57ea616c9edda91013c11. Readers reject duplicate keys, unsafe integers, noncanonical bytes, unknown fields, mismatched keys or input, wrong derived artifact paths, unsupported versions, and malformed identities. Marker v2 deterministically sorts roots, commands, dependencies, files, and builds; requires build_source exactly when builds are active; and keeps valid marker v1 current for pre-v6 packages. Focused pytest and strict mypy pass.
