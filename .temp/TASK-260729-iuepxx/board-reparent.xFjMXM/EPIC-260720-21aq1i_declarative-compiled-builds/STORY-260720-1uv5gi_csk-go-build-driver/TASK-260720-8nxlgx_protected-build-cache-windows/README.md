# Implement protected Windows build cache

## Description
Implement the Windows backend for the protected immutable build cache with DACL ownership, reparse-point containment, handle-safe access, and hard-link defenses.

## Scope
Own src/csk/builds/cache_windows.py and Windows-specific tests while preserving the backend contract from TASK-260720-2jfnz6. Use standard-library and ctypes Windows APIs as needed to create and verify a manager-owned cache boundary, inspect DACLs and file identity, reject reparse escapes and multiply linked files, and atomically publish entries. Keep the module import-safe on non-Windows hosts. Do not change portable receipt or cache-key semantics.

## Acceptance Criteria
Only the manager principal and trusted operating-system administrators may mutate the protected boundary and descendants. Pre-existing permissive or unverifiable roots, escaping reparse points, special files, hard-linked receipt or artifact files, ownership or DACL drift, and containment races force a miss or fail closed and never admit candidate bytes. Exact receipt, input, artifact path, hash, size, concurrent-winner, immutability, and dry-run rules match the POSIX backend. If protection cannot be established reliably, persistent reuse is disabled rather than opened. Windows CI exercises positive and negative backend cases, and full pytest plus strict mypy pass without non-Windows import regressions.
