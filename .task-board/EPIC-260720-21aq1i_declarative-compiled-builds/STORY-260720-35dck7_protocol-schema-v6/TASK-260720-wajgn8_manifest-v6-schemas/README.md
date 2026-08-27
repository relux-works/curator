# Add canonical and legacy manifest schema v6

## Description
Introduce strict manifest schema version 6 for both agent-skill.json and the legacy csk-skill.json read alias, together with their generated schema cases. Reuse existing script and system definitions, add a separate v6-only build union, and expose build_roots only in version 6 so each task revision keeps the repository validation gate green.

## Scope
Work in curator-spec origin/main commit 57c1f56846d221ecc55786bd3c2467ec32f11730. Own schemas/v1/common.schema.json, new agent-skill-v6.schema.json and csk-skill-v6.schema.json, their portions of tools/generate-vectors/main.go and main_test.go, their generated schema-case directories, index entries, and resulting manifest hashes. Do not broaden the existing common command used by schemas 1 through 5. Receipt, marker, and claim schemas remain separate tasks.

## Acceptance Criteria
Both v6 schemas require schema_version 6 and preserve v5 capabilities and dependencies; build_roots is v6-only and commandV6 accepts strict script and system objects plus exactly type build, driver go-v1, and source_dir; additionalProperties false rejects args, env, flags, tags, output, toolchain, scripts, hooks, and mixed shapes; canonical and legacy schemas differ only in stable id and title; generated valid and invalid cases cover the accepted build and structural rejection surfaces; schemas and generated cases for versions 1 through 5 are unchanged; go test ./tools/generate-vectors, make regenerate, make validate, and make regenerate-check pass.
