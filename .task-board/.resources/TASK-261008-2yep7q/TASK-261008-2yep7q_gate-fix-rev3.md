RECOVERY RUN — CR REV3 VALIDATION FIX (TASK-261008-2yep7q)

CR rev3 validation failed at command 1/1 (spec-gate.sh) with exactly 1 failure out of 689 tests:
test_validate.WorkflowRegenerationScopeTests.test_workflows_match_makefile_generated_file_inventory
Makefile scope has 5 entries (incl. conformance/draft-sources-v2/manifest.json); the test-side GENERATED_FILE_INVENTORY still listed 4.

ROOT CAUSE: rev3 added the draft manifest to the regenerate diff scope in Makefile + ci.yml + release.yml (the Go generator owns the draft manifest, so the scope extension is correct), but did not update the test-side inventory mirror. Rev3s full 689-run predated the regenerate-path change, so the mismatch surfaced only in CR validation. No product-code or corpus issue.

FIX (one line, tools/test_validate.py:1797-1803): added "conformance/draft-sources-v2/manifest.json" to GENERATED_FILE_INVENTORY in exact Makefile order. No other file touched this run.

EVIDENCE (dev venv, real exit codes; suite sharded because the full discover takes ~15 min and a single shell call is time-bounded; no product-code change between shards):
- python3 -B tools/validate.py -> exit 0 — validated 73 schemas and 1294 vector files
- focused (WorkflowRegenerationScope + SkillfileSourcesSuiteManifest + ManifestDependencyDirectory + AcceptedSourceCorpusIsolation + CandidateMetadata) -> 18 tests OK, exit 0
- test_validate shard B (8 classes) -> 56 OK, exit 0
- test_validate shard C (8 classes) -> 219 OK, exit 0
- test_validate shard D (8 classes) -> 103 OK, exit 0 (668s, CPU-bound)
- test_validate shard E (8 classes) -> 178 OK, exit 0
- other files (allowed_signers, implementation_coverage, muse, skillfile_sources_independence, verify_release_commit, verify_release_merge_policy) -> 75 OK, exit 0
- test_release_gate -> 40 OK, exit 0
- TOTAL: 18+56+219+103+178+75+40 = 689/689, matching the CR gate count exactly; the previously failing inventory test is green
- go test ./tools/... -> exit 0 (ok generate-vectors 3.049s)
- regenerate-check: validate.py --release-history-only exit 0; go run generate-vectors exit 0; pre/post shasum identical for both manifests + candidate.json; git diff gate on the 5-path scope exit 0
- python3 tools/verify_skillfile_sources_independence.py -> exit 0 (89/89, 31/31)
- gofmt -l tools/ clean; git diff --check clean
- NOT run: pytest (not installed; unittest discovery is the Makefile/CI gate); lychee (CI-only; no new external URLs)
- Working tree left uncommitted: 13 modified + 3 new paths (the rev3 set plus the one-line fix)
