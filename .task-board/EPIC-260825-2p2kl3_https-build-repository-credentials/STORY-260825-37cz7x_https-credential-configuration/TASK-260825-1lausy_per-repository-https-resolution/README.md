# TASK-260825-1lausy: per-repository-https-resolution

## Description
Resolve HTTPS credentials per external build repository in the install path, mirroring the SSH resolver in internal/install/buildssh.go. Precedence: the run-wide environment override, then the longest matching build_https scope, then anonymous. The override MUST be bindable to a host (core 12.2): with the pin set, only repositories on that host receive it and every other repository resolves as if the override were absent. Carry the captured override in a type whose secret is excluded from its diagnostic representation. Absence of any selection is NOT an error for HTTPS: anonymous HTTPS is a real transport and public repositories must keep working; this is deliberately asymmetric with SSH and the reason belongs in a comment. Each selected source must fail closed with a message naming the exact remedy when its material is missing.

## Scope
(define task scope)

## Acceptance Criteria
Precedence, host pin, anonymous fallback and transport skip covered by tests; the three fail-closed remedies tested; the captured override never renders its secret; go test green.
