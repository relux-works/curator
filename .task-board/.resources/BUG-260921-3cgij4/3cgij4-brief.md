# BUG-260921-3cgij4 brief (orchestrator, binding) — curator-spec corpus fixture

Repository: curator-spec (this control root). Story STORY-260921-atwkfi (corpus follow-ups)
under the Skillfile-sources epic; the board is shared with curator.

Defect: `conformance/draft-sources-v1/schema-cases/install-marker-v5/valid.json` (index
`valid: true`) has `"builds": {}` together with a top-level `build_source`. The normative rule
in `protocol/skillfile-sources.md` (~line 268): "Top-level `build_source` remains required
exactly for active local `go-v1` commands and absent otherwise". With zero builds the document
is normatively invalid although JSON-Schema-valid (the cross-field rule is not expressible in
the schema). The curator reader refuses it correctly (`internal/marker` `validBuildState`), and
the curator harness keeps this case as a documented bound instead of driving it
(BUG-260920-2eg8nv results: 111 driven / 4 bounds; discrepancy text attached as
valid-json-discrepancy.txt).

Operator decision (2026-09-21): do NOT weaken the rule. Fix the corpus:
1. `valid.json` becomes a marker WITH an active local `go-v1` build: add a consistent `builds`
   entry (the closed v5 shape: `package` + `lock_sha256` etc. as the schema and the prose
   require; mirror an existing valid v5 example with a local go-v1 command in the corpus if
   one exists) so its top-level `build_source` is legitimately required.
2. Add a sibling POSITIVE fixture (e.g. `valid-no-builds.json`) with `builds` absent/empty AND
   no top-level `build_source`, covering the "absent otherwise" branch; and, if the corpus has
   the pattern, a NEGATIVE fixture for the exact defect (empty builds + top-level
   `build_source` → invalid) so the rule is pinned by the corpus itself.
3. Update the case index accordingly (ids, `valid` flags, comments citing the rule), run the
   repository validation (`make validate` — the configured gate) and any generator/consistency
   checks the corpus has.
4. results.md: the exact rule citation, the before/after documents, the index delta, and the
   curator follow-up to name: after the next spec pin promotion, curator converts its bound to
   a driven row (`internal/crossconformance` schema cases; the curator SPEC_PIN is a fixed
   commit, so this fix takes effect there only after a pin bump).
Rules: no normative text changes beyond a clarifying comment if the prose is ambiguous (say
so in results.md if you think it is); publish the Change Request only when the gate is green.
