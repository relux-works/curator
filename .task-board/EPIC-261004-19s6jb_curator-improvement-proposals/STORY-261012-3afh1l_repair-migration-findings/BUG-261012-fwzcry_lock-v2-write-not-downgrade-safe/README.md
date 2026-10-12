# BUG-261012-fwzcry: lock-v2-write-not-downgrade-safe

## Description
rc.4 and later write the profile lock as schema 2 with hash_version 2 in place. v0.14.0 decodes the lock with DisallowUnknownFields and requires schema 1 (contextlock.go:291,295-296 at v0.14.0), and then reports a misleading profile_unknown: profile is not installed (managed.go:1510-1513 at v0.14.0). On a shared host with an older Curator first on PATH, one repair makes every launch through the older Curator fail; this happened on the host on 2026-10-12 and the lock was restored by hand. The spec (environments.md section 1.3 and section 13) has no rule for older managers reading lock v2 and no guidance on the backup or rollback of a one-way lock migration.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
Spec states the compatibility and rollback rule for lock v2 (for example a backup of the v1 lock beside it, or an explicit one-way migration that requires confirmation); the implementation follows it; older-reader failures are documented; the new release reports an unreadable newer lock with a precise code rather than profile_unknown going forward
