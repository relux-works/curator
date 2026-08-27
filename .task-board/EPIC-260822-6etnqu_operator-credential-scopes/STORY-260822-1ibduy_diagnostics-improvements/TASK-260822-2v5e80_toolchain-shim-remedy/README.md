# TASK-260822-2v5e80: toolchain-shim-remedy

## Description
The toolchain executable mismatch error for version-manager shims (goenv/asdf/mise wrappers outside a real GOROOT/bin) keeps its protocol string but gains an operator remedy note: put the real GOROOT/bin first on PATH, e.g. PATH="$(go env GOROOT)/bin:$PATH". Follow the existing pattern used by the fingerprint-deadline error.

## Scope
(define task scope)

## Acceptance Criteria
Remedy text asserted in tests; protocol string byte-identical; go test green.
