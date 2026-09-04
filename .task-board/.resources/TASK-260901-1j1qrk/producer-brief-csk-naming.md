# Producer brief: csk surface naming uniformity

## Scope (operator-narrowed, Decision 0010 D4)
Surface-level ONLY: prose, documentation, and human diagnostics in `relux-works/curator-spec` and `relux-works/curator` stop spelling `csk` except where they name a frozen §1.1 wire identifier (`.csk-install.json`, `.csk-managed.json`, `CSK_PROJECT_ROOT`, legacy alias `csk-skill.json`, schema filenames `csk-skill-v*.schema.json`). Wire identifiers, schema files/ids, vector fixture bytes, conformance expected outputs, release manifests: UNTOUCHED. No alias/retirement/deprecation work.

## Setup
Two worktrees from fresh origin/main of each repo:
- `~/Developer/ReluxWorks/.worktrees/curator-spec-csk-naming`, branch `draft/csk-surface-naming` (curator-spec main must be f8d7e7a or later);
- `~/Developer/ReluxWorks/curator/.temp/STORY-260901-2zdg81/worktree` if the board CR machinery provisions it, else `~/Developer/ReluxWorks/.worktrees/curator-csk-naming`, branch `draft/csk-surface-naming` from curator main (979fa36e or later).
Signed commits; do not push.

## Method
1. Inventory: `grep -rn -i csk` per repo; classify every hit wire-frozen (name of a frozen identifier / schema file / fixture byte / generated vector) vs surface (prose sentence, doc heading, CLI help text, human diagnostic wording, comment). Attach the classified inventory as a board resource — that IS the review surface.
2. Rewrite surface hits only: neutral or agent-* wording per Decision 0010 D4 ("the marker name deliberately joins the agent-skill.json family"). A sentence that NAMES a frozen identifier keeps the identifier spelling but may lose surrounding csk-flavored prose. Do not rename Go identifiers/packages/tests in curator that would change behavior — code identifiers are out of scope unless they leak verbatim into user-facing output (then change the output string, not the identifier).
3. Gates: curator-spec `make validate` green; curator `go build ./... && go vet ./... && go test ./... -count=1` green (or the repo's make targets).

## Deliverables
Signed commit(s) per repo; board resource `csk-naming-inventory.md` (classified hit list with dispositions); handoff to-review.
