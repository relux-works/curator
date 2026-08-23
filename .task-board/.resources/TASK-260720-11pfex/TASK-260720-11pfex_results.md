# TASK-260720-11pfex implementation evidence

## Provenance
- Isolated worktree: /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-11pfex/worktree
- Base: exact origin/main 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8
- Imported predecessor: complete accepted product diff from TASK-260720-2g0e3b
- External conformance input: TASK-260720-3ag6pi candidate root only; not release or pin evidence
- No commit or staging performed

## Task delta
- Closure exports script and build commands for full/runtime activation, rejects unknown and system-only narrowing, keeps system commands inactive, detects build/build and build/script collisions, and exposes provider-local bytewise lexical active command order.
- Whitelist computes a sorted static union of runtime_roots and build_roots, prunes excluded subtrees, and shares the same path boundary with prompt-visible Markdown discovery.
- Project/global installation pass that union into context copying before locale rendering. An inactive context-only build provider and dry-run compiler sentinel prove no Go compiler execution, no build-root runtime copy, and no prompt-context leak.
- Skill check emits stable skill.build_root_in_prompt_context warnings for POSIX and Windows source references, ignores Markdown inside excluded build roots, preserves existing runtime warning behavior, and applies the shell-neutral resolver contract to build commands.
- Added authoritative rc.4 conformance coverage for provider-first plus lexical build ordering and the go-build-skill context file/hash golden.
- No compiler, cache publication, marker schema, or live installer transaction implementation was added.

## Verification
- go test ./internal/closure ./internal/whitelist ./internal/skillcheck ./internal/install ./internal/interop: pass
- CURATOR_CONFORMANCE_ROOT=TASK-260720-3ag6pi/conformance/v1 go test ./internal/closure ./internal/whitelist ./internal/skillcheck: pass
- go test ./...: pass after initializing the exact tracked testing submodule at 21585d0e937cae47e54a788d8ae36b1780eae47f
- CURATOR_CONFORMANCE_ROOT=... go test -race ./internal/closure ./internal/whitelist ./internal/skillcheck ./internal/install: pass
- go test -cover focused packages: closure 73.8 percent, whitelist 88.2 percent, skillcheck 89.7 percent, install 75.9 percent; changed branches are directly covered
- gofmt check, git diff --check, go vet ./..., and go build ./...: pass
- Linux amd64 and Windows amd64 curator plus closure, whitelist, skillcheck, and install test binaries compile successfully

## Known external anomaly
The full repository run with the entire rc.4 candidate enabled still fails only internal/interop TestManagerLifecycleVectors because that downstream consumer has not yet modeled the candidate build_order_cases and related compiled lifecycle fields. This exact downstream mismatch was already recorded by accepted predecessor TASK-260720-2g0e3b. Task-owned rc.4 consumers pass, and this task does not widen scope into the downstream lifecycle installer refactor.