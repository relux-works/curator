# BUG-261017-8v470p: audit-cache-ignores-capabilities

## Description
internal/audit caches findings keyed by content hash, hash version, commit, directory and package identity, but findings come from detectWithOpaquePaths(snapshot, subject.Capabilities, ...). The same content audited under different effective capabilities (declared skill capabilities on the skill install path, implicit none on the profile and context paths) can reuse each other's cached findings, so a capability-dependent finding such as network-undeclared can be skipped or wrongly reported. A cross-lane review reproduced the crossover and offered a candidate correction that binds cached findings to all effective capability fields, keeps old records readable under the previous ruleset, and also rejects present null prompt_scope and null schema-3+ capability objects.

## Scope
internal/audit verdict cache key and records; internal/capabilities and internal/skillspec only as the spec requires; tests

## Acceptance Criteria
1. A cached verdict is reused only when the effective capabilities equal the subject's (all fields the detectors read). 2. Crossover controls: the same content audited with declared capabilities and with implicit none never reuses each other's findings, in both orders, across skill, context and MCP audit paths; identical inputs still hit. 3. Existing records keep their meaning under the previous ruleset; no silent reinterpretation. 4. Any null-declaration rejection is backed by the spec and paired with valid, absent and empty-object controls. 5. Narrowing mutants fail as intended; full remote gate green.
