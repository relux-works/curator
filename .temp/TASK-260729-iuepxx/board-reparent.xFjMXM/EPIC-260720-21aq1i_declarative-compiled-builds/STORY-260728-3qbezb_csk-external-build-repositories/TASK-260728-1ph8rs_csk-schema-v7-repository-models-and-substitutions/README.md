# TASK-260728-1ph8rs: csk-schema-v7-repository-models-and-substitutions

## Description
Add independent Python csk models and strict validation for schema-7 external build repositories, go-repository-v1 command references, curator-build.json targets, and Skillfile.dev schema 2 substitutions. Preserve existing script and local go-v1 behavior.

## Scope
cocoaskills Python manifest/config parsing, declared/effective canonical identity, immutable lock and safe ref/tag grammar, target/path containment, operator substitution selection, and unit/property tests. No Git acquisition, audit, build, or install mutation yet.

## Acceptance Criteria
csk accepts exactly the released rc.5 schemas and rejects package-controlled argv, environment, outputs, executable naming, credentials, signing, hooks, and unknown fields; command key and target ownership match the spec; canonical network/local identities and substitutions match shared vectors; schemas 1-6, Python scripts, and local go-v1 paths do not regress; focused and full Python validation passes.
