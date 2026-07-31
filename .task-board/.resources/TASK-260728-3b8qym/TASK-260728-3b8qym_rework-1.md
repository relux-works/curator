# TASK-260728-3b8qym developer rework evidence — review cycle 1

## Verdict blockers closed

1. **Normative CCJ-1 hashes**
   - `build-receipt-v2.json.cache_key` is now
     `sha256:012564909df8f333004eb5aec867210d8973c74d9b71948e1f4fdb0d00c76559`,
     computed as SHA-256 of CCJ-1 `receipt.input`.
   - The mixed marker and mixed-build plan now carry
     `receipt_sha256: sha256:d86ca882ff336480a277010e886bea46767bf9a23d928fe3b43ee1009c3bd0ed`,
     computed from the full corrected CCJ-1 receipt.
   - Go and Python semantic checks recompute both relationships and include
     mutation assertions that reject a false cache key or marker receipt hash.
   - All generated receipt-v2 examples now derive their cache key from their
     actual input instead of using placeholder bytes.

2. **Source-covering operation order**
   - Added executable `external-source-dry-run` and `external-audit-only`
     cases with exact ordered phases, results, and no-mutation properties.
   - Cache hit, cache miss, dry-run, repair, and audit-only paths are checked
     independently against the global whole-snapshot order.
   - Every source-covering path proves exact acquisition, whole-snapshot
     validation, and independent external audit before any cache lookup or
     compiler phase. Syntax-only offline validation remains disjoint and
     makes no source, audit, cache, or mutation claim.

3. **Cryptographic pack/index negatives**
   - `reject-index-checksum-mismatch` now references
     `valid-empty-pack-v2-sha1`, carries complete pack/index bytes, and declares
     an executable final-index-byte XOR mutation.
   - `reject-pack-hash-family-mismatch` now references the same exact SHA-1
     bytes and declares the deterministic SHA-1-to-SHA-256 repository-format
     substitution.
   - Python and Go harnesses parse PACK and index-v2 headers, version/count,
     256-entry fanout, pack trailers, embedded pack checksum, index trailer,
     filenames, and declared hash family. They prove that the checksum case
     changes exactly one intended byte and that the family case is valid SHA-1
     data but invalid under its SHA-256 declaration. Both retain the stable
     `build_repository_local_object_format_unsupported` error.

## Files and generated surfaces

- Generator and Go harness:
  `tools/generate-vectors/main.go`,
  `tools/generate-vectors/main_test.go`.
- Python validator and independent tests:
  `tools/validate.py`, `tools/test_validate.py`.
- Author/operator guidance:
  `docs/external-build-repositories.md`, `conformance/README.md`.
- Exact expected bytes:
  `conformance/v1/expected/external-repository/build-receipt-v2.json`,
  `install-marker-v3-mixed.json`, `mixed-build-plan.json`.
- Executable fixtures/vectors:
  `conformance/v1/fixtures/external-repository/pack-index.json`,
  `conformance/v1/vectors/external-repository-lifecycle.json`.
- Regenerated receipt-v2 and marker-v3 schema cases, suite manifest, and
  `release/1.0.0-rc.5.json`.

The largest external-repository fixture is `pack-index.json` at 18,377 bytes,
below the enforced 65,536-byte ceiling. Suite inventory remains 42 schemas and
411 manifest-listed files.

## Exact release and downstream pins

- Assigned worktree HEAD remains the accepted baseline:
  `57c1f56846d221ecc55786bd3c2467ec32f11730`.
- Exact rc.5 downstream candidate protocol pin:
  `sha256:33fd7aed900ec4f9f9c72be6298823a16674116b18e3b8efdd2d147574dba2b8`
  (SHA-256 of `conformance/v1/manifest.json`).
- rc.5 release metadata SHA-256:
  `1f4f6fcc7c57f26f050d730f7b84091cf3f58970871bd7f511bea03f4b2b8f31`.
- Exact receipt file SHA-256:
  `28d3340295731b4271ceb002ef3fc063bbf187237f4012e2269b0c702bf78622`.
- Exact mixed marker file SHA-256:
  `0088cb9536eaacea2c087efeb70a5bc453c923020c5c01d3a87146f6e85773e3`.
- Exact pack/index fixture SHA-256:
  `09768bac46eb9966e76baf7fbe613ed87a88dd665962f4d3c18f20dc6a98791c`.
- Exact lifecycle vector SHA-256:
  `175d709f183775b22ed3db27bc923ca78d6394d93a13096b5d6890c509aab072`.
- Downstream environment remains `CURATOR_CONFORMANCE_ROOT`;
  `committed_release_pin_advanced` remains `false`.
- Claim-v3 candidate claims remain empty. macOS and Windows remain pending
  downstream native evidence. Linux remains explicitly excluded until
  `TASK-260728-1skseh`.
- The obsolete `sha256:30a64ed0...` candidate pin is absent from the revised
  worktree (`rg` returned exit 1 because there were no matches).

## Validation evidence

All listed green gates were executed directly and returned exit code 0:

- `go run ./tools/generate-vectors -root .`
- `go test ./tools/generate-vectors`
- task-local Python `-B -m unittest discover -s tools -p 'test_*.py'`
  (18 tests)
- task-local Python `tools/validate.py`
  (`validated 42 schemas and 411 vector files`)
- `go test ./tools/...`
- `PATH=<validation-venv>:$PATH make validate`
- `go vet ./tools/...`
- `test -z "$(gofmt -l tools)"`
- task-local Python `-m compileall -q tools`
- `git diff --check`
- `git diff --cached --quiet`
- exact pinned-HEAD assertion
- `go build ./tools/...` (the generated binary was moved to task temp)
- byte comparison of the assigned worktree and clean release probe
- clean-probe `make regenerate-check`, first run
- clean-probe `make regenerate-check`, second consecutive run
- clean-probe `make release-check VERSION=1.0.0-rc.5`
  (validation, Python, Go, deterministic regeneration, exact metadata pin,
  clean checkout, version, and candidate release gate all passed)
- post-gate clean-probe status.

The clean release probe used an isolated scratch repository only. Its synthetic
commit `8554df212e952ae533afc8b4dc3e370a7b092d73` is not a source commit,
protocol pin, implementation pin, or downstream pin. The assigned worktree was
not staged or committed.

Expected-red and setup evidence:

- The first targeted `go test ./tools/generate-vectors` returned exit 1 because
  generated fixtures still had the pre-rework shape; after regeneration the
  same command returned exit 0.
- The first scratch-probe commit attempt was rejected by the workstation's
  passphrase-protected signing key. The setup wrapper itself ended at exit 0
  after printing the probe path, but Git reported the internal commit failure.
  The retry disabled signing only in the disposable probe and returned exit 0.
  No source repository state was affected.

## Scope and compatibility

- The rework delta against the prior reviewed candidate is limited to the
  rejected CCJ-1 receipt/marker oracles, their generated receipt-v2/marker-v3
  examples, executable pack negatives, lifecycle order vectors, harnesses,
  docs, manifest, and release metadata.
- Frozen schemas 1-6, go-v1, receipt-v1, markers v1-v2, claim-v2, and rc.4
  artifacts remain guarded by the Go compatibility tests and passed.
- No Curator/csk manager implementation, generic language driver, real remote
  contact, signer implementation, or fabricated platform evidence was added.
