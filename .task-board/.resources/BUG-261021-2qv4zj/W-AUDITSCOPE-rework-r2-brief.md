# Rework brief r2: audit Markdown scope, case-alias bypass (BUG-261021-2qv4zj, curator)

Review r1 (`BUG-261021-2qv4zj_review-r1.md`, board outcome resource) requested changes at `2c01da83`. Read it in full first; the finding and its witnesses are there. Continue on branch `fix/audit-markdown-scope`.

## P1 F1: case-insensitive root aliases bypass the scan

On a case-insensitive filesystem (default macOS), `References/guide.md`, `SKILL.MD` and seven neighbouring aliases are skipped by the new audit predicate (`ContextPathIncluded`), but `CopyContext` still publishes them into agent context. A strict/high CLI install exits 0 and the undeclared URL reaches native Claude skill context. The base detector blocked all nine. This is a security regression; fix the class, not the nine names.

## Required invariant

**The audit-scanned Markdown set is a superset of what context assembly can publish, on every filesystem.** Both must be computed by one predicate with one name-normalisation rule. A second matcher, even a case-folded copy, is a finding. Pick one of these and justify it in the results:
- (a) the audit enumerates exactly the entries `CopyContext` would publish (the same walk and predicate), then scans their Markdown; or
- (b) both copy and audit apply one shared, case-folding eligibility and exclusion predicate.

Either way, check the exclusion side too: an exclusion pattern must not exclude from the audit something that copy still publishes. `readme.md`, `Tests/` and `ReadMe.MD` inside an eligible root must behave the same in both. If an exclusion matches more loosely in copy than in audit, the audit is the side that has to be stricter.

## Cache

Cached ruleset-6 allowances must not survive the correction. Bump `RulesetVersion` to 7, or prove by test that a ruleset-6 allowance for a case-alias package is re-evaluated and refused. Prove it with a test seeded from a real ruleset-6 cache entry.

## Tests (each test states its claim)

- Regression witnesses from review r1: all nine aliases are refused by the gate and through the exact CLI install (strict/high). No lock is published, and no URL bytes appear in native context.
- Positive controls stay: `.research/`, `.task-board/.resources/` and build-root vendor READMEs still admit; canonical `SKILL.md`, `references/` and `assets/` are still refused; a declared host passes.
- Exclusion parity: for each §3 exclusion pattern, case variants behave identically in copy and audit.
- Mutants: restoring the case-sensitive predicate fails the alias regression; "every .md" still fails the `.research` regression.
- Run on the real case-insensitive volume (the default). If feasible, also on a case-sensitive APFS disk image or tmpfs; state which runs you did.

## Run

Build, vet, the focused packages, and the full default suite in a normal-.git validation clone with a private GOCACHE deleted after the run. Delete task scratch after the handover.

## Hand-off

`BUG-261021-2qv4zj_results-r2.md` attached as an outcome resource, branch pushed with signed commits, handover in topic `audit-md-scope`.
