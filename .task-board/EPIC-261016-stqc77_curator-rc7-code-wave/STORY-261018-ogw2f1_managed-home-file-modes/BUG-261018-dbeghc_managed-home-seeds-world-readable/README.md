# BUG-261018-dbeghc: managed-home-seeds-world-readable

## Description
Managed homes are provisioned under an environments root created 0755, and managed writes use mode 0644 (internal/envprofile/managed.go atomicManagedFile calls, including seed provisioning). A native seed that is 0600, such as the Codex config carrying project trust levels and hook state, becomes group- and other-readable in the managed home. Under per-agent OS users this exposes one agent's managed home to the others. Found during a native-home transition rehearsal.

## Scope
internal/envprofile managed-home provisioning and repair writes, environments root and managed home directory modes; tests; docs note. No spec change in this leaf (stricter modes are compatible); a separate spec clarification follows.

## Acceptance Criteria
1. A seed keeps the native file permission bits and is never group- or other-readable when the native file is 0600; no managed seed is wider than its source. 2. The environments root and every managed home directory are created 0700; repair tightens an existing wider directory without losing contents. 3. Other manager-written files in a managed home use modes from one documented table; trust-bearing or credential-adjacent files are 0600. 4. Negative tests: a 0600 native seed never yields a readable-by-others copy; a pre-existing 0755 home is tightened by repair; a symlinked or unreadable seed still refuses as today; a mutant that writes 0644 again fails for that reason. 5. Full local suite and full hosted gate green on the exact signed candidate.
