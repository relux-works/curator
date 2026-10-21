# Review brief: audit Markdown network scan limited to agent-facing context (BUG-261021-2qv4zj, curator)

Review the branch in the request at the exact head, against the producer brief `W-AUDITSCOPE-brief.md` and the results file `BUG-261021-2qv4zj_results-r1.md` (both board resources). This is a Gate A regression fix and goes into rc.8. Sweep every surface; don't stop at the first hit.

| Surface | Attack |
| --- | --- |
| Scope correctness | Markdown is scanned exactly where curator-spec core §3 makes content agent-facing. That means the eligible roots (including `scripts/` only when the package exports no commands), minus the §3 exclusion patterns at every depth, minus whole `runtime_roots`/`build_roots`. Probe each eligible root, each exclusion pattern at depth (for example `references/x/README.md`, `assets/tests/a.md`) and nested build roots. A detector that misses a root is a security regression, not a nicety. |
| Single source of truth | The eligibility must come from the same code that context assembly and `skill.build_root_in_prompt_context` use. A second copy that can diverge is a finding. Check that context assembly behaviour is unchanged. |
| Non-Markdown scan | The `scripts/` and manifest scanning are byte-for-byte the old behaviour. Any widening or narrowing is a finding. |
| Ruleset and cache | Ruleset 6: confirm the bump is needed under the CIP-0021 round-7 cache-reuse rule, and that a stale cached refusal from ruleset 5 is invalidated, not reused. A stale *allowance* must not survive either. Run `TestReviewPriorMarkdownAllowance` and its neighbours. |
| CIP-0021 witness change | The historical contact witness moved to `references/notes.md` with a declared-host control. Confirm it still proves the original CIP-0021 claim and was not weakened to pass. The vendored corpus is unchanged. |
| Regression and controls | The regression install (`.research/`, `.task-board/.resources/`, build-root vendor READMEs) admits. The positive controls (`SKILL.md`, `references/`, `assets/`) are still refused, and a declared host passes. Rerun the narrowing mutants yourself: "every .md" fails the regression, and dropping `references/` fails the positive control. |
| End to end | If feasible, install a package shaped like project-management through `curator profile install <path>` in a temporary HOME and show it admits. A temp-HOME install of a package with an undeclared URL in `SKILL.md` must refuse. |
| Hygiene | No personal paths in committed files; signed commits; the default suite is green at the exact head (private GOCACHE, deleted afterwards). |

Disk: use a private per-run GOCACHE and delete your task scratch after the handover.
