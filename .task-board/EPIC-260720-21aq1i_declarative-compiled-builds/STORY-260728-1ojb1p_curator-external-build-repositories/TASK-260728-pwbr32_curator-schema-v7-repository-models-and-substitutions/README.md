# TASK-260728-pwbr32: curator-schema-v7-repository-models-and-substitutions

## Description
Add Curator domain models and strict parsing for schema-7 build_repositories, go-repository-v1 command references, curator-build.json schema 1 targets, and Skillfile.dev schema 2 operator substitutions. Preserve schema-6 go-v1 parsing and execution as a separate path.

## Scope
Go manifest/config models, validators, canonical declared/effective repository identity, URL/ref/tag/object-format lock parsing, target/build_root/source_dir containment, substitution resolution, and unit/property tests. No network fetch, audit, compiler, or installation mutation in this task.

## Acceptance Criteria
Models accept exactly the released rc.5 schemas and reject unknown execution/output/credential/signing fields; command keys own final executable names while repository targets select only closed build roots; canonical HTTPS/SSH and local-selector identities match shared vectors; immutable locks and optional tags are represented without mutable-ref fallback; schemas 1-6 and existing config behavior remain byte and behavior compatible; focused and full Go tests pass.
