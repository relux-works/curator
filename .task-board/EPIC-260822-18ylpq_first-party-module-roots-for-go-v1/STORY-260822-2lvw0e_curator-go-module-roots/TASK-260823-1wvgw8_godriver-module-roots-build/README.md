# TASK-260823-1wvgw8: godriver-module-roots-build

## Description
curator (Go): extend the go-v1 driver to build with declared module roots per the candidate prose: validateModule accepts replaced first-party modules exactly when the replacement maps onto a declared module root (bijection already validated at parse time — driver re-verifies fail-closed), declared directories join the directive/cgo/assembly scan surface, vendored resolution otherwise unchanged, diagnostics per spec. Behavioral tests using the module-roots vector family from the candidate suite. Worktree from origin/main. Fully autonomous per the 2026-08-22 pre-authorization; every lane green pre-merge. Executor: claude only.

## Scope
(define task scope)

## Acceptance Criteria
go-v1 builds a multi-module vendored fixture with declared roots; module-roots behavioral vectors consumed and green; merged to main green.
