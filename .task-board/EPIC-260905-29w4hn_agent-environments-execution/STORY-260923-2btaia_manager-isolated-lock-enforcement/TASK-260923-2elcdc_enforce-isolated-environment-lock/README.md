# TASK-260923-2elcdc: enforce-isolated-environment-lock

## Description
Enforce the isolated environment lock in the environment manager (curator-spec fe2d1c6, F-S3): system-config-v2 environments.isolation admits isolated; silence resolves to the locked direction; an explicit shared under an engaged isolated lock refuses with environment_isolation_lock_conflict; an already-provisioned shared passthrough fails closed toward the explicit F-C2 migration (never silently migrates). Spec text: environments section 12.2, manager section 1 rule 1.

## Scope
(define task scope)

## Acceptance Criteria
1) isolated lock admitted from system-config-v2 with manager section 1 locked-list semantics; 2) silence resolves to isolated under the lock; 3) explicit shared under the lock refuses with environment_isolation_lock_conflict through the real CLI entry; 4) provisioned shared passthrough fails closed and names the migration; 5) one narrowing mutant per refusal killed; 6) CHANGELOG
