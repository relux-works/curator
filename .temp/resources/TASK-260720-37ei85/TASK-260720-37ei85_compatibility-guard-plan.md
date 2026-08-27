# Compatibility guard plan — TASK-260720-37ei85

**Task:** `TASK-260720-37ei85` — Add legacy schema compatibility guards  
**Repository baseline:** `curator-spec origin/main` at `57c1f56846d221ecc55786bd3c2467ec32f11730`

## Frozen surfaces

The compatibility task must compare semantic structure and regenerated evidence, not only top-level file hashes:

- `agent-skill-v1` through `agent-skill-v5` and `csk-skill-v1` through `csk-skill-v5` keep the existing script/system-only command union and reject `build_roots` plus `type: build`.
- `install-marker-v1` keeps its historical fields and meaning.
- `conformance-claim-v1` stays at schema version 1 and protocol `1.0.0-rc.3`; its valid/invalid generated cases stay byte-semantically stable.
- `manager-config-v1` and `system-config-v1` gain no driver, argv, environment, toolchain, output-path, hook, or build-policy override surface. Fixed go-v1 semantics remain protocol-owned.
- Registry and audit-record schemas gain no local build artifact attestation or receipt-provenance field. Registry evidence continues to cover the untrusted source snapshot.

## Expected comparison result

New schema/case/vector files and the generated suite manifest will legitimately change the rc.4 inventory and manifest hashes. The guard must distinguish those additions from a structural change to any frozen wire contract or legacy case outcome. Two independent regeneration passes must leave no diff.

## Verification

Run and record:

```text
go test ./tools/generate-vectors
make regenerate
make validate
make regenerate-check
make regenerate
make regenerate-check
```

Any detected legacy semantic change routes back to its owning schema task; this task does not rewrite frozen schemas to make a regression pass.

