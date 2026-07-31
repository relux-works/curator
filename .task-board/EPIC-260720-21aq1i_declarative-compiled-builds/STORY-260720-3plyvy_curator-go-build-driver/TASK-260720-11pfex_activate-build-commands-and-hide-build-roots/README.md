# Activate build commands and exclude build roots

## Description
Carry parsed build declarations through dependency closure, command selection, collision checks, context installation, and skill check so compiled commands behave like exported script commands while their source subtrees never become prompt-visible or runtime-copied.

## Scope
Own focused changes in internal/closure, internal/whitelist, and internal/skillcheck with unit and conformance tests. Full and runtime edges may activate script and build commands but never system commands; narrowed requirements may name an exported build command. Preserve provider-first closure order and define lexical command ordering per provider. Pass the union of runtime_roots and build_roots to static context exclusion before locale rendering, including inactive, cache-hit, and dry-run paths. Add author warnings when prompt-visible Markdown references excluded build source paths. Do not compile, cache, write markers, or refactor the installer in this task.

## Acceptance Criteria
Full activation exports every script and build command and runtime narrowing accepts either kind; unknown or system-only narrowed names fail and active command collisions cover build versus build and build versus script; provider order remains deterministic and build names are bytewise lexical within a node; every build_root subtree is absent from copied context and runtime output while SKILL.md and unrelated eligible assets remain present; build-root exclusions apply even when no compiler command runs; skill check reports stable warnings for prompt references to excluded build sources and keeps existing runtime warnings; closure, whitelist, and skillcheck tests pass.
