# TASK-260916-bn5kvb: declare-umbrella-requires-mcp

## Description
After relux-mcp lands and is tagged, declare requires.mcp in packages/relux-root-context-ivan/agent-context.json (relux-root-context repository): figma and safari from git@github.com:relux-works/relux-mcp.git with ranges ^1.0 (Decision 0012 D6 form, same shape as the existing requires.skills entries), update the umbrella README and scripts/validate.sh expectations, keep every other package byte-unchanged. Validate with the curator context parser oracle as B1 did.

## Scope
(define task scope)

## Acceptance Criteria
Umbrella manifest declares requires.mcp for figma and safari with ^1.0 ranges resolvable against relux-mcp tags; validate.sh and the parser oracle pass; no other package changes.
