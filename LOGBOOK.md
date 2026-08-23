
### 1307 — Ratified: operator credential selections are never lockable

`build_ssh` stays outside LockableKeys by decision of the spec owner (2026-08-23): credential material is operator-owned, and a system configuration must not select or constrain it. The manager-profile system-configuration clause now states this explicitly; the peer implementation records the same rule at its lockable-keys definition.
