# STORY-260905-5u97yt: snapshot-acquisition-byte-exactness

## Description
Step 2 (review M3): normative rule that snapshot production reproduces exact committed blob bytes (no working-tree conversion, no attribute-driven processing), plus a conformance vector with `* text=auto` and an export-subst entry hashing to the raw-blob value. Spec-only PR on curator-spec; the reference implementation (gitops.Archive -> object-database extraction) lands in implementation stage (a). Worktree ~/Developer/ReluxWorks/.worktrees/curator-spec-m3-byte-exact, branch draft/snapshot-byte-exactness, base b4f29cd.

## Scope
(define story scope)

## Acceptance Criteria
environments.md carries the byte-exactness rule; a generated vector with a text=auto + export-subst fixture exists whose expected hash equals the raw-blob content hash; make validate and regenerate-check pass; accepted by an independent reviewer; landed on main by fast-forward of the reviewed head.
