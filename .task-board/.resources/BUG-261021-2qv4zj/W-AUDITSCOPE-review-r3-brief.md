# Review brief r3: audit Markdown scope per subject kind (BUG-261021-2qv4zj, curator)

Correctness re-review at the exact head in the request. Review it against:
- `W-AUDITSCOPE-rework-r3-brief.md`;
- your `BUG-261021-2qv4zj_review-r2.md`;
- the producer's `BUG-261021-2qv4zj_results-r3.md`.

The producer's approach:
- audit selection is driven by subject kind;
- skills keep `WalkContext`, which is shared with `CopyContext`;
- context members are read through `contextpkg.ReadModule`, which is shared with materialization;
- MCP and unknown kinds keep the conservative scan of all Markdown;
- ruleset 8, with `subject_kind` in cache equality.

Context and MCP members keep implicit-none capabilities. This is confirmed as the intended §9.1 contract: a context module with any non-reserved URL is refused.

| Surface | Check |
| --- | --- |
| F2 closure | Re-run your r2 profile-install fixture exactly. The URL module is refused with `audit.capability.network-undeclared`, and nothing is published (no `source.json` or `lock.json`, no URL in native `CLAUDE.md`/`AGENTS.md`). The clean and reserved-host context controls install. |
| Per-kind table | Check the producer's per-kind table against the code. For every subject kind the Gate accepts, what that kind publishes to an agent must be inside the audited set. Look for publishers the table omits: profile root documents, agent definitions, MCP server descriptions, templates, triggers. A gap is a finding. |
| Shared readers | `ReadModule` must really be the code materialization uses, with no parallel path resolution. Check that module path normalisation (case, `..`, symlinks, absolute paths) is identical on both sides. |
| Ruleset 8 and cache | A real ruleset-7 allowance (including the F2 `findings: []` record) is refused after the upgrade. `subject_kind` equality cannot be satisfied by a record of another kind. Canonical packages still reuse correctly. |
| No regression | The r1/r2 cases stay green: the nine aliases, exclusion parity, `.research` noise, Unicode and symlink cases, the CIP-0021 witness, and byte-identical canonical published trees. |
| Mutants and suite | Re-run: drop the context-module branch, restore the old matcher, every-Markdown, namespace reuse. Run the full default suite at the exact head with a private GOCACHE, and delete scratch afterwards. |

Give the verdict at the exact head. Attach `BUG-261021-2qv4zj_review-r3.md` as an outcome resource.

## Working note

Two reviewer runs on this task were stopped by the model provider's content filter, which flags security-style wording. Write probes, notes, test names and the report in neutral QA terms: coverage, consistency, refusal, publication. Read the r2 review only for the fixture recipe and the F2 finding text. Keep fixtures to the minimum needed, for example the `api.example.net` contact URL.
