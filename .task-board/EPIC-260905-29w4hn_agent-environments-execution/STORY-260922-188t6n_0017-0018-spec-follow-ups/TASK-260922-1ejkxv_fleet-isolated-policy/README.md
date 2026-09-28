# TASK-260922-1ejkxv: fleet-isolated-policy

## Description
F-S3 (0017 choice 6): a reviewed policy revision that lets a fleet enforce the isolated credential mode. Today system-config-v2 locks only toward shared; enforcing isolated fleet-wide needs an explicit lockable direction with the same locked-list semantics as manager §1, a conflict rule when a profile requests shared under an isolated lock, and diagnostics. No knob reinterpretation.

## Scope
curator-spec: system-config-v2 schema, manager §1 locked list, environments §12.2 lockable set, schema cases, vectors if demanded, CHANGELOG. No curator code (enactment = curator follow-up leaf named in results.md).

## Acceptance Criteria
1) system-config-v2 admits the isolated lock direction with schema cases valid/invalid; 2) normative conflict rule and diagnostic name for a profile requesting shared under an isolated lock; 3) make validate green; 4) CHANGELOG entry; 5) results.md names the curator follow-up leaf.
