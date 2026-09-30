# TASK-260917-2vapkz review verdict rev1 — CHANGES REQUESTED

## Verified (independent python hashlib from core.md §8 text)
- Colliding pair: v1 both c518e668…970c (equal); v2 576c3d37…551c vs de107e6a…28d4 (differ) — matches content-hashes-v2.json.
- Empty tree v2 = sha256("curator-content-v2\0") = f34a9d36…0c12 — matches.
- Mechanism: explicit `hash_version: 2` member on new carrier versions; frozen shapes absent => 1 (schemas/v1/README.md:15-25).

## BLOCKING — released schemas modified in place (schema policy)
Both files exist at tag v1.0.0-rc.13 and are changed in this CR:
1. schemas/v1/log-response-v2.schema.json:15-18 — `entries.items.$ref` switched from registry-log-entry-v1 to registry-log-entry-v2. Every rc.13-valid log-response-v2 instance (v1 entries) now fails validation: a breaking reinterpretation of a released schema. schemas/v1/README.md:33-40 itself states this revision adds no member and nothing is "widened, or reinterpreted"; the repo convention (README:13, :37, :47, :106, :185) is byte-frozen released files + new version. Fix: add `log-response-v3.schema.json` (refs registry-log-entry-v2), keep v2 byte-identical to 4ad8042b, move log-response-v2 schema-cases to v3, update registry.md/profiles/registry-service.md references.
2. schemas/v1/manager-config-v2.schema.json — adds `hash_version` to waiver state and reformats the whole file (whitespace churn on every line). Must be a new `manager-config-v3` (or justify in README with an explicit repo rule permitting additive change to a released schema); revert v2 to byte-identical and move the new cases (valid-v2-state-hash-waiver, invalid-waiver-hash-version, invalid-v2-state-waiver-commit-length) to v3.
3. Add a validate.py guard (with negative test in tools/test_validate.py) that fails when a schema present at the latest release tag differs byte-wise, so this class cannot recur.

Also: release/1.0.0-rc.13.json is modified — a published release record; revert unless the change is required by the tooling and explained.

Rest of the areas (registry equal-version match, NUL-opaque interim rule, generator-driven vectors, CHANGELOG) not blocking at first look; re-review after rework.
