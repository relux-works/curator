## Scope and provenance

- Task worktree: /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3lo9jc/curator-spec-worktree
- Exact base: origin/main at 57c1f56846d221ecc55786bd3c2467ec32f11730.
- Accepted rc.4 import source: /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1u7hes/worktree at the same base commit. The complete tracked and untracked accepted product diff was imported; .temp, virtual environments, generated caches, alternate indexes, binaries, task-board configuration, and unrelated files were excluded.
- Post-edit recursive comparison against the accepted source, excluding .git and .temp, reported differences only in cli/curator.md, conformance/README.md, and schemas/v1/README.md. Nothing was staged or committed.

## Documentation result

The schema index now provides a complete schema 6 declaration, canonical and legacy compatibility, exact receipt/marker/claim schemas and vectors, fixed go-v1 prerequisites, and the future-driver standardization process. The conformance contract now defines rc.4 manager evidence, fixture/vector execution, suite hash semantics, claim v1/v2 handling, and fail-closed behavior. The Curator CLI guide now covers install/upgrade builds, context exclusion, cache and marker currentness, compiler-free dry-run results, diagnostics, repair, GC, security limits, and the absence of package argv/hooks or generic build options without claiming an existing manager release ships schema 6.

## Verification

- git diff --check: passed.
- python3 tools/validate.py under a task-local venv: passed; validated 35 schemas and 189 vector files, including local Markdown links.
- make validate with PATH preferring the task-local venv: passed; validator passed, 27 Python unit tests passed, and go test ./tools/... passed.
- The initial system-Python attempt failed only because jsonschema was not installed. requirements-dev.txt pins jsonschema 4.25.1; installing that declared dependency in .temp/validate-venv resolved the environment prerequisite. No repository files were changed for the environment.
- No new test code was added because the task owns documentation only; the existing link validator, schema/vector validation, Python unit suite, and Go generator suite cover the changed scope.