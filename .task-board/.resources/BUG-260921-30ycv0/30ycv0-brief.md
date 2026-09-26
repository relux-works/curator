# BUG-260921-30ycv0 brief (orchestrator, binding)

Story STORY-260915-3w11un (cross-PR Story: publish a Change Request as usual; the orchestrator
lands it by PR). Origin: BUG-260920-3vfwch results.md §1 run 35340496757 and §2 B3 — the only
bare-error git spawn on the install path is `internal/gitops` `writeBlobs` (`git -C <repo>
cat-file --batch`), whose `cmd.Start()` error is returned raw (`cmd.Dir` unset), so a
product-path spawn failure surfaces in `install.Result.Errors` as the naked
`fork/exec /opt/homebrew/bin/git: permission denied` with no operation context.

Rulings:
R1 Wrap/classify the `writeBlobs` spawn (and `Wait`) errors the way `gitops.run` does for its
invocations — operation name + argv shape, sanitized (no new path disclosure beyond what
`gitops.run` already reports); keep the error class the callers already map (do not invent a
new refusal class; check how `closure`/`install` classify gitops errors and keep that mapping).
R2 No retry in product code. No behaviour change on the success path; the frozen v1 lane is
affected only in the wording of a failure that was previously unformatted — declare it in
CHANGELOG `## Unreleased` → `### Fixed` and keep the legacy goldens green.
R3 Evidence: a row that drives `Extract` with an injected non-executable git (PATH-resolvable
0755 script whose shebang interpreter is 0644, the shape the flake has) and asserts the wrapped
message; a mutant that restores the bare return → row fails; unchanged behaviour rows for the
success path (existing Extract tests). Windows: declared skip only if a sibling row already
skips for the same reason (ledger vocabulary, `.github/ci/skip-classes.tsv`).
R4 If `gitops` has (or the closure lane has) a spawn-observability seam that could carry the
fixture's diagnostic facts (binary/dir stat, rlimits), say whether adding it is cheap; do NOT
add product diagnostics that print environment or paths beyond current sanitization rules
without a ruling — record it as a bound instead.
results.md: the before/after message at the production entry (`install.Project` over a git
source with the injected git), mutant table, ratio line if any corpus row changes (none
expected). Publish the Change Request only when the configured gate is green.
