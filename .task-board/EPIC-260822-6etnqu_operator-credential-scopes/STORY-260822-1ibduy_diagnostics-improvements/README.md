# STORY-260822-1ibduy: diagnostics-improvements

## Description
Actionable diagnostics: dependency-closure manifest errors name the skill and ref they come from; toolchain executable mismatch for version-manager shims (goenv/asdf/mise) carries an operator remedy pointing at the real GOROOT/bin.

## Scope
(define story scope)

## Acceptance Criteria
Error texts carry provenance/remedy; protocol error codes unchanged; tests assert the new texts; go test green.
