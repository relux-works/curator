# BUG-261014-351fcr: migrated-global-skill-audited-under-v1-hash

## Description
migrateGlobalSkills (internal/envprofile/envprofile.go) audits the migrated revision with hashing.VersionV1 hard-coded (contextaudit.DetectAtVersion(..., hashing.VersionV1, ...)) while the resulting lock is written under hashing.WriteVersion() (v2). Under v1, core section 8 interim reader rule makes any file containing a NUL byte a blocking, non-waivable opaque-file finding, regardless of directory. Every gzip header carries a NUL, so a skill repository that tracks .tar.gz board attachments (or any binary) cannot be migrated off a legacy branch declaration: profile_source_invalid. The normal global add --revision path already audits under WriteVersion() (internal/contextresolve/contextresolve.go). Rejected remedies: a path exemption such as .task-board/** contradicts core section 8 (the rule applies regardless of directory) and would let any repository hide bytes under a magic directory name; no manifest declaration for opaque files exists for skill sources.

## Scope
Audit migrated global skills under hashing.WriteVersion() and record HashVersion on the migrated Resolved, matching the global add --revision path. Follow-up check, separate leaf if confirmed: the profile strict audit passes capabilities.ImplicitNone(), so a migrated skill whose scripts contain URLs may then fail audit.capability.network-undeclared.

## Acceptance Criteria
1. A hermetic test migrates a legacy branch-declared global skill whose commit tracks a gzip file and a binary; migration succeeds under v2 and the lock records hash version 2. 2. The same fixture under a forced v1 write version still refuses with the opaque-file finding (no silent weakening). 3. No path-based audit exemption is introduced. 4. The network-undeclared follow-up is either covered by a test or filed as its own leaf.
