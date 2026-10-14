# BUG-261014-3oeu7d: migrated-global-skill-network-undeclared

## Description
migrateGlobalSkills runs strictAuditMember, whose audit subject carries capabilities.ImplicitNone(). A migrated legacy global skill whose scripts contain a URL therefore fails audit.capability.network-undeclared (high, blocking) and cannot migrate: profile_source_invalid: migrated global skill X audit blocked: high audit.capability.network-undeclared - contacts HOST, which no capabilities.network entry covers (file: scripts/fetch.sh). Pinned by TestMigratedGlobalSkillNetworkUndeclaredFollowUp in internal/envprofile. Legacy global skills had no capability declaration surface, so the migration cannot honestly supply one; the open question is what the spec (environments 9.1 strict audit and 9.4 migration) requires: refuse with an actionable diagnostic naming the declaration to add, admit the skill's own manifest capabilities if it carries any, or something else. Spec first if the text is silent.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
(define fix acceptance criteria)
