# BUG-261012-10fd8w: repair-migrates-current-homes

## Description
env resolve --repair with a hash-v1 lock calls repairUnderLock before the current-home check (managed.go:2523-2524 at v0.15.0-rc.4; same at rc.5 and main). prepareIdentityMigration then rewrites lock.json and the marker of every registered harness, including siblings outside --env (identity_migration.go:85,170,184-228). This contradicts environments.md section 8.2 (upgrade v1/v2 markers only when a marker publication is otherwise required) and section 10.1 (a current home under --repair emits its fragment without touching any lock). Expected: a current home under --repair emits its fragment and leaves the lock and the markers unchanged; the identity migration runs only through an explicit migration command or when a publication is required anyway.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
--repair on a current home with a hash-v1 lock leaves lock.json and every marker byte-identical; a regression test covers a sibling harness outside --env; migration remains reachable through env migrate --apply
