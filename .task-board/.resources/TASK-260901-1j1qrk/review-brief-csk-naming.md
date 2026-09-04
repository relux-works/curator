# Review brief: csk surface naming sweep

Subject: curator-spec worktree `~/Developer/ReluxWorks/.worktrees/curator-spec-csk-naming` (`draft/csk-surface-naming`, head 4d55698, base f8d7e7a) + curator story worktree `.temp/STORY-260901-2zdg81/worktree` (head 5914a9dd, base 979fa36e, LOGBOOK only). Inventory resource `csk-naming-inventory.md` on TASK-260901-1j1qrk — the review surface IS the classification.

1. **Completeness**: rerun the inventory grep yourself in both worktrees (and curator docs/README/CLI help specifically); every hit must appear in the inventory with a category; hunt for surface hits the classification may have miscategorized as wire (e.g. prose AROUND identifier mentions with csk-flavored wording, doc headings, human diagnostic strings in curator Go code).
2. **The one rewrite**: confirm `.csk-build.json` at core.md:1644 is purely illustrative (no schema, vector, fixture, or implementation references it — grep both repos and the tags), and the replacement `.agent-build.json` reads correctly in context.
3. **Untouched wire**: spot-check 3+ categories (fixtures, schema $ids, tools) byte-identical to base.
4. Gates: `make validate` green in the spec worktree; curator worktree `go build/vet` (LOGBOOK-only delta, but run them); signatures on both commits; LOGBOOK entry accurate and appends cleanly (no heading damage — check the previous entry survived).

Verdict: `review-findings-csk-1.md`; blocking/major -> development; else ACCEPT + accept_cr.
