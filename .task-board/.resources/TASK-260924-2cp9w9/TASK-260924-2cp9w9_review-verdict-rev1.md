# TASK-260924-2cp9w9 review verdict — CR rev1 — ACCEPTED

Reviewer: claude-opus-5-5 (low). Candidate tree a166d0bf over base 3d2c6114, verified from `git archive a166d0bf` into /tmp (disposable), not the control root. Context profile: full (role body present).

## Clauses
1. Registry (skillfile-sources.md §4 marker-attestation paragraph): positive evidence = conjunctive grant of name, canonical repository, locked commit, and raw frozen package-tree content_sha256 (explicitly NOT the lock-projected context hash). Revocation keeps the broader registry §3 match (content OR repository+commit) and is deny-wins under advisory as well as strict. Consistent with registry.md §3/§4 (lines ~106-129). Fail-closed. Existing wording "context hash" became "content hash", which fixes the ambiguity.
2. Marker v5: local go-v1 is a closed raw-JSON arm (external-only members rejected even when null); external declared_tag is optional, and when present it must be a non-empty Git ref name. Matches the schema's existing `gitRefName` $ref; no schema change.
3. Refresh (skillfile-sources.md §3): refresh uses the current resolved policy endpoint plan, and stored remote.origin.url is machine-private and never selects the endpoint. repository-transport.md §5: an SCP-like endpoint with an alias port fails with repository_policy_invalid before I/O (also added to the §6 table row), and the SSH URI + alias port rendering is stated. Fail-closed; the refusal choice is justified in the text.
Minimal: no status or namespace edits. CHANGELOG entry is under unreleased.

## Vectors → mutant each one kills
- invalid-local-{repository,substituted,substitution}-null: each is a single-member mutation of valid-local-mixed-builds. Kills "null == absent" decoding.
- invalid-external-declared-tag-empty / -invalid-refname ("release..candidate"): kills "any string" / "empty allowed".
- attested-network-current (projected hash differs from raw) + wrong-context (record matches projected hash only): kills "compare lock-projected context hash".
- wrong-name/-repository/-commit: now pinned as single-dimension mismatches under strict. Kills partial-grant readings.
- attestation-evidence-revoked (advisory, content-only match) + new -identity-commit-advisory: kills "revocation needs exact match" and "advisory ignores revocation".
- v2-refresh-current-endpoint-existing-checkout: kills "fetch stored origin".
- v2-scp-alias-port-refused vs v2-ssh-uri-alias-port: kills "guess ssh:// rendering" and also "refuse all alias ports".
Index and semantic-cases are consistent. No new section or family: all additions are inside the existing install-marker-v5 schema family and the existing semantic list.

## Gates rerun (zsh, set -o pipefail, clean copy with venv from requirements-dev.txt)
- Draft README gate: Schema cases 121/121; negatives 96/96; wire schemas 8/8; marker migration refusal mutants 18/18; snapshot 3/3; exit 0.
- tools/validate.py: validated 64 schemas / 1169 vector files, exit 0.
- go test ./tools/...: ok, exit 0.
- regenerate-check equivalent: regenerated in the copy, and `diff -r` of conformance/v1 and release/ against the candidate is clean.
- Bound: the Python unittest discover suite (~20 min) was NOT rerun.

## Lockstep impact (reported, not blocking)
Disposable curator clone at origin/main a48f584c, with the vendored internal/crossconformance/testdata/draft-sources-v1/corpus replaced by the candidate corpus. `go test ./internal/crossconformance/` FAILS:
- Pin/manifest/count failures (TestDraftSourcesPin, CorpusCounts 121 vs 115, SchemaCases, SemanticCases 98 vs 94). Part of this is PRE-EXISTING: curator vendors spec pin 802caee, which already lacks valid-no-builds.json and the base valid.json/semantic-cases. Any spec bump needs a curator re-vendor.
- NEW coverage gaps (TestDraftSourcesSemanticCoverage): no production-entry row for attestation-evidence-revoked-identity-commit-advisory (→ 10d3l1 registry leaf), v2-refresh-current-endpoint-existing-checkout, v2-scp-alias-port-refused, and v2-ssh-uri-alias-port (→ the curator transport/refresh leaf; none is named in the brief, so the orchestrator must assign one).
- The new marker v5 schema negatives (→ 2v4v2m, accepted, not landed) were not reached, because the count assertion fails first. Curator behaviour on those vectors is UNVERIFIED here.

## Hygiene
No root TASK-/BUG- files, test/ or ledger/ artefacts among the 10 changed paths.
