# TASK-260720-37ei85 compatibility blocker

## Constraint and evidence

The task requires both `agent-skill-v1.schema.json` and
`csk-skill-v1.schema.json` to reject `build_roots`, while also requiring legacy
schemas 1 through 5 to retain their frozen wire semantics and explicitly
forbidding this task from rewriting those schemas.

The frozen `origin/main` baseline is
`57c1f56846d221ecc55786bd3c2467ec32f11730`. On that baseline, both manifest-v1
schemas have `additionalProperties: true`. Their `allOf` gates forbid only
`runtime_roots`, `dependencies`, and `capabilities`; `build_roots` is not
forbidden. The accepted rc.4 composite from
`/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-2zc6k1/worktree`
preserves those two schemas byte-for-byte.

Validation with the predecessor task's installed Draft 2020-12 validator
produced:

```text
agent-skill-v1: build_roots_valid=True type_build_valid=False
agent-skill-v2: build_roots_valid=False type_build_valid=False
agent-skill-v3: build_roots_valid=False type_build_valid=False
agent-skill-v4: build_roots_valid=False type_build_valid=False
agent-skill-v5: build_roots_valid=False type_build_valid=False
csk-skill-v1: build_roots_valid=True type_build_valid=False
csk-skill-v2: build_roots_valid=False type_build_valid=False
csk-skill-v3: build_roots_valid=False type_build_valid=False
csk-skill-v4: build_roots_valid=False type_build_valid=False
csk-skill-v5: build_roots_valid=False type_build_valid=False
```

`go test ./tools/generate-vectors` passes on the accepted rc.4 composite, but a
truthful compatibility test for the stated v1 rejection would fail immediately.

## Failed assumptions

- Treating absence from `properties` as rejection is invalid when
  `additionalProperties` is `true`.
- A structural test that merely checks `build_roots` is undeclared would make
  the acceptance criterion appear satisfied while the real JSON Schema
  validator accepts the field.
- No separate manifest semantic-validation layer exists in this specification
  repository that could enforce the rejection without changing the wire
  schema or adding a new behavioral contract outside this task's scope.

## Viable options

1. Amend both v1 schemas with an explicit `not` gate for `build_roots` (and any
   other reserved build-only top-level fields). This satisfies rejection but
   changes frozen v1 semantics and violates the task's no-legacy-schema-rewrite
   constraint.
2. Preserve v1 exactly and narrow the requirement to schemas 2 through 5;
   assert that v1 continues its deployed extension behavior while all versions
   reject `type: build`. This preserves the baseline but requires correcting
   the acceptance criterion and the new rc.4 protocol sentence that currently
   says schemas 1 through 5 reject `build_roots`.
3. Define a separate reader-level reserved-field rejection rule for schema 1
   and add behavioral conformance evidence. This preserves the JSON Schema but
   expands reader semantics and behavioral-vector scope, so it needs an
   explicit architecture/product decision and ownership assignment.

## Recommendation and decision needed

Choose option 2 if frozen wire compatibility is the governing requirement.
Choose option 1 only if reserving newly introduced names is intentionally
allowed to tighten schema 1. The exact decision needed is whether v1's deployed
`additionalProperties: true` behavior remains authoritative or may be narrowed
for `build_roots`; implementation cannot honestly satisfy both requirements.

## Worktree and verification state

A task-scoped curator-spec worktree exists at
`/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-37ei85/worktree`, based
on the frozen SHA above and populated with only the accepted predecessor product
diff. No compatibility-task code changes were made, staged, or committed.

Not run because the stop-the-line constraint prevents a valid implementation:
`make regenerate`, `make validate`, and both deterministic
`make regenerate-check` passes.
