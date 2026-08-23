# TASK-260720-q5oy3o results

## Provenance and scope

- Worktree: /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-q5oy3o/curator-spec-worktree
- Exact fetched origin/main base: 57c1f56846d221ecc55786bd3c2467ec32f11730.
- Imported the accepted TASK-260720-3lo9jc product tree from /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3lo9jc/curator-spec-worktree. A checksum-aware mirror comparison is empty outside the four owned files.
- Task-only files: README.md, COMPATIBILITY.md, CHANGELOG.md, and RELEASE.md. No implementation pin, schema, vector, conformance manifest, decision, test, or generated artifact was changed by this task. The real Git index has no staged paths.

## Release metadata

- README and CHANGELOG identify protocol 1.0.0-rc.4 with actual local change date 2026-07-21.
- README indexes manifest schema 6, marker v2, receipt v1, claim v2, the generated rc.4 suite, and decision 0004.
- COMPATIBILITY contains explicit reader and writer rows for manifest schemas 1 through 6, marker v1 and v2, receipt v1, claim v1 rc.3, and claim v2 rc.4. It requires rejection by unsupported old readers and no package migration for existing schema 1 through 5 packages.
- CHANGELOG records the closed declarative go-v1 surface, context exclusion, compiler-free dry-run, transaction and rollback rules, generated conformance additions, and compiler/cache security tradeoffs.
- RELEASE keeps all rc.4 evidence boxes unchecked and requires exact-candidate validation, deterministic regeneration, claim-v2 suite identity, cross-manager and cross-platform interoperability, real downstream commits and pins, reviews, tag/signature/checksum/archive/attestation evidence.
- Version validation agrees across README, CHANGELOG, conformance/v1/manifest.json, and conformance-claim-v2. Claim v1 remains frozen at rc.3. The locally computed manifest identity is sha256:70cd274150d629f16ca04f2073a8eae5a0f3fff4584f905d807775040af21eae; this is suite identity evidence, not a release archive checksum or attestation.

## Verification

- make validate: PASS; 35 schemas, 189 vector files, 27 Python tests, and go test ./tools/... passed.
- make regenerate-check: PASS using a disposable alternate index seeded from the accepted uncommitted composite; the real index stayed untouched.
- Determinism: aggregate conformance/v1 hash was 41aa37774478c26377455877ee79ef74f8cb5cf5562ea5b1501e5c94fe9c1fa0 before generation and after each of two consecutive regeneration checks. A final post-edit regenerate-check also passed.
- Direct release metadata and rc.4 protocol-artifact validation: PASS.
- go vet ./tools/..., gofmt check, and git diff --check: PASS.

## Unmet external evidence

No downstream schema 6 manager release, new implementation pin, cross-platform interoperability result, claim-v2 file, independent review acceptance, candidate tag, signature, release archive checksum, or provenance attestation is claimed. Those prerequisites remain explicitly unchecked in RELEASE.md.
