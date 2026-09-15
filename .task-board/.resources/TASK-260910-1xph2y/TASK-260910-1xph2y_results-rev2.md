# F1 rework evidence, revision 2 candidate

Task: TASK-260910-1xph2y. Role: doc-writer implementer. Ready for independent review, not accepted.
Base/HEAD: d019f0e7179520b5c8dcde321c4fe51e04552f58.
Rework baseline: CR revision 1 tree 96b33ebf97e4dfe01a705124219fbec1adf2f5c1.

## Resolution

Preserved attestation and legacy skill substituted fields through frozen v4 schema references. Added explicit local/Git applicability, registry summary equality, non-authorizing evidence, strict substitution refusal before cache/compiler/publication, and currentness/repair/refresh behavior. An exhaustive v4 audit found external build field loss in the same migration; marker v5 now retains both complete v4 driver records with receipt version 3 and explicit wrapper/package equality. The normative migration table accounts for every v4 top-level field. Legacy v1 behavior and frozen artifacts remain unchanged.

Schema examples cover attested network/configured Git, deprecated status without key, legacy substitution, local mixed builds, and external local committed-source substitution. Negatives cover malformed attestation/substitution, local misuse, all 16 missing required external fields, commit length, receipt version and typed substitution state. Schema fixture hashes are explicitly synthetic, not claimed linked executable evidence.

## Checks rerun by this producer

- Existing task venv import readiness: exit 0. No installations.
- Exact documented draft command, extracted to .temp/rework-1xph2y/validate-draft.py: final exit 0; 102/102 schema cases, 82/82 negatives, 7/7 wire schemas, 3/3 snapshot byte vectors.
- Complete v4 migration audit: 25/25 fields and 2/2 build arms.
- Same Draft202012Validator.is_valid entrypoint: 18/18 narrowed refusal mutants detected (2 applicability and 16 required external fields). This is a specification schema call site, not manager production execution.
- PATH="$PWD/.temp/source-contract/venv/bin:$PATH" make validate: exit 0; 60 schemas, 1047 vectors, 227 Python tests, Go tooling tests pass. All rerun here; no pass accepted solely from previous evidence.
- Initial new draft mutation harness run: exit 1 (draft-01.log). The missing-object_format case also failed the existing commit-length conditional, hiding that mutant. Refined the single-defect structural fixture to a 64-character commit so the missing format gate is isolated; subsequent draft-02 and final draft-03 exit 0. No production refusal was weakened to pass.
- Boundary checks recorded in boundaries-01.log: whitespace, frozen JSON schema/vector/release/tool files, and real index unchanged. HEAD..main is 0.

## Limits and lifecycle

Manager semantic execution remains 0 cases, unverified. Added semantic requirements cover stale/unreadable/revoked/mismatching attestation, strict substitution with omitted marker evidence, currentness mismatches and external receipt comparisons. They are requirements for future real planning/status/repair/refresh tests, not executed manager tests. No new implementation support is claimed.

Review verdict and task-scoped adversarial reproduction were read; the old reproduction intentionally expects the unfixed schema and old candidate tree, so it was not misreported as a current pass. Its entrypoint and F1 cases are now covered by the shipped draft checks.

No commit, tag, push, PR, checkpoint, integration, switch, rebase or merge was run. No self-acceptance. Current files remain uncommitted for a new CR and independent reviewer. Rework finding recorded in task notes/logbook. Raw logs are attached separately.

## Exact rework paths (41; compared with CR revision 1)

- `conformance/draft-sources-v1/README.md`
- `conformance/draft-sources-v1/index.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-attestation-empty-registry.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-attestation-extra.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-attestation-key.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-attestation-missing-status.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-attestation-unknown-status.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-external-bad-commit.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-external-legacy-receipt.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-external-missing-artifact_path.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-external-missing-artifact_sha256.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-external-missing-build_source.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-external-missing-cache_key.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-external-missing-commit.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-external-missing-declared_identity.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-external-missing-declared_locked_commit.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-external-missing-descriptor_target.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-external-missing-driver.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-external-missing-effective_identity.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-external-missing-execution_policy.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-external-missing-object_format.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-external-missing-receipt_schema_version.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-external-missing-receipt_sha256.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-external-missing-repository.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-external-missing-substituted.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-external-missing-substitution.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-external-unexpected-substitution.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-local-attestation.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-local-substituted.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-substitution-boolean.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/invalid-substitution-empty.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/valid-attested-configured.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/valid-attested-network.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/valid-deprecated-network.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/valid-external-substitution.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/valid-legacy-substitution.json`
- `conformance/draft-sources-v1/schema-cases/install-marker-v5/valid-local-mixed-builds.json`
- `conformance/draft-sources-v1/semantic-cases.json`
- `protocol/skillfile-sources.md`
- `schemas/draft-sources-v1/README.md`
- `schemas/draft-sources-v1/install-marker-v5.schema.json`
